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
package utils

import (
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	TestCommon "github.com/ontio/ontology/testsuite/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBlock(t *testing.T) {
	ClearTestChain(t)
	rootShardId := common.NewShardIDUnchecked(0)
	TestCommon.InitConfig(t, rootShardId)
	TestCommon.CreateChain(t, "root", rootShardId, 0)
	blk := GenInitShardAssetBlock(t)
	blk.ShardTxs = make(map[common.ShardID][]*types.CrossShardTxInfos)
	sigData := make(map[uint32][]byte)
	sigData[0] = []byte("123456")
	sigData[1] = []byte("345678")
	hashes := make([]common.Uint256, 0)
	hashes = append(hashes, common.Uint256{1, 2, 3})
	crossShardMsgHash := &types.CrossShardMsgHash{
		ShardMsgHashs: hashes,
		SigData:       sigData,
	}
	crossShardMsgInfo := &types.CrossShardMsgInfo{
		SignMsgHeight:        1111,
		PreCrossShardMsgHash: common.Uint256{1, 2, 3},
		Index:                2,
		ShardMsgInfo:         crossShardMsgHash,
	}
	shard1 := common.NewShardIDUnchecked(1)
	blk.ShardTxs[shard1] = []*types.CrossShardTxInfos{{
		ShardMsg: crossShardMsgInfo,
		Tx:       blk.Transactions[0],
	}}
	sink := common.NewZeroCopySink(0)
	blk.Serialization(sink)
	source := common.NewZeroCopySource(sink.Bytes())
	newBlk := &types.Block{}
	err := newBlk.Deserialization(source)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, blk.ShardTxs[shard1], newBlk.ShardTxs[shard1])
	for i, tx := range blk.Transactions {
		assert.Equal(t, tx, newBlk.Transactions[i])
	}
}
