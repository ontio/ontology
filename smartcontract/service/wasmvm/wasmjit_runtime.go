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
#cgo LDFLAGS: -L. -lwasmjit_onto_interface -ldl -lc -lm
#include "wasmjit_runtime.h"
#include <stdlib.h>
*/
import "C"

import (
	"io"
	"math"
	"unsafe"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/payload"
	states2 "github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/event"
	native2 "github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/service/util"
	"github.com/ontio/ontology/smartcontract/states"
	"github.com/ontio/ontology/vm/crossvm_codec"
	neotypes "github.com/ontio/ontology/vm/neovm/types"
)

const (
	wasmjit_result_success      uint   = 0
	wasmjit_result_err_internal uint   = 1
	wasmjit_result_err_compile  uint   = 2
	wasmjit_result_err_link     uint   = 3
	wasmjit_result_err_trap     uint   = 4
	wasmjit_gas_mod             uint64 = 200
)

func getContractType(Service *WasmVmService, addr common.Address) (ContractType, error) {
	if utils.IsNativeContract(addr) {
		return NATIVE_CONTRACT, nil
	}

	dep, err := Service.CacheDB.GetContract(addr)
	if err != nil {
		return UNKOWN_CONTRACT, err
	}
	if dep == nil {
		return UNKOWN_CONTRACT, errors.NewErr("contract is not exist.")
	}
	if dep.VmType() == payload.WASMVM_TYPE {
		return WASMVM_CONTRACT, nil
	}

	return NEOVM_CONTRACT, nil
}

func jitSliceToBytes(slice C.wasmjit_slice_t) []byte {
	return C.GoBytes((unsafe.Pointer)(slice.data), C.int(slice.len))
}

func jitSliceWrite(data []byte, slice C.wasmjit_slice_t) {
	if len(data) == 0 {
		return
	}

	C.memcpy((unsafe.Pointer)(slice.data), ((unsafe.Pointer)(&data[0])), C.ulong(slice.len))
}

func jitErr(err error) C.wasmjit_result_t {
	s := err.Error()
	ptr := []byte(s)
	l := len(s)
	result := C.wasmjit_construct_result((*C.uchar)((unsafe.Pointer)(&ptr[0])), (C.uint)(l), C.uint(wasmjit_result_err_trap))
	return result
}

func jitService(vmctx *C.wasmjit_vmctx_t) *WasmVmService {
	index := C.wasmjit_service_index(vmctx)
	return GetWasmVmService(uint64(index))
}

func setCallOutPut(vmctx *C.wasmjit_vmctx_t, result []byte) {
	var output *C.uchar
	if len(result) != 0 {
		output = (*C.uchar)((unsafe.Pointer)(&result[0]))
	} else {
		output = (*C.uchar)((unsafe.Pointer)(nil))
	}
	C.wasmjit_set_call_output(vmctx, output, C.uint(len(result)))
}

// c to call go interface

//export ontio_contract_create_cgo
func ontio_contract_create_cgo(service_index C.ulonglong,
	code_s C.wasmjit_slice_t,
	vmType uint32,
	name_s C.wasmjit_slice_t,
	ver_s C.wasmjit_slice_t,
	author_s C.wasmjit_slice_t,
	email_s C.wasmjit_slice_t,
	desc_s C.wasmjit_slice_t,
	newaddress *C.address_t,
) C.wasmjit_result_t {
	Service := GetWasmVmService(uint64(service_index))

	code := jitSliceToBytes(code_s)
	name := jitSliceToBytes(name_s)
	version := jitSliceToBytes(ver_s)
	author := jitSliceToBytes(author_s)
	email := jitSliceToBytes(email_s)
	desc := jitSliceToBytes(desc_s)

	dep, errs := payload.CreateDeployCode(code, vmType, name, version, author, email, desc)
	if errs != nil {
		return jitErr(errs)
	}

	wasmCode, errs := dep.GetWasmCode()
	if errs != nil {
		return jitErr(errs)
	}

	_, errs = ReadWasmModule(wasmCode, true)
	if errs != nil {
		return jitErr(errs)
	}

	contractAddr := dep.Address()

	item, errs := Service.CacheDB.GetContract(contractAddr)
	if errs != nil {
		return jitErr(errs)
	}

	if item != nil {
		return jitErr(errors.NewErr("contract has been deployed"))
	}

	Service.CacheDB.PutContract(dep)

	C.memcpy((unsafe.Pointer)(newaddress), ((unsafe.Pointer)(&contractAddr[0])), C.ulong(20))

	return C.wasmjit_result_t{kind: C.uint(wasmjit_result_success)}
}

