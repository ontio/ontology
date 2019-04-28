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
	"os"
	"path"

	"github.com/ontio/ontology/common"
	exec "github.com/ontio/ontology/core/store"
	"github.com/ontio/ontology/core/store/leveldbstore"
	"github.com/ontio/ontology/core/types"
)

const (
	parentHeightKey = "parentHeight"
)

type BlockCacheStore struct {
	shardID    types.ShardID
	dbDir      string
	store      *leveldbstore.LevelDBStore
	execResult map[uint32]exec.ExecuteResult
}

func ResetBlockCacheStore(shardID types.ShardID, dbDir string) (*BlockCacheStore, error) {
	dbPath := path.Join(dbDir, fmt.Sprintf("%s_%d", DBDirBlockCache, shardID.ToUint64()))

	// reset block cache
	os.RemoveAll(dbPath)

	store, err := leveldbstore.NewLevelDBStore(dbPath)
	if err != nil {
		return nil, err
	}
	return &BlockCacheStore{
		shardID:    shardID,
		dbDir:      dbDir,
		store:      store,
		execResult: make(map[uint32]exec.ExecuteResult),
	}, nil
}

func (this *BlockCacheStore) Close() error {
	return this.store.Close()
}

func (this *BlockCacheStore) PutBlock(block *types.Block, stateMerkleRoot common.Uint256) error {
	if this.shardID.ToUint64() != block.Header.ShardID {
		return fmt.Errorf("unmatched shard id: %d vs %d", this.shardID, block.Header.ShardID)
	}
	mklKey := fmt.Sprintf("mkl-%d-%d", this.shardID.ToUint64(), block.Header.Height)
	blkKey := fmt.Sprintf("blk-%d-%d", this.shardID.ToUint64(), block.Header.Height)
	sink := common.NewZeroCopySink(0)
	block.Serialization(sink)
	this.store.Put([]byte(mklKey), stateMerkleRoot[:])
	this.store.Put([]byte(blkKey), sink.Bytes())
	return nil
}

func (this *BlockCacheStore) GetBlock(height uint32) (*types.Block, common.Uint256, error) {
	mklKey := fmt.Sprintf("mkl-%d-%d", this.shardID.ToUint64(), height)
	blkKey := fmt.Sprintf("blk-%d-%d", this.shardID.ToUint64(), height)
	data, err := this.store.Get([]byte(blkKey))
	if err != nil {
		return nil, common.UINT256_EMPTY, err
	}
	blk, err := types.BlockFromRawBytes(data)

	mkl, err := this.store.Get([]byte(mklKey))
	if err != nil {
		return nil, common.UINT256_EMPTY, err
	}

	mklHash, err := common.Uint256ParseFromBytes(mkl)
	return blk, mklHash, err
}

func (this *BlockCacheStore) DelBlock(height uint32) {
	mklKey := fmt.Sprintf("mkl-%d-%d", this.shardID.ToUint64(), height)
	blkKey := fmt.Sprintf("blk-%d-%d", this.shardID.ToUint64(), height)
	this.store.Delete([]byte(mklKey))
	this.store.Delete([]byte(blkKey))
}

func (this *BlockCacheStore) SaveBlockExecuteResult(height uint32, exec exec.ExecuteResult) {
	this.execResult[height] = exec
}

func (this *BlockCacheStore) GetBlockExecuteResult(height uint32) (exec.ExecuteResult, error) {
	if exec, present := this.execResult[height]; present {
		return exec, nil
	}
	return exec.ExecuteResult{}, fmt.Errorf("block execute not found height:%d", height)
}
