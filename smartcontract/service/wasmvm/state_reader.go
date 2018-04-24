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
	"github.com/ontio/ontology/core/store"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/vm/wasmvm/exec"
)

type WasmStateReader struct {
	serviceMap    map[string]func(*exec.ExecutionEngine) (bool, error)
	Notifications []*event.NotifyEventInfo
	ldgerStore    store.LedgerStore
}

func NewWasmStateReader(ldgerStore store.LedgerStore) *WasmStateReader {
	i := &WasmStateReader{
		ldgerStore: ldgerStore,
		serviceMap: make(map[string]func(*exec.ExecutionEngine) (bool, error)),
	}

	i.Register("GetBlockHeight", i.Getblockheight)
	i.Register("GetBlockHashByNumber", i.GetblockhashbyNumber)
	i.Register("GetTimeStamp", i.GetblockTimestamp)

	return i
}

func (i *WasmStateReader) Register(name string, handler func(*exec.ExecutionEngine) (bool, error)) bool {
	if _, ok := i.serviceMap[name]; ok {
		return false
	}
	i.serviceMap[name] = handler
	return true
}

func (i *WasmStateReader) Invoke(methodName string, engine *exec.ExecutionEngine) (bool, error) {

	if v, ok := i.serviceMap[methodName]; ok {
		return v(engine)
	}
	return true, errors.NewErr("Not supported method:" + methodName)
}

func (i *WasmStateReader) MergeMap(mMap map[string]func(*exec.ExecutionEngine) (bool, error)) bool {

	for k, v := range mMap {
		if _, ok := i.serviceMap[k]; !ok {
			i.serviceMap[k] = v
		}
	}
	return true
}

func (i *WasmStateReader) GetServiceMap() map[string]func(*exec.ExecutionEngine) (bool, error) {
	return i.serviceMap
}

func (i *WasmStateReader) Exists(name string) bool {
	_, ok := i.serviceMap[name]
	return ok
}

//============================block apis below============================/
func (i *WasmStateReader) Getblockheight(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()

	h := i.ldgerStore.GetCurrentBlockHeight()
	vm.RestoreCtx()
	if vm.GetEnvCall().GetReturns() {
		vm.PushResult(uint64(h))
	}
	return true, nil
}


func (i *WasmStateReader) GetblockTimestamp(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()

	hash := i.ldgerStore.GetCurrentBlockHash()
	header, err := i.ldgerStore.GetHeaderByHash(hash)
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[RuntimeGetTime] GetHeader error!.")
	}

	if vm.GetEnvCall().GetReturns() {
		vm.PushResult(uint64(header.Timestamp))
	}
	return true, nil
}

func (i *WasmStateReader) GetblockhashbyNumber(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()

	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false ,errors.NewErr("[GetblockhashbyNumber]get parameter count error!")
	}

	h := i.ldgerStore.GetBlockHash(uint32(params[0]))

	hashIdx,err := vm.SetPointerMemory(h.ToArray())
	if err != nil{
		return false ,errors.NewErr("[GetblockhashbyNumber]SetPointerMemory error!")
	}

	vm.RestoreCtx()
	if vm.GetEnvCall().GetReturns() {
		vm.PushResult(uint64(hashIdx))
	}
	return true, nil
}

func contains(addresses []common.Address, address common.Address) bool {
	for _, v := range addresses {
		if v == address {
			return true
		}
	}
	return false
}