//export ontio_contract_migrate_cgo
func ontio_contract_migrate_cgo(service_index C.ulonglong,
	code_s C.wasmjit_slice_t,
	vmType uint32,
	name_s C.wasmjit_slice_t,
	ver_s C.wasmjit_slice_t,
	author_s C.wasmjit_slice_t,
	email_s C.wasmjit_slice_t,
	desc_s C.wasmjit_slice_t,
	newaddress *C.address_t,
) C.wasmjit_result_t {
	Service := GetWasmVmService(uint64(service_index))

	code := jitSliceToBytes(code_s)
	name := jitSliceToBytes(name_s)
	version := jitSliceToBytes(ver_s)
	author := jitSliceToBytes(author_s)
	email := jitSliceToBytes(email_s)
	desc := jitSliceToBytes(desc_s)

	dep, errs := payload.CreateDeployCode(code, vmType, name, version, author, email, desc)
	if errs != nil {
		return jitErr(errs)
	}

	wasmCode, errs := dep.GetWasmCode()
	if errs != nil {
		return jitErr(errs)
	}
	_, errs = ReadWasmModule(wasmCode, true)
	if errs != nil {
		return jitErr(errs)
	}

	contractAddr := dep.Address()

	item, errs := Service.CacheDB.GetContract(contractAddr)
	if errs != nil {
		return jitErr(errs)
	}

	if item != nil {
		return jitErr(errors.NewErr("contract has been deployed"))
	}

	oldAddress := Service.ContextRef.CurrentContext().ContractAddress

	Service.CacheDB.PutContract(dep)
	Service.CacheDB.DeleteContract(oldAddress)

	iter := Service.CacheDB.NewIterator(oldAddress[:])
	for has := iter.First(); has; has = iter.Next() {
		key := iter.Key()
		val := iter.Value()

		newkey := serializeStorageKey(contractAddr, key[20:])

		Service.CacheDB.Put(newkey, val)
		Service.CacheDB.Delete(key)
	}

	iter.Release()
	if errs := iter.Error(); errs != nil {
		return jitErr(errs)
	}

	C.memcpy((unsafe.Pointer)(newaddress), ((unsafe.Pointer)(&contractAddr[0])), C.ulong(20))

	return C.wasmjit_result_t{kind: C.uint(wasmjit_result_success)}
}

//export ontio_contract_destroy_cgo
func ontio_contract_destroy_cgo(service_index C.ulonglong) C.wasmjit_result_t {
	Service := GetWasmVmService(uint64(service_index))

	contractAddress := Service.ContextRef.CurrentContext().ContractAddress
	iter := Service.CacheDB.NewIterator(contractAddress[:])

	for has := iter.First(); has; has = iter.Next() {
		Service.CacheDB.Delete(iter.Key())
	}
	iter.Release()
	if errs := iter.Error(); errs != nil {
		return jitErr(errs)
	}

	Service.CacheDB.DeleteContract(contractAddress)

	return C.wasmjit_result_t{kind: C.uint(wasmjit_result_success)}
}

