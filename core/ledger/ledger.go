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

package ledger

import (
	"fmt"
	"path"
	"sync"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/store"
	"github.com/ontio/ontology/core/store/ledgerstore"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/core/xshard_types"
	"github.com/ontio/ontology/events/message"
	"github.com/ontio/ontology/smartcontract/event"
	cstate "github.com/ontio/ontology/smartcontract/states"
)

var DefLedger *Ledger
var DefLedgerMgr *LedgerMgr

type Ledger struct {
	ShardID      common.ShardID
	ParentLedger *Ledger
	ldgStore     store.LedgerStore
	cshardStore  store.CrossShardStore
	ChildLedger  *Ledger
}

type LedgerMgr struct {
	Lock    sync.RWMutex
	Ledgers map[common.ShardID]*Ledger
}

func init() {
	DefLedgerMgr = &LedgerMgr{
		Ledgers: make(map[common.ShardID]*Ledger),
	}
}

//
// NewLedger : initialize ledger for main-chain
//
func NewLedger(dataDir string, stateHashHeight uint32) (*Ledger, error) {
	dbPath := path.Join(dataDir, fmt.Sprintf("shard_%d", config.DEFAULT_SHARD_ID))
	ldgStore, err := ledgerstore.NewLedgerStore(dbPath, stateHashHeight, nil)
	if err != nil {
		return nil, fmt.Errorf("NewLedgerStore error %s", err)
	}
	cshardStore, err := ledgerstore.NewCrossShardStore(dbPath)
	if err != nil {
		return nil, fmt.Errorf("NewCrossShardStore error %s", err)
	}
	lgr := &Ledger{
		ShardID:     common.NewShardIDUnchecked(config.DEFAULT_SHARD_ID),
		ldgStore:    ldgStore,
		cshardStore: cshardStore,
	}

	DefLedgerMgr.Lock.Lock()
	defer DefLedgerMgr.Lock.Unlock()
	DefLedgerMgr.Ledgers[lgr.ShardID] = lgr

	return lgr, nil
}

//
// NewLedger : initialize ledger for shard-chain
//
func NewShardLedger(shardID common.ShardID, dataDir string, mainLedger *Ledger) (*Ledger, error) {
	if shardID.ToUint64() == config.DEFAULT_SHARD_ID {
		return mainLedger, nil
	}

	// load parent ledger
	var parentLedger *Ledger
	var err error
	for shardID.ParentID().ToUint64() != config.DEFAULT_SHARD_ID {
		parentLedger, err = NewShardLedger(shardID.ParentID(), dataDir, mainLedger)
		if err != nil {
			return nil, fmt.Errorf("failed to load shard %d ledger %d: %s", shardID, shardID.ParentID(), err)
		}
	}
	if parentLedger == nil {
		parentLedger = mainLedger
	}

	// load shard ledger
	dbPath := path.Join(dataDir, fmt.Sprintf("shard_%d", shardID.ToUint64()))
	ldgStore, err := ledgerstore.NewLedgerStore(dbPath, 0, parentLedger.ldgStore)
	if err != nil {
		return nil, fmt.Errorf("NewLedgerStore %d error %s", shardID, err)
	}
	cshardStore, err := ledgerstore.NewCrossShardStore(dbPath)
	if err != nil {
		return nil, fmt.Errorf("NewCrossShardStore %d error %s", shardID, err)
	}

	lgr := &Ledger{
		ShardID:      shardID,
		ParentLedger: parentLedger,
		ldgStore:     ldgStore,
		cshardStore:  cshardStore,
	}
	parentLedger.ChildLedger = lgr
	DefLedgerMgr.Lock.Lock()
	defer DefLedgerMgr.Lock.Unlock()
	DefLedgerMgr.Ledgers[lgr.ShardID] = lgr

	return lgr, nil
}

func GetShardLedger(shardID common.ShardID) *Ledger {
	DefLedgerMgr.Lock.RLock()
	defer DefLedgerMgr.Lock.RUnlock()

	return DefLedgerMgr.Ledgers[shardID]
}

// Note: helper for test
func RemoveLedger(shardID common.ShardID) {
	DefLedgerMgr.Lock.RLock()
	defer DefLedgerMgr.Lock.RUnlock()

	delete(DefLedgerMgr.Ledgers, shardID)
}

func CloseLedgers() {
	DefLedgerMgr.Lock.Lock()
	defer DefLedgerMgr.Lock.Unlock()
	for _, lgr := range DefLedgerMgr.Ledgers {
		lgr.Close()
	}
	DefLedgerMgr.Ledgers = make(map[common.ShardID]*Ledger)
}

func (self *Ledger) GetStore() store.LedgerStore {
	return self.ldgStore
}

func (self *Ledger) GetCrossShardStore() store.CrossShardStore {
	return self.cshardStore
}

func (self *Ledger) Init(defaultBookkeeper []keypair.PublicKey, genesisBlock *types.Block) error {
	err := self.ldgStore.InitLedgerStoreWithGenesisBlock(genesisBlock, defaultBookkeeper)
	if err != nil {
		return fmt.Errorf("InitLedgerStoreWithGenesisBlock error %s", err)
	}
	return nil
}

