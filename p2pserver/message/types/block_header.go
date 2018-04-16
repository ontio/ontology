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

	"github.com/ontio/ontology/common/log"
	ct "github.com/ontio/ontology/core/types"
)

type BlkHeader struct {
	Hdr    MsgHdr
	Cnt    uint32
	BlkHdr []ct.Header
}

//Check whether header is correct
func (msg BlkHeader) Verify(buf []byte) error {
	err := msg.Hdr.Verify(buf)
	return err
}

//Serialize message payload
func (msg BlkHeader) Serialization() ([]byte, error) {
	hdrBuf, err := msg.Hdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	err = binary.Write(buf, binary.LittleEndian, msg.Cnt)
	if err != nil {
		return nil, err
	}

	for _, header := range msg.BlkHdr {
		header.Serialize(buf)
	}
	return buf.Bytes(), err
}

//Deserialize message payload
func (msg *BlkHeader) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.Hdr))
	if err != nil {
		return err
	}

	err = binary.Read(buf, binary.LittleEndian, &(msg.Cnt))
	if err != nil {
		return err
	}

	for i := 0; i < int(msg.Cnt); i++ {
		var headers ct.Header
		err := (&headers).Deserialize(buf)
		msg.BlkHdr = append(msg.BlkHdr, headers)
		if err != nil {
			log.Debug("blkHeader Deserialization failed")
			goto blkHdrErr
		}
	}

blkHdrErr:
	return err
}
