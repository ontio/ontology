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
	"fmt"

	"github.com/Ontology/common"
	"github.com/Ontology/core/states"
	scommon "github.com/Ontology/core/store/common"
	"github.com/Ontology/core/store"
	"github.com/Ontology/errors"
	"github.com/Ontology/smartcontract/storage"
	stypes "github.com/Ontology/smartcontract/types"
	vm "github.com/Ontology/vm/neovm"
	"github.com/Ontology/core/payload"
	vmtypes "github.com/Ontology/vm/types"
)

type StateMachine struct {
	*StateReader
	ldgerStore store.LedgerStore
	CloneCache *storage.CloneCache
	trigger    stypes.TriggerType
	time       uint32
}

func NewStateMachine(ldgerStore store.LedgerStore, dbCache scommon.StateStore, trigger stypes.TriggerType, time uint32) *StateMachine {
	var stateMachine StateMachine
	stateMachine.ldgerStore = ldgerStore
	stateMachine.CloneCache = storage.NewCloneCache(dbCache)
	stateMachine.StateReader = NewStateReader(ldgerStore, trigger)
	stateMachine.trigger = trigger
	stateMachine.time = time

	stateMachine.StateReader.Register("Neo.Runtime.GetTrigger", stateMachine.RuntimeGetTrigger)
	stateMachine.StateReader.Register("Neo.Runtime.GetTime", stateMachine.RuntimeGetTime)

	stateMachine.StateReader.Register("Neo.Contract.Create", stateMachine.ContractCreate)
	stateMachine.StateReader.Register("Neo.Contract.Migrate", stateMachine.ContractMigrate)
	stateMachine.StateReader.Register("Neo.Contract.GetStorageContext", stateMachine.GetStorageContext)
	stateMachine.StateReader.Register("Neo.Contract.GetScript", stateMachine.ContractGetCode)
	stateMachine.StateReader.Register("Neo.Contract.Destroy", stateMachine.ContractDestory)

	stateMachine.StateReader.Register("Neo.Storage.Get", stateMachine.StorageGet)
	stateMachine.StateReader.Register("Neo.Storage.Put", stateMachine.StoragePut)
	stateMachine.StateReader.Register("Neo.Storage.Delete", stateMachine.StorageDelete)
	return &stateMachine
}

func (s *StateMachine) RuntimeGetTrigger(engine *vm.ExecutionEngine) (bool, error) {
	vm.PushData(engine, int(s.trigger))
	return true, nil
}

func (s *StateMachine) RuntimeGetTime(engine *vm.ExecutionEngine) (bool, error) {
	vm.PushData(engine, s.time)
	return true, nil
}

func (s *StateMachine) ContractCreate(engine *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(engine) < 7 {
		return false, errors.NewErr("[ContractCreate] Too few input parameters")
	}
	code := vm.PopByteArray(engine); if len(code) > 1024 * 1024 {
		return false, errors.NewErr("[ContractCreate] Code too long!")
	}
	needStorage := vm.PopBoolean(engine)
	name := vm.PopByteArray(engine); if len(name) > 252 {
		return false, errors.NewErr("[ContractCreate] Name too long!")
	}
	version := vm.PopByteArray(engine); if len(version) > 252 {
		return false, errors.NewErr("[ContractCreate] Version too long!")
	}
	author := vm.PopByteArray(engine); if len(author) > 252 {
		return false, errors.NewErr("[ContractCreate] Author too long!")
	}
	email := vm.PopByteArray(engine); if len(email) > 252 {
		return false, errors.NewErr("[ContractCreate] Email too long!")
	}
	desc := vm.PopByteArray(engine); if len(desc) > 65536 {
		return false, errors.NewErr("[ContractCreate] Desc too long!")
	}
	vmCode := &vmtypes.VmCode{VmType:vmtypes.NEOVM, Code: code}
	contractState := &payload.DeployCode{
		Code:        vmCode,
		NeedStorage: needStorage,
		Name:        string(name),
		Version:     string(version),
		Author:      string(author),
		Email:       string(email),
		Description: string(desc),
	}
	contractAddress := vmCode.AddressFromVmCode()
	state, err := s.CloneCache.GetOrAdd(scommon.ST_CONTRACT, contractAddress[:], contractState)
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[ContractCreate] GetOrAdd error!")
	}
	vm.PushData(engine, state)
	return true, nil
}

