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
	"math"
	"reflect"
	"sync"
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology-eventbus/eventstream"
	"github.com/ontio/ontology-eventbus/remote"
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
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
)

const (
	CAP_LOCAL_SHARDMSG_CHNL  = 64
	CAP_REMOTE_SHARDMSG_CHNL = 64
	CAP_SHARD_BLOCK_POOL     = 16
	CONNECT_PARENT_TIMEOUT   = 5 * time.Second
)

const (
	CONN_TYPE_UNKNOWN = iota
	CONN_TYPE_SELF
	CONN_TYPE_PARENT
	CONN_TYPE_CHILD
	CONN_TYPE_SIB
)

var defaultChainManager *ChainManager = nil

//
// RemoteMsg: for messages received from event-bus.
//  @Sender : where the message is sent from
//  @Msg : msg payload
//
type RemoteMsg struct {
	Sender *actor.PID
	Msg    shardmsg.RemoteShardMsg
}

//
// MsgSendReq: request BroadcastLoop() to send msg to other shard
//
type MsgSendReq struct {
	targetShardID types.ShardID
	msg           *shardmsg.CrossShardMsg
}

//
// ShardInfo:
//  . Configuration of other shards
//  . EventBus ID of other shards
//
type ShardInfo struct {
	ShardID       types.ShardID
	ParentShardID types.ShardID
	ShardAddress  string
	SeedList      []string
	ConnType      int
	Connected     bool
	Config        *config.OntologyConfig
	Sender        *actor.PID
}

type ChainManager struct {
	shardID              types.ShardID
	shardPort            uint
	parentShardID        types.ShardID
	parentShardIPAddress string
	parentShardPort      uint

	// ShardInfo management, indexing shards with ShardID / Sender-Addr
	lock       sync.RWMutex
	shards     map[types.ShardID]*ShardInfo
	shardAddrs map[string]types.ShardID

	// BlockHeader and Cross-Shard Txs of other shards
	// FIXME: check if any lock needed (if not only accessed by remoteShardMsgLoop)
	// TODO: persistent
	blockPool *shardmsg.ShardBlockPool

	// last local block processed by ChainManager
	// FIXME: on restart, make sure catchup with latest blocks
	// TODO: persistent
	processedBlockHeight uint32

	account *account.Account

	// Ontology process command arguments, for child-shard ontology process creation
	cmdArgs map[string]string

	ledger *ledger.Ledger
	p2pPid *actor.PID

	// send transaction to local
	txPoolPid *actor.PID

	localBlockMsgC  chan *message.SaveBlockCompleteMsg
	remoteShardMsgC chan *RemoteMsg
	broadcastMsgC   chan *MsgSendReq
	parentConnWait  chan bool

	parentPid *actor.PID
	localPid  *actor.PID

	// subscribe local SHARD_EVENT from shard-system-contract and BLOCK-EVENT from ledger
	localEventSub *events.ActorSubscriber

	// subscribe event-bus disconnected event
	busEventSub *eventstream.Subscription

	quitC  chan struct{}
	quitWg sync.WaitGroup
}

