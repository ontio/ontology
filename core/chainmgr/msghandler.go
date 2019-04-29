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
	"math"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-eventbus/actor"
	cmdUtil "github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/chainmgr/message"
	"github.com/ontio/ontology/core/types"
	evtmsg "github.com/ontio/ontology/events/message"
	bcommon "github.com/ontio/ontology/http/base/common"
	shardsysmsg "github.com/ontio/ontology/smartcontract/service/native/shard_sysmsg"
	"github.com/ontio/ontology/smartcontract/service/native/shardgas"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	shardstates "github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
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

	self.shards[shardID].Config = cfg

	self.shardAddrs[sender.Address] = shardID

	shardSeeds := make(map[uint64]*message.SibShardInfo)
	for _, s := range self.shards {
		if s.Config == nil {
			log.Errorf("unknow config of shard: %d, %v", s.ShardID, s)
			continue
		}
		shardSeeds[s.ShardID.ToUint64()] = &message.SibShardInfo{
			SeedList: s.SeedList,
			GasPrice: s.Config.Common.GasPrice,
			GasLimit: s.Config.Common.GasLimit,
		}
	}

	buf := new(bytes.Buffer)
	if err := cfg.Serialize(buf); err != nil {
		return err
	}
	ackMsg, err := message.NewShardConfigMsg(accPayload, shardSeeds, buf.Bytes(), self.localPid)
	if err != nil {
		return fmt.Errorf("construct config to shard %d: %s", helloMsg.SourceShardID, err)
	}
	sender.Tell(ackMsg)
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
	if shardInfo.ShardID.ParentID() != self.shardID {
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

	if err := self.startChildShard(evt.ShardID, shardState); err != nil {
		return err
	}
	return nil
}

func (self *ChainManager) onWithdrawGasReq(evt *shardstates.WithdrawGasReqEvent) {
	param := &shardgas.PeerWithdrawGasParam{
		Signer:     self.account.Address,
		PeerPubKey: hex.EncodeToString(keypair.SerializePublicKey(self.account.PublicKey)),
		User:       evt.User,
		ShardId:    evt.SourceShardID,
		Amount:     evt.Amount,
		WithdrawId: evt.WithdrawId,
	}
	err := self.invokeRootNativeContract(nativeUtil.ShardGasMgmtContractAddress, shardgas.PEER_CONFIRM_WTIDHRAW_NAME,
		[]interface{}{param})
	if err != nil {
		log.Errorf("onWithdrawGasReq: failed, err: %s", err)
	}
}

func (self *ChainManager) onShardCommitDpos(evt *shardstates.ShardCommitDposEvent) {
	param := &shardgas.CommitDposParam{
		Signer:     self.account.Address,
		PeerPubKey: hex.EncodeToString(keypair.SerializePublicKey(self.account.PublicKey)),
		CommitDposParam: &shardmgmt.CommitDposParam{
			ShardID:   evt.SourceShardID,
			FeeAmount: evt.FeeAmount,
		},
	}
	err := self.invokeRootNativeContract(nativeUtil.ShardGasMgmtContractAddress, shardgas.COMMIT_DPOS_NAME,
		[]interface{}{param})
	if err != nil {
		log.Errorf("onShardCommitDpos: failed, err: %s", err)
	}
}

func (self *ChainManager) restartChildShardProcess(shardID types.ShardID) error {
	shardState, err := GetShardState(self.ledger, shardID)
	if err != nil {
		return fmt.Errorf("restartChildShard get shardmgmt state: %s", err)
	}
	return self.startChildShard(shardID, shardState)
}

func (self ChainManager) startChildShard(shardID types.ShardID, shardState *shardstates.ShardState) error {
	// TODO: start child shard if account.pubkey is in peer-list

	if _, err := self.initShardInfo(shardID, shardState); err != nil {
		return fmt.Errorf("startChildShard init shard %d info: %s", shardID, err)
	}
	shardInfo := self.shards[shardID]
	if shardInfo == nil {
		return fmt.Errorf("startChildShard shard %d, nil shard info", shardID)
	}

	if cfg, err := self.buildShardConfig(shardID, shardState); err != nil {
		return fmt.Errorf("startChildShard shard %d, build shard %d config: %s", self.shardID, shardID, err)
	} else {
		shardInfo.Config = cfg
	}
	log.Infof("startChildShard shard %d, received shard %d restart msg", self.shardID, shardID)

	return self.startChildShardProcess(shardInfo)
}

func (self *ChainManager) startChildShardProcess(shardInfo *ShardInfo) error {
	return nil
}