func (s *StateMachine) ContractMigrate(engine *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(engine) < 7 {
		return false, errors.NewErr("[ContractMigrate] Too few input parameters ")
	}
	code := vm.PopByteArray(engine); if len(code) > 1024 * 1024 {
		return false, errors.NewErr("[ContractMigrate] Code too long!")
	}
	vmCode := &vmtypes.VmCode{
		Code: code,
		VmType: vmtypes.NEOVM,
	}
	contractAddress := vmCode.AddressFromVmCode()
	item, err := s.CloneCache.Get(scommon.ST_CONTRACT, contractAddress[:]); if err != nil {
		return false, errors.NewErr("[ContractMigrate] Get Contract error!")
	}
	if item != nil {
		return false, errors.NewErr("[ContractMigrate] Migrate Contract has exist!")
	}

	nameByte := vm.PopByteArray(engine); if len(nameByte) > 252 {
		return false, errors.NewErr("[ContractMigrate] Name too long!")
	}
	versionByte := vm.PopByteArray(engine); if len(versionByte) > 252 {
		return false, errors.NewErr("[ContractMigrate] Version too long!")
	}
	authorByte := vm.PopByteArray(engine); if len(authorByte) > 252 {
		return false, errors.NewErr("[ContractMigrate] Author too long!")
	}
	emailByte := vm.PopByteArray(engine); if len(emailByte) > 252 {
		return false, errors.NewErr("[ContractMigrate] Email too long!")
	}
	descByte := vm.PopByteArray(engine); if len(descByte) > 65536 {
		return false, errors.NewErr("[ContractMigrate] Desc too long!")
	}
	contractState := &payload.DeployCode{
		Code:        vmCode,
		Name:        string(nameByte),
		Version:     string(versionByte),
		Author:      string(authorByte),
		Email:       string(emailByte),
		Description: string(descByte),
	}
	s.CloneCache.Add(scommon.ST_CONTRACT, contractAddress[:], contractState)
	stateValues, err := s.CloneCache.Store.Find(scommon.ST_CONTRACT, contractAddress[:]); if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[ContractMigrate] Find error!")
	}
	for _, v := range stateValues {
		key := new(states.StorageKey)
		bf := bytes.NewBuffer([]byte(v.Key))
		if err := key.Deserialize(bf); err != nil {
			return false, errors.NewErr("[ContractMigrate] Key deserialize error!")
		}
		key = &states.StorageKey{CodeHash: contractAddress, Key: key.Key}
		b := new(bytes.Buffer)
		if _, err := key.Serialize(b); err != nil {
			return false, errors.NewErr("[ContractMigrate] Key Serialize error!")
		}
		s.CloneCache.Add(scommon.ST_STORAGE, key.ToArray(), v.Value)
	}
	vm.PushData(engine, contractState)
	return s.ContractDestory(engine)
}

func (s *StateMachine) ContractDestory(engine *vm.ExecutionEngine) (bool, error) {
	context, err := engine.CurrentContext(); if err != nil {
		return false, err
	}
	hash, err := context.GetCodeHash(); if err != nil {
		return false, nil
	}
	item, err := s.CloneCache.Store.TryGet(scommon.ST_CONTRACT, hash[:]); if err != nil {
		return false, err
	}
	if item == nil {
		return false, nil
	}
	s.CloneCache.Delete(scommon.ST_CONTRACT, hash[:])
	stateValues, err := s.CloneCache.Store.Find(scommon.ST_CONTRACT, hash[:]); if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[ContractDestory] Find error!")
	}
	for _, v := range stateValues {
		s.CloneCache.Delete(scommon.ST_STORAGE, []byte(v.Key))
	}
	return true, nil
}

