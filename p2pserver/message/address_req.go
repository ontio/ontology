package message

import (
	"bytes"
	"encoding/binary"
)

type AddrReq struct {
	Hdr MsgHdr
}

func (msg AddrReq) Verify(buf []byte) error {
	err := msg.Hdr.Verify(buf)
	return err
}

func (msg AddrReq) Serialization() ([]byte, error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}

func (msg *AddrReq) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, msg)
	return err
}