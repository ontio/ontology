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

// Package exec provides functions for executing WebAssembly bytecode.
package exec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/vm/wasmvm/disasm"
	"github.com/ontio/ontology/vm/wasmvm/exec/internal/compile"
	"github.com/ontio/ontology/vm/wasmvm/memory"
	"github.com/ontio/ontology/vm/wasmvm/wasm"
	ops "github.com/ontio/ontology/vm/wasmvm/wasm/operators"
)

var (
	// ErrMultipleLinearMemories is returned by (*VM).NewVM when the module
	// has more then one entries in the linear memory space.
	ErrMultipleLinearMemories = errors.New("exec: more than one linear memories in module")
	// ErrInvalidArgumentCount is returned by (*VM).ExecCode when an invalid
	// number of arguments to the WebAssembly function are passed to it.
	ErrInvalidArgumentCount = errors.New("exec: invalid number of arguments to function")
)

// InvalidReturnTypeError is returned by (*VM).ExecCode when the module
// specifies an invalid return type value for the executed function.
type InvalidReturnTypeError int8

func (e InvalidReturnTypeError) Error() string {
	return fmt.Sprintf("Function has invalid return value_type: %d", int8(e))
}

// InvalidFunctionIndexError is returned by (*VM).ExecCode when the function
// index provided is invalid.
type InvalidFunctionIndexError int64

func (e InvalidFunctionIndexError) Error() string {
	return fmt.Sprintf("Invalid index to function index space: %d", int64(e))
}

type context struct {
	stack   []uint64
	locals  []uint64
	code    []byte
	pc      int64
	curFunc int64
}

//store env call message
type EnvCall struct {
	envParams  []uint64
	envReturns bool
	envPreCtx  context
	Message    []interface{} //the 'Message' field is for the EOS contract like parameters
}

func (ec *EnvCall) GetParams() []uint64 {
	return ec.envParams
}

func (ec *EnvCall) GetReturns() bool {
	return ec.envReturns
}

// VM is the execution context for executing WebAssembly bytecode.
type VM struct {
	ctx context

	module        *wasm.Module
	globals       []uint64
	compiledFuncs []compiledFunction
	funcTable     [256]func()
	Services      map[string]func(engine *ExecutionEngine) (bool, error)
	memory        *memory.VMmemory
	//store the env call parameters
	envCall *EnvCall
	//store a engine pointer
	ContractAddress common.Address
	Caller          common.Address
	Engine          *ExecutionEngine
	VMCode          []byte
}

// As per the WebAssembly spec: https://github.com/WebAssembly/design/blob/27ac254c854994103c24834a994be16f74f54186/Semantics.md#linear-memory
const wasmPageSize = 65536 // (64 KB)

var endianess = binary.LittleEndian

// NewVM creates a new VM from a given module. If the module defines a
// start function, it will be executed.
func NewVM(module *wasm.Module) (*VM, error) {
	var vm VM
	err := vm.loadModule(module)
	if err != nil {
		return nil, err
	}
	return &vm, nil
}

//alloc memory and return the first index
func (vm *VM) Malloc(size int) (int, error) {
	return vm.memory.Malloc(size)
}

//alloc memory for pointer and return the first index
func (vm *VM) MallocPointer(size int, ptype memory.PType) (int, error) {
	return vm.memory.MallocPointer(size, ptype)
}

func (vm *VM) GetPointerMemSize(addr uint64) int {
	return vm.memory.GetPointerMemSize(addr)
}

//when wasm returns a pointer, call this function to get the pointed memory
func (vm *VM) GetPointerMemory(addr uint64) ([]byte, error) {
	return vm.memory.GetPointerMemory(addr)
}

//alloc memory for any pointer type
func (vm *VM) SetPointerMemory(val interface{}) (int, error) {
	return vm.memory.SetPointerMemory(val)
}

