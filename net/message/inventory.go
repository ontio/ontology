package message

import (
	"GoOnchain/common"
	"GoOnchain/common/serialization"
	"GoOnchain/core/ledger"
	. "GoOnchain/net/protocol"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"unsafe"
)

type InventoryType byte

type blocksReq struct {
	Hdr             msgHdr
	HeaderHashCount []byte
	HashStart       []common.Uint256
	HashStop        common.Uint256
}

type invPayload struct {
	InvType uint8
	Blk     []byte
}

type Inv struct {
	Hdr msgHdr
	P   invPayload
}

func (msg blocksReq) Verify(buf []byte) error {

	// TODO verify the message Content
	err := msg.Hdr.Verify(buf)
	return err
}

func (msg blocksReq) Handle(node Noder) error {
	common.Trace()

	var starthash []common.Uint256
	var stophash common.Uint256
	starthash = msg.HashStart
	stophash = msg.HashStop
	//FIXME if HeaderHashCount > 1
	inv := GetInvFromBlockHash(starthash[0], stophash)
	buf, _ := NewInv(inv)
	go node.Tx(buf)
	return nil
}

func (msg blocksReq) Serialization() ([]byte, error) {
	var buf bytes.Buffer

	fmt.Printf("The size of messge is %d in serialization\n",
		uint32(unsafe.Sizeof(msg)))
	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *blocksReq) Deserialization(p []byte) error {
	fmt.Printf("The size of messge is %d in deserialization\n",
		uint32(unsafe.Sizeof(*msg)))

	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, msg)
	return err
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
	fmt.Printf("The size of messge is %d in serialization\n",
		uint32(unsafe.Sizeof(msg)))
	hdrBuf, err := msg.Hdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	msg.P.Serialization(buf)

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

func GetInvFromBlockHash(starthash common.Uint256, stophash common.Uint256) invPayload {
	//FIXME need add error handle for GetBlockWithHash
	var stopheight uint32
	var count uint32 = 0
	var i uint32
	/*wait for GetBlockWithHash commit
	var empty common.Uint256
	bkstart, _ := ledger.DefaultLedger.Blockchain.GetBlockWithHash(starthash)
	startheight := bkstart.Blockdata.Height
	if (stophash != empty){
		bkstop, _ := ledger.DefaultLedger.Blockchain.GetBlockWithHash(starthash)
		stopheight = bkstop.Blockdata.Height
		count = startheight - stopheight
		if (count >= 500) {
			count = 500
		}
	}else{
		count = 500
	}
	*/

	tmpBuffer := bytes.NewBuffer([]byte{})
	for i = 1; i <= count; i++ {
		//FIXME need add error handle for GetBlockWithHash
		hash, _ := ledger.DefaultLedger.Store.GetBlockHash(stopheight + i)
		hash.Serialize(tmpBuffer)
	}
	var inv invPayload
	inv.Blk = tmpBuffer.Bytes()
	inv.InvType = 0x02
	return inv
}

func NewInv(inv invPayload) ([]byte, error) {
	var msg Inv

	msg.P.Blk = inv.Blk
	msg.P.InvType = inv.InvType
	msg.Hdr.Magic = NETMAGIC
	cmd := "inv"
	copy(msg.Hdr.CMD[0:len(cmd)], cmd)
	tmpBuffer := bytes.NewBuffer([]byte{})
	inv.Serialization(tmpBuffer)

	b := new(bytes.Buffer)
	err := binary.Write(b, binary.LittleEndian, tmpBuffer.Bytes())
	if err != nil {
		fmt.Println("Binary Write failed at new Msg")
		return nil, err
	}
	s := sha256.Sum256(b.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(msg.Hdr.Checksum))
	msg.Hdr.Length = uint32(len(buf.Bytes()))
	fmt.Printf("The message payload length is %d\n", msg.Hdr.Length)

	m, err := msg.Serialization()
	if err != nil {
		fmt.Println("Error Convert net message ", err.Error())
		return nil, err
	}

	str := hex.EncodeToString(m)
	fmt.Printf("The message length is %d, %s\n", len(m), str)
	return m, nil
}

func (msg *invPayload) Serialization(w io.Writer) {
	serialization.WriteUint8(w, msg.InvType)
	serialization.WriteVarBytes(w, msg.Blk)
}
