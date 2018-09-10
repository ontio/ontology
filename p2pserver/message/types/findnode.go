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

type FindNode struct {
	FromID   types.NodeID
	TargetID types.NodeID
}

func (this *FindNode) CmdType() string {
	return pCom.DHT_FIND_NODE
}

//Serialize message payload
func (this FindNode) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteVarBytes(this.FromID[:])
	sink.WriteVarBytes(this.TargetID[:])
	return nil
}

//Deserialize message payload
func (this *FindNode) Deserialization(source *common.ZeroCopySource) error {
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

	buf, _, irregular, eof = source.NextVarBytes()
	if eof {
		return io.ErrUnexpectedEOF
	}
	if irregular {
		return common.ErrIrregularData
	}
	copy(this.TargetID[:], buf)

	return nil
}
