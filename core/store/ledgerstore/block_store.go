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
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	scom "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/leveldbstore"
	"github.com/ontio/ontology/core/types"
)

//Block store save the data of block & transaction
type BlockStore struct {
	enableCache bool                       //Is enable lru cache
	dbDir       string                     //The path of store file
	cache       *BlockCache                //The cache of block, if have.
	store       *leveldbstore.LevelDBStore //block store handler
}

//NewBlockStore return the block store instance
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

//NewBatch start a commit batch
func (this *BlockStore) NewBatch() {
	this.store.NewBatch()
}

//SaveBlock persist block to store
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
		this.SaveTransaction(tx, blockHeight)
	}
	return nil
}

//ContainBlock return the block specified by block hash save in store
func (this *BlockStore) ContainBlock(blockHash common.Uint256) (bool, error) {
	if this.enableCache {
		if this.cache.ContainBlock(blockHash) {
			return true, nil
		}
	}
	key := genHeaderKey(blockHash)
	_, err := this.store.Get(key)
	if err != nil {
		if err == scom.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

//GetBlock return block by block hash
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
			return nil, fmt.Errorf("GetTransaction %s error %s", txHash.ToHexString(), err)
		}
		if tx == nil {
			return nil, fmt.Errorf("cannot get transaction %s", txHash.ToHexString())
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
	key := genHeaderKey(blockHash)
	value, err := this.store.Get(key)
	if err != nil {
		return nil, nil, err
	}
	source := common.NewZeroCopySource(value)
	sysFee := new(common.Fixed64)
	err = sysFee.Deserialization(source)
	if err != nil {
		return nil, nil, err
	}
	header := new(types.Header)
	err = header.Deserialization(source)
	if err != nil {
		return nil, nil, err
	}
	txSize, eof := source.NextUint32()
	if eof {
		return nil, nil, io.ErrUnexpectedEOF
	}
	txHashes := make([]common.Uint256, 0, int(txSize))
	for i := uint32(0); i < txSize; i++ {
		txHash, eof := source.NextHash()
		if eof {
			return nil, nil, io.ErrUnexpectedEOF
		}
		txHashes = append(txHashes, txHash)
	}
	return header, txHashes, nil
}

//SaveHeader persist block header to store
func (this *BlockStore) SaveHeader(block *types.Block, sysFee common.Fixed64) error {
	blockHash := block.Hash()
	key := genHeaderKey(blockHash)
	sink := common.NewZeroCopySink(nil)
	sysFee.Serialization(sink)
	block.Header.Serialization(sink)
	sink.WriteUint32(uint32(len(block.Transactions)))
	for _, tx := range block.Transactions {
		txHash := tx.Hash()
		sink.WriteHash(txHash)
	}
	this.store.BatchPut(key, sink.Bytes())
	return nil
}

//GetHeader return the header specified by block hash
func (this *BlockStore) GetHeader(blockHash common.Uint256) (*types.Header, error) {
	if this.enableCache {
		block := this.cache.GetBlock(blockHash)
		if block != nil {
			return block.Header, nil
		}
	}
	return this.loadHeader(blockHash)
}

func (this *BlockStore) GetRawHeader(blockHash common.Uint256) (*types.RawHeader, error) {
	if this.enableCache {
		block := this.cache.GetBlock(blockHash)
		if block != nil {
			return block.Header.GetRawHeader(), nil
		}
	}
	return this.loadRawHeader(blockHash)
}

//GetSysFeeAmount return the sys fee for block by block hash
func (this *BlockStore) GetSysFeeAmount(blockHash common.Uint256) (common.Fixed64, error) {
	key := genHeaderKey(blockHash)
	data, err := this.store.Get(key)
	if err != nil {
		return common.Fixed64(0), err
	}
	source := common.NewZeroCopySource(data)
	var fee common.Fixed64
	err = fee.Deserialization(source)
	if err != nil {
		return common.Fixed64(0), err
	}
	return fee, nil
}

func (this *BlockStore) loadHeader(blockHash common.Uint256) (*types.Header, error) {
	rawHeader, err := this.loadRawHeader(blockHash)
	if err != nil {
		return nil, err
	}
	source := common.NewZeroCopySource(rawHeader.Payload)
	header := new(types.Header)
	err = header.Deserialization(source)
	if err != nil {
		return nil, err
	}
	return header, nil
}

func (this *BlockStore) loadRawHeader(blockHash common.Uint256) (*types.RawHeader, error) {
	key := genHeaderKey(blockHash)
	value, err := this.store.Get(key)
	if err != nil {
		return nil, err
	}
	source := common.NewZeroCopySource(value)
	sysFee := new(common.Fixed64)
	err = sysFee.Deserialization(source)
	if err != nil {
		return nil, err
	}
	header := &types.RawHeader{}
	err = header.Deserialization(source)
	if err != nil {
		return nil, err
	}

	return header, nil
}

//GetCurrentBlock return the current block hash and current block height
func (this *BlockStore) GetCurrentBlock() (common.Uint256, uint32, error) {
	key := genCurrentBlockKey()
	data, err := this.store.Get(key)
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

//SaveCurrentBlock persist the current block height and current block hash to store
func (this *BlockStore) SaveCurrentBlock(height uint32, blockHash common.Uint256) error {
	key := genCurrentBlockKey()
	value := common.NewZeroCopySink(nil)
	value.WriteHash(blockHash)
	value.WriteUint32(height)
	this.store.BatchPut(key, value.Bytes())
	return nil
}

//GetHeaderIndexList return the head index store in header index list
func (this *BlockStore) GetHeaderIndexList() (map[uint32]common.Uint256, error) {
	result := make(map[uint32]common.Uint256)
	iter := this.store.NewIterator([]byte{byte(scom.IX_HEADER_HASH_LIST)})
	defer iter.Release()
	for iter.Next() {
		startCount, err := genStartHeightByHeaderIndexKey(iter.Key())
		if err != nil {
			return nil, fmt.Errorf("genStartHeightByHeaderIndexKey error %s", err)
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
	if err := iter.Error(); err != nil {
		return nil, err
	}
	return result, nil
}

//SaveHeaderIndexList persist header index list to store
func (this *BlockStore) SaveHeaderIndexList(startIndex uint32, indexList []common.Uint256) {
	indexKey := genHeaderIndexListKey(startIndex)
	indexSize := uint32(len(indexList))
	value := common.NewZeroCopySink(nil)
	value.WriteUint32(indexSize)
	for _, hash := range indexList {
		value.WriteHash(hash)
	}

	this.store.BatchPut(indexKey, value.Bytes())
}

//GetBlockHash return block hash by block height
func (this *BlockStore) GetBlockHash(height uint32) (common.Uint256, error) {
	key := genBlockHashKey(height)
	value, err := this.store.Get(key)
	if err != nil {
		return common.UINT256_EMPTY, err
	}
	blockHash, err := common.Uint256ParseFromBytes(value)
	if err != nil {
		return common.UINT256_EMPTY, err
	}
	return blockHash, nil
}

//SaveBlockHash persist block height and block hash to store
func (this *BlockStore) SaveBlockHash(height uint32, blockHash common.Uint256) {
	key := genBlockHashKey(height)
	this.store.BatchPut(key, blockHash.ToArray())
}

//SaveTransaction persist transaction to store
func (this *BlockStore) SaveTransaction(tx *types.Transaction, height uint32) {
	if this.enableCache {
		this.cache.AddTransaction(tx, height)
	}
	this.putTransaction(tx, height)
}

func (this *BlockStore) putTransaction(tx *types.Transaction, height uint32) {
	txHash := tx.Hash()
	key := genTransactionKey(txHash)
	value := common.NewZeroCopySink(nil)
	value.WriteUint32(height)
	tx.Serialization(value)
	this.store.BatchPut(key, value.Bytes())
}

//GetTransaction return transaction by transaction hash
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
	key := genTransactionKey(txHash)

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
		return nil, 0, err
	}
	source := common.NewZeroCopySource(value)
	var eof bool
	height, eof = source.NextUint32()
	if eof {
		return nil, 0, io.ErrUnexpectedEOF
	}
	tx = new(types.Transaction)
	err = tx.Deserialization(source)
	if err != nil {
		return nil, 0, fmt.Errorf("transaction deserialize error %s", err)
	}
	return tx, height, nil
}

//IsContainTransaction return whether the transaction is in store
func (this *BlockStore) ContainTransaction(txHash common.Uint256) (bool, error) {
	key := genTransactionKey(txHash)

	if this.enableCache {
		if this.cache.ContainTransaction(txHash) {
			return true, nil
		}
	}
	_, err := this.store.Get(key)
	if err != nil {
		if err == scom.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

//GetVersion return the version of store
func (this *BlockStore) GetVersion() (byte, error) {
	key := genVersionKey()
	value, err := this.store.Get(key)
	if err != nil {
		return 0, err
	}
	reader := bytes.NewReader(value)
	return reader.ReadByte()
}

//SaveVersion persist version to store
func (this *BlockStore) SaveVersion(ver byte) error {
	key := genVersionKey()
	return this.store.Put(key, []byte{ver})
}

//ClearAll clear all the data of block store
func (this *BlockStore) ClearAll() error {
	this.NewBatch()
	iter := this.store.NewIterator(nil)
	for iter.Next() {
		this.store.BatchDelete(iter.Key())
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return err
	}
	return this.CommitTo()
}

//CommitTo commit the batch to store
func (this *BlockStore) CommitTo() error {
	return this.store.BatchCommit()
}

//Close block store
func (this *BlockStore) Close() error {
	return this.store.Close()
}

func genTransactionKey(txHash common.Uint256) []byte {
	key := bytes.NewBuffer(nil)
	key.WriteByte(byte(scom.DATA_TRANSACTION))
	txHash.Serialize(key)
	return key.Bytes()
}

func genHeaderKey(blockHash common.Uint256) []byte {
	data := blockHash.ToArray()
	key := make([]byte, 1+len(data))
	key[0] = byte(scom.DATA_HEADER)
	copy(key[1:], data)
	return key
}

func genBlockHashKey(height uint32) []byte {
	key := make([]byte, 5, 5)
	key[0] = byte(scom.DATA_BLOCK_HASH)
	binary.LittleEndian.PutUint32(key[1:], height)
	return key
}

func genCurrentBlockKey() []byte {
	return []byte{byte(scom.SYS_CURRENT_BLOCK)}
}

func genVersionKey() []byte {
	return []byte{byte(scom.SYS_VERSION)}
}

func genHeaderIndexListKey(startHeight uint32) []byte {
	sink := common.NewZeroCopySink(nil)
	sink.WriteByte(byte(scom.IX_HEADER_HASH_LIST))
	sink.WriteUint32(startHeight)
	return sink.Bytes()
}

func genStartHeightByHeaderIndexKey(key []byte) (uint32, error) {
	reader := bytes.NewReader(key[1:])
	height, err := serialization.ReadUint32(reader)
	if err != nil {
		return 0, err
	}
	return height, nil
}

func genBlockPruneHeightKey() []byte {
	return []byte{byte(scom.DATA_BLOCK_PRUNE_HEIGHT)}
}

func (this *BlockStore) GetBlockPrunedHeight() (uint32, error) {
	key := genBlockPruneHeightKey()
	data, err := this.store.Get(key)
	if err != nil {
		if err == scom.ErrNotFound {
			return 0, nil
		}
		return 0, err
	}
	height, eof := common.NewZeroCopySource(data).NextUint32()
	if eof {
		return 0, io.ErrUnexpectedEOF
	}

	return height, nil
}

func (this *BlockStore) SaveBlockPrunedHeight(height uint32) {
	key := genBlockPruneHeightKey()
	sink := common.NewZeroCopySink(nil)
	sink.WriteUint32(height)

	this.store.BatchPut(key, sink.Bytes())
}

func (this *BlockStore) PruneBlock(hash common.Uint256) []common.Uint256 {
	_, txHashes, err := this.loadHeaderWithTx(hash)
	if err != nil {
		return nil
	}
	for _, hash := range txHashes {
		key := genTransactionKey(hash)
		this.store.BatchDelete(key)
	}
	key := genHeaderKey(hash)
	this.store.BatchDelete(key)
	return txHashes
}
