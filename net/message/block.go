package message

import (
	"GoOnchain/common"
	"GoOnchain/core/ledger"
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
	blk []byte
	// TBD
	//event *events.Event
}

func (msg block) Handle(node Noder) error {
	common.Trace()

	fmt.Printf("RX block message\n")
	/*
		if !node.ExistedID(msg.blk.Hash()) {
			// TODO Update the currently ledger
			// FIXME the relative event should be attached to the message

			if msg.event != nil {
				msg.event.Notify(events.EventSaveBlock, msg.blk)
			}

		}
	*/
	return nil
}

func (msg dataReq) Handle(node Noder) error {
	common.Trace()
	reqtype := msg.dataType
	hash := msg.hash
	switch reqtype {
	case 0x01:
		buf, _ := NewBlock(hash)
		go node.LocalNode().Tx(buf)
	case 0x02:
		buf, _ := NewTx(hash)
		go node.LocalNode().Tx(buf)
	}
	return nil
}
func NewBlock(hash common.Uint256) ([]byte, error) {
	common.Trace()
	var msg block
	//FIXME no error
	bk, _ := ledger.DefaultLedger.Store.GetBlock(hash)
	msg.msgHdr.Magic = NETMAGIC
	ver := "block"
	copy(msg.msgHdr.CMD[0:7], ver)
	tmpBuffer := bytes.NewBuffer([]byte{})
	bk.Serialize(tmpBuffer)
	msg.blk = tmpBuffer.Bytes()
	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, &(msg.blk))
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
	var buf bytes.Buffer

	fmt.Printf("The size of messge is %d in serialization\n",
		uint32(unsafe.Sizeof(msg)))
	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *block) Deserialization(p []byte) error {
	fmt.Printf("The size of messge is %d in deserialization\n",
		uint32(unsafe.Sizeof(*msg)))

	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, msg)
	return err
}
