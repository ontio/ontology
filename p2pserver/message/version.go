package message

import (
	"bytes"
	//"crypto/sha256"
	"encoding/binary"
	"errors"
	//"fmt"
	//"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	//	"github.com/Ontology/core/ledger"
	"github.com/Ontology/crypto"
	//actor "github.com/Ontology/p2pserver/actor/req"
	//"time"
)

const (
	HTTPINFOFLAG = 0
)

type version struct {
	Hdr msgHdr
	P   struct {
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
	pk *crypto.PubKey
}

func (msg version) Verify(buf []byte) error {
	err := msg.Hdr.Verify(buf)
	// TODO verify the message Content
	// TODO check version compatible or not
	return err
}

func (msg version) Serialization() ([]byte, error) {
	hdrBuf, err := msg.Hdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	err = binary.Write(buf, binary.LittleEndian, msg.P)
	if err != nil {
		return nil, err
	}
	msg.pk.Serialize(buf)

	return buf.Bytes(), err
}

func (msg *version) Deserialization(p []byte) error {
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
	msg.pk = pk
	return err
}
