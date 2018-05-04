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
	"bytes"
	"encoding/hex"
	"testing"
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	tc "github.com/ontio/ontology/txnpool/common"
	"github.com/ontio/ontology/validator/stateless"
	vt "github.com/ontio/ontology/validator/types"
	"github.com/stretchr/testify/assert"
)

var (
	txn    *types.Transaction
	topic  string
	sender tc.SenderType
)

func init() {
	log.Init(log.PATH, log.Stdout)
	topic = "TXN"

	bookKeepingPayload := &payload.Bookkeeping{
		Nonce: uint64(time.Now().UnixNano()),
	}

	txn = &types.Transaction{
		Version:    0,
		Attributes: []*types.TxAttribute{},
		TxType:     types.Bookkeeper,
		Payload:    bookKeepingPayload,
	}

	tempStr := "3369930accc1ddd067245e8edadcd9bea207ba5e1753ac18a51df77a343bfe92"
	hex, _ := hex.DecodeString(tempStr)
	var hash common.Uint256
	hash.Deserialize(bytes.NewReader(hex))
	txn.SetHash(hash)

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
	s = NewTxPoolServer(tc.MAX_WORKER_NUM)
	if s == nil {
		t.Error("Test case: new tx pool server failed")
		return
	}
	defer s.Stop()

	// Case 1: Send nil txn to the server, server should reject it
	s.assignTxToWorker(nil, sender)
	/* Case 2: send non-nil txn to the server, server should assign
	 * it to the worker
	 */
	s.assignTxToWorker(txn, sender)

	/* Case 3: Duplicate input the tx, server should reject the second
	 * one
	 */
	time.Sleep(10 * time.Second)
	s.assignTxToWorker(txn, sender)
	s.assignTxToWorker(txn, sender)

	/* Case 4: Given the tx is in the tx pool, server can get the tx
	 * with the invalid hash
	 */
	time.Sleep(10 * time.Second)
	txEntry := &tc.TXEntry{
		Tx:    txn,
		Attrs: []*tc.TXAttr{},
		Fee:   txn.GetTotalFee(),
	}
	s.addTxList(txEntry)

	ret := s.checkTx(txn.Hash())
	if ret == false {
		t.Error("Failed to check the tx")
		return
	}

	entry := s.getTransaction(txn.Hash())
	if entry == nil {
		t.Error("Failed to get the transaction")
		return
	}

	t.Log("Ending test tx")
}

func TestAssignRsp2Worker(t *testing.T) {
	t.Log("Starting assign response to the worker testing")
	var s *TXPoolServer
	s = NewTxPoolServer(tc.MAX_WORKER_NUM)
	if s == nil {
		t.Error("Test case: new tx pool server failed")
		return
	}

	defer s.Stop()

	s.assignRspToWorker(nil)

	statelessRsp := &vt.CheckResponse{
		WorkerId: 0,
		ErrCode:  errors.ErrNoError,
		Hash:     txn.Hash(),
		Type:     vt.Stateless,
		Height:   0,
	}

	statefulRsp := &vt.CheckResponse{
		WorkerId: 0,
		ErrCode:  errors.ErrUnknown,
		Hash:     txn.Hash(),
		Type:     vt.Stateful,
		Height:   0,
	}
	s.assignRspToWorker(statelessRsp)
	s.assignRspToWorker(statefulRsp)

	statelessRsp = &vt.CheckResponse{
		WorkerId: 0,
		ErrCode:  errors.ErrUnknown,
		Hash:     txn.Hash(),
		Type:     vt.Stateless,
		Height:   0,
	}
	s.assignRspToWorker(statelessRsp)

	t.Log("Ending assign response to the worker testing")
}

func TestActor(t *testing.T) {
	t.Log("Starting actor testing")
	var s *TXPoolServer
	s = NewTxPoolServer(tc.MAX_WORKER_NUM)
	if s == nil {
		t.Error("Test case: new tx pool server failed")
		return
	}

	defer s.Stop()

	rspActor := NewVerifyRspActor(s)
	rspPid := startActor(rspActor)
	if rspPid == nil {
		t.Error("Fail to start verify rsp actor")
		return
	}
	s.RegisterActor(tc.VerifyRspActor, rspPid)

	tpa := NewTxPoolActor(s)
	txPoolPid := startActor(tpa)
	if txPoolPid == nil {
		t.Error("Fail to start txnpool actor")
		return
	}
	s.RegisterActor(tc.TxPoolActor, txPoolPid)

	pid := s.GetPID(tc.VerifyRspActor)
	if pid == nil {
		t.Error("Fail to get the pid")
		return
	}

	pid = s.GetPID(tc.TxPoolActor)
	if pid == nil {
		t.Error("Fail to get the pid")
		return
	}

	pid = s.GetPID(tc.TxActor)
	if pid != nil {
		t.Error("Fail to get the pid")
		return
	}

	s.UnRegisterActor(tc.TxPoolActor)
	s.UnRegisterActor(tc.VerifyRspActor)
	pid = s.GetPID(tc.TxPoolActor)
	if pid != nil {
		t.Error("Pid was not registered")
		return
	}

	pid = s.GetPID(tc.VerifyRspActor)
	if pid != nil {
		t.Error("Pid was not registered")
		return
	}

	pid = s.GetPID(tc.MaxActor)
	if pid != nil {
		t.Error("Invalid PID")
		return
	}

	t.Log("Ending actor testing")
}

func TestValidator(t *testing.T) {
	t.Log("Starting validator testing")
	var s *TXPoolServer
	s = NewTxPoolServer(tc.MAX_WORKER_NUM)
	if s == nil {
		t.Error("Test case: new tx pool server failed")
		return
	}

	defer s.Stop()

	rspActor := NewVerifyRspActor(s)
	rspPid := startActor(rspActor)
	if rspPid == nil {
		t.Error("Fail to start verify rsp actor")
		return
	}
	s.RegisterActor(tc.VerifyRspActor, rspPid)

	statelessV1, err := stateless.NewValidator("stateless1")
	if err != nil {
		t.Error("failed to new stateless valdiator", err)
		return
	}
	statelessV1.Register(rspPid)

	statelessV2, err := stateless.NewValidator("stateless2")
	if err != nil {
		t.Error("failed to new stateless valdiator", err)
		return
	}
	statelessV2.Register(rspPid)

	time.Sleep(1 * time.Second)

	ret := s.getNextValidatorPIDs()
	for _, v := range ret {
		assert.NotNil(t, v)
	}

	ret = s.getNextValidatorPIDs()
	for _, v := range ret {
		assert.NotNil(t, v)
	}

	statelessV1.UnRegister(rspPid)
	statelessV2.UnRegister(rspPid)

	time.Sleep(1 * time.Second)

	ret = s.getNextValidatorPIDs()
	for _, v := range ret {
		assert.NotNil(t, v)
	}

	t.Log("Ending validator testing")
}
