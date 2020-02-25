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
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/states"
	scom "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/leveldbstore"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/ontio/ontology/merkle"
	"github.com/ontio/ontology/smartcontract/service/native/ontid"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

var (
	BOOKKEEPER = []byte("Bookkeeper") //Bookkeeper store key
)

//StateStore saving the data of ledger states. Like balance of account, and the execution result of smart contract
type StateStore struct {
	dbDir                string                    //Store file path
	store                scom.PersistStore         //Store handler
	merklePath           string                    //Merkle tree store path
	merkleTree           *merkle.CompactMerkleTree //Merkle tree of block root
	deltaMerkleTree      *merkle.CompactMerkleTree //Merkle tree of delta state root
	merkleHashStore      merkle.HashStore
	stateHashCheckHeight uint32
}

//NewStateStore return state store instance
func NewStateStore(dbDir, merklePath string, stateHashCheckHeight uint32) (*StateStore, error) {
	var err error
	store, err := leveldbstore.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	stateStore := &StateStore{
		dbDir:                dbDir,
		store:                store,
		merklePath:           merklePath,
		stateHashCheckHeight: stateHashCheckHeight,
	}
	_, height, err := stateStore.GetCurrentBlock()
	if err != nil && err != scom.ErrNotFound {
		return nil, fmt.Errorf("GetCurrentBlock error %s", err)
	}
	err = stateStore.init(height)
	if err != nil {
		return nil, fmt.Errorf("init error %s", err)
	}
	return stateStore, nil
}

// for test
func NewMemStateStore(stateHashHeight uint32) *StateStore {
	store, _ := leveldbstore.NewMemLevelDBStore()
	stateStore := &StateStore{
		store:                store,
		merkleTree:           merkle.NewTree(0, nil, nil),
		deltaMerkleTree:      merkle.NewTree(0, nil, nil),
		stateHashCheckHeight: stateHashHeight,
	}

	return stateStore
}

//NewBatch start new commit batch
func (self *StateStore) NewBatch() {
	self.store.NewBatch()
}

func (self *StateStore) BatchPutRawKeyVal(key, val []byte) {
	self.store.BatchPut(key, val)
}

func (self *StateStore) BatchDeleteRawKey(key []byte) {
	self.store.BatchDelete(key)
}

func (self *StateStore) init(currBlockHeight uint32) error {
	treeSize, hashes, err := self.GetBlockMerkleTree()
	if err != nil && err != scom.ErrNotFound {
		return err
	}
	if treeSize > 0 && treeSize != currBlockHeight+1 {
		return fmt.Errorf("merkle tree size is inconsistent with blockheight: %d", currBlockHeight+1)
	}
	self.merkleHashStore, err = merkle.NewFileHashStore(self.merklePath, treeSize)
	if err != nil {
		log.Warn("merkle store is inconsistent with ChainStore. persistence will be disabled")
	}
	self.merkleTree = merkle.NewTree(treeSize, hashes, self.merkleHashStore)

	if currBlockHeight >= self.stateHashCheckHeight {
		treeSize, hashes, err := self.GetStateMerkleTree()
		if err != nil && err != scom.ErrNotFound {
			return err
		}
		if treeSize > 0 && treeSize != currBlockHeight-self.stateHashCheckHeight+1 {
			return fmt.Errorf("merkle tree size is inconsistent with blockheight: %d", currBlockHeight+1)
		}
		self.deltaMerkleTree = merkle.NewTree(treeSize, hashes, nil)
	}
	return nil
}

//GetStateMerkleTree return merkle tree size an tree node
func (self *StateStore) GetStateMerkleTree() (uint32, []common.Uint256, error) {
	key := self.genStateMerkleTreeKey()
	return self.getMerkleTree(key)
}

