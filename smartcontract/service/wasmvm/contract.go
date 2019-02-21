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
	"github.com/go-interpreter/wagon/exec"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/errors"
)

func (self *Runtime) ContractCreate(proc *exec.Process,
	codePtr uint32,
	codeLen uint32,
	needStorage uint32,
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

	code := make([]byte, codeLen)
	_, err := proc.ReadAt(code, int64(codePtr))
	if err != nil {
		panic(err)
	}

	cost := CONTRACT_CREATE_GAS + uint64(uint64(codeLen)/PER_UNIT_CODE_LEN)*UINT_DEPLOY_CODE_LEN_GAS
	self.checkGas(cost)

	name := make([]byte, nameLen)
	_, err = proc.ReadAt(name, int64(namePtr))
	if err != nil {
		panic(err)
	}

	version := make([]byte, verLen)
	_, err = proc.ReadAt(version, int64(verPtr))
	if err != nil {
		panic(err)
	}

	author := make([]byte, authorLen)
	_, err = proc.ReadAt(author, int64(authorPtr))
	if err != nil {
		panic(err)
	}

	email := make([]byte, emailLen)
	_, err = proc.ReadAt(email, int64(emailPtr))
	if err != nil {
		panic(err)
	}

	desc := make([]byte, descLen)
	_, err = proc.ReadAt(desc, int64(descPtr))
	if err != nil {
		panic(err)
	}

	dep, err := self.isContractValid(code, needStorage, name, version, author, email, desc)
	if err != nil {
		panic(err)
	}

	contractAddr := dep.Address()
	if self.isContractExist(contractAddr) {
		panic(errors.NewErr("contract has been deployed"))
	}

	err = self.Service.CacheDB.PutContract(dep)
	if err != nil {
		panic(err)
	}

	length, err := proc.WriteAt(contractAddr[:], int64(newAddressPtr))
	return uint32(length)

}

func (self *Runtime) ContractMigrate(proc *exec.Process,
	codePtr uint32,
	codeLen uint32,
	needStorage uint32,
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

	code := make([]byte, codeLen)
	_, err := proc.ReadAt(code, int64(codePtr))
	if err != nil {
		panic(err)
	}

	cost := CONTRACT_CREATE_GAS + uint64(uint64(codeLen)/PER_UNIT_CODE_LEN)*UINT_DEPLOY_CODE_LEN_GAS
	self.checkGas(cost)

	name := make([]byte, nameLen)
	_, err = proc.ReadAt(name, int64(namePtr))
	if err != nil {
		panic(err)
	}

	version := make([]byte, verLen)
	_, err = proc.ReadAt(version, int64(verPtr))
	if err != nil {
		panic(err)
	}

	author := make([]byte, authorLen)
	_, err = proc.ReadAt(author, int64(authorPtr))
	if err != nil {
		panic(err)
	}

	email := make([]byte, emailLen)
	_, err = proc.ReadAt(email, int64(emailPtr))
	if err != nil {
		panic(err)
	}

	desc := make([]byte, descLen)
	_, err = proc.ReadAt(desc, int64(descPtr))
	if err != nil {
		panic(err)
	}

	dep, err := self.isContractValid(code, needStorage, name, version, author, email, desc)
	if err != nil {
		panic(err)
	}

	contractAddr := dep.Address()
	if self.isContractExist(contractAddr) {
		panic(errors.NewErr("contract has been deployed"))
	}
	oldAddress := self.Service.ContextRef.CurrentContext().ContractAddress

	iter := self.Service.CacheDB.NewIterator(oldAddress[:])
	for has := iter.First(); has; has = iter.Next() {
		key := iter.Key()
		val := iter.Value()

		newkey, err := serializeStorageKey(contractAddr, key)
		if err != nil {
			panic(err)
		}
		self.Service.CacheDB.Put(newkey, val)
		self.Service.CacheDB.Delete(key)
	}

	iter.Release()
	if err := iter.Error(); err != nil {
		panic(err)
	}

	length, err := proc.WriteAt(contractAddr[:], int64(newAddressPtr))
	if err != nil {
		panic(err)
	}

	return uint32(length)
}

func (self *Runtime) ContractDelete(proc *exec.Process) {
	contractAddress := self.Service.ContextRef.CurrentContext().ContractAddress
	iter := self.Service.CacheDB.NewIterator(contractAddress[:])

	for has := iter.First(); has; has = iter.Next() {
		self.Service.CacheDB.Delete(iter.Key())
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		panic(err)
	}

}

func (self *Runtime) isContractValid(code []byte,
	needStorage uint32,
	name []byte,
	version []byte,
	author []byte,
	email []byte,
	desc []byte) (*payload.DeployCode, error) {

	if len(code) > 1024*1024 {
		return nil, errors.NewErr("[Contract] Code too long!")
	}

	if len(name) > 252 {
		return nil, errors.NewErr("[Contract] name too long!")
	}

	if len(version) > 252 {
		return nil, errors.NewErr("[Contract] version too long!")
	}

	if len(author) > 252 {
		return nil, errors.NewErr("[author] version too long!")
	}

	if len(email) > 252 {
		return nil, errors.NewErr("[author] emailPtr too long!")
	}

	if len(desc) > 65536 {
		return nil, errors.NewErr("[descPtr] emailPtr too long!")
	}

	contract := &payload.DeployCode{
		Code:        code,
		NeedStorage: byte(needStorage),
		Name:        string(name),
		Version:     string(version),
		Author:      string(author),
		Email:       string(email),
		Description: string(desc),
	}
	return contract, nil
}

func (self *Runtime) isContractExist(contractAddress common.Address) bool {
	item, err := self.Service.CacheDB.GetContract(contractAddress)
	if err != nil {
		panic(err)
	}
	return item != nil
}