func (self *Ledger) AddHeaders(headers []*types.Header) error {
	return self.ldgStore.AddHeaders(headers)
}

func (self *Ledger) AddBlock(block *types.Block, stateMerkleRoot common.Uint256) error {
	if block.Header.ShardID != self.ShardID {
		return fmt.Errorf("add block from shard %v on ledger %v", block.Header.ShardID, self.ShardID)
	}

	if self.ParentLedger != nil {
		currentParentHeight := self.ParentLedger.GetCurrentBlockHeight()
		if block.Header.ParentHeight > currentParentHeight {
			return fmt.Errorf("failed to add block(%d, %d) with parent height %d", block.Header.Height, block.Header.ShardID, currentParentHeight)
		}
	}

	// FIXME:
	// 1. ExecuteBlock/SubmitBlock requires saving block
	// 2. lock released and re-acquired after ExecuteBlock(block)
	execResult, err := self.ExecuteBlock(block)
	if err != nil {
		return fmt.Errorf("shard %d execute block %d failed: %s", self.ShardID, block.Header.Height, err)
	}

	if err := self.SubmitBlock(block, execResult); err != nil {
		return fmt.Errorf("shard %d submit block %d failed: %s", self.ShardID, block.Header.Height, err)
	}

	return nil
}

func (self *Ledger) ExecuteBlock(b *types.Block) (store.ExecuteResult, error) {
	if b.Header.ShardID != self.ShardID {
		return store.ExecuteResult{}, fmt.Errorf("execute block from shard %v on ledger %v", b.Header.ShardID, self.ShardID)
	}

	if self.ParentLedger != nil {
		currentParentHeight := self.ParentLedger.GetCurrentBlockHeight()
		if b.Header.ParentHeight > currentParentHeight {
			return store.ExecuteResult{}, fmt.Errorf("failed to execute block(%d, %d) with parent height %d", b.Header.Height, b.Header.ShardID, currentParentHeight)
		}
	}

	return self.ldgStore.ExecuteBlock(b)
}

func (self *Ledger) SubmitBlock(b *types.Block, exec store.ExecuteResult) error {
	if b.Header.ShardID != self.ShardID {
		return fmt.Errorf("submit block from shard %v on ledger %v", b.Header.ShardID, self.ShardID)
	}

	if self.ParentLedger != nil {
		currentParentHeight := self.ParentLedger.GetCurrentBlockHeight()
		if b.Header.ParentHeight > currentParentHeight {
			return fmt.Errorf("failed to submit block(%d, %d) with parent height %d", b.Header.Height, b.Header.ShardID, currentParentHeight)
		}
	}

	err := self.ldgStore.SubmitBlock(b, exec)
	if err != nil {
		log.Errorf("Ledger %d SubmitBlock BlockHeight:%d BlockHash:%x error:%s", self.ShardID, b.Header.Height, b.Hash(), err)
		return err
	}
	return nil
}

func (self *Ledger) GetStateMerkleRoot(height uint32) (result common.Uint256, err error) {
	return self.ldgStore.GetStateMerkleRoot(height)
}

func (self *Ledger) GetBlockRootWithNewTxRoots(startHeight uint32, txRoots []common.Uint256) common.Uint256 {
	return self.ldgStore.GetBlockRootWithNewTxRoots(startHeight, txRoots)
}

func (self *Ledger) GetBlockByHeight(height uint32) (*types.Block, error) {
	return self.ldgStore.GetBlockByHeight(height)
}

func (self *Ledger) GetBlockByHash(blockHash common.Uint256) (*types.Block, error) {
	return self.ldgStore.GetBlockByHash(blockHash)
}

func (self *Ledger) GetHeaderByHeight(height uint32) (*types.Header, error) {
	return self.ldgStore.GetHeaderByHeight(height)
}

func (self *Ledger) GetHeaderByHash(blockHash common.Uint256) (*types.Header, error) {
	return self.ldgStore.GetHeaderByHash(blockHash)
}
func (self *Ledger) GetRawHeaderByHash(blockHash common.Uint256) (*types.RawHeader, error) {
	return self.ldgStore.GetRawHeaderByHash(blockHash)
}

func (self *Ledger) GetBlockHash(height uint32) common.Uint256 {
	return self.ldgStore.GetBlockHash(height)
}

func (self *Ledger) GetTransaction(txHash common.Uint256) (*types.Transaction, error) {
	tx, _, err := self.ldgStore.GetTransaction(txHash)
	return tx, err
}

func (self *Ledger) GetTransactionWithHeight(txHash common.Uint256) (*types.Transaction, uint32, error) {
	return self.ldgStore.GetTransaction(txHash)
}

func (self *Ledger) GetCurrentBlockHeight() uint32 {
	return self.ldgStore.GetCurrentBlockHeight()
}

func (self *Ledger) GetCurrentBlockHash() common.Uint256 {
	return self.ldgStore.GetCurrentBlockHash()
}

func (self *Ledger) GetCurrentHeaderHeight() uint32 {
	return self.ldgStore.GetCurrentHeaderHeight()
}

