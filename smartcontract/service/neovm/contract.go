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
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	vm "github.com/ontio/ontology/vm/neovm"
)

// ContractCreate create a new smart contract on blockchain, and put it to vm stack
func ContractCreate(service *NeoVmService, engine *vm.ExecutionEngine) error {
	contract, err := isContractParamValid(engine)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[ContractCreate] contract parameters invalid!")
	}
	contractAddress := contract.Address()
	dep, err := service.CacheDB.GetContract(contractAddress)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[ContractCreate] GetOrAdd error!")
	}
	if dep == nil {
		service.CacheDB.PutContract(contract)
		dep = contract
	}
	vm.PushData(engine, dep)
	return nil
}

// InitMetaData init contract meta data, we don't help user to check witness of owner
func InitMetaData(service *NeoVmService, engine *vm.ExecutionEngine) error {
	contractAddress := service.ContextRef.CurrentContext().ContractAddress
	oldMeta, err := service.CacheDB.GetMetaData(contractAddress)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[InitMetaData] read meta data failed!")
	}
	if oldMeta != nil { // init contract meta data
		return errors.NewDetailErr(err, errors.ErrNoCode, "[InitMetaData] meta data has already initialized")
	}
	newMeta, err := getMetaData(engine)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, fmt.Sprintf("[InitMetaData] invalid param: %s", err))
	}
	if !checkInitMeta(service, newMeta) {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[InitMetaData] meta data should contain owner")
	}
	newMeta.OntVersion = common.VERSION_SUPPORT_SHARD
	newMeta.Contract = service.ContextRef.CurrentContext().ContractAddress
	service.CacheDB.PutMetaData(newMeta)
	vm.PushData(engine, oldMeta)
	return nil
}

// ContractMigrate migrate old smart contract to a new contract, and destroy old contract
func ContractMigrate(service *NeoVmService, engine *vm.ExecutionEngine) error {
	contract, err := isContractParamValid(engine)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[ContractMigrate] contract parameters invalid!")
	}
	newAddr := contract.Address()

	if err := isContractExist(service, newAddr); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[ContractMigrate] contract invalid!")
	}
	context := service.ContextRef.CurrentContext()
	oldAddr := context.ContractAddress

	service.CacheDB.PutContract(contract)
	service.CacheDB.DeleteContract(oldAddr)

	meta, err := service.CacheDB.GetMetaData(oldAddr)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[ContractMigrate] cannot get contract meta data!")
	}
	if meta != nil {
		meta.Contract = newAddr
		service.CacheDB.PutMetaData(meta)
		service.CacheDB.DeleteMetaData(oldAddr)
	}

	iter := service.CacheDB.NewIterator(oldAddr[:])
	for has := iter.First(); has; has = iter.Next() {
		key := iter.Key()
		val := iter.Value()

		newKey := genStorageKey(newAddr, key[20:])
		service.CacheDB.Put(newKey, val)
		service.CacheDB.Delete(key)
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return err
	}

	vm.PushData(engine, contract)
	return nil
}

// ContractDestory destroy a contract
func ContractDestory(service *NeoVmService, engine *vm.ExecutionEngine) error {
	context := service.ContextRef.CurrentContext()
	if context == nil {
		return errors.NewErr("[ContractDestory] current contract context invalid!")
	}
	addr := context.ContractAddress
	contract, err := service.CacheDB.GetContract(addr)
	if err != nil || contract == nil {
		return errors.NewErr("[ContractDestory] get current contract fail!")
	}

	service.CacheDB.DeleteContract(addr)

	iter := service.CacheDB.NewIterator(addr[:])
	for has := iter.First(); has; has = iter.Next() {
		key := iter.Key()
		service.CacheDB.Delete(key)
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return err
	}

	return nil
}

// ContractGetStorageContext put contract storage context to vm stack
func ContractGetStorageContext(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[GetStorageContext] Too few input parameter!")
	}
	opInterface, err := vm.PopInteropInterface(engine)
	if err != nil {
		return err
	}
	if opInterface == nil {
		return errors.NewErr("[GetStorageContext] Pop data nil!")
	}
	contractState, ok := opInterface.(*payload.DeployCode)
	if !ok {
		return errors.NewErr("[GetStorageContext] Pop data not contract!")
	}
	address := contractState.Address()
	item, err := service.CacheDB.GetContract(address)
	if err != nil || item == nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[GetStorageContext] Get StorageContext nil")
	}
	if address != service.ContextRef.CurrentContext().ContractAddress {
		return errors.NewErr("[GetStorageContext] CodeHash not equal!")
	}
	vm.PushData(engine, NewStorageContext(address))
	return nil
}

