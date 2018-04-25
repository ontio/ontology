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

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/store"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/storage"
	"github.com/ontio/ontology/vm/wasmvm/exec"
	"github.com/ontio/ontology/vm/wasmvm/memory"
	"github.com/ontio/ontology/vm/wasmvm/util"
)

type LogLevel byte

const (
	Debug LogLevel = iota
	Info
	Error
)

type ParamType byte

const (
	Json ParamType = iota
	Raw
)

type WasmStateMachine struct {
	*WasmStateReader
	ldgerStore store.LedgerStore
	CloneCache *storage.CloneCache
	time       uint32
}

func NewWasmStateMachine(ldgerStore store.LedgerStore, cloneCache *storage.CloneCache, time uint32) *WasmStateMachine {

	var stateMachine WasmStateMachine
	stateMachine.ldgerStore = ldgerStore
	stateMachine.CloneCache = cloneCache
	stateMachine.WasmStateReader = NewWasmStateReader(ldgerStore)
	stateMachine.time = time

	//todo remove the following later
	stateMachine.Register("PutStorage", stateMachine.putstore)
	stateMachine.Register("GetStorage", stateMachine.getstore)
	stateMachine.Register("DeleteStorage", stateMachine.deletestore)
	stateMachine.Register("ContractLogDebug", stateMachine.contractLogDebug)
	stateMachine.Register("ContractLogInfo", stateMachine.contractLogInfo)
	stateMachine.Register("ContractLogError", stateMachine.contractLogError)

	return &stateMachine
}

//======================store apis here============================================
func (s *WasmStateMachine) putstore(engine *exec.ExecutionEngine) (bool, error) {

	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 2 {
		return false, errors.NewErr("[putstore] parameter count error")
	}

	key, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}
	if len(key) > 1024 {
		return false, errors.NewErr("[putstore] Get Storage key to long")
	}

	value, err := vm.GetPointerMemory(params[1])
	if err != nil {
		return false, err
	}
	k, err := serializeStorageKey(vm.ContractAddress, []byte(util.TrimBuffToString(key)))
	if err != nil {
		return false, err
	}
	s.CloneCache.Add(scommon.ST_STORAGE, k, &states.StorageItem{Value: value})

	vm.RestoreCtx()

	return true, nil
}

func (s *WasmStateMachine) getstore(engine *exec.ExecutionEngine) (bool, error) {

	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()

	if len(params) != 1 {
		return false, errors.NewErr("[getstore] parameter count error ")
	}

	key, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}
	k, err := serializeStorageKey(vm.ContractAddress, []byte(util.TrimBuffToString(key)))
	if err != nil {
		return false, err
	}
	item, err := s.CloneCache.Get(scommon.ST_STORAGE, k)
	if err != nil {
		return false, err
	}

	if item == nil {
		vm.RestoreCtx()
		if envCall.GetReturns() {
			vm.PushResult(uint64(memory.VM_NIL_POINTER))
		}
		return true, nil
	}
	idx, err := vm.SetPointerMemory(item.(*states.StorageItem).Value)
	if err != nil {
		return false, err
	}

	vm.RestoreCtx()
	if envCall.GetReturns() {
		vm.PushResult(uint64(idx))
	}
	return true, nil
}

func (s *WasmStateMachine) deletestore(engine *exec.ExecutionEngine) (bool, error) {

	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()

	if len(params) != 1 {
		return false, errors.NewErr("[deletestore] parameter count error")
	}

	key, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}

	k, err := serializeStorageKey(vm.ContractAddress, []byte(util.TrimBuffToString(key)))
	if err != nil {
		return false, err
	}

	s.CloneCache.Delete(scommon.ST_STORAGE, k)
	vm.RestoreCtx()

	return true, nil
}

func (s *WasmStateMachine) contractLogDebug(engine *exec.ExecutionEngine) (bool, error) {
	_, err := contractLog(Debug, engine)
	if err != nil {
		return false, err
	}

	engine.GetVM().RestoreCtx()
	return true, nil
}

func (s *WasmStateMachine) contractLogInfo(engine *exec.ExecutionEngine) (bool, error) {
	_, err := contractLog(Info, engine)
	if err != nil {
		return false, err

	}
	engine.GetVM().RestoreCtx()
	return true, nil
}

func (s *WasmStateMachine) contractLogError(engine *exec.ExecutionEngine) (bool, error) {
	_, err := contractLog(Error, engine)
	if err != nil {
		return false, err
	}

	engine.GetVM().RestoreCtx()
	return true, nil
}

func contractLog(lv LogLevel, engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("parameter count error while call contractLong")
	}

	Idx := params[0]
	addr, err := vm.GetPointerMemory(Idx)
	if err != nil {
		return false, errors.NewErr("get Contract address failed")
	}

	msg := fmt.Sprintf("[WASM Contract] Address:%s message:%s", vm.ContractAddress.ToHexString(), util.TrimBuffToString(addr))

	switch lv {
	case Debug:
		log.Debug(msg)
	case Info:
		log.Info(msg)
	case Error:
		log.Error(msg)
	}
	return true, nil

}

func serializeStorageKey(contractAddress common.Address, key []byte) ([]byte, error) {
	bf := new(bytes.Buffer)
	storageKey := &states.StorageKey{CodeHash: contractAddress, Key: key}
	if _, err := storageKey.Serialize(bf); err != nil {
		return []byte{}, errors.NewErr("[serializeStorageKey] StorageKey serialize error!")
	}
	return bf.Bytes(), nil
}
