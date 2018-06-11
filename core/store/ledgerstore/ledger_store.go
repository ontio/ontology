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
//Storage of ledger
package ledgerstore

import (
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/core/states"
	scom "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/statestore"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/events"
	"github.com/ontio/ontology/events/message"
	"github.com/ontio/ontology/smartcontract"
	scommon "github.com/ontio/ontology/smartcontract/common"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/neovm"
	sstate "github.com/ontio/ontology/smartcontract/states"
	"github.com/ontio/ontology/smartcontract/storage"
	"math"
	"os"
	"sort"
	"strings"
	"sync"
)

const (
	SYSTEM_VERSION          = byte(1)      //Version of ledger store
	HEADER_INDEX_BATCH_SIZE = uint32(2000) //Bath size of saving header index
)

var (
	//Storage save path.
	DBDirEvent          = "ledgerevent"
	DBDirBlock          = "block"
	DBDirState          = "states"
	MerkleTreeStorePath = "merkle_tree.db"
)

//LedgerStoreImp is main store struct fo ledger
type LedgerStoreImp struct {
	blockStore       *BlockStore                      //BlockStore for saving block & transaction data
	stateStore       *StateStore                      //StateStore for saving state data, like balance, smart contract execution result, and so on.
	eventStore       *EventStore                      //EventStore for saving log those gen after smart contract executed.
	storedIndexCount uint32                           //record the count of have saved block index
	currBlockHeight  uint32                           //Current block height
	currBlockHash    common.Uint256                   //Current block hash
	headerCache      map[common.Uint256]*types.Header //BlockHash => Header
	headerIndex      map[uint32]common.Uint256        //Header index, Mapping header height => block hash
	savingBlock      bool                             //is saving block now
	lock             sync.RWMutex
}

//NewLedgerStore return LedgerStoreImp instance
func NewLedgerStore(dataDir string) (*LedgerStoreImp, error) {
	ledgerStore := &LedgerStoreImp{
		headerIndex: make(map[uint32]common.Uint256),
		headerCache: make(map[common.Uint256]*types.Header, 0),
	}

	blockStore, err := NewBlockStore(fmt.Sprintf("%s%s%s", dataDir, string(os.PathSeparator), DBDirBlock), true)
	if err != nil {
		return nil, fmt.Errorf("NewBlockStore error %s", err)
	}
	ledgerStore.blockStore = blockStore

	stateStore, err := NewStateStore(fmt.Sprintf("%s%s%s", dataDir, string(os.PathSeparator), DBDirState),
		fmt.Sprintf("%s%s%s", dataDir, string(os.PathSeparator), MerkleTreeStorePath))
	if err != nil {
		return nil, fmt.Errorf("NewStateStore error %s", err)
	}
	ledgerStore.stateStore = stateStore

	eventState, err := NewEventStore(fmt.Sprintf("%s%s%s", dataDir, string(os.PathSeparator), DBDirEvent))
	if err != nil {
		return nil, fmt.Errorf("NewEventStore error %s", err)
	}
	ledgerStore.eventStore = eventState

	return ledgerStore, nil
}

//InitLedgerStoreWithGenesisBlock init the ledger store with genesis block. It's the first operation after NewLedgerStore.
func (this *LedgerStoreImp) InitLedgerStoreWithGenesisBlock(genesisBlock *types.Block, defaultBookkeeper []keypair.PublicKey) error {
	hasInit, err := this.hasAlreadyInitGenesisBlock()
	if err != nil {
		return fmt.Errorf("hasAlreadyInit error %s", err)
	}
	if !hasInit {
		err = this.blockStore.ClearAll()
		if err != nil {
			return fmt.Errorf("blockStore.ClearAll error %s", err)
		}
		err = this.stateStore.ClearAll()
		if err != nil {
			return fmt.Errorf("stateStore.ClearAll error %s", err)
		}
		err = this.eventStore.ClearAll()
		if err != nil {
			return fmt.Errorf("eventStore.ClearAll error %s", err)
		}
		sort.Sort(keypair.NewPublicList(defaultBookkeeper))
		bookkeeperState := &states.BookkeeperState{
			CurrBookkeeper: defaultBookkeeper,
			NextBookkeeper: defaultBookkeeper,
		}
		err = this.stateStore.SaveBookkeeperState(bookkeeperState)
		if err != nil {
			return fmt.Errorf("SaveBookkeeperState error %s", err)
		}
		err = this.saveBlock(genesisBlock)
		if err != nil {
			return fmt.Errorf("save genesis block error %s", err)
		}
		err = this.initGenesisBlock()
		if err != nil {
			return fmt.Errorf("init error %s", err)
		}
		log.Infof("GenesisBlock init success. GenesisBlock hash:%x\n", genesisBlock.Hash())
	} else {
		genesisHash := genesisBlock.Hash()
		exist, err := this.blockStore.ContainBlock(genesisHash)
		if err != nil {
			return fmt.Errorf("HashBlockExist error %s", err)
		}
		if !exist {
			return fmt.Errorf("GenesisBlock arenot init correctly")
		}
		err = this.init()
		if err != nil {
			return fmt.Errorf("init error %s", err)
		}
	}
	return nil
}

