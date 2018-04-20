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
	"errors"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
)

type DataReq struct {
	MsgHdr
	DataType common.InventoryType
	Hash     common.Uint256
}

//Serialize message payload
func (msg DataReq) Serialization() ([]byte, error) {
	p := bytes.NewBuffer([]byte{})
	err := binary.Write(p, binary.LittleEndian, &(msg.DataType))
	msg.Hash.Serialize(p)
	if err != nil {
		return nil, err
	}

	checkSumBuf := CheckSum(p.Bytes())
	msg.Init("getdata", checkSumBuf, uint32(len(p.Bytes())))

	hdrBuf, err := msg.MsgHdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	err = binary.Write(buf, binary.LittleEndian, p.Bytes())
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}

//Deserialize message payload
func (msg *DataReq) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.MsgHdr))
	if err != nil {
		log.Warn("Parse dataReq message hdr error")
		return errors.New("Parse dataReq message hdr error ")
	}

	err = binary.Read(buf, binary.LittleEndian, &(msg.DataType))
	if err != nil {
		log.Warn("Parse dataReq message dataType error")
		return errors.New("Parse dataReq message dataType error ")
	}

	err = msg.Hash.Deserialize(buf)
	if err != nil {
		log.Warn("Parse dataReq message hash error")
		return errors.New("Parse dataReq message hash error ")
	}
	return nil
}
