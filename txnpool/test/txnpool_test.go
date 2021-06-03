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
	"os"
	"testing"
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
)

func TestMain(m *testing.M) {
	log.InitLog(log.InfoLog, log.Stdout)

	var err error
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

	// tear down
	ledger.DefLedger.Close()
	os.RemoveAll(config.DEFAULT_DATA_DIR)
}

func initTestTx() *types.Transaction {
	log.InitLog(log.InfoLog, log.Stdout)
	//topic := "TXN"

	mutable := &types.MutableTransaction{
		TxType:  types.InvokeNeo,
		Nonce:   uint32(time.Now().Unix()),
		Payload: &payload.InvokeCode{Code: []byte{}},
	}

	tx, _ := mutable.IntoImmutable()
	return tx
}

func startActor(obj interface{}) *actor.PID {
	props := actor.FromProducer(func() actor.Actor {
		return obj.(actor.Actor)
	})

	pid := actor.Spawn(props)
	return pid
}