//GetBlockMerkleTree return merkle tree size an tree node
func (self *StateStore) GetBlockMerkleTree() (uint32, []common.Uint256, error) {
	key := self.genBlockMerkleTreeKey()
	return self.getMerkleTree(key)
}
func (self *StateStore) getMerkleTree(key []byte) (uint32, []common.Uint256, error) {
	data, err := self.store.Get(key)
	if err != nil {
		return 0, nil, err
	}
	value := bytes.NewBuffer(data)
	treeSize, err := serialization.ReadUint32(value)
	if err != nil {
		return 0, nil, err
	}
	hashCount := (len(data) - 4) / common.UINT256_SIZE
	hashes := make([]common.Uint256, 0, hashCount)
	for i := 0; i < hashCount; i++ {
		var hash = new(common.Uint256)
		err = hash.Deserialize(value)
		if err != nil {
			return 0, nil, err
		}
		hashes = append(hashes, *hash)
	}
	return treeSize, hashes, nil
}

func (self *StateStore) GetStateMerkleRoot(height uint32) (result common.Uint256, err error) {
	if height < self.stateHashCheckHeight {
		return
	}
	key := self.genStateMerkleRootKey(height)
	var value []byte
	value, err = self.store.Get(key)
	if err != nil {
		return
	}
	source := common.NewZeroCopySource(value)
	_, eof := source.NextHash()
	result, eof = source.NextHash()
	if eof {
		err = io.ErrUnexpectedEOF
	}
	return
}

func (self *StateStore) AddStateMerkleTreeRoot(blockHeight uint32, writeSetHash common.Uint256) error {
	if blockHeight < self.stateHashCheckHeight {
		return nil
	} else if blockHeight == self.stateHashCheckHeight {
		self.deltaMerkleTree = merkle.NewTree(0, nil, nil)
	}
	key := self.genStateMerkleTreeKey()

	self.deltaMerkleTree.AppendHash(writeSetHash)
	treeSize := self.deltaMerkleTree.TreeSize()
	hashes := self.deltaMerkleTree.Hashes()
	value := common.NewZeroCopySink(make([]byte, 0, 4+len(hashes)*common.UINT256_SIZE))
	value.WriteUint32(treeSize)
	for _, hash := range hashes {
		value.WriteHash(hash)
	}
	self.store.BatchPut(key, value.Bytes())

	key = self.genStateMerkleRootKey(blockHeight)
	value.Reset()
	value.WriteHash(writeSetHash)
	value.WriteHash(self.deltaMerkleTree.Root())
	self.store.BatchPut(key, value.Bytes())

	return nil
}

//AddBlockMerkleTreeRoot add a new tree root
func (self *StateStore) AddBlockMerkleTreeRoot(txRoot common.Uint256) error {
	key := self.genBlockMerkleTreeKey()

	self.merkleTree.AppendHash(txRoot)
	treeSize := self.merkleTree.TreeSize()
	hashes := self.merkleTree.Hashes()
	value := common.NewZeroCopySink(make([]byte, 0, 4+len(hashes)*common.UINT256_SIZE))
	value.WriteUint32(treeSize)
	for _, hash := range hashes {
		value.WriteHash(hash)
	}
	self.store.BatchPut(key, value.Bytes())
	return nil
}

//GetMerkleProof return merkle proof of block
func (self *StateStore) GetMerkleProof(proofHeight, rootHeight uint32) ([]common.Uint256, error) {
	return self.merkleTree.InclusionProof(proofHeight, rootHeight+1)
}

func (self *StateStore) NewOverlayDB() *overlaydb.OverlayDB {
	return overlaydb.NewOverlayDB(self.store)
}

//CommitTo commit state batch to state store
func (self *StateStore) CommitTo() error {
	return self.store.BatchCommit()
}

//GetContractState return contract by contract address
func (self *StateStore) GetContractState(contractHash common.Address) (*payload.DeployCode, error) {
	key, err := self.getContractStateKey(contractHash)
	if err != nil {
		return nil, err
	}

	value, err := self.store.Get(key)
	if err != nil {
		return nil, err
	}
	source := common.NewZeroCopySource(value)
	contractState := new(payload.DeployCode)
	err = contractState.Deserialization(source)
	if err != nil {
		return nil, err
	}
	return contractState, nil
}

