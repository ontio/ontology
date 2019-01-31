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
	"bytes"
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

type RemoteMsg struct {
	Sender *actor.PID
	Msg    shardmsg.RemoteShardMsg
}

type MsgSendReq struct {
	targetShardID uint64
	msg           *shardmsg.CrossShardMsg
}

type ShardInfo struct {
	ShardID       uint64
	ParentShardID uint64
	ShardAddress  string
	ConnType      int
	Connected     bool
	Config        *config.OntologyConfig
	Sender        *actor.PID
}

type ShardStorageReqList map[common.Address]*shardmsg.StorageRequest

type ChainManager struct {
	shardID              uint64
	shardPort            uint
	parentShardID        uint64
	parentShardIPAddress string
	parentShardPort      uint

	lock       sync.RWMutex
	shards     map[uint64]*ShardInfo
	shardAddrs map[string]uint64
	blockPool  *shardmsg.ShardBlockPool

	account      *account.Account
	genesisBlock *types.Block
	cmd          string
	cmdArgs      map[string]string

	ledger    *ledger.Ledger
	p2pPid    *actor.PID
	txPoolPid *actor.PID

	localBlockMsgC  chan *types.Block
	localEventC     chan *shardstates.ShardEventState
	remoteShardMsgC chan *RemoteMsg
	broadcastMsgC   chan *MsgSendReq
	parentConnWait  chan bool

	txnReqC     chan shardmsg.ShardTxRequest
	txnRspC     chan shardmsg.ShardTxResponse
	pendingTxns map[common.Uint256]*shardmsg.TxRequest
	pendingStorageReqs map[uint64]ShardStorageReqList

	parentPid   *actor.PID
	localPid    *actor.PID
	sub         *events.ActorSubscriber
	endpointSub *eventstream.Subscription

	quitC  chan struct{}
	quitWg sync.WaitGroup
}

func Initialize(shardID, parentShardID uint64, parentAddr string, shardPort, parentPort uint, acc *account.Account, cmdArgs map[string]string) (*ChainManager, error) {
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
		parentShardID:        parentShardID,
		parentShardIPAddress: parentAddr,
		parentShardPort:      parentPort,
		shards:               make(map[uint64]*ShardInfo),
		shardAddrs:           make(map[string]uint64),
		blockPool:            blkPool,
		localBlockMsgC:       make(chan *types.Block, CAP_LOCAL_SHARDMSG_CHNL),
		localEventC:          make(chan *shardstates.ShardEventState, CAP_LOCAL_SHARDMSG_CHNL),
		remoteShardMsgC:      make(chan *RemoteMsg, CAP_REMOTE_SHARDMSG_CHNL),
		broadcastMsgC:        make(chan *MsgSendReq, CAP_REMOTE_SHARDMSG_CHNL),
		parentConnWait:       make(chan bool),
		quitC:                make(chan struct{}),
		txnReqC:              make(chan shardmsg.ShardTxRequest, CAP_LOCAL_SHARDMSG_CHNL),
		txnRspC:              make(chan shardmsg.ShardTxResponse, CAP_REMOTE_SHARDMSG_CHNL),
		pendingTxns:          make(map[common.Uint256]*shardmsg.TxRequest),
		pendingStorageReqs:   make(map[uint64]ShardStorageReqList),

		account: acc,
		cmdArgs: cmdArgs,
	}

	chainMgr.startRemoteEventbus()
	if err := chainMgr.startListener(); err != nil {
		return nil, fmt.Errorf("shard %d start listener failed: %s", chainMgr.shardID, err)
	}

	go chainMgr.localEventLoop()
	go chainMgr.remoteShardMsgLoop()
	go chainMgr.broadcastMsgLoop()
	go chainMgr.txnLoop()

	if err := chainMgr.connectParent(); err != nil {
		chainMgr.Stop()
		return nil, fmt.Errorf("connect parent shard failed: %s", err)
	}

	chainMgr.endpointSub = eventstream.Subscribe(chainMgr.remoteEndpointEvent).
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

