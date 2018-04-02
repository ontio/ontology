package message

import (
	"bytes"
	"encoding/binary"
	"errors"
	"crypto/sha256"

	"github.com/Ontology/common/log"
	"github.com/Ontology/crypto"
	msgCommon "github.com/Ontology/p2pserver/common"
)

const (
	HTTPINFOFLAG = 0
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
	P	VersionPayload
	pk *crypto.PubKey
}


func NewVersion(vpl VersionPayload, pk *crypto.PubKey) ([]byte, error) {
	log.Debug()
	var msg Version

/*	msg.P.Version = n.Version()
	msg.P.Services = n.Services()
	msg.P.HttpInfoPort = config.Parameters.HttpInfoPort
	msg.P.ConsensusPort = n.GetConsensusPort()
	msg.P.IsConsensus = isConsensus
	if config.Parameters.HttpInfoStart {
		msg.P.Cap[HTTPINFOFLAG] = 0x01
	} else {
		msg.P.Cap[HTTPINFOFLAG] = 0x00
	}

	// FIXME Time overflow
	msg.P.TimeStamp = uint32(time.Now().UTC().UnixNano())
	msg.P.Port = n.GetPort()
	msg.P.Nonce = n.GetID()
	msg.P.UserAgent = 0x00
	height, _ := actor.GetCurrentBlockHeight()
	msg.P.StartHeight = uint64(height)
	if n.GetRelay() {
		msg.P.Relay = 1
	} else {
		msg.P.Relay = 0
	}
	msg.pk = n.GetBookKeeperAddr()
*/
	msg.P = vpl
	msg.pk = pk
	log.Debug("new version msg.pk is ", msg.pk)
	// TODO the function to wrap below process
	// msg.HDR.init("version", n.GetID(), uint32(len(p.Bytes())))

	msg.Hdr.Magic = msgCommon.NETMAGIC
	copy(msg.Hdr.CMD[0:7], "version")
	p := bytes.NewBuffer([]byte{})
	err := binary.Write(p, binary.LittleEndian, &(msg.P))
	msg.pk.Serialize(p)
	if err != nil {
		log.Error("Binary Write failed at new Msg")
		return nil, err
	}
	s := sha256.Sum256(p.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(msg.Hdr.Checksum))
	msg.Hdr.Length = uint32(len(p.Bytes()))
	log.Debug("The message payload length is ", msg.Hdr.Length)

	m, err := msg.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}

	return m, nil
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
	msg.pk.Serialize(buf)

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
	msg.pk = pk
	return err
}
