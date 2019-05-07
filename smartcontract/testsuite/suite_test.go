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
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/chainmgr/xshard_state"
	"github.com/ontio/ontology/core/store/ledgerstore"
	types2 "github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/core/xshard_types"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/neovm"
	"github.com/ontio/ontology/smartcontract/storage"
	"github.com/ontio/ontology/vm/neovm/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRemoteNotifyPing(t *testing.T) {
	shardAContract := RandomAddress()
	InstallNativeContract(shardAContract, map[string]native.Handler{
		"remoteNotifyPing": RemoteNotifyPing,
		"handlePing":       HandlePing,
	})

	tx := BuildInvokeTx(shardAContract, "remoteNotifyPing", []interface{}{""})
	assert.NotNil(t, tx)

	overlay := NewOverlayDB()
	cache := storage.NewCacheDB(overlay)

	state, _, err := executeTransaction(tx, cache)

	assert.Nil(t, err)
	assert.Equal(t, len(state.ShardNotifies), 1)
	notify := state.ShardNotifies[0]
	sink := common.NewZeroCopySink(10)
	sink.WriteString(fmt.Sprintf("hello from shard: %d", tx.ShardID))
	expected := &xshard_types.XShardNotify{
		ShardMsgHeader: xshard_types.ShardMsgHeader{
			SourceShardID: common.NewShardIDUnchecked(tx.ShardID),
			TargetShardID: common.NewShardIDUnchecked(2),
			SourceTxHash:  tx.Hash(),
		},
		NotifyID: 0,
		Payer:    tx.Payer,
		Fee:      neovm.MIN_TRANSACTION_GAS,
		Method:   "handlePing",
		Args:     sink.Bytes(),
	}

	assert.Equal(t, expected, notify)
}

func TestRemoteInvokeAdd(t *testing.T) {
	shardAContract := RandomAddress()
	method := "remoteAddAndInc"
	InstallNativeContract(shardAContract, map[string]native.Handler{
		method: RemoteInvokeAddAndInc,
	})

	tx := BuildInvokeTx(shardAContract, method, []interface{}{""})
	assert.NotNil(t, tx)

	overlay := NewOverlayDB()
	cache := storage.NewCacheDB(overlay)

	state, _, err := executeTransaction(tx, cache)

	//assert.Equal(t, shardsysmsg.ErrYield, err) // error is wrapped
	assert.NotNil(t, err)
	assert.NotNil(t, state.PendingReq)
	sink := common.NewZeroCopySink(10)
	sink.WriteUint64(2)
	sink.WriteUint64(3)
	expected := &xshard_types.XShardTxReq{
		ShardMsgHeader: xshard_types.ShardMsgHeader{
			TargetShardID: ShardB,
			SourceTxHash:  tx.Hash(),
		},
		IdxInTx: 0,
		Payer:   tx.Payer,
		Fee:     neovm.MIN_TRANSACTION_GAS,
		Method:  "handlePing",
		Args:    sink.Bytes(),
	}

	assert.Equal(t, expected, state.PendingReq)
	hs := tx.Hash()
	shardTxID := xshard_state.ShardTxID(string(hs[:]))
	xshard_state.PutTxState(shardTxID, state)

	sink.Reset()
	sink.WriteUint64(5)
	rep := &xshard_types.XShardTxRsp{
		IdxInTx: expected.IdxInTx,
		Error:   false,
		Result:  sink.Bytes(),
	}

	state, res, err := resumeTx(shardTxID, rep)
	assert.Nil(t, err)
	sink.Reset()
	sink.WriteUint64(6)

	assert.Equal(t, res.(*types.ByteArray), types.NewByteArray(sink.Bytes()))
}

func TestLedgerRemoteInvokeAdd(t *testing.T) {
	shardAContract := RandomAddress()
	method := "remoteAddAndInc"
	InstallNativeContract(shardAContract, map[string]native.Handler{
		method: RemoteInvokeAddAndInc,
	})

	tx := BuildInvokeTx(shardAContract, method, []interface{}{""})
	assert.NotNil(t, tx)

	overlay := NewOverlayDB()
	cache := storage.NewCacheDB(overlay)
	xshardDB := storage.NewXShardDB(overlay)
	header := &types2.Header{}
	txHash := tx.Hash()
	txevent := &event.ExecuteNotify{TxHash: txHash, State: event.CONTRACT_STATE_FAIL}
	notify := &event.TransactionNotify{
		ContractEvent: txevent,
	}

	_, err := ledgerstore.HandleInvokeTransaction(nil, overlay, cache, xshardDB, tx, header, notify)
	assert.NotNil(t, err)

	shardID := xshard_state.ShardTxID(string(txHash[:]))
	state, err := xshardDB.GetXShardState(shardID)
	assert.Nil(t, err)
	assert.NotNil(t, state.PendingReq)
	sink := common.NewZeroCopySink(10)
	sink.WriteUint64(2)
	sink.WriteUint64(3)
	expected := &xshard_types.XShardTxReq{
		ShardMsgHeader: xshard_types.ShardMsgHeader{
			TargetShardID: ShardB,
			SourceTxHash:  txHash,
		},
		IdxInTx: 0,
		Payer:   tx.Payer,
		Fee:     neovm.MIN_TRANSACTION_GAS,
		Method:  "handlePing",
		Args:    sink.Bytes(),
	}

	assert.Equal(t, expected, state.PendingReq)
	hs := tx.Hash()

	sink.Reset()
	sink.WriteUint64(5)
	rep := &xshard_types.XShardTxRsp{
		IdxInTx: expected.IdxInTx,
		Error:   false,
		Result:  sink.Bytes(),
	}

	rep.SourceTxHash = hs
	msgs := []xshard_types.CommonShardMsg{rep}
	err = ledgerstore.HandleShardCallTransaction(nil, overlay, cache, xshardDB, msgs, header, notify)
	assert.Nil(t, err)
	sink.Reset()
	sink.WriteUint64(6)

	state, err = xshardDB.GetXShardState(state.TxID)
	assert.Nil(t, err)
	res, _ := json.Marshal(state.Notify.Notify[0].States)
	buf, _ := json.Marshal(sink.Bytes())
	assert.Equal(t, string(res), string(buf))

	commit := &xshard_types.XShardCommitMsg{}
	commit.SourceTxHash = hs

	msgs = []xshard_types.CommonShardMsg{commit}
	err = ledgerstore.HandleShardCallTransaction(nil, overlay, cache, xshardDB, msgs, header, notify)
	assert.Nil(t, err)
	sink.Reset()
	sink.WriteUint64(6)

	assert.Nil(t, err)
	res, _ = json.Marshal(notify.ContractEvent.Notify[0].States)
	buf, _ = json.Marshal(sink.Bytes())
	assert.Equal(t, string(res), string(buf))
}
