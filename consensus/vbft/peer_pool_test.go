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

package vbft

import (
	"testing"
	"time"

	"github.com/ontio/ontology/common"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
)

func constructPeerPool(connect bool) *PeerPool {
	peer := &Peer{
		Index:          uint32(1),
		LastUpdateTime: time.Unix(0, 0),
		connected:      connect,
	}
	peers := make(map[uint32]*Peer)
	peers[1] = peer
	peerpool := &PeerPool{
		maxSize: int(3),
		configs: make(map[uint32]*vconfig.PeerConfig),
		IDMap:   make(map[vconfig.NodeID]uint32),
		peers:   peers,
	}
	return peerpool
}
func TestIsNewPeer(t *testing.T) {
	peerpool := constructPeerPool(false)
	isnew := peerpool.isNewPeer(uint32(2))
	t.Logf("TestIsNewPeer: %v\n", isnew)
}

func TestAddPeer(t *testing.T) {
	nodeId, _ := vconfig.StringID("206520e7475798520164487f7e4586bb55790097ceb786aab6d5bc889d12991a5a204c6298bef1bf43c20680a3979a213392b99c97042ebae27d2a7af6442aa7c008")
	peerconfig := &vconfig.PeerConfig{
		Index: uint32(1),
		ID:    nodeId,
	}
	peerpool := constructPeerPool(false)
	res := peerpool.addPeer(peerconfig)
	t.Logf("TestAddPeer : %v", res)
}

func TestGetActivePeerCount(t *testing.T) {
	peerpool := constructPeerPool(true)
	count := peerpool.getActivePeerCount()
	t.Logf("TestGetActivePeerCount count:%v", count)
}

func TestPeerConnected(t *testing.T) {
	peerpool := constructPeerPool(false)
	err := peerpool.peerConnected(uint32(1))
	t.Logf("TestPeerConnected :%v", err)
}

func TestPeerDisconnected(t *testing.T) {
	peerpool := constructPeerPool(true)
	err := peerpool.peerDisconnected(uint32(1))
	t.Logf("TestPeerDisconnected :%v", err)
}

func TestPeerHandshake(t *testing.T) {
	nodeId, _ := vconfig.StringID("206520e7475798520164487f7e4586bb55790097ceb786aab6d5bc889d12991a5a204c6298bef1bf43c20680a3979a213392b99c97042ebae27d2a7af6442aa7c008")
	peerconfig := &vconfig.PeerConfig{
		Index: uint32(1),
		ID:    nodeId,
	}
	peerpool := constructPeerPool(false)
	peerpool.addPeer(peerconfig)
	handshakemsg := &peerHandshakeMsg{
		CommittedBlockNumber: uint64(2),
		CommittedBlockHash:   common.Uint256{},
		CommittedBlockLeader: uint32(1),
	}
	err := peerpool.peerHandshake(uint32(1), handshakemsg)
	t.Logf("TestPeerHandshake :%v", err)
}

func TestPeerHeartbeat(t *testing.T) {
	nodeId, _ := vconfig.StringID("206520e7475798520164487f7e4586bb55790097ceb786aab6d5bc889d12991a5a204c6298bef1bf43c20680a3979a213392b99c97042ebae27d2a7af6442aa7c008")
	peerconfig := &vconfig.PeerConfig{
		Index: uint32(1),
		ID:    nodeId,
	}
	peerpool := constructPeerPool(false)
	peerpool.addPeer(peerconfig)
	heartbeatmsg := &peerHeartbeatMsg{
		CommittedBlockNumber: uint64(2),
		CommittedBlockHash:   common.Uint256{},
		CommittedBlockLeader: uint32(1),
		ChainConfigView:      uint32(1),
	}
	err := peerpool.peerHeartbeat(uint32(1), heartbeatmsg)
	t.Logf("TestPeerHeartbeat: %v", err)
}

func TestGetNeighbours(t *testing.T) {
	peerpool := constructPeerPool(true)
	peers := peerpool.getNeighbours()
	t.Logf("TestGetNeighbours: %d", len(peers))
}

func TestGetPeerIndex(t *testing.T) {
	nodeId, _ := vconfig.StringID("12020298fe9f22e9df64f6bfcc1c2a14418846cffdbbf510d261bbc3fa6d47073df9a2")
	peerconfig := &vconfig.PeerConfig{
		Index: uint32(1),
		ID:    nodeId,
	}
	peerpool := constructPeerPool(false)
	peerpool.addPeer(peerconfig)
	idx, present := peerpool.GetPeerIndex(nodeId)
	if !present {
		t.Errorf("TestGetPeerIndex is not exist: %d", idx)
		return
	}
	t.Logf("TestGetPeerIndex: %d,%v", idx, present)
}

func TestGetPeer(t *testing.T) {
	nodeId, _ := vconfig.StringID("12020298fe9f22e9df64f6bfcc1c2a14418846cffdbbf510d261bbc3fa6d47073df9a2")
	peerconfig := &vconfig.PeerConfig{
		Index: uint32(1),
		ID:    nodeId,
	}
	peerpool := constructPeerPool(false)
	peerpool.addPeer(peerconfig)
	peer := peerpool.getPeer(uint32(1))
	if peer == nil {
		t.Errorf("TestGetPeer failed peer is nil")
		return
	}
	t.Logf("TestGetPeer: %v", peer.Index)
}
