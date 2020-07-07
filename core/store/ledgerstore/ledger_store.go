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
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hash"
	"math"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/store"
	scom "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/events"
	"github.com/ontio/ontology/events/message"
	"github.com/ontio/ontology/merkle"
	"github.com/ontio/ontology/smartcontract"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/service/neovm"
	"github.com/ontio/ontology/smartcontract/service/wasmvm"
	sstate "github.com/ontio/ontology/smartcontract/states"
	"github.com/ontio/ontology/smartcontract/storage"
	types2 "github.com/ontio/ontology/vm/neovm/types"
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

type PrexecuteParam struct {
	JitMode    bool
	WasmFactor uint64
	MinGas     bool
}

//LedgerStoreImp is main store struct fo ledger
type LedgerStoreImp struct {
	blockStore           *BlockStore                      //BlockStore for saving block & transaction data
	stateStore           *StateStore                      //StateStore for saving state data, like balance, smart contract execution result, and so on.
	eventStore           *EventStore                      //EventStore for saving log those gen after smart contract executed.
	crossChainStore      *CrossChainStore                 //crossChainStore for saving cross chain msg.
	storedIndexCount     uint32                           //record the count of have saved block index
	currBlockHeight      uint32                           //Current block height
	currBlockHash        common.Uint256                   //Current block hash
	headerCache          map[common.Uint256]*types.Header //BlockHash => Header
	headerIndex          map[uint32]common.Uint256        //Header index, Mapping header height => block hash
	vbftPeerInfoMap      map[uint32]map[string]uint32     //key:block height,value:peerInfo
	lock                 sync.RWMutex
	stateHashCheckHeight uint32

	savingBlockSemaphore       chan bool
	closing                    bool
	preserveBlockHistoryLength uint32 // block could be pruned if blockHeight + preserveBlockHistoryLength < currHeight , disable prune if equals 0
}

//NewLedgerStore return LedgerStoreImp instance
func NewLedgerStore(dataDir string, stateHashHeight uint32) (*LedgerStoreImp, error) {
	ledgerStore := &LedgerStoreImp{
		headerIndex:          make(map[uint32]common.Uint256),
		headerCache:          make(map[common.Uint256]*types.Header, 0),
		vbftPeerInfoMap:      make(map[uint32]map[string]uint32),
		savingBlockSemaphore: make(chan bool, 1),
		stateHashCheckHeight: stateHashHeight,
	}

	blockStore, err := NewBlockStore(fmt.Sprintf("%s%s%s", dataDir, string(os.PathSeparator), DBDirBlock), true)
	if err != nil {
		return nil, fmt.Errorf("NewBlockStore error %s", err)
	}
	ledgerStore.blockStore = blockStore

	crossChainStore, err := NewCrossChainStore(dataDir)
	if err != nil {
		return nil, fmt.Errorf("NewBlockStore error %s", err)
	}
	ledgerStore.crossChainStore = crossChainStore

	dbPath := fmt.Sprintf("%s%s%s", dataDir, string(os.PathSeparator), DBDirState)
	merklePath := fmt.Sprintf("%s%s%s", dataDir, string(os.PathSeparator), MerkleTreeStorePath)
	stateStore, err := NewStateStore(dbPath, merklePath, stateHashHeight)
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
		defaultBookkeeper = keypair.SortPublicKeys(defaultBookkeeper)
		bookkeeperState := &states.BookkeeperState{
			CurrBookkeeper: defaultBookkeeper,
			NextBookkeeper: defaultBookkeeper,
		}
		err = this.stateStore.SaveBookkeeperState(bookkeeperState)
		if err != nil {
			return fmt.Errorf("SaveBookkeeperState error %s", err)
		}

		result, err := this.executeBlock(genesisBlock)
		if err != nil {
			return err
		}
		err = this.submitBlock(genesisBlock, nil, result)
		if err != nil {
			return fmt.Errorf("save genesis block error %s", err)
		}
		err = this.initGenesisBlock()
		if err != nil {
			return fmt.Errorf("init error %s", err)
		}
		genHash := genesisBlock.Hash()
		log.Infof("GenesisBlock init success. GenesisBlock hash:%s\n", genHash.ToHexString())
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
	//load vbft peerInfo
	consensusType := strings.ToLower(config.DefConfig.Genesis.ConsensusType)
	if consensusType == "vbft" {
		header, err := this.GetHeaderByHash(this.currBlockHash)
		if err != nil {
			return err
		}
		blkInfo, err := vconfig.VbftBlock(header)
		if err != nil {
			return err
		}
		var cfg *vconfig.ChainConfig
		var chainConfigHeight uint32
		if blkInfo.NewChainConfig != nil {
			cfg = blkInfo.NewChainConfig
			chainConfigHeight = header.Height
		} else {
			cfgHeader, err := this.GetHeaderByHeight(blkInfo.LastConfigBlockNum)
			if err != nil {
				return err
			}
			Info, err := vconfig.VbftBlock(cfgHeader)
			if err != nil {
				return err
			}
			if Info.NewChainConfig == nil {
				return fmt.Errorf("getNewChainConfig error block num:%d", blkInfo.LastConfigBlockNum)
			}
			cfg = Info.NewChainConfig
			chainConfigHeight = cfgHeader.Height
		}
		this.lock.Lock()
		vbftPeerInfo := make(map[string]uint32)
		this.vbftPeerInfoMap = make(map[uint32]map[string]uint32)
		for _, p := range cfg.Peers {
			vbftPeerInfo[p.ID] = p.Index
		}
		this.vbftPeerInfoMap[chainConfigHeight] = vbftPeerInfo
		this.lock.Unlock()
		val, _ := json.Marshal(vbftPeerInfo)
		log.Infof("loading vbftPeerInfo at height: %s : %s", header.Height, string(val))
	}
	// check and fix imcompatible states
	err = this.stateStore.CheckStorage()
	return err
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
	err := this.loadCurrentBlock()
	if err != nil {
		return fmt.Errorf("loadCurrentBlock error %s", err)
	}
	err = this.loadHeaderIndexList()
	if err != nil {
		return fmt.Errorf("loadHeaderIndexList error %s", err)
	}
	err = this.recoverStore()
	if err != nil {
		return fmt.Errorf("recoverStore error %s", err)
	}
	return nil
}

