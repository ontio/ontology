package message

import (
	"bytes"
	"encoding/binary"

	"github.com/Ontology/core/types"
)

// Transaction message
type Trn struct {
	MsgHdr
	Txn types.Transaction
}


func (msg Trn) Serialization() ([]byte, error) {
	hdrBuf, err := msg.MsgHdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	msg.Txn.Serialize(buf)

	return buf.Bytes(), err
}

func (msg *Trn) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.MsgHdr))
	err = msg.Txn.Deserialize(buf)
	if err != nil {
		return err
	}
	return nil
}

type txnPool struct {
	MsgHdr
}
