package message

import (
	"GoOnchain/common"
	. "GoOnchain/net/protocol"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
)

type Messager interface {
	Verify([]byte) error
	Serialization() ([]byte, error)
	Deserialization([]byte) error
	Handle(Noder) error
}

// The network communication message header
type msgHdr struct {
	Magic    uint32
	//ID	 uint64
	CMD      [MSGCMDLEN]byte // The message type
	Length   uint32
	Checksum [CHECKSUMLEN]byte
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
	//TBD
}

type filterclear struct {
	msgHdr
	//TBD
}

type filterload struct {
	msgHdr
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
func AllocMsg(t string, length int) Messager {
	switch t {
	case "msgheader":
		var msg msgHdr
		return &msg
	case "version":
		var msg version
		// TODO fill the header and type
		copy(msg.Hdr.CMD[0:len(t)], t)
		return &msg
	case "verack":
		var msg verACK
		copy(msg.msgHdr.CMD[0:len(t)], t)
		return &msg
	case "getheaders":
		var msg headersReq
		// TODO fill the header and type
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
		// the 1 is the inv type lenght
		msg.P.Blk = make([]byte, length-MSGHDRLEN-1)
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
		//if (message.Payload.Length <= 1024 * 1024)
		//OnInventoryReceived(Transaction.DeserializeFrom(message.Payload));
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
		var msg blockReq
		copy(msg.msgHdr.CMD[0:len(t)], t)
		return &msg
	case "mempool":
		var msg memPool
		copy(msg.msgHdr.CMD[0:len(t)], t)
		return &msg
	case "alert":
		errors.New("Not supported message type")
		return nil
	case "merkleblock":
		errors.New("Not supported message type")
		return nil
	case "notfound":
		errors.New("Not supported message type")
		return nil
	case "ping":
		errors.New("Not supported message type")
		return nil
	case "pong":
		errors.New("Not supported message type")
		return nil
	case "reject":
		errors.New("Not supported message type")
		return nil
	default:
		errors.New("Unknown message type")
		return nil
	}
}

func MsgType(buf []byte) (string, error) {
	cmd := buf[CMDOFFSET : CMDOFFSET+MSGCMDLEN]
	n := bytes.IndexByte(cmd, 0)
	if n < 0 || n >= MSGCMDLEN {
		return "", errors.New("Unexpected length of CMD command")
	}
	s := string(cmd[:n])
	return s, nil
}

// TODO combine all of message alloc in one function via interface
func NewMsg(t string, n Noder) ([]byte, error) {
	switch t {
	case "version":
		return NewVersion(n)
	case "verack":
		return NewVerack()
	case "getheaders":
		return NewHeadersReq(n)
	case "getaddr":
		return newGetAddr()

	default:
		return nil, errors.New("Unknown message type")
	}
}

// FIXME the length exceed int32 case?
func HandleNodeMsg(node Noder, buf []byte, len int) error {
	if len < MSGHDRLEN {
		fmt.Println("Unexpected size of received message")
		return errors.New("Unexpected size of received message")
	}

	//str := hex.EncodeToString(buf[:len])
	//fmt.Printf("Received data len %d\n: %s \n  Received string: %v \n",
	//	len, str, string(buf[:len]))
	//fmt.Printf("Received data len %d : \"%v\" ", len, string(buf[:len]))

	s, err := MsgType(buf)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	msg := AllocMsg(s, len)
	if msg == nil {
		fmt.Println(err.Error())
		return err
	}
	// Todo attach a ndoe pointer to each message
	msg.Deserialization(buf[0:len])
	msg.Verify(buf[MSGHDRLEN:len])
	return msg.Handle(node)
}

func magicVerify(magic uint32) bool {
	if magic != NETMAGIC {
		return false
	}
	return true
}

func PayloadLen(buf []byte) int {
	var h msgHdr
	h.Deserialization(buf)
	return int(h.Length)
}

func checkSum(p []byte) []byte {
	t := sha256.Sum256(p)
	s := sha256.Sum256(t[:])

	// Currently we only need the front 4 bytes as checksum
	return s[:CHECKSUMLEN]
}

func reverse(input []byte) []byte {
	if len(input) == 0 {
		return input
	}
	return append(reverse(input[1:]), input[0])
}

func (hdr *msgHdr) init(cmd string, checksum []byte, length uint32) {
	hdr.Magic = NETMAGIC
	copy(hdr.CMD[0:uint32(len(cmd))], cmd)
	copy(hdr.Checksum[:], checksum[:CHECKSUMLEN])
	hdr.Length = length
	//hdr.ID = id
}

// Verify the message header information
// @p payload of the message
func (hdr msgHdr) Verify(buf []byte) error {
	if hdr.Magic != NETMAGIC {
		fmt.Printf("Unmatched magic number 0x%d\n", hdr.Magic)
		return errors.New("Unmatched magic number")
	}

	checkSum := checkSum(buf)
	if bytes.Equal(hdr.Checksum[:], checkSum[:]) == false {
		str1 := hex.EncodeToString(hdr.Checksum[:])
		str2 := hex.EncodeToString(checkSum[:])
		fmt.Printf("Message Checksum error, Received checksum %s Wanted checksum: %s\n",
			str1, str2)
		return errors.New("Message Checksum error")
	}

	return nil
}

func (msg *msgHdr) Deserialization(p []byte) error {

	buf := bytes.NewBuffer(p[0:MSGHDRLEN])
	err := binary.Read(buf, binary.LittleEndian, msg)
	return err
}

// FIXME how to avoid duplicate serial/deserial function as
// most of them are the same
func (hdr msgHdr) Serialization() ([]byte, error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, hdr)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (hdr msgHdr) Handle(n Noder) error {
	common.Trace()
	// TBD
	return nil
}
