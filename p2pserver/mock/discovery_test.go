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

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/message/types"
	msgTypes "github.com/ontio/ontology/p2pserver/message/types"
	"github.com/ontio/ontology/p2pserver/net/netserver"
	p2p "github.com/ontio/ontology/p2pserver/net/protocol"
	"github.com/ontio/ontology/p2pserver/peer"
	"github.com/ontio/ontology/p2pserver/protocols/bootstrap"
	"github.com/ontio/ontology/p2pserver/protocols/discovery"
	"github.com/ontio/ontology/p2pserver/protocols/utils"
	"github.com/stretchr/testify/assert"
)

func init() {
	common.Difficulty = 1
}

type DiscoveryProtocol struct {
	MaskPeers       []string
	RefleshInterval time.Duration
	seeds           []string

	discovery *discovery.Discovery
	bootstrap *bootstrap.BootstrapService
}

func NewDiscoveryProtocol(seeds []string, maskPeers []string) *DiscoveryProtocol {
	return &DiscoveryProtocol{seeds: seeds, MaskPeers: maskPeers}
}

func (self *DiscoveryProtocol) start(net p2p.P2P) {
	self.discovery = discovery.NewDiscovery(net, self.MaskPeers, p2p.NoneAddrFilter(), self.RefleshInterval)
	seeds, invalid := utils.NewHostsResolver(self.seeds)
	if len(invalid) != 0 {
		panic(fmt.Errorf("invalid seed list； %v", invalid))
	}
	self.bootstrap = bootstrap.NewBootstrapService(net, seeds)
	go self.discovery.Start()
	go self.bootstrap.Start()
}

func (self *DiscoveryProtocol) HandleSystemMessage(net p2p.P2P, msg p2p.SystemMessage) {
	switch m := msg.(type) {
	case p2p.NetworkStart:
		self.start(net)
	case p2p.PeerConnected:
		self.discovery.OnAddPeer(m.Info)
		self.bootstrap.OnAddPeer(m.Info)
	case p2p.PeerDisConnected:
		self.discovery.OnDelPeer(m.Info)
		self.bootstrap.OnDelPeer(m.Info)
	case p2p.NetworkStop:
		self.discovery.Stop()
		self.bootstrap.Stop()
	}
}

func (self *DiscoveryProtocol) HandlePeerMessage(ctx *p2p.Context, msg msgTypes.Message) {
	log.Trace("[p2p]receive message", ctx.Sender().GetAddr(), ctx.Sender().GetID())
	switch m := msg.(type) {
	case *types.AddrReq:
		self.discovery.AddrReqHandle(ctx)
	case *msgTypes.FindNodeResp:
		self.discovery.FindNodeResponseHandle(ctx, m)
	case *msgTypes.FindNodeReq:
		self.discovery.FindNodeHandle(ctx, m)
	default:
		msgType := msg.CmdType()
		log.Warn("unknown message handler for the msg: ", msgType)
	}
}

func TestDiscoveryNode(t *testing.T) {
	N := 5
	net := NewNetwork()
	seedNode := NewDiscoveryNode(nil, net)
	var nodes []*netserver.NetServer
	go seedNode.Start()
	seedAddr := seedNode.GetHostInfo().Addr
	log.Errorf("seed addr: %s", seedAddr)
	for i := 0; i < N; i++ {
		node := NewDiscoveryNode([]string{seedAddr}, net)
		net.AllowConnect(seedNode.GetHostInfo().Id, node.GetHostInfo().Id)
		go node.Start()
		nodes = append(nodes, node)
	}

	time.Sleep(time.Second * 1)
	assert.Equal(t, seedNode.GetConnectionCnt(), uint32(N))
	for i, node := range nodes {
		assert.Equal(t, node.GetConnectionCnt(), uint32(1), fmt.Sprintf("node %d", i))
	}

	log.Info("start allow node connection")
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			net.AllowConnect(nodes[i].GetHostInfo().Id, nodes[j].GetHostInfo().Id)
		}
	}
	time.Sleep(time.Second * 1)
	for i, node := range nodes {
		assert.True(t, node.GetConnectionCnt() > uint32(N/3), fmt.Sprintf("node %d", i))
	}
}

func NewDiscoveryNode(seeds []string, net Network) *netserver.NetServer {
	seedId := common.RandPeerKeyId()
	info := peer.NewPeerInfo(seedId.Id, 0, 0, true, 0,
		0, 0, "1.10", "")

	dis := NewDiscoveryProtocol(seeds, nil)
	dis.RefleshInterval = time.Millisecond * 10

	context := fmt.Sprintf("peer %s:, ", seedId.Id.ToHexString()[:6])
	logger := common.LoggerWithContext(common.NewGlobalLoggerWrapper(), context)
	return NewNode(seedId, "", info, dis, net, nil, p2p.AllAddrFilter(), logger)
}
