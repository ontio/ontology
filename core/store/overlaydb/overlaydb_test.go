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
	"encoding/binary"
	"math/rand"
	"strconv"
	"testing"

	"github.com/ontio/ontology/core/store/leveldbstore"
	"github.com/stretchr/testify/assert"
)

func makeKey(i int) []byte {
	key := make([]byte, 11)
	copy(key, "key")
	binary.BigEndian.PutUint64(key[3:], uint64(i))
	return key
}

func TestNewOverlayDB(t *testing.T) {
	store, err := leveldbstore.NewMemLevelDBStore()
	assert.Nil(t, err)

	N := 10000

	overlay := NewOverlayDB(store)
	for i := 0; i < N; i++ {
		overlay.Put(makeKey(i), []byte("val"+strconv.Itoa(i)))
	}

	for i := 0; i < N; i++ {
		val, err := overlay.Get(makeKey(i))
		assert.Nil(t, err)
		assert.Equal(t, val, []byte("val"+strconv.Itoa(i)))
	}

	for i := 0; i < N; i += 2 {
		overlay.Delete(makeKey(i))
	}

	iter := overlay.NewIterator([]byte("key"))
	hasfirst := iter.First()
	assert.True(t, hasfirst)
	for i := 1; i < N; i += 2 {
		key := iter.Key()
		val := iter.Value()
		assert.Equal(t, key, makeKey(i))
		assert.Equal(t, val, []byte("val"+strconv.Itoa(i)))
		n := iter.Next()
		assert.True(t, n || i+2 >= N)
	}
}

func BenchmarkOverlayDBSerialPut(b *testing.B) {
	store, _ := leveldbstore.NewMemLevelDBStore()

	N := 100000
	overlay := NewOverlayDB(store)
	for i := 0; i < b.N; i++ {
		overlay.Reset()
		for i := 0; i < N; i++ {
			overlay.Put(makeKey(i), []byte("val"+strconv.Itoa(i)))
		}

	}

}

func BenchmarkOverlayDBRandomPut(b *testing.B) {
	store, _ := leveldbstore.NewMemLevelDBStore()

	N := 100000
	keys := make([]int, N)
	for i := 0; i < N; i++ {
		k := rand.Int() % N
		keys[i] = k
	}
	overlay := NewOverlayDB(store)
	for i := 0; i < b.N; i++ {
		overlay.Reset()
		for i := 0; i < N; i++ {
			overlay.Put(makeKey(i), []byte("val"+strconv.Itoa(keys[i])))
		}

	}

}
