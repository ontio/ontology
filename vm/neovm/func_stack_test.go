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

func TestOpToDupFromAltStack(t *testing.T) {
	var e ExecutionEngine
	e.EvaluationStack = NewRandAccessStack()
	e.AltStack = NewRandAccessStack()
	e.AltStack.Push(types.NewInteger(big.NewInt(9999)))

	opToDupFromAltStack(&e)
	v, err := e.EvaluationStack.Pop().GetBigInteger()
	if err != nil {
		t.Fatal("NeoVM opToDupFromAltStack test failed.")
	}
	ret := v.Int64()

	if ret != 9999 {
		t.Fatal("NeoVM opToDupFromAltStack test failed.")
	}
}

func TestOpToAltStack(t *testing.T) {
	var e ExecutionEngine
	e.EvaluationStack = NewRandAccessStack()
	e.AltStack = NewRandAccessStack()
	//e.EvaluationStack.Push(NewElementImpl("aaa"))
	e.EvaluationStack.Push(types.NewInteger(big.NewInt(9999)))

	opToAltStack(&e)
	v, err := e.AltStack.Peek(0).GetBigInteger()
	if err != nil {
		t.Fatal("NeoVM opToDupFromAltStack test failed.")
	}
	alt := v.Int64()
	eval := e.EvaluationStack.Peek(0)

	if eval != nil || alt != 9999 {
		t.Fatal("NeoVM opToAltStack test failed.")
	}
}

func TestOpFromAltStack(t *testing.T) {
	var e ExecutionEngine
	e.EvaluationStack = NewRandAccessStack()
	e.AltStack = NewRandAccessStack()
	e.AltStack.Push(types.NewInteger(big.NewInt(9999)))

	opFromAltStack(&e)
	alt := e.AltStack.Peek(0)
	v, err := e.EvaluationStack.Peek(0).GetBigInteger()
	if err != nil {
		t.Fatal("NeoVM opToDupFromAltStack test failed.")
	}
	eval := v.Int64()

	if alt != nil || eval != 9999 {
		t.Fatal("NeoVM opFromAltStack test failed.")
	}
}

func TestOpXDrop(t *testing.T) {
	var e ExecutionEngine
	stack := NewRandAccessStack()
	stack.Push(types.NewInteger(big.NewInt(9999)))
	stack.Push(types.NewInteger(big.NewInt(8888)))
	stack.Push(types.NewInteger(big.NewInt(7777)))
	stack.Push(NewStackItem(types.NewInteger(big.NewInt(1))))
	e.EvaluationStack = stack

	opXDrop(&e)
	v1, err := stack.Peek(0).GetBigInteger()
	if err != nil {
		t.Fatal("NeoVM OpXDrop test failed.")
	}
	v2, err := stack.Peek(1).GetBigInteger()
	if err != nil {
		t.Fatal("NeoVM OpXDrop test failed.")
	}
	e1 := v1.Int64()
	e2 := v2.Int64()

	if stack.Count() != 2 || e1 != 7777 || e2 != 9999 {
		t.Fatal("NeoVM OpXDrop test failed.")
	}
}

func TestOpXSwap(t *testing.T) {
	var e ExecutionEngine
	stack := NewRandAccessStack()
	stack.Push(types.NewInteger(big.NewInt(9999)))
	stack.Push(types.NewInteger(big.NewInt(8888)))
	stack.Push(types.NewInteger(big.NewInt(7777)))
	stack.Push(NewStackItem(types.NewInteger(big.NewInt(1))))
	e.EvaluationStack = stack

	opXSwap(&e)
	v1, err := stack.Peek(0).GetBigInteger()
	if err != nil {
		t.Fatal("NeoVM OpXSwap test failed.")
	}
	v2, err := stack.Peek(1).GetBigInteger()
	if err != nil {
		t.Fatal("NeoVM OpXSwap test failed.")
	}
	e1 := v1.Int64()
	e2 := v2.Int64()

	if stack.Count() != 3 || e1 != 8888 || e2 != 7777 {
		t.Fatal("NeoVM OpXSwap test failed.")
	}
}

