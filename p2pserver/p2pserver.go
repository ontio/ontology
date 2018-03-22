package p2pserver

import (
	"github.com/Ontology/account"
	"github.com/Ontology/p2pserver/peer"
	"github.com/Ontology/p2pserver/protocol"
)

type P2pServer interface {
	Start(bool, bool) error
	Stop() error
	GetVersion() uint32
	GetConnectionCnt() uint64
	GetPort() (uint16, uint16)
	GetState() uint32
	GetId() uint64
	Services() uint64
	GetConnectionState() uint32
	GetTime() int64
	GetNeighborAddrs() ([]protocol.PeerAddr, uint64)
	Xmit(interface{}) error
	IsSyncing() bool
	IsStarted() bool
	EnableDual(bool) error
}

func NewServer(acc *account.Account) (P2pServer, error) {
	server := peer.NewPeer(acc.PubKey())
	err := server.Start(true, true)
	return server, err
}
