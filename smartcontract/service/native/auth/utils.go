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
	"bytes"
	"fmt"

	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/service/native"
)

var (
	RoleF        = []byte{0x01}
	RoleP        = []byte{0x02}
	DelegateList = []byte{0x04}
	Admin        = []byte{0x05}
)

//type(this.contractAddr.Admin) = []byte
func GetContractAdminKey(native *native.NativeService, contractAddr []byte) ([]byte, error) {
	this := native.ContextRef.CurrentContext().ContractAddress
	adminKey, err := packKeys(this[:], [][]byte{contractAddr, Admin})

	return adminKey, err
}

//type(this.contractAddr.RoleF.role) = roleFuncs
func GetRoleFKey(native *native.NativeService, contractAddr, role []byte) ([]byte, error) {
	this := native.ContextRef.CurrentContext().ContractAddress
	roleFKey, err := packKeys(this[:], [][]byte{contractAddr, RoleF, role})

	return roleFKey, err
}

//type(this.contractAddr.RoleP.ontID) = roleTokens
func GetRolePKey(native *native.NativeService, contractAddr, ontID []byte) ([]byte, error) {
	this := native.ContextRef.CurrentContext().ContractAddress
	rolePKey, err := packKeys(this[:], [][]byte{contractAddr, RoleP, ontID})

	return rolePKey, err
}

//type(this.contractAddr.DelegateList.role.ontID)
func GetDelegateListKey(native *native.NativeService, contractAddr, role, ontID []byte) ([]byte, error) {
	this := native.ContextRef.CurrentContext().ContractAddress
	delegateListKey, err := packKeys(this[:], [][]byte{contractAddr, DelegateList, role, ontID})

	return delegateListKey, err
}

//remote duplicates in the slice of string
func stringSliceUniq(s []string) []string {
	smap := make(map[string]int)
	for i, str := range s {
		if str == "" {
			continue
		}
		smap[str] = i
	}
	ret := make([]string, len(smap))
	i := 0
	for str, _ := range smap {
		ret[i] = str
		i++
	}
	return ret
}

func packKeys(field []byte, items [][]byte) ([]byte, error) {
	w := new(bytes.Buffer)
	for _, item := range items {
		err := serialization.WriteVarBytes(w, item)
		if err != nil {
			return nil, fmt.Errorf("packKeys failed when serialize %x", item)
		}
	}
	key := append(field, w.Bytes()...)
	return key, nil
}

//pack data to be used as a key in the kv storage
// key := field || ser_data
func packKey(field []byte, data []byte) ([]byte, error) {
	return packKeys(field, [][]byte{data})
}
