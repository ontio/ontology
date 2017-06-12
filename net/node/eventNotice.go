package node

import (
	"DNA/events"
	"fmt"
)

type eventQueue struct {
	Consensus  *events.Event
	Block      *events.Event
	Disconnect *events.Event
}

func (eq *eventQueue) init() {
	eq.Consensus = events.NewEvent()
	eq.Block = events.NewEvent()
	eq.Disconnect = events.NewEvent()
}

func (eq *eventQueue) GetEvent(eventName string) *events.Event {
	switch eventName {
	case "consensus":
		return eq.Consensus
	case "block":
		return eq.Block
	case "disconnect":
		return eq.Disconnect
	default:
		fmt.Printf("Unknow event registe")
		return nil
	}
}
