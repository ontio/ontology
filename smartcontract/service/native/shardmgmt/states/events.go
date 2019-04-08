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
	"fmt"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
)

const (
	EVENT_SHARD_CREATE = iota
	EVENT_SHARD_CONFIG_UPDATE
	EVENT_SHARD_PEER_JOIN
	EVENT_SHARD_ACTIVATED
	EVENT_SHARD_PEER_LEAVE
)

type ShardMgmtEvent interface {
	serialization.SerializableData
	GetType() uint32
	GetSourceShardID() types.ShardID
	GetTargetShardID() types.ShardID
	GetHeight() uint32
}

type ImplSourceTargetShardID struct {
	SourceShardID types.ShardID
	ShardID       types.ShardID
}

func (self *ImplSourceTargetShardID) GetSourceShardID() types.ShardID {
	return self.SourceShardID
}

func (self *ImplSourceTargetShardID) GetTargetShardID() types.ShardID {
	return self.ShardID
}

func (self *ImplSourceTargetShardID) Serialize(w io.Writer) error {
	if err := utils.SerializeShardId(w, self.SourceShardID); err != nil {
		return fmt.Errorf("serialize: write source shardId failed, err: %s", err)
	}
	if err := utils.SerializeShardId(w, self.ShardID); err != nil {
		return fmt.Errorf("serialize: write shardId failed, err: %s", err)
	}
	return nil
}

func (self *ImplSourceTargetShardID) Deserialize(r io.Reader) error {
	var err error = nil
	if self.SourceShardID, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read source shardId failed, err: %s", err)
	}
	if self.ShardID, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read shardId failed, err: %s", err)
	}
	return nil
}

type CreateShardEvent struct {
	SourceShardID types.ShardID
	Height        uint32
	NewShardID    types.ShardID
}

func (evt *CreateShardEvent) GetSourceShardID() types.ShardID {
	return evt.SourceShardID
}

func (evt *CreateShardEvent) GetTargetShardID() types.ShardID {
	return evt.NewShardID
}

func (evt *CreateShardEvent) GetHeight() uint32 {
	return evt.Height
}

func (evt *CreateShardEvent) GetType() uint32 {
	return EVENT_SHARD_CREATE
}

func (evt *CreateShardEvent) Serialize(w io.Writer) error {
	if err := utils.SerializeShardId(w, evt.SourceShardID); err != nil {
		return fmt.Errorf("serialize: write source shardId failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, uint64(evt.Height)); err != nil {
		return fmt.Errorf("serialize: write height failed, err: %s", err)
	}
	if err := utils.SerializeShardId(w, evt.NewShardID); err != nil {
		return fmt.Errorf("serialize: write new shardId failed, err: %s", err)
	}
	return nil
}

func (evt *CreateShardEvent) Deserialize(r io.Reader) error {
	var err error = nil
	if evt.SourceShardID, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read source shardId failed, err: %s", err)
	}
	height, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read height failed, err: %s", err)
	}
	evt.Height = uint32(height)
	if evt.NewShardID, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read new shardId failed, err: %s", err)
	}
	return nil
}

type ConfigShardEvent struct {
	ImplSourceTargetShardID
	Height uint32       `json:"height"`
	Config *ShardConfig `json:"config"`
}

func (evt *ConfigShardEvent) GetHeight() uint32 {
	return evt.Height
}

func (evt *ConfigShardEvent) GetType() uint32 {
	return EVENT_SHARD_CONFIG_UPDATE
}

func (evt *ConfigShardEvent) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, evt)
}

func (evt *ConfigShardEvent) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, evt)
}

type PeerJoinShardEvent struct {
	ImplSourceTargetShardID
	Height     uint32 `json:"height"`
	PeerPubKey string `json:"peer_pub_key"`
}

func (evt *PeerJoinShardEvent) GetHeight() uint32 {
	return evt.Height
}

func (evt *PeerJoinShardEvent) GetType() uint32 {
	return EVENT_SHARD_PEER_JOIN
}

func (evt *PeerJoinShardEvent) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, evt)
}

func (evt *PeerJoinShardEvent) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, evt)
}

type ShardActiveEvent struct {
	ImplSourceTargetShardID
	Height uint32 `json:"height"`
}

func (evt *ShardActiveEvent) GetHeight() uint32 {
	return evt.Height
}

func (evt *ShardActiveEvent) GetType() uint32 {
	return EVENT_SHARD_ACTIVATED
}

func (evt *ShardActiveEvent) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, evt)
}

func (evt *ShardActiveEvent) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, evt)
}

type XShardCommitMsg struct {
	MsgType int `json:"msg_type"`
}

func (msg *XShardCommitMsg) Type() int {
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
	return shardutil.SerJson(w, msg)
}

func (msg *XShardCommitMsg) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, msg)
}

func DecodeShardGasEvent(evtType uint32, evtPayload []byte) (ShardMgmtEvent, error) {
	switch evtType {
	case EVENT_SHARD_GAS_DEPOSIT:
		evt := &DepositGasEvent{}
		if err := evt.Deserialize(bytes.NewBuffer(evtPayload)); err != nil {
			return nil, fmt.Errorf("unmarshal gas deposit evt: %s", err)
		}
		return evt, nil
	case EVENT_SHARD_GAS_WITHDRAW_DONE:
		evt := &WithdrawGasDoneEvent{}
		if err := evt.Deserialize(bytes.NewBuffer(evtPayload)); err != nil {
			return nil, fmt.Errorf("unmarshal gas withdraw done evt: %s", err)
		}
		return evt, nil
	}

	return nil, fmt.Errorf("unknown remote event type: %d", evtType)
}
