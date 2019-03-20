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
	"bytes"
	"fmt"
	"reflect"

	"github.com/go-interpreter/wagon/exec"
	"github.com/go-interpreter/wagon/wasm"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/context"
	"github.com/ontio/ontology/smartcontract/event"
	native2 "github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/service/neovm"
	"github.com/ontio/ontology/smartcontract/states"
	neotypes "github.com/ontio/ontology/vm/neovm/types"
)

type ContractType byte

const (
	NATIVE_CONTRACT ContractType = iota
	NEOVM_CONTRACT
	WASMVM_CONTRACT
	UNKOWN_CONTRACT
)

type Runtime struct {
	Service    *WasmVmService
	Input      []byte
	Output     []byte
	CallOutPut []byte
}

func (self *Runtime) TimeStamp(proc *exec.Process) uint64 {
	self.checkGas(TIME_STAMP_GAS)
	return uint64(self.Service.Time)
}

func (self *Runtime) BlockHeight(proc *exec.Process) uint32 {
	self.checkGas(BLOCK_HEGHT_GAS)
	return self.Service.Height
}

func (self *Runtime) SelfAddress(proc *exec.Process, dst uint32) {
	self.checkGas(SELF_ADDRESS_GAS)
	selfaddr := self.Service.ContextRef.CurrentContext().ContractAddress
	_, err := proc.WriteAt(selfaddr[:], int64(dst))
	if err != nil {
		panic(err)
	}
}

func (self *Runtime) CallerAddress(proc *exec.Process, dst uint32) {
	self.checkGas(CALLER_ADDRESS_GAS)
	if self.Service.ContextRef.CallingContext() != nil {
		calleraddr := self.Service.ContextRef.CallingContext().ContractAddress
		_, err := proc.WriteAt(calleraddr[:], int64(dst))
		if err != nil {
			panic(err)
		}
	} else {
		_, err := proc.WriteAt(common.ADDRESS_EMPTY[:], int64(dst))
		if err != nil {
			panic(err)
		}
	}

}

func (self *Runtime) EntryAddress(proc *exec.Process, dst uint32) {
	self.checkGas(ENTRY_ADDRESS_GAS)
	entryAddress := self.Service.ContextRef.EntryContext().ContractAddress
	_, err := proc.WriteAt(entryAddress[:], int64(dst))
	if err != nil {
		panic(err)
	}
}

func (self *Runtime) Checkwitness(proc *exec.Process, dst uint32) uint32 {
	self.checkGas(CHECKWITNESS_GAS)
	var addr common.Address
	_, err := proc.ReadAt(addr[:], int64(dst))
	if err != nil {
		panic(err)
	}

	address, err := common.AddressParseFromBytes(addr[:])
	if err != nil {
		panic(err)
	}

	if self.Service.ContextRef.CheckWitness(address) {
		return 1
	}
	return 0
}

func (self *Runtime) Ret(proc *exec.Process, ptr uint32, len uint32) {
	bs := make([]byte, len)
	_, err := proc.ReadAt(bs, int64(ptr))
	if err != nil {
		panic(err)
	}

	self.Output = bs
	proc.Terminate()
}

func (self *Runtime) Debug(proc *exec.Process, ptr uint32, len uint32) {
	bs := make([]byte, len)
	_, err := proc.ReadAt(bs, int64(ptr))
	if err != nil {
		//do not panic on debug
		return
	}

	//log.Debugf("[WasmContract]Debug:%s\n", bs)
	fmt.Printf("%s", bs)
}

func (self *Runtime) Notify(proc *exec.Process, ptr uint32, len uint32) {
	bs := make([]byte, len)
	_, err := proc.ReadAt(bs, int64(ptr))
	if err != nil {
		panic(err)
	}

	notify := &event.NotifyEventInfo{self.Service.ContextRef.CurrentContext().ContractAddress, string(bs)}
	notifys := make([]*event.NotifyEventInfo, 1)
	notifys[0] = notify
	self.Service.ContextRef.PushNotifications(notifys)
}

func (self *Runtime) InputLength(proc *exec.Process) uint32 {
	return uint32(len(self.Input))
}

func (self *Runtime) GetInput(proc *exec.Process, dst uint32) {
	_, err := proc.WriteAt(self.Input, int64(dst))
	if err != nil {
		panic(err)
	}
}

func (self *Runtime) CallOutputLength(proc *exec.Process) uint32 {
	return uint32(len(self.CallOutPut))
}

