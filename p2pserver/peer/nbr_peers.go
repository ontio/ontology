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

package peer

import (
	"fmt"
	"sync"

	"github.com/ontio/ontology/p2pserver/common"
)

//NbrPeers: The neigbor list
type NbrPeers struct {
	sync.RWMutex
	List map[uint64]*Peer
}

//Broadcast tranfer msg buffer to all establish peer
func (nm *NbrPeers) Broadcast(buf []byte, isConsensus bool) {
	nm.RLock()
	defer nm.RUnlock()
	for _, node := range nm.List {
		if node.syncState == common.ESTABLISH && node.GetRelay() == true {
			node.Send(buf, isConsensus)
		}
	}
}

//NodeExisted return when peer in nbr list
func (nm *NbrPeers) NodeExisted(uid uint64) bool {
	_, ok := nm.List[uid]
	return ok
}

//GetPeer return peer according to id
func (nm *NbrPeers) GetPeer(id uint64) *Peer {
	nm.Lock()
	defer nm.Unlock()
	n, ok := nm.List[id]
	if ok == false {
		return nil
	}
	return n
}

//AddNbrNode add peer to nbr list
func (nm *NbrPeers) AddNbrNode(p *Peer) {
	nm.Lock()
	defer nm.Unlock()

	if nm.NodeExisted(p.GetID()) {
		fmt.Printf("Insert a existed node\n")
	} else {
		nm.List[p.GetID()] = p
	}
}

//DelNbrNode delete peer from nbr list
func (nm *NbrPeers) DelNbrNode(id uint64) (*Peer, bool) {
	nm.Lock()
	defer nm.Unlock()

	n, ok := nm.List[id]
	if ok == false {
		return nil, false
	}
	delete(nm.List, id)
	return n, true
}

//initialize nbr list
func (nm *NbrPeers) Init() {
	nm.List = make(map[uint64]*Peer)
}

//NodeEstablished whether peer established according to id
func (nm *NbrPeers) NodeEstablished(id uint64) bool {
	nm.RLock()
	defer nm.RUnlock()

	n, ok := nm.List[id]
	if ok == false {
		return false
	}

	if n.syncState != common.ESTABLISH {
		return false
	}

	return true
}

//GetNeighborAddrs return all establish peer address
func (nm *NbrPeers) GetNeighborAddrs() ([]common.PeerAddr, uint64) {
	nm.RLock()
	defer nm.RUnlock()

	var i uint64
	var addrs []common.PeerAddr
	for _, p := range nm.List {
		if p.GetSyncState() != common.ESTABLISH {
			continue
		}
		var addr common.PeerAddr
		addr.IpAddr, _ = p.GetAddr16()
		addr.Time = p.GetTimeStamp()
		addr.Services = p.GetServices()
		addr.Port = p.GetSyncPort()
		addr.ID = p.GetID()
		addrs = append(addrs, addr)

		i++
	}

	return addrs, i
}

//GetNeighborHeights return the id-height map of nbr peers
func (nm *NbrPeers) GetNeighborHeights() map[uint64]uint64 {
	nm.RLock()
	defer nm.RUnlock()

	hm := make(map[uint64]uint64)
	for _, n := range nm.List {
		if n.GetSyncState() == common.ESTABLISH {
			hm[n.GetID()] = n.GetHeight()
		}
	}
	return hm
}

//GetNeighbors return all establish peers in nbr list
func (nm *NbrPeers) GetNeighbors() []*Peer {
	nm.RLock()
	defer nm.RUnlock()
	peers := []*Peer{}
	for _, n := range nm.List {
		if n.GetSyncState() == common.ESTABLISH {
			node := n
			peers = append(peers, node)
		}
	}
	return peers
}

//GetNbrNodeCnt return count of establish peers in nbrlist
func (nm *NbrPeers) GetNbrNodeCnt() uint32 {
	nm.RLock()
	defer nm.RUnlock()
	var count uint32
	for _, n := range nm.List {
		if n.GetSyncState() == common.ESTABLISH {
			count++
		}
	}
	return count
}
