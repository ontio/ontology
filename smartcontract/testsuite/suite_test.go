package testsuite

import (
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native/shard_sysmsg"
	"github.com/ontio/ontology/smartcontract/service/neovm"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/ontio/ontology/core/chainmgr/xshard_state"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/storage"
)

func TestRemoteNotifyPing(t *testing.T) {
	shardAContract := RandomAddress()
	InstallNativeContract(shardAContract, map[string]native.Handler{
		"remoteNotifyPing": RemoteNotifyPing,
		"handlePing":       HandlePing,
	})

	tx := BuildInvokeTx(shardAContract, "remoteNotifyPing", []interface{}{""})
	assert.NotNil(t, tx)

	state, err := executeTransaction(tx)

	assert.Nil(t, err)
	assert.Equal(t, len(state.ShardNotifies), 1)
	notify := state.ShardNotifies[0]
	sink := common.NewZeroCopySink(10)
	sink.WriteString(fmt.Sprintf("hello from shard: %d", tx.ShardID))
	expected := &xshard_state.XShardNotify{
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
	InstallNativeContract(shardAContract, map[string]native.Handler{
		"remoteInvokeAdd": RemoteInvokeAdd,
	})

	tx := BuildInvokeTx(shardAContract, "remoteInvokeAdd", []interface{}{""})
	assert.NotNil(t, tx)

	state, err := executeTransaction(tx)

	//assert.Equal(t, shardsysmsg.ErrYield, err) // error is wrapped
	assert.NotNil(t, err)
	assert.NotNil(t, state.PendingReq)
	sink := common.NewZeroCopySink(10)
	sink.WriteUint64(2)
	sink.WriteUint64(3)
	expected := &xshard_state.XShardTxReq{
		IdxInTx: 0,
		Payer:   tx.Payer,
		Fee:     neovm.MIN_TRANSACTION_GAS,
		Method:  "handlePing",
		Args:    sink.Bytes(),
	}

	assert.Equal(t, expected, state.PendingReq)
}

func executeTransaction(tx *types.Transaction) (*xshard_state.TxState, error) {
	config := &smartcontract.Config{
		ShardID: types.NewShardIDUnchecked(tx.ShardID),
		Time:    uint32(time.Now().Unix()),
		Tx:      tx,
	}

	overlay := NewOverlayDB()
	cache := storage.NewCacheDB(overlay)

	txHash := tx.Hash()
	txState := xshard_state.CreateTxState(xshard_state.ShardTxID(string(txHash[:])))

	if tx.TxType == types.Invoke {
		invoke := tx.Payload.(*payload.InvokeCode)

		sc := smartcontract.SmartContract{
			Config:           config,
			Store:            nil,
			MainShardTxState: txState,
			CacheDB:          cache,
			Gas:              100000000000000,
			PreExec:          true,
		}

		//start the smart contract executive function
		engine, _ := sc.NewExecuteEngine(invoke.Code)
		_, err := engine.Invoke()

		if err != nil {
			//if err == shardsysmsg.ErrYield {
			//	return txState, err
			//}
			// todo: handle error check
			if txState.PendingReq != nil {
				return txState, shardsysmsg.ErrYield
			}
			return nil, err
		}

		return txState, nil
	}

	panic("unimplemented")
}
