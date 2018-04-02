package message

import (
	"bytes"
	"encoding/binary"
	"crypto/sha256"
	"github.com/Ontology/common/log"
	"github.com/Ontology/common/serialization"
	. "github.com/Ontology/p2pserver/common"
)

type Ping struct {
	msgHdr
	Height uint64
}

func NewPingMsg(height uint64) ([]byte, error) {
	var msg Ping
	msg.msgHdr.Magic = NETMAGIC
	copy(msg.msgHdr.CMD[0:7], "ping")
	msg.Height = uint64(height)
	tmpBuffer := bytes.NewBuffer([]byte{})
	serialization.WriteUint64(tmpBuffer, msg.Height)
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
	binary.Read(buf, binary.LittleEndian, &(msg.msgHdr.Checksum))
	msg.msgHdr.Length = uint32(len(b.Bytes()))

	m, err := msg.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}

func (msg Ping) Verify(buf []byte) error {
	err := msg.msgHdr.Verify(buf)
	// TODO verify the message Content
	return err
}

func (msg Ping) Serialization() ([]byte, error) {
	hdrBuf, err := msg.msgHdr.Serialization()
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
	err := binary.Read(buf, binary.LittleEndian, &(msg.msgHdr))
	if err != nil {
		return err
	}

	msg.Height, err = serialization.ReadUint64(buf)
	return err
}
