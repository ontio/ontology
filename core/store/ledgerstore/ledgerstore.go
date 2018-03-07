package ledgerstore

import (
	"fmt"
	"github.com/Ontology/common"
	"github.com/Ontology/core/payload"
	"github.com/Ontology/core/states"
	"github.com/Ontology/core/store/statestore"
	"github.com/Ontology/core/types"
	"github.com/Ontology/crypto"
	smcommon "github.com/Ontology/smartcontract/common"
	"github.com/Ontology/smartcontract/service"
	smtypes "github.com/Ontology/smartcontract/types"
	vmtypes "github.com/Ontology/smartcontract/types"
	"github.com/Ontology/vm/neovm"
	vm "github.com/Ontology/vm/neovm"
	log4 "github.com/alecthomas/log4go"
	"sort"
	"sync"
	"time"
)

const (
	SystemVersion        = byte(1)
	HeaderIndexBatchSize = uint32(2000)

	DBDirEvent          = "Chain/ledgerevent"
	DBDirBlock          = "Chain/block"
	DBDirState          = "Chain/states"
	DBDirMerkleTree     = "Chain/merkle"
	MerkleTreeStorePath = "Chain/merkle_tree.db"
)

type LedgerStore struct {
	blockStore       *BlockStore
	stateStore       *StateStore
	merkleStore      *MerkleTreeStore
	eventStore       *EventStore
	storedIndexCount uint32
	currBlockHeight  uint32
	currBlockHash    *common.Uint256
	headerCache      map[common.Uint256]*types.Header
	blockCache       map[common.Uint256]*types.Block
	headerIndex      map[uint32]*common.Uint256
	lock             sync.RWMutex
	exitCh           chan interface{}
}

func NewLedgerStore() (*LedgerStore, error) {
	ledgerStore := &LedgerStore{exitCh: make(chan interface{}, 0)}
	ledgerStore.headerCache = make(map[common.Uint256]*types.Header)

	blockStore, err := NewBlockStore(DBDirBlock, true)
	if err != nil {
		return nil, fmt.Errorf("NewBlockStore error %s", err)
	}
	ledgerStore.blockStore = blockStore

	stateStore, err := NewStateStore(DBDirState)
	if err != nil {
		return nil, fmt.Errorf("NewStateStore error %s", err)
	}
	ledgerStore.stateStore = stateStore

	err = ledgerStore.init()
	if err != nil {
		return nil, fmt.Errorf("init error %s", err)
	}

	merkleStore, err := NewMerkleTreeStore(DBDirMerkleTree, MerkleTreeStorePath, ledgerStore.GetCurrentBlockHeight())
	if err != nil {
		return nil, fmt.Errorf("NewMerkleTreeStore error %s", err)
	}
	ledgerStore.merkleStore = merkleStore

	eventState, err := NewEventStore(DBDirEvent)
	if err != nil {
		return nil, fmt.Errorf("NewEventStore error %s", err)
	}
	ledgerStore.eventStore = eventState

	go ledgerStore.start()
	return ledgerStore, nil
}

func (this *LedgerStore) InitLedgerStoreWithGenesisBlock(genesisBlock *types.Block, defaultBookKeeper []*crypto.PubKey) error {
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
		err = this.merkleStore.ClearAll()
		if err != nil {
			return fmt.Errorf("merkleStore.ClearAll error %s", err)
		}
		err = this.eventStore.ClearAll()
		if err != nil {
			return fmt.Errorf("eventStore.ClearAll error %s", err)
		}
		sort.Sort(crypto.PubKeySlice(defaultBookKeeper))
		bookKeeperState := &states.BookKeeperState{
			CurrBookKeeper: defaultBookKeeper,
			NextBookKeeper: defaultBookKeeper,
		}
		err = this.stateStore.SaveBookKeeperState(bookKeeperState)
		if err != nil {
			return fmt.Errorf("SaveBookKeeperState error %s", err)
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
		exist, err := this.blockStore.ContainBlock(&genesisHash)
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
	return nil
}

func (this *LedgerStore) initCurrentBlock() error {
	currentBlockHash, currentBlockHeight, err := this.blockStore.GetCurrentBlock()
	if err != nil {
		return fmt.Errorf("LoadCurrentBlock error %s", err)
	}
	this.currBlockHash = currentBlockHash
	this.currBlockHeight = currentBlockHeight
	return nil
}

func (this *LedgerStore) initHeaderIndexList() error {
	currBlockHeight, currBlockHash := this.GetCurrentBlock()
	if currBlockHash == nil {
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
		height := storeIndexCount
		blockHash, err := this.blockStore.GetBlockHash(height)
		if err != nil {
			return fmt.Errorf("LoadBlockHash height %d error %s", height, err)
		}
		if blockHash == nil {
			return fmt.Errorf("LoadBlockHash height %d hash nil", height)
		}
		this.headerIndex[height] = blockHash
	}
	this.storedIndexCount += currBlockHeight - storeIndexCount + 1
	return nil
}

func (this *LedgerStore) start() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-this.exitCh:
			return
		case <-ticker.C:
			this.clearCache()
		}
	}
}

