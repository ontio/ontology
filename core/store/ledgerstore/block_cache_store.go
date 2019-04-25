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
	"strconv"

	"github.com/ontio/ontology/common"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/leveldbstore"
	"github.com/ontio/ontology/core/types"
)

const (
	parentHeightKey = "parentHeight"
)

type BlockCacheStore struct {
	shardID types.ShardID
	dbDir   string
	store   *leveldbstore.LevelDBStore
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
		shardID: shardID,
		dbDir:   dbDir,
		store:   store,
	}, nil
}

func (this *BlockCacheStore) PutBlock(block *types.Block, stateMerkleRoot common.Uint256) error {
	if this.shardID.ToUint64() != block.Header.ShardID {
		return fmt.Errorf("unmatched shard id: %d vs %d", this.shardID, block.Header.ShardID)
	}
	currentHeight, err := this.GetCurrentParentHeight()
	if err != nil {
		return fmt.Errorf("PutBlockHeight err:%s", err)
	}
	if block.Header.Height <= currentHeight {
		return nil
	}
	if currentHeight+1 != block.Header.Height && currentHeight != 0 {
		return fmt.Errorf("block height %d not equal next block height %d", currentHeight, block.Header.Height)
	}
	this.PutBlockHeight(block.Header.Height)
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

func (this *BlockCacheStore) PutBlockHeight(height uint32) {
	parentHeight := fmt.Sprintf("%d", height)
	this.store.Delete([]byte(parentHeight))
	this.store.Put([]byte(parentHeightKey), []byte(parentHeight))
}

func (this *BlockCacheStore) GetCurrentParentHeight() (uint32, error) {
	parentHeight, err := this.store.Get([]byte(parentHeightKey))
	if err != nil {
		if err == scommon.ErrNotFound {
			return 0, nil
		}
		return 0, err
	}
	height, err := strconv.Atoi(string(parentHeight))
	if err != nil {
		return 0, err
	}
	return uint32(height), nil
}
