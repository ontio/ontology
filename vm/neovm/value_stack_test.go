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
	"math/big"
	"testing"

	"github.com/ontio/ontology/vm/neovm/types"
	"github.com/stretchr/testify/assert"
)

func TestValueStack_Count(t *testing.T) {
	r := NewValueStack(4)
	val, err := types.VmValueFromBigInt(big.NewInt(9999))
	assert.Equal(t, err, nil)
	err = r.Push(val)
	assert.Equal(t, err, nil)
	val2, err := types.VmValueFromBigInt(big.NewInt(8888))
	assert.Equal(t, err, nil)
	r.Push(val2)
	assert.Equal(t, r.Count(), 2)

}

func TestValueStack_Pop(t *testing.T) {
	r := NewValueStack(4)
	val, err := types.VmValueFromBigInt(big.NewInt(9999))
	assert.Equal(t, err, nil)
	err = r.Push(val)
	assert.Equal(t, err, nil)
	val2, err := types.VmValueFromBigInt(big.NewInt(8888))
	assert.Equal(t, err, nil)
	r.Push(val2)
	val, err = r.Remove(0)
	assert.Equal(t, err, nil)
	i, err := val.AsInt64()
	assert.Equal(t, err, nil)
	assert.Equal(t, i, int64(8888))

}

func TestValueStack_Peek(t *testing.T) {
	r := NewValueStack(4)
	val, err := types.VmValueFromBigInt(big.NewInt(9999))
	assert.Equal(t, err, nil)
	err = r.Push(val)
	assert.Equal(t, err, nil)
	val2, err := types.VmValueFromBigInt(big.NewInt(8888))
	assert.Equal(t, err, nil)
	r.Push(val2)
	val, err = r.Peek(0)
	assert.Equal(t, err, nil)
	val2, err = r.Peek(1)
	assert.Equal(t, err, nil)

	v, err := val.AsInt64()
	assert.Equal(t, err, nil)
	v2, err := val2.AsInt64()
	assert.Equal(t, err, nil)
	if v != int64(8888) || v2 != int64(9999) {
		t.Fatal("stack peek test failed.")
	}
}

func TestValueStack_Swap(t *testing.T) {
	r := NewValueStack(4)
	val, err := types.VmValueFromBigInt(big.NewInt(9999))
	assert.Equal(t, err, nil)
	err = r.Push(val)
	assert.Equal(t, err, nil)
	val2, err := types.VmValueFromBigInt(big.NewInt(8888))
	assert.Equal(t, err, nil)
	r.Push(val2)
	val3, err := types.VmValueFromBigInt(big.NewInt(7777))
	assert.Equal(t, err, nil)
	r.Push(val3)

	r.Swap(0, 2)

	val, err = r.Pop()
	assert.Equal(t, err, nil)
	v0, err := val.AsInt64()
	assert.Equal(t, err, nil)
	r.Pop()
	val2, err = r.Pop()
	assert.Equal(t, err, nil)
	v2, err := val2.AsInt64()
	assert.Equal(t, err, nil)
	assert.Equal(t, v0, int64(9999))
	assert.Equal(t, v2, int64(7777))
}

func TestValueStack_CopyTo(t *testing.T) {
	r := NewValueStack(0)
	val, err := types.VmValueFromBigInt(big.NewInt(9999))
	assert.Equal(t, err, nil)
	r.Push(val)
	val2, err := types.VmValueFromBigInt(big.NewInt(8888))
	assert.Equal(t, err, nil)
	r.Push(val2)

	e := NewValueStack(0)
	err = r.CopyTo(e)
	assert.Equal(t, err, nil)

	for k, v := range r.data {
		if !v.Equals(e.data[k]) {
			t.Fatal("stack copyto test failed.")
		}
	}
}
