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

package proc

import (
	"fmt"
	"testing"
	"time"

	"github.com/Ontology/core/types"
	"github.com/Ontology/events/message"
	tc "github.com/Ontology/txnpool/common"
	vt "github.com/Ontology/validator/types"
)

func TestTxActor(t *testing.T) {
	fmt.Println("Starting tx actor test")
	s := NewTxPoolServer(tc.MAX_WORKER_NUM)
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

	txReq := &txReq{
		Tx:     txn,
		Sender: tc.NilSender,
	}
	txPid.RequestFuture(txn)

	future := txPid.RequestFuture(&tc.GetTxnReq{Hash: txn.Hash()}, 1*time.Second)
	result, err := future.Result()
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
	s := NewTxPoolServer(tc.MAX_WORKER_NUM)
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
	s := NewTxPoolServer(tc.MAX_WORKER_NUM)
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
