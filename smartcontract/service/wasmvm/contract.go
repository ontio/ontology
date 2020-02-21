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
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/wagon/exec"
)

func migrateContractStorage(service *WasmVmService, newAddress common.Address) error {
	oldAddress := service.ContextRef.CurrentContext().ContractAddress
	service.CacheDB.DeleteContract(oldAddress)

	iter := service.CacheDB.NewIterator(oldAddress[:])
	for has := iter.First(); has; has = iter.Next() {
		key := iter.Key()
		val := iter.Value()

		newkey := serializeStorageKey(newAddress, key[20:])

		service.CacheDB.Put(newkey, val)
		service.CacheDB.Delete(key)
	}

	iter.Release()
	if err := iter.Error(); err != nil {
		return err
	}

	return nil
}

func deleteContractStorage(service *WasmVmService) error {
	contractAddress := service.ContextRef.CurrentContext().ContractAddress
	iter := service.CacheDB.NewIterator(contractAddress[:])

	for has := iter.First(); has; has = iter.Next() {
		service.CacheDB.Delete(iter.Key())
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return err
	}

	service.CacheDB.DeleteContract(contractAddress)
	return nil
}

func ContractCreate(proc *exec.Process,
	codePtr uint32,
	codeLen uint32,
	vmType uint32,
	namePtr uint32,
	nameLen uint32,
	verPtr uint32,
	verLen uint32,
	authorPtr uint32,
	authorLen uint32,
	emailPtr uint32,
	emailLen uint32,
	descPtr uint32,
	descLen uint32,
	newAddressPtr uint32) uint32 {
	self := proc.HostData().(*Runtime)
	code, err := ReadWasmMemory(proc, codePtr, codeLen)
	if err != nil {
		panic(err)
	}

	cost := CONTRACT_CREATE_GAS + uint64(uint64(codeLen)/PER_UNIT_CODE_LEN)*UINT_DEPLOY_CODE_LEN_GAS
	self.checkGas(cost)

	name, err := ReadWasmMemory(proc, namePtr, nameLen)
	if err != nil {
		panic(err)
	}

	version, err := ReadWasmMemory(proc, verPtr, verLen)
	if err != nil {
		panic(err)
	}

	author, err := ReadWasmMemory(proc, authorPtr, authorLen)
	if err != nil {
		panic(err)
	}

	email, err := ReadWasmMemory(proc, emailPtr, emailLen)
	if err != nil {
		panic(err)
	}

	desc, err := ReadWasmMemory(proc, descPtr, descLen)
	if err != nil {
		panic(err)
	}

	dep, err := payload.CreateDeployCode(code, vmType, name, version, author, email, desc)
	if err != nil {
		panic(err)
	}

	wasmCode, err := dep.GetWasmCode()
	if err != nil {
		panic(err)
	}
	_, err = ReadWasmModule(wasmCode, config.DefConfig.Common.WasmVerifyMethod)
	if err != nil {
		panic(err)
	}

	contractAddr := dep.Address()
	if self.isContractExist(contractAddr) {
		panic(errors.NewErr("contract has been deployed"))
	}

	self.Service.CacheDB.PutContract(dep)

	length, err := proc.WriteAt(contractAddr[:], int64(newAddressPtr))
	return uint32(length)

}

func ContractMigrate(proc *exec.Process,
	codePtr uint32,
	codeLen uint32,
	vmType uint32,
	namePtr uint32,
	nameLen uint32,
	verPtr uint32,
	verLen uint32,
	authorPtr uint32,
	authorLen uint32,
	emailPtr uint32,
	emailLen uint32,
	descPtr uint32,
	descLen uint32,
	newAddressPtr uint32) uint32 {

	self := proc.HostData().(*Runtime)

	code, err := ReadWasmMemory(proc, codePtr, codeLen)
	if err != nil {
		panic(err)
	}

	cost := CONTRACT_CREATE_GAS + uint64(uint64(codeLen)/PER_UNIT_CODE_LEN)*UINT_DEPLOY_CODE_LEN_GAS
	self.checkGas(cost)

	name, err := ReadWasmMemory(proc, namePtr, nameLen)
	if err != nil {
		panic(err)
	}

	version, err := ReadWasmMemory(proc, verPtr, verLen)
	if err != nil {
		panic(err)
	}

	author, err := ReadWasmMemory(proc, authorPtr, authorLen)
	if err != nil {
		panic(err)
	}

	email, err := ReadWasmMemory(proc, emailPtr, emailLen)
	if err != nil {
		panic(err)
	}

	desc, err := ReadWasmMemory(proc, descPtr, descLen)
	if err != nil {
		panic(err)
	}

	dep, err := payload.CreateDeployCode(code, vmType, name, version, author, email, desc)
	if err != nil {
		panic(err)
	}

	wasmCode, err := dep.GetWasmCode()
	if err != nil {
		panic(err)
	}
	_, err = ReadWasmModule(wasmCode, config.DefConfig.Common.WasmVerifyMethod)
	if err != nil {
		panic(err)
	}

	contractAddr := dep.Address()
	if self.isContractExist(contractAddr) {
		panic(errors.NewErr("contract has been deployed"))
	}
	self.Service.CacheDB.PutContract(dep)

	err = migrateContractStorage(self.Service, contractAddr)
	if err != nil {
		panic(err)
	}

	length, err := proc.WriteAt(contractAddr[:], int64(newAddressPtr))
	if err != nil {
		panic(err)
	}

	return uint32(length)
}

func ContractDestroy(proc *exec.Process) {
	self := proc.HostData().(*Runtime)
	err := deleteContractStorage(self.Service)
	if err != nil {
		panic(err)
	}
	//the contract has been deleted ,quit the contract operation
	proc.Terminate()
}

func (self *Runtime) isContractExist(contractAddress common.Address) bool {
	item, err := self.Service.CacheDB.GetContract(contractAddress)
	if err != nil {
		panic(err)
	}
	return item != nil
}
