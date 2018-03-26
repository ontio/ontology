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

package ledgerstore

import (
	"fmt"
	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/payload"
	"github.com/Ontology/core/states"
	"github.com/Ontology/core/store/statestore"
	"github.com/Ontology/core/types"
	"github.com/Ontology/crypto"
	"github.com/Ontology/events"
	"github.com/Ontology/events/message"
	"sort"
	"sync"
	"time"
	"github.com/Ontology/smartcontract/event"
)

const (
	SystemVersion        = byte(1)
	HeaderIndexBatchSize = uint32(2000)
	BlockCacheTimeout    = time.Minute * 30
)

var (
	DBDirEvent          = "Chain/ledgerevent"
	DBDirBlock          = "Chain/block"
	DBDirState          = "Chain/states"
	MerkleTreeStorePath = "Chain/merkle_tree.db"
)

type ledgerCacheItem struct {
	item      interface{}
	cacheTime time.Time
}

type LedgerStore struct {
	blockStore       *BlockStore
	stateStore       *StateStore
	eventStore       *EventStore
	storedIndexCount uint32
	currBlockHeight  uint32
	currBlockHash    common.Uint256
	headerCache      map[common.Uint256]*ledgerCacheItem
	blockCache       map[common.Uint256]*ledgerCacheItem
	headerIndex      map[uint32]common.Uint256
	savingBlockHashes  map[common.Uint256]bool
	lock             sync.RWMutex
	exitCh           chan interface{}
}

