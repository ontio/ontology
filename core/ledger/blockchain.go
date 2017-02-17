package ledger

import (
	. "GoOnchain/common"
	tx "GoOnchain/core/transaction"
	"GoOnchain/crypto"
	"GoOnchain/events"
	"sync"
	"GoOnchain/core/asset"
	. "GoOnchain/errors"
	"errors"
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

func NewBlockchainWithGenesisBlock() *Blockchain {
	blockchain := NewBlockchain()
	blockchain.AddBlock(GenesisBlockInit())

	return blockchain
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
	bc.BCEvents.Notify(events.EventBlockPersistCompleted, block)

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

func (bc *Blockchain) GetAsset(assetId Uint256) *asset.Asset {
	asset, _:= DefaultLedger.Store.GetAsset(assetId)
	return asset
}

func (bc *Blockchain) GetBlockWithHeight(height uint32) (*Block, error) {
	temp := DefaultLedger.Store.GetBlockHash(height)
	bk, err := DefaultLedger.Store.GetBlock(temp)
	if err != nil{
		return nil, NewDetailErr(errors.New("[Blockchain], GetBlockWithHeight failed."), ErrNoCode, "")
	}
	return bk, nil
}

func (bc *Blockchain) GetBlockWithHash(hash Uint256) (*Block, error) {
	bk, err := DefaultLedger.Store.GetBlock(hash)
	if err != nil{
		return nil, NewDetailErr(errors.New("[Blockchain], GetBlockWithHash failed."), ErrNoCode, "")
	}
	return bk, nil
}