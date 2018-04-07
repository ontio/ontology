package p2pserver

import (
	types "github.com/Ontology/p2pserver/common"
	P2Pnet "github.com/Ontology/p2pserver/net"
	"github.com/Ontology/p2pserver/peer"
)

//P2P represent the net interface of p2p package
type P2P interface {
	Start()
	Halt()
	Connect(addr string)
	GetVersion() uint32
	GetPort() uint16
	GetConsensusPort() uint16
	GetId() uint64
	GetTime() int64
	GetState() uint32
	GetServices() uint64
	GetNeighborAddrs() ([]types.PeerAddr, uint64)
	GetConnectionCnt() uint32
	IsPeerEstablished(p *peer.Peer) bool
	GetMsgCh() chan types.MsgPayload
	Send(p *peer.Peer, data []byte, isConsensus bool) error
}

//NewNetServer return the net object in p2p
func NewNetServer(p *peer.Peer) P2P {
	p.AttachEvent(P2Pnet.DisconnectNotify)
	n := &P2Pnet.NetServer{
		Self:        p,
		ReceiveChan: make(chan types.MsgPayload),
	}
	return n
}
