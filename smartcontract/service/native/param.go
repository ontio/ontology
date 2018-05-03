package native

import (
	"bytes"
	"sync"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/genesis"
	cstates "github.com/ontio/ontology/core/states"
	scommon "github.com/ontio/ontology/core/store/common"
	ctypes "github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native/states"
)

type ParamCache struct {
	lock   sync.RWMutex
	Params map[string]string
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
	for k, v := range GLOBAL_PARAM {
		native.CloneCache.Add(scommon.ST_STORAGE, getParamKey(contract, k, CURRENT_VALUE), getParamStorageItem(v))
		paramCache.Params[k] = v
	}
	admin = new(states.Admin)
	admin.Address = ctypes.AddressFromPubKey(account.GetBookkeepers()[0])
	native.CloneCache.Add(scommon.ST_STORAGE, getAdminKey(contract), getAdminStorageItem(admin))
	return nil
}

func TransferAdmin(native *NativeService) error {
	destinationAdmin := new(states.Admin)
	if err := destinationAdmin.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return errors.NewErr("[Transfer Admin]Deserialize Admins failed!")
	}
	if !native.ContextRef.CheckWitness(destinationAdmin.Address) {
		return errors.NewErr("[Transfer Admin]Authentication failed!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	initAdmin(native, contract)
	transferAdmin, err := getStorageAdmin(native, getTransferAdminKey(contract, admin.Address, destinationAdmin.Address))
	if err != nil || transferAdmin.Address != destinationAdmin.Address {
		return errors.NewDetailErr(err, errors.ErrNoCode,
			"[Transfer Admin] Destination account hasn't been approved!")
	}
	// delete transfer admin item
	native.CloneCache.Delete(scommon.ST_STORAGE, getTransferAdminKey(contract, admin.Address, destinationAdmin.Address))
	// modify admin in database
	native.CloneCache.Add(scommon.ST_STORAGE, getAdminKey(contract), getAdminStorageItem(destinationAdmin))

	admin = destinationAdmin
	return nil
}

func ApproveAdmin(native *NativeService) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	initAdmin(native, contract)
	if !native.ContextRef.CheckWitness(admin.Address) {
		return errors.NewErr("[Approve Admin]Authentication failed!")
	}
	destinationAdmin := new(states.Admin)
	if err := destinationAdmin.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return errors.NewErr("[Approve Admin]Deserialize Admins failed!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, getTransferAdminKey(contract, admin.Address, destinationAdmin.Address),
		getAdminStorageItem(destinationAdmin))
	return nil
}

func SetParam(native *NativeService) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	initAdmin(native, contract)
	if !native.ContextRef.CheckWitness(admin.Address) {
		return errors.NewErr("[Set Param]Authentication failed!")
	}
	params := new(states.Params)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return errors.NewErr("[Set Param]Deserialize failed!")
	}
	for _, param := range params.ParamList {
		native.CloneCache.Add(scommon.ST_STORAGE, getParamKey(contract, param.K, PREPARE_VALUE),
			getParamStorageItem(param.V))
		notifyParamSetSuccess(native, contract, param)
	}
	return nil
}

func EnforceParam(native *NativeService) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	initAdmin(native, contract)
	if !native.ContextRef.CheckWitness(admin.Address) {
		return errors.NewErr("[Enforce Param]Authentication failed!")
	}
	params := new(states.Params)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return errors.NewErr("[Enforce Param]Deserialize failed!")
	}
	for _, param := range params.ParamList {
		paramName := param.K
		// read prepare value
		value, err := native.CloneCache.Get(scommon.ST_STORAGE,
			getParamKey(genesis.ParamContractAddress, paramName, PREPARE_VALUE))
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[Enforce Param] storage error!")
		}
		if value == nil {
			return errors.NewErr("[Enforce Param] Prepare param doesn't exist!")
		}
		item, ok := value.(*cstates.StorageItem)
		if !ok {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[Enforce Param] storage error!")
		}
		// set prepare value to current value, make it effective
		native.CloneCache.Add(scommon.ST_STORAGE, getParamKey(contract, paramName, CURRENT_VALUE), value)
		paramValue := string(item.Value)
		setParamToCache(paramName, paramValue)
	}
	return nil
}

func initAdmin(native *NativeService, contract common.Address) {
	if admin.Address == *new(common.Address) {
		var err error
		// get admin from database
		admin, err = getStorageAdmin(native, getAdminKey(contract))
		// there are no admin in database
		if err != nil {
			admin.Address = ctypes.AddressFromPubKey(account.GetBookkeepers()[0])
		}
	}
}

func setParamToCache(key, value string) {
	paramCache.lock.Lock()
	defer paramCache.lock.Unlock()
	paramCache.Params[key] = value
}

func getParamFromCache(key string) string {
	paramCache.lock.RLock()
	defer paramCache.lock.RUnlock()
	return paramCache.Params[key]
}

func RegisterParamContract(native *NativeService) {
	native.Register("init", ParamInit)
	native.Register("transferAdmin", TransferAdmin)
	native.Register("approveAdmin", ApproveAdmin)
	native.Register("setParam", SetParam)
	native.Register("enforceParam", EnforceParam)
}

func GetGlobalPramValue(native *NativeService, paramName string) (string, error) {
	if value := getParamFromCache(paramName); value != "" {
		return value, nil
	}
	value, err := native.CloneCache.Get(scommon.ST_STORAGE,
		getParamKey(genesis.ParamContractAddress, paramName, CURRENT_VALUE))
	if err != nil {
		return "", errors.NewDetailErr(err, errors.ErrNoCode, "[Get Param] storage error!")
	}
	if value == nil {
		return "", nil
	}
	item, ok := value.(*cstates.StorageItem)
	if !ok {
		return "", errors.NewDetailErr(err, errors.ErrNoCode, "[Get Param] storage error!")
	}
	paramValue := string(item.Value)
	setParamToCache(paramName, paramValue)
	return paramValue, err
}
