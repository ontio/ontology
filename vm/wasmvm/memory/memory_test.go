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
package memory

import (
	"bytes"
	"testing"
)

func TestVMmemory_Malloc(t *testing.T) {
	tmpMem := &VMmemory{
		Memory:          make([]byte, 10),
		AllocedMemIdex:  -1,
		PointedMemIndex: 5,
	}

	idx, err := tmpMem.Malloc(6)
	if err == nil {
		t.Fatal("Malloc 6 should failed")
	}

	idx, err = tmpMem.Malloc(5)
	if err != nil {
		t.Fatal("Malloc 5 should not failed")
	}
	if idx != 0 {
		t.Fatal("idx should be 0")
	}
	idx, err = tmpMem.Malloc(1)
	if err == nil {
		t.Fatal("Malloc 5 and 1 should not failed")
	}

}

func TestVMmemory_SetPointerMemory(t *testing.T) {
	tmpMem := &VMmemory{
		Memory:          make([]byte, 10),
		AllocedMemIdex:  -1,
		PointedMemIndex: 5,
		MemPoints:       make(map[uint64]*TypeLength),
	}

	idx, err := tmpMem.SetPointerMemory(nil)
	if err != nil {
		t.Fatal("setPointerMemory should failed")
	}
	if idx != VM_NIL_POINTER {
		t.Fatal("setPointerMemory should failed")
	}

	idx, err = tmpMem.SetPointerMemory("abcdef")
	if err == nil {
		t.Fatal("setPointerMemory should failed")
	}

	idx, err = tmpMem.SetPointerMemory("abc")
	if err != nil {
		t.Fatal("setPointerMemory should failed")
	}
	if idx != 6 {
		t.Fatal("setPointerMemory idx should be 6")
	}

}

func TestVMmemory_SetPointerMemory2(t *testing.T) {
	tmpMem := &VMmemory{
		Memory:          make([]byte, 20),
		AllocedMemIdex:  -1,
		PointedMemIndex: 10,
		MemPoints:       make(map[uint64]*TypeLength),
	}

	intarr := []int{1, 2, 3}
	_, err := tmpMem.SetPointerMemory(intarr)
	if err == nil {
		t.Fatal("SetPointerMemory should failed")
	}
	intarr = []int{1, 2}
	idx, err := tmpMem.SetPointerMemory(intarr)
	if err != nil {
		t.Fatal("SetPointerMemory should not failed")
	}
	if idx != 11 {
		t.Fatal("idx should be 11")
	}

}
func TestVMmemory_SetPointerMemory3(t *testing.T) {
	tmpMem := &VMmemory{
		Memory:          make([]byte, 40),
		AllocedMemIdex:  -1,
		PointedMemIndex: 20,
		MemPoints:       make(map[uint64]*TypeLength),
	}

	i64arr := []int64{1, 2, 3}
	_, err := tmpMem.SetPointerMemory(i64arr)
	if err == nil {
		t.Fatal("SetPointerMemory should failed")
	}

	i64arr = []int64{1, 2}
	idx, err := tmpMem.SetPointerMemory(i64arr)
	if err != nil {
		t.Fatal("SetPointerMemory should not failed")
	}
	if idx != 21 {
		t.Fatal("SetPointerMemory idx should be 21")
	}

}

func TestVMmemory_SetPointerMemory4(t *testing.T) {
	tmpMem := &VMmemory{
		Memory:          make([]byte, 10),
		AllocedMemIdex:  -1,
		PointedMemIndex: 5,
		MemPoints:       make(map[uint64]*TypeLength),
	}

	idx, err := tmpMem.SetPointerMemory(nil)
	if err != nil {
		t.Fatal("setPointerMemory should failed")
	}
	if idx != VM_NIL_POINTER {
		t.Fatal("setPointerMemory should failed")
	}

	bf := bytes.NewBufferString("abcdefg")
	idx, err = tmpMem.SetPointerMemory(bf.Bytes())
	if err == nil {
		t.Fatal("setPointerMemory should failed")
	}

	bf = bytes.NewBufferString("abc")
	idx, err = tmpMem.SetPointerMemory(bf.Bytes())
	if err != nil {
		t.Fatal("setPointerMemory should failed")
	}
	if idx != 6 {
		t.Fatal("setPointerMemory idx should be 6")
	}

}

func TestVMmemory_MallocPointer(t *testing.T) {
	tmpMem := &VMmemory{
		Memory:          make([]byte, 10),
		AllocedMemIdex:  -1,
		PointedMemIndex: 5,
		MemPoints:       make(map[uint64]*TypeLength),
	}

	_, err := tmpMem.MallocPointer(6, PString)
	if err == nil {
		t.Fatal("MallocPointer should failed")
	}
}

func TestVMmemory_GetPointerMemSize(t *testing.T) {
	tmpMem := &VMmemory{
		Memory:          make([]byte, 10),
		AllocedMemIdex:  -1,
		PointedMemIndex: 5,
		MemPoints:       make(map[uint64]*TypeLength),
	}

	idx, err := tmpMem.SetPointerMemory("abc")
	if err != nil {
		t.Fatal("GetPointerMemSize should not failed")
	}

	size := tmpMem.GetPointerMemSize(uint64(idx))
	if size != 3 {
		t.Fatal("GetPointerMemSize size should be 3")
	}

	size = tmpMem.GetPointerMemSize(uint64(8))
	if size != 0 {
		t.Fatal("GetPointerMemSize size should be 0")
	}
}

func TestVMmemory_GetPointerMemory(t *testing.T) {
	tmpMem := &VMmemory{
		Memory:          make([]byte, 10),
		AllocedMemIdex:  -1,
		PointedMemIndex: 5,
		MemPoints:       make(map[uint64]*TypeLength),
	}

	idx, err := tmpMem.SetPointerMemory("abc")
	if err != nil {
		t.Fatal("GetPointerMemSize should not failed")
	}

	bts, err := tmpMem.GetPointerMemory(uint64(idx))
	if err != nil {
		t.Fatal("GetPointerMemSize should not failed")
	}
	if string(bts) != "abc" {
		t.Fatal("GetPointerMemSize bts should be 'abc'")
	}

	bts, err = tmpMem.GetPointerMemory(uint64(8))
	if err != nil {
		t.Fatal("GetPointerMemSize should not failed")
	}
	if bts != nil {
		t.Fatal("GetPointerMemSize bts should be nil")
	}
}
