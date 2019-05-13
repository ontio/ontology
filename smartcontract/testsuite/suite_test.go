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
package testsuite

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/chainmgr/xshard_state"
	"github.com/ontio/ontology/core/xshard_types"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/neovm"
	"github.com/stretchr/testify/assert"
)

func TestRemoteNotifyPing(t *testing.T) {
	shardAContract := RandomAddress()
	InstallNativeContract(shardAContract, map[string]native.Handler{
		"remoteNotifyPing": RemoteNotifyPing,
		"handlePing":       HandlePing,
	})

	shardContext := NewShardContext(common.NewShardIDUnchecked(1), shardAContract, t)
	txHash, notify := shardContext.InvokeShardContract("remoteNotifyPing", []interface{}{""})

	sink := common.NewZeroCopySink(10)
	sink.WriteString(fmt.Sprintf("hello from shard: %d", shardContext.shardID.ToUint64()))
	expected := &xshard_types.XShardNotify{
		ShardMsgHeader: xshard_types.ShardMsgHeader{
			SourceShardID: shardContext.shardID,
			TargetShardID: common.NewShardIDUnchecked(2),
			SourceTxHash:  txHash,
		},
		NotifyID: 0,
		Fee:      neovm.MIN_TRANSACTION_GAS,
		Method:   "handlePing",
		Contract: shardAContract,
		Args:     sink.Bytes(),
	}

	assert.Equal(t, len(notify.ShardMsg), 1)
	notifyMsg, ok := notify.ShardMsg[0].(*xshard_types.XShardNotify)
	assert.True(t, ok)
	expected.Fee = notifyMsg.Fee
	assert.Equal(t, expected, notifyMsg)
	t.Logf("notify fee is %d", notifyMsg.Fee)
}

func TestLedgerRemoteInvokeAdd(t *testing.T) {
	shardAContract := RandomAddress()
	method := "remoteAddAndInc"
	InstallNativeContract(shardAContract, map[string]native.Handler{
		method: RemoteInvokeAddAndInc,
	})

	shardContext := NewShardContext(common.NewShardIDUnchecked(1), shardAContract, t)
	txHash, notify := shardContext.InvokeShardContract(method, []interface{}{""})

	shardTxID := xshard_state.ShardTxID(string(txHash[:]))
	state, err := shardContext.GetXShardState(shardTxID)
	assert.Nil(t, err)
	assert.NotNil(t, state.PendingReq)
	sink := common.NewZeroCopySink(10)
	sink.WriteUint64(2)
	sink.WriteUint64(3)
	expected := &xshard_types.XShardTxReq{
		ShardMsgHeader: xshard_types.ShardMsgHeader{
			TargetShardID: ShardB,
			SourceTxHash:  txHash,
			SourceShardID: shardContext.shardID,
		},
		IdxInTx:  0,
		Fee:      neovm.MIN_TRANSACTION_GAS,
		Contract: shardAContract,
		Method:   "handlePing",
		Args:     sink.Bytes(),
	}

	reqMsg, ok := notify.ShardMsg[0].(*xshard_types.XShardTxReq)
	assert.True(t, ok)
	expected.Fee = reqMsg.Fee
	assert.Equal(t, expected, reqMsg)
	t.Logf("req fee is %d", reqMsg.Fee)

	sink.Reset()
	sink.WriteUint64(5)
	rep := &xshard_types.XShardTxRsp{
		IdxInTx: expected.IdxInTx,
		Error:   false,
		Result:  sink.Bytes(),
	}

	rep.SourceTxHash = txHash
	msgs := []xshard_types.CommonShardMsg{rep}
	shardContext.HandleShardCallMsgs(msgs)
	sink.Reset()
	sink.WriteUint64(6)

	state, err = shardContext.GetXShardState(shardTxID)
	assert.Nil(t, err)
	res, _ := json.Marshal(state.Notify.Notify[0].States)
	buf, _ := json.Marshal(sink.Bytes())
	assert.Equal(t, string(res), string(buf))

	commit := &xshard_types.XShardCommitMsg{}
	commit.SourceTxHash = txHash

	msgs = []xshard_types.CommonShardMsg{commit}
	notify = shardContext.HandleShardCallMsgs(msgs)
	sink.Reset()
	sink.WriteUint64(6)

	assert.Nil(t, err)
	res, _ = json.Marshal(notify.ContractEvent.Notify[0].States)
	buf, _ = json.Marshal(sink.Bytes())
	assert.Equal(t, string(res), string(buf))
}

// test shard transaction mode:
// shard0 -> invoke shard1
//        -> invoke shard2
//        -> ...
func TestShardReverseBytes(t *testing.T) {
	contract := RandomAddress()
	method := "shardReverseBytes"
	InstallNativeContract(contract, map[string]native.Handler{
		method: ShardReverseBytes,
	})

	shards := make(map[common.ShardID]*ShardContext, 3)
	shard0 := common.NewShardIDUnchecked(0)
	shard1 := common.NewShardIDUnchecked(1)
	shard2 := common.NewShardIDUnchecked(2)

	shards[shard0] = NewShardContext(shard0, contract, t)
	shards[shard1] = NewShardContext(shard1, contract, t)
	shards[shard2] = NewShardContext(shard2, contract, t)

	// shard0 -> invoke shard1
	//        -> invoke shard2
	args := common.SerializeToBytes(&ReverseStringParam{Origin: []byte("1234567"), Shards: []common.ShardID{shard1, shard2}})
	totalShardMsg := RunShardTxToComplete(shards, shard0, method, args)
	// 2 req, 2 rep, 2 prep, 2 preped, 2 commit = 10
	assert.Equal(t, 10, totalShardMsg)

	// shard0 -> invoke shard1
	//        -> invoke shard2
	//        -> invoke shard1
	args = common.SerializeToBytes(&ReverseStringParam{Origin: []byte("1234567"), Shards: []common.ShardID{shard1, shard2, shard1}})
	totalShardMsg = RunShardTxToComplete(shards, shard0, method, args)
	// 3 req, 3 rep, 2 prep, 2 preped, 2 commit = 12
	assert.Equal(t, 12, totalShardMsg)
}
