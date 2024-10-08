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
	"bytes"
	"fmt"
	"math"
	"reflect"
	"sync"
	"time"

	"github.com/ontio/ontology-crypto/vrf"
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	actorTypes "github.com/ontio/ontology/consensus/actor"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/core/utils"
	"github.com/ontio/ontology/events"
	"github.com/ontio/ontology/events/message"
	p2pmsg "github.com/ontio/ontology/p2pserver/message/types"
	p2p "github.com/ontio/ontology/p2pserver/net/protocol"
	gover "github.com/ontio/ontology/smartcontract/service/native/governance"
	ninit "github.com/ontio/ontology/smartcontract/service/native/init"
	nutils "github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/validator/increment"
)

type BftActionType uint8

const (
	MakeProposal BftActionType = iota
	EndorseBlock
	CommitBlock
	SealBlock
	FastForward // for syncer catch up
	ReBroadcast
	SubmitBlock
)

const (
	CAP_MESSAGE_CHANNEL  = 4096
	CAP_ACTION_CHANNEL   = 64
	CAP_MSG_SEND_CHANNEL = 16
)

type BftAction struct {
	Type     BftActionType
	BlockNum uint32
	Proposal *blockProposalMsg
	forEmpty bool
}

type BlockParticipantConfig struct {
	BlockNum    uint32
	ChainConfig *vconfig.ChainConfig
	Proposers   []uint32
	Endorsers   []uint32
	Committers  []uint32
}

type p2pMsgPayload struct {
	fromPeer uint32
	Data     []byte
}

type Server struct {
	Index         uint32
	account       *account.Account
	poolActor     *actorTypes.TxPoolActor
	p2p           p2p.P2P
	incrValidator *increment.IncrementValidator
	pid           *actor.PID

	//
	// Note:
	// 1. locking priority: metaLock > blockpool.Lock > peerpool.Lock
	// 2. should never take exclusive lock on both blockpool and peerpool at the same time.
	// 3. msgpool.Lock is independent, should have no exclusive overlap with other locks.
	//
	metaLock                 sync.RWMutex
	completedBlockNum        uint32 // ledger SaveBlockCompleted block num
	currentBlockNum          uint32
	LastConfigBlockNum       uint32
	dealfutureBlockNum       uint32
	config                   *vconfig.ChainConfig
	currentParticipantConfig *BlockParticipantConfig

	msgPool   *MsgPool   // consensus msg pool
	blockPool *BlockPool // received block proposals
	peerPool  *PeerPool  // consensus peers
	syncer    *Syncer
	stateMgr  *StateMgr
	timer     *EventTimer

	msgRecvC   *sync.Map // map[uint32]chan *p2pMsgPayload
	msgC       chan ConsensusMsg
	bftActionC chan *BftAction
	msgSendC   chan *SendMsgEvent
	sub        *events.ActorSubscriber
	quitC      chan struct{}
	quitWg     sync.WaitGroup
}

func NewVbftServer(account *account.Account, txpool *actor.PID, p2p p2p.P2P) (*Server, error) {
	server := &Server{
		account:       account,
		poolActor:     &actorTypes.TxPoolActor{Pool: txpool},
		p2p:           p2p,
		incrValidator: increment.NewIncrementValidator(20),
	}

	if err := server.initialize(); err != nil {
		return nil, fmt.Errorf("vbft server start failed: %s", err)
	}

	props := actor.FromProducer(func() actor.Actor {
		return server
	})

	pid, err := actor.SpawnNamed(props, "consensus_vbft")
	if err != nil {
		return nil, err
	}
	server.pid = pid
	server.sub = events.NewActorSubscriber(pid)
	server.sub.Subscribe(message.TOPIC_SAVE_BLOCK_COMPLETE)

	return server, nil
}

func (self *Server) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Restarting:
		log.Info("vbft actor restarting")
	case *actor.Stopping:
		log.Info("vbft actor stopping")
	case *actor.Stopped:
		log.Info("vbft actor stopped")
	case *actor.Started:
		log.Info("vbft actor started")
	case *actor.Restart:
		log.Info("vbft actor restart")
	case *actorTypes.StartConsensus:
		log.Info("vbft actor start consensus")
	case *actorTypes.StopConsensus:
		self.stop()
	case *message.SaveBlockCompleteMsg:
		log.Infof("vbft actor SaveBlockCompleteMsg receives block complete event. block height=%d, numtx=%d",
			msg.Block.Header.Height, len(msg.Block.Transactions))
		self.handleBlockPersistCompleted(msg.Block)
	case *message.BlockConsensusComplete:
		log.Infof("vbft actor  BlockConsensusComplete receives block complete event. block height=%d, numtx=%d",
			msg.Block.Header.Height, len(msg.Block.Transactions))
		self.handleBlockPersistCompleted(msg.Block)
	case *p2pmsg.ConsensusPayload:
		self.NewConsensusPayload(msg)

	default:
		log.Info("vbft actor: Unknown msg ", msg, "type", reflect.TypeOf(msg))
	}
}

func (self *Server) GetPID() *actor.PID {
	return self.pid
}

func (self *Server) Start() error {
	return self.start()
}

func (self *Server) Halt() error {
	self.pid.Tell(&actorTypes.StopConsensus{})
	return nil
}

func (self *Server) handleBlockPersistCompleted(block *types.Block) {
	log.Infof("persist block: %d, %x", block.Header.Height, block.Hash())

	if block.Header.Height <= self.GetCompletedBlockNum() {
		log.Infof("server %d, persist block %d, vs completed %d",
			self.Index, block.Header.Height, self.GetCompletedBlockNum())
		return
	}
	completedBlock := block.Header.Height
	self.SetCompletedBlockNum(completedBlock)
	self.incrValidator.AddBlock(block)
	if self.nonConsensusNode() {
		self.blockPool.ReloadFromLedger()
		if self.GetCommittedBlockNo() >= self.GetCurrentBlockNo() {
			self.SetCurrentBlockNo(self.GetCommittedBlockNo() + 1)
		}
	}

	if self.checkNeedUpdateChainConfig(completedBlock) || self.checkUpdateChainConfig(completedBlock) {
		err := self.updateChainConfig(completedBlock)
		if err != nil {
			log.Errorf("updateChainConfig failed:%s", err)
		}
	}
}

func (self *Server) CheckSubmitBlock(blkNum uint32, stateRoot common.Uint256) bool {
	cMsgs := self.msgPool.GetBlockSubmitMsgs(blkNum)
	var stateRootCnt uint32
	for _, msg := range cMsgs {
		c := msg.(*blockSubmitMsg)
		if c.BlockStateRoot == stateRoot {
			stateRootCnt++
		} else {
			continue
		}
	}

	cfg := self.GetChainConfig()
	m := cfg.N - (cfg.N-1)/3

	if stateRootCnt < uint32(m) {
		return false
	}
	return true
}

func (self *Server) NewConsensusPayload(payload *p2pmsg.ConsensusPayload) {
	peerID := vconfig.PubkeyID(payload.Owner)
	peerIdx, present := self.peerPool.GetPeerIndex(peerID)
	if !present {
		log.Debugf("invalid consensus node: %s", peerID)
		return
	}
	if !self.peerPool.IsPeerConnected(peerIdx) {
		self.peerPool.OnPeerConnected(peerIdx)
	}
	p2pid, present := self.peerPool.GetP2pId(peerIdx)
	if !present || p2pid != payload.PeerId {
		self.peerPool.AddP2pId(peerIdx, payload.PeerId)
	}

	if C := self.GetPeerMsgChan(peerIdx); C != nil {
		C <- &p2pMsgPayload{
			fromPeer: peerIdx,
			Data:     payload.Data,
		}
	} else {
		log.Errorf("consensus msg without receiver: %d node: %s", peerIdx, peerID)
		return
	}
}

func (self *Server) LoadChainConfig(store *ChainStore) error {
	blkNum := store.GetChainedBlockNum()
	block, _ := store.GetBlock(blkNum)
	if block == nil {
		return fmt.Errorf("getSealedBlock err height:%d", blkNum)
	}
	var cfg vconfig.ChainConfig
	if block.getNewChainConfig() != nil {
		cfg = *block.getNewChainConfig()
		self.LastConfigBlockNum = blkNum
	} else {
		lastConfigNum := block.getLastConfigBlockNum()
		cfgBlock, _ := store.GetBlock(lastConfigNum)
		if cfgBlock == nil {
			return fmt.Errorf("failed to get cfg block height:%d", lastConfigNum)
		}
		if cfgBlock.getNewChainConfig() == nil {
			panic("failed to get chain config from config block")
		}
		cfg = *cfgBlock.getNewChainConfig()
		self.LastConfigBlockNum = lastConfigNum
	}
	self.config = &cfg

	if self.config.View == 0 || self.config.MaxBlockChangeView == 0 {
		panic("invalid view or maxblockchangeview ")
	}

	// update timer params
	self.updateTimerParams(self.config)

	self.completedBlockNum = blkNum
	self.currentBlockNum = blkNum + 1

	log.Infof("current committed block no: %d", blkNum)

	pcfg := buildParticipantConfig(self.GetCurrentBlockNo(), block.Info.Proposer, block.Info.VrfValue, self.config)
	self.currentParticipantConfig = pcfg
	log.Infof("server %d, blkNum: %d, state: %d, participants config: %v, %v, %v", self.Index, blkNum,
		self.getState(), pcfg.Proposers, pcfg.Endorsers, pcfg.Committers)

	return nil
}

func (self *Server) nonConsensusNode() bool {
	return self.Index == math.MaxUint32
}

// updateChainCofig
func (self *Server) updateChainConfig(completedBlock uint32) error {
	block, _ := self.blockPool.getSealedBlock(completedBlock)
	if block == nil {
		return fmt.Errorf("GetBlockInfo failed,block is nil:%d", completedBlock)
	}
	config := block.Info.NewChainConfig
	if config == nil {
		return fmt.Errorf("GetNewChainConfig nil,%d", completedBlock)
	}

	pubkey := vconfig.PubkeyID(self.account.PublicKey)
	peermap := make(map[string]uint32)
	for _, p := range config.Peers {
		peermap[p.ID] = p.Index
		if self.Index == math.MaxUint32 && pubkey == p.ID {
			self.Index = p.Index
			log.Infof("updateChainConfig add index :%d", self.Index)
		}
		// check if peer pubkey support VRF
		publickey, err := vconfig.Pubkey(p.ID)
		if err != nil {
			return fmt.Errorf("failed to parse peer %d PeerID: %s", p.Index, err)
		} else if !vrf.ValidatePublicKey(publickey) {
			return fmt.Errorf("peer %d: invalid peer pubkey for VRF", p.Index)
		}
	}

	log.Infof("updateChainConfig blkNum:%d", completedBlock)
	self.metaLock.Lock()
	self.config = config
	self.LastConfigBlockNum = block.getLastConfigBlockNum()
	self.metaLock.Unlock()

	self.metaLock.RLock()
	defer self.metaLock.RUnlock()

	// . update timer param
	// . update peer pool
	// . remove nonparticipation consensus node
	// . update statemgr peers
	// . reset remove peer connections, create new connections with new peers
	self.updateTimerParams(config)

	added, removed := self.peerPool.ResetNewConsuensusPeers(peermap)
	for _, peerIdx := range added {
		self.CreatePeerMsgChan(peerIdx)
		go func() {
			if err := self.run(peerIdx); err != nil {
				log.Errorf("server %d, processor on peer %d failed: %s",
					self.Index, peerIdx, err)
			}
		}()
		log.Infof("updateChainConfig add peer index:%v", peerIdx)
	}

	for _, index := range removed {
		if index == self.Index {
			self.Index = math.MaxUint32
			log.Infof("updateChainConfig remove index :%d", index)
		} else {
			if C := self.GetPeerMsgChan(index); C != nil {
				log.Infof("updateChainConfig remove consensus:index:%d", index)
				C <- nil
			}
		}
	}

	return nil
}

