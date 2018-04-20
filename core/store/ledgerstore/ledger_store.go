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
	"sort"
	"sync"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/store/statestore"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/events"
	"github.com/ontio/ontology/events/message"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract"
	"github.com/ontio/ontology/smartcontract/context"
	vmtype "github.com/ontio/ontology/smartcontract/types"
	scommon "github.com/ontio/ontology/smartcontract/common"
	"github.com/ontio/ontology/smartcontract/storage"
)

const (
	SYSTEM_VERSION = byte(1)          //Version of ledger store
	HEADER_INDEX_BATCH_SIZE = uint32(2000)     //Bath size of saving header index
	BLOCK_CACHE_TIMEOUT = time.Minute * 15 //Cache time for block to save in sync block
	MAX_HEADER_CACHE_SIZE = 5000             //Max cache size of block header in sync block
	MAX_BLOCK_CACHE_SIZE = 500              //Max cache size of block in sync block
)

var (
	//Storage save path.
	DBDirEvent = "Chain/ledgerevent"
	DBDirBlock = "Chain/block"
	DBDirState = "Chain/states"
	MerkleTreeStorePath = "Chain/merkle_tree.db"
)

type ledgerCacheItem struct {
	item      interface{}
	cacheTime time.Time
}

//LedgerStoreImp is main store struct fo ledger
type LedgerStoreImp struct {
	blockStore       *BlockStore                         //BlockStore for saving block & transaction data
	stateStore       *StateStore                         //StateStore for saving state data, like balance, smart contract execution result, and so on.
	eventStore       *EventStore                         //EventStore for saving log those gen after smart contract executed.
	storedIndexCount uint32                              //record the count of have saved block index
	currBlockHeight  uint32                              //Current block height
	currBlockHash    common.Uint256                      //Current block hash
	headerCache      map[common.Uint256]*ledgerCacheItem //Cache header to saving in sync block. BlockHash => Header
	blockCache       map[common.Uint256]*ledgerCacheItem //Cache block to saving in sync block. BlockHash => Block
	headerIndex      map[uint32]common.Uint256           //Header index, Mapping header height => block hash
	savingBlock      bool                                //is saving block now
	lock             sync.RWMutex
	exitCh           chan interface{}
}

//NewLedgerStore return LedgerStoreImp instance
func NewLedgerStore() (*LedgerStoreImp, error) {
	ledgerStore := &LedgerStoreImp{
		exitCh:      make(chan interface{}, 0),
		headerIndex: make(map[uint32]common.Uint256),
		headerCache: make(map[common.Uint256]*ledgerCacheItem),
		blockCache:  make(map[common.Uint256]*ledgerCacheItem),
	}

	blockStore, err := NewBlockStore(DBDirBlock, true)
	if err != nil {
		return nil, fmt.Errorf("NewBlockStore error %s", err)
	}
	ledgerStore.blockStore = blockStore

	stateStore, err := NewStateStore(DBDirState, MerkleTreeStorePath)
	if err != nil {
		return nil, fmt.Errorf("NewStateStore error %s", err)
	}
	ledgerStore.stateStore = stateStore

	eventState, err := NewEventStore(DBDirEvent)
	if err != nil {
		return nil, fmt.Errorf("NewEventStore error %s", err)
	}
	ledgerStore.eventStore = eventState

	err = ledgerStore.init()
	if err != nil {
		return nil, fmt.Errorf("init error %s", err)
	}

	go ledgerStore.start()
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
	} else {
		genesisHash := genesisBlock.Hash()
		exist, err := this.blockStore.ContainBlock(genesisHash)
		if err != nil {
			return fmt.Errorf("HashBlockExist error %s", err)
		}
		if !exist {
			return fmt.Errorf("GenesisBlock arenot init correctly")
		}
	}
	return nil
}

