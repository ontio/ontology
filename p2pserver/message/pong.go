package message

import (
	"bytes"
	"encoding/binary"

	"github.com/Ontology/common/serialization"
)

type Pong struct {
	MsgHdr
	Height uint64
}


func (msg Pong) Verify(buf []byte) error {
	err := msg.MsgHdr.Verify(buf)
	return err
}

func (msg Pong) Serialization() ([]byte, error) {
	hdrBuf, err := msg.MsgHdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	err = serialization.WriteUint64(buf, msg.Height)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err

}

func (msg *Pong) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.MsgHdr))
	if err != nil {
		return err
	}

	msg.Height, err = serialization.ReadUint64(buf)
	return err
}
