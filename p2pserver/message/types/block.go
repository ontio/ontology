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

	"github.com/ontio/ontology/common/log"
	ct "github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
)

type Block struct {
	MsgHdr
	Blk ct.Block
}

//Check whether header is correct
func (this Block) Verify(buf []byte) error {
	err := this.MsgHdr.Verify(buf)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNetVerifyFail, fmt.Sprintf("verify error. buf:%v", buf))
	}
	return nil
}

//Serialize message payload
func (this Block) Serialization() ([]byte, error) {

	p := bytes.NewBuffer([]byte{})
	err := this.Blk.Serialize(p)

	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNetPackFail, fmt.Sprintf("serialize error. Blk:%v", this.Blk))
	}

	checkSumBuf := CheckSum(p.Bytes())
	this.Init("block", checkSumBuf, uint32(len(p.Bytes())))
	log.Debug("The message payload length is ", this.Length)

	hdrBuf, err := this.MsgHdr.Serialization()
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNetPackFail, fmt.Sprint("serialization error. MsgHdr:%v", this.MsgHdr))
	}
	buf := bytes.NewBuffer(hdrBuf)
	data := append(buf.Bytes(), p.Bytes()...)
	return data, nil
}

//Deserialize message payload
func (this *Block) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(this.MsgHdr))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNetUnPackFail, fmt.Sprint("read MsgHdr error. buf:%v", buf))
	}

	err = this.Blk.Deserialize(buf)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNetUnPackFail, fmt.Sprint("read Blk error. buf:%v", buf))
	}

	return nil
}
