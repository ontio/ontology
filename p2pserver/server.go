package p2pserver

import (
	types "github.com/Ontology/p2pserver/common"
	P2Pnet "github.com/Ontology/p2pserver/net"
	"github.com/Ontology/p2pserver/peer"
	"net"
)

type P2P interface {
	Halt()
	InitNonTLSListen() (net.Listener, net.Listener, error)
	InitTLSListen() (net.Listener, error)
	NonTLSDial(addr string) (net.Conn, error)
	TLSDial(addr string) (net.Conn, error)
	Connectseeds()
	GetVersion() uint32
	GetPort() uint16
	GetConsensusPort() uint16
	GetId() uint64
	GetTime() int64
	GetState() uint32
	GetServices() uint64
	GetNeighborAddrs() ([]types.PeerAddr, uint64)
	GetConnectionCnt() uint32
	IsPeerEstablished(uint64) bool
	GetMsgCh() chan types.MsgPayload
	Xmit(msg interface{}) error
	Tx(id uint64, data []byte)
}

func NewNetServer(p *peer.Peer) P2P {
	p.AttachEvent(P2Pnet.DisconnectNotify)
	n := &P2Pnet.NetServer{
		Self:        p,
		ReceiveChan: make(chan types.MsgPayload),
	}
	return n
}