//alloc memory for struct
//todo move to the SetPointerMemory
func (vm *VM) SetStructMemory(val interface{}) (int, error) {
	return vm.memory.SetStructMemory(val)

}

func (vm *VM) GetEnvCall() *EnvCall {
	return vm.envCall
}

func (vm *VM) RestoreCtx() bool {

	if vm.envCall != nil {
		vm.ctx = vm.envCall.envPreCtx
	}
	return true
}

//SetMessage
//for further extension
//support EOS like message
func (vm *VM) SetMessage(message []interface{}) {
	if message != nil {
		if vm.envCall == nil {
			vm.envCall = &EnvCall{}
		}
		vm.envCall.Message = message
	}
}

//GetMessageBytes
//for further extension
func (vm *VM) GetMessageBytes() ([]byte, error) {
	if vm.envCall.Message == nil || len(vm.envCall.Message) == 0 {
		return nil, nil
	}

	bytesbuf := bytes.NewBuffer(nil)

	for _, m := range vm.envCall.Message {
		switch m.(type) {
		case string:
			bytesbuf.WriteString(m.(string))
		case int:
			tmp := make([]byte, 4)
			binary.LittleEndian.PutUint32(tmp, uint32(m.(int)))
			bytesbuf.Write(tmp)
		case int64:
			tmp := make([]byte, 8)
			binary.LittleEndian.PutUint64(tmp, uint64(m.(int64)))
			bytesbuf.Write(tmp)
		case float32:
			bits := math.Float32bits(m.(float32))
			tmp := make([]byte, 4)
			binary.LittleEndian.PutUint32(tmp, bits)
			bytesbuf.Write(tmp)
		case float64:
			bits := math.Float64bits(m.(float64))
			tmp := make([]byte, 8)
			binary.LittleEndian.PutUint64(tmp, uint64(bits))
			bytesbuf.Write(tmp)

		default:
			//todo need support array types???
			return nil, errors.New("[GetMessageBytes] unsupported type")

		}
	}
	return bytesbuf.Bytes(), nil
}

func (vm *VM) SetMemory(val interface{}) (int, error) {
	return vm.memory.SetMemory(val)
}
func (vm *VM) GetMemory() *memory.VMmemory {
	return vm.memory
}

func (vm *VM) PushResult(res uint64) {
	vm.pushUint64(res)
}

// Memory returns the linear memory space for the VM.
func (vm *VM) Memory() []byte {
	return vm.memory.Memory
}

func (vm *VM) pushBool(v bool) {
	if v {
		vm.pushUint64(1)
	} else {
		vm.pushUint64(0)
	}
}

func (vm *VM) fetchBool() bool {
	return vm.fetchInt8() != 0
}

func (vm *VM) fetchInt8() int8 {
	i := int8(vm.ctx.code[vm.ctx.pc])
	vm.ctx.pc++
	return i
}

func (vm *VM) fetchUint32() uint32 {
	v := endianess.Uint32(vm.ctx.code[vm.ctx.pc:])
	vm.ctx.pc += 4
	return v
}

func (vm *VM) fetchInt32() int32 {
	return int32(vm.fetchUint32())
}

func (vm *VM) fetchFloat32() float32 {
	return math.Float32frombits(vm.fetchUint32())
}

func (vm *VM) fetchUint64() uint64 {
	v := endianess.Uint64(vm.ctx.code[vm.ctx.pc:])
	vm.ctx.pc += 8
	return v
}

func (vm *VM) fetchInt64() int64 {
	return int64(vm.fetchUint64())
}

func (vm *VM) fetchFloat64() float64 {
	return math.Float64frombits(vm.fetchUint64())
}

func (vm *VM) popUint64() uint64 {
	i := vm.ctx.stack[len(vm.ctx.stack)-1]
	vm.ctx.stack = vm.ctx.stack[:len(vm.ctx.stack)-1]
	return i
}

func (vm *VM) popInt64() int64 {
	return int64(vm.popUint64())
}

