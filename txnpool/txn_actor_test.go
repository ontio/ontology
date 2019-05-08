package txnpool

import (
	"testing"
	"time"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/events"
	tc "github.com/ontio/ontology/txnpool/common"
	"github.com/stretchr/testify/assert"
)

var (
	txn *types.Transaction
)

func init() {
	events.Init()

	code := []byte("ont")
	invokeCodePayload := &payload.InvokeCode{
		Code: code,
	}
	mutable := &types.MutableTransaction{
		TxType:  types.Invoke,
		Nonce:   uint32(time.Now().Unix()),
		Payload: invokeCodePayload,
	}

	txn, _ = mutable.IntoImmutable()

	ledger.DefLedger, _ = ledger.NewLedger(config.DEFAULT_DATA_DIR, 0)
	bookKeepers, _ := config.DefConfig.GetBookkeepers()
	genesisConfig := config.DefConfig.Genesis
	shardConfig := config.DefConfig.Shard
	genesisBlock, _ := genesis.BuildGenesisBlock(bookKeepers, genesisConfig, shardConfig)
	ledger.DefLedger.Init(bookKeepers, genesisBlock)
}

func TestTxActor(t *testing.T) {
	t.Log("Starting tx actor test")
	shardId := common.NewShardIDUnchecked(config.DEFAULT_SHARD_ID)
	mgr, err := NewTxnPoolManager(shardId, true, false)
	if err != nil {
		t.Fatalf("Test case: new tx pool server failed")
	}

	txActor := NewTxActor(mgr)
	txPid, err := startActor(txActor, "tx")
	if err != nil {
		t.Fatalf("Test case: start tx actor failed")
	}

	s, err := mgr.StartTxnPoolServer(shardId, ledger.DefLedger)

	txReq := &tc.TxReq{
		Tx:     txn,
		Sender: tc.NilSender,
	}
	txPid.Tell(txReq)

	time.Sleep(1 * time.Second)

	future := txPid.RequestFuture(&tc.GetTxnReq{Hash: txn.Hash()}, 1*time.Second)
	result, err := future.Result()
	assert.Nil(t, err)
	rsp := (result).(*tc.GetTxnRsp)
	assert.Nil(t, rsp.Txn)

	future = txPid.RequestFuture(&tc.GetTxnStats{}, 2*time.Second)
	result, err = future.Result()
	assert.Nil(t, err)
	future = txPid.RequestFuture(&tc.CheckTxnReq{Hash: txn.Hash()}, 1*time.Second)
	result, err = future.Result()
	assert.Nil(t, err)

	future = txPid.RequestFuture(&tc.GetTxnStatusReq{Hash: txn.Hash()}, 1*time.Second)
	result, err = future.Result()
	assert.Nil(t, err)

	// Given the tx in the pool, test again
	txEntry := &tc.TXEntry{
		Tx:    txn,
		Attrs: []*tc.TXAttr{},
	}
	s.AddTxList(txEntry)

	future = txPid.RequestFuture(&tc.GetTxnReq{Hash: txn.Hash()}, 1*time.Second)
	result, err = future.Result()
	assert.Nil(t, err)

	future = txPid.RequestFuture(&tc.GetTxnStats{}, 2*time.Second)
	result, err = future.Result()
	assert.Nil(t, err)
	future = txPid.RequestFuture(&tc.CheckTxnReq{Hash: txn.Hash()}, 1*time.Second)
	result, err = future.Result()
	assert.Nil(t, err)

	future = txPid.RequestFuture(&tc.GetTxnStatusReq{Hash: txn.Hash()}, 1*time.Second)
	result, err = future.Result()
	assert.Nil(t, err)

	txPid.Tell("test")
	s.Stop()
	t.Log("Ending tx actor test")
}
