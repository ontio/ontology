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

package ledgerstore

import (
	"bytes"
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	common2 "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/types"
)

var dbDir = "./testDB"

func TestResetBlockCacheStore(t *testing.T) {
	shardID, err := types.NewShardID(10)
	if err != nil {
		t.Fatalf("init shard id: %s", err)
	}

	cacheStore, err := ResetBlockCacheStore(shardID, dbDir)
	if err != nil {
		t.Fatalf("reset block cache failed: %s", err)
	}

	ite := cacheStore.store.NewIterator([]byte{})
	if ite.First() {
		t.Fatalf("reset block cache is not empty")
	}
	ite.Release()
}

func newTestBlock(height uint32, shardID types.ShardID) *types.Block {
	header := &types.Header{}
	header.Version = common.CURR_HEADER_VERSION
	header.Height = height
	header.ShardID = shardID.ToUint64()
	header.Bookkeepers = make([]keypair.PublicKey, 0)
	header.SigData = make([][]byte, 0)

	return &types.Block{
		Header:       header,
		ShardTxs:     make(map[uint64][]*types.Transaction),
		Transactions: make([]*types.Transaction, 0),
	}
}

func TestBlockCacheStore_Ops(t *testing.T) {
	shardID, err := types.NewShardID(10)
	if err != nil {
		t.Fatalf("init shard id: %s", err)
	}

	cacheStore, err := ResetBlockCacheStore(shardID, dbDir)
	if err != nil {
		t.Fatalf("reset block cache failed: %s", err)
	}

	var height uint32 = 123
	blk := newTestBlock(height, shardID)
	hashRoot := common.Uint256{1, 2, 3}
	err = cacheStore.PutBlock(blk, hashRoot)
	if err != nil {
		t.Fatalf("put block failed: %s", err)
	}

	blk2, hashRoot2, err := cacheStore.GetBlock(height)
	if err != nil {
		t.Fatalf("get block failed: %s", err)
	}
	if bytes.Compare(hashRoot[:], hashRoot2[:]) != 0 {
		t.Fatalf("get block unmatched hashroot")
	}
	if blk2.Header.ShardID != shardID.ToUint64() {
		t.Fatalf("get block unmatched shard id")
	}
	cacheStore.DelBlock(height)
	if _, _, err := cacheStore.GetBlock(height); err != common2.ErrNotFound {
		t.Fatalf("del block failed: %s", err)
	}
}