func (self *ChainManager) LoadFromLedger(lgr *ledger.Ledger) error {
	self.ledger = lgr

	// start listen on local actor
	self.sub = events.NewActorSubscriber(self.localPid)
	self.sub.Subscribe(message.TOPIC_SHARD_SYSTEM_EVENT)
	self.sub.Subscribe(message.TOPIC_SAVE_BLOCK_COMPLETE)

	globalState, err := self.getShardMgmtGlobalState()
	if err != nil {
		return fmt.Errorf("chainmgr: failed to read shard-mgmt global state: %s", err)
	}
	if globalState == nil {
		// not initialized from ledger
		log.Info("chainmgr: shard-mgmt not initialized, skipped loading from ledger")
		return nil
	}

	for i := uint64(1); i < globalState.NextShardID; i++ {
		shard, err := self.getShardState(i)
		if err != nil {
			return fmt.Errorf("get shard %d failed: %s", i, err)
		}
		if shard.State != shardstates.SHARD_STATE_ACTIVE {
			continue
		}
		if _, err := self.initShardInfo(i, shard); err != nil {
			return fmt.Errorf("init shard %d failed: %s", i, err)
		}
		// TODO: start shard process
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
	case *message.ShardSystemEventMsg:
		if msg == nil {
			return
		}
		evt := msg.Event
		log.Infof("chain mgr received shard system event: ver: %d, type: %d", evt.Version, evt.EventType)
		self.localEventC <- evt
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
		self.localBlockMsgC <- msg.Block

	case *shardmsg.TxRequest:
		self.txnReqC <- msg

	case *shardmsg.TxResult:
		self.txnRspC <- msg

	case *shardmsg.StorageRequest:
		log.Error("chain mgr recieved local storage request")
		self.txnReqC <- msg

	case *shardmsg.StorageResult:
		self.txnRspC <- msg

	default:
		log.Info("chain mgr actor: Unknown msg ", msg, "type", reflect.TypeOf(msg))
	}
}

