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
	"syscall"
	"time"

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
	log.Errorf("remote shard addr:%s disconnected", disconnMsg.Address)
	// TODO: clean pending remote-tx to disconnected shard
	if shardID, present := self.shardAddrs[disconnMsg.Address]; present {
		self.shards[shardID].Connected = false
		self.shards[shardID].Sender = nil
		err := self.restartChildShardProcess(shardID)
		if err != nil {
			log.Errorf("restart chaild shard failed shardID:%d,err:%s", shardID, err)
		}
	} else {
		if disconnMsg.Address == self.parentShardIPAddress+":"+fmt.Sprint(self.parentShardPort) {
			log.Infof("parentShard:%d has quit server", self.parentShardID)
			pid := os.Getpid()
			log.Infof("ShardId:%d,pid:%d quit server", self.shardID, pid)
			time.AfterFunc(3*time.Second, func() { syscall.Kill(pid, syscall.SIGKILL) })
		}
		log.Warnf("remote shard addr is not present:%s,parentShardID:%d,parentShardIPAddress:%s,parentShardPort:%d", disconnMsg.Address, self.parentShardID, self.parentShardIPAddress, self.parentShardPort)
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
	return self.startChildShard(evt.ShardID, shardState)
}

func (self *ChainManager) restartChildShardProcess(shardID types.ShardID) error {
	shardState, err := GetShardState(self.ledger, shardID)
	if err != nil {
		return fmt.Errorf("restartChildShard get shardmgmt state: %s", err)
	}
	return self.startChildShard(shardID, shardState)
}
func (self ChainManager) startChildShard(shardID types.ShardID, shardState *shardstates.ShardState) error {
	if _, err := self.initShardInfo(shardID, shardState); err != nil {
		return fmt.Errorf("startChildShard init shard %d info: %s", shardID, err)
	}
	if _, err := self.buildShardConfig(shardID, shardState); err != nil {
		return fmt.Errorf("startChildShard shard %d, build shard %d config: %s", self.shardID, shardID, err)
	}
	shardInfo := self.shards[shardID]
	log.Infof("startChildShard shard %d, received shard %d restart msg, parent %d", self.shardID, shardID, shardInfo.ParentShardID)
	if shardInfo == nil {
		return fmt.Errorf("startChildShard shard %d, nil shard info", shardID)
	}
	if shardInfo.ParentShardID != self.shardID {
		log.Warnf("startChildShard ParentShardID:%d,shardID:%d", shardInfo.ParentShardID, self.shardID)
		return nil
	}
	pubKey := hex.EncodeToString(keypair.SerializePublicKey(self.account.PublicKey))
	if _, has := shardState.Peers[pubKey]; !has {
		log.Warnf("startChildShard pubKey:%s is not exit shardState", pubKey)
		return nil
	}
	return self.startChildShardProcess(shardInfo)
}

func (self *ChainManager) startChildShardProcess(shardInfo *ShardInfo) error {
	// build sub-shard args
	shardportcfg := &cmdUtil.ShardPortConfig{
		ParentPort: self.parentShardPort,
		NodePort:   GetShardNodePortID(shardInfo.ShardID.ToUint64()),
		RpcPort:    GetShardRpcPortByShardID(shardInfo.ShardID.ToUint64()),
		RestPort:   GetShardRestPortByShardID(shardInfo.ShardID.ToUint64()),
	}
	shardArgs, err := cmdUtil.BuildShardCommandArgs(self.cmdArgs, shardInfo.ShardID, shardportcfg)
	if err != nil {
		return fmt.Errorf("shard %d, build shard %d command args: %s", self.shardID, shardInfo.ShardID, err)
	}
	cmd := exec.Command(os.Args[0], shardArgs...)
	if err := cmd.Start(); err != nil {
		log.Errorf("shard %d, failed to start %d: err:%s", self.shardID, shardInfo.ShardID, err)
		return fmt.Errorf("shard %d, failed to start %d: err:%s", self.shardID, shardInfo.ShardID, err)
	} else {
		log.Infof(">>>> staring shard %d, cmd: %s, args: %v", shardInfo.ShardID, os.Args[0], shardArgs)
	}
	return nil
}

func (self *ChainManager) handleBlockEvents(header *types.Header, shardEvts []*shardstates.ShardEventState) error {
	// construct one parent-block-completed message
	blkInfo := message.NewShardBlockInfo(self.shardID, header)
	shardTxs, err := constructShardBlockTx(shardEvts)
	if err != nil {
		return fmt.Errorf("shard %d, block %d, construct shard tx: %s", self.shardID, header.Height, err)
	}
	blkInfo.ShardTxs = shardTxs
	blkInfo.Events = shardEvts
	if err := self.addShardBlockInfo(blkInfo); err != nil {
		return fmt.Errorf("add shard block: %s", err)
	}

	// broadcast message to shards
	for shardID := range blkInfo.ShardTxs {
		msg, err := message.NewShardBlockRspMsg(self.shardID, header, shardTxs[shardID], self.localPid)
		if err != nil {
			return fmt.Errorf("build shard block msg: %s", err)
		}

		log.Infof("shard %d, send block %d to %d with shard tx: %v",
			self.shardID, header.Height, shardID, blkInfo.ShardTxs[shardID])

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

		msg, err := message.NewShardBlockRspMsg(self.shardID, header, shardTxs[shardID], self.localPid)
		if err != nil {
			return fmt.Errorf("build shard block msg: %s", err)
		}
		self.sendShardMsg(shardID, msg)
	}

	return nil
}

func (self *ChainManager) handleShardReqsInBlock(header *types.Header) error {

	defer func() {
		// TODO: update persisted ProcessedBlockHeight
	}()

	for height := self.processedBlockHeight + 1; height <= header.Height; height++ {
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
			self.sendCrossShardTx(tx, string(GetShardRpcPortByShardID(s.ToUint64())))
		}

		self.processedBlockHeight = height
	}

	return nil
}

func (self *ChainManager) onBlockPersistCompleted(blk *types.Block, shardEvts []*shardstates.ShardEventState) error {
	log.Infof("shard %d, get new block %d", self.shardID, blk.Header.Height)

	if err := self.handleBlockEvents(blk.Header, shardEvts); err != nil {
		log.Errorf("shard %d, handle block %d events: %s", self.shardID, blk.Header.Height, err)
	}
	if err := self.handleShardReqsInBlock(blk.Header); err != nil {
		log.Errorf("shard %d, handle shardReqs in block %d: %s", self.shardID, blk.Header.Height, err)
	}
	return nil
}

func constructShardBlockTx(evts []*shardstates.ShardEventState) (map[types.ShardID]*message.ShardBlockTx, error) {
	shardEvts := make(map[types.ShardID][]*shardstates.ShardEventState)

	// sort all ShardEvents by 'to-shard-id'
	for _, evt := range evts {
		toShard := evt.ToShard
		if _, present := shardEvts[toShard]; !present {
			shardEvts[toShard] = make([]*shardstates.ShardEventState, 0)
		}

		shardEvts[toShard] = append(shardEvts[toShard], evt)
	}

	// build one ShardTx with events to the shard
	shardTxs := make(map[types.ShardID]*message.ShardBlockTx)
	for shardId, evts := range shardEvts {
		tx, err := newShardBlockTx(evts)
		if err != nil {
			return nil, err
		}
		shardTxs[shardId] = tx
	}

	return shardTxs, nil
}

func newShardBlockTx(evts []*shardstates.ShardEventState) (*message.ShardBlockTx, error) {
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

	return &message.ShardBlockTx{Tx: tx}, nil

}
