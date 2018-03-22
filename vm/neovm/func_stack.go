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

func opToDupFromAltStack(e *ExecutionEngine) (VMState, error) {
	Push(e, e.altStack.Peek(0))
	return NONE, nil
}

func opToAltStack(e *ExecutionEngine) (VMState, error) {
	e.altStack.Push(Pop(e))
	return NONE, nil
}

func opFromAltStack(e *ExecutionEngine) (VMState, error) {
	Push(e, e.altStack.Pop())
	return NONE, nil
}

func opXDrop(e *ExecutionEngine) (VMState, error) {
	n := PopInt(e)
	e.evaluationStack.Remove(n)
	return NONE, nil
}

func opXSwap(e *ExecutionEngine) (VMState, error) {
	n := PopInt(e)
	if n == 0 {
		return NONE, nil
	}
	e.evaluationStack.Swap(0, n)
	return NONE, nil
}

func opXTuck(e *ExecutionEngine) (VMState, error) {
	n := PopInt(e)
	e.evaluationStack.Insert(n, Peek(e))
	return NONE, nil
}

func opDepth(e *ExecutionEngine) (VMState, error) {
	PushData(e, Count(e))
	return NONE, nil
}

func opDrop(e *ExecutionEngine) (VMState, error) {
	Pop(e)
	return NONE, nil
}

func opDup(e *ExecutionEngine) (VMState, error) {
	Push(e, Peek(e))
	return NONE, nil
}

func opNip(e *ExecutionEngine) (VMState, error) {
	x2 := Pop(e)
	Pop(e)
	Push(e, x2)
	return NONE, nil
}

func opOver(e *ExecutionEngine) (VMState, error) {
	x2 := Pop(e)
	x1 := Peek(e)

	Push(e, x2)
	Push(e, x1)
	return NONE, nil
}

func opPick(e *ExecutionEngine) (VMState, error) {
	n := PopInt(e)
	Push(e, e.evaluationStack.Peek(n))
	return NONE, nil
}

func opRoll(e *ExecutionEngine) (VMState, error) {
	n := PopInt(e)
	if n == 0 {
		return NONE, nil
	}
	Push(e, e.evaluationStack.Remove(n))
	return NONE, nil
}

func opRot(e *ExecutionEngine) (VMState, error) {
	x3 := Pop(e)
	x2 := Pop(e)
	x1 := Pop(e)
	Push(e, x2)
	Push(e, x3)
	Push(e, x1)
	return NONE, nil
}

func opSwap(e *ExecutionEngine) (VMState, error) {
	x2 := Pop(e)
	x1 := Pop(e)
	Push(e, x2)
	Push(e, x1)
	return NONE, nil
}

func opTuck(e *ExecutionEngine) (VMState, error) {
	x2 := Pop(e)
	x1 := Pop(e)
	Push(e, x2)
	Push(e, x1)
	Push(e, x2)
	return NONE, nil
}

