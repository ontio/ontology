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

package vbft

import (
	"os"
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/ledger"
)

func newChainStore(t *testing.T) *ChainStore {
	log.InitLog(log.InfoLog, log.Stdout)
	var err error
	acct := account.NewAccount("SHA256withECDSA")
	if acct == nil {
		t.Fatalf("GetDefaultAccount error: acc is nil")
	}

	db, err := ledger.NewLedger(config.DEFAULT_DATA_DIR, 0)
	if err != nil {
		t.Fatalf("NewLedger error %s", err)
	}
	acc1 := account.NewAccount("")
	acc2 := account.NewAccount("")
	acc3 := account.NewAccount("")
	acc4 := account.NewAccount("")
	acc5 := account.NewAccount("")
	acc6 := account.NewAccount("")
	acc7 := account.NewAccount("")

	bookkeepers := []keypair.PublicKey{acc1.PublicKey, acc2.PublicKey, acc3.PublicKey, acc4.PublicKey, acc5.PublicKey, acc6.PublicKey, acc7.PublicKey}
	genesisConfig := config.DefConfig.Genesis
	shardConfig := config.DefConfig.Shard
	block, err := genesis.BuildGenesisBlock(bookkeepers, genesisConfig, shardConfig)
	if err != nil {
		t.Fatalf("BuildGenesisBlock error %s", err)
	}
	err = db.Init(bookkeepers, block)
	if err != nil {
		t.Fatalf("InitLedgerStoreWithGenesisBlock error %s", err)
	}
	chainstore, err := OpenBlockStore(db, nil)
	if err != nil {
		t.Fatalf("openblockstore failed: %v\n", err)
	}
	return chainstore
}

func cleanChainStore(t *testing.T, chainstore *ChainStore) {
	chainstore.close()
	os.RemoveAll(config.DEFAULT_DATA_DIR)
}

func TestGetChainedBlockNum(t *testing.T) {
	chainstore := newChainStore(t)
	if chainstore == nil {
		t.Error("newChainStrore error")
		return
	}
	defer cleanChainStore(t, chainstore)

	blocknum := chainstore.GetChainedBlockNum()
	t.Logf("TestGetChainedBlockNum :%d", blocknum)
}