func (self *Runtime) GetCallOut(proc *exec.Process, dst uint32) {
	_, err := proc.WriteAt(self.CallOutPut, int64(dst))
	if err != nil {
		panic(err)
	}
}

func (self *Runtime) GetCurrentTxHash(proc *exec.Process, ptr uint32) uint32 {
	self.checkGas(CURRENT_TX_HASH_GAS)

	txhash := self.Service.Tx.Hash()

	length, err := proc.WriteAt(txhash[:], int64(ptr))
	if err != nil {
		panic(err)
	}

	return uint32(length)
}

func (self *Runtime) RaiseException(proc *exec.Process, ptr uint32, len uint32) {
	bs := make([]byte, len)
	_, err := proc.ReadAt(bs, int64(ptr))
	if err != nil {
		//do not panic on debug
		return
	}

	panic(fmt.Errorf("[RaiseException]Contract RaiseException:%s\n", bs))
}

func (self *Runtime) CallContract(proc *exec.Process, contractAddr uint32, inputPtr uint32, inputLen uint32) uint32 {

	self.checkGas(CALL_CONTRACT_GAS)
	contractAddrbytes := make([]byte, 20)
	_, err := proc.ReadAt(contractAddrbytes, int64(contractAddr))
	if err != nil {
		panic(err)
	}

	contractAddress, err := common.AddressParseFromBytes(contractAddrbytes)
	if err != nil {
		panic(err)
	}

	if uint32(proc.MemAllocated()) < inputLen {
		panic(errors.NewErr("inputLen is greater than memory size"))
	}

	inputs := make([]byte, inputLen)
	_, err = proc.ReadAt(inputs, int64(inputPtr))
	if err != nil {
		panic(err)
	}

	contracttype, err := self.getContractType(contractAddress)
	if err != nil {
		panic(err)
	}

	currentCtx := &context.Context{
		Code:            self.Service.Code,
		ContractAddress: self.Service.ContextRef.CurrentContext().ContractAddress,
	}
	self.Service.ContextRef.PushContext(currentCtx)

	var result []byte

	switch contracttype {
	case NATIVE_CONTRACT:
		bf := bytes.NewBuffer(inputs)
		ver, err := serialization.ReadByte(bf)
		if err != nil {
			panic(err)
		}

		method, err := serialization.ReadString(bf)
		if err != nil {
			panic(err)
		}

		args, err := serialization.ReadVarBytes(bf)
		if err != nil {
			panic(err)
		}

		contract := states.ContractInvokeParam{
			Version: ver,
			Address: contractAddress,
			Method:  method,
			Args:    args,
		}

		self.checkGas(NATIVE_INVOKE_GAS)
		native := &native2.NativeService{
			CacheDB:     self.Service.CacheDB,
			InvokeParam: contract,
			Tx:          self.Service.Tx,
			Height:      self.Service.Height,
			Time:        self.Service.Time,
			ContextRef:  self.Service.ContextRef,
			ServiceMap:  make(map[string]native2.Handler),
		}

		tmpRes, err := native.Invoke()
		if err != nil {
			panic(errors.NewErr("[nativeInvoke]AppCall failed:" + err.Error()))
		}

		result = tmpRes.([]byte)

	case WASMVM_CONTRACT:
		conParam := states.WasmContractParam{Address: contractAddress, Args: inputs}
		sink := common.NewZeroCopySink(nil)
		conParam.Serialization(sink)

		newservice, err := self.Service.ContextRef.NewExecuteEngine(sink.Bytes(), types.InvokeWasm)
		if err != nil {
			panic(err)
		}

		tmpRes, err := newservice.Invoke()
		if err != nil {
			panic(err)
		}
		result = tmpRes.([]byte)

	case NEOVM_CONTRACT:
		neoservice, err := self.Service.ContextRef.NewExecuteEngine(inputs, types.InvokeNeo)
		if err != nil {
			panic(err)
		}
		tmp, err := neoservice.Invoke()
		if err != nil {
			panic(err)
		}
		switch tmp.(type) {
		case neotypes.StackItems:
			result, err = tmp.(neotypes.StackItems).GetByteArray()
			if err != nil {
				result, err = neovm.SerializeStackItem(tmp.(neotypes.StackItems))
				if err != nil {
					panic(err)
				}
			}
		default:
			panic(errors.NewErr("Invalid return type of NeoVM"))
		}

	default:
		panic(errors.NewErr("Not a supported contract type"))
	}
	self.Service.ContextRef.PopContext()

	self.CallOutPut = result
	return uint32(len(self.CallOutPut))
}

