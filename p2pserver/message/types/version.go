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

	comm "github.com/ontio/ontology/v2/common"
	"github.com/ontio/ontology/v2/p2pserver/common"
)

type VersionPayload struct {
	Version      uint32
	Services     uint64
	TimeStamp    int64
	SyncPort     uint16
	HttpInfoPort uint16
	//TODO remove this legecy field
	ConsPort    uint16
	Cap         [32]byte
	Nonce       uint64
	StartHeight uint64
	Relay       uint8
	IsConsensus bool
	SoftVersion string
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
}

func (this *Version) CmdType() string {
	return common.VERSION_TYPE
}

//Deserialize message payload
func (this *Version) Deserialization(source *comm.ZeroCopySource) error {
	var irregular, eof bool
	this.P.Version, eof = source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}

	this.P.Services, eof = source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}

	this.P.TimeStamp, eof = source.NextInt64()
	if eof {
		return io.ErrUnexpectedEOF
	}

	this.P.SyncPort, eof = source.NextUint16()
	if eof {
		return io.ErrUnexpectedEOF
	}

	this.P.HttpInfoPort, eof = source.NextUint16()
	if eof {
		return io.ErrUnexpectedEOF
	}

	this.P.ConsPort, eof = source.NextUint16()
	if eof {
		return io.ErrUnexpectedEOF
	}

	var buf []byte
	buf, eof = source.NextBytes(uint64(len(this.P.Cap[:])))
	if eof {
		return io.ErrUnexpectedEOF
	}
	copy(this.P.Cap[:], buf)

	this.P.Nonce, eof = source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}

	this.P.StartHeight, eof = source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}

	this.P.Relay, eof = source.NextUint8()
	if eof {
		return io.ErrUnexpectedEOF
	}

	this.P.IsConsensus, irregular, eof = source.NextBool()
	if eof || irregular {
		return io.ErrUnexpectedEOF
	}

	this.P.SoftVersion, _, irregular, eof = source.NextString()
	if eof || irregular {
		this.P.SoftVersion = ""
	}

	return nil
}
