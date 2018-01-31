package ChainStore

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	. "github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"github.com/Ontology/common/serialization"
	"github.com/Ontology/core/contract/program"
	. "github.com/Ontology/core/ledger"
	"github.com/Ontology/core/states"
	. "github.com/Ontology/core/store"
	. "github.com/Ontology/core/store/LevelDBStore"
	"github.com/Ontology/core/store/statestore"
	tx "github.com/Ontology/core/transaction"
	"github.com/Ontology/core/transaction/payload"
	"github.com/Ontology/core/transaction/utxo"
	"github.com/Ontology/core/validation"
	"github.com/Ontology/crypto"
	"github.com/Ontology/events"
	"github.com/Ontology/merkle"
	httprestful "github.com/Ontology/net/httprestful/error"
	sc "github.com/Ontology/smartcontract"
	"github.com/Ontology/smartcontract/event"
	"github.com/Ontology/smartcontract/service"
	"github.com/Ontology/smartcontract/types"
	"sort"
	"sync"
	"time"
)

const (
	HeaderHashListCount = 2000
	CleanCacheThreshold = 2
	TaskChanCap         = 4
	DBDir               = "Chain"
	MerkleTreeStorePath = "Chain/merkle_tree.db"
	DEPLOY_TRANSACTION  = "DeployTransaction"
	INVOKE_TRANSACTION  = "InvokeTransaction"
)

var (
	ErrDBNotFound    = "leveldb: not found"
	CurrentStateRoot = []byte("Current-State-Root")
	BookerKeeper     = []byte("Booker-Keeper")
)

type persistTask interface{}
type persistHeaderTask struct {
	header *Header
}
type persistBlockTask struct {
	block  *Block
	ledger *Ledger
}

type ChainStore struct {
	st IStore

	taskCh chan persistTask
	quit   chan chan bool

	mu          sync.RWMutex // guard the following var
	headerIndex map[uint32]Uint256
	blockCache  map[Uint256]*Block
	headerCache map[Uint256]*Header

	merkleTree      *merkle.CompactMerkleTree
	merkleHashStore *merkle.FileHashStore

	currentBlockHeight uint32
	storedHeaderCount  uint32
}

func NewStore(file string) (IStore, error) {
	ldbs, err := NewLevelDBStore(file)

	return ldbs, err
}

func NewLedgerStore() (ILedgerStore, error) {
	// TODO: read config file decide which db to use.
	cs, err := NewChainStore(DBDir)
	if err != nil {
		return nil, err
	}

	return cs, nil
}

func NewChainStore(file string) (*ChainStore, error) {

	st, err := NewStore(file)
	if err != nil {
		return nil, err
	}

	chain := &ChainStore{
		st:                 st,
		headerIndex:        map[uint32]Uint256{},
		blockCache:         map[Uint256]*Block{},
		headerCache:        map[Uint256]*Header{},
		currentBlockHeight: 0,
		storedHeaderCount:  0,
		taskCh:             make(chan persistTask, TaskChanCap),
		quit:               make(chan chan bool, 1),
	}

	go chain.loop()

	return chain, nil
}

func (self *ChainStore) Close() {
	closed := make(chan bool)
	self.quit <- closed
	<-closed

	self.st.Close()
}

func (self *ChainStore) loop() {
	for {
		select {
		case t := <-self.taskCh:
			now := time.Now()
			switch task := t.(type) {
			case *persistHeaderTask:
				self.handlePersistHeaderTask(task.header)
				tcall := float64(time.Now().Sub(now)) / float64(time.Second)
				log.Debugf("handle header exetime: %g \n", tcall)

			case *persistBlockTask:
				self.handlePersistBlockTask(task.block, task.ledger)
				tcall := float64(time.Now().Sub(now)) / float64(time.Second)
				log.Debugf("handle block exetime: %g num transactions:%d \n", tcall, len(task.block.Transactions))
			}

		case closed := <-self.quit:
			closed <- true
			return
		}
	}
}

// can only be invoked by backend write goroutine
func (self *ChainStore) clearCache() {
	self.mu.Lock()
	defer self.mu.Unlock()

	currBlockHeight := self.currentBlockHeight
	for hash, header := range self.headerCache {
		if header.Height+CleanCacheThreshold < currBlockHeight {
			delete(self.headerCache, hash)
		}
	}

	for hash, block := range self.blockCache {
		if block.Header.Height+CleanCacheThreshold < currBlockHeight {
			delete(self.blockCache, hash)
		}
	}
}

