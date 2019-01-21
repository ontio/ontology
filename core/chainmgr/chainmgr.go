package chainmgr

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"sync"
	"time"

	"encoding/hex"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology-eventbus/eventstream"
	"github.com/ontio/ontology-eventbus/remote"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	shardmsg "github.com/ontio/ontology/core/chainmgr/message"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/events"
	"github.com/ontio/ontology/events/message"
	"github.com/ontio/ontology/smartcontract/service/native/shardgas/states"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
)

const (
	LOCAL_SHARDMSG_MAX_PENDING  = 64
	REMOTE_SHARDMSG_MAX_PENDING = 64
	CONNECT_PARENT_TIMEOUT      = 5 * time.Second
)

var defaultChainManager *ChainManager = nil

type RemoteMsg struct {
	Sender *actor.PID
	Msg    shardmsg.RemoteShardMsg
}

type ShardInfo struct {
	ShardAddress string
	Connected    bool
	Config       *config.OntologyConfig
	Sender *actor.PID
}

type ChainManager struct {
	ShardID              uint64
	ShardPort            uint
	ParentShardID        uint64
	ParentShardIPAddress string
	ParentShardPort      uint

	Lock        sync.RWMutex
	Shards      map[uint64]*ShardInfo
	ShardAddrs  map[string]uint64
	ShardBlocks map[uint64]shardmsg.ShardBlockMap // indexed by shardID

	account      *account.Account
	genesisBlock *types.Block

	ledger *ledger.Ledger
	p2pPid *actor.PID

	localBlockMsgC  chan *types.Block
	localShardMsgC  chan *shardstates.ShardEventState
	remoteShardMsgC chan *RemoteMsg
	broadcastMsgC   chan *shardmsg.CrossShardMsg
	parentConnWait  chan bool

	parentPid   *actor.PID
	localPid    *actor.PID
	sub         *events.ActorSubscriber
	endpointSub *eventstream.Subscription

	quitC  chan struct{}
	quitWg sync.WaitGroup
}