func NewLedgerStore() (*LedgerStore, error) {
	ledgerStore := &LedgerStore{
		exitCh:      make(chan interface{}, 0),
		headerIndex: make(map[uint32]common.Uint256),
		headerCache: make(map[common.Uint256]*ledgerCacheItem),
		blockCache:  make(map[common.Uint256]*ledgerCacheItem),
		savingBlockHashes:make(map[common.Uint256]bool, 0),
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

func (this *LedgerStore) InitLedgerStoreWithGenesisBlock(genesisBlock *types.Block, defaultBookkeeper []*crypto.PubKey) error {
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
		sort.Sort(crypto.PubKeySlice(defaultBookkeeper))
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

func (this *LedgerStore) hasAlreadyInitGenesisBlock() (bool, error) {
	version, err := this.blockStore.GetVersion()
	if err != nil {
		return false, fmt.Errorf("GetVersion error %s", err)
	}
	return version == SystemVersion, nil
}

func (this *LedgerStore) initGenesisBlock() error {
	return this.blockStore.SaveVersion(SystemVersion)
}

func (this *LedgerStore) init() error {
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

func (this *LedgerStore) initCurrentBlock() error {
	currentBlockHash, currentBlockHeight, err := this.blockStore.GetCurrentBlock()
	if err != nil {
		return fmt.Errorf("LoadCurrentBlock error %s", err)
	}
	log.Infof("InitCurrentBlock currentBlockHash %x currentBlockHeight %d", currentBlockHash, currentBlockHeight)
	this.currBlockHash = currentBlockHash
	this.currBlockHeight = currentBlockHeight
	return nil
}

func (this *LedgerStore) initHeaderIndexList() error {
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

func (this *LedgerStore) initStore() error {
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
	}
	return nil
}

func (this *LedgerStore) start() {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	timeoutTicker := time.NewTicker(time.Minute)
	defer timeoutTicker.Stop()
	for {
		select {
		case <-this.exitCh:
			return
		case <-ticker.C:
			go this.clearCache()
		case <-timeoutTicker.C:
			go this.clearTimeoutBlock()
		}
	}
}

func (this *LedgerStore) clearCache() {
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

func (this *LedgerStore) clearTimeoutBlock() {
	this.lock.Lock()
	defer this.lock.Unlock()
	timeoutBlocks := make([]common.Uint256, 0)
	now := time.Now()
	for blockHash, cacheItem := range this.blockCache {
		if now.Sub(cacheItem.cacheTime) < BlockCacheTimeout {
			continue
		}
		timeoutBlocks = append(timeoutBlocks, blockHash)
	}
	for _, blockHash := range timeoutBlocks {
		delete(this.blockCache, blockHash)
	}
}

func (this *LedgerStore) setHeaderIndex(height uint32, blockHash common.Uint256) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.headerIndex[height] = blockHash
}

func (this *LedgerStore) getHeaderIndex(height uint32) common.Uint256 {
	this.lock.RLock()
	defer this.lock.RUnlock()
	blockHash, ok := this.headerIndex[height]
	if !ok {
		return common.Uint256{}
	}
	return blockHash
}

func (this *LedgerStore) GetCurrentHeaderHeight() uint32 {
	this.lock.RLock()
	defer this.lock.RUnlock()
	size := len(this.headerIndex)
	if size == 0 {
		return 0
	}
	return uint32(size) - 1
}

func (this *LedgerStore) GetCurrentHeaderHash() common.Uint256 {
	this.lock.RLock()
	defer this.lock.RUnlock()
	size := len(this.headerIndex)
	if size == 0 {
		return common.Uint256{}
	}
	return this.headerIndex[uint32(size)-1]
}

func (this *LedgerStore) setCurrentBlock(height uint32, blockHash common.Uint256) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.currBlockHash = blockHash
	this.currBlockHeight = height
	return
}

func (this *LedgerStore) GetCurrentBlock() (uint32, common.Uint256) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.currBlockHeight, this.currBlockHash
}

func (this *LedgerStore) GetCurrentBlockHash() common.Uint256 {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.currBlockHash
}

func (this *LedgerStore) GetCurrentBlockHeight() uint32 {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.currBlockHeight
}

func (this *LedgerStore) addToHeaderCache(header *types.Header) {
	this.lock.Lock()
	defer this.lock.Unlock()
	cacheItem := &ledgerCacheItem{
		item:      header,
		cacheTime: time.Now(),
	}
	this.headerCache[header.Hash()] = cacheItem
}

func (this *LedgerStore) getFromHeaderCache(blockHash common.Uint256) *types.Header {
	this.lock.RLock()
	defer this.lock.RUnlock()
	cacheItem, ok := this.headerCache[blockHash]
	if !ok {
		return nil
	}
	return cacheItem.item.(*types.Header)
}

func (this *LedgerStore) addToBlockCache(block *types.Block) {
	this.lock.Lock()
	defer this.lock.Unlock()
	cacheItem := &ledgerCacheItem{
		item:      block,
		cacheTime: time.Now(),
	}
	this.blockCache[block.Hash()] = cacheItem
}

func (this *LedgerStore) getFromBlockCache(blockHash common.Uint256) *types.Block {
	this.lock.RLock()
	defer this.lock.RUnlock()
	cacheItem, ok := this.blockCache[blockHash]
	if !ok {
		return nil
	}
	return cacheItem.item.(*types.Block)
}

func (this *LedgerStore) delFromBlockHash(blockHash common.Uint256) {
	this.lock.Lock()
	defer this.lock.Unlock()
	delete(this.blockCache, blockHash)
}

func (this *LedgerStore) verifyHeader(header *types.Header) error {
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

	if prevHeader.Height+1 != header.Height {
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

	m := len(header.Bookkeepers) - (len(header.Bookkeepers)-1)/3
	hash := header.Hash()
	err = crypto.VerifyMultiSignature(hash[:], header.Bookkeepers, m, header.SigData)
	if err != nil {
		return err
	}

	return nil
}

//sync block header
func (this *LedgerStore) AddHeader(header *types.Header) error {
	nextHeaderHeight := this.GetCurrentHeaderHeight() + 1
	if header.Height != nextHeaderHeight {
		return fmt.Errorf("header height %d not equal next header height %d", header.Height, nextHeaderHeight)
	}
	err := this.verifyHeader(header)
	if err != nil {
		fmt.Errorf("verifyHeader error %s", err)
	}
	blockHash := header.Hash()
	this.setHeaderIndex(header.Height, blockHash)
	this.addToHeaderCache(header)
	return nil
}

func (this *LedgerStore) AddHeaders(headers []*types.Header) error {
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

func (this *LedgerStore) verifyBlock(block *types.Block) error {
	if block.Header.Height == 0 {
		return nil
	}
	if len(block.Transactions) == 0 {
		return fmt.Errorf("transaction is emtpy")
	}
	txs := block.Transactions
	size := len(txs)
	for i := 0; i < size; i++ {
		tx := txs[i]
		if i == 0 && tx.TxType != types.BookKeeping {
			return fmt.Errorf("first transaction type is not BookKeeping")
		}
		if i > 0 && tx.TxType == types.BookKeeping {
			return fmt.Errorf("too many BookKeeping transaction in block")
		}
	}
	return nil
}

func (this *LedgerStore) addSavingBlock(blockHash common.Uint256) bool {
	this.lock.Lock()
	defer this.lock.Unlock()

	_, ok := this.savingBlockHashes[blockHash]
	if ok {
		return false
	}

	this.savingBlockHashes[blockHash] = true
	return true
}

func (this *LedgerStore) deleteSavingBlock(blockHash common.Uint256) {
	this.lock.Lock()
	defer this.lock.Unlock()

	delete(this.savingBlockHashes, blockHash)
}

func (this *LedgerStore) AddBlock(block *types.Block) error {
	if !this.addSavingBlock(block.Hash()) {
		return nil
	}

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

	return nil
}

func (this *LedgerStore) saveBlockToBlockStore(block *types.Block) error {
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
	err = this.blockStore.CommitTo()
	if err != nil {
		return fmt.Errorf("blockStore.CommitTo error %s", err)
	}
	return nil
}

func (this *LedgerStore) saveBlockToStateStore(block *types.Block) error {
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
	err = this.stateStore.CommitTo()
	if err != nil {
		return fmt.Errorf("stateStore.CommitTo error %s", err)
	}
	return nil
}

func (this *LedgerStore) saveBlockToEventStore(block *types.Block) error {
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
	err = this.eventStore.CommitTo()
	if err != nil {
		return fmt.Errorf("eventStore.CommitTo error %s", err)
	}
	return nil
}

func (this *LedgerStore) saveBlock(block *types.Block) error {
	blockHash := block.Hash()
	blockHeight := block.Header.Height
	defer this.deleteSavingBlock(blockHash)

	this.blockStore.NewBatch()
	this.stateStore.NewBatch()
	this.eventStore.NewBatch()
	err := this.saveBlockToBlockStore(block)
	if err != nil {
		return fmt.Errorf("save to block store error:%s", err)
	}
	err = this.saveBlockToStateStore(block)
	if err != nil {
		return fmt.Errorf("save to state store error:%s", err)
	}
	err = this.saveBlockToEventStore(block)
	if err != nil {
		return fmt.Errorf("save to event store error:%s", err)
	}

	this.setCurrentBlock(blockHeight, blockHash)

	if events.DefActorPublisher != nil {
		events.DefActorPublisher.Publish(
			message.TopicSaveBlockComplete,
			&message.SaveBlockCompleteMsg{
				Block: block,
			})
	}
	return nil
}

func (this *LedgerStore) handleTransaction(stateBatch *statestore.StateBatch, block *types.Block, tx *types.Transaction) error {
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

func (this *LedgerStore) saveHeaderIndexList() error {
	this.lock.RLock()
	storeCount := this.storedIndexCount
	currHeight := this.currBlockHeight
	if currHeight-storeCount < HeaderIndexBatchSize {
		this.lock.RUnlock()
		return nil
	}

	headerList := make([]common.Uint256, HeaderIndexBatchSize)
	for i := uint32(0); i < HeaderIndexBatchSize; i++ {
		height := storeCount + i
		headerList[i] = this.headerIndex[height]
	}
	this.lock.RUnlock()

	err := this.blockStore.SaveHeaderIndexList(storeCount, headerList)
	if err != nil {
		return fmt.Errorf("SaveHeaderIndexList start %d error %s", storeCount, err)
	}

	this.lock.Lock()
	this.storedIndexCount += HeaderIndexBatchSize
	this.lock.Unlock()
	return nil
}

func (this *LedgerStore) IsContainBlock(blockHash common.Uint256) (bool, error) {
	block := this.getFromBlockCache(blockHash)
	if block != nil {
		return true, nil
	}
	return this.blockStore.ContainBlock(blockHash)
}

func (this *LedgerStore) IsContainTransaction(txHash common.Uint256) (bool, error) {
	return this.blockStore.ContainTransaction(txHash)
}

func (this *LedgerStore) GetBlockRootWithNewTxRoot(txRoot common.Uint256) common.Uint256 {
	return this.stateStore.GetBlockRootWithNewTxRoot(txRoot)
}

func (this *LedgerStore) GetBlockHash(height uint32) common.Uint256 {
	return this.getHeaderIndex(height)
}

func (this *LedgerStore) GetHeaderByHash(blockHash common.Uint256) (*types.Header, error) {
	header := this.getFromHeaderCache(blockHash)
	if header != nil {
		return header, nil
	}
	return this.blockStore.GetHeader(blockHash)
}

func (this *LedgerStore) GetHeaderByHeight(height uint32) (*types.Header, error) {
	blockHash := this.GetBlockHash(height)
	var empty common.Uint256
	if blockHash == empty {
		return nil, nil
	}
	return this.GetHeaderByHash(blockHash)
}

func (this *LedgerStore) GetSysFeeAmount(blockHash common.Uint256) (common.Fixed64, error) {
	return this.blockStore.GetSysFeeAmount(blockHash)
}

func (this *LedgerStore) GetTransaction(txHash common.Uint256) (*types.Transaction, uint32, error) {
	return this.blockStore.GetTransaction(txHash)
}

func (this *LedgerStore) GetBlockByHash(blockHash common.Uint256) (*types.Block, error) {
	block := this.getFromBlockCache(blockHash)
	if block != nil {
		return block, nil
	}
	return this.blockStore.GetBlock(blockHash)
}

func (this *LedgerStore) GetBlockByHeight(height uint32) (*types.Block, error) {
	blockHash := this.GetBlockHash(height)
	var empty common.Uint256
	if blockHash == empty {
		return nil, nil
	}
	return this.GetBlockByHash(blockHash)
}

func (this *LedgerStore) GetBookkeeperState() (*states.BookkeeperState, error) {
	return this.stateStore.GetBookkeeperState()
}

func (this *LedgerStore) GetContractState(contractHash common.Address) (*payload.DeployCode, error) {
	return this.stateStore.GetContractState(contractHash)
}

func (this *LedgerStore) GetStorageItem(key *states.StorageKey) (*states.StorageItem, error) {
	return this.stateStore.GetStorageState(key)
}

func (this *LedgerStore)GetEventNotifyByTx(tx common.Uint256)([]*event.NotifyEventInfo, error){
	return this.eventStore.GetEventNotifyByTx(tx)
}

func (this *LedgerStore)GetEventNotifyByBlock(height uint32)([]common.Uint256, error){
	return this.eventStore.GetEventNotifyByBlock(height)
}

func (this *LedgerStore) PreExecuteContract(tx *types.Transaction) ([]interface{}, error) {
//	if tx.TxType != types.Invoke {
//		return nil, fmt.Errorf("transaction type error")
//	}
//	invokeCode, ok := tx.Payload.(*payload.InvokeCode)
//	if !ok {
//		return nil, fmt.Errorf("transaction type error")
//	}
//
//	param := invokeCode.Code
//	address := param.AddressFromVmCode()
//	code := append(param.Code, 0x67)
//	code = append(param.Code, address.ToArray()...)
//	stateBatch := this.stateStore.NewStateBatch()
//
//	stateMachine := service.NewStateMachine(this, stateBatch, smtypes.Application, 0)
//	se := neovm.NewExecutionEngine(tx, new(neovm.ECDsaCrypto), &CacheCodeTable{stateBatch}, stateMachine)
//	se.LoadCode(code, false)
//	err := se.Execute()
//	if err != nil {
//		return nil, err
//	}
//	if se.GetEvaluationStackCount() == 0 {
//		return nil, err
//	}
//	if neovm.Peek(se).GetStackItem() == nil {
//		return nil, err
//	}
//	return smcommon.ConvertReturnTypes(neovm.Peek(se).GetStackItem()), nil
	return nil, nil
}

func (this *LedgerStore) Close() error {
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
