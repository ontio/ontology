package native

import (
	"bytes"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/genesis"
	cstates "github.com/ontio/ontology/core/states"
	scommon "github.com/ontio/ontology/core/store/common"
	ctypes "github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native/states"
	"sync"
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

var paramCache *ParamCache
var admin common.Address

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
		native.CloneCache.Add(scommon.ST_STORAGE, getParamKey(contract, k), getParamStorageItem(v))
		paramCache.Params[k] = v
	}
	admin = ctypes.AddressFromPubKey(account.GetBookkeepers()[0])
	native.CloneCache.Add(scommon.ST_STORAGE, getAdminKey(contract), getAdminStorageItem(admin))
	return nil
}

func TransferAdmin(native *NativeService) error {
	destinationAdmin := new(common.Address)
	if err := destinationAdmin.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return errors.NewErr("[Transfer Admin]Deserialize Admins failed!")
	}
	if !native.ContextRef.CheckWitness(*destinationAdmin) {
		return errors.NewErr("[Transfer Admin]Authentication failed!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	transferAdmin, err := getTransferAdmin(native, getTransferAdminKey(contract, *destinationAdmin))
	if err != nil || transferAdmin != *destinationAdmin {
		return errors.NewDetailErr(err, errors.ErrNoCode,
			"[Transfer Admin] Destination account hasn't been approved!")
	}
	// delete transfer admin item
	native.CloneCache.Delete(scommon.ST_STORAGE, getTransferAdminKey(contract, *destinationAdmin))
	// modify admin in database
	native.CloneCache.Add(scommon.ST_STORAGE, getAdminKey(contract), getAdminStorageItem(*destinationAdmin))

	admin = *destinationAdmin
	return nil
}

func ApproveAdmin(native *NativeService) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	if admin == *new(common.Address) {
		var err error
		admin, err = getAdmin(native, getAdminKey(contract)) // get admin from database
		if err != nil { // there are no admin in database
			admin = ctypes.AddressFromPubKey(account.GetBookkeepers()[0])
		}
	}
	if !native.ContextRef.CheckWitness(admin) {
		return errors.NewErr("[Approve Admin]Authentication failed!")
	}
	destinationAdmin := new(common.Address)
	if err := destinationAdmin.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return errors.NewErr("[Approve Admin]Deserialize Admins failed!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, getTransferAdminKey(contract, *destinationAdmin),
		getAdminStorageItem(*destinationAdmin))
	return nil
}

func SetParam(native *NativeService) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	if admin == *new(common.Address) {
		var err error
		admin, err = getAdmin(native, getAdminKey(contract))
		if err != nil { // there are no admin in database
			admin = ctypes.AddressFromPubKey(account.GetBookkeepers()[0])
		}
	}
	if !native.ContextRef.CheckWitness(admin) {
		return errors.NewErr("[Set Param]Authentication failed!")
	}
	params := new(states.Params)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return errors.NewErr("[Set Param]Deserialize failed!")
	}
	for _, param := range params.ParamList {
		deleteParamInCache(param.K) // delete the param in cache, in case of data inconsistencies
		native.CloneCache.Add(scommon.ST_STORAGE, getParamKey(contract, param.K), getParamStorageItem(param.V))
		notifyParamSetSucess(native, contract, param)
	}
	return nil
}

func deleteParamInCache(key string) {
	paramCache.lock.Lock()
	defer paramCache.lock.Unlock()
	delete(paramCache.Params, key)
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
}

func GetGlobalPramValue(native *NativeService, paramName string) (string, error) {
	if value := getParamFromCache(paramName); value != "" {
		return value, nil
	}
	value, err := native.CloneCache.Get(scommon.ST_STORAGE, getParamKey(genesis.ParamContractAddress, paramName))
	if err != nil {
		return "", errors.NewDetailErr(err, errors.ErrNoCode, "[Get Param] storage error!")
	}
	if value == nil {
		return "", errors.NewErr("[Get Param] param doesn't exist!")
	}
	item, ok := value.(*cstates.StorageItem)
	if !ok {
		return "", errors.NewDetailErr(err, errors.ErrNoCode, "[Get Param] storage error!")
	}
	paramValue := string(item.Value)
	setParamToCache(paramName, paramValue)
	return paramValue, err
}