func (this *LedgerStoreImp) hasAlreadyInitGenesisBlock() (bool, error) {
	version, err := this.blockStore.GetVersion()
	if err != nil && err != scom.ErrNotFound {
		return false, fmt.Errorf("GetVersion error %s", err)
	}
	return version == SYSTEM_VERSION, nil
}

func (this *LedgerStoreImp) initGenesisBlock() error {
	return this.blockStore.SaveVersion(SYSTEM_VERSION)
}

func (this *LedgerStoreImp) init() error {
	err := this.initCurrentBlock()
	if err != nil {
		return fmt.Errorf("initCurrentBlock error %s", err)
	}
	err = this.initHeaderIndexList()
	if err != nil {
		return fmt.Errorf("initHeaderIndexList error %s", err)
	}
	err = this.initStore()
	if err != nil {
		return fmt.Errorf("initStore error %s", err)
	}
	return nil
}

func (this *LedgerStoreImp) initCurrentBlock() error {
	currentBlockHash, currentBlockHeight, err := this.blockStore.GetCurrentBlock()
	if err != nil {
		return fmt.Errorf("LoadCurrentBlock error %s", err)
	}
	log.Infof("InitCurrentBlock currentBlockHash %x currentBlockHeight %d", currentBlockHash, currentBlockHeight)
	this.currBlockHash = currentBlockHash
	this.currBlockHeight = currentBlockHeight
	return nil
}

func (this *LedgerStoreImp) initHeaderIndexList() error {
	currBlockHeight := this.GetCurrentBlockHeight()
	headerIndex, err := this.blockStore.GetHeaderIndexList()
	if err != nil {
		return fmt.Errorf("LoadHeaderIndexList error %s", err)
	}
	storeIndexCount := uint32(len(headerIndex))
	this.headerIndex = headerIndex
	this.storedIndexCount = storeIndexCount

	for i := storeIndexCount; i <= currBlockHeight; i++ {
		height := i
		blockHash, err := this.blockStore.GetBlockHash(height)
		if err != nil {
			return fmt.Errorf("LoadBlockHash height %d error %s", height, err)
		}
		if blockHash == common.UINT256_EMPTY {
			return fmt.Errorf("LoadBlockHash height %d hash nil", height)
		}
		this.headerIndex[height] = blockHash
	}
	return nil
}

