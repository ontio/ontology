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

type TransmitConsensusMsgReq struct {
	Target uint64
	Msg    *Consensus
}

type Consensus struct {
	Cons ConsensusPayload
	Hop  uint8
}

//Serialize message payload
func (this *Consensus) Serialization(sink *comm.ZeroCopySink) error {
	err := this.Cons.Serialization(sink)
	if err != nil {
		return err
	}
	sink.WriteUint8(this.Hop)
	return nil
}

func (this *Consensus) CmdType() string {
	return common.CONSENSUS_TYPE
}

//Deserialize message payload
func (this *Consensus) Deserialization(source *comm.ZeroCopySource) error {
	err := this.Cons.Deserialization(source)
	if err != nil {
		return err
	}
	var eof bool
	this.Hop, eof = source.NextUint8()
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}
