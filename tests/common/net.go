
package TestCommon

import (
	"github.com/ontio/ontology/p2pserver/message/types"
)

var MockNet *MockNetwork

type MockNetwork struct {
	peers map[uint64]*MockPeer
}

func init() {
	MockNet = &MockNetwork{
		peers: make(map[uint64]*MockPeer),
	}
}

func (net *MockNetwork) GetPeer(id uint64) *MockPeer {
	return net.peers[id]
}

func (net *MockNetwork) RegisterPeer(peer *MockPeer) {
	peerID := peer.Local.GetID()
	net.peers[peerID] = peer
	for _, peer := range net.peers {
		if peer.Local.GetID() != peerID {
			peer.Connected(peerID)
		}
	}
}

func (net *MockNetwork) Broadcast(peerID uint64, msg types.Message) {
	for _, peer := range net.peers {
		if peer.Local.GetID() != peerID {
			peer.Receive(peerID, msg)
		}
	}
}
