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
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/vm/wasmvm/exec"
	"github.com/ontio/ontology/vm/wasmvm/util"
)

func (this *WasmVmService) blockChainGetHeight(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	vm.RestoreCtx()
	vm.PushResult(uint64(this.Store.GetCurrentBlockHeight()))
	return true, nil
}

func (this *WasmVmService) blockChainGetHeaderByHeight(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[blockChainGetHeaderByHeight]parameter count error ")
	}
	hash := this.Store.GetBlockHash(uint32(params[0]))
	header, err := this.Store.GetHeaderByHash(hash)
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[blockChainGetHeaderByHeight] GetHeader error!.")
	}

	idx, err := vm.SetPointerMemory(header.ToArray())
	if err != nil {
		return false, err
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(idx))
	return true, nil
}

func (this *WasmVmService) blockChainGetHeaderByHash(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[blockChainGetHeaderByHash]parameter count error ")
	}

	hashbytes, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}

	hash, err := common.Uint256ParseFromBytes(hashbytes)
	if err != nil {
		return false, err
	}
	header, err := this.Store.GetHeaderByHash(hash)
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[blockChainGetHeaderByHeight] GetHeader error!.")
	}

	idx, err := vm.SetPointerMemory(header.ToArray())
	if err != nil {
		return false, err
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(idx))
	return true, nil
}

func (this *WasmVmService) blockChainGetBlockByHeight(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[blockChainGetBlockByHeight]parameter count error ")
	}
	block, err := this.Store.GetBlockByHeight(uint32(params[0]))
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[blockChainGetBlockByHeight] GetHeader error!.")
	}

	idx, err := vm.SetPointerMemory(block.ToArray())
	if err != nil {
		return false, err
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(idx))
	return true, nil
}

func (this *WasmVmService) blockChainGetBlockByHash(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[blockChainGetBlockByHash]parameter count error ")
	}

	hashbytes, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}

	hash, err := common.Uint256ParseFromBytes(hashbytes)
	if err != nil {
		return false, err
	}
	block, err := this.Store.GetBlockByHash(hash)
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[blockChainGetBlockByHash] GetHeader error!.")
	}

	idx, err := vm.SetPointerMemory(block.ToArray())
	if err != nil {
		return false, err
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(idx))
	return true, nil
}

func (this *WasmVmService) blockChainGetContract(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[blockChainGetContract]parameter count error ")
	}

	addressBytes, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}
	address, err := common.AddressFromBase58(util.TrimBuffToString(addressBytes))
	if err != nil {
		return false, err
	}

	item, err := this.Store.GetContractState(address)
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[blockChainGetContract] GetAsset error!")
	}

	idx, err := vm.SetPointerMemory(item.ToArray())
	if err != nil {
		return false, err
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(idx))
	return true, nil
}
