package message

import (
	"fmt"
	"bytes"
	"unsafe"
	"encoding/binary"
	"encoding/hex"
	"GoOnchain/common"
	. "GoOnchain/net/protocol"
	sig "GoOnchain/core/signature"
)

type InventoryType byte

const (
	Transaction	InventoryType = 0x01
	Block		InventoryType = 0x02
	Consensus	InventoryType = 0xe0
)

//TODO: temp inventory
type Inventory interface {
	sig.SignableData
	Hash() common.Uint256
	Verify() error
	InvertoryType() InventoryType
}


type invPayload struct {
	invType uint8
	blk     []byte
}

type inv struct {
	hdr msgHdr
	p  invPayload
}

func (msg inv) Verify(buf []byte) error {
	// TODO verify the message Content
	err := msg.hdr.Verify(buf)
	return err
}

func (msg inv) Handle(node *Noder) error {
	common.Trace()
	str := hex.EncodeToString(msg.p.blk)
	fmt.Printf("The inv type: 0x%x block len: %d, %s\n",
		msg.p.invType, len(msg.p.blk), str)

	switch msg.p.invType {
	case TXN:
		fmt.Printf("RX TRX message\n")
	case BLOCK:
		fmt.Printf("RX block message\n")
	case CONSENSUS:
		fmt.Printf("RX consensus message\n")
	default:
		fmt.Printf("RX unknown inventory message\n")
		// Warning:
	}
	// notice event inventory
	return nil
}

func (msg inv) Serialization() ([]byte, error) {
	var buf bytes.Buffer

	fmt.Printf("The size of messge is %d in serialization\n",
		uint32(unsafe.Sizeof(msg)))

	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *inv) Deserialization(p []byte) error {
	fmt.Printf("The size of messge is %d in deserialization\n",
		uint32(unsafe.Sizeof(*msg)))

	err := msg.hdr.Deserialization(p)

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
