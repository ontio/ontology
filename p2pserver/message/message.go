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

package message

import (
	"bytes"
	"crypto/sha256"
	"errors"

	"github.com/Ontology/common/log"
	"github.com/Ontology/p2pserver/common"
)

type Message interface {
	Verify([]byte) error
	Serialization() ([]byte, error)
	Deserialization([]byte) error
}

// The message body and header
type msgCont struct {
	hdr MsgHdr
	p   interface{}
}

type varStr struct {
	len uint
	buf []byte
}

type filteradd struct {
	MsgHdr
	//TBD
}

type filterclear struct {
	MsgHdr
	//TBD
}

type filterload struct {
	MsgHdr
	//TBD
}

// Alloc different message stucture
// @t the message name or type
// @len the message length only valid for varible length structure
//
// Return:
// @messager the messager interface
// @error  error code
// FixMe fix the ugly multiple return.
func AllocMsg(t string, length int) Message {
	switch t {
	case "msgheader":
		var msg MsgHdr
		return &msg
	case "version":
		var msg Version
		// TODO fill the header and type
		copy(msg.Hdr.CMD[0:len(t)], t)
		return &msg
	case "verack":
		var msg VerACK
		copy(msg.MsgHdr.CMD[0:len(t)], t)
		return &msg
	case "getheaders":
		var msg HeadersReq
		// TODO fill the header and type
		copy(msg.Hdr.CMD[0:len(t)], t)
		return &msg
	case "headers":
		var msg BlkHeader
		copy(msg.Hdr.CMD[0:len(t)], t)
		return &msg
	case "getaddr":
		var msg AddrReq
		copy(msg.Hdr.CMD[0:len(t)], t)
		return &msg
	case "addr":
		var msg Addr
		copy(msg.Hdr.CMD[0:len(t)], t)
		return &msg
	case "inv":
		var msg Inv
		copy(msg.Hdr.CMD[0:len(t)], t)
		// the 1 is the inv type lenght
		msg.P.Blk = make([]byte, length-common.MSG_HDR_LEN-1)
		return &msg
	case "getdata":
		var msg DataReq
		copy(msg.MsgHdr.CMD[0:len(t)], t)
		return &msg
	case "block":
		var msg Block
		copy(msg.MsgHdr.CMD[0:len(t)], t)
		return &msg
	case "tx":
		var msg Trn
		copy(msg.MsgHdr.CMD[0:len(t)], t)
		//if (message.Payload.Length <= 1024 * 1024)
		//OnInventoryReceived(Transaction.DeserializeFrom(message.Payload));
		return &msg
	case "consensus":
		var msg Consensus
		copy(msg.MsgHdr.CMD[0:len(t)], t)
		return &msg
	case "filteradd":
		var msg filteradd
		copy(msg.MsgHdr.CMD[0:len(t)], t)
		return &msg
	case "filterclear":
		var msg filterclear
		copy(msg.MsgHdr.CMD[0:len(t)], t)
		return &msg
	case "filterload":
		var msg filterload
		copy(msg.MsgHdr.CMD[0:len(t)], t)
		return &msg
	case "getblocks":
		var msg BlocksReq
		copy(msg.MsgHdr.CMD[0:len(t)], t)
		return &msg
	case "txnpool":
		var msg txnPool
		copy(msg.MsgHdr.CMD[0:len(t)], t)
		return &msg
	case "alert":
		log.Warn("Not supported message type - alert")
		return nil
	case "merkleblock":
		log.Warn("Not supported message type - merkleblock")
		return nil
	case "notfound":
		var msg NotFound
		copy(msg.MsgHdr.CMD[0:len(t)], t)
		return &msg
	case "ping":
		var msg Ping
		copy(msg.Hdr.CMD[0:len(t)], t)
		return &msg
	case "pong":
		var msg Pong
		copy(msg.MsgHdr.CMD[0:len(t)], t)
		return &msg
	case "reject":
		log.Warn("Not supported message type - reject")
		return nil
	default:
		log.Warn("Unknown message type", t)
		return nil
	}
}

func MsgType(buf []byte) (string, error) {
	cmd := buf[common.CMD_OFFSET : common.CMD_OFFSET+common.MSG_CMD_LEN]
	n := bytes.IndexByte(cmd, 0)
	if n < 0 || n >= common.MSG_CMD_LEN {
		return "", errors.New("Unexpected length of CMD command")
	}
	s := string(cmd[:n])
	return s, nil
}

//check netmagic value
func magicVerify(magic uint32) bool {
	if magic != common.NETMAGIC {
		return false
	}
	return true
}

//check wether header is valid
func ValidMsgHdr(buf []byte) bool {
	var h MsgHdr
	h.Deserialization(buf)
	//TODO: verify hdr checksum
	return magicVerify(h.Magic)
}

//caculate payload length
func PayloadLen(buf []byte) int {
	var h MsgHdr
	h.Deserialization(buf)
	return int(h.Length)
}

//caculate checksum value
func CheckSum(p []byte) []byte {
	t := sha256.Sum256(p)
	s := sha256.Sum256(t[:])

	// Currently we only need the front 4 bytes as checksum
	return s[:common.CHECKSUM_LEN]
}

// reverse the input
func Reverse(input []byte) []byte {
	if len(input) == 0 {
		return input
	}
	return append(Reverse(input[1:]), input[0])
}
