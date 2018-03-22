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
	"fmt"
	"github.com/Ontology/common"
	"github.com/Ontology/common/serialization"
	. "github.com/Ontology/core/states"
	. "github.com/Ontology/core/store/common"
	"github.com/Ontology/core/store/leveldbstore"
	. "github.com/Ontology/core/store/statestore"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/Ontology/merkle"
	"github.com/Ontology/core/payload"
)

var (
	CurrentStateRoot = []byte("Current-State-Root")
	BookerKeeper     = []byte("Booker-Keeper")
)

type StateStore struct {
	dbDir string
	store IStore
	merklePath      string
	merkleTree      *merkle.CompactMerkleTree
	merkleHashStore *merkle.FileHashStore
}

func NewStateStore(dbDir, merklePath string) (*StateStore, error) {
	var err error
	store, err := leveldbstore.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	stateStore := &StateStore{
		dbDir: dbDir,
		store: store,
		merklePath:merklePath,
	}
	_, height, err :=stateStore.GetCurrentBlock()
	if err != nil {
		return nil, fmt.Errorf("GetCurrentBlock error %s", err)
	}
	err = stateStore.init(height)
	if err != nil {
		return nil,fmt.Errorf("init error %s", err)
	}
	return stateStore, nil
}

func (this *StateStore) NewBatch() error {
	err := this.store.NewBatch()
	if err != nil {
		return fmt.Errorf("NewBatch error %s", err)
	}
	return nil
}

func (this *StateStore) init(currBlockHeight uint32)error{
	treeSize, hashes, err := this.GetMerkleTree()
	if err != nil {
		return err
	}
	if treeSize > 0 && treeSize != currBlockHeight+1 {
		return fmt.Errorf("merkle tree size is inconsistent with blockheight: %d", currBlockHeight+1)
	}
	this.merkleHashStore, err = merkle.NewFileHashStore(this.merklePath, treeSize)
	if err != nil {
		return fmt.Errorf("merkle store is inconsistent with ChainStore. persistence will be disabled")
	}
	this.merkleTree = merkle.NewTree(treeSize, hashes, this.merkleHashStore)
	return nil
}

func (this *StateStore) GetMerkleTree() (uint32, []common.Uint256, error) {
	key, err := this.getMerkleTreeKey()
	if err != nil {
		return 0, nil, err
	}
	data, err := this.store.Get(key)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return 0, nil, nil
		}
		return 0, nil, err
	}
	value := bytes.NewBuffer(data)
	treeSize, err := serialization.ReadUint32(value)
	if err != nil {
		return 0, nil, err
	}
	hashCount := (len(data) - 4) / common.UINT256SIZE
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

func (this *StateStore) AddMerkleTreeRoot(txRoot common.Uint256) error {
	key, err := this.getMerkleTreeKey()
	if err != nil {
		return err
	}
	this.merkleTree.AppendHash(txRoot)
	err = this.merkleHashStore.Flush()
	if err != nil {
		return err
	}
	treeSize := this.merkleTree.TreeSize()
	hashes := this.merkleTree.Hashes()
	value := bytes.NewBuffer(make([]byte, 0, 4+len(hashes)*common.UINT256SIZE))
	err = serialization.WriteUint32(value, treeSize)
	if err != nil {
		return err
	}
	for _, hash := range hashes {
		_, err = hash.Serialize(value)
		if err != nil {
			return err
		}
	}
	return this.store.BatchPut(key, value.Bytes())
}

func (this *StateStore) NewStateBatch() *StateBatch {
	return NewStateStoreBatch(NewMemDatabase(), this.store)
}

func (this *StateStore) CommitTo() error {
	return this.store.BatchCommit()
}

func (this *StateStore) GetContractState(contractHash common.Address) (*payload.DeployCode, error) {
	key, err := this.getContractStateKey(contractHash)
	if err != nil {
		return nil, err
	}

	value, err := this.store.Get(key)
	if err != nil {
		if err == leveldb.ErrNotFound{
			return nil, nil
		}
		return nil, err
	}
	reader := bytes.NewReader(value)
	contractState := new(payload.DeployCode)
	err = contractState.Deserialize(reader)
	if err != nil {
		return nil, err
	}
	return contractState, nil
}

