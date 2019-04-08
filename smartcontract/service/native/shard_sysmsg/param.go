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

package shardsysmsg

import (
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/events/message"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type CrossShardMsgParam struct {
	Events []*message.ShardEventState
}

func SerializeEventState(w io.Writer, state *message.ShardEventState) error {
	if err := utils.WriteVarUint(w, uint64(state.Version)); err != nil {
		return fmt.Errorf("SerializeEventState: write version failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, uint64(state.EventType)); err != nil {
		return fmt.Errorf("SerializeEventState: write event type failed, err: %s", err)
	}
	if err := utils.SerializeShardId(w, state.ToShard); err != nil {
		return fmt.Errorf("SerializeEventState: write to shard failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, uint64(state.FromHeight)); err != nil {
		return fmt.Errorf("SerializeEventState: write from height failed, err: %s", err)
	}
	if err := serialization.WriteVarBytes(w, state.Payload); err != nil {
		return fmt.Errorf("SerializeEventState: write payload failed, err: %s", err)
	}
	return nil
}

func DeserializeEventState(r io.Reader) (*message.ShardEventState, error) {
	state := &message.ShardEventState{}
	version, err := utils.ReadVarUint(r)
	if err != nil {
		return state, fmt.Errorf("DeserializeEventState: read version failed, err: %s", err)
	}
	state.Version = uint32(version)
	evtType, err := utils.ReadVarUint(r)
	if err != nil {
		return state, fmt.Errorf("DeserializeEventState: read event type failed, err: %s", err)
	}
	state.EventType = uint32(evtType)
	state.ToShard, err = utils.DeserializeShardId(r)
	if err != nil {
		return state, fmt.Errorf("DeserializeEventState: read to shard failed, err: %s", err)
	}
	fromHeight, err := utils.ReadVarUint(r)
	if err != nil {
		return state, fmt.Errorf("DeserializeEventState: read from height failed, err: %s", err)
	}
	state.FromHeight = uint32(fromHeight)
	state.Payload, err = serialization.ReadVarBytes(r)
	if err != nil {
		return state, fmt.Errorf("DeserializeEventState: read payload failed, err: %s", err)
	}
	return state, nil
}

func (this *CrossShardMsgParam) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, uint64(len(this.Events))); err != nil {
		return fmt.Errorf("serialize: write events len failed, err: %s", err)
	}
	for index, evt := range this.Events {
		if err := SerializeEventState(w, evt); err != nil {
			return fmt.Errorf("serialize: write event failed, index %d, err: %s", index, err)
		}
	}
	return nil
}

func (this *CrossShardMsgParam) Deserialize(r io.Reader) error {
	num, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read events num failed, err: %s", err)
	}
	this.Events = make([]*message.ShardEventState, num)
	for i := uint64(0); i < num; i++ {
		if evt, err := DeserializeEventState(r); err != nil {
			return fmt.Errorf("deserialize: read event failed, index %d, err: %s", i, err)
		} else {
			this.Events[i] = evt
		}
	}
	return nil
}

type NotifyReqParam struct {
	ToShard    types.ShardID
	ToContract common.Address
	Method     string
	Args       []byte
}

func (this *NotifyReqParam) Serialize(w io.Writer) error {
	if err := utils.SerializeShardId(w, this.ToShard); err != nil {
		return fmt.Errorf("serialize: write to shard failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.ToContract); err != nil {
		return fmt.Errorf("serialize: write to contract failed, err: %s", err)
	}
	if err := serialization.WriteString(w, this.Method); err != nil {
		return fmt.Errorf("serialize: write method failed, err: %s", err)
	}
	if err := serialization.WriteVarBytes(w, this.Args); err != nil {
		return fmt.Errorf("serialize: write args failed, err: %s", err)
	}
	return nil
}

func (this *NotifyReqParam) Deserialize(r io.Reader) error {
	var err error
	if this.ToShard, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read to shard failed, err: %s", err)
	}
	if this.ToContract, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read to contract failed, err: %s", err)
	}
	if this.Method, err = serialization.ReadString(r); err != nil {
		return fmt.Errorf("deserialize: read method failed, err: %s", err)
	}
	if this.Args, err = serialization.ReadVarBytes(r); err != nil {
		return fmt.Errorf("deserialize: read args failed, err: %s", err)
	}
	return nil
}

type InvokeReqParam struct {
	ToShard    types.ShardID
	ToContract common.Address
	Args       []byte
}

func (this *InvokeReqParam) Serialize(w io.Writer) error {
	if err := utils.SerializeShardId(w, this.ToShard); err != nil {
		return fmt.Errorf("serialize: write to shard failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.ToContract); err != nil {
		return fmt.Errorf("serialize: write to contract failed, err: %s", err)
	}
	if err := serialization.WriteVarBytes(w, this.Args); err != nil {
		return fmt.Errorf("serialize: write args failed, err: %s", err)
	}
	return nil
}

func (this *InvokeReqParam) Deserialize(r io.Reader) error {
	var err error
	if this.ToShard, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read to shard failed, err: %s", err)
	}
	if this.ToContract, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read to contract failed, err: %s", err)
	}
	if this.Args, err = serialization.ReadVarBytes(r); err != nil {
		return fmt.Errorf("deserialize: read args failed, err: %s", err)
	}
	return nil
}
