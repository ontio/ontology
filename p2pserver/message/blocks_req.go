package message

import (
	"bytes"
	"encoding/binary"

	"github.com/Ontology/p2pserver/common"
)

type BlocksReq struct {
	MsgHdr
	P struct {
		HeaderHashCount uint8
		HashStart       [common.HASH_LEN]byte
		HashStop        [common.HASH_LEN]byte
	}
}

func (msg BlocksReq) Verify(buf []byte) error {
	err := msg.MsgHdr.Verify(buf)
	return err
}

func (msg BlocksReq) Serialization() ([]byte, error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}

func (msg *BlocksReq) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &msg)
	return err
}
