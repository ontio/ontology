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

package neovm

import (
	"testing"

	"github.com/ontio/ontology/vm/neovm/types"
	"github.com/stretchr/testify/assert"
	"math/big"
)

func TestRuntimeSerialize(t *testing.T) {
	a := types.NewArray(nil)
	b := types.NewArray([]types.StackItems{a})
	a.Add(b)

	_, err := SerializeStackItem(a)
	assert.NotNil(t, err)
}

func TestArrayRef(t *testing.T) {
	a := types.NewArray(nil)
	b := types.NewArray([]types.StackItems{a})

	assert.False(t, CircularRefDetection(a))
	assert.False(t, CircularRefDetection(b))

	a.Add(b)
	assert.True(t, CircularRefDetection(a))
	assert.True(t, CircularRefDetection(b))
}

func TestStructRef(t *testing.T) {
	ba1 := types.NewByteArray([]byte{1, 2, 3})
	ba2 := types.NewByteArray([]byte{4, 5, 6})
	bf := types.NewBoolean(false)
	bt := types.NewBoolean(true)

	assert.False(t, CircularRefDetection(ba1))
	assert.False(t, CircularRefDetection(ba2))
	assert.False(t, CircularRefDetection(bf))
	assert.False(t, CircularRefDetection(bt))

	array1 := types.NewArray([]types.StackItems{ba1, bf})
	array2 := types.NewArray([]types.StackItems{ba2, bf})
	struc := types.NewStruct([]types.StackItems{ba1, bf})

	assert.False(t, CircularRefDetection(struc))
	assert.False(t, CircularRefDetection(array1))
	assert.False(t, CircularRefDetection(array2))

	array1.Add(struc)
	assert.False(t, CircularRefDetection(array1))

	struc.Add(array2)
	assert.False(t, CircularRefDetection(array1))

	array2.Add(array1)
	assert.True(t, CircularRefDetection(array1))
	assert.True(t, CircularRefDetection(struc))
	assert.True(t, CircularRefDetection(array2))

	map1 := types.NewMap()
	assert.False(t, CircularRefDetection(map1))
	map1.Add(array1, bf)
	assert.True(t, CircularRefDetection(map1))

	map2 := types.NewMap()
	map2.Add(bf, array2)
	assert.True(t, CircularRefDetection(map2))

	map3 := types.NewMap()
	map4 := types.NewMap()
	map5 := types.NewMap()
	map3.Add(map4, map5)
	map3.Add(map5, map4)

	assert.False(t, CircularRefDetection(map3))

	map6 := types.NewMap()
	map7 := types.NewMap()
	map8 := types.NewMap()
	map6.Add(bf, bf)
	map7.Add(bt, bf)
	map8.Add(bf, map6)
	map8.Add(bt, map7)
	map8.Add(ba1, map7)

	assert.False(t, CircularRefDetection(map8))
}

func TestNestedRef(t *testing.T) {
	i := types.NewInteger(big.NewInt(int64(0)))
	j := types.NewInteger(big.NewInt(int64(1)))
	k := types.NewInteger(big.NewInt(int64(2)))

	arr1 := types.NewArray([]types.StackItems{i, j})
	arr2 := types.NewArray([]types.StackItems{j, k})

	struc := types.NewStruct([]types.StackItems{arr1, arr2})
	assert.False(t, CircularRefDetection(struc))

	struc1 := types.NewStruct([]types.StackItems{i, j, k})
	struc2 := types.NewStruct([]types.StackItems{i, struc1})
	assert.False(t, CircularRefDetection(struc2))

	struc3 := types.NewStruct([]types.StackItems{struc1, struc2})
	assert.False(t, CircularRefDetection(struc3))

	struc4 := types.NewStruct([]types.StackItems{struc3, struc1})
	struc5 := types.NewArray([]types.StackItems{i, struc4})
	struc4.Add(struc5)

	assert.True(t, CircularRefDetection(struc4))

	//check depth
	n := VM_SERIALIZE_DEPTH
	arrNest := makeNestedArr(n, arr1)
	assert.True(t, CircularRefDetection(arrNest))

	n = VM_SERIALIZE_DEPTH - 1
	arrNest = makeNestedArr(n, arr1)
	assert.False(t, CircularRefDetection(arrNest))

}

func TestMaxLenght(t *testing.T) {

	//OOM when 	cap := 100000000000
	cap := 1024 * 1000 * 10
	b := make([]byte, cap)
	//for i:=0;i<cap;i++{
	//	b[i] = byte('a')
	//}

	s := types.NewByteArray(b)

	_, err := SerializeStackItem(s)
	assert.Error(t, err, "")

}

func TestMapLength(t *testing.T) {
	m := types.NewMap()
	key1 := types.NewByteArray([]byte("key1"))
	key2 := types.NewByteArray([]byte("key2"))

	val1 := types.NewByteArray([]byte("val1"))
	val2 := types.NewByteArray([]byte("val2"))

	m.Add(key1, val1)
	m.Add(key2, val2)

	res, _ := SerializeStackItem(m)

	assert.NotNil(t, res, "")

}

func makeNestedArr(n int, arr types.StackItems) types.StackItems {
	if n == 1 {
		return arr
	}
	tmp := types.NewArray([]types.StackItems{arr})
	return makeNestedArr(n-1, tmp)
}
