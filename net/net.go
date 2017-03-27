package net

import (
	"DNA/common"
	"DNA/config"
	"DNA/core/transaction"
	"DNA/crypto"
	"DNA/events"
	"DNA/net/node"
	"DNA/net/protocol"
)

type Neter interface {
	GetTxnPool(cleanPool bool) map[common.Uint256]*transaction.Transaction
	SynchronizeTxnPool()
	Xmit(common.Inventory) error // The transmit interface
	GetEvent(eventName string) *events.Event
	GetMinersAddrs() ([]*crypto.PubKey, uint64)
}

func StartProtocol(pubKey *crypto.PubKey) (Neter, protocol.Noder) {
	seedNodes := config.Parameters.SeedList

	net := node.InitNode(pubKey)
	for _, nodeAddr := range seedNodes {
		go net.Connect(nodeAddr)
	}
	return net, net
}
