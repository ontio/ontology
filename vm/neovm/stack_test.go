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
	//ret.GetBigInteger()

	v, err := ret.GetBigInteger()
	if err != nil {
		t.Fatal("NeoVM stack pop test failed.")
	}

	if v.Int64() != 8888 {
		t.Fatalf("stack pop test failed: expect 8888, got %d.", v.Int64())
	}
}

func TestRandomAccessStack_Swap(t *testing.T) {
	r := NewRandAccessStack()
	r.Push(types.NewInteger(big.NewInt(9999)))
	r.Push(types.NewInteger(big.NewInt(8888)))
	r.Push(types.NewInteger(big.NewInt(7777)))

	r.Swap(0, 2)
	v0, err := r.Pop().GetBigInteger()
	if err != nil {
		t.Fatal("NeoVM stack swap test failed.")
	}
	e0 := v0.Int64()
	r.Pop()
	v2, err := r.Pop().GetBigInteger()
	if err != nil {
		t.Fatal("NeoVM stack swap test failed.")
	}
	e2 := v2.Int64()

	if e0 != 9999 || e2 != 7777 {
		t.Fatal("stack swap test failed.")
	}
}

func TestRandomAccessStack_Peek(t *testing.T) {
	r := NewRandAccessStack()
	r.Push(types.NewInteger(big.NewInt(9999)))
	r.Push(types.NewInteger(big.NewInt(8888)))

	v0, err := r.Peek(0).GetBigInteger()
	if err != nil {
		t.Fatal("NeoVM stack peek test failed.")
	}
	v1, err := r.Peek(1).GetBigInteger()
	if err != nil {
		t.Fatal("NeoVM stack peek test failed.")
	}

	e0 := v0.Int64()
	e1 := v1.Int64()

	if e0 != 8888 || e1 != 9999 {
		t.Fatal("stack peek test failed.")
	}
}

func TestRandomAccessStack_CopyTo(t *testing.T) {
	r := NewRandAccessStack()
	r.Push(types.NewInteger(big.NewInt(9999)))
	r.Push(types.NewInteger(big.NewInt(8888)))

	e := NewRandAccessStack()
	r.CopyTo(e)

	for k, v := range r.e {
		if !v.Equals(e.e[k]) {
			t.Fatal("stack copyto test failed.")
		}
	}
}
