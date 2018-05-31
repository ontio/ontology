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

package ledgerstore

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/types"
)

var testBlockStore *BlockStore
var testStateStore *StateStore
var testLedgerStore *LedgerStoreImp

func TestMain(m *testing.M) {
	log.InitLog(0)

	var err error
	testLedgerStore, err = NewLedgerStore("test/ledger")
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewLedgerStore error %s\n", err)
		return
	}

	testBlockDir := "test/block"
	testBlockStore, err = NewBlockStore(testBlockDir, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewBlockStore error %s\n", err)
		return
	}
	testStateDir := "test/state"
	merklePath := "test/" + MerkleTreeStorePath
	testStateStore, err = NewStateStore(testStateDir, merklePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewStateStore error %s\n", err)
		return
	}
	m.Run()
	err = testLedgerStore.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "testLedgerStore.Close error %s\n", err)
		return
	}
	err = testBlockStore.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "testBlockStore.Close error %s\n", err)
		return
	}
	err = testStateStore.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "testStateStore.Close error %s", err)
		return
	}
	err = os.RemoveAll("./test")
	if err != nil {
		fmt.Fprintf(os.Stderr, "os.RemoveAll error %s\n", err)
		return
	}
	os.RemoveAll("ActorLog")
}

func TestInitLedgerStoreWithGenesisBlock(t *testing.T) {
	acc1 := account.NewAccount("")
	acc2 := account.NewAccount("")
	acc3 := account.NewAccount("")
	acc4 := account.NewAccount("")

	bookkeepers := []keypair.PublicKey{acc1.PublicKey, acc2.PublicKey, acc3.PublicKey, acc4.PublicKey}
	bookkeeper, err := types.AddressFromBookkeepers(bookkeepers)
	if err != nil {
		t.Errorf("AddressFromBookkeepers error %s", err)
		return
	}
	header := &types.Header{
		Version:          123,
		PrevBlockHash:    common.Uint256{},
		TransactionsRoot: common.Uint256{},
		Timestamp:        uint32(uint32(time.Date(2017, time.February, 23, 0, 0, 0, 0, time.UTC).Unix())),
		Height:           uint32(0),
		ConsensusData:    1234567890,
		NextBookkeeper:   bookkeeper,
	}
	block := &types.Block{
		Header:       header,
		Transactions: []*types.Transaction{},
	}

	err = testLedgerStore.InitLedgerStoreWithGenesisBlock(block, bookkeepers)
	if err != nil {
		t.Errorf("TestInitLedgerStoreWithGenesisBlock error %s", err)
		return
	}

	curBlockHeight := testLedgerStore.GetCurrentBlockHeight()
	curBlockHash := testLedgerStore.GetCurrentBlockHash()
	if curBlockHeight != block.Header.Height {
		t.Errorf("TestInitLedgerStoreWithGenesisBlock failed CurrentBlockHeight %d != %d", curBlockHeight, block.Header.Height)
		return
	}
	if curBlockHash != block.Hash() {
		t.Errorf("TestInitLedgerStoreWithGenesisBlock failed CurrentBlockHash %x != %x", curBlockHash, block.Hash())
		return
	}
	block1, err := testLedgerStore.GetBlockByHeight(curBlockHeight)
	if err != nil {
		t.Errorf("TestInitLedgerStoreWithGenesisBlock failed GetBlockByHeight error %s", err)
		return
	}
	if block1.Hash() != block.Hash() {
		t.Errorf("TestInitLedgerStoreWithGenesisBlock failed blockhash %x != %x", block1.Hash(), block.Hash())
		return
	}
}
