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

package discovery

import (
	"time"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/dht"
	msgpack "github.com/ontio/ontology/p2pserver/message/msg_pack"
	"github.com/ontio/ontology/p2pserver/message/types"
	p2p "github.com/ontio/ontology/p2pserver/net/protocol"
	"github.com/ontio/ontology/p2pserver/peer"
)

type Discovery struct {
	dht  *dht.DHT
	net  p2p.P2P
	id   common.PeerId
	quit chan bool
}

func NewDiscovery(net p2p.P2P) *Discovery {
	return &Discovery{
		id:   net.GetID(),
		dht:  dht.NewDHT(net.GetID()),
		net:  net,
		quit: make(chan bool),
	}
}

func (self *Discovery) Start() {
	go self.findSelf()
	go self.refreshCPL()
}

func (self *Discovery) Stop() {
	close(self.quit)
}

func (self *Discovery) OnAddPeer(info *peer.PeerInfo) {
	self.dht.Update(info.Id)
}

func (self *Discovery) OnDelPeer(info *peer.PeerInfo) {
	self.dht.Remove(info.Id)
}

func (self *Discovery) findSelf() {
	tick := time.NewTicker(self.dht.RtRefreshPeriod)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			log.Debug("[dht] start to find myself")
			closer := self.dht.BetterPeers(self.id, dht.AlphaValue)
			for _, id := range closer {
				log.Debugf("[dht] find closr peer %x", id)

				var msg types.Message
				if id.IsPseudoPeerId() {
					msg = msgpack.NewAddrReq()
				} else {
					msg = msgpack.NewFindNodeReq(id)
				}
				self.net.Send(self.net.GetPeer(id), msg)
			}
		case <-self.quit:
			return
		}
	}
}

func (self *Discovery) refreshCPL() {
	tick := time.NewTicker(self.dht.RtRefreshPeriod)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			for curCPL := range self.dht.RouteTable().Buckets {
				log.Debugf("[dht] start to refresh bucket: %d", curCPL)
				randPeer := self.dht.RouteTable().GenRandKadId(uint(curCPL))
				closer := self.dht.BetterPeers(randPeer, dht.AlphaValue)
				for _, pid := range closer {
					log.Debugf("[dht] find closr peer %d", pid)
					var msg types.Message
					if pid.IsPseudoPeerId() {
						msg = msgpack.NewAddrReq()
					} else {
						msg = msgpack.NewFindNodeReq(randPeer)
					}
					self.net.Send(self.net.GetPeer(pid), msg)
				}
			}
		case <-self.quit:
			return
		}
	}
}

func (self *Discovery) FindNodeHandle(ctx *p2p.Context, freq *types.FindNodeReq) {
	// we recv message must from establised peer
	remotePeer := ctx.Sender()

	var fresp types.FindNodeResp
	// check the target is my self
	log.Debugf("[dht] find node for peerid: %d", freq.TargetID)
	p2p := ctx.Network()
	if freq.TargetID == self.id {
		fresp.Success = true
		fresp.TargetID = freq.TargetID
		// you've already connected with me so there's no need to give you my address
		// omit the address
		if err := remotePeer.Send(&fresp); err != nil {
			log.Warn(err)
		}
		return
	}
	// search dht
	closer := self.dht.BetterPeers(freq.TargetID, dht.AlphaValue)

	paddrs := p2p.GetPeerStringAddr()
	for _, pid := range closer {
		if addr, ok := paddrs[pid]; ok {
			curAddr := types.PeerAddr{
				Addr:   addr,
				PeerID: pid,
			}
			fresp.CloserPeers = append(fresp.CloserPeers, curAddr)

		}
	}
	fresp.TargetID = freq.TargetID
	log.Debugf("[dht] find %d more closer peers:", len(fresp.CloserPeers))
	for _, curpa := range fresp.CloserPeers {
		log.Debugf("    dht: pid: %d, addr: %s", curpa.PeerID, curpa.Addr)
	}

	if err := remotePeer.Send(&fresp); err != nil {
		log.Warn(err)
	}
}

func (self *Discovery) FindNodeResponseHandle(ctx *p2p.Context, fresp *types.FindNodeResp) {
	if fresp.Success {
		log.Debugf("[p2p dht] %s", "find peer success, do nothing")
		return
	}
	p2p := ctx.Network()
	// we should connect to closer peer to ask them them where should we go
	for _, curpa := range fresp.CloserPeers {
		// already connected
		if p2p.GetPeer(curpa.PeerID) != nil {
			continue
		}
		// do nothing about
		if curpa.PeerID == p2p.GetID() {
			continue
		}
		log.Debugf("[dht] try to connect to another peer by dht: %d ==> %s", curpa.PeerID, curpa.Addr)
		go p2p.Connect(curpa.Addr)
	}
}