func (this *LedgerStore) clearCache() {
	this.lock.Lock()
	blocks := make([]*types.Block, 0)
	currentBlockHeight := this.currBlockHeight
	for blockHash, header := range this.headerCache {
		if header.Height > currentBlockHeight {
			continue
		}
		delete(this.headerCache, blockHash)
	}
	for blockHash, block := range this.blockCache {
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
		block := this.blockCache[*nextBlockHash]
		if block == nil {
			break
		}
		blocks = append(blocks, block)
	}
	this.lock.Unlock()

	for _, block := range blocks {
		err := this.saveBlock(block)
		if err != nil {
			blockHash := block.Hash()
			this.delFromBlockHash(&blockHash)
			log4.Error("saveBlock in cache height:%d error %s", block.Header.Height, err)
			break
		}
	}
}

func (this *LedgerStore) getStoreIndexCount() uint32 {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.storedIndexCount
}

func (this *LedgerStore) addStoreIndexCount(delt uint32) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.storedIndexCount += delt
}

func (this *LedgerStore) setHeaderIndex(height uint32, blockHash *common.Uint256) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.headerIndex[height] = blockHash
}

func (this *LedgerStore) getHeaderIndex(height uint32) *common.Uint256 {
	this.lock.RLock()
	defer this.lock.RUnlock()
	blockHash, ok := this.headerIndex[height]
	if !ok {
		return nil
	}
	return blockHash
}

func (this *LedgerStore) GetCurrentHeaderHeight() uint32 {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return uint32(len(this.headerIndex)) - 1
}

func (this *LedgerStore) GetCurrentHeaderHash() *common.Uint256 {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.headerIndex[uint32(len(this.headerIndex))-1]
}

func (this *LedgerStore) setCurrentBlock(height uint32, blockHash *common.Uint256) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.currBlockHash = blockHash
	this.currBlockHeight = height
	return
}

func (this *LedgerStore) GetCurrentBlock() (uint32, *common.Uint256) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.currBlockHeight, this.currBlockHash
}

func (this *LedgerStore) GetCurrentBlockHash() *common.Uint256 {
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
	this.headerCache[header.Hash()] = header
}

func (this *LedgerStore) getFromHeaderCache(blockHash *common.Uint256) *types.Header {
	this.lock.RLock()
	defer this.lock.RUnlock()
	header, ok := this.headerCache[*blockHash]
	if !ok {
		return nil
	}
	return header
}

func (this *LedgerStore) addToBlockCache(block *types.Block) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.blockCache[block.Hash()] = block
}

func (this *LedgerStore) getFromBlockCache(blockHash *common.Uint256) *types.Block {
	this.lock.RLock()
	defer this.lock.RUnlock()
	block, ok := this.blockCache[*blockHash]
	if !ok {
		return nil
	}
	return block
}

func (this *LedgerStore) delFromBlockHash(blockHash *common.Uint256) {
	this.lock.Lock()
	defer this.lock.Unlock()
	delete(this.blockCache, *blockHash)
}

func (this *LedgerStore) verifyHeader(header *types.Header) error {
	if header.Height == 0 {
		return nil
	}

	var prevHeader *types.Header
	prevHeaderHash := header.PrevBlockHash
	prevHeader, err := this.GetHeaderByHash(&prevHeaderHash)
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

	programhash := common.ToCodeHash(header.Program.Code)
	if prevHeader.NextBookKeeper != types.Address(programhash) {
		return fmt.Errorf("bookkeeper address error")
	}

	err = this.verifyHeaderProgram(header)
	if err != nil {
		return fmt.Errorf("VerifyHeaderProgram error %s", err)
	}
	return nil
}

