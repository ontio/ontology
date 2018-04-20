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

	"github.com/ontio/ontology/common/serialization"
)

type Ping struct {
	Hdr    MsgHdr
	Height uint64
}

//Check whether header is correct
func (msg Ping) Verify(buf []byte) error {
	err := msg.Hdr.Verify(buf)
	return err
}

//Serialize message payload
func (msg Ping) Serialization() ([]byte, error) {
	tmpBuffer := bytes.NewBuffer([]byte{})
	serialization.WriteUint64(tmpBuffer, msg.Height)

	checkSumBuf := CheckSum(tmpBuffer.Bytes())
	msg.Hdr.Init("ping", checkSumBuf, uint32(len(tmpBuffer.Bytes())))

	hdrBuf, err := msg.Hdr.Serialization()
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
func (msg *Ping) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.Hdr))
	if err != nil {
		return err
	}

	msg.Height, err = serialization.ReadUint64(buf)
	return err
}
