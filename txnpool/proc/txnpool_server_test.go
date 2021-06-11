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
	"testing"
	"time"

	"github.com/ontio/ontology/cmd/utils"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
	tc "github.com/ontio/ontology/txnpool/common"
)

var (
	txn    *types.Transaction
	sender tc.SenderType
)

func init() {
	code := []byte("ont")

	invokeCodePayload := &payload.InvokeCode{
		Code: code,
	}

	acct := account.NewAccount("")

	mutable := &types.MutableTransaction{
		TxType:  types.InvokeNeo,
		Nonce:   uint32(time.Now().Unix()),
		Payload: invokeCodePayload,
		Payer:   acct.Address,
	}

	err := utils.SignTransaction(acct, mutable)
	if err != nil {
		panic(err)
	}

	txn, _ = mutable.IntoImmutable()

	sender = tc.NilSender
}

func startActor(obj interface{}) *actor.PID {
	props := actor.FromProducer(func() actor.Actor {
		return obj.(actor.Actor)
	})

	pid := actor.Spawn(props)

	return pid
}

func TestTxn(t *testing.T) {
	t.Log("Starting test tx")
	var s *TXPoolServer
	s = NewTxPoolServer(true, false)
	if s == nil {
		t.Error("Test case: new tx pool server failed")
		return
	}
	defer s.Stop()

	/* Case 2: send non-nil txn to the server, server should assign
	 * it to the worker
	 */
	s.startTxVerify(txn, sender, nil)

	/* Case 3: Duplicate input the tx, server should reject the second
	 * one
	 */
	time.Sleep(10 * time.Second)
	s.startTxVerify(txn, sender, nil)
	s.startTxVerify(txn, sender, nil)

	t.Log("Ending test tx")
}

func TestActor(t *testing.T) {
	t.Log("Starting actor testing")
	var s *TXPoolServer
	s = NewTxPoolServer(true, false)
	if s == nil {
		t.Error("Test case: new tx pool server failed")
		return
	}

	defer s.Stop()

	tpa := NewTxPoolActor(s)
	txPoolPid := startActor(tpa)
	if txPoolPid == nil {
		t.Error("Fail to start txnpool actor")
		return
	}
	s.RegisterActor(txPoolPid)

	pid := s.GetPID()
	if pid == nil {
		t.Error("Fail to get the pid")
		return
	}

	t.Log("Ending actor testing")
}
