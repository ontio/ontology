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
	"time"

	"github.com/ontio/ontology/common/log"
)

const (
	// TODO: move to config file
	MAX_PEER_CONNECTIONS      = 100
	MAX_SYNCING_CHECK_BLK_NUM = 10
)

type ServerState int

const (
	Init ServerState = iota
	LocalConfigured
	Configured       // config loaded from chain
	Syncing          // syncing block from neighbours
	WaitNetworkReady // sync reached, and keep synced, try connecting with more peers
	SyncReady        // start processing consensus msg, but not broadcasting proposal/endorse/commit
	Synced           // start bft
	SyncingCheck     // potentially lost syncing
)

func isReady(state ServerState) bool {
	return state >= SyncReady
}

func isActive(state ServerState) bool {
	return state >= Synced
}

type StateEventType int

const (
	ConfigLoaded     StateEventType = iota
	UpdatePeerConfig                // notify statemgmt on peer heartbeat
	UpdatePeerState                 // notify statemgmt on peer heartbeat
	SyncReadyTimeout
	SyncDone
	LiveTick
)

type StateEvent struct {
	Type      StateEventType
	peerState *PeerState
	blockNum  uint64
}

type PeerState struct {
	peerIdx           uint32
	chainConfigView   uint32
	committedBlockNum uint64
	connected         bool
}

type StateMgr struct {
	server           *Server
	syncReadyTimeout time.Duration
	currentState     ServerState
	StateEventC      chan *StateEvent
	peers            map[uint32]*PeerState

	liveTicker             *time.Timer
	lastTickChainHeight    uint64
	lastBlockSyncReqHeight uint64
}

func newStateMgr(server *Server) *StateMgr {
	return &StateMgr{
		server:           server,
		syncReadyTimeout: time.Second * 10,
		currentState:     Init,
		StateEventC:      make(chan *StateEvent, 16),
		peers:            make(map[uint32]*PeerState),
	}
}

func (self *StateMgr) getState() ServerState {
	return self.currentState
}

func (self *StateMgr) run() {
	self.liveTicker = time.AfterFunc(peerHandshakeTimeout*5, func() {
		self.StateEventC <- &StateEvent{
			Type:     LiveTick,
			blockNum: self.server.GetCommittedBlockNo(),
		}
		self.liveTicker.Reset(peerHandshakeTimeout * 3)
	})

	// wait config done
	self.server.quitWg.Add(1)
	defer self.server.quitWg.Done()

	for {
		select {
		case evt := <-self.StateEventC:
			switch evt.Type {
			case ConfigLoaded:
				if self.currentState == Init {
					self.currentState = LocalConfigured
				}
			case SyncReadyTimeout:
				if self.currentState == SyncReady {
					self.currentState = Synced
					if evt.blockNum == self.server.GetCurrentBlockNo() {
						self.server.startNewRound()
					}
				}
			case UpdatePeerConfig:
				peerIdx := evt.peerState.peerIdx
				self.peers[peerIdx] = evt.peerState

				if self.currentState >= LocalConfigured {
					v := self.getSyncedChainConfigView()
					if v == self.server.config.View && self.currentState < Syncing {
						log.Infof("server %d, start syncing", self.server.Index)
						self.currentState = Syncing
					} else if v > self.server.config.View {
						// update ChainConfig
						log.Errorf("todo: chain config changed, need update chain config from peers")
						// TODO: fetch config from neighbours, update chain config
						self.currentState = LocalConfigured
					}
				}
			case UpdatePeerState:
				if evt.peerState.connected {
					if err := self.onPeerUpdate(evt.peerState); err != nil {
						log.Errorf("statemgr process peer (%d) err: %s", evt.peerState.peerIdx, err)
					}
				} else {
					if err := self.onPeerDisconnected(evt.peerState.peerIdx); err != nil {
						log.Errorf("statmgr process peer (%d) disconn err: %s", evt.peerState.peerIdx, err)
					}
				}

			case SyncDone:
				log.Infof("server %d sync done, curr blkNum: %d", self.server.Index, self.server.GetCurrentBlockNo())
				self.setSyncedReady()

			case LiveTick:
				if err := self.onLiveTick(evt); err != nil {
					log.Errorf("server %d, live ticker: %s", self.server.Index, err)
				}
			}

		case <-self.server.quitC:
			log.Infof("server %d, state mgr quit", self.server.Index)
			return
		}
	}
}