//export ontio_storage_read_cgo
func ontio_storage_read_cgo(service_index C.ulonglong, key_s C.wasmjit_slice_t, val_s C.wasmjit_slice_t, offset uint32) C.wasmjit_u32 {
	Service := GetWasmVmService(uint64(service_index))

	keybytes := jitSliceToBytes(key_s)

	key := serializeStorageKey(Service.ContextRef.CurrentContext().ContractAddress, keybytes)

	raw, errs := Service.CacheDB.Get(key)
	if errs != nil {
		return C.wasmjit_u32{v: 0, res: jitErr(errs)}
	}

	if raw == nil {
		return C.wasmjit_u32{v: C.uint(math.MaxUint32), res: C.wasmjit_result_t{kind: C.uint(wasmjit_result_success)}}
	}

	item, errs := states2.GetValueFromRawStorageItem(raw)
	if errs != nil {
		return C.wasmjit_u32{v: 0, res: jitErr(errs)}
	}

	length := uint32(val_s.len)
	itemlen := uint32(len(item))
	if itemlen < uint32(val_s.len) {
		length = itemlen // choose the smaller one. so C.memcpy is safe.
	}

	if uint32(len(item)) < offset {
		return C.wasmjit_u32{v: 0, res: jitErr(errors.NewErr("offset is invalid"))}
	}

	item_s := item[offset : offset+length]
	C.memcpy((unsafe.Pointer)(val_s.data), ((unsafe.Pointer)(&item_s[0])), C.ulong(length))

	return C.wasmjit_u32{v: C.uint(len(item)), res: C.wasmjit_result_t{kind: C.uint(wasmjit_result_success)}}
}

//export ontio_storage_write_cgo
func ontio_storage_write_cgo(service_index C.ulonglong, key_s C.wasmjit_slice_t, val_s C.wasmjit_slice_t) {
	Service := GetWasmVmService(uint64(service_index))

	keybytes := jitSliceToBytes(key_s)

	valbytes := jitSliceToBytes(val_s)

	key := serializeStorageKey(Service.ContextRef.CurrentContext().ContractAddress, keybytes)

	Service.CacheDB.Put(key, states2.GenRawStorageItem(valbytes))
}

//export ontio_storage_delete_cgo
func ontio_storage_delete_cgo(service_index C.ulonglong, key_s C.wasmjit_slice_t) {
	Service := GetWasmVmService(uint64(service_index))

	//self.checkGas(STORAGE_DELETE_GAS)

	keybytes := jitSliceToBytes(key_s)

	key := serializeStorageKey(Service.ContextRef.CurrentContext().ContractAddress, keybytes)

	Service.CacheDB.Delete(key)
}

//export ontio_notify_cgo
func ontio_notify_cgo(service_index C.ulonglong, data C.wasmjit_slice_t) C.wasmjit_result_t {
	if uint(data.len) >= neotypes.MAX_NOTIFY_LENGTH {
		return jitErr(errors.NewErr("notify length over the uplimit"))
	}

	Service := GetWasmVmService(uint64(service_index))

	bs := jitSliceToBytes(data)

	notify := &event.NotifyEventInfo{ContractAddress: Service.ContextRef.CurrentContext().ContractAddress}
	val := crossvm_codec.DeserializeNotify(bs)
	notify.States = val

	notifys := make([]*event.NotifyEventInfo, 1)
	notifys[0] = notify
	Service.ContextRef.PushNotifications(notifys)

	return C.wasmjit_result_t{kind: C.uint(wasmjit_result_success)}
}

//export ontio_debug_cgo
func ontio_debug_cgo(data C.wasmjit_slice_t) {
	bs := jitSliceToBytes(data)

	log.Infof("[WasmContract]Debug:%s\n", bs)
}

