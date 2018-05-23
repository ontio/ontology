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
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
)

// The network communication message header
type MsgHdr struct {
	Magic    uint32
	CMD      [common.MSG_CMD_LEN]byte // The message type
	Length   uint32
	Checksum [common.CHECKSUM_LEN]byte
}

//initialize the header, assign netmagic and checksume value
func (this *MsgHdr) Init(cmd string, checksum []byte, length uint32) {
	this.Magic = common.NETMAGIC
	copy(this.CMD[0:uint32(len(cmd))], cmd)
	copy(this.Checksum[:], checksum[:common.CHECKSUM_LEN])
	this.Length = length
}

// Verify the message header information
// @p payload of the message
func (this MsgHdr) Verify(buf []byte) error {
	if magicVerify(this.Magic) == false {
		log.Warn(fmt.Sprintf("unmatched magic number 0x%0x", this.Magic))
		return errors.New("unmatched magic number ")
	}
	checkSum := CheckSum(buf)
	if bytes.Equal(this.Checksum[:], checkSum[:]) == false {
		str1 := hex.EncodeToString(this.Checksum[:])
		str2 := hex.EncodeToString(checkSum[:])
		log.Warn(fmt.Sprintf("message checksum error, received checksum %s Wanted checksum: %s",
			str1, str2))
		return errors.New("message checksum error ")
	}

	return nil
}

//serialize the header
func (this MsgHdr) Serialization() ([]byte, error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, this)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

//deserialize the header
func (this *MsgHdr) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p[0:common.MSG_HDR_LEN])
	err := binary.Read(buf, binary.LittleEndian, this)
	return err
}
