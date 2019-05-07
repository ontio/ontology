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
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
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
	GetSourceShardID() common.ShardID
	GetTargetShardID() common.ShardID
	GetHeight() uint32
	Serialization(sink *common.ZeroCopySink)
	Deserialization(source *common.ZeroCopySource) error
}

type ImplSourceTargetShardID struct {
	SourceShardID common.ShardID
	ShardID       common.ShardID
}

func (self *ImplSourceTargetShardID) GetSourceShardID() common.ShardID {
	return self.SourceShardID
}

func (self *ImplSourceTargetShardID) GetTargetShardID() common.ShardID {
	return self.ShardID
}

func (self ImplSourceTargetShardID) Serialize(w io.Writer) error {
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

func (this *ImplSourceTargetShardID) Serialization(sink *common.ZeroCopySink) {
	utils.SerializationShardId(sink, this.SourceShardID)
	utils.SerializationShardId(sink, this.ShardID)
}

func (this *ImplSourceTargetShardID) Deserialization(source *common.ZeroCopySource) error {
	this.SourceShardID = common.ShardID{}
	this.ShardID = common.ShardID{}
	var err error = nil
	if this.SourceShardID, err = utils.DeserializationShardId(source); err != nil {
		return fmt.Errorf("read source shard id err: %s", err)
	}
	if this.ShardID, err = utils.DeserializationShardId(source); err != nil {
		return fmt.Errorf("read shard id err: %s", err)
	}
	return nil
}

type CreateShardEvent struct {
	SourceShardID common.ShardID
	Height        uint32
	NewShardID    common.ShardID
}

func (evt *CreateShardEvent) GetSourceShardID() common.ShardID {
	return evt.SourceShardID
}

func (evt *CreateShardEvent) GetTargetShardID() common.ShardID {
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

func (this *CreateShardEvent) Serialization(sink *common.ZeroCopySink) {
	utils.SerializationShardId(sink, this.SourceShardID)
	sink.WriteUint32(this.Height)
	utils.SerializationShardId(sink, this.NewShardID)
}

func (this *CreateShardEvent) Deserialization(source *common.ZeroCopySource) error {
	this.SourceShardID = common.ShardID{}
	var err error = nil
	if this.SourceShardID, err = utils.DeserializationShardId(source); err != nil {
		return fmt.Errorf("read source shard id err: %s", err)
	}
	var eof bool
	this.Height, eof = source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.NewShardID = common.ShardID{}
	if this.NewShardID, err = utils.DeserializationShardId(source); err != nil {
		return fmt.Errorf("read shard id err: %s", err)
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
	if err := evt.ImplSourceTargetShardID.Serialize(w); err != nil {
		return fmt.Errorf("serialize: write impl shardId failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, uint64(evt.Height)); err != nil {
		return fmt.Errorf("serialize: write height failed, err: %s", err)
	}
	if err := evt.Config.Serialize(w); err != nil {
		return fmt.Errorf("serialize: write config failed, err: %s", err)
	}
	return nil
}

func (evt *ConfigShardEvent) Deserialize(r io.Reader) error {
	var err error = nil
	evt.ImplSourceTargetShardID = ImplSourceTargetShardID{}
	if err = evt.ImplSourceTargetShardID.Deserialize(r); err != nil {
		return fmt.Errorf("deserialize: read impl shardId failed, err: %s", err)
	}
	height, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read height failed, err: %s", err)
	}
	evt.Height = uint32(height)
	evt.Config = &ShardConfig{}
	if err = evt.Config.Deserialize(r); err != nil {
		return fmt.Errorf("deserialize: read config failed, err: %s", err)
	}
	return nil
}

func (this *ConfigShardEvent) Serialization(sink *common.ZeroCopySink) {
	this.ImplSourceTargetShardID.Serialization(sink)
	sink.WriteUint32(this.Height)
	this.Config.Serialization(sink)
}

func (this *ConfigShardEvent) Deserialization(source *common.ZeroCopySource) error {
	this.ImplSourceTargetShardID = ImplSourceTargetShardID{}
	if err := this.ImplSourceTargetShardID.Deserialization(source); err != nil {
		return fmt.Errorf("read impl err: %s", err)
	}
	var eof bool
	this.Height, eof = source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.Config = &ShardConfig{}
	if err := this.Config.Deserialization(source); err != nil {
		return fmt.Errorf("read config err: %s", err)
	}
	return nil
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
	if err := evt.ImplSourceTargetShardID.Serialize(w); err != nil {
		return fmt.Errorf("serialize: write impl shardId failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, uint64(evt.Height)); err != nil {
		return fmt.Errorf("serialize: write height failed, err: %s", err)
	}
	if err := serialization.WriteString(w, evt.PeerPubKey); err != nil {
		return fmt.Errorf("serialize: write peer pub key failed, err: %s", err)
	}
	return nil
}

func (evt *PeerJoinShardEvent) Deserialize(r io.Reader) error {
	var err error = nil
	evt.ImplSourceTargetShardID = ImplSourceTargetShardID{}
	if err = evt.ImplSourceTargetShardID.Deserialize(r); err != nil {
		return fmt.Errorf("deserialize: read impl shardId failed, err: %s", err)
	}
	height, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read height failed, err: %s", err)
	}
	evt.Height = uint32(height)
	if evt.PeerPubKey, err = serialization.ReadString(r); err != nil {
		return fmt.Errorf("deserialize: read peer pub key failed, err: %s", err)
	}
	return nil
}

func (this *PeerJoinShardEvent) Serialization(sink *common.ZeroCopySink) {
	this.ImplSourceTargetShardID.Serialization(sink)
	sink.WriteUint32(this.Height)
	sink.WriteString(this.PeerPubKey)
}

func (this *PeerJoinShardEvent) Deserialization(source *common.ZeroCopySource) error {
	this.ImplSourceTargetShardID = ImplSourceTargetShardID{}
	if err := this.ImplSourceTargetShardID.Deserialization(source); err != nil {
		return fmt.Errorf("read impl err: %s", err)
	}
	var irregular, eof bool
	this.Height, eof = source.NextUint32()
	this.PeerPubKey, _, irregular, eof = source.NextString()
	if irregular {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
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
	if err := evt.ImplSourceTargetShardID.Serialize(w); err != nil {
		return fmt.Errorf("serialize: write impl shardId failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, uint64(evt.Height)); err != nil {
		return fmt.Errorf("serialize: write height failed, err: %s", err)
	}
	return nil
}

func (evt *ShardActiveEvent) Deserialize(r io.Reader) error {
	var err error = nil
	evt.ImplSourceTargetShardID = ImplSourceTargetShardID{}
	if err = evt.ImplSourceTargetShardID.Deserialize(r); err != nil {
		return fmt.Errorf("deserialize: read impl shardId failed, err: %s", err)
	}
	height, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read height failed, err: %s", err)
	}
	evt.Height = uint32(height)
	return nil
}

func (this *ShardActiveEvent) Serialization(sink *common.ZeroCopySink) {
	this.ImplSourceTargetShardID.Serialization(sink)
	sink.WriteUint32(this.Height)
}

func (this *ShardActiveEvent) Deserialization(source *common.ZeroCopySource) error {
	this.ImplSourceTargetShardID = ImplSourceTargetShardID{}
	if err := this.ImplSourceTargetShardID.Deserialization(source); err != nil {
		return fmt.Errorf("read impl err: %s", err)
	}
	var eof bool
	this.Height, eof = source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func DecodeShardGasEvent(evtType uint32, evtPayload []byte) (ShardMgmtEvent, error) {
	switch evtType {
	case EVENT_SHARD_GAS_DEPOSIT:
		evt := &DepositGasEvent{}
		if err := evt.Deserialization(common.NewZeroCopySource(evtPayload)); err != nil {
			return nil, fmt.Errorf("DecodeShardGasEvent: %s", err)
		}
		return evt, nil
	case EVENT_SHARD_GAS_WITHDRAW_DONE:
		evt := &WithdrawGasDoneEvent{}
		if err := evt.Deserialization(common.NewZeroCopySource(evtPayload)); err != nil {
			return nil, fmt.Errorf("DecodeShardGasEvent: %s", err)
		}
		return evt, nil
	}

	return nil, fmt.Errorf("DecodeShardGasEvent: unknown remote event type: %d", evtType)
}
