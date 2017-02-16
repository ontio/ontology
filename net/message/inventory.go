package message

import (
	"GoOnchain/common"
	. "GoOnchain/net/protocol"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"unsafe"
)

type invPayload struct {
	InvType uint8
	Blk     []byte
}

type Inv struct {
	Hdr msgHdr
	P   invPayload
}

func (msg Inv) Verify(buf []byte) error {
	// TODO verify the message Content
	err := msg.Hdr.Verify(buf)
	return err
}

func (msg Inv) Handle(node Noder) error {
	common.Trace()
	var id common.Uint256
	str := hex.EncodeToString(msg.P.Blk)
	fmt.Printf("The inv type: 0x%x block len: %d, %s\n",
		msg.P.InvType, len(msg.P.Blk), str)

	invType := common.InventoryType(msg.P.InvType)
	switch invType {
	case common.TRANSACTION:
		fmt.Printf("RX TRX message\n")
		// TODO check the ID queue
		id.Deserialize(bytes.NewReader(msg.P.Blk[:32]))
		if !node.ExistedID(id) {
			reqTxnData(node, id)
		}
	case common.BLOCK:
		fmt.Printf("RX block message\n")
		id.Deserialize(bytes.NewReader(msg.P.Blk[:32]))
		if !node.ExistedID(id) {
			// send the block request
			reqBlkData(node, id)
		}
	case common.CONSENSUS:
		fmt.Printf("RX consensus message\n")
		id.Deserialize(bytes.NewReader(msg.P.Blk[:32]))
		reqConsensusData(node, id)
	default:
		fmt.Printf("RX unknown inventory message\n")
		// Warning:
	}
	return nil
}

func (msg Inv) Serialization() ([]byte, error) {
	var buf bytes.Buffer

	fmt.Printf("The size of messge is %d in serialization\n",
		uint32(unsafe.Sizeof(msg)))

	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *Inv) Deserialization(p []byte) error {
	fmt.Printf("The size of messge is %d in deserialization\n",
		uint32(unsafe.Sizeof(*msg)))

	err := msg.Hdr.Deserialization(p)

	msg.P.InvType = p[MSGHDRLEN]
	msg.P.Blk = p[MSGHDRLEN+1:]
	return err
}

func (msg Inv) invType() byte {
	return msg.P.InvType
}

//func (msg inv) invLen() (uint64, uint8) {
func (msg Inv) invLen() (uint64, uint8) {
	var val uint64
	var size uint8

	len := binary.LittleEndian.Uint64(msg.P.Blk[0:1])
	if len < 0xfd {
		val = len
		size = 1
	} else if len == 0xfd {
		val = binary.LittleEndian.Uint64(msg.P.Blk[1:3])
		size = 3
	} else if len == 0xfe {
		val = binary.LittleEndian.Uint64(msg.P.Blk[1:5])
		size = 5
	} else if len == 0xff {
		val = binary.LittleEndian.Uint64(msg.P.Blk[1:9])
		size = 9
	}

	return val, size
}
