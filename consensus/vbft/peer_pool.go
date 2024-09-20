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
	"sync"

	"github.com/ontio/ontology-crypto/keypair"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/p2pserver/common"
)

type Peer struct {
	Index            uint32
	PubKey           keypair.PublicKey
	CommittedBlockNo uint32
	connected        bool
}

type PeerPool struct {
	lock sync.RWMutex

	IDMap  map[string]uint32
	P2pMap map[uint32]common.PeerId //value: p2p random id
	peers  map[uint32]*Peer

	peerConnectionWaitings map[uint32]chan struct{}
}

func NewPeerPool(peers map[string]uint32) *PeerPool {
	pool := &PeerPool{
		IDMap:                  make(map[string]uint32),
		P2pMap:                 make(map[uint32]common.PeerId),
		peers:                  make(map[uint32]*Peer),
		peerConnectionWaitings: make(map[uint32]chan struct{}),
	}
	pool.ResetNewConsuensusPeers(peers)
	return pool
}

func (pool *PeerPool) clean() {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	pool.IDMap = make(map[string]uint32)
	pool.P2pMap = make(map[uint32]common.PeerId)
	pool.peers = make(map[uint32]*Peer)
}

func (pool *PeerPool) IsPeerConnected(peerIdx uint32) bool {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	p := pool.peers[peerIdx]
	return p != nil && p.connected
}

func (pool *PeerPool) ResetNewConsuensusPeers(peers map[string]uint32) (added []uint32, removed []uint32) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	for id, index := range peers {
		if _, has := pool.IDMap[id]; !has {
			added = append(added, index)
			peerPK := vconfig.MustPubkey(id)
			pool.peers[index] = &Peer{
				Index:     index,
				PubKey:    peerPK,
				connected: false,
			}
		}
	}

	for id, index := range pool.IDMap {
		if _, has := peers[id]; !has {
			removed = append(removed, index)
		}
	}

	pool.IDMap = peers
	return
}

func (pool *PeerPool) GetConnectedPeerCount() int {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	n := 0
	for _, p := range pool.peers {
		if p.connected {
			n++
		}
	}
	return n
}

func (pool *PeerPool) WaitPeerConnected(peerIdx uint32) {
	if pool.IsPeerConnected(peerIdx) {
		return
	}

	var C chan struct{}
	pool.lock.Lock()
	if _, present := pool.peerConnectionWaitings[peerIdx]; !present {
		C = make(chan struct{})
		pool.peerConnectionWaitings[peerIdx] = C
	} else {
		C = pool.peerConnectionWaitings[peerIdx]
	}
	pool.lock.Unlock()

	<-C
}

func (pool *PeerPool) OnPeerConnected(peerIdx uint32) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	// new peer, rather than modify
	pool.peers[peerIdx] = &Peer{
		Index:     peerIdx,
		PubKey:    pool.peers[peerIdx].PubKey,
		connected: true,
	}
	if C, present := pool.peerConnectionWaitings[peerIdx]; present {
		delete(pool.peerConnectionWaitings, peerIdx)
		close(C)
	}
}

func (pool *PeerPool) OnPeerDisconnected(peerIdx uint32) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	pool.peers[peerIdx] = &Peer{
		Index:     peerIdx,
		PubKey:    pool.peers[peerIdx].PubKey,
		connected: false,
	}
}

func (pool *PeerPool) UpdatePeerCommitBlockNo(peerIdx uint32, commitedBlockNo uint32) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	if C, present := pool.peerConnectionWaitings[peerIdx]; present {
		// wake up peer connection waitings
		delete(pool.peerConnectionWaitings, peerIdx)
		close(C)
	}

	pool.peers[peerIdx] = &Peer{
		Index:            peerIdx,
		PubKey:           pool.peers[peerIdx].PubKey,
		CommittedBlockNo: commitedBlockNo,
		connected:        true,
	}
}

func (pool *PeerPool) GetPeerIndex(nodeId string) (uint32, bool) {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	idx, present := pool.IDMap[nodeId]
	return idx, present
}

func (pool *PeerPool) GetPeerPubKey(peerIdx uint32) keypair.PublicKey {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	if p, present := pool.peers[peerIdx]; present && p != nil {
		return p.PubKey
	}

	return nil
}

func (pool *PeerPool) GetAllPubKeys() map[uint32]keypair.PublicKey {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	keys := make(map[uint32]keypair.PublicKey)
	for idx, peer := range pool.peers {
		keys[idx] = peer.PubKey
	}
	return keys
}

func (pool *PeerPool) getPeer(idx uint32) *Peer {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	peer := pool.peers[idx]
	if peer != nil {
		return peer
	}

	return nil
}

func (pool *PeerPool) AddP2pId(peerIdx uint32, p2pId common.PeerId) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	pool.P2pMap[peerIdx] = p2pId
}

func (pool *PeerPool) GetP2pId(peerIdx uint32) (common.PeerId, bool) {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	p2pid, present := pool.P2pMap[peerIdx]
	return p2pid, present
}
