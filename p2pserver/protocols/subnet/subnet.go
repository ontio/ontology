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

package subnet

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/ontio/ontology/account"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/message/types"
	p2p "github.com/ontio/ontology/p2pserver/net/protocol"
	"github.com/ontio/ontology/p2pserver/peer"
	"github.com/ontio/ontology/p2pserver/protocols/utils"
)

const MaxMemberRequests = 3
const MaxInactiveTime = 10 * time.Minute

var RefreshDuration = 1 * time.Minute

type SubNet struct {
	acct     *account.Account // nil if conenesus is not enabled
	seeds    *utils.HostsResolver
	gov      utils.GovNodeResolver
	unparker *utils.Parker
	selfAddr string

	seedNode bool
	closed   bool

	lock      sync.RWMutex
	connected map[string]*peer.PeerInfo // connected seed or gov node, listen address --> PeerInfo
	members   map[string]*MemberStatus  // gov node info, listen address --> pubkey hex string
	logger    common.Logger
}

func NewSubNet(acc *account.Account, seeds *utils.HostsResolver,
	gov utils.GovNodeResolver, logger common.Logger) *SubNet {
	return &SubNet{
		acct:     acc,
		seeds:    seeds,
		gov:      gov,
		unparker: utils.NewParker(),
		logger:   logger,

		connected: make(map[string]*peer.PeerInfo),
		members:   make(map[string]*MemberStatus),
	}
}

type MemberStatus struct {
	PubKey string
	Alive  time.Time
}

func (self *SubNet) Start(net p2p.P2P) {
	go self.maintainLoop(net)
}

func (self *SubNet) Stop() {
	self.closed = true
	self.unparker.Unpark()
}

func (self *SubNet) OnAddPeer(net p2p.P2P, info *peer.PeerInfo) {
	self.lock.Lock()
	defer self.lock.Unlock()
	listenAddr := info.RemoteListenAddress()
	member := self.members[listenAddr]
	if self.isSeed(listenAddr) || member != nil {
		self.connected[listenAddr] = info
		self.sendMembersRequest(net, info.Id)
	}
	if member != nil {
		member.Alive = time.Now()
	}
}

func (self *SubNet) OnDelPeer(info *peer.PeerInfo) {
	self.lock.Lock()
	defer self.lock.Unlock()
	listenAddr := info.RemoteListenAddress()
	member := self.members[listenAddr]
	if self.isSeed(listenAddr) || member != nil {
		delete(self.connected, listenAddr)
	}
	if member != nil {
		member.Alive = time.Now()
	}
}
func (self *SubNet) IpInMembers(ip string) bool {
	self.lock.RLock()
	defer self.lock.RUnlock()
	for addr := range self.members {
		if strings.HasPrefix(addr, ip+":") {
			return true
		}
	}

	return false
}

func (self *SubNet) isSeedIp(ip string) bool {
	hosts := self.seeds.GetHostAddrs()
	for _, host := range hosts {
		if strings.HasPrefix(host, ip+":") {
			return true
		}
	}

	return false
}

func (self *SubNet) isSeed(addr string) bool {
	hosts := self.seeds.GetHostAddrs()
	for _, host := range hosts {
		if host == addr {
			return true
		}
	}

	return false
}

func (self *SubNet) OnHostAddrDetected(listenAddr string) {
	self.selfAddr = listenAddr
	self.seedNode = self.isSeed(listenAddr)
}

func (self *SubNet) checkAuthority(listenAddr string, msg *types.SubnetMembersRequest) bool {
	if msg.FromSeed() {
		return self.isSeed(listenAddr)
	}

	return self.gov.IsGovNode(msg.PubKey)
}

func (self *SubNet) OnMembersRequest(ctx *p2p.Context, msg *types.SubnetMembersRequest) {
	sender := ctx.Sender()

	peerAddr := sender.Info.RemoteListenAddress()
	if !self.checkAuthority(peerAddr, msg) {
		self.logger.Info("[subnet] check authority for members request failed, peer: %s", peerAddr)
		return
	}

	members := make([]types.MemberInfo, 0, len(self.members))
	self.lock.Lock()

	for addr, status := range self.members {
		members = append(members, types.MemberInfo{PubKey: status.PubKey, Addr: addr})
	}

	//update self.members
	if self.gov.IsGovNode(msg.PubKey) && self.members[peerAddr] == nil {
		self.members[peerAddr] = &MemberStatus{
			PubKey: vconfig.PubkeyID(msg.PubKey),
			Alive:  time.Now(),
		}
	}
	self.lock.Unlock()

	reply := &types.SubnetMembers{Members: members}
	self.logger.Debugf("[subnet], send members to peer %s, value: %s", sender.Info.Id.ToHexString(), reply)
	ctx.Network().SendTo(sender.GetID(), reply)
}

