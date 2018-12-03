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
	"encoding/binary"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/vm/wasmvm/exec"
	"github.com/ontio/ontology/vm/wasmvm/util"
)

func (this *WasmVmService) runtimeGetTime(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	vm.RestoreCtx()
	vm.PushResult(uint64(this.Time))
	return true, nil
}
func (this *WasmVmService) runtimeLog(engine *exec.ExecutionEngine) (bool, error) {

	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[RuntimeLog]parameter count error ")
	}

	item, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}

	context := this.ContextRef.CurrentContext()
	txHash := this.Tx.Hash()
	event.PushSmartCodeEvent(txHash, 0, event.EVENT_LOG, &event.LogEventArgs{TxHash: txHash, ContractAddress: context.ContractAddress, Message: string(item)})
	vm.RestoreCtx()

	return true, nil
}

func (this *WasmVmService) runtimeCheckSig(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 3 {
		return false, errors.NewErr("[RuntimeCheckSig]parameter count error ")
	}
	pubKey, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}
	data, err := vm.GetPointerMemory(params[1])
	if err != nil {
		return false, err
	}
	sig, err := vm.GetPointerMemory(params[2])
	if err != nil {
		return false, err
	}
	res := 0
	err = signature.Verify(pubKey, data, sig)
	if err == nil {
		res = 1
	}

	vm.RestoreCtx()
	vm.PushResult(uint64(res))

	return true, nil
}

func (this *WasmVmService) runtimeNotify(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()

	if len(params) != 1 {

		return false, errors.NewErr("[RuntimeNotify]parameter count error ")
	}
	item, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}

	length := len(item) / 4

	notify := make([]string, length)

	for i := 0; i < length; i++ {

		tmp := item[i*4 : (i+1)*4]
		idx := binary.LittleEndian.Uint32(tmp)

		tmpitem, err := vm.GetPointerMemory(uint64(idx))
		if err != nil {
			return false, err
		}
		notify[i] = string(tmpitem)
	}

	context := this.ContextRef.CurrentContext()

	this.Notifications = append(this.Notifications, &event.NotifyEventInfo{ContractAddress: context.ContractAddress, States: notify})
	vm.RestoreCtx()
	return true, nil
}

func (this *WasmVmService) runtimeCheckWitness(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()

	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[CheckWitness]get parameter count error!")
	}
	data, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, errors.NewErr("[CheckWitness]" + err.Error())
	}
	address, err := common.AddressFromBase58(util.TrimBuffToString(data))
	if err != nil {
		return false, errors.NewErr("[CheckWitness]" + err.Error())
	}
	chkRes := this.ContextRef.CheckWitness(address)
	res := 0
	if chkRes == true {
		res = 1
	}
	vm.RestoreCtx()
	if vm.GetEnvCall().GetReturns() {
		vm.PushResult(uint64(res))
	}
	return true, nil
}

// raise an exception to terminate the vm execution
func (this *WasmVmService) runtimeRaiseException(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()

	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[runtimeRaiseException]get parameter count error!")
	}
	data, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, errors.NewErr("[runtimeRaiseException]" + err.Error())
	}

	return false, errors.NewErr(string(data))
}

//get current block hash
func (this *WasmVmService) runtimeGetCurrentBlockHash(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()

	hash := this.BlockHash.ToHexString()
	idx, err := vm.SetPointerMemory(hash)
	if err != nil {
		return false, err
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(idx))

	return true, nil
}

//get current tx
func (this *WasmVmService) runtimeGetCodeContainer(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()

	tx := this.Tx.Raw
	idx, err := vm.SetPointerMemory(tx)
	if err != nil {
		return false, err
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(idx))

	return true, nil
}

//get current contract address
func (this *WasmVmService) runtimeGetExecutingAddress(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()

	ctx := this.ContextRef.CurrentContext()
	if ctx == nil {
		return false, errors.NewErr("Calling context invalid")
	}

	addr := ctx.ContractAddress[:]
	//addr := vm.ContractAddress[:]
	idx, err := vm.SetPointerMemory(addr)
	if err != nil {
		return false, err
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(idx))

	return true, nil
}

//get current contract address
func (this *WasmVmService) runtimeGetCallingAddress(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()

	ctx := this.ContextRef.CallingContext()
	if ctx == nil {
		return false, errors.NewErr("Calling context invalid")
	}
	addr := ctx.ContractAddress[:]
	//addr := vm.ContractAddress[:]
	idx, err := vm.SetPointerMemory(addr)
	if err != nil {
		return false, err
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(idx))

	return true, nil
}

//get entry contract address
func (this *WasmVmService) runtimeGetEntryAddress(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()

	ctx := this.ContextRef.EntryContext()
	if ctx == nil {
		return false, errors.NewErr("Calling context invalid")
	}

	addr := ctx.ContractAddress[:]
	//addr := vm.ContractAddress[:]
	idx, err := vm.SetPointerMemory(addr)
	if err != nil {
		return false, err
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(idx))

	return true, nil
}

//change address to Base58 format
func (this *WasmVmService) runtimeAddressToBase58(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()

	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[runtimeAddressToBase58]get parameter count error!")
	}
	data, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, errors.NewErr("[runtimeAddressToBase58]" + err.Error())
	}
	address, err := common.AddressParseFromBytes(data)
	if err != nil {
		return false, errors.NewErr("[runtimeAddressToBase58]" + err.Error())
	}

	idx, err := vm.SetPointerMemory(address.ToBase58())
	if err != nil {
		return false, err
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(idx))
	return true, nil
}

//change address to hex format
func (this *WasmVmService) runtimeAddressToHex(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()

	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[runtimeAddressToHex]get parameter count error!")
	}
	data, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, errors.NewErr("[runtimeAddressToHex]" + err.Error())
	}
	address, err := common.AddressParseFromBytes(data)
	if err != nil {
		return false, errors.NewErr("[runtimeAddressToHex]" + err.Error())
	}

	idx, err := vm.SetPointerMemory(address.ToHexString())
	if err != nil {
		return false, err
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(idx))
	return true, nil
}
