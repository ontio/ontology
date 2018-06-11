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

package txnpool

import (
	"bytes"
	"encoding/hex"
	"sync"
	"testing"
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
	tc "github.com/ontio/ontology/txnpool/common"
	tp "github.com/ontio/ontology/txnpool/proc"
	"github.com/ontio/ontology/validator/stateful"
	"github.com/ontio/ontology/validator/stateless"
)

var (
	tx    *types.Transaction
	topic string
)

func init() {
	log.Init(log.PATH, log.Stdout)
	topic = "TXN"

	tx = &types.Transaction{
		Version: 0,
	}

	tempStr := "3369930accc1ddd067245e8edadcd9bea207ba5e1753ac18a51df77a343bfe92"
	hex, _ := hex.DecodeString(tempStr)
	var hash common.Uint256
	hash.Deserialize(bytes.NewReader(hex))
	tx.SetHash(hash)
}

func startActor(obj interface{}) *actor.PID {
	props := actor.FromProducer(func() actor.Actor {
		return obj.(actor.Actor)
	})

	pid := actor.Spawn(props)
	return pid
}

func Test_RCV(t *testing.T) {
	var s *tp.TXPoolServer
	var wg sync.WaitGroup
	var err error
	ledger.DefLedger, err = ledger.NewLedger(config.DEFAULT_DATA_DIR)
	if err != nil {
		t.Error("failed  to new ledger")
		return
	}

	bookKeepers, err := config.DefConfig.GetBookkeepers()
	if err != nil {
		t.Error("failed to get bookkeepers")
		return
	}
	genesisConfig := config.DefConfig.Genesis
	genesisBlock, err := genesis.BuildGenesisBlock(bookKeepers, genesisConfig)
	if err != nil {
		t.Error("failed to build genesis block")
		return
	}
	err = ledger.DefLedger.Init(bookKeepers, genesisBlock)
	if err != nil {
		t.Error("failed to initialize default ledger")
		return
	}

	// Start txnpool server to receive msgs from p2p, consensus and valdiators
	s = tp.NewTxPoolServer(tc.MAX_WORKER_NUM)

	// Initialize an actor to handle the msgs from valdiators
	rspActor := tp.NewVerifyRspActor(s)
	rspPid := startActor(rspActor)
	if rspPid == nil {
		t.Error("Fail to start verify rsp actor")
		return
	}
	s.RegisterActor(tc.VerifyRspActor, rspPid)

	// Initialize an actor to handle the msgs from consensus
	tpa := tp.NewTxPoolActor(s)
	txPoolPid := startActor(tpa)
	if txPoolPid == nil {
		t.Error("Fail to start txnpool actor")
		return
	}
	s.RegisterActor(tc.TxPoolActor, txPoolPid)

	// Initialize an actor to handle the msgs from p2p and api
	ta := tp.NewTxActor(s)
	txPid := startActor(ta)
	if txPid == nil {
		t.Error("Fail to start txn actor")
		return
	}
	s.RegisterActor(tc.TxActor, txPid)

	// Start stateless validator
	statelessV, err := stateless.NewValidator("stateless")
	if err != nil {
		t.Errorf("failed to new stateless valdiator", err)
		return
	}
	statelessV.Register(rspPid)

	statelessV2, err := stateless.NewValidator("stateless2")
	if err != nil {
		t.Errorf("failed to new stateless valdiator", err)
		return
	}
	statelessV2.Register(rspPid)

	statelessV3, err := stateless.NewValidator("stateless3")
	if err != nil {
		t.Errorf("failed to new stateless valdiator", err)
		return
	}
	statelessV3.Register(rspPid)

	statefulV, err := stateful.NewValidator("stateful")
	if err != nil {
		t.Errorf("failed to new stateful valdiator", err)
		return
	}
	statefulV.Register(rspPid)

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			var j int
			defer wg.Done()
			for {
				j++
				txReq := &tc.TxReq{
					Tx:     tx,
					Sender: tc.NilSender,
				}
				txPid.Tell(txReq)

				if j >= 4 {
					return
				}
			}
		}()
	}

	wg.Wait()
	time.Sleep(1 * time.Second)
	txPoolPid.Tell(&tc.GetTxnPoolReq{ByCount: true})
	txPoolPid.Tell(&tc.GetPendingTxnReq{ByCount: true})
	time.Sleep(2 * time.Second)

	statelessV.UnRegister(rspPid)
	statelessV2.UnRegister(rspPid)
	statelessV3.UnRegister(rspPid)
	statefulV.UnRegister(rspPid)
	s.Stop()
}
