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

package exec

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"reflect"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/types"
	"github.com/ontio/ontology/vm/neovm/interfaces"
	"github.com/ontio/ontology/vm/wasmvm/memory"
	"github.com/ontio/ontology/vm/wasmvm/util"
	"github.com/ontio/ontology/vm/wasmvm/validate"
	"github.com/ontio/ontology/vm/wasmvm/wasm"
)

const (
	CONTRACT_METHOD_NAME = "invoke"
	CONTRACT_INIT_METHOD = "init"
	VM_STACK_DEPTH       = 10
)

// backup vm while call other contracts
type vmstack struct {
	top   int
	stack []*VM
}

func (s *vmstack) push(vm *VM) error {
	if s.top+1 == len(s.stack) {
		return errors.NewErr(fmt.Sprintf("[vm stack push] stack is full, only support %d contracts calls", VM_STACK_DEPTH))
	}
	s.stack[s.top+1] = vm
	s.top += 1
	return nil
}

func (s *vmstack) pop() (*VM, error) {
	if s.top == -1 {
		return nil, errors.NewErr("[vm stack pop] stack is empty")
	}

	retvm := s.stack[s.top]
	s.top -= 1
	return retvm, nil
}

func newStack(depth int) *vmstack {
	return &vmstack{top: -1, stack: make([]*VM, depth)}
}

//todo add parameters
func NewExecutionEngine(container interfaces.CodeContainer, crypto interfaces.Crypto, service InteropServiceInterface) *ExecutionEngine {

	engine := &ExecutionEngine{
		crypto:        crypto,
		CodeContainer: container,
		service:       NewInteropService(),
	}
	if service != nil {
		engine.service.MergeMap(service.GetServiceMap())
	}

	engine.backupVM = newStack(VM_STACK_DEPTH)
	return engine
}

type ExecutionEngine struct {
	crypto        interfaces.Crypto
	service       *InteropService
	CodeContainer interfaces.CodeContainer
	vm            *VM
	backupVM      *vmstack
}

//GetVM return vm pointer
func (e *ExecutionEngine) GetVM() *VM {
	return e.vm
}

//for call other contract,
// 1.store current vm
// 2.load new vm
func (e *ExecutionEngine) SetNewVM(vm *VM) error {

	err := e.backupVM.push(e.vm)
	if err != nil {
		return err
	}
	e.vm = vm
	return nil
}

//for call other contract,
// 1.pop stored vm
// 2.reset vm
func (e *ExecutionEngine) RestoreVM() error {
	backupVM, err := e.backupVM.pop()
	if err != nil {
		return err
	}
	e.vm = backupVM
	return nil
}

