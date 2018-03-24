package p2pserver

import (
	"github.com/Ontology/account"
	types "github.com/Ontology/p2pserver/common"
	"github.com/Ontology/p2pserver/peer"
)

type P2PServer interface {
	Start(bool, bool) error
	Stop() error
	GetVersion() uint32
	GetConnectionCnt() uint
	GetPort() (uint16, uint16)
	GetState() uint32
	GetId() uint64
	Services() uint64
	GetConnectionState() uint32
	GetTime() int64
	GetNeighborAddrs() ([]types.PeerAddr, uint64)
	Xmit(interface{}) error
	IsSyncing() bool
	IsStarted() bool
	EnableDual(bool) error
}

func NewServer(acc *account.Account) (P2PServer, error) {
	server := peer.NewPeer(acc.PubKey())
	err := server.Start(true, true)
	return server, err
}
