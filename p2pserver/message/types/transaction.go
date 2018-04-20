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

	"github.com/ontio/ontology/core/types"
)

// Transaction message
type Trn struct {
	MsgHdr
	Txn types.Transaction
}

//Serialize message payload
func (msg Trn) Serialization() ([]byte, error) {
	tmpBuffer := bytes.NewBuffer([]byte{})
	msg.Txn.Serialize(tmpBuffer)
	checkSumBuf := CheckSum(tmpBuffer.Bytes())
	msg.MsgHdr.Init("tx", checkSumBuf, uint32(len(tmpBuffer.Bytes())))

	hdrBuf, err := msg.MsgHdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	err = binary.Write(buf, binary.LittleEndian, tmpBuffer.Bytes())
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

//Deserialize message payload
func (msg *Trn) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.MsgHdr))
	err = msg.Txn.Deserialize(buf)
	if err != nil {
		return err
	}
	return nil
}

type txnPool struct {
	MsgHdr
}