func (this *LedgerStoreImp) initStore() error {
	blockHeight := this.GetCurrentBlockHeight()

	_, stateHeight, err := this.stateStore.GetCurrentBlock()
	if err != nil {
		return fmt.Errorf("stateStore.GetCurrentBlock error %s", err)
	}
	for i := stateHeight; i < blockHeight; i++ {
		blockHash, err := this.blockStore.GetBlockHash(i)
		if err != nil {
			return fmt.Errorf("blockStore.GetBlockHash height:%d error:%s", i, err)
		}
		block, err := this.blockStore.GetBlock(blockHash)
		if err != nil {
			return fmt.Errorf("blockStore.GetBlock height:%d error:%s", i, err)
		}
		this.stateStore.NewBatch()
		err = this.saveBlockToStateStore(block)
		if err != nil {
			return fmt.Errorf("save to state store height:%d error:%s", i, err)
		}
		err = this.stateStore.CommitTo()
		if err != nil {
			return fmt.Errorf("stateStore.CommitTo height:%d error %s", i, err)
		}
	}

	_, eventHeight, err := this.eventStore.GetCurrentBlock()
	if err != nil {
		return fmt.Errorf("eventStore.GetCurrentBlock error:%s", err)
	}
	for i := eventHeight; i < blockHeight; i++ {
		blockHash, err := this.blockStore.GetBlockHash(i)
		if err != nil {
			return fmt.Errorf("blockStore.GetBlockHash height:%d error:%s", i, err)
		}
		block, err := this.blockStore.GetBlock(blockHash)
		if err != nil {
			return fmt.Errorf("blockStore.GetBlock height:%d error:%s", i, err)
		}
		this.eventStore.NewBatch()
		err = this.saveBlockToEventStore(block)
		if err != nil {
			return fmt.Errorf("save to event store height:%d error:%s", i, err)
		}
		err = this.eventStore.CommitTo()
		if err != nil {
			return fmt.Errorf("eventStore.CommitTo height:%d error %s", i, err)
		}
	}
	return nil
}

func (this *LedgerStoreImp) setHeaderIndex(height uint32, blockHash common.Uint256) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.headerIndex[height] = blockHash
}

func (this *LedgerStoreImp) getHeaderIndex(height uint32) common.Uint256 {
	this.lock.RLock()
	defer this.lock.RUnlock()
	blockHash, ok := this.headerIndex[height]
	if !ok {
		return common.Uint256{}
	}
	return blockHash
}

//GetCurrentHeaderHeight return the current header height.
//In block sync states, Header height is usually higher than block height that is has already committed to storage
func (this *LedgerStoreImp) GetCurrentHeaderHeight() uint32 {
	this.lock.RLock()
	defer this.lock.RUnlock()
	size := len(this.headerIndex)
	if size == 0 {
		return 0
	}
	return uint32(size) - 1
}

//GetCurrentHeaderHash return the current header hash. The current header means the latest header.
func (this *LedgerStoreImp) GetCurrentHeaderHash() common.Uint256 {
	this.lock.RLock()
	defer this.lock.RUnlock()
	size := len(this.headerIndex)
	if size == 0 {
		return common.Uint256{}
	}
	return this.headerIndex[uint32(size)-1]
}

func (this *LedgerStoreImp) setCurrentBlock(height uint32, blockHash common.Uint256) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.currBlockHash = blockHash
	this.currBlockHeight = height
	return
}

//GetCurrentBlock return the current block height, and block hash.
//Current block means the latest block in store.
func (this *LedgerStoreImp) GetCurrentBlock() (uint32, common.Uint256) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.currBlockHeight, this.currBlockHash
}

//GetCurrentBlockHash return the current block hash
func (this *LedgerStoreImp) GetCurrentBlockHash() common.Uint256 {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.currBlockHash
}

//GetCurrentBlockHeight return the current block height
func (this *LedgerStoreImp) GetCurrentBlockHeight() uint32 {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.currBlockHeight
}

func (this *LedgerStoreImp) addHeaderCache(header *types.Header) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.headerCache[header.Hash()] = header
}

func (this *LedgerStoreImp) delHeaderCache(blockHash common.Uint256) {
	this.lock.Lock()
	defer this.lock.Unlock()
	delete(this.headerCache, blockHash)
}

func (this *LedgerStoreImp) getHeaderCache(blockHash common.Uint256) *types.Header {
	this.lock.RLock()
	defer this.lock.RUnlock()
	header, ok := this.headerCache[blockHash]
	if !ok {
		return nil
	}
	return header
}

