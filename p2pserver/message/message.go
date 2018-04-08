package message

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
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

func magicVerify(magic uint32) bool {
	if magic != common.NETMAGIC {
		return false
	}
	return true
}

func ValidMsgHdr(buf []byte) bool {
	var h MsgHdr
	h.Deserialization(buf)
	//TODO: verify hdr checksum
	return magicVerify(h.Magic)
}

func PayloadLen(buf []byte) int {
	var h MsgHdr
	h.Deserialization(buf)
	return int(h.Length)
}

func LocateMsgHdr(buf []byte) []byte {
	var h MsgHdr
	for i := 0; i <= len(buf)-common.MSG_HDR_LEN; i++ {
		if magicVerify(binary.LittleEndian.Uint32(buf[i:])) {
			buf = append(buf[:0], buf[i:]...)
			h.Deserialization(buf)
			return buf
		}
	}
	return nil
}

func CheckSum(p []byte) []byte {
	t := sha256.Sum256(p)
	s := sha256.Sum256(t[:])

	// Currently we only need the front 4 bytes as checksum
	return s[:common.CHECKSUM_LEN]
}

func Reverse(input []byte) []byte {
	if len(input) == 0 {
		return input
	}
	return append(Reverse(input[1:]), input[0])
}
