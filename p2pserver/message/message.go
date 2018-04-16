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
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/protocol"
)

type Messager interface {
	Verify([]byte) error
	Serialization() ([]byte, error)
	Deserialization([]byte) error
	Handle(protocol.Noder) error
}

// The network communication message header
type msgHdr struct {
	Magic    uint32
	CMD      [protocol.MSG_CMD_LEN]byte // The message type
	Length   uint32
	Checksum [protocol.CHECKSUM_LEN]byte
}

// The message body and header
type msgCont struct {
	hdr msgHdr
	p   interface{}
}

type varStr struct {
	len uint
	buf []byte
}

type filteradd struct {
	msgHdr
}

type filterclear struct {
	msgHdr
}

type filterload struct {
	msgHdr
}

// Alloc different message structure
// @t the message name or type
// @len the message length only valid for varible length structure
//
// Return:
// @messager the messager interface
// @error  error code
// FixMe fix the ugly multiple return.
func AllocMsg(t string, length int) Messager {
	switch t {
	case "msgheader":
		var msg msgHdr
		return &msg
	case "version":
		var msg version
		copy(msg.Hdr.CMD[0:len(t)], t)
		return &msg
	case "verack":
		var msg verACK
		copy(msg.msgHdr.CMD[0:len(t)], t)
		return &msg
	case "getheaders":
		var msg headersReq
		copy(msg.hdr.CMD[0:len(t)], t)
		return &msg
	case "headers":
		var msg blkHeader
		copy(msg.hdr.CMD[0:len(t)], t)
		return &msg
	case "getaddr":
		var msg addrReq
		copy(msg.Hdr.CMD[0:len(t)], t)
		return &msg
	case "addr":
		var msg addr
		copy(msg.hdr.CMD[0:len(t)], t)
		return &msg
	case "inv":
		var msg Inv
		copy(msg.Hdr.CMD[0:len(t)], t)
		// the 1 is the inv type length
		msg.P.Blk = make([]byte, length-protocol.MSG_HDR_LEN-1)
		return &msg
	case "getdata":
		var msg dataReq
		copy(msg.msgHdr.CMD[0:len(t)], t)
		return &msg
	case "block":
		var msg block
		copy(msg.msgHdr.CMD[0:len(t)], t)
		return &msg
	case "tx":
		var msg trn
		copy(msg.msgHdr.CMD[0:len(t)], t)
		return &msg
	case "consensus":
		var msg consensus
		copy(msg.msgHdr.CMD[0:len(t)], t)
		return &msg
	case "filteradd":
		var msg filteradd
		copy(msg.msgHdr.CMD[0:len(t)], t)
		return &msg
	case "filterclear":
		var msg filterclear
		copy(msg.msgHdr.CMD[0:len(t)], t)
		return &msg
	case "filterload":
		var msg filterload
		copy(msg.msgHdr.CMD[0:len(t)], t)
		return &msg
	case "getblocks":
		var msg blocksReq
		copy(msg.msgHdr.CMD[0:len(t)], t)
		return &msg
	case "txnpool":
		var msg txnPool
		copy(msg.msgHdr.CMD[0:len(t)], t)
		return &msg
	case "alert":
		log.Warn("Not supported message type - alert")
		return nil
	case "merkleblock":
		log.Warn("Not supported message type - merkleblock")
		return nil
	case "notfound":
		var msg notFound
		copy(msg.msgHdr.CMD[0:len(t)], t)
		return &msg
	case "ping":
		var msg ping
		copy(msg.msgHdr.CMD[0:len(t)], t)
		return &msg
	case "pong":
		var msg pong
		copy(msg.msgHdr.CMD[0:len(t)], t)
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
	cmd := buf[protocol.CMD_OFFSET : protocol.CMD_OFFSET+protocol.MSG_CMD_LEN]
	n := bytes.IndexByte(cmd, 0)
	if n < 0 || n >= protocol.MSG_CMD_LEN {
		return "", errors.New("Unexpected length of CMD command")
	}
	s := string(cmd[:n])
	return s, nil
}

func NewMsg(t string, n protocol.Noder) ([]byte, error) {
	switch t {
	case "version":
		return NewVersion(n)
	case "verack":
		return NewVerack()
	case "getheaders":
		return NewHeadersReq()
	case "getaddr":
		return newGetAddr()
	default:
		return nil, fmt.Errorf("Unknown message type %v", t)
	}
}

func HandleNodeMsg(node protocol.Noder, buf []byte, len int) error {
	if len < protocol.MSG_HDR_LEN {
		log.Warn("Unexpected size of received message")
		return errors.New("Unexpected size of received message")
	}

	log.Debugf("Received data len:  %d\n%x", len, buf[:len])

	s, err := MsgType(buf)
	if err != nil {
		log.Error("Message type parsing error")
		return err
	}

	log.Debug("Message type:", s)

	msg := AllocMsg(s, len)
	if msg == nil {
		log.Error(fmt.Sprintf("Allocation message %s failed", s))
		return errors.New("Allocation message failed")
	}

	msg.Deserialization(buf[:len])
	msg.Verify(buf[protocol.MSG_HDR_LEN:len])

	return msg.Handle(node)
}

func magicVerify(magic uint32) bool {
	if magic != protocol.NET_MAGIC {
		return false
	}
	return true
}

func ValidMsgHdr(buf []byte) bool {
	var h msgHdr
	h.Deserialization(buf)
	//TODO: verify hdr checksum
	return magicVerify(h.Magic)
}

func PayloadLen(buf []byte) int {
	var h msgHdr
	h.Deserialization(buf)
	return int(h.Length)
}

func LocateMsgHdr(buf []byte) []byte {
	var h msgHdr
	for i := 0; i <= len(buf)-protocol.MSG_HDR_LEN; i++ {
		if magicVerify(binary.LittleEndian.Uint32(buf[i:])) {
			buf = append(buf[:0], buf[i:]...)
			h.Deserialization(buf)
			return buf
		}
	}
	return nil
}

func checkSum(p []byte) []byte {
	t := sha256.Sum256(p)
	s := sha256.Sum256(t[:])

	// Currently we only need the front 4 bytes as checksum
	return s[:protocol.CHECKSUM_LEN]
}

func reverse(input []byte) []byte {
	if len(input) == 0 {
		return input
	}
	return append(reverse(input[1:]), input[0])
}

func (hdr *msgHdr) init(cmd string, checksum []byte, length uint32) {
	hdr.Magic = protocol.NET_MAGIC
	copy(hdr.CMD[0:uint32(len(cmd))], cmd)
	copy(hdr.Checksum[:], checksum[:protocol.CHECKSUM_LEN])
	hdr.Length = length
}

// Verify the message header information
// @p payload of the message
func (hdr msgHdr) Verify(buf []byte) error {
	if magicVerify(hdr.Magic) == false {
		log.Warn(fmt.Sprintf("Unmatched magic number 0x%0x", hdr.Magic))
		return errors.New("Unmatched magic number")
	}
	checkSum := checkSum(buf)
	if bytes.Equal(hdr.Checksum[:], checkSum[:]) == false {
		str1 := hex.EncodeToString(hdr.Checksum[:])
		str2 := hex.EncodeToString(checkSum[:])
		log.Warn(fmt.Sprintf("Message Checksum error, Received checksum %s Wanted checksum: %s",
			str1, str2))
		return errors.New("Message Checksum error")
	}

	return nil
}

func (msg *msgHdr) Deserialization(p []byte) error {

	buf := bytes.NewBuffer(p[0:protocol.MSG_HDR_LEN])
	err := binary.Read(buf, binary.LittleEndian, msg)
	return err
}

func (hdr msgHdr) Serialization() ([]byte, error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, hdr)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (hdr msgHdr) Handle(n protocol.Noder) error {
	log.Debug()
	return nil
}
