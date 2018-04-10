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

// Copyright 2017 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package exec

import (
	"errors"
	"math"
)

// ErrOutOfBoundsMemoryAccess is the error value used while trapping the VM
// when it detects an out of bounds access to the linear memory.
var ErrOutOfBoundsMemoryAccess = errors.New("exec: out of bounds memory access")

func (vm *VM) fetchBaseAddr() int {
	return int(vm.fetchUint32() + uint32(vm.popInt32()))
}

// inBounds returns true when the next vm.fetchBaseAddr() + offset
// indices are in bounds accesses to the linear memory.
func (vm *VM) inBounds(offset int) bool {
	addr := endianess.Uint32(vm.ctx.code[vm.ctx.pc:]) + uint32(vm.ctx.stack[len(vm.ctx.stack)-1])
	return int(addr)+offset < len(vm.memory.Memory)
}

// curMem returns a slice to the memeory segment pointed to by
// the current base address on the bytecode stream.
func (vm *VM) curMem() []byte {
	return vm.memory.Memory[vm.fetchBaseAddr():]
}

func (vm *VM) i32Load() {
	if !vm.inBounds(3) {
		panic(ErrOutOfBoundsMemoryAccess)
	}
	vm.pushUint32(endianess.Uint32(vm.curMem()))
}

func (vm *VM) i32Load8s() {
	if !vm.inBounds(0) {
		panic(ErrOutOfBoundsMemoryAccess)
	}
	vm.pushInt32(int32(int8(vm.memory.Memory[vm.fetchBaseAddr()])))
}

func (vm *VM) i32Load8u() {
	if !vm.inBounds(0) {
		panic(ErrOutOfBoundsMemoryAccess)
	}
	vm.pushUint32(uint32(uint8(vm.memory.Memory[vm.fetchBaseAddr()])))
}

func (vm *VM) i32Load16s() {
	if !vm.inBounds(1) {
		panic(ErrOutOfBoundsMemoryAccess)
	}
	vm.pushInt32(int32(int16(endianess.Uint16(vm.curMem()))))
}

func (vm *VM) i32Load16u() {
	if !vm.inBounds(1) {
		panic(ErrOutOfBoundsMemoryAccess)
	}
	vm.pushUint32(uint32(endianess.Uint16(vm.curMem())))
}

func (vm *VM) i64Load() {
	if !vm.inBounds(7) {
		panic(ErrOutOfBoundsMemoryAccess)
	}
	vm.pushUint64(endianess.Uint64(vm.curMem()))
}

func (vm *VM) i64Load8s() {
	if !vm.inBounds(0) {
		panic(ErrOutOfBoundsMemoryAccess)
	}
	vm.pushInt64(int64(int8(vm.memory.Memory[vm.fetchBaseAddr()])))
}

func (vm *VM) i64Load8u() {
	if !vm.inBounds(0) {
		panic(ErrOutOfBoundsMemoryAccess)
	}
	vm.pushUint64(uint64(uint8(vm.memory.Memory[vm.fetchBaseAddr()])))
}

func (vm *VM) i64Load16s() {
	if !vm.inBounds(1) {
		panic(ErrOutOfBoundsMemoryAccess)
	}
	vm.pushInt64(int64(int16(endianess.Uint16(vm.curMem()))))
}

func (vm *VM) i64Load16u() {
	if !vm.inBounds(1) {
		panic(ErrOutOfBoundsMemoryAccess)
	}
	vm.pushUint64(uint64(endianess.Uint16(vm.curMem())))
}

func (vm *VM) i64Load32s() {
	if !vm.inBounds(3) {
		panic(ErrOutOfBoundsMemoryAccess)
	}
	vm.pushInt64(int64(int32(endianess.Uint32(vm.curMem()))))
}

func (vm *VM) i64Load32u() {
	if !vm.inBounds(3) {
		panic(ErrOutOfBoundsMemoryAccess)
	}
	vm.pushUint64(uint64(endianess.Uint32(vm.curMem())))
}

func (vm *VM) f32Store() {
	v := math.Float32bits(vm.popFloat32())
	if !vm.inBounds(3) {
		panic(ErrOutOfBoundsMemoryAccess)
	}
	endianess.PutUint32(vm.curMem(), v)
}

func (vm *VM) f32Load() {
	if !vm.inBounds(3) {
		panic(ErrOutOfBoundsMemoryAccess)
	}
	vm.pushFloat32(math.Float32frombits(endianess.Uint32(vm.curMem())))
}

func (vm *VM) f64Store() {
	v := math.Float64bits(vm.popFloat64())
	if !vm.inBounds(7) {
		panic(ErrOutOfBoundsMemoryAccess)
	}
	endianess.PutUint64(vm.curMem(), v)
}

func (vm *VM) f64Load() {
	if !vm.inBounds(7) {
		panic(ErrOutOfBoundsMemoryAccess)
	}
	vm.pushFloat64(math.Float64frombits(endianess.Uint64(vm.curMem())))
}

func (vm *VM) i32Store() {
	v := vm.popUint32()
	if !vm.inBounds(3) {
		panic(ErrOutOfBoundsMemoryAccess)
	}
	endianess.PutUint32(vm.curMem(), v)
}

func (vm *VM) i32Store8() {
	v := byte(uint8(vm.popUint32()))
	if !vm.inBounds(0) {
		panic(ErrOutOfBoundsMemoryAccess)
	}
	vm.memory.Memory[vm.fetchBaseAddr()] = v
}

func (vm *VM) i32Store16() {
	v := uint16(vm.popUint32())
	if !vm.inBounds(1) {
		panic(ErrOutOfBoundsMemoryAccess)
	}
	endianess.PutUint16(vm.curMem(), v)
}

func (vm *VM) i64Store() {
	v := vm.popUint64()
	if !vm.inBounds(7) {
		panic(ErrOutOfBoundsMemoryAccess)
	}
	endianess.PutUint64(vm.curMem(), v)
}

func (vm *VM) i64Store8() {
	v := byte(uint8(vm.popUint64()))
	if !vm.inBounds(0) {
		panic(ErrOutOfBoundsMemoryAccess)
	}
	vm.memory.Memory[vm.fetchBaseAddr()] = v
}

func (vm *VM) i64Store16() {
	v := uint16(vm.popUint64())
	if !vm.inBounds(1) {
		panic(ErrOutOfBoundsMemoryAccess)
	}
	endianess.PutUint16(vm.curMem(), v)
}

func (vm *VM) i64Store32() {
	v := uint32(vm.popUint64())
	if !vm.inBounds(3) {
		panic(ErrOutOfBoundsMemoryAccess)
	}
	endianess.PutUint32(vm.curMem(), v)
}

func (vm *VM) currentMemory() {
	_ = vm.fetchInt8() // reserved (https://github.com/WebAssembly/design/blob/27ac254c854994103c24834a994be16f74f54186/BinaryEncoding.md#memory-related-operators-described-here)
	vm.pushInt32(int32(len(vm.memory.Memory) / wasmPageSize))
}

func (vm *VM) growMemory() {
	_ = vm.fetchInt8() // reserved (https://github.com/WebAssembly/design/blob/27ac254c854994103c24834a994be16f74f54186/BinaryEncoding.md#memory-related-operators-described-here)
	curLen := len(vm.memory.Memory) / wasmPageSize
	n := vm.popInt32()
	vm.memory.Memory = append(vm.memory.Memory, make([]byte, n*wasmPageSize)...)
	vm.pushInt32(int32(curLen))
}