func (self *Server) initialize() error {
	self.dealfutureBlockNum = 0
	selfNodeId := vconfig.PubkeyID(self.account.PublicKey)
	log.Infof("server: %s starting", selfNodeId)

	store, err := OpenBlockStore(ledger.DefLedger, func(block *types.Block) {
		if self.pid != nil {
			self.pid.Tell(&message.BlockConsensusComplete{Block: block})
		}
	})
	if err != nil {
		log.Errorf("failed to open block store: %s", err)
		return fmt.Errorf("failed to open block store: %s", err)
	}
	log.Info("block store opened")

	var msgHistoryDuration uint32 = 64
	self.msgPool = newMsgPool(self, msgHistoryDuration)
	self.timer = NewEventTimer(self)
	self.syncer = newSyncer(self)
	self.stateMgr = newStateMgr(self)

	self.msgRecvC = new(sync.Map)
	self.msgC = make(chan ConsensusMsg, CAP_MESSAGE_CHANNEL)
	self.bftActionC = make(chan *BftAction, CAP_ACTION_CHANNEL)
	self.msgSendC = make(chan *SendMsgEvent, CAP_MSG_SEND_CHANNEL)
	self.quitC = make(chan struct{})
	if err := self.LoadChainConfig(store); err != nil {
		log.Errorf("failed to load config: %s", err)
		return fmt.Errorf("failed to load config: %s", err)
	}
	log.Infof("chain config loaded from local, current blockNum: %d", self.GetCurrentBlockNo())

	// add all consensus peers to peer_pool
	peermap := make(map[string]uint32)
	for _, p := range self.GetChainConfig().Peers {
		peermap[p.ID] = p.Index
		// check if peer pubkey support VRF
		if pk, err := vconfig.Pubkey(p.ID); err != nil {
			return fmt.Errorf("failed to parse peer %d PeerID: %s", p.Index, err)
		} else if !vrf.ValidatePublicKey(pk) {
			return fmt.Errorf("peer %d: invalid peer pubkey for VRF", p.Index)
		}
		log.Infof("added peer: %s", p.ID)
	}
	self.peerPool = NewPeerPool(peermap)

	self.blockPool, err = newBlockPool(self, msgHistoryDuration, store)
	if err != nil {
		log.Errorf("init blockpool: %s", err)
		return fmt.Errorf("init blockpool: %s", err)
	}

	//index equal math.MaxUint32  is noconsensus node
	id := vconfig.PubkeyID(self.account.PublicKey)
	index, present := self.peerPool.GetPeerIndex(id)
	if present {
		self.Index = index
	} else {
		self.Index = math.MaxUint32
	}
	go self.syncer.run()
	go self.stateMgr.run()
	go self.msgSendLoop()
	go self.timerLoop()
	go self.actionLoop()
	go self.processMsgLoop()

	log.Infof("peer %d started", self.Index)

	return nil
}

func (self *Server) start() error {
	// check if server pubkey support VRF
	if !vrf.ValidatePrivateKey(self.account.PrivateKey) || !vrf.ValidatePublicKey(self.account.PublicKey) {
		return fmt.Errorf("server %d consensus start failed: invalid account key for VRF", self.Index)
	}

	// start heartbeat ticker
	self.timer.startPeerTicker()

	// start peers msg handlers
	for _, p := range self.GetChainConfig().Peers {
		peerIdx := p.Index
		self.CreatePeerMsgChan(peerIdx)

		go func() {
			if err := self.run(peerIdx); err != nil {
				log.Errorf("server %d, processor on peer %d failed: %s", self.Index, peerIdx, err)
			}
		}()
	}

	return nil
}

func (self *Server) stop() {

	self.incrValidator.Clean()
	self.sub.Unsubscribe(message.TOPIC_SAVE_BLOCK_COMPLETE)
	// stop syncer, statemgr, msgSendLoop, timer, actionLoop, msgProcessingLoop
	close(self.quitC)
	self.quitWg.Wait()

	self.syncer.stop()
	self.timer.stop()
	self.msgPool.clean()
	self.blockPool.clean()
	self.peerPool.clean()
}

// go routine per net connection
func (self *Server) run(peerIdx uint32) error {
	// broadcast heartbeat
	self.heartbeat(math.MaxUint32)

	defer func() {
		// TODO: handle peer disconnection here
		log.Warnf("server %d: disconnected with peer %d", self.Index, peerIdx)
		self.ClosePeerMsgChan(peerIdx)

		self.peerPool.OnPeerDisconnected(peerIdx)
		self.stateMgr.StateEventC <- &StateEvent{
			Type: UpdatePeerState,
			peerState: &PeerState{
				peerIdx:   peerIdx,
				connected: false,
			},
		}
	}()

	for {
		fromPeer, msgData, err := self.receiveFromPeer(peerIdx)
		if err != nil {
			return err
		}
		msg, err := DeserializeVbftMsg(msgData)
		if err != nil {
			log.Errorf("server %d failed to deserialize vbft msg (len %d): %s", self.Index, len(msgData), err)
			continue
		}

		pk := self.peerPool.GetPeerPubKey(fromPeer)
		if pk == nil {
			log.Errorf("server %d failed to get peer %d pubkey", self.Index, fromPeer)
			continue
		}

		if msg.Type() == BlockProposalMessage {
			if proposal := msg.(*blockProposalMsg); proposal != nil {
				fromPeer = proposal.Block.getProposer()
				pk = self.peerPool.GetPeerPubKey(proposal.Block.getProposer())
			}
		} else if msg.Type() == BlockEndorseMessage {
			if endorseMsg := msg.(*blockEndorseMsg); endorseMsg != nil {
				proposal := self.findBlockProposal(msg.GetBlockNum(), endorseMsg.EndorsedProposer)
				if proposal != nil {
					if endorseMsg.EndorseForEmpty {
						if proposal.Block.EmptyBlock.Hash() != endorseMsg.EndorsedBlockHash {
							log.Errorf("server %d failed to compare blkHash, type %d,blk:%d,emptyBlockHash:%x,endorsedBlockHash:%x,proposer:%d,from:%d",
								self.Index, msg.Type(), msg.GetBlockNum(), proposal.Block.EmptyBlock.Hash(), endorseMsg.EndorsedBlockHash, endorseMsg.EndorsedProposer, fromPeer)
							continue
						}
					} else {
						if proposal.Block.Block.Hash() != endorseMsg.EndorsedBlockHash {
							log.Errorf("server %d failed to compare blkHash, type %d,blk:%d,blockHash:%x,endorsedBlockHash:%x,proposer:%d,from:%d",
								self.Index, msg.Type(), msg.GetBlockNum(), proposal.Block.Block.Hash(), endorseMsg.EndorsedBlockHash, endorseMsg.EndorsedProposer, fromPeer)
							continue
						}
					}
				}
			}
		} else if msg.Type() == BlockCommitMessage {
			if commitMsg := msg.(*blockCommitMsg); commitMsg != nil {
				proposal := self.findBlockProposal(msg.GetBlockNum(), commitMsg.BlockProposer)
				if proposal != nil {
					if commitMsg.CommitForEmpty {
						if proposal.Block.EmptyBlock.Hash() != commitMsg.CommitBlockHash {
							log.Errorf("server %d failed to compare blkHash, type %d,blk:%d,emptyBlockHash:%x,CommitBlockHash:%x,proposer:%d,from:%d",
								self.Index, msg.Type(), msg.GetBlockNum(), proposal.Block.EmptyBlock.Hash(), commitMsg.CommitBlockHash, commitMsg.BlockProposer, fromPeer)
							continue
						}
					} else {
						if proposal.Block.Block.Hash() != commitMsg.CommitBlockHash {
							log.Errorf("server %d failed to compare blkHash, type %d,blk:%d,blockHash:%x,CommitBlockHash:%x,proposer:%d,from:%d",
								self.Index, msg.Type(), msg.GetBlockNum(), proposal.Block.Block.Hash(), commitMsg.CommitBlockHash, commitMsg.BlockProposer, fromPeer)
							continue
						}
					}
				}
			}
		}

		if err := msg.Verify(pk, self.peerPool.GetAllPubKeys()); err != nil {
			log.Errorf("server %d failed to verify msg, type %d, err: %s",
				self.Index, msg.Type(), err)
			continue
		}

		if msg.Type() <= BlockCommitMessage || msg.Type() == BlockSubmitMessage {
			log.Infof("server %d received consensus msg, blk %d, type: %d from %d",
				self.Index, msg.GetBlockNum(), msg.Type(), fromPeer)
		}

		self.onConsensusMsg(fromPeer, msg, hashData(msgData))
	}
}

func (self *Server) getState() ServerState {
	return self.stateMgr.getState()
}

func (self *Server) updateParticipantConfig() error {
	blkNum := self.GetCurrentBlockNo()
	block, _ := self.blockPool.getSealedBlock(blkNum - 1)
	if block == nil {
		return fmt.Errorf("failed to get sealed block (%d)", blkNum-1)
	}

	chainconfig := block.Info.NewChainConfig
	if chainconfig == nil {
		chainCfg := self.GetChainConfig()
		chainconfig = &chainCfg
	}
	cfg := buildParticipantConfig(blkNum, block.Info.Proposer, block.Info.VrfValue, chainconfig)
	log.Infof("server %d, blkNum: %d, state: %d, participants config: %v, %v, %v", self.Index, blkNum,
		self.getState(), cfg.Proposers, cfg.Endorsers, cfg.Committers)

	self.metaLock.Lock()
	self.currentParticipantConfig = cfg
	self.metaLock.Unlock()

	return nil
}