func (vm *VM) popFloat64() float64 {
	return math.Float64frombits(vm.popUint64())
}

func (vm *VM) popUint32() uint32 {
	return uint32(vm.popUint64())
}

func (vm *VM) popInt32() int32 {
	return int32(vm.popUint32())
}

func (vm *VM) popFloat32() float32 {
	return math.Float32frombits(vm.popUint32())
}

func (vm *VM) pushUint64(i uint64) {
	vm.ctx.stack = append(vm.ctx.stack, i)
}

func (vm *VM) pushInt64(i int64) {
	vm.pushUint64(uint64(i))
}

func (vm *VM) pushFloat64(f float64) {
	vm.pushUint64(math.Float64bits(f))
}

func (vm *VM) pushUint32(i uint32) {
	vm.pushUint64(uint64(i))
}

func (vm *VM) pushInt32(i int32) {
	vm.pushUint64(uint64(i))
}

func (vm *VM) pushFloat32(f float32) {
	vm.pushUint32(math.Float32bits(f))
}

// ExecCode calls the function with the given index and arguments.
// fnIndex should be a valid index into the function index space of
// the VM's module.
//insideCall :true (call contract)
func (vm *VM) ExecCode(insideCall bool, fnIndex int64, args ...uint64) (interface{}, error) {

	if int(fnIndex) > len(vm.compiledFuncs) {
		return nil, InvalidFunctionIndexError(fnIndex)
	}

	if len(vm.module.GetFunction(int(fnIndex)).Sig.ParamTypes) != len(args) {
		return nil, ErrInvalidArgumentCount
	}

	compiled := vm.compiledFuncs[fnIndex]

	if len(vm.ctx.stack) < compiled.maxDepth {
		vm.ctx.stack = make([]uint64, 0, compiled.maxDepth)
	}
	vm.ctx.locals = make([]uint64, compiled.totalLocalVars)
	vm.ctx.pc = 0
	vm.ctx.code = compiled.code
	vm.ctx.curFunc = fnIndex

	for i, arg := range args {
		vm.ctx.locals[i] = arg
	}
	var rtrn interface{}
	res := vm.execCode(insideCall, compiled)
	// for the call contract case
	if insideCall {
		return res, nil
	}

	if compiled.returns {
		rtrnType := vm.module.GetFunction(int(fnIndex)).Sig.ReturnTypes[0]
		switch rtrnType {
		case wasm.ValueTypeI32:
			rtrn = uint32(res)
		case wasm.ValueTypeI64:
			rtrn = uint64(res)
		case wasm.ValueTypeF32:
			rtrn = math.Float32frombits(uint32(res))
		case wasm.ValueTypeF64:
			rtrn = math.Float64frombits(res)
		default:
			return nil, InvalidReturnTypeError(rtrnType)
		}
	}

	return rtrn, nil
}

