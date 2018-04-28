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
	"encoding/hex"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/Ontology/account"
	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	actorTypes "github.com/Ontology/consensus/actor"
	vconfig "github.com/Ontology/consensus/vbft/config"
	"github.com/Ontology/core/payload"
	"github.com/Ontology/core/types"
	"github.com/Ontology/crypto"
	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/events/message"
	p2pmsg "github.com/Ontology/net/message"
	"reflect"
	"github.com/Ontology/events"
	"github.com/Ontology/validator/increment"
)

type BftActionType uint8

const (
	MakeProposal BftActionType = iota
	EndorseBlock
	CommitBlock
	SealBlock
	FastForward // for syncer catch up
	ReBroadcast
)

type BftAction struct {
	Type     BftActionType
	Proposal *blockProposalMsg
	forEmpty bool
}

type BlockParticipantConfig struct {
	BlockNum    uint64
	L           uint32
	Vrf         vconfig.VRFValue
	ChainConfig *vconfig.ChainConfig
	Proposers   []uint32
	Endorsers   []uint32
	Committers  []uint32
}

type Server struct {
	Index         uint32
	account       *account.Account
	poolActor     *actorTypes.TxPoolActor
	p2p           *actorTypes.P2PActor
	ledger        *actorTypes.LedgerActor
	incrValidator *increment.IncrementValidator
	pid           *actor.PID

	privateKey keypair.PrivateKey

	// some config
	msgHistoryDuration uint64

	metaLock                 sync.RWMutex
	currentBlockNum          uint64
	config                   *vconfig.ChainConfig
	currentParticipantConfig *BlockParticipantConfig

	chainStore *ChainStore // block store
	msgPool    *MsgPool    // consensus msg pool
	blockPool  *BlockPool  // received block proposals
	peerPool   *PeerPool   // consensus peers
	syncer     *Syncer
	stateMgr   *StateMgr
	timer      *EventTimer

	msgRecvC   map[uint32]chan *p2pmsg.ConsensusPayload
	msgC       chan ConsensusMsg
	bftActionC chan *BftAction
	msgSendC   chan *SendMsgEvent

	sub   *events.ActorSubscriber
	quitC chan interface{}
	quit  bool
}

func NewVbftServer(account *account.Account, txpool, ledger, p2p *actor.PID) (*Server, error) {
	server := &Server{
		msgHistoryDuration: 64,
		account:            account,
		poolActor:          &actorTypes.TxPoolActor{Pool: txpool},
		p2p:                &actorTypes.P2PActor{P2P: p2p},
		ledger:             &actorTypes.LedgerActor{Ledger: ledger},
		incrValidator:      increment.NewIncrementValidator(10),
	}
	server.stateMgr = newStateMgr(server)

	props := actor.FromProducer(func() actor.Actor {
		return server
	})

	pid, err := actor.SpawnNamed(props, "consensus_vbft")
	if err != nil {
		return nil, err
	}
	server.pid = pid
	server.sub = events.NewActorSubscriber(pid)
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
		if err := self.start(); err != nil {
			log.Errorf("vbft start failed: %s", err)
		}
	case *actorTypes.StopConsensus:
		self.stop()
	case *message.SaveBlockCompleteMsg:
		log.Infof("vbft actor receives block complete event. block height=%d, numtx=%d",
			msg.Block.Header.Height, len(msg.Block.Transactions))
		self.handleBlockPersistCompleted(msg.Block)
	case *p2pmsg.PeerStateUpdate:
		self.handlePeerStateUpdate(msg)
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
	self.pid.Tell(&actorTypes.StartConsensus{})
	return nil
}

func (self *Server) Halt() error {
	self.pid.Tell(&actorTypes.StopConsensus{})
	return nil
}

func (self *Server) handlePeerStateUpdate(peer *p2pmsg.PeerStateUpdate) {
	if peer.PeerPubKey == nil {
		log.Errorf("server %d, invalid peer state update (no pk)", self.Index)
		return
	}
	peerID, err := vconfig.PubkeyID(peer.PeerPubKey)
	if err != nil {
		log.Errorf("failed to get peer ID for pubKey: %v", peer.PeerPubKey)
	}
	peerIdx, present := self.peerPool.GetPeerIndex(peerID)
	if !present {
		log.Errorf("invalid consensus node: %s", peerID.String())
	}

	log.Infof("peer state update: %d, connect: %t", peerIdx, peer.Connected)
	if peer.Connected {
		if _, present := self.msgRecvC[peerIdx]; !present {
			self.msgRecvC[peerIdx] = make(chan *p2pmsg.ConsensusPayload, 1024)
		}

		go func() {
			time.Sleep(5 * time.Second)
			if err := self.run(peer.PeerPubKey); err != nil {
				log.Errorf("server %d, processor on peer %d failed: %s",
					self.Index, peerIdx, err)
			}
		}()
	} else {
		if C, present := self.msgRecvC[peerIdx]; present {
			C <- nil
		}
	}
}

func (self *Server) handleBlockPersistCompleted(block *types.Block) {
	log.Infof("persist block: %x", block.Hash())

	self.incrValidator.AddBlock(block)

	// TODO: why this?
	//self.p2p.Broadcast(block.Hash())
}

func (self *Server) NewConsensusPayload(payload *p2pmsg.ConsensusPayload) {
	peerID, err := vconfig.PubkeyID(payload.Owner)
	if err != nil {
		log.Errorf("failed to get peer ID for pubKey: %v", payload.Owner)
	}
	peerIdx, present := self.peerPool.GetPeerIndex(peerID)
	if !present {
		log.Errorf("invalid consensus node: %s", peerID.String())
	}
	if C, present := self.msgRecvC[peerIdx]; present {
		C <- payload
	} else {
		log.Errorf("consensus msg without receiver: %d node: %s", peerIdx, peerID.String())
	}
}

func (self *Server) LoadChainConfig(chainStore *ChainStore) error {
	self.metaLock.Lock()
	defer self.metaLock.Unlock()

	block, err := chainStore.GetBlock(chainStore.GetChainedBlockNum())
	if err != nil {
		return err
	}

	var cfg vconfig.ChainConfig
	if block.getNewChainConfig() != nil {
		cfg = *block.getNewChainConfig()
	} else {
		cfgBlock := block
		if block.getLastConfigBlockNum() != math.MaxUint64 {
			cfgBlock, err = chainStore.GetBlock(block.getLastConfigBlockNum())
			if err != nil {
				return fmt.Errorf("failed to get cfg block: %s", err)
			}
		}
		if cfgBlock.getNewChainConfig() == nil {
			panic("failed to get chain config from config block")
		}
		cfg = *cfgBlock.getNewChainConfig()
	}
	self.config = &cfg
	if self.config.View == 0 {
		panic("invalid view")
	}

	// update msg delays
	makeProposalTimeout = time.Duration(cfg.BlockMsgDelay * 2)
	make2ndProposalTimeout = time.Duration(cfg.BlockMsgDelay)
	endorseBlockTimeout = time.Duration(cfg.HashMsgDelay * 2)
	commitBlockTimeout = time.Duration(cfg.HashMsgDelay * 3)
	peerHandshakeTimeout = time.Duration(cfg.PeerHandshakeTimeout)
	zeroTxBlockTimeout = time.Duration(cfg.BlockMsgDelay * 3)
	// TODO: load sealed blocks from chainStore

	// protected by server.metaLock
	self.currentBlockNum = self.GetCommittedBlockNo() + 1

	log.Infof("committed: %d, current block no: %d", self.GetCommittedBlockNo(), self.GetCurrentBlockNo())

	self.currentParticipantConfig, err = self.buildParticipantConfig(self.GetCurrentBlockNo(), self.config)
	if err != nil {
		return fmt.Errorf("failed to build participant config: %s", err)
	}

	return nil
}

