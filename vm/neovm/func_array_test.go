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
	v, err := PeekInt(&e)
	if err != nil {
		t.Fatalf("NeoVM OpArraySize test failed.")
	}
	if v != 5 {
		t.Fatalf("NeoVM OpArraySize test failed, expect 5, got %d.", v)
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

	arr, err := PeekArray(&e)
	if err != nil {
		t.Fatalf("NeoVM OpPack test failed.")
	}
	if len(arr) != 3 {
		t.Fatalf("NeoVM OpPack test failed, expect 3, got %d.", len(arr))
	}

	for i := 0; i < 3; i++ {
		v1, arrErr := arr[i].GetByteArray()
		v2, itemErr := items[i].GetByteArray()
		if arrErr != nil || itemErr != nil {
			t.Fatal("NeoVM OpPack test failed.")
		}
		if !bytes.Equal(v1, v2) {
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
	v, err := PopInt(&e)
	if err != nil {
		t.Fatalf("NeoVM OpUnpack test failed.")
	}
	if stack.Count() != 3 || v != 3 {
		t.Fatalf("NeoVM OpUnpack test failed, expect 3, got %d.", stack.Count())
	}

	for i := 0; i < 3; i++ {
		v1, err1 := PopStackItem(&e).GetByteArray()
		v2, err2 := items[i].GetByteArray()
		if err1 != nil || err2 != nil {
			t.Fatal("NeoVM OpUnpack test failed.")
		}
		if !bytes.Equal(v1, v2) {
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
	v, err := PeekStackItem(&e).GetByteArray()
	if err != nil {
		t.Fatal("NeoVM OpPickItem test failed.")
	}
	if stack.Count() != 1 || !bytes.Equal(v, []byte("aaa")) {
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

	opReverse(&e1)
	arr, err := PeekArray(&e2)
	if err != nil {
		t.Fatal("NeoVM OpReverse test failed.")
	}
	v, err := arr[0].GetByteArray()
	if err != nil {
		t.Fatal("NeoVM OpReverse test failed.")
	}

	if string(v) != "ccc" {
		t.Fatalf("NeoVM OpReverse test failed, expect ccc, get %s.", string(v))
	}
}

func TestOpRemove(t *testing.T) {
	var e1 ExecutionEngine
	e1.EvaluationStack = NewRandAccessStack()


	items := make([]vtypes.StackItems, 0)
	items = append(items, vtypes.NewByteArray([]byte("aaa")))
	items = append(items, vtypes.NewByteArray([]byte("bbb")))
	items = append(items, vtypes.NewByteArray([]byte("ccc")))

	arr := vtypes.NewArray(items)
	PushData(&e1, arr)
	opDup(&e1)

	PushData(&e1, vtypes.NewInteger(big.NewInt(1)))

	opRemove(&e1)

	arr1  , err := PeekArray(&e1)
	if err != nil {
		t.Fatal("NeoVM OpRemove test failed.")
	}

	if len(arr1) != 2 {
		t.Fatal("NeoVM OpRemove test failed. expect len 2, get %v", len(arr1))
	}

	v, err := arr1[0].GetByteArray()

	if err != nil {
		t.Fatal("NeoVM OpRemove test failed.")
	}

	if string(v) != "bbb" {
		t.Fatalf("NeoVM OpRemove test failed, expect bbb, get %s.", string(v))
	}
}
