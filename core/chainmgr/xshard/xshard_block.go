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

package xshard

import (
	"fmt"
	"sync"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/types"
)

// cross  shard pool

type CrossShardPool struct {
	lock        sync.RWMutex
	ShardID     common.ShardID
	Shards      map[uint64]map[common.Uint256]*types.CrossShardTxInfos // indexed by FromShardID
	MaxBlockCap uint32
}

// BlockHeader and Cross-Shard Txs of other shards
var crossShardPool *CrossShardPool

func InitCrossShardPool(shardID common.ShardID, historyCap uint32) {
	crossShardPool = &CrossShardPool{
		ShardID:     shardID,
		Shards:      make(map[uint64]map[common.Uint256]*types.CrossShardTxInfos),
		MaxBlockCap: historyCap,
	}
}

func AddCrossShardInfo(crossShardMsg *types.CrossShardMsg, tx *types.Transaction) error {
	pool := crossShardPool
	crossShardTxInfo := &types.CrossShardTxInfos{
		ShardMsg: crossShardMsg,
		Tx:       tx,
	}
	pool.lock.Lock()
	defer pool.lock.Unlock()
	if _, present := pool.Shards[crossShardMsg.FromShardID.ToUint64()]; !present {
		pool.Shards[crossShardMsg.FromShardID.ToUint64()] = make(map[common.Uint256]*types.CrossShardTxInfos)
	}
	m := pool.Shards[crossShardMsg.FromShardID.ToUint64()]
	if m == nil {
		return fmt.Errorf("add shard cross tx, nil map")
	}
	m[crossShardTxInfo.ShardMsg.CrossShardMsgRoot] = crossShardTxInfo
	log.Infof("chainmgr AddBlock from shard %d, block %d", crossShardMsg.FromShardID.ToUint64(), crossShardMsg.MsgHeight)
	return nil
}

//
// GetShardTxsByParentHeight
// Get cross-shard Tx/Events from parent shard.
// Cross-shard Tx/events of parent shard are delivered to child shards with parent-block propagation.
// NOTE: all cross-shard tx/events should be indexed with (parentHeight, shardHeight)
//

func GetCrossShardTxs() map[uint64][]*types.CrossShardTxInfos {
	pool := crossShardPool
	if pool.ShardID.IsRootShard() {
		return nil
	}
	pool.lock.RLock()
	defer pool.lock.RUnlock()
	crossShardInfo := make([]*types.CrossShardTxInfos, 0)
	crossShardMapInfos := make(map[uint64][]*types.CrossShardTxInfos)
	for shardID, shardTxs := range pool.Shards {
		for _, shardTx := range shardTxs {
			crossShardInfo = append(crossShardInfo, shardTx)
		}
		crossShardMapInfos[shardID] = crossShardInfo
	}
	return crossShardMapInfos
}

func DelCrossShardTxs(shardID common.ShardID, msgHash common.Uint256) error {
	pool := crossShardPool
	if pool.ShardID.IsRootShard() {
		return nil
	}
	pool.lock.RLock()
	defer pool.lock.RUnlock()
	if crossShardTxInfos, present := pool.Shards[shardID.ToUint64()]; !present {
		log.Infof("delcrossshardtxs shardID:%v,not exist", shardID)
		return nil
	} else {
		delete(crossShardTxInfos, msgHash)
	}
	return nil
}
