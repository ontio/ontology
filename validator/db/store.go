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

package db

import (
	"errors"
	"fmt"
	"sync"

	"github.com/ontio/ontology/common"
	storcomm "github.com/ontio/ontology/core/store/common"
	leveldb "github.com/ontio/ontology/core/store/leveldbstore"
	"github.com/ontio/ontology/core/types"
	pool "github.com/valyala/bytebufferpool"
)

var keyPool pool.Pool
var valuePool pool.Pool

type Store struct {
	db storcomm.PersistStore

	mutex           sync.RWMutex // guard the following var
	bestBlockHeader *types.Header
	genesisBlock    *types.Block
}

func NewStore(path string) (*Store, error) {
	ldb, err := leveldb.NewLevelDBStore(path)
	if err != nil {
		return nil, err
	}

	st := &Store{db: ldb}
	err = st.init()
	if err != nil {
		return nil, err
	}

	return st, nil
}

func (self *Store) init() error {
	prefix := []byte{byte(SYS_VERSION)}
	version, err := self.db.Get(prefix)
	if err != nil {
		version = []byte{0x00}
	}

	if version[0] == 0x01 {
		//test if genesis block in db
		genesis, err := self.db.Get([]byte{byte(SYS_GENESIS_BLOCK)})
		if err != nil {
			self.bestBlockHeader = nil
			self.genesisBlock = nil
			return nil
		}

		self.genesisBlock, err = types.BlockFromRawBytes(genesis)
		if err != nil {
			return errors.New(fmt.Sprint("inconsist db: genesis block deserialize failed. cause of:\n ", err.Error()))
		}

		best, err := self.db.Get([]byte{byte(SYS_BEST_BLOCK_HEADER)})
		if err != nil {
			return errors.New("inconsist db: best blockheader not in db")
		}

		self.bestBlockHeader, err = types.HeaderFromRawBytes(best)
		if err != nil {
			return errors.New(fmt.Sprint("inconsist db: best blockheader deserialize failed. cause of:\n ", err.Error()))
		}

		return nil
	} else {
		self.bestBlockHeader = nil
		self.genesisBlock = nil
		// can not find version info
		iter := self.db.NewIterator(nil)
		if iter.Next() {
			iter.Release()
			return errors.New("not a fresh db")
		}
		iter.Release()

		// put version to db
		err := self.db.Put(prefix, []byte{0x01})

		return err
	}

}

func (self *Store) GetBestBlock() (BestBlock, error) {
	if self.bestBlockHeader == nil {
		return BestBlock{}, errors.New("fresh db")
	}
	return BestBlock{
		Height: self.bestBlockHeader.Height,
		Hash:   self.bestBlockHeader.Hash(),
	}, nil
}

func (self *Store) GetBestHeader() (*types.Header, error) {
	if self.bestBlockHeader == nil {
		return nil, errors.New("fresh db")
	}

	return self.bestBlockHeader, nil
}

// implement  TransactionProvider interface
func (self *Store) ContainTransaction(hash common.Uint256) bool {
	_, err := self.GetTransactionBytes(hash)
	return err == nil
}

func (self *Store) GetTransactionBytes(hash common.Uint256) ([]byte, error) {
	key := GenDataTransactionKey(hash)
	defer keyPool.Put(key)
	txn, err := self.db.Get(key.Bytes())

	return txn, err
}

func (self *Store) GetTransaction(hash common.Uint256) (*types.Transaction, error) {
	buf, err := self.GetTransactionBytes(hash)
	if err != nil {
		return nil, err
	}
	txn, err := types.TransactionFromRawBytes(buf)
	if err != nil {
		return nil, err
	}
	return txn, nil
}

func (self *Store) Close() error {
	err := self.db.Close()
	self.db = nil
	return err
}

func (self *Store) saveTransaction(tx *types.Transaction, height uint32) error {
	// generate key with DATA_TRANSACTION prefix
	key := GenDataTransactionKey(tx.Hash())
	defer keyPool.Put(key)

	value := common.NewZeroCopySink(nil)
	value.WriteUint32(height)
	tx.Serialization(value)

	// put value
	self.db.BatchPut(key.Bytes(), value.Bytes())
	return nil
}

func (self *Store) PersistBlock(block *types.Block) error {
	height := block.Header.Height
	if !((self.bestBlockHeader == nil && height == 0) || height == self.bestBlockHeader.Height+1) {
		return errors.New("can't persist discontinuous block")
	}

	self.mutex.Lock()
	defer self.mutex.Unlock()

	self.db.NewBatch()
	for _, txn := range block.Transactions {
		err := self.saveTransaction(txn, height)
		if err != nil {
			return err
		}
	}

	// is genesis block
	if self.bestBlockHeader == nil {
		key := GenGenesisBlockKey()
		defer keyPool.Put(key)
		value := valuePool.Get()
		defer valuePool.Put(value)

		sink := common.NewZeroCopySink(nil)
		block.Serialization(sink)
		value.Write(sink.Bytes())
		self.db.BatchPut(key.Bytes(), value.Bytes())
	}

	key := GenBestBlockHeaderKey()
	defer keyPool.Put(key)
	value := valuePool.Get()
	defer valuePool.Put(value)

	header := block.Header

	sink := common.NewZeroCopySink(nil)
	header.Serialization(sink)
	value.Write(sink.Bytes())
	self.db.BatchPut(key.Bytes(), value.Bytes())

	err := self.db.BatchCommit()

	if err != nil {
		return err
	}

	if self.bestBlockHeader == nil {
		self.genesisBlock = block
	}
	self.bestBlockHeader = block.Header

	return err
}
