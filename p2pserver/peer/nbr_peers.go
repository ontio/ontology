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

	comm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/message/types"
)

//NbrPeers: The neigbor list
type NbrPeers struct {
	sync.RWMutex
	List map[uint64]*Peer
}

//Broadcast tranfer msg buffer to all establish peer
func (this *NbrPeers) Broadcast(msg types.Message) {
	sink := comm.NewZeroCopySink(nil)
	types.WriteMessage(sink, msg)

	this.RLock()
	defer this.RUnlock()
	for _, node := range this.List {
		if node.linkState == common.ESTABLISH && node.GetRelay() {
			node.SendRaw(msg.CmdType(), sink.Bytes())
		}
	}
}

//NodeExisted return when peer in nbr list
func (this *NbrPeers) NodeExisted(uid uint64) bool {
	_, ok := this.List[uid]
	return ok
}

//GetPeer return peer according to id
func (this *NbrPeers) GetPeer(id uint64) *Peer {
	this.Lock()
	defer this.Unlock()
	n, exist := this.List[id]
	if !exist {
		return nil
	}
	return n
}

//AddNbrNode add peer to nbr list
func (this *NbrPeers) AddNbrNode(p *Peer) {
	this.Lock()
	defer this.Unlock()

	if this.NodeExisted(p.GetID()) {
		fmt.Printf("[p2p]insert an existed node\n")
	} else {
		this.List[p.GetID()] = p
	}
}

//DelNbrNode delete peer from nbr list
func (this *NbrPeers) DelNbrNode(id uint64) (*Peer, bool) {
	this.Lock()
	defer this.Unlock()

	n, exist := this.List[id]
	if !exist {
		return nil, false
	}
	delete(this.List, id)
	return n, true
}

//initialize nbr list
func (this *NbrPeers) Init() {
	this.List = make(map[uint64]*Peer)
}

//NodeEstablished whether peer established according to id
func (this *NbrPeers) NodeEstablished(id uint64) bool {
	this.RLock()
	defer this.RUnlock()

	n, exist := this.List[id]
	if !exist {
		return false
	}

	return n.linkState == common.ESTABLISH
}

//GetNeighborAddrs return all establish peer address
func (this *NbrPeers) GetNeighborAddrs() []common.PeerAddr {
	this.RLock()
	defer this.RUnlock()

	var addrs []common.PeerAddr
	for _, p := range this.List {
		if p.GetState() != common.ESTABLISH {
			continue
		}
		var addr common.PeerAddr
		addr.IpAddr, _ = p.GetAddr16()
		addr.Time = p.GetTimeStamp()
		addr.Services = p.GetServices()
		addr.Port = p.GetPort()
		addr.ID = p.GetID()
		addrs = append(addrs, addr)
	}

	return addrs
}

//GetNeighborHeights return the id-height map of nbr peers
func (this *NbrPeers) GetNeighborHeights() map[uint64]uint64 {
	this.RLock()
	defer this.RUnlock()

	hm := make(map[uint64]uint64)
	for _, n := range this.List {
		if n.GetState() == common.ESTABLISH {
			hm[n.GetID()] = n.GetHeight()
		}
	}
	return hm
}

//GetNeighborMostHeight return the most height of nbr peers
func (this *NbrPeers) GetNeighborMostHeight() uint64 {
	this.RLock()
	defer this.RUnlock()
	mostHeight := uint64(0)
	for _, n := range this.List {
		if n.GetState() == common.ESTABLISH {
			height := n.GetHeight()
			if mostHeight < height {
				mostHeight = height
			}
		}
	}
	return mostHeight
}

//GetNeighbors return all establish peers in nbr list
func (this *NbrPeers) GetNeighbors() []*Peer {
	this.RLock()
	defer this.RUnlock()
	peers := []*Peer{}
	for _, n := range this.List {
		if n.GetState() == common.ESTABLISH {
			node := n
			peers = append(peers, node)
		}
	}
	return peers
}

//GetNbrNodeCnt return count of establish peers in nbrlist
func (this *NbrPeers) GetNbrNodeCnt() uint32 {
	this.RLock()
	defer this.RUnlock()
	var count uint32
	for _, n := range this.List {
		if n.GetState() == common.ESTABLISH {
			count++
		}
	}
	return count
}
