package ledgerstore

import (
	"bytes"
	"fmt"
	"github.com/Ontology/common"
	"github.com/Ontology/common/serialization"
	"github.com/Ontology/core/store/leveldbstore"
	. "github.com/Ontology/core/store/common"
	"github.com/Ontology/merkle"
	"github.com/syndtr/goleveldb/leveldb"
)

type MerkleTreeStore struct {
	dbDir string
	merklePath string
	store           IStore
	merkleTree      *merkle.CompactMerkleTree
	merkleHashStore *merkle.FileHashStore
}

func NewMerkleTreeStore(dbDir, merklePath string, currentHeight uint32) (*MerkleTreeStore, error) {
	store, err := leveldbstore.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	merkleTreeStore := &MerkleTreeStore{
		dbDir:dbDir,
		merklePath:merklePath,
		store: store,
	}
	err = merkleTreeStore.init(currentHeight)
	if err != nil {
		return nil, fmt.Errorf("init error %s", err)
	}
	return merkleTreeStore, nil
}

func (this *MerkleTreeStore) init(currBlockHeight uint32) error {
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

func (this *MerkleTreeStore) NewBatch() error {
	return this.store.NewBatch()
}

func (this *MerkleTreeStore) GetMerkleTree() (uint32, []common.Uint256, error) {
	key, err := this.getMerkleTreeKey()
	if err != nil {
		return 0, nil, err
	}
	data, err := this.store.Get(key)
	if err != nil {
		if err == leveldb.ErrNotFound{
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

func (this *MerkleTreeStore) AddMerkleTreeRoot(txRoot common.Uint256) error {
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

func (this *MerkleTreeStore) ClearAll() error {
	err := this.NewBatch()
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
	return this.CommitTo()
}

func (this *MerkleTreeStore) CommitTo() error {
	return this.store.BatchCommit()
}

func (this *MerkleTreeStore) Close()error{
	this.merkleHashStore.Close()
	return this.store.Close()
}

func (this *MerkleTreeStore) GetBlockRootWithNewTxRoot(txRoot common.Uint256) common.Uint256 {
	return this.merkleTree.GetRootWithNewLeaf(txRoot)
}

func (this *MerkleTreeStore) getMerkleTreeKey() ([]byte, error) {
	return []byte{byte(SYS_BlockMerkleTree)}, nil
}