//use this method just for test
func (e *ExecutionEngine) CallInf(caller common.Address,
	code []byte,
	input []interface{},
	message []interface{}) ([]byte, error) {
	methodName := input[0].(string)

	//1. read code
	bf := bytes.NewBuffer(code)

	//2. read module
	m, err := wasm.ReadModule(bf, importer)
	if err != nil {
		return nil, errors.NewErr("Verify wasm failed!" + err.Error())
	}

	//3. verify the module
	//already verified in step 2

	//4. check the export
	//every wasm should have at least 1 export
	if m.Export == nil {
		return nil, errors.NewErr("No export in wasm!")
	}

	vm, err := NewVM(m)
	if err != nil {
		return nil, err
	}
	vm.Engine = e
	if e.service != nil {
		vm.Services = e.service.GetServiceMap()
	}
	e.vm = vm
	vm.Engine = e

	vm.SetMessage(message)

	vm.Caller = caller
	//this is only for test
	//vm.CodeHash = common.ToCodeHash(code)

	entry, ok := m.Export.Entries[methodName]
	if ok == false {
		return nil, errors.NewErr("Method:" + methodName + " does not exist!")
	}
	//get entry index
	index := int64(entry.Index)

	//get function index
	fidx := m.Function.Types[int(index)]

	//get  function type
	ftype := m.Types.Entries[int(fidx)]

	paramlength := len(input) - 1
	if len(ftype.ParamTypes) != paramlength {
		return nil, errors.NewErr("parameter count is not right")
	}
	params := make([]uint64, paramlength)
	for i, param := range input[1:] {
		//if type is struct
		if reflect.TypeOf(param).Kind() == reflect.Struct {
			offset, err := vm.SetStructMemory(param)
			if err != nil {
				return nil, err
			}
			params[i] = uint64(offset)
		} else {
			switch param.(type) {
			case string:
				offset, err := vm.SetPointerMemory(param.(string))
				if err != nil {
					return nil, err
				}
				params[i] = uint64(offset)
			case int:
				params[i] = uint64(param.(int))
			case int64:
				params[i] = uint64(param.(int64))
			case float32:
				bits := math.Float32bits(param.(float32))
				params[i] = uint64(bits)
			case float64:
				bits := math.Float64bits(param.(float64))
				params[i] = uint64(bits)

			case []int:
				idx := 0
				for i, v := range param.([]int) {
					offset, err := vm.SetMemory(v)
					if err != nil {
						return nil, err
					}
					if i == 0 {
						idx = offset
					}
				}
				vm.GetMemory().MemPoints[uint64(idx)] = &memory.TypeLength{Ptype: memory.PInt32, Length: len(param.([]int)) * 4}
				params[i] = uint64(idx)

			case []int64:
				idx := 0
				for i, v := range param.([]int64) {
					offset, err := vm.SetMemory(v)
					if err != nil {
						return nil, err
					}
					if i == 0 {
						idx = offset
					}
				}
				vm.GetMemory().MemPoints[uint64(idx)] = &memory.TypeLength{Ptype: memory.PInt64, Length: len(param.([]int64)) * 8}
				params[i] = uint64(idx)

			case []float32:
				idx := 0
				for i, v := range param.([]float32) {
					offset, err := vm.SetMemory(v)
					if err != nil {
						return nil, err
					}
					if i == 0 {
						idx = offset
					}
				}
				vm.GetMemory().MemPoints[uint64(idx)] = &memory.TypeLength{Ptype: memory.PFloat32, Length: len(param.([]float32)) * 4}
				params[i] = uint64(idx)
			case []float64:
				idx := 0
				for i, v := range param.([]float64) {
					offset, err := vm.SetMemory(v)
					if err != nil {
						return nil, err
					}
					if i == 0 {
						idx = offset
					}
				}
				vm.GetMemory().MemPoints[uint64(idx)] = &memory.TypeLength{Ptype: memory.PFloat64, Length: len(param.([]float64)) * 8}
				params[i] = uint64(idx)
			}
		}

	}

	res, err := vm.ExecCode(false, index, params...)
	if err != nil {
		return nil, errors.NewErr("ExecCode error!" + err.Error())
	}

	if len(ftype.ReturnTypes) == 0 {
		return nil, nil
	}

	switch ftype.ReturnTypes[0] {
	case wasm.ValueTypeI32:
		return util.Int32ToBytes(res.(uint32)), nil
	case wasm.ValueTypeI64:
		return util.Int64ToBytes(res.(uint64)), nil
	case wasm.ValueTypeF32:
		return util.Float32ToBytes(res.(float32)), nil
	case wasm.ValueTypeF64:
		return util.Float64ToBytes(res.(float64)), nil
	default:
		return nil, errors.NewErr("the return type is not supported")
	}

	return nil, nil
}

func (e *ExecutionEngine) GetMemory() *memory.VMmemory {
	return e.vm.memory
}

func (e *ExecutionEngine) Create(caller common.Address, code []byte) ([]byte, error) {
	return code, nil
}

//
//Call Main interface of wasm vm excution engine
// caller      common.Address :call address
// code        []byte         :wasm smart contract code
// actionName  string         :action name of the contract
// input       []byte         :arguments
// ver         byte           :contract version  require > 0 for production
func (e *ExecutionEngine) Call(caller common.Address,
	code []byte,
	actionName string,
	input []byte,
	ver byte) (returnbytes []byte, er error) {

	//catch the panic to avoid crash the whole node
	defer func() {
		if err := recover(); err != nil {
			returnbytes = nil
			er = errors.NewErr("[Call] error happened while call wasmvm")
		}
	}()

	return e.call(caller, code, input, actionName, ver)

}

