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
	"github.com/go-interpreter/wagon/exec"
	"github.com/go-interpreter/wagon/wasm"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/event"
	native2 "github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/states"
	"reflect"
	"encoding/gob"
)

type ContractType byte

const (
	NATIVE_CONTRACT ContractType = iota
	NEOVM_CONTRACT
	WASMVM_CONTRACT
	UNKOWN_CONTRACT
)

type Runtime struct {
	Service *WasmVmService
	Input   []byte
	Output  []byte
}

func (self *Runtime) TimeStamp(proc *exec.Process, v uint32) uint64 {
	return uint64(self.Service.Time)
}

func (self *Runtime) BlockHeight(proc *exec.Process) uint32 {
	return self.Service.Height
}

func (self *Runtime) SelfAddress(proc *exec.Process, dst uint32) {
	selfaddr := self.Service.ContextRef.CurrentContext().ContractAddress
	_, err := proc.WriteAt(selfaddr[:], int64(dst))
	if err != nil {
		panic(err)
	}
}

func (self *Runtime) Calleraddress(proc *exec.Process, dst uint32) {
	calleraddr := self.Service.ContextRef.CallingContext().ContractAddress
	_, err := proc.WriteAt(calleraddr[:], int64(dst))
	if err != nil {
		panic(err)
	}
}

func (self *Runtime) Checkwitness(proc *exec.Process, dst uint32) uint32 {
	addrbytes := make([]byte, 20)
	_, err := proc.ReadAt(addrbytes, int64(dst))
	if err != nil {
		panic(err)
	}

	address, err := common.AddressParseFromBytes(addrbytes)
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

	self.Output = make([]byte, len)
	copy(self.Output, bs)
}

func (self *Runtime) Notify(proc *exec.Process, ptr uint32, len uint32) {
	bs := make([]byte, len)
	_, err := proc.ReadAt(bs, int64(ptr))
	if err != nil {
		panic(err)
	}

	notify := &event.NotifyEventInfo{self.Service.ContextRef.CurrentContext().ContractAddress, bs}
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

func (self *Runtime) Call_contract(proc *exec.Process, contractAddr uint32, inputPtr uint32, inputLen uint32,outPtr uint32) uint32 {
	contractAddrbytes := make([]byte, 20)
	_, err := proc.ReadAt(contractAddrbytes, int64(contractAddr))
	if err != nil {
		panic(err)
	}

	contractAddress, err := common.AddressParseFromBytes(contractAddrbytes)
	if err != nil {
		panic(err)
	}

	inputs := make([]byte, inputLen)
	_, err = proc.ReadAt(inputs, int64(inputPtr))
	if err != nil {
		panic(err)
	}

	bf := bytes.NewBuffer(inputs)
	ver, err := serialization.ReadUint32(bf)
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

	contracttype, err := self.getContractType(contractAddress)
	if err != nil {
		panic(err)
	}
	switch contracttype {
	case NATIVE_CONTRACT:
		cip := states.ContractInvokeParam{
			Version: byte(ver),
			Address: contractAddress,
			Method:  method,
			Args:    args,
		}

		native := &native2.NativeService{
			CacheDB:     self.Service.CacheDB,
			InvokeParam: cip,
			Tx:          self.Service.Tx,
			Height:      self.Service.Height,
			Time:        self.Service.Time,
			ContextRef:  self.Service.ContextRef,
			ServiceMap:  make(map[string]native2.Handler),
		}
		result, err := native.Invoke()
		if err != nil {
			panic(errors.NewErr("[nativeInvoke]AppCall failed:" + err.Error()))
		}

		buf := bytes.NewBuffer(nil)
		enc := gob.NewEncoder(buf)
		err = enc.Encode(result)
		if err != nil {
			panic(errors.NewErr("[nativeInvoke]AppCall failed:" + err.Error()))
		}

		proc.WriteAt(buf.Bytes(),int64(outPtr))
		return outPtr

	case NEOVM_CONTRACT:

	case WASMVM_CONTRACT:

	default:
		panic(errors.NewErr("Not a supported contract type"))
	}

}

func NewHostModule(host *Runtime) *wasm.Module {
	m := wasm.NewModule()
	m.Types = &wasm.SectionTypes{
		Entries: []wasm.FunctionSig{
			{
				Form:       0, // value for the 'func' type constructor
				ParamTypes: []wasm.ValueType{wasm.ValueTypeI32},
			},
			{
				Form:        0, // value for the 'func' type constructor
				ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32},
			},
		},
	}
	m.FunctionIndexSpace = []wasm.Function{
		{
			Sig:  &m.Types.Entries[0],
			Host: reflect.ValueOf(host.Print),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{
			Sig:  &m.Types.Entries[0],
			Host: reflect.ValueOf(host.GetInput),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{
			Sig:  &m.Types.Entries[1],
			Host: reflect.ValueOf(host.BlockHeight),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{
			Sig:  &m.Types.Entries[1],
			Host: reflect.ValueOf(host.InputLength),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
	}

	m.Export = &wasm.SectionExports{
		Entries: map[string]wasm.ExportEntry{
			"print": {
				FieldStr: "print",
				Kind:     wasm.ExternalFunction,
				Index:    0,
			},
			"get_input": {
				FieldStr: "get_input",
				Kind:     wasm.ExternalFunction,
				Index:    1,
			},
			"block_height": {
				FieldStr: "block_height",
				Kind:     wasm.ExternalFunction,
				Index:    2,
			},
			"input_length": {
				FieldStr: "input_length",
				Kind:     wasm.ExternalFunction,
				Index:    3,
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

	if dep.NeedStorage == byte(3) {
		return WASMVM_CONTRACT, nil
	}

	return NEOVM_CONTRACT, nil

}
