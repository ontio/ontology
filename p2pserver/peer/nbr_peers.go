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
	types "github.com/Ontology/p2pserver/common"
	"sync"
)

//NbrPeers: The neigbor list
type NbrPeers struct {
	sync.RWMutex
	List map[uint64]*Peer
}

func (nm *NbrPeers) Broadcast(buf []byte) {
	nm.RLock()
	defer nm.RUnlock()
	for _, node := range nm.List {
		if node.state == types.ESTABLISH && node.GetRelay() == true {
			node.Send(buf)
		}
	}
}

func (nm *NbrPeers) NodeExisted(uid uint64) bool {
	_, ok := nm.List[uid]
	return ok
}

func (nm *NbrPeers) AddNbrNode(p *Peer) {
	nm.Lock()
	defer nm.Unlock()

	if nm.NodeExisted(p.GetID()) {
		fmt.Printf("Insert a existed node\n")
	} else {
		nm.List[p.GetID()] = p
	}
}

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

func (nm *NbrPeers) GetConnectionCnt() uint {
	nm.RLock()
	defer nm.RUnlock()

	var cnt uint
	for _, node := range nm.List {
		if node.state == types.ESTABLISH {
			cnt++
		}
	}
	return cnt
}

func (nm *NbrPeers) init() {
	nm.List = make(map[uint64]*Peer)
}

func (nm *NbrPeers) NodeEstablished(id uint64) bool {
	nm.RLock()
	defer nm.RUnlock()

	n, ok := nm.List[id]
	if ok == false {
		return false
	}

	if n.state != types.ESTABLISH {
		return false
	}

	return true
}

func (nm *NbrPeers) GetNeighborAddrs() ([]types.PeerAddr, uint64) {
	nm.RLock()
	defer nm.RUnlock()

	var i uint64
	var addrs []types.PeerAddr
	for _, p := range nm.List {
		if p.GetState() != types.ESTABLISH {
			continue
		}
		var addr types.PeerAddr
		addr.IpAddr, _ = p.GetAddr16()
		addr.Time = p.GetTime()
		addr.Services = p.Services()
		addr.Port = p.GetPort()
		addr.ID = p.GetID()
		addrs = append(addrs, addr)

		i++
	}

	return addrs, i
}

func (nm *NbrPeers) GetNeighborHeights() map[uint64]uint64 {
	nm.RLock()
	defer nm.RUnlock()

	hm := make(map[uint64]uint64)
	for _, n := range nm.List {
		if n.GetState() == types.ESTABLISH {
			hm[n.id] = n.height
		}
	}
	return hm
}

func (nm *NbrPeers) GetNeighbors() []*Peer {
	nm.RLock()
	defer nm.RUnlock()
	peers := []*Peer{}
	for _, n := range nm.List {
		if n.GetState() == types.ESTABLISH {
			node := n
			peers = append(peers, node)
		}
	}
	return peers
}

func (nm *NbrPeers) GetNbrNodeCnt() uint32 {
	nm.RLock()
	defer nm.RUnlock()
	var count uint32
	for _, n := range nm.List {
		if n.GetState() == types.ESTABLISH {
			count++
		}
	}
	return count
}