func (this *LedgerStoreImp) verifyHeader(header *types.Header) error {
	if header.Height == 0 {
		return nil
	}

	var prevHeader *types.Header
	prevHeaderHash := header.PrevBlockHash
	prevHeader, err := this.GetHeaderByHash(prevHeaderHash)
	if err != nil && err != scom.ErrNotFound {
		return fmt.Errorf("get prev header error %s", err)
	}
	if prevHeader == nil {
		return fmt.Errorf("cannot find pre header by blockHash %x", prevHeaderHash)
	}

	if prevHeader.Height+1 != header.Height {
		return fmt.Errorf("block height is incorrect")
	}

	if prevHeader.Timestamp >= header.Timestamp {
		return fmt.Errorf("block timestamp is incorrect")
	}
	consensusType := strings.ToLower(config.DefConfig.Genesis.ConsensusType)
	if consensusType != "vbft" {
		address, err := types.AddressFromBookkeepers(header.Bookkeepers)
		if err != nil {
			return err
		}
		if prevHeader.NextBookkeeper != address {
			return fmt.Errorf("bookkeeper address error")
		}

		m := len(header.Bookkeepers) - (len(header.Bookkeepers)-1)/3
		hash := header.Hash()
		err = signature.VerifyMultiSignature(hash[:], header.Bookkeepers, m, header.SigData)
		if err != nil {
			return err
		}
	}
	return nil
}

//AddHeader add header to cache, and add the mapping of block height to block hash. Using in block sync
func (this *LedgerStoreImp) AddHeader(header *types.Header) error {
	nextHeaderHeight := this.GetCurrentHeaderHeight() + 1
	if header.Height != nextHeaderHeight {
		return fmt.Errorf("header height %d not equal next header height %d", header.Height, nextHeaderHeight)
	}
	err := this.verifyHeader(header)
	if err != nil {
		return fmt.Errorf("verifyHeader error %s", err)
	}
	this.addHeaderCache(header)
	this.setHeaderIndex(header.Height, header.Hash())
	return nil
}

//AddHeaders bath add header.
func (this *LedgerStoreImp) AddHeaders(headers []*types.Header) error {
	sort.Slice(headers, func(i, j int) bool {
		return headers[i].Height < headers[j].Height
	})
	var err error
	for _, header := range headers {
		err = this.AddHeader(header)
		if err != nil {
			return err
		}
	}
	return nil
}

//AddBlock add the block to store.
//When the block is not the next block, it will be cache. until the missing block arrived
func (this *LedgerStoreImp) AddBlock(block *types.Block) error {
	currBlockHeight := this.GetCurrentBlockHeight()
	blockHeight := block.Header.Height
	if blockHeight <= currBlockHeight {
		return nil
	}
	nextBlockHeight := currBlockHeight + 1
	if blockHeight != nextBlockHeight {
		return fmt.Errorf("block height %d not equal next block height %d", blockHeight, nextBlockHeight)
	}

	err := this.verifyHeader(block.Header)
	if err != nil {
		return fmt.Errorf("verifyHeader error %s", err)
	}

	err = this.saveBlock(block)
	if err != nil {
		return fmt.Errorf("saveBlock error %s", err)
	}
	this.delHeaderCache(block.Hash())
	return nil
}

func (this *LedgerStoreImp) saveBlockToBlockStore(block *types.Block) error {
	blockHash := block.Hash()
	blockHeight := block.Header.Height

	this.setHeaderIndex(blockHeight, blockHash)
	err := this.saveHeaderIndexList()
	if err != nil {
		return fmt.Errorf("saveHeaderIndexList error %s", err)
	}
	err = this.blockStore.SaveCurrentBlock(blockHeight, blockHash)
	if err != nil {
		return fmt.Errorf("SaveCurrentBlock error %s", err)
	}
	this.blockStore.SaveBlockHash(blockHeight, blockHash)
	err = this.blockStore.SaveBlock(block)
	if err != nil {
		return fmt.Errorf("SaveBlock height %d hash %x error %s", blockHeight, blockHash, err)
	}
	return nil
}

