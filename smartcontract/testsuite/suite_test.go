package testsuite

import (
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
	notify := &event.ExecuteNotify{TxHash: txHash, State: event.CONTRACT_STATE_FAIL}

	state := xshard_state.CreateTxState(xshard_state.ShardTxID(string(txHash[:])))
	_, err := ledgerstore.HandleInvokeTransaction(nil, overlay, cache, xshardDB, state, tx, header, notify)
	//state, _, err := executeTransaction(tx, cache)

	//assert.Equal(t, shardsysmsg.ErrYield, err) // error is wrapped
	assert.NotNil(t, err)
	assert.NotNil(t, state.PendingReq)
	sink := common.NewZeroCopySink(10)
	sink.WriteUint64(2)
	sink.WriteUint64(3)
	expected := &xshard_types.XShardTxReq{
		IdxInTx: 0,
		Payer:   tx.Payer,
		Fee:     neovm.MIN_TRANSACTION_GAS,
		Method:  "handlePing",
		Args:    sink.Bytes(),
	}

	assert.Equal(t, expected, state.PendingReq.Req)
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
