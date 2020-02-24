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

/*
#cgo CFLAGS: -I.
#cgo darwin LDFLAGS: -L. -lwasmjit_onto_interface_darwin -ldl -lc -lm
#cgo linux LDFLAGS: -L. -lwasmjit_onto_interface -ldl -lc -lm
#cgo windows LDFLAGS: -Wl,-rpath,${SRCDIR} -L. -lwasmjit_onto_interface
#include "wasmjit_runtime.h"
#include <stdlib.h>
*/
import "C"

import (
	"math"
	"unsafe"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/payload"
	states2 "github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/states"
)

const (
	wasmjit_result_success      uint   = 0
	wasmjit_result_err_internal uint   = 1
	wasmjit_result_err_compile  uint   = 2
	wasmjit_result_err_link     uint   = 3
	wasmjit_result_err_trap     uint   = 4
	wasmjit_gas_mod             uint64 = 200
)

func WasmjitValidate(wasmCode []byte) error {
	codeSlice := C.wasmjit_slice_t{data: (*C.uint8_t)((unsafe.Pointer)(&wasmCode[0])), len: C.uint32_t(len(wasmCode))}
	result := C.wasmjit_validate(codeSlice)
	if result.kind != C.wasmjit_result_kind(wasmjit_result_success) {
		err := errors.NewErr(C.GoStringN((*C.char)((unsafe.Pointer)(result.msg.data)), C.int(result.msg.len)))
		C.wasmjit_bytes_destroy(result.msg)
		return err
	}

	return nil
}

func jitSliceToBytes(slice C.wasmjit_slice_t) []byte {
	return C.GoBytes((unsafe.Pointer)(slice.data), C.int(slice.len))
}

func jitErr(err error) C.wasmjit_result_t {
	s := err.Error()
	ptr := []byte(s)
	l := len(s)
	result := C.wasmjit_construct_result((*C.uint8_t)((unsafe.Pointer)(&ptr[0])), (C.uint32_t)(l), C.wasmjit_result_kind(wasmjit_result_err_trap))
	return result
}

func jitService(vmctx *C.wasmjit_vmctx_t) *WasmVmService {
	index := C.wasmjit_service_index(vmctx)
	return getWasmVmService(uint64(index))
}

func setCallOutPut(vmctx *C.wasmjit_vmctx_t, result []byte) {
	var output *C.uint8_t
	if len(result) != 0 {
		output = (*C.uint8_t)((unsafe.Pointer)(&result[0]))
	} else {
		output = (*C.uint8_t)((unsafe.Pointer)(nil))
	}
	C.wasmjit_set_calloutput(vmctx, output, C.uint32_t(len(result)))
}

func jitContractCreate(serviceIndex C.uint64_t,
	codeSlice C.wasmjit_slice_t,
	vmType C.uint32_t,
	nameSlice C.wasmjit_slice_t,
	verSlice C.wasmjit_slice_t,
	authorSlice C.wasmjit_slice_t,
	emailSlice C.wasmjit_slice_t,
	descSlice C.wasmjit_slice_t,
	newAddress *C.address_t,
) (C.wasmjit_result_t, *WasmVmService, common.Address) {
	service := getWasmVmService(uint64(serviceIndex))

	code := jitSliceToBytes(codeSlice)
	name := jitSliceToBytes(nameSlice)
	version := jitSliceToBytes(verSlice)
	author := jitSliceToBytes(authorSlice)
	email := jitSliceToBytes(emailSlice)
	desc := jitSliceToBytes(descSlice)

	dep, errs := payload.CreateDeployCode(code, uint32(vmType), name, version, author, email, desc)
	if errs != nil {
		return jitErr(errs), nil, common.ADDRESS_EMPTY
	}

	wasmCode, errs := dep.GetWasmCode()
	if errs != nil {
		return jitErr(errs), nil, common.ADDRESS_EMPTY
	}

	errs = WasmjitValidate(wasmCode)
	if errs != nil {
		return jitErr(errs), nil, common.ADDRESS_EMPTY
	}

	contractAddr := dep.Address()

	item, errs := service.CacheDB.GetContract(contractAddr)
	if errs != nil {
		return jitErr(errs), nil, common.ADDRESS_EMPTY
	}

	if item != nil {
		return jitErr(errors.NewErr("contract has been deployed")), nil, common.ADDRESS_EMPTY
	}

	service.CacheDB.PutContract(dep)
	C.memcpy((unsafe.Pointer)(newAddress), ((unsafe.Pointer)(&contractAddr[0])), C.ulong(20))
	return C.wasmjit_result_t{kind: C.wasmjit_result_kind(wasmjit_result_success)}, service, contractAddr
}

// c to call go interface

