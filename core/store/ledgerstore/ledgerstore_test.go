package ledgerstore

import (
	"fmt"
	"github.com/Ontology/common"
	"github.com/Ontology/core/payload"
	"github.com/Ontology/core/types"
	"github.com/Ontology/core/utils"
	"github.com/Ontology/crypto"
	"os"
	"testing"
	"time"
)

var testBlockStore *BlockStore
var testStateStore *StateStore
var testLedgerStore *LedgerStore

func TestMain(m *testing.M) {
	var err error
	DBDirEvent = "test/ledger/ledgerevent"
	DBDirBlock = "test/ledger/block"
	DBDirState = "test/ledger/states"
	DBDirMerkleTree = "test/ledger/merkle"
	MerkleTreeStorePath = "test/ledger/merkle_tree.db"
	testLedgerStore, err = NewLedgerStore()
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
	testStateStore, err = NewStateStore(testStateDir)
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
}

func TestInitLedgerStoreWithGenesisBlock(t *testing.T) {
	_, pubKey1, _ := crypto.GenKeyPair()
	_, pubKey2, _ := crypto.GenKeyPair()
	_, pubKey3, _ := crypto.GenKeyPair()
	_, pubKey4, _ := crypto.GenKeyPair()

	bookKeepers := []*crypto.PubKey{&pubKey1, &pubKey2, &pubKey3, &pubKey4}
	bookKeeper, err := types.AddressFromBookKeepers(bookKeepers)
	if err != nil {
		t.Errorf("AddressFromBookKeepers error %s", err)
		return
	}
	header := &types.Header{
		Version:          123,
		PrevBlockHash:    common.Uint256{},
		TransactionsRoot: common.Uint256{},
		Timestamp:        uint32(uint32(time.Date(2017, time.February, 23, 0, 0, 0, 0, time.UTC).Unix())),
		Height:           uint32(0),
		ConsensusData:    1234567890,
		NextBookKeeper:   bookKeeper,
	}
	tx1 := &types.Transaction{
		TxType: types.BookKeeping,
		Payload: &payload.BookKeeping{
			Nonce: 1234567890,
		},
		Attributes: []*types.TxAttribute{},
	}
	block := &types.Block{
		Header:       header,
		Transactions: []*types.Transaction{tx1},
	}

	err = testLedgerStore.InitLedgerStoreWithGenesisBlock(block, bookKeepers)
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
