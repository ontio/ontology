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

func (self *XShardDB) GetXShardState(id xshard_state.ShardTxID) (*xshard_state.TxState, error) {
	val, err := self.cacheDB.get(common.XSHARD_STATE, []byte(string(id)))
	if err != nil {
		return nil, err
	}

	if len(val) == 0 { // not found
		state := xshard_state.CreateTxState(id)

		return state, nil
	}

	source := comm.NewZeroCopySource(val)
	state := &xshard_state.TxState{}
	err = state.Deserialization(source)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (self *XShardDB) SetXShardState(state *xshard_state.TxState) {
	buf := comm.SerializeToBytes(state)
	self.cacheDB.put(common.XSHARD_STATE, []byte(string(state.TxID)), buf)
}

func (self *XShardDB) SetXShardMsgInBlock(blockHeight uint32, msgs []xshard_types.CommonShardMsg) {
	shardMsgMap := make(map[comm.ShardID][]xshard_types.CommonShardMsg)
	for _, msg := range msgs {
		shardMsgMap[msg.GetTargetShardID()] = append(shardMsgMap[msg.GetTargetShardID()], msg)
	}
	keys := comm.NewZeroCopySink(8)
	val := comm.NewZeroCopySink(1024)
	shards := comm.NewZeroCopySink(2 + 8*len(shardMsgMap))
	shards.WriteUint32(uint32(len(shardMsgMap)))
	for shardID, shardMsgs := range shardMsgMap {
		shards.WriteUint64(shardID.ToUint64())
		keys.Reset()
		keys.WriteUint32(blockHeight)
		keys.WriteUint64(shardID.ToUint64())

		val.Reset()
		xshard_types.EncodeShardCommonMsgs(val, shardMsgs)
		self.cacheDB.put(common.XSHARD_KEY_REQS_IN_BLOCK, keys.Bytes(), val.Bytes())
	}
	keys.Reset()
	keys.WriteUint32(blockHeight)

	self.cacheDB.put(common.XSHARD_KEY_SHARDS_IN_BLOCK, keys.Bytes(), shards.Bytes())
}
