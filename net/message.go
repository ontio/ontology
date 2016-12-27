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
)

// The network and module communication message buffer
type Msg struct {
	Magic	 uint32
	CMD	 [MSGCMDLEN]byte 	// the message command (message type)
	Length   uint32
	Checksum uint32
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
	hashStart	[32]byte
	hashEnd		[32]byte
}

// Sample function, shoule be called from ledger module
func ledgerGetHeader() [32]byte {
	var t [32]byte
	return t
}

// TODO combine all of message alloc in one function via interface
func newMsg(p interface{}) (*Msg, error) {
	msg := new(Msg)
	switch t := p.(type) {
	case version:
		log.Printf("Port is %d", t.port)
	case verACK:
	case headersReq:
		t.hashStart = ledgerGetHeader()
	default:
		return nil, errors.New("Unknown message type")
	}

	//msg.payloader = p
	return msg, nil
}

func newVersionMsg() (*Msg, error) {
	var v version

	// TODO Need Node read lock or channel
	v.version = nodes.node.version
	v.services = nodes.node.services
	v.timeStamp = uint32(time.Now().UTC().UnixNano())
	v.port = nodes.node.port
	v.nonce = nodes.node.nonce
	//v.userAgent =
	// Fixme Get the block height from ledger
	v.startHeight = 0
	if nodes.node.relay {
		v.relay = 1
	} else {
		v.relay = 0
	}

	msg := new(Msg)
	msg.Magic = NETMAGIC
	ver := "version"
	copy(msg.CMD[0:7], ver)
	msg.Length = uint32(unsafe.Sizeof(v))
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, &v)
	if err != nil {
		log.Println("Binary Write failed at new Msg")
		return nil, err
	}

	//msg.Checksum = rc32.ChecksumIEEE(buf.Bytes())
	//msg.payloader = &ver

	s := sha256.Sum256(buf.Bytes())
	buf = bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(msg.Checksum))

	return msg, nil
}

func newVersionBuf() ([]byte, error) {
	common.Trace()
	var v version

	// TODO Need Node read lock or channel
	v.version = nodes.node.version
	v.services = nodes.node.services
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

func newVerackBuf() []byte {
	//var verACK verACK
	msg := new(Msg)

	msg.Magic = NETMAGIC
	v := "verack"
	copy(msg.CMD[0:6], v)
	msg.Length = 0
	msg.Checksum = 0
	//msg.payloader = &verACK

	buf, err := msg.serialization()
	if (err != nil) {
		log.Println("Error Convert net message ", err.Error())
		return nil
	}

	return buf
}

func newHeadersReqBuf() []byte {
	//var headersReq headersReq
	msg := new(Msg)
	//msg.payloader = &headersReq

	buf, err := msg.serialization()
	if (err != nil) {
		log.Println("Error Convert net message ", err.Error())
		return nil
	}

	return buf
}

func newVerackMsg() *Msg {
	//var verACK verACK
	msg := new(Msg)

	msg.Magic = NETMAGIC
	v := "verack"
	copy(msg.CMD[0:6], v)
	msg.Length = 0
	msg.Checksum = 0
	//msg.payloader = &verACK
	return msg
}

func newHeadersReqMsg() *Msg {
	//var headersReq headersReq
	msg := new(Msg)
	//msg.payloader = &headersReq
	return msg
}

func (msg Msg) serialization() ([]byte, error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, msg)
	return buf.Bytes(), err
}

func (msg *Msg) deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, msg)
	return err
}
