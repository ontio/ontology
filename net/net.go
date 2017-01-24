package net

import (
	"time"
	"os"
	"fmt"
	"GoOnchain/common"
	"GoOnchain/config"
	"GoOnchain/events"
	"GoOnchain/core/transaction"
	. "GoOnchain/net/protocol"
	. "GoOnchain/net/message"
	. "GoOnchain/net/node"
)

type neter interface {
	GetMemoryPool() map[common.Uint256]*transaction.Transaction
	SynchronizeMemoryPool()
	Xmit(inv Inventory) error // The transmit interface
}

type net struct {
	ConsensusEvent  *events.Event
	BlockEvent	*events.Event
	// Other Event register
}

func (net *net) init() {
	net.ConsensusEvent = events.NewEvent()
	net.BlockEvent = events.NewEvent()
}

// Trigger handshake
func handshake(n *Noder) error {
	node := *n
	node.SetHandshakeTime(time.Now())
	buf, _ := NewVersion(node)
	go node.Tx(buf)

	timer := time.NewTimer(time.Second * HELLOTIMEOUT)
	go func() {
		<-timer.C
		r := node.GetHandshakeRetry()
		if ((r < MAXHELLORETYR) && (node.GetState() != ESTABLISH)) {
			r++
			node.SetHandshakeRetry(r)
			fmt.Println("Handshake with %s timeout", node.GetID())
			handshake(n)
		}
	} ()

	// TODO Does the timer should be recollected?
	return nil
}

func txBlockHeadersReq(n *Noder) {
	// TODO Need Lock
	node := *n
	if (node.GetState() != ESTABLISH) {
		fmt.Println("Incorrectly node state to send get Header message")
		return
	}

	buf, _ := NewHeadersReq()
	go node.Tx(buf)
}

func txInventory(node *Noder) {
	// TODO get transaction entity TX/Block/Consensus

}

func keepAlive(from *Noder, dst *Noder) {
	// Need move to node function or keep here?
}

// Fixme the Nodes should be a parameter
func updateNodeInfo() {
	ticker := time.NewTicker(time.Second * PERIODUPDATETIME)
	quit := make(chan struct{})

	for {
		select {
		case <- ticker.C:
			common.Trace()
			for _, node := range Nodes.List {
				h1 := node.GetHeight()
				h2 := Nodes.Node.GetHeight()
				if (node.GetState() == ESTABLISH) && (h1 > h2) {
					//buf, _ := newMsg("version")
					buf, _ := NewMsg("getheaders", node)
					//buf, _ := newMsg("getaddr")
					go node.Tx(buf)
				}
			}
		case <- quit:
			ticker.Stop()
			return
		}
	}
	// TODO when to close the timer
	//close(quit)
}

// Fixme the Nodes should be a parameter
func StartProtocol() {
	seedNodes, err := config.SeedNodes()

	if err != nil {
		fmt.Println("Access the config file failure")
		os.Exit(1)
		// TODO should we kick off a blind connection in this case
	}

	for _, nodeAddr := range seedNodes {
		Nodes.Node.Connect(nodeAddr)
	}

	updateNodeInfo()
}
