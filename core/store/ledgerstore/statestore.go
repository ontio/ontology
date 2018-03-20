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

func NewStateStore(dbDir, merklePath string, currentHeight uint32) (*StateStore, error) {
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
	err = stateStore.init(currentHeight)
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

func (this *StateStore) NewStateBatch(stateRoot common.Uint256) (*StateBatch, error) {
	return NewStateStoreBatch(NewMemDatabase(), this.store, stateRoot)
}

func (this *StateStore) CommitTo() error {
	return this.store.BatchCommit()
}

func (this *StateStore) GetCurrentStateRoot() (common.Uint256, error) {
	key, err := this.getCurrentStateRootKey()
	if err != nil {
		return common.Uint256{}, err
	}
	value, err := this.store.Get(key)
	if err != nil {
		if err == leveldb.ErrNotFound{
			return common.Uint256{}, nil
		}
		return common.Uint256{}, err
	}
	stateRoot, err := common.Uint256ParseFromBytes(value)
	if err != nil {
		return common.Uint256{}, err
	}
	return stateRoot, nil
}


func (this *StateStore) GetContractState(contractHash common.Uint160) (*payload.DeployCode, error) {
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

func (this *StateStore) SaveCurrentStateRoot(stateRoot common.Uint256) error {
	key, err := this.getCurrentStateRootKey()
	if err != nil {
		return err
	}

	return this.store.BatchPut(key, stateRoot.ToArray())
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

func (this *StateStore) GetVoteStates() (map[common.Uint160]*VoteState, error) {
	votes := make(map[common.Uint160]*VoteState)
	iter := this.store.NewIterator([]byte{byte(ST_Vote)})
	for iter.Next() {
		rk := bytes.NewReader(iter.Key())
		// read prefix
		_, err := serialization.ReadBytes(rk, 1)
		if err != nil {
			return nil, fmt.Errorf("ReadBytes error %s", err)
		}
		var programHash common.Uint160
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

func (this *StateStore) getCurrentStateRootKey() ([]byte, error) {
	key := make([]byte, 1+len(CurrentStateRoot))
	key[0] = byte(SYS_CurrentStateRoot)
	copy(key[1:], []byte(CurrentStateRoot))
	return key, nil
}

func (this *StateStore) getBookKeeperKey() ([]byte, error) {
	key := make([]byte, 1+len(BookerKeeper))
	key[0] = byte(ST_BookKeeper)
	copy(key[1:], []byte(BookerKeeper))
	return key, nil
}

func (this *StateStore) getContractStateKey(contractHash common.Uint160) ([]byte, error) {
	data := contractHash.ToArray()
	key := make([]byte, 1+len(data))
	key[0] = byte(ST_Contract)
	copy(key[1:], []byte(data))
	return key, nil
}

func (this *StateStore) getStorageKey(key *StorageKey) ([]byte, error) {
	data := key.ToArray()
	storeKey := make([]byte, 1+len(data))
	storeKey[0] = byte(ST_Storage)
	copy(storeKey[1:], []byte(data))
	return storeKey, nil
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
