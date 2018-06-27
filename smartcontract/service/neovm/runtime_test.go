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

	assert.False(t, CircularRefAndDepthDetection(a))
	assert.False(t, CircularRefAndDepthDetection(b))

	a.Add(b)
	assert.True(t, CircularRefAndDepthDetection(a))
	assert.True(t, CircularRefAndDepthDetection(b))
}

func TestStructRef(t *testing.T) {
	ba1 := types.NewByteArray([]byte{1, 2, 3})
	ba2 := types.NewByteArray([]byte{4, 5, 6})
	bf := types.NewBoolean(false)
	bt := types.NewBoolean(true)

	assert.False(t, CircularRefAndDepthDetection(ba1))
	assert.False(t, CircularRefAndDepthDetection(ba2))
	assert.False(t, CircularRefAndDepthDetection(bf))
	assert.False(t, CircularRefAndDepthDetection(bt))

	array1 := types.NewArray([]types.StackItems{ba1, bf})
	array2 := types.NewArray([]types.StackItems{ba2, bf})
	struc := types.NewStruct([]types.StackItems{ba1, bf})

	assert.False(t, CircularRefAndDepthDetection(struc))
	assert.False(t, CircularRefAndDepthDetection(array1))
	assert.False(t, CircularRefAndDepthDetection(array2))

	array1.Add(struc)
	assert.False(t, CircularRefAndDepthDetection(array1))

	struc.Add(array2)
	assert.False(t, CircularRefAndDepthDetection(array1))

	array2.Add(array1)
	assert.True(t, CircularRefAndDepthDetection(array1))
	assert.True(t, CircularRefAndDepthDetection(struc))
	assert.True(t, CircularRefAndDepthDetection(array2))

	map1 := types.NewMap()
	assert.False(t, CircularRefAndDepthDetection(map1))
	map1.Add(array1, bf)
	assert.True(t, CircularRefAndDepthDetection(map1))

	map2 := types.NewMap()
	map2.Add(bf, array2)
	assert.True(t, CircularRefAndDepthDetection(map2))

	map3 := types.NewMap()
	map4 := types.NewMap()
	map5 := types.NewMap()
	map3.Add(map4, map5)
	map3.Add(map5, map4)

	assert.False(t, CircularRefAndDepthDetection(map3))

	map6 := types.NewMap()
	map7 := types.NewMap()
	map8 := types.NewMap()
	map6.Add(bf, bf)
	map7.Add(bt, bf)
	map8.Add(bf, map6)
	map8.Add(bt, map7)
	map8.Add(ba1, map7)

	assert.False(t, CircularRefAndDepthDetection(map8))
}
