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
	"os"
	"testing"
	"time"

	p2pcomm "github.com/ontio/ontology/p2pserver/common"
)

var (
	nm *NbrPeers
)

const (
	startID     = 0xff
	startHeight = 434923
)

func TestMain(m *testing.M) {
	nm = initTestNbrPeers()
	os.Exit(m.Run())
}

func createPeers(cnt uint16) []*Peer {
	np := []*Peer{}
	var syncport uint16
	var id uint64
	var height uint64
	for i := uint16(0); i < cnt; i++ {
		syncport = 20224 + i
		id = startID + uint64(i)
		height = startHeight + uint64(i)
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
	if !nm.nodeExisted(startID) {
		t.Fatalf("%d should in nbr peers", startID)
	}
	if nm.nodeExisted(startID - 1) {
		t.Fatalf("%d should not in nbr peers", startID-1)
	}
}

func TestGetPeer(t *testing.T) {
	p := nm.GetPeer(startID)
	if p == nil {
		t.Fatal("TestGetPeer error")
	}
}

func TestAddAndDelNbrNode(t *testing.T) {
	p := NewPeer()
	var newID uint64
	newID = startID - 1
	p.UpdateInfo(time.Now(), 2, 3, 10335, newID, 0, 100, "1.5.2")
	p.SetState(p2pcomm.HAND_SHAKE)
	p.SetHttpInfoState(true)
	p.Link.SetAddr("127.0.0.1")
	orignLen := len(nm.List)
	nm.AddNbrNode(p)
	if !nm.nodeExisted(newID) {
		t.Fatalf("%d should be added in nbr peer", newID)
	}
	if len(nm.List) != orignLen+1 {
		t.Fatal("0x7123456 should be added in nbr peer")
	}

	cnt := len(nm.List)
	p, delOK := nm.DelNbrNode(newID)
	if p == nil || !delOK {
		t.Fatal("TestDelNbrNode err")
	}
	if len(nm.List) != cnt-1 {
		t.Fatal("TestDelNbrNode not work")
	}
	if p.GetID() != newID {
		t.Fatal("TestDelNbrNode return ID not valid")
	}
}

func TestNodeEstablished(t *testing.T) {
	p := nm.GetPeer(startID)
	if p == nil {
		t.Fatal("TestNodeEstablished:get peer error")
	}
	p.SetState(p2pcomm.ESTABLISH)
	if !nm.NodeEstablished(startID) {
		t.Fatal("TestNodeEstablished error")
	}
}

func TestGetNeighborAddrs(t *testing.T) {
	// all to init stat
	for _, v := range nm.List {
		v.SetState(p2pcomm.INIT)
	}

	p := nm.GetPeer(startID)
	if p == nil {
		t.Fatal("TestGetNeighborAddrs:get peer error")
	}
	p.SetState(p2pcomm.ESTABLISH)

	pList := nm.GetNeighborAddrs()
	for i := 0; i < len(pList); i++ {
		fmt.Printf("peer id = %x \n", pList[i].ID)
	}
	if len(pList) != 1 {
		t.Fatal("TestGetNeighborAddrs error")
	}
	if pList[0].ID != startID {
		t.Fatal("TestGetNeighborAddrs error")
	}

}

func TestGetNeighborHeights(t *testing.T) {
	p := nm.GetPeer(startID)
	if p == nil {
		t.Fatal("TestGetNeighborHeights:get peer error")
	}
	p.SetState(p2pcomm.ESTABLISH)

	pMap := nm.GetNeighborHeights()
	if len(pMap) != 1 {
		t.Fatal("GetNeighborHeights test fail")
	}
	if _, ok := pMap[startID]; !ok {
		t.Fatal("GetNeighborHeights test fail")
	}
	if pMap[startID] != startHeight {
		t.Fatal("GetNeighborHeights test fail")
	}
}

func TestGetNeighbors(t *testing.T) {
	for _, v := range nm.List {
		v.SetState(p2pcomm.INIT)
	}

	p := nm.GetPeer(startID)
	if p == nil {
		t.Fatal("TestGetNeighbors:get peer error")
	}
	p.SetState(p2pcomm.ESTABLISH)

	pList := nm.GetNeighbors()
	for _, v := range pList {
		v.DumpInfo()
	}
	if len(pList) != 1 {
		t.Fatalf("GetNeighbors test fail, expect size: %d, got %d\n", 1, len(pList))
	}
	if pList[0].GetID() != startID {
		t.Fatalf("GetNeighbors test fail, expect id: %d, got %d\n", startID, pList[0].GetID())
	}
}

func TestGetNbrNodeCnt(t *testing.T) {
	for _, v := range nm.List {
		v.SetState(p2pcomm.INIT)
	}

	p := nm.GetPeer(startID)
	if p == nil {
		t.Fatal("TestGetNbrNodeCnt:get peer error")
	}
	p.SetState(p2pcomm.ESTABLISH)

	if nm.GetNbrNodeCnt() != 1 {
		t.Fatal("TestGetNbrNodeCnt error")
	}
}
