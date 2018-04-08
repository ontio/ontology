package message

import (
	"bytes"
	"encoding/binary"

	"github.com/Ontology/p2pserver/common"
)

type HeadersReq struct {
	Hdr MsgHdr
	P   struct {
		Len       uint8
		HashStart [common.HASH_LEN]byte
		HashEnd   [common.HASH_LEN]byte
	}
}

func (msg HeadersReq) Verify(buf []byte) error {
	err := msg.Hdr.Verify(buf)
	return err
}

func (msg HeadersReq) Serialization() ([]byte, error) {
	hdrBuf, err := msg.Hdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	err = binary.Write(buf, binary.LittleEndian, msg.P.Len)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.LittleEndian, msg.P.HashStart)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.LittleEndian, msg.P.HashEnd)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *HeadersReq) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.Hdr))
	if err != nil {
		return err
	}

	err = binary.Read(buf, binary.LittleEndian, &(msg.P.Len))
	if err != nil {
		return err
	}

	err = binary.Read(buf, binary.LittleEndian, &(msg.P.HashStart))
	if err != nil {
		return err
	}

	err = binary.Read(buf, binary.LittleEndian, &(msg.P.HashEnd))
	return err
}
