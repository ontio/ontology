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
	"github.com/ontio/ontology/core/chainmgr/xshard"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/ledger"
	com "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/types"
	ontErr "github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/events"
	"github.com/ontio/ontology/events/message"
	actor2 "github.com/ontio/ontology/http/base/actor"
	hserver "github.com/ontio/ontology/http/base/actor"
	bcomm "github.com/ontio/ontology/http/base/common"
	"github.com/ontio/ontology/p2pserver/actor/req"
	"github.com/ontio/ontology/p2pserver/actor/server"
	p2p "github.com/ontio/ontology/p2pserver/common"
	shardstates "github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/txnpool"
	tc "github.com/ontio/ontology/txnpool/common"
	"github.com/ontio/ontology/validator/stateful"
	"github.com/ontio/ontology/validator/stateless"
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
	ShardID  common.ShardID
	SeedList []string
	Config   *config.OntologyConfig
}

type ChainManager struct {
	shardID common.ShardID

	// ShardInfo management, indexing shards with ShardID / Sender-Addr
	lock       sync.RWMutex
	shards     map[common.ShardID]*ShardInfo
	mainLedger *ledger.Ledger
	consensus  consensus.ConsensusService

	account *account.Account

	// send transaction to local
	txPoolPid *actor.PID
	p2pPid    *actor.PID
	localPid  *actor.PID
	mgr       *txnpool.TxnPoolManager

	// subscribe local SHARD_EVENT from shard-system-contract and BLOCK-EVENT from ledger
	localEventSub  *events.ActorSubscriber
	localBlockMsgC chan *message.SaveBlockCompleteMsg

	quitC  chan struct{}
	quitWg sync.WaitGroup
}

