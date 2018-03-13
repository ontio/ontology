package ledgerstore

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/Ontology/common"
	"github.com/Ontology/common/serialization"
	. "github.com/Ontology/core/store/common"
	"github.com/Ontology/core/store/leveldbstore"
	"github.com/Ontology/core/types"
	"github.com/syndtr/goleveldb/leveldb"
)

type BlockStore struct {
	enableCache bool
	dbDir       string
	cache       *BlockCache
	store       *leveldbstore.LevelDBStore
}

func NewBlockStore(dbDir string, enableCache bool) (*BlockStore, error) {
	var cache *BlockCache
	var err error
	if enableCache {
		cache, err = NewBlockCache()
		if err != nil {
			return nil, fmt.Errorf("NewBlockCache error %s", err)
		}
	}

	store, err := leveldbstore.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	blockStore := &BlockStore{
		dbDir:       dbDir,
		enableCache: enableCache,
		store:       store,
		cache:       cache,
	}
	return blockStore, nil
}

func (this *BlockStore) NewBatch() error {
	return this.store.NewBatch()
}

func (this *BlockStore) SaveBlock(block *types.Block) error {
	if this.enableCache {
		this.cache.AddBlock(block)
	}

	blockHeight := block.Header.Height
	err := this.SaveHeader(block, 0)
	if err != nil {
		return fmt.Errorf("SaveHeader error %s", err)
	}
	for _, tx := range block.Transactions {
		err = this.SaveTransaction(tx, blockHeight)
		if err != nil {
			return fmt.Errorf("SaveTransaction block height %d tx %x err %s", blockHeight, tx.Hash(), err)
		}
	}
	return nil
}

func (this *BlockStore) ContainBlock(blockHash common.Uint256) (bool, error) {
	if this.enableCache {
		if this.cache.ContainBlock(blockHash) {
			return true, nil
		}
	}
	key := this.getHeaderKey(blockHash)
	_, err := this.store.Get(key)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (this *BlockStore) GetBlock(blockHash common.Uint256) (*types.Block, error) {
	var block *types.Block
	if this.enableCache {
		block = this.cache.GetBlock(blockHash)
		if block != nil {
			return block, nil
		}
	}
	header, txHashes, err := this.loadHeaderWithTx(blockHash)
	if err != nil {
		return nil, err
	}
	txList := make([]*types.Transaction, 0, len(txHashes))
	for _, txHash := range txHashes {
		tx, _, err := this.GetTransaction(txHash)
		if err != nil {
			return nil, fmt.Errorf("GetTransaction %x error %s", txHash, err)
		}
		if tx == nil {
			return nil, fmt.Errorf("cannot get transaction %x", txHash)
		}
		txList = append(txList, tx)
	}
	block = &types.Block{
		Header:       header,
		Transactions: txList,
	}
	return block, nil
}

func (this *BlockStore) loadHeaderWithTx(blockHash common.Uint256) (*types.Header, []common.Uint256, error) {
	key := this.getHeaderKey(blockHash)
	value, err := this.store.Get(key)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	reader := bytes.NewBuffer(value)
	sysFee := new(common.Fixed64)
	err = sysFee.Deserialize(reader)
	if err != nil {
		return nil, nil, err
	}
	header := new(types.Header)
	err = header.Deserialize(reader)
	if err != nil {
		return nil, nil, err
	}
	txSize, err := serialization.ReadUint32(reader)
	if err != nil {
		return nil, nil, err
	}
	txHashes := make([]common.Uint256, 0, int(txSize))
	for i := uint32(0); i < txSize; i++ {
		txHash := common.Uint256{}
		err = txHash.Deserialize(reader)
		if err != nil {
			return nil, nil, err
		}
		txHashes = append(txHashes, txHash)
	}
	return header, txHashes, nil
}

func (this *BlockStore) SaveHeader(block *types.Block, sysFee common.Fixed64) error {
	blockHash := block.Hash()
	key := this.getHeaderKey(blockHash)
	value := bytes.NewBuffer(nil)
	err := sysFee.Serialize(value)
	if err != nil {
		return err
	}
	block.Header.Serialize(value)
	serialization.WriteUint32(value, uint32(len(block.Transactions)))
	for _, tx := range block.Transactions {
		txHash := tx.Hash()
		_, err = txHash.Serialize(value)
		if err != nil {
			return err
		}
	}
	return this.store.BatchPut(key, value.Bytes())
}

func (this *BlockStore) GetHeader(blockHash common.Uint256) (*types.Header, error) {
	if this.enableCache {
		block := this.cache.GetBlock(blockHash)
		if block != nil {
			return block.Header, nil
		}
	}
	return this.loadHeader(blockHash)
}

func (this *BlockStore) GetSysFeeAmount(blockHash common.Uint256) (common.Fixed64, error) {
	key := this.getHeaderKey(blockHash)
	data, err := this.store.Get(key)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return common.Fixed64(0), nil
		}
		return common.Fixed64(0), err
	}
	reader := bytes.NewBuffer(data)
	var fee common.Fixed64
	err = fee.Deserialize(reader)
	if err != nil {
		return common.Fixed64(0), err
	}
	return fee, nil
}

