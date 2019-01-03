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

type ChainManager struct {
	ShardID              uint64
	ShardPort            uint
	ParentShardID        uint64
	ParentShardIPAddress string
	ParentShardPort      uint
	ChildShards          []uint64

	account *account.Account
	genesisBlock *types.Block

	p2pPid *actor.PID

	remoteShardMsgC chan *RemoteMsg

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
		ChildShards:          make([]uint64, 0),
		remoteShardMsgC:      make(chan *RemoteMsg, REMOTE_SHARDMSG_MAX_PENDING),
		quitC:                make(chan struct{}),

		account: acc,
	}

	chainMgr.startRemoteEventbus()
	if err := chainMgr.connectParent(); err != nil {
		return nil, fmt.Errorf("connect parent shard failed: %s", err)
	}
	if err := chainMgr.startListener(); err != nil {
		return nil, fmt.Errorf("shard %d start listener failed: %s", chainMgr.ShardID, err)
	}

	go chainMgr.run()
	go chainMgr.remoteShardMsgLoop()

	defaultChainManager = chainMgr
	return defaultChainManager, nil
}

func GetChainManager() *ChainManager {
	return defaultChainManager
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
		log.Infof("chain mgr received shard msg")
		shardmsg, err := shardmsg.Decode(msg.Type, msg.Data)
		if err != nil {
			log.Errorf("decode shard msg: %s", err)
			return
		}
		self.remoteShardMsgC <- &RemoteMsg{
			Sender: msg.Sender,
			Msg:    shardmsg,
		}

	default:
		log.Info("chain mgr actor: Unknown msg ", msg, "type", reflect.TypeOf(msg))
	}
}

func (self *ChainManager) run() error {

	// start listener
	// connect to parent
	// get genesis block
	// init ledger if needed
	// verify genesis block if needed

	//

	return nil
}

func (self *ChainManager) remoteShardMsgLoop() {
	for {
		if err := self.processRemoteShardMsg(); err != nil {
			log.Errorf("chain mgr process remote shard msg failed: %s", err)
		}
		select {
		case <-self.quitC:
			break
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
			// response ack
			ackMsg, err := shardmsg.NewShardHelloAckMsg()
			if err != nil {
				return fmt.Errorf("construct hello ack to %d: %s", helloMsg.SourceShardID, err)
			}
			remoteMsg.Sender.Tell(ackMsg)
			return nil
		case shardmsg.HELLO_ACK_MSG:
			helloAckMsg, ok := msg.(*shardmsg.ShardHelloAckMsg)
			if !ok {
				return fmt.Errorf("invalid hello ack msg")
			}
			acc, err := DeserializeAccount(helloAckMsg.Account)
			if err != nil {
				return fmt.Errorf("unmarshal account: %s", err)
			}
			self.account = acc
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
	parentAddr := fmt.Sprintf("%s:%d", self.ParentShardIPAddress, self.ParentShardPort)
	parentPid := actor.NewPID(parentAddr, "parentChainMgr")
	hellomsg, err := shardmsg.NewShardHelloMsg(self.ShardID, self.ParentShardID)
	if err != nil {
		return fmt.Errorf("build hello msg: %s", err)
	}
	if err := parentPid.RequestFuture(hellomsg, CONNECT_PARENT_TIMEOUT).Wait(); err != nil {
		return fmt.Errorf("connect parent chain failed: %s", err)
	}
	self.parentPid = parentPid
	return nil
}

func (self *ChainManager) startListener() error {

	// start local
	props := actor.FromProducer(func() actor.Actor {
		return self
	})
	pid, err := actor.SpawnNamed(props, "chain-manager")
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
