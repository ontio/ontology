package net

import (
	"log"
	"bytes"
	"errors"
	"time"
	"unsafe"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"GoOnchain/common"
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
	TX	= 0x01
	BLOCK	= 0x02
	CONSENSUS = 0xe0
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

type inv struct {
	hdr msgHdr
	p struct {
		invType uint8
		blk     []byte
	}
}

type dataReq struct {
	msgHdr
	// TBD
}

type block struct {
	msgHdr
	// TBD
}

type transaction struct {
	msgHdr
	// TBD
}

// TODO Sample function, shoule be called from ledger module
func ledgerGetHeader() ([]byte, error) {
	genesisHeader := "b3181718ef6167105b70920e4a8fbbd0a0a56aacf460d70e10ba6fa1668f1fef"

	h, err := hex.DecodeString(genesisHeader)
	if err != nil {
		log.Printf("Decode Header hash error")
		return nil, err
	}
	return h, nil
}

// Alloc different message stucture
// @t the message name or type
// @len the message length only valid for varible length structure
//
// Return:
// @messager the messager structure
// @error  error code
// FixMe fix the ugly multiple return.
func allocMsg(t string, length uint64) (messager, error) {
	switch t {
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
		var msg transaction
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

	log.Printf("The message payload length is %d", hdr.Length)
	log.Printf("The message header length is %d", uint32(unsafe.Sizeof(*hdr)))
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
	log.Printf("The nonce is 0x%x", msg.P.Nonce)
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
		log.Println("Binary Write failed at new Msg")
		return nil, err
	}
	s := sha256.Sum256(p.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(msg.Hdr.Checksum))
	msg.Hdr.Length = uint32(len(p.Bytes()))
	log.Printf("The message payload length is %d", msg.Hdr.Length)

	m, err := msg.serialization()
	if (err != nil) {
		log.Println("Error Convert net message ", err.Error())
		return nil, err
	}

	str := hex.EncodeToString(m)
	log.Printf("The message length is %d, %s", len(m), str)
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
	log.Printf("The message tx verack length is %d, %s", len(buf), str)

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
	log.Printf("The message get addr length is %d, %s", len(buf), str)

	return buf, err
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
	return s[:CHECKSUMLEN]
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
	buf, err := ledgerGetHeader()
	if (err != nil) {
		return nil, err
	}
	copy(h.p.hashStart[:], reverse(buf))

	p := new(bytes.Buffer)
	err = binary.Write(p, binary.LittleEndian, &(h.p))
	if err != nil {
		log.Println("Binary Write failed at new headersReq")
		return nil, err
	}

	s := checkSum(p.Bytes())
	h.hdr.init("getheaders", s, uint32(len(p.Bytes())))

	m, err := h.serialization()
	str := hex.EncodeToString(m)
	log.Printf("The message length is %d, %s", len(m), str)
	return m, err
}

// Verify the message header information
// @p payload of the message
func (hdr msgHdr) verify(buf []byte) error {
	// TODO verify the message header
	// checksum,version magic number
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

	log.Printf("The size of messge is %d in serialization",
		uint32(unsafe.Sizeof(hdr)))
	err := binary.Write(&buf, binary.LittleEndian, hdr)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *msgHdr) deserialization(p []byte) error {
	log.Printf("The size of messge is %d in deserialization",
		uint32(unsafe.Sizeof(*msg)))

	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, msg)
	return err
}

func (msg version) serialization() ([]byte, error) {
	var buf bytes.Buffer

	log.Printf("The size of messge is %d in serialization",
		uint32(unsafe.Sizeof(msg)))
	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *version) deserialization(p []byte) error {
	log.Printf("The size of messge is %d in deserialization",
		uint32(unsafe.Sizeof(*msg)))

	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, msg)
	return err
}

func (msg headersReq) serialization() ([]byte, error) {
	var buf bytes.Buffer

	log.Printf("The size of messge is %d in serialization",
		uint32(unsafe.Sizeof(msg)))
	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *headersReq) deserialization(p []byte) error {
	log.Printf("The size of messge is %d in deserialization",
		uint32(unsafe.Sizeof(*msg)))

	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, msg)
	return err
}

func (msg blkHeader) serialization() ([]byte, error) {
	var buf bytes.Buffer

	log.Printf("The size of messge is %d in serialization",
		uint32(unsafe.Sizeof(msg)))
	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *blkHeader) deserialization(p []byte) error {
	log.Printf("The size of messge is %d in deserialization",
		uint32(unsafe.Sizeof(*msg)))

	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, msg)
	return err
}

func (msg addrReq) serialization() ([]byte, error) {
	var buf bytes.Buffer

	log.Printf("The size of messge is %d in serialization",
		uint32(unsafe.Sizeof(msg)))
	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *addrReq) deserialization(p []byte) error {
	log.Printf("The size of messge is %d in deserialization",
		uint32(unsafe.Sizeof(*msg)))

	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, msg)
	return err
}

func (msg inv) serialization() ([]byte, error) {
	var buf bytes.Buffer

	log.Printf("The size of messge is %d in serialization",
		uint32(unsafe.Sizeof(msg)))
	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *inv) deserialization(p []byte) error {
	log.Printf("The size of messge is %d in deserialization",
		uint32(unsafe.Sizeof(*msg)))

	buf := bytes.NewBuffer(p[0 : MSGHDRLEN])
	err := binary.Read(buf, binary.LittleEndian, msg.hdr)

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
