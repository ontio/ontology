package vbft

import (
	"fmt"
	"math"

	. "github.com/Ontology/common"
	"github.com/Ontology/consensus/actor"
	"github.com/Ontology/core/ledger"
)

type ChainInfo struct {
	Num      uint64  `json:"num"`
	Hash     Uint256 `json:"hash"`
	Proposer uint32  `json:"proposer"`
}

type ChainStore struct {
	Info *ChainInfo
	db   *actor.LedgerActor
}

func OpenBlockStore(ledger *actor.LedgerActor) (*ChainStore, error) {
	info := &ChainInfo{
		Num: math.MaxUint64, // as invalid blockNum
	}
	return &ChainStore{
		Info: info,
		db:   ledger,
	}, nil
}

func (self *ChainStore) Close() {
	// TODO: any action on ledger actor??
	self.Info = nil
}

func (self *ChainStore) GetChainedBlockNum() uint64 {
	return self.Info.Num
}

func (self *ChainStore) setChainInfo(blockNum uint64, blockHash Uint256, proposer uint32) error {
	info := &ChainInfo{
		Num:      blockNum,
		Hash:     blockHash,
		Proposer: proposer,
	}

	if blockNum > self.GetChainedBlockNum() {
		self.Info = info
	}

	return nil
}

func (self *ChainStore) AddBlock(block *Block, blockHash Uint256) error {
	if block == nil {
		return fmt.Errorf("try add nil block")
	}

	if err := ledger.DefLedger.AddBlock(block.Block); err != nil {
		return fmt.Errorf("ledger add blk failed: %s", err)
	}

	// update chain Info
	return self.setChainInfo(block.getBlockNum(), blockHash, block.getProposer())
}

//
// SetBlock is used when recovering from fork-chain
//
func (self *ChainStore) SetBlock(block *Block, blockHash Uint256) error {

	if err := ledger.DefLedger.AddBlock(block.Block); err != nil {
		return fmt.Errorf("ledger failed to add block: %s", err)
	}

	if uint64(block.getBlockNum()) == self.Info.Num || self.Info.Num == math.MaxUint64 {
		return self.setChainInfo(block.getBlockNum(), blockHash, block.getProposer())
	}

	return nil
}

func (self *ChainStore) GetBlock(blockNum uint64) (*Block, error) {

	block, err := ledger.DefLedger.GetBlockByHeight(uint32(blockNum))
	if err != nil {
		return nil, err
	}

	return initVbftBlock(block)
}
