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
package wasmvm

import (
	"sync"

	lru "github.com/hashicorp/golang-lru"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/context"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/states"
	"github.com/ontio/ontology/smartcontract/storage"
	"github.com/ontio/wagon/exec"
)

type WasmVmService struct {
	CacheDB       *storage.CacheDB
	ContextRef    context.ContextRef
	Notifications []*event.NotifyEventInfo
	Code          []byte
	Tx            *types.Transaction
	Time          uint32
	Height        uint32
	BlockHash     common.Uint256
	PreExec       bool
	GasPrice      uint64
	GasLimit      *uint64
	ExecStep      *uint64
	GasFactor     uint64
	IsTerminate   bool
	JitMode       bool
	ServiceIndex  uint64
	vm            *exec.VM
}

var (
	ERR_CHECK_STACK_SIZE  = errors.NewErr("[WasmVmService] vm over max stack size!")
	ERR_EXECUTE_CODE      = errors.NewErr("[WasmVmService] vm execute code invalid!")
	ERR_GAS_INSUFFICIENT  = errors.NewErr("[WasmVmService] gas insufficient")
	VM_EXEC_STEP_EXCEED   = errors.NewErr("[WasmVmService] vm execute step exceed!")
	CONTRACT_NOT_EXIST    = errors.NewErr("[WasmVmService] Get contract code from db fail")
	DEPLOYCODE_TYPE_ERROR = errors.NewErr("[WasmVmService] DeployCode type error!")
	VM_EXEC_FAULT         = errors.NewErr("[WasmVmService] vm execute state fault!")
	VM_INIT_FAULT         = errors.NewErr("[WasmVmService] vm init state fault!")

	CODE_CACHE_SIZE      = 100
	CONTRACT_METHOD_NAME = "invoke"

	//max memory size of wasm vm
	WASM_MEM_LIMITATION  uint64 = 10 * 1024 * 1024
	VM_STEP_LIMIT               = 40000000
	WASM_CALLSTACK_LIMIT        = 1024

	CodeCache *lru.ARCCache

	serviceData        = make(map[uint64]*WasmVmService)
	nextServiceDataIdx uint64
	serviceDataMtx     sync.RWMutex
)

func init() {
	CodeCache, _ = lru.NewARC(CODE_CACHE_SIZE)
	nextServiceDataIdx = 1
	//if err != nil{
	//	log.Info("NewARC block error %s", err)
	//}
}

func GetAddressBuff(addrs []common.Address) ([]byte, int) {
	sink := common.NewZeroCopySink(nil)
	for _, addr := range addrs {
		sink.WriteAddress(addr)
	}

	return sink.Bytes(), int(sink.Size())
}

func registerWasmVmService(this *WasmVmService) uint64 {
	defer func() {
		nextServiceDataIdx++
		if nextServiceDataIdx == 0 {
			nextServiceDataIdx++
		}
		serviceDataMtx.Unlock()
	}()
	serviceDataMtx.Lock()
	serviceData[nextServiceDataIdx] = this
	this.ServiceIndex = nextServiceDataIdx
	return nextServiceDataIdx
}

func getWasmVmService(index uint64) *WasmVmService {
	defer serviceDataMtx.Unlock()
	serviceDataMtx.Lock()
	return serviceData[index]
}

func unregisterWasmVmService(index uint64) {
	defer serviceDataMtx.Unlock()
	serviceDataMtx.Lock()
	delete(serviceData, index)
}

func (this *WasmVmService) Invoke() (interface{}, error) {
	if len(this.Code) == 0 {
		return nil, ERR_EXECUTE_CODE
	}

	contract := &states.WasmContractParam{}
	sink := common.NewZeroCopySource(this.Code)
	err := contract.Deserialization(sink)
	if err != nil {
		return nil, err
	}

	code, err := this.CacheDB.GetContract(contract.Address)
	if err != nil {
		return nil, err
	}

	if code == nil {
		return nil, errors.NewErr("wasm contract does not exist")
	}

	wasmCode, err := code.GetWasmCode()
	if err != nil {
		return nil, errors.NewErr("not a wasm contract")
	}

	this.ContextRef.PushContext(&context.Context{ContractAddress: contract.Address, Code: wasmCode})

	var output []byte
	if this.JitMode {
		output, err = invokeJit(this, contract, wasmCode)
	} else {
		output, err = invokeInterpreter(this, contract, wasmCode)
	}

	if err != nil {
		return nil, err
	}

	this.ContextRef.PopContext()
	return output, nil
}

func invokeInterpreter(this *WasmVmService, contract *states.WasmContractParam, wasmCode []byte) ([]byte, error) {
	host := &Runtime{Service: this, Input: contract.Args}

	var compiled *exec.CompiledModule
	if CodeCache != nil {
		cached, ok := CodeCache.Get(contract.Address.ToHexString())
		if ok {
			compiled = cached.(*exec.CompiledModule)
		}
	}

	if compiled == nil {
		module, err := ReadWasmModule(wasmCode, config.NoneVerifyMethod)
		if err != nil {
			return nil, err
		}
		compiled = module
		CodeCache.Add(contract.Address.ToHexString(), compiled)
	}

	vm, err := exec.NewVMWithCompiled(compiled, WASM_MEM_LIMITATION)
	if err != nil {
		return nil, VM_INIT_FAULT
	}

	vm.HostData = host

	vm.ExecMetrics = &exec.Gas{GasLimit: this.GasLimit, LocalGasCounter: 0, GasPrice: this.GasPrice, GasFactor: this.GasFactor, ExecStep: this.ExecStep}
	vm.CallStackDepth = uint32(WASM_CALLSTACK_LIMIT)
	vm.RecoverPanic = true

	entryName := CONTRACT_METHOD_NAME

	entry, ok := compiled.RawModule.Export.Entries[entryName]

	if ok == false {
		return nil, errors.NewErr("[Call]Method:" + entryName + " does not exist!")
	}

	//get entry index
	index := int64(entry.Index)

	//get function index
	fidx := compiled.RawModule.Function.Types[int(index)]

	//get  function type
	ftype := compiled.RawModule.Types.Entries[int(fidx)]

	//no returns of the entry function
	if len(ftype.ReturnTypes) > 0 {
		return nil, errors.NewErr("[Call]ExecCode error! Invoke function sig error")
	}

	//no args for passed in, all args in runtime input buffer
	this.vm = vm

	_, err = vm.ExecCode(index)

	if err != nil {
		return nil, errors.NewErr("[Call]ExecCode error!" + err.Error())
	}

	return host.Output, nil
}
