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
package mock

import (
	"fmt"
	"testing"
	"time"

	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/p2pserver/common"
	msgTypes "github.com/ontio/ontology/p2pserver/message/types"
	"github.com/ontio/ontology/p2pserver/net/netserver"
	p2p "github.com/ontio/ontology/p2pserver/net/protocol"
	"github.com/ontio/ontology/p2pserver/peer"
	"github.com/ontio/ontology/p2pserver/protocols/bootstrap"
	"github.com/ontio/ontology/p2pserver/protocols/discovery"
	"github.com/ontio/ontology/p2pserver/protocols/subnet"
	"github.com/ontio/ontology/p2pserver/protocols/utils"
	"github.com/stretchr/testify/assert"
)

func TestSubnetAllGovAreSeed(t *testing.T) {
	subnet.RefreshDuration = time.Millisecond * 1000
	log.Info("test subnet all gov are seed")
	//topo
	/**
		normal —————— normal
		  \           /
		   \         /
		    \      /
	          gov
	*/
	SG, N := 4, 2
	T := N + SG
	acct := make([]*account.Account, 0, N)
	for i := 0; i < T; i++ {
		acct = append(acct, account.NewAccount(""))
	}
	var gov []string
	var seedList []string
	for i := 0; i < SG; i++ {
		gov = append(gov, vconfig.PubkeyID(acct[i].PubKey()))
		seedList = append(seedList, fmt.Sprintf("127.0.0.%d:%d", i, i))
	}

	net := NewNetwork()
	var nodes []*netserver.NetServer
	for i := 0; i < SG; i++ {
		seedNode := NewSubnetNode(acct[0], seedList[i], seedList, gov, net, nil, "seedgov")
		go seedNode.Start()
		nodes = append(nodes, seedNode)
	}

	for i := SG; i < T; i++ {
		node := NewSubnetNode(acct[i], fmt.Sprintf("127.0.0.%d:%d", i, i), seedList, gov, net, nil, "norm")
		go node.Start()
		nodes = append(nodes, node)
	}

	for i := 0; i < T; i++ {
		for j := i; j < T; j++ {
			net.AllowConnect(nodes[i].GetHostInfo().Id, nodes[j].GetHostInfo().Id)
		}
	}

	//need some time for seed node detected it's identity
	time.Sleep(time.Second * 20)
	for i := 0; i < SG; i++ {
		assert.Equal(t, len(getSubnetMemberInfo(nodes[i].Protocol())), SG, i)
	}
	for i := 0; i < T; i++ {
		assert.Equal(t, uint32(T)-1, nodes[i].GetConnectionCnt(), i)
	}
}

func TestSubnet(t *testing.T) {
	subnet.RefreshDuration = time.Millisecond * 500
	log.Info("test subnet start")
	//topo
	/**
	normal —————— normal
	  \    gov     /
	   \    |     /
	    \  seed  /
	        |
	        |
	       gov
	*/
	S, G, N := 2, 4, 2
	T := N + S + G
	acct := make([]*account.Account, 0, N)
	for i := 0; i < T; i++ {
		acct = append(acct, account.NewAccount(""))
	}
	var gov []string
	for i := S; i < S+G; i++ {
		gov = append(gov, vconfig.PubkeyID(acct[i].PubKey()))
	}

	var seedList []string
	for i := 0; i < S; i++ {
		seedList = append(seedList, fmt.Sprintf("127.0.0.%d:%d", i, i))
	}
	net := NewNetwork()
	var nodes []*netserver.NetServer
	for i := 0; i < S; i++ {
		seedNode := NewSubnetNode(acct[0], seedList[i], seedList, gov, net, nil, "seed")
		go seedNode.Start()
		nodes = append(nodes, seedNode)
	}

	for i := S; i < T; i++ {
		prefix := "norm"
		if i < S+G {
			prefix = "gov"
		}
		node := NewSubnetNode(acct[i], fmt.Sprintf("127.0.0.%d:%d", i, i), seedList, gov, net, nil, prefix)
		go node.Start()
		nodes = append(nodes, node)
	}

	for i := 0; i < T; i++ {
		for j := i; j < T; j++ {
			net.AllowConnect(nodes[i].GetHostInfo().Id, nodes[j].GetHostInfo().Id)
		}
	}

	time.Sleep(time.Second * 10)
	for i := 0; i < S+G; i++ {
		assert.Equal(t, len(getSubnetMemberInfo(nodes[i].Protocol())), G, i)
	}
	for i := 0; i < T; i++ {
		if i < S {
			assert.Equal(t, uint32(T)-1, nodes[i].GetConnectionCnt(), i)
		} else if i < S+G {
			assert.Equal(t, uint32(S+G-1), nodes[i].GetConnectionCnt(), i)
		} else {
			assert.Equal(t, uint32(S+N-1), nodes[i].GetConnectionCnt(), i)
		}
	}
}

