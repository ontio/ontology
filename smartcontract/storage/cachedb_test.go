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
	"bytes"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/leveldbstore"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
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
	prefix := common.ST_STORAGE
	mem := make(map[string]string)
	memback, _ := leveldbstore.NewMemLevelDBStore()
	overlay := overlaydb.NewOverlayDB(memback)

	cache := NewCacheDB(overlay)
	for i := 0; i < N; i++ {
		key, val := genRandKeyVal()
		item := &states.StorageItem{Value: []byte(val)}
		cache.Add(prefix, []byte(key), item)
		mem[key] = val
	}

	for key := range mem {
		op := rand.Int() % 2
		if op == 0 {
			//delete
			delete(mem, key)
			cache.Delete(prefix, []byte(key))
		} else if op == 1 {
			//update
			_, val := genRandKeyVal()
			mem[key] = val
			item := &states.StorageItem{Value: []byte(val)}
			cache.Add(prefix, []byte(key), item)
		}
	}

	for key, val := range mem {
		item, err := cache.Get(prefix, []byte(key))
		assert.Nil(t, err)
		assert.NotNil(t, item)
		v := item.(*states.StorageItem).Value
		assert.Equal(t, []byte(val), v)
	}
	cache.Commit()

	for key, val := range mem {
		raw, err := overlay.Get(append([]byte{byte(prefix)}, key...))
		assert.Nil(t, err)
		assert.NotNil(t, raw)
		item := new(states.StorageItem)
		item.Deserialize(bytes.NewBuffer(raw))

		assert.Equal(t, []byte(val), item.Value)
	}

}
