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

package vbft

import (
	"fmt"
	"sort"
	"sync/atomic"
	"time"

	"github.com/ontio/ontology/common/log"
)

const (
	MAX_PEER_CONNECTIONS      = 10
	MAX_SYNCING_CHECK_BLK_NUM = 10
)

type ServerState uint32

const (
	Initialized      ServerState = iota
	Syncing                      // syncing block from neighbours
	WaitNetworkReady             // sync reached, and keep synced, try connecting with more peers
	SyncReady                    // start processing consensus msg, but not broadcasting proposal/endorse/commit
	Active                       // start bft
)

func (state ServerState) IsReady() bool {
	return state == SyncReady || state == Active
}

func (state ServerState) IsActive() bool {
	return state == Active
}

func (state ServerState) String() string {
	switch state {
	case Initialized:
		return "Initialized"
	case Syncing:
		return "Syncing"
	case WaitNetworkReady:
		return "WaitNetworkReady"
	case SyncReady:
		return "SyncReady"
	case Active:
		return "Active"
	default:
		return "unknown state"
	}
}

type StateEventType int

const (
	UpdatePeerState StateEventType = iota
	SyncReadyTimeout
	ForceSyncing
	LiveTick
)

type StateEvent struct {
	Type      StateEventType
	peerState *PeerState
	blockNum  uint32
}

type PeerState struct {
	peerIdx           uint32
	chainConfigView   uint32
	committedBlockNum uint32
	connected         bool
}

func (peer *PeerState) String() string {
	return fmt.Sprintf("{%d: (%d, %d)},", peer.peerIdx, peer.chainConfigView, peer.committedBlockNum)
}

type StateMgr struct {
	server           *Server
	syncReadyTimeout time.Duration
	currentState     ServerState
	StateEventC      chan *StateEvent
	peers            map[uint32]*PeerState

	liveTicker            *time.Timer
	lastTickHeight        uint32
	lastSyncRequestHeight uint32
}

func newStateMgr(server *Server) *StateMgr {
	return &StateMgr{
		server:           server,
		syncReadyTimeout: time.Second * 10,
		currentState:     Initialized,
		StateEventC:      make(chan *StateEvent, 16),
		peers:            make(map[uint32]*PeerState),
	}
}

func (self *StateMgr) getState() ServerState {
	state := atomic.LoadUint32((*uint32)(&self.currentState))
	return ServerState(state)
}

func (self *StateMgr) setState(newstate ServerState) {
	atomic.StoreUint32((*uint32)(&self.currentState), uint32(newstate))
}

func (self *StateMgr) run() {
	liveTimeout := time.Duration(atomic.LoadInt64(&peerHandshakeTimeout) * 5)
	self.liveTicker = time.AfterFunc(liveTimeout, func() {
		self.StateEventC <- &StateEvent{
			Type:     LiveTick,
			blockNum: self.server.GetCurrentBlockNo() - 1,
		}
		liveTimeout = time.Duration(atomic.LoadInt64(&peerHandshakeTimeout) * 3)
		self.liveTicker.Reset(liveTimeout)
	})

	// wait config done
	self.server.quitWg.Add(1)
	defer self.server.quitWg.Done()

	for {
		select {
		case evt := <-self.StateEventC:
			switch evt.Type {
			case SyncReadyTimeout:
				if self.getState() == SyncReady {
					self.setState(Active)
					if evt.blockNum == self.server.GetCurrentBlockNo() {
						self.server.startNewRound()
					}
				}
			case UpdatePeerState:
				if evt.peerState.connected {
					self.onPeerUpdate(evt.peerState)
				} else {
					self.onPeerDisconnected(evt.peerState.peerIdx)
				}
			case ForceSyncing:
				if self.server.nonConsensusNode() {
					// non-consensus node, block-syncer do the syncing
					return
				}

				self.setState(Syncing)
			case LiveTick:
				log.Infof("server %d peer update, current blk: %d, state: %s. received peer states: %v",
					self.server.Index, self.server.GetCurrentBlockNo(), self.getState(), self.peers)
				self.onLiveTick(evt)
			}

		case <-self.server.quitC:
			log.Infof("server %d, state mgr quit", self.server.Index)
			return
		}
	}
}

