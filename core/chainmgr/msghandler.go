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
	utils3 "github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/chainmgr/message"
	"github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/core/utils"
	"github.com/ontio/ontology/smartcontract/service/native/shard_sysmsg"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	utils2 "github.com/ontio/ontology/smartcontract/service/native/utils"
	tcomn "github.com/ontio/ontology/txnpool/common"
)

func (self *ChainManager) onNewShardConnected(sender *actor.PID, helloMsg *message.ShardHelloMsg) error {
	accPayload, err := serializeShardAccount(self.account)
	if err != nil {
		return err
	}

	shardID := helloMsg.SourceShardID
	shardState, err := self.getShardState(shardID)
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

	shardState, err := self.getShardState(evt.ShardID)
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

	return self.startUpSubShardProcess(shardInfo)
}

func (self *ChainManager) onShardActivated(evt *shardstates.ShardActiveEvent) error {
	// build shard config
	// start local shard
	shardState, err := self.getShardState(evt.ShardID)
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
	if _, has := shardState.Peers[pubKey]; has == false {
		return nil
	}

	return self.startUpSubShardProcess(shardInfo)
}

func (self *ChainManager) startUpSubShardProcess(shardInfo *ShardInfo) error {
	// build sub-shard args
	shardArgs, err := utils3.BuildShardCommandArgs(self.cmdArgs, shardInfo.ShardID, self.shardID, uint64(self.shardPort))
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

	for height := self.processedBlockHeight+1; height <= uint64(blk.Header.Height); height++ {
		shards, err := self.getRemoteMsgShards(height)
		if err != nil {
			return fmt.Errorf("get remoteMsgShards of height %d: %s", height, err)
		}
		if shards == nil || len(shards) == 0 {
			self.processedBlockHeight = height
			continue
		}

		for _, s := range shards {
			reqs, err := self.GetRemoteMsg(height, s)
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
		shardTxs[shardId] = &message.ShardBlockTx{Tx: tx}
	}

	return shardTxs, nil
}

func (self *ChainManager) onTxnRequest(txnReq *message.TxRequest) error {
	if txnReq == nil || txnReq.Tx == nil {
		return fmt.Errorf("nil Tx request")
	}
	if txnReq.Tx.ShardID == self.shardID {
		// should be processed by txnpool
		return fmt.Errorf("self Tx Request")
	}

	// check if tx is for child-shards
	childShards := self.getChildShards()
	if _, present := childShards[txnReq.Tx.ShardID]; present {
		msg, err := message.NewTxnRequestMessage(txnReq, self.localPid)
		if err != nil {
			return fmt.Errorf("failed to construct TxRequest Msg: %s", err)
		}
		self.sendShardMsg(txnReq.Tx.ShardID, msg)
		self.pendingTxns[txnReq.Tx.Hash()] = txnReq
		return nil
	}

	return fmt.Errorf("unreachable Tx request")
}

func (self *ChainManager) onTxnResponse(txnRsp *message.TxResult) error {
	if txnRsp == nil {
		return fmt.Errorf("nil txn response")
	}

	if txnReq, present := self.pendingTxns[txnRsp.Hash]; present {
		txnReq.TxResultCh <- txnRsp
		delete(self.pendingTxns, txnRsp.Hash)
		return nil
	}

	return fmt.Errorf("not found in pending tx list")
}

func (self *ChainManager) onRemoteTxnRequest(sender *actor.PID, txReq *message.TxRequest) {
	if txReq == nil || txReq.Tx == nil {
		return
	}
	if txReq.Tx.ShardID != self.shardID {
		log.Errorf("invalid remote TxRequest")
		return
	}

	// send tx to txpool
	ch := make(chan *tcomn.TxResult, 1)
	txPoolReq := &tcomn.TxReq{txReq.Tx, tcomn.ShardSender, ch}
	self.txPoolPid.Tell(txPoolReq)
	go func() {
		// FIXME: one go-routine per remote-tx ??
		if msg, ok := <-ch; ok {
			rsp := &message.TxResult{
				Err:  msg.Err,
				Hash: msg.Hash,
				Desc: msg.Desc,
			}
			// TODO: handle error
			msg, _ := message.NewTxnResponseMessage(rsp, sender)
			sender.Tell(msg)
		}
	}()
}

func (self *ChainManager) onRemoteTxnResponse(txRsp *message.TxResult) {
	if txRsp == nil {
		return
	}

	txReq, present := self.pendingTxns[txRsp.Hash]
	if !present {
		log.Errorf("invalid remote TxResponse")
		return
	}

	txReq.TxResultCh <- txRsp
}

func (self *ChainManager) onRemoteRelayTx(tx *types.Transaction) error {
	childShards := self.getChildShards()
	if _, present := childShards[tx.ShardID]; present {
		reqTxMsg := &message.TxRequest{
			Tx: tx,
		}
		msg, err := message.NewTxnRequestMessage(reqTxMsg, self.localPid)
		if err != nil {
			return fmt.Errorf("failed to build TxRequest msg: %s", err)
		}
		self.sendShardMsg(tx.ShardID, msg)
	}

	return nil
}

func (self *ChainManager) onStorageRequest(storageReq *message.StorageRequest) error {
	if storageReq.ShardId == self.shardID {
		return fmt.Errorf("self storage request")
	}

	log.Errorf("chain mgr onStorage request to shard %d", storageReq.ShardID())

	childShards := self.getChildShards()
	if _, present := childShards[storageReq.ShardID()]; present {
		msg, err := message.NewStorageRequestMessage(storageReq, self.localPid)
		if err != nil {
			return fmt.Errorf("failed to construct StorageRequest Msg: %s", err)
		}
		self.sendShardMsg(storageReq.ShardID(), msg)
		if _, present := self.pendingStorageReqs[storageReq.ShardID()]; !present {
			self.pendingStorageReqs[storageReq.ShardID()] = make(ShardStorageReqList)
		}
		self.pendingStorageReqs[storageReq.ShardID()][storageReq.Address] = storageReq
		return nil
	}

	return fmt.Errorf("unreachable storage request")
}

func (self *ChainManager) onStorageResponse(rsp *message.StorageResult) error {
	if rsp == nil {
		return fmt.Errorf("nil storage response")
	}
	reqList := self.pendingStorageReqs[rsp.ShardID]
	if reqList == nil {
		return fmt.Errorf("shard not found in pending storage reqs")
	}
	req := reqList[rsp.Address]
	if req == nil {
		return fmt.Errorf("req not found in pending storage req list")
	}
	req.ResultCh <- rsp
	delete(reqList, rsp.Address)
	return nil
}

func (self *ChainManager) onRemoteStorageRequest(sender *actor.PID, req *message.StorageRequest) {
	if req == nil {
		return
	}
	if req.ShardId != self.shardID {
		return
	}

	// get storage from local ledger
	var errStr string
	data, err := self.ledger.GetStorageItem(req.Address, req.Key)
	log.Errorf("shard %d get storage addr %v, key %v, data %v, err: %s", self.shardID, req.Address, req.Key, data, err)
	if err == common.ErrNotFound {
		err = nil
	}
	if err != nil {
		errStr = err.Error()
	}
	rsp := &message.StorageResult{
		ShardID: req.ShardId,
		Address: req.Address,
		Key:     req.Key,
		Data:    data,
		Err:     errStr,
	}
	msg, _ := message.NewStorageResponseMessage(rsp, sender)
	sender.Tell(msg)
}

func (self *ChainManager) onRemoteStorageResponse(rsp *message.StorageResult) {
	if rsp == nil {
		return
	}
	reqList := self.pendingStorageReqs[rsp.ShardID]
	if reqList == nil {
		log.Errorf("shard not found in pending storage reqs")
		return
	}
	req := reqList[rsp.Address]
	if req == nil {
		log.Errorf("req not found in pending storage req list")
		return
	}
	req.ResultCh <- rsp
}