func (self *Server) startNewRound() error {
	blkNum := self.GetCurrentBlockNo()

	if err := self.updateParticipantConfig(); err != nil {
		log.Errorf("startNewRound error:%s", err)
		return err
	}
	// check proposals in msgpool
	var proposal *blockProposalMsg
	for _, p := range self.msgPool.GetProposalMsgs(blkNum) {
		msg := p.(*blockProposalMsg)
		if self.isProposer(msg.Block.getProposer()) {
			// get proposal from proposer, process it
			proposal = msg
		} else {
			// add other proposals to blockpool
			if err := self.blockPool.AddBlockProposal(msg); err != nil {
				log.Errorf("starting new round, failed to add proposal from %d: %s",
					msg.Block.getProposer(), err)
			}
		}
	}

	endorses := self.msgPool.GetEndorsementsMsgs(blkNum)
	for _, e := range endorses {
		msg := e.(*blockEndorseMsg)
		self.blockPool.AddBlockEndorseMsg(msg)
	}

	commits := self.msgPool.GetCommitMsgs(blkNum)
	for _, c := range commits {
		msg := c.(*blockCommitMsg)
		if err := self.blockPool.AddBlockCommitMsg(msg); err != nil {
			log.Infof("start new round, failed to add commit, blk %d, commit for %d: %s",
				blkNum, msg.BlockProposer, err)
		}
	}

	chainCfg := self.GetChainConfig()
	if _, _, done := self.blockPool.commitDone(blkNum, chainCfg.C, chainCfg.N); done && len(commits) > 0 {
		// resend commit msg to msg-processor to restart commit-done processing
		// Note: commitDone will set Done flag in block-pool, so removed Done flag checking
		// in commit msg processing.
		self.blockPool.setCommitDone(blkNum)
		self.processConsensusMsg(commits[0])
		return nil
	} else if _, _, done := self.blockPool.endorseDone(blkNum, chainCfg.C); done && len(endorses) > 0 {
		// resend endorse msg to msg-processor to restart endorse-done processing
		self.processConsensusMsg(endorses[0])
		return nil
	} else if proposal != nil {
		self.processProposalMsg(proposal)
		return nil
	}
	self.timer.StartEventTimer(EventTxPool, blkNum)
	self.timer.StartEventTimer(EventTxBlockTimeout, blkNum)
	return nil
}

func (self *Server) startNewProposal(blkNum uint32) {
	// make proposal
	if self.isProposer(self.Index) {
		log.Infof("server %d, proposer for block %d", self.Index, blkNum)
		// FIXME: possible deadlock on channel
		self.bftActionC <- &BftAction{
			Type:     MakeProposal,
			BlockNum: blkNum,
			forEmpty: false,
		}
	} else if self.is2ndProposer(blkNum, self.Index) {
		log.Infof("server %d, 2nd proposer for block %d", self.Index, blkNum)
		self.timer.StartEventTimer(EventProposalBackoff, blkNum)
	}

	// TODO: if new round block proposal has received, go endorsing/committing directly
	self.timer.StartEventTimer(EventProposeBlockTimeout, blkNum)
}

// verify consensus messsage, then send msg to processMsgEvent
func (self *Server) onConsensusMsg(peerIdx uint32, msg ConsensusMsg, msgHash common.Uint256) {
	if self.msgPool.HasMsg(msg, msgHash) && msg.Type() != BlockCommitMessage {
		// dup msg checking
		log.Debugf("dup msg with msg type %d from %d", msg.Type(), peerIdx)
		return
	}

	if self.dealfutureBlockNum != self.GetCurrentBlockNo() {
		self.processMsg(self.GetCurrentBlockNo())
	}

	switch msg.Type() {
	case BlockProposalMessage:
		pMsg := msg.(*blockProposalMsg)

		msgBlkNum := pMsg.GetBlockNum()
		if msgBlkNum > self.GetCurrentBlockNo() {
			// for concurrency, support two active consensus round
			if err := self.msgPool.AddMsg(msg, msgHash); err != nil {
				if err != errDropFarFutureMsg {
					log.Errorf("failed to add proposal msg (%d) to pool: %s", msgBlkNum, err)
				}
				return
			}
		} else if msgBlkNum < self.GetCurrentBlockNo() {
			if msgBlkNum+MAX_SYNCING_CHECK_BLK_NUM < self.GetCommittedBlockNo() {
				log.Infof("server %d get proposal msg for block %d, from %d, current committed %d",
					self.Index, msgBlkNum, pMsg.Block.getProposer(), self.GetCommittedBlockNo())
				self.timer.C <- &TimerEvent{
					evtType:  EventPeerHeartbeat,
					blockNum: pMsg.Block.getProposer(),
				}
			}
		} else {
			if err := self.msgPool.AddMsg(msg, msgHash); err != nil {
				log.Errorf("failed to add proposal msg (%d) to pool", msgBlkNum)
				return
			}
			self.processProposalMsg(pMsg)
		}

	case BlockEndorseMessage:
		pMsg := msg.(*blockEndorseMsg)

		// TODO: verify msg

		msgBlkNum := pMsg.GetBlockNum()
		if msgBlkNum > self.GetCurrentBlockNo() {
			// for concurrency, support two active consensus round
			if err := self.msgPool.AddMsg(msg, msgHash); err != nil {
				if err != errDropFarFutureMsg {
					log.Errorf("failed to add endorse msg (%d) to pool: %s", msgBlkNum, err)
				}
				return
			}
		} else if msgBlkNum < self.GetCurrentBlockNo() {
			if msgBlkNum+MAX_SYNCING_CHECK_BLK_NUM < self.GetCommittedBlockNo() {
				log.Infof("server %d get endorse msg for block %d, from %d, current committed %d",
					self.Index, msgBlkNum, pMsg.Endorser, self.GetCommittedBlockNo())
				self.timer.C <- &TimerEvent{
					evtType:  EventPeerHeartbeat,
					blockNum: pMsg.Endorser,
				}
			}
		} else {
			// add to msg pool
			if err := self.msgPool.AddMsg(msg, msgHash); err != nil {
				log.Errorf("failed to add endorse msg (%d) to pool", msgBlkNum)
				return
			}
			self.processConsensusMsg(msg)
		}

	case BlockCommitMessage:
		pMsg := msg.(*blockCommitMsg)

		// TODO: verify msg

		msgBlkNum := pMsg.GetBlockNum()
		if msgBlkNum > self.GetCurrentBlockNo() {
			if err := self.msgPool.AddMsg(msg, msgHash); err != nil {
				if err != errDropFarFutureMsg {
					log.Errorf("failed to add commit msg (%d) to pool: %s", msgBlkNum, err)
				}
				return
			}
		} else if msgBlkNum < self.GetCurrentBlockNo() {
			if msgBlkNum+MAX_SYNCING_CHECK_BLK_NUM < self.GetCommittedBlockNo() {
				log.Infof("server %d get commit msg for block %d, from %d, current committed %d",
					self.Index, msgBlkNum, pMsg.Committer, self.GetCommittedBlockNo())
				self.timer.C <- &TimerEvent{
					evtType:  EventPeerHeartbeat,
					blockNum: pMsg.Committer,
				}
			}
		} else {
			// add to msg pool
			if err := self.msgPool.AddMsg(msg, msgHash); err != nil {
				log.Errorf("failed to add commit msg (%d) to pool", msgBlkNum)
				return
			}
			self.processConsensusMsg(msg)
		}
	case PeerHeartbeatMessage:
		pMsg := msg.(*peerHeartbeatMsg)
		self.processHeartbeatMsg(peerIdx, pMsg)
		if pMsg.CommittedBlockNumber+MAX_SYNCING_CHECK_BLK_NUM < self.GetCommittedBlockNo() {
			// delayed peer detected, response heartbeat with our chain Info
			self.timer.C <- &TimerEvent{
				evtType:  EventPeerHeartbeat,
				blockNum: peerIdx,
			}
		}

	case ProposalFetchMessage:
		pMsg := msg.(*proposalFetchMsg)
		var pmsg *blockProposalMsg
		if self.Index == pMsg.ProposerID || pMsg.BlockNum == self.GetCurrentBlockNo() {
			pMsgs := self.msgPool.GetProposalMsgs(pMsg.BlockNum)
			for _, msg := range pMsgs {
				p := msg.(*blockProposalMsg)
				if p != nil && p.Block.getProposer() == pMsg.ProposerID {
					log.Infof("server %d rebroadcast proposal to %d, blk %d",
						self.Index, peerIdx, p.Block.getBlockNum())
					pmsg = p
				}
			}
		}
		if self.Index == pMsg.ProposerID {
			if pmsg == nil {
				blk, _ := self.blockPool.getSealedBlock(pMsg.BlockNum)
				if blk != nil {
					pmsg = &blockProposalMsg{
						Block: blk,
					}
				}
			}
		}

		if pmsg != nil {
			log.Infof("server %d, handle proposal fetch %d from %d",
				self.Index, pMsg.BlockNum, peerIdx)
			self.msgSendC <- &SendMsgEvent{
				ToPeer: peerIdx,
				Msg:    pmsg,
			}
		}

	case BlockFetchMessage:
		// handle block fetch msg
		pMsg := msg.(*blockFetchMsg)
		blk, blkHash := self.blockPool.getChainedBlock(pMsg.BlockNum)
		if blk == nil {
			return
		}
		msg := self.constructBlockFetchRespMsg(pMsg.BlockNum, blk, blkHash)
		log.Infof("server %d, handle blockfetch %d from %d",
			self.Index, pMsg.BlockNum, peerIdx)
		self.msgSendC <- &SendMsgEvent{
			ToPeer: peerIdx,
			Msg:    msg,
		}

	case BlockFetchRespMessage:
		self.syncer.syncMsgC <- &SyncMsg{
			fromPeer: peerIdx,
			msg:      msg.(*BlockFetchRespMsg),
		}
	case BlockSubmitMessage:
		pMsg := msg.(*blockSubmitMsg)
		msgBlkNum := pMsg.GetBlockNum()
		if self.GetCurrentBlockNo() > msgBlkNum+1 {
			return
		}
		if err := self.msgPool.AddMsg(msg, msgHash); err != nil {
			if err != errDropFarFutureMsg {
				log.Errorf("failed to add submit msg (%d) to pool: %s", msgBlkNum, err)
			}
			return
		}
		if self.CheckSubmitBlock(msgBlkNum, pMsg.BlockStateRoot) {
			self.makeBlockSubmit(msgBlkNum)
		}
	}
}

func (self *Server) verifyCrossChainMsg(msg *blockProposalMsg) bool {
	root, err := self.blockPool.getCrossStatesRoot(msg.Block.Block.Header.Height - 1)
	if err != nil {
		log.Errorf("verifyCrossChainMsg:%s", err)
		return false
	}
	//malicious consensus node may create a nil cross chain msg proposal, but it is actual not nil.
	if root != common.UINT256_EMPTY && msg.Block.CrossChainMsg == nil {
		return false
	}
	if msg.Block.CrossChainMsg == nil {
		return true
	}
	if msg.Block.CrossChainMsg.StatesRoot != root ||
		msg.Block.CrossChainMsg.Version != types.CURR_CROSS_STATES_VERSION {
		return false
	}
	return true
}

