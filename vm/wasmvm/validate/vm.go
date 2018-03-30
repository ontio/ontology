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

package validate

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/ontio/ontology/vm/wasmvm/wasm"
	"github.com/ontio/ontology/vm/wasmvm/wasm/leb128"
	ops "github.com/ontio/ontology/vm/wasmvm/wasm/operators"
)

// mockVM is a minimal implementation of a virtual machine to
// validate WebAssembly code
type mockVM struct {
	stack      []operand
	stackTop   int // the top of the operand stack
	origLength int // the original length of the bytecode stream

	code *bytes.Reader

	polymorphic bool    // whether the base implicit block has a polymorphic stack
	blocks      []block // a stack of encountered blocks

	curFunc *wasm.FunctionSig
}

// a block reprsents an instruction sequence preceded by a control flow operator
// it is used to verify that the block signature set by the operator is the correct
// one when the block ends
type block struct {
	pc          int            // the pc where the control flow operator starting the block is located
	stackTop    int            // stack top when the block started
	blockType   wasm.BlockType // block_type signature of the control operator
	op          byte           // opcode for the operator starting the new block
	polymorphic bool           // whether the block has a polymorphic stack
	loop        bool           // whether the block is the body of a loop instruction
}

func (vm *mockVM) fetchVarUint() (uint32, error) {
	return leb128.ReadVarUint32(vm.code)
}

func (vm *mockVM) fetchVarInt() (int32, error) {
	return leb128.ReadVarint32(vm.code)
}

func (vm *mockVM) fetchVarInt64() (int64, error) {
	return leb128.ReadVarint64(vm.code)
}

func (vm *mockVM) fetchUint32() (uint32, error) {
	var buf [4]byte
	_, err := io.ReadFull(vm.code, buf[:])
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:]), nil
}

func (vm *mockVM) fetchUint64() (uint64, error) {
	var buf [8]byte
	_, err := io.ReadFull(vm.code, buf[:])
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(buf[:]), nil
}

func (vm *mockVM) pushBlock(op byte, blockType wasm.BlockType) {
	logger.Printf("Pushing block %v", blockType)
	vm.blocks = append(vm.blocks, block{
		pc:          vm.pc(),
		stackTop:    vm.stackTop,
		blockType:   blockType,
		polymorphic: vm.isPolymorphic(),
		op:          op,
		loop:        op == ops.Loop,
	})
}

// Get a block from it's relative nesting depth
func (vm *mockVM) getBlockFromDepth(depth int) *block {
	if depth >= len(vm.blocks) {
		return nil
	}

	return &vm.blocks[len(vm.blocks)-1-depth]
}

// Returns nil if depth is a valid nesting depth value that can be
// branched to.
func (vm *mockVM) canBranch(depth int) error {
	blockType := wasm.BlockTypeEmpty

	block := vm.getBlockFromDepth(depth)
	// jumping to the start of a loop block doesn't push a value
	// on the stack.
	if block == nil {
		if depth == len(vm.blocks) {
			//equivalent to a `return', as the function
			//body is an "implicit" block
			if len(vm.curFunc.ReturnTypes) != 0 {
				blockType = wasm.BlockType(vm.curFunc.ReturnTypes[0])
			}
		} else {
			return InvalidLabelError(uint32(depth))
		}
	} else if !block.loop {
		blockType = block.blockType
	}

	if blockType != wasm.BlockTypeEmpty {
		top, under := vm.topOperand()
		if under || top.Type != wasm.ValueType(blockType) {
			return InvalidTypeError{wasm.ValueType(blockType), top.Type}
		}
	}

	return nil
}

// returns nil in case of an underflow
func (vm *mockVM) popBlock() *block {
	if len(vm.blocks) == 0 {
		return nil
	}

	stackTop := len(vm.blocks) - 1
	block := vm.blocks[stackTop]
	vm.blocks = append(vm.blocks[:stackTop], vm.blocks[stackTop+1:]...)

	return &block
}

func (vm *mockVM) topBlock() *block {
	if len(vm.blocks) == 0 {
		return nil
	}

	return &vm.blocks[len(vm.blocks)-1]
}

func (vm *mockVM) topOperand() (o operand, under bool) {
	stackTop := vm.stackTop - 1
	if stackTop == -1 {
		under = true
		return
	}
	o = vm.stack[stackTop]
	return
}

func (vm *mockVM) popOperand() (operand, bool) {
	var o operand
	stackTop := vm.stackTop - 1
	if stackTop == -1 {
		return o, true
	}
	o = vm.stack[stackTop]
	vm.stackTop--

	logger.Printf("Stack after pop is %v. Popped %v", vm.stack[:vm.stackTop], o)
	return o, false
}

func (vm *mockVM) pushOperand(t wasm.ValueType) {
	o := operand{t}
	logger.Printf("Stack top: %d, Len of stack :%d", vm.stackTop, len(vm.stack))
	if vm.stackTop == len(vm.stack) {
		vm.stack = append(vm.stack, o)
	} else {
		vm.stack[vm.stackTop] = o
	}
	vm.stackTop++

	logger.Printf("Stack after push is %v. Pushed %v", vm.stack[:vm.stackTop], o)
}

func (vm *mockVM) adjustStack(op ops.Op) error {
	for _, t := range op.Args {
		op, under := vm.popOperand()
		if !vm.isPolymorphic() && (under || op.Type != t) {
			return InvalidTypeError{t, op.Type}
		}
	}

	if op.Returns != wasm.ValueType(wasm.BlockTypeEmpty) {
		vm.pushOperand(op.Returns)
	}

	return nil
}

// setPolymorphic sets the current block as having a polymorphic stack
// blocks created under it will be polymorphic too. All type-checking
// is ignored in a polymorhpic stack.
// (See https://github.com/WebAssembly/design/blob/27ac254c854994103c24834a994be16f74f54186/Semantics.md#validation)
func (vm *mockVM) setPolymorphic() {
	if len(vm.blocks) == 0 {
		vm.polymorphic = true
	} else {

		vm.blocks[len(vm.blocks)-1].polymorphic = true
	}
}

func (vm *mockVM) isPolymorphic() bool {
	if len(vm.blocks) == 0 {
		return vm.polymorphic
	}

	return vm.topBlock().polymorphic
}

func (vm *mockVM) pc() int {
	return vm.origLength - vm.code.Len()
}
