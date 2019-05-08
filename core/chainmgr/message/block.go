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

package message

import (
	"bytes"
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/events/message"
)

const (
	ShardBlockNew = iota
	ShardBlockReceived
	ShardBlockProcessed
)

//
// Marshal-Helper for transaction
//
type ShardBlockTx struct {
	Tx *types.Transaction
}

func (this *ShardBlockTx) Serialization(sink *common.ZeroCopySink) {
	this.Tx.Serialization(sink)
}

func (this *ShardBlockTx) Deserialization(source *common.ZeroCopySource) error {
	this.Tx = &types.Transaction{}
	return this.Tx.Deserialization(source)
}

//
// ShardBlockInfo contains:
//  .Block: block
//  .ShardTxs: Cross-Shard Tx from the block
//  .Events: shard events generated from the block (only for local block)
//
type ShardBlockInfo struct {
	FromShardID common.ShardID                   `json:"from_shard_id"`
	Height      uint32                           `json:"height"`
	State       uint                             `json:"state"`
	Block       *types.Block                     `json:"block"`
	ShardTxs    map[common.ShardID]*ShardBlockTx `json:"shard_txs"` // indexed by ToShardID
}

func (this *ShardBlockInfo) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteUint64(this.FromShardID.ToUint64())
	sink.WriteUint32(this.Height)
	sink.WriteUint64(uint64(this.State))
	this.Block.Serialization(sink)
	return nil
}

func (this *ShardBlockInfo) Deserialization(source *common.ZeroCopySource) error {
	fromShard, eof := source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	id, err := common.NewShardID(fromShard)
	if err != nil {
		return fmt.Errorf("deserialization: generate from shard id failed, err: %s", err)
	}
	this.FromShardID = id
	this.Height, eof = source.NextUint32()
	state, eof := source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.State = uint(state)
	this.Block = &types.Block{}
	if err := this.Block.Deserialization(source); err != nil {
		return fmt.Errorf("deserialization: read header failed, err: %s", err)
	}
	eventNum, eof := source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	for i := uint64(0); i < eventNum; i++ {
		evt := &message.ShardEventState{}
		if err := evt.Deserialization(source); err != nil {
			return fmt.Errorf("deserialization: read event failed, index %d, err: %s", i, err)
		}
	}
	return nil
}

////////////////////////////////////
//
//  shard block pool
//
////////////////////////////////////

type ShardBlockMap map[uint32]*ShardBlockInfo // indexed by BlockHeight

type ShardBlockPool struct {
	Shards      map[common.ShardID]ShardBlockMap // indexed by FromShardID
	MaxBlockCap uint32
}

func NewShardBlockPool(historyCap uint32) *ShardBlockPool {
	return &ShardBlockPool{
		Shards:      make(map[common.ShardID]ShardBlockMap),
		MaxBlockCap: historyCap,
	}
}

func (pool *ShardBlockPool) GetBlockInfo(shardID common.ShardID, height uint32) *ShardBlockInfo {
	if m, present := pool.Shards[shardID]; present && m != nil {
		return m[height]
	}
	return nil
}

func (pool *ShardBlockPool) AddBlockInfo(blkInfo *ShardBlockInfo) error {
	if _, present := pool.Shards[blkInfo.FromShardID]; !present {
		pool.Shards[blkInfo.FromShardID] = make(ShardBlockMap)
	}

	m := pool.Shards[blkInfo.FromShardID]
	if m == nil {
		return fmt.Errorf("add shard block, nil map")
	}
	if blk, present := m[blkInfo.Height]; present {
		if blk.State != ShardBlockNew {
			return fmt.Errorf("add shard block, new block on block state %d", blk.State)
		}
		hdr := blk.Block.Header
		if hdr != nil && bytes.Compare(hdr.BlockRoot[:], blkInfo.Block.Header.BlockRoot[:]) == 0 {
			return fmt.Errorf("add shard block, dup blk")
		}
	}

	log.Infof("chainmgr AddBlock from shard %d, block %d", blkInfo.FromShardID, blkInfo.Height)
	m[blkInfo.Height] = blkInfo

	// if too much block cached in map, drop old blocks
	if uint32(len(m)) < pool.MaxBlockCap {
		return nil
	}
	h := blkInfo.Height
	for _, blk := range m {
		if blk.Height > h {
			h = blk.Height
		}
	}

	toDrop := make([]uint32, 0)
	for _, blk := range m {
		if blk.Height < h-uint32(pool.MaxBlockCap) {
			toDrop = append(toDrop, blk.Height)
		}
	}
	for _, blkHeight := range toDrop {
		delete(m, blkHeight)
	}

	return nil
}
