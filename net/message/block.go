package message

import (
	"GoOnchain/common"
	"GoOnchain/core/ledger"
	"GoOnchain/events"
	//"GoOnchain/events"
	. "GoOnchain/net/protocol"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"unsafe"
)

type blockReq struct {
	msgHdr
	//TBD
}

type block struct {
	msgHdr
	blk ledger.Block
	// TBD
	//event *events.Event
}

func (msg block) Handle(node Noder) error {
	common.Trace()

	fmt.Printf("RX block message\n")
	node.LocalNode().GetEvent("block").Notify(events.EventNewInventory, &msg.blk)
	return nil
}

func (msg dataReq) Handle(node Noder) error {
	common.Trace()
	reqtype := msg.dataType
	hash := msg.hash
	switch reqtype {
	case 0x01:
		block := NewBlockFromHash(hash)
		buf, _ := NewBlock(block)
		go node.Tx(buf)

	case 0x02:
		tx := NewTxFromHash(hash)
		buf, _ := NewTx(tx)
		go node.Tx(buf)
	}
	return nil
}

func NewBlockFromHash(hash common.Uint256) *ledger.Block {
	bk, _ := ledger.DefaultLedger.Store.GetBlock(hash)
	return bk
}

func NewBlock(bk *ledger.Block) ([]byte, error) {
	common.Trace()
	var msg block
	msg.blk = *bk
	msg.msgHdr.Magic = NETMAGIC
	cmd := "block"
	copy(msg.msgHdr.CMD[0:len(cmd)], cmd)
	tmpBuffer := bytes.NewBuffer([]byte{})
	bk.Serialize(tmpBuffer)
	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, tmpBuffer.Bytes())
	if err != nil {
		fmt.Println("Binary Write failed at new Msg")
		return nil, err
	}
	s := sha256.Sum256(p.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(msg.msgHdr.Checksum))
	msg.msgHdr.Length = uint32(len(p.Bytes()))
	fmt.Printf("The message payload length is %d\n", msg.msgHdr.Length)

	m, err := msg.Serialization()
	if err != nil {
		fmt.Println("Error Convert net message ", err.Error())
		return nil, err
	}

	str := hex.EncodeToString(m)
	fmt.Printf("The message length is %d, %s\n", len(m), str)
	return m, nil
}

func reqBlkData(node Noder, hash common.Uint256) error {
	var msg dataReq
	msg.dataType = common.BLOCK
	// TODO handle the hash array case
	msg.hash = hash

	buf, _ := msg.Serialization()
	go node.Tx(buf)

	return nil
}

func (msg block) Verify(buf []byte) error {
	err := msg.msgHdr.Verify(buf)
	// TODO verify the message Content
	return err
}

func (msg block) Serialization() ([]byte, error) {

	fmt.Printf("The size of messge is %d in serialization\n",
		uint32(unsafe.Sizeof(msg)))
	hdrBuf, err := msg.msgHdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	msg.blk.Serialize(buf)

	return buf.Bytes(), err
}

func (msg *block) Deserialization(p []byte) error {
	fmt.Printf("The size of messge is %d in deserialization\n",
		uint32(unsafe.Sizeof(*msg)))

	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, msg.msgHdr)
	msg.blk.Deserialize(buf)
	return err
}
