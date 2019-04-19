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
	"errors"
	"fmt"
	"github.com/ontio/ontology/core/xshard_types"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/ontio/ontology/smartcontract/event"
)

const (
	TxUnknownState = iota
	TxExec
	TxWait
	TxPrepared
	TxAbort
	TxCommit
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

type ExecuteState uint8

const ExecYielded = ExecuteState(1)
const ExecPrepared = ExecuteState(2)
const ExecCompleted = ExecuteState(3)

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
type XShardReqMsg struct {
	SourceShardID common.ShardID
	SourceHeight  uint32
	TargetShardID common.ShardID
	SourceTxHash  common.Uint256
	Req           *xshard_types.XShardTxReq
}

type TxState struct {
	State         int
	TxID          ShardTxID // cross shard tx id: userTxHash+notify1+notify2...
	Shards        map[common.ShardID]int
	TxPayload     []byte
	NumNotifies   uint32
	ShardNotifies []*xshard_types.XShardNotify
	NextReqID     uint32
	InReqResp     map[common.ShardID][]*XShardTxReqResp // todo: request id may conflict
	TotalInReq    uint32
	OutReqResp    []*XShardTxReqResp
	PendingReq    *XShardReqMsg
	ExecuteState
	Result    []byte
	ResultErr error
	WriteSet  *overlaydb.MemDB
	Notify    *event.ExecuteNotify
}

type ShardTxInfo struct {
	Index uint32
	State *TxState
}

type ShardTxID string // cross shard tx id: userTxHash+notify1+notify2...
//
// ShardTxStateMap
// stores all intermediate states of cross-shard transactions
//
type ShardTxStateMap struct {
	TxStates map[ShardTxID]*TxState
}

var shardTxStateTable = ShardTxStateMap{
	TxStates: make(map[ShardTxID]*TxState),
}

func (self *TxState) GetTxShards() []common.ShardID {
	shards := make([]common.ShardID, 0, len(self.Shards))
	for id := range self.Shards {
		shards = append(shards, id)
	}
	return shards
}

//
// GetTxShards
// get shards which participant with the procession of transaction
//
func GetTxShards(txid ShardTxID) ([]common.ShardID, error) {
	if state, present := shardTxStateTable.TxStates[txid]; present {
		shards := make([]common.ShardID, 0, len(state.Shards))
		for id := range state.Shards {
			shards = append(shards, id)
		}
		return shards, nil
	}

	return nil, ErrNotFound
}

func (self *TxState) AddTxShard(id common.ShardID) error {
	if state, present := self.Shards[id]; !present {
		self.Shards[id] = TxExec
	} else if state != TxExec {
		return ErrInvalidTxState
	}

	return nil
}

func (self *TxState) IsTxExecutionPaused(tx common.Uint256) bool {
	return self.State != TxExec
}

func IsTxExecutionPaused(tx common.Uint256) (bool, error) {
	txState, err := GetTxState(tx)
	if err != nil {
		return false, err
	}

	return txState.State != TxExec, nil
}

func (txState *TxState) SetTxExecutionPaused() {
	switch txState.State {
	case TxExec:
		txState.State = TxWait
	}
}

func SetTxExecutionPaused(tx common.Uint256) error {
	txState, err := GetTxState(tx)
	if err != nil {
		return err
	}
	switch txState.State {
	case TxExec:
		txState.State = TxWait
	}
	return nil
}

func (txState *TxState) SetTxExecutionContinued() {
	switch txState.State {
	case TxWait:
		txState.State = TxExec
	}
}

func SetTxExecutionContinued(tx common.Uint256) error {
	txState, err := GetTxState(tx)
	if err != nil {
		return err
	}
	switch txState.State {
	case TxWait:
		txState.State = TxExec
	}
	return nil
}

func (self *TxState) SetShardPrepared(shardId common.ShardID) error {
	if _, ok := self.Shards[shardId]; !ok {
		return fmt.Errorf("invalid shard ID %d, in tx commit", shardId)
	}
	self.Shards[shardId] = TxPrepared
	return nil
}

func (self *TxState) IsTxCommitReady() bool {
	if self.State != TxPrepared {
		return false
	}
	for _, state := range self.Shards {
		if state != TxPrepared {
			return false
		}
	}
	return true
}

func (self *TxState) Clone() *TxState {
	txs := &TxState{
		State:      self.State,
		Shards:     make(map[common.ShardID]int),
		TxPayload:  self.TxPayload,
		NextReqID:  self.NextReqID,
		Result:     make([]byte, len(self.Result)),
		InReqResp:  make(map[common.ShardID][]*XShardTxReqResp),
		PendingReq: self.PendingReq,
		ResultErr:  self.ResultErr,
		WriteSet:   nil,
		Notify:     self.Notify,
	}

	for k, v := range self.Shards {
		txs.Shards[k] = v
	}
	// todo: need deep clone?
	for k, v := range self.InReqResp {
		for _, res := range v {
			txs.InReqResp[k] = append(txs.InReqResp[k], res)
		}
	}

	for _, v := range self.OutReqResp {
		txs.OutReqResp = append(txs.OutReqResp, v)
	}
	//todo: need clone?
	txs.WriteSet = self.WriteSet

	return txs
}

// CreateTxState
// If txState available, return it.  Otherwise, Create txState.
func CreateTxState(tx ShardTxID) *TxState {
	if state, present := shardTxStateTable.TxStates[tx]; present {
		return state
	}
	state := &TxState{
		State:  TxExec,
		Shards: make(map[common.ShardID]int),
		TxID:   tx,
	}
	shardTxStateTable.TxStates[tx] = state
	return state
}

func PutTxState(txid ShardTxID, state *TxState) {
	shardTxStateTable.TxStates[txid] = state
}

func GetTxState(tx common.Uint256) (*TxState, error) {
	//todo:
	txID := ShardTxID(string(tx[:]))
	if state, present := shardTxStateTable.TxStates[txID]; present {
		return state, nil
	}
	return nil, ErrNotFound
}

func (self *TxState) SetTxPrepared() error {
	if self.State != TxExec {
		return fmt.Errorf("invalid state to prepared: %s", self.State)
	}

	self.State = TxPrepared
	return nil
}

func SetTxResult(tx common.Uint256, result []byte, resultErr error) error {
	txState, err := GetTxState(tx)
	if err != nil {
		return err
	}

	txState.Result = result
	txState.ResultErr = resultErr
	return nil
}

func (self *TxState) VerifyStates() error {
	// TODO
	return nil
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
