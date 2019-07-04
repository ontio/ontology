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

	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/chainmgr/message"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/core/xshard_types"
	"github.com/stretchr/testify/assert"
)

func newTestShardMsg(t *testing.T) *types.CrossShardMsg {
	shardMsg := []xshard_types.CommonShardMsg{&xshard_types.XShardCommitMsg{
		ShardMsgHeader: xshard_types.ShardMsgHeader{
			SourceShardID: common.NewShardIDUnchecked(1),
			TargetShardID: common.NewShardIDUnchecked(0),
			SourceTxHash:  common.Uint256{1, 2, 3},
			ShardTxID:     "2",
		},
	}}
	sigData := make(map[uint32][]byte)
	sigData[0] = []byte("123456")
	sigData[1] = []byte("345678")
	hashes := make([]common.Uint256, 0)
	hashes = append(hashes, common.Uint256{1, 2, 3})
	crossShardMsgHash := &types.CrossShardMsgHash{
		ShardMsgHashs: hashes,
		SigData:       sigData,
	}
	crossShardMsg := &types.CrossShardMsg{
		CrossShardMsgInfo: &types.CrossShardMsgInfo{
			SignMsgHeight:        uint32(100),
			PreCrossShardMsgHash: common.Uint256{},
			Index:                1,
			ShardMsgInfo:         crossShardMsgHash,
		},
		ShardMsg: shardMsg,
	}
	return crossShardMsg
}

func TestAddShardInfo(t *testing.T) {
	shardID := common.NewShardIDUnchecked(0)
	InitCrossShardPool(shardID, 10)
	shardID = common.NewShardIDUnchecked(1)
	ldg, err := ledger.NewLedger(config.DEFAULT_DATA_DIR, 0)
	if err != nil {
		t.Errorf("failed to new ledger err:%s", err)
		return
	}
	addShardInfo(ldg, shardID)
	shardInfo := GetShardInfo()
	if shardId, present := shardInfo[shardID]; !present {
		t.Errorf("shardID not found:%v", shardID)
	} else {
		t.Logf("shardId found:%v", shardId)
	}
	ldg.Close()
}

func TestCrossShardTxInfos_Serialize(t *testing.T) {
	acc := account.NewAccount("")
	if acc == nil {
		t.Fatalf("failed to new account")
	}
	shardMsg := []xshard_types.CommonShardMsg{&xshard_types.XShardCommitMsg{
		ShardMsgHeader: xshard_types.ShardMsgHeader{
			SourceShardID: common.NewShardIDUnchecked(1),
			TargetShardID: common.NewShardIDUnchecked(0),
			SourceTxHash:  common.Uint256{1, 2, 3},
			ShardTxID:     "2",
		},
	}}
	sigData := make(map[uint32][]byte)
	sigData[0] = []byte("123456")
	sigData[1] = []byte("345678")
	hashes := make([]common.Uint256, 0)
	hashes = append(hashes, common.Uint256{1, 2, 3})
	crossShardMsgHash := &types.CrossShardMsgHash{
		ShardMsgHashs: hashes,
		SigData:       sigData,
	}
	tx, err := message.NewCrossShardTxMsg(acc, 100, common.NewShardIDUnchecked(10), 500, 20000, shardMsg)
	if err != nil {
		t.Fatalf("failed to build cross shard tx: %s", err)
	}
	crossShardTxInfos := &types.CrossShardTxInfos{
		ShardMsg: &types.CrossShardMsgInfo{
			SignMsgHeight:        uint32(100),
			PreCrossShardMsgHash: common.Uint256{1, 2, 3},
			Index:                1,
			ShardMsgInfo:         crossShardMsgHash,
		},
		Tx: tx,
	}
	sink := common.NewZeroCopySink(0)
	crossShardTxInfos.Serialization(sink)
	msg := sink.Bytes()
	var sharTxInfos types.CrossShardTxInfos
	source := common.NewZeroCopySource(msg)
	err = sharTxInfos.Deserialization(source)
	assert.Nil(t, err)
	if crossShardTxInfos.ShardMsg.Index != sharTxInfos.ShardMsg.Index {
		t.Errorf("cross shardTx info index:%d not equal index:%d", crossShardTxInfos.ShardMsg.Index, sharTxInfos.ShardMsg.Index)
	}
}
