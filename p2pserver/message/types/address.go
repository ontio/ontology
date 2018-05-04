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
	NodeAddrs []comm.PeerAddr
}

//Check whether header is correct
func (this Addr) Verify(buf []byte) error {
	err := this.Hdr.Verify(buf)
	return err
}

//Serialize message payload
func (this Addr) Serialization() ([]byte, error) {
	p := new(bytes.Buffer)
	num := uint64(len(this.NodeAddrs))
	err := binary.Write(p, binary.LittleEndian, num)
	if err != nil {
		return nil, err
	}

	err = binary.Write(p, binary.LittleEndian, this.NodeAddrs)
	if err != nil {
		return nil, err
	}

	checkSumBuf := CheckSum(p.Bytes())
	this.Hdr.Init("addr", checkSumBuf, uint32(len(p.Bytes())))

	var buf bytes.Buffer
	err = binary.Write(&buf, binary.LittleEndian, this.Hdr)

	if err != nil {
		return nil, err
	}
	data := append(buf.Bytes(), p.Bytes()...)
	return data, nil
}

//Deserialize message payload
func (this *Addr) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(this.Hdr))
	var NodeCnt uint64
	err = binary.Read(buf, binary.LittleEndian, &NodeCnt)
	this.NodeAddrs = make([]comm.PeerAddr, NodeCnt)
	for i := 0; i < int(NodeCnt); i++ {
		err := binary.Read(buf, binary.LittleEndian, &(this.NodeAddrs[i]))
		if err != nil {
			goto err
		}
	}
err:
	return err
}
