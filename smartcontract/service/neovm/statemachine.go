package service

import (
	"bytes"
	"fmt"
	"github.com/Ontology/common"
	"github.com/Ontology/core/states"
	scommon "github.com/Ontology/core/store/common"
	"github.com/Ontology/core/store"
	"github.com/Ontology/errors"
	"github.com/Ontology/smartcontract/storage"
	stypes "github.com/Ontology/smartcontract/types"
	vm "github.com/Ontology/vm/neovm"
	"github.com/Ontology/core/payload"
	vmtypes "github.com/Ontology/vm/types"
)

type StateMachine struct {
	*StateReader
	ldgerStore store.ILedgerStore
	CloneCache *storage.CloneCache
	trigger    stypes.TriggerType
	time       uint32
}

func NewStateMachine(ldgerStore store.ILedgerStore, dbCache scommon.IStateStore, trigger stypes.TriggerType, time uint32) *StateMachine {
	var stateMachine StateMachine
	stateMachine.ldgerStore = ldgerStore
	stateMachine.CloneCache = storage.NewCloneCache(dbCache)
	stateMachine.StateReader = NewStateReader(ldgerStore,trigger)
	stateMachine.trigger = trigger
	stateMachine.time = time

	stateMachine.StateReader.Register("Neo.Runtime.GetTrigger", stateMachine.RuntimeGetTrigger)
	stateMachine.StateReader.Register("Neo.Runtime.GetTime", stateMachine.RuntimeGetTime)

	stateMachine.StateReader.Register("Neo.Contract.Create", stateMachine.ContractCreate)
	stateMachine.StateReader.Register("Neo.Contract.Migrate", stateMachine.ContractMigrate)
	stateMachine.StateReader.Register("Neo .Contract.GetStorageContext", stateMachine.GetStorageContext)
	stateMachine.StateReader.Register("Neo.Contract.GetScript", stateMachine.ContractGetCode)
	stateMachine.StateReader.Register("Neo.Contract.Destroy", stateMachine.ContractDestory)

	stateMachine.StateReader.Register("Neo.Storage.Get", stateMachine.StorageGet)
	stateMachine.StateReader.Register("Neo.Storage.Put", stateMachine.StoragePut)
	stateMachine.StateReader.Register("Neo.Storage.Delete", stateMachine.StorageDelete)
	return &stateMachine
}

func (s *StateMachine) RuntimeGetTrigger(engine *vm.ExecutionEngine) (bool, error) {
	vm.PushData(engine, int(s.trigger))
	return true, nil
}

func (s *StateMachine) RuntimeGetTime(engine *vm.ExecutionEngine) (bool, error) {
	vm.PushData(engine, s.time)
	return true, nil
}

func (s *StateMachine) ContractCreate(engine *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(engine) < 8 {
		return false, errors.NewErr("[ContractCreate] Too few input parameters")
	}
	codeByte := vm.PopByteArray(engine)
	if len(codeByte) > 1024*1024 {
		return false, errors.NewErr("[ContractCreate] Code too long!")
	}
	nameByte := vm.PopByteArray(engine)
	if len(nameByte) > 252 {
		return false, errors.NewErr("[ContractCreate] Name too long!")
	}
	versionByte := vm.PopByteArray(engine)
	if len(versionByte) > 252 {
		return false, errors.NewErr("[ContractCreate] Version too long!")
	}
	authorByte := vm.PopByteArray(engine)
	if len(authorByte) > 252 {
		return false, errors.NewErr("[ContractCreate] Author too long!")
	}
	emailByte := vm.PopByteArray(engine)
	if len(emailByte) > 252 {
		return false, errors.NewErr("[ContractCreate] Email too long!")
	}
	descByte := vm.PopByteArray(engine)
	if len(descByte) > 65536 {
		return false, errors.NewErr("[ContractCreate] Desc too long!")
	}
	contractState := &payload.DeployCode{
		Code:        codeByte,
		Name:        string(nameByte),
		Version:     string(versionByte),
		Author:      string(authorByte),
		Email:       string(emailByte),
		Description: string(descByte),
	}
	codeHash := common.ToCodeHash(codeByte)
	state, err := s.CloneCache.GetOrAdd(scommon.ST_Contract, codeHash.ToArray(), contractState)
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[ContractCreate] GetOrAdd error!")
	}
	vm.PushData(engine, state)
	return true, nil
}

