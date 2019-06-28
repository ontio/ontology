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
	"math/rand"
	"reflect"
	"testing"

	"github.com/ontio/ontology/common"
	vbftcfg "github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/payload"
	msg "github.com/ontio/ontology/events/message"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
)

func TestSaveContractMetaData(t *testing.T) {
	var addr common.Address
	rand.Read(addr[:])
	metaDataCode := &payload.MetaDataCode{
		OntVersion: 1,
		Contract:   addr,
		ShardId:    1,
	}
	metaDataEvent := &msg.MetaDataEvent{
		Height:   123,
		MetaData: metaDataCode,
	}
	testEventStore.NewBatch()
	err := testEventStore.SaveContractMetaDataEvent(metaDataEvent.Height, metaDataEvent.MetaData)
	if err != nil {
		t.Errorf("SaveContractMetaDataEvent err :%s", err)
		return
	}
	err = testEventStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo err :%s", err)
		return
	}
	msg, err := testEventStore.GetContractMetaDataEvent(metaDataEvent.Height, metaDataCode.Contract)
	if err != nil {
		t.Errorf("GetMeteEvent err:%s", err)
	}
	if metaDataCode.ShardId != msg.ShardId {
		t.Errorf("metaData shardId:%d not match msg shardID:%d", metaDataCode.ShardId, msg.ShardId)
	}
}

func TestSaveDeployCode(t *testing.T) {
	var addr common.Address
	rand.Read(addr[:])
	deployCode := &payload.DeployCode{
		NeedStorage: true,
		Name:        "code",
	}
	contractEvent := &msg.ContractLifetimeEvent{
		DeployHeight: 123,
		Contract:     deployCode,
	}
	testEventStore.NewBatch()
	err := testEventStore.SaveContractEvent(contractEvent)
	if err != nil {
		t.Errorf("SaveContractEvent err:%s", err)
		return
	}
	err = testEventStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo err :%s", err)
		return
	}
	msg, err := testEventStore.GetContractEvent(deployCode.Address())
	if err != nil {
		t.Errorf("GetMeteEvent err:%s", err)
	}
	if deployCode.Name != msg.Contract.Name {
		t.Errorf("metaData name:%s not match msg name:%s", deployCode.Name, msg.Contract.Name)
	}
}

func TestAddShardConsensusHeight(t *testing.T) {
	shardID := common.NewShardIDUnchecked(1)
	heights := []uint32{100, 120, 150}
	testEventStore.NewBatch()
	testEventStore.AddShardConsensusHeight(shardID, heights)
	err := testEventStore.CommitTo()
	if err != nil {
		t.Errorf("TestAddShardConsensusHeight CommitTo err :%s", err)
		return
	}
	blkHeights, err := testEventStore.GetShardConsensusHeight(shardID)
	if err != nil {
		t.Errorf("TestAddShardConsensusHeight failed GetShardConsensusHeight err:%s", err)
		return
	}
	if !reflect.DeepEqual(heights, blkHeights) {
		t.Errorf("TestAddShardConsensusHeight faied heigts:%v,blkHeigts:%v", heights, blkHeights)
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
	testEventStore.NewBatch()
	testEventStore.AddShardConsensusConfig(shardID, uint32(height), sink.Bytes())
	err := testEventStore.CommitTo()
	if err != nil {
		t.Errorf("TestAddShardConsensusHeight CommitTo err :%s", err)
		return
	}
	data, err := testEventStore.GetShardConsensusConfig(shardID, uint32(height))
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

func TestSaveContractMetaHeights(t *testing.T) {
	var addr common.Address
	rand.Read(addr[:])
	heights := []uint32{100, 120, 150}
	testEventStore.NewBatch()
	testEventStore.SaveContractMetaHeights(addr, heights)
	err := testEventStore.CommitTo()
	if err != nil {
		t.Errorf("TestSaveContractMetaHeights CommitTo err :%s", err)
		return
	}
	blkHeights, err := testEventStore.GetContractMetaHeights(addr)
	if err != nil {
		t.Errorf("TestSaveContractMetaHeights failed GetContractMetaHeights err:%s", err)
		return
	}
	if !reflect.DeepEqual(heights, blkHeights) {
		t.Errorf("TestSaveContractMetaHeights faied heigts:%v,blkHeigts:%v", heights, blkHeights)
		return
	}
}
