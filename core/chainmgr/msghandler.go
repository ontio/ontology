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
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-eventbus/actor"
	cmdUtil "github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/chainmgr/message"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/core/utils"
	"github.com/ontio/ontology/smartcontract/service/native/shard_sysmsg"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	nativeUtil "github.com/ontio/ontology/smartcontract/service/native/utils"
)

func (self *ChainManager) onNewShardConnected(sender *actor.PID, helloMsg *message.ShardHelloMsg) error {
	accPayload, err := serializeShardAccount(self.account)
	if err != nil {
		return err
	}

	shardID := helloMsg.SourceShardID
	shardState, err := GetShardState(self.ledger, shardID)
	if err != nil {
		return fmt.Errorf("get shardmgmt state: %s", err)
	}

	cfg, err := self.buildShardConfig(shardID, shardState)
	if err != nil {
		return err
	}

	if _, present := self.shards[shardID]; !present {
		if _, err := self.initShardInfo(shardID, shardState); err != nil {
			return fmt.Errorf("new shard connected, init: %s", err)
		}
		if self.shards[shardID] == nil {
			return nil
		}
	}

	self.shards[shardID].ShardAddress = sender.Address
	self.shards[shardID].Connected = true
	self.shards[shardID].Config = cfg
	self.shards[shardID].Sender = sender

	self.shardAddrs[sender.Address] = shardID

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

	// TODO: clean pending remote-tx to disconnected shard

	if shardID, present := self.shardAddrs[disconnMsg.Address]; present {
		self.shards[shardID].Connected = false
		self.shards[shardID].Sender = nil
	}

	return nil
}

func (self *ChainManager) onShardConfig(sender *actor.PID, shardCfgMsg *message.ShardConfigMsg) error {
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
	pubKey := hex.EncodeToString(keypair.SerializePublicKey(self.account.PublicKey))
	if evt.PeerPubKey != pubKey {
		return nil
	}

	shardState, err := GetShardState(self.ledger, evt.ShardID)
	if err != nil {
		return fmt.Errorf("get shardmgmt state: %s", err)
	}

	if shardState.State != shardstates.SHARD_STATE_ACTIVE {
		return nil
	}

	shardInfo := self.shards[evt.ShardID]
	if shardInfo == nil {
		return fmt.Errorf("shard %d, nil shard info", evt.ShardID)
	}
	if shardInfo.ParentShardID != self.shardID {
		return nil
	}

	return self.startChildShardProcess(shardInfo)
}

func (self *ChainManager) onShardActivated(evt *shardstates.ShardActiveEvent) error {
	// build shard config
	// start local shard
	shardState, err := GetShardState(self.ledger, evt.ShardID)
	if err != nil {
		return fmt.Errorf("get shardmgmt state: %s", err)
	}

	if shardState.State != shardstates.SHARD_STATE_ACTIVE {
		return fmt.Errorf("shard %d state %d is not active", evt.ShardID, shardState.State)
	}

	if _, err := self.initShardInfo(evt.ShardID, shardState); err != nil {
		return fmt.Errorf("init shard %d info: %s", evt.ShardID, err)
	}

	if _, err := self.buildShardConfig(evt.ShardID, shardState); err != nil {
		return fmt.Errorf("shard %d, build shard %d config: %s", self.shardID, evt.ShardID, err)
	}

	shardInfo := self.shards[evt.ShardID]
	log.Infof("shard %d, received shard %d activated, parent %d", self.shardID, evt.ShardID, shardInfo.ParentShardID)
	if shardInfo == nil {
		return fmt.Errorf("shard %d, nil shard info", evt.ShardID)
	}
	if shardInfo.ParentShardID != self.shardID {
		return nil
	}

	pubKey := hex.EncodeToString(keypair.SerializePublicKey(self.account.PublicKey))
	if _, has := shardState.Peers[pubKey]; !has {
		return nil
	}

	return self.startChildShardProcess(shardInfo)
}

