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
	"github.com/ontio/ontology/events/message"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/types"
)

type CrossShardMsgParam struct {
	Events []*message.ShardEventState
}

func (this *CrossShardMsgParam) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, uint64(len(this.Events))); err != nil {
		return fmt.Errorf("serialize: write events len failed, err: %s", err)
	}
	for index, evt := range this.Events {
		if err := evt.Serialize(w); err != nil {
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
		evt := &message.ShardEventState{}
		if err := evt.Deserialize(r); err != nil {
			return fmt.Errorf("deserialize: read event failed, index %d, err: %s", i, err)
		}
		this.Events[i] = evt
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
