package ledger

import (
	. "github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/genesis"
	"github.com/Ontology/core/types"
	"github.com/Ontology/crypto"
	. "github.com/Ontology/errors"
	"github.com/Ontology/events"
	"sync"
)

type Blockchain struct {
	BlockHeight uint32
	BCEvents    *events.Event
	mutex       sync.Mutex
}

func NewBlockchain(height uint32) *Blockchain {
	return &Blockchain{
		BlockHeight: height,
		BCEvents:    events.NewEvent(),
	}
}

func NewBlockchainWithGenesisBlock(defaultBookKeeper []*crypto.PubKey) (*Blockchain, error) {
	genesisBlock, err := genesis.GenesisBlockInit(defaultBookKeeper)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "[Blockchain], NewBlockchainWithGenesisBlock failed.")
	}
	genesisBlock.RebuildMerkleRoot()
	genesisBlock.Header.BlockRoot = genesisBlock.Header.TransactionsRoot

	height, err := DefaultLedger.Store.InitLedgerStoreWithGenesisBlock(genesisBlock, defaultBookKeeper)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "[Blockchain], InitLevelDBStoreWithGenesisBlock failed.")
	}
	blockchain := NewBlockchain(height)
	return blockchain, nil
}

func (bc *Blockchain) AddBlock(block *types.Block) error {
	log.Debug()
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	err := bc.SaveBlock(block)
	if err != nil {
		return err
	}

	return nil
}

func (bc *Blockchain) GetHeader(hash Uint256) (*types.Header, error) {
	header, err := DefaultLedger.Store.GetHeader(hash)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "[Blockchain], GetHeader failed.")
	}
	return header, nil
}

func (bc *Blockchain) SaveBlock(block *types.Block) error {
	log.Debugf("Save block, block hash %x", block.Hash())
	err := DefaultLedger.Store.SaveBlock(block, DefaultLedger)
	if err != nil {
		log.Warn("Save block failure , ", err)
		return err
	}

	return nil
}

func (bc *Blockchain) ContainsTransaction(hash Uint256) bool {
	//TODO: implement error catch
	_, err := DefaultLedger.Store.GetTransaction(hash)
	if err != nil {
		return false
	}
	return true
}

func (bc *Blockchain) GetBookKeepersByTXs(others []*types.Transaction) []*crypto.PubKey {
	//TODO: GetBookKeepers()
	//TODO: Just for TestUse

	return StandbyBookKeepers
}

func (bc *Blockchain) GetBookKeepers() []*crypto.PubKey {
	//TODO: GetBookKeepers()
	//TODO: Just for TestUse

	return StandbyBookKeepers
}

func (bc *Blockchain) CurrentBlockHash() Uint256 {
	return DefaultLedger.Store.GetCurrentBlockHash()
}
