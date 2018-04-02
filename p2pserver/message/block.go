package message

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/types"
	. "github.com/Ontology/p2pserver/common"
)

type BlockReq struct {
	msgHdr
	//TBD
}

type Block struct {
	msgHdr
	Blk types.Block
	// TBD
	//event *events.Event
}

func NewBlock(bk *types.Block) ([]byte, error) {
	log.Debug()
	var msg Block
	msg.Blk = *bk
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
	log.Debug("The message payload length is ", msg.msgHdr.Length)

	m, err := msg.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}

	return m, nil
}

func (msg Block) Verify(buf []byte) error {
	err := msg.msgHdr.Verify(buf)
	// TODO verify the message Content
	return err
}

func (msg Block) Serialization() ([]byte, error) {
	hdrBuf, err := msg.msgHdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	msg.Blk.Serialize(buf)

	return buf.Bytes(), err
}

func (msg *Block) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)

	err := binary.Read(buf, binary.LittleEndian, &(msg.msgHdr))
	if err != nil {
		log.Warn("Parse block message hdr error")
		return errors.New("Parse block message hdr error")
	}

	err = msg.Blk.Deserialize(buf)
	if err != nil {
		log.Warn("Parse block message error")
		return errors.New("Parse block message error")
	}

	return err
}