func (s *StateMachine) CheckStorageContext(context *StorageContext) (bool, error) {
	item, err := s.CloneCache.Get(scommon.ST_CONTRACT, context.codeHash[:])
	if err != nil {
		return false, err
	}
	if item == nil {
		return false, errors.NewErr(fmt.Sprintf("get contract by codehash=%v nil", context.codeHash))
	}
	return true, nil
}

func (s *StateMachine) StoragePut(engine *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(engine) < 3 {
		return false, errors.NewErr("[StoragePut] Too few input parameters ")
	}
	opInterface := vm.PopInteropInterface(engine); if opInterface == nil {
		return false, errors.NewErr("[StoragePut] Get StorageContext nil")
	}
	context := opInterface.(*StorageContext)
	key := vm.PopByteArray(engine)
	if len(key) > 1024 {
		return false, errors.NewErr("[StoragePut] Get Storage key to long")
	}
	value := vm.PopByteArray(engine)
	k, err := serializeStorageKey(context.codeHash, key); if err != nil {
		return false, err
	}
	s.CloneCache.Add(scommon.ST_STORAGE, k, &states.StorageItem{Value: value})
	return true, nil
}

func (s *StateMachine) StorageDelete(engine *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(engine) < 2 {
		return false, errors.NewErr("[StorageDelete] Too few input parameters ")
	}
	opInterface := vm.PopInteropInterface(engine)
	if opInterface == nil {
		return false, errors.NewErr("[StorageDelete] Get StorageContext nil")
	}
	context := opInterface.(*StorageContext)
	key := vm.PopByteArray(engine)
	k, err := serializeStorageKey(context.codeHash, key); if err != nil {
		return false, err
	}
	s.CloneCache.Delete(scommon.ST_STORAGE, k)
	return true, nil
}

func (s *StateMachine) StorageGet(engine *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(engine) < 2 {
		return false, errors.NewErr("[StorageGet] Too few input parameters ")
	}
	opInterface := vm.PopInteropInterface(engine)
	if opInterface == nil {
		return false, errors.NewErr("[StorageGet] Get StorageContext error!")
	}
	context := opInterface.(*StorageContext)
	if exist, err := s.CheckStorageContext(context); !exist {
		return false, err
	}
	key := vm.PopByteArray(engine)
	k, err := serializeStorageKey(context.codeHash, key); if err != nil {
		return false, err
	}
	item, err := s.CloneCache.Get(scommon.ST_STORAGE, k); if err != nil {
		return false, err
	}
	if item == nil {
		vm.PushData(engine, []byte{})
	} else {
		vm.PushData(engine, item.(*states.StorageItem).Value)
	}
	return true, nil
}

func (s *StateMachine) GetStorageContext(engine *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(engine) < 1 {
		return false, errors.NewErr("[GetStorageContext] Too few input parameters ")
	}
	opInterface := vm.PopInteropInterface(engine)
	if opInterface == nil {
		return false, errors.NewErr("[GetStorageContext] Get StorageContext nil")
	}
	contractState := opInterface.(*payload.DeployCode)
	codeHash := contractState.Code.AddressFromVmCode()
	item, err := s.CloneCache.Store.TryGet(scommon.ST_CONTRACT, codeHash[:])
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[GetStorageContext] Get StorageContext nil")
	}
	context, err := engine.CurrentContext(); if err != nil {
		return false, err
	}
	if item == nil {
		return false, errors.NewErr(fmt.Sprintf("[GetStorageContext] Get contract by codehash:%v nil", codeHash))
	}
	currentHash, err := context.GetCodeHash(); if err != nil {
		return false, err
	}
	if codeHash != currentHash {
		return false, errors.NewErr("[GetStorageContext] CodeHash not equal!")
	}
	vm.PushData(engine, &StorageContext{codeHash: codeHash})
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

func serializeStorageKey(codeHash common.Address, key []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.Write(codeHash[:])
	buf.Write(key)
	return buf.Bytes(), nil
}
