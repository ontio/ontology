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
	pCom "github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/dht/types"
)

type Neighbors struct {
	FromID types.NodeID
	Nodes  []types.Node
}

func (this *Neighbors) CmdType() string {
	return pCom.DHT_NEIGHBORS
}

//Serialize message payload
func (this Neighbors) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteVarBytes(this.FromID[:])
	sink.WriteUint32(uint32(len(this.Nodes)))

	for _, node := range this.Nodes {
		sink.WriteVarBytes(node.ID[:])
		sink.WriteString(node.IP)
		sink.WriteUint16(node.UDPPort)
		sink.WriteUint16(node.TCPPort)
	}
	return nil
}

//Deserialize message payload
func (this *Neighbors) Deserialization(source *common.ZeroCopySource) error {
	var (
		eof       bool
		irregular bool
		buf       []byte
	)

	buf, _, irregular, eof = source.NextVarBytes()
	if eof {
		return io.ErrUnexpectedEOF
	}
	if irregular {
		return common.ErrIrregularData
	}
	copy(this.FromID[:], buf)

	num, eof := source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}

	for i := 0; i < int(num); i++ {
		node := new(types.Node)

		buf, _, irregular, eof = source.NextVarBytes()
		if eof {
			return io.ErrUnexpectedEOF
		}
		if irregular {
			return common.ErrIrregularData
		}
		copy(node.ID[:], buf)

		node.IP, _, irregular, eof = source.NextString()
		if eof {
			return io.ErrUnexpectedEOF
		}
		if irregular {
			return common.ErrIrregularData
		}

		node.UDPPort, eof = source.NextUint16()
		if eof {
			return io.ErrUnexpectedEOF
		}
		node.TCPPort, eof = source.NextUint16()
		if eof {
			return io.ErrUnexpectedEOF
		}
		this.Nodes = append(this.Nodes, *node)
	}

	if num > types.BUCKET_SIZE {
		this.Nodes = this.Nodes[:types.BUCKET_SIZE]
	}

	return nil
}