//GetBookkeeperState return current book keeper states
func (self *StateStore) GetBookkeeperState() (*states.BookkeeperState, error) {
	key, err := self.getBookkeeperKey()
	if err != nil {
		return nil, err
	}

	value, err := self.store.Get(key)
	if err != nil {
		return nil, err
	}
	reader := common.NewZeroCopySource(value)
	bookkeeperState := new(states.BookkeeperState)
	err = bookkeeperState.Deserialization(reader)
	if err != nil {
		return nil, err
	}
	return bookkeeperState, nil
}

//SaveBookkeeperState persist book keeper state to store
func (self *StateStore) SaveBookkeeperState(bookkeeperState *states.BookkeeperState) error {
	key, err := self.getBookkeeperKey()
	if err != nil {
		return err
	}
	value := common.NewZeroCopySink(nil)
	bookkeeperState.Serialization(value)

	return self.store.Put(key, value.Bytes())
}

//GetStorageItem return the storage value of the key in smart contract.
func (self *StateStore) GetStorageState(key *states.StorageKey) (*states.StorageItem, error) {
	storeKey, err := self.getStorageKey(key)
	if err != nil {
		return nil, err
	}

	data, err := self.store.Get(storeKey)
	if err != nil {
		return nil, err
	}
	reader := common.NewZeroCopySource(data)
	storageState := new(states.StorageItem)
	err = storageState.Deserialization(reader)
	if err != nil {
		return nil, err
	}
	return storageState, nil
}

//GetCurrentBlock return current block height and current hash in state store
func (self *StateStore) GetCurrentBlock() (common.Uint256, uint32, error) {
	key := self.getCurrentBlockKey()
	data, err := self.store.Get(key)
	if err != nil {
		return common.Uint256{}, 0, err
	}
	reader := bytes.NewReader(data)
	blockHash := common.Uint256{}
	err = blockHash.Deserialize(reader)
	if err != nil {
		return common.Uint256{}, 0, err
	}
	height, err := serialization.ReadUint32(reader)
	if err != nil {
		return common.Uint256{}, 0, err
	}
	return blockHash, height, nil
}

//SaveCurrentBlock persist current block to state store
func (self *StateStore) SaveCurrentBlock(height uint32, blockHash common.Uint256) error {
	key := self.getCurrentBlockKey()
	value := bytes.NewBuffer(nil)
	blockHash.Serialize(value)
	serialization.WriteUint32(value, height)
	self.store.BatchPut(key, value.Bytes())
	return nil
}

func (self *StateStore) GetCrossStates(height uint32) ([]common.Uint256, error) {
	key := self.genCrossStatesKey(height)
	data, err := self.store.Get(key)
	if err != nil {
		return []common.Uint256{}, err
	}
	source := common.NewZeroCopySource(data)
	l := len(data) / common.UINT256_SIZE
	hashes := make([]common.Uint256, 0, l)
	for i := 0; i < l; i++ {
		u256, eof := source.NextHash()
		if eof {
			return []common.Uint256{}, fmt.Errorf("%s", "Get states hash error!")
		}
		hashes = append(hashes, u256)
	}
	return hashes, nil
}

func (self *StateStore) GetCrossStatesRoot(height uint32) (common.Uint256, error) {
	states, err := self.GetCrossStates(height)
	if err != nil && err != scom.ErrNotFound {
		return common.UINT256_EMPTY, err
	}
	if err == scom.ErrNotFound {
		return common.UINT256_EMPTY, nil
	}
	return merkle.TreeHasher{}.HashFullTreeWithLeafHash(states), nil
}

func (self *StateStore) SaveCrossStates(height uint32, crossStates []common.Uint256) error {
	//save cross states hash
	if len(crossStates) == 0 {
		return nil
	}
	key := self.genCrossStatesKey(height)
	sink := common.NewZeroCopySink(make([]byte, 0, len(crossStates)*common.UINT256_SIZE))
	for _, v := range crossStates {
		sink.WriteHash(v)
	}
	self.store.BatchPut(key, sink.Bytes())
	return nil
}