func (s *StateMachine) ContractMigrate(engine *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(engine) < 8 {
		return false, errors.NewErr("[ContractMigrate] Too few input parameters ")
	}
	codeByte := vm.PopByteArray(engine)
	if len(codeByte) > 1024*1024 {
		return false, errors.NewErr("[ContractMigrate] Code too long!")
	}
	codeHash := common.ToCodeHash(codeByte)
	item, err := s.CloneCache.Get(scommon.ST_Contract, codeHash.ToArray())
	if err != nil {
		return false, errors.NewErr("[ContractMigrate] Get Contract error!")
	}
	if item != nil {
		return false, errors.NewErr("[ContractMigrate] Migrate Contract has exist!")
	}

	nameByte := vm.PopByteArray(engine)
	if len(nameByte) > 252 {
		return false, errors.NewErr("[ContractMigrate] Name too long!")
	}
	versionByte := vm.PopByteArray(engine)
	if len(versionByte) > 252 {
		return false, errors.NewErr("[ContractMigrate] Version too long!")
	}
	authorByte := vm.PopByteArray(engine)
	if len(authorByte) > 252 {
		return false, errors.NewErr("[ContractMigrate] Author too long!")
	}
	emailByte := vm.PopByteArray(engine)
	if len(emailByte) > 252 {
		return false, errors.NewErr("[ContractMigrate] Email too long!")
	}
	descByte := vm.PopByteArray(engine)
	if len(descByte) > 65536 {
		return false, errors.NewErr("[ContractMigrate] Desc too long!")
	}
	contractState := &payload.DeployCode{
		Code:        codeByte,
		Name:        string(nameByte),
		Version:     string(versionByte),
		Author:      string(authorByte),
		Email:       string(emailByte),
		Description: string(descByte),
	}
	s.CloneCache.Add(scommon.ST_Contract, codeHash.ToArray(), contractState)
	stateValues, err := s.CloneCache.Store.Find(scommon.ST_Contract, codeHash.ToArray())
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[ContractMigrate] Find error!")
	}
	for _, v := range stateValues {
		key := new(states.StorageKey)
		bf := bytes.NewBuffer([]byte(v.Key))
		if err := key.Deserialize(bf); err != nil {
			return false, errors.NewErr("[ContractMigrate] Key deserialize error!")
		}
		key = &states.StorageKey{CodeHash: codeHash, Key: key.Key}
		b := new(bytes.Buffer)
		if _, err := key.Serialize(b); err != nil {
			return false, errors.NewErr("[ContractMigrate] Key Serialize error!")
		}
		s.CloneCache.Add(scommon.ST_Storage, key.ToArray(), v.Value)
	}
	vm.PushData(engine, contractState)
	return s.ContractDestory(engine)
}

func (s *StateMachine) ContractDestory(engine *vm.ExecutionEngine) (bool, error) {
	context, err := engine.CurrentContext()
	if err != nil {
		return false, err
	}
	hash, err := context.GetCodeHash()
	if err != nil {
		return false, nil
	}
	item, err := s.CloneCache.Store.TryGet(scommon.ST_Contract, hash.ToArray())
	if err != nil {
		return false, err
	}
	if item == nil {
		return false, nil
	}
	s.CloneCache.Delete(scommon.ST_Contract, hash.ToArray())
	stateValues, err := s.CloneCache.Store.Find(scommon.ST_Contract, hash.ToArray())
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[ContractDestory] Find error!")
	}
	for _, v := range stateValues {
		s.CloneCache.Delete(scommon.ST_Storage, []byte(v.Key))
	}
	return true, nil
}

func (s *StateMachine) CheckStorageContext(context *StorageContext) (bool, error) {
	item, err := s.CloneCache.Get(scommon.ST_Contract, context.codeHash.ToArray())
	if err != nil {
		return false, err
	}
	if item == nil {
		return false, errors.NewErr(fmt.Sprintf("get contract by codehash=%v nil", context.codeHash))
	}
	return true, nil
}

