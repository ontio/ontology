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
	"io"

	"github.com/ontio/ontology/common/serialization"
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
	GetSourceShardID() uint64
	GetTargetShardID() uint64
	GetHeight() uint64
}

type CreateShardEvent struct {
	SourceShardID uint64 `json:"source_shard_id"`
	Height        uint64 `json:"height"`
	NewShardID    uint64 `json:"new_shard_id"`
}

func (evt *CreateShardEvent) GetSourceShardID() uint64 {
	return evt.SourceShardID
}

func (evt *CreateShardEvent) GetTargetShardID() uint64 {
	return evt.NewShardID
}

func (evt *CreateShardEvent) GetHeight() uint64 {
	return evt.Height
}

func (evt *CreateShardEvent) GetType() uint32 {
	return EVENT_SHARD_CREATE
}

func (evt *CreateShardEvent) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, evt)
}

func (evt *CreateShardEvent) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, evt)
}

type ConfigShardEvent struct {
	SourceShardID uint64       `json:"source_shard_id"`
	Height        uint64       `json:"height"`
	ShardID       uint64       `json:"shard_id"`
	Config        *ShardConfig `json:"config"`
}

func (evt *ConfigShardEvent) GetSourceShardID() uint64 {
	return evt.SourceShardID
}

func (evt *ConfigShardEvent) GetTargetShardID() uint64 {
	return evt.ShardID
}

func (evt *ConfigShardEvent) GetHeight() uint64 {
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
	SourceShardID uint64 `json:"source_shard_id"`
	Height        uint64 `json:"height"`
	ShardID       uint64 `json:"shard_id"`
	PeerPubKey    string `json:"peer_pub_key"`
}

func (evt *PeerJoinShardEvent) GetSourceShardID() uint64 {
	return evt.SourceShardID
}

func (evt *PeerJoinShardEvent) GetTargetShardID() uint64 {
	return evt.ShardID
}

func (evt *PeerJoinShardEvent) GetHeight() uint64 {
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
	SourceShardID uint64 `json:"source_shard_id"`
	Height        uint64 `json:"height"`
	ShardID       uint64 `json:"shard_id"`
}

func (evt *ShardActiveEvent) GetSourceShardID() uint64 {
	return evt.SourceShardID
}

func (evt *ShardActiveEvent) GetTargetShardID() uint64 {
	return evt.ShardID
}

func (evt *ShardActiveEvent) GetHeight() uint64 {
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

type ShardEventState struct {
	Version    uint32 `json:"version"`
	EventType  uint32 `json:"event_type"`
	ToShard    uint64 `json:"to_shard"`
	FromHeight uint64 `json:"from_height"`
	Payload    []byte `json:"payload"`
}

func DecodeShardEvent(evtType uint32, evtPayload []byte) (ShardMgmtEvent, error) {
	switch evtType {
	case EVENT_SHARD_GAS_DEPOSIT:
		evt := &DepositGasEvent{}
		if err := evt.Deserialize(bytes.NewBuffer(evtPayload)); err != nil {
			return nil, fmt.Errorf("unmarshal gas deposit evt: %s", err)
		}
		return evt, nil
	case EVENT_SHARD_GAS_WITHDRAW_REQ:
		evt := &WithdrawGasReqEvent{}
		if err := evt.Deserialize(bytes.NewBuffer(evtPayload)); err != nil {
			return nil, fmt.Errorf("unmarshal gas withdraw req: %s", err)
		}
		return evt, nil
	case EVENT_SHARD_GAS_WITHDRAW_DONE:
		// TODO
		return nil, nil
	case EVENT_SHARD_REQ_COMMON:
		evt := &CommonShardReq{}
		if err := evt.Deserialize(bytes.NewBuffer(evtPayload)); err != nil {
			return nil, fmt.Errorf("unmarshal common shard req: %s", err)
		}
		return evt, nil
	}

	return nil, fmt.Errorf("unknown remote event type: %d", evtType)
}
