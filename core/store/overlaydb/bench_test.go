// Copyright (c) 2012, Suryandaru Triandana <syndtr@gmail.com>
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package overlaydb

import (
	"encoding/binary"
	"math/rand"
	"testing"

	"fmt"
	"github.com/stretchr/testify/assert"
)

func TestDot(t *testing.T) {
	buf := make(map[string]string)
	for i := 0; i < 100; i++ {
		k := fmt.Sprintf("k%d", i)
		v := fmt.Sprintf("v%d", i)
		buf[k] = v
	}
	db := NewMemDB(10)
	for k, v := range buf {
		db.Put([]byte(k), []byte(v))
	}

	dot := db.DumpToDot()

	fmt.Println(dot)
}

func TestDB(t *testing.T) {
	buf := make([][1]byte, 10)
	for i := range buf {
		buf[i][0] = byte(i)
	}

	db := NewMemDB(10)
	for i := range buf {
		db.Put(buf[i][:], nil)
	}

	for i := range buf {
		val, err := db.Get(buf[i][:])
		assert.Nil(t, err)
		assert.Equal(t, len(val), 0)
	}

	fmt.Println(db.nodeData)
}

func BenchmarkPut(b *testing.B) {
	buf := make([][4]byte, b.N)
	for i := range buf {
		binary.LittleEndian.PutUint32(buf[i][:], uint32(i))
	}

	b.ResetTimer()
	p := NewMemDB(0)
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
	p := NewMemDB(0)
	for i := range buf {
		p.Put(buf[i][:], nil)
	}
}

func BenchmarkGet(b *testing.B) {
	buf := make([][4]byte, b.N)
	for i := range buf {
		binary.LittleEndian.PutUint32(buf[i][:], uint32(i))
	}

	p := NewMemDB(0)
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

	p := NewMemDB(0)
	for i := range buf {
		p.Put(buf[i][:], nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Get(buf[rand.Int()%b.N][:])
	}
}
