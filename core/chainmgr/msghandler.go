package chainmgr

import (
	"bytes"
	"fmt"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/chainmgr/message"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/service/native/shardgas/states"
	"github.com/ontio/ontology/core/types"
)

func (self *ChainManager) onNewShardConnected(sender *actor.PID, helloMsg *message.ShardHelloMsg) error {
	accPayload, err := serializeShardAccount(self.account)
	if err != nil {
		return err
	}
	cfg, err := self.buildShardConfig(helloMsg.SourceShardID)
	if err != nil {
		return err
	}

	self.Shards[helloMsg.SourceShardID] = &ShardInfo{
		ShardAddress: sender.Address,
		Connected:    true,
		Config:       cfg,
		Sender: sender,
	}
	self.ShardAddrs[sender.Address] = helloMsg.SourceShardID

	buf := new(bytes.Buffer)
	if err := cfg.Serialize(buf); err != nil {
		return err
	}
	ackMsg, err := message.NewShardConfigMsg(accPayload, buf.Bytes(), self.localPid)
	if err != nil {
		return fmt.Errorf("construct config to shard %d: %s", helloMsg.SourceShardID, err)
	}
	sender.Tell(ackMsg)
	return nil
}

func (self *ChainManager) onShardDisconnected(disconnMsg *message.ShardDisconnectedMsg) error {
	log.Errorf("remote shard %s disconnected", disconnMsg.Address)

	if shardID, present := self.ShardAddrs[disconnMsg.Address]; present {
		self.Shards[shardID].Connected = false
		self.Shards[shardID].Sender = nil
	}

	return nil
}

func (self *ChainManager) onShardConfigRequest(sender *actor.PID, shardCfgMsg *message.ShardConfigMsg) error {
	acc, err := deserializeShardAccount(shardCfgMsg.Account)
	if err != nil {
		return fmt.Errorf("unmarshal account: %s", err)
	}
	config, err := deserializeShardConfig(shardCfgMsg.Config)
	if err != nil {
		return fmt.Errorf("unmarshal shard config: %s", err)
	}
	self.account = acc
	if err := self.setShardConfig(config.Shard.ShardID, config); err != nil {
		return fmt.Errorf("add shard %d config: %s", config.Shard.ShardID, err)
	}
	self.notifyParentConnected()
	return nil
}

func (self *ChainManager) onShardBlockReceived(sender *actor.PID, blkRspMsg *message.ShardBlockRspMsg) error {

	log.Infof("shard %d, got block header from %d, height %d", self.ShardID, blkRspMsg.ShardID, blkRspMsg.Height)

	return nil
}

func (self *ChainManager) onShardCreated(evt *shardstates.CreateShardEvent) error {
	return nil
}

func (self *ChainManager) onShardConfigured(evt *shardstates.ConfigShardEvent) error {
	return nil
}

func (self *ChainManager) onShardPeerJoint(evt *shardstates.JoinShardEvent) error {
	return nil
}

func (self *ChainManager) onShardActivated(evt *shardstates.ShardActiveEvent) error {
	// build shard config
	// start local shard
	return nil
}

func (self *ChainManager) onShardGasDeposited(evt *shardgas_states.DepositGasEvent) error {
	return nil
}

/////////////
//
// local shard processors
//
/////////////

func (self *ChainManager) onBlockPersistCompleted(blk *types.Block) error {
	if blk == nil {
		return fmt.Errorf("notification with nil blk on shard %d", self.ShardID)
	}
	log.Infof("shard %d, get new block %d", self.ShardID, blk.Header.Height)

	// construct one parent-block-completed message
	blockInfo, err := message.NewShardBlockInfo(self.ShardID, blk)
	if err != nil {
		return fmt.Errorf("init shard block info: %s", err)
	}
	if err := self.addShardBlockInfo(blockInfo); err != nil {
		return fmt.Errorf("add shard block: %s", err)
	}

	msg, err := message.NewShardBlockRspMsg(self.ShardID, blockInfo.BlockHeight, blk.Header, self.localPid)
	if err != nil {
		return fmt.Errorf("build shard block msg: %s", err)
	}

	// send msg to child shards
	return self.broadcastShardMsg(msg)
}
