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
	shardmsg "github.com/ontio/ontology/core/chainmgr/message"
	"github.com/ontio/ontology/core/types"
)

// cross  shard pool
//type ShardTxMap map[uint32]*types.Transaction // indexed by BlockHeight

//type ShardBlockMap map[uint32]*shardmsg.ShardBlockInfo // indexed by BlockHeight

type ShardMsgMap map[uint32]*shardmsg.CrossShardMsgInfo // indexed by BlockHeight
type CrossShardPool struct {
	lock        sync.RWMutex
	ShardID     common.ShardID
	Shards      map[common.ShardID]ShardMsgMap // indexed by FromShardID
	MaxBlockCap uint32
}

// BlockHeader and Cross-Shard Txs of other shards
var crossShardPool *CrossShardPool

func InitCrossShardPool(shard common.ShardID, historyCap uint32) {
	crossShardPool = &CrossShardPool{
		ShardID:     shard,
		Shards:      make(map[common.ShardID]ShardMsgMap),
		MaxBlockCap: historyCap,
	}
}

func AddCrossShardInfo(crossShardMsg *shardmsg.CrossShardMsg, tx *types.Transaction) error {
	pool := crossShardPool
	pool.lock.Lock()
	defer pool.lock.Unlock()
	if _, present := pool.Shards[crossShardMsg.FromShardID]; !present {
		pool.Shards[crossShardMsg.FromShardID] = make(ShardMsgMap)
	}
	m := pool.Shards[crossShardMsg.FromShardID]
	if m == nil {
		return fmt.Errorf("add shard cross tx, nil map")
	}
	if crossMsg, present := m[crossShardMsg.MsgHeight]; present {
		if crossMsg.ShardMsg.CrossShardMsgRoot == crossShardMsg.CrossShardMsgRoot {
			return fmt.Errorf("add shard block, dup blk")
		}
	}
	log.Infof("chainmgr AddBlock from shard %d, block %d", crossShardMsg.FromShardID.ToUint64(), crossShardMsg.MsgHeight)
	crossShardMsgInfo := &shardmsg.CrossShardMsgInfo{
		ShardMsg: crossShardMsg,
		ShardTx: &shardmsg.ShardBlockTx{
			Tx: tx,
		},
	}
	m[crossShardMsg.MsgHeight] = crossShardMsgInfo
	// if too much block cached in map, drop old blocks
	if uint32(len(m)) < pool.MaxBlockCap {
		return nil
	}
	h := crossShardMsg.MsgHeight
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
	for shardID, shardBlocks := range pool.Shards {
		txs := make([]*types.Transaction, 0)
		for _, shard := range shardBlocks {
			txs = append(txs, shard.ShardTx.Tx)
		}
		shardTxs[shardID.ToUint64()] = txs
	}
	return shardTxs
}