func (vm *VM) execCode(isinside bool, compiled compiledFunction) uint64 {
outer:
	for int(vm.ctx.pc) < len(vm.ctx.code) {
		op := vm.ctx.code[vm.ctx.pc]
		vm.ctx.pc++

		switch op {
		case ops.Return:
			break outer
		case compile.OpJmp:
			vm.ctx.pc = vm.fetchInt64()
			continue
		case compile.OpJmpZ:
			target := vm.fetchInt64()
			if vm.popUint32() == 0 {
				vm.ctx.pc = target
				continue
			}
		case compile.OpJmpNz:
			target := vm.fetchInt64()
			preserveTop := vm.fetchBool()
			discard := vm.fetchInt64()
			if vm.popUint32() != 0 {
				vm.ctx.pc = target
				var top uint64
				if preserveTop {
					top = vm.ctx.stack[len(vm.ctx.stack)-1]
				}
				vm.ctx.stack = vm.ctx.stack[:len(vm.ctx.stack)-int(discard)]
				if preserveTop {
					vm.pushUint64(top)
				}
				continue
			}
		case ops.BrTable:
			index := vm.fetchInt64()
			label := vm.popInt32()
			table := vm.compiledFuncs[vm.ctx.curFunc].branchTables[index]
			var target compile.Target
			if label >= 0 && label < int32(len(table.Targets)) {
				target = table.Targets[int32(label)]
			} else {
				target = table.DefaultTarget
			}

			if target.Return {
				break outer
			}
			vm.ctx.pc = target.Addr
			var top uint64
			if target.PreserveTop {
				top = vm.ctx.stack[len(vm.ctx.stack)-1]
			}
			vm.ctx.stack = vm.ctx.stack[:len(vm.ctx.stack)-int(target.Discard)]
			if target.PreserveTop {
				vm.pushUint64(top)
			}
			continue
		case compile.OpDiscard:
			place := vm.fetchInt64()
			if len(vm.ctx.stack)-int(place) > 0 {
				vm.ctx.stack = vm.ctx.stack[:len(vm.ctx.stack)-int(place)]
			}

		case compile.OpDiscardPreserveTop:
			top := vm.ctx.stack[len(vm.ctx.stack)-1]
			place := vm.fetchInt64()
			if len(vm.ctx.stack)-int(place) > 0 {
				vm.ctx.stack = vm.ctx.stack[:len(vm.ctx.stack)-int(place)]
			}
			vm.pushUint64(top)
		default:
			vm.funcTable[op]()
		}
	}

	if compiled.returns {
		return vm.ctx.stack[len(vm.ctx.stack)-1]
	}
	return 0
}

//CallContract
//start a new vm
//this method is replaced with wasm_service :callContract
func (vm *VM) CallContract(caller common.Address, contractAddress common.Address, module *wasm.Module, actionName []byte, arg []byte) (uint64, error) {

	methodName := CONTRACT_METHOD_NAME

	//1. exec the method code
	entry, ok := module.Export.Entries[methodName]
	if ok == false {
		return uint64(0), errors.New("Method:" + methodName + " does not exist!")
	}

	//get entry index
	index := int64(entry.Index)

	//new vm
	newvm, err := NewVM(module)
	if err != nil {
		return uint64(0), err
	}

	newvm.Caller = caller
	newvm.ContractAddress = contractAddress

	newvm.Services = vm.Services

	engine := vm.Engine
	newvm.Engine = engine

	engine.SetNewVM(newvm)

	actionIdx, err := newvm.SetPointerMemory(actionName)
	if err != nil {
		return uint64(0), err
	}
	argIdx, err := newvm.SetPointerMemory(arg)
	if err != nil {
		return uint64(0), err
	}
	res, err := newvm.ExecCode(true, int64(index), uint64(actionIdx), uint64(argIdx))
	if err != nil {
		return uint64(0), err
	}
	resBytes, err := newvm.GetPointerMemory(res.(uint64))
	if err != nil {
		return uint64(0), err
	}
	//copy memory if need!!!
	engine.RestoreVM()
	idx, err := vm.SetPointerMemory(resBytes)
	if err != nil {
		return uint64(0), err
	}

	return uint64(idx), nil
}