func (self *ChainManager) remoteEndpointEvent(evt interface{}) {
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

func (self *ChainManager) localEventLoop() {

	self.quitWg.Add(1)
	defer self.quitWg.Done()

	for {
		select {
		case shardEvt := <-self.localEventC:
			switch shardEvt.EventType {
			case shardstates.EVENT_SHARD_CREATE:
				createEvt := &shardstates.CreateShardEvent{}
				if err := createEvt.Deserialize(bytes.NewBuffer(shardEvt.Payload)); err != nil {
					log.Errorf("deserialize create shard event: %s", err)
					continue
				}
				if err := self.onShardCreated(createEvt); err != nil {
					log.Errorf("processing create shard event: %s", err)
				}
			case shardstates.EVENT_SHARD_CONFIG_UPDATE:
				cfgEvt := &shardstates.ConfigShardEvent{}
				if err := cfgEvt.Deserialize(bytes.NewBuffer(shardEvt.Payload)); err != nil {
					log.Errorf("deserialize create shard event: %s", err)
					continue
				}
				if err := self.onShardConfigured(cfgEvt); err != nil {
					log.Errorf("processing create shard event: %s", err)
				}
			case shardstates.EVENT_SHARD_PEER_JOIN:
				jointEvt := &shardstates.PeerJoinShardEvent{}
				if err := jointEvt.Deserialize(bytes.NewBuffer(shardEvt.Payload)); err != nil {
					log.Errorf("deserialize join shard event: %s", err)
					continue
				}
				if err := self.onShardPeerJoint(jointEvt); err != nil {
					log.Errorf("processing join shard event: %s", err)
				}
			case shardstates.EVENT_SHARD_ACTIVATED:
				evt := &shardstates.ShardActiveEvent{}
				if err := evt.Deserialize(bytes.NewBuffer(shardEvt.Payload)); err != nil {
					log.Errorf("deserialize shard activation event: %s", err)
					continue
				}
				if err := self.onShardActivated(evt); err != nil {
					log.Errorf("processing shard activation event: %s", err)
				}
			case shardstates.EVENT_SHARD_PEER_LEAVE:
			case shardstates.EVENT_SHARD_GAS_DEPOSIT:
				fallthrough
			case shardstates.EVENT_SHARD_GAS_WITHDRAW_REQ:
				fallthrough
			case shardstates.EVENT_SHARD_GAS_WITHDRAW_DONE:
				if err := self.onLocalShardEvent(shardEvt); err != nil {
					log.Errorf("processing shard %d gas deposit: %s", shardEvt.ToShard, err)
				}
			}
			break
		case blk := <-self.localBlockMsgC:
			if err := self.onBlockPersistCompleted(blk); err != nil {
				log.Errorf("processing shard %d, block %d, err: %s", self.shardID, blk.Header.Height, err)
			}
		case <-self.quitC:
			return
		}
	}
}

func (self *ChainManager) broadcastMsgLoop() {
	self.quitWg.Add(1)
	defer self.quitWg.Done()

	for {
		select {
		case msg := <-self.broadcastMsgC:
			if msg.targetShardID == math.MaxUint64 {
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

func (self *ChainManager) txnLoop() {
	self.quitWg.Add(1)
	defer self.quitWg.Done()

	// TODO: make sure tx requests are responsed in all cases

	for {
		select {
		case req := <-self.txnReqC:
			childShards := self.getChildShards()
			if _, present := childShards[req.ShardID()]; !present {
				// request not for child shards
				log.Errorf("shard %d dropped request type %d to shard %d", self.shardID, req.Type(), req.ShardID())
				continue
			}

			switch req.Type() {
			case shardmsg.TXN_REQ_MSG:
				// TODO: err handling
				txreq, _ := req.(*shardmsg.TxRequest)
				if err := self.onTxnRequest(txreq); err != nil {
					log.Errorf("processing txn request: %s", err)
					if txreq != nil && txreq.TxResultCh != nil {
						close(txreq.TxResultCh)
					}
				}
			case shardmsg.STORAGE_REQ_MSG:
				storageReq, _ := req.(*shardmsg.StorageRequest)
				if err := self.onStorageRequest(storageReq); err != nil {
					log.Errorf("processing storage request: %s", err)
					if storageReq != nil && storageReq.ResultCh != nil {
						close(storageReq.ResultCh)
					}
				}
			}
		case rsp := <-self.txnRspC:
			switch rsp.Type() {
			case shardmsg.TXN_RSP_MSG:
				txrsp, _ := rsp.(*shardmsg.TxResult)
				if err := self.onTxnResponse(txrsp); err != nil {
					log.Errorf("processing Txn response: %s", err)
				}
			case shardmsg.STORAGE_RSP_MSG:
				storageRsp, _ := rsp.(*shardmsg.StorageResult)
				if err := self.onStorageResponse(storageRsp); err != nil {
					log.Errorf("processing storage response: %s", err)
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

func (self *ChainManager) processRemoteShardMsg() error {
	select {
	case remoteMsg := <-self.remoteShardMsgC:
		msg := remoteMsg.Msg
		log.Errorf(">>>>>> shard %d received remote shard msg type %d", self.shardID, msg.Type())
		switch msg.Type() {
		case shardmsg.HELLO_MSG:
			helloMsg, ok := msg.(*shardmsg.ShardHelloMsg)
			if !ok {
				return fmt.Errorf("invalid hello msg")
			}
			if helloMsg.TargetShardID != self.shardID {
				return fmt.Errorf("hello msg with invalid target %d, from %d", helloMsg.TargetShardID, helloMsg.SourceShardID)
			}
			log.Infof(">>>>>> received hello msg from %d", helloMsg.SourceShardID)
			// response ack
			return self.onNewShardConnected(remoteMsg.Sender, helloMsg)
		case shardmsg.CONFIG_MSG:
			shardCfgMsg, ok := msg.(*shardmsg.ShardConfigMsg)
			if !ok {
				return fmt.Errorf("invalid config msg")
			}
			log.Infof(">>>>>> shard %d received config msg", self.shardID)
			return self.onShardConfig(remoteMsg.Sender, shardCfgMsg)
		case shardmsg.BLOCK_REQ_MSG:
		case shardmsg.BLOCK_RSP_MSG:
			blkMsg, ok := msg.(*shardmsg.ShardBlockRspMsg)
			if !ok {
				return fmt.Errorf("invalid block rsp msg")
			}
			return self.onShardBlockReceived(remoteMsg.Sender, blkMsg)
		case shardmsg.PEERINFO_REQ_MSG:
		case shardmsg.PEERINFO_RSP_MSG:
			return nil
		case shardmsg.TXN_REQ_MSG:
			txReq, ok := msg.(*shardmsg.TxRequest)
			if !ok {
				return fmt.Errorf("invalid txn req msg")
			}
			self.onRemoteTxnRequest(remoteMsg.Sender, txReq)
		case shardmsg.TXN_RSP_MSG:
			txRsp, ok := msg.(*shardmsg.TxResult)
			if !ok {
				return fmt.Errorf("invalid txn rsp msg")
			}
			self.onRemoteTxnResponse(txRsp)
		case shardmsg.STORAGE_REQ_MSG:
			storageReq, ok := msg.(*shardmsg.StorageRequest)
			if !ok {
				return fmt.Errorf("invalid storage req msg")
			}
			self.onRemoteStorageRequest(remoteMsg.Sender, storageReq)
		case shardmsg.STORAGE_RSP_MSG:
			rsp, ok := msg.(*shardmsg.StorageResult)
			if !ok {
				return fmt.Errorf("invalid storage rsp msg")
			}
			self.onRemoteStorageResponse(rsp)
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

func (self *ChainManager) connectParent() error {

	// connect to parent
	if self.parentShardID == math.MaxUint64 {
		return nil
	}
	if self.localPid == nil {
		return fmt.Errorf("shard %d connect parent with nil localPid", self.shardID)
	}

	parentAddr := fmt.Sprintf("%s:%d", self.parentShardIPAddress, self.parentShardPort)
	parentPid := actor.NewPID(parentAddr, fmt.Sprintf("shard-%d", self.parentShardID))
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
	self.parentConnWait <- true
}

func (self *ChainManager) startListener() error {

	// start local
	props := actor.FromProducer(func() actor.Actor {
		return self
	})
	pid, err := actor.SpawnNamed(props, fmt.Sprintf("shard-%d", self.shardID))
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
		targetShardID: math.MaxUint64,
		msg:           msg,
	}
}

func (self *ChainManager) sendShardMsg(shardId uint64, msg *shardmsg.CrossShardMsg) {
	log.Infof("send shard msg type %d from %d to %d", msg.GetType(), self.shardID, shardId)
	self.broadcastMsgC <- &MsgSendReq{
		targetShardID: shardId,
		msg:           msg,
	}
}
