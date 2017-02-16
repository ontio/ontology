package message

import (
	"GoOnchain/common"
	. "GoOnchain/net/protocol"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"
	"unsafe"
)

type version struct {
	Hdr msgHdr
	P   struct {
		Version   uint32
		Services  uint64
		TimeStamp uint32
		Port      uint16
		Nonce     uint32
		// TODO remove tempory to get serilization function passed
		UserAgent   uint8
		StartHeight uint32
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
	msg.P.Nonce = n.GetNonce()
	fmt.Printf("The nonce is 0x%x", msg.P.Nonce)
	msg.P.UserAgent = 0x00
	// Fixme Get the block height from ledger
	msg.P.StartHeight = 1
	if n.GetRelay() {
		msg.P.Relay = 1
	} else {
		msg.P.Relay = 0
	}

	msg.Hdr.Magic = NETMAGIC
	ver := "version"
	copy(msg.Hdr.CMD[0:7], ver)
	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, &(msg.P))
	if err != nil {
		fmt.Println("Binary Write failed at new Msg")
		return nil, err
	}
	s := sha256.Sum256(p.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(msg.Hdr.Checksum))
	msg.Hdr.Length = uint32(len(p.Bytes()))
	fmt.Printf("The message payload length is %d\n", msg.Hdr.Length)

	m, err := msg.Serialization()
	if err != nil {
		fmt.Println("Error Convert net message ", err.Error())
		return nil, err
	}

	str := hex.EncodeToString(m)
	fmt.Printf("The message length is %d, %s\n", len(m), str)
	return m, nil
}

func (msg version) Verify(buf []byte) error {
	err := msg.Hdr.Verify(buf)
	// TODO verify the message Content
	return err
}

func (msg version) Serialization() ([]byte, error) {
	var buf bytes.Buffer

	fmt.Printf("The size of messge is %d in serialization\n",
		uint32(unsafe.Sizeof(msg)))
	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *version) Deserialization(p []byte) error {
	fmt.Printf("The size of messge is %d in deserialization\n",
		uint32(unsafe.Sizeof(*msg)))

	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, msg)
	return err
}

/*
 * The node state switch table after rx message, there is time limitation for each action
 * The Hanshark status will switch to INIT after TIMEOUT if not received the VerACK
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
 * The node state switch table after TX message, there is time limitation for each action
 *  ____________________________________________________________
 * |          |    INIT   | HANDSHAKE  | ESTABLISH | INACTIVITY |
 * |------------------------------------------------------------|
 * | version  |           |  INIT      | None      |            |
 * |          | Update    |  Update    |           | Update     |
 * |          | helloTime |  helloTime |           | helloTime  |
 * |------------------------------------------------------------|
 */
// TODO The process should be adjusted based on above table
func (msg version) Handle(node Noder) error {
	common.Trace()
	t := time.Now()
	// TODO check version compatible or not
	s := node.GetState()
	if s == HANDSHAKEING {
		node.SetState(HANDSHAKED)
		buf, _ := newVerack()
		fmt.Println("TX verack")
		go node.Tx(buf)
	} else if s != ESTABLISH {
		node.SetHandshakeTime(t)
		node.SetState(HANDSHAKEING)
		buf, _ := NewVersion(node.LocalNode())
		go node.Tx(buf)
	}

	// TODO Update other node information
	fmt.Printf("Node %s state is %d", node.GetID(), node.GetState())
	node.UpdateTime(t)

	return nil
}
