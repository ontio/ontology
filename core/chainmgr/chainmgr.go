package chainmgr

import (
	"fmt"
	"math"
	"reflect"
	"sync"
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology-eventbus/remote"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	shardmsg "github.com/ontio/ontology/core/chainmgr/message"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/account"
)

const (
	REMOTE_SHARDMSG_MAX_PENDING = 64
	CONNECT_PARENT_TIMEOUT      = 5 * time.Second
)

var defaultChainManager *ChainManager = nil

type RemoteMsg struct {
	Sender *actor.PID
	Msg    shardmsg.RemoteShardMsg
}

type ShardInfo struct {
	Config *config.OntologyConfig
}

type ChainManager struct {
	ShardID              uint64
	ShardPort            uint
	ParentShardID        uint64
	ParentShardIPAddress string
	ParentShardPort      uint

	Shards map[uint64]*ShardInfo

	account *account.Account
	genesisBlock *types.Block

	p2pPid *actor.PID

	remoteShardMsgC chan *RemoteMsg
	parentConnWait chan bool

	parentPid *actor.PID
	localPid  *actor.PID
	quitC     chan struct{}
	quitWg    sync.WaitGroup
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
		remoteShardMsgC:      make(chan *RemoteMsg, REMOTE_SHARDMSG_MAX_PENDING),
		parentConnWait:       make(chan bool),
		quitC:                make(chan struct{}),

		account: acc,
	}

	chainMgr.startRemoteEventbus()
	if err := chainMgr.startListener(); err != nil {
		return nil, fmt.Errorf("shard %d start listener failed: %s", chainMgr.ShardID, err)
	}

	go chainMgr.run()
	go chainMgr.remoteShardMsgLoop()

	if err := chainMgr.connectParent(); err != nil {
		chainMgr.Stop()
		return nil, fmt.Errorf("connect parent shard failed: %s", err)
	}

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
	case *shardmsg.CrossShardMsg:
		if msg == nil {
			return
		}
		log.Infof("chain mgr received shard msg: %v", msg)
		smsg, err := shardmsg.Decode(msg.Type, msg.Data)
		if err != nil {
			log.Errorf("decode shard msg: %s", err)
			return
		}
		self.remoteShardMsgC <- &RemoteMsg{
			Sender: msg.Sender,
			Msg:    smsg,
		}

	default:
		log.Info("chain mgr actor: Unknown msg ", msg, "type", reflect.TypeOf(msg))
	}
}

func (self *ChainManager) run() error {

	self.quitWg.Add(1)
	defer self.quitWg.Done()

	for {
		select {
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
			accPayload, err := serializeShardAccount(self.account)
			if err != nil {
				return err
			}
			configPayload, err := self.buildShardConfig(helloMsg.SourceShardID)
			if err != nil {
				return err
			}
			ackMsg, err := shardmsg.NewShardConfigMsg(accPayload, configPayload, self.localPid)
			if err != nil {
				return fmt.Errorf("construct config to shard %d: %s", helloMsg.SourceShardID, err)
			}
			remoteMsg.Sender.Tell(ackMsg)
			return nil
		case shardmsg.CONFIG_MSG:
			shardCfgMsg, ok := msg.(*shardmsg.ShardConfigMsg)
			if !ok {
				return fmt.Errorf("invalid config msg")
			}
			log.Infof(">>>>>> shard %d received config msg", self.ShardID)
			acc, err := deserializeShardAccount(shardCfgMsg.Account)
			if err != nil {
				return fmt.Errorf("unmarshal account: %s", err)
			}
			config, err := deserializeShardConfig(shardCfgMsg.Config)
			if err != nil {
				return fmt.Errorf("unmarshal shard config: %s", err)
			}
			self.account = acc
			if err := self.addShardConfig(config.Shard.ShardID, config); err != nil {
				return fmt.Errorf("add shard %d config: %s", config.Shard.ShardID, err)
			}
			self.notifyParentConnected()
			return nil
		case shardmsg.BLOCK_REQ_MSG:
		case shardmsg.BLOCK_RSP_MSG:
		case shardmsg.PEERINFO_REQ_MSG:
		case shardmsg.PEERINFO_RSP_MSG:
			return nil
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
	case <- time.After(timeout):
		return fmt.Errorf("wait parent connection timeout")
	case connected := <- self.parentConnWait:
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

func (self *ChainManager) GetShardConfig(shardID uint64) *config.OntologyConfig {
	if s := self.Shards[shardID]; s != nil {
		return s.Config
	}
	return nil
}

func (self *ChainManager) addShardConfig(shardID uint64, cfg *config.OntologyConfig) error {
	self.Shards[shardID] = &ShardInfo{
		Config: cfg,
	}
	return nil
}