//export ontio_call_contract_cgo
func ontio_call_contract_cgo(vmctx *C.wasmjit_vmctx_t, contractAddr *C.address_t, input C.wasmjit_slice_t) C.wasmjit_result_t {
	var contractAddress common.Address

	Service := jitService(vmctx)

	exec_step := C.wasmjit_get_exec_step(vmctx)
	gas_left := C.wasmjit_get_gas(vmctx)
	*Service.ExecStep = uint64(exec_step)
	*Service.GasLimit = uint64(gas_left)

	buff := jitSliceToBytes(C.wasmjit_slice_t{data: ((*C.uchar)((unsafe.Pointer)(contractAddr))), len: 20})

	copy(contractAddress[:], buff[:])

	inputs := jitSliceToBytes(input)

	contracttype, errs := getContractType(Service, contractAddress)
	if errs != nil {
		return jitErr(errs)
	}

	var result []byte

	switch contracttype {
	case NATIVE_CONTRACT:
		source := common.NewZeroCopySource(inputs)
		ver, eof := source.NextByte()
		if eof {
			return jitErr(io.ErrUnexpectedEOF)
		}
		method, _, irregular, eof := source.NextString()
		if irregular {
			return jitErr(common.ErrIrregularData)
		}
		if eof {
			return jitErr(io.ErrUnexpectedEOF)
		}

		args, _, irregular, eof := source.NextVarBytes()
		if irregular {
			return jitErr(common.ErrIrregularData)
		}
		if eof {
			return jitErr(io.ErrUnexpectedEOF)
		}

		contract := states.ContractInvokeParam{
			Version: ver,
			Address: contractAddress,
			Method:  method,
			Args:    args,
		}

		native := &native2.NativeService{
			CacheDB:     Service.CacheDB,
			InvokeParam: contract,
			Tx:          Service.Tx,
			Height:      Service.Height,
			Time:        Service.Time,
			ContextRef:  Service.ContextRef,
			ServiceMap:  make(map[string]native2.Handler),
		}

		tmpRes, err := native.Invoke()
		C.wasmjit_set_gas(vmctx, C.ulong(*Service.GasLimit))
		C.wasmjit_set_exec_step(vmctx, C.ulong(*Service.ExecStep))
		if err != nil {
			return jitErr(errors.NewErr("[nativeInvoke]AppCall failed:" + err.Error()))
		}

		result = tmpRes

	case WASMVM_CONTRACT:
		conParam := states.WasmContractParam{Address: contractAddress, Args: inputs}
		param := common.SerializeToBytes(&conParam)

		newservice, err := Service.ContextRef.NewExecuteEngine(param, types.InvokeWasm)
		if err != nil {
			return jitErr(err)
		}

		tmpRes, err := newservice.Invoke()
		C.wasmjit_set_gas(vmctx, C.ulong(*Service.GasLimit))
		C.wasmjit_set_exec_step(vmctx, C.ulong(*Service.ExecStep))
		if err != nil {
			return jitErr(err)
		}

		result = tmpRes.([]byte)

	case NEOVM_CONTRACT:
		evalstack, err := util.GenerateNeoVMParamEvalStack(inputs)
		if err != nil {
			return jitErr(err)
		}

		neoservice, err := Service.ContextRef.NewExecuteEngine([]byte{}, types.InvokeNeo)
		if err != nil {
			return jitErr(err)
		}

		err = util.SetNeoServiceParamAndEngine(contractAddress, neoservice, evalstack)
		if err != nil {
			return jitErr(err)
		}

		tmp, err := neoservice.Invoke()
		C.wasmjit_set_gas(vmctx, C.ulong(*Service.GasLimit))
		C.wasmjit_set_exec_step(vmctx, C.ulong(*Service.ExecStep))
		if err != nil {
			return jitErr(err)
		}

		if tmp != nil {
			val := tmp.(*neotypes.VmValue)
			source := common.NewZeroCopySink([]byte{byte(crossvm_codec.VERSION)})

			err = neotypes.BuildResultFromNeo(*val, source)
			if err != nil {
				return jitErr(err)
			}
			result = source.Bytes()
		}

	default:
		return jitErr(errors.NewErr("Not a supported contract type"))
	}

	setCallOutPut(vmctx, result)
	return C.wasmjit_result_t{kind: C.uint(wasmjit_result_success)}
}