func (this *LedgerStoreImp) hasAlreadyInitGenesisBlock() (bool, error) {
	version, err := this.blockStore.GetVersion()
	if err != nil {
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
	currBlockHeight, currBlockHash := this.GetCurrentBlock()
	var empty common.Uint256
	if currBlockHash == empty {
		return nil
	}
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
		if blockHash == empty {
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

func (this *LedgerStoreImp) start() {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	timeoutTicker := time.NewTicker(time.Minute)
	defer timeoutTicker.Stop()
	for {
		select {
		case <-this.exitCh:
			return
		case <-ticker.C:
			go this.clearBlockCache()
		case <-timeoutTicker.C:
			go this.clearTimeoutBlock()
		}
	}
}

func (this *LedgerStoreImp) clearBlockCache() {
	this.lock.Lock()
	blocks := make([]*types.Block, 0)
	currentBlockHeight := this.currBlockHeight
	for blockHash, cacheItem := range this.headerCache {
		header := cacheItem.item.(*types.Header)
		if header.Height > currentBlockHeight {
			continue
		}
		delete(this.headerCache, blockHash)
	}
	for blockHash, cacheItem := range this.blockCache {
		block := cacheItem.item.(*types.Block)
		if block.Header.Height > currentBlockHeight {
			continue
		}
		delete(this.blockCache, blockHash)
	}
	for nextBlockHeight := currentBlockHeight + 1; ; nextBlockHeight++ {
		nextBlockHash, ok := this.headerIndex[nextBlockHeight]
		if !ok {
			break
		}
		cacheItem := this.blockCache[nextBlockHash]
		if cacheItem == nil {
			break
		}
		block := cacheItem.item.(*types.Block)
		blocks = append(blocks, block)
	}
	this.lock.Unlock()

	for _, block := range blocks {
		err := this.saveBlock(block)
		if err != nil {
			blockHash := block.Hash()
			this.delFromBlockHash(blockHash)
			log.Errorf("saveBlock in cache height:%d error %s", block.Header.Height, err)
			break
		}
	}
}

func (this *LedgerStoreImp) clearTimeoutBlock() {
	this.lock.Lock()
	defer this.lock.Unlock()
	timeoutBlocks := make([]common.Uint256, 0)
	now := time.Now()
	for blockHash, cacheItem := range this.blockCache {
		if now.Sub(cacheItem.cacheTime) < BLOCK_CACHE_TIMEOUT {
			continue
		}
		timeoutBlocks = append(timeoutBlocks, blockHash)
	}
	for _, blockHash := range timeoutBlocks {
		delete(this.blockCache, blockHash)
	}
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
	return this.headerIndex[uint32(size) - 1]
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

func (this *LedgerStoreImp) addToHeaderCache(header *types.Header) bool {
	this.lock.Lock()
	defer this.lock.Unlock()

	if len(this.headerCache) > MAX_HEADER_CACHE_SIZE {
		return false
	}

	cacheItem := &ledgerCacheItem{
		item:      header,
		cacheTime: time.Now(),
	}
	this.headerCache[header.Hash()] = cacheItem
	return true
}

func (this *LedgerStoreImp) getFromHeaderCache(blockHash common.Uint256) *types.Header {
	this.lock.RLock()
	defer this.lock.RUnlock()
	cacheItem, ok := this.headerCache[blockHash]
	if !ok {
		return nil
	}
	return cacheItem.item.(*types.Header)
}

func (this *LedgerStoreImp) addToBlockCache(block *types.Block) bool {
	this.lock.Lock()
	defer this.lock.Unlock()

	if len(this.blockCache) > MAX_BLOCK_CACHE_SIZE {
		return false
	}

	cacheItem := &ledgerCacheItem{
		item:      block,
		cacheTime: time.Now(),
	}
	this.blockCache[block.Hash()] = cacheItem
	return true
}

func (this *LedgerStoreImp) getFromBlockCache(blockHash common.Uint256) *types.Block {
	this.lock.RLock()
	defer this.lock.RUnlock()
	cacheItem, ok := this.blockCache[blockHash]
	if !ok {
		return nil
	}
	return cacheItem.item.(*types.Block)
}

func (this *LedgerStoreImp) delFromBlockHash(blockHash common.Uint256) {
	this.lock.Lock()
	defer this.lock.Unlock()
	delete(this.blockCache, blockHash)
}

func (this *LedgerStoreImp) verifyHeader(header *types.Header) error {
	if header.Height == 0 {
		return nil
	}

	var prevHeader *types.Header
	prevHeaderHash := header.PrevBlockHash
	prevHeader, err := this.GetHeaderByHash(prevHeaderHash)
	if err != nil {
		return fmt.Errorf("get prev header error %s", err)
	}
	if prevHeader == nil {
		return fmt.Errorf("cannot find pre header by blockHash %x", prevHeaderHash)
	}

	if prevHeader.Height + 1 != header.Height {
		return fmt.Errorf("block height is incorrect")
	}

	if prevHeader.Timestamp >= header.Timestamp {
		return fmt.Errorf("block timestamp is incorrect")
	}

	address, err := types.AddressFromBookkeepers(header.Bookkeepers)
	if err != nil {
		return err
	}
	if prevHeader.NextBookkeeper != address {
		return fmt.Errorf("bookkeeper address error")
	}

	m := len(header.Bookkeepers) - (len(header.Bookkeepers) - 1) / 3
	hash := header.Hash()
	err = signature.VerifyMultiSignature(hash[:], header.Bookkeepers, m, header.SigData)
	if err != nil {
		return err
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
		fmt.Errorf("verifyHeader error %s", err)
	}
	blockHash := header.Hash()
	if this.addToHeaderCache(header) {
		this.setHeaderIndex(header.Height, blockHash)
	}
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

func (this *LedgerStoreImp) verifyBlock(block *types.Block) error {
	if block.Header.Height == 0 {
		return nil
	}
	if len(block.Transactions) == 0 {
		return fmt.Errorf("transaction is empty")
	}
	txs := block.Transactions
	size := len(txs)
	for i := 0; i < size; i++ {
		tx := txs[i]
		if i == 0 && tx.TxType != types.BookKeeping {
			return fmt.Errorf("first transaction type is not Bookkeeping")
		}
		if i > 0 && tx.TxType == types.BookKeeping {
			return fmt.Errorf("too many Bookkeeping transaction in block")
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
	blockHash := this.getHeaderIndex(blockHeight)
	var empty common.Uint256
	if blockHeight > nextBlockHeight && blockHash == empty {
		return fmt.Errorf("block height %d larger than next block height %d", blockHeight, nextBlockHeight)
	}

	err := this.verifyHeader(block.Header)
	if err != nil {
		return fmt.Errorf("verifyHeader error %s", err)
	}
	err = this.verifyBlock(block)
	if err != nil {
		return fmt.Errorf("verifyBlock error %s", err)
	}

	if blockHeight != nextBlockHeight {
		//sync block
		this.addToBlockCache(block)
		return nil
	}

	err = this.saveBlock(block)
	if err != nil {
		return fmt.Errorf("saveBlock error %s", err)
	}

	go this.clearBlockCache()
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
	if this.isSavingBlock(){
		//hash already saved or is saving
		return nil
	}
	defer this.resetSavingBlock()
	if (blockHeight > 0 && blockHeight != (this.GetCurrentBlockHeight()+1)) {
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
	var err error
	txHash := tx.Hash()
	switch tx.TxType {
	case types.Deploy:
		err = this.stateStore.HandleDeployTransaction(stateBatch, tx)
		if err != nil {
			return fmt.Errorf("HandleDeployTransaction tx %x error %s", txHash, err)
		}
	case types.Invoke:
		err = this.stateStore.HandleInvokeTransaction(this, stateBatch, tx, block, this.eventStore)
		if err != nil {
			fmt.Printf("HandleInvokeTransaction tx %x error %s \n", txHash, err)
		}
	case types.Claim:
	case types.Enrollment:
	case types.Vote:
	}
	return nil
}

func (this *LedgerStoreImp) saveHeaderIndexList() error {
	this.lock.RLock()
	storeCount := this.storedIndexCount
	currHeight := this.currBlockHeight
	if currHeight - storeCount < HEADER_INDEX_BATCH_SIZE {
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
	block := this.getFromBlockCache(blockHash)
	if block != nil {
		return true, nil
	}
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
	header := this.getFromHeaderCache(blockHash)
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
	block := this.getFromBlockCache(blockHash)
	if block != nil {
		return block, nil
	}
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
func (this *LedgerStoreImp) GetEventNotifyByTx(tx common.Uint256) ([]*event.NotifyEventInfo, error) {
	return this.eventStore.GetEventNotifyByTx(tx)
}

//GetEventNotifyByBlock return the transaction hash which have event notice after execution of smart contract. Wrap function of EventStore.GetEventNotifyByBlock
func (this *LedgerStoreImp) GetEventNotifyByBlock(height uint32) ([]common.Uint256, error) {
	return this.eventStore.GetEventNotifyByBlock(height)
}

//PreExecuteContract return the result of smart contract execution without commit to store
func (this *LedgerStoreImp) PreExecuteContract(tx *types.Transaction) (interface{}, error) {
	if tx.TxType != types.Invoke {
		return nil, errors.NewErr("transaction type error")
	}

	invoke, ok := tx.Payload.(*payload.InvokeCode)
	if !ok {
		return nil, errors.NewErr("transaction type error")
	}

	header, err := this.GetHeaderByHeight(this.GetCurrentBlockHeight()); if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[PreExecuteContract] Get current block error!")
	}
	// init smart contract configuration info
	config := &smartcontract.Config{
		Time:    header.Timestamp,
		Height:  header.Height,
		Tx:      tx,
	}

	//init smart contract context info
	ctx := &context.Context{
		Code:            invoke.Code,
		ContractAddress: invoke.Code.AddressFromVmCode(),
	}

	//init smart contract info
	sc := smartcontract.SmartContract{
		Config: config,
		Store: this,
		CloneCache: storage.NewCloneCache(this.stateStore.NewStateBatch()),
	}

	//load current context to smart contract
	sc.PushContext(ctx)

	//start the smart contract executive function
	result, err := sc.Execute(); if err != nil {
		return nil, err
	}

	prefix := ctx.ContractAddress[0]
	if prefix == byte(vmtype.NEOVM) {
		result = scommon.ConvertNeoVmTypeHexString(result)
	} else if prefix == byte(vmtype.WASMVM) {
		if v, ok := result.([]byte); ok {
			result = common.ToHexString(v)
		}
	}
	return result, nil

}

//Close ledger store.
func (this *LedgerStoreImp) Close() error {
	close(this.exitCh)
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