func (bd *ChainStore) InitLedgerStoreWithGenesisBlock(genesisBlock *Block, defaultBookKeeper []*crypto.PubKey) (uint32, error) {
	hash := genesisBlock.Hash()
	bd.headerIndex[0] = hash

	prefix := []byte{byte(SYS_Version)}
	version, err := bd.st.Get(prefix)
	if err != nil {
		version = []byte{0x00}
	}

	if version[0] == 0x01 {
		// GenesisBlock should exist in chain
		// Or the bookkeepers are not consistent with the chain
		if !bd.IsBlockInStore(hash) {
			return 0, errors.New("bookkeepers are not consistent with the chain")
		}
		// Get Current Block
		currentBlockPrefix := []byte{byte(SYS_CurrentBlock)}
		data, err := bd.st.Get(currentBlockPrefix)
		if err != nil {
			return 0, err
		}

		r := bytes.NewReader(data)
		var blockHash Uint256
		blockHash.Deserialize(r)
		bd.currentBlockHeight, err = serialization.ReadUint32(r)
		current_Header_Height := bd.currentBlockHeight

		var listHash Uint256
		iter := bd.st.NewIterator([]byte{byte(IX_HeaderHashList)})
		for iter.Next() {
			rk := bytes.NewReader(iter.Key())
			// read prefix
			_, _ = serialization.ReadBytes(rk, 1)
			startNum, err := serialization.ReadUint32(rk)
			if err != nil {
				return 0, err
			}
			log.Debugf("start index: %d\n", startNum)

			r = bytes.NewReader(iter.Value())
			listNum, err := serialization.ReadVarUint(r, 0)
			if err != nil {
				return 0, err
			}

			for i := 0; i < int(listNum); i++ {
				listHash.Deserialize(r)
				bd.headerIndex[startNum+uint32(i)] = listHash
				bd.storedHeaderCount++
				//log.Debug( fmt.Sprintf( "listHash %d: %x\n", startNum+uint32(i), listHash ) )
			}
		}

		if bd.storedHeaderCount == 0 {
			iter = bd.st.NewIterator([]byte{byte(DATA_Block)})
			for iter.Next() {
				rk := bytes.NewReader(iter.Key())
				// read prefix
				_, _ = serialization.ReadBytes(rk, 1)
				listheight, err := serialization.ReadUint32(rk)
				if err != nil {
					return 0, err
				}
				//log.Debug(fmt.Sprintf( "DATA_BlockHash block height: %d\n", listheight ))

				r := bytes.NewReader(iter.Value())
				listHash.Deserialize(r)
				//log.Debug(fmt.Sprintf( "DATA_BlockHash block hash: %x\n", listHash ))

				bd.headerIndex[listheight] = listHash
			}
		} else if current_Header_Height >= bd.storedHeaderCount {
			hash = blockHash
			for {
				if hash == bd.headerIndex[bd.storedHeaderCount-1] {
					break
				}

				header, err := bd.GetHeader(hash)
				if err != nil {
					return 0, err
				}

				//log.Debug(fmt.Sprintf( "header height: %d\n", header.Height ))
				//log.Debug(fmt.Sprintf( "header hash: %x\n", hash ))

				bd.headerIndex[header.Height] = hash
				hash = header.PrevBlockHash
			}
		}
		buf, _ := bd.st.Get([]byte{byte(SYS_BlockMerkleTree)})
		tree_size := binary.BigEndian.Uint32(buf[0:4])
		if tree_size != bd.currentBlockHeight+1 {
			return 0, errors.New("Merkle tree size is inconsistent with blockheight")
		}
		nhashes := (len(buf) - 4) / UINT256SIZE
		hashes := make([]Uint256, nhashes, nhashes)
		for i := 0; i < nhashes; i++ {
			copy(hashes[i][:], buf[4+i*UINT256SIZE:])
		}

		bd.merkleHashStore, err = merkle.NewFileHashStore(MerkleTreeStorePath, tree_size)
		if err != nil {
			log.Error("merkle_tree.db is inconsistent with ChainStore. persistence will be disabled")
		}
		bd.merkleTree = merkle.NewTree(tree_size, hashes, bd.merkleHashStore)
		return bd.currentBlockHeight, nil

	} else {
		// batch delete old data
		bd.st.NewBatch()
		iter := bd.st.NewIterator(nil)
		for iter.Next() {
			bd.st.BatchDelete(iter.Key())
		}
		iter.Release()

		err := bd.st.BatchCommit()
		if err != nil {
			return 0, err
		}

		///////////////////////////////////////////////////
		// process defaultBookKeeper
		///////////////////////////////////////////////////
		// sort defaultBookKeeper
		sort.Sort(crypto.PubKeySlice(defaultBookKeeper))

		// currBookKeeper key
		bkListKey := bytes.NewBuffer(nil)
		bkListKey.Write(append([]byte{byte(ST_BookKeeper)}, BookerKeeper...))

		bkListValue := bytes.NewBuffer(nil)
		bookKeeper := new(states.BookKeeperState)
		bookKeeper.CurrBookKeeper = defaultBookKeeper
		bookKeeper.NextBookKeeper = defaultBookKeeper
		err = bookKeeper.Serialize(bkListValue)
		if err != nil {
			return 0, err
		}

		// defaultBookKeeper put value
		bd.st.Put(bkListKey.Bytes(), bkListValue.Bytes())
		///////////////////////////////////////////////////

		// Init merkle tree and hash store
		bd.merkleHashStore, _ = merkle.NewFileHashStore(MerkleTreeStorePath, 0)
		bd.merkleTree = merkle.NewTree(0, nil, bd.merkleHashStore)

		// persist genesis block
		bd.persist(genesisBlock)

		// put version to db
		err = bd.st.Put(prefix, []byte{0x01})
		if err != nil {
			return 0, err
		}

		return 0, nil
	}
}

func (bd *ChainStore) InitLedgerStore(l *Ledger) error {
	// TODO: InitLedgerStore
	return nil
}

func (bd *ChainStore) IsTxHashDuplicate(txhash Uint256) bool {
	prefix := []byte{byte(DATA_Transaction)}
	_, err_get := bd.st.Get(append(prefix, txhash.ToArray()...))
	if err_get != nil {
		return false
	} else {
		return true
	}
}

