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

	"github.com/ontio/ontology/common/log"
)

var nm *NbrPeers

func creatPeers(cnt uint16) []*Peer {
	np := []*Peer{}
	var syncport uint16
	var consport uint16
	var id uint64
	var height uint64
	for i := uint16(0); i < cnt; i++ {
		syncport = 20224 + i
		consport = 20335 + i
		id = 0x7533345 + uint64(i)
		height = 434923 + uint64(i)
		p = NewPeer()
		p.UpdateInfo(time.Now(), 2, 3, syncport, consport, id, 0, height)
		p.SetConsState(2)
		p.SetSyncState(3)
		p.SetHttpInfoState(true)
		p.SyncLink.SetAddr("127.0.0.1:10338")
		np = append(np, p)
	}
	return np

}

func init() {
	log.Init(log.Stdout)
	nm = &NbrPeers{}
	nm.Init()
	np := creatPeers(5)
	for _, v := range np {
		nm.List[v.GetID()] = v
	}
}

func TestNodeExisted(t *testing.T) {
	if nm.NodeExisted(0x7533345) == false {
		t.Fatal("0x7533345 should in nbr peers")
	}
	if nm.NodeExisted(0x5533345) == true {
		t.Fatal("0x5533345 should not in nbr peers")
	}
}

func TestGetPeer(t *testing.T) {
	p := nm.GetPeer(0x7533345)
	if p == nil {
		t.Fatal("TestGetPeer error")
	}
}

func TestAddNbrNode(t *testing.T) {
	p := NewPeer()
	p.UpdateInfo(time.Now(), 2, 3, 10335, 10336, 0x7123456, 0, 100)
	p.SetConsState(2)
	p.SetSyncState(3)
	p.SetHttpInfoState(true)
	p.SyncLink.SetAddr("127.0.0.1")
	nm.AddNbrNode(p)
	if nm.NodeExisted(0x7123456) == false {
		t.Fatal("0x7123456 should be added in nbr peer")
	}
	if len(nm.List) != 6 {
		t.Fatal("0x7123456 should be added in nbr peer")
	}
}

func TestDelNbrNode(t *testing.T) {
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
	p := nm.GetPeer(0x7533346)
	if p == nil {
		t.Fatal("TestNodeEstablished:get peer error")
	}
	p.SetSyncState(4)
	if nm.NodeEstablished(0x7533346) == false {
		t.Fatal("TestNodeEstablished error")
	}
}

func TestGetNeighborAddrs(t *testing.T) {
	p := nm.GetPeer(0x7533346)
	if p == nil {
		t.Fatal("TestGetNeighborAddrs:get peer error")
	}
	p.SetSyncState(4)

	p = nm.GetPeer(0x7533347)
	if p == nil {
		t.Fatal("TestGetNeighborAddrs:get peer error")
	}
	p.SetSyncState(4)

	pList := nm.GetNeighborAddrs()
	for i := 0; i < int(cnt); i++ {
		fmt.Printf("peer id = %x \n", pList[i].ID)
	}
	if cnt != 2 {
		t.Fatal("TestGetNeighborAddrs error")
	}
}

func TestGetNeighborHeights(t *testing.T) {
	p := nm.GetPeer(0x7533346)
	if p == nil {
		t.Fatal("TestGetNeighborHeights:get peer error")
	}
	p.SetSyncState(4)

	p = nm.GetPeer(0x7533347)
	if p == nil {
		t.Fatal("TestGetNeighborHeights:get peer error")
	}
	p.SetSyncState(4)

	pMap := nm.GetNeighborHeights()
	for k, v := range pMap {
		fmt.Printf("peer id = %x height = %d \n", k, v)
	}
}

func TestGetNeighbors(t *testing.T) {
	p := nm.GetPeer(0x7533346)
	if p == nil {
		t.Fatal("TestGetNeighbors:get peer error")
	}
	p.SetSyncState(4)

	p = nm.GetPeer(0x7533347)
	if p == nil {
		t.Fatal("TestGetNeighbors:get peer error")
	}
	p.SetSyncState(4)

	pList := nm.GetNeighbors()
	for _, v := range pList {
		v.DumpInfo()
	}
}

func TestGetNbrNodeCnt(t *testing.T) {
	p := nm.GetPeer(0x7533346)
	if p == nil {
		t.Fatal("TestGetNbrNodeCnt:get peer error")
	}
	p.SetSyncState(4)

	p = nm.GetPeer(0x7533347)
	if p == nil {
		t.Fatal("TestGetNbrNodeCnt:get peer error")
	}
	p.SetSyncState(4)

	if nm.GetNbrNodeCnt() != 2 {
		t.Fatal("TestGetNbrNodeCnt error")
	}
}