func (this *LedgerStoreImp) saveBlockToStateStore(block *types.Block) error {
	blockHash := block.Hash()
	blockHeight := block.Header.Height

	stateBatch := this.stateStore.NewStateBatch()

	for _, tx := range block.Transactions {
		err := this.handleTransaction(stateBatch, block, tx)
		if err != nil {
			return fmt.Errorf("handleTransaction error %s", err)
		}
	}

	err := this.stateStore.AddMerkleTreeRoot(block.Header.TransactionsRoot)
	if err != nil {
		return fmt.Errorf("AddMerkleTreeRoot error %s", err)
	}

	err = this.stateStore.SaveCurrentBlock(blockHeight, blockHash)
	if err != nil {
		return fmt.Errorf("SaveCurrentBlock error %s", err)
	}
	err = stateBatch.CommitTo()
	if err != nil {
		return fmt.Errorf("stateBatch.CommitTo error %s", err)
	}
	return nil
}

func (this *LedgerStoreImp) saveBlockToEventStore(block *types.Block) error {
	blockHash := block.Hash()
	blockHeight := block.Header.Height
	invokeTxs := make([]common.Uint256, 0)
	for _, tx := range block.Transactions {
		txHash := tx.Hash()
		if tx.TxType == types.Invoke {
			invokeTxs = append(invokeTxs, txHash)
		}
	}
	if len(invokeTxs) > 0 {
		err := this.eventStore.SaveEventNotifyByBlock(block.Header.Height, invokeTxs)
		if err != nil {
			return fmt.Errorf("SaveEventNotifyByBlock error %s", err)
		}
	}
	err := this.eventStore.SaveCurrentBlock(blockHeight, blockHash)
	if err != nil {
		return fmt.Errorf("SaveCurrentBlock error %s", err)
	}
	return nil
}

func (this *LedgerStoreImp) isSavingBlock() bool {
	this.lock.Lock()
	defer this.lock.Unlock()

	if !this.savingBlock {
		this.savingBlock = true
		return false
	}
	return true
}

func (this *LedgerStoreImp) resetSavingBlock() {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.savingBlock = false
}

//saveBlock do the job of execution samrt contract and commit block to store.
func (this *LedgerStoreImp) saveBlock(block *types.Block) error {
	blockHash := block.Hash()
	blockHeight := block.Header.Height
	if this.isSavingBlock() {
		//hash already saved or is saving
		return nil
	}
	defer this.resetSavingBlock()
	if blockHeight > 0 && blockHeight != (this.GetCurrentBlockHeight()+1) {
		return nil
	}

	this.blockStore.NewBatch()
	this.stateStore.NewBatch()
	this.eventStore.NewBatch()
	err := this.saveBlockToBlockStore(block)
	if err != nil {
		return fmt.Errorf("save to block store height:%d error:%s", blockHeight, err)
	}
	err = this.saveBlockToStateStore(block)
	if err != nil {
		return fmt.Errorf("save to state store height:%d error:%s", blockHeight, err)
	}
	err = this.saveBlockToEventStore(block)
	if err != nil {
		return fmt.Errorf("save to event store height:%d error:%s", blockHeight, err)
	}
	err = this.blockStore.CommitTo()
	if err != nil {
		return fmt.Errorf("blockStore.CommitTo height:%d error %s", blockHeight, err)
	}
	err = this.stateStore.CommitTo()
	if err != nil {
		return fmt.Errorf("stateStore.CommitTo height:%d error %s", blockHeight, err)
	}
	err = this.eventStore.CommitTo()
	if err != nil {
		return fmt.Errorf("eventStore.CommitTo height:%d error %s", blockHeight, err)
	}
	this.setCurrentBlock(blockHeight, blockHash)

	if events.DefActorPublisher != nil {
		events.DefActorPublisher.Publish(
			message.TOPIC_SAVE_BLOCK_COMPLETE,
			&message.SaveBlockCompleteMsg{
				Block: block,
			})
	}
	return nil
}

func (this *LedgerStoreImp) handleTransaction(stateBatch *statestore.StateBatch, block *types.Block, tx *types.Transaction) error {
	txHash := tx.Hash()
	switch tx.TxType {
	case types.Deploy:
		err := this.stateStore.HandleDeployTransaction(this, stateBatch, tx, block, this.eventStore)
		if err != nil {
			if stateBatch.Error() == nil {
				log.Debugf("HandleDeployTransaction tx %x error %s", txHash, err)
				SaveNotify(this.eventStore, txHash, []*event.NotifyEventInfo{}, false)
			} else {
				return fmt.Errorf("HandleDeployTransaction tx %x error %s", txHash, stateBatch.Error())
			}
		}
	case types.Invoke:
		err := this.stateStore.HandleInvokeTransaction(this, stateBatch, tx, block, this.eventStore)
		if err != nil {
			if stateBatch.Error() == nil {
				log.Debugf("HandleInvokeTransaction tx %x error %s", txHash, err)
				SaveNotify(this.eventStore, txHash, []*event.NotifyEventInfo{}, false)
			} else {
				return fmt.Errorf("HandleInvokeTransaction tx %x error %s", txHash, stateBatch.Error())
			}
		}
	}
	return nil
}

