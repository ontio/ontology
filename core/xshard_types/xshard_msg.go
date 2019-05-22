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
	"crypto/sha256"
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

func IsXShardMsgEqual(msg1, msg2 CommonShardMsg) bool {
	buf1 := common.SerializeToBytes(msg1)
	buf2 := common.SerializeToBytes(msg2)

	return bytes.Equal(buf1, buf2)
}

type XShardNotify struct {
	ShardMsgHeader
	NotifyID uint32
	Contract common.Address
	Payer    common.Address
	Fee      uint64
	Method   string
	Args     []byte
}

func (msg *XShardNotify) Type() uint32 {
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
	msg.ShardMsgHeader.Serialization(sink)
	sink.WriteUint32(msg.NotifyID)
	sink.WriteAddress(msg.Contract)
	sink.WriteAddress(msg.Payer)
	sink.WriteUint64(msg.Fee)
	sink.WriteString(msg.Method)
	sink.WriteVarBytes(msg.Args)
}

func (msg *XShardNotify) Deserialization(source *common.ZeroCopySource) error {
	err := msg.ShardMsgHeader.Deserialization(source)
	if err != nil {
		return err
	}
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
	ShardMsgHeader
	IdxInTx  uint64
	Contract common.Address
	Payer    common.Address
	Fee      uint64
	Method   string
	Args     []byte
}

func (msg *XShardTxReq) Type() uint32 {
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
	msg.ShardMsgHeader.Serialization(sink)
	sink.WriteUint64(msg.IdxInTx)
	sink.WriteAddress(msg.Contract)
	sink.WriteAddress(msg.Payer)
	sink.WriteUint64(msg.Fee)
	sink.WriteString(msg.Method)
	sink.WriteVarBytes(msg.Args)
}

func (msg *XShardTxReq) Deserialization(source *common.ZeroCopySource) error {
	err := msg.ShardMsgHeader.Deserialization(source)
	if err != nil {
		return err
	}
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
	ShardMsgHeader
	IdxInTx uint64
	FeeUsed uint64
	Error   bool
	Result  []byte
}

func (msg *XShardTxRsp) Type() uint32 {
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
	msg.ShardMsgHeader.Serialization(sink)
	sink.WriteUint64(msg.IdxInTx)
	sink.WriteUint64(msg.FeeUsed)
	sink.WriteBool(msg.Error)
	sink.WriteVarBytes(msg.Result)
}

func (msg *XShardTxRsp) Deserialization(source *common.ZeroCopySource) error {
	err := msg.ShardMsgHeader.Deserialization(source)
	if err != nil {
		return err
	}
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
	ShardMsgHeader
}

func (msg *XShardCommitMsg) Type() uint32 {
	return EVENT_SHARD_COMMIT
}

type XShardAbortMsg struct {
	ShardMsgHeader
}

func (msg *XShardAbortMsg) Type() uint32 {
	return EVENT_SHARD_ABORT
}

type XShardPrepareMsg struct {
	ShardMsgHeader
}

type XShardPreparedMsg struct {
	ShardMsgHeader
}

func (msg *XShardPrepareMsg) Type() uint32 {
	return EVENT_SHARD_PREPARE
}

func (msg *XShardPreparedMsg) Type() uint32 {
	return EVENT_SHARD_PREPARED
}

type ShardMsgHeader struct {
	ShardTxID     ShardTxID
	SourceShardID common.ShardID
	TargetShardID common.ShardID
	SourceTxHash  common.Uint256
}

func (evt *ShardMsgHeader) GetSourceTxHash() common.Uint256 {
	return evt.SourceTxHash
}

func (evt *ShardMsgHeader) GetShardTxID() ShardTxID {
	return evt.ShardTxID
}

func (evt *ShardMsgHeader) GetSourceShardID() common.ShardID {
	return evt.SourceShardID
}

func (evt *ShardMsgHeader) GetTargetShardID() common.ShardID {
	return evt.TargetShardID
}

