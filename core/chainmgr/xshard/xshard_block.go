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
type ShardTxMap map[uint32]*types.Transaction // indexed by BlockHeight

type CrossShardPool struct {
	lock        sync.RWMutex
	ShardID     common.ShardID
	Shards      map[uint64]ShardTxMap // indexed by FromShardID
	MaxBlockCap uint32
}

// BlockHeader and Cross-Shard Txs of other shards
var crossShardPool *CrossShardPool

func InitCrossShardPool(shard common.ShardID, historyCap uint32) {
	crossShardPool = &CrossShardPool{
		ShardID:     shard,
		Shards:      make(map[uint64]ShardTxMap),
		MaxBlockCap: historyCap,
	}
}

func AddCrossShardInfo(shardID uint64, height uint32, tx *types.Transaction) error {
	pool := crossShardPool

	pool.lock.Lock()
	defer pool.lock.Unlock()
	if _, present := pool.Shards[shardID]; !present {
		pool.Shards[shardID] = make(ShardTxMap)
	}
	m := pool.Shards[shardID]
	if m == nil {
		return fmt.Errorf("add shard cross tx, nil map")
	}
	if crossTx, present := m[height]; present {
		txhash := crossTx.Hash()
		if tx != nil && txhash != tx.Hash() {
			return fmt.Errorf("add shard cross tx, dup tx")
		}
	}
	log.Infof("chainmgr AddBlock from shard %d, block %d", shardID, height)
	m[height] = tx
	// if too much block cached in map, drop old blocks
	if uint32(len(m)) < pool.MaxBlockCap {
		return nil
	}
	h := height
	for height, _ := range m {
		if height > h {
			h = height
		}
	}
	toDrop := make([]uint32, 0)
	for height, _ := range m {
		if height < h-uint32(pool.MaxBlockCap) {
			toDrop = append(toDrop, height)
		}
	}
	for _, blkHeight := range toDrop {
		delete(m, blkHeight)
	}
	return nil
}

//
// GetShardTxsByParentHeight
// Get cross-shard Tx/Events from parent shard.
// Cross-shard Tx/events of parent shard are delivered to child shards with parent-block propagation.
// NOTE: all cross-shard tx/events should be indexed with (parentHeight, shardHeight)
//
func GetCrossShardTxs() map[uint64][]*types.Transaction {
	pool := crossShardPool
	shardID := pool.ShardID
	if shardID.IsRootShard() {
		return nil
	}
	pool.lock.RLock()
	defer pool.lock.RUnlock()
	shardTxs := make(map[uint64][]*types.Transaction)
	for shardID, shardtxs := range pool.Shards {
		txs := make([]*types.Transaction, 0)
		for _, tx := range shardtxs {
			txs = append(txs, tx)
		}
		shardTxs[shardID] = txs
	}
	return shardTxs
}
