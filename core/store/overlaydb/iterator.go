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
package overlaydb

import (
	"github.com/ontio/ontology/core/store/common"
	"github.com/syndtr/goleveldb/leveldb/comparer"
)

type KeyOrigin byte

const (
	FromMem  KeyOrigin = iota
	FromBack           = iota
	FromBoth           = iota
)

type JoinIter struct {
	backend     common.StoreIterator
	memdb       common.StoreIterator
	key, value  []byte
	keyOrigin   KeyOrigin
	nextMemEnd  bool
	nextBackEnd bool
	cmp         comparer.BasicComparer
}

func NewJoinIter(memIter, backendIter common.StoreIterator) *JoinIter {
	return &JoinIter{
		backend: backendIter,
		memdb:   memIter,
		cmp:     comparer.DefaultComparer,
	}
}

func (iter *JoinIter) First() bool {
	f := iter.first()
	if f == false {
		return false
	}
	for len(iter.value) == 0 {
		if iter.next() == false {
			return false
		}
	}

	return true
}

func (iter *JoinIter) first() bool {
	var bkey, bval, mkey, mval []byte
	back := iter.backend.First()
	mem := iter.memdb.First()
	// check error
	if iter.Error() != nil {
		return false
	}
	if back {
		bkey = iter.backend.Key()
		bval = iter.backend.Value()
		if mem == false {
			iter.key = bkey
			iter.value = bval
			iter.keyOrigin = FromBack
			return true
		}
		mkey = iter.memdb.Key()
		mval = iter.memdb.Value()
		cmp := iter.cmp.Compare(mkey, bkey)
		if cmp < 1 {
			iter.key = mkey
			iter.value = mval
			if cmp == 0 {
				iter.keyOrigin = FromBoth
			} else {
				iter.keyOrigin = FromMem
			}
		} else {
			iter.key = bkey
			iter.value = bval
			iter.keyOrigin = FromBack
		}
		return true
	} else {
		if mem {
			iter.key = iter.memdb.Key()
			iter.value = iter.memdb.Value()
			iter.keyOrigin = FromMem
			return true
		}
		return false
	}
}

func (iter *JoinIter) Key() []byte {
	return iter.key
}

func (iter *JoinIter) Value() []byte {
	return iter.value
}

func (iter *JoinIter) Next() bool {
	f := iter.next()
	if f == false {
		return false
	}

	for len(iter.value) == 0 {
		if iter.next() == false {
			return false
		}
	}

	return true
}

func (iter *JoinIter) next() bool {
	if (iter.keyOrigin == FromMem || iter.keyOrigin == FromBoth) && iter.nextMemEnd == false {
		iter.nextMemEnd = !iter.memdb.Next()
	}
	if (iter.keyOrigin == FromBack || iter.keyOrigin == FromBoth) && iter.nextBackEnd == false {
		iter.nextBackEnd = !iter.backend.Next()
	}

	// check error
	if iter.Error() != nil {
		return false
	}

	if iter.nextBackEnd {
		if iter.nextMemEnd {
			iter.key = nil
			iter.value = nil
			return false
		} else {
			iter.key = iter.memdb.Key()
			iter.value = iter.memdb.Value()
			iter.keyOrigin = FromMem
		}
	} else {
		if iter.nextMemEnd {
			iter.key = iter.backend.Key()
			iter.value = iter.backend.Value()
			iter.keyOrigin = FromBack
		} else {
			bkey := iter.backend.Key()
			mkey := iter.memdb.Key()
			cmp := iter.cmp.Compare(mkey, bkey)
			switch cmp {
			case -1:
				iter.key = mkey
				iter.value = iter.memdb.Value()
				iter.keyOrigin = FromMem
			case 0:
				iter.key = mkey
				iter.value = iter.memdb.Value()
				iter.keyOrigin = FromBoth
			case 1:
				iter.key = bkey
				iter.value = iter.backend.Value()
				iter.keyOrigin = FromBack
			default:
				panic("unreachable")
			}
		}
	}

	return true
}

func (iter *JoinIter) Release() {
	iter.memdb.Release()
	iter.backend.Release()
}

func (iter *JoinIter) Error() error {
	if iter.backend.Error() != nil {
		return iter.backend.Error()
	} else if iter.memdb.Error() != nil {
		return iter.memdb.Error()
	}
	return nil
}
