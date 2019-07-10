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

package ledgerstore

import (
	"fmt"

	"github.com/hashicorp/golang-lru"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
)

const (
	BLOCK_CACHE_SIZE       = 10    //Block cache size
	SHARD_TX_CACHE_SIZE    = 20    //block shardTx cache size
	TRANSACTION_CACHE_SIZE = 10000 //Transaction cache size
)

//Value of transaction cache
type TransactionCacheaValue struct {
	Tx     *types.Transaction
	Height uint32
}

type ShardTxCacheValue struct {
	ShardTx *types.CrossShardTxInfos
	Height  uint32
}

//BlockCache with block cache and transaction hash
type BlockCache struct {
	blockCache        *lru.ARCCache
	shardTxCache      *lru.ARCCache
	transactionCache  *lru.ARCCache
}

//NewBlockCache return BlockCache instance
func NewBlockCache() (*BlockCache, error) {
	blockCache, err := lru.NewARC(BLOCK_CACHE_SIZE)
	if err != nil {
		return nil, fmt.Errorf("NewARC block error %s", err)
	}
	shardTxCache, err := lru.NewARC(SHARD_TX_CACHE_SIZE)
	if err != nil {
		return nil, fmt.Errorf("NewARC shardTx err :%s", err)
	}
	transactionCache, err := lru.NewARC(TRANSACTION_CACHE_SIZE)
	if err != nil {
		return nil, fmt.Errorf("NewARC header error %s", err)
	}
	return &BlockCache{
		blockCache:        blockCache,
		shardTxCache:      shardTxCache,
		transactionCache:  transactionCache,
	}, nil
}

//AddBlock to cache
func (this *BlockCache) AddBlock(block *types.Block) {
	blockHash := block.Hash()
	this.blockCache.Add(string(blockHash.ToArray()), block)
}

//GetBlock return block by block hash from cache
func (this *BlockCache) GetBlock(blockHash common.Uint256) *types.Block {
	block, ok := this.blockCache.Get(string(blockHash.ToArray()))
	if !ok {
		return nil
	}
	return block.(*types.Block)
}

//ContainBlock return whether block is in cache
func (this *BlockCache) ContainBlock(blockHash common.Uint256) bool {
	return this.blockCache.Contains(string(blockHash.ToArray()))
}

func (this *BlockCache) AddShardTx(tx *types.CrossShardTxInfos, height uint32) {
	shardTxHash := tx.Tx.Hash()
	this.shardTxCache.Add(string(shardTxHash.ToArray()), &ShardTxCacheValue{
		ShardTx: tx,
		Height:  height,
	})
}

func (this *BlockCache) GetShardTx(shardTxHash common.Uint256) (*types.CrossShardTxInfos, uint32) {
	value, ok := this.shardTxCache.Get(string(shardTxHash.ToArray()))
	if !ok {
		return nil, 0
	}
	txValue := value.(*ShardTxCacheValue)
	return txValue.ShardTx, txValue.Height
}

func (this *BlockCache) ContainShardTx(shardTxHash common.Uint256) bool {
	return this.shardTxCache.Contains(string(shardTxHash.ToArray()))
}

//AddTransaction add transaction to block cache
func (this *BlockCache) AddTransaction(tx *types.Transaction, height uint32) {
	txHash := tx.Hash()
	this.transactionCache.Add(string(txHash.ToArray()), &TransactionCacheaValue{
		Tx:     tx,
		Height: height,
	})
}

//GetTransaction return transaction by transaction hash from cache
func (this *BlockCache) GetTransaction(txHash common.Uint256) (*types.Transaction, uint32) {
	value, ok := this.transactionCache.Get(string(txHash.ToArray()))
	if !ok {
		return nil, 0
	}
	txValue := value.(*TransactionCacheaValue)
	return txValue.Tx, txValue.Height
}

//ContainTransaction return whether transaction is in cache
func (this *BlockCache) ContainTransaction(txHash common.Uint256) bool {
	return this.transactionCache.Contains(string(txHash.ToArray()))
}