func (self *Server) updateChainConfig() error {
	self.metaLock.Lock()
	defer self.metaLock.Unlock()

	// TODO

	// 1. update peer pool
	// 2. reset all peer connections, create new connections with new peers
	// 3. check connections num for state-mgmt
	// 4. update chainStore

	return nil
}

func (self *Server) start() error {
	// TODO: load config from chain

	// load private key from node config
	self.privateKey = self.account.PrivateKey

	// TODO: configurable log

	store, err := OpenBlockStore(self.ledger)
	if err != nil {
		log.Errorf("failed to open block store: %s", err)
		return fmt.Errorf("failed to open block store: %s", err)
	}
	self.chainStore = store
	log.Info("block store opened")

	self.blockPool, err = newBlockPool(self.msgHistoryDuration, store)
	if err != nil {
		log.Errorf("init blockpool: %s", err)
		return fmt.Errorf("init blockpool: %s", err)
	}
	self.msgPool = newMsgPool(self, self.msgHistoryDuration)
	self.peerPool = NewPeerPool(0, self) // FIXME: maxSize
	self.timer = NewEventTimer(self)
	self.syncer = newSyncer(self)

	self.msgRecvC = make(map[uint32]chan *p2pmsg.ConsensusPayload)
	self.msgC = make(chan ConsensusMsg, 64)
	self.bftActionC = make(chan *BftAction, 8)
	self.msgSendC = make(chan *SendMsgEvent, 16)
	self.quitC = make(chan interface{})
	if err := self.LoadChainConfig(store); err != nil {
		log.Errorf("failed to load config: %s", err)
		return fmt.Errorf("failed to load config: %s", err)
	}
	log.Infof("chain config loaded from local, current blockNum: %d", self.GetCurrentBlockNo())

	// add all consensus peers to peer_pool
	for _, p := range self.config.Peers {
		self.peerPool.addPeer(p)
		log.Infof("added peer: %s", p.ID.String())
	}

	id, _ := vconfig.PubkeyID(self.account.PublicKey)
	self.Index, _ = self.peerPool.GetPeerIndex(id)
	self.sub.Subscribe(message.TopicSaveBlockComplete)
	go self.syncer.run()
	go self.stateMgr.run()
	go self.msgSendLoop()
	go self.timerLoop()
	go self.actionLoop()
	go func() {
		for {
			if err := self.processMsgEvent(); err != nil {
				log.Errorf("server %d: %s", self.Index, err)
			}
			if self.quit {
				break
			}
		}
	}()

	self.stateMgr.StateEventC <- &StateEvent{
		Type: ConfigLoaded,
	}

	log.Infof("peer %d started", self.Index)

	return nil
}

func (self *Server) stop() error {

	// TODO
<<<<<<< HEAD
	self.sub.Unsubscribe(message.TopicSaveBlockComplete)
=======
	self.incrValidator.Clean()
	self.sub.Unsubscribe(message.TOPIC_SAVE_BLOCK_COMPLETE)
>>>>>>> 44663ad... add tx validator
	return nil
}

func (self *Server) run(peerPubKey *crypto.PubKey) error {
	peerID, err := vconfig.PubkeyID(peerPubKey)
	if err != nil {
		return fmt.Errorf("failed to get peer ID for pubKey: %v", peerPubKey)
	}
	peerIdx, present := self.peerPool.GetPeerIndex(peerID)
	if !present {
		return fmt.Errorf("invalid consensus node: %s", peerID.String())
	}
	if self.peerPool.isNewPeer(peerIdx) {
		if err := self.peerPool.peerConnected(peerIdx); err != nil {
			return err
		}

		// initHandshake
		if err := self.initHandshake(peerIdx, peerPubKey); err != nil {
			log.Errorf("failed to initHandshake with peer %s: %s", peerID, err)
			return err
		}

		log.Infof("server %d: completed handshake with peer %d, %s", self.Index, peerIdx,
			peerID.String())

		// new peer connected
		self.timer.startPeerTicker(peerIdx)
		self.stateMgr.StateEventC <- &StateEvent{
			Type: UpdatePeerState,
			peerState: &PeerState{
				peerIdx:           peerIdx,
				connected:         true,
				chainConfigView:   self.peerPool.getPeer(peerIdx).handShake.ChainConfig.View,
				committedBlockNum: self.peerPool.getPeer(peerIdx).handShake.CommittedBlockNumber,
			},
		}
	} else {
		return fmt.Errorf("dup peer")
	}

	defer func() {
		// TODO: handle peer disconnection here

		self.peerPool.peerDisconnected(peerIdx)
		self.stateMgr.StateEventC <- &StateEvent{
			Type: UpdatePeerState,
			peerState: &PeerState{
				peerIdx:   peerIdx,
				connected: false,
			},
		}
		delete(self.msgRecvC, peerIdx)
	}()

	errC := make(chan error)
	go func() {
		for {
			msgData, err := self.receiveFromPeer(peerIdx)
			if err != nil {
				errC <- err
				return
			}
			msg, err := DeserializeVbftMsg(msgData)

			if err != nil {
				log.Errorf("server %d failed to deserialize vbft msg (len %d): %s", self.Index, len(msgData), err)
			} else {
				if err := msg.Verify(peerPubKey); err != nil {
					log.Errorf("server %d failed to verify msg, type %d, err: %s",
						self.Index, msg.Type(), err)
					return
				}

				if msg.Type() < 4 {
					log.Debugf("server %d received consensus msg, type: %d", self.Index, msg.Type())
				}
				self.onConsensusMsg(peerIdx, msg)
			}
		}
	}()

	return <-errC
}

func (self *Server) getState() ServerState {
	return self.stateMgr.getState()
}

func (self *Server) updateParticipantConfig() error {
	blkNum := self.GetCurrentBlockNo()

	var err error
	self.metaLock.Lock()
	if cfg, err := self.buildParticipantConfig(blkNum, self.config); err == nil {
		self.currentParticipantConfig = cfg
	}
	self.metaLock.Unlock()

	if err != nil {
		return fmt.Errorf("failed to build participant config (%d): %s", blkNum, err)
	}

	// TODO: if server is not in new config, self.stop()
	return nil
}

func (self *Server) startNewRound() error {
	blkNum := self.GetCurrentBlockNo()

	if err := self.updateParticipantConfig(); err != nil {
		return err
	}

	// check proposals in msgpool
	var proposal *blockProposalMsg
	if proposals := self.msgPool.GetProposalMsgs(blkNum); len(proposals) > 0 {
		for _, p := range proposals {
			msg := p.(*blockProposalMsg)
			if msg == nil {
				continue
			}
			if self.isProposer(blkNum, msg.Block.getProposer()) {
				// get proposal from proposer, process it
				proposal = msg
			} else {
				// add other proposals to blockpool
				if err := self.blockPool.newBlockProposal(msg); err != nil {
					log.Errorf("starting new round, failed to add proposal from %d: %s",
						msg.Block.getProposer(), err)
				}
			}
		}
	}

	if endorses := self.msgPool.GetEndorsementsMsgs(blkNum); len(endorses) > 0 {
		for _, e := range endorses {
			msg := e.(*blockEndorseMsg)
			if msg == nil {
				continue
			}
			if err := self.blockPool.newBlockEndorsement(msg); err != nil {
				log.Infof("starting new round, failed to add endorse, blk %d, endorse for %d: %s",
					blkNum, msg.EndorsedProposer, err)
			}
		}
		if proposal == nil {
			if p, proposer, done := self.blockPool.endorseDone(blkNum, self.config.F); done {
				if p != nil {
					proposal = p
				} else {
					self.fetchProposal(blkNum, proposer)
				}
			}
		}
	}
	if commits := self.msgPool.GetCommitMsgs(blkNum); len(commits) > 0 {
		for _, c := range commits {
			msg := c.(*blockCommitMsg)
			if msg == nil {
				continue
			}
			if err := self.blockPool.newBlockCommitment(msg); err != nil {
				log.Infof("start new round, failed to add commit, blk %d, commit for %d: %s",
					blkNum, msg.BlockProposer, err)
			}
		}
	}
	if proposal != nil {
		self.processConsensusMsg(proposal)
		return nil
	}

	txpool := self.poolActor.GetTxnPool(true, uint32(blkNum-1))
	if len(txpool) != 0 {
		self.packProposal(blkNum)
	} else {
		self.timer.startTxTicker(blkNum)
	}
	return nil
}

