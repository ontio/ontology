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

func (this *WasmVmService) headerGetHash(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[transactionGetHash] parameter count error")
	}

	header, err := this.Store.GetHeaderByHeight(uint32(params[0]))
	if err != nil {
		return false, err
	}
	hash := header.Hash()
	idx, err := vm.SetPointerMemory(common.ToHexString(hash.ToArray()))
	if err != nil {
		return false, err
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(idx))
	return true, nil
}

func (this *WasmVmService) headerGetVersionByHeight(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[transactionGetHash] parameter count error")
	}
	header, err := this.Store.GetHeaderByHeight(uint32(params[0]))
	if err != nil {
		return false, err
	}
	version := header.Version
	vm.RestoreCtx()
	vm.PushResult(uint64(version))
	return true, nil
}

func (this *WasmVmService) headerGetVersionByHash(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[transactionGetHash] parameter count error")
	}

	hexbytes, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}

	hashbytes, err := common.HexToBytes(util.TrimBuffToString(hexbytes))
	if err != nil {
		return false, err
	}

	hash, err := common.Uint256ParseFromBytes(hashbytes)
	if err != nil {
		return false, err
	}

	header, err := this.Store.GetHeaderByHash(hash)
	if err != nil {
		return false, err
	}
	version := header.Version

	vm.RestoreCtx()
	vm.PushResult(uint64(version))
	return true, nil
}

func (this *WasmVmService) headerGetPrevHashByHash(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[transactionGetHash] parameter count error")
	}

	hexbytes, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}

	hashbytes, err := common.HexToBytes(util.TrimBuffToString(hexbytes))
	if err != nil {
		return false, err
	}

	hash, err := common.Uint256ParseFromBytes(hashbytes)
	if err != nil {
		return false, err
	}

	header, err := this.Store.GetHeaderByHash(hash)
	if err != nil {
		return false, err
	}

	prevhash := header.PrevBlockHash.ToArray()
	idx, err := vm.SetPointerMemory(common.ToHexString(prevhash))
	if err != nil {
		return false, err
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(idx))
	return true, nil
}

func (this *WasmVmService) headerGetPrevHashByHeight(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[transactionGetHash] parameter count error")
	}

	header, err := this.Store.GetHeaderByHeight(uint32(params[0]))
	if err != nil {
		return false, err
	}

	hash := header.PrevBlockHash.ToArray()
	idx, err := vm.SetPointerMemory(common.ToHexString(hash))
	if err != nil {
		return false, err
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(idx))
	return true, nil
}

func (this *WasmVmService) headerGetMerkleRootByHeight(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[transactionGetHash] parameter count error")
	}

	header, err := this.Store.GetHeaderByHeight(uint32(params[0]))
	if err != nil {
		return false, err
	}

	hash := header.TransactionsRoot.ToArray()
	idx, err := vm.SetPointerMemory(common.ToHexString(hash))
	if err != nil {
		return false, err
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(idx))
	return true, nil
}

func (this *WasmVmService) headerGetMerkleRootByHash(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[transactionGetHash] parameter count error")
	}

	hexbytes, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}

	hashbytes, err := common.HexToBytes(util.TrimBuffToString(hexbytes))
	if err != nil {
		return false, err
	}

	hash, err := common.Uint256ParseFromBytes(hashbytes)
	if err != nil {
		return false, err
	}

	header, err := this.Store.GetHeaderByHash(hash)
	if err != nil {
		return false, err
	}

	merkel := header.TransactionsRoot.ToArray()
	idx, err := vm.SetPointerMemory(common.ToHexString(merkel))
	if err != nil {
		return false, err
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(idx))
	return true, nil
}

func (this *WasmVmService) headerGetIndexByHash(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[transactionGetHash] parameter count error")
	}

	hexbytes, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}

	hashbytes, err := common.HexToBytes(util.TrimBuffToString(hexbytes))
	if err != nil {
		return false, err
	}

	hash, err := common.Uint256ParseFromBytes(hashbytes)
	if err != nil {
		return false, err
	}

	header, err := this.Store.GetHeaderByHash(hash)
	if err != nil {
		return false, err
	}

	height := header.Height

	vm.RestoreCtx()
	vm.PushResult(uint64(height))
	return true, nil
}

func (this *WasmVmService) headerGetTimestampByHeight(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[transactionGetHash] parameter count error")
	}

	header, err := this.Store.GetHeaderByHeight(uint32(params[0]))
	if err != nil {
		return false, err
	}

	tm := header.Timestamp

	vm.RestoreCtx()
	vm.PushResult(uint64(tm))
	return true, nil
}

func (this *WasmVmService) headerGetTimestampByHash(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[transactionGetHash] parameter count error")
	}

	hexbytes, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}

	hashbytes, err := common.HexToBytes(util.TrimBuffToString(hexbytes))
	if err != nil {
		return false, err
	}

	hash, err := common.Uint256ParseFromBytes(hashbytes)
	if err != nil {
		return false, err
	}

	header, err := this.Store.GetHeaderByHash(hash)
	if err != nil {
		return false, err
	}

	tm := header.Timestamp

	vm.RestoreCtx()
	vm.PushResult(uint64(tm))
	return true, nil
}

func (this *WasmVmService) headerGetConsensusDataByHash(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[transactionGetHash] parameter count error")
	}

	hexbytes, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}

	hashbytes, err := common.HexToBytes(util.TrimBuffToString(hexbytes))
	if err != nil {
		return false, err
	}

	hash, err := common.Uint256ParseFromBytes(hashbytes)
	if err != nil {
		return false, err
	}

	header, err := this.Store.GetHeaderByHash(hash)
	if err != nil {
		return false, err
	}
	cd := header.ConsensusData

	vm.RestoreCtx()
	vm.PushResult(cd)
	return true, nil
}

func (this *WasmVmService) headerGetConsensusDataByHeight(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[transactionGetHash] parameter count error")
	}

	header, err := this.Store.GetHeaderByHeight(uint32(params[0]))
	if err != nil {
		return false, err
	}

	cd := header.ConsensusData

	vm.RestoreCtx()
	vm.PushResult(cd)
	return true, nil
}

func (this *WasmVmService) headerGetNextConsensusByHeight(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[transactionGetHash] parameter count error")
	}

	header, err := this.Store.GetHeaderByHeight(uint32(params[0]))
	if err != nil {
		return false, err
	}

	cd := header.NextBookkeeper[:]
	idx, err := vm.SetPointerMemory(common.ToHexString(cd))
	if err != nil {
		return false, err
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(idx))
	return true, nil
}

func (this *WasmVmService) headerGetNextConsensusByHash(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[transactionGetHash] parameter count error")
	}

	hexbytes, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}

	hashbytes, err := common.HexToBytes(util.TrimBuffToString(hexbytes))
	if err != nil {
		return false, err
	}

	hash, err := common.Uint256ParseFromBytes(hashbytes)
	if err != nil {
		return false, err
	}

	header, err := this.Store.GetHeaderByHash(hash)
	if err != nil {
		return false, err
	}

	cd := header.NextBookkeeper[:]
	idx, err := vm.SetPointerMemory(common.ToHexString(cd))
	if err != nil {
		return false, err
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(idx))
	return true, nil
}
