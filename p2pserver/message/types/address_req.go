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
)

type AddrReq struct {
	Hdr MsgHdr
}

//Check whether header is correct
func (this AddrReq) Verify(buf []byte) error {
	err := this.Hdr.Verify(buf)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNetVerifyFail, fmt.Sprintf("verify error. buf:%v", buf))
	}
	return nil
}

//Serialize message payload
func (this AddrReq) Serialization() ([]byte, error) {
	var buf bytes.Buffer
	var sum []byte
	sum = []byte{0x5d, 0xf6, 0xe0, 0xe2}
	this.Hdr.Init("getaddr", sum, 0)

	err := binary.Write(&buf, binary.LittleEndian, this)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNetPackFail, fmt.Sprintf("write error. AddrReq:%v", this))
	}
	return buf.Bytes(), nil
}

//Deserialize message payload
func (this *AddrReq) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, this)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNetVerifyFail, fmt.Sprintf("read AddrReq error. buf:%v", buf))
	}
	return nil
}