func (self *ChainManager) startChildShardProcess(shardInfo *ShardInfo) error {
	// build sub-shard args
	shardArgs, err := cmdUtil.BuildShardCommandArgs(self.cmdArgs, shardInfo.ShardID, uint64(self.shardPort))
	if err != nil {
		return fmt.Errorf("shard %d, build shard %d command args: %s", self.shardID, shardInfo.ShardID, err)
	}

	// create new process
	cmd := exec.Command(os.Args[0], shardArgs...)
	if false {
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("shard %d, failed to start %d: %s", self.shardID, shardInfo.ShardID, err)
		}
	} else {
		log.Infof(">>>> staring shard %d, cmd: %s, args: %v", shardInfo.ShardID, os.Args[0], shardArgs)
	}

	return nil
}

func (self *ChainManager) onLocalShardEvent(evt *shardstates.ShardEventState) error {
	if evt == nil {
		return fmt.Errorf("notification with nil evt on shard %d", self.shardID)
	}
	log.Infof("shard %d, get new event type %d", self.shardID, evt.EventType)

	return self.addShardEvent(evt)
}

func (self *ChainManager) handleBlockEvents(blk *types.Block) error {
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

func (self *ChainManager) handleShardReqsInBlock(blk *types.Block) error {

	defer func() {
		// TODO: update persisted ProcessedBlockHeight
	}()

	for height := self.processedBlockHeight + 1; height <= uint64(blk.Header.Height); height++ {
		shards, err := GetRequestedRemoteShards(self.ledger, height)
		if err != nil {
			return fmt.Errorf("get remoteMsgShards of height %d: %s", height, err)
		}
		if shards == nil || len(shards) == 0 {
			self.processedBlockHeight = height
			continue
		}

		for _, s := range shards {
			reqs, err := GetRequestsToRemoteShard(self.ledger, height, s)
			if err != nil {
				return fmt.Errorf("get remoteMsg of height %d to shard %d: %s", height, s, err)
			}
			if len(reqs) == 0 {
				continue
			}
			tx, err := message.NewCrossShardTxMsg(self.account, height, s, reqs)
			if err != nil {
				return fmt.Errorf("construct remoteTxMsg of height %d to shard %d: %s", height, s, err)
			}
			self.sendCrossShardTx(s, tx)
		}

		self.processedBlockHeight = height
	}

	return nil
}

func (self *ChainManager) onBlockPersistCompleted(blk *types.Block) error {
	if blk == nil {
		return fmt.Errorf("notification with nil blk on shard %d", self.shardID)
	}
	log.Infof("shard %d, get new block %d", self.shardID, blk.Header.Height)

	if err := self.handleBlockEvents(blk); err != nil {
		log.Errorf("shard %d, handle block %d events: %s", self.shardID, blk.Header.Height, err)
	}
	if err := self.handleShardReqsInBlock(blk); err != nil {
		log.Errorf("shard %d, handle shardReqs in block %d: %s", self.shardID, blk.Header.Height, err)
	}
	return nil
}

func (self *ChainManager) constructShardBlockTx(block *message.ShardBlockInfo) (map[types.ShardID]*message.ShardBlockTx, error) {
	shardEvts := make(map[types.ShardID][]*shardstates.ShardEventState)

	// sort all ShardEvents by 'to-shard-id'
	for _, evt := range block.Events {
		toShard := evt.ToShard
		if _, present := shardEvts[toShard]; !present {
			shardEvts[toShard] = make([]*shardstates.ShardEventState, 0)
		}

		shardEvts[toShard] = append(shardEvts[toShard], evt)
	}

	// build one ShardTx with events to the shard
	shardTxs := make(map[types.ShardID]*message.ShardBlockTx)
	for shardId, evts := range shardEvts {
		params := &shardsysmsg.CrossShardMsgParam{
			Events: evts,
		}
		payload := new(bytes.Buffer)
		if err := params.Serialize(payload); err != nil {
			return nil, fmt.Errorf("construct shardTx, serialize shard sys msg: %s", err)
		}

		mutable := utils.BuildNativeTransaction(nativeUtil.ShardSysMsgContractAddress, shardsysmsg.PROCESS_CROSS_SHARD_MSG, payload.Bytes())
		tx, err := mutable.IntoImmutable()
		if err != nil {
			return nil, fmt.Errorf("construct shardTx: %s", err)
		}
		shardTxs[shardId] = &message.ShardBlockTx{Tx: tx}
	}

	return shardTxs, nil
}
