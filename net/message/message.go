package message

import (
	"fmt"
	"bytes"
	"errors"
	"unsafe"
	"encoding/binary"
	"crypto/sha256"
	"encoding/hex"
	"GoOnchain/common"
	. "GoOnchain/net/protocol"
)

// The Inventory type
const (
	TXN		= 0x01	// Transaction
	BLOCK		= 0x02
	CONSENSUS	= 0xe0
)

type Messager interface {
	Verify([]byte) error
	Serialization() ([]byte, error)
	Deserialization([]byte) error
	Handle(*Noder) error
}

// The network communication message header
type msgHdr struct {
	Magic	 uint32
	CMD	 [MSGCMDLEN]byte 	// The message type
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

type headersReq struct {
	hdr msgHdr
	p struct {
		len		uint8
		hashStart	[HASHLEN]byte
		hashEnd		[HASHLEN]byte
	}
}

type addrReq struct {
	Hdr msgHdr
	// No payload
}

type blkHeader struct {
	hdr msgHdr
	blkHdr []byte
}

type addr struct {
	msgHdr
	// TBD
}

type dataReq struct {
	msgHdr
	// TBD
}

type block struct {
	msgHdr
	// TBD
}

// Transaction message
type trn struct {
	msgHdr
	// TBD
}

type consensus struct {
	msgHdr
	//TBD
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

type blockReq struct {
	msgHdr
	//TBD
}

type memPool struct {
	msgHdr
	//TBD
}

func (msg *msgHdr) Deserialization(p []byte) error {

	buf := bytes.NewBuffer(p[0 : MSGHDRLEN])
	err := binary.Read(buf, binary.LittleEndian, msg)
	return err
}

// Alloc different message stucture
// @t the message name or type
// @len the message length only valid for varible length structure
//
// Return:
// @messager the messager interface
// @error  error code
// FixMe fix the ugly multiple return.
func AllocMsg(t string, length int) (Messager, error) {
	switch t {
	case "msgheader":
		var msg msgHdr
		return &msg, nil
	case "version":
		var msg version
		return &msg, nil
	case "verack":
		var msg verACK
		return &msg, nil
	case "getheaders":
		var msg headersReq
		return &msg, nil
	case "headers":
		var msg blkHeader
		return &msg, nil
	case "getaddr":
		var msg addrReq
		return &msg, nil
	case "addr":
		var msg addr
		return &msg, nil
	case "inv":
		var msg inv
		// the 1 is the inv type lenght
		msg.p.blk = make([]byte, length - MSGHDRLEN - 1)
		return &msg, nil
	case "getdata":
		var msg dataReq
		return &msg, nil
	case "block":
		var msg block
		return &msg, nil
	case "tx":
		var msg trn
		//if (message.Payload.Length <= 1024 * 1024)
		//OnInventoryReceived(Transaction.DeserializeFrom(message.Payload));
		return &msg, nil
	case "consensus":
		var msg consensus
		return &msg, nil
	case "filteradd":
		var msg filteradd
		return &msg, nil
	case "filterclear":
		var msg filterclear
		return &msg, nil
	case "filterload":
		var msg filterload
		return &msg, nil
	case "getblocks":
		var msg blockReq
		return &msg, nil
	case "mempool":
		var msg memPool
		return &msg, nil
	case "alert":
		return nil, errors.New("Not supported message type")
	case "merkleblock":
		return nil, errors.New("Not supported message type")
	case "notfound":
		return nil, errors.New("Not supported message type")
	case "ping":
		return nil, errors.New("Not supported message type")
	case "pong":
		return nil, errors.New("Not supported message type")
	case "reject":
		return nil, errors.New("Not supported message type")
	default:
		return nil, errors.New("Unknown message type")
	}
}

func MsgType(buf []byte) (string, error) {
	cmd := buf[CMDOFFSET : CMDOFFSET + MSGCMDLEN]
	n := bytes.IndexByte(cmd, 0)
	if (n < 0 || n >= MSGCMDLEN) {
		return "", errors.New("Unexpected length of CMD command")
	}
	s := string(cmd[:n])
	return s,  nil
}

// TODO combine all of message alloc in one function via interface
func NewMsg(t string, n Noder) ([]byte, error) {
	switch t {
	case "version":
		return NewVersion(n)
	case "verack":
		return newVerack()
	case "getheaders":
		return NewHeadersReq()
	case "getaddr":
		return newGetAddr()

	default:
		return nil, errors.New("Unknown message type")
	}
}

// FIXME the length exceed int32 case?
func HandleNodeMsg(node Noder, buf []byte, len int) error {
	if (len < MSGHDRLEN) {
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

	msg, err := AllocMsg(s, len)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	msg.Deserialization(buf[0 : len])
	msg.Verify(buf[MSGHDRLEN : len])
	return msg.Handle(&node)
}

func (hdr *msgHdr) init(cmd string, checksum []byte, length uint32) {
	hdr.Magic = NETMAGIC
	copy(hdr.CMD[0: uint32(len(cmd))], cmd)
	copy(hdr.Checksum[:], checksum[:CHECKSUMLEN])
	hdr.Length = length

	fmt.Printf("The message payload length is %d\n", hdr.Length)
	fmt.Printf("The message header length is %d\n", uint32(unsafe.Sizeof(*hdr)))
}

func newGetAddr() ([]byte, error) {
	var msg addrReq
	// Fixme the check is the []byte{0} instead of 0
	var sum []byte
	sum = []byte{0x5d, 0xf6, 0xe0, 0xe2}
	msg.Hdr.init("getaddr", sum, 0)

	buf, err := msg.Serialization()
	if (err != nil) {
		return nil, err
	}

	str := hex.EncodeToString(buf)
	fmt.Printf("The message get addr length is %d, %s", len(buf), str)

	return buf, err
}

func magicVerify(magic uint32) bool {
	if (magic != NETMAGIC) {
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
	return s[: CHECKSUMLEN]
}

func reverse(input []byte) []byte {
    if len(input) == 0 {
        return input
    }
    return append(reverse(input[1:]), input[0])
}

func NewHeadersReq() ([]byte, error) {
	var h headersReq

	// Fixme correct with the exactly request length
	h.p.len = 1
	buf, err := LedgerGetHeader()
	if (err != nil) {
		return nil, err
	}
	copy(h.p.hashStart[:], reverse(buf))

	p := new(bytes.Buffer)
	err = binary.Write(p, binary.LittleEndian, &(h.p))
	if err != nil {
		fmt.Println("Binary Write failed at new headersReq")
		return nil, err
	}

	s := checkSum(p.Bytes())
	h.hdr.init("getheaders", s, uint32(len(p.Bytes())))

	m, err := h.Serialization()
	str := hex.EncodeToString(m)
	fmt.Printf("The message length is %d, %s\n", len(m), str)
	return m, err
}

// Verify the message header information
// @p payload of the message
func (hdr msgHdr) Verify(buf []byte) error {
	if (hdr.Magic != NETMAGIC) {
		fmt.Printf("Unmatched magic number 0x%d\n", hdr.Magic)
		return errors.New("Unmatched magic number")
	}

	checkSum := checkSum(buf)
	if (bytes.Equal(hdr.Checksum[:], checkSum[:]) == false) {
		str1 := hex.EncodeToString(hdr.Checksum[:])
		str2 := hex.EncodeToString(checkSum[:])
		fmt.Printf("Message Checksum error, Received checksum %s Wanted checksum: %s\n",
			str1, str2)
		return errors.New("Message Checksum error")
	}

	return nil
}

func (msg headersReq) Verify(buf []byte) error {
	// TODO Verify the message Content
	err := msg.hdr.Verify(buf)
	return err
}

func (msg blkHeader) Verify(buf []byte) error {
	// TODO Verify the message Content
	err := msg.hdr.Verify(buf)
	return err
}

func (msg addrReq) Verify(buf []byte) error {
	// TODO Verify the message Content
	err := msg.Hdr.Verify(buf)
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

func (msg headersReq) Serialization() ([]byte, error) {
	var buf bytes.Buffer

	fmt.Printf("The size of messge is %d in serialization\n",
		uint32(unsafe.Sizeof(msg)))
	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *headersReq) Deserialization(p []byte) error {
	fmt.Printf("The size of messge is %d in deserialization\n",
		uint32(unsafe.Sizeof(*msg)))

	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, msg)
	return err
}

func (msg blkHeader) Serialization() ([]byte, error) {
	var buf bytes.Buffer

	fmt.Printf("The size of messge is %d in serialization\n",
		uint32(unsafe.Sizeof(msg)))
	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	// TODO serilization the header, then the payload
	return buf.Bytes(), err
}

func (msg *blkHeader) Deserialization(p []byte) error {
	fmt.Printf("The size of messge is %d in deserialization\n",
		uint32(unsafe.Sizeof(*msg)))

	err := msg.hdr.Deserialization(p)
	msg.blkHdr = p[MSGHDRLEN : ]
	return err
}

func (msg addrReq) Serialization() ([]byte, error) {
	var buf bytes.Buffer

	fmt.Printf("The size of messge is %d in serialization\n",
		uint32(unsafe.Sizeof(msg)))
	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *addrReq) Deserialization(p []byte) error {
	fmt.Printf("The size of messge is %d in deserialization\n",
		uint32(unsafe.Sizeof(*msg)))

	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, msg)
	return err
}

func (hdr msgHdr) Handle(n *Noder) error {
	common.Trace()
	// TBD
	return nil
}

func (msg headersReq) Handle(node *Noder) error {
	common.Trace()
	// TBD
	return nil
}

func (msg blkHeader) Handle(node *Noder) error {
	common.Trace()
	// TBD
	return nil
}

func (msg addrReq) Handle(node *Noder) error {
	common.Trace()
	// TBD
	return nil
}