func (this *StateStore) GetBookKeeperState() (*BookKeeperState, error) {
	key, err := this.getBookKeeperKey()
	if err != nil {
		return nil, err
	}

	value, err := this.store.Get(key)
	if err != nil {
		if err == leveldb.ErrNotFound{
			return nil, nil
		}
		return nil, err
	}
	reader := bytes.NewReader(value)
	bookKeeperState := new(BookKeeperState)
	err = bookKeeperState.Deserialize(reader)
	if err != nil {
		return nil, err
	}
	return bookKeeperState, nil
}

func (this *StateStore) SaveBookKeeperState(bookKeeperState *BookKeeperState) error {
	key, err := this.getBookKeeperKey()
	if err != nil {
		return err
	}
	value := bytes.NewBuffer(nil)
	err = bookKeeperState.Serialize(value)
	if err != nil {
		return err
	}

	return this.store.Put(key, value.Bytes())
}

func (this *StateStore) GetStorageState(key *StorageKey) (*StorageItem, error) {
	storeKey, err := this.getStorageKey(key)
	if err != nil {
		return nil, err
	}

	data, err := this.store.Get(storeKey)
	if err != nil {
		if err == leveldb.ErrNotFound{
			return nil,nil
		}
		return nil, err
	}
	reader := bytes.NewReader(data)
	storageState := new(StorageItem)
	err = storageState.Deserialize(reader)
	if err != nil {
		return nil, err
	}
	return storageState, nil
}

func (this *StateStore) GetVoteStates() (map[common.Address]*VoteState, error) {
	votes := make(map[common.Address]*VoteState)
	iter := this.store.NewIterator([]byte{byte(ST_Vote)})
	for iter.Next() {
		rk := bytes.NewReader(iter.Key())
		// read prefix
		_, err := serialization.ReadBytes(rk, 1)
		if err != nil {
			return nil, fmt.Errorf("ReadBytes error %s", err)
		}
		var programHash common.Address
		if err := programHash.Deserialize(rk); err != nil {
			return nil, err
		}
		vote := new(VoteState)
		r := bytes.NewReader(iter.Value())
		if err := vote.Deserialize(r); err != nil {
			return nil, err
		}
		votes[programHash] = vote
	}
	return votes, nil
}

func (this *StateStore) GetCurrentBlock() (common.Uint256, uint32, error) {
	key := this.getCurrentBlockKey()
	data, err := this.store.Get(key)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return common.Uint256{}, 0, nil
		}
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

func (this *StateStore) SaveCurrentBlock(height uint32, blockHash common.Uint256) error {
	key := this.getCurrentBlockKey()
	value := bytes.NewBuffer(nil)
	blockHash.Serialize(value)
	serialization.WriteUint32(value, height)
	err := this.store.BatchPut(key, value.Bytes())
	if err != nil {
		return fmt.Errorf("BatchPut error %s", err)
	}
	return nil
}

func (this *StateStore) getCurrentBlockKey() []byte {
	return []byte{byte(SYS_CurrentBlock)}
}

func (this *StateStore) getBookKeeperKey() ([]byte, error) {
	key := make([]byte, 1+len(BookerKeeper))
	key[0] = byte(ST_BookKeeper)
	copy(key[1:], []byte(BookerKeeper))
	return key, nil
}

func (this *StateStore) getContractStateKey(contractHash common.Address) ([]byte, error) {
	data := contractHash.ToArray()
	key := make([]byte, 1+len(data))
	key[0] = byte(ST_Contract)
	copy(key[1:], []byte(data))
	return key, nil
}

func (this *StateStore) getStorageKey(key *StorageKey) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte( byte(ST_Storage))
	buf.Write(key.CodeHash.ToArray())
	buf.Write(key.Key)
	return buf.Bytes(), nil
}

func (this *StateStore) GetBlockRootWithNewTxRoot(txRoot common.Uint256) common.Uint256 {
	return this.merkleTree.GetRootWithNewLeaf(txRoot)
}

func (this *StateStore) getMerkleTreeKey() ([]byte, error) {
	return []byte{byte(SYS_BlockMerkleTree)}, nil
}

func (this *StateStore) ClearAll() error {
	err := this.store.NewBatch()
	if err != nil {
		return err
	}
	iter := this.store.NewIterator(nil)
	for iter.Next() {
		err = this.store.BatchDelete(iter.Key())
		if err != nil {
			return fmt.Errorf("BatchDelete error %s", err)
		}
	}
	iter.Release()
	return this.store.BatchCommit()
}

func (this *StateStore) Close() error {
	return this.store.Close()
}
