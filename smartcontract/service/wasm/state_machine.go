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
	vmtypes "github.com/Ontology/smartcontract/types"
	"github.com/Ontology/smartcontract/storage"
	"github.com/Ontology/core/store"
	scommon "github.com/Ontology/core/store/common"
	"github.com/Ontology/core/types"
	"github.com/Ontology/core/ledger"
	"github.com/Ontology/vm/wasmvm/exec"
	"bytes"
	"github.com/Ontology/common"
	"github.com/Ontology/core/states"
	"errors"
)

type WasmStateMachine struct {
	*WasmStateReader
	ldgerStore store.LedgerStore
	CloneCache *storage.CloneCache
	trigger    vmtypes.TriggerType
	block       *types.Block
}


func NewWasmStateMachine(ldgerStore store.LedgerStore, dbCache scommon.StateStore, trigger vmtypes.TriggerType, block *types.Block) *WasmStateMachine {
	var stateMachine WasmStateMachine
	stateMachine.ldgerStore = ldgerStore
	stateMachine.CloneCache = storage.NewCloneCache(dbCache)
	stateMachine.WasmStateReader = NewWasmStateReader(ldgerStore,trigger)
	stateMachine.trigger = trigger
	stateMachine.block = block

	stateMachine.Register("getBlockHeight",bcGetHeight)
	stateMachine.Register("PutStorage",putstore)
	stateMachine.Register("GetStorage",getstore)
	stateMachine.Register("DeleteStorage",deletestore)
	//todo add and register services
	return &stateMachine
}

//======================some block api ===============
func  bcGetHeight(engine *exec.ExecutionEngine) (bool, error) {
/*	vm := engine.GetVM()
	var i uint32
	if ledger.DefaultLedger == nil {
		i = 0
	} else {
		i = ledger.DefaultLedger.PersistStore.GetHeight()
	}
	//engine.vm.ctx = envCall.envPreCtx
	vm.RestoreCtx()
	if vm.GetEnvCall().GetReturns(){
		vm.PushResult(uint64(i))
	}*/
	return true,nil
}

func putstore(engine *exec.ExecutionEngine) (bool, error) {
	return true,nil
}

func getstore(engine *exec.ExecutionEngine) (bool, error) {
	return true,nil
}


func deletestore(engine *exec.ExecutionEngine) (bool, error) {
	return true,nil
}

func serializeStorageKey(codeHash common.Address, key []byte) ([]byte, error) {
	bf := new(bytes.Buffer)
	storageKey := &states.StorageKey{CodeHash: codeHash, Key: key}
	if _, err := storageKey.Serialize(bf); err != nil {
		return []byte{}, errors.New("[serializeStorageKey] StorageKey serialize error!")
	}
	return bf.Bytes(), nil
}