func (self *Server) processMsg(blockNum uint32) {
	consensusMsgs := self.msgPool.DropBftMsgs(blockNum)
	for _, msg := range consensusMsgs {
		if msg.Type() == BlockCommitMessage {
			if commit := msg.(*blockCommitMsg); commit != nil {
				pk := self.peerPool.GetPeerPubKey(commit.Committer)
				if pk == nil {
					log.Errorf("processMsg GetPeerPubKey is nil,committer:%d,blk:%d", commit.Committer, blockNum)
					continue
				}
				if err := msg.Verify(pk, self.peerPool.GetAllPubKeys()); err != nil {
					log.Errorf("server:%d  processMsg failed to verify commit msg, type %d, err: %s",
						self.Index, msg.Type(), err)
					continue
				}
			}
		}
		if err := self.msgPool.AddMsg(msg, MustHashMsg(msg)); err != nil {
			log.Errorf("processMsg failed to add commit msg blk:%d to pool,err:%s", blockNum, err)
			continue
		}
		self.processConsensusMsg(msg)
	}
	self.dealfutureBlockNum = blockNum
}

func (self *Server) processProposalMsg(msg *blockProposalMsg) {
	msgBlkNum := msg.GetBlockNum()
	blk, prevBlkHash := self.blockPool.getSealedBlock(msg.GetBlockNum() - 1)
	if blk == nil {
		log.Errorf("BlockProposal failed to GetPreBlock:%d", msg.GetBlockNum()-1)
		return
	}

	msgPrevBlkHash := msg.Block.getPrevBlockHash()
	if prevBlkHash != msgPrevBlkHash {
		log.Errorf("BlockPrposalMessage check blocknum:%d,prevhash:%s,msg prevhash:%s", msg.GetBlockNum(), prevBlkHash.ToHexString(), msgPrevBlkHash.ToHexString())
		self.msgPool.DropMsg(msg)
		return
	}
	if self.LastConfigBlockNum != math.MaxUint32 && blk.Info.LastConfigBlockNum != self.LastConfigBlockNum {
		log.Errorf("BlockPrposalMessage  check LastConfigBlockNum blocknum:%d,prvLastConfigBlockNum:%d,self LastConfigBlockNum:%d", msg.GetBlockNum(), blk.Info.LastConfigBlockNum, self.LastConfigBlockNum)
		return
	}
	merkleRoot, err := self.blockPool.getExecMerkleRoot(msgBlkNum - 1)
	if err != nil {
		log.Errorf("failed to GetExecMerkleRoot: %s,blkNum:%d", err, msgBlkNum-1)
		return
	}
	if msg.Block.getPrevExecMerkleRoot() != merkleRoot {
		self.msgPool.DropMsg(msg)
		msgMerkleRoot := msg.Block.getPrevExecMerkleRoot()
		log.Errorf("BlockPrposalMessage check MerkleRoot blocknum:%d,msg MerkleRoot:%s,self MerkleRoot:%s", msg.GetBlockNum(), msgMerkleRoot.ToHexString(), merkleRoot.ToHexString())
		return
	}
	cfg := vconfig.ChainConfig{}
	if blk.getNewChainConfig() != nil {
		cfg = *blk.getNewChainConfig()
		chainCfg := self.GetChainConfig()
		if cfg.Hash() != chainCfg.Hash() {
			log.Errorf("processProposalMsg chainconfig unqeual to blockinfo cfg,view:(%d,%d),N:(%d,%d),C:(%d,%d),BlockMsgDelay:(%d,%d),HashMsgDelay:(%d,%d),PeerHandshakeTimeout:(%d,%d),posTable:(%v,%v),MaxBlockChangeView:(%d,%d)",
				cfg.View, chainCfg.View,
				cfg.N, chainCfg.N,
				cfg.C, chainCfg.C,
				cfg.BlockMsgDelay, chainCfg.BlockMsgDelay,
				cfg.HashMsgDelay, chainCfg.HashMsgDelay,
				cfg.PeerHandshakeTimeout, chainCfg.PeerHandshakeTimeout,
				cfg.PosTable, chainCfg.PosTable,
				cfg.MaxBlockChangeView, chainCfg.MaxBlockChangeView)
			self.msgPool.DropMsg(msg)
			return
		}
	}

	prevBlockTimestamp := blk.Block.Header.Timestamp
	currentBlockTimestamp := msg.Block.Block.Header.Timestamp
	if currentBlockTimestamp <= prevBlockTimestamp || currentBlockTimestamp > uint32(time.Now().Add(time.Minute*10).Unix()) {
		log.Errorf("BlockPrposalMessage check  blocknum:%d,prevBlockTimestamp:%d,currentBlockTimestamp:%d", msg.GetBlockNum(), prevBlockTimestamp, currentBlockTimestamp)
		self.msgPool.DropMsg(msg)
		return
	}

	// verify VRF
	proposerPk := self.peerPool.GetPeerPubKey(msg.Block.getProposer())
	if proposerPk == nil {
		log.Errorf("server %d failed to get proposer %d pk of block %d",
			self.Index, msg.Block.getProposer(), msgBlkNum)
		self.msgPool.DropMsg(msg)
		return
	}
	if err := verifyVrf(proposerPk, msgBlkNum, blk.getVrfValue(), msg.Block.getVrfValue(), msg.Block.getVrfProof()); err != nil {
		log.Errorf("server %d failed to verify vrf of block %d proposal from %d",
			self.Index, msgBlkNum, msg.Block.getProposer())
		self.msgPool.DropMsg(msg)
		return
	}
	if !self.verifyCrossChainMsg(msg) {
		log.Errorf("verify cross chain message error:%+v\n", msg.Block.CrossChainMsg)
		self.msgPool.DropMsg(msg)
		return
	}
	txs := msg.Block.Block.Transactions
	if len(txs) > 0 && self.nonSystxs(txs, msgBlkNum) {
		height := msgBlkNum - 1
		start, end := self.incrValidator.BlockRange()
		if msg.GetBlockNum() <= self.GetCompletedBlockNum() {
			log.Infof("processProposalMsg failed: MsgBlockNum:%d,CompletedBlockNum:%d", msg.GetBlockNum(), self.GetCompletedBlockNum())
			return
		}
		validHeight := height
		if height+1 == end {
			validHeight = start
		} else {
			self.incrValidator.Clean()
			//log.Infof("incr validator block height %v != ledger block height %v", int(end)-1, height)
			log.Infof("incr validator block height %v != ledger block height %v, CompletedBlockNum:%d", int(end)-1, height, self.GetCompletedBlockNum())
		}
		// start new routine to verify txs in proposal block
		go func() {
			if err := self.poolActor.VerifyBlock(txs, validHeight); err != nil && err != actor.ErrTimeout {
				log.Errorf("server %d verify proposal blk from %d failed, blk %d, txs %d, err: %s",
					self.Index, msg.Block.getProposer(), msgBlkNum, len(txs), err)
				self.msgPool.DropMsg(msg)
				return
			} else if err == actor.ErrTimeout {
				log.Errorf("server %d verify proposal blk from %d timedout, blk %d, txs %d, err: %s",
					self.Index, msg.Block.getProposer(), msgBlkNum, len(txs), err)
			}
			nonceCtx := make(map[common.Address]uint64)
			for _, tx := range txs {
				if err := self.incrValidator.Verify(tx, validHeight, nonceCtx); err != nil {
					log.Errorf("server %d verify proposal tx from %d failed, blk %d, txs %d, err: %s",
						self.Index, msg.Block.getProposer(), msgBlkNum, len(txs), err)
					self.msgPool.DropMsg(msg)
					return
				}
			}
			self.processConsensusMsg(msg)
		}()
	} else {
		// empty block, process directly
		self.processConsensusMsg(msg)
	}
}

func (self *Server) processConsensusMsg(msg ConsensusMsg) {
	if self.getState().IsReady() {
		self.msgC <- msg
	}
}

func (self *Server) processMsgLoop() {
	self.quitWg.Add(1)
	defer self.quitWg.Done()

	for {
		select {
		case msg := <-self.msgC:
			self.processMsgEvent(msg)
		case <-self.quitC:
			log.Infof("server %d, processMsgEvent loop quit", self.Index)
			return
		}
	}
}

