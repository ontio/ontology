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
package bootstrap

import (
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/ontio/ontology/p2pserver/common"
	msgpack "github.com/ontio/ontology/p2pserver/message/msg_pack"
	"github.com/ontio/ontology/p2pserver/message/types"
	p2p "github.com/ontio/ontology/p2pserver/net/protocol"
	"github.com/ontio/ontology/p2pserver/peer"
	"github.com/ontio/ontology/p2pserver/protocols/utils"
)

const activeConnect = 4 // when connection num less than this value, we connect seeds node actively.

type BootstrapService struct {
	seeds     *utils.HostsResolver
	connected uint32
	net       p2p.P2P
	quit      chan bool
}

func NewBootstrapService(net p2p.P2P, seeds *utils.HostsResolver) *BootstrapService {
	return &BootstrapService{
		seeds: seeds,
		net:   net,
		quit:  make(chan bool),
	}
}

func (self *BootstrapService) Start() {
	go self.connectSeedService()
}

func (self *BootstrapService) Stop() {
	close(self.quit)
}

func (self *BootstrapService) OnAddPeer(info *peer.PeerInfo) {
	atomic.AddUint32(&self.connected, 1)
}

func (self *BootstrapService) OnDelPeer(info *peer.PeerInfo) {
	atomic.AddUint32(&self.connected, ^uint32(0))
}

//connectSeedService make sure seed peer be connected
func (self *BootstrapService) connectSeedService() {
	t := time.NewTimer(0) // let it timeout to start connect immediately
	for {
		select {
		case <-t.C:
			self.connectSeeds()
			t.Stop()
			connected := atomic.LoadUint32(&self.connected)
			if connected >= activeConnect {
				t.Reset(time.Second * time.Duration(10*common.CONN_MONITOR))
			} else {
				t.Reset(time.Second * common.CONN_MONITOR)
			}
		case <-self.quit:
			t.Stop()
			return
		}
	}
}

//connectSeeds connect the seeds in seedlist and call for nbr list
func (self *BootstrapService) connectSeeds() {
	connPeers := make(map[string]*peer.Peer)
	nps := self.net.GetNeighbors()
	for _, tn := range nps {
		listenAddr := tn.Info.RemoteListenAddress()
		connPeers[listenAddr] = tn
	}

	seedConnList := make([]*peer.Peer, 0)
	seedDisconn := make([]string, 0)
	isSeed := false
	for _, nodeAddr := range self.seeds.GetHostAddrs() {
		if p, ok := connPeers[nodeAddr]; ok {
			seedConnList = append(seedConnList, p)
		} else {
			seedDisconn = append(seedDisconn, nodeAddr)
		}

		if self.net.IsOwnAddress(nodeAddr) {
			isSeed = true
		}
	}

	if len(seedConnList) > 0 {
		rand.Seed(time.Now().UnixNano())
		// close NewAddrReq
		index := rand.Intn(len(seedConnList))
		self.reqNbrList(seedConnList[index])
		if isSeed && len(seedDisconn) > 0 {
			index := rand.Intn(len(seedDisconn))
			go self.net.Connect(seedDisconn[index])
		}
	} else { //not found
		for _, nodeAddr := range self.seeds.GetHostAddrs() {
			go self.net.Connect(nodeAddr)
		}
	}
}

func (this *BootstrapService) reqNbrList(p *peer.Peer) {
	id := p.GetID()
	var msg types.Message
	if id.IsPseudoPeerId() {
		msg = msgpack.NewAddrReq()
	} else {
		msg = msgpack.NewFindNodeReq(this.net.GetID())
	}

	go this.net.SendTo(id, msg)
}
