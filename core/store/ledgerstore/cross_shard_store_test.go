package ledgerstore

import (
	"reflect"
	"testing"

	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/service/native/utils"

	"github.com/ontio/ontology/common"
	vbftcfg "github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/types"
)

func TestSaveCrossShardMsgByShardID(t *testing.T) {
	shardID := common.NewShardIDUnchecked(1)
	shardMsg := &types.CrossShardMsg{
		FromShardID:   common.NewShardIDUnchecked(2),
		MsgHeight:     109,
		SignMsgHeight: 1111,
	}
	crossShardTxInfo := &types.CrossShardTxInfos{
		ShardMsg: shardMsg,
		Tx:       nil,
	}
	crossShardTxInfos := make([]*types.CrossShardTxInfos, 0)
	crossShardTxInfos = append(crossShardTxInfos, crossShardTxInfo)
	testCrossShardStore.NewBatch()
	testCrossShardStore.SaveCrossShardMsgByShardID(shardID, crossShardTxInfos)
	err := testCrossShardStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo err:%s", err)
		return
	}
	infos, err := testCrossShardStore.GetCrossShardMsgByShardID(shardID)
	if err != nil {
		t.Errorf("getCrossShardMsgKeyByShard failed,shardID:%v,err:%s", shardID, err)
		return
	}
	if len(infos) != len(crossShardTxInfos) {
		t.Errorf("crossShardMsg len not match:%d,%d", len(crossShardTxInfos), len(infos))
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
		ShardChangeView: &utils.ChangeView{
			View:   3,
			Height: 109,
		},
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
	heights := make([]uint32, 0)
	blkNums := []uint32{100, 120, 150}
	heights = append(heights, blkNums...)
	value := common.NewZeroCopySink(16)
	value.WriteUint32(uint32(len(heights)))
	for _, number := range heights {
		value.WriteUint32(number)
	}
	testCrossShardStore.NewBatch()
	testCrossShardStore.AddShardConsensusHeight(shardID, value.Bytes())
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
