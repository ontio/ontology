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
	"bytes"
	"fmt"
	"sync"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	shardmsg "github.com/ontio/ontology/core/chainmgr/message"
	"github.com/ontio/ontology/core/types"
)

////////////////////////////////////
//
//  shard block pool
//
////////////////////////////////////

type ShardBlockMap map[uint32]*shardmsg.ShardBlockInfo // indexed by BlockHeight

type ShardBlockPool struct {
	lock        sync.RWMutex
	ShardID     common.ShardID
	Shards      map[common.ShardID]ShardBlockMap // indexed by FromShardID
	MaxBlockCap uint32
}

// BlockHeader and Cross-Shard Txs of other shards
var shardBlockPool *ShardBlockPool

func InitShardBlockPool(shard common.ShardID, historyCap uint32) {
	shardBlockPool = &ShardBlockPool{
		ShardID:     shard,
		Shards:      make(map[common.ShardID]ShardBlockMap),
		MaxBlockCap: historyCap,
	}
}

func GetBlockInfo(shardID common.ShardID, height uint32) *shardmsg.ShardBlockInfo {
	pool := shardBlockPool

	pool.lock.RLock()
	defer pool.lock.RUnlock()
	if m, present := pool.Shards[shardID]; present && m != nil {
		return m[height]
	}
	return nil
}

func AddBlockInfo(blkInfo *shardmsg.ShardBlockInfo) error {
	pool := shardBlockPool

	pool.lock.Lock()
	defer pool.lock.Unlock()
	if _, present := pool.Shards[blkInfo.FromShardID]; !present {
		pool.Shards[blkInfo.FromShardID] = make(ShardBlockMap)
	}

	m := pool.Shards[blkInfo.FromShardID]
	if m == nil {
		return fmt.Errorf("add shard block, nil map")
	}
	if blk, present := m[blkInfo.Height]; present {
		hdr := blk.Block.Header
		if hdr != nil && bytes.Compare(hdr.BlockRoot[:], blkInfo.Block.Header.BlockRoot[:]) == 0 {
			return fmt.Errorf("add shard block, dup blk")
		}
	}

	log.Infof("chainmgr AddBlock from shard %d, block %d", blkInfo.FromShardID, blkInfo.Height)
	m[blkInfo.Height] = blkInfo

	// if too much block cached in map, drop old blocks
	if uint32(len(m)) < pool.MaxBlockCap {
		return nil
	}
	h := blkInfo.Height
	for _, blk := range m {
		if blk.Height > h {
			h = blk.Height
		}
	}

	toDrop := make([]uint32, 0)
	for _, blk := range m {
		if blk.Height < h-uint32(pool.MaxBlockCap) {
			toDrop = append(toDrop, blk.Height)
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
// TODO:
//
// @start : startHeight of parent block
// @end : endHeight of parent block
//
func GetShardTxsByParentHeight(height uint32) map[common.ShardID][]*types.Transaction {
	pool := shardBlockPool
	shardID := pool.ShardID
	if shardID.IsRootShard() {
		return nil
	}
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	parentShard := shardID.ParentID()
	log.Infof("shard %d get parent shard %d tx, %d", shardID, parentShard, height)
	m := pool.Shards[parentShard]
	if m == nil {
		return nil
	}
	shardTxs := make(map[common.ShardID][]*types.Transaction)
	if blk, present := m[height]; present && blk != nil {
		if shardTx, present := blk.ShardTxs[shardID]; present && shardTx != nil && shardTx.Tx != nil {
			if shardTxs[parentShard] == nil {
				shardTxs[parentShard] = make([]*types.Transaction, 0)
			}
			shardTxs[parentShard] = append(shardTxs[parentShard], shardTx.Tx)
			log.Infof(">>>> shard %d got remote Tx from parent %d, height: %d",
				shardID, parentShard, height)
		}
	} else {
		log.Infof(">>>> shard %d got remote Tx from parent %d, height: %d, nil block",
			shardID, parentShard, height)
	}
	return shardTxs
}
