
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

func (net *MockNetwork) RegisterPeer(newPeer *MockPeer) {
	peerID := newPeer.Local.GetID()
	net.peers[peerID] = newPeer
	for _, peer := range net.peers {
		if peer.Local.GetID() != peerID {
			peer.Connected(peerID)
			newPeer.Connected(peer.Local.GetID())
		}
	}
}

func (net *MockNetwork) Broadcast(fromPeerID uint64, msg types.Message) {
	for _, peer := range net.peers {
		if peer.Local.GetID() != fromPeerID {
			peer.Receive(fromPeerID, msg)
		}
	}
}

func (net *MockNetwork) Send(from, to uint64, msg types.Message) {
	for _, peer := range net.peers {
		if peer.Local.GetID() == to {
			peer.Receive(from, msg)
		}
	}
}
