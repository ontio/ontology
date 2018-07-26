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
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type paramType byte

const (
	VERSION_CONTRACT_GLOBAL_PARAMS           = byte(0)
	CURRENT_VALUE                  paramType = 0x00
	PREPARE_VALUE                  paramType = 0x01
	INIT_NAME                                = "init"
	ACCEPT_ADMIN_NAME                        = "acceptAdmin"
	TRANSFER_ADMIN_NAME                      = "transferAdmin"
	SET_OPERATOR                             = "setOperator"
	SET_GLOBAL_PARAM_NAME                    = "setGlobalParam"
	GET_GLOBAL_PARAM_NAME                    = "getGlobalParam"
	CREATE_SNAPSHOT_NAME                     = "createSnapshot"
)

func InitGlobalParams() {
	native.Contracts[utils.ParamContractAddress] = RegisterParamContract
}

func RegisterParamContract(native *native.NativeService) {
	native.Register(INIT_NAME, ParamInit)
	native.Register(ACCEPT_ADMIN_NAME, AcceptAdmin)
	native.Register(TRANSFER_ADMIN_NAME, TransferAdmin)
	native.Register(SET_OPERATOR, SetOperator)
	native.Register(SET_GLOBAL_PARAM_NAME, SetGlobalParam)
	native.Register(GET_GLOBAL_PARAM_NAME, GetGlobalParam)
	native.Register(CREATE_SNAPSHOT_NAME, CreateSnapshot)
}

func ParamInit(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	storageAdmin, _ := GetStorageRole(native, generateAdminKey(contract, false))
	storageOperator, _ := GetStorageRole(native, generateAdminKey(contract, false))
	if storageAdmin != common.ADDRESS_EMPTY || storageOperator != common.ADDRESS_EMPTY {
		return utils.BYTE_FALSE, errors.NewErr("init param, admin or operator has already existed!")
	}

	initParams := Params{}
	args, err := serialization.ReadVarBytes(bytes.NewBuffer(native.Input))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "init param, read native input failed!")
	}
	argsBuffer := bytes.NewBuffer(args)
	if err := initParams.Deserialize(argsBuffer); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "init param, deserialize params failed!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, generateParamKey(contract, CURRENT_VALUE), getParamStorageItem(initParams))
	native.CloneCache.Add(scommon.ST_STORAGE, generateParamKey(contract, PREPARE_VALUE), getParamStorageItem(initParams))

	var admin common.Address
	if admin, err = utils.ReadAddress(argsBuffer); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "init param, deserialize admin failed!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, generateAdminKey(contract, false), getRoleStorageItem(admin))
	operator := admin
	native.CloneCache.Add(scommon.ST_STORAGE, GenerateOperatorKey(contract), getRoleStorageItem(operator))
	return utils.BYTE_TRUE, nil
}

