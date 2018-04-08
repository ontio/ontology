package message

import (
	"bytes"
	"encoding/binary"

	"github.com/Ontology/common/log"
)

type Consensus struct {
	MsgHdr
	Cons ConsensusPayload
}

func (msg *Consensus) Serialization() ([]byte, error) {
	hdrBuf, err := msg.MsgHdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	err = msg.Cons.Serialize(buf)
	return buf.Bytes(), err
}

func (msg *Consensus) Deserialization(p []byte) error {
	log.Debug()
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.MsgHdr))
	err = msg.Cons.Deserialize(buf)
	return err
}