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
	"fmt"

	"github.com/ontio/ontology/errors"
	comm "github.com/ontio/ontology/p2pserver/common"
)

type Addr struct {
	NodeAddrs []comm.PeerAddr
}

//Serialize message payload
func (this Addr) Serialization() ([]byte, error) {
	p := new(bytes.Buffer)
	num := uint64(len(this.NodeAddrs))
	err := binary.Write(p, binary.LittleEndian, num)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNetPackFail, fmt.Sprintf("write error. num:%v", num))
	}

	err = binary.Write(p, binary.LittleEndian, this.NodeAddrs)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNetPackFail, fmt.Sprintf("write error. NodeAddrs:%v", this.NodeAddrs))
	}

	return p.Bytes(), nil
}

func (this *Addr) CmdType() string {
	return comm.ADDR_TYPE
}

//Deserialize message payload
func (this *Addr) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)

	var NodeCnt uint64
	err := binary.Read(buf, binary.LittleEndian, &NodeCnt)
	if NodeCnt > comm.MAX_ADDR_NODE_CNT {
		NodeCnt = comm.MAX_ADDR_NODE_CNT
	}
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNetUnPackFail, fmt.Sprintf("read NodeCnt error. buf:%v", buf))
	}

	for i := 0; i < int(NodeCnt); i++ {
		var addr comm.PeerAddr
		err := binary.Read(buf, binary.LittleEndian, &addr)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNetUnPackFail, fmt.Sprintf("read NodeAddrs error. buf:%v", buf))
		}
		this.NodeAddrs = append(this.NodeAddrs, addr)
	}
	return nil
}
