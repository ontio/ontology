package net

import (
	"github.com/DNAProject/DNA/common"
	"github.com/DNAProject/DNA/config"
	"github.com/DNAProject/DNA/core/transaction"
	"github.com/DNAProject/DNA/crypto"
	"github.com/DNAProject/DNA/events"
	"github.com/DNAProject/DNA/net/node"
	"github.com/DNAProject/DNA/net/protocol"
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