func (self *SubNet) OnMembersResponse(ctx *p2p.Context, msg *types.SubnetMembers) {
	self.lock.Lock()
	defer self.lock.Unlock()
	self.logger.Debugf("[subnet], receive members: %s ", msg.String())

	listen := ctx.Sender().Info.RemoteListenAddress()
	if self.connected[listen] == nil {
		self.logger.Info("[subnet] receive members response from unkown node: %s", listen)
		return
	}

	for _, info := range msg.Members {
		if self.members[info.Addr] == nil {
			self.members[info.Addr] = &MemberStatus{
				PubKey: info.PubKey,
				Alive:  time.Now(),
			}
		}
	}

	self.unparker.Unpark()
}

func (self *SubNet) getUnconnectedGovNode() []string {
	self.lock.Lock()
	defer self.lock.Unlock()

	var addrs []string
	for addr := range self.members {
		if self.connected[addr] == nil {
			addrs = append(addrs, addr)
		}
	}

	return addrs
}

func (self *SubNet) newMembersRequest() *types.SubnetMembersRequest {
	var request *types.SubnetMembersRequest
	if self.seedNode {
		request = types.NewMembersRequestFromSeed()
	} else if self.acct != nil && self.gov.IsGovNode(self.acct.PublicKey) {
		var err error
		request, err = types.NewMembersRequest(self.acct)
		if err != nil {
			return nil
		}
	}

	return request
}

func (self *SubNet) sendMembersRequestToRandNodes(net p2p.P2P) {
	request := self.newMembersRequest()
	if request == nil {
		return
	}

	count := 0
	// note map iteration is randomized
	for _, peer := range self.connected {
		net.SendTo(peer.Id, request)
		count += 1
		if count == MaxMemberRequests {
			break
		}
	}
}

func (self *SubNet) sendMembersRequest(net p2p.P2P, peer common.PeerId) {
	request := self.newMembersRequest()
	if request == nil {
		return
	}

	net.SendTo(peer, request)
}

func (self *SubNet) cleanStaleGovNode() {
	now := time.Now()
	self.lock.Lock()
	defer self.lock.Unlock()

	for addr, member := range self.members {
		if member.Alive.Add(MaxInactiveTime).Before(now) {
			delete(self.members, addr)
		}
	}
}

func (self *SubNet) maintainLoop(net p2p.P2P) {
	parker := self.unparker
	for {
		self.lock.Lock()
		for _, peer := range net.GetNeighbors() {
			listen := peer.Info.RemoteListenAddress()
			if self.members[listen] != nil && self.connected[listen] == nil {
				self.connected[listen] = peer.Info
				self.members[listen].Alive = time.Now()
			}
		}
		self.lock.Unlock()

		for _, addr := range self.getUnconnectedGovNode() {
			self.logger.Infof("[subnet] try connect gov node: %s", addr)
			go net.Connect(addr)
		}

		self.cleanStaleGovNode()
		self.sendMembersRequestToRandNodes(net)

		if self.seedNode || (self.acct != nil && self.gov.IsGovNode(self.acct.PublicKey)) {
			members := self.GetMembersInfo()
			buf, _ := json.Marshal(members)
			self.logger.Infof("[subnet] current members: %s", string(buf))
		}

		parker.ParkTimeout(RefreshDuration)
		if self.closed {
			return
		}
	}
}

func (self *SubNet) GetReservedAddrFilter() p2p.AddressFilter {
	return &SubNetReservedAddrFilter{
		subnet: self,
	}
}

func (self *SubNet) GetMaskAddrFilter() p2p.AddressFilter {
	return &SubNetMaskAddrFilter{
		subnet: self,
	}
}

//restful api
func (self *SubNet) GetMembersInfo() []common.SubnetMemberInfo {
	self.lock.RLock()
	defer self.lock.RUnlock()

	var members []common.SubnetMemberInfo
	for addr, mem := range self.members {
		members = append(members, common.SubnetMemberInfo{
			PubKey:     mem.PubKey,
			ListenAddr: addr,
			Connected:  self.connected[addr] != nil || self.selfAddr == addr,
		})
	}

	return members
}