func (self *Server) packProposal(blkNum uint64) error {
	// make proposal
	if self.isProposer(blkNum, self.Index) {
		log.Infof("server %d, proposer for block %d", self.Index, blkNum)
		// FIXME: possible deadlock on channel
		self.bftActionC <- &BftAction{
			Type:     MakeProposal,
			forEmpty: false,
		}
	} else if self.is2ndProposer(blkNum, self.Index) {
		log.Infof("server %d, 2nd proposer for block %d", self.Index, blkNum)
		self.timer.StartProposalBackoffTimer(blkNum)
	}

	// TODO: if new round block proposal has received, go endorsing/committing directly

	return self.timer.StartProposalTimer(blkNum)
}

// verify consensus messsage, then send msg to processMsgEvent
func (self *Server) onConsensusMsg(peerIdx uint32, msg ConsensusMsg) {

	if self.msgPool.HasMsg(msg) {
		// dup msg checking
		log.Debugf("dup msg with msg type %d from %d", msg.Type(), peerIdx)
		return
	}

	switch msg.Type() {
	case BlockProposalMessage:
		pMsg, ok := msg.(*blockProposalMsg)
		if !ok {
			log.Error("invalid msg with proposal msg type")
			return
		}
		// TODO: verify msg

		msgBlkNum := pMsg.GetBlockNum()
		if msgBlkNum > self.GetCurrentBlockNo() {
			// for concurrency, support two active consensus round
			if err := self.msgPool.AddMsg(msg); err != nil {
				log.Errorf("failed to add proposal msg (%d) to pool", msgBlkNum)
				return
			}

			if isReady(self.getState()) {
				// set the peer as syncing-check trigger from current round
				// start syncing check from current round
				self.syncer.syncCheckReqC <- &SyncCheckReq{
					msg:      msg,
					blockNum: self.GetCurrentBlockNo(),
				}
			}
		} else if msgBlkNum < self.GetCurrentBlockNo() {

			if msgBlkNum <= self.GetCommittedBlockNo() {
				if msgBlkNum+MAX_SYNCING_CHECK_BLK_NUM < self.GetCommittedBlockNo() {
					log.Infof("server %d get proposal msg for block %d, from %d, current committed %d",
						self.Index, msgBlkNum, pMsg.Block.getProposer(), self.GetCommittedBlockNo())
					self.timer.C <- &TimerEvent{
						evtType:  EventPeerHeartbeat,
						blockNum: uint64(pMsg.Block.getProposer()),
					}
				}
				return
			}

			if isReady(self.getState()) {
				// start syncing check from proposed block round
				self.syncer.syncCheckReqC <- &SyncCheckReq{
					msg:      msg,
					blockNum: msgBlkNum,
				}
			}

		} else {
			if err := self.msgPool.AddMsg(msg); err != nil {
				log.Errorf("failed to add proposal msg (%d) to pool", msgBlkNum)
				return
			}
			self.processConsensusMsg(msg)
		}

	case BlockEndorseMessage:
		pMsg, ok := msg.(*blockEndorseMsg)
		if !ok {
			log.Error("invalid msg with endorse msg type")
			return
		}

		// TODO: verify msg

		msgBlkNum := pMsg.GetBlockNum()
		if msgBlkNum > self.GetCurrentBlockNo() {
			// for concurrency, support two active consensus round
			if err := self.msgPool.AddMsg(msg); err != nil {
				log.Errorf("failed to add endorse msg (%d) to pool", msgBlkNum)
				return
			}

			if isReady(self.getState()) {
				// set the peer as syncing-check trigger from current round
				// start syncing check from current round
				self.syncer.syncCheckReqC <- &SyncCheckReq{
					msg:      msg,
					blockNum: self.GetCurrentBlockNo(),
				}
			}
		} else if msgBlkNum < self.GetCurrentBlockNo() {
			if msgBlkNum <= self.GetCommittedBlockNo() {
				if msgBlkNum+MAX_SYNCING_CHECK_BLK_NUM < self.GetCommittedBlockNo() {
					log.Infof("server %d get endorse msg for block %d, from %d, current committed %d",
						self.Index, msgBlkNum, pMsg.Endorser, self.GetCommittedBlockNo())
					self.timer.C <- &TimerEvent{
						evtType:  EventPeerHeartbeat,
						blockNum: uint64(pMsg.Endorser),
					}
				}
				return
			}

			if isReady(self.getState()) {
				// start syncing check from proposed block round
				self.syncer.syncCheckReqC <- &SyncCheckReq{
					msg:      msg,
					blockNum: msgBlkNum,
				}
			}

		} else {
			// add to msg pool
			if err := self.msgPool.AddMsg(msg); err != nil {
				log.Errorf("failed to add endorse msg (%d) to pool", msgBlkNum)
				return
			}
			self.processConsensusMsg(msg)
		}

	case BlockCommitMessage:
		pMsg, ok := msg.(*blockCommitMsg)
		if !ok {
			log.Error("invalid msg with commit msg type")
			return
		}

		// TODO: verify msg

		msgBlkNum := pMsg.GetBlockNum()
		if msgBlkNum > self.GetCurrentBlockNo() {
			if err := self.msgPool.AddMsg(msg); err != nil {
				log.Errorf("failed to add commit msg (%d) to pool", msgBlkNum)
				return
			}

			if isReady(self.getState()) {
				// set the peer as syncing-check trigger from current round
				// start syncing check from current round
				self.syncer.syncCheckReqC <- &SyncCheckReq{
					msg:      msg,
					blockNum: self.GetCurrentBlockNo(),
				}
			}

		} else if msgBlkNum < self.GetCurrentBlockNo() {
			if msgBlkNum <= self.GetCommittedBlockNo() {
				if msgBlkNum+MAX_SYNCING_CHECK_BLK_NUM < self.GetCommittedBlockNo() {
					log.Infof("server %d get commit msg for block %d, from %d, current committed %d",
						self.Index, msgBlkNum, pMsg.Committer, self.GetCommittedBlockNo())
					self.timer.C <- &TimerEvent{
						evtType:  EventPeerHeartbeat,
						blockNum: uint64(pMsg.Committer),
					}
				}
				return
			}

			if isReady(self.getState()) {
				// start syncing check from proposed block round
				self.syncer.syncCheckReqC <- &SyncCheckReq{
					msg:      msg,
					blockNum: msgBlkNum,
				}
			}
		} else {
			// add to msg pool
			if err := self.msgPool.AddMsg(msg); err != nil {
				log.Errorf("failed to add commit msg (%d) to pool", msgBlkNum)
				return
			}
			self.processConsensusMsg(msg)
		}
	case PeerHeartbeatMessage:
		pMsg, ok := msg.(*peerHeartbeatMsg)
		if !ok {
			log.Errorf("invalid msg with heartbeat msg type")
			return
		}
		if err := self.processHeartbeatMsg(peerIdx, pMsg); err != nil {
			log.Errorf("server %d, failed to process heartbeat %d: %s", self.Index, peerIdx, err)
		}
		if pMsg.CommittedBlockNumber+MAX_SYNCING_CHECK_BLK_NUM < self.GetCommittedBlockNo() {
			// delayed peer detected, response heartbeat with our chain Info
			self.timer.C <- &TimerEvent{
				evtType:  EventPeerHeartbeat,
				blockNum: uint64(peerIdx),
			}
		}

	case ProposalFetchMessage:
		pMsg, ok := msg.(*proposalFetchMsg)
		if !ok {
			log.Errorf("invalid msg with proposal fetch msg type")
			return
		}
		var pmsg *blockProposalMsg
		pMsgs := self.msgPool.GetProposalMsgs(self.GetCurrentBlockNo())
		for _, msg := range pMsgs {
			p := msg.(*blockProposalMsg)
			if p != nil && p.Block.getProposer() == self.Index {
				log.Infof("server %d rebroadcast proposal to %d, blk %d",
					self.Index, peerIdx, p.Block.getBlockNum())
				pmsg = p
			}
		}

		if pmsg == nil {
			blk, _ := self.blockPool.getSealedBlock(pMsg.BlockNum)
			if blk != nil {
				pmsg = &blockProposalMsg{
					Block: blk,
				}
			}
		}
		if pmsg != nil {
			self.msgSendC <- &SendMsgEvent{
				ToPeer: peerIdx,
				Msg:    pmsg,
			}
		} else {
			log.Errorf("server %d, failed to handle proposal fetch %d from %d: %s",
				self.Index, pMsg.BlockNum, peerIdx)
		}

	case BlockFetchMessage:
		// handle block fetch msg
		pMsg, ok := msg.(*blockFetchMsg)
		if !ok {
			log.Errorf("invalid msg with blockfetch msg type")
			return
		}
		blk, blkHash := self.blockPool.getSealedBlock(pMsg.BlockNum)
		msg, err := self.constructBlockFetchRespMsg(pMsg.BlockNum, blk, blkHash)
		if err != nil {
			log.Errorf("server %d, failed to handle blockfetch %d from %d: %s",
				self.Index, pMsg.BlockNum, peerIdx, err)
		} else {
			self.msgSendC <- &SendMsgEvent{
				ToPeer: peerIdx,
				Msg:    msg,
			}
		}

	case BlockFetchRespMessage:
		self.syncer.syncMsgC <- &SyncMsg{
			fromPeer: peerIdx,
			msg:      msg,
		}

	case BlockInfoFetchMessage:
		// handle block Info fetch msg
		pMsg, ok := msg.(*BlockInfoFetchMsg)
		if !ok {
			log.Errorf("invalid msg with blockinfo fetch msg type")
			return
		}
		maxCnt := 64
		blkInfos := make([]*BlockInfo_, 0)
		targetBlkNum := self.GetCommittedBlockNo()
		for startBlkNum := pMsg.StartBlockNum; startBlkNum <= targetBlkNum; startBlkNum++ {
			blk, _ := self.blockPool.getSealedBlock(startBlkNum)
			if blk == nil {
				break
			}
			blkInfos = append(blkInfos, &BlockInfo_{
				BlockNum: startBlkNum,
				Proposer: blk.getProposer(),
			})
			if len(blkInfos) >= maxCnt {
				break
			}
		}
		msg, err := self.constructBlockInfoFetchRespMsg(blkInfos)
		if err != nil {
			log.Errorf("server %d, failed to handle blockinfo fetch %d to %d: %s",
				self.Index, pMsg.StartBlockNum, peerIdx, err)
		} else {
			log.Infof("server %d, response blockinfo fetch to %d, blk %d, len %d",
				self.Index, peerIdx, pMsg.StartBlockNum, len(blkInfos))
			self.msgSendC <- &SendMsgEvent{
				ToPeer: peerIdx,
				Msg:    msg,
			}
		}

	case BlockInfoFetchRespMessage:
		self.syncer.syncMsgC <- &SyncMsg{
			fromPeer: peerIdx,
			msg:      msg,
		}
	}
}

