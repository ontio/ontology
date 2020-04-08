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
package storage

import (
	"math/rand"
	"testing"

	"github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/leveldbstore"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/stretchr/testify/assert"
)

func genRandKeyVal() (string, string) {
	p := make([]byte, 100)
	rand.Read(p)
	key := string(p)
	rand.Read(p)
	val := string(p)
	return key, val
}

func TestCacheDB(t *testing.T) {
	N := 10000
	mem := make(map[string]string)
	memback, _ := leveldbstore.NewMemLevelDBStore()
	overlay := overlaydb.NewOverlayDB(memback)

	cache := NewCacheDB(overlay)
	for i := 0; i < N; i++ {
		key, val := genRandKeyVal()
		cache.Put([]byte(key), []byte(val))
		mem[key] = val
	}

	for key := range mem {
		op := rand.Int() % 2
		if op == 0 {
			//delete
			delete(mem, key)
			cache.Delete([]byte(key))
		} else if op == 1 {
			//update
			_, val := genRandKeyVal()
			mem[key] = val
			cache.Put([]byte(key), []byte(val))
		}
	}

	for key, val := range mem {
		value, err := cache.Get([]byte(key))
		assert.Nil(t, err)
		assert.NotNil(t, value)
		assert.Equal(t, []byte(val), value)
	}
	cache.Commit()

	prefix := common.ST_STORAGE
	for key, val := range mem {
		pkey := make([]byte, 1+len(key))
		pkey[0] = byte(prefix)
		copy(pkey[1:], key)
		raw, err := overlay.Get(pkey)
		assert.Nil(t, err)
		assert.NotNil(t, raw)
		assert.Equal(t, []byte(val), raw)
	}

}
