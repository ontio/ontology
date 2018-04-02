package message

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	//"errors"
	"github.com/Ontology/common/log"
	"github.com/Ontology/common/serialization"
	. "github.com/Ontology/p2pserver/common"
)

type VerACK struct {
	msgHdr
	IsConsensus bool
}

func NewVerack(isConsensus bool) ([]byte, error) {
	var msg VerACK
	msg.msgHdr.Magic = NETMAGIC
	copy(msg.msgHdr.CMD[0:7], "verack")
	msg.IsConsensus = isConsensus
	tmpBuffer := bytes.NewBuffer([]byte{})
	serialization.WriteBool(tmpBuffer, msg.IsConsensus)
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

func (msg VerACK) Serialization() ([]byte, error) {
	hdrBuf, err := msg.msgHdr.Serialization()
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
	err := binary.Read(buf, binary.LittleEndian, &(msg.msgHdr))
	if err != nil {
		return err
	}

	msg.IsConsensus, err = serialization.ReadBool(buf)
	return err
}
