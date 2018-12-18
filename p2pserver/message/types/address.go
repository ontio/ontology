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

	"github.com/ontio/ontology/common"
	comm "github.com/ontio/ontology/p2pserver/common"
)

type Addr struct {
	NodeAddrs []comm.PeerAddr
}

//Serialize message payload
func (this Addr) Serialization(sink *common.ZeroCopySink) error {
	num := uint64(len(this.NodeAddrs))
	sink.WriteUint64(num)

	for _, addr := range this.NodeAddrs {
		sink.WriteInt64(addr.Time)
		sink.WriteUint64(addr.Services)
		sink.WriteBytes(addr.IpAddr[:])
		sink.WriteUint16(addr.Port)
		sink.WriteUint16(addr.ConsensusPort)
		sink.WriteUint64(addr.ID)
	}

	return nil
}

func (this *Addr) CmdType() string {
	return comm.ADDR_TYPE
}

func (this *Addr) Deserialization(source *common.ZeroCopySource) error {
	count, eof := source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}

	for i := 0; i < int(count); i++ {
		var addr comm.PeerAddr
		addr.Time, eof = source.NextInt64()
		addr.Services, eof = source.NextUint64()
		buf, _ := source.NextBytes(uint64(len(addr.IpAddr[:])))
		copy(addr.IpAddr[:], buf)
		addr.Port, eof = source.NextUint16()
		addr.ConsensusPort, eof = source.NextUint16()
		addr.ID, eof = source.NextUint64()
		if eof {
			return io.ErrUnexpectedEOF
		}

		this.NodeAddrs = append(this.NodeAddrs, addr)
	}

	if count > comm.MAX_ADDR_NODE_CNT {
		count = comm.MAX_ADDR_NODE_CNT
	}
	this.NodeAddrs = this.NodeAddrs[:count]

	return nil
}