func (self *Server) processMsgEvent(msg ConsensusMsg) {
	log.Debugf("server %d process msg, block %d, type %d, current blk %d",
		self.Index, msg.GetBlockNum(), msg.Type(), self.GetCurrentBlockNo())

	msgBlkNum := msg.GetBlockNum()
	if msgBlkNum != self.GetCurrentBlockNo() {
		return
	}
	switch msg.Type() {
	case BlockProposalMessage:
		pMsg := msg.(*blockProposalMsg)
		// add proposal to block-pool
		if err := self.blockPool.AddBlockProposal(pMsg); err != nil {
			// if err == errDupProposal {
			// 	// TODO: faulty proposer detected
			// }
			log.Errorf("failed to add block proposal (%d): %s", msgBlkNum, err)
			return
		}

		if self.isProposer(pMsg.Block.getProposer()) {
			// check if agreed on prev-blockhash
			if err := self.verifyPrevBlockHash(msgBlkNum, pMsg); err != nil {
				// continue
				log.Errorf("failed verify prevBlockHash from proposer %d, blk %d",
					pMsg.Block.getProposer(), msgBlkNum)
				return
			}

			// stop proposal timer
			self.timer.CancelEventTimer(EventProposeBlockTimeout, msgBlkNum)
			if self.isEndorser(self.Index) {
				if err := self.endorseBlock(pMsg, false); err != nil {
					log.Errorf("failed to endorse block proposal (%d): %s", msgBlkNum, err)
				}
			}
		} else {
			if self.isProposer(self.Index) {
				for _, msg := range self.msgPool.GetProposalMsgs(msgBlkNum) {
					p := msg.(*blockProposalMsg)
					if p != nil && p.Block.getProposer() == self.Index {
						log.Infof("server %d rebroadcast proposal to %d, blk %d",
							self.Index, pMsg.Block.getProposer(), msgBlkNum)
						self.broadcast(msg)
						break
					}
				}
			}
			// makeProposalTimeout handles non-leader proposals
		}
	case BlockEndorseMessage:
		pMsg := msg.(*blockEndorseMsg)

		if pMsg.EndorsedProposer != self.Index && len(self.msgPool.GetProposalMsgs(msgBlkNum)) == 0 {
			self.fetchProposal(msgBlkNum, pMsg.EndorsedProposer)
		}

		// add endorse to block-pool
		self.blockPool.AddBlockEndorseMsg(pMsg)
		log.Infof("server %d received endorse from %d, for proposer %d, block %d, empty: %t",
			self.Index, pMsg.Endorser, pMsg.EndorsedProposer, msgBlkNum, pMsg.EndorseForEmpty)

		// if had committed for current round, skip the following steps
		if self.blockPool.committedForBlock(msgBlkNum) {
			// get more endorse msg after committed, trigger seal-block-timeout
			self.timer.StartEventTimer(EventCommitBlockTimeout, msgBlkNum)
			return
		}

		if self.isEndorser(pMsg.Endorser) {
			//              if countOfEndrosement(msg.proposal) >= 2C + 1:
			//                      stop WaitEndorsementTimer
			//                      commitBlock(msg.BlockHash)
			//              else if WaitEndorsementTimer has not started:
			//                      start WaitEndorsementTimer

			// TODO: should only count endorsements from endorsers
			if proposer, forEmpty, done := self.blockPool.endorseDone(msgBlkNum, self.GetChainConfig().C); done {
				// stop endorse timer
				self.timer.CancelEventTimer(EventEndorseBlockTimeout, msgBlkNum)
				// stop empty endorse timer
				self.timer.CancelEventTimer(EventEndorseEmptyBlockTimeout, msgBlkNum)
				proposal := self.findBlockProposal(msgBlkNum, proposer)
				if proposal == nil {
					log.Infof("server %d endorse %d done, waiting proposal from %d", self.Index, msgBlkNum, proposer)
				} else if self.isCommitter(self.Index) {
					// make endorsement
					if err := self.commitBlock(proposal, forEmpty); err != nil {
						log.Errorf("failed to endorse for block %d: %s", msgBlkNum, err)
						return
					}
				}
			} // else {
			// 	// wait until endorse timeout
			// }
		} // else {
		// 	// makeEndorsementTimeout handles non-endorser endorsements
		// }
		if self.blockPool.endorseFailed(msgBlkNum, self.GetChainConfig().C) {
			// endorse failed, start empty endorsing
			self.timer.C <- &TimerEvent{
				evtType:  EventEndorseBlockTimeout,
				blockNum: msgBlkNum,
			}
		}
	case BlockCommitMessage:
		pMsg := msg.(*blockCommitMsg)
		//              if countOfCommitment(msg.proposal) >= 2C + 1:
		//                      stop WaitCommitsTimer
		//                      sealProposal(msg.BlockHash)
		//              else if WaitCommitsTimer has not started:
		//                      start WaitCommitsTimer
		if err := self.blockPool.AddBlockCommitMsg(pMsg); err != nil {
			log.Errorf("failed to add commit msg (%d): %s", msgBlkNum, err)
			return
		}

		log.Infof("server %d received commit from %d, for proposer %d, block %d, empty: %t",
			self.Index, pMsg.Committer, pMsg.BlockProposer, msgBlkNum, pMsg.CommitForEmpty)

		chainCfg := self.GetChainConfig()
		if proposer, forEmpty, done := self.blockPool.commitDone(msgBlkNum, chainCfg.C, chainCfg.N); done {
			self.blockPool.setCommitDone(msgBlkNum)
			proposal := self.findBlockProposal(msgBlkNum, proposer)
			if proposal == nil {
				// TODO: commit done, but we not have the proposal, should request proposal from neighbours
				//       commitTimeout handle this
				log.Infof("server %d commit %d done, waiting proposal",
					self.Index, msgBlkNum)
				return
			}

			if self.isCommitter(self.Index) {
				// make sure committer broadcasting his commit msg
				if err := self.commitBlock(proposal, forEmpty); err != nil {
					log.Errorf("server %d consensused %d, committer broadcast commit msg: %s", self.Index, msgBlkNum, err)
				}
			}
			if !self.blockPool.checkBlockSign(proposal.Block, forEmpty, self.config.N-(self.config.N-1)/3) {
				log.Errorf("server %d received commit checkBlockSign insufficient at blk: %d", self.Index, msgBlkNum)
				return
			}
			// stop commit timer
			self.timer.CancelEventTimer(EventCommitBlockTimeout, msgBlkNum)

			if err := self.makeSealed(proposal, forEmpty); err != nil {
				log.Errorf("failed to seal block %d, err: %s", msgBlkNum, err)
			}
		} // else {
		// 	// wait commit timeout, nothing to do
		// }
	}
}

func (self *Server) processBftAction(action *BftAction) {
	switch action.Type {
	case MakeProposal:
		// this may triggered when block sealed or random backoff of 2nd proposer
		blkNum := self.GetCurrentBlockNo()
		if blkNum > action.BlockNum {
			return
		}

		var proposal *blockProposalMsg
		msgs := self.msgPool.GetProposalMsgs(blkNum)
		for _, m := range msgs {
			if p, ok := m.(*blockProposalMsg); ok && p.Block.getProposer() == self.Index {
				proposal = p
				break
			}
		}
		if proposal == nil {
			if err := self.makeProposal(blkNum, action.forEmpty); err != nil {
				log.Errorf("server %d failed to making proposal (%d): %s",
					self.Index, blkNum, err)
			}
		}

	case EndorseBlock:
		// endorse the proposal
		blkNum := action.Proposal.GetBlockNum()
		if err := self.endorseBlock(action.Proposal, action.forEmpty); err != nil {
			log.Errorf("server %d failed to endorse block proposal (%d): %s",
				self.Index, blkNum, err)
			return
		}

	case CommitBlock:
		blkNum := action.Proposal.GetBlockNum()
		if err := self.commitBlock(action.Proposal, action.forEmpty); err != nil {
			log.Errorf("server %d failed to commit block proposal (%d): %s",
				self.Index, blkNum, err)
			return
		}
	case SealBlock:
		if action.Proposal.GetBlockNum() < self.GetCurrentBlockNo() {
			return
		}
		if err := self.sealProposal(action.Proposal, action.forEmpty); err != nil {
			log.Errorf("server %d failed to seal block (%d): %s",
				self.Index, action.Proposal.GetBlockNum(), err)
		}
	case FastForward:
		// 1. from current block num, check commit msgs in msg pool
		// 2. if commit consensused, seal the proposal
		for {
			blkNum := self.GetCurrentBlockNo()
			chainCfg := self.GetChainConfig()

			if err := self.updateParticipantConfig(); err != nil {
				log.Errorf("server %d update config failed in forwarding: %s", self.Index, err)
			}

			// get pending msgs from msgpool
			pMsgs := self.msgPool.GetProposalMsgs(blkNum)
			for _, msg := range pMsgs {
				p := msg.(*blockProposalMsg)
				if err := self.blockPool.AddBlockProposal(p); err != nil {
					log.Errorf("server %d failed add proposal in fastforwarding: %s",
						self.Index, err)
				}
			}

			cMsgs := self.msgPool.GetCommitMsgs(blkNum)
			commitMsgs := make([]*blockCommitMsg, 0)
			for _, msg := range cMsgs {
				c := msg.(*blockCommitMsg)
				if err := self.blockPool.AddBlockCommitMsg(c); err == nil {
					commitMsgs = append(commitMsgs, c)
				} else {
					log.Errorf("server %d failed to add commit in fastforwarding: %s",
						self.Index, err)
				}
			}

			log.Infof("server %d fastforwarding from %d, (%d, %d)",
				self.Index, self.GetCurrentBlockNo(), len(cMsgs), len(pMsgs))
			if len(pMsgs) == 0 && len(cMsgs) == 0 {
				log.Infof("server %d fastforward done, no msg", self.Index)
				self.startNewRound()
				break
			}

			// check if consensused
			proposer, forEmpty := getCommitConsensus(commitMsgs, int(chainCfg.C), int(chainCfg.N))
			if proposer == math.MaxUint32 {
				if err := self.catchConsensus(blkNum); err != nil {
					log.Infof("server %d fastforward done, catch consensus: %s", self.Index, err)
				}
				log.Infof("server %d fastforward done at blk %d, no consensus", self.Index, blkNum)
				break
			}

			// get consensused proposal
			var proposal *blockProposalMsg
			for _, m := range pMsgs {
				p, ok := m.(*blockProposalMsg)
				if !ok {
					continue
				}
				if p.Block.getProposer() == proposer {
					proposal = p
					break
				}
			}
			if proposal == nil {
				log.Infof("server %d fastforward stopped at blk %d, no proposal", self.Index, blkNum)
				self.fetchProposal(blkNum, proposer)
				self.timer.StartEventTimer(EventCommitBlockTimeout, blkNum)
				break
			}
			//check block sign num
			if !self.blockPool.checkBlockSign(proposal.Block, forEmpty, self.config.N-(self.config.N-1)/3) {
				log.Errorf("server %d fastforward checkBlockSign insufficient at blk: %d", self.Index, blkNum)
				break
			}

			log.Infof("server %d fastforwarding block %d, proposer %d",
				self.Index, blkNum, proposal.Block.getProposer())

			// fastforward the block
			if err := self.sealBlock(proposal.Block, forEmpty, true); err != nil {
				log.Errorf("server %d fastforward stopped at blk %d, seal failed: %s",
					self.Index, blkNum, err)
				break
			}
		}

	case ReBroadcast:
		blkNum := self.GetCurrentBlockNo()
		if blkNum > action.BlockNum {
			return
		}

		proposals := make([]*blockProposalMsg, 0)
		for _, msg := range self.msgPool.GetProposalMsgs(blkNum) {
			p := msg.(*blockProposalMsg)
			if p != nil {
				proposals = append(proposals, p)
			}
		}

		for _, p := range proposals {
			if p.Block.getProposer() == self.Index {
				log.Infof("server %d rebroadcast proposal, blk %d",
					self.Index, p.Block.getBlockNum())
				self.broadcast(p)
			}
		}
		if self.isEndorser(self.Index) {
			rebroadcasted := false
			endorseFailed := self.blockPool.endorseFailed(blkNum, self.GetChainConfig().C)
			eMsgs := self.msgPool.GetEndorsementsMsgs(blkNum)
			for _, msg := range eMsgs {
				e := msg.(*blockEndorseMsg)
				if e != nil && e.Endorser == self.Index && e.EndorseForEmpty == endorseFailed {
					log.Infof("server %d rebroadcast endorse, blk %d for %d, %t",
						self.Index, e.GetBlockNum(), e.EndorsedProposer, e.EndorseForEmpty)
					self.broadcast(e)
					rebroadcasted = true
				}
			}
			if !rebroadcasted {
				proposal := self.getHighestRankProposal(blkNum, proposals)
				if proposal != nil {
					if err := self.endorseBlock(proposal, false); err != nil {
						log.Errorf("server %d rebroadcasting failed to endorse (%d): %s",
							self.Index, blkNum, err)
					}
				} else {
					log.Errorf("server %d rebroadcasting failed to endorse(%d), no proposal found(%d)",
						self.Index, blkNum, len(proposals))
				}
			}
		} else if proposal, forEmpty := self.blockPool.GetEndorsedProposal(blkNum); proposal != nil {
			// construct endorse msg
			if endorseMsg, _ := self.constructEndorseMsg(proposal, forEmpty); endorseMsg != nil {
				self.broadcast(endorseMsg)
			}
		}
		if self.isCommitter(self.Index) {
			committed := false
			cMsgs := self.msgPool.GetCommitMsgs(self.GetCurrentBlockNo())
			for _, msg := range cMsgs {
				c := msg.(*blockCommitMsg)
				if c != nil && c.Committer == self.Index {
					log.Infof("server %d rebroadcast commit, blk %d for %d, %t",
						self.Index, c.GetBlockNum(), c.BlockProposer, c.CommitForEmpty)
					self.broadcast(msg)
					committed = true
				}
			}
			if !committed {
				if proposer, forEmpty, done := self.blockPool.endorseDone(blkNum, self.GetChainConfig().C); done {
					proposal := self.findBlockProposal(blkNum, proposer)

					// consensus ok, make endorsement
					if proposal == nil {
						self.fetchProposal(blkNum, proposer)
						// restart endorsing timer
						self.timer.StartEventTimer(EventEndorseBlockTimeout, blkNum)
						log.Errorf("server %d endorse %d done, but no proposal", self.Index, blkNum)
					} else if err := self.commitBlock(proposal, forEmpty); err != nil {
						log.Errorf("server %d failed to commit block %d on rebroadcasting: %s",
							self.Index, blkNum, err)
					}
				} else if self.blockPool.endorseFailed(blkNum, self.GetChainConfig().C) {
					// endorse failed, start empty endorsing
					self.timer.C <- &TimerEvent{
						evtType:  EventEndorseBlockTimeout,
						blockNum: blkNum,
					}
				}
			}
		}
	case SubmitBlock:
		blkNum := self.GetCurrentBlockNo()
		if action.BlockNum > blkNum {
			return
		}
		stateRoot, err := self.blockPool.getExecMerkleRoot(action.BlockNum)
		if err != nil {
			log.Infof("handleBlockSubmit failed:%s", err)
			return
		}
		if self.CheckSubmitBlock(action.BlockNum, stateRoot) {
			if err := self.blockPool.submitBlock(action.BlockNum); err != nil {
				log.Errorf("SubmitBlock err:%s", err)
			}
		}
	}
}