//Call init method on deployment
// caller      common.Address :call address
// code        []byte         :wasm smart contract code
// input       []byte         :arguments
// ver         byte           :contract version  require > 0 for production
//TODO TBD whether we need to call init method after deploy automatically
func (e *ExecutionEngine) InitCall(caller common.Address,
	code []byte,
	input []byte,
	ver byte) (returnbytes []byte, er error) {

	//catch the panic to avoid crash the whole node
	defer func() {
		if err := recover(); err != nil {
			returnbytes = nil
			er = errors.NewErr("[Call] error happened while call wasmvm")
		}
	}()

	return e.call(caller, code, input, CONTRACT_INIT_METHOD, ver)

}

// call to execute wasm vm
func (e *ExecutionEngine) call(caller common.Address,
	code []byte,
	input []byte,
	actionName string,
	ver byte) (returnbytes []byte, er error) {
	if ver > 0 { //production contract version
		methodName := CONTRACT_METHOD_NAME //fix to "invoke"
		//1. read code
		bf := bytes.NewBuffer(code)

		//2. read module
		m, err := wasm.ReadModule(bf, importer)
		if err != nil {
			return nil, errors.NewErr("[Call]Verify wasm failed!" + err.Error())
		}
		//3. verify the module
		//already verified in step 2

		//4. check the export
		//every wasm should have at least 1 export
		if m.Export == nil {
			return nil, errors.NewErr("[Call]No export in wasm!")
		}

		vm, err := NewVM(m)
		if err != nil {
			return nil, err
		}
		if e.service != nil {
			vm.Services = e.service.GetServiceMap()
		}
		e.vm = vm
		vm.Engine = e
		//no message support for now
		// vm.SetMessage(message)

		vm.Caller = caller

		vmcode := types.VmCode{VmType: types.WASMVM, Code: code}
		vm.VMCode = vmcode
		vm.ContractAddress = vmcode.AddressFromVmCode()

		entry, ok := m.Export.Entries[methodName]

		if ok == false {
			return nil, errors.NewErr("[Call]Method:" + methodName + " does not exist!")
		}
		//get entry index
		index := int64(entry.Index)

		//get function index
		fidx := m.Function.Types[int(index)]

		//get  function type
		ftype := m.Types.Entries[int(fidx)]
		//method ,param bytes
		params := make([]uint64, 2)

		actIdx, err := vm.SetPointerMemory(actionName)
		if err != nil {
			return nil, err
		}
		params[0] = uint64(actIdx)

		args := input
		argIdx, err := vm.SetPointerMemory(args)
		if err != nil {
			return nil, err
		}
		//init paramIndex
		vm.memory.ParamIndex = 0

		params[1] = uint64(argIdx)

		res, err := vm.ExecCode(false, index, params...)
		if err != nil {
			return nil, errors.NewErr("[Call]ExecCode error!" + err.Error())
		}

		if len(ftype.ReturnTypes) == 0 {
			return nil, nil
		}

		switch ftype.ReturnTypes[0] {
		case wasm.ValueTypeI32:
			return util.Int32ToBytes(res.(uint32)), nil
		case wasm.ValueTypeI64:
			return util.Int64ToBytes(res.(uint64)), nil
		case wasm.ValueTypeF32:
			return util.Float32ToBytes(res.(float32)), nil
		case wasm.ValueTypeF64:
			return util.Float64ToBytes(res.(float64)), nil
		default:
			return nil, errors.NewErr("[Call]the return type is not supported")
		}

	} else {
		//for test contract version
		methodName, err := getCallMethodName(input)
		if err != nil {
			return nil, err
		}

		//1. read code
		bf := bytes.NewBuffer(code)

		//2. read module
		m, err := wasm.ReadModule(bf, importer)
		if err != nil {
			return nil, errors.NewErr("[Call]Verify wasm failed!" + err.Error())
		}

		//3. verify the module
		//already verified in step 2

		//4. check the export
		//every wasm should have at least 1 export
		if m.Export == nil {
			return nil, errors.NewErr("[Call]No export in wasm!")
		}

		vm, err := NewVM(m)
		if err != nil {
			return nil, err
		}
		if e.service != nil {
			vm.Services = e.service.GetServiceMap()
		}
		e.vm = vm
		vm.Engine = e
		entry, ok := m.Export.Entries[methodName]
		if ok == false {
			return nil, errors.NewErr("[Call]Method:" + methodName + " does not exist!")
		}
		//get entry index
		index := int64(entry.Index)

		//get function index
		fidx := m.Function.Types[int(index)]

		//get  function type
		ftype := m.Types.Entries[int(fidx)]

		//paramtypes := ftype.ParamTypes

		params, err := getParams(input)
		if err != nil {
			return nil, err
		}

		if len(params) != len(ftype.ParamTypes) {
			return nil, errors.NewErr("[Call]Parameters count is not right")
		}

		res, err := vm.ExecCode(false, index, params...)
		if err != nil {
			return nil, errors.NewErr("[Call]ExecCode error!" + err.Error())
		}

		if len(ftype.ReturnTypes) == 0 {
			return nil, nil
		}

		switch ftype.ReturnTypes[0] {
		case wasm.ValueTypeI32:
			return util.Int32ToBytes(res.(uint32)), nil
		case wasm.ValueTypeI64:
			return util.Int64ToBytes(res.(uint64)), nil
		case wasm.ValueTypeF32:
			return util.Float32ToBytes(res.(float32)), nil
		case wasm.ValueTypeF64:
			return util.Float64ToBytes(res.(float64)), nil
		default:
			return nil, errors.NewErr("[Call]the return type is not supported")
		}

	}
}