func (self *StateMgr) onPeerUpdate(peerState *PeerState) error {
	peerIdx := peerState.peerIdx
	newPeer := false
	if _, present := self.peers[peerIdx]; !present {
		newPeer = true
	}

	log.Infof("server %d peer update, current blk %d, state %d, from peer %d, committed %d",
		self.server.Index, self.server.GetCurrentBlockNo(), self.currentState, peerState.peerIdx, peerState.committedBlockNum)

	// update peer state
	self.peers[peerIdx] = peerState

	if !newPeer {
		if isActive(self.currentState) && peerState.committedBlockNum > self.server.GetCurrentBlockNo()+MAX_SYNCING_CHECK_BLK_NUM {
			log.Warnf("server %d seems lost sync: %d(%d) vs %d", self.server.Index,
				peerState.committedBlockNum, peerState.peerIdx, self.server.GetCurrentBlockNo())
			if err := self.checkStartSyncing(self.server.GetCommittedBlockNo()+MAX_SYNCING_CHECK_BLK_NUM, false); err != nil {
				log.Errorf("server %d start syncing check failed", self.server.Index)
			}
			return nil
		}
	}

	switch self.currentState {
	case LocalConfigured:
		v := self.getSyncedChainConfigView()
		log.Infof("server %d statemgr update, current state: %d, from peer: %d, peercnt: %d, v1: %d, v2: %d",
			self.server.Index, self.currentState, peerIdx, len(self.peers), v, self.server.config.View)

		if v == self.server.config.View {
			self.currentState = Syncing
		}
	case Configured:
	case Syncing:
		if peerState.committedBlockNum > self.server.GetCommittedBlockNo() {

			committedBlkNum, ok := self.getConsensusedCommittedBlockNum()
			if ok && committedBlkNum > self.server.GetCommittedBlockNo() {
				fastforward := self.canFastForward(committedBlkNum)
				log.Infof("server %d, syncing %d, target %d, fastforward %t",
					self.server.Index, self.server.GetCommittedBlockNo(), committedBlkNum, fastforward)
				if fastforward {
					self.server.makeFastForward()
				} else {
					self.checkStartSyncing(self.server.GetCommittedBlockNo(), false)
				}
			}
		}
		if self.isSyncedReady() {
			log.Infof("server %d synced from syncing", self.server.Index)
			self.setSyncedReady()
		}
	case WaitNetworkReady:
		if self.isSyncedReady() {
			log.Infof("server %d synced from sync-ready", self.server.Index)
			self.setSyncedReady()
		}
	case SyncReady:
	case Synced:
		committedBlkNum, ok := self.getConsensusedCommittedBlockNum()
		if ok && committedBlkNum > self.server.GetCommittedBlockNo()+1 {
			log.Infof("server %d synced try fastforward from %d",
				self.server.Index, self.server.GetCommittedBlockNo())
			self.server.makeFastForward()
		}
	case SyncingCheck:
		if self.isSyncedReady() {
			self.setSyncedReady()
		} else {
			self.checkStartSyncing(self.server.GetCommittedBlockNo()+MAX_SYNCING_CHECK_BLK_NUM, false)
		}
	}

	return nil
}

func (self *StateMgr) onPeerDisconnected(peerIdx uint32) error {

	if _, present := self.peers[peerIdx]; !present {
		return nil
	}
	delete(self.peers, peerIdx)

	// start another connection if necessary
	if self.currentState == Synced || self.currentState == SyncingCheck {
		if self.server.peerPool.getActivePeerCount() < self.getMinActivePeerCount() {
			self.currentState = WaitNetworkReady
		}
	}

	return nil
}

func (self *StateMgr) onLiveTick(evt *StateEvent) error {
	if evt.blockNum > self.lastTickChainHeight {
		self.lastTickChainHeight = evt.blockNum
		return nil
	}

	if self.lastTickChainHeight == 0 {
		self.lastTickChainHeight = evt.blockNum
		return nil
	}

	if self.getState() != Synced {
		return nil
	}

	log.Errorf("server %d detected consensus halt %d",
		self.server.Index, self.server.GetCurrentBlockNo())

	committedBlkNum, ok := self.getConsensusedCommittedBlockNum()
	if ok && committedBlkNum > self.server.GetCommittedBlockNo() {
		fastforward := self.canFastForward(committedBlkNum)
		log.Infof("server %d, syncing %d, target %d, fast-forward %t",
			self.server.Index, self.server.GetCommittedBlockNo(), committedBlkNum, fastforward)
		if fastforward {
			self.server.makeFastForward()
		} else {
			self.checkStartSyncing(self.server.GetCommittedBlockNo(), false)
		}
	}

	return self.server.reBroadcastCurrentRoundMsgs()
}

