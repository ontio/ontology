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
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	shardmsg "github.com/ontio/ontology/core/chainmgr/message"
	"github.com/ontio/ontology/core/ledger"
	com "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/events"
	"github.com/ontio/ontology/events/message"
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
func (self *ChainManager) LoadFromLedger(mainLedger *ledger.Ledger) error {
	processedBlockHeight, err := mainLedger.GetShardProcessedBlockHeight()
	if err != nil {
		return fmt.Errorf("chainmgr: failed to read processed block height: %s", err)
	}
	self.processedParentBlockHeight = processedBlockHeight
	self.mainLedger = mainLedger
	self.SetShardLedger(types.NewShardIDUnchecked(0), mainLedger)
	// TODO: load parent-shard/sib-shard blockhdrs from ledger

	// start listen on local shard events
	self.localEventSub = events.NewActorSubscriber(self.localPid)
	self.localEventSub.Subscribe(message.TOPIC_SHARD_SYSTEM_EVENT)
	self.localEventSub.Subscribe(message.TOPIC_SAVE_BLOCK_COMPLETE)

	// get child-shards from shard-mgmt contract
	globalState, err := GetShardMgmtGlobalState(mainLedger)
	if err != nil {
		return fmt.Errorf("chainmgr: failed to read shard-mgmt global state: %s", err)
	}
	if globalState == nil {
		// not initialized from ledger
		log.Info("chainmgr: shard-mgmt not initialized, skipped loading from ledger")
		return nil
	}

	// load all child-shard from shard-mgmt contract
	for i := uint16(1); i < globalState.NextSubShardIndex; i++ {
		subShardID, err := self.shardID.GenSubShardID(i)
		if err != nil {
			return err
		}
		shard, err := GetShardState(mainLedger, subShardID)
		if err == com.ErrNotFound {
			continue
		}
		if err != nil {
			return fmt.Errorf("get shard %d failed: %s", subShardID, err)
		}
		// skip if shard is not active
		if shard.State != shardstates.SHARD_STATE_ACTIVE {
			continue
		}
		if _, err := self.initShardInfo(subShardID, shard); err != nil {
			return fmt.Errorf("init shard %d failed: %s", subShardID, err)
		}
		if shardInfo := self.shards[subShardID]; shardInfo != nil {
			cfg, err := self.buildShardConfig(subShardID, shard)
			if err != nil {
				return fmt.Errorf("init shard %d, failed to build config: %s", subShardID, err)
			}
			shardInfo.Config = cfg
		}
	}
	return nil
}

func (self *ChainManager) StartShardServer() error {
	return nil
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

func (self *ChainManager) Stop() {
	close(self.quitC)
	self.quitWg.Wait()
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