//export ontio_contract_create_cgo
func ontio_contract_create_cgo(serviceIndex C.uint64_t,
	codeSlice C.wasmjit_slice_t,
	vmType C.uint32_t,
	nameSlice C.wasmjit_slice_t,
	verSlice C.wasmjit_slice_t,
	authorSlice C.wasmjit_slice_t,
	emailSlice C.wasmjit_slice_t,
	descSlice C.wasmjit_slice_t,
	newAddress *C.address_t,
) C.wasmjit_result_t {
	cResult, _, _ := jitContractCreate(serviceIndex, codeSlice, vmType, nameSlice, verSlice, authorSlice, emailSlice, descSlice, newAddress)
	return cResult
}

//export ontio_contract_migrate_cgo
func ontio_contract_migrate_cgo(serviceIndex C.uint64_t,
	codeSlice C.wasmjit_slice_t,
	vmType C.uint32_t,
	nameSlice C.wasmjit_slice_t,
	verSlice C.wasmjit_slice_t,
	authorSlice C.wasmjit_slice_t,
	emailSlice C.wasmjit_slice_t,
	descSlice C.wasmjit_slice_t,
	newAddress *C.address_t,
) C.wasmjit_result_t {
	cResult, service, contractAddr := jitContractCreate(serviceIndex, codeSlice, vmType, nameSlice, verSlice, authorSlice, emailSlice, descSlice, newAddress)
	if cResult.kind != C.wasmjit_result_kind(wasmjit_result_success) {
		return cResult
	}

	errs := migrateContractStorage(service, contractAddr)
	if errs != nil {
		return jitErr(errs)
	}

	return C.wasmjit_result_t{kind: C.wasmjit_result_kind(wasmjit_result_success)}
}

//export ontio_contract_destroy_cgo
func ontio_contract_destroy_cgo(service_index C.uint64_t) C.wasmjit_result_t {
	service := getWasmVmService(uint64(service_index))

	errs := deleteContractStorage(service)
	if errs != nil {
		return jitErr(errs)
	}

	return C.wasmjit_result_t{kind: C.wasmjit_result_kind(wasmjit_result_success)}
}

//export ontio_storage_read_cgo
func ontio_storage_read_cgo(serviceIndex C.uint64_t, keySlice C.wasmjit_slice_t, valSlice C.wasmjit_slice_t, offset C.uint32_t) C.wasmjit_u32 {
	service := getWasmVmService(uint64(serviceIndex))

	keybytes := jitSliceToBytes(keySlice)

	itemWrite, originLen, err := storageRead(service, keybytes, uint32(keySlice.len), uint32(valSlice.len), uint32(offset))
	if err != nil {
		return C.wasmjit_u32{v: 0, res: jitErr(err)}
	}

	if originLen != math.MaxUint32 {
		C.memcpy((unsafe.Pointer)(valSlice.data), ((unsafe.Pointer)(&itemWrite[0])), C.ulong(len(itemWrite)))
	}

	return C.wasmjit_u32{v: C.uint32_t(originLen), res: C.wasmjit_result_t{kind: C.wasmjit_result_kind(wasmjit_result_success)}}
}

//export ontio_storage_write_cgo
func ontio_storage_write_cgo(service_index C.uint64_t, key_s C.wasmjit_slice_t, val_s C.wasmjit_slice_t) {
	service := getWasmVmService(uint64(service_index))
	keybytes := jitSliceToBytes(key_s)
	valbytes := jitSliceToBytes(val_s)

	key := serializeStorageKey(service.ContextRef.CurrentContext().ContractAddress, keybytes)
	service.CacheDB.Put(key, states2.GenRawStorageItem(valbytes))
}

//export ontio_storage_delete_cgo
func ontio_storage_delete_cgo(service_index C.uint64_t, key_s C.wasmjit_slice_t) {
	service := getWasmVmService(uint64(service_index))
	keybytes := jitSliceToBytes(key_s)

	key := serializeStorageKey(service.ContextRef.CurrentContext().ContractAddress, keybytes)
	service.CacheDB.Delete(key)
}

//export ontio_notify_cgo
func ontio_notify_cgo(service_index C.uint64_t, data C.wasmjit_slice_t) C.wasmjit_result_t {
	service := getWasmVmService(uint64(service_index))
	bs := jitSliceToBytes(data)

	err := notify(service, bs)
	if err != nil {
		return jitErr(err)
	}
	return C.wasmjit_result_t{kind: C.wasmjit_result_kind(wasmjit_result_success)}
}

//export ontio_debug_cgo
func ontio_debug_cgo(data C.wasmjit_slice_t) {
	bs := jitSliceToBytes(data)
	debugLog(bs)
}

