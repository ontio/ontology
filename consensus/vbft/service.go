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
	"github.com/ontio/ontology/core/store"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/core/utils"
	"github.com/ontio/ontology/events"
	"github.com/ontio/ontology/events/message"
	p2pmsg "github.com/ontio/ontology/p2pserver/message/types"
	p2p "github.com/ontio/ontology/p2pserver/net/protocol"
	gover "github.com/ontio/ontology/smartcontract/service/native/governance"
	nutils "github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/validator/increment"
)

const (
	CAP_MESSAGE_CHANNEL  = 4096
	CAP_ACTION_CHANNEL   = 64
	CAP_MSG_SEND_CHANNEL = 16
)

type BlockAndExecteInfo struct {
	Block           *types.Block
	Info            *vconfig.VbftBlockInfo
	WriteSet        *overlaydb.MemDB
	MerkleRoot      common.Uint256
	CrossStatesRoot common.Uint256
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
	metaLock sync.RWMutex
	vbftCtx  *VbftContext

	msgPool   *MsgPool   // consensus msg pool
	blockPool *BlockPool // received block proposals
	peerPool  *PeerPool  // consensus peers
	syncer    *Syncer
	stateMgr  *StateMgr
	timer     *EventTimer

	msgRecvC     *sync.Map // map[uint32]chan *p2pMsgPayload
	msgC         chan ConsensusMsg
	rebroadcastC chan uint32
	msgSendC     chan *SendMsgEvent
	sub          *events.ActorSubscriber
	quitC        chan struct{}
	quitWg       sync.WaitGroup
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
		self.handleBlockPersistCompleted(msg.Block, msg.ExecResult)
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

func (self *Server) handleBlockPersistCompleted(block *types.Block, exec *store.ExecuteResult) {
	log.Infof("persist block complete: height=%d, hash=%x, numtx=%d", block.Header.Height, block.Hash(),
		len(block.Transactions))
	blkInfo, err := vconfig.VbftBlock(block.Header)
	if err != nil {
		log.Errorf("load vbft block info failed:%s", err)
		return
	}
	if self.updateVbftContext(block, blkInfo, exec) {
		// p2p synced before seal block: 1. not consensus node; 2. consensus node in syncing state
		self.blockPool.ReloadFromLedger()
	}
}

func (self *Server) CheckAndSubmitBlock(blkNum uint32, stateRoot common.Uint256) {
	cMsgs := self.msgPool.GetBlockSubmitMsgs(blkNum)
	var stateRootCnt uint32
	for _, msg := range cMsgs {
		c := msg.(*blockSubmitMsg)
		if c.BlockStateRoot == stateRoot {
			stateRootCnt++
		}
	}

	cfg := self.GetChainConfig()
	m := cfg.N - (cfg.N-1)/3

	if stateRootCnt >= m {
		log.Infof("receive enough submit msg for block %d, start submit block", blkNum)
		if err := self.blockPool.SubmitBlock(blkNum); err != nil {
			log.Errorf("SubmitBlock err:%s", err)
		}
	}
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
	blkNum := store.ChainedBlockNum
	block, _ := store.GetBlock(blkNum)
	if block == nil {
		return fmt.Errorf("getSealedBlock err height:%d", blkNum)
	}
	crossStateRoot, err := store.GetCrossStatesRoot(blkNum)
	if err != nil {
		return fmt.Errorf("get cross state root  err height:%d", blkNum)
	}
	stateRoot, err := store.GetExecMerkleRoot(blkNum)
	if err != nil {
		return fmt.Errorf("get state root  err height:%d", blkNum)
	}

	var cfg vconfig.ChainConfig
	configBlk := blkNum
	if block.getNewChainConfig() != nil {
		cfg = *block.getNewChainConfig()
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
		configBlk = lastConfigNum
	}

	if cfg.View == 0 || cfg.MaxBlockChangeView == 0 {
		panic("invalid view or maxblockchangeview ")
	}

	// update timer params
	self.updateTimerParams(&cfg)

	log.Infof("current committed block no: %d", blkNum)
	proposers, endorsers, committers := buildPeerRoles(blkNum+1, block.Info.Proposer, block.Info.VrfValue, &cfg)
	log.Infof("server %d, blkNum: %d, state: %d, participants: %v, %v, %v", self.Index, blkNum,
		self.getState(), proposers, endorsers, committers)

	self.vbftCtx = &VbftContext{
		BlockNum:  blkNum + 1,
		Config:    &cfg,
		ConfigNum: configBlk,
		PrevBlockInfo: &BlockAndExecteInfo{
			Block:           block.Block,
			Info:            block.Info,
			WriteSet:        nil,
			MerkleRoot:      stateRoot,
			CrossStatesRoot: crossStateRoot,
		},
		Proposers:  proposers,
		Endorsers:  endorsers,
		Committers: committers,
	}
	self.incrValidator.AddBlock(block.Block)

	return nil
}

func (self *Server) nonConsensusNode() bool {
	return self.Index == math.MaxUint32
}

func (self *Server) updateVbftContext(block *types.Block, info *vconfig.VbftBlockInfo, result *store.ExecuteResult) (updated bool) {
	blkNum := block.Header.Height
	vbftCtx := *self.GetVbftContext()
	if info.NewChainConfig != nil {
		vbftCtx.Config = info.NewChainConfig
		vbftCtx.ConfigNum = blkNum
	}

	vbftCtx.Proposers, vbftCtx.Endorsers, vbftCtx.Committers = buildPeerRoles(blkNum+1, info.Proposer, info.VrfValue, vbftCtx.Config)

	vbftCtx.BlockNum = blkNum + 1
	vbftCtx.PrevBlockInfo = &BlockAndExecteInfo{
		Block:           block,
		Info:            info,
		WriteSet:        result.WriteSet,
		MerkleRoot:      result.MerkleRoot,
		CrossStatesRoot: result.CrossStatesRoot,
	}

	self.metaLock.Lock()
	if self.vbftCtx.BlockNum+1 == vbftCtx.BlockNum {
		self.vbftCtx = &vbftCtx
		self.incrValidator.AddBlock(block)
		start, end := self.incrValidator.BlockRange()
		log.Infof("update vbft context, blkNum:%d, incr validator range: [%d, %d)", blkNum+1, start, end)
		updated = true
	}
	self.metaLock.Unlock()
	if updated {
		log.Infof("server %d, blkNum: %d, state: %d, participants: %v, %v, %v", self.Index, blkNum+1,
			self.getState(), vbftCtx.Proposers, vbftCtx.Endorsers, vbftCtx.Committers)
		if info.NewChainConfig != nil {
			self.updateTimerAndPeerPool(info.NewChainConfig)
		}
	}
	return updated
}

func (self *Server) updateTimerAndPeerPool(config *vconfig.ChainConfig) {
	pubkey := vconfig.PubkeyID(self.account.PublicKey)
	peermap := make(map[string]uint32)
	for _, p := range config.Peers {
		peermap[p.ID] = p.Index
		if self.Index == math.MaxUint32 && pubkey == p.ID {
			self.Index = p.Index
			log.Infof("updateTimerAndPeerPool add index :%d", self.Index)
		}
		// check if peer pubkey support VRF
		publickey, err := vconfig.Pubkey(p.ID)
		if err != nil || !vrf.ValidatePublicKey(publickey) {
			panic(fmt.Errorf("peer pubkey is ensured to be valid for VRF:%s", p.ID))
		}
	}

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
		log.Infof("updateTimerAndPeerPool add peer index:%v", peerIdx)
	}

	for _, index := range removed {
		if index == self.Index {
			self.Index = math.MaxUint32
			log.Infof("updateTimerAndPeerPool remove index :%d", index)
		} else {
			if C := self.GetPeerMsgChan(index); C != nil {
				log.Infof("updateTimerAndPeerPool remove consensus:index:%d", index)
				C <- nil
			}
		}
	}
}

