package message

import (
	"GoOnchain/common"
	"GoOnchain/core/ledger"
	"GoOnchain/events"
	. "GoOnchain/net/protocol"
	"fmt"
)

type blockReq struct {
	msgHdr
	//TBD
}

type block struct {
	msgHdr
	blk ledger.Block
	// TBD
	event *events.Event
}

func (msg block) Handle(node Noder) error {
	common.Trace()

	fmt.Printf("RX block message\n")
	if !node.ExistedID(msg.blk.Hash()) {
		// TODO Update the currently ledger
		// FIXME the relative event should be attached to the message
		if msg.event != nil {
			msg.event.Notify(events.EventSaveBlock, msg.blk)
		}
	}
	return nil
}

func reqBlkData(node Noder, hash common.Uint256) error {
	var msg dataReq
	msg.dataType = common.BLOCK
	// TODO handle the hash array case
	msg.hash = hash

	buf, _ := msg.Serialization()
	go node.Tx(buf)

	return nil
}
