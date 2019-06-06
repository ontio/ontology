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

	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	crossshard "github.com/ontio/ontology/core/chainmgr/message"
	"github.com/ontio/ontology/core/ledger"
	com "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/types"
)

// cross  shard pool

type CrossShardPool struct {
	lock        sync.RWMutex
	ShardID     common.ShardID
	Shards      map[uint64]map[common.Uint256]*types.CrossShardMsg // key:indexed by FromShardID key:preMsgHash
	MaxBlockCap uint32
	ShardInfo   map[common.ShardID]bool
}

// BlockHeader and Cross-Shard Txs of other shards
var crossShardPool *CrossShardPool

func InitCrossShardPool(shardID common.ShardID, historyCap uint32) {
	crossShardPool = &CrossShardPool{
		ShardID:     shardID,
		Shards:      make(map[uint64]map[common.Uint256]*types.CrossShardMsg),
		ShardInfo:   make(map[common.ShardID]bool),
		MaxBlockCap: historyCap,
	}
}

func InitShardInfo(lgr *ledger.Ledger) error {
	pool := crossShardPool
	pool.lock.Lock()
	defer pool.lock.Unlock()
	shardIds, err := lgr.GetAllShardIDs()
	if err != nil {
		if err != com.ErrNotFound {
			return fmt.Errorf("GetAllShardIDs failed err:%s", err)
		}
	}
	for _, shardId := range shardIds {
		pool.ShardInfo[shardId] = true
		msgHash, err := GetCrossShardHashByShardID(shardId)
		if err != nil {
			if err != com.ErrNotFound {
				return fmt.Errorf("InitShardInfo GetCrossShardHashByShardID shardID:%v,err:%s", shardId, err)
			} else {
				break
			}
		}
		for {
			msg, err := lgr.GetCrossShardMsgByHash(msgHash)
			if err != nil {
				if err != com.ErrNotFound {
					return fmt.Errorf("InitShardInfo GetCrossShardMsgByHash hash:%s,err:%s", msgHash.ToHexString(), err)
				} else {
					break
				}
			}
			if _, present := pool.Shards[shardId.ToUint64()]; !present {
				pool.Shards[shardId.ToUint64()] = make(map[common.Uint256]*types.CrossShardMsg)
			}
			m := pool.Shards[shardId.ToUint64()]
			if m == nil {
				return fmt.Errorf("pool shard shardId:%v, nil map", shardId)
			}
			if _, present := m[msg.CrossShardMsgInfo.PreCrossShardMsgHash]; present {
				log.Debugf("InitShardInfo msgHash:%s had exist", msgHash.ToHexString())
				continue
			}
			m[msg.CrossShardMsgInfo.PreCrossShardMsgHash] = msg
			msgHash = msg.CrossShardMsgInfo.CrossShardMsgRoot
		}
	}
	return nil
}

func AddShardInfo(shardID common.ShardID) {
	pool := crossShardPool
	pool.lock.Lock()
	defer pool.lock.Unlock()
	if _, present := pool.ShardInfo[shardID]; present {
		return
	}
	pool.ShardInfo[shardID] = true

	shardIds := make([]common.ShardID, 0)
	for shardId, _ := range pool.ShardInfo {
		shardIds = append(shardIds, shardId)
	}
	err := ledger.DefLedger.SaveAllShardIDs(shardIds)
	if err != nil {
		log.Errorf("SaveAllShardIDs shardId:%v,err:%s", shardID, err)
		return
	}
}

func GetShardInfo() map[common.ShardID]bool {
	pool := crossShardPool
	pool.lock.Lock()
	defer pool.lock.Unlock()
	return pool.ShardInfo
}

func GetCrossShardHashByShardID(shardID common.ShardID) (common.Uint256, error) {
	pool := crossShardPool
	pool.lock.Lock()
	defer pool.lock.Unlock()
	return ledger.DefLedger.GetCrossShardHash(shardID)
}
func SaveCrossShardHash(shardID common.ShardID, msgHash common.Uint256) error {
	pool := crossShardPool
	pool.lock.Lock()
	defer pool.lock.Unlock()
	return ledger.DefLedger.SaveCrossShardHash(shardID, msgHash)
}

