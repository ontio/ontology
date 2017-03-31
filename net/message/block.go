package message

import (
	"DNA/common"
	"DNA/common/log"
	"DNA/core/ledger"
	"DNA/events"
	. "DNA/net/protocol"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
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

	log.Debug("RX block message")
	err := ledger.DefaultLedger.Blockchain.AddBlock(&msg.blk)
	if err != nil {
		log.Warn("Add block error")
		return errors.New("Add block error before Xmit\n")
	}
	node.LocalNode().GetEvent("block").Notify(events.EventNewInventory, &msg.blk)
	return nil
}

func (msg dataReq) Handle(node Noder) error {
	common.Trace()
	reqtype := common.InventoryType(msg.dataType)
	hash := msg.hash
	switch reqtype {
	case common.BLOCK:
		block, err := NewBlockFromHash(hash)
		if err != nil {
			return err
		}
		buf, err := NewBlock(block)
		if err != nil {
			return err
		}
		go node.Tx(buf)

	case common.TRANSACTION:
		txn, err := NewTxnFromHash(hash)
		if err != nil {
			return err
		}
		buf, err := NewTxn(txn)
		if err != nil {
			return err
		}
		go node.Tx(buf)
	}
	return nil
}

func NewBlockFromHash(hash common.Uint256) (*ledger.Block, error) {
	bk, err := ledger.DefaultLedger.Store.GetBlock(hash)
	if err != nil {
		log.Error("Get Block error: ", err.Error())
		return nil, err
	}
	return bk, nil
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
		log.Error("Binary Write failed at new Msg")
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
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}

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
	hdrBuf, err := msg.msgHdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	msg.blk.Serialize(buf)

	return buf.Bytes(), err
}

func (msg *block) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)

	err := binary.Read(buf, binary.LittleEndian, &(msg.msgHdr))
	if err != nil {
		log.Warn("Parse block message hdr error")
		return errors.New("Parse block message hdr error")
	}

	err = msg.blk.Deserialize(buf)
	if err != nil {
		log.Warn("Parse block message error")
		return errors.New("Parse block message error")
	}

	return err
}