func (self *StateMgr) getMinActivePeerCount() int {
	n := int(self.server.config.C) * 2 // plus self
	if n > MAX_PEER_CONNECTIONS {
		// FIXME: C vs. maxConnections
		return MAX_PEER_CONNECTIONS
	}
	return n
}

func (self *StateMgr) getSyncedChainConfigView() uint32 {
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

func (self *StateMgr) isSyncedReady() bool {
	// check action peer connections
	if len(self.peers) < self.getMinActivePeerCount() {
		return false
	}

	// check chain consensus
	committedBlkNum, ok := self.getConsensusedCommittedBlockNum()
	if !ok {
		return false
	}
	if self.server.GetCommittedBlockNo() >= committedBlkNum {
		return true
	}

	return self.canFastForward(committedBlkNum)
}

func (self *StateMgr) setSyncedReady() error {
	prevState := self.currentState
	self.currentState = SyncReady
	if prevState <= SyncReady {
		log.Infof("server %d start sync ready", self.server.Index)
		blkNum := self.server.GetCurrentBlockNo()
		time.AfterFunc(self.syncReadyTimeout, func() {
			self.StateEventC <- &StateEvent{
				Type:     SyncReadyTimeout,
				blockNum: blkNum,
			}
		})
		self.server.makeFastForward()
	}

	return nil
}

func (self *StateMgr) checkStartSyncing(startBlkNum uint64, forceSync bool) error {

	var maxCommitted uint64
	peers := make(map[uint64][]uint32)
	for _, p := range self.peers {
		n := p.committedBlockNum
		if n > startBlkNum {
			if _, present := peers[n]; !present {
				peers[n] = make([]uint32, 0)
			}
			for k := range peers {
				if n >= k {
					peers[k] = append(peers[k], p.peerIdx)
				}
			}
			if len(peers[n]) > int(self.server.config.C) {
				maxCommitted = n
			}
		}
	}

	if maxCommitted > startBlkNum || forceSync {
		self.currentState = Syncing
		startBlkNum = self.server.GetCommittedBlockNo() + 1

		if maxCommitted > self.lastBlockSyncReqHeight {
			// syncer is much slower than peer-update, too much SyncReq can make channel full
			log.Infof("server %d, start syncing %d - %d, with %v", self.server.Index, startBlkNum, maxCommitted, peers)
			self.lastBlockSyncReqHeight = maxCommitted
			self.server.syncer.blockSyncReqC <- &BlockSyncReq{
				targetPeers:    peers[maxCommitted],
				startBlockNum:  startBlkNum,
				targetBlockNum: maxCommitted,
			}
		}
	} else if self.currentState == Synced {
		log.Infof("server %d, start syncing check %v, %d", self.server.Index, peers, self.server.GetCurrentBlockNo())
		self.currentState = SyncingCheck
	}

	return nil
}

// return 0 if consensus not reached yet
func (self *StateMgr) getConsensusedCommittedBlockNum() (uint64, bool) {
	C := int(self.server.config.C)

	consensused := false
	var maxCommitted uint64
	myCommitted := self.server.GetCommittedBlockNo()
	peers := make(map[uint64][]uint32)
	for _, p := range self.peers {
		n := p.committedBlockNum
		if n >= myCommitted {
			if _, present := peers[n]; !present {
				peers[n] = make([]uint32, 0)
			}
			for k := range peers {
				if n >= k {
					peers[k] = append(peers[k], p.peerIdx)
				}
			}
			if len(peers[n]) > C {
				maxCommitted = n
				consensused = true
			}
		}
	}

	return maxCommitted, consensused
}

func (self *StateMgr) canFastForward(targetBlkNum uint64) bool {
	if targetBlkNum > self.server.GetCommittedBlockNo()+MAX_SYNCING_CHECK_BLK_NUM*4 {
		return false
	}

	C := int(self.server.config.C)
	// one block less than targetBlkNum is also acceptable for fastforward
	for blkNum := self.server.GetCurrentBlockNo(); blkNum < targetBlkNum; blkNum++ {
		if len(self.server.msgPool.GetProposalMsgs(blkNum)) == 0 {
			log.Infof("server %d check fastforward false, no proposal for block %d",
				self.server.Index, blkNum)
			return false
		}
		cMsgs := self.server.msgPool.GetCommitMsgs(blkNum)
		if len(cMsgs) <= C {
			log.Infof("server %d check fastforward false, only %d commit msg for block %d",
				self.server.Index, len(cMsgs), blkNum)
			return false
		}
	}

	if self.server.syncer.isActive() {
		return false
	}

	return true
}
