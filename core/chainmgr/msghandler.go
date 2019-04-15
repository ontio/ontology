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
	"encoding/hex"
	"fmt"
	"math"

	"github.com/ontio/ontology-crypto/keypair"
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

	lgr := GetShardLedger(evt.ShardID)
	if lgr == nil {
		return fmt.Errorf("failed to get ledger of shard %d", evt.ShardID)
	}

	shardState, err := GetShardState(lgr, evt.ShardID)
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

	return nil
}

func (self *ChainManager) onShardActivated(evt *shardstates.ShardActiveEvent) error {
	// build shard config
	// start local shard
	lgr := GetShardLedger(evt.ShardID)
	if lgr == nil {
		return fmt.Errorf("failed to get ledger of shard %d", evt.ShardID)
	}
	shardState, err := GetShardState(lgr, evt.ShardID)
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

func (self ChainManager) startChildShard(shardID types.ShardID, shardState *shardstates.ShardState) error {
	// TODO: start consensus / syncer / http / txpool

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

	return nil
}

func (self *ChainManager) handleBlockEvents(block *types.Block, shardEvts []*evtmsg.ShardEventState) error {
	// construct one parent-block-completed message
	header := block.Header
	blkInfo := message.NewShardBlockInfo(self.shardID, block)
	shardTxs, err := constructShardBlockTx(shardEvts)
	if err != nil {
		return fmt.Errorf("shard %d, block %d, construct shard tx: %s", self.shardID, header.Height, err)
	}
	blkInfo.ShardTxs = shardTxs
	blkInfo.Events = shardEvts
	if err := self.addShardBlockInfo(blkInfo); err != nil {
		return fmt.Errorf("add shard block: %s", err)
	}

	return nil
}

func (self *ChainManager) handleShardReqsInBlock(header *types.Header) error {
	shardID, err := types.NewShardID(header.ShardID)
	if err != nil {
		return fmt.Errorf("invalid shard id %d", header.ShardID)
	}
	lgr := GetShardLedger(shardID)
	if lgr == nil {
		return fmt.Errorf("failed to get ledger of shard %d", header.ShardID)
	}

	for height := self.processedParentBlockHeight + 1; height <= header.Height; height++ {
		shards, err := GetRequestedRemoteShards(lgr, height)
		if err != nil {
			return fmt.Errorf("get remoteMsgShards of height %d: %s", height, err)
		}
		log.Infof("chainmgr get remote shards: height: %d, shards: %v", height, shards)
		if shards == nil || len(shards) == 0 {
			self.processedParentBlockHeight = height
			continue
		}

		for _, s := range shards {
			reqs, err := GetRequestsToRemoteShard(lgr, height, s)
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
				if err := self.sendCrossShardTx(tx, shardInfo.SeedList, self.getShardRPCPort(s)); err != nil {
					log.Errorf("send xshardTx to %d, ip %v, failed: %s", s.ToUint64(), shardInfo.SeedList, err)
				}
			}()
		}

		self.processedParentBlockHeight = height
	}

	return nil
}

func (self *ChainManager) onBlockPersistCompleted(blk *types.Block, shardEvts []*evtmsg.ShardEventState) error {
	log.Infof("shard %d, get new block %d", self.shardID, blk.Header.Height)

	if err := self.handleBlockEvents(blk, shardEvts); err != nil {
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
	go self.sendCrossShardTx(tx, []string{"127.0.0.1"}, self.getShardRPCPort(self.shardID.ParentID()))
	return nil
}
