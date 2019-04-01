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
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/types"
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

type ShardEventState struct {
	Version    uint32
	EventType  uint32
	ToShard    types.ShardID
	FromHeight uint32
	Payload    []byte
}

func (this *ShardEventState) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, this.Version); err != nil {
		return fmt.Errorf("serialize: write version failed, err: %s", err)
	}
	if err := serialization.WriteUint32(w, this.EventType); err != nil {
		return fmt.Errorf("serialize: write event type failed, err: %s", err)
	}
	if err := serialization.WriteUint64(w, this.ToShard.ToUint64()); err != nil {
		return fmt.Errorf("serialize: write to shard failed, err: %s", err)
	}
	if err := serialization.WriteUint32(w, this.FromHeight); err != nil {
		return fmt.Errorf("serialize: write from height failed, err: %s", err)
	}
	if err := serialization.WriteVarBytes(w, this.Payload); err != nil {
		return fmt.Errorf("serialize: write payload failed, err: %s", err)
	}
	return nil
}

func (this *ShardEventState) Deserialize(r io.Reader) error {
	var err error = nil
	if this.Version, err = serialization.ReadUint32(r); err != nil {
		return fmt.Errorf("deserialize: read version failed, err: %s", err)
	}
	if this.EventType, err = serialization.ReadUint32(r); err != nil {
		return fmt.Errorf("deserialize: read event type failed, err: %s", err)
	}
	shardId, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("deserialize: read to shard failed, err: %s", err)
	}
	toShardId, err := types.NewShardID(shardId)
	if err != nil {
		return fmt.Errorf("deserialize: generate to shard id failed, err: %s", err)
	}
	this.ToShard = toShardId
	if this.FromHeight, err = serialization.ReadUint32(r); err != nil {
		return fmt.Errorf("deserialize: read from height failed, err: %s", err)
	}
	if this.Payload, err = serialization.ReadVarBytes(r); err != nil {
		return fmt.Errorf("deserialize: read payload failed, err: %s", err)
	}
	return nil
}

func (this *ShardEventState) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(this.Version)
	sink.WriteUint32(this.EventType)
	sink.WriteUint64(this.ToShard.ToUint64())
	sink.WriteUint32(this.FromHeight)
	sink.WriteVarBytes(this.Payload)
}

func (this *ShardEventState) Deserialization(source *common.ZeroCopySource) error {
	var irregular, eof bool
	this.Version, eof = source.NextUint32()
	this.EventType, eof = source.NextUint32()
	toShard, eof := source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	id, err := types.NewShardID(toShard)
	if err != nil {
		return fmt.Errorf("generate shardId faield, err: %s", err)
	}
	this.ToShard = id
	this.FromHeight, eof = source.NextUint32()
	this.Payload, _, irregular, eof = source.NextVarBytes()
	if irregular {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}
