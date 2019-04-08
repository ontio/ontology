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

package xshard_state

import (
	"bytes"
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
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
	Serialize(w io.Writer) error
	Deserialize(r io.Reader) error
}

type XShardNotify struct {
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

func (msg *XShardNotify) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, msg.Contract); err != nil {
		return fmt.Errorf("serialize: write contract failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, msg.Payer); err != nil {
		return fmt.Errorf("serialize: write payer failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, msg.Fee); err != nil {
		return fmt.Errorf("serialize: write fee failed, err: %s", err)
	}
	if err := serialization.WriteString(w, msg.Method); err != nil {
		return fmt.Errorf("serialize: write method failed, err: %s", err)
	}
	if err := serialization.WriteVarBytes(w, msg.Args); err != nil {
		return fmt.Errorf("serialize: write args failed, err: %s", err)
	}
	return nil
}

func (msg *XShardNotify) Deserialize(r io.Reader) error {
	var err error = nil
	if msg.Contract, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read contract failed, err: %s", err)
	}
	if msg.Payer, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read payer failed, err: %s", err)
	}
	if msg.Fee, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read fee failed, err: %s", err)
	}
	if msg.Method, err = serialization.ReadString(r); err != nil {
		return fmt.Errorf("deserialize: read method failed, err: %s", err)
	}
	if msg.Args, err = serialization.ReadVarBytes(r); err != nil {
		return fmt.Errorf("deserialize: read args failed, err: %s", err)
	}
	return nil
}

type XShardTxReq struct {
	IdxInTx  uint64
	Payer    common.Address
	Fee      uint64
	Contract common.Address
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

func (msg *XShardTxReq) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, msg.IdxInTx); err != nil {
		return fmt.Errorf("serialize: write IdxInTx failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, msg.Payer); err != nil {
		return fmt.Errorf("serialize: write payer failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, msg.Fee); err != nil {
		return fmt.Errorf("serialize: write fee failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, msg.Contract); err != nil {
		return fmt.Errorf("serialize: write contract failed, err: %s", err)
	}
	if err := serialization.WriteString(w, msg.Method); err != nil {
		return fmt.Errorf("serialize: write method failed, err: %s", err)
	}
	if err := serialization.WriteVarBytes(w, msg.Args); err != nil {
		return fmt.Errorf("serialize: write args failed, err: %s", err)
	}
	return nil
}

func (msg *XShardTxReq) Deserialize(r io.Reader) error {
	var err error = nil
	if msg.IdxInTx, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read IdxInTx failed, err: %s", err)
	}
	if msg.Payer, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read payer failed, err: %s", err)
	}
	if msg.Fee, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read fee failed, err: %s", err)
	}
	if msg.Contract, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read contract failed, err: %s", err)
	}
	if msg.Method, err = serialization.ReadString(r); err != nil {
		return fmt.Errorf("deserialize: read method failed, err: %s", err)
	}
	if msg.Args, err = serialization.ReadVarBytes(r); err != nil {
		return fmt.Errorf("deserialize: read args failed, err: %s", err)
	}
	return nil
}

type XShardTxRsp struct {
	IdxInTx uint64
	FeeUsed uint64
	Result  []byte
	Error   bool
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

func (msg *XShardTxRsp) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, msg.IdxInTx); err != nil {
		return fmt.Errorf("serialize: write IdxInTx failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, msg.FeeUsed); err != nil {
		return fmt.Errorf("serialize: write fee used failed, err: %s", err)
	}
	if err := serialization.WriteVarBytes(w, msg.Result); err != nil {
		return fmt.Errorf("serialize: write result failed, err: %s", err)
	}
	if err := serialization.WriteBool(w, msg.Error); err != nil {
		return fmt.Errorf("serialize: write error failed, err: %s", err)
	}
	return nil
}

