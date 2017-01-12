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

type messager interface {
	verify() error
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
	msgHdr
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
	msgHdr
	p struct {
		version		uint32
		services	uint64
		timeStamp	uint32
		port		uint16
		nonce		uint32
		// TODO remove tempory to get serilization function passed
		userAgent	uint8
		startHeight	uint32
		// FIXME check with the specify relay type length
		relay		uint8
	}
}

type headersReq struct {
	msgHdr
	p struct {
		len		uint8
		hashStart	[HASHLEN]byte
		hashEnd		[HASHLEN]byte
	}
}

type addrReq struct {
	msgHdr
	// No payload
}

// TODO Sample function, shoule be called from ledger module
func ledgerGetHeader() ([]byte, error) {
	genesisHeader := "38c41224b9d791d382d896a426d94bb68ef0a0208eca1156cb9e6d52287a12dd"

	h, err := hex.DecodeString(genesisHeader)
	if err != nil {
		log.Printf("Decode Header hash error")
		return nil, err
	}
	return h, nil
}

func allocMsg(t string) (messager, error) {
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
	case "getaddr":
		var msg addrReq
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
	var v version

	// TODO Need Node read lock or channel
	v.p.version = nodes.node.version
	v.p.services = nodes.node.services
	// FIXME Time overflow
	v.p.timeStamp = uint32(time.Now().UTC().UnixNano())
	v.p.port = nodes.node.port
	v.p.nonce = nodes.node.nonce
	log.Printf("The nonce is 0x%x", v.p.nonce)
	v.p.userAgent = 0x00
	// Fixme Get the block height from ledger
	v.p.startHeight = 1
	if nodes.node.relay {
		v.p.relay = 1
	} else {
		v.p.relay = 0
	}

	v.Magic = NETMAGIC
	ver := "version"
	copy(v.CMD[0:7], ver)
	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, &(v.p))
	if err != nil {
		log.Println("Binary Write failed at new Msg")
		return nil, err
	}
	s := sha256.Sum256(p.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(v.Checksum))
	v.Length = uint32(len(p.Bytes()))
	log.Printf("The message payload length is %d", v.Length)

	m, err := v.serialization()
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
	msg.msgHdr.init("getaddr", sum, 0)

	buf, err := msg.serialization()
	if (err != nil) {
		return nil, err
	}

	str := hex.EncodeToString(buf)
	log.Printf("The message get addr length is %d, %s", len(buf), str)

	return buf, err
}

func msgType(buf []byte) string {
	cmd := buf[CMDOFFSET : CMDOFFSET + MSGCMDLEN]
	n := bytes.IndexByte(cmd, 0)
	s := string(cmd[:n])

	return s
}

func checkSum(p []byte) []byte {
	t := sha256.Sum256(p)
	s := sha256.Sum256(t[:])

	// Currently we only need the front 4 bytes as checksum
	return s[:CHECKSUMLEN]
}

func newHeadersReq() ([]byte, error) {
	var h headersReq

	// Fixme correct with the exactly request length
	h.p.len = 1
	buf, err := ledgerGetHeader()
	if (err != nil) {
		return nil, err
	}
	copy(h.p.hashStart[:], buf[:32])

	p := new(bytes.Buffer)
	err = binary.Write(p, binary.LittleEndian, &(h.p))
	if err != nil {
		log.Println("Binary Write failed at new headersReq")
		return nil, err
	}

	s := checkSum(p.Bytes())
	h.msgHdr.init("getheaders", s, uint32(len(p.Bytes())))

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

func (v version) verify() error {
	// TODO verify the message Content
	return nil
}

func (v verACK) verify() error {
	// TODO verify the message Content
	return nil
}

func (v headersReq) verify() error {
	// TODO verify the message Content
	return nil
}

func (v addrReq) verify() error {
	// TODO verify the message Content
	return nil
}

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

func (hdr *msgHdr) deserialization(p []byte) error {
//func (hdr *msgHdr) deserialization() error {
	log.Printf("The size of messge is %d in deserialization",
		uint32(unsafe.Sizeof(*hdr)))

	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, hdr)
	return err
}

func (v version) serialization() ([]byte, error) {
	var buf bytes.Buffer

	log.Printf("The size of messge is %d in serialization",
		uint32(unsafe.Sizeof(v)))
	err := binary.Write(&buf, binary.LittleEndian, v)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}


func (h headersReq) serialization() ([]byte, error) {
	var buf bytes.Buffer

	log.Printf("The size of messge is %d in serialization",
		uint32(unsafe.Sizeof(h)))
	err := binary.Write(&buf, binary.LittleEndian, h)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}
