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
	"time"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology-crypto/keypair"
)

type Peer struct {
	Index          uint32
	PubKey         *crypto.PubKey
	handShake      *peerHandshakeMsg
	LatestInfo     *peerHeartbeatMsg // latest heartbeat msg
	LastUpdateTime time.Time         // time received heartbeat from peer
	connected      bool
}

type PeerPool struct {
	lock    sync.RWMutex
	maxSize int

	server  *Server
	configs map[uint32]*vconfig.PeerConfig // peer index to peer
	IDMap   map[vconfig.NodeID]uint32

	peers map[uint32]*Peer
}

func NewPeerPool(maxSize int, server *Server) *PeerPool {
	return &PeerPool{
		maxSize: maxSize,
		server:  server,
		configs: make(map[uint32]*vconfig.PeerConfig),
		IDMap:   make(map[vconfig.NodeID]uint32),
		peers:   make(map[uint32]*Peer),
	}
}

func (pool *PeerPool) isNewPeer(peerIdx uint32) bool {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	if _, present := pool.peers[peerIdx]; present {
		return !pool.peers[peerIdx].connected
	}

	return true
}

func (pool *PeerPool) addPeer(config *vconfig.PeerConfig) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	pool.configs[config.Index] = config
	pool.IDMap[config.ID] = config.Index
	pool.peers[config.Index] = &Peer{
		Index:          config.Index,
		LastUpdateTime: time.Unix(0, 0),
		connected:      false,
	}
	return nil
}

func (pool *PeerPool) getActivePeerCount() int {
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

func (pool *PeerPool) peerConnected(peerIdx uint32) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	// new peer, rather than modify
	pool.peers[peerIdx] = &Peer{
		Index:     peerIdx,
		connected: true,
	}
	return nil
}

func (pool *PeerPool) peerDisconnected(peerIdx uint32) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	var lastUpdateTime time.Time
	if p, present := pool.peers[peerIdx]; present {
		lastUpdateTime = p.LastUpdateTime
	}

	pool.peers[peerIdx] = &Peer{
		Index:          peerIdx,
		LastUpdateTime: lastUpdateTime,
		connected:      false,
	}
	return nil
}

func (pool *PeerPool) peerHandshake(peerIdx uint32, msg *peerHandshakeMsg) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	pool.peers[peerIdx] = &Peer{
		Index:          peerIdx,
		handShake:      msg,
		LatestInfo:     pool.peers[peerIdx].LatestInfo,
		LastUpdateTime: time.Now(),
		connected:      true,
	}

	return nil
}

func (pool *PeerPool) peerHeartbeat(peerIdx uint32, msg *peerHeartbeatMsg) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	pool.peers[peerIdx] = &Peer{
		Index:          peerIdx,
		handShake:      pool.peers[peerIdx].handShake,
		LatestInfo:     msg,
		LastUpdateTime: time.Now(),
		connected:      true,
	}

	return nil
}

func (pool *PeerPool) getNeighbours() []*Peer {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	peers := make([]*Peer, 0)
	for _, p := range pool.peers {
		if p.connected {
			peers = append(peers, p)
		}
	}
	return peers
}

func (pool *PeerPool) GetPeerIndex(nodeId vconfig.NodeID) (uint32, bool) {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	idx, present := pool.IDMap[nodeId]
	return idx, present
}

func (pool *PeerPool) isPeerAlive(peerIdx uint32) bool {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	p := pool.peers[peerIdx]
	if p == nil || !p.connected {
		return false
	}
	if time.Now().Sub(p.LastUpdateTime) > peerHandshakeTimeout*2 {
		if p.LastUpdateTime.Unix() > 0 {
			log.Errorf("server %d: peer %d seems disconnected, %v, %v", pool.server.Index, time.Now(), p.LastUpdateTime)
		}
		return false
	}
	return true
}

func (pool *PeerPool) getPeer(idx uint32) *Peer {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	peer := pool.peers[idx]
	if peer != nil {
		peerPK, _ := pool.configs[idx].ID.Pubkey()
		peer.PubKey = peerPK
		return peer
	}

	return nil
}