func AddCrossShardInfo(ledger *ledger.Ledger, crossShardMsg *types.CrossShardMsg) error {
	pool := crossShardPool
	pool.lock.Lock()
	defer pool.lock.Unlock()
	fromShardID := crossShardMsg.CrossShardMsgInfo.FromShardID.ToUint64()
	if _, present := pool.Shards[fromShardID]; !present {
		pool.Shards[fromShardID] = make(map[common.Uint256]*types.CrossShardMsg)
	}
	m := pool.Shards[fromShardID]
	if m == nil {
		return fmt.Errorf("add shard cross shard msg:%d, nil map", fromShardID)
	}
	if _, present := m[crossShardMsg.CrossShardMsgInfo.PreCrossShardMsgHash]; present {
		log.Debugf("SaveCrossShardMsgByShardID premsgHash:%s had save", crossShardMsg.CrossShardMsgInfo.PreCrossShardMsgHash.ToHexString())
		return nil
	}
	m[crossShardMsg.CrossShardMsgInfo.PreCrossShardMsgHash] = crossShardMsg
	err := ledger.SaveCrossShardMsgByHash(crossShardMsg.CrossShardMsgInfo.PreCrossShardMsgHash, crossShardMsg)
	if err != nil {
		return fmt.Errorf("SaveCrossShardMsgByShardID shardID:%v,msgHash:%s,err:%s", crossShardMsg.CrossShardMsgInfo.FromShardID, crossShardMsg.CrossShardMsgInfo.PreCrossShardMsgHash.ToHexString(), err)
	}
	AddShardInfo(crossShardMsg.CrossShardMsgInfo.FromShardID)
	log.Infof("chainmgr AddBlock from shard %d,msgHash:%v, block height %d", fromShardID, crossShardMsg.CrossShardMsgInfo.PreCrossShardMsgHash.ToHexString(), crossShardMsg.CrossShardMsgInfo.MsgHeight)
	return nil
}

//
// GetShardTxsByParentHeight
// Get cross-shard Tx/Events from parent shard.
// Cross-shard Tx/events of parent shard are delivered to child shards with parent-block propagation.
// NOTE: all cross-shard tx/events should be indexed with (parentHeight, shardHeight)
//

func GetCrossShardTxs(account *account.Account, FromShardID common.ShardID) (map[uint64][]*types.CrossShardTxInfos, error) {
	pool := crossShardPool
	pool.lock.RLock()
	defer pool.lock.RUnlock()
	crossShardInfo := make([]*types.CrossShardTxInfos, 0)
	crossShardMapInfos := make(map[uint64][]*types.CrossShardTxInfos)
	for shardID, shardMsgs := range pool.Shards {
		id, err := common.NewShardID(shardID)
		if err != nil {
			log.Errorf("shardID new shardID:%d,err:%s", shardID, err)
			continue
		}
		msgHash, err := GetCrossShardHashByShardID(common.NewShardIDUnchecked(shardID))
		if err != nil {
			if err != com.ErrNotFound {
				log.Errorf("GetCrossShardHashByShardID shardID:%v,err:%s", shardID, err)
				continue
			}
		}
		crossShardMsgs := make([]*types.CrossShardMsg, 0)
		for {
			if shardMsg, persent := shardMsgs[msgHash]; !persent {
				msg, err := ledger.DefLedger.GetCrossShardMsgByHash(msgHash)
				if err != nil {
					if err != com.ErrNotFound {
						break
					} else {
						return nil, fmt.Errorf("GetCrossShardMsgByHash msgHash:%s,err:%v", msgHash, err)
					}
				} else {
					shardMsg = msg
				}
			} else {
				crossShardMsgs = append(crossShardMsgs, shardMsg)
				msgHash = shardMsg.CrossShardMsgInfo.PreCrossShardMsgHash
			}
		}
		for _, msg := range crossShardMsgs {
			if id.IsParentID() && msg.CrossShardMsgInfo.SignMsgHeight < ledger.GetShardLedger(id).GetCurrentBlockHeight() {
				break
			}
			tx, err := crossshard.NewCrossShardTxMsg(account, msg.CrossShardMsgInfo.MsgHeight, FromShardID, config.DefConfig.Common.GasPrice, config.DefConfig.Common.GasLimit, msg.ShardMsg)
			if err != nil {
				log.Errorf("handleCrossShardMsg NewCrossShardTxMsg height:%d,err:%s", msg.CrossShardMsgInfo.MsgHeight, err)
				break
			}
			shardTxInfo := &types.CrossShardTxInfos{
				ShardMsg: msg.CrossShardMsgInfo,
				Tx:       tx,
			}
			crossShardInfo = append(crossShardInfo, shardTxInfo)
		}
		crossShardMapInfos[shardID] = crossShardInfo
	}
	return crossShardMapInfos, nil
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
				SaveCrossShardHash(common.NewShardIDUnchecked(shardID), shardTx.ShardMsg.PreCrossShardMsgHash)
			}
		}
	}
	return nil
}
