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
	"bytes"

	"github.com/ontio/ontology/common"
	cstates "github.com/ontio/ontology/core/states"
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
	bf := new(bytes.Buffer)
	utils.WriteAddress(bf, role)
	return &cstates.StorageItem{Value: bf.Bytes()}
}

func getParamStorageItem(params *Params) *cstates.StorageItem {
	bf := new(bytes.Buffer)
	params.Serialize(bf)
	return &cstates.StorageItem{Value: bf.Bytes()}
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

func getStorageParam(native *native.NativeService, key []byte) (*Params, error) {
	item, err := utils.GetStorageItem(native, key)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	params := new(Params)
	bf := bytes.NewBuffer(item.Value)
	params.Deserialize(bf)
	return params, nil
}

func GetStorageRole(native *native.NativeService, key []byte) (common.Address, error) {
	item, err := utils.GetStorageItem(native, key)
	var role common.Address
	if err != nil || item == nil {
		return role, err
	}
	bf := bytes.NewBuffer(item.Value)
	role, err = utils.ReadAddress(bf)
	return role, err
}
