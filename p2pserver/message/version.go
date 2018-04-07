package message

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/Ontology/common/log"
	"github.com/Ontology/crypto"
)

const (
	HTTP_INFO_FLAG = 0
)

type VersionPayload struct {
	Version       uint32
	Services      uint64
	TimeStamp     uint32
	Port          uint16
	HttpInfoPort  uint16
	ConsensusPort uint16
	Cap           [32]byte
	Nonce         uint64
	// TODO remove tempory to get serilization function passed
	UserAgent   uint8
	StartHeight uint64
	// FIXME check with the specify relay type length
	Relay       uint8
	IsConsensus bool
}
type Version struct {
	Hdr msgHdr
	P   VersionPayload
	PK  *crypto.PubKey
}

func (msg Version) Verify(buf []byte) error {
	err := msg.Hdr.Verify(buf)
	// TODO verify the message Content
	// TODO check version compatible or not
	return err
}

func (msg Version) Serialization() ([]byte, error) {
	hdrBuf, err := msg.Hdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	err = binary.Write(buf, binary.LittleEndian, msg.P)
	if err != nil {
		return nil, err
	}
	msg.PK.Serialize(buf)

	return buf.Bytes(), err
}

func (msg *Version) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)

	err := binary.Read(buf, binary.LittleEndian, &(msg.Hdr))
	if err != nil {
		log.Warn("Parse version message hdr error")
		return errors.New("Parse version message hdr error")
	}

	err = binary.Read(buf, binary.LittleEndian, &(msg.P))
	if err != nil {
		log.Warn("Parse version P message error")
		return errors.New("Parse version P message error")
	}

	pk := new(crypto.PubKey)
	err = pk.DeSerialize(buf)
	if err != nil {
		return errors.New("Parse pubkey Deserialize failed.")
	}
	msg.PK = pk
	return err
}
