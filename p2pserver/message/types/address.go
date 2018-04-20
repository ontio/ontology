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
	"bytes"
	"encoding/binary"

	comm "github.com/ontio/ontology/p2pserver/common"
)

type Addr struct {
	Hdr       MsgHdr
	NodeCnt   uint64
	NodeAddrs []comm.PeerAddr
}

//Check whether header is correct
func (msg Addr) Verify(buf []byte) error {
	err := msg.Hdr.Verify(buf)
	return err
}

//Serialize message payload
func (msg Addr) Serialization() ([]byte, error) {
	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, msg.NodeCnt)
	if err != nil {
		return nil, err
	}

	err = binary.Write(p, binary.LittleEndian, msg.NodeAddrs)
	if err != nil {
		return nil, err
	}

	checkSumBuf := CheckSum(p.Bytes())
	msg.Hdr.Init("addr", checkSumBuf, uint32(len(p.Bytes())))

	var buf bytes.Buffer
	err = binary.Write(&buf, binary.LittleEndian, msg.Hdr)

	if err != nil {
		return nil, err
	}
	err = binary.Write(&buf, binary.LittleEndian, p.Bytes())
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}

//Deserialize message payload
func (msg *Addr) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.Hdr))
	err = binary.Read(buf, binary.LittleEndian, &(msg.NodeCnt))
	msg.NodeAddrs = make([]comm.PeerAddr, msg.NodeCnt)
	for i := 0; i < int(msg.NodeCnt); i++ {
		err := binary.Read(buf, binary.LittleEndian, &(msg.NodeAddrs[i]))
		if err != nil {
			goto err
		}
	}
err:
	return err
}
