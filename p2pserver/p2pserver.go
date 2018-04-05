package p2pserver

import (
	"github.com/Ontology/account"
	"github.com/Ontology/common/log"
	types "github.com/Ontology/p2pserver/common"
	"github.com/Ontology/p2pserver/peer"
)

type P2PServer struct {
	Self      *peer.Peer
	network   P2P
	msgRouter *MessageRouter
	//TODO: add infoupdater and syncer
}

func NewServer(acc *account.Account) (*P2PServer, error) {
	self, err := peer.NewPeer(acc.PubKey())
	if err != nil {
		return nil, err
	}
	n := NewNetServer(self)

	p := &P2PServer{
		Self:    self,
		network: n,
	}

	p.msgRouter = NewMsgRouter(p)

	// Fixme: implement the message handler for each msg type
	p.msgRouter.RegisterMsgHandler(types.VERSION_TYPE, VersionHandle)

	return p, nil
}

func (this *P2PServer) GetConnectionCnt() uint32 {
	return this.network.GetConnectionCnt()
}
func (this *P2PServer) Start(isSync bool) error {
	if this != nil {
		this.network.Start()
	}
	return nil
}
func (this *P2PServer) Stop() error {
	this.network.Halt()
	return nil
}
func (this *P2PServer) IsSyncing() bool {
	return false
}
func (this *P2PServer) GetPort() (uint16, uint16) {
	return this.network.GetPort(), this.network.GetConsensusPort()
}
func (this *P2PServer) GetVersion() uint32 {
	return this.network.GetVersion()
}
func (this *P2PServer) GetNeighborAddrs() ([]types.PeerAddr, uint64) {
	return this.network.GetNeighborAddrs()
}
func (this *P2PServer) Xmit(msg interface{}) error {
	return nil
}
func (this *P2PServer) Send(id uint64, buf []byte, isConsensus bool) {
	if this.network.IsPeerEstablished(id) {
		this.network.Send(id, buf, isConsensus)
	}
	log.Errorf("P2PServer send error: peer %x is not established.", id)
}
func (this *P2PServer) GetId() uint64 {
	return this.network.GetId()
}
func (this *P2PServer) GetConnectionState() uint32 {
	return this.network.GetState()
}
func (this *P2PServer) GetTime() int64 {
	return this.network.GetTime()
}