func (s *StateMachine) StoragePut(engine *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(engine) < 3 {
		return false, errors.NewErr("[StoragePut] Too few input parameters ")
	}
	opInterface := vm.PopInteropInterface(engine)
	if opInterface == nil {
		return false, errors.NewErr("[StoragePut] Get StorageContext nil")
	}
	context := opInterface.(*StorageContext)
	key := vm.PopByteArray(engine)
	if len(key) > 1024 {
		return false, errors.NewErr("[StoragePut] Get Storage key to long")
	}
	value := vm.PopByteArray(engine)
	k, err := serializeStorageKey(context.codeHash, key)
	if err != nil {
		return false, err
	}
	s.CloneCache.Add(scommon.ST_Storage, k, &states.StorageItem{Value: value})
	return true, nil
}

func (s *StateMachine) StorageDelete(engine *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(engine) < 2 {
		return false, errors.NewErr("[StorageDelete] Too few input parameters ")
	}
	opInterface := vm.PopInteropInterface(engine)
	if opInterface == nil {
		return false, errors.NewErr("[StorageDelete] Get StorageContext nil")
	}
	context := opInterface.(*StorageContext)
	key := vm.PopByteArray(engine)
	k, err := serializeStorageKey(context.codeHash, key)
	if err != nil {
		return false, err
	}
	s.CloneCache.Delete(scommon.ST_Storage, k)
	return true, nil
}

func (s *StateMachine) StorageGet(engine *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(engine) < 2 {
		return false, errors.NewErr("[StorageGet] Too few input parameters ")
	}
	opInterface := vm.PopInteropInterface(engine)
	if opInterface == nil {
		return false, errors.NewErr("[StorageGet] Get StorageContext error!")
	}
	context := opInterface.(*StorageContext)
	if exist, err := s.CheckStorageContext(context); !exist {
		return false, err
	}
	key := vm.PopByteArray(engine)
	k, err := serializeStorageKey(context.codeHash, key)
	if err != nil {
		return false, err
	}
	item, err := s.CloneCache.Get(scommon.ST_Storage, k)
	if err != nil {
		return false, err
	}
	if item == nil {
		vm.PushData(engine, []byte{})
	} else {
		vm.PushData(engine, item.(*states.StorageItem).Value)
	}
	return true, nil
}

func (s *StateMachine) GetStorageContext(engine *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(engine) < 1 {
		return false, errors.NewErr("[GetStorageContext] Too few input parameters ")
	}
	opInterface := vm.PopInteropInterface(engine)
	if opInterface == nil {
		return false, errors.NewErr("[GetStorageContext] Get StorageContext nil")
	}
	contractState := opInterface.(*payload.DeployCode)
	code := &vmtypes.VmCode{
		VmType: contractState.VmType,
		Code: contractState.Code,
	}
	codeHash := code.AddressFromVmCode()
	item, err := s.CloneCache.Store.TryGet(scommon.ST_Contract, codeHash.ToArray())
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[GetStorageContext] Get StorageContext nil")
	}
	context, err := engine.CurrentContext()
	if err != nil {
		return false, err
	}
	if item == nil {
		return false, errors.NewErr(fmt.Sprintf("[GetStorageContext] Get contract by codehash:%v nil", codeHash))
	}
	currentHash, err := context.GetCodeHash()
	if err != nil {
		return false, err
	}
	if codeHash.CompareTo(currentHash) != 0 {
		return false, errors.NewErr("[GetStorageContext] CodeHash not equal!")
	}
	vm.PushData(engine, &StorageContext{codeHash: codeHash})
	return true, nil
}

func contains(programHashes []common.Uint160, programHash common.Uint160) bool {
	for _, v := range programHashes {
		if v.CompareTo(programHash) == 0 {
			return true
		}
	}
	return false
}

func serializeStorageKey(codeHash common.Uint160, key []byte) ([]byte, error) {
	bf := new(bytes.Buffer)
	storageKey := &states.StorageKey{CodeHash: codeHash, Key: key}
	if _, err := storageKey.Serialize(bf); err != nil {
		return []byte{}, errors.NewErr("[serializeStorageKey] StorageKey serialize error!")
	}
	return bf.Bytes(), nil
}