func (bd *ChainStore) GetBlockRootWithNewTxRoot(txRoot Uint256) Uint256 {
	return bd.merkleTree.GetRootWithNewLeaf(txRoot)
}

func (bd *ChainStore) IsDoubleSpend(tx *tx.Transaction) bool {
	if len(tx.UTXOInputs) == 0 {
		return false
	}

	unspentPrefix := []byte{byte(ST_Coin)}
	for k, group := range groupInputs(tx.UTXOInputs) {
		unspentValue, err_get := bd.st.Get(append(unspentPrefix, k.ToArray()...))
		if err_get != nil {
			return true
		}
		unspentcoin := new(states.UnspentCoinState)
		bf := bytes.NewBuffer(unspentValue)
		if err := unspentcoin.Deserialize(bf); err != nil {
			log.Error("[IsDoubleSpend] error:", err)
			return true
		}
		for _, u := range group {
			index := int(u.ReferTxOutputIndex)
			if index >= len(unspentcoin.Item) || unspentcoin.Item[index] == states.Spent {
				return true
			}
		}
		return false
	}

	return false
}

func (bd *ChainStore) GetBlockHash(height uint32) (Uint256, error) {
	queryKey := bytes.NewBuffer(nil)
	queryKey.WriteByte(byte(DATA_Block))
	err := serialization.WriteUint32(queryKey, height)

	if err != nil {
		return Uint256{}, err
	}
	blockHash, err_get := bd.st.Get(queryKey.Bytes())
	if err_get != nil {
		//TODO: implement error process
		return Uint256{}, err_get
	}
	blockHash256, err_parse := Uint256ParseFromBytes(blockHash)
	if err_parse != nil {
		return Uint256{}, err_parse
	}

	return blockHash256, nil
}

func (bd *ChainStore) GetCurrentBlockHash() Uint256 {
	bd.mu.RLock()
	defer bd.mu.RUnlock()

	return bd.headerIndex[bd.currentBlockHeight]
}

func (bd *ChainStore) getHeaderWithCache(hash Uint256) *Header {
	if _, ok := bd.headerCache[hash]; ok {
		return bd.headerCache[hash]
	}

	header, _ := bd.GetHeader(hash)

	return header
}

func (bd *ChainStore) verifyHeader(header *Header) bool {
	prevHeader := bd.getHeaderWithCache(header.PrevBlockHash)

	if prevHeader == nil {
		log.Error("[verifyHeader] failed, not found prevHeader.")
		return false
	}

	if prevHeader.Height+1 != header.Height {
		log.Error("[verifyHeader] failed, prevHeader.Height + 1 != header.Height")
		return false
	}

	if prevHeader.Timestamp >= header.Timestamp {
		log.Error("[verifyHeader] failed, prevHeader.Timestamp >= header.Timestamp")
		return false
	}

	err := validation.VerifySignableDataSignature(header)
	if err != nil {
		log.Error("[verifyHeader] failed, VerifySignableDataSignature failed.", err.Error())
		return false
	}
	err = validation.VerifySignableDataProgramHashes(header)
	if err != nil {
		log.Error("[verifyHeader] failed, VerifySignableDataProgramHashes failed.", err.Error())
		return false
	}

	return true
}

func (self *ChainStore) AddHeaders(headers []Header, ledger *Ledger) error {

	sort.Slice(headers, func(i, j int) bool {
		return headers[i].Height < headers[j].Height
	})

	for i := 0; i < len(headers); i++ {
		self.taskCh <- &persistHeaderTask{header: &headers[i]}
	}

	return nil

}

func (bd *ChainStore) GetHeader(hash Uint256) (*Header, error) {
	bd.mu.RLock()
	if header, ok := bd.headerCache[hash]; ok {
		bd.mu.RUnlock()
		return header, nil
	}
	bd.mu.RUnlock()

	var h *Header = new(Header)

	h.Program = new(program.Program)

	prefix := []byte{byte(DATA_Header)}
	data, err_get := bd.st.Get(append(prefix, hash.ToArray()...))
	if err_get != nil {
		return nil, err_get
	}

	r := bytes.NewReader(data)

	// first 8 bytes is sys_fee
	sysfee, err := serialization.ReadUint64(r)
	if err != nil {
		return nil, err
	}
	log.Debug(fmt.Sprintf("sysfee: %d\n", sysfee))

	// Deserialize block data
	err = h.Deserialize(r)
	if err != nil {
		return nil, err
	}

	return h, err
}

func (bd *ChainStore) GetAsset(hash Uint256) (*states.AssetState, error) {
	log.Debug(fmt.Sprintf("GetAsset Hash: %x\n", hash))

	asset := new(states.AssetState)

	prefix := []byte{byte(ST_Asset)}
	data, err_get := bd.st.Get(append(prefix, hash.ToArray()...))

	log.Debug(fmt.Sprintf("GetAsset Data: %x\n", data))
	if err_get != nil {
		//TODO: implement error process
		return nil, err_get
	}

	r := bytes.NewReader(data)
	asset.Deserialize(r)

	return asset, nil
}

