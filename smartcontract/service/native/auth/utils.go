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

package auth

import (
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/smartcontract/service/native"
)

var (
	RoleF        = []byte{0x01}
	RoleP        = []byte{0x02}
	FuncPerson   = []byte{0x03}
	DelegateList = []byte{0x04}
	Admin        = []byte{0x05}
)

//this.contractAddr.Admin
func GetContractAdminKey(native *native.NativeService, contractAddr []byte) ([]byte, error) {
	this := native.ContextRef.CurrentContext().ContractAddress
	adminKey, err := PackKeys(this[:], [][]byte{contractAddr, Admin})

	return adminKey, err
}

//this.contractAddr.RoleF.role
func GetRoleFKey(native *native.NativeService, contractAddr, role []byte) ([]byte, error) {
	this := native.ContextRef.CurrentContext().ContractAddress
	roleFKey, err := PackKeys(this[:], [][]byte{contractAddr, RoleF, role})

	return roleFKey, err
}

//this.contractAddr.RoleP.role
func GetRolePKey(native *native.NativeService, contractAddr, role []byte) ([]byte, error) {
	this := native.ContextRef.CurrentContext().ContractAddress
	rolePKey, err := PackKeys(this[:], [][]byte{contractAddr, RoleP, role})

	return rolePKey, err
}

//this.contractAddr.FuncOntID.func.ontID
func GetFuncOntIDKey(native *native.NativeService, contractAddr, fn, ontID []byte) ([]byte, error) {
	this := native.ContextRef.CurrentContext().ContractAddress
	funcOntIDKey, err := PackKeys(this[:], [][]byte{contractAddr, FuncPerson, fn, ontID})

	return funcOntIDKey, err
}

//this.contractAddr.DelegateList.role.ontID
func GetDelegateListKey(native *native.NativeService, contractAddr, role, ontID []byte) ([]byte, error) {
	this := native.ContextRef.CurrentContext().ContractAddress
	delegateListKey, err := PackKeys(this[:], [][]byte{contractAddr, DelegateList, role, ontID})

	return delegateListKey, err
}

func PutBytes(native *native.NativeService, key []byte, value []byte) {
	native.CloneCache.Add(common.ST_STORAGE, key, &states.StorageItem{Value: value})
}

func writeAuthToken(native *native.NativeService, contractAddr, fn, ontID, auth []byte) error {
	key, err := GetFuncOntIDKey(native, contractAddr, fn, ontID)
	if err != nil {
		return err
	}
	PutBytes(native, key, auth)
	return nil
}
