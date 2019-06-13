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
	"reflect"
	"testing"

	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"

	"bytes"

	"github.com/ontio/ontology/common"
	vbftcfg "github.com/ontio/ontology/common/config"
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
		t.Errorf("crossShardMsg len not match:%d,%d", crossShardRoot.ToHexString(), msgHash)
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
func TestAddShardConsensusConfig(t *testing.T) {
	shardID := common.NewShardIDUnchecked(1)
	height := 110
	cfg := &shardstates.ShardConfig{
		GasPrice: 20000,
		GasLimit: 10000,
		VbftCfg: &vbftcfg.VBFTConfig{
			N: 1,
			C: 7,
		},
	}
	shardEvent := &shardstates.ConfigShardEvent{
		Height: 120,
		Config: cfg,
	}
	sink := common.ZeroCopySink{}
	shardEvent.Serialization(&sink)
	testCrossShardStore.NewBatch()
	testCrossShardStore.AddShardConsensusConfig(shardID, uint32(height), sink.Bytes())
	err := testCrossShardStore.CommitTo()
	if err != nil {
		t.Errorf("TestAddShardConsensusHeight CommitTo err :%s", err)
		return
	}
	data, err := testCrossShardStore.GetShardConsensusConfig(shardID, uint32(height))
	if err != nil {
		t.Errorf("GetShardConsensusConfig failed shardID:%v,height:%d", shardID, height)
		return
	}
	source := common.NewZeroCopySource(data)
	shardEventInfo := &shardstates.ConfigShardEvent{}
	err = shardEventInfo.Deserialization(source)
	if err != nil {
		t.Errorf("Deserialization failed:%s", err)
		return
	}
	if shardEventInfo.Height != shardEvent.Height {
		t.Errorf("height not match:%d,%d", shardEventInfo.Height, shardEvent.Height)
		return
	}
}

func TestAddShardConsensusHeight(t *testing.T) {
	shardID := common.NewShardIDUnchecked(1)
	heights := []uint32{100, 120, 150}
	testCrossShardStore.NewBatch()
	testCrossShardStore.AddShardConsensusHeight(shardID, heights)
	err := testCrossShardStore.CommitTo()
	if err != nil {
		t.Errorf("TestAddShardConsensusHeight CommitTo err :%s", err)
		return
	}
	blkHeights, err := testCrossShardStore.GetShardConsensusHeight(shardID)
	if err != nil {
		t.Errorf("TestAddShardConsensusHeight failed GetShardConsensusHeight err:%s", err)
		return
	}
	if !reflect.DeepEqual(heights, blkHeights) {
		t.Errorf("TestAddShardConsensusHeight faied heigts:%v,blkHeigts:%v", heights, blkHeights)
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