func (self *StateStore) genCrossStatesKey(height uint32) []byte {
	key := make([]byte, 5)
	key[0] = byte(scom.SYS_CURRENT_CROSS_STATES)
	binary.LittleEndian.PutUint32(key[1:], height)
	return key
}

func (self *StateStore) getCurrentBlockKey() []byte {
	return []byte{byte(scom.SYS_CURRENT_BLOCK)}
}

func (self *StateStore) getBookkeeperKey() ([]byte, error) {
	key := make([]byte, 1+len(BOOKKEEPER))
	key[0] = byte(scom.ST_BOOKKEEPER)
	copy(key[1:], BOOKKEEPER)
	return key, nil
}

func (self *StateStore) getContractStateKey(contractHash common.Address) ([]byte, error) {
	data := contractHash[:]
	key := make([]byte, 1+len(data))
	key[0] = byte(scom.ST_CONTRACT)
	copy(key[1:], data)
	return key, nil
}

func (self *StateStore) getStorageKey(key *states.StorageKey) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(byte(scom.ST_STORAGE))
	buf.Write(key.ContractAddress[:])
	buf.Write(key.Key)
	return buf.Bytes(), nil
}

func (self *StateStore) GetStateMerkleRootWithNewHash(writeSetHash common.Uint256) common.Uint256 {
	return self.deltaMerkleTree.GetRootWithNewLeaf(writeSetHash)
}

func (self *StateStore) GetBlockRootWithNewTxRoots(txRoots []common.Uint256) common.Uint256 {
	return self.merkleTree.GetRootWithNewLeaves(txRoots)
}

func (self *StateStore) genBlockMerkleTreeKey() []byte {
	return []byte{byte(scom.SYS_BLOCK_MERKLE_TREE)}
}

func (self *StateStore) genStateMerkleTreeKey() []byte {
	return []byte{byte(scom.SYS_STATE_MERKLE_TREE)}
}

func (self *StateStore) genStateMerkleRootKey(height uint32) []byte {
	key := make([]byte, 5, 5)
	key[0] = byte(scom.DATA_STATE_MERKLE_ROOT)
	binary.LittleEndian.PutUint32(key[1:], height)
	return key
}

//ClearAll clear all data in state store
func (self *StateStore) ClearAll() error {
	self.store.NewBatch()
	iter := self.store.NewIterator(nil)
	for iter.Next() {
		self.store.BatchDelete(iter.Key())
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		self.store.NewBatch() // reset the batch
		return err
	}
	return self.store.BatchCommit()
}

//Close state store
func (self *StateStore) Close() error {
	self.merkleHashStore.Close()
	return self.store.Close()
}

func (self *StateStore) CheckStorage() error {
	db := self.store

	prefix := append([]byte{byte(scom.ST_STORAGE)}, utils.OntIDContractAddress[:]...) //prefix of new storage key
	flag := append(prefix, ontid.FIELD_VERSION)
	val, err := db.Get(flag)
	if err == nil {
		item := &states.StorageItem{}
		source := common.NewZeroCopySource(val)
		err := item.Deserialization(source)
		if err == nil && item.Value[0] == ontid.FLAG_VERSION {
			return nil
		} else if err == nil {
			return errors.New("check ontid storage: invalid version flag")
		} else {
			return err
		}
	}

	prefix1 := []byte{byte(scom.ST_STORAGE), 0x2a, 0x64, 0x69, 0x64} //prefix of old storage key

	iter := db.NewIterator(prefix1)
	db.NewBatch()
	for ok := iter.First(); ok; ok = iter.Next() {
		key := append(prefix, iter.Key()[1:]...)
		db.BatchPut(key, iter.Value())
		db.BatchDelete(iter.Key())
	}
	iter.Release()
	err = iter.Error()
	if err != nil {
		return err
	}

	tag := states.StorageItem{}
	tag.Value = []byte{ontid.FLAG_VERSION}
	buf := common.NewZeroCopySink(nil)
	tag.Serialization(buf)
	db.BatchPut(flag, buf.Bytes())
	err = db.BatchCommit()

	return err
}
