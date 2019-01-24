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

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/chainmgr/message"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/core/utils"
	utils2 "github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/service/native/shard_sysmsg"
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

	self.shards[helloMsg.SourceShardID] = &ShardInfo{
		ShardAddress: sender.Address,
		Connected:    true,
		Config:       cfg,
		Sender:       sender,
	}
	self.shardAddrs[sender.Address] = helloMsg.SourceShardID

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

	if shardID, present := self.shardAddrs[disconnMsg.Address]; present {
		self.shards[shardID].Connected = false
		self.shards[shardID].Sender = nil
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

func (self *ChainManager) onShardBlockReceived(sender *actor.PID, blkMsg *message.ShardBlockRspMsg) error {
	blkInfo, err := message.NewShardBlockInfoFromRemote(self.shardID, blkMsg)
	if err != nil {
		return fmt.Errorf("construct shard blockInfo for %d: %s", blkMsg.FromShardID, err)
	}

	log.Infof("shard %d, got block header from %d, height: %d, tx %v",
		self.shardID, blkMsg.FromShardID, blkMsg.BlockHeader.Header.Height, blkInfo.ShardTxs)

	return self.addShardBlockInfo(blkInfo)
}

/////////////
//
// local shard processors
//
/////////////

func (self *ChainManager) onShardCreated(evt *shardstates.CreateShardEvent) error {
	return nil
}

func (self *ChainManager) onShardConfigured(evt *shardstates.ConfigShardEvent) error {
	return nil
}

func (self *ChainManager) onShardPeerJoint(evt *shardstates.PeerJoinShardEvent) error {
	// if current node is the peer, and shard is active, start the shard
	return nil
}

func (self *ChainManager) onShardActivated(evt *shardstates.ShardActiveEvent) error {
	// build shard config
	// start local shard
	return nil
}

func (self *ChainManager) onLocalShardEvent(evt *shardstates.ShardEventState) error {
	if evt == nil {
		return fmt.Errorf("notification with nil evt on shard %d", self.shardID)
	}
	log.Infof("shard %d, get new event type %d", self.shardID, evt.EventType)

	return self.addShardEvent(evt)
}

func (self *ChainManager) onBlockPersistCompleted(blk *types.Block) error {
	if blk == nil {
		return fmt.Errorf("notification with nil blk on shard %d", self.shardID)
	}
	log.Infof("shard %d, get new block %d", self.shardID, blk.Header.Height)

	// construct one parent-block-completed message
	blkInfo := self.getShardBlockInfo(self.shardID, uint64(blk.Header.Height))
	if blkInfo == nil {
		newBlkInfo, err := message.NewShardBlockInfo(self.shardID, blk)
		if err != nil {
			return fmt.Errorf("init shard block info: %s", err)
		}
		if err := self.addShardBlockInfo(newBlkInfo); err != nil {
			return fmt.Errorf("add shard block: %s", err)
		}
		blkInfo = newBlkInfo
	} else {
		shardTxs, err := self.constructShardBlockTx(blkInfo)
		if err != nil {
			return fmt.Errorf("shard %d, block %d, construct shard tx: %s", self.shardID, blkInfo.Height, err)
		}

		log.Infof("shard %d, block %d with shard tx: %v", self.shardID, blk.Header.Height, shardTxs)
		self.updateShardBlockInfo(self.shardID, uint64(blk.Header.Height), blk, shardTxs)
	}

	// broadcast message to shards
	for shardID := range blkInfo.ShardTxs {
		msg, err := message.NewShardBlockRspMsg(self.shardID, shardID, blkInfo, self.localPid)
		if err != nil {
			return fmt.Errorf("build shard block msg: %s", err)
		}

		log.Infof("shard %d, send block %d to %d with shard tx: %v",
			self.shardID, blk.Header.Height, shardID, blkInfo.ShardTxs[shardID])

		// send msg to shard
		self.sendShardMsg(shardID, msg)
	}

	// broadcast to all other child shards
	for shardID := range self.shards {
		if shardID == self.shardID || shardID == self.parentShardID {
			continue
		}
		if _, present := blkInfo.ShardTxs[shardID]; present {
			continue
		}

		msg, err := message.NewShardBlockRspMsg(self.shardID, shardID, blkInfo, self.localPid)
		if err != nil {
			return fmt.Errorf("build shard block msg: %s", err)
		}
		self.sendShardMsg(shardID, msg)
	}

	return nil
}

func (self *ChainManager) constructShardBlockTx(block *message.ShardBlockInfo) (map[uint64]*message.ShardBlockTx, error) {
	shardEvts := make(map[uint64][]*shardstates.ShardEventState)

	// sort all ShardEvents by 'to-shard-id'
	for _, evt := range block.Events {
		toShard := evt.ToShard
		if _, present := shardEvts[toShard]; !present {
			shardEvts[toShard] = make([]*shardstates.ShardEventState, 0)
		}

		shardEvts[toShard] = append(shardEvts[toShard], evt)
	}

	// build one ShardTx with events to the shard
	shardTxs := make(map[uint64]*message.ShardBlockTx)
	for shardId, evts := range shardEvts {
		params := &shardsysmsg.CrossShardMsgParam{
			Events: evts,
		}
		payload := new(bytes.Buffer)
		if err := params.Serialize(payload); err != nil {
			return nil, fmt.Errorf("construct shardTx, serialize shard sys msg: %s", err)
		}

		mutable := utils.BuildNativeTransaction(utils2.ShardSysMsgContractAddress, shardsysmsg.PROCESS_CROSS_SHARD_MSG, payload.Bytes())
		tx, err := mutable.IntoImmutable()
		if err != nil {
			return nil, fmt.Errorf("construct shardTx: %s", err)
		}
		shardTxs[shardId] = &message.ShardBlockTx{ Tx: tx }
	}

	return shardTxs, nil
}
