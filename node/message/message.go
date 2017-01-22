package message

import (
	"fmt"
	"bytes"
	"errors"
	"time"
	"unsafe"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"GoOnchain/common"
	"GoOnchain/node"
)

const (
	MSGCMDLEN	= 12
	CMDOFFSET	= 4
	CHECKSUMLEN	= 4
	HASHLEN		= 32	// hash length in byte
	MSGHDRLEN	= 24
)

// The Inventory type
const (
	TXN		= 0x01	// Transaction
	BLOCK		= 0x02
	CONSENSUS	= 0xe0
)

type messager interface {
	verify([]byte) error
	serialization() ([]byte, error)
	deserialization([]byte) error
	handle(*node) error
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

type verACK struct {
	msgHdr
	// No payload
}

type version struct {
	Hdr msgHdr
	P  struct {
		Version		uint32
		Services	uint64
		TimeStamp	uint32
		Port		uint16
		Nonce		uint32
		// TODO remove tempory to get serilization function passed
		UserAgent	uint8
		StartHeight	uint32
		// FIXME check with the specify relay type length
		Relay		uint8
	}
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

type invPayload struct {
	invType uint8
	blk     []byte
}

type inv struct {
	hdr msgHdr
	p  invPayload
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

// Alloc different message stucture
// @t the message name or type
// @len the message length only valid for varible length structure
//
// Return:
// @messager the messager structure
// @error  error code
// FixMe fix the ugly multiple return.
func allocMsg(t string, length int) (messager, error) {
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
		return &msg, nil
	default:
		return nil, errors.New("Unknown message type")
	}
}


// TODO combine all of message alloc in one function via interface
func newMsg(t string) ([]byte, error) {
	switch t {
	case "version":
		return newVersion()
	case "verack":
		return newVerack()
	case "getheaders":
		return newHeadersReq()
	case "getaddr":
		return newGetAddr()

	default:
		return nil, errors.New("Unknown message type")
	}
}

func (hdr *msgHdr) init(cmd string, checksum []byte, length uint32) {
	hdr.Magic = NETMAGIC
	copy(hdr.CMD[0: uint32(len(cmd))], cmd)
	copy(hdr.Checksum[:], checksum[:CHECKSUMLEN])
	hdr.Length = length

	fmt.Printf("The message payload length is %d\n", hdr.Length)
	fmt.Printf("The message header length is %d\n", uint32(unsafe.Sizeof(*hdr)))
}


func (msg *version) init(n node) {
	// Do the init
}

func newVersion() ([]byte, error) {
	common.Trace()
	var msg version

	// TODO Need Node read lock or channel
	msg.P.Version = nodes.node.version
	msg.P.Services = nodes.node.services
	// FIXME Time overflow
	msg.P.TimeStamp = uint32(time.Now().UTC().UnixNano())
	msg.P.Port = nodes.node.port
	msg.P.Nonce = nodes.node.nonce
	fmt.Printf("The nonce is 0x%x", msg.P.Nonce)
	msg.P.UserAgent = 0x00
	// Fixme Get the block height from ledger
	msg.P.StartHeight = 1
	if nodes.node.relay {
		msg.P.Relay = 1
	} else {
		msg.P.Relay = 0
	}

	msg.Hdr.Magic = NETMAGIC
	ver := "version"
	copy(msg.Hdr.CMD[0:7], ver)
	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, &(msg.P))
	if err != nil {
		fmt.Println("Binary Write failed at new Msg")
		return nil, err
	}
	s := sha256.Sum256(p.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(msg.Hdr.Checksum))
	msg.Hdr.Length = uint32(len(p.Bytes()))
	fmt.Printf("The message payload length is %d\n", msg.Hdr.Length)

	m, err := msg.serialization()
	if (err != nil) {
		fmt.Println("Error Convert net message ", err.Error())
		return nil, err
	}

	str := hex.EncodeToString(m)
	fmt.Printf("The message length is %d, %s\n", len(m), str)
	return m, nil
}

func newVerack() ([]byte, error) {
	var msg verACK
	// Fixme the check is the []byte{0} instead of 0
	var sum []byte
	sum = []byte{0x5d, 0xf6, 0xe0, 0xe2}
	msg.msgHdr.init("verack", sum, 0)

	buf, err := msg.serialization()
	if (err != nil) {
		return nil, err
	}

	str := hex.EncodeToString(buf)
	fmt.Printf("The message tx verack length is %d, %s", len(buf), str)

	return buf, err
}

func newGetAddr() ([]byte, error) {
	var msg addrReq
	// Fixme the check is the []byte{0} instead of 0
	var sum []byte
	sum = []byte{0x5d, 0xf6, 0xe0, 0xe2}
	msg.Hdr.init("getaddr", sum, 0)

	buf, err := msg.serialization()
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

func payloadLen(buf []byte) int {
	var h msgHdr
	h.deserialization(buf)
	return int(h.Length)
}

func msgType(buf []byte) (string, error) {
	cmd := buf[CMDOFFSET : CMDOFFSET + MSGCMDLEN]
	n := bytes.IndexByte(cmd, 0)
	if (n < 0 || n >= MSGCMDLEN) {
		return "", errors.New("Unexpected length of CMD command")
	}
	s := string(cmd[:n])
	return s,  nil
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

func newHeadersReq() ([]byte, error) {
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

	m, err := h.serialization()
	str := hex.EncodeToString(m)
	fmt.Printf("The message length is %d, %s\n", len(m), str)
	return m, err
}

// Verify the message header information
// @p payload of the message
func (hdr msgHdr) verify(buf []byte) error {
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

func (msg version) verify(buf []byte) error {
	err := msg.Hdr.verify(buf)
	// TODO verify the message Content
	return err
}

func (msg headersReq) verify(buf []byte) error {
	// TODO verify the message Content
	err := msg.hdr.verify(buf)
	return err
}

func (msg blkHeader) verify(buf []byte) error {
	// TODO verify the message Content
	err := msg.hdr.verify(buf)
	return err
}

func (msg addrReq) verify(buf []byte) error {
	// TODO verify the message Content
	err := msg.Hdr.verify(buf)
	return err
}

func (msg inv) verify(buf []byte) error {
	// TODO verify the message Content
	err := msg.hdr.verify(buf)
	return err
}

// FIXME how to avoid duplicate serial/deserial function as
// most of them are the same
func (hdr msgHdr) serialization() ([]byte, error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, hdr)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *msgHdr) deserialization(p []byte) error {

	buf := bytes.NewBuffer(p[0 : MSGHDRLEN])
	err := binary.Read(buf, binary.LittleEndian, msg)
	return err
}

func (msg version) serialization() ([]byte, error) {
	var buf bytes.Buffer

	fmt.Printf("The size of messge is %d in serialization\n",
		uint32(unsafe.Sizeof(msg)))
	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *version) deserialization(p []byte) error {
	fmt.Printf("The size of messge is %d in deserialization\n",
		uint32(unsafe.Sizeof(*msg)))

	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, msg)
	return err
}

func (msg headersReq) serialization() ([]byte, error) {
	var buf bytes.Buffer

	fmt.Printf("The size of messge is %d in serialization\n",
		uint32(unsafe.Sizeof(msg)))
	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *headersReq) deserialization(p []byte) error {
	fmt.Printf("The size of messge is %d in deserialization\n",
		uint32(unsafe.Sizeof(*msg)))

	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, msg)
	return err
}

func (msg blkHeader) serialization() ([]byte, error) {
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

func (msg *blkHeader) deserialization(p []byte) error {
	fmt.Printf("The size of messge is %d in deserialization\n",
		uint32(unsafe.Sizeof(*msg)))

	err := msg.hdr.deserialization(p)
	msg.blkHdr = p[MSGHDRLEN : ]
	return err
}

func (msg addrReq) serialization() ([]byte, error) {
	var buf bytes.Buffer

	fmt.Printf("The size of messge is %d in serialization\n",
		uint32(unsafe.Sizeof(msg)))
	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *addrReq) deserialization(p []byte) error {
	fmt.Printf("The size of messge is %d in deserialization\n",
		uint32(unsafe.Sizeof(*msg)))

	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, msg)
	return err
}

func (msg inv) serialization() ([]byte, error) {
	var buf bytes.Buffer

	fmt.Printf("The size of messge is %d in serialization\n",
		uint32(unsafe.Sizeof(msg)))

	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *inv) deserialization(p []byte) error {
	fmt.Printf("The size of messge is %d in deserialization\n",
		uint32(unsafe.Sizeof(*msg)))

	err := msg.hdr.deserialization(p)

	msg.p.invType = p[MSGHDRLEN]
	msg.p.blk = p[MSGHDRLEN + 1 :]
	return err
}

func (msg inv) invType() byte {
	return msg.p.invType
}

//func (msg inv) invLen() (uint64, uint8) {
func (msg inv) invLen() (uint64, uint8) {
	var val uint64
	var size uint8

	len := binary.LittleEndian.Uint64(msg.p.blk[0:1])
	if (len < 0xfd) {
		val = len
		size = 1
	} else if (len == 0xfd) {
		val = binary.LittleEndian.Uint64(msg.p.blk[1 : 3])
		size = 3
	} else if (len == 0xfe) {
		val = binary.LittleEndian.Uint64(msg.p.blk[1 : 5])
		size = 5
	} else if (len == 0xff) {
		val = binary.LittleEndian.Uint64(msg.p.blk[1 : 9])
		size = 9
	}

	return val, size
}

