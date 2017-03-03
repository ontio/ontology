package node

import (
	"GoOnchain/common"
	"GoOnchain/events"
	"fmt"
)

type eventQueue struct {
	Consensus *events.Event
	Block     *events.Event
}

func (eq *eventQueue) init() {
	eq.Consensus = events.NewEvent()
	eq.Block = events.NewEvent()
}

func (eq eventQueue) SubscribeMsgQueue(common.InventoryType) {
	//TODO
}

func (eq *eventQueue) GetEvent(eventName string) *events.Event {
	switch eventName {
	case "consensus":
		return eq.Consensus
	case "block":
		return eq.Block
	default:
		fmt.Printf("Unknow event registe")
		return nil
	}
}
