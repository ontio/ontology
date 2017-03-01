package ledger

import (
	. "GoOnchain/common"
	tx "GoOnchain/core/transaction"
	"GoOnchain/crypto"
	. "GoOnchain/errors"
	"GoOnchain/events"
	"errors"
	"sync"
)

const (
	EventBlockPersistCompleted events.EventType = iota
)

type Blockchain struct {
	BlockCache  map[Uint256]*Block
	BlockHeight uint32
	BCEvents    *events.Event
	mutex       sync.Mutex
}

func NewBlockchain() *Blockchain {
	return &Blockchain{
		BlockHeight: 0,
		BlockCache: make(map[Uint256]*Block),
		BCEvents:   events.NewEvent(),
	}
}

func NewBlockchainWithGenesisBlock() (*Blockchain,error) {
	blockchain := NewBlockchain()
	genesisBlock,err:=GenesisBlockInit()
	if err != nil{
		return nil,NewDetailErr(err, ErrNoCode, "[Blockchain], NewBlockchainWithGenesisBlock failed.")
	}
	genesisBlock.RebuildMerkleRoot()
	hashx :=genesisBlock.Hash()
	genesisBlock.hash = &hashx
	blockchain.AddBlock(genesisBlock)
	return blockchain,nil
}

func (bc *Blockchain) AddBlock(block *Block) error {
	//TODO: implement AddBlock

	//set block cache
	bc.AddBlockCache(block)

	//Block header verfiy

	//save block
	err := bc.SaveBlock(block)
	if err != nil {
		return err
	}

	return nil
}

func (bc *Blockchain) AddBlockCache(block *Block) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	if _, ok := bc.BlockCache[block.Hash()]; !ok {
		bc.BlockCache[block.Hash()] = block
	}
}

func (bc *Blockchain) ContainsBlock(hash Uint256) bool {
	//TODO: implement ContainsBlock
	return false
}

func (bc *Blockchain) GetHeader(hash Uint256) (*Header,error) {
	 header,err:=DefaultLedger.Store.GetHeader(hash)
	if err != nil{
		return nil, NewDetailErr(errors.New("[Blockchain], GetHeader failed."), ErrNoCode, "")
	}
	return header,nil
}

func (bc *Blockchain) SaveBlock(block *Block) error {
	err := DefaultLedger.Store.SaveBlock(block)
	if err != nil {
		return err
	}
	bc.BCEvents.Notify(EventBlockPersistCompleted, block)

	return nil
}

func (bc *Blockchain) ContainsTransaction(hash Uint256) bool {
	//TODO: implement error catch
	tx ,_ := DefaultLedger.Store.GetTransaction(hash)
	if tx != nil{
		return true
	}
	return false
}

func (bc *Blockchain) GetMinersByTXs(others []*tx.Transaction) []*crypto.PubKey {
	//TODO: GetMiners()
	//TODO: Just for TestUse

	return StandbyMiners
}

func (bc *Blockchain) GetMiners() []*crypto.PubKey {
	//TODO: GetMiners()
	//TODO: Just for TestUse

	return StandbyMiners
}

func (bc *Blockchain) CurrentBlockHash() Uint256 {
	return DefaultLedger.Store.GetCurrentBlockHash()
}

