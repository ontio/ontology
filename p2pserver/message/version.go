package message

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	//	"github.com/Ontology/core/ledger"
	"github.com/Ontology/crypto"
	"github.com/Ontology/p2pserver/actor"
	. "github.com/Ontology/p2pserver/protocol"
	"time"
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

func (msg *version) init(n Noder) {
	// Do the init
}

func NewVersion(n Noder, isConsensus bool) ([]byte, error) {
	log.Debug()
	var msg version

	msg.P.Version = n.Version()
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
	//msg.P.StartHeight = 0 //uint64(ledger.DefaultLedger.GetLocalBlockChainHeight())
	height, _ := actor.GetCurrentBlockHeight()
	msg.P.StartHeight = uint64(height)
	if n.GetRelay() {
		msg.P.Relay = 1
	} else {
		msg.P.Relay = 0
	}

	msg.pk = n.GetBookKeeperAddr()
	log.Debug("new version msg.pk is ", msg.pk)
	// TODO the function to wrap below process
	// msg.HDR.init("version", n.GetID(), uint32(len(p.Bytes())))

	msg.Hdr.Magic = NETMAGIC
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
		if msg.P.IsConsensus == false {
			log.Warn("The node handshark with itself")
			node.CloseConn()
			return errors.New("The node handshark with itself")
		}
		if msg.P.IsConsensus == true {
			log.Warn("The node handshark with itself")
			node.CloseConsensusConn()
			return errors.New("The node handshark with itself")
		}
	}

	if msg.P.IsConsensus == true {
		s := node.GetConsensusState()
		if s != INIT && s != HAND {
			log.Warn("Unknow status to received version")
			return errors.New("Unknow status to received version")
		}

		//	n, ok := LocalNode.GetNbrNode(msg.P.Nonce)
		//	if ok == false {
		//		log.Warn("nbr node is not exsit")
		//		return errors.New("nbr node is not exsit")
		//	}

		//	n.SetConsensusConn(node.GetConsensusConn())
		//	n.SetConsensusPort(node.GetConsensusPort())
		//	n.SetConsensusState(node.GetConsensusState())

		node.UpdateInfo(time.Now(), msg.P.Version, msg.P.Services,
			msg.P.Port, msg.P.Nonce, msg.P.Relay, msg.P.StartHeight)
		node.SetConsensusPort(msg.P.ConsensusPort)

		var buf []byte
		if s == INIT {
			node.SetConsensusState(HANDSHAKE)
			buf, _ = NewVersion(localNode, true)
		} else if s == HAND {
			node.SetConsensusState(HANDSHAKED)
			buf, _ = NewVerack(true)
		}
		node.ConsensusTx(buf)
		return nil
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
	if msg.P.Cap[HTTPINFOFLAG] == 0x01 {
		node.SetHttpInfoState(true)
	} else {
		node.SetHttpInfoState(false)
	}
	node.SetHttpInfoPort(msg.P.HttpInfoPort)
	node.SetConsensusPort(msg.P.ConsensusPort)
	node.SetBookKeeperAddr(msg.pk)
	// if  msg.P.Port == msg.P.ConsensusPort don't updateInfo
	node.UpdateInfo(time.Now(), msg.P.Version, msg.P.Services,
		msg.P.Port, msg.P.Nonce, msg.P.Relay, msg.P.StartHeight)
	localNode.AddNbrNode(node)

	var buf []byte
	if s == INIT {
		node.SetState(HANDSHAKE)
		buf, _ = NewVersion(localNode, false)
	} else if s == HAND {
		node.SetState(HANDSHAKED)
		buf, _ = NewVerack(false)
	}
	node.Tx(buf)

	return nil
}