//
// Initialize chain manager when ontology starting
//
func Initialize(shardID common.ShardID, acc *account.Account) (*ChainManager, error) {
	if defaultChainManager != nil {
		return nil, fmt.Errorf("chain manager had been initialized for shard: %d", defaultChainManager.shardID)
	}

	xshard.InitShardBlockPool(shardID, CAP_SHARD_BLOCK_POOL)

	chainMgr := &ChainManager{
		shardID:        shardID,
		shards:         make(map[common.ShardID]*ShardInfo),
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

	shardState, err := xshard.GetShardState(self.mainLedger, self.shardID)
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
	cfg := config.DefConfig
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

	mainShardID := common.NewShardIDUnchecked(config.DEFAULT_SHARD_ID)
	mainShardInfo := &ShardInfo{
		ShardID:  mainShardID,
		SeedList: cfg.Genesis.SeedList,
		Config:   cfg,
	}
	self.shards[mainShardID] = mainShardInfo
	self.mainLedger = lgr
	ledger.DefLedger = lgr
	log.Infof("main ledger init success")
	return nil
}

func (self *ChainManager) initShardLedger(shardInfo *ShardInfo) error {
	if shardInfo != nil && ledger.GetShardLedger(shardInfo.ShardID) != nil {
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
	return nil
}

func (self *ChainManager) GetActiveShards() []common.ShardID {
	shards := make([]common.ShardID, 0)
	for _, shardInfo := range self.shards {
		shards = append(shards, shardInfo.ShardID)
	}
	return shards
}

func (self *ChainManager) GetDefaultLedger() *ledger.Ledger {
	if shardInfo := self.shards[self.shardID]; shardInfo != nil {
		return ledger.GetShardLedger(self.shardID)
	}
	return self.mainLedger
}

func (self *ChainManager) startConsensus() error {
	if self.consensus != nil {
		return nil
	}

	// start consensus
	shardInfo := self.shards[self.shardID]
	if shardInfo == nil {
		log.Infof("shard %d starting consensus, shard info not available", self.shardID.ToUint64())
		return nil
	}
	if shardInfo.Config == nil {
		log.Infof("shard %d starting consensus, shard config not available", self.shardID.ToUint64())
		return nil
	}
	lgr := ledger.GetShardLedger(self.shardID)
	if lgr == nil {
		log.Infof("shard %d starting consensus, shard ledger not available", self.shardID.ToUint64())
		return nil
	}

	// TODO: check if peer should start consensus
	if !shardInfo.Config.Consensus.EnableConsensus {
		return nil
	}

	consensusType := shardInfo.Config.Genesis.ConsensusType
	consensusService, err := consensus.NewConsensusService(consensusType, self.shardID, self.account,
		self.txPoolPid, lgr, self.p2pPid)
	if err != nil {
		return fmt.Errorf("NewConsensusService:%s error:%s", consensusType, err)
	}
	consensusService.Start()
	self.consensus = consensusService

	actor2.SetConsensusPid(consensusService.GetPID())
	req.SetConsensusPid(consensusService.GetPID())
	return nil
}

func (self *ChainManager) initTxPool() (*actor.PID, error) {
	lgr := ledger.GetShardLedger(self.shardID)
	if lgr == nil {
		log.Infof("shard %d starting consensus, shard ledger not available", self.shardID.ToUint64())
		return nil, nil
	}
	srv, err := self.mgr.StartTxnPoolServer(self.shardID, lgr)
	if err != nil {
		return nil, fmt.Errorf("Init txpool error:%s", err)
	}
	stlValidator, _ := stateless.NewValidator(fmt.Sprintf("stateless_validator_%d", self.shardID.ToUint64()))
	stlValidator.Register(srv.GetPID(tc.VerifyRspActor))
	stlValidator2, _ := stateless.NewValidator(fmt.Sprintf("stateless_validator2_%d", self.shardID.ToUint64()))
	stlValidator2.Register(srv.GetPID(tc.VerifyRspActor))
	stfValidator, _ := stateful.NewValidator(fmt.Sprintf("stateful_validator_%d", self.shardID.ToUint64()), lgr)
	stfValidator.Register(srv.GetPID(tc.VerifyRspActor))

	hserver.SetTxnPoolPid(srv.GetPID(tc.TxPoolActor))
	hserver.SetTxPid(srv.GetPID(tc.TxActor))
	SetTxPool(srv.GetPID(tc.TxActor))
	return self.mgr.GetPID(self.shardID, tc.TxActor), nil
}

func (self *ChainManager) Start(txPoolPid, p2pPid *actor.PID, txPoolMgr *txnpool.TxnPoolManager) error {
	self.txPoolPid = txPoolPid
	self.p2pPid = p2pPid
	self.mgr = txPoolMgr
	// start listen on local shard events
	self.localEventSub = events.NewActorSubscriber(self.localPid)
	self.localEventSub.Subscribe(message.TOPIC_SHARD_SYSTEM_EVENT)
	self.localEventSub.Subscribe(message.TOPIC_SAVE_BLOCK_COMPLETE)

	syncerToStart := make([]common.ShardID, 0)
	for shardId := self.shardID; shardId.ToUint64() != config.DEFAULT_SHARD_ID; shardId = shardId.ParentID() {
		syncerToStart = append(syncerToStart, shardId)
	}
	// start syncing root-chain
	syncerToStart = append(syncerToStart, common.NewShardIDUnchecked(config.DEFAULT_SHARD_ID))

	// start syncing shard
	for i := len(syncerToStart) - 1; i >= 0; i-- {
		shardId := syncerToStart[i]
		if self.shards[shardId] != nil {
			p2pPid.Tell(&server.StartSync{
				ShardID: shardId.ToUint64(),
			})
			log.Infof("start sync %d", shardId)
		}
	}

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
		}
	}

	return gasEvents
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
			ledgerSize := len(ledger.DefLedgerMgr.Ledgers)
			evts := self.handleShardSysEvents(msg.ShardSysEvents)
			blk := msg.Block
			if err := self.onBlockPersistCompleted(blk, evts); err != nil {
				log.Errorf("processing shard %d, block %d, err: %s", self.shardID, blk.Header.Height, err)
			}
			if ledgerSize < 2 {
				self.p2pPid.Tell(
					&p2p.AddBlock{
						Height:  blk.Header.Height,
						ShardID: blk.Header.ShardID,
					})
			}
		case <-self.quitC:
			return
		}
	}
}

func (self *ChainManager) Close() {
	close(self.quitC)
	self.quitWg.Wait()
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

	if tx.ShardID == self.shardID.ToUint64() {
		errCode, errString := bcomm.SendTxToPool(tx)
		if errCode != ontErr.ErrNoError {
			return fmt.Errorf(errString)
		}
	} else {
		if len(shardPeerIPList) == 0 {
			return fmt.Errorf("send raw tx failed: no shard peers")
		}
		if err := sendRawTx(tx, shardPeerIPList[0], shardPort); err != nil {
			return fmt.Errorf("send raw tx failed: %s, shardAddr %s:%d", err, shardPeerIPList[0], shardPort)
		}
	}
	return nil
}

func (self *ChainManager) getShardRPCPort(shardID common.ShardID) uint {
	// TODO: get from shardinfo
	return 20336
}
