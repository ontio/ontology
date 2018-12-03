package wasmvm

import (
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/vm/wasmvm/exec"
)

func (this *WasmVmService) contractCreate(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()

	contract, err := IsContractValid(vm)
	if err != nil {
		return false, err
	}
	contractAddress := contract.Address()

	dep, err := this.CacheDB.GetContract(contractAddress)
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[ContractCreate] GetOrAdd error!")
	}
	if dep == nil {
		this.CacheDB.PutContract(contract)
		dep = contract
	}

	idx, err := vm.SetPointerMemory(contractAddress)
	if err != nil {
		return false, err
	}

	vm.RestoreCtx()
	if vm.GetEnvCall().GetReturns() {
		vm.PushResult(uint64(idx))
	}
	return true, nil

}

func (this *WasmVmService) contractMigrate(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()

	contract, err := IsContractValid(vm)
	if err != nil {
		return false, err
	}
	newAddr := contract.Address()

	if err := isContractExist(this, newAddr); err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[ContractMigrate] contract invalid!")
	}

	context := this.ContextRef.CurrentContext()
	oldAddr := context.ContractAddress

	this.CacheDB.PutContract(contract)
	this.CacheDB.DeleteContract(oldAddr)

	iter := this.CacheDB.NewIterator(oldAddr[:])
	for has := iter.First(); has; has = iter.Next() {
		key := iter.Key()
		val := iter.Value()

		newKey, err := serializeStorageKey(newAddr, key[20:])
		if err != nil {
			return false, err
		}
		this.CacheDB.Put(newKey, val)
		this.CacheDB.Delete(key)
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return false, err
	}

	idx, err := vm.SetPointerMemory(newAddr)
	if err != nil {
		return false, err
	}

	vm.RestoreCtx()
	if vm.GetEnvCall().GetReturns() {
		vm.PushResult(uint64(idx))
	}
	return true, nil

}

func (this *WasmVmService) contractDelete(engine *exec.ExecutionEngine) (bool, error) {
	context := this.ContextRef.CurrentContext()
	if context == nil {
		return false, errors.NewErr("[ContractDestory] current contract context invalid!")
	}
	addr := context.ContractAddress
	contract, err := this.CacheDB.GetContract(addr)
	if err != nil || contract == nil {
		return false, errors.NewErr("[ContractDestory] get current contract fail!")
	}

	this.CacheDB.DeleteContract(addr)

	iter := this.CacheDB.NewIterator(addr[:])
	for has := iter.First(); has; has = iter.Next() {
		key := iter.Key()
		this.CacheDB.Delete(key)
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return false, err
	}
	return true, nil

}

func IsContractValid(vm *exec.VM) (*payload.DeployCode, error) {
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 7 {
		return nil, errors.NewErr("[contractMigrate] parameter count error")
	}

	code, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return nil, err
	}
	if len(code) > 1024*1024 {
		return nil, errors.NewErr("[Contract] Code too long!")
	}

	needStorage := true
	tmp := int(params[1])
	if tmp == 0 {
		needStorage = false
	}

	name, err := vm.GetPointerMemory(params[2])
	if err != nil {
		return nil, err
	}

	if len(name) > 252 {
		return nil, errors.NewErr("[Contract] Name too long!")
	}

	version, err := vm.GetPointerMemory(params[3])
	if err != nil {
		return nil, err
	}

	author, err := vm.GetPointerMemory(params[4])
	if err != nil {
		return nil, err
	}

	email, err := vm.GetPointerMemory(params[5])
	if err != nil {
		return nil, err
	}

	desc, err := vm.GetPointerMemory(params[5])
	if err != nil {
		return nil, err
	}

	if len(desc) > 65536 {
		return nil, errors.NewErr("[Contract] Desc too long!")
	}
	contract := &payload.DeployCode{
		Code:        code,
		NeedStorage: needStorage,
		Name:        string(name),
		Version:     string(version),
		Author:      string(author),
		Email:       string(email),
		Description: string(desc),
	}

	return contract, nil
}

func isContractExist(service *WasmVmService, addr common.Address) error {
	item, err := service.CacheDB.GetContract(addr)

	if err != nil || item != nil {
		return fmt.Errorf("[Contract] Get contract %x error or contract exist!", addr)
	}
	return nil
}
