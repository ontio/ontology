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

package neovm

import (
	"bytes"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/states"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/errors"
	vm "github.com/ontio/ontology/vm/neovm"
)

// StoragePut put smart contract storage item to cache
func StoragePut(service *NeoVmService, engine *vm.ExecutionEngine) error {
	context, err := getContext(engine)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[StoragePut] get pop context error!")
	}
	if err := checkStorageContext(service, context); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[StoragePut] check context error!")
	}

	key := vm.PopByteArray(engine)
	if len(key) > 1024 {
		return errors.NewErr("[StoragePut] Storage key to long")
	}

	value := vm.PopByteArray(engine)
	service.CloneCache.Add(scommon.ST_STORAGE, getStorageKey(context.address, key), &states.StorageItem{Value: value})
	return nil
}

// StorageDelete delete smart contract storage item from cache
func StorageDelete(service *NeoVmService, engine *vm.ExecutionEngine) error {
	context, err := getContext(engine)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[StorageDelete] get pop context error!")
	}
	if err := checkStorageContext(service, context); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[StorageDelete] check context error!")
	}

	service.CloneCache.Delete(scommon.ST_STORAGE, getStorageKey(context.address, vm.PopByteArray(engine)))

	return nil
}

// StorageGet push smart contract storage item from cache to vm stack
func StorageGet(service *NeoVmService, engine *vm.ExecutionEngine) error {
	context, err := getContext(engine)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[StorageGet] get pop context error!")
	}

	item, err := service.CloneCache.Get(scommon.ST_STORAGE, getStorageKey(context.address, vm.PopByteArray(engine)))
	if err != nil {
		return err
	}

	if item == nil {
		vm.PushData(engine, []byte{})
	} else {
		vm.PushData(engine, item.(*states.StorageItem).Value)
	}
	return nil
}

// StorageGetContext push smart contract storage context to vm stack
func StorageGetContext(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, NewStorageContext(service.ContextRef.CurrentContext().ContractAddress))
	return nil
}

func checkStorageContext(service *NeoVmService, context *StorageContext) error {
	item, err := service.CloneCache.Get(scommon.ST_CONTRACT, context.address[:])
	if err != nil || item == nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[CheckStorageContext] get context fail!")
	}
	return nil
}

func getContext(engine *vm.ExecutionEngine) (*StorageContext, error) {
	if vm.EvaluationStackCount(engine) < 2 {
		return nil, errors.NewErr("[Context] Too few input parameters ")
	}
	opInterface := vm.PopInteropInterface(engine)
	if opInterface == nil {
		return nil, errors.NewErr("[Context] Get storageContext nil")
	}
	context, ok := opInterface.(*StorageContext)
	if !ok {
		return nil, errors.NewErr("[Context] Get storageContext invalid")
	}
	return context, nil
}

func getStorageKey(codeHash common.Address, key []byte) []byte {
	buf := bytes.NewBuffer(nil)
	buf.Write(codeHash[:])
	buf.Write(key)
	return buf.Bytes()
}
