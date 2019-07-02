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

package TestCommon

import (
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/store"
	"github.com/ontio/ontology/core/types"
)

type BlockchainResult map[uint32]*store.ExecuteResult

var resultCache map[common.ShardID]BlockchainResult

func init() {
	resultCache = make(map[common.ShardID]BlockchainResult)
}

func AddResult(t *testing.T, shardID common.ShardID, height uint32, result *store.ExecuteResult) {
	if _, present := resultCache[shardID]; !present {
		resultCache[shardID] = make(BlockchainResult)
	}
	if _, present := resultCache[shardID][height]; present {
		t.Fatalf("add block result %d:%d, already exists", shardID, height)
	}
	resultCache[shardID][height] = result
}

func GetResult(t *testing.T, shardID common.ShardID, height uint32) *store.ExecuteResult {
	if _, present := resultCache[shardID]; !present {
		return nil
	}

	return resultCache[shardID][height]
}

func DelResult(t *testing.T, shardID common.ShardID, height uint32) {
	if _, present := resultCache[shardID]; !present {
		t.Fatalf("del block result %d:%d, shard not exists", shardID, height)
	}
	if _, present := resultCache[shardID][height]; !present {
		t.Fatalf("del block result %d:%d, blk not exists", shardID, height)
	}
	delete(resultCache[shardID], height)
}

func ExecBlock(t *testing.T, shardID common.ShardID, blk *types.Block) common.Uint256 {
	lgr := ledger.GetShardLedger(shardID)
	if lgr == nil {
		t.Fatalf("exec block failed to get shard ledger: %d", shardID)
	}

	result, err := lgr.ExecuteBlock(blk)
	if err != nil {
		t.Fatalf("execut shard %d block: %s", shardID, err)
	}
	AddResult(t, shardID, blk.Header.Height, &result)

	return result.MerkleRoot
}

func SubmitBlock(t *testing.T, shardID common.ShardID, blk *types.Block) {
	lgr := ledger.GetShardLedger(shardID)
	if lgr == nil {
		t.Fatalf("submit block failed to get shard ledger: %d", shardID)
	}

	result := GetResult(t, shardID, blk.Header.Height)
	if result == nil {
		t.Fatalf("submit block failed to get exec result: %d:%d", shardID, blk.Header.Height)
	}

	if err := lgr.SubmitBlock(blk, *result); err != nil {
		t.Fatalf("submit block %d:%d failed, %s", shardID, blk.Header.Height, err)
	}
}

func AddBlock(t *testing.T, shardID common.ShardID, block *types.Block, merkleRoot common.Uint256) {
	if lgr := ledger.GetShardLedger(shardID); lgr != nil {
		if err := lgr.AddBlock(block, merkleRoot); err != nil {
			t.Fatalf("add block to shard %d: %s", shardID, err)
		}
	}
	t.Fatalf("get height with invalid shard %d", shardID)
}
