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
	"fmt"
	"testing"
	"time"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/peer"
	"github.com/stretchr/testify/require"
)

func init() {
	log.InitLog(log.InfoLog, log.Stdout)
	fmt.Println("Start test the netserver...")
}

func creatPeers(cnt uint16) []*peer.Peer {
	np := []*peer.Peer{}
	var syncport uint16
	var id uint64
	var height uint64
	for i := uint16(0); i < cnt; i++ {
		syncport = 20224 + i
		id = 0x7533345 + uint64(i)
		height = 434923 + uint64(i)
		p := peer.NewPeer()
		p.UpdateInfo(time.Now(), 2, 3, syncport, id, 0, height, "1.5.2")
		p.SetState(4)
		p.SetHttpInfoState(true)
		p.Link.SetAddr("127.0.0.1:10338")
		np = append(np, p)
	}
	return np

}
func TestNewNetServer(t *testing.T) {
	server := NewNetServer()
	server.Start()
	defer server.Halt()

	server.SetHeight(1000)
	if server.GetHeight() != 1000 {
		t.Error("TestNewNetServer set server height error")
	}

	if !server.GetRelay() {
		t.Error("TestNewNetServer server relay state error", server.GetRelay())
	}
	if server.GetServices() != 1 {
		t.Error("TestNewNetServer server service state error", server.GetServices())
	}
	if server.GetVersion() != common.PROTOCOL_VERSION {
		t.Error("TestNewNetServer server version error", server.GetVersion())
	}
	if server.GetPort() != 20338 {
		t.Error("TestNewNetServer sync port error", server.GetPort())
	}

	fmt.Printf("lastest server time is %s\n", time.Unix(server.GetTime()/1e9, 0).String())
}

func TestNetServerNbrPeer(t *testing.T) {
	server := NewNetServer()
	server.Start()
	defer server.Halt()

	nm := &peer.NbrPeers{}
	nm.Init()
	np := creatPeers(5)
	for _, v := range np {
		server.AddNbrNode(v)
	}
	if server.GetConnectionCnt() != 5 {
		t.Error("TestNetServerNbrPeer GetConnectionCnt error", server.GetConnectionCnt())
	}
	addrs := server.GetNeighborAddrs()
	if len(addrs) != 5 {
		t.Error("TestNetServerNbrPeer GetNeighborAddrs error")
	}
	if !server.NodeEstablished(0x7533345) {
		t.Error("TestNetServerNbrPeer NodeEstablished error")
	}
	if server.GetPeer(0x7533345) == nil {
		t.Error("TestNetServerNbrPeer GetPeer error")
	}
	p, ok := server.DelNbrNode(0x7533345)
	if !ok || p == nil {
		t.Error("TestNetServerNbrPeer DelNbrNode error")
	}
	if len(server.GetNeighbors()) != 4 {
		t.Error("TestNetServerNbrPeer GetNeighbors error")
	}
	sp := &peer.Peer{}
	server.AddPeerAddress("127.0.0.1:10338", sp)
	if server.GetPeerFromAddr("127.0.0.1:10338") != sp {
		t.Error("TestNetServerNbrPeer Get/AddPeerConsAddress error")
	}

}

func TestConnectingNodeAPI(t *testing.T) {
	a := require.New(t)
	server := NewNetServer()

	a.Equal(server.GetOutConnectingListLen(), uint(0), "fail to test GetOutConnectingListLen")

	addOK := server.AddOutConnectingList("192.168.1.1:28339")
	a.Equal(server.GetOutConnectingListLen(), uint(1), "fail to test AddOutConnectingList")
	a.Equal(addOK, true, "fail to test AddOutConnectingList")

	// add same
	addOK = server.AddOutConnectingList("192.168.1.1:28339")
	a.Equal(server.GetOutConnectingListLen(), uint(1), "fail to test AddOutConnectingList")
	a.Equal(addOK, false, "fail to test AddOutConnectingList")

	// add new
	server.AddOutConnectingList("192.168.2.2:2")
	a.Equal(server.GetOutConnectingListLen(), uint(2), "fail to test AddOutConnectingList")

	// test exist
	a.Equal(server.IsAddrFromConnecting("192.168.2.2:2"), true, "fail to test IsAddrFromConnecting")
	a.Equal(server.IsAddrFromConnecting("192.168.2.3:3"), false, "fail to test IsAddrFromConnecting")

	server.RemoveFromConnectingList("192.168.2.2:2")
	a.Equal(server.GetOutConnectingListLen(), uint(1), "fail to test RemoveFromConnectingList")
}

func TestInConnAPI(t *testing.T) {
	a := require.New(t)
	si := NewNetServer()
	server, ok := si.(*NetServer)
	a.True(ok, "fail to cast P2PServer")

	a.Equal(server.GetInConnRecordLen(), int(0), "fail to test GetInConnRecordLen")
	server.AddInConnRecord("192.168.1.1:1024")
	a.Equal(server.GetInConnRecordLen(), int(1), "fail to test AddInConnRecord")
	server.AddInConnRecord("192.168.1.1:1024")
	a.Equal(server.GetInConnRecordLen(), int(1), "fail to test GetInConnRecordLen")
	server.AddInConnRecord("192.168.1.2:2048")
	a.Equal(server.GetInConnRecordLen(), int(2), "fail to test AddInConnRecord")
	server.RemoveFromInConnRecord("192.168.1.2:2048")
	a.Equal(server.GetInConnRecordLen(), int(1), "fail to test RemoveFromInConnRecord")
	// same IP, different port
	server.AddInConnRecord("192.168.1.1:2048")
	a.Equal(server.GetInConnRecordLen(), int(2), "fail to test RemoveFromInConnRecord")

	a.Equal(server.GetIpCountInInConnRecord("192.168.1.1"), uint(2), "fail to test GetIpCountInInConnRecord")
}

func TestOutConnAPI(t *testing.T) {
	a := require.New(t)
	si := NewNetServer()
	server, ok := si.(*NetServer)
	a.True(ok, "fail to case P2PServer")

	a.Equal(server.GetOutConnRecordLen(), int(0), "fail to test GetOutConnRecordLen")
	server.AddOutConnRecord("192.168.1.1:200")
	a.Equal(server.GetOutConnRecordLen(), int(1), "fail to test AddOutConnRecord")
	server.AddOutConnRecord("192.168.1.1:200")
	a.Equal(server.GetOutConnRecordLen(), int(1), "fail to test AddOutConnRecord")
	server.AddOutConnRecord("192.168.1.1:300")
	a.Equal(server.GetOutConnRecordLen(), int(2), "fail to test AddOutConnRecord")
	server.RemoveFromOutConnRecord("192.168.1.1:300")
	a.Equal(server.GetOutConnRecordLen(), int(1), "fail to test RemoveFromOutConnRecord")
	a.Equal(server.IsAddrInOutConnRecord("192.168.1.1:300"), false, "fail to test IsAddrInOutConnRecord")
	a.Equal(server.IsAddrInOutConnRecord("192.168.1.1:200"), true, "fail to test IsAddrInOutConnRecord")
}
