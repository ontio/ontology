package proc

import (
	"fmt"
	"github.com/Ontology/core/types"
	"github.com/Ontology/events/message"
	tc "github.com/Ontology/txnpool/common"
	vt "github.com/Ontology/validator/types"
	"testing"
	"time"
)

func TestTxActor(t *testing.T) {
	fmt.Println("Starting tx actor test")
	s := NewTxPoolServer(tc.MAXWORKERNUM)
	if s == nil {
		t.Error("Test case: new tx pool server failed")
		return
	}

	txActor := NewTxActor(s)
	txPid := startActor(txActor)
	if txPid == nil {
		t.Error("Test case: start tx actor failed")
		s.Stop()
		return
	}
	future := txPid.RequestFuture(txn, 5*time.Second)
	result, err := future.Result()
	fmt.Println(result, err)

	future = txPid.RequestFuture(&tc.GetTxnReq{Hash: txn.Hash()}, 1*time.Second)
	result, err = future.Result()
	fmt.Println(result, err)

	future = txPid.RequestFuture(&tc.GetTxnStats{}, 2*time.Second)
	result, err = future.Result()
	fmt.Println(result, err)
	future = txPid.RequestFuture(&tc.CheckTxnReq{Hash: txn.Hash()}, 1*time.Second)
	result, err = future.Result()
	fmt.Println(result, err)

	future = txPid.RequestFuture(&tc.GetTxnStatusReq{Hash: txn.Hash()}, 1*time.Second)
	result, err = future.Result()
	fmt.Println(result, err)

	// Given the tx in the pool, test again
	txEntry := &tc.TXEntry{
		Tx:    txn,
		Attrs: []*tc.TXAttr{},
		Fee:   txn.GetTotalFee(),
	}
	s.addTxList(txEntry)
	future = txPid.RequestFuture(txn, 5*time.Second)
	result, err = future.Result()
	fmt.Println(result, err)

	future = txPid.RequestFuture(&tc.GetTxnReq{Hash: txn.Hash()}, 1*time.Second)
	result, err = future.Result()
	fmt.Println(result, err)

	future = txPid.RequestFuture(&tc.GetTxnStats{}, 2*time.Second)
	result, err = future.Result()
	fmt.Println(result, err)
	future = txPid.RequestFuture(&tc.CheckTxnReq{Hash: txn.Hash()}, 1*time.Second)
	result, err = future.Result()
	fmt.Println(result, err)

	future = txPid.RequestFuture(&tc.GetTxnStatusReq{Hash: txn.Hash()}, 1*time.Second)
	result, err = future.Result()
	fmt.Println(result, err)

	txPid.Tell("test")
	s.Stop()
	fmt.Println("Ending tx actor test")
}

func TestTxPoolActor(t *testing.T) {
	fmt.Println("Starting tx pool actor test")
	s := NewTxPoolServer(tc.MAXWORKERNUM)
	if s == nil {
		t.Error("Test case: new tx pool server failed")
		return
	}

	txPoolActor := NewTxPoolActor(s)
	txPoolPid := startActor(txPoolActor)
	if txPoolPid == nil {
		t.Error("Test case: start tx actor failed")
		s.Stop()
		return
	}

	txPoolPid.Tell(txn)

	future := txPoolPid.RequestFuture(&tc.GetTxnPoolReq{ByCount: false}, 2*time.Second)
	result, err := future.Result()
	fmt.Println(result, err)

	future = txPoolPid.RequestFuture(&tc.GetPendingTxnReq{ByCount: false}, 2*time.Second)
	result, err = future.Result()
	fmt.Println(result, err)

	bk := &tc.VerifyBlockReq{
		Height: 1,
		Txs:    []*types.Transaction{txn},
	}
	future = txPoolPid.RequestFuture(bk, 10*time.Second)
	result, err = future.Result()
	fmt.Println(result, err)

	sbc := &message.SaveBlockCompleteMsg{}
	txPoolPid.Tell(sbc)

	s.Stop()
	fmt.Println("Ending tx pool actor test")
}

func TestVerifyRspActor(t *testing.T) {
	fmt.Println("Starting validator response actor test")
	s := NewTxPoolServer(tc.MAXWORKERNUM)
	if s == nil {
		t.Error("Test case: new tx pool server failed")
		return
	}

	validatorActor := NewVerifyRspActor(s)
	validatorPid := startActor(validatorActor)
	if validatorPid == nil {
		t.Error("Test case: start tx actor failed")
		s.Stop()
		return
	}

	validatorPid.Tell(txn)

	registerMsg := &vt.RegisterValidator{}
	validatorPid.Tell(registerMsg)

	unRegisterMsg := &vt.UnRegisterValidator{}
	validatorPid.Tell(unRegisterMsg)

	rsp := &vt.CheckResponse{}
	validatorPid.Tell(rsp)

	s.Stop()
	fmt.Println("Ending validator response actor test")
}
