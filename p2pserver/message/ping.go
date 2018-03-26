package message

import (
	"bytes"
	//"crypto/sha256"
	"encoding/binary"
	//"github.com/Ontology/common/log"
	"github.com/Ontology/common/serialization"
	//. "github.com/Ontology/p2pserver/common"
)

type ping struct {
	msgHdr
	height uint64
}

func (msg ping) Verify(buf []byte) error {
	err := msg.msgHdr.Verify(buf)
	// TODO verify the message Content
	return err
}

func (msg ping) Serialization() ([]byte, error) {
	hdrBuf, err := msg.msgHdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	err = serialization.WriteUint64(buf, msg.height)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err

}

func (msg *ping) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.msgHdr))
	if err != nil {
		return err
	}

	msg.height, err = serialization.ReadUint64(buf)
	return err
}