//
// Initialize chain manager when ontology starting
//
func Initialize(shardID types.ShardID, parentAddr string, shardPort, parentPort uint, acc *account.Account, cmdArgs map[string]string) (*ChainManager, error) {
	// fixme: change to sync.once
	if defaultChainManager != nil {
		return nil, fmt.Errorf("chain manager had been initialized for shard: %d", defaultChainManager.shardID)
	}

	blkPool := shardmsg.NewShardBlockPool(CAP_SHARD_BLOCK_POOL)
	if blkPool == nil {
		return nil, fmt.Errorf("chainmgr init: failed to init block pool")
	}

	chainMgr := &ChainManager{
		shardID:              shardID,
		shardPort:            shardPort,
		parentShardID:        shardID.ParentID(),
		parentShardIPAddress: parentAddr,
		parentShardPort:      parentPort,
		shards:               make(map[types.ShardID]*ShardInfo),
		shardAddrs:           make(map[string]types.ShardID),
		blockPool:            blkPool,
		localBlockMsgC:       make(chan *message.SaveBlockCompleteMsg, CAP_LOCAL_SHARDMSG_CHNL),
		remoteShardMsgC:      make(chan *RemoteMsg, CAP_REMOTE_SHARDMSG_CHNL),
		broadcastMsgC:        make(chan *MsgSendReq, CAP_REMOTE_SHARDMSG_CHNL),
		parentConnWait:       make(chan bool),
		quitC:                make(chan struct{}),

		account: acc,
		cmdArgs: cmdArgs,
	}

	// start remote-eventBus
	chainMgr.startRemoteEventbus()
	// listening on remote-eventBus
	if err := chainMgr.startListener(); err != nil {
		return nil, fmt.Errorf("shard %d start listener failed: %s", chainMgr.shardID, err)
	}

	go chainMgr.localEventLoop()
	go chainMgr.remoteShardMsgLoop()
	go chainMgr.broadcastMsgLoop()

	if err := chainMgr.connectParent(); err != nil {
		chainMgr.Stop()
		return nil, fmt.Errorf("connect parent shard failed: %s", err)
	}

	// subscribe on event-bus disconnected event
	chainMgr.busEventSub = eventstream.Subscribe(chainMgr.processEventBusEvent).
		WithPredicate(func(m interface{}) bool {
			switch m.(type) {
			case *remote.EndpointTerminatedEvent:
				return true
			default:
				return false
			}
		})

	defaultChainManager = chainMgr
	return defaultChainManager, nil
}

//
// LoadFromLedger when ontology starting, after ledger initialized.
//
func (self *ChainManager) LoadFromLedger(lgr *ledger.Ledger) error {
	self.ledger = lgr

	processedBlockHeight, err := self.ledger.GetShardProcessedBlockHeight()
	if err != nil {
		return fmt.Errorf("chainmgr: failed to read processed block height: %s", err)
	}
	self.processedBlockHeight = processedBlockHeight

	// TODO: load parent-shard/sib-shard blockhdrs from ledger

	// start listen on local shard events
	self.localEventSub = events.NewActorSubscriber(self.localPid)
	self.localEventSub.Subscribe(message.TOPIC_SHARD_SYSTEM_EVENT)
	self.localEventSub.Subscribe(message.TOPIC_SAVE_BLOCK_COMPLETE)

	// get child-shards from shard-mgmt contract
	globalState, err := GetShardMgmtGlobalState(self.ledger)
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
		shard, err := GetShardState(self.ledger, subShardID)
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
	for shardID, _ := range self.shards {
		err := self.restartChildShardProcess(shardID)
		if err != nil {
			return fmt.Errorf("StartShardServer shardId:%d,err:%s", shardID, err)
		}
		log.Infof("StartShardServer shardId succ:%d", shardID)
	}
	return nil
}

func (self *ChainManager) startRemoteEventbus() {
	localRemote := fmt.Sprintf("%s:%d", config.DEFAULT_PARENTSHARD_IPADDR, self.shardPort)
	remote.Start(localRemote)
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
	case *shardmsg.CrossShardMsg:
		if msg == nil {
			return
		}
		log.Tracef("chain mgr received shard msg: %v", msg)
		smsg, err := shardmsg.DecodeShardMsg(msg.Type, msg.Data)
		if err != nil {
			log.Errorf("decode shard msg: %s", err)
			return
		}
		self.remoteShardMsgC <- &RemoteMsg{
			Sender: msg.Sender,
			Msg:    smsg,
		}

	case *message.SaveBlockCompleteMsg:
		self.localBlockMsgC <- msg

	default:
		log.Info("chain mgr actor: Unknown msg ", msg, "type", reflect.TypeOf(msg))
	}
}

