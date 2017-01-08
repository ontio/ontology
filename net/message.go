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
	MSGCMDLEN = 12
	CHECKSUMLEN = 4
)

// The network and module communication message buffer
type Msg struct {
	Magic	 uint32
	CMD	 [MSGCMDLEN]byte 	// The message type
	Length   uint32
	Checksum [CHECKSUMLEN]byte
	//payloader interface{}
}

type varStr struct {
	len uint
	buf []byte
}

type verACK struct {
	// No payload
}

type version struct {
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

type headersReq struct {
	len		uint8
	hashStart	[32]byte
	hashEnd		[32]byte
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

func newMsgHeader(cmd string, checksum []byte, length  uint32) ([]byte, error) {
	msg := new(Msg)
	msg.Magic = NETMAGIC
	copy(msg.CMD[0:len(cmd)], cmd)
	copy(msg.Checksum[:], checksum[:CHECKSUMLEN])
	//binary.Read(buf, binary.LittleEndian, &(msg.Checksum))
	msg.Length = length

	log.Printf("The message payload length is %d", msg.Length)
	log.Printf("The message header length is %d", uint32(unsafe.Sizeof(*msg)))

	m, err := msg.serialization()
	if (err != nil) {
		log.Println("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}

func newVersion() ([]byte, error) {
	common.Trace()
	var v version

	// TODO Need Node read lock or channel
	v.version = nodes.node.version
	v.services = nodes.node.services
	// FIXME Time overflow
	v.timeStamp = uint32(time.Now().UTC().UnixNano())
	v.port = nodes.node.port
	v.nonce = nodes.node.nonce
	log.Printf("The nonce is 0x%x", v.nonce)
	v.userAgent = 0x00
	// Fixme Get the block height from ledger
	v.startHeight = 1
	if nodes.node.relay {
		v.relay = 1
	} else {
		v.relay = 0
	}

	msg := new(Msg)
	msg.Magic = NETMAGIC
	ver := "version"
	copy(msg.CMD[0:7], ver)
	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, &v)
	if err != nil {
		log.Println("Binary Write failed at new Msg")
		return nil, err
	}
	s := sha256.Sum256(p.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(msg.Checksum))
	msg.Length = uint32(len(p.Bytes()))
	log.Printf("The message payload length is %d", msg.Length)
	log.Printf("The message header len is %d", uint32(unsafe.Sizeof(*msg)))

	m, err := msg.serialization()
	if (err != nil) {
		log.Println("Error Convert net message ", err.Error())
		return nil, err
	}

	m = append(m, p.Bytes()...)

	str := hex.EncodeToString(m)
	log.Printf("The message length is %d, %s", len(m), str)
	return m, nil
}

func newVerack() ([]byte, error) {
	//var verACK verACK
	var sum []byte
	sum = []byte{0x5d, 0xf6, 0xe0, 0xe2}
	msg := new(Msg)

	msg.Magic = NETMAGIC
	v := "verack"
	copy(msg.CMD[0:6], v)
	msg.Length = 0
	log.Printf("The checksum should be 0 or not in this case")
	copy(msg.Checksum[0:4], sum)
	//msg.payloader = &verACK

	buf, err := msg.serialization()
	if (err != nil) {
		return nil, err
	}

	str := hex.EncodeToString(buf)
	log.Printf("The message tx verack length is %d, %s", len(buf), str)

	return buf, err
}

func newGetAddr() ([]byte, error) {
	//var verACK verACK
	msg := new(Msg)

	msg.Magic = NETMAGIC
	v := "getaddr"
	copy(msg.CMD[0:7], v)
	msg.Length = 0
	//msg.Checksum = 0
	//msg.payloader = &verACK

	buf, err := msg.serialization()
	if (err != nil) {
		return nil, err
	}

	return buf, err
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
	h.len = 1
	buf, err := ledgerGetHeader()
	if (err != nil) {
		return nil, err
	}
	copy(h.hashStart[:], buf[:32])

	p := new(bytes.Buffer)
	err = binary.Write(p, binary.LittleEndian, &h)
	if err != nil {
		log.Println("Binary Write failed at new headersReq")
		return nil, err
	}

	s := checkSum(p.Bytes())
	m, err := newMsgHeader("getheaders", s, uint32(len(p.Bytes())))
	m = append(m, p.Bytes()...)

	str := hex.EncodeToString(m)
	log.Printf("The message length is %d, %s", len(m), str)
	return m, nil
}

func (msg Msg) preVerify() error {
	// TODO verify the message header
	// checksum,version magic number
	return nil
}

func (msg Msg) verify(t string) error {
	// TODO verify the message Content
	switch t {
	case "version":
	case "verACK":
	case "inventory":
	default:
		log.Println("Unknow message type to parse")
		return errors.New("Unknown message type to parse")
	}
	return nil
}

func (msg Msg) serialization() ([]byte, error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *Msg) deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, msg)
	return err
}
