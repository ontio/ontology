package message

import (
	"bytes"
	"encoding/binary"

	"github.com/Ontology/common/log"
	"github.com/Ontology/core/types"
)

type BlkHeader struct {
	Hdr    MsgHdr
	Cnt    uint32
	BlkHdr []types.Header
}

func (msg BlkHeader) Verify(buf []byte) error {
	err := msg.Hdr.Verify(buf)
	return err
}

func (msg BlkHeader) Serialization() ([]byte, error) {
	hdrBuf, err := msg.Hdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	err = binary.Write(buf, binary.LittleEndian, msg.Cnt)
	if err != nil {
		return nil, err
	}

	for _, header := range msg.BlkHdr {
		header.Serialize(buf)
	}
	return buf.Bytes(), err
}

func (msg *BlkHeader) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.Hdr))
	if err != nil {
		return err
	}

	err = binary.Read(buf, binary.LittleEndian, &(msg.Cnt))
	if err != nil {
		return err
	}

	for i := 0; i < int(msg.Cnt); i++ {
		var headers types.Header
		err := (&headers).Deserialize(buf)
		msg.BlkHdr = append(msg.BlkHdr, headers)
		if err != nil {
			log.Debug("blkHeader Deserialization failed")
			goto blkHdrErr
		}
	}

blkHdrErr:
	return err
}


