package chainmgr

import (
	"math"

	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/common/log"
)

func IsRootShard(shardId uint64) bool {
	return shardId == 0
}

func GetChainManager() *ChainManager {
	return defaultChainManager
}

func GetAccount() *account.Account {
	chainmgr := GetChainManager()
	return chainmgr.account
}

func GetShardID() uint64 {
	return GetChainManager().shardID
}

func GetParentBlockHeight() uint64 {
	chainmgr := GetChainManager()
	chainmgr.lock.RLock()
	defer chainmgr.lock.RUnlock()

	if IsRootShard(chainmgr.shardID) {
		return 0
	}

	m := chainmgr.blockPool.Shards[chainmgr.parentShardID]
	if m == nil {
		return math.MaxUint64
	}

	h := uint64(0)
	for _, blk := range m {
		if blk.Height > h {
			h = blk.Height
		}
	}

	return h
}

func GetParentBlockHeader(height uint64) *types.Header {
	chainmgr := GetChainManager()
	chainmgr.lock.RLock()
	defer chainmgr.lock.RUnlock()
	if IsRootShard(chainmgr.shardID) {
		return nil
	}

	m := chainmgr.blockPool.Shards[chainmgr.parentShardID]
	if m == nil {
		return nil
	}
	if blk, present := m[height]; present && blk != nil {
		return blk.Header.Header
	}

	return nil
}

func GetShardTxsByParentHeight(start, end uint64) map[uint64][]*types.Transaction {
	chainmgr := GetChainManager()

	chainmgr.lock.RLock()
	defer chainmgr.lock.RUnlock()
	if IsRootShard(chainmgr.shardID) {
		return nil
	}

	parentShard := chainmgr.parentShardID
	m := chainmgr.blockPool.Shards[parentShard]
	if m == nil {
		return nil
	}
	shardTxs := make(map[uint64][]*types.Transaction)
	for ; start < end+1; start++ {
		if blk, present := m[start]; present && blk != nil {
			if shardTx, present := blk.ShardTxs[chainmgr.shardID]; present && shardTx != nil && shardTx.Tx != nil {
				if shardTxs[parentShard] == nil {
					shardTxs[parentShard] = make([]*types.Transaction, 0)
				}
				shardTxs[parentShard] = append(shardTxs[parentShard], shardTx.Tx)
				log.Infof(">>>> shard %d got remote Tx from parent %d, height: %d",
					chainmgr.shardID, parentShard, start)
			}
		}
	}

	return shardTxs
}
