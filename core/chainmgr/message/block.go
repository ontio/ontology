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
	"encoding/json"
	"fmt"
	"github.com/ontio/ontology/events/message"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/types"
)

const (
	ShardBlockNew = iota
	ShardBlockReceived
	ShardBlockProcessed
)

//
// Marshal-Helper for Block
//
type ShardBlockHeader struct {
	Header *types.Header
}

func (this *ShardBlockHeader) Serialize(sink *common.ZeroCopySink) error {
	return this.Header.Serialization(sink)
}

func (this *ShardBlockHeader) Deserialize(sink *common.ZeroCopySource) error {
	return this.Header.Deserialization(sink)
}

//
// Marshal-Helper for transaction
//
type ShardBlockTx struct {
	Tx *types.Transaction
}

//
// ShardBlockInfo contains:
//  .Header: block header
//  .ShardTxs: Cross-Shard Tx from the block
//  .Events: shard events generated from the block (only for local block)
//
type ShardBlockInfo struct {
	FromShardID types.ShardID                   `json:"from_shard_id"`
	Height      uint32                          `json:"height"`
	State       uint                            `json:"state"`
	Header      *ShardBlockHeader               `json:"header"`
	ShardTxs    map[types.ShardID]*ShardBlockTx `json:"shard_txs"` // indexed by ToShardID
	Events      []*message.ShardEventState
}

type shardBlkMarshalHelper struct {
	Payload []byte `json:"payload"`
}

func (this *ShardBlockHeader) MarshalJSON() ([]byte, error) {
	sink := common.NewZeroCopySink(0)
	if this.Header != nil {
		if err := this.Header.Serialization(sink); err != nil {
			return nil, fmt.Errorf("shard block hdr marshal: %s", err)
		}
	}

	return json.Marshal(&shardBlkMarshalHelper{
		Payload: sink.Bytes(),
	})
}

func (this *ShardBlockHeader) UnmarshalJSON(data []byte) error {
	helper := &shardBlkMarshalHelper{}
	if err := json.Unmarshal(data, helper); err != nil {
		return fmt.Errorf("shard block hdr helper: %s", err)
	}

	if len(helper.Payload) > 0 {
		hdr := &types.Header{}
		if err := hdr.Deserialization(common.NewZeroCopySource(helper.Payload)); err != nil {
			return fmt.Errorf("shard block hdr unmarshal: %s", err)
		}
		this.Header = hdr
	}
	return nil
}

func (this *ShardBlockTx) MarshalJSON() ([]byte, error) {
	sink := common.NewZeroCopySink(0)
	if this.Tx != nil {
		if err := this.Tx.Serialization(sink); err != nil {
			return nil, fmt.Errorf("shard block tx marshal: %x", err)
		}
	}

	return json.Marshal(&shardBlkMarshalHelper{
		Payload: sink.Bytes(),
	})
}

func (this *ShardBlockTx) UnmarshalJSON(data []byte) error {
	helper := &shardBlkMarshalHelper{}
	if err := json.Unmarshal(data, helper); err != nil {
		return fmt.Errorf("shard block tx helper: %s", err)
	}

	if len(helper.Payload) > 0 {
		tx := &types.Transaction{Raw: helper.Payload}
		if err := tx.Deserialization(common.NewZeroCopySource(helper.Payload)); err != nil {
			return fmt.Errorf("shard block tx unmarshal: %s", err)
		}
		this.Tx = tx
	}
	return nil
}

func (this *ShardBlockInfo) Serialize(w io.Writer) error {
	return SerJson(w, this)
}

func (this *ShardBlockInfo) Deserialize(r io.Reader) error {
	return DesJson(r, this)
}

////////////////////////////////////
//
//  shard block pool
//
////////////////////////////////////

type ShardBlockMap map[uint32]*ShardBlockInfo // indexed by BlockHeight

type ShardBlockPool struct {
	Shards      map[types.ShardID]ShardBlockMap // indexed by FromShardID
	MaxBlockCap uint32
}

func NewShardBlockPool(historyCap uint32) *ShardBlockPool {
	return &ShardBlockPool{
		Shards:      make(map[types.ShardID]ShardBlockMap),
		MaxBlockCap: historyCap,
	}
}

func (pool *ShardBlockPool) GetBlockInfo(shardID types.ShardID, height uint32) *ShardBlockInfo {
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
		if blk.Header != nil &&
			bytes.Compare(blk.Header.Header.BlockRoot[:], blkInfo.Header.Header.BlockRoot[:]) == 0 {
			return fmt.Errorf("add shard block, dup blk")
		}

		// replace events
		blkInfo.Events = blk.Events
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

////////////////////////////////////
//
//  json helpers
//
////////////////////////////////////

func SerJson(w io.Writer, v interface{}) error {
	buf, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("json marshal failed: %s", err)
	}

	if err := serialization.WriteVarBytes(w, buf); err != nil {
		return fmt.Errorf("json serialize write failed: %s", err)
	}
	return nil
}

func DesJson(r io.Reader, v interface{}) error {
	buf, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("json deserialize read failed: %s", err)
	}
	if err := json.Unmarshal(buf, v); err != nil {
		return fmt.Errorf("json unmarshal failed: %s", err)
	}
	return nil
}