func (self *Server) processConsensusMsg(msg ConsensusMsg) {
	if isReady(self.getState()) {
		self.msgC <- msg
	}
}

func (self *Server) processMsgEvent() error {
	select {
	case msg := <-self.msgC:

		log.Debugf("server %d process msg, block %d, type %d, current blk %d",
			self.Index, msg.GetBlockNum(), msg.Type(), self.GetCurrentBlockNo())

		switch msg.Type() {
		case BlockProposalMessage:
			pMsg := msg.(*blockProposalMsg)

			if err := self.validateTxsInProposal(pMsg); err != nil {
				// TODO: report faulty proposal
				return fmt.Errorf("failed to validate tx in proposal: %s", err)
			}

			msgBlkNum := pMsg.GetBlockNum()
			if msgBlkNum == self.GetCurrentBlockNo() {
				// add proposal to block-pool
				if err := self.blockPool.newBlockProposal(pMsg); err != nil {
					if err == errDupProposal {
						// TODO: faulty proposer detected
					}
					return fmt.Errorf("failed to add block proposal (%d): %s", msgBlkNum, err)
				}

				if self.isProposer(msgBlkNum, pMsg.Block.getProposer()) {
					// check if agreed on prev-blockhash
					prevBlk, prevBlkHash := self.blockPool.getSealedBlock(msgBlkNum - 1)
					if prevBlk == nil {
						// TODO: has no candidate proposals for prevBlock, should restart syncing
						return fmt.Errorf("failed to get prevBlock of current round (%d)", msgBlkNum)
					}
					prevBlkHash2 := pMsg.Block.getPrevBlockHash()
					if bytes.Compare(prevBlkHash.ToArray(), prevBlkHash2.ToArray()) != 0 {
						// continue waiting for more proposals
						// FIXME
						return fmt.Errorf("inconsistent prev-block hash with proposer (%d)", msgBlkNum)
					}

					// stop proposal timer
					if err := self.timer.CancelProposalTimer(msgBlkNum); err != nil {
						log.Errorf("failed to cancel proposal timer, blockNum %d, err: %s", msgBlkNum, err)
					}

					if self.isEndorser(msgBlkNum, self.Index) {
						if err := self.endorseBlock(pMsg, false); err != nil {
							log.Errorf("failed to endorse block proposal (%d): %s", msgBlkNum, err)
						}
					}
				} else {
					// nothing to do.
					// makeProposalTimeout handles non-leader proposals
				}
			} else {
				// process new proposal when
				// 1. we have endorsed for current BlockNum
				// 2. proposal is from next potential-leader

				// TODO
			}

		case BlockEndorseMessage:
			pMsg := msg.(*blockEndorseMsg)
			msgBlkNum := pMsg.GetBlockNum()

			// if had committed for current round, ignore the endorsement
			if self.blockPool.committedForBlock(msgBlkNum) {
				return nil
			}

			if msgBlkNum == self.GetCurrentBlockNo() {
				// add endorse to block-pool
				if err := self.blockPool.newBlockEndorsement(pMsg); err != nil {
					return fmt.Errorf("failed to add endorsement (%d): %s", msgBlkNum, err)
				}
				log.Infof("server %d received endorse from %d, for proposer %d, block %d, empty: %t",
					self.Index, pMsg.Endorser, pMsg.EndorsedProposer, msgBlkNum, pMsg.EndorseForEmpty)

				if self.isEndorser(msgBlkNum, pMsg.Endorser) {
					//              if countOfEndrosement(msg.proposal) >= 2F + 1:
					//                      stop WaitEndorsementTimer
					//                      commitBlock(msg.BlockHash)
					//              else if WaitEndorsementTimer has not started:
					//                      start WaitEndorsementTimer

					// TODO: should only count endorsements from endorsers
					if proposal, proposer, done := self.blockPool.endorseDone(msgBlkNum, self.config.F); done {
						// stop endorse timer
						if err := self.timer.CancelEndorseMsgTimer(msgBlkNum); err != nil {
							log.Errorf("failed to cancel endorse timer, blockNum %d, err: %s", msgBlkNum, err)
						}
						// stop empty endorse timer
						if err := self.timer.CancelEndorseEmptyBlockTimer(msgBlkNum); err != nil {
							log.Errorf("failed to cancel empty endorse timer, blockNum %d, err: %s", msgBlkNum, err)
						}

						if proposal == nil {
							self.fetchProposal(msgBlkNum, proposer)
							log.Infof("server %d endorse %d done without proposal, fetching proposal", self.Index, msgBlkNum)
						} else if self.isCommitter(msgBlkNum, self.Index) {
							// make endorsement
							if err := self.makeCommitment(proposal, msgBlkNum); err != nil {
								return fmt.Errorf("failed to endorse for block %d: %s", msgBlkNum, err)
							}
						}
					} else {
						// wait until endorse timeout
					}

				} else {
					// nothing to do
					// makeEndorsementTimeout handles non-endorser endorsements
				}
			} else {
				// process new endorsement when
				// 1. we have committed for current BlockNum
				// 2. endorsed proposal is from next potential-leader
			}

		case BlockCommitMessage:
			pMsg := msg.(*blockCommitMsg)
			msgBlkNum := pMsg.GetBlockNum()

			if msgBlkNum == self.GetCurrentBlockNo() {
				//              if countOfCommitment(msg.proposal) >= 2F + 1:
				//                      stop WaitCommitsTimer
				//                      sealProposal(msg.BlockHash)
				//              else if WaitCommitsTimer has not started:
				//                      start WaitCommitsTimer
				if err := self.blockPool.newBlockCommitment(pMsg); err != nil {
					return fmt.Errorf("failed to add commit msg (%d): %s", msgBlkNum, err)
				}

				log.Infof("server %d received commit from %d, for proposer %d, block %d, empty: %t",
					self.Index, pMsg.Committer, pMsg.BlockProposer, msgBlkNum, pMsg.CommitForEmpty)

				if !self.blockPool.isCommitHadDone(msgBlkNum) {
					if proposal, forEmpty, done := self.blockPool.commitDone(msgBlkNum, self.config.F); done {
						if proposal == nil {
							// TODO: commit done, but we not have the proposal, should request proposal from neighbours
							//       commitTimeout handle this
							return fmt.Errorf("commit done, but proposal not available")
						}

						// stop commit timer
						if err := self.timer.CancelCommitMsgTimer(msgBlkNum); err != nil {
							log.Errorf("failed to cancel endorse timer, blockNum: %d, err: %s", msgBlkNum, err)
						}

						return self.makeSealed(proposal, forEmpty)
					} else {
						// wait commit timeout, nothing to do
					}
				}

			} else {
				// nothing to do besides adding to msg pool

				// FIXME: add msg from msg-pool to block-pool when starting new block-round
			}
		}
	case <-self.quitC:
		self.quit = true
	}
	return nil
}