func (self *Server) actionLoop() {
	self.quitWg.Add(1)
	defer self.quitWg.Done()

	for {
		select {
		case action := <-self.bftActionC:
			self.processBftAction(action)
		case <-self.quitC:
			log.Infof("server %d actionLoop quit", self.Index)
			return
		}
	}
}

func (self *Server) timerLoop() {
	self.quitWg.Add(1)
	defer self.quitWg.Done()

	for {
		select {
		case evt := <-self.timer.C:
			if err := self.processTimerEvent(evt); err != nil {
				log.Errorf("failed to process timer evt: %d, err: %s", evt.evtType, err)
			}
		case <-self.quitC:
			log.Infof("server %d timerLoop quit", self.Index)
			return
		}
	}
}

func (self *Server) processTimerEvent(evt *TimerEvent) error {
	switch evt.evtType {
	case EventProposalBackoff:
		// 1. if endorsed, return
		// 2. if no proposal received,
		// 		if 2nd proposer, make proposal, start endorse timeout, return
		//		else, return (wait proposal timeout)
		// 3. else:
		// 		if no proposal from leader, return (wait proposal timeout will make endorse anyway)
		// 		else, return (endorsing on leader-proposal done when received the proposal)
		//

		if self.blockPool.HasEndorsedForBlock(evt.blockNum) {
			return nil
		}
		if !self.getState().IsReady() {
			return nil
		}
		proposals := self.blockPool.GetBlockProposals(evt.blockNum)
		if len(proposals) == 0 {
			// no proposal received, make proposal, start endorse timeout
			if self.is2ndProposer(evt.blockNum, self.Index) {
				if err := self.makeProposal(evt.blockNum, false); err != nil {
					return fmt.Errorf("failed to make 2nd proposal (%d): %s", evt.blockNum, err)
				}
			}
		}

	case EventProposeBlockTimeout:
		// 1. if endorsed, return
		// 2. check proposal from leader, if there is, endorse the proposal, start endorse timeout, return
		// 3. check proposal from 2nd proposer, endorse for first 2nd proposal, start endorse timeout, return
		// 4. if not, random backoff, return
		//
		// then propose for empty block, start 2ndProposal timeout, return
		//
		return self.handleProposalTimeout(evt)

	case EventRandomBackoff:
		// 1. if endorsed, return
		// 2. if any valid proposal, endorse on high-priority one (priority from vrf), start endorse timeout, return
		// 3. make empty proposal, broadcast, start 2nd proposal timeout, return
		//
		return self.handleProposalTimeout(evt)

	case EventPropose2ndBlockTimeout:
		// 1. if endorsed, return
		// 2. there must some valid proposal, if not, force resync, reset peer neighbours
		// 3. endorse on highest-priority one, start endorse timeout, return
		//
		return self.handleProposalTimeout(evt)

	case EventEndorseBlockTimeout:
		// 1. if committed, return
		// 2. check endorse quorum
		// 3. if quorum reached, endorse the proposal, start commit timeout, return
		// 4. broadcast endorse on highest-priority proposal empty, start empty endorse timeout, return
		//
		if self.blockPool.committedForBlock(evt.blockNum) {
			return nil
		}
		if !self.getState().IsReady() {
			return nil
		}
		if proposer, forEmpty, done := self.blockPool.endorseDone(evt.blockNum, self.GetChainConfig().C); done {
			proposal := self.findBlockProposal(evt.blockNum, proposer)

			// consensus ok, make endorsement
			if proposal == nil {
				self.fetchProposal(evt.blockNum, proposer)
				// restart endorsing timer
				self.timer.StartEventTimer(EventEndorseBlockTimeout, evt.blockNum)
				return fmt.Errorf("endorse %d done, but no proposal available", evt.blockNum)
			}
			if err := self.verifyPrevBlockHash(evt.blockNum, proposal); err != nil {
				// restart endorsing timer
				self.timer.StartEventTimer(EventEndorseBlockTimeout, evt.blockNum)
				return fmt.Errorf("endorse %d done, but prev blk hash inconsistency: %s", evt.blockNum, err)
			}
			if err := self.commitBlock(proposal, forEmpty); err != nil {
				return fmt.Errorf("failed to endorse for block %d on endorse timeout: %s", evt.blockNum, err)
			}
			return nil
		}
		if !self.getState().IsActive() {
			// not active yet, waiting active peers making decision
			return nil
		}
		if self.blockPool.HasEndorsedForEmptyBlock(evt.blockNum) {
			return nil
		}
		proposals := self.blockPool.GetBlockProposals(evt.blockNum)
		if len(proposals) == 0 {
			log.Errorf("endorsing timeout, without any proposal. restarting syncing")
			self.restartSyncing()
			return nil
		}
		proposal := self.getHighestRankProposal(evt.blockNum, proposals)
		if proposal != nil {
			if err := self.endorseBlock(proposal, true); err != nil {
				return fmt.Errorf("failed to endorse block proposal (%d): %s", evt.blockNum, err)
			}
		}
		return nil

	case EventEndorseEmptyBlockTimeout:
		// 1. if committed, return
		// 2. check endorse quorum
		// 3. if quorum reached, commit the proposal, start commit timeout, return
		// 4. check empty endorse quorum
		// 5. if empty endorse quorum reached, commit the empty proposal, start commit timeout, return
		//
		if self.blockPool.committedForBlock(evt.blockNum) {
			return nil
		}
		if !self.getState().IsReady() {
			return nil
		}
		if proposer, forEmpty, done := self.blockPool.endorseDone(evt.blockNum, self.GetChainConfig().C); done {
			proposal := self.findBlockProposal(evt.blockNum, proposer)

			// consensus ok, make endorsement
			if proposal == nil {
				self.fetchProposal(evt.blockNum, proposer)
				// restart timer
				self.timer.StartEventTimer(EventEndorseEmptyBlockTimeout, evt.blockNum)
			} else if err := self.commitBlock(proposal, forEmpty); err != nil {
				return fmt.Errorf("failed to endorse for block %d on empty endorse timeout: %s", evt.blockNum, err)
			}
			return nil
		} else {
			log.Errorf("server %d: empty endorse timeout, no quorum", self.Index)
			if !self.getState().IsActive() {
				proposals := self.blockPool.GetBlockProposals(evt.blockNum)
				proposal := self.getHighestRankProposal(evt.blockNum, proposals)
				if proposal != nil {
					if err := self.endorseBlock(proposal, true); err != nil {
						return fmt.Errorf("failed to endorse block proposal (%d): %s", evt.blockNum, err)
					}
				}
			} else {
				self.timer.StartEventTimer(EventEndorseEmptyBlockTimeout, evt.blockNum)
			}
		}
		return nil

	case EventCommitBlockTimeout:
		// 1. if sealed, return
		// 2. check commit quorum
		// 3. if quorum reached, seal the commit, start new round, return
		// 4. else: there must have some network issues, force resync, reset all neighbours
		//
		if blk, _ := self.blockPool.getSealedBlock(evt.blockNum); blk != nil {
			return nil
		}
		if !self.getState().IsReady() {
			return nil
		}
		if !self.blockPool.isCommitHadDone(evt.blockNum) {
			chainCfg := self.GetChainConfig()
			if proposer, forEmpty, done := self.blockPool.commitDone(evt.blockNum, chainCfg.C, chainCfg.N); done {
				self.blockPool.setCommitDone(evt.blockNum)
				proposal := self.findBlockProposal(evt.blockNum, proposer)
				if proposal == nil {
					self.restartSyncing()
					return fmt.Errorf("commit timeout, consensused proposal not available. need resync")
				}
				if !self.blockPool.checkBlockSign(proposal.Block, forEmpty, self.config.N-(self.config.N-1)/3) {
					self.restartSyncing()
					log.Errorf("server %d commit timeout checkBlockSign insufficient at blk: %d", self.Index, evt.blockNum)
					return fmt.Errorf("commit timeout, consensused blockSign not enough. need resync")
				}

				if err := self.makeSealed(proposal, forEmpty); err != nil {
					return fmt.Errorf("commit timeout, failed to seal block %d: %s", evt.blockNum, err)
				}
				return nil
			} else {
				log.Errorf("server %d commit blk %d timeout without consensus", self.Index, evt.blockNum)
				self.restartSyncing()
			}
		}

	case EventPeerHeartbeat:
		self.heartbeat(evt.blockNum)

	case EventTxPool:
		self.timer.CancelEventTimer(EventTxPool, evt.blockNum)
		if self.GetCompletedBlockNum()+1 == evt.blockNum {
			validHeight := self.validHeight(evt.blockNum)
			newProposal := false
			nonceCtx := make(map[common.Address]uint64)
			for _, e := range self.poolActor.GetTxnPool(true, validHeight) {
				if err := self.incrValidator.Verify(e.Tx, validHeight, nonceCtx); err == nil {
					newProposal = true
					break
				}
			}
			if newProposal {
				self.timer.CancelEventTimer(EventTxBlockTimeout, evt.blockNum)
				self.startNewProposal(evt.blockNum)
			} else {
				//reset timer, continue waiting txs from txnpool
				self.timer.StartEventTimer(EventTxPool, evt.blockNum)
			}
		} else {
			self.timer.StartEventTimer(EventTxPool, evt.blockNum)
		}
	case EventTxBlockTimeout:
		self.timer.CancelEventTimer(EventTxPool, evt.blockNum)
		self.timer.CancelEventTimer(EventTxBlockTimeout, evt.blockNum)
		self.startNewProposal(evt.blockNum)
	}
	return nil
}