//
// only process 'Disconnect' event from event stream
//
func (self *ChainManager) processEventBusEvent(evt interface{}) {
	switch evt := evt.(type) {
	case *remote.EndpointTerminatedEvent:
		self.remoteShardMsgC <- &RemoteMsg{
			Msg: &shardmsg.ShardDisconnectedMsg{
				Address: evt.Address,
			},
		}
	default:
		return
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

//
// broadcastMsgLoop: help broadcasting message to remote shards
//  TODO: send msg to sibling-shards
//
func (self *ChainManager) broadcastMsgLoop() {
	self.quitWg.Add(1)
	defer self.quitWg.Done()

	for {
		select {
		case msg := <-self.broadcastMsgC:
			if msg.targetShardID.ToUint64() == math.MaxUint64 {
				// broadcast
				for _, s := range self.shards {
					if s.Connected && s.Sender != nil {
						s.Sender.Tell(msg.msg)
					}
				}
			} else {
				// send to shard
				if s, present := self.shards[msg.targetShardID]; present {
					if s.Connected && s.Sender != nil {
						s.Sender.Tell(msg.msg)
					}
				} else {
					// other shards
					// TODO: send msg to sib shards
				}
			}
		case <-self.quitC:
			return
		}
	}
}

func (self *ChainManager) remoteShardMsgLoop() {
	self.quitWg.Add(1)
	defer self.quitWg.Done()

	for {
		if err := self.processRemoteShardMsg(); err != nil {
			log.Errorf("chain mgr process remote shard msg failed: %s", err)
		}
		select {
		case <-self.quitC:
			return
		default:
		}
	}
}

//
// processRemoteShardMsg: process messages from all remote shards
//
func (self *ChainManager) processRemoteShardMsg() error {
	select {
	case remoteMsg := <-self.remoteShardMsgC:
		msg := remoteMsg.Msg
		log.Debugf("shard %d received remote shard msg type %d", self.shardID, msg.Type())
		switch msg.Type() {
		case shardmsg.HELLO_MSG:
			helloMsg, ok := msg.(*shardmsg.ShardHelloMsg)
			if !ok {
				return fmt.Errorf("invalid hello msg")
			}
			if helloMsg.TargetShardID != self.shardID {
				return fmt.Errorf("hello msg with invalid target %d, from %d", helloMsg.TargetShardID, helloMsg.SourceShardID)
			}
			log.Infof("received hello msg from %d", helloMsg.SourceShardID)
			// response ack
			return self.onNewShardConnected(remoteMsg.Sender, helloMsg)
		case shardmsg.CONFIG_MSG:
			shardCfgMsg, ok := msg.(*shardmsg.ShardConfigMsg)
			if !ok {
				return fmt.Errorf("invalid config msg")
			}
			log.Infof("shard %d received config msg", self.shardID)
			return self.onShardConfig(remoteMsg.Sender, shardCfgMsg)
		case shardmsg.BLOCK_REQ_MSG:
			var header *types.Header
			var shardTx *shardmsg.ShardBlockTx
			var err error

			req := msg.(*shardmsg.ShardBlockReqMsg)
			info := self.getShardBlockInfo(self.shardID, req.BlockHeight)
			if info != nil {
				header = info.Header.Header
				shardTx = info.ShardTxs[req.ShardID]
			} else {
				header, err = self.ledger.GetHeaderByHeight(req.BlockHeight)
				if err != nil {
					return err
				}
				evts, err := self.ledger.GetBlockShardEvents(req.BlockHeight)
				if err != nil {
					return err
				}

				shardEvts := make([]*message.ShardEventState, 0)
				for _, evt := range evts {
					shardEvt := evt.Event
					if isShardGasEvent(shardEvt) && shardEvt.ToShard == req.ShardID {
						shardEvts = append(shardEvts, shardEvt)
					}
				}
				shardTx, err = newShardBlockTx(shardEvts)
				if err != nil {
					return err
				}
			}
			msg, err := shardmsg.NewShardBlockRspMsg(self.shardID, header, shardTx, self.localPid)
			if err != nil {
				return fmt.Errorf("build shard block msg: %s", err)
			}

			// send msg to shard
			self.sendShardMsg(req.ShardID, msg)
		case shardmsg.BLOCK_RSP_MSG:
			blkMsg, ok := msg.(*shardmsg.ShardBlockRspMsg)
			if !ok {
				return fmt.Errorf("invalid block rsp msg")
			}
			return self.onShardBlockReceived(remoteMsg.Sender, blkMsg)
		case shardmsg.PEERINFO_REQ_MSG:
		case shardmsg.PEERINFO_RSP_MSG:
			return nil
		case shardmsg.DISCONNECTED_MSG:
			disconnMsg, ok := msg.(*shardmsg.ShardDisconnectedMsg)
			if !ok {
				return fmt.Errorf("invalid disconnect message")
			}
			return self.onShardDisconnected(disconnMsg)
		default:
			return nil
		}
	case <-self.quitC:
		return nil
	}
	return nil
}

//
// Connect to parent shard, send Hello message.
// (root shard does not have parent shard)
//
func (self *ChainManager) connectParent() error {
	// connect to parent
	if self.parentShardID.ToUint64() == math.MaxUint64 {
		return nil
	}
	if self.localPid == nil {
		return fmt.Errorf("shard %d connect parent with nil localPid", self.shardID)
	}

	parentAddr := fmt.Sprintf("%s:%d", self.parentShardIPAddress, self.parentShardPort)
	parentPid := actor.NewPID(parentAddr, GetShardName(self.parentShardID))
	hellomsg, err := shardmsg.NewShardHelloMsg(self.shardID, self.parentShardID, self.localPid)
	if err != nil {
		return fmt.Errorf("build hello msg: %s", err)
	}
	parentPid.Request(hellomsg, self.localPid)
	if err := self.waitConnectParent(CONNECT_PARENT_TIMEOUT); err != nil {
		return fmt.Errorf("wait connection with parent err: %s", err)
	}

	self.parentPid = parentPid
	log.Infof("shard %d connected with parent shard %d", self.shardID, self.parentShardID)
	return nil
}

func (self *ChainManager) waitConnectParent(timeout time.Duration) error {
	select {
	case <-time.After(timeout):
		return fmt.Errorf("wait parent connection timeout")
	case connected := <-self.parentConnWait:
		if connected {
			return nil
		}
		return fmt.Errorf("connection failed")
	}
	return nil
}

func (self *ChainManager) notifyParentConnected() {
	select {
	case self.parentConnWait <- true:
	default:
		return
	}
}

//
// start listening on remote event bus
//
func (self *ChainManager) startListener() error {

	// start local
	props := actor.FromProducer(func() actor.Actor {
		return self
	})
	pid, err := actor.SpawnNamed(props, GetShardName(self.shardID))
	if err != nil {
		return fmt.Errorf("init chain manager actor: %s", err)
	}
	self.localPid = pid

	log.Infof("chain %d started listen on port %d", self.shardID, self.shardPort)
	return nil
}

func (self *ChainManager) Stop() {
	close(self.quitC)
	self.quitWg.Wait()
}

func (self *ChainManager) broadcastShardMsg(msg *shardmsg.CrossShardMsg) {
	self.broadcastMsgC <- &MsgSendReq{
		targetShardID: types.NewShardIDUnchecked(math.MaxUint64),
		msg:           msg,
	}
}

func (self *ChainManager) sendShardMsg(shardId types.ShardID, msg *shardmsg.CrossShardMsg) {
	log.Infof("send shard msg type %d from %d to %d", msg.GetType(), self.shardID, shardId)
	self.broadcastMsgC <- &MsgSendReq{
		targetShardID: shardId,
		msg:           msg,
	}
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
