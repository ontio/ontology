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
func (hdr *MsgHdr) Init(cmd string, checksum []byte, length uint32) {
	hdr.Magic = common.NETMAGIC
	copy(hdr.CMD[0:uint32(len(cmd))], cmd)
	copy(hdr.Checksum[:], checksum[:common.CHECKSUM_LEN])
	hdr.Length = length
}

// Verify the message header information
// @p payload of the message
func (hdr MsgHdr) Verify(buf []byte) error {
	if magicVerify(hdr.Magic) == false {
		log.Warn(fmt.Sprintf("Unmatched magic number 0x%0x", hdr.Magic))
		return errors.New("Unmatched magic number ")
	}
	checkSum := CheckSum(buf)
	if bytes.Equal(hdr.Checksum[:], checkSum[:]) == false {
		str1 := hex.EncodeToString(hdr.Checksum[:])
		str2 := hex.EncodeToString(checkSum[:])
		log.Warn(fmt.Sprintf("Message Checksum error, Received checksum %s Wanted checksum: %s",
			str1, str2))
		return errors.New("Message Checksum error ")
	}

	return nil
}

//serialize the header
func (hdr MsgHdr) Serialization() ([]byte, error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, hdr)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

//deserialize the header
func (hdr *MsgHdr) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p[0:common.MSG_HDR_LEN])
	err := binary.Read(buf, binary.LittleEndian, hdr)
	return err
}
