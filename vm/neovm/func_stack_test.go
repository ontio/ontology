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
	ret := e.EvaluationStack.Pop().GetBigInteger().Int64()

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
	alt := e.AltStack.Peek(0).GetBigInteger().Int64()
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
	eval := e.EvaluationStack.Peek(0).GetBigInteger().Int64()

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
	e1 := stack.Peek(0).GetBigInteger().Int64()
	e2 := stack.Peek(1).GetBigInteger().Int64()

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
	e1 := stack.Peek(0).GetBigInteger().Int64()
	e2 := stack.Peek(1).GetBigInteger().Int64()

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
	e1 := stack.Peek(0).GetBigInteger().Int64()
	e2 := stack.Peek(2).GetBigInteger().Int64()

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
	if e.EvaluationStack.Count() != 3 || PeekBigInteger(&e).Int64() != 2 {
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
	e1 := stack.Peek(0).GetBigInteger().Int64()
	e2 := stack.Peek(1).GetBigInteger().Int64()

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
	e1 := stack.Peek(0).GetBigInteger().Int64()

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
	e1 := stack.Peek(0).GetBigInteger().Int64()
	e2 := stack.Peek(1).GetBigInteger().Int64()

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
	e1 := stack.Peek(0).GetBigInteger().Int64()
	e2 := stack.Peek(1).GetBigInteger().Int64()

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
	e1 := stack.Peek(0).GetBigInteger().Int64()
	e2 := stack.Peek(1).GetBigInteger().Int64()
	e3 := stack.Peek(2).GetBigInteger().Int64()

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
	e1 := stack.Peek(0).GetBigInteger().Int64()
	e2 := stack.Peek(1).GetBigInteger().Int64()

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
	e1 := stack.Peek(0).GetBigInteger().Int64()
	e2 := stack.Peek(1).GetBigInteger().Int64()
	e3 := stack.Peek(2).GetBigInteger().Int64()

	if stack.Count() != 3 || e1 != 8888 || e2 != 9999 || e3 != 8888 {
		t.Fatal("NeoVM OpTuck test failed.")
	}
}
