package message

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
)

type NotFound struct {
	MsgHdr
	Hash common.Uint256
}



func (msg NotFound) Verify(buf []byte) error {
	err := msg.MsgHdr.Verify(buf)
	return err
}

func (msg NotFound) Serialization() ([]byte, error) {
	hdrBuf, err := msg.MsgHdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	msg.Hash.Serialize(buf)

	return buf.Bytes(), err
}

func (msg *NotFound) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)

	err := binary.Read(buf, binary.LittleEndian, &(msg.MsgHdr))
	if err != nil {
		log.Warn("Parse notFound message hdr error")
		return errors.New("Parse notFound message hdr error ")
	}

	err = msg.Hash.Deserialize(buf)
	if err != nil {
		log.Warn("Parse notFound message error")
		return errors.New("Parse notFound message error ")
	}

	return err
}
