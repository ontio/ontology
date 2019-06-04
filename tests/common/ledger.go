
package TestCommon

import (
	"testing"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/store"
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
