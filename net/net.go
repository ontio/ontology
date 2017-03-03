package net

import (
	"GoOnchain/common"
	"GoOnchain/config"
	"GoOnchain/events"
	"GoOnchain/core/transaction"
	"GoOnchain/net/node"
)

type Neter interface {
	GetMemoryPool() map[common.Uint256]*transaction.Transaction
	SynchronizeMemoryPool()
	Xmit(common.Inventory) error // The transmit interface
	GetEvent(eventName string) *events.Event
}

func StartProtocol() Neter {
	seedNodes := config.Parameters.SeedList

	net := node.InitNode()
	for _, nodeAddr := range seedNodes {
		net.Connect(nodeAddr)
	}
	return net
}
