package message

import (
	"DNA/common"
	"DNA/common/log"
	"DNA/core/ledger"
	. "DNA/net/protocol"
	"encoding/hex"
	"errors"
	"time"
)

type verACK struct {
	msgHdr
	// No payload
}

func NewVerack() ([]byte, error) {
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
	log.Info("The message tx verack length is %d, %s\n", len(buf), str)

	return buf, err
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
	common.Trace()

	t := time.Now()
	s := node.GetState()
	if s != HANDSHAKE && s != HANDSHAKED {
		log.Warn("Unknow status to received verack")
		return errors.New("Unknow status to received verack")
	}

	node.SetState(ESTABLISH)
	if (s == HANDSHAKE) {
		buf, _ := NewVerack()
		node.Tx(buf)
	}

	// TODO update other node info
	node.UpdateTime(t)
	node.DumpInfo()
	// Fixme, there is a race condition here,
	// but it doesn't matter to access the invalid
	// node which will trigger a warning
	node.ReqNeighborList()

	// FIXME compact to a seperate function
	if uint64(ledger.DefaultLedger.Blockchain.BlockHeight) < node.GetHeight() {
		buf, err := NewHeadersReq(node)
		if err != nil {
			log.Error("failed build a new headersReq")
		} else {
			node.Tx(buf)
		}
	}
	return nil
}
