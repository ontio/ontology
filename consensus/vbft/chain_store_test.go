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
	"fmt"
	"os"
	"testing"

	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
)

func newChainStore() *ChainStore {
	log.Init(log.PATH, log.Stdout)
	var err error
	acct := account.NewAccount("SHA256withECDSA")
	if acct == nil {
		fmt.Println("GetDefaultAccount error: acc is nil")
		os.Exit(1)
	}

	ledger.DefLedger, err = ledger.NewLedger()
	if err != nil {
		log.Fatalf("NewLedger error %s", err)
		os.Exit(1)
	}
	store, err := OpenBlockStore(ledger.DefLedger)
	if err != nil {
		fmt.Printf("openblockstore failed: %v\n", err)
		return nil
	}
	return store
}

func TestGetChainedBlockNum(t *testing.T) {
	chainstore := newChainStore()
	if chainstore == nil {
		t.Error("newChainStrore error")
		return
	}
	blocknum := chainstore.GetChainedBlockNum()
	t.Logf("TestGetChainedBlockNum :%d", blocknum)
}
