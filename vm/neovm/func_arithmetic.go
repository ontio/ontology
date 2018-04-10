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

func opBigInt(e *ExecutionEngine) (VMState, error) {
	x := PopBigInt(e)
	PushData(e, BigIntOp(x, e.opCode))
	return NONE, nil
}

func opSign(e *ExecutionEngine) (VMState, error) {
	x := PopBigInt(e)
	PushData(e, x.Sign())
	return NONE, nil
}

func opNot(e *ExecutionEngine) (VMState, error) {
	x := PopBoolean(e)
	PushData(e, !x)
	return NONE, nil
}

func opNz(e *ExecutionEngine) (VMState, error) {
	x := PopBigInt(e)
	PushData(e, BigIntComp(x, e.opCode))
	return NONE, nil
}

func opBigIntZip(e *ExecutionEngine) (VMState, error) {
	x2 := PopBigInt(e)
	x1 := PopBigInt(e)
	b := BigIntZip(x1, x2, e.opCode)
	PushData(e, b)
	return NONE, nil
}

func opBoolZip(e *ExecutionEngine) (VMState, error) {
	x2 := PopBoolean(e)
	x1 := PopBoolean(e)
	PushData(e, BoolZip(x1, x2, e.opCode))
	return NONE, nil
}

func opBigIntComp(e *ExecutionEngine) (VMState, error) {
	x2 := PopBigInt(e)
	x1 := PopBigInt(e)
	PushData(e, BigIntMultiComp(x1, x2, e.opCode))
	return NONE, nil
}

func opWithIn(e *ExecutionEngine) (VMState, error) {
	b := PopBigInt(e)
	a := PopBigInt(e)
	c := PopBigInt(e)
	PushData(e, WithInOp(c, a, b))
	return NONE, nil
}