func NewHostModule(host *Runtime) *wasm.Module {
	m := wasm.NewModule()
	m.Types = &wasm.SectionTypes{
		Entries: []wasm.FunctionSig{
			//func()uint64    [0]
			{
				Form:        0, // value for the 'func' type constructor
				ReturnTypes: []wasm.ValueType{wasm.ValueTypeI64},
			},
			//func()uint32     [1]
			{
				Form:        0, // value for the 'func' type constructor
				ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32},
			},
			//func(uint32)     [2]
			{
				Form:       0, // value for the 'func' type constructor
				ParamTypes: []wasm.ValueType{wasm.ValueTypeI32},
			},
			//func(uint32)uint32  [3]
			{
				Form:        0, // value for the 'func' type constructor
				ParamTypes:  []wasm.ValueType{wasm.ValueTypeI32},
				ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32},
			},
			//func(uint32,uint32)  [4]
			{
				Form:       0, // value for the 'func' type constructor
				ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32},
			},
			//func(uint32,uint32,uint32)uint32  [5]
			{
				Form:        0, // value for the 'func' type constructor
				ParamTypes:  []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32},
				ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32},
			},
			//func(uint32,uint32,uint32,uint32,uint32)uint32  [6]
			{
				Form:        0, // value for the 'func' type constructor
				ParamTypes:  []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32},
				ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32},
			},
			//func(uint32,uint32,uint32,uint32)  [7]
			{
				Form:       0, // value for the 'func' type constructor
				ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32},
			},
			//func(uint32,uint32)uint32   [8]
			{
				Form:        0, // value for the 'func' type constructor
				ParamTypes:  []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32},
				ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32},
			},
			//func(uint32 * 12)uint32   [9]
			{
				Form: 0, // value for the 'func' type constructor
				ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32,
					wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32,
					wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32,
					wasm.ValueTypeI32, wasm.ValueTypeI32},
				ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32},
			},
			//funct()   [10]
			{
				Form: 0, // value for the 'func' type constructor
			},
		},
	}
	m.FunctionIndexSpace = []wasm.Function{
		{ //0
			Sig:  &m.Types.Entries[0],
			Host: reflect.ValueOf(host.TimeStamp),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //1
			Sig:  &m.Types.Entries[1],
			Host: reflect.ValueOf(host.BlockHeight),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //2
			Sig:  &m.Types.Entries[1],
			Host: reflect.ValueOf(host.InputLength),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //3
			Sig:  &m.Types.Entries[1],
			Host: reflect.ValueOf(host.CallOutputLength),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //4
			Sig:  &m.Types.Entries[2],
			Host: reflect.ValueOf(host.SelfAddress),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //5
			Sig:  &m.Types.Entries[2],
			Host: reflect.ValueOf(host.CallerAddress),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //6
			Sig:  &m.Types.Entries[2],
			Host: reflect.ValueOf(host.EntryAddress),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //7
			Sig:  &m.Types.Entries[2],
			Host: reflect.ValueOf(host.GetInput),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //8
			Sig:  &m.Types.Entries[2],
			Host: reflect.ValueOf(host.GetCallOut),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //9
			Sig:  &m.Types.Entries[3],
			Host: reflect.ValueOf(host.Checkwitness),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //10
			Sig:  &m.Types.Entries[3],
			Host: reflect.ValueOf(host.GetCurrentBlockHash),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //11
			Sig:  &m.Types.Entries[3],
			Host: reflect.ValueOf(host.GetCurrentTxHash),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //12
			Sig:  &m.Types.Entries[4],
			Host: reflect.ValueOf(host.Ret),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //13
			Sig:  &m.Types.Entries[4],
			Host: reflect.ValueOf(host.Notify),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //14
			Sig:  &m.Types.Entries[4],
			Host: reflect.ValueOf(host.Debug),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //15
			Sig:  &m.Types.Entries[5],
			Host: reflect.ValueOf(host.CallContract),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //16
			Sig:  &m.Types.Entries[6],
			Host: reflect.ValueOf(host.StorageRead),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //17
			Sig:  &m.Types.Entries[7],
			Host: reflect.ValueOf(host.StorageWrite),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //18
			Sig:  &m.Types.Entries[4],
			Host: reflect.ValueOf(host.StorageDelete),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //19
			Sig:  &m.Types.Entries[9],
			Host: reflect.ValueOf(host.ContractCreate),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //20
			Sig:  &m.Types.Entries[9],
			Host: reflect.ValueOf(host.ContractMigrate),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //21
			Sig:  &m.Types.Entries[10],
			Host: reflect.ValueOf(host.ContractDelete),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //22
			Sig:  &m.Types.Entries[4],
			Host: reflect.ValueOf(host.RaiseException),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
	}

	m.Export = &wasm.SectionExports{
		Entries: map[string]wasm.ExportEntry{
			"timestamp": {
				FieldStr: "timestamp",
				Kind:     wasm.ExternalFunction,
				Index:    0,
			},
			"block_height": {
				FieldStr: "block_height",
				Kind:     wasm.ExternalFunction,
				Index:    1,
			},
			"input_length": {
				FieldStr: "input_length",
				Kind:     wasm.ExternalFunction,
				Index:    2,
			},
			"call_output_length": {
				FieldStr: "call_output_length",
				Kind:     wasm.ExternalFunction,
				Index:    3,
			},
			"self_address": {
				FieldStr: "self_address",
				Kind:     wasm.ExternalFunction,
				Index:    4,
			},
			"caller_address": {
				FieldStr: "caller_address",
				Kind:     wasm.ExternalFunction,
				Index:    5,
			},
			"entry_address": {
				FieldStr: "entry_address",
				Kind:     wasm.ExternalFunction,
				Index:    6,
			},
			"get_input": {
				FieldStr: "get_input",
				Kind:     wasm.ExternalFunction,
				Index:    7,
			},
			"get_call_output": {
				FieldStr: "get_call_output",
				Kind:     wasm.ExternalFunction,
				Index:    8,
			},
			"check_witness": {
				FieldStr: "check_witness",
				Kind:     wasm.ExternalFunction,
				Index:    9,
			},
			"current_blockhash": {
				FieldStr: "current_blockhash",
				Kind:     wasm.ExternalFunction,
				Index:    10,
			},
			"current_txhash": {
				FieldStr: "current_txhash",
				Kind:     wasm.ExternalFunction,
				Index:    11,
			},
			"ret": {
				FieldStr: "ret",
				Kind:     wasm.ExternalFunction,
				Index:    12,
			},
			"notify": {
				FieldStr: "notify",
				Kind:     wasm.ExternalFunction,
				Index:    13,
			},
			"debug": {
				FieldStr: "debug",
				Kind:     wasm.ExternalFunction,
				Index:    14,
			},
			"call_contract": {
				FieldStr: "call_contract",
				Kind:     wasm.ExternalFunction,
				Index:    15,
			},
			"storage_read": {
				FieldStr: "storage_read",
				Kind:     wasm.ExternalFunction,
				Index:    16,
			},
			"storage_write": {
				FieldStr: "storage_write",
				Kind:     wasm.ExternalFunction,
				Index:    17,
			},
			"storage_delete": {
				FieldStr: "storage_delete",
				Kind:     wasm.ExternalFunction,
				Index:    18,
			},
			"contract_create": {
				FieldStr: "contract_create",
				Kind:     wasm.ExternalFunction,
				Index:    19,
			},
			"contract_migrate": {
				FieldStr: "contract_migrate",
				Kind:     wasm.ExternalFunction,
				Index:    20,
			},
			"contract_delete": {
				FieldStr: "contract_delete",
				Kind:     wasm.ExternalFunction,
				Index:    21,
			},
			"panic": {
				FieldStr: "panic",
				Kind:     wasm.ExternalFunction,
				Index:    22,
			},
		},
	}

	return m
}

func (self *Runtime) getContractType(addr common.Address) (ContractType, error) {
	if utils.IsNativeContract(addr) {
		return NATIVE_CONTRACT, nil
	}

	dep, err := self.Service.CacheDB.GetContract(addr)
	if err != nil {
		return UNKOWN_CONTRACT, err
	}
	if dep == nil {
		return UNKOWN_CONTRACT, errors.NewErr("contract is not exist.")
	}
	if dep.VmType == payload.WASMVM_TYPE {
		return WASMVM_CONTRACT, nil
	}

	return NEOVM_CONTRACT, nil

}

func (self *Runtime) checkGas(gaslimit uint64) {
	gas := self.Service.vm.AvaliableGas
	if gas.GasLimit >= gaslimit {
		gas.GasLimit -= gaslimit
	} else {
		panic(errors.NewErr("[wasm_Service]Insufficient gas limit"))
	}
}

func serializeStorageKey(contractAddress common.Address, key []byte) []byte {
	bf := new(bytes.Buffer)

	bf.Write(contractAddress[:])
	bf.Write(key)

	return bf.Bytes()
}
