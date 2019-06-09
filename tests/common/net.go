/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package TestCommon

import (
	"github.com/ontio/ontology/p2pserver/message/types"
	"github.com/ontio/ontology/common/log"
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
	if len(net.peers) < 2 {
		log.Errorf("less than two peers in network")
	}
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
