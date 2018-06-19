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

func TestOpBigInt(t *testing.T) {
	var e ExecutionEngine
	e.EvaluationStack = NewRandAccessStack()

	for _, code := range []OpCode{INC, DEC, NEGATE, ABS, PUSH0} {
		e.EvaluationStack.Push(NewStackItem(types.NewInteger(big.NewInt(-10))))
		e.OpCode = code
		opBigInt(&e)
		v, err := PopBigInt(&e)
		if err != nil {
			t.Fatal("NeoVM OpBigInt test failed.")
		}
		if code == INC && !(v.Cmp(big.NewInt(-9)) == 0) {
			t.Fatal("NeoVM OpBigInt test failed.")
		} else if code == DEC && !(v.Cmp(big.NewInt(-11)) == 0) {
			t.Fatal("NeoVM OpBigInt test failed.")
		} else if code == NEGATE && !(v.Cmp(big.NewInt(10)) == 0) {
			t.Fatal("NeoVM OpBigInt test failed.")
		} else if code == ABS && !(v.Cmp(big.NewInt(10)) == 0) {
			t.Fatal("NeoVM OpBigInt test failed.")
		} else if code == PUSH0 && !(v.Cmp(big.NewInt(-10)) == 0) {
			t.Fatal("NeoVM OpBigInt test failed.")
		}
	}
}

func TestOpSign(t *testing.T) {
	var e ExecutionEngine
	e.EvaluationStack = NewRandAccessStack()
	i := big.NewInt(10)
	e.EvaluationStack.Push(NewStackItem(types.NewInteger(i)))

	opSign(&e)
	v, err := PopInt(&e)
	if err != nil {
		t.Fatal("NeoVM OpSign test failed.")
	}
	if !(v == i.Sign()) {
		t.Fatal("NeoVM OpSign test failed.")
	}
}

func TestOpNot(t *testing.T) {
	var e ExecutionEngine
	e.EvaluationStack = NewRandAccessStack()
	e.EvaluationStack.Push(NewStackItem(types.NewBoolean(true)))

	opNot(&e)
	v, err := PopBoolean(&e)
	if err != nil {
		t.Fatal("NeoVM OpNot test failed.")
	}
	if !(v == false) {
		t.Fatal("NeoVM OpNot test failed.")
	}
}

func TestOpNz(t *testing.T) {
	var e ExecutionEngine
	e.EvaluationStack = NewRandAccessStack()

	e.EvaluationStack.Push(NewStackItem(types.NewInteger(big.NewInt(0))))
	e.OpCode = NZ
	opNz(&e)
	v, err := PopBoolean(&e)
	if err != nil {
		t.Fatal("NeoVM OpNz test failed.")
	}
	if v == true {
		t.Fatal("NeoVM OpNz test failed.")
	}
	e.EvaluationStack.Push(NewStackItem(types.NewInteger(big.NewInt(10))))
	opNz(&e)

	v, err = PopBoolean(&e)
	if err != nil {
		t.Fatal("NeoVM OpNz test failed.")
	}
	if v == false {
		t.Fatal("NeoVM OpNz test failed.")
	}
	e.EvaluationStack.Push(NewStackItem(types.NewInteger(big.NewInt(0))))
	e.OpCode = PUSH0
	opNz(&e)

	v, err = PopBoolean(&e)
	if err != nil {
		t.Fatal("NeoVM OpNz test failed.")
	}
	if v == true {
		t.Fatal("NeoVM OpNz test failed.")
	}
}