func (self *Server) processHeartbeatMsg(peerIdx uint32, msg *peerHeartbeatMsg) {
	self.peerPool.UpdatePeerCommitBlockNo(peerIdx, msg.CommittedBlockNumber)
	log.Debugf("server %d received heartbeat from peer %d, chainview %d, blkNum %d",
		self.Index, peerIdx, msg.ChainConfigView, msg.CommittedBlockNumber)
	self.stateMgr.StateEventC <- &StateEvent{
		Type: UpdatePeerState,
		peerState: &PeerState{
			peerIdx:           peerIdx,
			connected:         true,
			chainConfigView:   msg.ChainConfigView,
			committedBlockNum: msg.CommittedBlockNumber,
		},
	}
}

func (self *Server) endorseBlock(proposal *blockProposalMsg, forEmpty bool) error {
	// for each round, one node can only endorse one block, or empty block
	if proposal.Block.getProposer() == self.Index {
		return nil
	}
	blkNum := proposal.GetBlockNum()

	// check if has endorsed
	if !forEmpty && self.blockPool.HasEndorsedForBlock(blkNum) {
		return nil
	} else if forEmpty && self.blockPool.HasEndorsedForEmptyBlock(blkNum) {
		return nil
	}

	if !forEmpty {
		if self.blockPool.endorseFailed(blkNum, self.GetChainConfig().C) {
			forEmpty = true
			log.Errorf("server %d, endorsing %d, changed from true to false", self.Index, blkNum)
		}
	}

	// build endorsement msg
	endorseMsg, err := self.constructEndorseMsg(proposal, forEmpty)
	if err != nil {
		return fmt.Errorf("failed to construct endorse msg: %s", err)
	}

	// set the block as self-endorsed-block
	if err := self.blockPool.setProposalEndorsed(proposal, forEmpty); err != nil {
		return fmt.Errorf("failed to set proposal as endorsed: %s", err)
	}

	self.processConsensusMsg(endorseMsg)
	// if node is endorser of current round
	if forEmpty || self.isEndorser(self.Index) {
		h, _ := HashMsg(endorseMsg)
		self.msgPool.AddMsg(endorseMsg, h)
		log.Infof("endorser %d, endorsed block %d, from server %d",
			self.Index, blkNum, proposal.Block.getProposer())
		// broadcast my endorsement
		self.broadcast(endorseMsg)
		return nil
	}

	// start endorsing timer
	// TODO: endorsing may have reached consensus before received proposal, handle this
	if !forEmpty {
		self.timer.StartEventTimer(EventEndorseBlockTimeout, blkNum)
	} else {
		self.timer.StartEventTimer(EventEndorseEmptyBlockTimeout, blkNum)
	}

	return nil
}

func (self *Server) commitBlock(proposal *blockProposalMsg, forEmpty bool) error {
	// for each round, we can only commit one block
	if proposal.Block.getProposer() == self.Index {
		return nil
	}
	blkNum := proposal.GetBlockNum()
	if self.blockPool.committedForBlock(blkNum) {
		return nil
	}

	var blkHash common.Uint256
	if !forEmpty {
		blkHash = proposal.Block.Block.Hash()
	} else {
		if proposal.Block.EmptyBlock == nil {
			return fmt.Errorf("blk %d proposal from %d has no empty proposal", blkNum, proposal.Block.getProposer())
		}
		blkHash = proposal.Block.EmptyBlock.Hash()
	}

	endorses := make([]*blockEndorseMsg, 0)
	for _, msg := range self.msgPool.GetEndorsementsMsgs(blkNum) {
		if e := msg.(*blockEndorseMsg); e != nil {
			if bytes.Compare(blkHash[:], e.EndorsedBlockHash[:]) == 0 && e.EndorseForEmpty == forEmpty {
				endorses = append(endorses, e)
			}
		}
	}

	// build commit msg
	commitMsg, err := self.constructCommitMsg(proposal, endorses, forEmpty)
	if err != nil {
		return fmt.Errorf("failed to construct commit msg: %s", err)
	}

	// set the block as committed-block
	if err := self.blockPool.SetProposalCommitted(proposal, forEmpty); err != nil {
		return fmt.Errorf("failed to set proposal as committed: %s", err)
	}

	self.processConsensusMsg(commitMsg)
	// if node is committer of current round
	if forEmpty || self.isCommitter(self.Index) {
		h, _ := HashMsg(commitMsg)
		self.msgPool.AddMsg(commitMsg, h)
		log.Infof("committer %d, set block %d committed, from server %d",
			self.Index, blkNum, proposal.Block.getProposer())
		// broadcast my commitment
		self.broadcast(commitMsg)
		return nil
	}

	// start commit timer
	// TODO: committing may have reached consensus before received endorsement, handle this
	self.timer.StartEventTimer(EventCommitBlockTimeout, blkNum)
	return nil
}

// Note: sealProposal updates self.currentBlockNum, make sure not concurrency
// (only called by sealProposal action)
func (self *Server) sealProposal(proposal *blockProposalMsg, empty bool) error {
	// for each round, we can only seal one block
	if err := self.sealBlock(proposal.Block, empty, true); err != nil {
		return err
	}

	if self.hasBlockConsensused() {
		self.makeFastForward()
	} else {
		return self.startNewRound()
	}

	return nil
}

func (self *Server) fastForwardBlock(block *VbftBlock) error {

	// TODO: update chainconfig when forwarding

	if self.getState().IsActive() {
		return fmt.Errorf("server %d: invalid fastforward, current state: %d", self.Index, self.getState())
	}
	if self.GetCurrentBlockNo() > block.getBlockNum() {
		return nil
	}
	if self.GetCurrentBlockNo() == block.getBlockNum() {
		// block from peer syncer, there should only one candidate block
		flag := false
		if len(block.Block.Header.SigData) <= 1 {
			flag = true
		}
		return self.sealBlock(block, false, flag)
	}
	return fmt.Errorf("server %d: fastforward blk %d failed, current blkNum: %d",
		self.Index, block.getBlockNum(), self.GetCurrentBlockNo())
}

func (self *Server) sealBlock(block *VbftBlock, empty bool, sigdata bool) error {
	sealedBlkNum := block.getBlockNum()
	if sealedBlkNum < self.GetCurrentBlockNo() {
		// we already in future round
		log.Errorf("late seal of %d, current blkNum: %d", sealedBlkNum, self.GetCurrentBlockNo())
		return nil
	} else if sealedBlkNum > self.GetCurrentBlockNo() {
		// we have lost sync, restarting syncing
		self.restartSyncing()
		return fmt.Errorf("future seal of %d, current blknum: %d", sealedBlkNum, self.GetCurrentBlockNo())
	}

	if err := self.blockPool.SetBlockSealed(block, empty, sigdata); err != nil {
		return fmt.Errorf("failed to seal proposal: %s", err)
	}

	// TODO: also persistent the block endorsers and committer msgs

	// notify other modules that block sealed
	self.timer.OnBlockSealed(sealedBlkNum)
	self.msgPool.OnBlockSealed(sealedBlkNum)

	_, h := self.blockPool.getSealedBlock(sealedBlkNum)
	prevBlkHash := block.getPrevBlockHash()
	log.Infof("server %d, sealed block %d, proposer %d, prevhash: %s, hash: %s", self.Index,
		sealedBlkNum, block.getProposer(), prevBlkHash.ToHexString(), h.ToHexString())

	// broadcast to other modules
	// TODO: block committed, update tx pool, notify block-listeners
	if sealedBlkNum >= self.GetCurrentBlockNo() {
		self.SetCurrentBlockNo(sealedBlkNum + 1)
	}
	return nil
}

func (self *Server) msgSendLoop() {
	self.quitWg.Add(1)
	defer self.quitWg.Done()

	for {
		select {
		case evt := <-self.msgSendC:
			if self.nonConsensusNode() {
				continue
			}
			if evt.ToPeer == math.MaxUint32 {
				// broadcast
				self.broadcastToAll(evt.Msg)
			} else {
				if err := self.sendToPeer(evt.ToPeer, evt.Msg); err != nil {
					log.Errorf("server %d xmit to peer %d failed: %s", self.Index, evt.ToPeer, err)
				}
			}

		case <-self.quitC:
			log.Infof("server %d msgSendLoop quit", self.Index)
			return
		}
	}
}

// creategovernaceTransaction invoke governance native contract commit_pos
func (self *Server) creategovernaceTransaction(blkNum uint32) (*types.Transaction, error) {
	mutable := utils.BuildNativeTransaction(nutils.GovernanceContractAddress, gover.COMMIT_DPOS, []byte{})
	mutable.Nonce = blkNum
	tx, err := mutable.IntoImmutable()
	return tx, err
}

// checkNeedUpdateChainConfig use blockcount
func (self *Server) checkNeedUpdateChainConfig(blockNum uint32) bool {
	prevBlk, _ := self.blockPool.getSealedBlock(blockNum - 1)
	if prevBlk == nil {
		log.Errorf("failed to get prevBlock (%d)", blockNum-1)
		return false
	}
	lastConfigBlkNum := prevBlk.getLastConfigBlockNum()
	if (blockNum - lastConfigBlkNum) >= self.GetChainConfig().MaxBlockChangeView {
		return true
	}
	return false
}

// checkUpdateChainConfig query leveldb check is force update
func (self *Server) checkUpdateChainConfig(blkNum uint32) bool {
	force, err := isUpdate(self.blockPool.getExecWriteSet(blkNum-1), self.GetChainConfig().View)
	if err != nil {
		log.Errorf("checkUpdateChainConfig err:%s", err)
		return false
	}
	log.Debugf("checkUpdateChainConfig force: %v", force)
	return force
}

