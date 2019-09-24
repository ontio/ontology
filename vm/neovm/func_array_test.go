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

	vtypes "github.com/ontio/ontology/vm/neovm/types"
)

func TestOpRemove(t *testing.T) {
	var e1 ExecutionEngine
	e1.EvaluationStack = NewRandAccessStack()

	m1 := vtypes.NewMap()

	m1.Add(vtypes.NewByteArray([]byte("aaa")), vtypes.NewByteArray([]byte("aaa")))
	m1.Add(vtypes.NewByteArray([]byte("bbb")), vtypes.NewByteArray([]byte("bbb")))
	m1.Add(vtypes.NewByteArray([]byte("ccc")), vtypes.NewByteArray([]byte("ccc")))

	PushData(&e1, m1)
	opDup(&e1)
	PushData(&e1, vtypes.NewByteArray([]byte("aaa")))
	opRemove(&e1)

	mm := e1.EvaluationStack.Peek(0)

	v := mm.(*vtypes.Map).TryGetValue(vtypes.NewByteArray([]byte("aaa")))

	if v != nil {
		t.Fatal("NeoVM OpRemove remove map failed.")
	}
}

func TestStruct_Clone(t *testing.T) {
	var e1 ExecutionEngine
	e1.EvaluationStack = NewRandAccessStack()
	a := vtypes.NewStruct(nil)
	for i := 0; i < 1024; i++ {
		a.Add(vtypes.NewStruct(nil))
	}
	b := vtypes.NewStruct(nil)
	for i := 0; i < 1024; i++ {
		b.Add(a)
	}
	PushData(&e1, b)
	for i := 0; i < 1024; i++ {
		opDup(&e1)
		opDup(&e1)
		opAppend(&e1)
	}

}

func TestOpValues(t *testing.T) {
	var e ExecutionEngine
	stack := NewRandAccessStack()
	e.EvaluationStack = stack
	a := vtypes.NewMap()
	key1 := NewStackItem(vtypes.NewByteArray([]byte("aaa")))
	key2 := NewStackItem(vtypes.NewByteArray([]byte("bbb")))

	value1 := NewStackItem(vtypes.NewByteArray([]byte("xxx")))
	value2 := NewStackItem(vtypes.NewByteArray([]byte("yyy")))

	a.Add(key1, value1)
	a.Add(key2, value2)

	PushData(&e, a)
	opValues(&e)

	arr, err := PeekArray(&e)
	if err != nil {
		t.Fatal("NeoVM OpValues test failed.")
	}

	v, err := arr[0].GetByteArray()
	if err != nil {
		t.Fatal("NeoVM OpValues test failed.")
	}

	if string(v) != "xxx" {
		t.Fatalf("NeoVM OpValues test failed, expect xxx, get %s.", string(v))
	}

	v1, err := arr[1].GetByteArray()
	if err != nil {
		t.Fatal("NeoVM OpValues test failed.")
	}

	if string(v1) != "yyy" {
		t.Fatalf("NeoVM OpValues test failed, expect xxx, get %s.", string(v1))
	}
}

func TestOpKeys(t *testing.T) {
	var e ExecutionEngine
	stack := NewRandAccessStack()
	e.EvaluationStack = stack
	a := vtypes.NewMap()
	key1 := NewStackItem(vtypes.NewByteArray([]byte("aaa")))
	key2 := NewStackItem(vtypes.NewByteArray([]byte("bbb")))

	value1 := NewStackItem(vtypes.NewByteArray([]byte("xxx")))
	value2 := NewStackItem(vtypes.NewByteArray([]byte("yyy")))

	a.Add(key1, value1)
	a.Add(key2, value2)

	PushData(&e, a)
	opKeys(&e)

	arr, err := PeekArray(&e)
	if err != nil {
		t.Fatal("NeoVM OpValues test failed.")
	}

	v, err := arr[0].GetByteArray()
	if err != nil {
		t.Fatal("NeoVM OpValues test failed.")
	}

	if string(v) != "aaa" {
		t.Fatalf("NeoVM OpValues test failed, expect xxx, get %s.", string(v))
	}

	v1, err := arr[1].GetByteArray()
	if err != nil {
		t.Fatal("NeoVM OpValues test failed.")
	}

	if string(v1) != "bbb" {
		t.Fatalf("NeoVM OpValues test failed, expect xxx, get %s.", string(v1))
	}
}

func TestOpHasKey(t *testing.T) {
	var e ExecutionEngine
	stack := NewRandAccessStack()
	e.EvaluationStack = stack
	a := vtypes.NewMap()
	key1 := NewStackItem(vtypes.NewByteArray([]byte("aaa")))
	key2 := NewStackItem(vtypes.NewByteArray([]byte("bbb")))
	key3 := NewStackItem(vtypes.NewByteArray([]byte("ccc")))

	value1 := NewStackItem(vtypes.NewByteArray([]byte("xxx")))
	value2 := NewStackItem(vtypes.NewByteArray([]byte("yyy")))

	a.Add(key1, value1)
	a.Add(key2, value2)

	PushData(&e, a)
	PushData(&e, key2)
	opHasKey(&e)
	arr, err := PopBoolean(&e)
	if err != nil {
		t.Fatal("NeoVM OpHaskey test failed.")
	}

	if !arr {
		t.Fatalf("NeoVM OpHaskey test failed, expect true, get false.")
	}

	PushData(&e, a)
	PushData(&e, key3)
	opHasKey(&e)
	arr1, err := PopBoolean(&e)

	if err != nil {
		t.Fatal("NeoVM OpHaskey test failed.")
	}

	if arr1 {
		t.Fatalf("NeoVM OpHaskey test failed, expect false , get true.")
	}

}
