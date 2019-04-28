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
	"github.com/ontio/ontology/events/message"
	"github.com/ontio/ontology/smartcontract/event"
	cstate "github.com/ontio/ontology/smartcontract/states"
)

var DefLedger *Ledger
var DefLedgerMgr *LedgerMgr

type Ledger struct {
	ShardID          types.ShardID
	ParentLedger     *Ledger
	ParentBlockCache *ledgerstore.BlockCacheStore
	ldgStore         store.LedgerStore
	ChildLedger      *Ledger
}

type LedgerMgr struct {
	Lock    sync.RWMutex
	Ledgers map[types.ShardID]*Ledger
}

func init() {
	DefLedgerMgr = &LedgerMgr{
		Ledgers: make(map[types.ShardID]*Ledger),
	}
}

//
// NewLedger : initialize ledger for main-chain
//
func NewLedger(dataDir string, stateHashHeight uint32) (*Ledger, error) {
	dbPath := path.Join(dataDir, fmt.Sprintf("shard_%d", config.DEFAULT_SHARD_ID))
	ldgStore, err := ledgerstore.NewLedgerStore(dbPath, stateHashHeight)
	if err != nil {
		return nil, fmt.Errorf("NewLedgerStore error %s", err)
	}
	lgr := &Ledger{
		ShardID:  types.NewShardIDUnchecked(config.DEFAULT_SHARD_ID),
		ldgStore: ldgStore,
	}

	DefLedgerMgr.Lock.Lock()
	defer DefLedgerMgr.Lock.Unlock()
	DefLedgerMgr.Ledgers[lgr.ShardID] = lgr

	return lgr, nil
}

//
// NewLedger : initialize ledger for shard-chain
//
func NewShardLedger(shardID types.ShardID, dataDir string, mainLedger *Ledger) (*Ledger, error) {
	if shardID.ToUint64() == config.DEFAULT_SHARD_ID {
		return mainLedger, nil
	}

	// load parent ledger
	var parentLedger *Ledger
	var err error
	for shardID.ParentID().ToUint64() != config.DEFAULT_SHARD_ID {
		parentLedger, err = NewShardLedger(shardID.ParentID(), dataDir, mainLedger)
		if err != nil {
			return nil, fmt.Errorf("failed to load shard ledger %d: %s", shardID.ParentID(), err)
		}
	}
	if parentLedger == nil {
		parentLedger = mainLedger
	}

	// load shard ledger
	dbPath := path.Join(dataDir, fmt.Sprintf("shard_%d", shardID.ToUint64()))
	ldgStore, err := ledgerstore.NewLedgerStore(dbPath, 0)
	if err != nil {
		return nil, fmt.Errorf("NewLedgerStore error %s", err)
	}

	// init parent block cache
	parentBlockCache, err := ledgerstore.ResetBlockCacheStore(shardID.ParentID(), dataDir)
	if err != nil {
		return nil, fmt.Errorf("reset shard %d parent blockcache failed: %s", shardID, err)
	}

	lgr := &Ledger{
		ShardID:          shardID,
		ParentLedger:     parentLedger,
		ParentBlockCache: parentBlockCache,
		ldgStore:         ldgStore,
	}
	parentLedger.ChildLedger = lgr
	DefLedgerMgr.Lock.Lock()
	defer DefLedgerMgr.Lock.Unlock()
	DefLedgerMgr.Ledgers[lgr.ShardID] = lgr

	return lgr, nil
}

func GetShardLedger(shardID types.ShardID) *Ledger {
	DefLedgerMgr.Lock.RLock()
	defer DefLedgerMgr.Lock.RUnlock()

	return DefLedgerMgr.Ledgers[shardID]
}

func CloseLedgers() {
	DefLedgerMgr.Lock.Lock()
	defer DefLedgerMgr.Lock.Unlock()
	for _, lgr := range DefLedgerMgr.Ledgers {
		lgr.Close()
	}
	DefLedgerMgr.Ledgers = make(map[types.ShardID]*Ledger)
}

func (self *Ledger) GetStore() store.LedgerStore {
	return self.ldgStore
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
	if len(DefLedgerMgr.Ledgers) > 1 && self.ShardID.IsRootShard() {
		return self.ChildLedger.ParentBlockCache.PutBlock(block, stateMerkleRoot)
	} else {
		err := self.ldgStore.AddBlock(block, stateMerkleRoot)
		if err != nil {
			log.Errorf("Ledger AddBlock BlockHeight:%d BlockHash:%x error:%s", block.Header.Height, block.Hash(), err)
		}
		return err
	}
}

func (self *Ledger) ExecuteBlock(b *types.Block) (store.ExecuteResult, error) {
	if !self.ShardID.IsRootShard() {
		parentBlock, merkleRoot, err := self.ParentBlockCache.GetBlock(b.Header.ParentHeight)
		if err != nil {
			log.Errorf("Ledger ExecuteBlock GetBlock sharad height:%d,ParentHeight:%d error:%s", b.Header.Height, b.Header.ParentHeight, err)
			return store.ExecuteResult{}, err
		}
		result, err := self.ParentLedger.ldgStore.ExecuteBlock(parentBlock)
		if err != nil {
			return result, err
		}
		if merkleRoot != result.MerkleRoot {
			log.Errorf("ExecuteBlock check parentblock cache MerkleRoot blocknum:%d,MerkleRoot:%s,execute MerkleRoot:%s", b.Header.ParentHeight, merkleRoot.ToHexString(), result.MerkleRoot.ToHexString())
			return store.ExecuteResult{}, fmt.Errorf("merkleroot not match")
		}
		self.ParentBlockCache.SaveBlockExecuteResult(b.Header.ParentHeight, result)
	}
	return self.ldgStore.ExecuteBlock(b)
}

func (self *Ledger) SubmitBlock(b *types.Block, exec store.ExecuteResult) error {
	if !self.ShardID.IsRootShard() {
		parentBlock, _, err := self.ParentBlockCache.GetBlock(b.Header.ParentHeight)
		if err != nil {
			log.Errorf("Ledger SubmitBlock GetBlock sharad height:%d,ParentHeight:%d error:%s", b.Header.Height, b.Header.ParentHeight, err)
			return err
		}
		result, err := self.ParentBlockCache.GetBlockExecuteResult(b.Header.ParentHeight)
		if err != nil {
			return err
		}
		if err := self.ParentLedger.ldgStore.SubmitBlock(parentBlock, result); err != nil {
			return err
		} else {
			self.ParentBlockCache.DelBlock(b.Header.ParentHeight)
		}
	}
	err := self.ldgStore.SubmitBlock(b, exec)
	if err != nil {
		log.Errorf("Ledger SubmitBlock BlockHeight:%d BlockHash:%x error:%s", b.Header.Height, b.Hash(), err)
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

func (self *Ledger) Close() error {
	if self.ParentBlockCache != nil {
		self.ParentBlockCache.Close()
	}
	return self.ldgStore.Close()
}

func (self *Ledger) GetParentHeight() uint32 {
	if self.ParentLedger != nil {
		return self.ParentLedger.GetCurrentBlockHeight()
	}
	return 0
}
