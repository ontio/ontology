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
package common

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/ontio/ontology/common/serialization"
	"github.com/stretchr/testify/assert"
)

func BenchmarkZeroCopySource(b *testing.B) {
	const N = 12000
	buf := make([]byte, N)
	rand.Read(buf)

	for i := 0; i < b.N; i++ {
		source := NewZeroCopySource(buf)
		for j := 0; j < N/100; j++ {
			source.NextUint16()
			source.NextByte()
			source.NextUint64()
			source.NextVarUint()
			source.NextBytes(20)
		}
	}

}

func BenchmarkDerserialize(b *testing.B) {
	const N = 12000
	buf := make([]byte, N)
	rand.Read(buf)

	for i := 0; i < b.N; i++ {
		reader := bytes.NewBuffer(buf)
		for j := 0; j < N/100; j++ {
			serialization.ReadUint16(reader)
			serialization.ReadByte(reader)
			serialization.ReadUint64(reader)
			serialization.ReadVarUint(reader, 0)
			serialization.ReadBytes(reader, 20)
		}
	}

}

func TestReadFromNil(t *testing.T) {
	s := NewZeroCopySource(nil)
	_, _, _, eof := s.NextString()
	assert.True(t, eof)
}

func TestReadVarInt(t *testing.T) {
	s := NewZeroCopySource([]byte{0xfd})
	_, _, _, eof := s.NextString()
	assert.True(t, eof)
}
