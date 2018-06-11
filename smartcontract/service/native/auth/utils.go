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
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

var (
	PreAdmin          = []byte{0x01}
	PreRoleFunc       = []byte{0x02}
	PreRoleToken      = []byte{0x03}
	PreDelegateStatus = []byte{0x04}
	//PreDelegateList   = []byte{0x04}
)

//type(this.contractAddr.Admin) = []byte
func concatContractAdminKey(native *native.NativeService, contractAddr []byte) ([]byte, error) {
	this := native.ContextRef.CurrentContext().ContractAddress
	adminKey, err := packKeys(this, [][]byte{contractAddr, PreAdmin})

	return adminKey, err
}

func getContractAdmin(native *native.NativeService, contractAddr []byte) ([]byte, error) {
	key, err := concatContractAdminKey(native, contractAddr)
	if err != nil {
		return nil, err
	}
	item, err := utils.GetStorageItem(native, key)
	if err != nil {
		return nil, err
	}
	if item == nil { //is not set
		return nil, nil
	}
	return item.Value, nil
}

func putContractAdmin(native *native.NativeService, contractAddr, adminOntID []byte) error {
	key, err := concatContractAdminKey(native, contractAddr)
	if err != nil {
		return err
	}
	utils.PutBytes(native, key, adminOntID)
	return nil
}

//type(this.contractAddr.RoleFunc.role) = roleFuncs
func concatRoleFuncKey(native *native.NativeService, contractAddr, role []byte) ([]byte, error) {
	this := native.ContextRef.CurrentContext().ContractAddress
	roleFuncKey, err := packKeys(this, [][]byte{contractAddr, PreRoleFunc, role})

	return roleFuncKey, err
}

func getRoleFunc(native *native.NativeService, contractAddr, role []byte) (*roleFuncs, error) {
	key, err := concatRoleFuncKey(native, contractAddr, role)
	if err != nil {
		return nil, err
	}
	item, err := utils.GetStorageItem(native, key)
	if err != nil {
		return nil, err
	}
	if item == nil { //is not set
		return nil, nil
	}
	rd := bytes.NewReader(item.Value)
	rF := new(roleFuncs)
	err = rF.Deserialize(rd)
	if err != nil {
		return nil, fmt.Errorf("deserialize roleFuncs object failed. data: %x", item.Value)
	}
	return rF, nil
}

func putRoleFunc(native *native.NativeService, contractAddr, role []byte, funcs *roleFuncs) error {
	key, _ := concatRoleFuncKey(native, contractAddr, role)
	bf := new(bytes.Buffer)
	err := funcs.Serialize(bf)
	if err != nil {
		return fmt.Errorf("serialize roleFuncs failed, caused by %v", err)
	}
	utils.PutBytes(native, key, bf.Bytes())
	return nil
}

//type(this.contractAddr.RoleP.ontID) = roleTokens
func concatOntIDTokenKey(native *native.NativeService, contractAddr, ontID []byte) ([]byte, error) {
	this := native.ContextRef.CurrentContext().ContractAddress
	tokenKey, err := packKeys(this, [][]byte{contractAddr, PreRoleToken, ontID})

	return tokenKey, err
}

func getOntIDToken(native *native.NativeService, contractAddr, ontID []byte) (*roleTokens, error) {
	key, err := concatOntIDTokenKey(native, contractAddr, ontID)
	if err != nil {
		return nil, err
	}
	item, err := utils.GetStorageItem(native, key)
	if err != nil {
		return nil, err
	}
	if item == nil { //is not set
		return nil, nil
	}
	rd := bytes.NewReader(item.Value)
	rT := new(roleTokens)
	err = rT.Deserialize(rd)
	if err != nil {
		return nil, fmt.Errorf("deserialize roleTokens object failed. data: %x", item.Value)
	}
	return rT, nil
}

func putOntIDToken(native *native.NativeService, contractAddr, ontID []byte, tokens *roleTokens) error {
	key, _ := concatOntIDTokenKey(native, contractAddr, ontID)
	bf := new(bytes.Buffer)
	err := tokens.Serialize(bf)
	if err != nil {
		return fmt.Errorf("serialize roleFuncs failed, caused by %v", err)
	}
	utils.PutBytes(native, key, bf.Bytes())
	return nil
}

//type(this.contractAddr.DelegateStatus.ontID)
func concatDelegateStatusKey(native *native.NativeService, contractAddr, ontID []byte) ([]byte, error) {
	this := native.ContextRef.CurrentContext().ContractAddress
	key, err := packKeys(this, [][]byte{contractAddr, PreDelegateStatus, ontID})

	return key, err
}

func getDelegateStatus(native *native.NativeService, contractAddr, ontID []byte) (*Status, error) {
	key, err := concatDelegateStatusKey(native, contractAddr, ontID)
	if err != nil {
		return nil, err
	}
	item, err := utils.GetStorageItem(native, key)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	status := new(Status)
	rd := bytes.NewReader(item.Value)
	err = status.Deserialize(rd)
	if err != nil {
		return nil, fmt.Errorf("deserialize Status object failed. data: %x", item.Value)
	}
	return status, nil
}

func putDelegateStatus(native *native.NativeService, contractAddr, ontID []byte, status *Status) error {
	key, _ := concatDelegateStatusKey(native, contractAddr, ontID)
	bf := new(bytes.Buffer)
	err := status.Serialize(bf)
	if err != nil {
		return fmt.Errorf("serialize Status failed, caused by %v", err)
	}
	utils.PutBytes(native, key, bf.Bytes())
	return nil
}

/*
 * pack data to be used as a key in the kv storage
 * key := field || ser_items[1] || ... || ser_items[n]
 */
func packKeys(field common.Address, items [][]byte) ([]byte, error) {
	w := new(bytes.Buffer)
	for _, item := range items {
		err := serialization.WriteVarBytes(w, item)
		if err != nil {
			return nil, fmt.Errorf("packKeys failed when serialize %x", item)
		}
	}
	key := append(field[:], w.Bytes()...)
	return key, nil
}

/*
 * pack data to be used as a key in the kv storage
 * key := field || ser_data
 */
func packKey(field common.Address, data []byte) ([]byte, error) {
	return packKeys(field, [][]byte{data})
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

func pushEvent(native *native.NativeService, s interface{}) {
	event := new(event.NotifyEventInfo)
	event.ContractAddress = native.ContextRef.CurrentContext().ContractAddress
	event.States = s
	native.Notifications = append(native.Notifications, event)
}

func invokeEvent(native *native.NativeService, fn string, ret bool) {
	pushEvent(native, []interface{}{fn, ret})
}

func serializeAddress(w io.Writer, addr common.Address) error {
	err := serialization.WriteVarBytes(w, addr[:])
	if err != nil {
		return err
	}
	return nil
}
