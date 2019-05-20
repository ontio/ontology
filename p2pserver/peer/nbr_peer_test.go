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
	"testing"
	"time"
)

func createPeers(cnt uint16) []*Peer {
	np := []*Peer{}
	var syncport uint16
	var id uint64
	var height uint64
	for i := uint16(0); i < cnt; i++ {
		syncport = 20224 + i
		id = 0x7533345 + uint64(i)
		height = 434923 + uint64(i)
		p := NewPeer()
		p.UpdateInfo(time.Now(), 2, 3, syncport, id, 0, height, "1.5.2")
		p.SetState(3)
		p.SetHttpInfoState(true)
		p.Link.SetAddr("127.0.0.1:10338")
		np = append(np, p)
	}
	return np

}

func initTestNbrPeers() *NbrPeers {
	nm := &NbrPeers{}
	nm.Init()
	np := createPeers(5)
	for _, v := range np {
		nm.List[v.GetID()] = v
	}
	return nm
}

func TestNodeExisted(t *testing.T) {
	nm := initTestNbrPeers()

	if nm.NodeExisted(0x7533345) == false {
		t.Fatal("0x7533345 should in nbr peers")
	}
	if nm.NodeExisted(0x5533345) == true {
		t.Fatal("0x5533345 should not in nbr peers")
	}
}

func TestGetPeer(t *testing.T) {
	nm := initTestNbrPeers()

	p := nm.GetPeer(0x7533345)
	if p == nil {
		t.Fatal("TestGetPeer error")
	}
}

func TestAddNbrNode(t *testing.T) {
	nm := initTestNbrPeers()

	p := NewPeer()
	p.UpdateInfo(time.Now(), 2, 3, 10335, 0x7123456, 0, 100, "1.5.2")
	p.SetState(3)
	p.SetHttpInfoState(true)
	p.Link.SetAddr("127.0.0.1")
	nm.AddNbrNode(p)
	if nm.NodeExisted(0x7123456) == false {
		t.Fatal("0x7123456 should be added in nbr peer")
	}
	if len(nm.List) != 6 {
		t.Fatal("0x7123456 should be added in nbr peer")
	}
}

func TestDelNbrNode(t *testing.T) {
	nm := initTestNbrPeers()

	cnt := len(nm.List)
	p, ret := nm.DelNbrNode(0x7533345)
	if p == nil || ret != true {
		t.Fatal("TestDelNbrNode err")
	}
	if len(nm.List) != cnt-1 {
		t.Fatal("TestDelNbrNode not work")
	}
	p.DumpInfo()
}

func TestNodeEstablished(t *testing.T) {
	nm := initTestNbrPeers()

	p := nm.GetPeer(0x7533346)
	if p == nil {
		t.Fatal("TestNodeEstablished:get peer error")
	}
	p.SetState(4)
	if nm.NodeEstablished(0x7533346) == false {
		t.Fatal("TestNodeEstablished error")
	}
}

func TestGetNeighborAddrs(t *testing.T) {
	nm := initTestNbrPeers()

	p := nm.GetPeer(0x7533346)
	if p == nil {
		t.Fatal("TestGetNeighborAddrs:get peer error")
	}
	p.SetState(4)

	p = nm.GetPeer(0x7533347)
	if p == nil {
		t.Fatal("TestGetNeighborAddrs:get peer error")
	}
	p.SetState(4)

	pList := nm.GetNeighborAddrs()
	for i := 0; i < len(pList); i++ {
		fmt.Printf("peer id = %x \n", pList[i].ID)
	}
	if len(pList) != 2 {
		t.Fatal("TestGetNeighborAddrs error")
	}
}

func TestGetNeighborHeights(t *testing.T) {
	nm := initTestNbrPeers()
	p := nm.GetPeer(0x7533346)
	if p == nil {
		t.Fatal("TestGetNeighborHeights:get peer error")
	}
	p.SetState(4)

	p = nm.GetPeer(0x7533347)
	if p == nil {
		t.Fatal("TestGetNeighborHeights:get peer error")
	}
	p.SetState(4)

	pMap := nm.GetNeighborHeights()
	for k, v := range pMap {
		fmt.Printf("peer id = %x height = %d \n", k, v)
	}
}

func TestGetNeighbors(t *testing.T) {
	nm := initTestNbrPeers()

	p := nm.GetPeer(0x7533346)
	if p == nil {
		t.Fatal("TestGetNeighbors:get peer error")
	}
	p.SetState(4)

	p = nm.GetPeer(0x7533347)
	if p == nil {
		t.Fatal("TestGetNeighbors:get peer error")
	}
	p.SetState(4)

	pList := nm.GetNeighbors()
	for _, v := range pList {
		v.DumpInfo()
	}
}

func TestGetNbrNodeCnt(t *testing.T) {
	nm := initTestNbrPeers()

	p := nm.GetPeer(0x7533346)
	if p == nil {
		t.Fatal("TestGetNbrNodeCnt:get peer error")
	}
	p.SetState(4)

	p = nm.GetPeer(0x7533347)
	if p == nil {
		t.Fatal("TestGetNbrNodeCnt:get peer error")
	}
	p.SetState(4)

	if nm.GetNbrNodeCnt() != 2 {
		t.Fatal("TestGetNbrNodeCnt error")
	}
}
