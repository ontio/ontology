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

package types

import (
	"github.com/ontio/ontology/common"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestBoolean_Equals(t *testing.T) {
	a := NewBoolean(false)
	b := NewBoolean(true)
	c := NewBoolean(true)
	d := NewBoolean(false)

	assert.True(t, a.Equals(d))
	assert.True(t, a.Equals(a))
	assert.True(t, b.Equals(c))
	assert.True(t, b.Equals(b))

	assert.False(t, a.Equals(b))
	assert.False(t, b.Equals(d))
}

func TestArray_Equals(t *testing.T) {
	a := NewArray(nil)
	b := NewArray(nil)

	assert.False(t, a.Equals(b))
	assert.True(t, a.Equals(a))

	i := NewInteger(big.NewInt(int64(0)))
	j := NewInteger(big.NewInt(int64(0)))
	k := NewInteger(big.NewInt(int64(1)))

	m1 := NewArray([]StackItems{i, j, k})
	m2 := NewArray([]StackItems{i, j, k})

	assert.False(t, m1.Equals(m2))
	assert.True(t, m1.Equals(m1))

}

func TestInteger_Equals(t *testing.T) {
	i := NewInteger(big.NewInt(int64(0)))
	j := NewInteger(big.NewInt(int64(0)))

	assert.True(t, i.Equals(j))
	k := NewInteger(big.NewInt(int64(100000)))
	assert.False(t, i.Equals(k))
}

func TestNewByteArray(t *testing.T) {
	i := NewByteArray([]byte("abcde"))
	j := NewByteArray([]byte{'a', 'b', 'c', 'd', 'e'})

	assert.True(t, i.Equals(j))

	k := NewByteArray(nil)
	assert.True(t, k.Equals(NewByteArray(nil)))
}

func TestStruct_Equals(t *testing.T) {

	a := NewStruct(nil)
	b := NewStruct(nil)

	assert.True(t, a.Equals(b))
	assert.True(t, a.Equals(a))

	i := NewInteger(big.NewInt(int64(0)))
	j := NewInteger(big.NewInt(int64(0)))
	k := NewInteger(big.NewInt(int64(1)))

	m1 := NewStruct([]StackItems{i, j, k})
	m2 := NewStruct([]StackItems{i, j, k})

	assert.True(t, m1.Equals(m2))
	assert.True(t, m1.Equals(m1))

}

func TestMap_Equals(t *testing.T) {
	a := NewMap()
	b := NewMap()

	assert.False(t, a.Equals(b))
	assert.True(t, a.Equals(a))

	k1 := NewInteger(big.NewInt(int64(0)))
	k2 := NewInteger(big.NewInt(int64(0)))

	v1 := NewByteArray([]byte("abcde"))
	v2 := NewByteArray([]byte{'a', 'b', 'c', 'd', 'e'})

	a.Add(k1, v1)
	b.Add(k2, v2)

	assert.False(t, a.Equals(b))
	assert.True(t, b.Equals(b))

}

func TestInterop_Equals(t *testing.T) {
	a := NewInteropInterface(nil)
	b := NewInteropInterface(nil)

	assert.False(t, a.Equals(b))

}

func TestCmp(t *testing.T) {
	a := NewBoolean(false)
	b := NewInteger(big.NewInt(0))
	c := NewBoolean(true)
	d := NewInteger(big.NewInt(1))
	assert.False(t, a.Equals(b)) //????
	assert.True(t, c.Equals(d))  //????

	arr := NewArray(nil)
	stt := NewStruct(nil)
	assert.False(t, arr.Equals(stt))
	arr.Add(NewInteger(big.NewInt(0)))
	stt.Add(NewInteger(big.NewInt(0)))
	assert.False(t, arr.Equals(stt))

	ba := NewByteArray(common.BigIntToNeoBytes(big.NewInt(0)))
	assert.True(t, ba.Equals(b))

	k := NewByteArray(nil)
	assert.False(t, c.Equals(k))

}
