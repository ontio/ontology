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
	"math"
	"reflect"
	"sync"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
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
	Block      *types.Block
	Info       *vconfig.VbftBlockInfo
	WriteSet   *overlaydb.MemDB
	MerkleRoot common.Uint256
}

type p2pMsgPayload struct {
	fromPeer uint32
	Data     ConsensusMsg
}

type Server struct {
	Index         uint32
	account       *account.Account
	poolActor     *actorTypes.TxPoolActor
	p2p           p2p.P2P
	incrValidator *increment.IncrementValidator
	pid           *actor.PID

	lock    sync.RWMutex
	vbftCtx *VbftContext

	msgPool        *MsgPool // consensus msg pool
	chainStore     *ChainStore
	peerPool       *PeerPool // consensus peers
	stateMgr       *StateMgr
	timer          *EventTimer
	inMakeProgress bool // local value owned by bft loop routine

	totalPeerWorkers uint32
	peerWorkerChan   []chan *p2pMsgPayload
	msgC             chan ConsensusMsg
	blockSynced      chan *VbftBlock
	msgSendC         chan *SendMsgEvent
	sub              *events.ActorSubscriber
	quitC            chan struct{}
	quitWg           sync.WaitGroup
}

func NewVbftServer(account *account.Account, txpool *actor.PID, p2p p2p.P2P) (*Server, error) {
	server := &Server{
		account:          account,
		poolActor:        &actorTypes.TxPoolActor{Pool: txpool},
		p2p:              p2p,
		incrValidator:    increment.NewIncrementValidator(20),
		totalPeerWorkers: 4,
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
	log.Infof("persist block complete: height=%d, hash=%s, state root:%s, numtx=%d", block.Header.Height, block.Hash().ToHexString(),
		exec.MerkleRoot.ToHexString(), len(block.Transactions))
	blkInfo, err := vconfig.VbftBlock(block.Header)
	if err != nil {
		log.Errorf("load vbft block info failed:%s", err)
		return
	}
	if self.updateVbftContext(block, blkInfo, exec) {
		// p2p synced before seal block: 1. not consensus node; 2. consensus node in syncing state
		self.lock.Lock()
		self.chainStore.ReloadFromLedger()
		self.lock.Unlock()
	}
}

func (self *Server) CheckAndSubmitBlock(blkNum uint32, stateRoot common.Uint256) {
	cMsgs := self.msgPool.GetBlockSubmitMsgs(blkNum)
	var stateRootCnt uint32 = 1 // include self
	for _, msg := range cMsgs {
		c := msg.(*blockSubmitMsg)
		if c.BlockStateRoot == stateRoot && c.Submitter != self.Index {
			stateRootCnt++
		}
	}

	cfg := self.GetVbftContext().Config
	if stateRootCnt >= cfg.Quorum() {
		log.Infof("receive enough submit msg for block %d, start submit block", blkNum)
		if err := self.SubmitBlock(blkNum); err != nil {
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
	payload.BookkeeperIndex = uint16(peerIdx) // TODO: when all node upgraded, remove this
	msg, err := DeserializeVbftMsg(payload)
	if err != nil {
		log.Errorf("server %d failed to deserialize vbft msg (len %d): %s", self.Index, len(payload.Data), err)
		return
	}
	if !self.peerPool.IsPeerConnected(peerIdx) {
		self.peerPool.OnPeerConnected(peerIdx)
	}
	p2pid, present := self.peerPool.GetP2pId(peerIdx)
	if !present || p2pid != payload.PeerId {
		self.peerPool.AddP2pId(peerIdx, payload.PeerId)
	}
	switch pMsg := msg.(type) {
	case *blockProposalMsg:
		peerIdx = pMsg.Block.getProposer()
	case *blockProposalMsgV2:
		peerIdx = pMsg.Proposer
	}

	if info := self.GetVbftContext().PeerKeys[peerIdx]; info != nil {
		self.peerWorkerChan[info.Id] <- &p2pMsgPayload{
			fromPeer: peerIdx,
			Data:     msg,
		}
	} else {
		log.Errorf("consensus msg without receiver: %d node: %s", peerIdx, peerID)
		return
	}
}

func (self *Server) LoadChainConfig(store *ChainStore, block *VbftBlock, stateRoot common.Uint256) error {
	blkNum := store.ChainedBlockNum
	cfg, configBlk, err := ledger.DefLedger.LoadCfgFromBlock(blkNum)
	if err != nil {
		panic(err)
	}
	if cfg.View == 0 || cfg.MaxBlockChangeView == 0 {
		panic("invalid view or maxblockchangeview ")
	}

	// update timer params
	self.updateTimerParams(cfg)

	log.Infof("current committed block no: %d", blkNum)
	proposers, endorsers, committers := buildPeerRoles(blkNum+1, block.Info.Proposer, block.Info.VrfValue, cfg)
	log.Infof("server %d, blkNum: %d, state: %s, participants: %v, %v, %v", self.Index, blkNum+1,
		self.getState(), proposers, endorsers, committers)

	peermap := make(map[uint32]*KeyAndTaskId)
	for i, p := range cfg.Peers {
		// check if peer pubkey support VRF
		publickey, err := vconfig.Pubkey(p.ID)
		if err != nil || !vrf.ValidatePublicKey(publickey) {
			panic(fmt.Errorf("peer pubkey is ensured to be valid for VRF:%s", p.ID))
		}
		peermap[p.Index] = &KeyAndTaskId{Key: publickey, Id: uint32(i) % self.totalPeerWorkers}
	}
	self.vbftCtx = &VbftContext{
		BlockNum:  blkNum + 1,
		Config:    cfg,
		ConfigNum: configBlk,
		PeerKeys:  peermap,
		PrevBlockInfo: &BlockAndExecteInfo{
			Block:      block.Block,
			Info:       block.Info,
			WriteSet:   nil,
			MerkleRoot: stateRoot,
		},
		BftStatus:  NewBftStatus(),
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
		peermap := make(map[uint32]*KeyAndTaskId)
		for i, p := range info.NewChainConfig.Peers {
			// check if peer pubkey support VRF
			publickey, err := vconfig.Pubkey(p.ID)
			if err != nil || !vrf.ValidatePublicKey(publickey) {
				panic(fmt.Errorf("peer pubkey is ensured to be valid for VRF:%s", p.ID))
			}
			peermap[p.Index] = &KeyAndTaskId{Key: publickey, Id: uint32(i) % self.totalPeerWorkers}
		}
		vbftCtx.PeerKeys = peermap
	}

	vbftCtx.Proposers, vbftCtx.Endorsers, vbftCtx.Committers = buildPeerRoles(blkNum+1, info.Proposer, info.VrfValue, vbftCtx.Config)

	vbftCtx.BlockNum = blkNum + 1
	vbftCtx.BftStatus = NewBftStatus()
	vbftCtx.PrevBlockInfo = &BlockAndExecteInfo{
		Block:      block,
		Info:       info,
		WriteSet:   result.WriteSet,
		MerkleRoot: result.MerkleRoot,
	}

	self.lock.Lock()
	if self.vbftCtx.BlockNum+1 == vbftCtx.BlockNum {
		self.vbftCtx = &vbftCtx
		self.incrValidator.AddBlock(block)
		start, end := self.incrValidator.BlockRange()
		log.Infof("update vbft context, blkNum:%d, incr validator range: [%d, %d)", blkNum+1, start, end)
		updated = true
	}
	self.lock.Unlock()
	if updated {
		log.Infof("server %d, blkNum: %d, state: %s, participants: %v, %v, %v", self.Index, blkNum+1,
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

	_, removed := self.peerPool.ResetNewConsuensusPeers(peermap)
	for _, index := range removed {
		if index == self.Index {
			self.Index = math.MaxUint32
			log.Infof("updateTimerAndPeerPool remove index :%d", index)
		} else {
			self.peerPool.OnPeerDisconnected(index)
			self.stateMgr.StateEventC <- &StateEvent{
				Type: UpdatePeerState,
				peerState: &PeerState{
					peerIdx:   index,
					connected: false,
				},
			}
		}
	}
}

func (self *Server) initialize() error {
	selfNodeId := vconfig.PubkeyID(self.account.PublicKey)
	log.Infof("server: %s starting", selfNodeId)

	store, block, root, err := OpenBlockStore(ledger.DefLedger)
	if err != nil {
		log.Errorf("failed to open block store: %s", err)
		return fmt.Errorf("failed to open block store: %s", err)
	}
	log.Info("block store opened")

	var msgHistoryDuration uint32 = 64
	self.msgPool = newMsgPool(self, msgHistoryDuration)
	self.timer = NewEventTimer(self)
	self.stateMgr = newStateMgr(self)

	self.msgC = make(chan ConsensusMsg, CAP_MESSAGE_CHANNEL)
	self.msgSendC = make(chan *SendMsgEvent, CAP_MSG_SEND_CHANNEL)
	self.peerWorkerChan = make([]chan *p2pMsgPayload, self.totalPeerWorkers)
	for i := uint32(0); i < self.totalPeerWorkers; i += 1 {
		self.peerWorkerChan[i] = make(chan *p2pMsgPayload, 1024)
	}

	self.quitC = make(chan struct{})
	if err := self.LoadChainConfig(store, block, root); err != nil {
		log.Errorf("failed to load config: %s", err)
		return fmt.Errorf("failed to load config: %s", err)
	}
	log.Infof("chain config loaded from local, current blockNum: %d", self.GetCurrentBlockNo())

	// add all consensus peers to peer_pool
	peermap := make(map[string]uint32)
	for index, p := range self.GetVbftContext().PeerKeys {
		peermap[vconfig.PubkeyID(p.Key)] = index
	}
	self.peerPool = NewPeerPool(peermap)
	self.chainStore = store

	//index equal math.MaxUint32  is noconsensus node
	id := vconfig.PubkeyID(self.account.PublicKey)
	index, present := self.peerPool.GetPeerIndex(id)
	if present {
		self.Index = index
	} else {
		self.Index = math.MaxUint32
	}
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
	for i := uint32(0); i < self.totalPeerWorkers; i += 1 {
		go self.peerWorkerLoop(self.peerWorkerChan[i])
	}

	return nil
}

func (self *Server) stop() {
	self.incrValidator.Clean()
	self.sub.Unsubscribe(message.TOPIC_SAVE_BLOCK_COMPLETE)
	// stop syncer, statemgr, msgSendLoop, timer, actionLoop, msgProcessingLoop
	close(self.quitC)
	self.quitWg.Wait()

	self.timer.stop()
	self.msgPool.clean()
	self.peerPool.clean()
}

// go routine per net connection
func (self *Server) peerWorkerLoop(msgChan chan *p2pMsgPayload) {
	for {
		select {
		case payload := <-msgChan:
			if payload != nil {
				self.onConsensusMsg(payload.fromPeer, payload.Data)
			}
		case <-self.quitC:
			return
		}
	}
}

func (self *Server) getState() ServerState {
	return self.stateMgr.getState()
}

func (self *Server) startNewRound() {
	vbftCtx := self.GetVbftContext()
	blkNum := vbftCtx.BlockNum
	self.timer.StartEventTimer(EventTxPool, blkNum)
	self.timer.StartEventTimer(EventTxBlockTimeout, blkNum)

	msgs := self.msgPool.DropBftMsgs(blkNum)
	go func() {
		for _, msg := range msgs {
			self.processBftMsgFromPeer(vbftCtx, msg)
		}
	}()
	return
}

func (self *Server) startNewProposal(vbftCtx *VbftContext) {
	blkNum := vbftCtx.BlockNum
	// make proposal
	if self.isProposer(self.Index) {
		log.Infof("server %d, proposer for block %d", self.Index, blkNum)
		proposal := vbftCtx.BftStatus.GetBlockProposal(self.Index)
		if proposal == nil {
			if err := self.makeProposal(blkNum, false); err != nil {
				log.Errorf("server %d failed to making proposal (%d): %s", self.Index, blkNum, err)
			}
		}
	} else if vbftCtx.Is2ndProposer(self.Index) {
		log.Infof("server %d, 2nd proposer for block %d", self.Index, blkNum)
		self.timer.StartEventTimer(EventProposalBackoff, blkNum)
	}

	self.timer.StartEventTimer(EventProposeBlockTimeout, blkNum)
}

func (self *Server) processBftMsgFromPeer(vbftCtx *VbftContext, msg ConsensusMsg) {
	if !self.getState().IsReady() {
		return
	}
	if err := msg.(BftConsensusMsg).Verify(vbftCtx); err != nil {
		log.Errorf("server %d failed to verify msg, type %s, err: %s", self.Index, msg.Type(), err)
		return
	}
	switch pMsg := msg.(type) {
	case *blockProposalMsgV2:
		proposal, err := self.verifyProposalMsgV2(vbftCtx, pMsg, true)
		if err != nil {
			log.Errorf("verify proposal error: %v", err)
			return
		}
		msg = proposal
	case *blockProposalMsg:
		err := self.verifyProposalMsg(vbftCtx, pMsg, true)
		if err != nil {
			log.Errorf("verify proposal error: %v", err)
			return
		}
	case *blockEndorseMsg, *blockCommitMsg:
	default:
		panic(fmt.Errorf("unknown bft msg:%v", msg.Type()))
	}
	if err := self.msgPool.AddMsg(msg); err != nil {
		log.Errorf("failed to add msg (%d) to pool", vbftCtx.BlockNum)
		return
	}
	self.msgC <- msg
}

func (self *Server) onConsensusMsg(peerIdx uint32, msg ConsensusMsg) {
	if _, ok := msg.(BftConsensusMsg); ok {
		log.Infof("server %d received bft msg, blk %d, type: %s from %d",
			self.Index, msg.GetBlockNum(), msg.Type(), peerIdx)
	}

	if self.msgPool.HasMsg(msg) {
		// dup msg checking
		log.Debugf("dup msg with msg type %s from %d", msg.Type(), peerIdx)
		return
	}

	switch pMsg := msg.(type) {
	case *blockProposalMsg, *blockProposalMsgV2, *blockEndorseMsg, *blockCommitMsg:
		vbftCtx := self.GetVbftContext()
		msgBlkNum := msg.GetBlockNum()
		if msgBlkNum != vbftCtx.BlockNum {
			if msgBlkNum > vbftCtx.BlockNum {
				self.msgPool.AddMsg(msg)
			}
			return
		}
		self.processBftMsgFromPeer(vbftCtx, msg)
	case *peerHeartbeatMsg:
		self.processHeartbeatMsg(peerIdx, pMsg)
		vbftCtx := self.GetVbftContext()
		if pMsg.CommittedBlockNumber == vbftCtx.BlockNum {
			self.msgC <- msg
		}
	case *proposalFetchMsg:
		pmsg := self.msgPool.GetProposalMsg(pMsg.BlockNum, pMsg.ProposerID)
		if pmsg != nil {
			switch p := pmsg.(type) {
			case *blockProposalMsg:
				log.Infof("server %d, handle proposal fetch %d from %d", self.Index, pMsg.BlockNum, peerIdx)
				self.msgSendC <- &SendMsgEvent{
					ToPeer: peerIdx,
					Msg:    p,
				}
			case *blockProposalMsgV2:
				// TODO when upgraded
			}
		}
	case *blockSubmitMsg:
		msgBlkNum := pMsg.GetBlockNum()
		vbftCtx := self.GetVbftContext()
		if vbftCtx.BlockNum > msgBlkNum+1 {
			return
		}
		if err := pMsg.Verify(vbftCtx); err != nil {
			log.Errorf("server %d failed to verify msg, type %s, err: %s", self.Index, msg.Type(), err)
			return
		}
		if err := self.msgPool.AddMsg(msg); err != nil {
			return
		}
		if vbftCtx.BlockNum == msgBlkNum+1 {
			self.CheckAndSubmitBlock(msgBlkNum, vbftCtx.PrevBlockInfo.MerkleRoot)
		}
	}
}

func (self *Server) verifyProposalMsgV2(vbftCtx *VbftContext, msg *blockProposalMsgV2, verifyTx bool) (*blockProposalMsg, error) {
	blkNum := vbftCtx.BlockNum
	blockTime := msg.BlockTime
	if blockTime <= vbftCtx.PrevBlockInfo.Block.Header.Timestamp || blockTime > uint32(time.Now().Add(time.Minute*10).Unix()) {
		return nil, fmt.Errorf("proposal block timestamp failed, blocknum:%d, timestamp:%d", blkNum, blockTime)
	}

	proposer := msg.Proposer
	proposerPk := vbftCtx.GetPeerPubKey(proposer)
	if proposerPk == nil {
		return nil, fmt.Errorf("server %d failed to get proposer %d pk of block %d",
			self.Index, proposer, blkNum)
	}
	vrfValue, vrfProof := msg.VrfValue, msg.VrfProof
	if err := verifyVrf(proposerPk, blkNum, vbftCtx.PrevBlockInfo.Info.VrfValue, msg.VrfValue, msg.VrfProof); err != nil {
		return nil, fmt.Errorf("server %d failed to verify vrf of block %d proposal from %d",
			self.Index, blkNum, proposer)
	}
	txs := msg.Transactions
	cfg, err := self.GetNewBlockConfig(vbftCtx)
	if err != nil {
		return nil, fmt.Errorf("get new block config failed:%s", err)
	}
	if vbftCtx.NeedUpdateChainConfigTx() {
		if len(txs) != 1 || CreateGovernaceTransaction(vbftCtx.BlockNum).Hash() != txs[0].Hash() {
			return nil, fmt.Errorf("update chain config block must has 1 commit dpos transaction, blk: %d", blkNum)
		}
		txs = txs[1:]
	}

	proposal := BuildProposalMsg(vbftCtx, txs, cfg, uint64(blkNum),
		blockTime, proposer, vrfValue, vrfProof)
	if proposal.Block.Block.Hash() != msg.BlockHash || proposal.Block.EmptyBlock.Hash() != msg.EmptyBlockHash {
		return nil, fmt.Errorf("generated proposal block hash mismatch, blk: %d", blkNum)
	}
	proposal.Block.Block.Header.Bookkeepers = []keypair.PublicKey{proposerPk}
	proposal.Block.EmptyBlock.Header.Bookkeepers = []keypair.PublicKey{proposerPk}
	proposal.Block.Block.Header.SigData = [][]byte{msg.Sig}
	proposal.Block.EmptyBlock.Header.SigData = [][]byte{msg.EmptySig}

	if verifyTx && len(txs) > 0 {
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
		if err := self.poolActor.VerifyBlock(txs, validHeight); err != nil && err != actor.ErrTimeout {
			return nil, fmt.Errorf("server %d verify proposal blk from %d failed, blk %d, txs %d, err: %s",
				self.Index, proposer, blkNum, len(txs), err)
		} else if err == actor.ErrTimeout {
			return nil, fmt.Errorf("server %d verify proposal blk from %d timedout, blk %d, txs %d, err: %s",
				self.Index, proposer, blkNum, len(txs), err)
		}
		nonceCtx := make(map[common.Address]uint64)
		for _, tx := range txs {
			if err := self.incrValidator.Verify(tx, validHeight, nonceCtx); err != nil {
				return nil, fmt.Errorf("server %d verify proposal tx from %d failed, blk %d, txs %d, err: %s",
					self.Index, proposer, blkNum, len(txs), err)
			}
		}
	}
	return proposal, nil
}

func (self *Server) verifyProposalMsg(vbftCtx *VbftContext, msg *blockProposalMsg, verifyTx bool) error {
	blkNum := vbftCtx.BlockNum
	blockTime := msg.Block.Block.Header.Timestamp
	if blockTime <= vbftCtx.PrevBlockInfo.Block.Header.Timestamp || blockTime > uint32(time.Now().Add(time.Minute*10).Unix()) {
		return fmt.Errorf("proposal block timestamp failed, blocknum:%d, timestamp:%d", blkNum, blockTime)
	}

	proposer := msg.Block.getProposer()
	proposerPk := vbftCtx.GetPeerPubKey(proposer)
	if proposerPk == nil {
		return fmt.Errorf("server %d failed to get proposer %d pk of block %d",
			self.Index, proposer, blkNum)
	}
	vrfValue, vrfProof := msg.Block.getVrfValue(), msg.Block.getVrfProof()
	if err := verifyVrf(proposerPk, blkNum, vbftCtx.PrevBlockInfo.Info.VrfValue, vrfValue, vrfProof); err != nil {
		return fmt.Errorf("server %d failed to verify vrf of block %d proposal from %d",
			self.Index, blkNum, proposer)
	}
	txs := msg.Block.Block.Transactions
	cfg, err := self.GetNewBlockConfig(vbftCtx)
	if err != nil {
		return fmt.Errorf("get new block config failed:%s", err)
	}
	if vbftCtx.NeedUpdateChainConfigTx() {
		if len(txs) != 1 || CreateGovernaceTransaction(vbftCtx.BlockNum).Hash() != txs[0].Hash() {
			return fmt.Errorf("update chain config block must has 1 commit dpos transaction, blk: %d", blkNum)
		}
		txs = txs[1:]
	}

	proposal := BuildProposalMsg(vbftCtx, txs, cfg, msg.Block.Block.Header.ConsensusData,
		blockTime, proposer, vrfValue, vrfProof)
	if proposal.Block.Block.Hash() != msg.Block.Block.Hash() || proposal.Block.EmptyBlock.Hash() != msg.Block.EmptyBlock.Hash() {
		return fmt.Errorf("generated proposal block hash mismatch, blk: %d", blkNum)
	}

	if verifyTx && len(txs) > 0 {
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
		if err := self.poolActor.VerifyBlock(txs, validHeight); err != nil && err != actor.ErrTimeout {
			return fmt.Errorf("server %d verify proposal blk from %d failed, blk %d, txs %d, err: %s",
				self.Index, proposer, blkNum, len(txs), err)
		} else if err == actor.ErrTimeout {
			return fmt.Errorf("server %d verify proposal blk from %d timedout, blk %d, txs %d, err: %s",
				self.Index, proposer, blkNum, len(txs), err)
		}
		nonceCtx := make(map[common.Address]uint64)
		for _, tx := range txs {
			if err := self.incrValidator.Verify(tx, validHeight, nonceCtx); err != nil {
				return fmt.Errorf("server %d verify proposal tx from %d failed, blk %d, txs %d, err: %s",
					self.Index, proposer, blkNum, len(txs), err)
			}
		}
	}
	return nil
}

func (self *Server) makeProgress(vbftCtx *VbftContext) {
	// makeProgress can be called recursively, use this to only allow the outer one call this
	if self.inMakeProgress {
		return
	}
	self.inMakeProgress = true // makeProgress can be called recursively, use this to only allow the outer one call this
	defer func() {
		self.inMakeProgress = false
	}()
	log.Infof("bft before progress status: %s, block: %d", vbftCtx.BftStatus.String(), vbftCtx.BlockNum)
	defer func() {
		// when sealed, the vbft context will be updated, so get the new one
		vbftCtx := self.GetVbftContext()
		log.Infof("bft after progress status: %s, block: %d", vbftCtx.BftStatus.String(), vbftCtx.BlockNum)
	}()
	blkNum := vbftCtx.BlockNum
	bftStatus := vbftCtx.BftStatus
	proposal := bftStatus.GetBlockProposal(self.GetActiveProposer())
	if proposal != nil {
		// stop proposal timer
		self.timer.CancelEventTimer(EventProposeBlockTimeout, blkNum)
		if vbftCtx.IsEndorser(self.Index) {
			if err := self.endorseBlock(vbftCtx, proposal, false); err != nil {
				log.Errorf("failed to endorse block proposal (%d): %s", blkNum, err)
			}
		}
	}

	// TODO: should only count endorsements from endorsers
	if bftStatus.SelfCommitMsg == nil {
		if proposer, forEmpty, done := bftStatus.EndorseDone(vbftCtx.Config.C); done {
			// stop endorse timer
			self.timer.CancelEventTimer(EventEndorseBlockTimeout, blkNum)
			// stop empty endorse timer
			self.timer.CancelEventTimer(EventEndorseEmptyBlockTimeout, blkNum)
			proposal := bftStatus.GetBlockProposal(proposer)
			if proposal == nil {
				self.fetchProposal(blkNum, proposer)
				log.Infof("server %d endorse %d done, waiting proposal from %d", self.Index, blkNum, proposer)
			} else if vbftCtx.IsCommitter(self.Index) {
				// make endorsement
				if err := self.commitBlock(vbftCtx, proposal, forEmpty); err != nil {
					log.Errorf("failed to endorse for block %d: %s", blkNum, err)
					return
				}
			}
		}
	}
	if bftStatus.EndorseFailed(vbftCtx.Config.C) {
		// endorse failed, start empty endorsing
		self.processTimerEvent(&TimerEvent{evtType: EventEndorseBlockTimeout, blockNum: blkNum})
	}

	chainCfg := vbftCtx.Config
	if proposer, forEmpty, done := bftStatus.CommitDone(vbftCtx); done {
		proposal := bftStatus.GetBlockProposal(proposer)
		if proposal == nil {
			log.Infof("server %d commit %d done, waiting proposal", self.Index, blkNum)
			self.timer.StartEventTimer(EventCommitBlockTimeout, blkNum)
			return
		}

		if vbftCtx.IsCommitter(self.Index) {
			// make sure committer broadcasting his commit msg
			if err := self.commitBlock(vbftCtx, proposal, forEmpty); err != nil {
				log.Errorf("server %d consensused %d, committer broadcast commit msg: %s", self.Index, blkNum, err)
			}
		}
		if !bftStatus.checkBlockSign(vbftCtx, proposal.Block, forEmpty, chainCfg.Quorum()) {
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
	bftStatus := vbftCtx.BftStatus
	msgBlkNum := msg.GetBlockNum()
	if msgBlkNum != vbftCtx.BlockNum {
		return
	}
	log.Debugf("server %d start process bft msg, block %d, type %s", self.Index, msg.GetBlockNum(), msg.Type())
	switch pMsg := msg.(type) {
	case *peerHeartbeatMsg:
		log.Infof("server %d process heatbeat signature for block %d", self.Index, msgBlkNum)
		proposal := bftStatus.GetBlockProposal(pMsg.CommittedBlockProposer)
		if proposal == nil {
			// if node is syncing, proposal will not be in bft status
			prop := self.msgPool.GetProposalMsg(msgBlkNum, pMsg.CommittedBlockProposer)
			if prop != nil {
				switch pMsg := prop.(type) {
				case *blockProposalMsg:
					if self.verifyProposalMsg(vbftCtx, pMsg, false) == nil {
						proposal = pMsg
					}
				case *blockProposalMsgV2:
					if p, err := self.verifyProposalMsgV2(vbftCtx, pMsg, false); err == nil {
						proposal = p
					}
				}
			}
		}
		if proposal != nil && proposal.Block.Block.Hash() == pMsg.CommittedBlockHash {
			block := proposal.Block.Block
			var pubkeys []keypair.PublicKey
			for _, k := range pMsg.Endorsers {
				pub, err := keypair.DeserializePublicKey(k)
				if err != nil {
					return
				}
				pubkeys = append(pubkeys, pub)
			}
			pubInfos := make(map[string]bool)
			for _, p := range vbftCtx.PeerKeys {
				pubInfos[common.PubKeyToHex(p.Key)] = true
			}
			block.Header.Bookkeepers = pubkeys
			block.Header.SigData = pMsg.EndorsersSig
			err := block.Header.VerifyMultiSignature(pubInfos, false, vbftCtx.Config.Quorum())
			if err != nil {
				return
			}
			if err = self.sealBlock(proposal.Block, false, false); err != nil {
				log.Errorf("server %d failed to seal block (%d): %s", self.Index, block.Header.Height, err)
			}
		}
		return
	case *blockProposalMsg:
		log.Infof("server %d received proposal from %d, block %d, txnum %d",
			self.Index, pMsg.Block.getProposer(), msgBlkNum, len(pMsg.Block.Block.Transactions))
		if err := bftStatus.AddBlockProposal(pMsg); err != nil {
			// TODO: faulty proposer detected
			log.Errorf("failed to add block proposal (%d): %s", msgBlkNum, err)
			return
		}

		if self.Index != pMsg.Block.getProposer() && self.isProposer(self.Index) {
			p := bftStatus.GetBlockProposal(self.Index)
			if p != nil {
				self.broadcast(p)
			}
		}
	case *blockEndorseMsg:
		if err := bftStatus.AddBlockEndorseMsg(pMsg); err != nil {
			log.Errorf("failed to add endorse msg (%d): %s", msgBlkNum, err)
			return
		}
		log.Infof("server %d received endorse from %d, for proposer %d, block %d, empty: %t",
			self.Index, pMsg.Endorser, pMsg.EndorsedProposer, msgBlkNum, pMsg.EndorseForEmpty)
	case *blockCommitMsg:
		if err := bftStatus.AddBlockCommitMsg(pMsg); err != nil {
			log.Errorf("failed to add commit msg (%d): %s", msgBlkNum, err)
			return
		}

		log.Infof("server %d received commit from %d, for proposer %d, block %d, empty: %t",
			self.Index, pMsg.Committer, pMsg.BlockProposer, msgBlkNum, pMsg.CommitForEmpty)
	}
	self.makeProgress(vbftCtx)
}

func (self *Server) RebroadcastMsgs(vbftCtx *VbftContext, blkNum uint32) {
	bftStatus := vbftCtx.BftStatus
	proposals := bftStatus.Proposals
	for _, p := range proposals {
		if p.Block.getProposer() == self.Index {
			log.Infof("server %d rebroadcast proposal, blk %d", self.Index, blkNum)
			self.broadcast(p)
			break
		}
	}
	if vbftCtx.IsEndorser(self.Index) {
		rebroadcasted := false
		endorseMsg := bftStatus.GetSelfEndorseMsg()
		if endorseMsg != nil {
			self.broadcast(endorseMsg)
			rebroadcasted = true
		}
		if !rebroadcasted {
			proposal := vbftCtx.GetHighestRankProposal(proposals)
			if proposal != nil {
				if err := self.endorseBlock(vbftCtx, proposal, false); err != nil {
					log.Errorf("server %d rebroadcasting failed to endorse (%d): %s",
						self.Index, blkNum, err)
				}
			} else {
				log.Errorf("server %d rebroadcasting failed to endorse(%d), no proposal found(%d)",
					self.Index, blkNum, len(proposals))
			}
		}
	} else if endorseMsg := bftStatus.GetSelfEndorseMsg(); endorseMsg != nil {
		self.broadcast(endorseMsg)
	}

	if vbftCtx.IsCommitter(self.Index) {
		commitMsg := bftStatus.SelfCommitMsg
		if commitMsg != nil {
			log.Infof("server %d rebroadcast commit, blk %d for %d, %t",
				self.Index, vbftCtx.BlockNum, commitMsg.BlockProposer, commitMsg.CommitForEmpty)
			self.broadcast(commitMsg)
		} else {
			if proposer, forEmpty, done := bftStatus.EndorseDone(vbftCtx.Config.C); done {
				proposal := bftStatus.GetBlockProposal(proposer)
				// consensus ok, make endorsement
				if proposal == nil {
					self.fetchProposal(blkNum, proposer)
					// restart endorsing timer
					self.timer.StartEventTimer(EventEndorseBlockTimeout, blkNum)
					log.Errorf("server %d endorse %d done, but no proposal", self.Index, blkNum)
				} else if err := self.commitBlock(vbftCtx, proposal, forEmpty); err != nil {
					log.Errorf("server %d failed to commit block %d on rebroadcasting: %s",
						self.Index, blkNum, err)
				}
			} else if bftStatus.EndorseFailed(vbftCtx.Config.C) {
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
	log.Infof("start process timer event: %s, block: %d", evt.evtType, evt.blockNum)
	bftStatus := vbftCtx.BftStatus
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

		if bftStatus.HasEndorsedForBlock() {
			return nil
		}
		if !self.getState().IsReady() {
			return nil
		}
		proposals := bftStatus.Proposals
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
		if bftStatus.SelfCommitMsg != nil {
			return nil
		}
		if !self.getState().IsReady() {
			return nil
		}
		if proposer, forEmpty, done := bftStatus.EndorseDone(vbftCtx.Config.C); done {
			proposal := bftStatus.GetBlockProposal(proposer)
			// consensus ok, make endorsement
			if proposal == nil {
				self.fetchProposal(evt.blockNum, proposer)
				// restart endorsing timer
				self.timer.StartEventTimer(EventEndorseBlockTimeout, evt.blockNum)
				return fmt.Errorf("endorse %d done, but no proposal available", evt.blockNum)
			}
			if err := self.commitBlock(vbftCtx, proposal, forEmpty); err != nil {
				return fmt.Errorf("failed to endorse for block %d on endorse timeout: %s", evt.blockNum, err)
			}
			return nil
		}
		if !self.getState().IsActive() {
			// not active yet, waiting active peers making decision
			return nil
		}
		if bftStatus.HasEndorsedForEmptyBlock() {
			return nil
		}
		proposals := bftStatus.Proposals
		if len(proposals) == 0 {
			log.Errorf("endorsing timeout, without any proposal. restarting syncing")
			self.restartSyncing()
			return nil
		}
		proposal := vbftCtx.GetHighestRankProposal(proposals)
		if proposal != nil {
			if err := self.endorseBlock(vbftCtx, proposal, true); err != nil {
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
		if bftStatus.SelfCommitMsg != nil {
			return nil
		}
		if !self.getState().IsReady() {
			return nil
		}
		if proposer, forEmpty, done := bftStatus.EndorseDone(vbftCtx.Config.C); done {
			proposal := bftStatus.GetBlockProposal(proposer)

			// consensus ok, make endorsement
			if proposal == nil {
				self.fetchProposal(evt.blockNum, proposer)
				// restart timer
				self.timer.StartEventTimer(EventEndorseEmptyBlockTimeout, evt.blockNum)
			} else if err := self.commitBlock(vbftCtx, proposal, forEmpty); err != nil {
				return fmt.Errorf("failed to endorse for block %d on empty endorse timeout: %s", evt.blockNum, err)
			}
			return nil
		} else {
			log.Errorf("server %d: empty endorse timeout, no quorum", self.Index)
			if !self.getState().IsActive() {
				proposals := bftStatus.Proposals
				proposal := vbftCtx.GetHighestRankProposal(proposals)
				if proposal != nil {
					if err := self.endorseBlock(vbftCtx, proposal, true); err != nil {
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
		if proposer, forEmpty, done := bftStatus.CommitDone(vbftCtx); done {
			proposal := bftStatus.GetBlockProposal(proposer)
			if proposal == nil {
				self.restartSyncing()
				return fmt.Errorf("commit timeout, consensused proposal not available. need resync")
			}
			if !bftStatus.checkBlockSign(vbftCtx, proposal.Block, forEmpty, chainCfg.Quorum()) {
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
	case EventRebroadcast:
		self.RebroadcastMsgs(vbftCtx, evt.blockNum)
	}
	return nil
}

func (self *Server) processHeartbeatMsg(peerIdx uint32, msg *peerHeartbeatMsg) {
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

func (self *Server) endorseBlock(vbftCtx *VbftContext, proposal *blockProposalMsg, forEmpty bool) error {
	// for each round, one node can only endorse one block, or empty block
	if proposal.Block.getProposer() == self.Index {
		return nil
	}
	blkNum := proposal.GetBlockNum()
	bft := vbftCtx.BftStatus
	if !forEmpty {
		if bft.EndorseFailed(vbftCtx.Config.C) {
			forEmpty = true
			log.Errorf("server %d, endorsing %d, changed from true to false", self.Index, blkNum)
		}
	}

	// check if has endorsed
	if !forEmpty && bft.HasEndorsedForBlock() {
		return nil
	} else if forEmpty && bft.HasEndorsedForEmptyBlock() {
		return nil
	}

	// build endorsement msg
	endorseMsg, err := self.constructEndorseMsg(proposal, forEmpty)
	if err != nil {
		return fmt.Errorf("failed to construct endorse msg: %s", err)
	}

	bft.setProposalEndorsed(endorseMsg)
	// if node is endorser of current round
	if forEmpty || vbftCtx.IsEndorser(self.Index) {
		log.Infof("endorser %d, endorsed block %d, from server %d",
			self.Index, blkNum, proposal.Block.getProposer())
		// broadcast my endorsement
		self.broadcast(endorseMsg)
	}
	self.processMsgEvent(endorseMsg)

	if !forEmpty {
		self.timer.StartEventTimer(EventEndorseBlockTimeout, blkNum)
	} else {
		self.timer.StartEventTimer(EventEndorseEmptyBlockTimeout, blkNum)
	}

	return nil
}

func (self *Server) commitBlock(vbftCtx *VbftContext, proposal *blockProposalMsg, forEmpty bool) error {
	// for each round, we can only commit one block
	if proposal.Block.getProposer() == self.Index {
		return nil
	}
	blkNum := proposal.GetBlockNum()
	bft := vbftCtx.BftStatus
	if bft.SelfCommitMsg != nil {
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
	endorses := bft.GetEndorseSigInfos(blkHash)
	// build commit msg
	commitMsg, err := self.constructCommitMsg(proposal, endorses, forEmpty)
	if err != nil {
		return fmt.Errorf("failed to construct commit msg: %s", err)
	}

	bft.SelfCommitMsg = commitMsg
	// if node is committer of current round
	if forEmpty || vbftCtx.IsCommitter(self.Index) {
		log.Infof("committer %d, commit block %d for proposer %d",
			self.Index, blkNum, proposal.Block.getProposer())
		self.broadcast(commitMsg)
	}
	self.processMsgEvent(commitMsg)

	self.timer.StartEventTimer(EventCommitBlockTimeout, blkNum)
	return nil
}

func (self *Server) sealBlock(block *VbftBlock, empty bool, sigdata bool) error {
	sealedBlkNum := block.getBlockNum()
	vbftCtx := self.GetVbftContext()

	sealedBlock, result, err := self.SetBlockSealed(vbftCtx, block, empty, sigdata)
	if err != nil {
		return fmt.Errorf("failed to seal proposal: %s", err)
	}

	self.timer.OnBlockSealed(sealedBlkNum)
	self.msgPool.OnBlockSealed(sealedBlkNum)

	h := sealedBlock.Block.Hash()
	prevBlkHash := sealedBlock.getPrevBlockHash()
	log.Infof("server %d, sealed block %d, proposer %d, prevhash: %s, hash: %s, root: %s", self.Index,
		sealedBlkNum, block.getProposer(), prevBlkHash.ToHexString(), h.ToHexString(), result.MerkleRoot.ToHexString())
	submitMsg, err := self.constructBlockSubmitMsg(sealedBlkNum, result.MerkleRoot)
	if err != nil {
		log.Errorf("constructBlockSubmitMsg blockNum:%d,err:%s", sealedBlkNum, err)
	} else {
		self.broadcast(submitMsg)
	}

	self.updateVbftContext(sealedBlock.Block, sealedBlock.Info, result)
	self.CheckAndSubmitBlock(sealedBlkNum, self.GetVbftContext().PrevBlockInfo.MerkleRoot)
	self.startNewRound()
	self.heartbeat(math.MaxUint32)
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

func CreateGovernaceTransaction(blkNum uint32) *types.Transaction {
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

func (self *Server) IsViewIncreased(view uint32, writeSet *overlaydb.MemDB) bool {
	goveranceview, err := GetGovernanceView(writeSet)
	if err != nil {
		log.Errorf("IsViewIncreased err:%s", err)
		return false
	}
	return goveranceview.View > view
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

func (self *Server) GetNewBlockConfig(vbftCtx *VbftContext) (*vconfig.ChainConfig, error) {
	writeSet := vbftCtx.PrevBlockInfo.WriteSet
	blkNum := vbftCtx.BlockNum
	needUpdateTx := vbftCtx.NeedUpdateChainConfigTx()
	if needUpdateTx || self.IsViewIncreased(vbftCtx.Config.View, writeSet) {
		chainconfig, err := getChainConfig(writeSet, blkNum, needUpdateTx)
		if err != nil {
			return nil, fmt.Errorf("getChainConfig failed:%s", err)
		}
		return chainconfig, nil
	}
	return nil, nil
}

func (self *Server) makeProposal(blkNum uint32, forEmpty bool) error {
	vbftCtx := self.GetVbftContext()
	if blkNum != vbftCtx.BlockNum {
		return fmt.Errorf("server %d ignore deprecatd blk proposal %d, current %d",
			self.Index, blkNum, vbftCtx.BlockNum)
	}

	if self.nonConsensusNode() {
		return fmt.Errorf("%d quit consensus node", self.Index)
	}

	cfg, err := self.GetNewBlockConfig(vbftCtx)
	if err != nil {
		return fmt.Errorf("getChainConfig failed:%s", err)
	}

	validHeight := self.validHeight(blkNum)
	userTxs := make([]*types.Transaction, 0)
	if cfg != nil {
		forEmpty = true
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

	proposal, err := self.constructProposalMsg(vbftCtx, userTxs, cfg)
	if err != nil {
		return fmt.Errorf("failed to construct proposal: %s", err)
	}

	log.Infof("server %d make proposal for block %d", self.Index, blkNum)

	self.processMsgEvent(proposal)
	self.broadcast(proposal)
	return nil
}

func (self *Server) makeSealed(proposal *blockProposalMsg, forEmpty bool) error {
	blkNum := proposal.GetBlockNum()

	log.Infof("server %d ready to seal block %d, for proposer %d, empty: %t",
		self.Index, blkNum, proposal.Block.getProposer(), forEmpty)

	// seal the block
	if proposal.GetBlockNum() != self.GetCurrentBlockNo() {
		return nil
	}
	// for each round, we can only seal one block
	return self.sealBlock(proposal.Block, forEmpty, true)
}

func (self *Server) reBroadcastCurrentRoundMsgs() {
	go func() {
		self.timer.C <- &TimerEvent{
			evtType:  EventRebroadcast,
			blockNum: self.GetCurrentBlockNo(),
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

func (self *Server) handleProposalTimeout(vbftCtx *VbftContext, evt *TimerEvent) error {
	if vbftCtx.BftStatus.HasEndorsedForBlock() {
		return nil
	}
	if !self.getState().IsReady() {
		return nil
	}
	proposals := vbftCtx.BftStatus.Proposals

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
				log.Infof("server %d proposed empty block for blk %d", self.Index, evt.blockNum)
			}
			self.timer.StartEventTimer(EventPropose2ndBlockTimeout, evt.blockNum)
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

	if err := self.endorseBlock(vbftCtx, proposal, false); err != nil {
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
