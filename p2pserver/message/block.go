package message

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/Ontology/common/log"
	"github.com/Ontology/core/types"
)

type Block struct {
	MsgHdr
	Blk types.Block
}

func (msg Block) Verify(buf []byte) error {
	err := msg.MsgHdr.Verify(buf)
	return err
}

func (msg Block) Serialization() ([]byte, error) {
	hdrBuf, err := msg.MsgHdr.Serialization()
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(hdrBuf)
	msg.Blk.Serialize(buf)
	return buf.Bytes(), err
}

func (msg *Block) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.MsgHdr))
	if err != nil {
		log.Warn("Parse block message hdr error")
		return errors.New("Parse block message hdr error ")
	}

	err = msg.Blk.Deserialize(buf)
	if err != nil {
		log.Warn("Parse block message error")
		return errors.New("Parse block message error ")
	}

	return err
}