func (this *BlockStore) loadHeader(blockHash common.Uint256) (*types.Header, error) {
	key := this.getHeaderKey(blockHash)
	value, err := this.store.Get(key)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	reader := bytes.NewBuffer(value)
	sysFee := new(common.Fixed64)
	err = sysFee.Deserialize(reader)
	if err != nil {
		return nil, err
	}
	header := new(types.Header)
	err = header.Deserialize(reader)
	if err != nil {
		return nil, err
	}
	return header, nil
}

func (this *BlockStore) GetCurrentBlock() (common.Uint256, uint32, error) {
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

func (this *BlockStore) SaveCurrentBlock(height uint32, blockHash common.Uint256) error {
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

func (this *BlockStore) GetHeaderIndexList() (map[uint32]common.Uint256, error) {
	result := make(map[uint32]common.Uint256)
	iter := this.store.NewIterator([]byte{byte(IX_HeaderHashList)})
	for iter.Next() {
		startCount, err := this.getStartHeightByHeaderIndexKey(iter.Key())
		if err != nil {
			return nil, fmt.Errorf("getStartHeightByHeaderIndexKey error %s", err)
		}
		reader := bytes.NewReader(iter.Value())
		count, err := serialization.ReadUint32(reader)
		if err != nil {
			return nil, fmt.Errorf("serialization.ReadUint32 count error %s", err)
		}
		for i := uint32(0); i < count; i++ {
			height := startCount + i
			blockHash := common.Uint256{}
			err = blockHash.Deserialize(reader)
			if err != nil {
				return nil, fmt.Errorf("blockHash.Deserialize error %s", err)
			}
			result[height] = blockHash
		}
	}
	iter.Release()
	return result, nil
}

func (this *BlockStore) SaveHeaderIndexList(startIndex uint32, indexList []common.Uint256) error {
	indexKey := this.getHeaderIndexListKey(startIndex)
	indexSize := uint32(len(indexList))
	value := bytes.NewBuffer(nil)
	serialization.WriteUint32(value, indexSize)
	for _, hash := range indexList {
		hash.Serialize(value)
	}

	return this.store.BatchPut(indexKey, value.Bytes())
}

func (this *BlockStore) GetBlockHash(height uint32) (common.Uint256, error) {
	key, err := this.getBlockHashKey(height)
	if err != nil {
		return common.Uint256{}, err
	}
	value, err := this.store.Get(key)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return common.Uint256{}, nil
		}
		return common.Uint256{}, err
	}
	blockHash, err := common.Uint256ParseFromBytes(value)
	if err != nil {
		return common.Uint256{}, err
	}
	return blockHash, nil
}

func (this *BlockStore) SaveBlockHash(height uint32, blockHash common.Uint256) error {
	key, err := this.getBlockHashKey(height)
	if err != nil {
		return err
	}
	return this.store.BatchPut(key, blockHash.ToArray())
}

func (this *BlockStore) SaveTransaction(tx *types.Transaction, height uint32) error {
	if this.enableCache {
		this.cache.AddTransaction(tx, height)
	}
	return this.putTransaction(tx, height)
}

