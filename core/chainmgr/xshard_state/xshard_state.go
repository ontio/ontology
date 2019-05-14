/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package xshard_state

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/ontio/ontology/core/xshard_types"
	"github.com/ontio/ontology/smartcontract/event"
)

const MaxRemoteReqPerTx = 8

var (
	ErrNotFound            = errors.New("not found")
	ErrInvalidTxState      = errors.New("invalid transaction state")
	ErrTooMuchRemoteReq    = errors.New("too much remote request")
	ErrInvalidRemoteRsp    = errors.New("invalid remotes response")
	ErrMismatchedTxPayload = errors.New("mismatched Tx Payload")
	ErrMismatchedRequest   = errors.New("mismatched request")
	ErrMismatchedResponse  = errors.New("mismatched response")
)

var ErrYield = errors.New("transaction execution yielded")

type ExecState uint8

const ExecNone = ExecState(0)
const ExecYielded = ExecState(1)
const ExecPrepared = ExecState(2)
const ExecCommited = ExecState(3)
const ExecAborted = ExecState(4)

func state2string(state ExecState) string {
	switch state {
	case ExecNone:
		return "none"
	case ExecPrepared:
		return "prepared"
	case ExecAborted:
		return "aborted"
	case ExecCommited:
		return "commited"
	case ExecYielded:
		return "yielded"
	}
	return "unkown"
}

//
// TxState
// stores intermediate states of one cross-shard transaction
//
// * State: current execution state : exec -> (wait) -> prepare -> prepared -> commit
// * Shards: store shards which participant with the transaction
// * TxPayload: serialized transaction
// * NextReqID: index of next transactional remote request
// * Reqs: sent remote requests in this transaction
// * Rsps: received remote responses in this transaction
// * Result: execution result of the transaction
// * ResultErr: execution error of the transaction
// * WriteSet:
// * Notify:
//
type XShardTxReqResp struct {
	Req   *xshard_types.XShardTxReq
	Resp  *xshard_types.XShardTxRsp
	Index uint32
}

type TxState struct {
	TxID           xshard_types.ShardTxID       // cross shard tx id: userTxHash+notify1+notify2...
	Shards         map[common.ShardID]ExecState // shards in this shard transaction, not include notification
	NumNotifies    uint32
	ShardNotifies  []*xshard_types.XShardNotify
	NextReqID      uint32
	InReqResp      map[common.ShardID][]*XShardTxReqResp // todo: request id may conflict
	PendingInReq   *xshard_types.XShardTxReq
	TotalInReq     uint32
	OutReqResp     []*XShardTxReqResp
	TxPayload      []byte
	PendingOutReq  *xshard_types.XShardTxReq
	PendingPrepare *xshard_types.XShardPrepareMsg
	ExecState      ExecState
	Result         []byte
	ResultErr      string
	WriteSet       *overlaydb.MemDB
	Notify         *event.ExecuteNotify
}

// fot debug
func (self *TxState) DumpInfo() string {
	//buf, _ := json.MarshalIndent(self, "", "\t")
	//return string(buf)
	requestShards := ""
	for shard, state := range self.Shards {
		requestShards += fmt.Sprintf("  shard%d: %s\n", shard.ToUint64(), state2string(state))
	}

	return fmt.Sprintf(`
	txstate:{
	txid:%x,
	reqShards[%d]: %s
	payload:%x,
	num notifies:%d,
	
	}`, self.TxID, len(self.Shards), requestShards, self.TxPayload, self.NumNotifies)

}