func AcceptAdmin(native *native.NativeService) ([]byte, error) {
	var destinationAdmin common.Address
	destinationAdmin, err := utils.ReadAddress(bytes.NewBuffer(native.Input))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("accept admin, deserialize admin failed!")
	}
	if !native.ContextRef.CheckWitness(destinationAdmin) {
		return utils.BYTE_FALSE, errors.NewErr("accept admin, authentication failed!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	transferAdmin, err := GetStorageRole(native, generateAdminKey(contract, true))
	if err != nil || transferAdmin == common.ADDRESS_EMPTY || transferAdmin != destinationAdmin {
		return utils.BYTE_FALSE, fmt.Errorf("accept admin, destination account hasn't been approved, caused by %v", err)
	}
	// delete transfer admin item
	native.CloneCache.Delete(scommon.ST_STORAGE, generateAdminKey(contract, true))
	// modify admin in database
	native.CloneCache.Add(scommon.ST_STORAGE, generateAdminKey(contract, false), getRoleStorageItem(destinationAdmin))

	NotifyRoleChange(native, contract, ACCEPT_ADMIN_NAME, destinationAdmin)
	return utils.BYTE_TRUE, nil
}

func TransferAdmin(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	admin, err := GetStorageRole(native, generateAdminKey(contract, false))
	if err != nil || admin == common.ADDRESS_EMPTY {
		return utils.BYTE_FALSE, fmt.Errorf("transfer admin, admin doesn't exist, caused by %v", err)
	}
	if !native.ContextRef.CheckWitness(admin) {
		return utils.BYTE_FALSE, errors.NewErr("transfer admin, authentication failed!")
	}
	destinationAdmin, err := utils.ReadAddress(bytes.NewBuffer(native.Input))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("transfer admin, deserialize admin failed!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, generateAdminKey(contract, true),
		getRoleStorageItem(destinationAdmin))

	NotifyTransferAdmin(native, contract, TRANSFER_ADMIN_NAME, admin, destinationAdmin)
	return utils.BYTE_TRUE, nil
}

func SetOperator(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	admin, err := GetStorageRole(native, generateAdminKey(contract, false))
	if err != nil || admin == common.ADDRESS_EMPTY {
		return utils.BYTE_FALSE, fmt.Errorf("set operator, admin doesn't exist, caused by %v", err)
	}
	if !native.ContextRef.CheckWitness(admin) {
		return utils.BYTE_FALSE, errors.NewErr("set operator, authentication failed!")
	}
	destinationOperator, err := utils.ReadAddress(bytes.NewBuffer(native.Input))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("set operator, deserialize operator failed!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, GenerateOperatorKey(contract), getRoleStorageItem(destinationOperator))

	NotifyRoleChange(native, contract, SET_OPERATOR, destinationOperator)
	return utils.BYTE_TRUE, nil
}

func SetGlobalParam(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	operator, err := GetStorageRole(native, GenerateOperatorKey(contract))
	if err != nil || operator == common.ADDRESS_EMPTY {
		return utils.BYTE_FALSE, fmt.Errorf("set param, operator doesn't exist, caused by %v", err)
	}
	if !native.ContextRef.CheckWitness(operator) {
		return utils.BYTE_FALSE, errors.NewErr("set param, authentication failed!")
	}
	params := Params{}
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("set param, deserialize failed!")
	}
	if len(params) == 0 {
		return utils.BYTE_FALSE, errors.NewErr("set param, params is nil!")
	}
	// read old param from database
	storageParams, err := getStorageParam(native, generateParamKey(contract, PREPARE_VALUE))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode,
			"set param, read storage prepare param error!")
	}
	// update param
	for _, param := range params {
		storageParams.SetParam(param)
	}
	native.CloneCache.Add(scommon.ST_STORAGE, generateParamKey(contract, PREPARE_VALUE),
		getParamStorageItem(storageParams))

	NotifyParamChange(native, contract, SET_GLOBAL_PARAM_NAME, params)
	return utils.BYTE_TRUE, nil
}

func GetGlobalParam(native *native.NativeService) ([]byte, error) {
	var paramNameList ParamNameList
	if err := paramNameList.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("get param, deserialize failed!")
	}
	if len(paramNameList) == 0 {
		return utils.BYTE_FALSE, errors.NewErr("get param, required params is nil!")
	}
	// read from db
	contract := native.ContextRef.CurrentContext().ContractAddress
	storageParams, err := getStorageParam(native, generateParamKey(contract, CURRENT_VALUE))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode,
			"get param, read storage current param error!")
	}
	if len(storageParams) == 0 {
		return utils.BYTE_FALSE, errors.NewErr("get param, there are no params!")
	}
	params := new(Params)
	for _, paramName := range paramNameList { // read param not in cache
		if index, value := storageParams.GetParam(paramName); index >= 0 {
			params.SetParam(value)
		} else {
			params.SetParam(Param{Key: paramName, Value: ""})
		}
	}
	result := new(bytes.Buffer)
	err = params.Serialize(result)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "get param, serialize result error!")
	}
	return result.Bytes(), nil
}

func CreateSnapshot(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	operator, err := GetStorageRole(native, GenerateOperatorKey(contract))
	if err != nil || operator == common.ADDRESS_EMPTY {
		return utils.BYTE_FALSE, fmt.Errorf("create snapshot, operator doesn't exist, caused by %v", err)
	}
	if !native.ContextRef.CheckWitness(operator) {
		return utils.BYTE_FALSE, errors.NewErr("create snapshot, authentication failed!")
	}
	// read prepare param
	prepareParam, err := getStorageParam(native, generateParamKey(contract, PREPARE_VALUE))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode,
			"create snapshot, read storage prepare param error!")
	}
	if len(prepareParam) == 0 {
		return utils.BYTE_FALSE, errors.NewErr("create snapshot, prepare param doesn't exist!")
	}
	// set prepare value to current value, make it effective
	native.CloneCache.Add(scommon.ST_STORAGE, generateParamKey(contract, CURRENT_VALUE), getParamStorageItem(prepareParam))

	NotifyParamChange(native, contract, CREATE_SNAPSHOT_NAME, prepareParam)
	return utils.BYTE_TRUE, nil
}
