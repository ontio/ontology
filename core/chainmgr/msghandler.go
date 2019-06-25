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
	"encoding/json"
	"fmt"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/chainmgr/xshard"
	"github.com/ontio/ontology/core/ledger"
	com "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/p2pserver/actor/server"
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
	if evt.ShardID.ParentID() == self.shardID {
		return nil
	}
	self.AddShardEventConfig(evt.Height, evt.ShardID, evt.Config, evt.Peers)
	return self.updateShardConfig(evt.ShardID, evt.Config)
}

func (self *ChainManager) onShardPeerJoint(evt *shardstates.PeerJoinShardEvent) error {
	if self.account == nil {
		// peer is in sync mode, skip peer joint event
		return nil
	}

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
	self.AddShardEventConfig(0, evt.ShardID, shardState.Config, shardState.Peers)
	if evt.ShardID != self.shardID || self.shardID.IsRootShard() {
		log.Infof("self shardID equal evt shardID or is rootshard:%v", evt.ShardID)
		return nil
	}
	if err := self.startChildShard(evt.ShardID, shardState); err != nil {
		return err
	}
	return nil
}

func (self ChainManager) startChildShard(shardID common.ShardID, shardState *shardstates.ShardState) error {
	// TODO: start consensus / syncer / http / txpool
	shardInfo := self.initShardInfo(shardState)
	if cfg, err := self.buildShardConfig(shardID, shardState); err != nil {
		return fmt.Errorf("startChildShard shard %d, build shard %d config: %s", self.shardID, shardID, err)
	} else {
		shardInfo.Config = cfg
	}
	config.DefConfig = shardInfo.Config
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

	self.p2pPid.Tell(&server.StartSync{
		ShardID:    shardID.ToUint64(),
		ShardSeeds: shardInfo.SeedList,
	})
	log.Infof("chainmgr starting shard-sync %d", shardID)

	return self.startConsensus()
}

func (self *ChainManager) onBlockPersistCompleted(blk *types.Block) {
	if self.shardID.IsRootShard() {
		// main-chain has no parent-chain, and not support xshard-txn
		return
	}
	log.Infof("chainmgr shard %d, get new block %d from shard %d", self.shardID, blk.Header.Height, blk.Header.ShardID)
	if err := self.handleRootChainConfig(blk); err != nil {
		log.Errorf("shard %d, handle rootchain chainConfig block in block %d: %s", self.shardID, blk.Header.Height, err)
	}
	if err := self.handleRootChainBlock(); err != nil {
		log.Errorf("shard %d, handle rootchain block in block %d: %s", self.shardID, blk.Header.Height, err)
	}
}
func (self *ChainManager) handleRootChainConfig(block *types.Block) error {
	if config.DefConfig.Genesis.ConsensusType == config.CONSENSUS_TYPE_SOLO {
		return nil
	}
	blkInfo := &vconfig.VbftBlockInfo{}
	if err := json.Unmarshal(block.Header.ConsensusPayload, blkInfo); err != nil {
		return fmt.Errorf("unmarshal blockInfo: %s", err)
	}
	if blkInfo.LastConfigBlockNum != block.Header.Height {
		return nil
	}
	config := &shardstates.ShardConfig{
		VbftCfg: &config.VBFTConfig{
			N: blkInfo.NewChainConfig.N,
			C: blkInfo.NewChainConfig.C,
		},
	}
	peers := make(map[string]*shardstates.PeerShardStakeInfo)
	for _, peer := range blkInfo.NewChainConfig.Peers {
		peers[peer.ID] = &shardstates.PeerShardStakeInfo{
			Index:      peer.Index,
			PeerPubKey: peer.ID,
			NodeType:   shardstates.CONSENSUS_NODE,
		}
	}
	self.AddShardEventConfig(block.Header.Height, common.NewShardIDUnchecked(block.Header.ShardID), config, peers)
	return nil
}

func (self *ChainManager) handleRootChainBlock() error {
	shardState, err := xshard.GetShardState(ledger.GetShardLedger(common.NewShardIDUnchecked(config.DEFAULT_SHARD_ID)), self.shardID)
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

func (self *ChainManager) AddShardEventConfig(height uint32, shardID common.ShardID, cfg *shardstates.ShardConfig, peers map[string]*shardstates.PeerShardStakeInfo) {
	shardEvent := &shardstates.ConfigShardEvent{
		Height: height,
		Config: cfg,
		Peers:  peers,
	}
	sink := common.ZeroCopySink{}
	shardEvent.Serialization(&sink)
	lgr := ledger.GetShardLedger(self.shardID)
	if lgr == nil {
		lgr = ledger.GetShardLedger(common.NewShardIDUnchecked(config.DEFAULT_SHARD_ID))
	}

	err := lgr.AddShardConsensusConfig(shardID, height, sink.Bytes())
	if err != nil {
		log.Errorf("AddShardConsensusConfig err:%s", err)
		return
	}

	heights, err := lgr.GetShardConsensusHeight(shardID)
	if err != nil {
		if err != com.ErrNotFound {
			log.Errorf("GetShardConsensusHeight shardID:%v, err:%s", shardID, err)
			return
		}
	}
	heights_db := make([]uint32, 0)
	heights_db = append(heights_db, heights...)
	heights_db = append(heights_db, height)
	err = lgr.AddShardConsensusHeight(shardID, heights_db)
	if err != nil {
		log.Errorf("AddShardConsensusHeight err:%s", err)
		return
	}
}
