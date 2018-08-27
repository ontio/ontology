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

type Pong struct {
	Height uint64
}

//Serialize message payload
func (this Pong) Serialization(sink *comm.ZeroCopySink) error {
	sink.WriteUint64(this.Height)
	return nil
}

func (this Pong) CmdType() string {
	return common.PONG_TYPE
}

//Deserialize message payload
func (this *Pong) Deserialization(source *comm.ZeroCopySource) error {
	var eof bool
	this.Height, eof = source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}

	return nil
}