func (self *ChainManager) handleBlockEvents(header *types.Header, shardEvts []*evtmsg.ShardEventState) error {
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
		if shardID == self.shardID || shardID == self.shardID.ParentID() {
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
		err := self.ledger.PutShardProcessedBlockHeight(self.processedBlockHeight)
		if err != nil {
			log.Infof("save processed block height err:%v", err)
		}
	}()

	for height := self.processedBlockHeight + 1; height <= header.Height; height++ {
		shards, err := GetRequestedRemoteShards(self.ledger, height)
		if err != nil {
			return fmt.Errorf("get remoteMsgShards of height %d: %s", height, err)
		}
		log.Infof("chainmgr get remote shards: height: %d, shards: %v", height, shards)
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
			shardInfo, _ := self.shards[s]
			if shardInfo == nil {
				return fmt.Errorf("to send xshard tx to %d, no seeds", s)
			}
			if shardInfo.Config == nil || shardInfo.Config.Common == nil {
				return fmt.Errorf("to send xshard tx to %d, mal-formed shard info", s)
			}

			gasPrice := shardInfo.Config.Common.GasPrice
			gasLimit := shardInfo.Config.Common.GasLimit
			tx, err := message.NewCrossShardTxMsg(self.account, height, s, gasPrice, gasLimit, reqs)
			if err != nil {
				return fmt.Errorf("construct remoteTxMsg of height %d to shard %d: %s", height, s, err)
			}
			go func() {
				if err := self.sendCrossShardTx(tx, shardInfo.SeedList, GetShardRpcPortByShardID(s.ToUint64())); err != nil {
					log.Errorf("send xshardTx to %d, ip %v, failed: %s", s.ToUint64(), shardInfo.SeedList, err)
				}
			}()
		}

		self.processedBlockHeight = height
	}

	return nil
}

func (self *ChainManager) onBlockPersistCompleted(blk *types.Block, shardEvts []*evtmsg.ShardEventState) error {
	log.Infof("shard %d, get new block %d", self.shardID, blk.Header.Height)

	if err := self.handleBlockEvents(blk.Header, shardEvts); err != nil {
		log.Errorf("shard %d, handle block %d events: %s", self.shardID, blk.Header.Height, err)
	}
	if err := self.handleShardReqsInBlock(blk.Header); err != nil {
		log.Errorf("shard %d, handle shardReqs in block %d: %s", self.shardID, blk.Header.Height, err)
	}
	return nil
}

func constructShardBlockTx(evts []*evtmsg.ShardEventState) (map[types.ShardID]*message.ShardBlockTx, error) {
	shardEvts := make(map[types.ShardID][]*evtmsg.ShardEventState)

	// sort all ShardEvents by 'to-shard-id'
	for _, evt := range evts {
		toShard := evt.ToShard
		if _, present := shardEvts[toShard]; !present {
			shardEvts[toShard] = make([]*evtmsg.ShardEventState, 0)
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

func newShardBlockTx(evts []*evtmsg.ShardEventState) (*message.ShardBlockTx, error) {
	params := &shardsysmsg.CrossShardMsgParam{
		Events: evts,
	}
	// build transaction
	mutable, err := bcommon.NewNativeInvokeTransaction(0, math.MaxUint32, nativeUtil.ShardSysMsgContractAddress,
		byte(0), shardsysmsg.PROCESS_CROSS_SHARD_MSG, []interface{}{params})
	if err != nil {
		return nil, fmt.Errorf("newShardBlockTx: build tx failed, err: %s", err)
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		return nil, fmt.Errorf("construct shardTx: %s", err)
	}

	return &message.ShardBlockTx{Tx: tx}, nil

}

func (self *ChainManager) invokeRootNativeContract(contract common.Address, method string, args []interface{}) error {
	mutable, err := bcommon.NewNativeInvokeTransaction(0, math.MaxUint32, contract, byte(0), method, args)
	if err != nil {
		return fmt.Errorf("invokeRootNativeContract: generate tx failed, err: %s", err)
	}
	err = cmdUtil.SignTransaction(self.account, mutable)
	if err != nil {
		return fmt.Errorf("invokeRootNativeContract: sign tx failed, err: %s", err)
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		return fmt.Errorf("invokeRootNativeContract: parse tx failed, err: %s", err)
	}

	// TODO: handle send-tx failure
	// TODO: change 127.0.0.1 to seeds of root-shard
	go self.sendCrossShardTx(tx, []string{"127.0.0.1"}, GetShardRpcPortByShardID(self.shardID.ParentID().ToUint64()))
	return nil
}
