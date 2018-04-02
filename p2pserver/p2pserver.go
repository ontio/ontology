package p2pserver

import (
	"strings"
	"errors"

	"github.com/Ontology/account"
	"github.com/Ontology/common/config"
	"github.com/Ontology/p2pserver/peer"
	msg "github.com/Ontology/p2pserver/message"
)

type P2PServer struct {
	isRunning bool
	syncing   bool
	self      *peer.Peer
	nbrPeers  peer.NbrPeers
}

func (p2p *P2PServer) Start(bool) error {
	if p2p != nil {
		p2p.connectseeds()
	}
	return errors.New("p2p server invalid")
}
func (p2p *P2PServer) Stop() error {
	return nil
}
func (p2p *P2PServer) GetVersion() uint32 {
	if p2p.self != nil {
		return p2p.self.Version()
	}
	return 0
}
func (p2p *P2PServer) GetConnectionCnt() uint32 {
	if p2p.self != nil {
		return p2p.nbrPeers.GetNbrNodeCnt()
	}
	return 0
}
func (p2p *P2PServer) GetPort() (uint16, uint16) {
	return 0, 0
}
func (p2p *P2PServer) GetState() uint32 {
	if p2p.self != nil {
		return p2p.self.GetState()
	}
	return 0
}
func (p2p *P2PServer) GetId() uint64 {
	if p2p.self != nil {
		return p2p.self.GetID()
	}
	return 0
}
func (p2p *P2PServer) Services() uint64 {
	if p2p.self != nil {
		return p2p.self.Services()
	}
	return 0
}
func (p2p *P2PServer) GetConnectionState() uint32 {
	return 0
}
func (p2p *P2PServer) GetTime() int64 {
	if p2p.self != nil {
		return p2p.self.GetTime()
	}
	return 0
}
func (p2p *P2PServer) GetNeighborAddrs() ([]msg.PeerAddr, uint64) {
	return p2p.nbrPeers.GetNeighborAddrs()
}
func (p2p *P2PServer) Xmit(interface{}) error {
	//TODO
	return nil
}
func (p2p *P2PServer) IsSyncing() bool {
	return p2p.syncing
}
func (p2p *P2PServer) IsStarted() bool {
	return p2p.isRunning
}

func NewServer(acc *account.Account) (*P2PServer, error) {
	p, err := peer.NewPeer(acc.PubKey())
	if err != nil {
		return nil, err
	}
	server := &P2PServer{
		isRunning: false,
		syncing:   false,
		self:      p,
	}
	return server, nil
}

func (p2p *P2PServer) IsUptoMinNodeCount() bool {
	consensusType := strings.ToLower(config.Parameters.ConsensusType)
	if consensusType == "" {
		consensusType = "dbft"
	}
	minCount := config.DBFTMINNODENUM
	switch consensusType {
	case "dbft":
	case "solo":
		minCount = config.SOLOMINNODENUM
	}
	return int(p2p.GetConnectionCnt())+1 >= minCount
}
func (p2p *P2PServer) connectseeds() {

}
