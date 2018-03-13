package message

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"github.com/Ontology/common/log"
	"github.com/Ontology/common/serialization"
	. "github.com/Ontology/net/protocol"
	"strconv"
)

type verACK struct {
	msgHdr
	isConsensus bool
}

func NewVerack(isConsensus bool) ([]byte, error) {
	var msg verACK
	msg.msgHdr.Magic = NETMAGIC
	copy(msg.msgHdr.CMD[0:7], "verack")
	msg.isConsensus = isConsensus
	tmpBuffer := bytes.NewBuffer([]byte{})
	serialization.WriteBool(tmpBuffer, msg.isConsensus)
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

func (msg verACK) Serialization() ([]byte, error) {
	hdrBuf, err := msg.msgHdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	err = serialization.WriteBool(buf, msg.isConsensus)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err

}

func (msg *verACK) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.msgHdr))
	if err != nil {
		return err
	}

	msg.isConsensus, err = serialization.ReadBool(buf)
	return err
}

/*
 * The node state switch table after rx message, there is time limitation for each action
 * The Hanshake status will switch to INIT after TIMEOUT if not received the VerACK
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
 *
 */
// TODO The process should be adjusted based on above table
func (msg verACK) Handle(node Noder) error {
	log.Debug()

	if msg.isConsensus == true {
		s := node.GetConsensusState()
		if s != HANDSHAKE && s != HANDSHAKED {
			log.Warn("Unknow status to received verack")
			return errors.New("Unknow status to received verack")
		}

		localNode := node.LocalNode()
		n, ok := localNode.GetNbrNode(node.GetID())
		if ok == false {
			log.Warn("nbr node is not exsit")
			return errors.New("nbr node is not exsit")
		}

		node.SetConsensusState(ESTABLISH)
		n.SetConsensusState(node.GetConsensusState())
		n.SetConsensusConn(node.GetConsensusConn())
		//	n.SetConsensusPort(node.GetConsensusPort())
		//	n.SetConsensusState(node.GetConsensusState())

		if s == HANDSHAKE {
			buf, _ := NewVerack(true)
			node.ConsensusTx(buf)
		}
		return nil
	}
	s := node.GetState()
	if s != HANDSHAKE && s != HANDSHAKED {
		log.Warn("Unknow status to received verack")
		return errors.New("Unknow status to received verack")
	}

	node.SetState(ESTABLISH)

	if s == HANDSHAKE {
		buf, _ := NewVerack(false)
		node.Tx(buf)
	}

	node.DumpInfo()
	// Fixme, there is a race condition here,
	// but it doesn't matter to access the invalid
	// node which will trigger a warning
	//TODO JQ: only master p2p port request neighbor list
	node.ReqNeighborList()
	addr := node.GetAddr()
	port := node.GetPort()
	nodeAddr := addr + ":" + strconv.Itoa(int(port))
	//TODO JQ： only master p2p port remove the list
	node.LocalNode().RemoveAddrInConnectingList(nodeAddr)
	//connect consensus port

	if s == HANDSHAKED {
		consensusPort := node.GetConsensusPort()
		nodeConsensusAddr := addr + ":" + strconv.Itoa(int(consensusPort))
		go node.Connect(nodeConsensusAddr, true)
	}
	return nil
}