func (self *Server) actionLoop() {
	for {
		select {
		case action := <-self.bftActionC:
			switch action.Type {
			case MakeProposal:
				// this may triggered when block sealed or random backoff of 2nd proposer
				blkNum := self.GetCurrentBlockNo()

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
					continue
				}

			case CommitBlock:
				blkNum := action.Proposal.GetBlockNum()
				if err := self.commitBlock(action.Proposal, action.forEmpty); err != nil {
					log.Errorf("server %d failed to commit block proposal (%d): %s",
						self.Index, blkNum, err)
					continue
				}
			case SealBlock:
				if action.Proposal.GetBlockNum() < self.GetCurrentBlockNo() {
					continue
				}
				if err := self.sealProposal(action.Proposal, action.forEmpty); err != nil {
					log.Errorf("%d failed to seal block (%d): %s",
						self.Index, action.Proposal.GetBlockNum(), err)
				}
			case FastForward:
				// 1. from current block num, check commit msgs in msg pool
				// 2. if commit consensused, seal the proposal
				for {
					blkNum := self.GetCurrentBlockNo()
					F := int(self.config.F)

					if err := self.updateParticipantConfig(); err != nil {
						log.Errorf("server %d update config failed in forwarding: %s", self.Index, err)
					}

					// get pending msgs from msgpool
					pMsgs := self.msgPool.GetProposalMsgs(blkNum)
					cMsgs := self.msgPool.GetCommitMsgs(blkNum)
					if len(cMsgs) <= F || len(pMsgs) == 0 {
						for _, msg := range pMsgs {
							p := msg.(*blockProposalMsg)
							if p != nil {
								if err := self.blockPool.newBlockProposal(p); err != nil {
									log.Errorf("server %d failed add proposal in fastforwarding: %s",
										self.Index, err)
								}
							}
						}
						for _, msg := range cMsgs {
							c := msg.(*blockCommitMsg)
							if c != nil {
								if err := self.blockPool.newBlockCommitment(c); err != nil {
									log.Errorf("server %d failed to add commit in fastforwarding: %s",
										self.Index, err)
								}
							}
						}

						// catcher doesnot participant in the latest round before catched up
						// rebroadcast timer will reinit the consensus when halt
						if err := self.catchConsensus(blkNum); err != nil {
							log.Infof("server %d fastforward done, catch consensus: %s", self.Index, err)
						}

						log.Infof("server %d fastforward done at blk %d, no msg", self.Index, blkNum)
						break
					}
					log.Infof("server %d fastforwarding from %d, (%d, %d)",
						self.Index, self.GetCurrentBlockNo(), len(cMsgs), len(pMsgs))

					// convert to blockCommitMsg array
					commitMsgs := make([]*blockCommitMsg, 0)
					for _, m := range cMsgs {
						c, ok := m.(*blockCommitMsg)
						if !ok {
							continue
						}
						commitMsgs = append(commitMsgs, c)
					}

					// check if consensused
					proposer, forEmpty := getCommitConsensus(commitMsgs, F)
					if proposer == math.MaxUint32 {
						if err := self.catchConsensus(blkNum); err != nil {
							log.Infof("server %d fastforward done, catch consensus: %s", self.Index, err)
						}
						log.Errorf("server %d fastforward done at blk %d, no consensus", self.Index, blkNum)
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
						log.Errorf("server %d fastforward stopped at blk %d, no proposal", self.Index, blkNum)
						break
					}

					log.Infof("server %d fastforwarding block %d, proposer %d",
						self.Index, blkNum, proposal.Block.getProposer())

					// fastforward the block
					if err := self.sealBlock(proposal.Block, forEmpty); err != nil {
						log.Errorf("server %d fastforward stopped at blk %d, seal failed: %s",
							self.Index, blkNum, err)
						break
					}
				}

			case ReBroadcast:
				blkNum := self.GetCurrentBlockNo()
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
				if self.isEndorser(blkNum, self.Index) {
					endorsed := false
					eMsgs := self.msgPool.GetEndorsementsMsgs(blkNum)
					for _, msg := range eMsgs {
						e := msg.(*blockEndorseMsg)
						if e != nil && e.Endorser == self.Index {
							log.Infof("server %d rebroadcast endorse, blk %d for %d, %t",
								self.Index, e.GetBlockNum(), e.EndorsedProposer, e.EndorseForEmpty)
							self.broadcast(e)
							endorsed = true
						}
					}
					if !endorsed {
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
				} else if proposal, forEmpty := self.blockPool.getEndorsedProposal(blkNum); proposal != nil {
					// construct endorse msg
					blkHash, _ := HashBlock(proposal.Block)
					if endorseMsg, _ := self.constructEndorseMsg(proposal, blkHash, forEmpty); endorseMsg != nil {
						self.broadcast(endorseMsg)
					}
				}
				if self.isCommitter(blkNum, self.Index) {
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
						if proposal, proposer, done := self.blockPool.endorseDone(blkNum, self.config.F); done {
							// consensus ok, make endorsement
							if proposal == nil {
								self.fetchProposal(blkNum, proposer)
								// restart endorsing timer
								self.timer.StartEndorsingTimer(blkNum)
								log.Errorf("server %d endorse %d done, but no proposal", self.Index, blkNum)
							} else if err := self.makeCommitment(proposal, blkNum); err != nil {
								log.Errorf("server %d failed to commit block %d on rebroadcasting: %s",
									self.Index, blkNum, err)
							}
						}
					}
				}
			}
		}
	}
}

