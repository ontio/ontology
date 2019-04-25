/*
 * Copyright (C) 2018 The ontology Authors
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

package storage

import (
	"fmt"
	comm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/chainmgr/xshard_state"
	"github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/ontio/ontology/core/xshard_types"
)

// CacheDB is smart contract execute cache, it contain transaction cache and block cache
// When smart contract execute finish, need to commit transaction cache to block cache
type XShardDB struct {
	cacheDB *CacheDB
	states  map[xshard_state.ShardTxID]*xshard_state.TxState
}

func NewXShardDB(store *overlaydb.OverlayDB) *XShardDB {
	return &XShardDB{cacheDB: NewCacheDB(store)}
}

func (self *XShardDB) Reset() {
	self.cacheDB.Reset()
}

// Commit current transaction cache to block cache
func (self *XShardDB) Commit() {
	self.cacheDB.Commit()
}

func (self *XShardDB) GetXshardReqsInBlock(blockHeight uint32, shardID comm.ShardID) ([][]byte, error) {
	panic("unimplemented")
	return nil, nil
}

func (self *XShardDB) AddXShardMsgInBlock(blockHeight uint32, req *xshard_types.CommonShardMsg) error {
	fmt.Println("AddXShardMsgInBlock: unimplemented")
	return nil

}

func (self *XShardDB) SetShardTxState(id xshard_state.ShardTxID, state *xshard_state.TxState) {
	self.states[id] = state
}

func (self *XShardDB) GetShardTxState(id xshard_state.ShardTxID) *xshard_state.TxState {
	return self.states[id]
}

func (self *XShardDB) AddToShard(blockHeight uint32, to comm.ShardID) error {
	shardIDs, err := self.GetToShards(blockHeight)
	if err != nil {
		return err
	}

	shardIDs = append(shardIDs, to)

	sink := comm.NewZeroCopySink(len(shardIDs) * 8)
	for _, id := range shardIDs {
		sink.WriteUint64(id.ToUint64())
	}

	keys := comm.NewZeroCopySink(4)
	keys.WriteUint32(blockHeight)
	self.cacheDB.put(common.XSHARD_KEY_SHARDS_IN_BLOCK, keys.Bytes(), sink.Bytes())
	return nil
}

func (self *XShardDB) GetToShards(blockHeight uint32) ([]comm.ShardID, error) {
	keys := comm.NewZeroCopySink(4)
	keys.WriteUint32(blockHeight)
	val, err := self.cacheDB.get(common.XSHARD_KEY_SHARDS_IN_BLOCK, keys.Bytes())
	if err != nil {
		return nil, err
	}

	source := comm.NewZeroCopySource(val)

	shardIDs := make([]comm.ShardID, 0, len(val)/8)
	for {
		id, eof := source.NextUint64()
		if eof {
			break
		}
		shardID, err := comm.NewShardID(id)
		if err != nil {
			return nil, err
		}

		shardIDs = append(shardIDs, shardID)
	}

	return shardIDs, nil
}
