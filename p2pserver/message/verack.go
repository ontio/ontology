package message

import (
	"bytes"
	"encoding/binary"

	"github.com/Ontology/common/serialization"
)

type VerACK struct {
	MsgHdr
	IsConsensus bool
}


func (msg VerACK) Serialization() ([]byte, error) {
	hdrBuf, err := msg.MsgHdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	err = serialization.WriteBool(buf, msg.IsConsensus)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}

func (msg *VerACK) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.MsgHdr))
	if err != nil {
		return err
	}

	msg.IsConsensus, err = serialization.ReadBool(buf)
	return err
}
