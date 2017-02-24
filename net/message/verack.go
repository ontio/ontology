package message

import (
	"GoOnchain/common"
	. "GoOnchain/net/protocol"
	"encoding/hex"
	"fmt"
	"time"
)

type verACK struct {
	msgHdr
	// No payload
}

func newVerack() ([]byte, error) {
	var msg verACK
	// Fixme the check is the []byte{0} instead of 0
	var sum []byte
	sum = []byte{0x5d, 0xf6, 0xe0, 0xe2}
	msg.msgHdr.init("verack", sum, 0)

	buf, err := msg.Serialization()
	if err != nil {
		return nil, err
	}

	str := hex.EncodeToString(buf)
	fmt.Printf("The message tx verack length is %d, %s", len(buf), str)

	return buf, err
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
func (msg verACK) Handle(node Noder) error {
	common.Trace()

	t := time.Now()
	// TODO we loading the state&time without consider race case
	th := node.GetHandshakeTime()
	s := node.GetState()

	m, _ := msg.Serialization()
	str := hex.EncodeToString(m)
	fmt.Printf("The message rx verack length is %d, %s\n", len(m), str)

	// TODO take care about the time duration overflow
	tDelta := t.Sub(th)
	if tDelta.Seconds() < HELLOTIMEOUT {
		if s == HANDSHAKEING {
			node.SetState(ESTABLISH)
			buf, _ := newVerack()
			go node.Tx(buf)
		} else if s == HANDSHAKED {
			node.SetState(ESTABLISH)
		}
	}
	// TODO update other node info
	node.UpdateTime(t)
	node.DumpInfo()
	if node.GetState() == ESTABLISH {
		node.ReqNeighborList()
	}

	return nil
}
