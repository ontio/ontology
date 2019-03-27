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

package shardstates

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
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
	Type() int
	GetContract() common.Address
	GetMethod() string
	GetArgs() []byte
	Serialize(w io.Writer) error
	Deserialize(r io.Reader) error
}

type XShardNotify struct {
	Contract common.Address `json:"contract"`
	Payer    common.Address `json:"payer"`
	Fee      uint64         `json:"fee"`
	Method   string         `json:"method"`
	Args     []byte         `json:"payload"`
}

func (msg *XShardNotify) Type() int {
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
	return shardutil.SerJson(w, msg)
}

func (msg *XShardNotify) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, msg)
}

type XShardTxReq struct {
	IdxInTx  int32          `json:"idx_in_tx"`
	Payer    common.Address `json:"payer"`
	Fee      uint64         `json:"fee"`
	Contract common.Address `json:"contract"`
	Method   string         `json:"method"`
	Args     []byte         `json:"payload"`
}

func (msg *XShardTxReq) Type() int {
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
	return shardutil.SerJson(w, msg)
}

func (msg *XShardTxReq) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, msg)
}

type XShardTxRsp struct {
	IdxInTx int32  `json:"idx_in_tx"`
	FeeUsed uint64 `json:"fee_used"`
	Result  []byte `json:"result"`
	Error   bool   `json:"error"`
}

func (msg *XShardTxRsp) Type() int {
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
	return shardutil.SerJson(w, msg)
}

func (msg *XShardTxRsp) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, msg)
}

type CommonShardMsg struct {
	SourceShardID types.ShardID  `json:"source_shard_id"`
	SourceHeight  uint64         `json:"source_height"`
	TargetShardID types.ShardID  `json:"target_shard_id"`
	SourceTxHash  common.Uint256 `json:"source_tx_hash"`
	Type          int            `json:"type"`
	Payload       []byte         `json:"payload"`
	Msg           XShardMsg      `json:"-"`
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
	msgBuf := new(bytes.Buffer)
	if err := evt.Msg.Serialize(msgBuf); err != nil {
		return err
	}
	evt.Type = evt.Msg.Type()
	evt.Payload = msgBuf.Bytes()
	return shardutil.SerJson(w, evt)
}

func (evt *CommonShardMsg) Deserialize(r io.Reader) error {
	if err := shardutil.DesJson(r, evt); err != nil {
		return err
	}

	switch evt.Type {
	case EVENT_SHARD_NOTIFY:
		notify := &XShardNotify{}
		if err := shardutil.DesJson(bytes.NewBuffer(evt.Payload), notify); err != nil {
			return err
		}
		evt.Msg = notify
	case EVENT_SHARD_TXREQ:
		req := &XShardTxReq{}
		if err := shardutil.DesJson(bytes.NewBuffer(evt.Payload), req); err != nil {
			return err
		}
		evt.Msg = req
	case EVENT_SHARD_TXRSP:
		rsp := &XShardTxRsp{}
		if err := shardutil.DesJson(bytes.NewBuffer(evt.Payload), rsp); err != nil {
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
		if err := shardutil.DesJson(bytes.NewBuffer(evt.Payload), xcommitMsg); err != nil {
			return err
		}
		if xcommitMsg.Type() != evt.Type {
			return fmt.Errorf("invalid xcommit msg %d vs %d", xcommitMsg.Type(), evt.Type)
		}
		evt.Msg = xcommitMsg
	}

	return nil
}

type _CrossShardTx struct {
	Txs [][]byte `json:"txs"`
}

func DecodeShardCommonReqs(payload []byte) ([]*CommonShardMsg, error) {
	txs := &_CrossShardTx{}
	// FIXME: fix marshaling
	if err := json.Unmarshal(payload, txs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal txs: %s", err)
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
