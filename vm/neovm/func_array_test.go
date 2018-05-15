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

	"bytes"
	"math/big"

	vtypes "github.com/ontio/ontology/vm/neovm/types"
)

func TestOpArraySize(t *testing.T) {
	var e ExecutionEngine
	stack := NewRandAccessStack()
	stack.Push(NewStackItem(vtypes.NewByteArray([]byte("aaaaa"))))
	e.EvaluationStack = stack

	opArraySize(&e)
	if PeekInt(&e) != 5 {
		t.Fatalf("NeoVM OpArraySize test failed, expect 5, got %d.", PeekInt(&e))
	}
}

func TestOpPack(t *testing.T) {
	var e ExecutionEngine
	stack := NewRandAccessStack()
	stack.Push(NewStackItem(vtypes.NewByteArray([]byte("aaa"))))
	stack.Push(NewStackItem(vtypes.NewByteArray([]byte("bbb"))))
	stack.Push(NewStackItem(vtypes.NewByteArray([]byte("ccc"))))
	stack.Push(NewStackItem(vtypes.NewInteger(big.NewInt(3))))
	e.EvaluationStack = stack

	opPack(&e)
	if stack.Count() != 1 {
		t.Fatalf("NeoVM OpPack test failed, expect 3, got %d.", stack.Count())
	}

	items := make([]vtypes.StackItems, 0)
	items = append(items, vtypes.NewByteArray([]byte("ccc")))
	items = append(items, vtypes.NewByteArray([]byte("bbb")))
	items = append(items, vtypes.NewByteArray([]byte("aaa")))

	arr := PeekArray(&e)
	if len(arr) != 3 {
		t.Fatalf("NeoVM OpPack test failed, expect 3, got %d.", len(arr))
	}

	for i := 0; i < 3; i++ {
		if !bytes.Equal(arr[i].GetByteArray(), items[i].GetByteArray()) {
			t.Fatal("NeoVM OpPack test failed")
		}
	}
}

func TestOpUnpack(t *testing.T) {
	var e ExecutionEngine
	stack := NewRandAccessStack()
	e.EvaluationStack = stack

	items := make([]vtypes.StackItems, 0)
	items = append(items, vtypes.NewByteArray([]byte("aaa")))
	items = append(items, vtypes.NewByteArray([]byte("bbb")))
	items = append(items, vtypes.NewByteArray([]byte("ccc")))
	PushData(&e, items)

	opUnpack(&e)
	if stack.Count() != 4 || PopInt(&e) != 3 {
		t.Fatalf("NeoVM OpUnpack test failed, expect 3, got %d.", stack.Count())
	}

	for i := 0; i < 3; i++ {
		if !bytes.Equal(PopStackItem(&e).GetByteArray(), items[i].GetByteArray()) {
			t.Fatal("NeoVM OpUnpack test failed")
		}
	}
}

func TestOpPickItem(t *testing.T) {
	var e ExecutionEngine
	stack := NewRandAccessStack()
	e.EvaluationStack = stack

	items := make([]vtypes.StackItems, 0)
	items = append(items, vtypes.NewByteArray([]byte("aaa")))
	items = append(items, vtypes.NewByteArray([]byte("bbb")))
	items = append(items, vtypes.NewByteArray([]byte("ccc")))
	PushData(&e, items)
	stack.Push(NewStackItem(vtypes.NewInteger(big.NewInt(0))))

	opPickItem(&e)
	if stack.Count() != 1 || !bytes.Equal(PeekStackItem(&e).GetByteArray(), []byte("aaa")) {
		t.Fatal("NeoVM OpPickItem test failed.")
	}
}

func TestOpReverse(t *testing.T) {
	var e1 ExecutionEngine
	var e2 ExecutionEngine
	e1.EvaluationStack = NewRandAccessStack()
	e2.EvaluationStack = NewRandAccessStack()

	items := make([]vtypes.StackItems, 0)
	items = append(items, vtypes.NewByteArray([]byte("aaa")))
	items = append(items, vtypes.NewByteArray([]byte("bbb")))
	items = append(items, vtypes.NewByteArray([]byte("ccc")))
	PushData(&e1, items)
	PushData(&e2, items)

	t.Log("=======Before===========")

	t.Log(string(PeekArray(&e2)[0].GetByteArray()))
	t.Log(string(PeekArray(&e2)[1].GetByteArray()))
	t.Log(string(PeekArray(&e2)[2].GetByteArray()))

	opReverse(&e1)

	t.Log("=======After===========")
	t.Log(string(PeekArray(&e2)[0].GetByteArray()))
	t.Log(string(PeekArray(&e2)[1].GetByteArray()))
	t.Log(string(PeekArray(&e2)[2].GetByteArray()))

	if string(PeekArray(&e2)[0].GetByteArray()) != "ccc" {
		t.Fatalf("NeoVM OpReverse test failed, expect ccc, get %s.", string(PeekArray(&e2)[0].GetByteArray()))
	}
}
