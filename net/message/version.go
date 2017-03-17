package message

import (
	"GoOnchain/common"
	"GoOnchain/common/log"
	. "GoOnchain/net/protocol"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

type version struct {
	Hdr msgHdr
	P   struct {
		Version   uint32
		Services  uint64
		TimeStamp uint32
		Port      uint16
		Nonce     uint64
		// TODO remove tempory to get serilization function passed
		UserAgent   uint8
		StartHeight uint64
		// FIXME check with the specify relay type length
		Relay uint8
	}
}

func (msg *version) init(n Noder) {
	// Do the init
}

func NewVersion(n Noder) ([]byte, error) {
	common.Trace()
	var msg version

	msg.P.Version = n.Version()
	msg.P.Services = n.Services()
	// FIXME Time overflow
	msg.P.TimeStamp = uint32(time.Now().UTC().UnixNano())
	msg.P.Port = n.GetPort()
	msg.P.Nonce = n.GetID()
	msg.P.UserAgent = 0x00
	// Fixme Get the block height from ledger
	msg.P.StartHeight = uint64(n.GetLedger().GetLocalBlockChainHeight())
	if n.GetRelay() {
		msg.P.Relay = 1
	} else {
		msg.P.Relay = 0
	}

	// TODO the function to wrap below process
	// msg.HDR.init("version", n.GetID(), uint32(len(p.Bytes())))

	msg.Hdr.Magic = NETMAGIC
	copy(msg.Hdr.CMD[0:7], "version")
	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, &(msg.P))
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
	var buf bytes.Buffer

	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *version) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, msg)
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
	common.Trace()
	localNode := node.LocalNode()

	// Exclude the node itself
	if msg.P.Nonce == localNode.GetID() {
		log.Warn("The node handshark with itself")
		return errors.New("The node handshark with itself")
	}

	s := node.GetState()
	if s != INIT {
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

	node.SetState(HANDSHAKE)
	node.UpdateInfo(time.Now(), msg.P.Version, msg.P.Services,
		msg.P.Port, msg.P.Nonce, msg.P.Relay, msg.P.StartHeight)
	localNode.AddNbrNode(node)

	buf, _ := NewVersion(localNode)
	node.Tx(buf)

	time.Sleep(2 * time.Second)
	buf, _ = NewVerack()
	node.Tx(buf)

	return nil
}