func (self *Server) timerLoop() {
	for {
		select {
		case evt := <-self.timer.C:
			if err := self.processTimerEvent(evt); err != nil {
				log.Errorf("failed to process timer evt: %d, err: %s", evt.evtType, err)
			}

		case <-self.quitC:
			break
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

		if self.blockPool.endorsedForBlock(evt.blockNum) {
			return nil
		}
		if !isReady(self.getState()) {
			return nil
		}
		proposals := self.blockPool.getBlockProposals(evt.blockNum)
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
		if !isReady(self.getState()) {
			return nil
		}
		if proposal, proposer, done := self.blockPool.endorseDone(evt.blockNum, self.config.F); done {
			// consensus ok, make endorsement
			if proposal == nil {
				self.fetchProposal(evt.blockNum, proposer)
				// restart endorsing timer
				self.timer.StartEndorsingTimer(evt.blockNum)
				return fmt.Errorf("endorse %d done, but no proposal available", evt.blockNum)
			} else if err := self.makeCommitment(proposal, evt.blockNum); err != nil {
				return fmt.Errorf("failed to endorse for block %d on endorse timeout: %s", evt.blockNum, err)
			}
			return nil
		}
		if !isActive(self.getState()) {
			// not active yet, waiting active peers making decision
			return nil
		}
		if self.blockPool.endorsedForEmptyBlock(evt.blockNum) {
			return nil
		}
		proposals := self.blockPool.getBlockProposals(evt.blockNum)
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
		if !isReady(self.getState()) {
			return nil
		}
		if proposal, proposer, done := self.blockPool.endorseDone(evt.blockNum, self.config.F); done {
			// consensus ok, make endorsement
			if proposal == nil {
				self.fetchProposal(evt.blockNum, proposer)
				// restart timer
				self.timer.StartEndorseEmptyBlockTimer(evt.blockNum)
			} else if err := self.makeCommitment(proposal, evt.blockNum); err != nil {
				return fmt.Errorf("failed to endorse for block %d on empty endorse timeout: %s", evt.blockNum, err)
			}
			return nil
		} else {
			log.Errorf("server %d: empty endorse timeout, no quorum", self.Index)
			if !isActive(self.getState()) {
				proposals := self.blockPool.getBlockProposals(evt.blockNum)
				proposal := self.getHighestRankProposal(evt.blockNum, proposals)
				if proposal != nil {
					if err := self.endorseBlock(proposal, true); err != nil {
						return fmt.Errorf("failed to endorse block proposal (%d): %s", evt.blockNum, err)
					}
				}
			} else {
				if err := self.timer.StartEndorseEmptyBlockTimer(evt.blockNum); err != nil {
					return fmt.Errorf("failed to start empty endorse timer (%d): %s", evt.blockNum, err)
				}
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
		if !isReady(self.getState()) {
			return nil
		}
		if !self.blockPool.isCommitHadDone(evt.blockNum) {
			if proposal, forEmpty, done := self.blockPool.commitDone(evt.blockNum, self.config.F); done {
				if proposal == nil {
					self.restartSyncing()
					return fmt.Errorf("commit timeout, consensused proposal not available. need resync")
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
		//	build heartbeat msg
		msg, err := self.constructHeartbeatMsg()
		if err != nil {
			return fmt.Errorf("failed to build heartbeat msg: %s", err)
		}

		// TODO: check remote peer lastUpdateTime
		log.Debugf("send heartbeat to peer: %d", evt.blockNum)

		//	send to peer
		self.msgSendC <- &SendMsgEvent{
			ToPeer: uint32(evt.blockNum),
			Msg:    msg,
		}
	case EventTxPool:
		blockNum := self.GetCurrentBlockNo()
		txpool := self.poolActor.GetTxnPool(true, uint32(blockNum-1))
		if len(txpool) != 0 {
			self.packProposal(blockNum)
			self.timer.stopTxTicker()
		} else {
			self.timer.stopTxTicker()
			self.timer.StartTxBlockTimeout(blockNum)
		}
	case EventTxBlockTimeout:
		self.packProposal(evt.blockNum)
		if err := self.timer.CancelTxBlockTimeout(evt.blockNum); err != nil {
			log.Errorf("failed to cancel txBlockTimeout timer, blockNum %d, err: %s", evt.blockNum, err)
		}
	}
	return nil
}

func (self *Server) processHandshakeMsg(peerIdx uint32, msg *peerHandshakeMsg) error {
	if err := self.peerPool.peerHandshake(peerIdx, msg); err != nil {
		return fmt.Errorf("failed to update peer %d: %s", peerIdx, err)
	}
	self.stateMgr.StateEventC <- &StateEvent{
		Type: UpdatePeerConfig,
		peerState: &PeerState{
			peerIdx:           peerIdx,
			connected:         true,
			chainConfigView:   msg.ChainConfig.View,
			committedBlockNum: msg.CommittedBlockNumber,
		},
	}

	return nil
}

func (self *Server) processHeartbeatMsg(peerIdx uint32, msg *peerHeartbeatMsg) error {

	if err := self.peerPool.peerHeartbeat(peerIdx, msg); err != nil {
		return fmt.Errorf("failed to update peer %d: %s", peerIdx, err)
	}
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

	return nil
}

func (self *Server) endorseBlock(proposal *blockProposalMsg, forEmpty bool) error {
	// for each round, one node can only endorse one block, or empty block

	blkNum := proposal.GetBlockNum()

	// check if has endorsed
	if !forEmpty && self.blockPool.endorsedForBlock(blkNum) {
		return nil
	} else if forEmpty && self.blockPool.endorsedForEmptyBlock(blkNum) {
		return nil
	}

	blkHash, err := HashBlock(proposal.Block)
	if err != nil {
		return fmt.Errorf("failed to hash proposal block: %s", err)
	}

	// build endorsement msg
	endorseMsg, err := self.constructEndorseMsg(proposal, blkHash, forEmpty)
	if err != nil {
		return fmt.Errorf("failed to construct endorse msg: %s", err)
	}

	// set the block as self-endorsed-block
	if err := self.blockPool.setProposalEndorsed(proposal, forEmpty); err != nil {
		return fmt.Errorf("failed to set proposal as endorsed: %s", err)
	}

	self.processConsensusMsg(endorseMsg)
	// if node is endorser of current round
	if forEmpty || self.isEndorser(blkNum, self.Index) {
		self.msgPool.AddMsg(endorseMsg)
		log.Infof("endorser %d, endorsed block %d, from server %d",
			self.Index, blkNum, proposal.Block.getProposer())
		// broadcast my endorsement
		return self.broadcast(endorseMsg)
	}

	// start endorsing timer
	// TODO: endorsing may have reached consensus before received proposal, handle this
	if !forEmpty {
		if err := self.timer.StartEndorsingTimer(blkNum); err != nil {
			return fmt.Errorf("server %d failed to start endorser timer, blockNum %d, err: %s",
				self.Index, blkNum, err)
		}
	} else {
		if err := self.timer.StartEndorseEmptyBlockTimer(blkNum); err != nil {
			return fmt.Errorf("server %d failed to start empty endorse timer (%d): %s",
				self.Index, blkNum, err)
		}
	}

	return nil
}

func (self *Server) commitBlock(proposal *blockProposalMsg, forEmpty bool) error {
	// for each round, we can only commit one block

	blkNum := proposal.GetBlockNum()
	if self.blockPool.committedForBlock(blkNum) {
		return nil
	}

	blkHash, err := HashBlock(proposal.Block)
	if err != nil {
		return fmt.Errorf("failed to hash proposal proposal: %s", err)
	}

	// build commit msg
	commitMsg, err := self.constructCommitMsg(proposal, blkHash, forEmpty)
	if err != nil {
		return fmt.Errorf("failed to construct commit msg: %s", err)
	}

	// set the block as commited-block
	if err := self.blockPool.setProposalCommitted(proposal, forEmpty); err != nil {
		return fmt.Errorf("failed to set proposal as committed: %s", err)
	}

	self.processConsensusMsg(commitMsg)
	// if node is committer of current round
	if forEmpty || self.isCommitter(blkNum, self.Index) {
		self.msgPool.AddMsg(commitMsg)
		log.Infof("committer %d, set block %d committed, from server %d",
			self.Index, blkNum, proposal.Block.getProposer())
		// broadcast my commitment
		return self.broadcast(commitMsg)
	}

	// start commit timer
	// TODO: committing may have reached consensus before received endorsement, handle this
	if err := self.timer.StartCommitTimer(blkNum); err != nil {
		return fmt.Errorf("failed to start commit timer (%d): %s", blkNum, err)
	}

	return nil
}

//
// Note: sealProposal updates self.currentBlockNum, make sure not concurrency
// (only called by sealProposal action)
//
func (self *Server) sealProposal(proposal *blockProposalMsg, empty bool) error {
	// for each round, we can only seal one block
	if err := self.sealBlock(proposal.Block, empty); err != nil {
		return err
	}

	if self.hasBlockConsensused() {
		return self.makeFastForward()
	} else {
		return self.startNewRound()
	}

	return nil
}

func (self *Server) fastForwardBlock(block *Block) error {

	// TODO: update chainconfig when forwarding

	if isActive(self.getState()) {
		return fmt.Errorf("server %d: invalid fastforward, current state: %d", self.Index, self.getState())
	}
	if self.GetCurrentBlockNo() > block.getBlockNum() {
		return nil
	}
	if self.GetCurrentBlockNo() == block.getBlockNum() {
		return self.sealBlock(block, block.isEmpty())
	}
	return fmt.Errorf("server %d: fastforward blk %d failed, current blkNum: %d",
		self.Index, block.getBlockNum(), self.GetCurrentBlockNo())
}

func (self *Server) sealBlock(block *Block, empty bool) error {
	sealedBlkNum := block.getBlockNum()
	if sealedBlkNum < self.GetCurrentBlockNo() {
		// we already in future round
		return fmt.Errorf("late seal of %d, current blkNum: %d", sealedBlkNum, self.GetCurrentBlockNo())
	} else if sealedBlkNum > self.GetCurrentBlockNo() {
		// we have lost sync, restarting syncing
		self.restartSyncing()
		return fmt.Errorf("future seal of %d, current blknum: %d", sealedBlkNum, self.GetCurrentBlockNo())
	}

	if err := self.blockPool.setBlockSealed(block, empty); err != nil {
		return fmt.Errorf("failed to seal proposal: %s", err)
	}

	// TODO: also persistent the block endorsers and committer msgs

	// notify other modules that block sealed
	self.timer.onBlockSealed(sealedBlkNum)
	self.msgPool.onBlockSealed(sealedBlkNum)

	_, h := self.blockPool.getSealedBlock(sealedBlkNum)
	if h.CompareTo(common.Uint256{}) == 0 {
		return fmt.Errorf("failed to seal proposal: nil hash")
	}

	prevBlkHash := block.getPrevBlockHash()
	log.Infof("server %d, sealed block %d, proposer %d, prevhash: %s, hash: %s", self.Index,
		sealedBlkNum, block.getProposer(),
		hex.EncodeToString(prevBlkHash.ToArray()[:4]), hex.EncodeToString(h[:4]))

	// broadcast to other modules
	// TODO: block committed, update tx pool, notify block-listeners

	{
		self.metaLock.Lock()
		if sealedBlkNum >= self.currentBlockNum {
			self.currentBlockNum = sealedBlkNum + 1
		}
		defer self.metaLock.Unlock()
	}
	return nil
}

func (self *Server) msgSendLoop() {
	for {
		select {
		case evt := <-self.msgSendC:
			payload, err := SerializeVbftMsg(evt.Msg)
			if err != nil {
				log.Errorf("server %d failed to serialized msg (type: %d): %s", self.Index, evt.Msg.Type(), err)
				continue
			}
			if evt.ToPeer == math.MaxUint32 {
				// broadcast
				if err := self.broadcastToAll(payload); err != nil {
					log.Errorf("server %d xmit msg (type %d): %s",
						self.Index, evt.Msg.Type(), err)
				}
			} else {
				if err := self.sendToPeer(evt.ToPeer, payload); err != nil {
					log.Errorf("server %d xmit to peer %d failed: %s", self.Index, evt.ToPeer, err)
				}
			}

		case <-self.quitC:
			break
		}
	}
}

// FIXME
//    copied from dbft
func (self *Server) createBookkeepingTransaction(nonce uint64, fee uint64) *types.Transaction {
	log.Debug()
	//TODO: sysfee
	bookKeepingPayload := &payload.BookKeeping{
		Nonce: uint64(time.Now().UnixNano()),
	}
	return &types.Transaction{
		TxType:     types.BookKeeping,
		Payload:    bookKeepingPayload,
		Attributes: []*types.TxAttribute{},
	}
}

func (self *Server) makeProposal(blkNum uint64, forEmpty bool) error {
	var txs []*types.Transaction

	height := uint32(blkNum) - 1
	validHeight := height
	start, end := self.incrValidator.BlockRange()
	if height+1 == end {
		validHeight = start
	} else {
		self.incrValidator.Clean()
	}

	// FIXME: self.index as nonce??
	// FIXME: fix feesum calculation
	txBookkeeping := self.createBookkeepingTransaction(uint64(self.Index), 0)
	txs = append(txs, txBookkeeping)
	if !forEmpty {
		for _, e := range self.poolActor.GetTxnPool(true, uint32(validHeight)) {
			if err := self.incrValidator.Verify(e.Tx, uint32(validHeight)); err == nil {
				txs = append(txs, e.Tx)
			}
		}
	}

	proposal, err := self.constructProposalMsg(blkNum, txs)
	if err != nil {
		return fmt.Errorf("failed to construct proposal: %s", err)
	}

	log.Infof("server %d make proposal for block %d", self.Index, blkNum)

	// add proposal to self
	self.msgPool.AddMsg(proposal)
	self.processConsensusMsg(proposal)
	return self.broadcast(proposal)
}

func (self *Server) makeCommitment(proposal *blockProposalMsg, blkNum uint64) error {
	forEmpty := false
	if proposal == nil {
		// find empty proposal
		var err error
		proposal, err = self.blockPool.findConsensusEmptyProposal(blkNum)
		if err != nil {
			return fmt.Errorf("failed to get consensused empty proposal: %s", err)
		}
		forEmpty = true
	}

	if err := self.commitBlock(proposal, forEmpty); err != nil {
		return fmt.Errorf("failed to commit block proposal (%d): %s", blkNum, err)
	}
	return nil
}

func (self *Server) makeSealed(proposal *blockProposalMsg, forEmpty bool) error {
	blkNum := proposal.GetBlockNum()

	// check if agreed on prev-blockhash
	prevBlk, prevBlkHash := self.blockPool.getSealedBlock(blkNum - 1)
	if prevBlk == nil {
		// TODO: has no candidate proposals for prevBlock, should restart syncing
		return fmt.Errorf("failed to get prevBlock of current round (%d)", blkNum)
	}
	prevBlkHash2 := proposal.Block.getPrevBlockHash()
	if bytes.Compare(prevBlkHash.ToArray(), prevBlkHash2.ToArray()) != 0 {
		// TODO: in-consistency with prev-blockhash, resync-required
		self.restartSyncing()
		return fmt.Errorf("inconsistent prev-block detected when commiting (%d)", blkNum)
	}

	log.Infof("server %d ready to seal block %d, for proposer %d, empty: %t",
		self.Index, blkNum, proposal.Block.getProposer(), forEmpty)

	// seal the block
	self.bftActionC <- &BftAction{
		Type:     SealBlock,
		Proposal: proposal,
		forEmpty: forEmpty,
	}
	return nil
}

func (self *Server) makeFastForward() error {
	self.bftActionC <- &BftAction{
		Type: FastForward,
	}
	return nil
}

func (self *Server) reBroadcastCurrentRoundMsgs() error {
	self.bftActionC <- &BftAction{
		Type: ReBroadcast,
	}
	return nil
}

func (self *Server) fetchProposal(blkNum uint64, proposer uint32) error {
	msg, err := self.constructProposalFetchMsg(blkNum)
	if err != nil {
		return nil
	}
	self.msgSendC <- &SendMsgEvent{
		ToPeer: proposer,
		Msg:    msg,
	}
	return nil
}

func (self *Server) handleProposalTimeout(evt *TimerEvent) error {
	if self.blockPool.endorsedForBlock(evt.blockNum) {
		return nil
	}
	if !isReady(self.getState()) {
		return nil
	}
	proposals := self.blockPool.getBlockProposals(evt.blockNum)

	log.Infof("server %d proposal timeout, known proposals %d, timeout: %d", self.Index, len(proposals), evt.evtType)

	// if no proposal available, random backoff
	if len(proposals) == 0 {
		log.Infof("no proposal available for block %d, timeout: %d", evt.blockNum, evt.evtType)

		switch evt.evtType {
		case EventProposeBlockTimeout:
			self.timer.StartBackoffTimer(evt.blockNum)
			log.Infof("server %d started backoff timer for blk %d", self.Index, evt.blockNum)
			return nil
		case EventRandomBackoff:
			if err := self.makeProposal(evt.blockNum, true); err != nil {
				return fmt.Errorf("failed to propose empty block: %s", err)
			}
			if err := self.timer.Start2ndProposalTimer(evt.blockNum); err != nil {
				return fmt.Errorf("failed to start 2nd proposal timer: %s", err)
			}
			log.Infof("server %d proposed empty block for blk %d", self.Index, evt.blockNum)
			return nil
		case EventPropose2ndBlockTimeout:
			// 2nd proposal without any proposal, force resync
			self.restartSyncing()
			return nil
		}
	}
	// find highest rank proposal
	proposal := self.getHighestRankProposal(evt.blockNum, proposals)
	if proposal != nil {
		if self.isProposer(evt.blockNum, proposal.Block.getProposer()) {
			// proposal msg handler will do the endorsement
			return nil
		}

		self.bftActionC <- &BftAction{
			Type:     EndorseBlock,
			Proposal: proposal,
			forEmpty: false,
		}
	} else {
		log.Errorf("server: %d, blkNum: %d, failed get better proposal, first proposal: %d, %d",
			self.Index, evt.blockNum, proposals[0].Block.getProposer(), proposals[0].Block.getBlockNum())
	}
	return nil
}

func (self *Server) initHandshake(peerIdx uint32, peerPubKey *crypto.PubKey) error {
	msg, err := self.constructHandshakeMsg()
	if err != nil {
		return fmt.Errorf("build handshake msg: %s", err)
	}

	msgPayload, err := SerializeVbftMsg(msg)
	if err != nil {
		return fmt.Errorf("marshal handshake msg: %s", err)
	}

	errC := make(chan error)
	msgC := make(chan *peerHandshakeMsg)
	go func() {
		for {
			// FIXME: peer receive with timeout
			msgData, err := self.receiveFromPeer(peerIdx)
			if err != nil {
				log.Errorf("read initHandshake msg from peer: %s", err)
				errC <- fmt.Errorf("read initHandshake msg from peer: %s", err)
			}
			msg, err := DeserializeVbftMsg(msgData)
			if err != nil {
				log.Errorf("unmarshal msg failed: %s", err)
				errC <- fmt.Errorf("unmarshal msg failed: %s", err)
			}
			if err := msg.Verify(peerPubKey); err != nil {
				log.Errorf("msg verify failed in initHandshake: %s", err)
				errC <- fmt.Errorf("msg verify failed in initHandshake: %s", err)
			}
			if msg.Type() == PeerHandshakeMessage {
				if shakeMsg, ok := msg.(*peerHandshakeMsg); ok {
					self.sendToPeer(peerIdx, msgPayload)
					msgC <- shakeMsg
				}
				break
			}
		}
	}()
	if err := self.sendToPeer(peerIdx, msgPayload); err != nil {
		return fmt.Errorf("send initHandshake msg: %s", err)
	}

	timeout := time.NewTimer(peerHandshakeTimeout)
	defer timeout.Stop()

	select {
	case msg := <-msgC:
		if err := self.processHandshakeMsg(peerIdx, msg); err != nil {
			return fmt.Errorf("process initHandshake msg failed: %s", err)
		}
	case err := <-errC:
		return fmt.Errorf("peer initHandshake failed: %s", err)
	case <-timeout.C:
		return fmt.Errorf("peer initHandshake timeout")
	}

	return nil
}

// TODO: refactor this
func (self *Server) catchConsensus(blkNum uint64) error {
	if !self.isEndorser(blkNum, self.Index) && !self.isCommitter(blkNum, self.Index) {
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

	F := int(self.config.F)
	eMsgs := self.msgPool.GetEndorsementsMsgs(blkNum)
	var proposal *blockProposalMsg
	endorseDone := false
	endorseEmpty := false
	if len(eMsgs) > F {
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
		if maxCnt > F {
			endorseDone = true
		}
		if emptyCnt > F {
			endorseDone = true
			endorseEmpty = true
		}
	}
	if proposal != nil && self.isProposer(blkNum, proposal.Block.getProposer()) {
		self.processConsensusMsg(proposal)
	}

	if self.isEndorser(blkNum, self.Index) && !endorseDone && proposal != nil {
		return self.endorseBlock(proposal, endorseEmpty)
	}

	if !endorseDone {
		return fmt.Errorf("server %d catch consensus with endorse failed", self.Index)
	}

	if !self.isCommitter(blkNum, self.Index) {
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

func (self *Server) hasBlockConsensused() bool {
	blkNum := self.GetCurrentBlockNo()

	F := int(self.config.F)
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
		if proposers[c.BlockProposer] > F {
			return true
		}
	}

	return emptyCnt > F
}
