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
	"bytes"
	"math/big"
	"testing"

	vtypes "github.com/ontio/ontology/vm/neovm/types"
)

func TestOpCat(t *testing.T) {
	var e ExecutionEngine
	stack := NewRandAccessStack()
	stack.Push(NewStackItem(vtypes.NewByteArray([]byte("aaa"))))
	stack.Push(NewStackItem(vtypes.NewByteArray([]byte("bbb"))))
	e.EvaluationStack = stack

	opCat(&e)
	if Count(&e) != 1 || !bytes.Equal(PeekNByteArray(0, &e), []byte("aaabbb")) {
		t.Fatalf("NeoVM OpCat test failed, expect aaabbb, got %s.", string(PeekNByteArray(0, &e)))
	}
}

func TestOpSubStr(t *testing.T) {
	var e ExecutionEngine
	stack := NewRandAccessStack()
	stack.Push(NewStackItem(vtypes.NewByteArray([]byte("12345"))))
	stack.Push(NewStackItem(vtypes.NewInteger(big.NewInt(1))))
	stack.Push(NewStackItem(vtypes.NewInteger(big.NewInt(4))))
	e.EvaluationStack = stack

	opSubStr(&e)
	if !bytes.Equal(PeekNByteArray(0, &e), []byte("2345")) {
		t.Fatalf("NeoVM OpSubStr test failed, expect 234, got %s.", string(PeekNByteArray(0, &e)))
	}
}

func TestOpLeft(t *testing.T) {
	var e ExecutionEngine
	stack := NewRandAccessStack()
	stack.Push(NewStackItem(vtypes.NewByteArray([]byte("12345"))))
	stack.Push(NewStackItem(vtypes.NewInteger(big.NewInt(4))))
	e.EvaluationStack = stack

	opLeft(&e)
	if !bytes.Equal(PeekNByteArray(0, &e), []byte("1234")) {
		t.Fatalf("NeoVM OpLeft test failed, expect 1234, got %s.", string(PeekNByteArray(0, &e)))
	}
}

func TestOpRight(t *testing.T) {
	var e ExecutionEngine
	stack := NewRandAccessStack()
	stack.Push(NewStackItem(vtypes.NewByteArray([]byte("12345"))))
	stack.Push(NewStackItem(vtypes.NewInteger(big.NewInt(3))))
	e.EvaluationStack = stack

	opRight(&e)
	if !bytes.Equal(PeekNByteArray(0, &e), []byte("345")) {
		t.Fatalf("NeoVM OpRight test failed, expect 345, got %s.", string(PeekNByteArray(0, &e)))
	}
}

func TestOpSize(t *testing.T) {
	var e ExecutionEngine
	stack := NewRandAccessStack()
	stack.Push(NewStackItem(vtypes.NewByteArray([]byte("12345"))))
	e.EvaluationStack = stack

	opSize(&e)
	if PeekInt(&e) != 5 {
		t.Fatalf("NeoVM OpSize test failed, expect 5, got %d.", PeekInt(&e))
	}
}