func TestOpXTuck(t *testing.T) {
	var e ExecutionEngine
	stack := NewRandAccessStack()
	stack.Push(types.NewInteger(big.NewInt(9999)))
	stack.Push(types.NewInteger(big.NewInt(8888)))
	stack.Push(types.NewInteger(big.NewInt(7777)))

	stack.Push(NewStackItem(types.NewInteger(big.NewInt(2))))
	e.EvaluationStack = stack

	opXSwap(&e)
	v1, err := stack.Peek(0).GetBigInteger()
	if err != nil {
		t.Fatal("NeoVM OpXTuck test failed.")
	}
	v2, err := stack.Peek(2).GetBigInteger()
	if err != nil {
		t.Fatal("NeoVM OpXTuck test failed.")
	}
	e1 := v1.Int64()
	e2 := v2.Int64()

	if stack.Count() != 3 || e1 != 9999 || e2 != 7777 {
		t.Fatal("NeoVM OpXTuck test failed.")
	}
}

func TestOpDepth(t *testing.T) {
	var e ExecutionEngine
	stack := NewRandAccessStack()
	stack.Push(types.NewInteger(big.NewInt(9999)))
	stack.Push(types.NewInteger(big.NewInt(8888)))
	e.EvaluationStack = stack

	opDepth(&e)
	v, err := PeekBigInteger(&e)
	if err != nil {
		t.Fatal("NeoVM OpDepth test failed.")
	}
	if e.EvaluationStack.Count() != 3 || v.Int64() != 2 {
		t.Fatal("NeoVM OpDepth test failed.")
	}
}

func TestOpDrop(t *testing.T) {
	var e ExecutionEngine
	stack := NewRandAccessStack()
	stack.Push(types.NewInteger(big.NewInt(9999)))
	e.EvaluationStack = stack

	opDrop(&e)
	if e.EvaluationStack.Count() != 0 {
		t.Fatal("NeoVM OpDrop test failed.")
	}
}

func TestOpDup(t *testing.T) {
	var e ExecutionEngine
	stack := NewRandAccessStack()
	stack.Push(types.NewInteger(big.NewInt(9999)))
	e.EvaluationStack = stack

	opDup(&e)
	v1, err := stack.Peek(0).GetBigInteger()
	if err != nil {
		t.Fatal("NeoVM OpDup test failed.")
	}
	v2, err := stack.Peek(1).GetBigInteger()
	if err != nil {
		t.Fatal("NeoVM OpDup test failed.")
	}
	e1 := v1.Int64()
	e2 := v2.Int64()

	if stack.Count() != 2 || e1 != 9999 || e2 != 9999 {
		t.Fatal("NeoVM OpDup test failed.")
	}
}

func TestOpNip(t *testing.T) {
	var e ExecutionEngine
	stack := NewRandAccessStack()
	stack.Push(types.NewInteger(big.NewInt(9999)))
	stack.Push(types.NewInteger(big.NewInt(8888)))
	e.EvaluationStack = stack

	opNip(&e)
	v, err := stack.Peek(0).GetBigInteger()
	if err != nil {
		t.Fatal("NeoVM OpNip test failed.")
	}
	e1 := v.Int64()

	if stack.Count() != 1 || e1 != 8888 {
		t.Fatal("NeoVM OpNip test failed.")
	}
}

func TestOpOver(t *testing.T) {
	var e ExecutionEngine
	stack := NewRandAccessStack()
	stack.Push(types.NewInteger(big.NewInt(9999)))
	stack.Push(types.NewInteger(big.NewInt(8888)))
	e.EvaluationStack = stack

	opOver(&e)
	v1, err := stack.Peek(0).GetBigInteger()
	if err != nil {
		t.Fatal("NeoVM OpOver test failed.")
	}
	v2, err := stack.Peek(1).GetBigInteger()
	if err != nil {
		t.Fatal("NeoVM OpOver test failed.")
	}
	e1 := v1.Int64()
	e2 := v2.Int64()

	if stack.Count() != 3 || e1 != 9999 || e2 != 8888 {
		t.Fatal("NeoVM OpOver test failed.")
	}
}