func (vm *VM) loadModule(module *wasm.Module) error {

	vm.memory = &memory.VMmemory{}
	if module.Memory != nil && len(module.Memory.Entries) != 0 {
		if len(module.Memory.Entries) > 1 {
			return ErrMultipleLinearMemories
		}
		vm.memory.Memory = make([]byte, uint(module.Memory.Entries[0].Limits.Initial)*wasmPageSize)
		copy(vm.memory.Memory, module.LinearMemoryIndexSpace[0])
	} else if len(module.LinearMemoryIndexSpace) > 0 {
		//add imported memory ,all mem access will be on the imported mem
		vm.memory.Memory = module.LinearMemoryIndexSpace[0]
	}

	//give a default memory even if no memory section exist in wasm file
	if vm.memory.Memory == nil {
		vm.memory.Memory = make([]byte, 1*wasmPageSize)
	}

	vm.memory.MemPoints = make(map[uint64]*memory.TypeLength) //init the pointer map

	//solve the Data section
	//this section is for some const strings, just like heap
	if module.Data != nil {
		var tmpIdx int
		for _, entry := range module.Data.Entries {
			if entry.Index != 0 {
				return errors.New("invalid data index")
			}
			val, err := module.ExecInitExpr(entry.Offset)
			if err != nil {
				return err
			}
			offset, ok := val.(int32)
			tmpIdx += int(offset) + len(entry.Data)
			if !ok {
				return errors.New("invalid data index")
			}
			// for the case of " (data (get_global 0) "init\00init success!\00add\00int"))"
			if bytes.Contains(entry.Data, []byte{byte(0)}) {
				splited := bytes.Split(entry.Data, []byte{byte(0)})
				var tmpoffset = int(offset)
				for _, tmp := range splited {
					vm.memory.MemPoints[uint64(tmpoffset)] = &memory.TypeLength{Ptype: memory.PString, Length: len(tmp) + 1}
					tmpoffset += len(tmp) + 1
				}
			} else {
				vm.memory.MemPoints[uint64(offset)] = &memory.TypeLength{Ptype: memory.PString, Length: len(entry.Data)}
			}
		}
		//
		vm.memory.AllocedMemIdex = tmpIdx
		vm.memory.PointedMemIndex = (len(vm.memory.Memory) + tmpIdx) / 2
	} else {
		//default pointed memory
		vm.memory.AllocedMemIdex = -1
		vm.memory.PointedMemIndex = len(vm.memory.Memory) / 2 //the second half memory is reserved for the pointed objects,string,array,structs
	}

	vm.compiledFuncs = make([]compiledFunction, len(module.FunctionIndexSpace))
	vm.globals = make([]uint64, len(module.GlobalIndexSpace))
	vm.newFuncTable()
	vm.module = module

	for i, fn := range module.FunctionIndexSpace {
		disassembly, err := disasm.Disassemble(fn, module)
		if err != nil {
			return err
		}

		totalLocalVars := 0
		totalLocalVars += len(fn.Sig.ParamTypes)
		for _, entry := range fn.Body.Locals {
			totalLocalVars += int(entry.Count)
		}
		code, table := compile.Compile(disassembly.Code)

		if fn.IsEnvFunc {
			vm.compiledFuncs[i] = compiledFunction{
				code:           code,
				branchTables:   table,
				maxDepth:       disassembly.MaxDepth,
				totalLocalVars: totalLocalVars,
				args:           len(fn.Sig.ParamTypes),
				returns:        len(fn.Sig.ReturnTypes) != 0,
				isEnv:          true,
				name:           fn.Name,
			}
		} else {
			vm.compiledFuncs[i] = compiledFunction{
				code:           code,
				branchTables:   table,
				maxDepth:       disassembly.MaxDepth,
				totalLocalVars: totalLocalVars,
				args:           len(fn.Sig.ParamTypes),
				returns:        len(fn.Sig.ReturnTypes) != 0,
			}
		}
	}

	for i, global := range module.GlobalIndexSpace {
		val, err := module.ExecInitExpr(global.Init)
		if err != nil {
			return err
		}
		switch v := val.(type) {
		case int32:
			vm.globals[i] = uint64(v)
		case int64:
			vm.globals[i] = uint64(v)
		case float32:
			vm.globals[i] = uint64(math.Float32bits(v))
		case float64:
			vm.globals[i] = uint64(math.Float64bits(v))
		}
	}

	if module.Start != nil {
		_, err := vm.ExecCode(false, int64(module.Start.Index))
		if err != nil {
			return err
		}
	}
	return nil

}
