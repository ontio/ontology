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

package p2pserver

import (
	"time"
	"sort"
	"sync"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
	p2pComm "github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	"github.com/ontio/ontology/p2pserver/peer"
)

//HeaderSyncMgr is the manager class to deal with header sync
type HeaderSyncMgr struct {
	server         *P2PServer             //Pointer to the local node
	syncHeaderLock bool                   //Help to avoid send header sync request duplicate
	exitCh         chan interface{}       //ExitCh to receive exit signal
	ledger         *ledger.Ledger         //ledger
	lock           sync.RWMutex           //lock
	nodeWeights    map[uint64]*NodeWeight //Map NodeID => NodeStatus, using for getNextNode
}

func NewHeaderSyncMgr(server *P2PServer) *HeaderSyncMgr {
	return &HeaderSyncMgr{
		server:        server,
		ledger:        server.ledger,
		exitCh:        make(chan interface{}, 1),
		nodeWeights:   make(map[uint64]*NodeWeight, 0),
	}
}

//Start to sync
func (this *HeaderSyncMgr) Start() {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-this.exitCh:
			return
		case <-ticker.C:
			go this.syncHeader()
		}
	}
}

func (this *HeaderSyncMgr) syncHeader() {
	if !this.server.reachMinConnection() {
		return
	}
	if this.tryGetSyncHeaderLock() {
		return
	}
	defer this.releaseSyncHeaderLock()

	curBlockHeight := this.ledger.GetCurrentBlockHeight()

	curHeaderHeight := this.ledger.GetCurrentHeaderHeight()

	if curHeaderHeight-curBlockHeight >= SYNC_MAX_HEADER_FORWARD_SIZE {
		return
	}

	nextHeaderId := curHeaderHeight + 1
	reqNode := this.getNextNode(nextHeaderId)
	if reqNode == nil {
		return
	}

	headerHash := this.ledger.GetCurrentHeaderHash()
	msg := msgpack.NewHeadersReq(headerHash)
	err := this.server.Send(reqNode, msg, false)
	if err != nil {
		log.Warn("[p2p]syncHeader failed to send a new headersReq")
	} else {
		this.appendReqTime(reqNode.GetID())
	}

	log.Infof("Header sync request height:%d", nextHeaderId)
}

//OnHeaderReceive receive header from net
func (this *HeaderSyncMgr) OnHeaderReceive(fromID uint64, headers []*types.Header) {
	if len(headers) == 0 {
		return
	}
	log.Infof("Header receive height:%d - %d", headers[0].Height, headers[len(headers)-1].Height)
	height := headers[0].Height
	curHeaderHeight := this.ledger.GetCurrentHeaderHeight()

	//Means another gorountinue is adding header
	if height <= curHeaderHeight {
		return
	}
	err := this.ledger.AddHeaders(headers)
	if err != nil {
		log.Warnf("[p2p]OnHeaderReceive AddHeaders error:%s", err)
		return
	}
	this.syncHeader()
}

func (this *HeaderSyncMgr) getNextNode(nextBlockHeight uint32) *peer.Peer {
	weights := this.getAllNodeWeights()
	sort.Sort(sort.Reverse(weights))
	nodelist := make([]uint64, 0)
	for _, n := range weights {
		nodelist = append(nodelist, n.id)
	}
	nextNodeIndex := 0
	triedNode := make(map[uint64]bool, 0)
	for {
		var nextNodeId uint64
		nextNodeIndex, nextNodeId = getNextNodeId(nextNodeIndex, nodelist)
		if nextNodeId == 0 {
			return nil
		}
		_, ok := triedNode[nextNodeId]
		if ok {
			return nil
		}
		triedNode[nextNodeId] = true
		n := this.server.getNode(nextNodeId)
		if n == nil {
			continue
		}
		if n.GetSyncState() != p2pComm.ESTABLISH {
			continue
		}
		nodeBlockHeight := n.GetHeight()
		if nextBlockHeight <= uint32(nodeBlockHeight) {
			return n
		}
	}
}

func (this *HeaderSyncMgr) getAllNodeWeights() NodeWeights {
	this.lock.RLock()
	defer this.lock.RUnlock()
	weights := make(NodeWeights, 0, len(this.nodeWeights))
	for _, w := range this.nodeWeights {
		weights = append(weights, w)
	}
	return weights
}

func (this *HeaderSyncMgr) tryGetSyncHeaderLock() bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.syncHeaderLock {
		return true
	}
	this.syncHeaderLock = true
	return false
}

func (this *HeaderSyncMgr) releaseSyncHeaderLock() {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.syncHeaderLock = false
}

func (this *HeaderSyncMgr) getNodeWeight(nodeId uint64) *NodeWeight {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.nodeWeights[nodeId]
}

//appendReqTime append a node's request time
func (this *HeaderSyncMgr) appendReqTime(nodeId uint64) {
	n := this.getNodeWeight(nodeId)
	if n != nil {
		n.AppendNewReqtime()
	}
}