func (this *BlockStore) putTransaction(tx *types.Transaction, height uint32) error {
	txHash := tx.Hash()
	key := this.getTransactionKey(txHash)
	value := bytes.NewBuffer(nil)
	serialization.WriteUint32(value, height)
	err := tx.Serialize(value)
	if err != nil {
		return err
	}
	return this.store.BatchPut(key, value.Bytes())
}

func (this *BlockStore) GetTransaction(txHash common.Uint256) (*types.Transaction, uint32, error) {
	if this.enableCache {
		tx, height := this.cache.GetTransaction(txHash)
		if tx != nil {
			return tx, height, nil
		}
	}
	return this.loadTransaction(txHash)
}

func (this *BlockStore) loadTransaction(txHash common.Uint256) (*types.Transaction, uint32, error) {
	key := this.getTransactionKey(txHash)

	var tx *types.Transaction
	var height uint32
	if this.enableCache {
		tx, height = this.cache.GetTransaction(txHash)
		if tx != nil {
			return tx, height, nil
		}
	}

	value, err := this.store.Get(key)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, 0, nil
		}
		return nil, 0, err
	}
	reader := bytes.NewBuffer(value)
	height, err = serialization.ReadUint32(reader)
	if err != nil {
		return nil, 0, fmt.Errorf("ReadUint32 error %s", err)
	}
	tx = new(types.Transaction)
	err = tx.Deserialize(reader)
	if err != nil {
		return nil, 0, fmt.Errorf("transaction deserialize error %s", err)
	}
	return tx, height, nil
}

func (this *BlockStore) ContainTransaction(txHash common.Uint256) (bool, error) {
	key := this.getTransactionKey(txHash)

	if this.enableCache {
		if this.cache.ContainTransaction(txHash) {
			return true, nil
		}
	}

	_, err := this.store.Get(key)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (this *BlockStore) GetVersion() (byte, error) {
	key := this.getVersionKey()
	value, err := this.store.Get(key)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return 0, nil
		}
		return 0, err
	}
	reader := bytes.NewReader(value)
	return reader.ReadByte()
}

func (this *BlockStore) SaveVersion(ver byte) error {
	key := this.getVersionKey()
	return this.store.Put(key, []byte{ver})
}

func (this *BlockStore) ClearAll() error {
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

func (this *BlockStore) CommitTo() error {
	return this.store.BatchCommit()
}

func (this *BlockStore) Close() error {
	return this.store.Close()
}

func (this *BlockStore) getTransactionKey(txHash common.Uint256) []byte {
	key := bytes.NewBuffer(nil)
	key.WriteByte(byte(DATA_Transaction))
	txHash.Serialize(key)
	return key.Bytes()
}

func (this *BlockStore) getHeaderKey(blockHash common.Uint256) []byte {
	data := blockHash.ToArray()
	key := make([]byte, 1+len(data))
	key[0] = byte(DATA_Header)
	copy(key[1:], data)
	return key
}

func (this *BlockStore) getBlockHashKey(height uint32) ([]byte, error) {
	key := make([]byte, 5, 5)
	key[0] = byte(DATA_Block)
	binary.LittleEndian.PutUint32(key[1:], height)
	return key, nil
}

func (this *BlockStore) getCurrentBlockKey() []byte {
	return []byte{byte(SYS_CurrentBlock)}
}

func (this *BlockStore) getBlockMerkleTreeKey() []byte {
	return []byte{byte(SYS_BlockMerkleTree)}
}

func (this *BlockStore) getVersionKey() []byte {
	return []byte{byte(SYS_Version)}
}

func (this *BlockStore) getHeaderIndexListKey(startHeight uint32) []byte {
	key := bytes.NewBuffer(nil)
	key.WriteByte(byte(IX_HeaderHashList))
	serialization.WriteUint32(key, startHeight)
	return key.Bytes()
}

func (this *BlockStore) getStartHeightByHeaderIndexKey(key []byte) (uint32, error) {
	reader := bytes.NewReader(key[1:])
	height, err := serialization.ReadUint32(reader)
	if err != nil {
		return 0, err
	}
	return height, nil
}