func (self *Ledger) GetCurrentHeaderHash() common.Uint256 {
	return self.ldgStore.GetCurrentHeaderHash()
}

func (self *Ledger) IsContainTransaction(txHash common.Uint256) (bool, error) {
	return self.ldgStore.IsContainTransaction(txHash)
}

func (self *Ledger) IsContainBlock(blockHash common.Uint256) (bool, error) {
	return self.ldgStore.IsContainBlock(blockHash)
}

func (self *Ledger) GetCurrentStateRoot() (common.Uint256, error) {
	return common.Uint256{}, nil
}

func (self *Ledger) GetBookkeeperState() (*states.BookkeeperState, error) {
	return self.ldgStore.GetBookkeeperState()
}

func (self *Ledger) GetStorageItem(codeHash common.Address, key []byte) ([]byte, error) {
	storageKey := &states.StorageKey{
		ContractAddress: codeHash,
		Key:             key,
	}
	storageItem, err := self.ldgStore.GetStorageItem(storageKey)
	if err != nil {
		return nil, err
	}
	if storageItem == nil {
		return nil, nil
	}
	return storageItem.Value, nil
}

func (self *Ledger) GetContractState(contractHash common.Address) (*payload.DeployCode, error) {
	return self.ldgStore.GetContractState(contractHash)
}

func (self *Ledger) GetMerkleProof(proofHeight, rootHeight uint32) ([]common.Uint256, error) {
	return self.ldgStore.GetMerkleProof(proofHeight, rootHeight)
}

func (self *Ledger) PreExecuteContract(tx *types.Transaction) (*cstate.PreExecResult, error) {
	return self.ldgStore.PreExecuteContract(tx)
}

func (self *Ledger) GetEventNotifyByTx(tx common.Uint256) (*event.ExecuteNotify, error) {
	return self.ldgStore.GetEventNotifyByTx(tx)
}

func (self *Ledger) GetEventNotifyByBlock(height uint32) ([]*event.ExecuteNotify, error) {
	return self.ldgStore.GetEventNotifyByBlock(height)
}

func (self *Ledger) GetBlockShardEvents(height uint32) (events []*message.ShardSystemEventMsg, err error) {
	return self.ldgStore.GetBlockShardEvents(height)
}

func (self *Ledger) GetShardMsgsInBlock(blockHeight uint32, shardID common.ShardID) ([]xshard_types.CommonShardMsg, error) {
	return self.ldgStore.GetShardMsgsInBlock(blockHeight, shardID)
}

func (self *Ledger) GetRelatedShardIDsInBlock(blockHeight uint32) ([]common.ShardID, error) {
	return self.ldgStore.GetRelatedShardIDsInBlock(blockHeight)
}

func (self *Ledger) GetContractMetaDataEvent(blockHeight uint32, contractAddr common.Address) (*payload.MetaDataCode, error) {
	return self.ldgStore.GetParentMetaData(blockHeight, contractAddr)
}

func (self *Ledger) GetParentContract(blockHeight uint32, addr common.Address) (*payload.DeployCode, error) {
	return self.ldgStore.GetParentContract(blockHeight, addr)
}

func (self *Ledger) GetShardConsensusConfig(shardID common.ShardID, height uint32) ([]byte, error) {
	return self.ldgStore.GetShardConsensusConfig(shardID, height)
}

func (self *Ledger) SaveCrossShardMsgByHash(msgHash common.Uint256, crossShardMsg *types.CrossShardMsg) error {
	return self.cshardStore.SaveCrossShardMsgByHash(msgHash, crossShardMsg)
}
func (self *Ledger) GetCrossShardMsgByHash(msgHash common.Uint256) (*types.CrossShardMsg, error) {
	return self.cshardStore.GetCrossShardMsgByHash(msgHash)
}

func (self *Ledger) SaveAllShardIDs(shardIDs []common.ShardID) error {
	return self.cshardStore.SaveAllShardIDs(shardIDs)
}
func (self *Ledger) GetAllShardIDs() ([]common.ShardID, error) {
	return self.cshardStore.GetAllShardIDs()
}

func (self *Ledger) SaveCrossShardHash(shardID common.ShardID, msgHash common.Uint256) error {
	return self.cshardStore.SaveCrossShardHash(shardID, msgHash)
}
func (self *Ledger) GetCrossShardHash(shardID common.ShardID) (common.Uint256, error) {
	return self.cshardStore.GetCrossShardHash(shardID)
}

func (self *Ledger) SaveShardMsgHash(shardID common.ShardID, msgHash common.Uint256) error {
	return self.cshardStore.SaveShardMsgHash(shardID, msgHash)
}

func (self *Ledger) GetShardMsgHash(shardID common.ShardID) (common.Uint256, error) {
	return self.cshardStore.GetShardMsgHash(shardID)
}

func (self *Ledger) Close() error {
	err := self.ldgStore.Close()
	if err != nil {
		return err
	}
	return self.cshardStore.Close()
}

func (self *Ledger) GetParentHeight() uint32 {
	if self.ParentLedger != nil {
		return self.ParentLedger.GetCurrentBlockHeight()
	}
	return 0
}
