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
	"sync"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type ParamCache struct {
	lock   sync.RWMutex
	Params Params
}

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

var paramCache *ParamCache

func InitGlobalParams() {
	native.Contracts[utils.ParamContractAddress] = RegisterParamContract
	paramCache = new(ParamCache)
	paramCache.Params = make([]*Param, 0)
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
	paramCache = new(ParamCache)
	paramCache.Params = make([]*Param, 0)
	contract := native.ContextRef.CurrentContext().ContractAddress
	initParams := new(Params)
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

	admin := new(Role)
	if err := admin.Deserialize(argsBuffer); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "init param, deserialize admin failed!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, generateAdminKey(contract, false), getRoleStorageItem(admin))
	operator := admin
	native.CloneCache.Add(scommon.ST_STORAGE, GenerateOperatorKey(contract), getRoleStorageItem(operator))
	return utils.BYTE_TRUE, nil
}

func AcceptAdmin(native *native.NativeService) ([]byte, error) {
	destinationAdmin := new(Role)
	if err := destinationAdmin.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("accept admin, deserialize admin failed!")
	}
	if !native.ContextRef.CheckWitness(common.Address(*destinationAdmin)) {
		return utils.BYTE_FALSE, errors.NewErr("accept admin, authentication failed!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	transferAdmin, err := GetStorageRole(native, generateAdminKey(contract, true))
	if err != nil || transferAdmin == nil || *transferAdmin != *destinationAdmin {
		return utils.BYTE_FALSE, fmt.Errorf("accept admin, destination account hasn't been approved, casused by %v", err)
	}
	// delete transfer admin item
	native.CloneCache.Delete(scommon.ST_STORAGE, generateAdminKey(contract, true))
	// modify admin in database
	native.CloneCache.Add(scommon.ST_STORAGE, generateAdminKey(contract, false), getRoleStorageItem(destinationAdmin))

	return utils.BYTE_TRUE, nil
}

func TransferAdmin(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	admin, err := GetStorageRole(native, generateAdminKey(contract, false))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "transfer admin, admin doesn't exist!")
	}
	if !native.ContextRef.CheckWitness(common.Address(*admin)) {
		return utils.BYTE_FALSE, errors.NewErr("transfer admin, authentication failed!")
	}
	destinationAdmin := new(Role)
	if err := destinationAdmin.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("transfer admin, deserialize admin failed!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, generateAdminKey(contract, true),
		getRoleStorageItem(destinationAdmin))
	return utils.BYTE_TRUE, nil
}

func SetOperator(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	admin, err := GetStorageRole(native, generateAdminKey(contract, false))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "set operator, admin doesn't exist!")
	}
	if !native.ContextRef.CheckWitness(common.Address(*admin)) {
		return utils.BYTE_FALSE, errors.NewErr("set operator, authentication failed!")
	}
	destinationOperator := new(Role)
	if err := destinationOperator.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("set operator, deserialize operator failed!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, GenerateOperatorKey(contract), getRoleStorageItem(destinationOperator))
	return utils.BYTE_TRUE, nil
}

func SetGlobalParam(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	operator, err := GetStorageRole(native, GenerateOperatorKey(contract))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "set param, operator doesn't exist!")
	}
	if !native.ContextRef.CheckWitness(common.Address(*operator)) {
		return utils.BYTE_FALSE, errors.NewErr("set param, authentication failed!")
	}
	params := new(Params)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("set param, deserialize failed!")
	}
	// read old param from database
	storageParams, err := getStorageParam(native, generateParamKey(contract, PREPARE_VALUE))
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	// update param
	for _, param := range *params {
		storageParams.SetParam(param)
	}
	native.CloneCache.Add(scommon.ST_STORAGE, generateParamKey(contract, PREPARE_VALUE),
		getParamStorageItem(storageParams))
	return utils.BYTE_TRUE, nil
}

func GetGlobalParam(native *native.NativeService) ([]byte, error) {
	paramNameList := new(ParamNameList)
	if err := paramNameList.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("get param, deserialize failed!")
	}
	params := new(Params)
	var paramNotInCache = make([]string, 0)
	// read from cache
	for _, paramName := range *paramNameList {
		if index, value := getParamFromCache(paramName); index >= 0 {
			params.SetParam(value)
		} else {
			paramNotInCache = append(paramNotInCache, paramName)
		}
	}
	result := new(bytes.Buffer)
	if len(paramNotInCache) == 0 { // all request param exist in cache
		if err := params.Serialize(result); err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "get param, results seriealize error!")
		}
		return result.Bytes(), nil
	}
	// read from db
	contract := native.ContextRef.CurrentContext().ContractAddress
	storageParams, err := getStorageParam(native, generateParamKey(contract, CURRENT_VALUE))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "get param, storage error!")
	}
	if len(*storageParams) == 0 {
		return utils.BYTE_FALSE, errors.NewErr("get param, there are no params!")
	}
	setCache(storageParams)                     // set param to cache
	for _, paramName := range paramNotInCache { // read param not in cache
		if index, value := storageParams.GetParam(paramName); index >= 0 {
			params.SetParam(value)
		} else {
			return utils.BYTE_FALSE, errors.NewErr(fmt.Sprintf("get param, param %v doesn't exist!", paramName))
		}
	}
	err = params.Serialize(result)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "get param, results to json error!")
	}
	return result.Bytes(), nil
}

func CreateSnapshot(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	operator, err := GetStorageRole(native, GenerateOperatorKey(contract))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "create snapshot, operator doesn't exist!")
	}
	if !native.ContextRef.CheckWitness(common.Address(*operator)) {
		return utils.BYTE_FALSE, errors.NewErr("create snapshot, authentication failed!")
	}
	// read prepare param
	prepareParam, err := getStorageParam(native, generateParamKey(contract, PREPARE_VALUE))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "create snapshot, storage error!")
	}
	if len(*prepareParam) == 0 {
		return utils.BYTE_FALSE, errors.NewErr("create snapshot, prepare param doesn't exist!")
	}
	// set prepare value to current value, make it effective
	native.CloneCache.Add(scommon.ST_STORAGE, generateParamKey(contract, CURRENT_VALUE), getParamStorageItem(prepareParam))
	// clear memory cache
	clearCache()
	return utils.BYTE_TRUE, nil
}

func clearCache() {
	paramCache.lock.Lock()
	defer paramCache.lock.Unlock()
	paramCache.Params = make([]*Param, 0)
}

func setCache(params *Params) {
	paramCache.lock.Lock()
	defer paramCache.lock.Unlock()
	paramCache.Params = *params
}

func getParamFromCache(key string) (int, *Param) {
	paramCache.lock.RLock()
	defer paramCache.lock.RUnlock()
	return paramCache.Params.GetParam(key)
}
