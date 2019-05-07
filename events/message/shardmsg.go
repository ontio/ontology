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
	ToShard    common.ShardID
	FromHeight uint32
	Payload    []byte
}

func (this *ShardEventState) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(this.Version)
	sink.WriteUint32(this.EventType)
	sink.WriteUint64(this.ToShard.ToUint64())
	sink.WriteUint32(this.FromHeight)
	sink.WriteVarBytes(this.Payload)
}

func (this *ShardEventState) Deserialization(source *common.ZeroCopySource) error {
	var irregualr, eof bool
	this.Version, eof = source.NextUint32()
	this.EventType, eof = source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	toShard, eof := source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	id, err := common.NewShardID(toShard)
	if err != nil {
		return fmt.Errorf("serialization: generate shard id failed, err: %s", err)
	}
	this.ToShard = id
	this.FromHeight, eof = source.NextUint32()
	this.Payload, _, irregualr, eof = source.NextVarBytes()
	if irregualr {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}
