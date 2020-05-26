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

package netserver

import (
	"net"
	"sync"
	"sync/atomic"

	comm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/message/types"
	"github.com/ontio/ontology/p2pserver/peer"
)

// Conn is a net.Conn wrapper to do some clean up when Close.
type Conn struct {
	net.Conn
	session   uint64
	id        common.PeerId
	closed    bool
	netServer *NetServer
}

// Close overwrite net.Conn
func (self *Conn) Close() error {
	self.netServer.Np.Lock()
	defer self.netServer.Np.Unlock()
	if self.closed {
		return nil
	}

	n := self.netServer.Np.List[self.id]
	if n.Peer == nil {
		self.netServer.logger.Fatalf("connection %s not in net server", self.id.ToHexString())
	} else if n.session == self.session { // connection not replaced
		delete(self.netServer.Np.List, self.id)
		// need handle asynchronously since we hold Np.Lock
		self.netServer.logger.Infof("remove peer %s from net server", self.id.ToHexString())
		go self.netServer.notifyPeerDisconnected(n.Peer.Info)
	}

	self.closed = true
	return self.Conn.Close()
}

type connectedPeer struct {
	session uint64
	Peer    *peer.Peer
}

//NbrPeers: The neigbor list
type NbrPeers struct {
	sync.RWMutex
	List map[common.PeerId]connectedPeer

	nextSessionId uint64
}

func (self *NbrPeers) getSessionId() uint64 {
	return atomic.AddUint64(&self.nextSessionId, 1)
}

func NewNbrPeers() *NbrPeers {
	return &NbrPeers{
		List: make(map[common.PeerId]connectedPeer),
	}
}

//Broadcast tranfer msg buffer to all establish Peer
func (this *NbrPeers) Broadcast(msg types.Message) {
	sink := comm.NewZeroCopySink(nil)
	types.WriteMessage(sink, msg)

	this.RLock()
	defer this.RUnlock()
	for _, node := range this.List {
		if node.Peer.GetRelay() {
			go node.Peer.SendRaw(msg.CmdType(), sink.Bytes())
		}
	}
}

//NodeExisted return when Peer in nbr list
func (this *NbrPeers) NodeExisted(uid common.PeerId) bool {
	_, ok := this.List[uid]
	return ok
}

//GetPeer return Peer according to id
func (this *NbrPeers) GetPeer(id common.PeerId) *peer.Peer {
	this.Lock()
	defer this.Unlock()
	n, exist := this.List[id]
	if !exist {
		return nil
	}
	return n.Peer
}

func (self *NbrPeers) ReplacePeer(p *peer.Peer, net *NetServer) *peer.Peer {
	var result *peer.Peer
	self.Lock()
	defer self.Unlock()

	n := self.List[p.Info.Id]
	result = n.Peer

	conn := &Conn{
		Conn:      p.Link.GetConn(),
		session:   self.getSessionId(),
		id:        p.Info.Id,
		netServer: net,
	}
	p.Link.SetConn(conn)
	self.List[p.Info.Id] = connectedPeer{session: conn.session, Peer: p}

	return result
}

//GetNeighborAddrs return all establish Peer address
func (this *NbrPeers) GetNeighborAddrs() []common.PeerAddr {
	this.RLock()
	defer this.RUnlock()

	var addrs []common.PeerAddr
	for _, node := range this.List {
		p := node.Peer
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
func (this *NbrPeers) GetNeighborHeights() map[common.PeerId]uint64 {
	this.RLock()
	defer this.RUnlock()

	hm := make(map[common.PeerId]uint64)
	for _, p := range this.List {
		n := p.Peer
		hm[n.GetID()] = n.GetHeight()
	}
	return hm
}

//GetNeighborMostHeight return the most height of nbr peers
func (this *NbrPeers) GetNeighborMostHeight() uint64 {
	this.RLock()
	defer this.RUnlock()
	mostHeight := uint64(0)
	for _, p := range this.List {
		n := p.Peer
		height := n.GetHeight()
		if mostHeight < height {
			mostHeight = height
		}
	}
	return mostHeight
}

//GetNeighbors return all establish peers in nbr list
func (this *NbrPeers) GetNeighbors() []*peer.Peer {
	this.RLock()
	defer this.RUnlock()
	var peers []*peer.Peer
	for _, p := range this.List {
		n := p.Peer
		peers = append(peers, n)
	}
	return peers
}

//GetNbrNodeCnt return count of establish peers in nbrlist
func (this *NbrPeers) GetNbrNodeCnt() uint32 {
	this.RLock()
	defer this.RUnlock()
	return uint32(len(this.List))
}
