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

package xshard

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

func TestShardBlockPool(t *testing.T) {
	InitShardBlockPool(common.NewShardIDUnchecked(1), 100)
	blk := newTestShardBlockInfo(t)

	shardID := blk.FromShardID
	height := blk.Height

	if err := AddBlockInfo(blk); err != nil {
		t.Fatalf("failed add block: %s", err)
	}

	blk2 := GetBlockInfo(shardID, height)
	if blk2 == nil {
		t.Fatalf("failed get block")
	}

	if blk != blk2 {
		t.Fatalf("unmatched blk")
	}
}
