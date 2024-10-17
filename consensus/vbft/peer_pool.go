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

	"github.com/ontio/ontology/p2pserver/common"
)

type PeerPool struct {
	lock sync.RWMutex

	IDMap     map[string]uint32
	P2pMap    map[uint32]common.PeerId //value: p2p random id
	Connected map[uint32]bool
}

func NewPeerPool(peers map[string]uint32) *PeerPool {
	pool := &PeerPool{
		IDMap:     make(map[string]uint32),
		P2pMap:    make(map[uint32]common.PeerId),
		Connected: make(map[uint32]bool),
	}
	pool.ResetNewConsuensusPeers(peers)
	return pool
}

func (pool *PeerPool) clean() {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	pool.IDMap = make(map[string]uint32)
	pool.P2pMap = make(map[uint32]common.PeerId)
	pool.Connected = make(map[uint32]bool)
}

func (pool *PeerPool) IsPeerConnected(peerIdx uint32) bool {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	return pool.Connected[peerIdx]
}

func (pool *PeerPool) ResetNewConsuensusPeers(peers map[string]uint32) (added []uint32, removed []uint32) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	for id, index := range peers {
		if _, has := pool.IDMap[id]; !has {
			added = append(added, index)
			pool.Connected[index] = false
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

func (pool *PeerPool) OnPeerConnected(peerIdx uint32) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	pool.Connected[peerIdx] = true
}

func (pool *PeerPool) OnPeerDisconnected(peerIdx uint32) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	pool.Connected[peerIdx] = false
}

func (pool *PeerPool) GetPeerIndex(nodeId string) (uint32, bool) {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	idx, present := pool.IDMap[nodeId]
	return idx, present
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
