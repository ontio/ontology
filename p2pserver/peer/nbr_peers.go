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
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"sync"

	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/message/types"
)

//NbrPeers: The neigbor list
type NbrPeers struct {
	sync.RWMutex
	List map[uint64]*Peer
}

//Broadcast tranfer msg buffer to all establish peer
func (this *NbrPeers) Broadcast(msg types.Message, isConsensus bool) {
	this.RLock()
	defer this.RUnlock()
	for _, node := range this.List {
		if node.GetSyncState(node.GetTransportType()) == common.ESTABLISH && node.GetRelay() == true {
			log.Tracef("Send msg %s to node %s, msgType=%t", msg.CmdType(), node.GetAddr(node.GetTransportType()), isConsensus)
			node.Send(msg, isConsensus, node.GetTransportType())
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
	n, ok := this.List[id]
	if ok == false {
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

	n, ok := this.List[id]
	if ok == false {
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
func (this *NbrPeers) NodeEstablished(id uint64, tspType byte) bool {
	this.RLock()
	defer this.RUnlock()

	n, ok := this.List[id]
	if ok == false {
		return false
	}

	if n.GetSyncState(tspType) != common.ESTABLISH {
		return false
	}

	return true
}

//GetNeighborAddrs return all establish peer address
func (this *NbrPeers) GetNeighborAddrs() []common.PeerAddr {
	this.RLock()
	defer this.RUnlock()

	var addrs []common.PeerAddr
	tspType := config.DefConfig.P2PNode.TransportType
	for _, p := range this.List {
		if p.GetTransportType() != tspType {
			continue
		}
		if p.GetSyncState(tspType) != common.ESTABLISH {
			continue
		}
		var addr common.PeerAddr
		addr.IpAddr, _ = p.GetAddr16(tspType)
		addr.Time = p.GetTimeStamp(tspType)
		addr.Services = p.GetServices()
		addr.Port = p.GetSyncPort(tspType)
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
	tspType := config.DefConfig.P2PNode.TransportType
	for _, n := range this.List {
		if n.GetSyncState(tspType) == common.ESTABLISH {
			hm[n.GetID()] = n.GetHeight()
		}
	}
	return hm
}

//GetNeighbors return all establish peers in nbr list
func (this *NbrPeers) GetNeighbors() []*Peer {
	this.RLock()
	defer this.RUnlock()
	peers := []*Peer{}
	for _, n := range this.List {
		if n.GetSyncState(common.LegacyTSPType) == common.ESTABLISH {
			node := n
			peers = append(peers, node)
		}
		if n.GetSyncState(config.DefConfig.P2PNode.TransportType) == common.ESTABLISH {
			node := n
			peers = append(peers, node)
		}
	}
	return peers
}

//GetNbrNodeCnt return count of establish peers in nbrlist
func (this *NbrPeers) GetNbrNodeCnt() (uint32, uint32) {
	this.RLock()
	defer this.RUnlock()
	var countLegacy uint32
    var count uint32
	for _, n := range this.List {
		if n.GetSyncState(common.LegacyTSPType) == common.ESTABLISH {
			countLegacy++
		}
		if n.GetSyncState(config.DefConfig.P2PNode.TransportType) == common.ESTABLISH {
			count++
		}
	}
	return countLegacy, count
}
