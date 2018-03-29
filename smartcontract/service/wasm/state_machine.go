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

package wasm

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/Ontology/common"
	"github.com/Ontology/core/states"
	"github.com/Ontology/core/store"
	scommon "github.com/Ontology/core/store/common"
	"github.com/Ontology/smartcontract/storage"
	vmtypes "github.com/Ontology/smartcontract/types"
	"github.com/Ontology/vm/wasmvm/exec"
	"github.com/Ontology/vm/wasmvm/util"
	"github.com/Ontology/vm/wasmvm/wasm"
)

type WasmStateMachine struct {
	*WasmStateReader
	ldgerStore store.LedgerStore
	CloneCache *storage.CloneCache
	trigger    vmtypes.TriggerType
	time       uint32
}

func NewWasmStateMachine(ldgerStore store.LedgerStore, dbCache scommon.StateStore, trigger vmtypes.TriggerType, time uint32) *WasmStateMachine {

	var stateMachine WasmStateMachine
	stateMachine.ldgerStore = ldgerStore
	stateMachine.CloneCache = storage.NewCloneCache(dbCache)
	stateMachine.WasmStateReader = NewWasmStateReader(ldgerStore, trigger)
	stateMachine.trigger = trigger
	stateMachine.time = time

	stateMachine.Register("PutStorage", stateMachine.putstore)
	stateMachine.Register("GetStorage", stateMachine.getstore)
	stateMachine.Register("DeleteStorage", stateMachine.deletestore)
	stateMachine.Register("callContract", callContract)

	return &stateMachine
}

//======================store apis here============================================
func (s *WasmStateMachine) putstore(engine *exec.ExecutionEngine) (bool, error) {

	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()

	if len(params) != 2 {
		return false, errors.New("[putstore] parameter count error")
	}

	key, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}

	if len(key) > 1024 {
		return false, errors.New("[putstore] Get Storage key to long")
	}

	value, err := vm.GetPointerMemory(params[1])
	if err != nil {
		return false, err
	}

	k, err := serializeStorageKey(vm.CodeHash, key)
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
		return false, errors.New("[getstore] parameter count error ")
	}

	key, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}

	k, err := serializeStorageKey(vm.CodeHash, key)
	if err != nil {
		return false, err
	}
	item, err := s.CloneCache.Get(scommon.ST_STORAGE, k)
	if err != nil {
		return false, err
	}

	// idx = int64.max value if item is nil
	//todo need more  test about the nil case
	idx, err := vm.SetPointerMemory(item)
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
		return false, errors.New("[deletestore] parameter count error")
	}

	key, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}

	k, err := serializeStorageKey(vm.CodeHash, key)
	if err != nil {
		return false, err
	}

	s.CloneCache.Delete(scommon.ST_STORAGE, k)
	vm.RestoreCtx()

	return true, nil
}

//call other contract
func callContract(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 3 {
		return false, errors.New("parameter count error while call readMessage")
	}
	contractAddressIdx := params[0]
	addr, err := vm.GetPointerMemory(contractAddressIdx)
	if err != nil {
		return false, errors.New("get Contract address failed")
	}
	//the contract codes
	contractBytes, err := getContractFromAddr(addr)
	if err != nil {
		return false, err
	}
	bf := bytes.NewBuffer(contractBytes)
	module, err := wasm.ReadModule(bf, emptyImporter)
	if err != nil {
		return false, errors.New("load Module failed")
	}

	methodName, err := vm.GetPointerMemory(params[1])
	if err != nil {
		return false, errors.New("[callContract]get Contract methodName failed")
	}

	arg, err := vm.GetPointerMemory(params[2])
	if err != nil {
		return false, errors.New("[callContract]get Contract arg failed")
	}

	res, err := vm.CallProductContract(module, methodName, arg)

	vm.RestoreCtx()
	if envCall.GetReturns() {
		vm.PushResult(uint64(res))
	}
	return true, nil
}

func serializeStorageKey(codeHash common.Address, key []byte) ([]byte, error) {
	bf := new(bytes.Buffer)
	storageKey := &states.StorageKey{CodeHash: codeHash, Key: key}
	if _, err := storageKey.Serialize(bf); err != nil {
		return []byte{}, errors.New("[serializeStorageKey] StorageKey serialize error!")
	}
	return bf.Bytes(), nil
}

func getContractFromAddr(addr []byte) ([]byte, error) {

	//just for test
	contract := util.TrimBuffToString(addr)
	code, err := ioutil.ReadFile(fmt.Sprintf("./testdata2/%s.wasm", contract))
	if err != nil {
		fmt.Printf("./testdata2/%s.wasm is not exist", contract)
		return nil, err
	}

	return code, nil
	//Fixme get the contract code from ledger
	/*
		codeHash, err := common.Uint160ParseFromBytes(addr)
		if err != nil {
			return nil, errors.New("get address Code hash failed")
		}

		contract, err := ledger.DefLedger.GetContractState(codeHash)
		if err != nil {
			return nil, errors.New("get contract state failed")
		}

		if contract.VmType != types.WASMVM {
			return nil, errors.New(" contract is not a wasm contract")
		}

		return contract.Code, nil
	*/

}

func emptyImporter(name string) (*wasm.Module, error) {
	return nil, nil
}