func (this *LedgerStoreImp) saveHeaderIndexList() error {
	this.lock.RLock()
	storeCount := this.storedIndexCount
	currHeight := this.currBlockHeight
	if currHeight-storeCount < HEADER_INDEX_BATCH_SIZE {
		this.lock.RUnlock()
		return nil
	}

	headerList := make([]common.Uint256, HEADER_INDEX_BATCH_SIZE)
	for i := uint32(0); i < HEADER_INDEX_BATCH_SIZE; i++ {
		height := storeCount + i
		headerList[i] = this.headerIndex[height]
	}
	this.lock.RUnlock()

	err := this.blockStore.SaveHeaderIndexList(storeCount, headerList)
	if err != nil {
		return fmt.Errorf("SaveHeaderIndexList start %d error %s", storeCount, err)
	}

	this.lock.Lock()
	this.storedIndexCount += HEADER_INDEX_BATCH_SIZE
	this.lock.Unlock()
	return nil
}

//IsContainBlock return whether the block is in store
func (this *LedgerStoreImp) IsContainBlock(blockHash common.Uint256) (bool, error) {
	return this.blockStore.ContainBlock(blockHash)
}

//IsContainTransaction return whether the transaction is in store. Wrap function of BlockStore.ContainTransaction
func (this *LedgerStoreImp) IsContainTransaction(txHash common.Uint256) (bool, error) {
	return this.blockStore.ContainTransaction(txHash)
}

//GetBlockRootWithNewTxRoot return the block root(merkle root of blocks) after add a new tx root of block
func (this *LedgerStoreImp) GetBlockRootWithNewTxRoot(txRoot common.Uint256) common.Uint256 {
	return this.stateStore.GetBlockRootWithNewTxRoot(txRoot)
}

//GetBlockHash return the block hash by block height
func (this *LedgerStoreImp) GetBlockHash(height uint32) common.Uint256 {
	return this.getHeaderIndex(height)
}

//GetHeaderByHash return the block header by block hash
func (this *LedgerStoreImp) GetHeaderByHash(blockHash common.Uint256) (*types.Header, error) {
	header := this.getHeaderCache(blockHash)
	if header != nil {
		return header, nil
	}
	return this.blockStore.GetHeader(blockHash)
}

//GetHeaderByHash return the block header by block height
func (this *LedgerStoreImp) GetHeaderByHeight(height uint32) (*types.Header, error) {
	blockHash := this.GetBlockHash(height)
	var empty common.Uint256
	if blockHash == empty {
		return nil, nil
	}
	return this.GetHeaderByHash(blockHash)
}

//GetSysFeeAmount return the sys fee for block by block hash. Wrap function of BlockStore.GetSysFeeAmount
func (this *LedgerStoreImp) GetSysFeeAmount(blockHash common.Uint256) (common.Fixed64, error) {
	return this.blockStore.GetSysFeeAmount(blockHash)
}

//GetTransaction return transaction by transaction hash. Wrap function of BlockStore.GetTransaction
func (this *LedgerStoreImp) GetTransaction(txHash common.Uint256) (*types.Transaction, uint32, error) {
	return this.blockStore.GetTransaction(txHash)
}

//GetBlockByHash return block by block hash. Wrap function of BlockStore.GetBlockByHash
func (this *LedgerStoreImp) GetBlockByHash(blockHash common.Uint256) (*types.Block, error) {
	return this.blockStore.GetBlock(blockHash)
}