func (self *Server) validHeight(blkNum uint32) uint32 {
	height := blkNum - 1
	validHeight := height
	start, end := self.incrValidator.BlockRange()
	if height+1 == end {
		validHeight = start
	} else {
		self.incrValidator.Clean()
	}
	return validHeight
}

func (self *Server) nonSystxs(sysTxs []*types.Transaction, blkNum uint32) bool {
	if self.checkNeedUpdateChainConfig(blkNum) && len(sysTxs) == 1 {
		invoke := sysTxs[0].Payload.(*payload.InvokeCode)
		if invoke == nil {
			log.Errorf("nonSystxs invoke is nil,blocknum:%d", blkNum)
			return true
		}
		if bytes.Compare(invoke.Code, ninit.COMMIT_DPOS_BYTES) == 0 {
			return false
		}
	}
	return true
}

func (self *Server) makeProposal(blkNum uint32, forEmpty bool) error {
	if blkNum < self.GetCurrentBlockNo() {
		return fmt.Errorf("server %d ignore deprecatd blk proposal %d, current %d",
			self.Index, blkNum, self.GetCurrentBlockNo())
	}

	validHeight := self.validHeight(blkNum)
	sysTxs := make([]*types.Transaction, 0)
	userTxs := make([]*types.Transaction, 0)

	//check need upate chainconfig
	var cfg *vconfig.ChainConfig
	if self.checkNeedUpdateChainConfig(blkNum) || self.checkUpdateChainConfig(blkNum) {
		chainconfig, err := getChainConfig(self.blockPool.getExecWriteSet(blkNum-1), blkNum)
		if err != nil {
			return fmt.Errorf("getChainConfig failed:%s", err)
		}
		//add transaction invoke governance native commit_pos contract
		if self.checkNeedUpdateChainConfig(blkNum) {
			tx, err := self.creategovernaceTransaction(blkNum)
			if err != nil {
				return fmt.Errorf("construct governace transaction error: %v", err)
			}
			sysTxs = append(sysTxs, tx)
			chainconfig.View++
		}
		forEmpty = true
		cfg = chainconfig
	}
	if self.nonConsensusNode() {
		return fmt.Errorf("%d quit consensus node", self.Index)
	}

	if !forEmpty {
		nonceCtx := make(map[common.Address]uint64)
		for _, e := range self.poolActor.GetTxnPool(true, validHeight) {
			if err := self.incrValidator.Verify(e.Tx, validHeight, nonceCtx); err == nil {
				userTxs = append(userTxs, e.Tx)
			}
		}
		log.Infof("make proposal get %d valid tx from pool", len(userTxs))
	}

	proposal, err := self.constructProposalMsg(blkNum, sysTxs, userTxs, cfg)
	if err != nil {
		return fmt.Errorf("failed to construct proposal: %s", err)
	}

	log.Infof("server %d make proposal for block %d", self.Index, blkNum)

	// add proposal to self
	h, _ := HashMsg(proposal)
	self.msgPool.AddMsg(proposal, h)
	self.processProposalMsg(proposal)
	self.broadcast(proposal)
	return nil
}

func (self *Server) makeSealed(proposal *blockProposalMsg, forEmpty bool) error {
	blkNum := proposal.GetBlockNum()

	if err := self.verifyPrevBlockHash(blkNum, proposal); err != nil {
		// TODO: in-consistency with prev-blockhash, resync-required
		self.restartSyncing()
		return fmt.Errorf("verify prev block hash failed: %s", err)
	}

	log.Infof("server %d ready to seal block %d, for proposer %d, empty: %t",
		self.Index, blkNum, proposal.Block.getProposer(), forEmpty)

	// seal the block
	self.bftActionC <- &BftAction{
		Type:     SealBlock,
		BlockNum: blkNum,
		Proposal: proposal,
		forEmpty: forEmpty,
	}
	return nil
}

func (self *Server) makeFastForward() {
	go func() {
		self.bftActionC <- &BftAction{
			Type: FastForward,
		}
	}()
}

func (self *Server) reBroadcastCurrentRoundMsgs() {
	go func() {
		self.bftActionC <- &BftAction{
			Type:     ReBroadcast,
			BlockNum: self.GetCurrentBlockNo(),
		}
	}()
}

func (self *Server) makeBlockSubmit(blknum uint32) {
	go func() {
		self.bftActionC <- &BftAction{
			Type:     SubmitBlock,
			BlockNum: blknum,
		}
	}()
}

func (self *Server) fetchProposal(blkNum uint32, proposer uint32) {
	msg := self.constructProposalFetchMsg(blkNum, proposer)
	self.msgSendC <- &SendMsgEvent{
		ToPeer: math.MaxUint32,
		Msg:    msg,
	}
}

func (self *Server) handleProposalTimeout(evt *TimerEvent) error {
	if self.blockPool.HasEndorsedForBlock(evt.blockNum) {
		return nil
	}
	if !self.getState().IsReady() {
		return nil
	}
	proposals := self.blockPool.GetBlockProposals(evt.blockNum)

	log.Infof("server %d proposal timeout, known proposals %d, timeout: %d", self.Index, len(proposals), evt.evtType)

	// if no proposal available, random backoff
	if len(proposals) == 0 {
		log.Infof("no proposal available for block %d, timeout: %d", evt.blockNum, evt.evtType)

		switch evt.evtType {
		case EventProposeBlockTimeout:
			self.timer.StartEventTimer(EventRandomBackoff, evt.blockNum)
			log.Infof("server %d started backoff timer for blk %d", self.Index, evt.blockNum)
			return nil
		case EventRandomBackoff:
			if self.is2ndProposer(evt.blockNum, self.Index) {
				if err := self.makeProposal(evt.blockNum, true); err != nil {
					return fmt.Errorf("failed to propose empty block: %s", err)
				}
				self.timer.StartEventTimer(EventPropose2ndBlockTimeout, evt.blockNum)
				log.Infof("server %d proposed empty block for blk %d", self.Index, evt.blockNum)
			}
			return nil
		case EventPropose2ndBlockTimeout:
			// 2nd proposal without any proposal, force resync
			self.restartSyncing()
			return nil
		}
	}

	if evt.evtType == EventRandomBackoff {
		// proposal available, no proposing
		return nil
	}

	// find highest rank proposal
	proposal := self.getHighestRankProposal(evt.blockNum, proposals)
	if self.isProposer(proposal.Block.getProposer()) {
		// proposal msg handler will do the endorsement
		return nil
	}

	self.bftActionC <- &BftAction{
		Type:     EndorseBlock,
		BlockNum: evt.blockNum,
		Proposal: proposal,
		forEmpty: false,
	}
	return nil
}

// TODO: refactor this
func (self *Server) catchConsensus(blkNum uint32) error {
	if !self.isEndorser(self.Index) && !self.isCommitter(self.Index) {
		return nil
	}

	proposals := make(map[uint32]*blockProposalMsg)
	pMsgs := self.msgPool.GetProposalMsgs(blkNum)
	for _, msg := range pMsgs {
		p, ok := msg.(*blockProposalMsg)
		if !ok {
			continue
		}
		proposals[p.Block.getProposer()] = p
	}

	chainCfg := self.GetChainConfig()
	eMsgs := self.msgPool.GetEndorsementsMsgs(blkNum)
	var proposal *blockProposalMsg
	endorseDone := false
	endorseEmpty := false
	if len(eMsgs) > int(chainCfg.C) {
		var maxProposer uint32
		emptyCnt := 0
		maxCnt := 0
		proposers := make(map[uint32]int)
		for _, msg := range eMsgs {
			c, ok := msg.(*blockEndorseMsg)
			if !ok {
				continue
			}
			if c.EndorseForEmpty {
				emptyCnt++
			}
			proposers[c.EndorsedProposer] += 1
			if proposers[c.EndorsedProposer] > maxCnt {
				maxProposer = c.EndorsedProposer
				maxCnt = proposers[c.EndorsedProposer]
			}
		}
		proposal = proposals[maxProposer]
		if maxCnt > int(chainCfg.C) {
			endorseDone = true
		}
		if emptyCnt > int(chainCfg.C) {
			endorseDone = true
			endorseEmpty = true
		}
	}
	if proposal != nil && self.isProposer(proposal.Block.getProposer()) {
		self.processProposalMsg(proposal)
	}

	if self.isEndorser(self.Index) && !endorseDone && proposal != nil {
		return self.endorseBlock(proposal, endorseEmpty)
	}

	if !endorseDone {
		return fmt.Errorf("server %d catch consensus with endorse failed", self.Index)
	}

	if !self.isCommitter(self.Index) {
		return nil
	}

	var maxProposer uint32
	maxCnt := 0
	emptyCnt := 0
	proposers := make(map[uint32]int)
	cMsgs := self.msgPool.GetCommitMsgs(blkNum)
	for _, msg := range cMsgs {
		c, ok := msg.(*blockCommitMsg)
		if !ok {
			continue
		}
		if c.CommitForEmpty {
			emptyCnt++
		}
		proposers[c.BlockProposer] += 1
		if proposers[c.BlockProposer] > maxCnt {
			maxProposer = c.BlockProposer
		}
	}

	if p := proposals[maxProposer]; p != nil {
		return self.commitBlock(p, emptyCnt > 0)
	}

	return nil
}

func (self *Server) verifyPrevBlockHash(blkNum uint32, proposal *blockProposalMsg) error {
	prevBlk, prevBlkHash := self.blockPool.getSealedBlock(blkNum - 1)
	if prevBlk == nil {
		// TODO: has no candidate proposals for prevBlock, should restart syncing
		return fmt.Errorf("failed to get prevBlock of current round (%d)", blkNum)
	}
	prevBlkHash2 := proposal.Block.getPrevBlockHash()
	if prevBlkHash != prevBlkHash2 {
		// continue waiting for more proposals
		// FIXME
		return fmt.Errorf("inconsistent prev-block hash %s vs %s (blk %d)",
			prevBlkHash.ToHexString(), prevBlkHash2.ToHexString(), blkNum)
	}

	return nil
}

func (self *Server) hasBlockConsensused() bool {
	blkNum := self.GetCurrentBlockNo()

	C := int(self.GetChainConfig().C)
	cMsgs := self.msgPool.GetCommitMsgs(blkNum)
	emptyCnt := 0
	proposers := make(map[uint32]int)
	for _, msg := range cMsgs {
		c, ok := msg.(*blockCommitMsg)
		if !ok {
			continue
		}
		if c.CommitForEmpty {
			emptyCnt++
		}
		proposers[c.BlockProposer] += 1
		if proposers[c.BlockProposer] > C {
			return true
		}
	}

	return emptyCnt > C
}

func (self *Server) restartSyncing() {
	// send sync request to self.sync, go syncing-state immediately
	// stop all bft timers

	self.stateMgr.StateEventC <- &StateEvent{Type: ForceSyncing}
}