func (bd *ChainStore) GetContract(hash Uint160) (*states.ContractState, error) {
	contract := new(states.ContractState)
	data, err := bd.st.Get(append([]byte{byte(ST_Contract)}, hash.ToArray()...))
	if err != nil {
		return nil, err
	}

	bf := bytes.NewBuffer(data)
	if err := contract.Deserialize(bf); err != nil {
		return nil, err
	}
	return contract, nil
}

func (bd *ChainStore) GetTransaction(hash Uint256) (*tx.Transaction, error) {
	log.Debugf("GetTransaction Hash: %x\n", hash)

	t := new(tx.Transaction)
	_, err := bd.getTx(t, hash)

	if err != nil {
		return nil, err
	}

	return t, nil
}

func (bd *ChainStore) GetTransactionWithHeight(hash Uint256) (*tx.Transaction, uint32, error) {
	t := new(tx.Transaction)
	height, err := bd.getTx(t, hash)
	if err != nil {
		return nil, 0, err
	}
	return t, height, nil
}

func (bd *ChainStore) getTx(tx *tx.Transaction, hash Uint256) (uint32, error) {
	prefix := []byte{byte(DATA_Transaction)}
	tHash, err_get := bd.st.Get(append(prefix, hash.ToArray()...))
	if err_get != nil {
		//TODO: implement error process
		return 0, err_get
	}

	r := bytes.NewReader(tHash)

	// get height
	height, err := serialization.ReadUint32(r)
	if err != nil {
		return 0, err
	}

	// Deserialize Transaction
	if err := tx.Deserialize(r); err != nil {
		log.Error("[getTx] error:", err)
		return 0, err
	}

	return height, err
}