func (self *TxState) Deserialization(source *common.ZeroCopySource) error {
	id, _, irr, eof := source.NextString()
	if irr {
		return common.ErrIrregularData
	}
	self.TxID = xshard_types.ShardTxID(id)
	lenShards, _, irr, eof := source.NextVarUint()
	if irr {
		return common.ErrIrregularData
	}
	self.Shards = make(map[common.ShardID]ExecState)
	for i := uint64(0); i < lenShards; i++ {
		id, err := source.NextShardID()
		if err != nil {
			return err
		}
		state, eof := source.NextUint32()
		if eof {
			return io.ErrUnexpectedEOF
		}

		self.Shards[id] = ExecState(state)
	}
	self.NumNotifies, eof = source.NextUint32()
	len, _, irr, eof := source.NextVarUint()
	if irr {
		return common.ErrIrregularData
	}

	for i := uint64(0); i < len; i++ {
		notify := &xshard_types.XShardNotify{}
		err := notify.Deserialization(source)
		if err != nil {
			return err
		}

		self.ShardNotifies = append(self.ShardNotifies, notify)
	}
	self.NextReqID, eof = source.NextUint32()

	self.InReqResp = make(map[common.ShardID][]*XShardTxReqResp)
	len, _, irr, eof = source.NextVarUint()
	if irr {
		return common.ErrIrregularData
	}
	for i := uint64(0); i < len; i++ {
		id, err := source.NextShardID()
		if err != nil {
			return err
		}
		inLen, _, irr, _ := source.NextVarUint()
		if irr {
			return common.ErrIrregularData
		}

		var inReqResp []*XShardTxReqResp
		for j := uint64(0); j < inLen; j++ {
			req := &xshard_types.XShardTxReq{}
			err := req.Deserialization(source)
			if err != nil {
				return err
			}
			resp := &xshard_types.XShardTxRsp{}
			err = resp.Deserialization(source)
			if err != nil {
				return err
			}
			index, eof := source.NextUint32()
			if eof {
				return io.ErrUnexpectedEOF
			}
			inReqResp = append(inReqResp, &XShardTxReqResp{
				Index: index,
				Req:   req,
				Resp:  resp,
			})
		}

		self.InReqResp[id] = inReqResp
	}

	hasPending, irr, eof := source.NextBool()
	if irr {
		return common.ErrIrregularData
	}
	if hasPending {
		inReq := &xshard_types.XShardTxReq{}
		err := inReq.Deserialization(source)
		if err != nil {
			return err
		}
		self.PendingInReq = inReq
	}
	self.TotalInReq, eof = source.NextUint32()
	lenOutReqResp, _, irr, eof := source.NextVarUint()
	if irr {
		return common.ErrIrregularData
	}
	for i := uint64(0); i < lenOutReqResp; i++ {
		req := &xshard_types.XShardTxReq{}
		err := req.Deserialization(source)
		if err != nil {
			return err
		}
		resp := &xshard_types.XShardTxRsp{}
		err = resp.Deserialization(source)
		if err != nil {
			return err
		}
		index, eof := source.NextUint32()
		if eof {
			return io.ErrUnexpectedEOF
		}
		self.OutReqResp = append(self.OutReqResp, &XShardTxReqResp{
			Index: index,
			Req:   req,
			Resp:  resp,
		})
	}

	self.TxPayload, _, irr, eof = source.NextVarBytes()
	if irr {
		return common.ErrIrregularData
	}
	hasPending, irr, eof = source.NextBool()
	if irr {
		return common.ErrIrregularData
	}
	if hasPending {
		self.PendingOutReq = &xshard_types.XShardTxReq{}
		err := self.PendingOutReq.Deserialization(source)
		if err != nil {
			return err
		}
	}
	hasPending, irr, eof = source.NextBool()
	if irr {
		return common.ErrIrregularData
	}
	if hasPending {
		prepMsg := &xshard_types.XShardPrepareMsg{}
		err := prepMsg.Deserialization(source)
		if err != nil {
			return err
		}
		self.PendingPrepare = prepMsg
	}
	st, eof := source.NextUint8()
	self.ExecState = ExecState(st)
	self.Result, _, irr, eof = source.NextVarBytes()
	if irr {
		return common.ErrIrregularData
	}
	self.ResultErr, _, irr, eof = source.NextString()
	if irr {
		return common.ErrIrregularData
	}
	if self.WriteSet == nil {
		self.WriteSet = overlaydb.NewMemDB(1024, 10)
	}
	err := self.WriteSet.Deserialization(source)
	if err != nil {
		return err
	}

	buf, _, irr, eof := source.NextVarBytes()
	if irr {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}

	return json.Unmarshal(buf, &self.Notify)
}

