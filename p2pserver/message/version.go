package message

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/p2pserver/actor"
	"github.com/ontio/ontology/p2pserver/protocol"
)

const (
	HTTP_INFO_FLAG = 0
)

type versionPayload struct {
	Version      uint32
	Services     uint64
	TimeStamp    uint32
	Port         uint16
	HttpInfoPort uint16
	Cap          [32]byte
	Nonce        uint64
	UserAgent    uint8
	StartHeight  uint64
	Relay        uint8
	IsConsensus  bool
}

type version struct {
	Hdr msgHdr
	P   versionPayload
	pk  keypair.PublicKey
}

func NewVersion(n protocol.Noder) ([]byte, error) {
	log.Debug()
	var msg version

	msg.P.Version = n.Version()
	msg.P.Services = n.Services()
	msg.P.HttpInfoPort = config.Parameters.HttpInfoPort
	if config.Parameters.HttpInfoPort > 0 {
		msg.P.Cap[HTTP_INFO_FLAG] = 0x01
	} else {
		msg.P.Cap[HTTP_INFO_FLAG] = 0x00
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

	msg.pk = n.GetBookkeeperAddr()
	log.Debug("new version msg.pk is ", msg.pk)

	msg.Hdr.Magic = protocol.NET_MAGIC
	copy(msg.Hdr.CMD[0:7], "version")
	p := bytes.NewBuffer([]byte{})
	err := binary.Write(p, binary.LittleEndian, &(msg.P))
	serialization.WriteVarBytes(p, keypair.SerializePublicKey(msg.pk))
	if err != nil {
		log.Error("Binary Write failed at new Msg")
		return nil, err
	}
	s := sha256.Sum256(p.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:protocol.CHECKSUM_LEN])
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

func (msg version) Verify(buf []byte) error {
	err := msg.Hdr.Verify(buf)
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
	keyBuf := keypair.SerializePublicKey(msg.pk)
	err = serialization.WriteVarBytes(buf, keyBuf)
	if err != nil {
		return nil, err
	}

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

	keyBuf, err := serialization.ReadVarBytes(buf)
	if err != nil {
		return errors.New("Parse pubkey Deserialize failed.")
	}
	pk, err := keypair.DeserializePublicKey(keyBuf)
	if err != nil {
		return errors.New("Parse pubkey Deserialize failed.")
	}
	msg.pk = pk
	return err
}

func (msg version) Handle(node protocol.Noder) error {
	log.Debug()
	localNode := node.LocalNode()

	// Exclude the node itself
	if msg.P.Nonce == localNode.GetID() {
		log.Warn("The node handshake with itself")
		node.CloseConn()
		return errors.New("The node handshake with itself")
	}

	s := node.GetState()
	if s != protocol.INIT && s != protocol.HAND {
		log.Warn("Unknow status to received version")
		return errors.New("Unknow status to received version")
	}

	// Obsolete node
	n, ret := localNode.OnDelNode(msg.P.Nonce)
	if ret == true {
		log.Info(fmt.Sprintf("Node reconnect 0x%x", msg.P.Nonce))
		// Close the connection and release the node soure
		n.SetState(protocol.INACTIVITY)
		NotifyPeerState(n.GetPubKey(), false)
		n.CloseConn()
	}

	log.Debug("handle version msg.pk is ", msg.pk)
	if msg.P.Cap[HTTP_INFO_FLAG] == 0x01 {
		node.SetHttpInfoState(true)
	} else {
		node.SetHttpInfoState(false)
	}
	node.SetHttpInfoPort(msg.P.HttpInfoPort)
	node.SetBookkeeperAddr(msg.pk)
	node.UpdateInfo(time.Now(), msg.P.Version, msg.P.Services,
		msg.P.Port, msg.P.Nonce, msg.P.Relay, msg.P.StartHeight)
	localNode.OnAddNode(node)

	var buf []byte
	if s == protocol.INIT {
		node.SetState(protocol.HAND_SHAKE)
		buf, _ = NewVersion(localNode)
	} else if s == protocol.HAND {
		node.SetState(protocol.HAND_SHAKED)
		buf, _ = NewVerack()
	}
	node.Tx(buf)

	return nil
}
