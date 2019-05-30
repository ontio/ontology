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
	"github.com/ontio/ontology/core/ledger"
	com "github.com/ontio/ontology/core/store/common"
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

func AddCrossShardInfo(ledger *ledger.Ledger, crossShardMsg *types.CrossShardMsg, tx *types.Transaction) error {
	pool := crossShardPool
	crossShardTxInfo := &types.CrossShardTxInfos{
		ShardMsg: crossShardMsg.CrossShardMsgInfo,
		Tx:       tx,
	}
	pool.lock.Lock()
	defer pool.lock.Unlock()
	if _, present := pool.Shards[crossShardMsg.CrossShardMsgInfo.FromShardID.ToUint64()]; !present {
		pool.Shards[crossShardMsg.CrossShardMsgInfo.FromShardID.ToUint64()] = make(map[common.Uint256]*types.CrossShardTxInfos)
	}
	m := pool.Shards[crossShardMsg.CrossShardMsgInfo.FromShardID.ToUint64()]
	if m == nil {
		return fmt.Errorf("add shard cross tx, nil map")
	}
	m[crossShardTxInfo.ShardMsg.CrossShardMsgRoot] = crossShardTxInfo
	shardTxInfos, err := ledger.GetCrossShardMsgByShardID(crossShardMsg.CrossShardMsgInfo.FromShardID)
	if err != nil {
		if err != com.ErrNotFound {
			return fmt.Errorf("GetCrossShardMsgByShardID shardID:%v,err:%s", crossShardMsg.CrossShardMsgInfo.FromShardID, err)
		}
	}
	shardTxInfos = append(shardTxInfos, crossShardTxInfo)

	err = ledger.SaveCrossShardMsgByShardID(crossShardMsg.CrossShardMsgInfo.FromShardID, shardTxInfos)
	if err != nil {
		return fmt.Errorf("SaveCrossShardMsgByShardID shardID:%v,err:%s", crossShardMsg.CrossShardMsgInfo.FromShardID, err)
	}
	log.Infof("chainmgr AddBlock from shard %d, block %d", crossShardMsg.CrossShardMsgInfo.FromShardID.ToUint64(), crossShardMsg.CrossShardMsgInfo.MsgHeight)
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
	pool.lock.RLock()
	defer pool.lock.RUnlock()
	crossShardInfo := make([]*types.CrossShardTxInfos, 0)
	crossShardMapInfos := make(map[uint64][]*types.CrossShardTxInfos)
	for shardID, shardTxs := range pool.Shards {
		id, err := common.NewShardID(shardID)
		if err != nil {
			log.Errorf("shardID new shardID:%d,err:%s", shardID, err)
			continue
		}
		for _, shardTx := range shardTxs {
			if id.IsParentID() && shardTx.ShardMsg.SignMsgHeight < ledger.GetShardLedger(id).GetCurrentBlockHeight() {
				continue
			}
			crossShardInfo = append(crossShardInfo, shardTx)
		}
		crossShardMapInfos[shardID] = crossShardInfo
	}
	return crossShardMapInfos
}

func DelCrossShardTxs(crossShardTxs map[uint64][]*types.CrossShardTxInfos) error {
	pool := crossShardPool
	pool.lock.RLock()
	defer pool.lock.RUnlock()
	for shardID, shardTxs := range crossShardTxs {
		for _, shardTx := range shardTxs {
			if crossShardTxInfos, present := pool.Shards[shardID]; !present {
				log.Infof("delcrossshardtxs shardID:%d,not exist", shardID)
				return nil
			} else {
				delete(crossShardTxInfos, shardTx.ShardMsg.CrossShardMsgRoot)
			}
		}
	}
	return nil
}
