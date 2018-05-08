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

package native

import (
	"bytes"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/genesis"
	scommon "github.com/ontio/ontology/core/store/common"
	ctypes "github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native/states"
	"sync"
	"fmt"
	"github.com/ontio/ontology/common/log"
)

type ParamCache struct {
	lock   sync.RWMutex
	Params states.Params
}

var GLOBAL_PARAM = map[string]string{
	"init-key1": "init-value1",
	"init-key2": "init-value2",
	"init-key3": "init-value3",
	"init-key4": "init-value4",
}

type paramType byte

const (
	CURRENT_VALUE paramType = 0x00
	PREPARE_VALUE paramType = 0x01
)

var paramCache *ParamCache
var admin *states.Admin

func init() {
	Contracts[genesis.ParamContractAddress] = RegisterParamContract
	paramCache = new(ParamCache)
	paramCache.Params = make(map[string]string)
}

func ParamInit(native *NativeService) error {
	paramCache = new(ParamCache)
	paramCache.Params = make(map[string]string)
	contract := native.ContextRef.CurrentContext().ContractAddress
	initParams := new(states.Params)
	*initParams = make(map[string]string)
	for k, v := range GLOBAL_PARAM {
		(*initParams)[k] = v
	}
	native.CloneCache.Add(scommon.ST_STORAGE, getParamKey(contract, CURRENT_VALUE), getParamStorageItem(initParams))
	native.CloneCache.Add(scommon.ST_STORAGE, getParamKey(contract, PREPARE_VALUE), getParamStorageItem(initParams))
	admin = new(states.Admin)

	bookKeeepers, err := config.DefConfig.GetBookkeepers()
	if err != nil {
		return fmt.Errorf("GetBookkeepers error:%s", err)
	}
	initAddress := ctypes.AddressFromPubKey(bookKeeepers[0])
	copy((*admin)[:], initAddress[:])
	native.CloneCache.Add(scommon.ST_STORAGE, getAdminKey(contract, false), getAdminStorageItem(admin))
	return nil
}

func AcceptAdmin(native *NativeService) error {
	destinationAdmin := new(states.Admin)
	if err := destinationAdmin.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return errors.NewErr("[Accept Admin]Deserialize Admins failed!")
	}
	if !native.ContextRef.CheckWitness(common.Address(*destinationAdmin)) {
		return errors.NewErr("[Accept Admin]Authentication failed!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	getAdmin(native, contract)
	transferAdmin, err := getStorageAdmin(native, getAdminKey(contract, true))
	if err != nil || *transferAdmin != *destinationAdmin {
		return errors.NewDetailErr(err, errors.ErrNoCode,
			"[Accept Admin] Destination account hasn't been approved!")
	}
	// delete transfer admin item
	native.CloneCache.Delete(scommon.ST_STORAGE, getAdminKey(contract, true))
	// modify admin in database
	native.CloneCache.Add(scommon.ST_STORAGE, getAdminKey(contract, false), getAdminStorageItem(destinationAdmin))

	admin = destinationAdmin
	return nil
}

func TransferAdmin(native *NativeService) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	getAdmin(native, contract)
	if !native.ContextRef.CheckWitness(common.Address(*admin)) {
		return errors.NewErr("[Transfer Admin]Authentication failed!")
	}
	destinationAdmin := new(states.Admin)
	if err := destinationAdmin.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return errors.NewErr("[Transfer Admin]Deserialize Admins failed!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, getAdminKey(contract, true),
		getAdminStorageItem(destinationAdmin))
	return nil
}

func SetGlobalParam(native *NativeService) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	getAdmin(native, contract)
	if !native.ContextRef.CheckWitness(common.Address(*admin)) {
		return errors.NewErr("[Set Param]Authentication failed!")
	}
	params := new(states.Params)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return errors.NewErr("[Set Param]Deserialize failed!")
	}
	// read old param from database
	storageParams, err := getStorageParam(native, getParamKey(contract, PREPARE_VALUE))
	if err != nil {
		return err
	}
	// update param
	for key, value := range *params {
		(*storageParams)[key] = value
	}
	native.CloneCache.Add(scommon.ST_STORAGE, getParamKey(contract, PREPARE_VALUE),
		getParamStorageItem(storageParams))
	notifyParamSetSuccess(native, contract, *params)
	return nil
}

func CreateSnapshot(native *NativeService) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	getAdmin(native, contract)
	if !native.ContextRef.CheckWitness(common.Address(*admin)) {
		return errors.NewErr("[Create Snapshot]Authentication failed!")
	}
	// read prepare param
	prepareParam, err := getStorageParam(native, getParamKey(contract, PREPARE_VALUE))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Create Snapshot] storage error!")
	}
	if len(*prepareParam) == 0 {
		return errors.NewErr("[Create Snapshot] Prepare param doesn't exist!")
	}
	// set prepare value to current value, make it effective
	native.CloneCache.Add(scommon.ST_STORAGE, getParamKey(contract, CURRENT_VALUE), getParamStorageItem(prepareParam))
	// clear memory cache
	clearCache()
	return nil
}

func getAdmin(native *NativeService, contract common.Address) {
	if admin == nil || *admin == *new(states.Admin) {
		var err error
		// get admin from database
		admin, err = getStorageAdmin(native, getAdminKey(contract, false))
		// there are no admin in database
		if err != nil {
			bookKeeepers, err := config.DefConfig.GetBookkeepers()
			if err != nil {
				log.Errorf("GetBookkeepers error:%s", err)
				return
			}
			initAddress := ctypes.AddressFromPubKey(bookKeeepers[0])
			copy((*admin)[:], initAddress[:])
		}
	}
}

func clearCache() {
	paramCache.lock.Lock()
	defer paramCache.lock.Unlock()
	paramCache.Params = make(map[string]string)
}

func setCache(params *states.Params) {
	paramCache.lock.Lock()
	defer paramCache.lock.Unlock()
	paramCache.Params = *params
}

func getParamFromCache(key string) string {
	paramCache.lock.RLock()
	defer paramCache.lock.RUnlock()
	return paramCache.Params[key]
}

func RegisterParamContract(native *NativeService) {
	native.Register("init", ParamInit)
	native.Register("acceptAdmin", AcceptAdmin)
	native.Register("transferAdmin", TransferAdmin)
	native.Register("setGlobalParam", SetGlobalParam)
	native.Register("createSnapshot", CreateSnapshot)
}

func GetGlobalParam(native *NativeService, paramName string) (string, error) {
	if value := getParamFromCache(paramName); value != "" {
		return value, nil
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	storageParams, err := getStorageParam(native, getParamKey(contract, CURRENT_VALUE))
	if err != nil {
		return "", errors.NewDetailErr(err, errors.ErrNoCode, "[Get Param] storage error!")
	}
	if len(*storageParams) == 0 {
		return "", nil
	}
	// set param to cache
	setCache(storageParams)
	if value, ok := (*storageParams)[paramName]; ok {
		return value, nil
	} else {
		return "", errors.NewErr("[Get Param] param doesn't exist!")
	}
}