func (self *TxState) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(string(self.TxID))
	type shardState struct {
		shard common.ShardID
		state ExecState
	}
	var shards []shardState
	for id, state := range self.Shards {
		shards = append(shards, shardState{shard: id, state: state})
	}
	sort.Slice(shards, func(i, j int) bool {
		return shards[i].shard.ToUint64() < shards[j].shard.ToUint64()
	})
	sink.WriteVarUint(uint64(len(shards)))
	for _, s := range shards {
		sink.WriteUint64(s.shard.ToUint64())
		sink.WriteUint32(uint32(s.state))
	}

	sink.WriteUint32(self.NumNotifies)
	sink.WriteVarUint(uint64(len(self.ShardNotifies)))
	for _, notify := range self.ShardNotifies {
		notify.Serialization(sink)
	}
	sink.WriteUint32(self.NextReqID)

	type shardReqResp struct {
		shard     common.ShardID
		InReqResp []*XShardTxReqResp
	}
	var shardInReqResp []shardReqResp
	for shard, re := range self.InReqResp {
		shardInReqResp = append(shardInReqResp, shardReqResp{shard: shard, InReqResp: re})
	}
	sort.Slice(shardInReqResp, func(i, j int) bool {
		return shardInReqResp[i].shard.ToUint64() < shardInReqResp[j].shard.ToUint64()
	})

	sink.WriteVarUint(uint64(len(shardInReqResp)))
	for _, s := range shardInReqResp {
		sink.WriteUint64(s.shard.ToUint64())
		sink.WriteVarUint(uint64(len(s.InReqResp)))
		for _, reqResp := range s.InReqResp {
			reqResp.Req.Serialization(sink)
			reqResp.Resp.Serialization(sink)
			sink.WriteUint32(reqResp.Index)
		}
	}
	sink.WriteBool(self.PendingInReq != nil)
	if self.PendingInReq != nil {
		self.PendingInReq.Serialization(sink)
	}
	sink.WriteUint32(self.TotalInReq)
	sink.WriteVarUint(uint64(len(self.OutReqResp)))
	for _, reqResp := range self.OutReqResp {
		reqResp.Req.Serialization(sink)
		reqResp.Resp.Serialization(sink)
		sink.WriteUint32(reqResp.Index)
	}
	sink.WriteVarBytes(self.TxPayload)
	sink.WriteBool(self.PendingOutReq != nil)
	if self.PendingOutReq != nil {
		self.PendingOutReq.Serialization(sink)
	}
	sink.WriteBool(self.PendingPrepare != nil)
	if self.PendingPrepare != nil {
		self.PendingPrepare.Serialization(sink)
	}
	sink.WriteUint8(uint8(self.ExecState))
	sink.WriteVarBytes(self.Result)
	sink.WriteString(self.ResultErr)
	if self.WriteSet == nil {
		self.WriteSet = overlaydb.NewMemDB(1024, 10)
	}
	self.WriteSet.Serialization(sink)
	buf, _ := json.Marshal(self.Notify)
	sink.WriteVarBytes(buf)
}

type ShardTxInfo struct {
	Index uint32
	State *TxState
}

func (self *TxState) GetTxShards() []common.ShardID {
	shards := make([]common.ShardID, 0, len(self.Shards))
	for id := range self.Shards {
		shards = append(shards, id)
	}
	return shards
}

func (self *TxState) IsCommitReady() bool {
	if self.ExecState != ExecPrepared {
		return false
	}
	for _, state := range self.Shards {
		if state != ExecPrepared {
			return false
		}
	}
	return true
}

func (self *TxState) AddTxShard(id common.ShardID) error {
	if state, present := self.Shards[id]; !present {
		self.Shards[id] = ExecNone
	} else if state != ExecNone {
		return ErrInvalidTxState
	}

	return nil
}

func (self *TxState) SetShardPrepared(shardId common.ShardID) error {
	if _, ok := self.Shards[shardId]; !ok {
		return fmt.Errorf("invalid shard ID %d, in tx commit", shardId)
	}
	self.Shards[shardId] = ExecPrepared
	return nil
}

// CreateTxState
// If txState available, return it.  Otherwise, Create txState.
func CreateTxState(tx xshard_types.ShardTxID) *TxState {
	state := &TxState{
		Shards:    make(map[common.ShardID]ExecState),
		InReqResp: make(map[common.ShardID][]*XShardTxReqResp),
		TxID:      tx,
	}
	return state
}

func GetTxContracts(tx common.Uint256) ([]common.Address, error) {
	// TODO
	return []common.Address{}, nil
}

func LockContract(contract common.Address) error {
	// TODO: lock contract if it does not support concurrency (shard-sysmsg contract support concurrency)
	return nil
}

func UnlockContract(contract common.Address) error {
	return nil
}
