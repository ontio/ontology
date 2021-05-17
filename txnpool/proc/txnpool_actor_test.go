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
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/ledger"
	"os"
	"testing"
	"time"

	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/events/message"
	tc "github.com/ontio/ontology/txnpool/common"
	vt "github.com/ontio/ontology/validator/types"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	log.InitLog(log.InfoLog, log.Stdout)
	bookKeepers, err := config.DefConfig.GetBookkeepers()
	if err != nil {
		return
	}
	genesisConfig := config.DefConfig.Genesis
	genesisBlock, err := genesis.BuildGenesisBlock(bookKeepers, genesisConfig)
	if err != nil {
		return
	}
	ledger.DefLedger, err = ledger.InitLedger(config.DEFAULT_DATA_DIR, 0, bookKeepers, genesisBlock)
	if err != nil {
		return
	}

	m.Run()

	ledger.DefLedger.Close()
	os.RemoveAll(config.DEFAULT_DATA_DIR)
}

func TestTxActor(t *testing.T) {
	t.Log("Starting tx actor test")
	s := NewTxPoolServer(true, false)
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

	future = txPid.RequestFuture(&tc.GetTxnStatusReq{Hash: txn.Hash()}, 1*time.Second)
	result, err = future.Result()
	assert.Nil(t, err)

	// Given the tx in the pool, test again
	txEntry := &tc.TXEntry{
		Tx:    txn,
		Attrs: []*tc.TXAttr{},
	}
	s.addTxList(txEntry)

	future = txPid.RequestFuture(&tc.GetTxnReq{Hash: txn.Hash()}, 1*time.Second)
	result, err = future.Result()
	assert.Nil(t, err)

	future = txPid.RequestFuture(&tc.GetTxnStatusReq{Hash: txn.Hash()}, 1*time.Second)
	result, err = future.Result()
	assert.Nil(t, err)

	txPid.Tell("test")
	s.Stop()
	t.Log("Ending tx actor test")
}

func TestTxPoolActor(t *testing.T) {
	t.Log("Starting tx pool actor test")
	s := NewTxPoolServer(true, false)
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

	txEntry := &tc.TXEntry{
		Tx:    txn,
		Attrs: []*tc.TXAttr{},
	}

	retAttr := &tc.TXAttr{
		Height:  0,
		Type:    vt.Stateful,
		ErrCode: errors.ErrNoError,
	}
	txEntry.Attrs = append(txEntry.Attrs, retAttr)
	s.addTxList(txEntry)

	future := txPoolPid.RequestFuture(&tc.GetTxnPoolReq{ByCount: false}, 2*time.Second)
	result, err := future.Result()
	assert.Nil(t, err)
	rsp := (result).(*tc.GetTxnPoolRsp)
	assert.NotNil(t, rsp.TxnPool)

	bk := &tc.VerifyBlockReq{
		Height: 0,
		Txs:    []*types.Transaction{txn},
	}
	future = txPoolPid.RequestFuture(bk, 10*time.Second)
	result, err = future.Result()
	assert.Nil(t, err)

	sbc := &message.SaveBlockCompleteMsg{}
	txPoolPid.Tell(sbc)

	s.Stop()
	t.Log("Ending tx pool actor test")
}
