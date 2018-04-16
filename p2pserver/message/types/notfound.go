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

type NotFound struct {
	MsgHdr
	Hash common.Uint256
}

//Check whether header is correct
func (msg NotFound) Verify(buf []byte) error {
	err := msg.MsgHdr.Verify(buf)
	return err
}

//Serialize message payload
func (msg NotFound) Serialization() ([]byte, error) {
	hdrBuf, err := msg.MsgHdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	msg.Hash.Serialize(buf)

	return buf.Bytes(), err
}

//Deserialize message payload
func (msg *NotFound) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)

	err := binary.Read(buf, binary.LittleEndian, &(msg.MsgHdr))
	if err != nil {
		log.Warn("Parse notFound message hdr error")
		return errors.New("Parse notFound message hdr error ")
	}

	err = msg.Hash.Deserialize(buf)
	if err != nil {
		log.Warn("Parse notFound message error")
		return errors.New("Parse notFound message error ")
	}

	return err
}