func (self *ShardMsgHeader) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(string(self.ShardTxID))
	sink.WriteUint64(self.SourceShardID.ToUint64())
	sink.WriteUint64(self.TargetShardID.ToUint64())
	sink.WriteHash(self.SourceTxHash)
}

func (self *ShardMsgHeader) Deserialization(source *common.ZeroCopySource) error {
	s, _, irr, eof := source.NextString()
	if irr {
		return common.ErrIrregularData
	}
	self.ShardTxID = ShardTxID(s)
	shardID, eof := source.NextUint64()
	id, err := common.NewShardID(shardID)
	if err != nil {
		return err
	}
	self.SourceShardID = id
	shardID, eof = source.NextUint64()
	id, err = common.NewShardID(shardID)
	if err != nil {
		return err
	}
	self.TargetShardID = id
	self.SourceTxHash, eof = source.NextHash()
	if eof {
		return io.ErrUnexpectedEOF
	}

	return nil
}

type CommonShardMsg interface {
	GetSourceTxHash() common.Uint256
	GetShardTxID() ShardTxID
	GetSourceShardID() common.ShardID
	GetTargetShardID() common.ShardID
	Type() uint32
	Serialization(sink *common.ZeroCopySink)
	Deserialization(source *common.ZeroCopySource) error
}

func EncodeShardCommonMsgToBytes(msg CommonShardMsg) []byte {
	sink := common.NewZeroCopySink(256)
	sink.WriteUint32(msg.Type())
	msg.Serialization(sink)
	return sink.Bytes()
}

func EncodeShardCommonMsg(sink *common.ZeroCopySink, msg CommonShardMsg) {
	sink.WriteUint32(msg.Type())
	msg.Serialization(sink)
}

func DecodeShardCommonMsg(source *common.ZeroCopySource) (CommonShardMsg, error) {
	msgType, eof := source.NextUint32()
	if eof {
		return nil, io.ErrUnexpectedEOF
	}

	var msg CommonShardMsg
	switch msgType {
	case EVENT_SHARD_NOTIFY:
		msg = &XShardNotify{}
	case EVENT_SHARD_TXREQ:
		msg = &XShardTxReq{}
	case EVENT_SHARD_TXRSP:
		msg = &XShardTxRsp{}
	case EVENT_SHARD_PREPARE:
		msg = &XShardPrepareMsg{}
	case EVENT_SHARD_PREPARED:
		msg = &XShardPreparedMsg{}
	case EVENT_SHARD_COMMIT:
		msg = &XShardCommitMsg{}
	case EVENT_SHARD_ABORT:
		msg = &XShardAbortMsg{}
	default:
		return nil, fmt.Errorf("unsupported msg type:%d", msgType)
	}
	err := msg.Deserialization(source)
	if err != nil {
		return nil, err
	}
	return msg, nil
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

func EncodeShardCommonMsgs(sink *common.ZeroCopySink, msgs []CommonShardMsg) {
	sink.WriteUint32(uint32(len(msgs)))
	for _, msg := range msgs {
		EncodeShardCommonMsg(sink, msg)
	}
}

func DecodeShardCommonMsgs(payload []byte) ([]CommonShardMsg, error) {
	source := common.NewZeroCopySource(payload)
	len, eof := source.NextUint32()
	if eof {
		return nil, io.ErrUnexpectedEOF
	}

	reqs := make([]CommonShardMsg, 0)
	for i := uint32(0); i < len; i++ {
		req, err := DecodeShardCommonMsg(source)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal req-tx: %s", err)
		}
		reqs = append(reqs, req)
	}

	return reqs, nil
}

func NewShardTxID(h common.Uint256) ShardTxID {
	return ShardTxID(string(h[:]))
}

type ShardTxID string // cross shard tx id: userTxHash+notify1+notify2...
func GetShardCommonMsgsHash(msgs []CommonShardMsg) common.Uint256 {
	sink := &common.ZeroCopySink{}
	EncodeShardCommonMsgs(sink, msgs)
	return common.Uint256(sha256.Sum256(sink.Bytes()))
}
