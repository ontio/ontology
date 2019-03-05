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
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDot(t *testing.T) {
	buf := make(map[string]string)
	for i := 0; i < 100; i++ {
		k := fmt.Sprintf("k%d", i)
		v := fmt.Sprintf("v%d", i)
		buf[k] = v
	}
	db := NewMemDB(10, 10)
	for k, v := range buf {
		db.Put([]byte(k), []byte(v))
	}

	//dot := db.DumpToDot()

	//fmt.Println(dot)
}

func TestDB(t *testing.T) {
	buf := make([][1]byte, 10)
	for i := range buf {
		buf[i][0] = byte(i)
	}

	db := NewMemDB(10, 2)
	for i := range buf {
		db.Put(buf[i][:], nil)
	}

	for i := range buf {
		val, unknown := db.Get(buf[i][:])
		assert.False(t, unknown)
		assert.Equal(t, len(val), 0)
	}

	//fmt.Println(db.nodeData)
}

func BenchmarkPut(b *testing.B) {
	buf := make([][4]byte, b.N)
	for i := range buf {
		binary.LittleEndian.PutUint32(buf[i][:], uint32(i))
	}

	b.ResetTimer()
	p := NewMemDB(0, 0)
	for i := range buf {
		p.Put(buf[i][:], nil)
	}
}

func BenchmarkPutRandom(b *testing.B) {
	buf := make([][4]byte, b.N)
	for i := range buf {
		binary.LittleEndian.PutUint32(buf[i][:], uint32(rand.Int()))
	}

	b.ResetTimer()
	p := NewMemDB(0, 0)
	for i := range buf {
		p.Put(buf[i][:], nil)
	}
}

func BenchmarkGet(b *testing.B) {
	buf := make([][4]byte, b.N)
	for i := range buf {
		binary.LittleEndian.PutUint32(buf[i][:], uint32(i))
	}

	p := NewMemDB(0, 0)
	for i := range buf {
		p.Put(buf[i][:], nil)
	}

	b.ResetTimer()
	for i := range buf {
		p.Get(buf[i][:])
	}
}

func BenchmarkGetRandom(b *testing.B) {
	buf := make([][4]byte, b.N)
	for i := range buf {
		binary.LittleEndian.PutUint32(buf[i][:], uint32(i))
	}

	p := NewMemDB(0, 0)
	for i := range buf {
		p.Put(buf[i][:], nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Get(buf[rand.Int()%b.N][:])
	}
}
