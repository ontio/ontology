package neovm

import (
	"bytes"

	vm "github.com/ontio/ontology/vm/neovm"
	"github.com/ontio/ontology/errors"
	stypes "github.com/ontio/ontology/smartcontract/types"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/common"
)

// create a new contract
func ContractCreate(service *NeoVmService, engine *vm.ExecutionEngine) error {
	contract, err := isContractParamValid(engine); if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[ContractCreate] contract parameters invalid!")
	}
	contractAddress := contract.Code.AddressFromVmCode()
	state, err := service.CloneCache.GetOrAdd(scommon.ST_CONTRACT, contractAddress[:], contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[ContractCreate] GetOrAdd error!")
	}
	vm.PushData(engine, state)
	return nil
}

// migrate older contract to new contract
func ContractMigrate(service *NeoVmService, engine *vm.ExecutionEngine) error {
	contract, err := isContractParamValid(engine); if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[ContractMigrate] contract parameters invalid!")
	}
	contractAddress := contract.Code.AddressFromVmCode()

	if err := isContractExist(service, contractAddress); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[ContractMigrate] contract invalid!")
	}

	service.CloneCache.Add(scommon.ST_CONTRACT, contractAddress[:], contract)
	if err := storeMigration(service, contractAddress); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[ContractMigrate] contract store migration error!")
	}
	vm.PushData(engine, contract)
	return ContractDestory(service, engine)
}

// destory a contract
func ContractDestory(service *NeoVmService, engine *vm.ExecutionEngine) error {
	context := service.ContextRef.CurrentContext(); if context == nil {
		return errors.NewErr("[ContractDestory] current contract context invalid!")
	}
	item, err := service.CloneCache.Store.TryGet(scommon.ST_CONTRACT, context.ContractAddress[:])

	if err != nil || item == nil {
		return errors.NewErr("[ContractDestory] get current contract fail!")
	}

	service.CloneCache.Delete(scommon.ST_CONTRACT, context.ContractAddress[:])
	stateValues, err := service.CloneCache.Store.Find(scommon.ST_CONTRACT, context.ContractAddress[:]); if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[ContractDestory] find error!")
	}
	for _, v := range stateValues {
		service.CloneCache.Delete(scommon.ST_STORAGE, []byte(v.Key))
	}
	return nil
}

// get contract storage context
func ContractGetStorageContext(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[GetStorageContext] Too few input parameter!")
	}
	opInterface := vm.PopInteropInterface(engine); if opInterface == nil {
		return errors.NewErr("[GetStorageContext] Pop data nil!")
	}
	contractState, ok := opInterface.(*payload.DeployCode); if !ok {
		return errors.NewErr("[GetStorageContext] Pop data not contract!")
	}
	address := contractState.Code.AddressFromVmCode()
	item, err := service.CloneCache.Store.TryGet(scommon.ST_CONTRACT, address[:])
	if err != nil || item == nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[GetStorageContext] Get StorageContext nil")
	}
	if address != service.ContextRef.CurrentContext().ContractAddress {
		return errors.NewErr("[GetStorageContext] CodeHash not equal!")
	}
	vm.PushData(engine, &StorageContext{address: address})
	return nil
}

// get contract code
func ContractGetCode(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, vm.PopInteropInterface(engine).(*payload.DeployCode).Code)
	return nil
}

func isContractParamValid(engine *vm.ExecutionEngine) (*payload.DeployCode, error) {
	if vm.EvaluationStackCount(engine) < 7 {
		return nil, errors.NewErr("[Contract] Too few input parameters")
	}
	code := vm.PopByteArray(engine); if len(code) > 1024 * 1024 {
		return nil, errors.NewErr("[Contract] Code too long!")
	}
	needStorage := vm.PopBoolean(engine)
	name := vm.PopByteArray(engine); if len(name) > 252 {
		return nil, errors.NewErr("[Contract] Name too long!")
	}
	version := vm.PopByteArray(engine); if len(version) > 252 {
		return nil, errors.NewErr("[Contract] Version too long!")
	}
	author := vm.PopByteArray(engine); if len(author) > 252 {
		return nil, errors.NewErr("[Contract] Author too long!")
	}
	email := vm.PopByteArray(engine); if len(email) > 252 {
		return nil, errors.NewErr("[Contract] Email too long!")
	}
	desc := vm.PopByteArray(engine); if len(desc) > 65536 {
		return nil, errors.NewErr("[Contract] Desc too long!")
	}
	contract := &payload.DeployCode{
		Code:        stypes.VmCode{VmType:stypes.NEOVM, Code: code},
		NeedStorage: needStorage,
		Name:        string(name),
		Version:     string(version),
		Author:      string(author),
		Email:       string(email),
		Description: string(desc),
	}
	return contract, nil
}

func isContractExist(service *NeoVmService, contractAddress common.Address) error {
	item, err := service.CloneCache.Get(scommon.ST_CONTRACT, contractAddress[:])

	if err != nil || item != nil {
		return errors.NewErr("[Contract] Get contract error or contract exist!")
	}
	return nil
}

func storeMigration(service *NeoVmService, contractAddress common.Address) error {
	stateValues, err := service.CloneCache.Store.Find(scommon.ST_CONTRACT, contractAddress[:]); if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Contract] Find error!")
	}
	for _, v := range stateValues {
		key := new(states.StorageKey)
		bf := bytes.NewBuffer([]byte(v.Key))
		if err := key.Deserialize(bf); err != nil {
			return errors.NewErr("[Contract] Key deserialize error!")
		}
		key = &states.StorageKey{CodeHash: contractAddress, Key: key.Key}
		b := new(bytes.Buffer)
		if _, err := key.Serialize(b); err != nil {
			return errors.NewErr("[Contract] Key Serialize error!")
		}
		service.CloneCache.Add(scommon.ST_STORAGE, key.ToArray(), v.Value)
	}
	return nil
}


