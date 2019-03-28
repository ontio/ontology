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

package message

import (
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"io"
)

const (
	ShardGetGenesisBlockReq = iota
	ShardGetGenesisBlockRsp
	ShardGetPeerInfoReq
	ShardGetPeerInfoRsp
)

type ShardSystemEventMsg struct {
	FromAddress common.Address
	Event       *ShardEventState
}

func (this *ShardSystemEventMsg) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.FromAddress); err != nil {
		return fmt.Errorf("serialize: write from addr failed, err: %s", err)
	}
	if err := this.Event.Serialize(w); err != nil {
		return fmt.Errorf("serialize: write evt failed, err: %s", err)
	}
	return nil
}

func (this *ShardSystemEventMsg) Deserialize(r io.Reader) error {
	var err error = nil
	if this.FromAddress, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read from addr failed, err: %s", err)
	}
	if err = this.Event.Deserialize(r); err != nil {
		return fmt.Errorf("deserialize: read event failed, err: %s", err)
	}
	return nil
}

type ShardEventState struct {
	Version    uint32
	EventType  uint32
	ToShard    types.ShardID
	FromHeight uint32
	Payload    []byte
}

func (this *ShardEventState) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, uint64(this.Version)); err != nil {
		return fmt.Errorf("serialize: write version failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.EventType)); err != nil {
		return fmt.Errorf("serialize: write event type failed, err: %s", err)
	}
	if err := utils.SerializeShardId(w, this.ToShard); err != nil {
		return fmt.Errorf("serialize: write to shard failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.FromHeight)); err != nil {
		return fmt.Errorf("serialize: write from height failed, err: %s", err)
	}
	if err := serialization.WriteVarBytes(w, this.Payload); err != nil {
		return fmt.Errorf("serialize: write payload failed, err: %s", err)
	}
	return nil
}

func (this *ShardEventState) Deserialize(r io.Reader) error {
	version, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read version failed, err: %s", err)
	}
	this.Version = uint32(version)
	evtType, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read event type failed, err: %s", err)
	}
	this.EventType = uint32(evtType)
	if this.ToShard, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read to shard failed, err: %s", err)
	}
	height, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read from height failed, err: %s", err)
	}
	this.FromHeight = uint32(height)
	payload, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("deserialize: read payload failed, err: %s", err)
	}
	this.Payload = payload
	return nil
}
