package message

import (
	"bytes"
	"encoding/binary"

	"github.com/Ontology/common/serialization"
)

type Ping struct {
	Hdr    MsgHdr
	Height uint64
}


func (msg Ping) Verify(buf []byte) error {
	err := msg.Hdr.Verify(buf)
	return err
}

func (msg Ping) Serialization() ([]byte, error) {
	hdrBuf, err := msg.Hdr.Serialization()
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

func (msg *Ping) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.Hdr))
	if err != nil {
		return err
	}

	msg.Height, err = serialization.ReadUint64(buf)
	return err
}