func (bd *ChainStore) SaveTransaction(tx *tx.Transaction, height uint32) error {
	//////////////////////////////////////////////////////////////
	// generate key with DATA_Transaction prefix
	txhash := bytes.NewBuffer(nil)
	// add transaction header prefix.
	txhash.WriteByte(byte(DATA_Transaction))
	// get transaction hash
	txHashValue := tx.Hash()
	txHashValue.Serialize(txhash)
	log.Debug(fmt.Sprintf("transaction header + hash: %x\n", txhash))

	// generate value
	w := bytes.NewBuffer(nil)
	serialization.WriteUint32(w, height)
	tx.Serialize(w)
	log.Debug(fmt.Sprintf("transaction tx data: %x\n", w))

	// put value
	err := bd.st.BatchPut(txhash.Bytes(), w.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (bd *ChainStore) GetBlock(hash Uint256) (*Block, error) {
	bd.mu.RLock()
	if block, ok := bd.blockCache[hash]; ok {
		bd.mu.RUnlock()
		return block, nil
	}
	bd.mu.RUnlock()

	var b *Block = new(Block)

	b.Header = new(Header)
	b.Header.Program = new(program.Program)

	prefix := []byte{byte(DATA_Header)}
	bHash, err_get := bd.st.Get(append(prefix, hash.ToArray()...))
	if err_get != nil {
		//TODO: implement error process
		return nil, err_get
	}

	r := bytes.NewReader(bHash)

	// first 8 bytes is sys_fee
	_, err := serialization.ReadUint64(r)
	if err != nil {
		return nil, err
	}

	// Deserialize block data
	err = b.FromTrimmedData(r)
	if err != nil {
		return nil, err
	}

	// Deserialize transaction
	for i := 0; i < len(b.Transactions); i++ {
		_, err = bd.getTx(b.Transactions[i], b.Transactions[i].Hash())
		log.Debugf("tx hash =%x\n", b.Transactions[i].Hash())
		if err != nil {
			log.Debugf("NRF tx hash =%x , error=%s\n", b.Transactions[i].Hash(), err)
			return nil, err
		}
	}

	return b, nil
}

func (self *ChainStore) GetBookKeeperList() ([]*crypto.PubKey, []*crypto.PubKey, error) {
	val, err := self.st.Get(append([]byte{byte(ST_BookKeeper)}, BookerKeeper...))
	if err != nil {
		return nil, nil, err
	}
	bookKeeper := new(states.BookKeeperState)
	bf := bytes.NewBuffer(val)
	err = bookKeeper.Deserialize(bf)
	if err != nil {
		return nil, nil, err
	}
	return bookKeeper.CurrBookKeeper, bookKeeper.NextBookKeeper, nil
}

func (bd *ChainStore) persist(b *Block) error {
	bd.st.NewBatch()
	stateStore := NewStateStore(statestore.NewMemDatabase(), bd, statestore.NewTrieStore(bd.st), b.Header.StateRoot)
	state, err := stateStore.TryGet(ST_BookKeeper, BookerKeeper)
	if err != nil {
		log.Error("[persist] TryGet ST_BookKeeper error:", err)
		return err
	}
	bookKeeper := state.Value.(*states.BookKeeperState)
	handleBookKeeper(stateStore, bookKeeper)
	for _, t := range b.Transactions {
		bd.SaveTransaction(t, b.Header.Height)
		tx_id := t.Hash()
		if len(t.Outputs) > 0 {
			stateStore.TryAdd(ST_Coin, tx_id.ToArray(), &states.UnspentCoinState{Item: repeat(len(t.Outputs))}, false)
		}
		if err := handleOutputs(t.Hash(), t.Outputs, stateStore); err != nil {
			return err
		}
 		if err := handleInputs(t.UTXOInputs, stateStore, b.Header.Height, bd); err != nil {
			return err
		}
		switch t.TxType {
		case tx.RegisterAsset:
			p := t.Payload.(*payload.RegisterAsset)
			if err := stateStore.TryGetOrAdd(ST_Asset, tx_id.ToArray(), &states.AssetState{
				AssetId:    tx_id,
				AssetType:  p.Asset.AssetType,
				Name:       p.Asset.Name,
				Amount:     p.Amount,
				Available:  p.Amount,
				Precision:  p.Asset.Precision,
				Owner:      p.Issuer,
				Admin:      p.Controller,
				Issuer:     p.Controller,
				Expiration: b.Header.Height + 2*2000000,
				IsFrozen:   false,
			}, false); err != nil {
				log.Error("[persist] TryAdd ST_Asset error:", err)
				return err
			}
		case tx.IssueAsset:
			results := t.GetMergedAssetIDValueFromOutputs()
			for k, r := range results {
				state, err := stateStore.TryGetAndChange(ST_Asset, k.ToArray(), false)
				if err != nil {
					log.Errorf("[persist] TryGet ST_Asset error:", err)
					return err
				}
				asset := state.(*states.AssetState)
				asset.Available -= r
			}
		case tx.Claim:
			p := t.Payload.(*payload.Claim)
			for _, c := range p.Claims {
				state, err := stateStore.TryGetAndChange(ST_SpentCoin, c.ReferTxID.ToArray(), false)
				if err != nil {
					log.Errorf("[persist] TryGet ST_SpentCoin error:", err)
					return err
				}
				spentcoins := state.(*states.SpentCoinState)
				spentcoins.Items = remove(spentcoins.Items, int(c.ReferTxOutputIndex))
			}
		case tx.BookKeeper:
			bk := t.Payload.(*payload.BookKeeper)
			switch bk.Action {
			case payload.BookKeeperAction_ADD:
				if crypto.ContainPubKey(bk.PubKey, bookKeeper.NextBookKeeper) < 0 {
					bookKeeper.NextBookKeeper = append(bookKeeper.NextBookKeeper, bk.PubKey)
					sort.Sort(crypto.PubKeySlice(bookKeeper.NextBookKeeper))
				}
				stateStore.memoryStore.Change(byte(ST_BookKeeper), BookerKeeper, false)
			case payload.BookKeeperAction_SUB:
				index := crypto.ContainPubKey(bk.PubKey, bookKeeper.NextBookKeeper)
				if index >= 0 {
					bookKeeper.NextBookKeeper = append(bookKeeper.NextBookKeeper[:index], bookKeeper.NextBookKeeper[index+1:]...)
				}
				stateStore.memoryStore.Change(byte(ST_BookKeeper), BookerKeeper, false)
			}
		case tx.Deploy:
			deploy := t.Payload.(*payload.DeployCode)
			codeHash := deploy.Code.CodeHash()
			if err := stateStore.TryGetOrAdd(ST_Contract, codeHash.ToArray(), &states.ContractState{
				Code:        deploy.Code,
				VmType:      deploy.VmType,
				NeedStorage: deploy.NeedStorage,
				Name:        deploy.Name,
				Version:     deploy.CodeVersion,
				Author:      deploy.Author,
				Email:       deploy.Email,
				Description: deploy.Description,
			}, false); err != nil {
				log.Error("[persist] TryAdd ST_Contract error:", err)
				return err
			}
		case tx.Invoke:
			invoke := t.Payload.(*payload.InvokeCode)
			cs, err := stateStore.TryGet(ST_Contract, invoke.CodeHash.ToArray())
			if err != nil {
				log.Error("[persist] TryGet ST_Contract error:", err)
				return err
			}
			if cs == nil {
				event.PushSmartCodeEvent(t.Hash(), 0, INVOKE_TRANSACTION, "Contract not found!")
				continue
			}
			contract := cs.Value.(*states.ContractState)
			stateMachine := service.NewStateMachine(stateStore, types.Application, b)
			smc, err := sc.NewSmartContract(&sc.Context{
				VmType:         contract.VmType,
				StateMachine:   stateMachine,
				SignableData:   t,
				CacheCodeTable: &CacheCodeTable{stateStore},
				Input:          invoke.Code,
				Code:           contract.Code.Code,
				ReturnType:     contract.Code.ReturnType,
			})
			if err != nil {
				log.Error("[persist] NewSmartContract error:", err)
				return err
			}
			ret, err := smc.InvokeContract()
			if err != nil {
				log.Error("[persist] InvokeContract error:", err)
				event.PushSmartCodeEvent(t.Hash(), httprestful.SMARTCODE_ERROR, INVOKE_TRANSACTION, err)
				continue
			}
			log.Error("result:", ret)
			stateMachine.CloneCache.Commit()
			event.PushSmartCodeEvent(t.Hash(), 0, INVOKE_TRANSACTION, ret)
		}
	}
	if err := stateStore.CommitTo(); err != nil {
		return err
	}
	stateRoot, err := stateStore.trie.CommitTo()
	if err != nil {
		return nil
	}
	if err := addCurrentStateRoot(bd, stateRoot); err != nil {
		return nil
	}
	if err := addSysCurrentBlock(bd, b); err != nil {
		return err
	}
	if err := addHeader(bd, b, 0); err != nil {
		return err
	}
	if err := addDataBlock(bd, b); err != nil {
		return err
	}

	addMerkleRoot(bd, b)

	err = bd.st.BatchCommit()
	if err != nil {
		return err
	}
	return nil
}

// can only be invoked by backend write goroutine
func (bd *ChainStore) addHeader(header *Header) {

	log.Debugf("addHeader(), Height=%d\n", header.Height)

	hash := header.Hash()

	bd.mu.Lock()
	bd.headerCache[header.Hash()] = header
	bd.headerIndex[header.Height] = hash
	bd.mu.Unlock()

	log.Debug("[addHeader]: finish, header height:", header.Height)
}

func (self *ChainStore) handlePersistHeaderTask(header *Header) {

	if header.Height != uint32(len(self.headerIndex)) {
		return
	}

	if !self.verifyHeader(header) {
		return
	}

	self.addHeader(header)
}

func (self *ChainStore) SaveBlock(b *Block, ledger *Ledger) error {
	log.Debug("SaveBlock()")

	self.mu.RLock()
	headerHeight := uint32(len(self.headerIndex))
	currBlockHeight := self.currentBlockHeight
	self.mu.RUnlock()

	if b.Header.Height <= currBlockHeight {
		return nil
	}

	if b.Header.Height > headerHeight {
		return errors.New(fmt.Sprintf("Info: [SaveBlock] block height - headerIndex.count >= 1, block height:%d, headerIndex.count:%d",
			b.Header.Height, headerHeight))
	}

	if b.Header.Height == headerHeight {
		err := validation.VerifyBlock(b, ledger, false)
		if err != nil {
			log.Error("VerifyBlock error!")
			return err
		}

		self.taskCh <- &persistHeaderTask{header: b.Header}
	} else {
		err := validation.VerifySignableDataSignature(b)
		if err != nil {
			log.Error("VerifyBlock Signature error!")
			return err
		}
		err = validation.VerifySignableDataProgramHashes(b)
		if err != nil {
			log.Error("VerifyBlock ProgramHashes error!")
			return err
		}
	}

	self.taskCh <- &persistBlockTask{block: b, ledger: ledger}
	return nil
}

func (self *ChainStore) handlePersistBlockTask(b *Block, ledger *Ledger) {
	if b.Header.Height <= self.currentBlockHeight {
		return
	}

	self.mu.Lock()
	self.blockCache[b.Hash()] = b
	self.mu.Unlock()

	if b.Header.Height < uint32(len(self.headerIndex)) {
		self.persistBlocks(ledger)

		self.st.NewBatch()
		storedHeaderCount := self.storedHeaderCount
		for self.currentBlockHeight-storedHeaderCount >= HeaderHashListCount {
			hashBuffer := new(bytes.Buffer)
			serialization.WriteVarUint(hashBuffer, uint64(HeaderHashListCount))
			var hashArray []byte
			for i := 0; i < HeaderHashListCount; i++ {
				index := storedHeaderCount + uint32(i)
				thash := self.headerIndex[index]
				thehash := thash.ToArray()
				hashArray = append(hashArray, thehash...)
			}
			hashBuffer.Write(hashArray)

			hhlPrefix := bytes.NewBuffer(nil)
			hhlPrefix.WriteByte(byte(IX_HeaderHashList))
			serialization.WriteUint32(hhlPrefix, storedHeaderCount)

			self.st.BatchPut(hhlPrefix.Bytes(), hashBuffer.Bytes())
			storedHeaderCount += HeaderHashListCount
		}

		err := self.st.BatchCommit()
		if err != nil {
			log.Error("failed to persist header hash list:", err)
			return
		}
		self.mu.Lock()
		self.storedHeaderCount = storedHeaderCount
		self.mu.Unlock()

		self.clearCache()
	}
}

func (bd *ChainStore) persistBlocks(ledger *Ledger) {
	stopHeight := uint32(len(bd.headerIndex))
	for h := bd.currentBlockHeight + 1; h <= stopHeight; h++ {
		hash := bd.headerIndex[h]
		block, ok := bd.blockCache[hash]
		if !ok {
			break
		}
		err := bd.persist(block)
		if err != nil {
			log.Fatal("[persistBlocks]: error to persist block:", err.Error())
			return
		}

		// PersistCompleted event
		ledger.Blockchain.BlockHeight = block.Header.Height
		bd.mu.Lock()
		bd.currentBlockHeight = block.Header.Height
		bd.mu.Unlock()

		ledger.Blockchain.BCEvents.Notify(events.EventBlockPersistCompleted, block)
		log.Tracef("The latest block height:%d, block hash: %x", block.Header.Height, hash)
	}

}

func (bd *ChainStore) BlockInCache(hash Uint256) bool {
	bd.mu.RLock()
	defer bd.mu.RUnlock()

	_, ok := bd.blockCache[hash]
	return ok
}

func (bd *ChainStore) GetQuantityIssued(assetId Uint256) (Fixed64, error) {
	prefix := []byte{byte(ST_Asset)}
	data, err_get := bd.st.Get(append(prefix, assetId.ToArray()...))

	asset := new(states.AssetState)
	if err_get != nil {
		return Fixed64(0), err_get
	}

	r := bytes.NewReader(data)
	if err := asset.Deserialize(r); err != nil {
		return Fixed64(0), err
	}
	return asset.Amount - asset.Available, nil
}

func (bd *ChainStore) GetUnspent(txid Uint256, index uint16) (*utxo.TxOutput, error) {
	if ok, _ := bd.ContainsUnspent(txid, index); ok {
		Tx, err := bd.GetTransaction(txid)
		if err != nil {
			return nil, err
		}

		return Tx.Outputs[index], nil
	}

	return nil, errors.New("[GetUnspent] NOT ContainsUnspent.")
}

func (bd *ChainStore) ContainsUnspent(txid Uint256, index uint16) (bool, error) {
	unspentPrefix := []byte{byte(ST_Coin)}
	unspentValue, err_get := bd.st.Get(append(unspentPrefix, txid.ToArray()...))

	if err_get != nil {
		return false, err_get
	}

	unpsentcoin := new(states.UnspentCoinState)
	bf := bytes.NewBuffer(unspentValue)

	if err := unpsentcoin.Deserialize(bf); err != nil {
		return false, err
	}

	if index >= uint16(len(unpsentcoin.Item)) || unpsentcoin.Item[index] == states.Spent {
		return false, nil
	}

	return true, nil
}

func (bd *ChainStore) GetCurrentHeaderHash() Uint256 {
	bd.mu.RLock()
	defer bd.mu.RUnlock()
	return bd.headerIndex[uint32(len(bd.headerIndex)-1)]
}

func (bd *ChainStore) GetHeaderHashByHeight(height uint32) Uint256 {
	bd.mu.RLock()
	defer bd.mu.RUnlock()

	return bd.headerIndex[height]
}

func (bd *ChainStore) GetHeaderHeight() uint32 {
	bd.mu.RLock()
	defer bd.mu.RUnlock()

	return uint32(len(bd.headerIndex) - 1)
}

func (bd *ChainStore) GetHeight() uint32 {
	bd.mu.RLock()
	defer bd.mu.RUnlock()

	return bd.currentBlockHeight
}

func (bd *ChainStore) GetAccount(programHash Uint160) (*states.AccountState, error) {
	accountPrefix := []byte{byte(ST_Account)}

	state, err := bd.st.Get(append(accountPrefix, programHash.ToArray()...))

	if err != nil {
		return nil, err
	}

	accountState := new(states.AccountState)
	accountState.Deserialize(bytes.NewBuffer(state))

	return accountState, nil
}

func (bd *ChainStore) IsBlockInStore(hash Uint256) bool {

	b := new(Block)

	b.Header = new(Header)
	b.Header.Program = new(program.Program)

	prefix := []byte{byte(DATA_Header)}
	blockData, err_get := bd.st.Get(append(prefix, hash.ToArray()...))
	if err_get != nil {
		return false
	}

	r := bytes.NewReader(blockData)

	// first 8 bytes is sys_fee
	_, err := serialization.ReadUint64(r)
	if err != nil {
		return false
	}

	// Deserialize block data
	err = b.FromTrimmedData(r)
	if err != nil {
		return false
	}

	if b.Header.Height > bd.currentBlockHeight {
		return false
	}

	return true
}

func (bd *ChainStore) GetUnspentFromProgramHash(programHash Uint160, assetid Uint256) ([]*utxo.UTXOUnspent, error) {
	prefix := []byte{byte(ST_Program_Coin)}

	key := append(prefix, programHash.ToArray()...)
	key = append(key, assetid.ToArray()...)
	data, err := bd.st.Get(key)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(data)
	programCoin := new(states.ProgramUnspentCoin)
	if err := programCoin.Deserialize(r); err != nil {
		return nil, err
	}
	return programCoin.Unspents, nil
}

func (bd *ChainStore) GetAssets() map[Uint256]*states.AssetState {
	assets := make(map[Uint256]*states.AssetState)

	iter := bd.st.NewIterator([]byte{byte(ST_Asset)})
	for iter.Next() {
		rk := bytes.NewReader(iter.Key())

		// read prefix
		_, _ = serialization.ReadBytes(rk, 1)
		var assetid Uint256
		assetid.Deserialize(rk)
		log.Tracef("[GetAssets] assetid: %x\n", assetid.ToArray())

		asset := new(states.AssetState)
		r := bytes.NewReader(iter.Value())
		asset.Deserialize(r)
		assets[assetid] = asset
	}

	return assets
}

func (bd *ChainStore) GetUnclaimed(hash Uint256) (map[uint16]*utxo.SpentCoin, error) {
	transaction_ := new(tx.Transaction)
	_, err := bd.getTx(transaction_, hash)
	if err != nil {
		return nil, err
	}
	claimable := make(map[uint16]*utxo.SpentCoin)
	prefix := []byte{byte(ST_SpentCoin)}
	key := append(prefix, hash.ToArray()...)
	claimabledata, err := bd.st.Get(key)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(claimabledata)
	SpentCoinState_ := new(states.SpentCoinState)
	err = SpentCoinState_.Deserialize(r)
	if err != nil {
		return nil, err
	}
	for _, v := range SpentCoinState_.Items {
		claimable[v.PrevIndex] = &utxo.SpentCoin{
			Output:      transaction_.Outputs[v.PrevIndex],
			StartHeight: SpentCoinState_.TransactionHeight,
			EndHeight:   v.EndHeight,
		}
	}
	return claimable, nil
}

func (bd *ChainStore) GetCurrentStateRoot() Uint256 {
	u256 := new(Uint256)
	data, err := bd.st.Get(append([]byte{byte(Sys_CurrentStateRoot)}, CurrentStateRoot...))
	if err != nil {
		return Uint256{}
	}
	b := bytes.NewBuffer(data)
	err = u256.Deserialize(b)
	if err != nil {
		log.Errorf("[GetCurrentStateRoot] deserialize: %v", err)
		return Uint256{}
	}
	return *u256
}

func (bd *ChainStore) GetIdentity(ontId []byte) ([]byte, error) {
	idPrefix := []byte{byte(ST_Identity)}
	idKey := append(idPrefix, ontId...)
	return bd.st.Get(idKey)
}

func (bd *ChainStore) SetIdentity(ontId, ddo []byte) error {
	idPrefix := []byte{byte(ST_Identity)}
	idKey := append(idPrefix, ontId...)
	return bd.st.BatchPut(idKey, ddo)
}

func (bd *ChainStore) GetStorageItem(key *states.StorageKey) (*states.StorageItem, error) {
	v, err := bd.st.Get(append(append([]byte{byte(ST_Storage)}, key.ToArray()...)))
	if err != nil {
		return nil, err
	}
	item := new(states.StorageItem)
	if err := item.Deserialize(bytes.NewBuffer(v)); err != nil {
		return nil, err
	}
	return item, nil
}

func (bd *ChainStore) GetSysFeeAmount(hash Uint256) (Fixed64, error) {
	amount := new(Fixed64)
	data, err := bd.st.Get(append([]byte{byte(DATA_Header)}, hash.ToArray()...))
	if err != nil {
		return Fixed64(0), nil 
	}
	b := bytes.NewReader(data)
	err = amount.Deserialize(b)
	if err != nil {
		return Fixed64(0), nil
	}
	return *amount, nil
}

func (bd *ChainStore) GetVoteStates() (map[Uint160]*states.VoteState, error) {
	votes := make(map[Uint160]*states.VoteState)
	iter := bd.st.NewIterator([]byte{byte(ST_Vote)})
	for iter.Next() {
		rk := bytes.NewReader(iter.Key())

		// read prefix
		_, _ = serialization.ReadBytes(rk, 1)
		var programHash Uint160
		if err := programHash.Deserialize(rk); err != nil {
			return nil, err
		}

		vote := new(states.VoteState)
		r := bytes.NewReader(iter.Value())
		if err := vote.Deserialize(r); err != nil {
			return nil, err
		}
		votes[programHash] = vote
	}
	return votes, nil
}

func (bd *ChainStore) GetVotesAndEnrollments(txs []*tx.Transaction) ([]*states.VoteState, []*crypto.PubKey, error) {
	var votes []*states.VoteState
	result, votesBlock, enrollsBlock, err := bd.getBlockTransactionResult(txs)
	if err != nil {
		return nil, nil, err
	}
	voteStates, err := bd.GetVoteStates()
	if err != nil {
		return nil, nil, err
	}
	for k, v := range votesBlock {
		voteStates[k] = v
	}

	for k, v := range voteStates {
		account, err := bd.GetAccount(k)
		if err != nil {
			return nil, nil, err
		}
		v.Count = account.Balances[tx.ONTTokenID]
		if s, ok := result[k]; ok {
			v.Count += s
		}
		if v.Count <= 0 || v.Count < Fixed64(len(v.PublicKeys)) {
			continue
		}
		votes = append(votes, v)
	}
	enrolls, err := bd.getEnrollments()
	if err != nil {
		return nil, nil, err
	}

	enrolls = append(enrolls, enrollsBlock...)
	return votes, enrolls, nil
}

func (bd *ChainStore) getBlockTransactionResult(txs []*tx.Transaction) (map[Uint160]Fixed64,
map[Uint160]*states.VoteState, []*crypto.PubKey, error) {
	r := make(map[Uint160]Fixed64)
	votes := make(map[Uint160]*states.VoteState)
	var enrolls []*crypto.PubKey
	for _, t := range txs {
		for _, i := range t.UTXOInputs {
			if i.ReferTxID.CompareTo(tx.ONTTokenID) != 0 {
				continue
			}
			tran, err := bd.GetTransaction(i.ReferTxID)
			if err != nil {
				return nil, nil, nil, err
			}
			output := tran.Outputs[i.ReferTxOutputIndex]
			r[output.ProgramHash] -= output.Value
		}
		for _, o := range t.Outputs {
			if o.AssetID.CompareTo(tx.ONTTokenID) != 0 {
				continue
			}
			r[o.ProgramHash] += o.Value
		}

		if t.TxType == tx.Vote {
			vote := t.Payload.(*payload.Vote)
			votes[vote.Account] = &states.VoteState{PublicKeys: vote.PubKeys}
		} else if t.TxType == tx.Enrollment {
			enroll := t.Payload.(*payload.Enrollment)
			enrolls = append(enrolls, enroll.PublicKey)
		}
	}
	return r, votes, enrolls, nil
}

func (bd *ChainStore) getEnrollments() ([]*crypto.PubKey, error) {
	var validators []*crypto.PubKey
	iter := bd.st.NewIterator([]byte{byte(ST_Validator)})
	for iter.Next() {
		validator := new(states.ValidatorState)
		r := bytes.NewReader(iter.Value())
		if err := validator.Deserialize(r); err != nil {
			return nil, err
		}
		validators = append(validators, validator.PublicKey)
	}
	return append(StandbyBookKeepers, validators...), nil
}