//GetBlockByHeight return block by height.
func (this *LedgerStoreImp) GetBlockByHeight(height uint32) (*types.Block, error) {
	blockHash := this.GetBlockHash(height)
	var empty common.Uint256
	if blockHash == empty {
		return nil, nil
	}
	return this.GetBlockByHash(blockHash)
}

//GetBookkeeperState return the bookkeeper state. Wrap function of StateStore.GetBookkeeperState
func (this *LedgerStoreImp) GetBookkeeperState() (*states.BookkeeperState, error) {
	return this.stateStore.GetBookkeeperState()
}

//GetMerkleProof return the block merkle proof. Wrap function of StateStore.GetMerkleProof
func (this *LedgerStoreImp) GetMerkleProof(proofHeight, rootHeight uint32) ([]common.Uint256, error) {
	return this.stateStore.GetMerkleProof(proofHeight, rootHeight)
}

//GetContractState return contract by contract address. Wrap function of StateStore.GetContractState
func (this *LedgerStoreImp) GetContractState(contractHash common.Address) (*payload.DeployCode, error) {
	return this.stateStore.GetContractState(contractHash)
}

//GetStorageItem return the storage value of the key in smart contract. Wrap function of StateStore.GetStorageState
func (this *LedgerStoreImp) GetStorageItem(key *states.StorageKey) (*states.StorageItem, error) {
	return this.stateStore.GetStorageState(key)
}

//GetEventNotifyByTx return the events notify gen by executing of smart contract.  Wrap function of EventStore.GetEventNotifyByTx
func (this *LedgerStoreImp) GetEventNotifyByTx(tx common.Uint256) (*event.ExecuteNotify, error) {
	return this.eventStore.GetEventNotifyByTx(tx)
}

//GetEventNotifyByBlock return the transaction hash which have event notice after execution of smart contract. Wrap function of EventStore.GetEventNotifyByBlock
func (this *LedgerStoreImp) GetEventNotifyByBlock(height uint32) ([]*event.ExecuteNotify, error) {
	return this.eventStore.GetEventNotifyByBlock(height)
}

//PreExecuteContract return the result of smart contract execution without commit to store
func (this *LedgerStoreImp) PreExecuteContract(tx *types.Transaction) (*sstate.PreExecResult, error) {
	if tx.TxType != types.Invoke {
		return nil, errors.NewErr("transaction type error")
	}

	invoke, ok := tx.Payload.(*payload.InvokeCode)
	if !ok {
		return nil, errors.NewErr("transaction type error")
	}

	header, err := this.GetHeaderByHeight(this.GetCurrentBlockHeight())
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[PreExecuteContract] Get current block error!")
	}
	// init smart contract configuration info
	config := &smartcontract.Config{
		Time:   header.Timestamp,
		Height: header.Height,
		Tx:     tx,
	}

	//init smart contract info
	sc := smartcontract.SmartContract{
		Config:     config,
		Store:      this,
		CloneCache: storage.NewCloneCache(this.stateStore.NewStateBatch()),
		Gas:        math.MaxUint64,
	}

	//start the smart contract executive function
	engine, err := sc.NewExecuteEngine(invoke.Code)
	if err != nil {
		return nil, err
	}
	result, err := engine.Invoke()
	if err != nil {
		return nil, err
	}
	gasCost := math.MaxUint64 - sc.Gas
	if gasCost < neovm.TRANSACTION_GAS {
		gasCost = neovm.TRANSACTION_GAS
	}
	if err != nil {
		return &sstate.PreExecResult{State: event.CONTRACT_STATE_FAIL, Gas: gasCost, Result: nil}, err
	}
	return &sstate.PreExecResult{State: event.CONTRACT_STATE_SUCCESS, Gas: gasCost, Result: scommon.ConvertNeoVmTypeHexString(result)}, nil
}

//Close ledger store.
func (this *LedgerStoreImp) Close() error {
	err := this.blockStore.Close()
	if err != nil {
		return fmt.Errorf("blockStore close error %s", err)
	}
	err = this.stateStore.Close()
	if err != nil {
		return fmt.Errorf("stateStore close error %s", err)
	}
	err = this.eventStore.Close()
	if err != nil {
		return fmt.Errorf("eventStore close error %s", err)
	}
	return nil
}
