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

package leveldbstore

import (
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

//Iterator of leveldb. Warp struct of leveldb iterator
type Iterator struct {
	iter iterator.Iterator
}

func (it *Iterator) Next() bool {
	return it.iter.Next()
}

func (it *Iterator) Prev() bool {
	return it.iter.Prev()
}

func (it *Iterator) First() bool {
	return it.iter.First()
}

func (it *Iterator) Last() bool {
	return it.iter.Last()
}

func (it *Iterator) Seek(key []byte) bool {
	return it.iter.Seek(key)
}

func (it *Iterator) Key() []byte {
	return it.iter.Key()
}

func (it *Iterator) Value() []byte {
	return it.iter.Value()
}

func (it *Iterator) Release() {
	it.iter.Release()
}
