/*
 * Copyright (C) 2019 The ontology Authors
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

package chainmgr

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/consensus"
	shardmsg "github.com/ontio/ontology/core/chainmgr/message"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/ledger"
	com "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/events"
	"github.com/ontio/ontology/events/message"
	actor2 "github.com/ontio/ontology/http/base/actor"
	"github.com/ontio/ontology/p2pserver/actor/req"
	shardstates "github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
)

const (
	CAP_LOCAL_SHARDMSG_CHNL = 64
	CAP_SHARD_BLOCK_POOL    = 16
)

var defaultChainManager *ChainManager = nil

//
// ShardInfo:
//  . Configuration of other shards
//  . seed list of other shards
//
type ShardInfo struct {
	ShardID  types.ShardID
	SeedList []string
	Config   *config.OntologyConfig
	Ledger   *ledger.Ledger
}

type ChainManager struct {
	shardID types.ShardID

	// ShardInfo management, indexing shards with ShardID / Sender-Addr
	lock       sync.RWMutex
	shards     map[types.ShardID]*ShardInfo
	mainLedger *ledger.Ledger
	consensus  consensus.ConsensusService

	// BlockHeader and Cross-Shard Txs of other shards
	// FIXME: check if any lock needed (if not only accessed by remoteShardMsgLoop)
	// TODO: persistent
	blockPool *shardmsg.ShardBlockPool

	// last local block processed by ChainManager
	processedParentBlockHeight uint32
	account                    *account.Account

	// send transaction to local
	txPoolPid *actor.PID
	p2pPid    *actor.PID
	localPid  *actor.PID

	// subscribe local SHARD_EVENT from shard-system-contract and BLOCK-EVENT from ledger
	localEventSub  *events.ActorSubscriber
	localBlockMsgC chan *message.SaveBlockCompleteMsg

	quitC  chan struct{}
	quitWg sync.WaitGroup
}

//
// Initialize chain manager when ontology starting
//
func Initialize(shardID types.ShardID, acc *account.Account) (*ChainManager, error) {
	if defaultChainManager != nil {
		return nil, fmt.Errorf("chain manager had been initialized for shard: %d", defaultChainManager.shardID)
	}

	blkPool := shardmsg.NewShardBlockPool(CAP_SHARD_BLOCK_POOL)
	if blkPool == nil {
		return nil, fmt.Errorf("chainmgr init: failed to init block pool")
	}

	chainMgr := &ChainManager{
		shardID:        shardID,
		shards:         make(map[types.ShardID]*ShardInfo),
		blockPool:      blkPool,
		localBlockMsgC: make(chan *message.SaveBlockCompleteMsg, CAP_LOCAL_SHARDMSG_CHNL),
		quitC:          make(chan struct{}),

		account: acc,
	}
	go chainMgr.localEventLoop()
	props := actor.FromProducer(func() actor.Actor {
		return chainMgr
	})
	pid, err := actor.SpawnNamed(props, GetShardName(shardID))
	if err == nil {
		chainMgr.localPid = pid
	}
	defaultChainManager = chainMgr
	return defaultChainManager, nil
}

//
// LoadFromLedger when ontology starting, after ledger initialized.
//
func (self *ChainManager) LoadFromLedger(stateHashHeight uint32) error {
	if err := self.initMainLedger(stateHashHeight); err != nil {
		return err
	}

	if self.shardID.ToUint64() == config.DEFAULT_SHARD_ID {
		// main-chain node, not need to process shard-events
		return nil
	}

	processedBlockHeight, err := self.mainLedger.GetShardProcessedBlockHeight()
	if err != nil {
		return fmt.Errorf("chainmgr: failed to read processed block height: %s", err)
	}
	self.processedParentBlockHeight = processedBlockHeight

	shardState, err := GetShardState(self.mainLedger, self.shardID)
	if err == com.ErrNotFound {
		return nil
	}
	if err != nil {
		return fmt.Errorf("get shard %d failed: %s", self.shardID, err)
	}
	// skip if shard is not active
	if shardState.State != shardstates.SHARD_STATE_ACTIVE {
		return nil
	}
	if _, err := self.initShardInfo(self.shardID, shardState); err != nil {
		return fmt.Errorf("init shard %d failed: %s", self.shardID, err)
	}
	shardInfo := self.shards[self.shardID]
	if shardInfo == nil {
		return nil
	}
	cfg, err := self.buildShardConfig(self.shardID, shardState)
	if err != nil {
		return fmt.Errorf("init shard %d, failed to build config: %s", self.shardID, err)
	}
	shardInfo.Config = cfg

	if err := self.initShardLedger(shardInfo); err != nil {
		return fmt.Errorf("init shard %d, failed to init ledger: %s", self.shardID, err)
	}

	return nil
}

func (self *ChainManager) initMainLedger(stateHashHeight uint32) error {
	dbDir := utils.GetStoreDirPath(config.DefConfig.Common.DataDir, config.DefConfig.P2PNode.NetworkName)
	lgr, err := ledger.NewLedger(dbDir, stateHashHeight)
	if err != nil {
		return fmt.Errorf("NewLedger error:%s", err)
	}
	bookKeepers, err := config.DefConfig.GetBookkeepers()
	if err != nil {
		return fmt.Errorf("GetBookkeepers error:%s", err)
	}
	genesisConfig := config.DefConfig.Genesis
	shardConfig := config.DefConfig.Shard
	genesisBlock, err := genesis.BuildGenesisBlock(bookKeepers, genesisConfig, shardConfig)
	if err != nil {
		return fmt.Errorf("genesisBlock error %s", err)
	}
	err = lgr.Init(bookKeepers, genesisBlock)
	if err != nil {
		return fmt.Errorf("Init ledger error:%s", err)
	}

	mainShardID := types.NewShardIDUnchecked(config.DEFAULT_SHARD_ID)
	mainShardInfo := &ShardInfo{
		ShardID:  mainShardID,
		SeedList: config.DefConfig.Genesis.SeedList,
		Config:   config.DefConfig,
		Ledger:   lgr,
	}
	self.shards[mainShardID] = mainShardInfo
	self.mainLedger = lgr
	log.Infof("main ledger init success")
	return nil
}

func (self *ChainManager) initShardLedger(shardInfo *ShardInfo) error {
	if shardInfo.Ledger != nil {
		return nil
	}
	if self.shardID.ToUint64() == config.DEFAULT_SHARD_ID {
		return fmt.Errorf("init main ledger as shard ledger")
	}
	if self.mainLedger == nil {
		return fmt.Errorf("init shard ledger with nil main ledger")
	}
	dbDir := utils.GetStoreDirPath(config.DefConfig.Common.DataDir, config.DefConfig.P2PNode.NetworkName)
	lgr, err := ledger.NewShardLedger(self.shardID, dbDir, self.mainLedger)
	if err != nil {
		return fmt.Errorf("init shard ledger: %s", err)
	}

	bookKeepers, err := shardInfo.Config.GetBookkeepers()
	if err != nil {
		return fmt.Errorf("init shard ledger: GetBookkeepers error:%s", err)
	}
	genesisConfig := shardInfo.Config.Genesis
	shardConfig := shardInfo.Config.Shard
	genesisBlock, err := genesis.BuildGenesisBlock(bookKeepers, genesisConfig, shardConfig)
	if err != nil {
		return fmt.Errorf("init shard ledger: genesisBlock error %s", err)
	}
	err = lgr.Init(bookKeepers, genesisBlock)
	if err != nil {
		return fmt.Errorf("init shard ledger: :%s", err)
	}
	shardInfo.Ledger = lgr
	return nil
}

func (self *ChainManager) GetActiveShards() map[types.ShardID]*ledger.Ledger {
	lgrs := make(map[types.ShardID]*ledger.Ledger)
	for _, shardInfo := range self.shards {
		lgrs[shardInfo.ShardID] = shardInfo.Ledger
	}
	return lgrs
}

func (self *ChainManager) GetDefaultLedger() *ledger.Ledger {
	if shardInfo := self.shards[self.shardID]; shardInfo != nil {
		if shardInfo.Ledger != nil {
			return shardInfo.Ledger
		}
	}
	return self.mainLedger
}

func (self *ChainManager) startConsensus() error {
	if self.consensus != nil {
		return nil
	}

	// start consensus
	if shardInfo := self.shards[self.shardID]; shardInfo.Config != nil && shardInfo.Ledger != nil {
		// TODO: check if peer should start consensus
		if !shardInfo.Config.Consensus.EnableConsensus {
			return nil
		}

		consensusType := shardInfo.Config.Genesis.ConsensusType
		consensusService, err := consensus.NewConsensusService(consensusType, self.shardID, self.account,
			self.txPoolPid, shardInfo.Ledger, self.p2pPid)
		if err != nil {
			return fmt.Errorf("NewConsensusService:%s error:%s", consensusType, err)
		}
		consensusService.Start()
		self.consensus = consensusService

		actor2.SetConsensusPid(consensusService.GetPID())
		req.SetConsensusPid(consensusService.GetPID())
	}
	return nil
}

func (self *ChainManager) Start(txPoolPid, p2pPid *actor.PID) error {
	self.txPoolPid = txPoolPid
	self.p2pPid = p2pPid

	// start listen on local shard events
	self.localEventSub = events.NewActorSubscriber(self.localPid)
	self.localEventSub.Subscribe(message.TOPIC_SHARD_SYSTEM_EVENT)
	self.localEventSub.Subscribe(message.TOPIC_SAVE_BLOCK_COMPLETE)

	return self.startConsensus()
}

func (self *ChainManager) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Restarting:
		log.Info("chain mgr actor restarting")
	case *actor.Stopping:
		log.Info("chain mgr actor stopping")
	case *actor.Stopped:
		log.Info("chain mgr actor stopped")
	case *actor.Started:
		log.Info("chain mgr actor started")
	case *actor.Restart:
		log.Info("chain mgr actor restart")
	case *message.SaveBlockCompleteMsg:
		self.localBlockMsgC <- msg

	default:
		log.Info("chain mgr actor: Unknown msg ", msg, "type", reflect.TypeOf(msg))
	}
}

// handle shard system contract event, other events are not handled and returned
func (self *ChainManager) handleShardSysEvents(shardEvts []*message.ShardSystemEventMsg) []*message.ShardEventState {
	var gasEvents []*message.ShardEventState
	for _, evt := range shardEvts {
		shardEvt := evt.Event
		if isShardGasEvent(shardEvt) {
			gasEvents = append(gasEvents, shardEvt)
			continue
		}

		switch shardEvt.EventType {

		case shardstates.EVENT_SHARD_CREATE:
			createEvt := &shardstates.CreateShardEvent{}
			if err := createEvt.Deserialization(common.NewZeroCopySource(shardEvt.Payload)); err != nil {
				log.Errorf("deserialize create shard event: %s", err)
				continue
			}
			if err := self.onShardCreated(createEvt); err != nil {
				log.Errorf("processing create shard event: %s", err)
			}
		case shardstates.EVENT_SHARD_CONFIG_UPDATE:
			cfgEvt := &shardstates.ConfigShardEvent{}
			if err := cfgEvt.Deserialization(common.NewZeroCopySource(shardEvt.Payload)); err != nil {
				log.Errorf("deserialize update shard config event: %s", err)
				continue
			}
			if err := self.onShardConfigured(cfgEvt); err != nil {
				log.Errorf("processing update shard config event: %s", err)
			}
		case shardstates.EVENT_SHARD_PEER_JOIN:
			jointEvt := &shardstates.PeerJoinShardEvent{}
			if err := jointEvt.Deserialization(common.NewZeroCopySource(shardEvt.Payload)); err != nil {
				log.Errorf("deserialize join shard event: %s", err)
				continue
			}
			if err := self.onShardPeerJoint(jointEvt); err != nil {
				log.Errorf("processing join shard event: %s", err)
			}
		case shardstates.EVENT_SHARD_ACTIVATED:
			evt := &shardstates.ShardActiveEvent{}
			if err := evt.Deserialization(common.NewZeroCopySource(shardEvt.Payload)); err != nil {
				log.Errorf("deserialize shard activation event: %s", err)
				continue
			}
			if err := self.onShardActivated(evt); err != nil {
				log.Errorf("processing shard activation event: %s", err)
			}
		case shardstates.EVENT_SHARD_PEER_LEAVE:
		case shardstates.EVENT_SHARD_GAS_WITHDRAW_REQ:
			evt := &shardstates.WithdrawGasReqEvent{}
			if err := evt.Deserialization(common.NewZeroCopySource(shardEvt.Payload)); err != nil {
				log.Errorf("deserialize shard activation event: %s", err)
				continue
			}
			self.onWithdrawGasReq(evt)
		case shardstates.EVENT_SHARD_COMMIT_DPOS:
		}
	}

	return gasEvents
}

func isShardGasEvent(evt *message.ShardEventState) bool {
	switch evt.EventType {
	case shardstates.EVENT_SHARD_GAS_DEPOSIT, shardstates.EVENT_SHARD_GAS_WITHDRAW_DONE:
		return true
	}
	return false
}

//
// localEventLoop: process all local shard-event.
//   shard-events are from shard system contracts (shard-mgmt, shard-gas, shard-mq, shard-ccmc)
//
func (self *ChainManager) localEventLoop() {
	self.quitWg.Add(1)
	defer self.quitWg.Done()

	for {
		select {
		case msg := <-self.localBlockMsgC:
			evts := self.handleShardSysEvents(msg.ShardSysEvents)
			blk := msg.Block
			if err := self.onBlockPersistCompleted(blk, evts); err != nil {
				log.Errorf("processing shard %d, block %d, err: %s", self.shardID, blk.Header.Height, err)
			}
		case <-self.quitC:
			return
		}
	}
}

func (self *ChainManager) Close() {
	close(self.quitC)
	self.quitWg.Wait()

	// close ledgers
	lgr := self.GetDefaultLedger()
	lgr.Close()
}

func (self *ChainManager) Stop() {
	// TODO
}

//
// send Cross-Shard Tx to remote shard
// TODO: get ip-address of remote shard node
//
func (self *ChainManager) sendCrossShardTx(tx *types.Transaction, shardPeerIPList []string, shardPort uint) error {
	// FIXME: broadcast Tx to seed nodes of target shard

	// relay with parent shard
	//payload := new(bytes.Buffer)
	//if err := tx.Serialize(payload); err != nil {
	//	return fmt.Errorf("failed to serialize tx: %s", err)
	//}
	//
	//msg := &shardmsg.CrossShardMsg{
	//	Version: shardmsg.SHARD_PROTOCOL_VERSION,
	//	Type:    shardmsg.TXN_RELAY_MSG,
	//	Sender:  self.parentPid,
	//	Data:    payload.Bytes(),
	//}
	//self.sendShardMsg(self.parentShardID, msg)
	//return nil
	if len(shardPeerIPList) == 0 {
		return fmt.Errorf("send raw tx failed: no shard peers")
	}
	if err := sendRawTx(tx, shardPeerIPList[0], shardPort); err != nil {
		return fmt.Errorf("send raw tx failed: %s, shardAddr %s:%d", err, shardPeerIPList[0], shardPort)
	}

	return nil
}

func (self *ChainManager) getShardRPCPort(shardID types.ShardID) uint {
	// TODO: get from shardinfo
	return 0
}

func (self *ChainManager) SetShardLedger(shardID types.ShardID, ledger *ledger.Ledger) {
	self.lock.Lock()
	defer self.lock.Unlock()
	self.shards[shardID] = &ShardInfo{
		Ledger: ledger,
	}
}
