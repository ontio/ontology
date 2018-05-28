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
package rocksdbstore

import (
	"fmt"
	"github.com/ontio/ontology/core/store/common"
	"github.com/tecbot/gorocksdb"
	"os"
)

// used to compute the size of bloom filter bits array .
// too small will lead to high false positive rate.
const (
	BITSPERKEY        = 10
	CACHE_SIZE        = 256 << 20
	MAX_BG_COMPACTION = 3
	MAX_BG_FLUSH      = 3
)

type RocksDBStore struct {
	db    *gorocksdb.DB
	batch *gorocksdb.WriteBatch
	ro    *gorocksdb.ReadOptions
	wo    *gorocksdb.WriteOptions
}

func NewRocksDBStore(dir string) (*RocksDBStore, error) {
	bbto := gorocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetBlockCache(gorocksdb.NewLRUCache(CACHE_SIZE))
	filter := gorocksdb.NewBloomFilter(BITSPERKEY)
	bbto.SetFilterPolicy(filter)
	opts := gorocksdb.NewDefaultOptions()
	opts.SetBlockBasedTableFactory(bbto)
	opts.SetCreateIfMissing(true)
	opts.SetMaxBackgroundCompactions(MAX_BG_COMPACTION)
	opts.SetMaxBackgroundFlushes(MAX_BG_FLUSH)

	rocksDB := &RocksDBStore{
		batch: nil,
		ro:    gorocksdb.NewDefaultReadOptions(),
		wo:    gorocksdb.NewDefaultWriteOptions(),
	}
	if !rocksDB.isDirExists(dir) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return nil, fmt.Errorf("mkdir:%s error:%s", dir, err)
		}
	}
	db, err := gorocksdb.OpenDb(opts, dir)
	if err != nil {
		return nil, err
	}
	rocksDB.db = db
	return rocksDB, nil
}

func (this *RocksDBStore) isDirExists(dir string) bool {
	_, err := os.Stat(dir)
	return err == nil || os.IsExist(err)
}

func (this *RocksDBStore) Put(key, value []byte) error {
	return this.db.Put(this.wo, key, value)
}

func (this *RocksDBStore) Get(key []byte) ([]byte, error) {
	v, err := this.db.GetBytes(this.ro, key)
	if err != nil {
		return nil, err
	}
	//Compatible with leveldb
	if len(v) == 0 {
		return nil, common.ErrNotFound
	}
	return v, nil
}

func (this *RocksDBStore) Has(key []byte) (bool, error) {
	v, err := this.db.GetBytes(this.ro, key)
	if err != nil {
		return false, err
	}
	return len(v) != 0, nil
}

func (this *RocksDBStore) Delete(key []byte) error {
	return this.db.Delete(this.wo, key)
}

func (this *RocksDBStore) NewBatch() {
	this.batch = gorocksdb.NewWriteBatch()
}

func (this *RocksDBStore) BatchPut(key, value []byte) {
	this.batch.Put(key, value)
}

func (this *RocksDBStore) BatchDelete(key []byte) {
	this.batch.Delete(key)
}

func (this *RocksDBStore) BatchCommit() error {
	err := this.db.Write(this.wo, this.batch)
	if err != nil {
		this.batch.Destroy()
		return err
	}
	this.batch.Destroy()
	return nil
}

func (this *RocksDBStore) Close() error {
	this.db.Close()
	this.ro.Destroy()
	this.wo.Destroy()
	return nil
}

func (this *RocksDBStore) NewIterator(prefix []byte) common.StoreIterator {
	ro := gorocksdb.NewDefaultReadOptions()
	ro.SetFillCache(false)
	iter := this.db.NewIterator(ro)
	iter.Seek(prefix)
	return &RocksDBIterator{
		iter:    iter,
		ro:      ro,
		isFirst: true,
		prefix:  prefix,
	}
}