func (self *Server) initialize() error {
	selfNodeId := vconfig.PubkeyID(self.account.PublicKey)
	log.Infof("server: %s starting", selfNodeId)

	store, err := OpenBlockStore(ledger.DefLedger)
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
	self.rebroadcastC = make(chan uint32, CAP_ACTION_CHANNEL)
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

	self.blockPool, err = newBlockPool(msgHistoryDuration, store)
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
	go self.vbftLoop()

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

func (self *Server) startNewRound() {
	blkNum := self.GetVbftContext().BlockNum

	self.timer.StartEventTimer(EventTxPool, blkNum)
	self.timer.StartEventTimer(EventTxBlockTimeout, blkNum)

	msgs := self.msgPool.DropBftMsgs(blkNum)
	go func() {
		for _, msg := range msgs {
			self.processBftMsgFromPeer(msg, MustHashMsg(msg))
		}
	}()
	return
}

func (self *Server) startNewProposal(vbftCtx *VbftContext) {
	blkNum := vbftCtx.BlockNum
	// make proposal
	if self.isProposer(self.Index) {
		log.Infof("server %d, proposer for block %d", self.Index, blkNum)
		proposal := self.blockPool.GetBlockProposal(blkNum, self.Index)
		if proposal == nil {
			if err := self.makeProposal(blkNum, false); err != nil {
				log.Errorf("server %d failed to making proposal (%d): %s",
					self.Index, blkNum, err)
			}
		}
	} else if vbftCtx.Is2ndProposer(self.Index) {
		log.Infof("server %d, 2nd proposer for block %d", self.Index, blkNum)
		self.timer.StartEventTimer(EventProposalBackoff, blkNum)
	}

	// TODO: if new round block proposal has received, go endorsing/committing directly
	self.timer.StartEventTimer(EventProposeBlockTimeout, blkNum)
}

func (self *Server) processBftMsgFromPeer(msg ConsensusMsg, msgHash common.Uint256) {
	vbftCtx := self.GetVbftContext()
	msgBlkNum := msg.GetBlockNum()
	if msgBlkNum != vbftCtx.BlockNum {
		if msgBlkNum > vbftCtx.BlockNum {
			if err := self.msgPool.AddMsg(msg, msgHash); err != nil {
				if err != errDropFarFutureMsg {
					log.Errorf("failed to add proposal msg (%d) to pool: %s", msgBlkNum, err)
				}
			}
		}
		return
	}
	switch msg.Type() {
	case BlockProposalMessage:
		pMsg := msg.(*blockProposalMsg)
		if err := self.msgPool.AddMsg(msg, msgHash); err != nil {
			log.Errorf("failed to add proposal msg (%d) to pool", msgBlkNum)
			return
		}
		self.processProposalMsg(vbftCtx, pMsg)
	case BlockEndorseMessage, BlockCommitMessage:
		// TODO: verify msg

		// add to msg pool
		if err := self.msgPool.AddMsg(msg, msgHash); err != nil {
			log.Errorf("failed to add msg (%d) to pool", msgBlkNum)
			return
		}
		self.processConsensusMsg(msg)
	default:
		panic(fmt.Errorf("unknown bft msg:%v", msg.Type()))
	}
}

// verify consensus messsage, then send msg to processMsgEvent
func (self *Server) onConsensusMsg(peerIdx uint32, msg ConsensusMsg, msgHash common.Uint256) {
	if self.msgPool.HasMsg(msg, msgHash) && msg.Type() != BlockCommitMessage {
		// dup msg checking
		log.Debugf("dup msg with msg type %d from %d", msg.Type(), peerIdx)
		return
	}

	switch msg.Type() {
	case BlockProposalMessage, BlockEndorseMessage, BlockCommitMessage:
		self.processBftMsgFromPeer(msg, msgHash)
	case PeerHeartbeatMessage:
		pMsg := msg.(*peerHeartbeatMsg)
		self.processHeartbeatMsg(peerIdx, pMsg)
		if pMsg.CommittedBlockNumber+MAX_SYNCING_CHECK_BLK_NUM < self.GetCurrentBlockNo() {
			// delayed peer detected, response heartbeat with our chain Info
			self.heartbeat(peerIdx)
		}

	case ProposalFetchMessage:
		pMsg := msg.(*proposalFetchMsg)
		pmsg := self.blockPool.GetBlockProposal(pMsg.BlockNum, pMsg.ProposerID)
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
		vbftCtx := self.GetVbftContext()
		if vbftCtx.BlockNum > msgBlkNum+1 {
			return
		}
		if err := self.msgPool.AddMsg(msg, msgHash); err != nil {
			if err != errDropFarFutureMsg {
				log.Errorf("failed to add submit msg (%d) to pool: %s", msgBlkNum, err)
			}
			return
		}
		if vbftCtx.BlockNum == msgBlkNum+1 {
			self.CheckAndSubmitBlock(msgBlkNum, vbftCtx.PrevBlockInfo.MerkleRoot)
		}
	}
}

func (self *Server) verifyCrossChainMsg(msg *blockProposalMsg, root common.Uint256) bool {
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

func (self *Server) processProposalMsg(vbftCtx *VbftContext, msg *blockProposalMsg) {
	blkNum := vbftCtx.BlockNum
	blkInfo := vbftCtx.PrevBlockInfo.Info

	msgPrevBlkHash := msg.Block.getPrevBlockHash()
	if vbftCtx.PrevBlockInfo.Block.Hash() != msgPrevBlkHash {
		log.Errorf("BlockPrposalMessage check blocknum:%d,prevhash:%s,msg prevhash:%s", msg.GetBlockNum(), vbftCtx.PrevBlockInfo.Block.Hash().ToHexString(), msgPrevBlkHash.ToHexString())
		self.msgPool.DropMsg(msg)
		return
	}
	configNum := vbftCtx.ConfigNum
	if configNum != math.MaxUint32 && msg.Block.Info.LastConfigBlockNum != configNum {
		log.Errorf("BlockPrposalMessage  check LastConfigBlockNum blocknum:%d,prvLastConfigBlockNum:%d,self LastConfigBlockNum:%d", msg.GetBlockNum(), msg.Block.Info.LastConfigBlockNum, configNum)
		return
	}
	merkleRoot := vbftCtx.PrevBlockInfo.MerkleRoot
	if msg.Block.getPrevExecMerkleRoot() != merkleRoot {
		self.msgPool.DropMsg(msg)
		msgMerkleRoot := msg.Block.getPrevExecMerkleRoot()
		log.Errorf("BlockPrposalMessage check MerkleRoot blocknum:%d,msg MerkleRoot:%s,self MerkleRoot:%s", msg.GetBlockNum(), msgMerkleRoot.ToHexString(), merkleRoot.ToHexString())
		return
	}
	cfg := vconfig.ChainConfig{}
	if blkInfo.NewChainConfig != nil {
		cfg = *blkInfo.NewChainConfig
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

	prevBlockTimestamp := vbftCtx.PrevBlockInfo.Block.Header.Timestamp
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
			self.Index, msg.Block.getProposer(), blkNum)
		self.msgPool.DropMsg(msg)
		return
	}
	if err := verifyVrf(proposerPk, blkNum, blkInfo.VrfValue, msg.Block.getVrfValue(), msg.Block.getVrfProof()); err != nil {
		log.Errorf("server %d failed to verify vrf of block %d proposal from %d",
			self.Index, blkNum, msg.Block.getProposer())
		self.msgPool.DropMsg(msg)
		return
	}
	if !self.verifyCrossChainMsg(msg, vbftCtx.PrevBlockInfo.CrossStatesRoot) {
		log.Errorf("verify cross chain message error:%+v\n", msg.Block.CrossChainMsg)
		self.msgPool.DropMsg(msg)
		return
	}
	txs := msg.Block.Block.Transactions
	if vbftCtx.NeedUpdateChainConfigTx() {
		if len(txs) != 1 || self.CreateGovernaceTransaction(vbftCtx.BlockNum).Hash() != txs[0].Hash() {
			log.Errorf("update chain config block must has 1 commit dpos transaction, blk: %s", blkNum)
			self.msgPool.DropMsg(msg)
			return
		}
	}
	if len(txs) > 0 && !vbftCtx.NeedUpdateChainConfigTx() {
		height := blkNum - 1
		start, end := self.incrValidator.BlockRange()
		validHeight := height
		if blkNum <= end {
			validHeight = start
		} else {
			self.incrValidator.Clean()
			//log.Infof("incr validator block height %v != ledger block height %v", int(end)-1, height)
			log.Infof("incr validator block height %v != ledger block height %v, CompletedBlockNum:%d", int(end)-1, height, vbftCtx.BlockNum-1)
		}
		// start new routine to verify txs in proposal block
		go func() {
			if err := self.poolActor.VerifyBlock(txs, validHeight); err != nil && err != actor.ErrTimeout {
				log.Errorf("server %d verify proposal blk from %d failed, blk %d, txs %d, err: %s",
					self.Index, msg.Block.getProposer(), blkNum, len(txs), err)
				self.msgPool.DropMsg(msg)
				return
			} else if err == actor.ErrTimeout {
				log.Errorf("server %d verify proposal blk from %d timedout, blk %d, txs %d, err: %s",
					self.Index, msg.Block.getProposer(), blkNum, len(txs), err)
			}
			nonceCtx := make(map[common.Address]uint64)
			for _, tx := range txs {
				if err := self.incrValidator.Verify(tx, validHeight, nonceCtx); err != nil {
					log.Errorf("server %d verify proposal tx from %d failed, blk %d, txs %d, err: %s",
						self.Index, msg.Block.getProposer(), blkNum, len(txs), err)
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

func (self *Server) makeProgress(vbftCtx *VbftContext) {
	defer func() {
		log.Infof("bft progress status: %s", self.blockPool.Info(vbftCtx.BlockNum))
	}()
	blkNum := vbftCtx.BlockNum
	proposal := self.blockPool.GetBlockProposal(vbftCtx.BlockNum, self.GetActiveProposer())
	if proposal != nil {
		// stop proposal timer
		self.timer.CancelEventTimer(EventProposeBlockTimeout, blkNum)
		if vbftCtx.IsEndorser(self.Index) {
			if err := self.endorseBlock(proposal, false); err != nil {
				log.Errorf("failed to endorse block proposal (%d): %s", blkNum, err)
			}
		}
	}

	// TODO: should only count endorsements from endorsers
	if !self.blockPool.committedForBlock(blkNum) {
		if proposer, forEmpty, done := self.blockPool.endorseDone(blkNum, self.GetChainConfig().C); done {
			// stop endorse timer
			self.timer.CancelEventTimer(EventEndorseBlockTimeout, blkNum)
			// stop empty endorse timer
			self.timer.CancelEventTimer(EventEndorseEmptyBlockTimeout, blkNum)
			proposal := self.blockPool.GetBlockProposal(blkNum, proposer)
			if proposal == nil {
				self.fetchProposal(blkNum, proposer)
				log.Infof("server %d endorse %d done, waiting proposal from %d", self.Index, blkNum, proposer)
			} else if vbftCtx.IsCommitter(self.Index) {
				// make endorsement
				if err := self.commitBlock(proposal, forEmpty); err != nil {
					log.Errorf("failed to endorse for block %d: %s", blkNum, err)
					return
				}
			}
		}
	}
	if self.blockPool.endorseFailed(blkNum, self.GetChainConfig().C) {
		// endorse failed, start empty endorsing
		self.processTimerEvent(&TimerEvent{evtType: EventEndorseBlockTimeout, blockNum: blkNum})
	}

	chainCfg := vbftCtx.Config
	if proposer, forEmpty, done := self.blockPool.commitDone(vbftCtx, blkNum); done {
		proposal := self.blockPool.GetBlockProposal(blkNum, proposer)
		if proposal == nil {
			log.Infof("server %d commit %d done, waiting proposal", self.Index, blkNum)
			self.fetchProposal(blkNum, proposer)
			self.timer.StartEventTimer(EventCommitBlockTimeout, blkNum)
			return
		}

		if vbftCtx.IsCommitter(self.Index) {
			// make sure committer broadcasting his commit msg
			if err := self.commitBlock(proposal, forEmpty); err != nil {
				log.Errorf("server %d consensused %d, committer broadcast commit msg: %s", self.Index, blkNum, err)
			}
		}
		if !self.blockPool.checkBlockSign(vbftCtx, proposal.Block, forEmpty, chainCfg.N-(chainCfg.N-1)/3) {
			log.Errorf("server %d received commit checkBlockSign insufficient at blk: %d", self.Index, blkNum)
			return
		}
		// stop commit timer
		self.timer.CancelEventTimer(EventCommitBlockTimeout, blkNum)

		if err := self.makeSealed(proposal, forEmpty); err != nil {
			log.Errorf("failed to seal block %d, err: %s", blkNum, err)
		}
	}
}

func (self *Server) processMsgEvent(msg ConsensusMsg) {
	vbftCtx := self.GetVbftContext()
	msgBlkNum := msg.GetBlockNum()
	if msgBlkNum != vbftCtx.BlockNum {
		return
	}
	log.Debugf("server %d process msg, block %d, type %d", self.Index, msg.GetBlockNum(), msg.Type())
	switch msg.Type() {
	case BlockProposalMessage:
		pMsg := msg.(*blockProposalMsg)
		if err := self.blockPool.AddBlockProposal(pMsg); err != nil {
			// TODO: faulty proposer detected
			log.Errorf("failed to add block proposal (%d): %s", msgBlkNum, err)
			return
		}

		if self.Index != pMsg.Block.getProposer() && self.isProposer(self.Index) {
			p := self.blockPool.GetBlockProposal(msgBlkNum, self.Index)
			if p != nil {
				self.broadcast(msg)
			}
		}
	case BlockEndorseMessage:
		pMsg := msg.(*blockEndorseMsg)

		if pMsg.EndorsedProposer != self.Index && self.blockPool.GetBlockProposal(msgBlkNum, pMsg.EndorsedProposer) == nil {
			self.fetchProposal(msgBlkNum, pMsg.EndorsedProposer)
		}

		self.blockPool.AddBlockEndorseMsg(pMsg)
		log.Infof("server %d received endorse from %d, for proposer %d, block %d, empty: %t",
			self.Index, pMsg.Endorser, pMsg.EndorsedProposer, msgBlkNum, pMsg.EndorseForEmpty)
	case BlockCommitMessage:
		pMsg := msg.(*blockCommitMsg)
		if err := self.blockPool.AddBlockCommitMsg(pMsg); err != nil {
			log.Errorf("failed to add commit msg (%d): %s", msgBlkNum, err)
			return
		}

		log.Infof("server %d received commit from %d, for proposer %d, block %d, empty: %t",
			self.Index, pMsg.Committer, pMsg.BlockProposer, msgBlkNum, pMsg.CommitForEmpty)
	}
	self.makeProgress(vbftCtx)
}

func (self *Server) RebroadcastMsgs(blkNum uint32) {
	vbftCtx := self.GetVbftContext()
	if blkNum != vbftCtx.BlockNum {
		return
	}

	proposals := self.blockPool.GetBlockProposals(blkNum)
	for _, p := range proposals {
		if p.Block.getProposer() == self.Index {
			log.Infof("server %d rebroadcast proposal, blk %d", self.Index, blkNum)
			self.broadcast(p)
			break
		}
	}
	if vbftCtx.IsEndorser(self.Index) {
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
			proposal := vbftCtx.GetHighestRankProposal(proposals)
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
	if vbftCtx.IsCommitter(self.Index) {
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
				proposal := self.blockPool.GetBlockProposal(blkNum, proposer)
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
				self.processTimerEvent(&TimerEvent{evtType: EventEndorseBlockTimeout, blockNum: blkNum})
			}
		}
	}
}

func (self *Server) vbftLoop() {
	self.quitWg.Add(1)
	defer self.quitWg.Done()

	for {
		select {
		case blkNum := <-self.rebroadcastC:
			self.RebroadcastMsgs(blkNum)
		case msg := <-self.msgC:
			self.processMsgEvent(msg)
		case evt := <-self.timer.C:
			if err := self.processTimerEvent(evt); err != nil {
				log.Errorf("failed to process timer evt: %d, err: %s", evt.evtType, err)
			}
		case <-self.quitC:
			log.Infof("server %d vbftLoop quit", self.Index)
			return
		}
	}
}

func (self *Server) processTimerEvent(evt *TimerEvent) error {
	vbftCtx := self.GetVbftContext()
	if vbftCtx.BlockNum != evt.blockNum {
		return nil
	}
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
			if err := self.makeProposal(evt.blockNum, false); err != nil {
				return fmt.Errorf("failed to make 2nd proposal (%d): %s", evt.blockNum, err)
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
		return self.handleProposalTimeout(vbftCtx, evt)

	case EventRandomBackoff:
		// 1. if endorsed, return
		// 2. if any valid proposal, endorse on high-priority one (priority from vrf), start endorse timeout, return
		// 3. make empty proposal, broadcast, start 2nd proposal timeout, return
		//
		return self.handleProposalTimeout(vbftCtx, evt)

	case EventPropose2ndBlockTimeout:
		// 1. if endorsed, return
		// 2. there must some valid proposal, if not, force resync, reset peer neighbours
		// 3. endorse on highest-priority one, start endorse timeout, return
		//
		return self.handleProposalTimeout(vbftCtx, evt)

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
			proposal := self.blockPool.GetBlockProposal(evt.blockNum, proposer)
			// consensus ok, make endorsement
			if proposal == nil {
				self.fetchProposal(evt.blockNum, proposer)
				// restart endorsing timer
				self.timer.StartEventTimer(EventEndorseBlockTimeout, evt.blockNum)
				return fmt.Errorf("endorse %d done, but no proposal available", evt.blockNum)
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
		proposal := vbftCtx.GetHighestRankProposal(proposals)
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
			proposal := self.blockPool.GetBlockProposal(evt.blockNum, proposer)

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
				proposal := vbftCtx.GetHighestRankProposal(proposals)
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
		if !self.getState().IsReady() {
			return nil
		}
		chainCfg := vbftCtx.Config
		if proposer, forEmpty, done := self.blockPool.commitDone(vbftCtx, evt.blockNum); done {
			proposal := self.blockPool.GetBlockProposal(evt.blockNum, proposer)
			if proposal == nil {
				self.restartSyncing()
				return fmt.Errorf("commit timeout, consensused proposal not available. need resync")
			}
			if !self.blockPool.checkBlockSign(vbftCtx, proposal.Block, forEmpty, chainCfg.N-(chainCfg.N-1)/3) {
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
	case EventTxPool:
		self.timer.CancelEventTimer(EventTxPool, evt.blockNum)
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
			self.startNewProposal(vbftCtx)
		} else {
			//reset timer, continue waiting txs from txnpool
			self.timer.StartEventTimer(EventTxPool, evt.blockNum)
		}
	case EventTxBlockTimeout:
		self.timer.CancelEventTimer(EventTxPool, evt.blockNum)
		self.timer.CancelEventTimer(EventTxBlockTimeout, evt.blockNum)
		self.startNewProposal(vbftCtx)
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

	self.processMsgEvent(endorseMsg)
	// if node is endorser of current round
	if forEmpty || self.GetVbftContext().IsEndorser(self.Index) {
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

	self.processMsgEvent(commitMsg)
	// if node is committer of current round
	if forEmpty || self.GetVbftContext().IsCommitter(self.Index) {
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

func (self *Server) fastForwardBlock(block *VbftBlock) error {
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
	vbftCtx := self.GetVbftContext()
	if sealedBlkNum < vbftCtx.BlockNum {
		// we already in future round
		log.Errorf("late seal of %d, current blkNum: %d", sealedBlkNum, vbftCtx.BlockNum)
		return nil
	} else if sealedBlkNum > vbftCtx.BlockNum {
		// we have lost sync, restarting syncing
		self.restartSyncing()
		return fmt.Errorf("future seal of %d, current blknum: %d", sealedBlkNum, vbftCtx.BlockNum)
	}

	sealedBlock, result, err := self.blockPool.SetBlockSealed(vbftCtx, block, empty, sigdata)
	if err != nil {
		return fmt.Errorf("failed to seal proposal: %s", err)
	}

	self.timer.OnBlockSealed(sealedBlkNum)
	self.msgPool.OnBlockSealed(sealedBlkNum)

	h := sealedBlock.Block.Hash()
	prevBlkHash := sealedBlock.getPrevBlockHash()
	log.Infof("server %d, sealed block %d, proposer %d, prevhash: %s, hash: %s", self.Index,
		sealedBlkNum, block.getProposer(), prevBlkHash.ToHexString(), h.ToHexString())

	self.updateVbftContext(sealedBlock.Block, sealedBlock.Info, result)
	self.CheckAndSubmitBlock(sealedBlkNum, self.GetVbftContext().PrevBlockInfo.MerkleRoot)
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

func (self *Server) CreateGovernaceTransaction(blkNum uint32) *types.Transaction {
	mutable := utils.BuildNativeTransaction(nutils.GovernanceContractAddress, gover.COMMIT_DPOS, []byte{})
	mutable.Nonce = blkNum
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic(err)
	}
	return tx
}

func (self *VbftContext) NeedUpdateChainConfigTx() bool {
	lastConfigNum := self.PrevBlockInfo.Info.LastConfigBlockNum
	if self.PrevBlockInfo.Info.NewChainConfig != nil {
		lastConfigNum = self.BlockNum - 1
	}
	return (self.BlockNum - lastConfigNum) >= self.Config.MaxBlockChangeView
}

func (self *Server) checkUpdateChainConfig(view uint32, writeSet *overlaydb.MemDB) bool {
	force, err := isUpdate(writeSet, view)
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
	if blkNum <= end {
		validHeight = start
	}
	return validHeight
}

func (self *Server) nonSystxs(sysTxs []*types.Transaction, vbftCtx *VbftContext) bool {
	if vbftCtx.NeedUpdateChainConfigTx() {
		return len(sysTxs) == 1 && self.CreateGovernaceTransaction(vbftCtx.BlockNum).Hash() == sysTxs[0].Hash()
	}
	return true
}

func (self *Server) makeProposal(blkNum uint32, forEmpty bool) error {
	vbftCtx := self.GetVbftContext()
	if blkNum != vbftCtx.BlockNum {
		return fmt.Errorf("server %d ignore deprecatd blk proposal %d, current %d",
			self.Index, blkNum, vbftCtx.BlockNum)
	}

	validHeight := self.validHeight(blkNum)
	sysTxs := make([]*types.Transaction, 0)
	userTxs := make([]*types.Transaction, 0)

	//check need upate chainconfig
	var cfg *vconfig.ChainConfig
	writeSet := vbftCtx.PrevBlockInfo.WriteSet
	if vbftCtx.NeedUpdateChainConfigTx() || self.checkUpdateChainConfig(vbftCtx.Config.View, writeSet) {
		chainconfig, err := getChainConfig(writeSet, blkNum)
		if err != nil {
			return fmt.Errorf("getChainConfig failed:%s", err)
		}
		//add transaction invoke governance native commit_pos contract
		if vbftCtx.NeedUpdateChainConfigTx() {
			tx := self.CreateGovernaceTransaction(blkNum)
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

	proposal, err := self.constructProposalMsg(vbftCtx, sysTxs, userTxs, cfg)
	if err != nil {
		return fmt.Errorf("failed to construct proposal: %s", err)
	}

	log.Infof("server %d make proposal for block %d", self.Index, blkNum)

	// add proposal to self
	h, _ := HashMsg(proposal)
	self.msgPool.AddMsg(proposal, h)
	self.processMsgEvent(proposal)
	self.broadcast(proposal)
	return nil
}

func (self *Server) makeSealed(proposal *blockProposalMsg, forEmpty bool) error {
	blkNum := proposal.GetBlockNum()

	log.Infof("server %d ready to seal block %d, for proposer %d, empty: %t",
		self.Index, blkNum, proposal.Block.getProposer(), forEmpty)

	// seal the block
	if proposal.GetBlockNum() < self.GetCurrentBlockNo() {
		return nil
	}
	// for each round, we can only seal one block
	if err := self.sealBlock(proposal.Block, forEmpty, true); err != nil {
		log.Errorf("server %d failed to seal block (%d): %s",
			self.Index, proposal.GetBlockNum(), err)
		return nil
	}

	self.startNewRound()
	return nil
}

func (self *Server) reBroadcastCurrentRoundMsgs() {
	go func() {
		self.rebroadcastC <- self.GetCurrentBlockNo()
	}()
}

func (self *Server) fetchProposal(blkNum uint32, proposer uint32) {
	msg := self.constructProposalFetchMsg(blkNum, proposer)
	self.msgSendC <- &SendMsgEvent{
		ToPeer: math.MaxUint32,
		Msg:    msg,
	}
}

func (self *Server) handleProposalTimeout(vbftCtx *VbftContext, evt *TimerEvent) error {
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
			if vbftCtx.Is2ndProposer(self.Index) {
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
	proposal := vbftCtx.GetHighestRankProposal(proposals)
	if self.isProposer(proposal.Block.getProposer()) {
		// proposal msg handler will do the endorsement
		return nil
	}

	if err := self.endorseBlock(proposal, false); err != nil {
		log.Errorf("server %d failed to endorse block proposal (%d): %s",
			self.Index, proposal.GetBlockNum(), err)
	}
	return nil
}

func (self *Server) restartSyncing() {
	// send sync request to self.sync, go syncing-state immediately
	// stop all bft timers

	self.stateMgr.StateEventC <- &StateEvent{Type: ForceSyncing}
}
