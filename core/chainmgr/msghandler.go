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
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/chainmgr/message"
	"github.com/ontio/ontology/core/chainmgr/xshard"
	"github.com/ontio/ontology/core/ledger"
	com "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/types"
	shardstates "github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
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

	lgr := ledger.GetShardLedger(evt.ShardID)
	if lgr == nil {
		return fmt.Errorf("failed to get ledger of shard %d", evt.ShardID)
	}

	shardState, err := xshard.GetShardState(lgr, evt.ShardID)
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
	lgr := ledger.GetShardLedger(evt.ShardID.ParentID())
	if lgr == nil {
		return fmt.Errorf("failed to get ledger of shard %d", evt.ShardID)
	}
	shardState, err := xshard.GetShardState(lgr, evt.ShardID)
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

func (self ChainManager) startChildShard(shardID common.ShardID, shardState *shardstates.ShardState) error {
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

	if err := self.initShardLedger(shardInfo); err != nil {
		return fmt.Errorf("init shard %d, failed to init ledger: %s", self.shardID, err)
	}
	// set Default Ledger
	if lgr := ledger.GetShardLedger(shardID); lgr != nil {
		ledger.DefLedger = lgr
	}

	if err := self.initShardTxPool(); err != nil {
		return fmt.Errorf("init initShardTxPool %d, failed to init initShardTxPool: %s", self.shardID, err)
	}
	self.startConsensus()
	return nil
}

func (self *ChainManager) onBlockPersistCompleted(blk *types.Block) {
	if self.shardID.IsRootShard() {
		// main-chain has no parent-chain, and not support xshard-txn
		return
	}
	log.Infof("chainmgr shard %d, get new block %d from shard %d", self.shardID, blk.Header.Height, blk.Header.ShardID)

	if err := self.addShardBlock(blk); err != nil {
		log.Errorf("shard %d, handle block %d events: %s", self.shardID, blk.Header.Height, err)
	}
	if err := self.handleShardReqsInBlock(blk.Header); err != nil {
		log.Errorf("shard %d, handle shardReqs in block %d: %s", self.shardID, blk.Header.Height, err)
	}
	if err := self.handleRootChainBlock(); err != nil {
		log.Errorf("shard %d, handle rootchain block in block %d: %s", self.shardID, blk.Header.Height, err)
	}
}

func (self *ChainManager) addShardBlock(block *types.Block) error {
	// construct one parent-block-completed message
	blkInfo := message.NewShardBlockInfo(self.shardID, block)
	if err := self.addShardBlockInfo(blkInfo); err != nil {
		return fmt.Errorf("add shard block: %s", err)
	}

	return nil
}

func (self *ChainManager) handleShardReqsInBlock(header *types.Header) error {
	shardID, err := common.NewShardID(header.ShardID)
	if err != nil {
		return fmt.Errorf("invalid shard id %d", header.ShardID)
	}
	lgr := ledger.GetShardLedger(shardID)
	if lgr == nil {
		return fmt.Errorf("failed to get ledger of shard %d", header.ShardID)
	}
	height := header.Height
	shards, err := lgr.GetRelatedShardIDsInBlock(height)
	if err != nil {
		return fmt.Errorf("get remoteMsgShards of height %d: %s", height, err)
	}
	if shards == nil || len(shards) == 0 {
		return nil
	}

	log.Infof("[chainmgr]handleShardReqsInBlock: height: %d, shards: %v", height, shards)
	for _, s := range shards {
		// don't handle other shard req when block is from parent
		if self.shardID.ParentID().ToUint64() == header.ShardID && s.ToUint64() != self.shardID.ToUint64() {
			continue
		}
		reqs, err := lgr.GetShardMsgsInBlock(height, s)
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

	return nil
}

func (self *ChainManager) handleRootChainBlock() error {
	shardState, err := xshard.GetShardState(self.mainLedger, self.shardID)
	if err == com.ErrNotFound {
		log.Debugf("get shard %d failed: %s", self.shardID, err)
		return nil
	}
	if err != nil {
		return fmt.Errorf("get shard %d failed: %s", self.shardID, err)
	}
	if shardState.State != shardstates.SHARD_STATE_ACTIVE {
		return nil
	}
	if cfg, err := self.buildShardConfig(self.shardID, shardState); err != nil {
		return fmt.Errorf("startChildShard shard %d,config: %s", self.shardID, err)
	} else {
		if err := self.setShardConfig(self.shardID, cfg); err != nil {
			return fmt.Errorf("add shard %d config: %s", self.shardID, err)
		}
	}
	return nil
}
