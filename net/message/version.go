package message

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	"github.com/Ontology/common/serialization"
	"github.com/Ontology/net/actor"
	. "github.com/Ontology/net/protocol"
	"github.com/ontio/ontology-crypto/keypair"
)

const (
	HTTP_INFO_FLAG = 0
)

type version struct {
	Hdr msgHdr
	P   struct {
		Version      uint32
		Services     uint64
		TimeStamp    uint32
		Port         uint16
		HttpInfoPort uint16
		Cap          [32]byte
		Nonce        uint64
		// TODO remove tempory to get serilization function passed
		UserAgent   uint8
		StartHeight uint64
		// FIXME check with the specify relay type length
		Relay       uint8
		IsConsensus bool
	}
	pk keypair.PublicKey
}

func (msg *version) init(n Noder) {
	// Do the init
}

func NewVersion(n Noder) ([]byte, error) {
	log.Debug()
	var msg version

	msg.P.Version = n.Version()
	msg.P.Services = n.Services()
	msg.P.HttpInfoPort = config.Parameters.HttpInfoPort
	if config.Parameters.HttpInfoStart {
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
	// TODO the function to wrap below process
	// msg.HDR.init("version", n.GetID(), uint32(len(p.Bytes())))

	msg.Hdr.Magic = NET_MAGIC
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

/*
 * The node state switch table after rx message, there is time limitation for each action
 * The Handshake status will switch to INIT after TIMEOUT if not received the VerACK
 * in this time window
 *  _______________________________________________________________________
 * |          |    INIT         | HANDSHAKE |  ESTABLISH | INACTIVITY      |
 * |-----------------------------------------------------------------------|
 * | version  | HANDSHAKE(timer)|           |            | HANDSHAKE(timer)|
 * |          | if helloTime > 3| Tx verack | Depend on  | if helloTime > 3|
 * |          | Tx version      |           | node update| Tx version      |
 * |          | then Tx verack  |           |            | then Tx verack  |
 * |-----------------------------------------------------------------------|
 * | verack   |                 | ESTABLISH |            |                 |
 * |          |   No Action     |           | No Action  | No Action       |
 * |------------------------------------------------------------------------
 */
func (msg version) Handle(node Noder) error {
	log.Debug()
	localNode := node.LocalNode()

	// Exclude the node itself
	if msg.P.Nonce == localNode.GetID() {
		log.Warn("The node handshake with itself")
		node.CloseConn()
		return errors.New("The node handshake with itself")
	}

	s := node.GetState()
	if s != INIT && s != HAND {
		log.Warn("Unknow status to received version")
		return errors.New("Unknow status to received version")
	}

	// Obsolete node
	n, ret := localNode.DelNbrNode(msg.P.Nonce)
	if ret == true {
		log.Info(fmt.Sprintf("Node reconnect 0x%x", msg.P.Nonce))
		// Close the connection and release the node soure
		n.SetState(INACTIVITY)
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
	localNode.AddNbrNode(node)

	var buf []byte
	if s == INIT {
		node.SetState(HAND_SHAKE)
		buf, _ = NewVersion(localNode)
	} else if s == HAND {
		node.SetState(HAND_SHAKED)
		buf, _ = NewVerack()
	}
	node.Tx(buf)

	return nil
}
