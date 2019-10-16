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

package global_params

import (
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	cstates "github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	PARAM    = "param"
	TRANSFER = "transfer"
	ADMIN    = "admin"
	OPERATOR = "operator"
)

func getRoleStorageItem(role common.Address) *cstates.StorageItem {
	bf := common.NewZeroCopySink(nil)
	utils.EncodeAddress(bf, role)
	return &cstates.StorageItem{Value: bf.Bytes()}
}

func getParamStorageItem(params Params) *cstates.StorageItem {
	return &cstates.StorageItem{Value: common.SerializeToBytes(&params)}
}

func generateParamKey(contract common.Address, valueType paramType) []byte {
	key := append(contract[:], PARAM...)
	key = append(key[:], byte(valueType))
	return key
}

func generateAdminKey(contract common.Address, isTransferAdmin bool) []byte {
	if isTransferAdmin {
		return append(contract[:], TRANSFER...)
	} else {
		return append(contract[:], ADMIN...)
	}
}

func GenerateOperatorKey(contract common.Address) []byte {
	return append(contract[:], OPERATOR...)
}

func getStorageParam(native *native.NativeService, key []byte) (Params, error) {
	item, err := utils.GetStorageItem(native, key)
	params := Params{}
	if err != nil || item == nil {
		return params, err
	}
	err = params.Deserialization(common.NewZeroCopySource(item.Value))
	return params, err
}

func GetStorageRole(native *native.NativeService, key []byte) (common.Address, error) {
	item, err := utils.GetStorageItem(native, key)
	var role common.Address
	if err != nil || item == nil {
		return role, err
	}
	bf := common.NewZeroCopySource(item.Value)
	role, err = utils.DecodeAddress(bf)
	return role, err
}

func NotifyRoleChange(native *native.NativeService, contract common.Address, functionName string,
	newAddr common.Address) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          []interface{}{functionName, newAddr.ToBase58()},
		})
}

func NotifyTransferAdmin(native *native.NativeService, contract common.Address, functionName string,
	originAdmin, newAdmin common.Address) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          []interface{}{functionName, originAdmin.ToBase58(), newAdmin.ToBase58()},
		})
}

func NotifyParamChange(native *native.NativeService, contract common.Address, functionName string, params Params) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	paramsString := ""
	for _, param := range params {
		paramsString += param.Key + "," + param.Value + ";"
	}
	paramsString = paramsString[:len(paramsString)-1]
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          []interface{}{functionName, paramsString},
		})
}