func (self *StateMgr) onPeerUpdate(peerState *PeerState) {
	peerIdx := peerState.peerIdx
	self.peers[peerIdx] = peerState

	vbftCtx := self.server.GetVbftContext()
	log.Debugf("server %d peer update, current blk %d, state %s, received peer state: %v",
		self.server.Index, vbftCtx.BlockNum, self.getState(), peerState)

	switch self.getState() {
	case Initialized:
		v := self.getPeersView()
		log.Infof("server %d statemgr update, current state: %s, from peer: %d, peercnt: %d, v1: %d, v2: %d",
			self.server.Index, self.getState(), peerIdx, len(self.peers), v, vbftCtx.Config.View)

		if v == vbftCtx.Config.View {
			self.setState(Syncing)
		}
	case Syncing:
		self.trySetSyncedReady()
	case WaitNetworkReady:
		self.trySetSyncedReady()
	case SyncReady:
	case Active:
	}
}

func (self *StateMgr) onPeerDisconnected(peerIdx uint32) {
	if _, present := self.peers[peerIdx]; !present {
		return
	}
	delete(self.peers, peerIdx)
	if self.getState().IsActive() {
		if len(self.peers) < self.getMinActivePeerCount() {
			self.setState(WaitNetworkReady)
		}
	}
}

func (self *StateMgr) onLiveTick(evt *StateEvent) {
	if evt.blockNum > self.lastTickHeight || self.lastTickHeight == 0 {
		self.lastTickHeight = evt.blockNum
		return
	}

	if !self.getState().IsReady() {
		return
	}

	log.Warnf("server %d detected consensus halt %d", self.server.Index, self.server.GetCurrentBlockNo())
	if self.checkFallBehind() {
		self.setState(Syncing)
	}

	self.server.reBroadcastCurrentRoundMsgs()
}

func (self *StateMgr) checkFallBehind() (needSync bool) {
	committedBlkNum, ok := self.getConsensusedCommittedBlockNum()
	return ok && committedBlkNum > self.server.GetCurrentBlockNo()-1
}

func (self *StateMgr) getMinActivePeerCount() int {
	n := int(self.server.GetVbftContext().Config.C) * 2 // plus self
	if n > MAX_PEER_CONNECTIONS {
		// FIXME: C vs. maxConnections
		return MAX_PEER_CONNECTIONS
	}
	return n
}

func (self *StateMgr) getPeersView() uint32 {
	if len(self.peers) < self.getMinActivePeerCount() {
		return 0
	}

	views := make(map[uint32]int)
	for _, p := range self.peers {
		views[p.chainConfigView] += 1
	}

	for k, v := range views {
		if v >= self.getMinActivePeerCount() {
			return k
		}
	}

	return 0
}

func (self *StateMgr) trySetSyncedReady() {
	committedBlkNum, ok := self.getConsensusedCommittedBlockNum()
	if !ok || len(self.peers) < self.getMinActivePeerCount() {
		return
	}

	if self.server.GetCurrentBlockNo() > committedBlkNum {
		self.setState(SyncReady)
		log.Infof("server %d start sync ready", self.server.Index)
		blkNum := self.server.GetCurrentBlockNo()
		time.AfterFunc(self.syncReadyTimeout, func() {
			self.StateEventC <- &StateEvent{
				Type:     SyncReadyTimeout,
				blockNum: blkNum,
			}
		})
	}
}

// return (0, false) if consensus not reached yet, or else (committedNum, true)
func (self *StateMgr) getConsensusedCommittedBlockNum() (uint32, bool) {
	list := self.getPeersCommittedBlockNoSorted()
	vbftCtx := self.server.GetVbftContext()
	c := int(vbftCtx.Config.C)
	if len(list) >= c+1 && list[c] >= vbftCtx.BlockNum-1 {
		return list[c], true
	}

	return 0, false
}

func (self *StateMgr) getPeersCommittedBlockNoSorted() []uint32 {
	list := make([]uint32, 0, len(self.peers))
	for _, p := range self.peers {
		list = append(list, p.committedBlockNum)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i] > list[j]
	})
	return list
}
