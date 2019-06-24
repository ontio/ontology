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
package ledgerstore

import (
	"bytes"
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
)

func TestSaveCrossShardMsgByHash(t *testing.T) {
	crossShardRoot := common.Uint256{}
	crossShardMsgInfo := &types.CrossShardMsgInfo{
		FromShardID:       common.NewShardIDUnchecked(2),
		MsgHeight:         109,
		SignMsgHeight:     1111,
		CrossShardMsgRoot: crossShardRoot,
	}
	crossShardMsg := &types.CrossShardMsg{
		CrossShardMsgInfo: crossShardMsgInfo,
	}
	testCrossShardStore.NewBatch()
	testCrossShardStore.SaveCrossShardMsgByHash(crossShardRoot, crossShardMsg)
	err := testCrossShardStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo err:%s", err)
		return
	}
	msg, err := testCrossShardStore.GetCrossShardMsgByHash(crossShardRoot)
	if err != nil {
		t.Errorf("getCrossShardMsgKeyByShard failed,crossmsghash:%v,err:%s", crossShardRoot.ToHexString(), err)
		return
	}
	msgHash := msg.CrossShardMsgInfo.CrossShardMsgRoot.ToHexString()
	if crossShardRoot.ToHexString() != msgHash {
		t.Errorf("crossShardMsg len not match:%s, %s", crossShardRoot.ToHexString(), msgHash)
		return
	}
}

func TestSaveAllShardIDs(t *testing.T) {
	shardIds := make([]common.ShardID, 0)
	shardIds = append(shardIds, common.NewShardIDUnchecked(1))
	shardIds = append(shardIds, common.NewShardIDUnchecked(2))
	testCrossShardStore.SaveAllShardIDs(shardIds)
	err := testCrossShardStore.CommitTo()
	if err != nil {
		t.Errorf("TestSaveAllShardIDs CommitTo err :%s", err)
		return
	}
	data, err := testCrossShardStore.GetAllShardIDs()
	if err != nil {
		t.Errorf("GetAllShardIDs err:%s", err)
		return
	}
	if len(shardIds) != len(data) {
		t.Errorf("shardId len not match")
		return
	}
}

func TestSaveCrossShardHash(t *testing.T) {
	shardID := common.NewShardIDUnchecked(1)
	msgHash := common.Uint256{1, 2, 3}
	testCrossShardStore.NewBatch()
	testCrossShardStore.SaveCrossShardHash(shardID, msgHash)
	err := testCrossShardStore.CommitTo()
	if err != nil {
		t.Errorf("TestSaveCrossShardHash CommitTo err :%s", err)
		return
	}
	hash, err := testCrossShardStore.GetCrossShardHash(shardID)
	if err != nil {
		t.Errorf("GetCrossShardHash shardID:%v,err:%v", shardID, err)
	}
	if bytes.Compare(msgHash[:], hash[:]) != 0 {
		t.Errorf("msg hash not match")
	}
}

func TestSaveShardMsgHash(t *testing.T) {
	shardID := common.NewShardIDUnchecked(1)
	msgHash := common.Uint256{1, 2, 3}
	testCrossShardStore.NewBatch()
	testCrossShardStore.SaveShardMsgHash(shardID, msgHash)
	err := testCrossShardStore.CommitTo()
	if err != nil {
		t.Errorf("TestSaveShardMsgHash CommitTo err :%s", err)
		return
	}
	hash, err := testCrossShardStore.GetShardMsgHash(shardID)
	if bytes.Compare(msgHash[:], hash[:]) != 0 {
		t.Errorf("msg hash not match")
	}
}
