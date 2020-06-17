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
	"net"
	"strconv"
	"time"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/dht"
	msgpack "github.com/ontio/ontology/p2pserver/message/msg_pack"
	"github.com/ontio/ontology/p2pserver/message/types"
	p2p "github.com/ontio/ontology/p2pserver/net/protocol"
	"github.com/ontio/ontology/p2pserver/peer"
	"github.com/scylladb/go-set/strset"
)

type Discovery struct {
	dht        *dht.DHT
	net        p2p.P2P
	id         common.PeerId
	quit       chan bool
	maskSet    *strset.Set
	maskFilter p2p.AddressFilter //todo : conbine with maskSet
}

func NewDiscovery(net p2p.P2P, maskLst []string, maskFilter p2p.AddressFilter, refleshInterval time.Duration) *Discovery {
	dht := dht.NewDHT(net.GetID())
	if refleshInterval != 0 {
		dht.RtRefreshPeriod = refleshInterval
	}
	return &Discovery{
		id:         net.GetID(),
		dht:        dht,
		net:        net,
		quit:       make(chan bool),
		maskSet:    strset.New(maskLst...),
		maskFilter: maskFilter,
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
	self.dht.Update(info.Id, info.RemoteListenAddress())
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
			for _, curPair := range closer {
				log.Debugf("[dht] find closr peer %s", curPair.ID.ToHexString())

				var msg types.Message
				if curPair.ID.IsPseudoPeerId() {
					msg = msgpack.NewAddrReq()
				} else {
					msg = msgpack.NewFindNodeReq(self.id)
				}
				self.net.SendTo(curPair.ID, msg)
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
				for _, pair := range closer {
					log.Debugf("[dht] find closr peer %s", pair.ID.ToHexString())
					var msg types.Message
					if pair.ID.IsPseudoPeerId() {
						msg = msgpack.NewAddrReq()
					} else {
						msg = msgpack.NewFindNodeReq(randPeer)
					}
					self.net.SendTo(pair.ID, msg)
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

	fresp.TargetID = freq.TargetID
	// search dht
	fresp.CloserPeers = self.dht.BetterPeers(freq.TargetID, dht.AlphaValue)

	//hide mask node if necessary
	remoteAddr, _ := remotePeer.GetAddr16()
	remoteIP := net.IP(remoteAddr[:])

	// mask peer see everyone, but other's will not see mask node
	// if remotePeer is in msk-list, give them everything
	// not in mask set means they are in the other side
	if !self.maskSet.Has(remoteIP.String()) && !self.maskFilter.Contains(remotePeer.Info.RemoteListenAddress()) {
		unmaskedAddrs := make([]common.PeerIDAddressPair, 0)
		// filter out the masked node
		for _, pair := range fresp.CloserPeers {
			ip, _, err := net.SplitHostPort(pair.Address)
			if err != nil {
				continue
			}
			// hide mask node
			if self.maskSet.Has(ip) || self.maskFilter.Contains(pair.Address) {
				continue
			}
			unmaskedAddrs = append(unmaskedAddrs, pair)
		}
		// replace with masked nodes
		fresp.CloserPeers = unmaskedAddrs
	}

	log.Debugf("[dht] find %d more closer peers:", len(fresp.CloserPeers))
	for _, curpa := range fresp.CloserPeers {
		log.Debugf("    dht: pid: %s, addr: %s", curpa.ID.ToHexString(), curpa.Address)
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
		if p2p.GetPeer(curpa.ID) != nil {
			continue
		}
		// do nothing about
		if curpa.ID == p2p.GetID() {
			continue
		}
		log.Debugf("[dht] try to connect to another peer by dht: %s ==> %s", curpa.ID.ToHexString(), curpa.Address)
		go p2p.Connect(curpa.Address)
	}
}

// neighborAddresses get address from dht routing table
func (self *Discovery) neighborAddresses() []common.PeerAddr {
	// e.g. ["127.0.0.1:20338"]
	ipPortAdds := self.dht.RouteTable().ListPeers()
	ret := []common.PeerAddr{}
	for _, curIPPort := range ipPortAdds {
		host, port, err := net.SplitHostPort(curIPPort.Address)
		if err != nil {
			continue
		}

		ipadd := net.ParseIP(host)
		if ipadd == nil {
			continue
		}

		p, err := strconv.Atoi(port)
		if err != nil {
			continue
		}

		curAddr := common.PeerAddr{
			Port: uint16(p),
		}
		copy(curAddr.IpAddr[:], ipadd.To16())

		ret = append(ret, curAddr)
	}

	return ret
}

func (self *Discovery) AddrReqHandle(ctx *p2p.Context) {
	remotePeer := ctx.Sender()

	addrs := self.neighborAddresses()

	// get remote peer IP
	// if get remotePeerAddr failed, do masking anyway
	remoteAddr, _ := remotePeer.GetAddr16()
	remoteIP := net.IP(remoteAddr[:])

	// mask peer see everyone, but other's will not see mask node
	// if remotePeer is in msk-list, give them everthing
	// not in mask set means they are in the other side
	if self.maskSet.Size() > 0 && !self.maskSet.Has(remoteIP.String()) {
		mskedAddrs := make([]common.PeerAddr, 0)
		for _, addr := range addrs {
			ip := net.IP(addr.IpAddr[:])
			address := ip.To16().String()
			// hide mask node
			if self.maskSet.Has(address) {
				continue
			}
			mskedAddrs = append(mskedAddrs, addr)
		}
		// replace with mskedAddrs
		addrs = mskedAddrs
	}

	msg := msgpack.NewAddrs(addrs)
	err := remotePeer.Send(msg)

	if err != nil {
		log.Warn(err)
		return
	}
}

func (self *Discovery) AddrHandle(ctx *p2p.Context, msg *types.Addr) {
	p2p := ctx.Network()
	for _, v := range msg.NodeAddrs {
		if v.Port == 0 || v.ID == p2p.GetID() {
			continue
		}
		ip := net.IP(v.IpAddr[:])
		address := ip.To16().String() + ":" + strconv.Itoa(int(v.Port))

		if self.dht.Contains(v.ID) {
			continue
		}

		log.Debug("[p2p]connect ip address:", address)
		go p2p.Connect(address)
	}
}
