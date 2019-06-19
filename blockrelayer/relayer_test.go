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

package blockrelayer

import (
	"encoding/binary"
	"fmt"
	"os"
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/stretchr/testify/assert"
)

func BenchmarkStorage_SaveBlock(b *testing.B) {
	path := "test.db"
	os.RemoveAll(path)
	db, err := Open(path)
	assert.Nil(b, err)
	for i := 0; i < b.N; i++ {
		db.SaveBlockTest(uint32(i))
	}
}

func BenchmarkStorageBackend_GetBlockByHash(b *testing.B) {
	b.StopTimer()
	path := "test.db"
	os.RemoveAll(path)
	db, err := Open(path)
	assert.Nil(b, err)
	const N = 10000
	for i := 0; i < N; i++ {
		db.SaveBlockTest(uint32(i))
	}

	db.Flush()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		height := i % N
		var blockHash common.Uint256
		binary.LittleEndian.PutUint32(blockHash[:], uint32(height))
		raw, err := db.GetBlockByHash(blockHash)
		assert.Nil(b, err, "error at", i)
		assert.Equal(b, raw.Hash, blockHash)
	}

}

func TestFile(t *testing.T) {
	path := "test.db"
	os.RemoveAll(path)
	db, err := Open(path)
	assert.Nil(t, err)

	const N = 10
	for i := uint32(0); i < N; i++ {
		if i%1000 == 0 {
			fmt.Println("begin save ", i)
		}
		db.SaveBlockTest(i)
	}

	db.Flush()

	fmt.Println("begin get")
	for i := uint32(0); i < N; i++ {
		if i%10000 == 0 {
			fmt.Println("begin get ", i)
		}
		var blockHash common.Uint256
		binary.LittleEndian.PutUint32(blockHash[:], i)
		raw, err := db.GetBlockByHash(blockHash)
		assert.Nil(t, err, "error at", i)
		assert.Equal(t, raw.Hash, blockHash)
	}

	fmt.Println("done")
}
