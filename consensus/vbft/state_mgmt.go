package vbft

import (
	"time"
)

const (
	// TODO: move to config file
	maxPeerConnections    = 100
	maxSyncingCheckBlkNum = 10
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

	liveTicker          *time.Timer
	lastTickChainHeight uint64
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
	self.liveTicker = time.AfterFunc(peerHandshakeTimeout, func() {
		self.StateEventC <- &StateEvent{
			Type:     LiveTick,
			blockNum: self.server.GetCommittedBlockNo(),
		}
		self.liveTicker.Reset(peerHandshakeTimeout)
	})

	// wait config done
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
						self.server.log.Infof("server %d, start syncing", self.server.Index)
						self.currentState = Syncing
					} else if v > self.server.config.View {
						// update ChainConfig
						self.server.log.Errorf("todo: chain config changed, need update chain config from peers")
						// TODO: fetch config from neighbours, update chain config
						self.currentState = LocalConfigured
					}
				}
			case UpdatePeerState:
				if evt.peerState.connected {
					if err := self.onPeerUpdate(evt.peerState); err != nil {
						self.server.log.Errorf("statemgr process peer (%d) err: %s", evt.peerState.peerIdx, err)
					}
				} else {
					if err := self.onPeerDisconnected(evt.peerState.peerIdx); err != nil {
						self.server.log.Errorf("statmgr process peer (%d) disconn err: %s", evt.peerState.peerIdx, err)
					}
				}

			case SyncDone:
				self.server.log.Infof("server %d sync done, curr blkNum: %d", self.server.Index, self.server.GetCurrentBlockNo())
				self.setSyncedReady()

			case LiveTick:
				if err := self.onLiveTick(evt); err != nil {
					self.server.log.Errorf("server %d, live ticker: %s", self.server.Index, err)
				}
			}
		}
	}
}

func (self *StateMgr) onPeerUpdate(peerState *PeerState) error {
	peerIdx := peerState.peerIdx
	newPeer := false
	if _, present := self.peers[peerIdx]; !present {
		newPeer = true
	}

	self.server.log.Infof("server %d peer update, current blk %d, state %d, from peer %d, committed %d",
		self.server.Index, self.server.GetCurrentBlockNo(), self.currentState, peerState.peerIdx, peerState.committedBlockNum)

	// update peer state
	self.peers[peerIdx] = peerState

	if !newPeer {
		if isActive(self.currentState) && peerState.committedBlockNum > self.server.GetCurrentBlockNo()+maxSyncingCheckBlkNum {
			self.server.log.Warnf("server %d seems lost sync: %d(%d) vs %d", self.server.Index,
				peerState.committedBlockNum, peerState.peerIdx, self.server.GetCurrentBlockNo())
			if err := self.checkStartSyncing(self.server.GetCommittedBlockNo() + maxSyncingCheckBlkNum); err != nil {
				self.server.log.Errorf("server %d start syncing check failed", self.server.Index)
			}
			return nil
		}
	}

	switch self.currentState {
	case LocalConfigured:
		v := self.getSyncedChainConfigView()
		self.server.log.Infof("server %d statemgr update, current state: %d, from peer: %d, peercnt: %d, v1: %d, v2: %d",
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
				self.server.log.Infof("server %d, syncing %d, target %d, fastforward %t",
					self.server.Index, self.server.GetCommittedBlockNo(), committedBlkNum, fastforward)
				if fastforward {
					self.server.makeFastForward()
				} else {
					self.checkStartSyncing(self.server.GetCommittedBlockNo())
				}
			}
		}
		if self.isSyncedReady() {
			self.server.log.Infof("server %d synced from syncing", self.server.Index)
			self.setSyncedReady()
		}
	case WaitNetworkReady:
		if self.isSyncedReady() {
			self.server.log.Infof("server %d synced from sync-ready", self.server.Index)
			self.setSyncedReady()
		}
	case SyncReady:
	case Synced:
		committedBlkNum, ok := self.getConsensusedCommittedBlockNum()
		if ok && committedBlkNum > self.server.GetCommittedBlockNo()+1 {
			self.server.makeFastForward()
			self.server.log.Infof("server %d synced try fastforward from %d",
				self.server.Index, self.server.GetCommittedBlockNo())
		}
	case SyncingCheck:
		if self.isSyncedReady() {
			self.setSyncedReady()
		} else {
			self.checkStartSyncing(self.server.GetCommittedBlockNo() + maxSyncingCheckBlkNum)
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

	self.server.log.Infof("server %d detected consensus halt %d",
		self.server.Index, self.server.GetCurrentBlockNo())

	return self.server.reBroadcastCurrentRoundMsgs()
}

func (self *StateMgr) getMinActivePeerCount() int {
	n := int(self.server.config.F) * 2 // plus self
	if n > maxPeerConnections {
		// FIXME: F vs. maxConnections
		return maxPeerConnections
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
		self.server.log.Infof("server %d start sync ready", self.server.Index)
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

func (self *StateMgr) checkStartSyncing(startBlkNum uint64) error {

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
			if len(peers[n]) > int(self.server.config.F) {
				maxCommitted = n
			}
		}
	}

	if maxCommitted > startBlkNum {
		self.currentState = Syncing
		startBlkNum = self.server.GetCommittedBlockNo() + 1
		self.server.log.Infof("server %d, start syncing %d - %d, with %v", self.server.Index, startBlkNum, maxCommitted, peers)
		self.server.syncer.blockSyncReqC <- &BlockSyncReq{
			targetPeers:    peers[maxCommitted],
			startBlockNum:  startBlkNum,
			targetBlockNum: maxCommitted,
		}
	} else if self.currentState == Synced {
		self.server.log.Infof("server %d, start syncing check %v, %d", self.server.Index, peers, self.server.GetCurrentBlockNo())
		self.currentState = SyncingCheck
	}

	return nil
}

func (self *Server) restartSyncing() {

	// send sync request to self.sync, go syncing-state immediately
	// stop all bft timers

	self.log.Errorf("todo: server %d restart syncing", self.Index)

}

// return 0 if consensus not reached yet
func (self *StateMgr) getConsensusedCommittedBlockNum() (uint64, bool) {
	F := int(self.server.config.F)

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
			if len(peers[n]) > F {
				maxCommitted = n
				consensused = true
			}
		}
	}

	return maxCommitted, consensused
}

func (self *StateMgr) canFastForward(targetBlkNum uint64) bool {
	if self.getState() != Syncing {
		// fastword check only support syncing state
		return false
	}

	if targetBlkNum > self.server.GetCommittedBlockNo()+maxSyncingCheckBlkNum*8 {
		return false
	}

	F := int(self.server.config.F)
	// one block less than targetBlkNum is also acceptable for fastforward
	for blkNum := self.server.GetCurrentBlockNo(); blkNum < targetBlkNum; blkNum++ {
		if len(self.server.msgPool.GetProposalMsgs(blkNum)) == 0 {
			self.server.log.Info("server %d check fastforward false, no proposal for block %d",
				self.server.Index, blkNum)
			return false
		}
		if len(self.server.msgPool.GetCommitMsgs(blkNum)) <= F {
			self.server.log.Info("server %d check fastforward false, no commit msg for block %d",
				self.server.Index, blkNum)
			return false
		}
	}

	if self.server.syncer.isActive() {
		return false
	}

	return true
}
