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
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/ledger"
)

var testBookkeeperAccounts []*account.Account

func newTestChainStore(t *testing.T) *ChainStore {
	log.InitLog(log.InfoLog, log.Stdout)
	var err error
	acct := account.NewAccount("SHA256withECDSA")
	if acct == nil {
		t.Fatalf("GetDefaultAccount error: acc is nil")
	}
	os.RemoveAll(config.DEFAULT_DATA_DIR)
	db, err := ledger.NewLedger(config.DEFAULT_DATA_DIR, 0)
	if err != nil {
		t.Fatalf("NewLedger error %s", err)
	}

	var bookkeepers []keypair.PublicKey
	if len(testBookkeeperAccounts) == 0 {
		for i := 0; i < 7; i++ {
			acc := account.NewAccount("")
			testBookkeeperAccounts = append(testBookkeeperAccounts, acc)
			bookkeepers = append(bookkeepers, acc.PublicKey)
		}
	}

	genesisConfig := config.DefConfig.Genesis

	// update peers in genesis
	for i, p := range genesisConfig.VBFT.Peers {
		if i > 0 && i <= len(testBookkeeperAccounts) {
			p.PeerPubkey = vconfig.PubkeyID(testBookkeeperAccounts[i-1].PublicKey)
		}
	}
	block, err := genesis.BuildGenesisBlock(bookkeepers, genesisConfig)
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

func cleanTestChainStore() {
	os.RemoveAll(config.DEFAULT_DATA_DIR)
	testBookkeeperAccounts = make([]*account.Account, 0)
}

func TestGetChainedBlockNum(t *testing.T) {
	chainstore := newTestChainStore(t)
	if chainstore == nil {
		t.Error("newChainStrore error")
		return
	}
	defer cleanTestChainStore()

	blocknum := chainstore.GetChainedBlockNum()
	t.Logf("TestGetChainedBlockNum :%d", blocknum)
}