func (msg *XShardTxRsp) Deserialize(r io.Reader) error {
	var err error = nil
	if msg.IdxInTx, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read IdxInTx failed, err: %s", err)
	}
	if msg.FeeUsed, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read fee used failed, err: %s", err)
	}
	if msg.Result, err = serialization.ReadVarBytes(r); err != nil {
		return fmt.Errorf("deserialize: read result failed, err: %s", err)
	}
	if msg.Error, err = serialization.ReadBool(r); err != nil {
		return fmt.Errorf("deserialize: read error failed, err: %s", err)
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

func (msg *XShardCommitMsg) Serialize(w io.Writer) error {
	return utils.WriteVarUint(w, msg.MsgType)
}

func (msg *XShardCommitMsg) Deserialize(r io.Reader) error {
	msgType, err := utils.ReadVarUint(r)
	if err != nil {
		return err
	}
	msg.MsgType = msgType
	return nil
}

type CommonShardMsg struct {
	SourceShardID types.ShardID
	SourceHeight  uint64
	TargetShardID types.ShardID
	SourceTxHash  common.Uint256
	Type          uint64
	Payload       []byte
	Msg           XShardMsg
}

func (evt *CommonShardMsg) GetSourceShardID() types.ShardID {
	return evt.SourceShardID
}

func (evt *CommonShardMsg) GetTargetShardID() types.ShardID {
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

func (evt *CommonShardMsg) Serialize(w io.Writer) error {
	if err := utils.SerializeShardId(w, evt.SourceShardID); err != nil {
		return fmt.Errorf("serialize: write source shard id failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, evt.SourceHeight); err != nil {
		return fmt.Errorf("serialize: write source height failed, err: %s", err)
	}
	if err := utils.SerializeShardId(w, evt.TargetShardID); err != nil {
		return fmt.Errorf("serialize: write target shard id failed, err: %s", err)
	}
	if err := evt.SourceTxHash.Serialize(w); err != nil {
		return fmt.Errorf("serialize: write source tx hash failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, evt.Type); err != nil {
		return fmt.Errorf("serialize: write type failed, err: %s", err)
	}
	if len(evt.Payload) > 0 {
		if err := serialization.WriteVarBytes(w, evt.Payload); err != nil {
			return fmt.Errorf("serialize: write payload failed, err: %s", err)
		}
	} else {
		if err := evt.Msg.Serialize(w); err != nil {
			return fmt.Errorf("serialize: write msg failed, err: %s", err)
		}
	}
	return nil
}

func (evt *CommonShardMsg) Deserialize(r io.Reader) error {
	var err error = nil
	if evt.SourceShardID, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read source shard id failed, err: %s", err)
	}
	if evt.SourceHeight, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read source height failed, err: %s", err)
	}
	if evt.TargetShardID, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read target shard id failed, err: %s", err)
	}
	evt.SourceTxHash = common.Uint256{}
	if err := evt.SourceTxHash.Deserialize(r); err != nil {
		return fmt.Errorf("deserialize: read source tx hash failed, err: %s", err)
	}
	if evt.Type, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read type failed, err: %s", err)
	}
	if evt.Payload, err = serialization.ReadVarBytes(r); err != nil {
		return fmt.Errorf("deserialize: read payload failed, err: %s", err)
	}

	switch evt.Type {
	case EVENT_SHARD_NOTIFY:
		notify := &XShardNotify{}
		if err := notify.Deserialize(bytes.NewBuffer(evt.Payload)); err != nil {
			return err
		}
		evt.Msg = notify
	case EVENT_SHARD_TXREQ:
		req := &XShardTxReq{}
		if err := req.Deserialize(bytes.NewBuffer(evt.Payload)); err != nil {
			return err
		}
		evt.Msg = req
	case EVENT_SHARD_TXRSP:
		rsp := &XShardTxRsp{}
		if err := rsp.Deserialize(bytes.NewBuffer(evt.Payload)); err != nil {
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
		if err := xcommitMsg.Deserialize(bytes.NewBuffer(evt.Payload)); err != nil {
			return err
		}
		if xcommitMsg.Type() != evt.Type {
			return fmt.Errorf("invalid xcommit msg %d vs %d", xcommitMsg.Type(), evt.Type)
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
		if err := req.Deserialize(bytes.NewBuffer(tx)); err != nil {
			return nil, fmt.Errorf("failed to unmarshal req-tx: %s, %s", err, string(tx))
		}
		reqs = append(reqs, req)
	}

	return reqs, nil
}
