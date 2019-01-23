package chainmgr

import (
	"math"

	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
)

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

	m := chainmgr.blockPool.Shards[chainmgr.parentShardID]
	if m == nil {
		return nil
	}
	if blk, present := m[height]; present && blk != nil {
		return blk.Header.Header
	}

	return nil
}

func GetBlockEventsByParentHeight(height uint64) map[uint64][]shardstates.ShardMgmtEvent {
	chainmgr := GetChainManager()

	chainmgr.lock.RLock()
	defer chainmgr.lock.RUnlock()

	shardEvts := make(map[uint64][]shardstates.ShardMgmtEvent)

	m := chainmgr.blockPool.Shards[chainmgr.parentShardID]
	if m == nil {
		return shardEvts
	}
	if blk, present := m[height]; present && blk != nil {
		shardEvts[chainmgr.parentShardID] = blk.Events
	}

	// TODO: add shard event from sibling shards

	return shardEvts
}
