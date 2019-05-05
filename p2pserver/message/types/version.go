/*
 * Copyright (C) 2018 The ontology Authors
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

package types

import (
	"io"

	comm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/p2pserver/common"
)

type VersionPayload struct {
	Version      uint32
	Services     uint64
	TimeStamp    int64
	SyncPort     uint16
	HttpInfoPort uint16
	ConsPort     uint16
	Cap          [32]byte
	Nonce        uint64
	StartHeight  uint64
	Relay        uint8
	IsConsensus  bool
	SoftVersion  string
	ShardHeights map[uint64]uint32
}

type Version struct {
	P VersionPayload
}

//Serialize message payload
func (this *Version) Serialization(sink *comm.ZeroCopySink) {
	sink.WriteUint32(this.P.Version)
	sink.WriteUint64(this.P.Services)
	sink.WriteInt64(this.P.TimeStamp)
	sink.WriteUint16(this.P.SyncPort)
	sink.WriteUint16(this.P.HttpInfoPort)
	sink.WriteUint16(this.P.ConsPort)
	sink.WriteBytes(this.P.Cap[:])
	sink.WriteUint64(this.P.Nonce)
	sink.WriteUint64(this.P.StartHeight)
	sink.WriteUint8(this.P.Relay)
	sink.WriteBool(this.P.IsConsensus)
	sink.WriteString(this.P.SoftVersion)
	sink.WriteUint32(uint32(len(this.P.ShardHeights)))
	for id, h := range this.P.ShardHeights {
		sink.WriteUint64(id)
		sink.WriteUint32(h)
	}
}

func (this *Version) CmdType() string {
	return common.VERSION_TYPE
}

//Deserialize message payload
func (this *Version) Deserialization(source *comm.ZeroCopySource) error {
	var irregular, eof bool
	this.P.Version, eof = source.NextUint32()
	this.P.Services, eof = source.NextUint64()
	this.P.TimeStamp, eof = source.NextInt64()
	this.P.SyncPort, eof = source.NextUint16()
	this.P.HttpInfoPort, eof = source.NextUint16()
	this.P.ConsPort, eof = source.NextUint16()
	var buf []byte
	buf, eof = source.NextBytes(uint64(len(this.P.Cap[:])))
	copy(this.P.Cap[:], buf)

	this.P.Nonce, eof = source.NextUint64()
	this.P.StartHeight, eof = source.NextUint64()
	this.P.Relay, eof = source.NextUint8()
	this.P.IsConsensus, irregular, eof = source.NextBool()
	if eof {
		return io.ErrUnexpectedEOF
	}
	if irregular {
		return comm.ErrIrregularData
	}

	this.P.SoftVersion, _, irregular, eof = source.NextString()
	if eof || irregular {
		this.P.SoftVersion = ""
	}

	this.P.ShardHeights = make(map[uint64]uint32)
	shardCnt, eof := source.NextUint32()
	if !eof {
		for i := uint32(0); i < shardCnt; i++ {
			shardId, eof := source.NextUint64()
			if eof {
				break
			}
			shardHeight, eof := source.NextUint32()
			if eof {
				break
			}
			if _, err := comm.NewShardID(shardId); err != nil {
				break
			}
			this.P.ShardHeights[shardId] = shardHeight
		}
	}

	return nil
}
