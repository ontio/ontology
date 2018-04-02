package message

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	//"errors"
	//"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"github.com/Ontology/common/serialization"
	"github.com/Ontology/core/types"
	. "github.com/Ontology/p2pserver/common"
)

type HeadersReq struct {
	hdr msgHdr
	P   struct {
		Len       uint8
		HashStart [HASHLEN]byte
		HashEnd   [HASHLEN]byte
	}
}

type BlkHeader struct {
	hdr    msgHdr
	Cnt    uint32
	BlkHdr []types.Header
}

func (msg HeadersReq) Verify(buf []byte) error {
	// TODO Verify the message Content
	err := msg.hdr.Verify(buf)
	return err
}

func (msg BlkHeader) Verify(buf []byte) error {
	// TODO Verify the message Content
	err := msg.hdr.Verify(buf)
	return err
}

func (msg HeadersReq) Serialization() ([]byte, error) {
	hdrBuf, err := msg.hdr.Serialization()
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
	err := binary.Read(buf, binary.LittleEndian, &(msg.hdr))
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

func (msg BlkHeader) Serialization() ([]byte, error) {
	hdrBuf, err := msg.hdr.Serialization()
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
	err := binary.Read(buf, binary.LittleEndian, &(msg.hdr))
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

func NewHeaders(headers []types.Header, count uint32) ([]byte, error) {
	var msg BlkHeader
	msg.Cnt = count
	msg.BlkHdr = headers
	msg.hdr.Magic = NETMAGIC
	cmd := "headers"
	copy(msg.hdr.CMD[0:len(cmd)], cmd)

	tmpBuffer := bytes.NewBuffer([]byte{})
	serialization.WriteUint32(tmpBuffer, msg.Cnt)
	for _, header := range headers {
		header.Serialize(tmpBuffer)
	}
	b := new(bytes.Buffer)
	err := binary.Write(b, binary.LittleEndian, tmpBuffer.Bytes())
	if err != nil {
		log.Error("Binary Write failed at new Msg")
		return nil, err
	}
	s := sha256.Sum256(b.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(msg.hdr.Checksum))
	msg.hdr.Length = uint32(len(b.Bytes()))

	m, err := msg.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}