func (this *LedgerStoreImp) loadCurrentBlock() error {
	currentBlockHash, currentBlockHeight, err := this.blockStore.GetCurrentBlock()
	if err != nil {
		return fmt.Errorf("LoadCurrentBlock error %s", err)
	}
	log.Infof("InitCurrentBlock currentBlockHash %s currentBlockHeight %d", currentBlockHash.ToHexString(), currentBlockHeight)
	this.currBlockHash = currentBlockHash
	this.currBlockHeight = currentBlockHeight
	return nil
}

func (this *LedgerStoreImp) loadHeaderIndexList() error {
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

func (this *LedgerStoreImp) recoverStore() error {
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
		this.eventStore.NewBatch()
		this.stateStore.NewBatch()
		result, err := this.executeBlock(block)
		if err != nil {
			return err
		}
		err = this.saveBlockToStateStore(block, result)
		if err != nil {
			return fmt.Errorf("save to state store height:%d error:%s", i, err)
		}
		this.saveBlockToEventStore(block)
		err = this.eventStore.CommitTo()
		if err != nil {
			return fmt.Errorf("eventStore.CommitTo height:%d error %s", i, err)
		}
		err = this.stateStore.CommitTo()
		if err != nil {
			return fmt.Errorf("stateStore.CommitTo height:%d error %s", i, err)
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
		return fmt.Errorf("cannot find pre header by blockHash %s", prevHeaderHash.ToHexString())
	}

	if prevHeader.Height+1 != header.Height {
		return fmt.Errorf("block height is incorrect")
	}

	if prevHeader.Timestamp >= header.Timestamp {
		return fmt.Errorf("block timestamp is incorrect")
	}
	consensusType := strings.ToLower(config.DefConfig.Genesis.ConsensusType)
	if consensusType == "vbft" {
		blkInfo, err := vconfig.VbftBlock(header)
		if err != nil {
			return err
		}
		var chainConfigHeight uint32
		if blkInfo.NewChainConfig != nil {
			prevBlockInfo, err := vconfig.VbftBlock(prevHeader)
			if err != nil {
				return err
			}
			if prevBlockInfo.NewChainConfig != nil {
				chainConfigHeight = prevHeader.Height
			} else {
				chainConfigHeight = prevBlockInfo.LastConfigBlockNum
			}
		} else {
			chainConfigHeight = blkInfo.LastConfigBlockNum
		}
		this.lock.RLock()
		vbftPeerInfo, ok := this.vbftPeerInfoMap[chainConfigHeight]
		if !ok {
			this.lock.RUnlock()
			return fmt.Errorf("chainconfig height:%d not found", chainConfigHeight)
		}
		this.lock.RUnlock()
		m := len(vbftPeerInfo) - (len(vbftPeerInfo)*6)/7
		if len(header.Bookkeepers) < m {
			return fmt.Errorf("header Bookkeepers %d more than 6/7 len vbftPeerInfo%d", len(header.Bookkeepers), len(vbftPeerInfo))
		}
		for _, bookkeeper := range header.Bookkeepers {
			pubkey := vconfig.PubkeyID(bookkeeper)
			_, present := vbftPeerInfo[pubkey]
			if !present {
				val, _ := json.Marshal(vbftPeerInfo)
				log.Errorf("verify header error: invalid pubkey :%v, height:%d, current vbftPeerInfo :%s",
					pubkey, header.Height, string(val))
				return fmt.Errorf("verify header error: invalid pubkey : %v", pubkey)
			}
		}
		hash := header.Hash()
		err = signature.VerifyMultiSignature(hash[:], header.Bookkeepers, m, header.SigData)
		if err != nil {
			log.Errorf("VerifyMultiSignature:%s,Bookkeepers:%d,pubkey:%d,heigh:%d", err, len(header.Bookkeepers), len(vbftPeerInfo), header.Height)
			return err
		}
		if blkInfo.NewChainConfig != nil {
			peerInfo := make(map[string]uint32)
			for _, p := range blkInfo.NewChainConfig.Peers {
				peerInfo[p.ID] = p.Index
			}
			this.lock.Lock()
			this.vbftPeerInfoMap[header.Height] = peerInfo
			this.lock.Unlock()
		}
	} else {
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

func (this *LedgerStoreImp) verifyCrossChainMsg(crossChainMsg *types.CrossChainMsg, bookkeepers []keypair.PublicKey) error {
	consensusType := strings.ToLower(config.DefConfig.Genesis.ConsensusType)
	hash := crossChainMsg.Hash()
	if consensusType == "vbft" {
		err := signature.VerifyMultiSignature(hash[:], bookkeepers, len(bookkeepers), crossChainMsg.SigData)
		if err != nil {
			log.Errorf("vbft VerifyMultiSignature:%s,heigh:%d", err, crossChainMsg.Height)
			return err
		}
	} else {
		m := len(bookkeepers) - (len(bookkeepers)-1)/3
		err := signature.VerifyMultiSignature(hash[:], bookkeepers, m, crossChainMsg.SigData)
		if err != nil {
			log.Errorf("VerifyMultiSignature:%s,heigh:%d", err, crossChainMsg.Height)
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
	//this.vbftPeerInfoheader, err = this.verifyHeader(header, this.vbftPeerInfoheader)
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

func (this *LedgerStoreImp) GetStateMerkleRoot(height uint32) (common.Uint256, error) {
	return this.stateStore.GetStateMerkleRoot(height)
}

func (this *LedgerStoreImp) ExecuteBlock(block *types.Block) (result store.ExecuteResult, err error) {
	this.getSavingBlockLock()
	defer this.releaseSavingBlockLock()
	currBlockHeight := this.GetCurrentBlockHeight()
	blockHeight := block.Header.Height
	if blockHeight <= currBlockHeight {
		result.MerkleRoot, err = this.GetStateMerkleRoot(blockHeight)
		return
	}
	nextBlockHeight := currBlockHeight + 1
	if blockHeight != nextBlockHeight {
		err = fmt.Errorf("block height %d not equal next block height %d", blockHeight, nextBlockHeight)
		return
	}

	result, err = this.executeBlock(block)
	return
}

func (this *LedgerStoreImp) SubmitBlock(block *types.Block, ccMsg *types.CrossChainMsg, result store.ExecuteResult) error {
	this.getSavingBlockLock()
	defer this.releaseSavingBlockLock()
	if this.closing {
		return errors.NewErr("save block error: ledger is closing")
	}
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
	if ccMsg != nil {
		if ccMsg.Height != currBlockHeight {
			return fmt.Errorf("cross chain msg height %d not equal next block height %d", blockHeight, ccMsg.Height)
		}
		if ccMsg.Version != types.CURR_CROSS_STATES_VERSION {
			return fmt.Errorf("error cross chain msg version excepted:%d actual:%d", types.CURR_CROSS_STATES_VERSION, ccMsg.Version)
		}
		root, err := this.stateStore.GetCrossStatesRoot(ccMsg.Height)
		if err != nil {
			return fmt.Errorf("get cross states root fail:%s", err)
		}
		if root != ccMsg.StatesRoot {
			return fmt.Errorf("cross state root compare fail, expected:%x actual:%x", ccMsg.StatesRoot, root)
		}
		if err := this.verifyCrossChainMsg(ccMsg, block.Header.Bookkeepers); err != nil {
			return fmt.Errorf("verifyCrossChainMsg error: %s", err)
		}
	}

	err = this.submitBlock(block, ccMsg, result)
	if err != nil {
		return fmt.Errorf("saveBlock error %s", err)
	}
	this.delHeaderCache(block.Hash())
	return nil
}

//AddBlock add the block to store.
//When the block is not the next block, it will be cache. until the missing block arrived
func (this *LedgerStoreImp) AddBlock(block *types.Block, ccMsg *types.CrossChainMsg, stateMerkleRoot common.Uint256) error {
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
	if ccMsg != nil {
		if ccMsg.Height != currBlockHeight {
			return fmt.Errorf("cross chain msg height %d not equal next block height %d", blockHeight, ccMsg.Height)
		}
		if ccMsg.Version != types.CURR_CROSS_STATES_VERSION {
			return fmt.Errorf("error cross chain msg version excepted:%d actual:%d", types.CURR_CROSS_STATES_VERSION, ccMsg.Version)
		}
		root, err := this.stateStore.GetCrossStatesRoot(ccMsg.Height)
		if err != nil {
			return fmt.Errorf("get cross states root fail:%s", err)
		}
		if root != ccMsg.StatesRoot {
			return fmt.Errorf("cross state root compare fail, expected:%x actual:%x", ccMsg.StatesRoot, root)
		}
		if err := this.verifyCrossChainMsg(ccMsg, block.Header.Bookkeepers); err != nil {
			return fmt.Errorf("verifyCrossChainMsg error: %s", err)
		}
	}
	err = this.saveBlock(block, ccMsg, stateMerkleRoot)
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
		return fmt.Errorf("SaveBlock height %d hash %s error %s", blockHeight, blockHash.ToHexString(), err)
	}
	return nil
}

func (this *LedgerStoreImp) executeBlock(block *types.Block) (result store.ExecuteResult, err error) {
	overlay := this.stateStore.NewOverlayDB()
	if block.Header.Height != 0 {
		config := &smartcontract.Config{
			Time:   block.Header.Timestamp,
			Height: block.Header.Height,
			Tx:     &types.Transaction{},
		}

		err = refreshGlobalParam(config, storage.NewCacheDB(this.stateStore.NewOverlayDB()), this)
		if err != nil {
			return
		}
	}
	gasTable := make(map[string]uint64)
	neovm.GAS_TABLE.Range(func(k, value interface{}) bool {
		key := k.(string)
		val := value.(uint64)
		gasTable[key] = val

		return true
	})

	cache := storage.NewCacheDB(overlay)
	for _, tx := range block.Transactions {
		cache.Reset()
		notify, crossStateHashes, e := this.handleTransaction(overlay, cache, gasTable, block, tx)
		if e != nil {
			err = e
			return
		}
		result.Notify = append(result.Notify, notify)
		result.CrossStates = append(result.CrossStates, crossStateHashes...)
	}
	result.Hash = overlay.ChangeHash()
	result.WriteSet = overlay.GetWriteSet()
	if len(result.CrossStates) != 0 {
		log.Infof("executeBlock: %d cross states generated at block height:%d", len(result.CrossStates), block.Header.Height)
		result.CrossStatesRoot = merkle.TreeHasher{}.HashFullTreeWithLeafHash(result.CrossStates)
	} else {
		result.CrossStatesRoot = common.UINT256_EMPTY
	}
	if block.Header.Height < this.stateHashCheckHeight {
		result.MerkleRoot = common.UINT256_EMPTY
	} else if block.Header.Height == this.stateHashCheckHeight {
		res, e := calculateTotalStateHash(overlay)
		if e != nil {
			err = e
			return
		}

		result.MerkleRoot = res
		result.Hash = result.MerkleRoot
	} else {
		result.MerkleRoot = this.stateStore.GetStateMerkleRootWithNewHash(result.Hash)
	}

	return
}

func calculateTotalStateHash(overlay *overlaydb.OverlayDB) (result common.Uint256, err error) {
	stateDiff := sha256.New()
	iter := overlay.NewIterator([]byte{byte(scom.ST_CONTRACT)})
	err = accumulateHash(stateDiff, iter)
	iter.Release()
	if err != nil {
		return
	}

	iter = overlay.NewIterator([]byte{byte(scom.ST_STORAGE)})
	err = accumulateHash(stateDiff, iter)
	iter.Release()
	if err != nil {
		return
	}

	stateDiff.Sum(result[:0])
	return
}

func accumulateHash(hasher hash.Hash, iter scom.StoreIterator) error {
	for has := iter.First(); has; has = iter.Next() {
		key := iter.Key()
		val := iter.Value()
		hasher.Write(key)
		hasher.Write(val)
	}
	return iter.Error()
}

func (this *LedgerStoreImp) saveBlockToStateStore(block *types.Block, result store.ExecuteResult) error {
	blockHash := block.Hash()
	blockHeight := block.Header.Height

	for _, notify := range result.Notify {
		SaveNotify(this.eventStore, notify.TxHash, notify)
	}

	err := this.stateStore.AddStateMerkleTreeRoot(blockHeight, result.Hash)
	if err != nil {
		return fmt.Errorf("AddBlockMerkleTreeRoot error %s", err)
	}

	err = this.stateStore.AddBlockMerkleTreeRoot(block.Header.TransactionsRoot)
	if err != nil {
		return fmt.Errorf("AddBlockMerkleTreeRoot error %s", err)
	}

	err = this.stateStore.SaveCurrentBlock(blockHeight, blockHash)
	if err != nil {
		return fmt.Errorf("SaveCurrentBlock error %s", err)
	}

	err = this.stateStore.SaveCrossStates(blockHeight, result.CrossStates)
	if err != nil {
		return fmt.Errorf("SaveCrossStates error %s", err)
	}

	log.Debugf("the state transition hash of block %d is:%s", blockHeight, result.Hash.ToHexString())

	result.WriteSet.ForEach(func(key, val []byte) {
		if len(val) == 0 {
			this.stateStore.BatchDeleteRawKey(key)
		} else {
			this.stateStore.BatchPutRawKeyVal(key, val)
		}
	})

	return nil
}

func (this *LedgerStoreImp) saveBlockToEventStore(block *types.Block) {
	blockHash := block.Hash()
	blockHeight := block.Header.Height
	txs := make([]common.Uint256, 0)
	for _, tx := range block.Transactions {
		txHash := tx.Hash()
		txs = append(txs, txHash)
	}
	if len(txs) > 0 {
		this.eventStore.SaveEventNotifyByBlock(block.Header.Height, txs)
	}
	this.eventStore.SaveCurrentBlock(blockHeight, blockHash)
}

func (this *LedgerStoreImp) tryGetSavingBlockLock() (hasLocked bool) {
	select {
	case this.savingBlockSemaphore <- true:
		return false
	default:
		return true
	}
}

func (this *LedgerStoreImp) getSavingBlockLock() {
	this.savingBlockSemaphore <- true
}

func (this *LedgerStoreImp) releaseSavingBlockLock() {
	select {
	case <-this.savingBlockSemaphore:
		return
	default:
		panic("can not release in unlocked state")
	}
}

const pruneBatchSize = 10

func (this *LedgerStoreImp) tryPruneBlock(header *types.Header) bool {
	if this.preserveBlockHistoryLength == 0 {
		return false
	}
	height := this.maxAllowedPruneHeight(header)
	if height+this.preserveBlockHistoryLength >= header.Height {
		height = header.Height - this.preserveBlockHistoryLength
	}
	pruned, err := this.blockStore.GetBlockPrunedHeight()
	if err != nil {
		return false
	}
	if pruned+1 >= height {
		return false
	}

	pruneHeight := pruned + 1
	for ; pruneHeight-pruned < pruneBatchSize && pruneHeight < height; pruneHeight++ {
		hash := this.GetBlockHash(pruneHeight)
		txHashes := this.blockStore.PruneBlock(hash)
		this.eventStore.PruneBlock(pruneHeight, txHashes)
	}
	this.blockStore.SaveBlockPrunedHeight(pruneHeight)
	return true
}

//saveBlock do the job of execution samrt contract and commit block to store.
func (this *LedgerStoreImp) submitBlock(block *types.Block, crossChainMsg *types.CrossChainMsg, result store.ExecuteResult) error {
	blockHash := block.Hash()
	blockHeight := block.Header.Height
	blockRoot := this.GetBlockRootWithNewTxRoots(block.Header.Height, []common.Uint256{block.Header.TransactionsRoot})
	if block.Header.Height != 0 && blockRoot != block.Header.BlockRoot {
		return fmt.Errorf("wrong block root at height:%d, expected:%s, got:%s",
			block.Header.Height, blockRoot.ToHexString(), block.Header.BlockRoot.ToHexString())
	}

	this.blockStore.NewBatch()
	this.stateStore.NewBatch()
	this.eventStore.NewBatch()
	err := this.saveBlockToBlockStore(block)
	if err != nil {
		return fmt.Errorf("save to block store height:%d error:%s", blockHeight, err)
	}
	this.tryPruneBlock(block.Header)
	err = this.crossChainStore.SaveMsgToCrossChainStore(crossChainMsg)
	if err != nil {
		return fmt.Errorf("save to msg cross chain store height:%d error:%s", blockHeight, err)
	}
	err = this.saveBlockToStateStore(block, result)
	if err != nil {
		return fmt.Errorf("save to state store height:%d error:%s", blockHeight, err)
	}
	this.saveBlockToEventStore(block)
	err = this.blockStore.CommitTo()
	if err != nil {
		return fmt.Errorf("blockStore.CommitTo height:%d error %s", blockHeight, err)
	}
	// event store is idempotent to re-save when in recovering process, so save first before stateStore
	err = this.eventStore.CommitTo()
	if err != nil {
		return fmt.Errorf("eventStore.CommitTo height:%d error %s", blockHeight, err)
	}
	err = this.stateStore.CommitTo()
	if err != nil {
		return fmt.Errorf("stateStore.CommitTo height:%d error %s", blockHeight, err)
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

//saveBlock do the job of execution samrt contract and commit block to store.
func (this *LedgerStoreImp) saveBlock(block *types.Block, ccMsg *types.CrossChainMsg, stateMerkleRoot common.Uint256) error {
	blockHeight := block.Header.Height
	if blockHeight > 0 && blockHeight <= this.GetCurrentBlockHeight() {
		return nil
	}
	this.getSavingBlockLock()
	defer this.releaseSavingBlockLock()
	if this.closing {
		return errors.NewErr("save block error: ledger is closing")
	}
	if blockHeight > 0 && blockHeight != (this.GetCurrentBlockHeight()+1) {
		return nil
	}

	result, err := this.executeBlock(block)
	if err != nil {
		return err
	}

	//empty block does not check stateMerkleRoot
	if len(block.Transactions) != 0 && result.MerkleRoot != stateMerkleRoot {
		log.Infof("state mismatch at block height: %d, changeset: %s", block.Header.Height, result.WriteSet.DumpToDot())
		return fmt.Errorf("state merkle root mismatch. expected: %s, got: %s",
			result.MerkleRoot.ToHexString(), stateMerkleRoot.ToHexString())
	}

	return this.submitBlock(block, ccMsg, result)
}

func (this *LedgerStoreImp) handleTransaction(overlay *overlaydb.OverlayDB, cache *storage.CacheDB, gasTable map[string]uint64,
	block *types.Block, tx *types.Transaction) (*event.ExecuteNotify, []common.Uint256, error) {
	txHash := tx.Hash()
	notify := &event.ExecuteNotify{TxHash: txHash, State: event.CONTRACT_STATE_FAIL}
	var crossStateHashes []common.Uint256
	var err error
	switch tx.TxType {
	case types.Deploy:
		err = this.stateStore.HandleDeployTransaction(this, overlay, gasTable, cache, tx, block, notify)
		if overlay.Error() != nil {
			return nil, nil, fmt.Errorf("HandleDeployTransaction tx %s error %s", txHash.ToHexString(), overlay.Error())
		}
		if err != nil {
			log.Debugf("HandleDeployTransaction tx %s error %s", txHash.ToHexString(), err)
		}
	case types.InvokeNeo, types.InvokeWasm:
		crossStateHashes, err = this.stateStore.HandleInvokeTransaction(this, overlay, gasTable, cache, tx, block, notify)
		if overlay.Error() != nil {
			return nil, nil, fmt.Errorf("HandleInvokeTransaction tx %s error %s", txHash.ToHexString(), overlay.Error())
		}
		if err != nil {
			log.Debugf("HandleInvokeTransaction tx %s error %s", txHash.ToHexString(), err)
		}
	}
	return notify, crossStateHashes, nil
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

	this.blockStore.SaveHeaderIndexList(storeCount, headerList)

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

//GetBlockRootWithNewTxRoots return the block root(merkle root of blocks) after add a new tx root of block
func (this *LedgerStoreImp) GetBlockRootWithNewTxRoots(startHeight uint32, txRoots []common.Uint256) common.Uint256 {
	this.lock.RLock()
	defer this.lock.RUnlock()
	// the block height in consensus is far behind ledger, this case should be rare
	if this.currBlockHeight > startHeight+uint32(len(txRoots))-1 {
		// or return error?
		return common.UINT256_EMPTY
	} else if this.currBlockHeight+1 < startHeight {
		// this should never happen in normal case
		log.Fatalf("GetBlockRootWithNewTxRoots: invalid param: curr height: %d, start height: %d",
			this.currBlockHeight, startHeight)

		return common.UINT256_EMPTY
	}

	needs := txRoots[this.currBlockHeight+1-startHeight:]
	return this.stateStore.GetBlockRootWithNewTxRoots(needs)
}

func (this *LedgerStoreImp) GetCrossStatesRoot(height uint32) (common.Uint256, error) {
	return this.stateStore.GetCrossStatesRoot(height)
}

func (this *LedgerStoreImp) GetCrossChainMsg(height uint32) (*types.CrossChainMsg, error) {
	return this.crossChainStore.GetCrossChainMsg(height)
}

func (this *LedgerStoreImp) GetCrossStatesProof(height uint32, key []byte) ([]byte, error) {
	hashes, err := this.stateStore.GetCrossStates(height)
	if err != nil {
		return nil, fmt.Errorf("GetCrossStates:%s", err)
	}
	item, err := this.stateStore.GetStorageState(&states.StorageKey{ContractAddress: utils.CrossChainContractAddress, Key: key})
	if err != nil {
		return nil, fmt.Errorf("GetStorageState key:%x", key)
	}
	path, err := merkle.MerkleLeafPath(item.Value, hashes)
	if err != nil {
		return nil, err
	}
	return path, nil
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

func (this *LedgerStoreImp) GetRawHeaderByHash(blockHash common.Uint256) (*types.RawHeader, error) {
	header := this.getHeaderCache(blockHash)
	if header != nil {
		return header.GetRawHeader(), nil
	}
	return this.blockStore.GetRawHeader(blockHash)
}

//GetHeaderByHash return the block header by block height
func (this *LedgerStoreImp) GetHeaderByHeight(height uint32) (*types.Header, error) {
	blockHash := this.GetBlockHash(height)
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
func (this *LedgerStoreImp) PreExecuteContractBatch(txes []*types.Transaction, atomic bool) ([]*sstate.PreExecResult, uint32, error) {
	if atomic {
		this.getSavingBlockLock()
		defer this.releaseSavingBlockLock()
	}
	height := this.GetCurrentBlockHeight()
	results := make([]*sstate.PreExecResult, 0, len(txes))
	for _, tx := range txes {
		res, err := this.PreExecuteContract(tx)
		if err != nil {
			return nil, height, err
		}

		results = append(results, res)
	}

	return results, height, nil
}

//PreExecuteContract return the result of smart contract execution without commit to store
func (this *LedgerStoreImp) PreExecuteContractWithParam(tx *types.Transaction, preParam PrexecuteParam) (*sstate.PreExecResult, error) {
	height := this.GetCurrentBlockHeight()
	// use previous block time to make it predictable for easy test
	blockTime := uint32(time.Now().Unix())
	if header, err := this.GetHeaderByHeight(height); err == nil {
		blockTime = header.Timestamp + 1
	}
	stf := &sstate.PreExecResult{State: event.CONTRACT_STATE_FAIL, Gas: neovm.MIN_TRANSACTION_GAS, Result: nil}

	sconfig := &smartcontract.Config{
		Time:      blockTime,
		Height:    height + 1,
		Tx:        tx,
		BlockHash: this.GetBlockHash(height),
	}

	overlay := this.stateStore.NewOverlayDB()
	cache := storage.NewCacheDB(overlay)
	gasTable := make(map[string]uint64)
	neovm.GAS_TABLE.Range(func(k, value interface{}) bool {
		key := k.(string)
		val := value.(uint64)
		if key == config.WASM_GAS_FACTOR && preParam.WasmFactor != 0 {
			gasTable[key] = preParam.WasmFactor
		} else {
			gasTable[key] = val
		}

		return true
	})

	if tx.TxType == types.InvokeNeo || tx.TxType == types.InvokeWasm {
		invoke := tx.Payload.(*payload.InvokeCode)

		sc := smartcontract.SmartContract{
			Config:       sconfig,
			Store:        this,
			CacheDB:      cache,
			GasTable:     gasTable,
			Gas:          math.MaxUint64 - calcGasByCodeLen(len(invoke.Code), gasTable[neovm.UINT_INVOKE_CODE_LEN_NAME]),
			WasmExecStep: config.DEFAULT_WASM_MAX_STEPCOUNT,
			JitMode:      preParam.JitMode,
			PreExec:      true,
		}
		//start the smart contract executive function
		engine, _ := sc.NewExecuteEngine(invoke.Code, tx.TxType)

		result, err := engine.Invoke()
		if err != nil {
			return stf, err
		}
		gasCost := math.MaxUint64 - sc.Gas

		if preParam.MinGas {
			mixGas := neovm.MIN_TRANSACTION_GAS
			if gasCost < mixGas {
				gasCost = mixGas
			}
			gasCost = tuneGasFeeByHeight(sconfig.Height, gasCost, neovm.MIN_TRANSACTION_GAS, math.MaxUint64)
		}

		var cv interface{}
		if tx.TxType == types.InvokeNeo { //neovm
			if result != nil {
				val := result.(*types2.VmValue)
				cv, err = val.ConvertNeoVmValueHexString()
				if err != nil {
					return stf, err
				}
			}
		} else { //wasmvm
			cv = common.ToHexString(result.([]byte))
		}

		return &sstate.PreExecResult{State: event.CONTRACT_STATE_SUCCESS, Gas: gasCost, Result: cv, Notify: sc.Notifications}, nil
	} else if tx.TxType == types.Deploy {
		deploy := tx.Payload.(*payload.DeployCode)

		if deploy.VmType() == payload.WASMVM_TYPE {
			wasmCode := deploy.GetRawCode()
			err := wasmvm.WasmjitValidate(wasmCode)
			if err != nil {
				return stf, err
			}
		} else {
			wasmMagicversion := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}

			if len(deploy.GetRawCode()) >= len(wasmMagicversion) {
				if bytes.Compare(wasmMagicversion, deploy.GetRawCode()[:8]) == 0 {
					return stf, errors.NewErr("this code is wasm binary. can not deployed as neo contract")
				}
			}
		}

		return &sstate.PreExecResult{State: event.CONTRACT_STATE_SUCCESS, Gas: gasTable[neovm.CONTRACT_CREATE_NAME] + calcGasByCodeLen(len(deploy.GetRawCode()), gasTable[neovm.UINT_DEPLOY_CODE_LEN_NAME]), Result: nil}, nil
	} else {
		return stf, errors.NewErr("transaction type error")
	}
}

//PreExecuteContract return the result of smart contract execution without commit to store
func (this *LedgerStoreImp) PreExecuteContract(tx *types.Transaction) (*sstate.PreExecResult, error) {
	param := PrexecuteParam{
		JitMode:    false,
		WasmFactor: 0,
		MinGas:     true,
	}

	return this.PreExecuteContractWithParam(tx, param)
}

//Close ledger store.
func (this *LedgerStoreImp) Close() error {
	// wait block saving complete, and get the lock to avoid subsequent block saving
	this.getSavingBlockLock()
	defer this.releaseSavingBlockLock()

	this.closing = true

	err := this.blockStore.Close()
	if err != nil {
		return fmt.Errorf("blockStore close error %s", err)
	}
	err = this.eventStore.Close()
	if err != nil {
		return fmt.Errorf("eventStore close error %s", err)
	}
	err = this.stateStore.Close()
	if err != nil {
		return fmt.Errorf("stateStore close error %s", err)
	}
	return nil
}

const minPruneBlocksBeforeCurr = 1000

func (this *LedgerStoreImp) EnableBlockPrune(numBeforeCurr uint32) {
	if numBeforeCurr < minPruneBlocksBeforeCurr {
		numBeforeCurr = minPruneBlocksBeforeCurr
	}
	this.getSavingBlockLock()
	defer this.releaseSavingBlockLock()

	this.preserveBlockHistoryLength = numBeforeCurr
}

func (this *LedgerStoreImp) maxAllowedPruneHeight(currHeader *types.Header) uint32 {
	if currHeader.Height <= config.GetContractApiDeprecateHeight() {
		return 0
	}
	info, err := vconfig.VbftBlock(currHeader)
	if err != nil {
		return 0
	}
	lastReferHeight := info.LastConfigBlockNum
	if info.NewChainConfig != nil {
		lastReferHeight = currHeader.Height
	}

	if lastReferHeight == 0 {
		return 0
	}
	return lastReferHeight - 1
}