func Initialize(shardID, parentShardID uint64, parentAddr string, shardPort, parentPort uint, acc *account.Account) (*ChainManager, error) {
	// fixme: change to sync.once
	if defaultChainManager != nil {
		return nil, fmt.Errorf("chain manager had been initialized for shard: %d", defaultChainManager.ShardID)
	}
	chainMgr := &ChainManager{
		ShardID:              shardID,
		ShardPort:            shardPort,
		ParentShardID:        parentShardID,
		ParentShardIPAddress: parentAddr,
		ParentShardPort:      parentPort,
		Shards:               make(map[uint64]*ShardInfo),
		ShardAddrs:           make(map[string]uint64),
		ShardBlocks:          make(map[uint64]shardmsg.ShardBlockMap),
		localBlockMsgC:       make(chan *types.Block, LOCAL_SHARDMSG_MAX_PENDING),
		localShardMsgC:       make(chan *shardstates.ShardEventState, LOCAL_SHARDMSG_MAX_PENDING),
		remoteShardMsgC:      make(chan *RemoteMsg, REMOTE_SHARDMSG_MAX_PENDING),
		broadcastMsgC:        make(chan *shardmsg.CrossShardMsg, REMOTE_SHARDMSG_MAX_PENDING),
		parentConnWait:       make(chan bool),
		quitC:                make(chan struct{}),

		account: acc,
	}

	chainMgr.startRemoteEventbus()
	if err := chainMgr.startListener(); err != nil {
		return nil, fmt.Errorf("shard %d start listener failed: %s", chainMgr.ShardID, err)
	}

	go chainMgr.localShardMsgLoop()
	go chainMgr.remoteShardMsgLoop()
	go chainMgr.broadcastMsgLoop()

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

func GetChainManager() *ChainManager {
	return defaultChainManager
}

func (self *ChainManager) GetAccount() *account.Account {
	return self.account
}

func (self *ChainManager) SetP2P(p2p *actor.PID) error {
	if defaultChainManager == nil {
		return fmt.Errorf("uninitialized chain manager")
	}

	defaultChainManager.p2pPid = p2p
	return nil
}

func (self *ChainManager) LoadFromLedger(lgr *ledger.Ledger) error {
	// TODO: get all shards from local ledger

	self.ledger = lgr

	// start listen on local actor
	self.sub = events.NewActorSubscriber(self.localPid)
	self.sub.Subscribe(message.TOPIC_SHARD_SYSTEM_EVENT)
	self.sub.Subscribe(message.TOPIC_SAVE_BLOCK_COMPLETE)

	globalState, err := self.getShardMgmtGlobalState()
	if err != nil {

	}
	if globalState == nil {
		// not initialized from ledger
		log.Info("chainmgr: shard-mgmt not initialized, skipped loading from ledger")
		return nil
	}

	peerPK := hex.EncodeToString(keypair.SerializePublicKey(self.account.PublicKey))

	for i := uint64(1); i < globalState.NextShardID; i++ {
		shard, err := self.getShardState(i)
		if err != nil {
			log.Errorf("get shard %d failed: %s", i, err)
		}
		if shard.State != shardstates.SHARD_STATE_ACTIVE {
			continue
		}
		if _, present := shard.Peers[peerPK]; present {
			// peer is in the shard
			// build shard config
		}
	}

	return nil
}

func (self *ChainManager) startRemoteEventbus() {
	localRemote := fmt.Sprintf("%s:%d", config.DEFAULT_PARENTSHARD_IPADDR, self.ShardPort)
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
		self.localShardMsgC <- evt
	case *shardmsg.CrossShardMsg:
		if msg == nil {
			return
		}
		log.Tracef("chain mgr received shard msg: %v", msg)
		smsg, err := shardmsg.Decode(msg.Type, msg.Data)
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

func (self *ChainManager) localShardMsgLoop() error {

	self.quitWg.Add(1)
	defer self.quitWg.Done()

	for {
		select {
		case shardEvt := <-self.localShardMsgC:
			switch shardEvt.EventType {
			case shardstates.EVENT_SHARD_CREATE:
				createEvt := &shardstates.CreateShardEvent{}
				if err := createEvt.Deserialize(bytes.NewBuffer(shardEvt.Info)); err != nil {
					log.Errorf("deserialize create shard event: %s", err)
				}
				if err := self.onShardCreated(createEvt); err != nil {
					log.Errorf("processing create shard event: %s", err)
				}
			case shardstates.EVENT_SHARD_CONFIG_UPDATE:
				cfgEvt := &shardstates.ConfigShardEvent{}
				if err := cfgEvt.Deserialize(bytes.NewBuffer(shardEvt.Info)); err != nil {
					log.Errorf("deserialize create shard event: %s", err)
				}
				if err := self.onShardConfigured(cfgEvt); err != nil {
					log.Errorf("processing create shard event: %s", err)
				}
			case shardstates.EVENT_SHARD_PEER_JOIN:
				jointEvt := &shardstates.JoinShardEvent{}
				if err := jointEvt.Deserialize(bytes.NewBuffer(shardEvt.Info)); err != nil {
					log.Errorf("deserialize join shard event: %s", err)
				}
				if err := self.onShardPeerJoint(jointEvt); err != nil {
					log.Errorf("processing join shard event: %s", err)
				}
			case shardstates.EVENT_SHARD_ACTIVATED:
				evt := &shardstates.ShardActiveEvent{}
				if err := evt.Deserialize(bytes.NewBuffer(shardEvt.Info)); err != nil {
					log.Errorf("deserialize join shard event: %s", err)
				}
				if err := self.onShardActivated(evt); err != nil {
					log.Errorf("processing join shard event: %s", err)
				}
			case shardstates.EVENT_SHARD_PEER_LEAVE:
			case shardgas_states.EVENT_SHARD_GAS_DEPOSIT:
				evt := &shardgas_states.DepositGasEvent{}
				if err := evt.Deserialize(bytes.NewBuffer(shardEvt.Info)); err != nil {
					log.Errorf("deserialize shard gas deposit event: %s", err)
				}
				if err := self.onShardGasDeposited(evt); err != nil {
					log.Errorf("processing shard %d gas deposit: %s", evt.ShardID, err)
				}
			case shardgas_states.EVENT_SHARD_GAS_WITHDRAW:
			}
			break
		case blk := <-self.localBlockMsgC:
			if err := self.onBlockPersistCompleted(blk); err != nil {
				log.Errorf("processing shard %d, block %d, err: %s", self.ShardID, blk.Header.Height, err)
			}
		case <-self.quitC:
			return nil
		}
	}
	// get genesis block
	// init ledger if needed
	// verify genesis block if needed

	//

	return nil
}

func (self *ChainManager) broadcastMsgLoop() {
	self.quitWg.Add(1)
	defer self.quitWg.Done()

	for {
		select {
		case msg := <-self.broadcastMsgC:
			for _, s := range self.Shards {
				if s.Connected && s.Sender != nil {
					s.Sender.Tell(msg)
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
		switch msg.Type() {
		case shardmsg.HELLO_MSG:
			helloMsg, ok := msg.(*shardmsg.ShardHelloMsg)
			if !ok {
				return fmt.Errorf("invalid hello msg")
			}
			if helloMsg.TargetShardID != self.ShardID {
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
			log.Infof(">>>>>> shard %d received config msg", self.ShardID)
			return self.onShardConfigRequest(remoteMsg.Sender, shardCfgMsg)
		case shardmsg.BLOCK_REQ_MSG:
		case shardmsg.BLOCK_RSP_MSG:
			blkMsg, ok := msg.(*shardmsg.ShardBlockRspMsg)
			if !ok {
				return fmt.Errorf("invalid block rsp msg")
			}
			log.Info(">>>> shard %d received block info from %d", self.ShardID, blkMsg.ShardID)
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

func (self *ChainManager) connectParent() error {

	// connect to parent
	if self.ParentShardID == math.MaxUint64 {
		return nil
	}
	if self.localPid == nil {
		return fmt.Errorf("shard %d connect parent with nil localPid", self.ShardID)
	}

	parentAddr := fmt.Sprintf("%s:%d", self.ParentShardIPAddress, self.ParentShardPort)
	parentPid := actor.NewPID(parentAddr, fmt.Sprintf("shard-%d", self.ParentShardID))
	hellomsg, err := shardmsg.NewShardHelloMsg(self.ShardID, self.ParentShardID, self.localPid)
	if err != nil {
		return fmt.Errorf("build hello msg: %s", err)
	}
	parentPid.Request(hellomsg, self.localPid)
	if err := self.waitConnectParent(CONNECT_PARENT_TIMEOUT); err != nil {
		return fmt.Errorf("wait connection with parent err: %s", err)
	}

	self.parentPid = parentPid
	log.Infof("shard %d connected with parent shard %d", self.ShardID, self.ParentShardID)
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
	pid, err := actor.SpawnNamed(props, fmt.Sprintf("shard-%d", self.ShardID))
	if err != nil {
		return fmt.Errorf("init chain manager actor: %s", err)
	}
	self.localPid = pid

	log.Infof("chain %d started listen on port %d", self.ShardID, self.ShardPort)
	return nil
}

func (self *ChainManager) Stop() {
	close(self.quitC)
	self.quitWg.Wait()
}

func (self *ChainManager) broadcastShardMsg(msg *shardmsg.CrossShardMsg) error {
	self.broadcastMsgC <- msg
	return nil
}
