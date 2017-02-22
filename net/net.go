package net

import (
	"GoOnchain/common"
	"GoOnchain/config"
	"GoOnchain/events"
	"GoOnchain/core/transaction"
	. "GoOnchain/net/message"
	"GoOnchain/net/node"
	. "GoOnchain/net/protocol"
	"fmt"
	"os"
	"time"
)

type Neter interface {
	GetMemoryPool() map[common.Uint256]*transaction.Transaction
	SynchronizeMemoryPool()
	Xmit(common.Inventory) error // The transmit interface
	GetEvent(eventName string) *events.Event
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
		if (r < MAXHELLORETYR) && (node.GetState() != ESTABLISH) {
			r++
			node.SetHandshakeRetry(r)
			fmt.Println("Handshake with %s timeout", node.GetID())
			handshake(n)
		}
	}()

	// TODO Does the timer should be recollected?
	return nil
}

func txBlockHeadersReq(n *Noder) {
	// TODO Need Lock
	node := *n
	if node.GetState() != ESTABLISH {
		fmt.Println("Incorrectly node state to send get Header message")
		return
	}

	buf, _ := NewHeadersReq(node)
	go node.Tx(buf)
}

func StartProtocol() Neter {
	seedNodes, err := config.SeedNodes()
	// TODO alloc the local node, init the nodeMap, EventQueue, TXn pool and idcache
	if err != nil {
		fmt.Println("Access the config file failure")
		os.Exit(1)
		// TODO should we kick off a blind connection in this case
	}

	net := node.InitNode()
	for _, nodeAddr := range seedNodes {
		net.Connect(nodeAddr)
	}
	return net
}
