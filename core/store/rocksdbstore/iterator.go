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
	"github.com/tecbot/gorocksdb"
)

type RocksDBIterator struct {
	ro      *gorocksdb.ReadOptions
	iter    *gorocksdb.Iterator
	prefix  []byte
	isFirst bool
}

func (this *RocksDBIterator) Next() bool {
	//Compatible with leveldb. In leveldb the cursor is invalid before next
	if this.isFirst {
		this.isFirst = false
		return this.iter.ValidForPrefix(this.prefix)
	}
	this.iter.Next()
	return this.iter.ValidForPrefix(this.prefix)
}

func (this *RocksDBIterator) Prev() bool {
	this.iter.Prev()
	return this.iter.ValidForPrefix(this.prefix)
}

func (this *RocksDBIterator) First() bool {
	this.iter.SeekToFirst()
	return this.iter.ValidForPrefix(this.prefix)
}

func (this *RocksDBIterator) Last() bool {
	this.iter.SeekToLast()
	return this.iter.ValidForPrefix(this.prefix)
}

func (this *RocksDBIterator) Seek(key []byte) bool {
	this.iter.Seek(key)
	return this.iter.ValidForPrefix(this.prefix)
}

func (this *RocksDBIterator) Key() []byte {
	k := this.iter.Key()
	defer k.Free()
	return k.Data()
}

func (this *RocksDBIterator) Value() []byte {
	v := this.iter.Value()
	defer v.Free()
	return v.Data()
}

func (this *RocksDBIterator) Release() {
	this.iter.Close()
	this.ro.Destroy()
}
