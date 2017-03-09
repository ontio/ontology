package net

import (
	"GoOnchain/common"
	"GoOnchain/config"
	"GoOnchain/core/transaction"
	"GoOnchain/events"
	"GoOnchain/net/node"
	"GoOnchain/net/protocol"
)

type Neter interface {
	GetMemoryPool() map[common.Uint256]*transaction.Transaction
	SynchronizeMemoryPool()
	Xmit(common.Inventory) error // The transmit interface
	GetEvent(eventName string) *events.Event
}

func StartProtocol() (Neter, protocol.Noder) {
	seedNodes := config.Parameters.SeedList

	net := node.InitNode()
	for _, nodeAddr := range seedNodes {
		net.Connect(nodeAddr)
	}
	return net, net
}