//export ontio_call_contract_cgo
func ontio_call_contract_cgo(vmctx *C.wasmjit_vmctx_t, contractAddr *C.address_t, input C.wasmjit_slice_t) C.wasmjit_result_t {
	var contractAddress common.Address

	service := jitService(vmctx)

	exec_step := C.wasmjit_get_exec_step(vmctx)
	gas_left := C.wasmjit_get_gas(vmctx)
	*service.ExecStep = uint64(exec_step)
	*service.GasLimit = uint64(gas_left)

	buff := jitSliceToBytes(C.wasmjit_slice_t{data: ((*C.uint8_t)((unsafe.Pointer)(contractAddr))), len: 20})

	copy(contractAddress[:], buff[:])

	inputs := jitSliceToBytes(input)

	result, errs := callContractInner(service, contractAddress, inputs)
	// here need set the Gas back to JIT context before err return.
	C.wasmjit_set_gas(vmctx, C.uint64_t(*service.GasLimit))
	C.wasmjit_set_exec_step(vmctx, C.uint64_t(*service.ExecStep))
	if errs != nil {
		return jitErr(errs)
	}

	setCallOutPut(vmctx, result)
	return C.wasmjit_result_t{kind: C.wasmjit_result_kind(wasmjit_result_success)}
}

func destroyWasmjitRet(ret C.wasmjit_ret) {
	buffer := ret.buffer
	msg := ret.res.msg
	if buffer.data != (*C.uint8_t)((unsafe.Pointer)(nil)) {
		C.wasmjit_bytes_destroy(buffer)
	}

	if msg.data != (*C.uint8_t)((unsafe.Pointer)(nil)) {
		C.wasmjit_bytes_destroy(msg)
	}
}

// call to c
func invokeJit(this *WasmVmService, contract *states.WasmContractParam, wasmCode []byte) ([]byte, error) {
	index := registerWasmVmService(this)
	defer unregisterWasmVmService(index)

	txHash := this.Tx.Hash()
	witnessAddrBuff, witness_len := GetAddressBuff(this.Tx.GetSignatureAddresses())
	callersAddrBuff, callers_len := GetAddressBuff(this.ContextRef.GetCallerAddress())

	var witnessPtr, callersPtr, inputPtr *C.uint8_t

	if witness_len == 0 {
		witnessPtr = (*C.uint8_t)((unsafe.Pointer)(nil))
	} else {
		witnessPtr = (*C.uint8_t)((unsafe.Pointer)(&witnessAddrBuff[0]))
	}

	if callers_len == 0 {
		callersPtr = (*C.uint8_t)((unsafe.Pointer)(nil))
	} else {
		callersPtr = (*C.uint8_t)((unsafe.Pointer)(&callersAddrBuff[0]))
	}

	input_len := len(contract.Args)
	if len(contract.Args) == 0 {
		inputPtr = (*C.uint8_t)((unsafe.Pointer)(nil))
	} else {
		inputPtr = (*C.uint8_t)((unsafe.Pointer)(&contract.Args[0]))
	}

	height := C.uint32_t(this.Height)
	block_hash := (*C.h256_t)((unsafe.Pointer)(&this.BlockHash[0]))
	timestamp := C.uint64_t(this.Time)
	tx_hash := (*C.h256_t)((unsafe.Pointer)(&(txHash[0])))
	caller_raw := C.wasmjit_slice_t{data: callersPtr, len: C.uint32_t(callers_len)}
	witness_raw := C.wasmjit_slice_t{data: witnessPtr, len: C.uint32_t(witness_len)}
	input_raw := C.wasmjit_slice_t{data: inputPtr, len: C.uint32_t(input_len)}
	service_index := C.uint64_t(this.ServiceIndex)
	exec_step := C.uint64_t(*this.ExecStep)
	gas_factor := C.uint64_t(this.GasFactor)
	gas_left := C.uint64_t(*this.GasLimit)
	depth_left := C.uint64_t(WASM_CALLSTACK_LIMIT)
	codeSlice := C.wasmjit_slice_t{data: (*C.uint8_t)((unsafe.Pointer)(&wasmCode[0])), len: C.uint32_t(len(wasmCode))}

	ctx := C.wasmjit_chain_context_create(height, block_hash, timestamp, tx_hash, caller_raw, witness_raw, input_raw, exec_step, gas_factor, gas_left, depth_left, service_index)
	jit_ret := C.wasmjit_invoke(codeSlice, ctx)
	*this.ExecStep = uint64(jit_ret.exec_step)
	*this.GasLimit = uint64(jit_ret.gas_left)

	if jit_ret.res.kind != C.wasmjit_result_kind(wasmjit_result_success) {
		err := errors.NewErr(C.GoStringN((*C.char)((unsafe.Pointer)(jit_ret.res.msg.data)), C.int(jit_ret.res.msg.len)))
		destroyWasmjitRet(jit_ret)
		if jit_ret.res.kind != C.wasmjit_result_kind(wasmjit_result_err_trap) {
			this.ContextRef.SetInternalErr()
		}
		return nil, err
	}

	output := C.GoBytes((unsafe.Pointer)(jit_ret.buffer.data), (C.int)(jit_ret.buffer.len))
	destroyWasmjitRet(jit_ret)
	return output, nil
}
