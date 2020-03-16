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
	ncomm "github.com/ontio/ontology/p2pserver/common"
)

type FindNodeReq struct {
	TargetID ncomm.PeerId
}

// Serialization message payload
func (req FindNodeReq) Serialization(sink *common.ZeroCopySink) {
	req.TargetID.Serialization(sink)
}

// CmdType return this message type
func (req *FindNodeReq) CmdType() string {
	return ncomm.FINDNODE_TYPE
}

// Deserialization message payload
func (req *FindNodeReq) Deserialization(source *common.ZeroCopySource) error {
	return req.TargetID.Deserialization(source)
}

type FindNodeResp struct {
	TargetID    ncomm.PeerId
	Success     bool
	Address     string
	CloserPeers []ncomm.PeerIDAddressPair
}

// Serialization message payload
func (resp FindNodeResp) Serialization(sink *common.ZeroCopySink) {
	resp.TargetID.Serialization(sink)
	sink.WriteBool(resp.Success)
	sink.WriteString(resp.Address)
	sink.WriteUint32(uint32(len(resp.CloserPeers)))
	for _, curPeer := range resp.CloserPeers {
		curPeer.ID.Serialization(sink)
		sink.WriteString(curPeer.Address)
	}
}

// CmdType return this message type
func (resp *FindNodeResp) CmdType() string {
	return ncomm.FINDNODE_RESP_TYPE
}

// Deserialization message payload
func (resp *FindNodeResp) Deserialization(source *common.ZeroCopySource) error {
	err := resp.TargetID.Deserialization(source)
	if err != nil {
		return err
	}

	succ, _, eof := source.NextBool()
	if eof {
		return io.ErrUnexpectedEOF
	}
	resp.Success = succ

	addr, _, _, eof := source.NextString()
	if eof {
		return io.ErrUnexpectedEOF
	}
	resp.Address = addr

	numCloser, eof := source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}

	for i := 0; i < int(numCloser); i++ {
		var curpa ncomm.PeerIDAddressPair
		id := ncomm.PeerId{}
		err = id.Deserialization(source)
		if err != nil {
			return err
		}
		curpa.ID = id
		addr, _, _, eof := source.NextString()
		if eof {
			return io.ErrUnexpectedEOF
		}
		curpa.Address = addr

		resp.CloserPeers = append(resp.CloserPeers, curpa)
	}

	return nil
}
