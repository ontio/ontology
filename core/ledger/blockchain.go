package ledger

import (
	. "GoOnchain/common"
	tx "GoOnchain/core/transaction"
	"GoOnchain/crypto"
	"GoOnchain/events"
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
		BlockCache: make(map[Uint256]*Block),
		BCEvents:   events.NewEvent(),
	}
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

func (bc *Blockchain) GetHeader(hash Uint256) *Header {
	//TODO: implement GetHeader
	return nil
}

func (bc *Blockchain) SaveBlock(block *Block) error {
	//TODO: implement PersistBlock
	err := DefaultLedger.Store.SaveBlock(block)
	if err != nil {
		return err
	}
	bc.BCEvents.Notify(EventBlockPersistCompleted, block)

	return nil
}

func (bc *Blockchain) ContainsTransaction(hash Uint256) bool {
	//TODO: implement ContainsTransaction
	return false
}

func (bc *Blockchain) GetMinersByTXs(others []*tx.Transaction) []*crypto.PubKey {
	//TODO: GetMiners()
	return nil
}

func (bc *Blockchain) GetMiners() []*crypto.PubKey {
	//TODO: GetMiners()
	return nil
}

func (bc *Blockchain) CurrentBlockHash() Uint256 {
	//TODO: CurrentBlockHash()
	return Uint256{}
}
