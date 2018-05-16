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

	vtypes "github.com/ontio/ontology/vm/neovm/types"
)

func TestOpInvert(t *testing.T) {
	var e ExecutionEngine
	stack := NewRandAccessStack()
	stack.Push(NewStackItem(vtypes.NewInteger(big.NewInt(123456789))))
	e.EvaluationStack = stack

	opInvert(&e)
	i := big.NewInt(123456789)

	if PeekBigInteger(&e).Cmp(i.Not(i)) != 0 {
		t.Fatal("NeoVM OpInvert test failed.")
	}
}

func TestOpEqual(t *testing.T) {
	var e ExecutionEngine
	stack := NewRandAccessStack()
	stack.Push(NewStackItem(vtypes.NewInteger(big.NewInt(123456789))))
	stack.Push(NewStackItem(vtypes.NewInteger(big.NewInt(123456789))))
	e.EvaluationStack = stack

	opEqual(&e)
	if !PopBoolean(&e) {
		t.Fatal("NeoVM OpEqual test failed.")
	}
}