func TestOpPick(t *testing.T) {
	var e ExecutionEngine
	stack := NewRandAccessStack()
	stack.Push(types.NewInteger(big.NewInt(9999)))
	stack.Push(types.NewInteger(big.NewInt(8888)))
	stack.Push(types.NewInteger(big.NewInt(7777)))
	stack.Push(types.NewInteger(big.NewInt(6666)))

	stack.Push(NewStackItem(types.NewInteger(big.NewInt(3))))
	e.EvaluationStack = stack

	opPick(&e)
	v1, err := stack.Peek(0).GetBigInteger()
	if err != nil {
		t.Fatal("NeoVM OpPick test failed.")
	}
	v2, err := stack.Peek(1).GetBigInteger()
	if err != nil {
		t.Fatal("NeoVM OpPick test failed.")
	}
	e1 := v1.Int64()
	e2 := v2.Int64()

	if stack.Count() != 5 || e1 != 9999 || e2 != 6666 {
		t.Fatal("NeoVM OpPick test failed.")
	}
}

func TestOpRot(t *testing.T) {
	var e ExecutionEngine
	stack := NewRandAccessStack()
	stack.Push(types.NewInteger(big.NewInt(9999)))
	stack.Push(types.NewInteger(big.NewInt(8888)))
	stack.Push(types.NewInteger(big.NewInt(7777)))
	e.EvaluationStack = stack

	opRot(&e)
	v1, err := stack.Peek(0).GetBigInteger()
	if err != nil {
		t.Fatal("NeoVM OpRot test failed.")
	}
	v2, err := stack.Peek(1).GetBigInteger()
	if err != nil {
		t.Fatal("NeoVM OpRot test failed.")
	}
	v3, err := stack.Peek(2).GetBigInteger()
	if err != nil {
		t.Fatal("NeoVM OpRot test failed.")
	}
	e1 := v1.Int64()
	e2 := v2.Int64()
	e3 := v3.Int64()

	if stack.Count() != 3 || e1 != 9999 || e2 != 7777 || e3 != 8888 {
		t.Fatal("NeoVM OpRot test failed.")
	}
}

func TestOpSwap(t *testing.T) {
	var e ExecutionEngine
	stack := NewRandAccessStack()
	stack.Push(types.NewInteger(big.NewInt(9999)))
	stack.Push(types.NewInteger(big.NewInt(8888)))
	e.EvaluationStack = stack

	opSwap(&e)
	v1, err := stack.Peek(0).GetBigInteger()
	if err != nil {
		t.Fatal("NeoVM OpSwap test failed.")
	}
	v2, err := stack.Peek(1).GetBigInteger()
	if err != nil {
		t.Fatal("NeoVM OpSwap test failed.")
	}
	e1 := v1.Int64()
	e2 := v2.Int64()

	if stack.Count() != 2 || e1 != 9999 || e2 != 8888 {
		t.Fatal("NeoVM OpSwap test failed.")
	}
}

func TestOpTuck(t *testing.T) {
	var e ExecutionEngine
	stack := NewRandAccessStack()
	stack.Push(types.NewInteger(big.NewInt(9999)))
	stack.Push(types.NewInteger(big.NewInt(8888)))
	e.EvaluationStack = stack

	opTuck(&e)
	v1, err := stack.Peek(0).GetBigInteger()
	if err != nil {
		t.Fatal("NeoVM OpTuck test failed.")
	}
	v2, err := stack.Peek(1).GetBigInteger()
	if err != nil {
		t.Fatal("NeoVM OpTuck test failed.")
	}
	v3, err := stack.Peek(2).GetBigInteger()
	if err != nil {
		t.Fatal("NeoVM OpTuck test failed.")
	}
	e1 := v1.Int64()
	e2 := v2.Int64()
	e3 := v3.Int64()

	if stack.Count() != 3 || e1 != 8888 || e2 != 9999 || e3 != 8888 {
		t.Fatal("NeoVM OpTuck test failed.")
	}
}
