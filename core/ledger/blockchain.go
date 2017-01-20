package ledger

import (
	tx "GoOnchain/core/transaction"
	"GoOnchain/common"
	"sync"
	"GoOnchain/events"
	"GoOnchain/crypto"
)


// Store provides storage for State data
type BlockchainStore interface {
	//TODO: define the state store func
	SaveBlock(*Block) error
}

const (
	EventBlockPersistCompleted events.EventType = iota
)

type Blockchain struct {
	Store BlockchainStore
	TxPool tx.TransactionPool
	Transactions []*tx.Transaction

	BlockCache map[common.Uint256]*Block

	BlockHeight uint32

	BCEvents *events.Event
	mutex sync.Mutex

}

func NewBlockchain() *Blockchain {
	return &Blockchain{
		BlockCache: make(map[common.Uint256]*Block),
		BCEvents: events.NewEvent(),
	}
}

func (bc *Blockchain) AddBlock(block *Block) error {
	//TODO: implement AddBlock


	//set block cache
	bc.AddBlockCache(block)


	//Block header verfiy

	//save block
	err := bc.SaveBlock(block)
	if err != nil {return err}

	return nil
}


func (bc *Blockchain) AddBlockCache(block *Block)  {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	if _,ok := bc.BlockCache[block.GetHash()]; !ok{
		bc.BlockCache[block.GetHash()] = block
	}
}

func (bc *Blockchain) ContainsBlock(hash common.Uint256) bool {
	//TODO: implement ContainsBlock
	return false
}

func (bc *Blockchain) GetHeader(hash common.Uint256) *Header {
	//TODO: implement GetHeader
	return nil
}

func (bc *Blockchain) SaveBlock(block *Block) error {
	//TODO: implement PersistBlock

	err := bc.Store.SaveBlock(block)
	if err != nil {return err}

	bc.BCEvents.Notify(EventBlockPersistCompleted,block)

	return nil
}

func (bc *Blockchain) ContainsTransaction(hash common.Uint256) bool {
	//TODO: implement ContainsTransaction
	return false
}

func (bc *Blockchain) GetMinersByTXs(others []*tx.Transaction) []*crypto.PubKey{
	//TODO: GetMiners()
	return nil
}

func (bc *Blockchain) GetMiners() []*crypto.PubKey{
	//TODO: GetMiners()
	return nil
}

func (bc *Blockchain) CurrentBlockHash() common.Uint256{
	//TODO: CurrentBlockHash()
	return common.Uint256{}
}