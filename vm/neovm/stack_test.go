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

	"math/big"

	"github.com/ontio/ontology/vm/neovm/types"
)

func TestRandomAccessStack_Count(t *testing.T) {
	r := NewRandAccessStack()
	r.Push(types.NewInteger(big.NewInt(9999)))
	r.Push(types.NewInteger(big.NewInt(8888)))

	if r.Count() != 2 {
		t.Fatalf("stack count test failed: expected 2, got %d ", r.Count())
	}
}

func TestRandomAccessStack_Pop(t *testing.T) {
	r := NewRandAccessStack()
	r.Push(types.NewInteger(big.NewInt(9999)))
	r.Push(types.NewInteger(big.NewInt(8888)))

	ret := r.Remove(0)
	ret.GetBigInteger()
	if ret.GetBigInteger().Int64() != 8888 {
		t.Fatalf("stack remove test failed: expect aaaa, got %d.", ret.GetBigInteger().Int64())
	}
}

func TestRandomAccessStack_Swap(t *testing.T) {
	r := NewRandAccessStack()
	r.Push(types.NewInteger(big.NewInt(9999)))
	r.Push(types.NewInteger(big.NewInt(8888)))
	r.Push(types.NewInteger(big.NewInt(7777)))

	r.Swap(0, 2)

	e0 := r.Pop().GetBigInteger().Int64()
	r.Pop()
	e2 := r.Pop().GetBigInteger().Int64()

	if e0 != 9999 || e2 != 7777 {
		t.Fatal("stack swap test failed.")
	}
}

func TestRandomAccessStack_Peek(t *testing.T) {
	r := NewRandAccessStack()
	r.Push(types.NewInteger(big.NewInt(9999)))
	r.Push(types.NewInteger(big.NewInt(8888)))

	e0 := r.Peek(0).GetBigInteger().Int64()
	e1 := r.Peek(1).GetBigInteger().Int64()

	if e0 != 8888 || e1 != 9999 {
		t.Fatal("stack peek test failed.")
	}
}
