package net

import (
	. "github.com/Ontology/common"
	"github.com/Ontology/core/ledger"
	"github.com/Ontology/core/transaction"
	"github.com/Ontology/crypto"
	. "github.com/Ontology/errors"
	"github.com/Ontology/events"
	"github.com/Ontology/net/node"
	"github.com/Ontology/net/protocol"
)

type Neter interface {
	GetTxnPool(byCount bool) (map[Uint256]*transaction.Transaction, Fixed64)
	Xmit(interface{}) error
	GetEvent(eventName string) *events.Event
	GetBookKeepersAddrs() ([]*crypto.PubKey, uint64)
	CleanSubmittedTransactions(block *ledger.Block) error
	GetNeighborNoder() []protocol.Noder
	Tx(buf []byte)
	AppendTxnPool(*transaction.Transaction) ErrCode
}

func StartProtocol(pubKey *crypto.PubKey) protocol.Noder {
	net := node.InitNode(pubKey)
	net.ConnectSeeds()

	return net
}