// ContractGetCode put contract to vm stack
func ContractGetCode(service *NeoVmService, engine *vm.ExecutionEngine) error {
	i, err := vm.PopInteropInterface(engine)
	if err != nil {
		return err
	}
	vm.PushData(engine, i.(*payload.DeployCode).Code)
	return nil
}

func isContractParamValid(engine *vm.ExecutionEngine) (*payload.DeployCode, error) {
	if vm.EvaluationStackCount(engine) < 7 {
		return nil, errors.NewErr("[Contract] Too few input parameters")
	}
	code, err := vm.PopByteArray(engine)
	if err != nil {
		return nil, err
	}
	if len(code) > 1024*1024 {
		return nil, errors.NewErr("[Contract] Code too long!")
	}
	needStorage, err := vm.PopBoolean(engine)
	if err != nil {
		return nil, err
	}
	name, err := vm.PopByteArray(engine)
	if err != nil {
		return nil, err
	}
	if len(name) > 252 {
		return nil, errors.NewErr("[Contract] Name too long!")
	}
	version, err := vm.PopByteArray(engine)
	if err != nil {
		return nil, err
	}
	if len(version) > 252 {
		return nil, errors.NewErr("[Contract] Version too long!")
	}
	author, err := vm.PopByteArray(engine)
	if err != nil {
		return nil, err
	}
	if len(author) > 252 {
		return nil, errors.NewErr("[Contract] Author too long!")
	}
	email, err := vm.PopByteArray(engine)
	if err != nil {
		return nil, err
	}
	if len(email) > 252 {
		return nil, errors.NewErr("[Contract] Email too long!")
	}
	desc, err := vm.PopByteArray(engine)
	if err != nil {
		return nil, err
	}
	if len(desc) > 65536 {
		return nil, errors.NewErr("[Contract] Desc too long!")
	}
	contract := &payload.DeployCode{
		Code:        code,
		NeedStorage: needStorage,
		Name:        string(name),
		Version:     string(version),
		Author:      string(author),
		Email:       string(email),
		Description: string(desc),
	}
	return contract, nil
}

// param is owner, allShard, isFrozen, ShardId
func getMetaData(engine *vm.ExecutionEngine) (*payload.MetaDataCode, error) {
	if vm.EvaluationStackCount(engine) < 4 {
		return nil, fmt.Errorf("too few input parameters")
	}
	meta := payload.NewDefaultMetaData()
	owner, err := vm.PopByteArray(engine)
	if err != nil {
		return nil, fmt.Errorf("read owner failed, err: %s", err)
	}
	meta.Owner, err = common.AddressParseFromBytes(owner)
	if err != nil {
		return nil, fmt.Errorf("parse owner failed, err: %s", err)
	}
	meta.AllShard, err = vm.PopBoolean(engine)
	if err != nil {
		return nil, fmt.Errorf("read allShard failed, err: %s", err)
	}
	meta.IsFrozen, err = vm.PopBoolean(engine)
	if err != nil {
		return nil, fmt.Errorf("read isFrozen failed, err: %s", err)
	}
	shardId, err := vm.PopBigInt(engine)
	if err != nil {
		return nil, fmt.Errorf("read shardId failed, err: %s", err)
	}
	meta.ShardId = shardId.Uint64()
	return meta, nil
}

func checkInitMeta(service *NeoVmService, meta *payload.MetaDataCode) bool {
	if meta.Owner == common.ADDRESS_EMPTY {
		return false
	}
	if _, err := types.NewShardID(meta.ShardId); err != nil {
		return false
	}
	// shard contract can only run at self shard while init meta
	if !service.ShardID.IsRootShard() {
		return service.ShardID.ToUint64() == meta.ShardId
	}
	return true
}

func isContractExist(service *NeoVmService, contractAddress common.Address) error {
	item, err := service.CacheDB.GetContract(contractAddress)

	if err != nil || item != nil {
		return fmt.Errorf("[Contract] Get contract %x error or contract exist!", contractAddress)
	}
	return nil
}
