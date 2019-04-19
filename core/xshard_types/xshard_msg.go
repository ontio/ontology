/*
 * Copyright (C) 2019 The ontology Authors
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

package xshard_types

import (
	"bytes"
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
)

const (
	EVENT_SHARD_MSG_COMMON = iota + 256

	EVENT_SHARD_NOTIFY
	EVENT_SHARD_TXREQ
	EVENT_SHARD_TXRSP
	EVENT_SHARD_PREPARE
	EVENT_SHARD_PREPARED
	EVENT_SHARD_COMMIT
	EVENT_SHARD_ABORT
)

type XShardMsg interface {
	Type() uint64
	GetContract() common.Address
	GetMethod() string
	GetArgs() []byte
	Serialization(sink *common.ZeroCopySink)
	Deserialization(source *common.ZeroCopySource) error
}

func IsXShardMsgEqual(msg1, msg2 XShardMsg) bool {
	buf1 := common.SerializeToBytes(msg1)
	buf2 := common.SerializeToBytes(msg2)

	return bytes.Equal(buf1, buf2)
}

type XShardNotify struct {
	NotifyID uint32
	Contract common.Address
	Payer    common.Address
	Fee      uint64
	Method   string
	Args     []byte
}

func (msg *XShardNotify) Type() uint64 {
	return EVENT_SHARD_NOTIFY
}

func (msg *XShardNotify) GetContract() common.Address {
	return msg.Contract
}

func (msg *XShardNotify) GetMethod() string {
	return msg.Method
}

func (msg *XShardNotify) GetArgs() []byte {
	return msg.Args
}

func (msg *XShardNotify) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(msg.NotifyID)
	sink.WriteAddress(msg.Contract)
	sink.WriteAddress(msg.Payer)
	sink.WriteUint64(msg.Fee)
	sink.WriteString(msg.Method)
	sink.WriteVarBytes(msg.Args)
}

func (msg *XShardNotify) Deserialization(source *common.ZeroCopySource) error {
	var eof, irregular bool
	msg.NotifyID, eof = source.NextUint32()
	msg.Contract, eof = source.NextAddress()
	msg.Payer, eof = source.NextAddress()
	msg.Fee, eof = source.NextUint64()
	msg.Method, _, irregular, eof = source.NextString()
	if irregular {
		return common.ErrIrregularData
	}
	msg.Args, _, irregular, eof = source.NextVarBytes()
	if irregular {
		return common.ErrIrregularData
	}

	if eof {
		return io.ErrUnexpectedEOF
	}

	return nil
}

type XShardTxReq struct {
	IdxInTx  uint64
	Contract common.Address
	Payer    common.Address
	Fee      uint64
	Method   string
	Args     []byte
}

func (msg *XShardTxReq) Type() uint64 {
	return EVENT_SHARD_TXREQ
}

func (msg *XShardTxReq) GetContract() common.Address {
	return msg.Contract
}

func (msg *XShardTxReq) GetMethod() string {
	return msg.Method
}

func (msg *XShardTxReq) GetArgs() []byte {
	return msg.Args
}

func (msg *XShardTxReq) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(msg.IdxInTx)
	sink.WriteAddress(msg.Contract)
	sink.WriteAddress(msg.Payer)
	sink.WriteUint64(msg.Fee)
	sink.WriteString(msg.Method)
	sink.WriteVarBytes(msg.Args)
}

func (msg *XShardTxReq) Deserialization(source *common.ZeroCopySource) error {
	var eof, irregular bool
	msg.IdxInTx, eof = source.NextUint64()
	msg.Contract, eof = source.NextAddress()
	msg.Payer, eof = source.NextAddress()
	msg.Fee, eof = source.NextUint64()
	msg.Method, _, irregular, eof = source.NextString()
	if irregular {
		return common.ErrIrregularData
	}
	msg.Args, _, irregular, eof = source.NextVarBytes()
	if irregular {
		return common.ErrIrregularData
	}

	if eof {
		return io.ErrUnexpectedEOF
	}

	return nil
}

type XShardTxRsp struct {
	IdxInTx uint64
	FeeUsed uint64
	Error   bool
	Result  []byte
}

func (msg *XShardTxRsp) Type() uint64 {
	return EVENT_SHARD_TXRSP
}

func (msg *XShardTxRsp) GetContract() common.Address {
	return common.ADDRESS_EMPTY
}

func (msg *XShardTxRsp) GetMethod() string {
	return ""
}

func (msg *XShardTxRsp) GetArgs() []byte {
	return msg.Result
}

func (msg *XShardTxRsp) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(msg.IdxInTx)
	sink.WriteUint64(msg.FeeUsed)
	sink.WriteBool(msg.Error)
	sink.WriteVarBytes(msg.Result)
}

func (msg *XShardTxRsp) Deserialization(source *common.ZeroCopySource) error {
	var eof, irregular bool
	msg.IdxInTx, eof = source.NextUint64()
	msg.FeeUsed, eof = source.NextUint64()
	msg.Error, irregular, eof = source.NextBool()
	if irregular {
		return common.ErrIrregularData
	}

	msg.Result, _, irregular, eof = source.NextVarBytes()
	if irregular {
		return common.ErrIrregularData
	}

	if eof {
		return io.ErrUnexpectedEOF
	}

	return nil
}

type XShardCommitMsg struct {
	MsgType uint64
}

func (msg *XShardCommitMsg) Type() uint64 {
	return msg.MsgType
}

func (msg *XShardCommitMsg) GetContract() common.Address {
	return common.ADDRESS_EMPTY
}

func (msg *XShardCommitMsg) GetMethod() string {
	return ""
}

func (msg *XShardCommitMsg) GetArgs() []byte {
	return nil
}

func (msg *XShardCommitMsg) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(msg.MsgType)
}

func (msg *XShardCommitMsg) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	msg.MsgType, eof = source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}

type CommonShardMsg struct {
	SourceShardID common.ShardID
	SourceHeight  uint64
	TargetShardID common.ShardID
	SourceTxHash  common.Uint256
	Type          uint64
	Payload       []byte
	Msg           XShardMsg
}

func (evt *CommonShardMsg) GetSourceShardID() common.ShardID {
	return evt.SourceShardID
}

func (evt *CommonShardMsg) GetTargetShardID() common.ShardID {
	return evt.TargetShardID
}

func (evt *CommonShardMsg) GetHeight() uint64 {
	return evt.SourceHeight
}

func (evt *CommonShardMsg) GetType() uint32 {
	return EVENT_SHARD_MSG_COMMON
}

func (evt *CommonShardMsg) IsTransactional() bool {
	return evt.Msg.Type() != EVENT_SHARD_NOTIFY
}

func (evt *CommonShardMsg) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(evt.SourceShardID.ToUint64())
	sink.WriteUint64(evt.SourceHeight)
	sink.WriteUint64(evt.TargetShardID.ToUint64())
	sink.WriteHash(evt.SourceTxHash)
	sink.WriteUint64(evt.Type)
	if len(evt.Payload) > 0 {
		sink.WriteVarBytes(evt.Payload)
	} else {
		buf := common.SerializeToBytes(evt.Msg)
		sink.WriteVarBytes(buf)
	}
}

func (evt *CommonShardMsg) Deserialization(source *common.ZeroCopySource) error {
	var irregular bool
	shardID, eof := source.NextUint64()
	id, err := common.NewShardID(shardID)
	if err != nil {
		return err
	}
	evt.SourceShardID = id
	evt.SourceHeight, eof = source.NextUint64()
	shardID, eof = source.NextUint64()
	id, err = common.NewShardID(shardID)
	if err != nil {
		return err
	}
	evt.TargetShardID = id
	evt.SourceTxHash, eof = source.NextHash()
	evt.Type, eof = source.NextUint64()
	evt.Payload, _, irregular, eof = source.NextVarBytes()
	if irregular {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}

	switch evt.Type {
	case EVENT_SHARD_NOTIFY:
		notify := &XShardNotify{}
		if err := notify.Deserialization(common.NewZeroCopySource(evt.Payload)); err != nil {
			return err
		}
		evt.Msg = notify
	case EVENT_SHARD_TXREQ:
		req := &XShardTxReq{}
		if err := req.Deserialization(common.NewZeroCopySource(evt.Payload)); err != nil {
			return err
		}
		evt.Msg = req
	case EVENT_SHARD_TXRSP:
		rsp := &XShardTxRsp{}
		if err := rsp.Deserialization(common.NewZeroCopySource(evt.Payload)); err != nil {
			return err
		}
		evt.Msg = rsp
	case EVENT_SHARD_PREPARE:
		fallthrough
	case EVENT_SHARD_PREPARED:
		fallthrough
	case EVENT_SHARD_COMMIT:
		fallthrough
	case EVENT_SHARD_ABORT:
		xcommitMsg := &XShardCommitMsg{}
		if err := xcommitMsg.Deserialization(common.NewZeroCopySource(evt.Payload)); err != nil {
			return err
		}
		if xcommitMsg.Type() != evt.Type {
			return fmt.Errorf("invalid xcommit evt %d vs %d", xcommitMsg.Type(), evt.Type)
		}
		evt.Msg = xcommitMsg
	}

	return nil
}

type CrossShardTx struct {
	Txs [][]byte `json:"txs"`
}

func (this *CrossShardTx) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(uint64(len(this.Txs)))
	for _, tx := range this.Txs {
		sink.WriteVarBytes(tx)
	}
}

func (this *CrossShardTx) Deserialization(source *common.ZeroCopySource) error {
	num, eof := source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.Txs = make([][]byte, num)
	for i := uint64(0); i < num; i++ {
		data, _, irr, eof := source.NextVarBytes()
		if irr {
			return common.ErrIrregularData
		}
		if eof {
			return io.ErrUnexpectedEOF
		}
		this.Txs[i] = data
	}
	return nil
}

func DecodeShardCommonReqs(payload []byte) ([]*CommonShardMsg, error) {
	txs := &CrossShardTx{}
	source := common.NewZeroCopySource(payload)
	if err := txs.Deserialization(source); err != nil {
		return nil, fmt.Errorf("deserialization payload failed, err: %s", err)
	}

	reqs := make([]*CommonShardMsg, 0)
	for _, tx := range txs.Txs {
		req := &CommonShardMsg{}
		if err := req.Deserialization(common.NewZeroCopySource(tx)); err != nil {
			return nil, fmt.Errorf("failed to unmarshal req-tx: %s, %s", err, string(tx))
		}
		reqs = append(reqs, req)
	}

	return reqs, nil
}