func (this *LedgerStore) verifyHeaderProgram(header *types.Header) error {
	program := header.Program
	cryptos := new(vm.ECDsaCrypto)
	stateReader := service.NewStateReader(this, vmtypes.Verification)
	se := vm.NewExecutionEngine(header, cryptos, nil, stateReader)
	se.LoadCode(program.Code, false)
	se.LoadCode(program.Parameter, true)
	se.Execute()

	if se.GetState() != vm.HALT {
		return fmt.Errorf("VM] Finish State not equal to HALT")
	}
	if se.GetEvaluationStack().Count() != 1 {
		return fmt.Errorf("[VM] Execute Engine Stack Count Error")
	}
	flag := se.GetExecuteResult()
	if !flag {
		return fmt.Errorf("[VM] Check Sig FALSE")
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
	this.setHeaderIndex(header.Height, &blockHash)
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

func (this *LedgerStore) AddBlock(block *types.Block) error {
	currBlockHeight := this.GetCurrentBlockHeight()
	blockHeight := block.Header.Height
	if blockHeight <= currBlockHeight {
		return fmt.Errorf("block height %d not larger then current block height %d", blockHeight, currBlockHeight)
	}

	nextBlockHeight := currBlockHeight + 1
	if blockHeight > nextBlockHeight {
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
		this.addToBlockCache(block)
		return nil
	}

	err = this.saveBlock(block)
	if err != nil {
		return fmt.Errorf("saveBlock error %s", err)
	}

	return nil
}

func (this *LedgerStore) saveBlock(block *types.Block) error {
	blockHash := block.Hash()
	blockHeight := block.Header.Height

	err := this.blockStore.NewBatch()
	if err != nil {
		return fmt.Errorf("blockStore.NewBatch error %s", err)
	}
	stateRoot := block.Header.StateRoot
	err = this.stateStore.NewBatch()
	if err != nil {
		return fmt.Errorf("stateStore.NewBatch error %s", err)
	}
	stateBatch, err := this.stateStore.NewStateBatch(&stateRoot)
	if err != nil {
		return fmt.Errorf("NewStateBatch error %s", err)
	}
	err = this.merkleStore.NewBatch()
	if err != nil {
		return fmt.Errorf("merkleStore.NewBatch error %s", err)
	}

	err = this.stateStore.HandleBookKeeper(stateBatch)
	if err != nil {
		return fmt.Errorf("handleBookKeeper error %s", err)
	}

	invokeTxs := make([]*common.Uint256, 0)
	for _, tx := range block.Transactions {
		err = this.handleTransaction(stateBatch, block, tx)
		if err != nil {
			return fmt.Errorf("handleTransaction error %s", err)
		}
		txHash := tx.Hash()
		if tx.TxType == types.Invoke {
			invokeTxs = append(invokeTxs, &txHash)
		}
	}

	err = this.saveHeaderIndexList()
	if err != nil {
		return fmt.Errorf("saveHeaderIndexList error %s", err)
	}
	err = this.blockStore.SaveCurrentBlock(blockHeight, &blockHash)
	if err != nil {
		return fmt.Errorf("SaveCurrentBlock error %s", err)
	}
	err = this.blockStore.SaveBlockHash(blockHeight, &blockHash)
	if err != nil {
		return fmt.Errorf("SaveBlockHash height %s hash %v error%s", blockHeight, blockHash, err)
	}
	err = this.blockStore.SaveBlock(block)
	if err != nil {
		return fmt.Errorf("SaveBlock height %d hash %x error %s", blockHeight, blockHash, err)
	}

	err = this.stateStore.SaveCurrentStateRoot(&block.Header.StateRoot)
	if err != nil {
		return fmt.Errorf("SaveCurrentStateRoot error %s", err)
	}
	err = this.merkleStore.AddMerkleTreeRoot(block.Header.TransactionsRoot)
	if err != nil {
		return fmt.Errorf("AddMerkleTreeRoot error %s", err)
	}

	if len(invokeTxs) > 0 {
		err = this.eventStore.SaveEventNotifyByBlock(blockHeight, invokeTxs)
		if err != nil {
			return fmt.Errorf("SaveEventNotifyByBlock error %s", err)
		}
	}

	newStateRoot, err := stateBatch.CommitTo()
	if err != nil {
		return fmt.Errorf("stateBatch.CommitTo error %s", err)
	}
	if *newStateRoot != stateRoot {
		err = this.stateStore.SaveCurrentStateRoot(newStateRoot)
		if err != nil {
			return fmt.Errorf("SaveCurrentStateRoot error %s", err)
		}
	}
	err = this.stateStore.CommitTo()
	if err != nil {
		return fmt.Errorf("stateStore.CommitTo error %s", err)
	}
	err = this.blockStore.CommitTo()
	if err != nil {
		return fmt.Errorf("blockStore.CommitTo error %s", err)
	}
	err = this.merkleStore.CommitTo()
	if err != nil {
		return fmt.Errorf("merkleStore.CommitTo error %s", err)
	}
	err = this.eventStore.CommitTo()
	if err != nil {
		return fmt.Errorf("eventStore.CommitTo error %s", err)
	}
	this.setCurrentBlock(blockHeight, &blockHash)

	return nil
}

func (this *LedgerStore) handleTransaction(stateBatch *statestore.StateBatch, block *types.Block, tx *types.Transaction) error {
	var err error
	//blockHeight := block.Header.Height
	txHash := tx.Hash()
	//err := this.stateStore.HandleTxOutput(tx)
	//if err != nil {
	//	return fmt.Errorf("handleTxOutput block height %d tx %x err %s", blockHeight, txHash, err)
	//}
	//err = this.stateStore.HandleTxInput(tx, blockHeight, this.blockStore)
	//if err != nil {
	//	return fmt.Errorf("handleTxInput block height %d tx %x err %s", blockHeight, txHash, err)
	//}
	switch tx.TxType {
	case types.BookKeeper:
		err = this.stateStore.HandleBookKeeperTransaction(stateBatch, tx)
		if err != nil {
			return fmt.Errorf("HandleBookKeeperTransaction tx %x error %s", txHash, err)
		}
	//case types.RegisterAsset:
	//	err = this.stateStore.HandleRegisterAssertTransaction(tx, blockHeight)
	//	if err != nil {
	//		return fmt.Errorf("HandleRegisterAssertTransaction tx %x error %s", txHash, err)
	//	}
	//case types.IssueAsset:
	//	err = this.stateStore.HandleIssueAssetTransaction(tx)
	//	if err != nil {
	//		return fmt.Errorf("HandleIssueAssetTransaction tx %x error %s", txHash, err)
	//	}
	case types.Deploy:
		err = this.stateStore.HandleDeployTransaction(stateBatch, tx)
		if err != nil {
			return fmt.Errorf("HandleDeployTransaction tx %x error %s", txHash, err)
		}
	case types.Invoke:
		err = this.stateStore.HandleInvokeTransaction(stateBatch, tx, block, this.eventStore)
		if err != nil {
			return fmt.Errorf("HandleInvokeTransaction tx %x error %s", txHash, err)
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

	headerList := make(map[uint32]*common.Uint256, HeaderIndexBatchSize)
	for i := uint32(0); i < HeaderIndexBatchSize; i++ {
		height := storeCount + i
		blockHash := this.headerIndex[height]
		headerList[height] = blockHash
	}
	this.lock.RUnlock()

	err := this.blockStore.SaveHeaderIndexList(storeCount, headerList)
	if err != nil {
		return fmt.Errorf("SaveHeaderIndexList start %d error %s", storeCount, err)
	}

	this.addStoreIndexCount(HeaderIndexBatchSize)
	return nil
}

func (this *LedgerStore) IsContainBlock(blockHash *common.Uint256) (bool, error) {
	block := this.getFromBlockCache(blockHash)
	if block != nil {
		return true, nil
	}
	return this.blockStore.ContainBlock(blockHash)
}

func (this *LedgerStore) GetCurrentStateRoot() (*common.Uint256, error) {
	return this.stateStore.GetCurrentStateRoot()
}

func (this *LedgerStore) IsContainTransaction(txHash *common.Uint256) (bool, error) {
	return this.blockStore.ContainTransaction(txHash)
}

func (this *LedgerStore) GetBlockRootWithNewTxRoot(txRoot *common.Uint256) *common.Uint256 {
	newTxRoot := this.merkleStore.GetBlockRootWithNewTxRoot(*txRoot)
	return &newTxRoot
}

func (this *LedgerStore) GetBlockHash(height uint32) *common.Uint256 {
	return this.getHeaderIndex(height)
}

func (this *LedgerStore) GetHeaderByHash(blockHash *common.Uint256) (*types.Header, error) {
	header := this.getFromHeaderCache(blockHash)
	if header != nil {
		return header, nil
	}
	return this.blockStore.GetHeader(blockHash)
}

func (this *LedgerStore) GetHeaderByHeight(height uint32) (*types.Header, error) {
	blockHash := this.GetBlockHash(height)
	if blockHash == nil {
		return nil, nil
	}
	return this.GetHeaderByHash(blockHash)
}

func (this *LedgerStore) GetSysFeeAmount(blockHash *common.Uint256) (common.Fixed64, error) {
	return this.blockStore.GetSysFeeAmount(blockHash)
}

func (this *LedgerStore) GetTransaction(txHash *common.Uint256) (*types.Transaction, uint32, error) {
	return this.blockStore.GetTransaction(txHash)
}

func (this *LedgerStore) GetBlockByHash(blockHash *common.Uint256) (*types.Block, error) {
	block := this.getFromBlockCache(blockHash)
	if block != nil {
		return block, nil
	}
	return this.blockStore.GetBlock(blockHash)
}

func (this *LedgerStore) GetBlockByHeight(height uint32) (*types.Block, error) {
	blockHash := this.GetBlockHash(height)
	if blockHash == nil {
		return nil, nil
	}
	return this.GetBlockByHash(blockHash)
}

func (this *LedgerStore) GetUnspentCoinState(refTxId *common.Uint256) (*states.UnspentCoinState, error) {
	return this.stateStore.GetUnspentCoinState(refTxId)
}

func (this *LedgerStore) GetSpentCoinState(refTxId *common.Uint256) (*states.SpentCoinState, error) {
	return this.stateStore.GetSpentCoinState(refTxId)
}

func (this *LedgerStore) GetAccountState(programHash *common.Uint160) (*states.AccountState, error) {
	return this.stateStore.GetAccountState(programHash)
}

func (this *LedgerStore) GetBookKeeperState() (*states.BookKeeperState, error) {
	return this.stateStore.GetBookKeeperState()
}

func (this *LedgerStore) GetAssetState(assetId *common.Uint256) (*states.AssetState, error) {
	return this.stateStore.GetAssetState(assetId)
}

func (this *LedgerStore) GetContractState(contractHash *common.Uint160) (*states.ContractState, error) {
	return this.stateStore.GetContractState(contractHash)
}

func (this *LedgerStore) GetUnspentCoinStateByProgramHash(programHash *common.Uint160, assetId *common.Uint256) (*states.ProgramUnspentCoin, error) {
	return this.stateStore.GetUnspentCoinStateByProgramHash(programHash, assetId)
}

func (this *LedgerStore) GetStorageItem(key *states.StorageKey) (*states.StorageItem, error) {
	return this.stateStore.GetStorageState(key)
}

func (this *LedgerStore) GetAllAssetState() (map[common.Uint256]*states.AssetState, error) {
	return this.stateStore.GetAllAssetState()
}

func (this *LedgerStore) PreExecuteContract(tx *types.Transaction) ([]interface{}, error) {
	if tx.TxType != types.Invoke {
		return nil, fmt.Errorf("transaction type error")
	}
	invokeCode, ok := tx.Payload.(*payload.InvokeCode)
	if !ok {
		return nil, fmt.Errorf("transaction type error")
	}

	param := invokeCode.Code.Code
	codeHash := common.ToCodeHash(param)
	param = append(param, 0x67)
	param = append(param, codeHash.ToArray()...)
	stateBatch, err := this.stateStore.NewStateBatch(&common.Uint256{})
	if err != nil {
		return nil, fmt.Errorf("NewStateBatch error %s", err)
	}

	stateMachine := service.NewStateMachine(this, stateBatch, smtypes.Application, nil)
	se := neovm.NewExecutionEngine(tx, new(neovm.ECDsaCrypto), &CacheCodeTable{stateBatch}, stateMachine)
	se.LoadCode(param, false)
	err = se.Execute()
	if err != nil {
		return nil, err
	}
	if se.GetEvaluationStackCount() == 0 {
		return nil, err
	}
	if neovm.Peek(se).GetStackItem() == nil {
		return nil, err
	}
	return smcommon.ConvertReturnTypes(neovm.Peek(se).GetStackItem()), nil
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
	err = this.merkleStore.Close()
	if err != nil {
		return fmt.Errorf("merkleStore close error %s", err)
	}
	err = this.eventStore.Close()
	if err != nil {
		return fmt.Errorf("eventStore close error %s", err)
	}
	return nil
}
