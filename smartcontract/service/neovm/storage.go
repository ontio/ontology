package neovm

import (
	"bytes"

	vm "github.com/ontio/ontology/vm/neovm"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/common"
)

// put value to contract storage
func StoragePut(service *NeoVmService, engine *vm.ExecutionEngine) error {
	context, err := getContext(engine); if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[StoragePut] get pop context error!")
	}
	if err := checkStorageContext(service, context); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[StoragePut] check context error!")
	}

	key := vm.PopByteArray(engine)
	if len(key) > 1024 {
		return errors.NewErr("[StoragePut] Storage key to long")
	}

	value := vm.PopByteArray(engine)
	service.CloneCache.Add(scommon.ST_STORAGE, getStorageKey(context.address, key), &states.StorageItem{Value: value})
	return nil
}

// delete value in contract storage
func StorageDelete(service *NeoVmService, engine *vm.ExecutionEngine) error {
	context, err := getContext(engine); if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[StorageDelete] get pop context error!")
	}
	if err := checkStorageContext(service, context); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[StorageDelete] check context error!")
	}

	service.CloneCache.Delete(scommon.ST_STORAGE, getStorageKey(context.address, vm.PopByteArray(engine)))

	return nil
}

// get value from contract storage
func StorageGet(service *NeoVmService, engine *vm.ExecutionEngine) error {
	context, err := getContext(engine); if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[StorageGet] get pop context error!")
	}

	item, err := service.CloneCache.Get(scommon.ST_STORAGE, getStorageKey(context.address, vm.PopByteArray(engine))); if err != nil {
		return err
	}

	if item == nil {
		vm.PushData(engine, []byte{})
	} else {
		vm.PushData(engine, item.(*states.StorageItem).Value)
	}
	return nil
}

// put current contract context to stack
func StorageGetContext(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, NewStorageContext(service.ContextRef.CurrentContext().ContractAddress))
	return nil
}

func checkStorageContext(service *NeoVmService, context *StorageContext) error {
	item, err := service.CloneCache.Get(scommon.ST_CONTRACT, context.address[:])
	if err != nil || item == nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[CheckStorageContext] get context fail!")
	}
	return nil
}

func getContext(engine *vm.ExecutionEngine) (*StorageContext, error) {
	if vm.EvaluationStackCount(engine) < 2 {
		return nil, errors.NewErr("[Context] Too few input parameters ")
	}
	opInterface := vm.PopInteropInterface(engine); if opInterface == nil {
		return nil, errors.NewErr("[Context] Get storageContext nil")
	}
	context, ok := opInterface.(*StorageContext); if !ok {
		return nil, errors.NewErr("[Context] Get storageContext invalid")
	}
	return context, nil
}

func getStorageKey(codeHash common.Address, key []byte) []byte {
	buf := bytes.NewBuffer(nil)
	buf.Write(codeHash[:])
	buf.Write(key)
	return buf.Bytes()
}