func getSubnetMemberInfo(protocol p2p.Protocol) []common.SubnetMemberInfo {
	handler, ok := protocol.(*TestSubnetProtocalHandler)
	if !ok {
		return nil
	}

	return handler.subnet.GetMembersInfo()
}

func NewSubnetNode(acct *account.Account, listenAddr string, seeds, govs []string, net Network, reservedPeers []string,
	logPrefix string) *netserver.NetServer {
	seedId := common.RandPeerKeyId()
	info := peer.NewPeerInfo(seedId.Id, 0, 0, true, 0,
		0, 0, "v2.0.0", "")
	context := fmt.Sprintf("peer %s-%s: ", logPrefix, seedId.Id.ToHexString()[:6])
	logger := common.LoggerWithContext(common.NewGlobalLoggerWrapper(), context)
	protocal := NewTestSubnetProtocalHandler(acct, seeds, govs, logger)
	resvFilter := protocal.GetReservedAddrFilter(len(reservedPeers) != 0)
	return NewNode(seedId, listenAddr, info, protocal, net, reservedPeers, resvFilter, logger)
}

type TestSubnetProtocalHandler struct {
	seeds     *utils.HostsResolver
	discovery *discovery.Discovery
	bootstrap *bootstrap.BootstrapService
	subnet    *subnet.SubNet
	acct      *account.Account // nil if conenesus is not enabled
}

func NewTestSubnetProtocalHandler(acct *account.Account, seedList, govs []string, logger common.Logger) *TestSubnetProtocalHandler {
	gov := utils.NewGovNodeMockResolver(govs)
	seeds, invalid := utils.NewHostsResolver(seedList)
	if invalid != nil {
		panic(fmt.Errorf("invalid seed list； %v", invalid))
	}
	subNet := subnet.NewSubNet(acct, seeds, gov, logger)
	return &TestSubnetProtocalHandler{seeds: seeds, subnet: subNet, acct: acct}
}

func (self *TestSubnetProtocalHandler) GetReservedAddrFilter(staticFilterEnabled bool) p2p.AddressFilter {
	return self.subnet.GetReservedAddrFilter(staticFilterEnabled)
}

func (self *TestSubnetProtocalHandler) GetMaskAddrFilter() p2p.AddressFilter {
	return self.subnet.GetMaskAddrFilter()
}

func (self *TestSubnetProtocalHandler) start(net p2p.P2P) {
	maskFilter := self.subnet.GetMaskAddrFilter()
	self.discovery = discovery.NewDiscovery(net, config.DefConfig.P2PNode.ReservedCfg.MaskPeers,
		maskFilter, time.Millisecond*1000)
	self.bootstrap = bootstrap.NewBootstrapService(net, self.seeds)
	go self.discovery.Start()
	go self.bootstrap.Start()
	go self.subnet.Start(net)
}

func (self *TestSubnetProtocalHandler) stop() {
	self.discovery.Stop()
	self.bootstrap.Stop()
	self.subnet.Stop()
}

func (self *TestSubnetProtocalHandler) HandleSystemMessage(net p2p.P2P, msg p2p.SystemMessage) {
	switch m := msg.(type) {
	case p2p.NetworkStart:
		self.start(net)
	case p2p.PeerConnected:
		self.discovery.OnAddPeer(m.Info)
		self.bootstrap.OnAddPeer(m.Info)
		self.subnet.OnAddPeer(net, m.Info)
	case p2p.PeerDisConnected:
		self.discovery.OnDelPeer(m.Info)
		self.bootstrap.OnDelPeer(m.Info)
		self.subnet.OnDelPeer(m.Info)
	case p2p.NetworkStop:
		self.stop()
	case p2p.HostAddrDetected:
		self.subnet.OnHostAddrDetected(m.ListenAddr)
	}
}

func (self *TestSubnetProtocalHandler) HandlePeerMessage(ctx *p2p.Context, msg msgTypes.Message) {
	log.Trace("[p2p]receive message", ctx.Sender().GetAddr(), ctx.Sender().GetID())
	switch m := msg.(type) {
	case *msgTypes.AddrReq:
		self.discovery.AddrReqHandle(ctx)
	case *msgTypes.FindNodeResp:
		self.discovery.FindNodeResponseHandle(ctx, m)
	case *msgTypes.FindNodeReq:
		self.discovery.FindNodeHandle(ctx, m)
	case *msgTypes.Addr:
		self.discovery.AddrHandle(ctx, m)
	case *msgTypes.SubnetMembersRequest:
		self.subnet.OnMembersRequest(ctx, m)
	case *msgTypes.SubnetMembers:
		self.subnet.OnMembersResponse(ctx, m)
	case *msgTypes.NotFound:
		log.Debug("[p2p]receive notFound message, hash is ", m.Hash)
	default:
		msgType := msg.CmdType()
		log.Warn("unknown message handler for the msg: ", msgType)
	}
}