//FIXME NOT IN USE BUT DON'T DELETE IT
//current we only support the ONT SYSTEM module import
//other imports will raise an error
func importer(name string) (*wasm.Module, error) {
	//TODO add the path into config file
	if name != "ONT" {
		return nil, errors.NewErr("import [" + name + "] is not supported! ")
	}
	f, err := os.Open(name + ".wasm")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	m, err := wasm.ReadModule(f, nil)
	err = validate.VerifyModule(m)
	if err != nil {
		return nil, err
	}
	return m, nil

}

// getCallMethodName only used for testing case
//get call method name from the input bytes
//the input should be:[Namelength][methodName][paramcount][param1Length][param2Length].....[param1Data][Param2Data][....]
//input[0] should be the name length
//next n bytes should the be the method name
func getCallMethodName(input []byte) (string, error) {

	if len(input) <= 1 {
		return "", errors.NewErr("[Call]input format error!")
	}

	length := int(input[0])

	if length > len(input[1:]) {
		return "", errors.NewErr("[Call]input method name length error!")
	}

	return string(input[1 : length+1]), nil
}

//getParams only used for testing case
func getParams(input []byte) ([]uint64, error) {

	nameLength := int(input[0])

	paramCnt := int(input[1+nameLength])

	res := make([]uint64, paramCnt)

	paramlengthSlice := input[1+nameLength+1 : 1+1+nameLength+paramCnt]

	paramSlice := input[1+nameLength+1+paramCnt:]

	for i := 0; i < paramCnt; i++ {
		//get param length
		pl := int(paramlengthSlice[i])

		if (i+1)*pl > len(paramSlice) {
			return nil, errors.NewErr("[Call]get param failed!")
		}
		param := paramSlice[i*pl : (i+1)*pl]

		if len(param) < 8 {
			temp := make([]byte, 8)
			copy(temp, param)
			res[i] = binary.LittleEndian.Uint64(temp)
		} else {
			res[i] = binary.LittleEndian.Uint64(param)
		}
	}
	return res, nil
}
