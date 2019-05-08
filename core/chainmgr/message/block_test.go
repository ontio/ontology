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

package message_test

import (
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/chainmgr/message"
	"github.com/ontio/ontology/core/types"
)

func newTestBlock() *types.Block {
	hdr := &types.Header{}
	hdr.Version = common.VERSION_SUPPORT_SHARD
	hdr.Bookkeepers = make([]keypair.PublicKey, 0)
	hdr.SigData = make([][]byte, 0)

	return &types.Block{Header: hdr}
}

func newTestShardBlockInfo(t *testing.T) *message.ShardBlockInfo {
	height := uint32(123)
	parentHeight := uint32(321)
	shardBlk := newTestBlock()
	shardBlk.Header.Height = height
	shardBlk.Header.ParentHeight = parentHeight

	blkInfo := &message.ShardBlockInfo{
		FromShardID: common.NewShardIDUnchecked(100),
		Height:      uint32(height),
		Block:       shardBlk,
	}

	return blkInfo
}

func TestShardBlockHeaderSerialize(t *testing.T) {
	height := uint32(123)
	parentHeight := uint32(321)

	shardBlk := newTestBlock()
	shardBlk.Header.Height = height
	shardBlk.Header.ParentHeight = parentHeight
	sink := common.NewZeroCopySink(0)
	shardBlk.Serialization(sink)
	bs := sink.Bytes()

	source := common.NewZeroCopySource(bs)
	shardBlk2 := new(types.Block)
	if err := shardBlk2.Deserialization(source); err != nil {
		t.Fatalf("deser shard header: %s", err)
	}

	if shardBlk2.Header.ParentHeight != parentHeight {
		t.Fatalf("unmatched parent height: %d vs %d", shardBlk2.Header.ParentHeight, parentHeight)
	}

	if shardBlk2.Header.Height != height {
		t.Fatalf("unmatched height: %d vs %d", shardBlk2.Header.Height, height)
	}
}

func TestShardBlockPool(t *testing.T) {
	pool := message.NewShardBlockPool(100)
	blk := newTestShardBlockInfo(t)

	shardID := blk.FromShardID
	height := blk.Height

	if err := pool.AddBlockInfo(blk); err != nil {
		t.Fatalf("failed add block: %s", err)
	}

	blk2 := pool.GetBlockInfo(shardID, height)
	if blk2 == nil {
		t.Fatalf("failed get block")
	}

	if blk != blk2 {
		t.Fatalf("unmatched blk")
	}
}