func tuneGas(gas uint64, mod uint64) uint64 {
	return mod*(gas/mod) + mod
}

func destroy_wasmjit_ret(ret C.wasmjit_ret) {
	buffer := ret.buffer
	msg := ret.res.msg
	if buffer.data != (*C.uchar)((unsafe.Pointer)(nil)) {
		C.wasmjit_bytes_destroy(buffer)
	}

	if msg.data != (*C.uchar)((unsafe.Pointer)(nil)) {
		C.wasmjit_bytes_destroy(msg)
	}
}

// call to c
func invokeJit(this *WasmVmService, contract *states.WasmContractParam, wasmCode []byte) ([]byte, error) {
	txHash := this.Tx.Hash()
	witnessAddrBuff, witness_len := GetAddressBuff(this.Tx.GetSignatureAddresses())
	callersAddrBuff, callers_len := GetAddressBuff(this.ContextRef.GetCallerAddress())

	var witnessPtr, callersPtr, inputPtr *C.uchar

	if witness_len == 0 {
		witnessPtr = (*C.uchar)((unsafe.Pointer)(nil))
	} else {
		witnessPtr = (*C.uchar)((unsafe.Pointer)(&witnessAddrBuff[0]))
	}

	if callers_len == 0 {
		callersPtr = (*C.uchar)((unsafe.Pointer)(nil))
	} else {
		callersPtr = (*C.uchar)((unsafe.Pointer)(&callersAddrBuff[0]))
	}

	input_len := len(contract.Args)
	if len(contract.Args) == 0 {
		inputPtr = (*C.uchar)((unsafe.Pointer)(nil))
	} else {
		inputPtr = (*C.uchar)((unsafe.Pointer)(&contract.Args[0]))
	}

	// note here uint64 should condsider as ulonglong.
	height := C.uint(this.Height)
	block_hash := (*C.h256_t)((unsafe.Pointer)(&this.BlockHash[0]))
	timestamp := C.ulong(this.Time)
	tx_hash := (*C.h256_t)((unsafe.Pointer)(&(txHash[0])))
	caller_raw := C.wasmjit_slice_t{data: callersPtr, len: C.uint(callers_len)}
	witness_raw := C.wasmjit_slice_t{data: witnessPtr, len: C.uint(witness_len)}
	input_raw := C.wasmjit_slice_t{data: inputPtr, len: C.uint(input_len)}
	service_index := C.ulong(this.ServiceIndex)
	exec_step := C.ulong(*this.ExecStep)
	gas_factor := C.ulong(this.GasFactor)
	gas_left := C.ulong(*this.GasLimit)
	codeSlice := C.wasmjit_slice_t{data: (*C.uchar)((unsafe.Pointer)(&wasmCode[0])), len: C.uint(len(wasmCode))}

	ctx := C.wasmjit_chain_context_create(height, block_hash, timestamp, tx_hash, caller_raw, witness_raw, input_raw, exec_step, gas_factor, gas_left, service_index)
	jit_ret := C.wasmjit_invoke(codeSlice, ctx)
	*this.ExecStep = uint64(jit_ret.exec_step)
	*this.GasLimit = tuneGas(uint64(jit_ret.gas_left), wasmjit_gas_mod)

	if jit_ret.res.kind != C.uint(wasmjit_result_success) {
		err := errors.NewErr(C.GoStringN((*C.char)((unsafe.Pointer)(jit_ret.res.msg.data)), C.int(jit_ret.res.msg.len)))
		destroy_wasmjit_ret(jit_ret)
		return nil, err
	}

	output := C.GoBytes((unsafe.Pointer)(jit_ret.buffer.data), (C.int)(jit_ret.buffer.len))
	destroy_wasmjit_ret(jit_ret)
	return output, nil
}
