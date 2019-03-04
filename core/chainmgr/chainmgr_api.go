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
	"fmt"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/types"
)

const shard_port_gap = 10

func GetShardName(shardID types.ShardID) string {
	return fmt.Sprintf("shard_%d", shardID.ToUint64())
}

func GetChainManager() *ChainManager {
	return defaultChainManager
}

func GetAccount() *account.Account {
	chainmgr := GetChainManager()
	return chainmgr.account
}

func GetShardID() types.ShardID {
	return GetChainManager().shardID
}

func GetParentShardID() types.ShardID {
	chainmgr := GetChainManager()
	return chainmgr.parentShardID
}

func GetChildShards() []types.ShardID {
	childShards := make([]types.ShardID, 0)
	chainmgr := GetChainManager()
	for id := range chainmgr.getChildShards() {
		childShards = append(childShards, id)
	}
	return childShards
}

func GetPID() *actor.PID {
	return GetChainManager().localPid
}

func GetShardRestPort() uint {
	return config.DefConfig.Restful.HttpRestPort + uint(GetShardID().ToUint64())*shard_port_gap
}

func GetShardRpcPort() uint {
	return config.DefConfig.Rpc.HttpJsonPort + uint(GetShardID().ToUint64())*shard_port_gap
}

func GetShardRpcPortByShardID(shardId uint64) uint {
	return config.DefConfig.Rpc.HttpJsonPort + uint(shardId)*shard_port_gap
}

func SetP2P(p2p *actor.PID) error {
	if defaultChainManager == nil {
		return fmt.Errorf("uninitialized chain manager")
	}

	defaultChainManager.p2pPid = p2p
	return nil
}

func SetTxPool(txPool *actor.PID) error {
	if defaultChainManager == nil {
		return fmt.Errorf("uninitialized chain manager")
	}
	defaultChainManager.txPoolPid = txPool
	return nil
}

//
// GetParentShardHeight
// get height of parent shard
//
func GetParentShardHeight() (uint32, error) {
	chainmgr := GetChainManager()
	chainmgr.lock.RLock()
	defer chainmgr.lock.RUnlock()

	if chainmgr.shardID.IsRootShard() {
		return 0, nil
	}

	m := chainmgr.blockPool.Shards[chainmgr.parentShardID]
	if m == nil {
		return 0, fmt.Errorf("no parent block info found in blockPool")
	}

	h := uint32(0)
	if cfg := chainmgr.GetShardConfig(chainmgr.shardID); cfg != nil {
		h = cfg.Shard.GenesisParentHeight
	} else {
		log.Errorf("failed to get self shard config")
	}
	for _, blk := range m {
		if blk.Height > h {
			h = blk.Height
		}
	}

	return h, nil
}

func GetParentBlockHeader(height uint32) *types.Header {
	chainmgr := GetChainManager()
	chainmgr.lock.RLock()
	defer chainmgr.lock.RUnlock()
	if chainmgr.shardID.IsRootShard() {
		return nil
	}

	m := chainmgr.blockPool.Shards[chainmgr.parentShardID]
	if m == nil {
		return nil
	}
	if blk, present := m[height]; present && blk != nil {
		return blk.Header.Header
	}

	return nil
}

//
// GetShardTxsByParentHeight
// Get cross-shard Tx/Events from parent shard.
// Cross-shard Tx/events of parent shard are delivered to child shards with parent-block propagation.
// NOTE: all cross-shard tx/events should be indexed with (parentHeight, shardHeight)
// TODO:
//
// @start : startHeight of parent block
// @end : endHeight of parent block
//
func GetShardTxsByParentHeight(start, end uint32) map[types.ShardID][]*types.Transaction {
	chainmgr := GetChainManager()

	chainmgr.lock.RLock()
	defer chainmgr.lock.RUnlock()
	if chainmgr.shardID.IsRootShard() {
		return nil
	}

	parentShard := chainmgr.parentShardID
	log.Infof("shard %d get parent shard %d tx, %d - %d", chainmgr.shardID, parentShard, start, end)
	m := chainmgr.blockPool.Shards[parentShard]
	if m == nil {
		return nil
	}
	shardTxs := make(map[types.ShardID][]*types.Transaction)
	for ; start < end+1; start++ {
		if blk, present := m[start]; present && blk != nil {
			if shardTx, present := blk.ShardTxs[chainmgr.shardID]; present && shardTx != nil && shardTx.Tx != nil {
				if shardTxs[parentShard] == nil {
					shardTxs[parentShard] = make([]*types.Transaction, 0)
				}
				shardTxs[parentShard] = append(shardTxs[parentShard], shardTx.Tx)
				log.Infof(">>>> shard %d got remote Tx from parent %d, height: %d",
					chainmgr.shardID, parentShard, start)
			}
		} else {
			log.Infof(">>>> shard %d got remote Tx from parent %d, height: %d, nil block",
				chainmgr.shardID, parentShard, start)
		}
	}

	return shardTxs
}
