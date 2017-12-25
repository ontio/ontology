package service

import (
	"bytes"
	"fmt"
	"github.com/Ontology/common"
	"github.com/Ontology/core/asset"
	"github.com/Ontology/core/code"
	"github.com/Ontology/core/contract"
	"github.com/Ontology/core/ledger"
	"github.com/Ontology/core/states"
	"github.com/Ontology/core/store"
	"github.com/Ontology/core/transaction"
	"github.com/Ontology/crypto"
	"github.com/Ontology/errors"
	. "github.com/Ontology/smartcontract/errors"
	"github.com/Ontology/smartcontract/storage"
	vm "github.com/Ontology/vm/neovm"
	"math"
)

type StateMachine struct {
	*StateReader
	CloneCache *storage.CloneCache
}

func NewStateMachine(dbCache store.IStateStore) *StateMachine {
	var stateMachine StateMachine
	stateMachine.CloneCache = storage.NewCloneCache(dbCache)
	stateMachine.StateReader = NewStateReader()

	stateMachine.StateReader.Register("Neo.Asset.Create", stateMachine.CreateAsset)
	stateMachine.StateReader.Register("Neo.Asset.Renew", stateMachine.AssetRenew)

	stateMachine.StateReader.Register("Neo.Contract.Create", stateMachine.ContractCreate)
	stateMachine.StateReader.Register("Neo.Contract.Migrate", stateMachine.ContractMigrate)
	stateMachine.StateReader.Register("Neo.Contract.GetStorageContext", stateMachine.GetStorageContext)
	stateMachine.StateReader.Register("Neo.Contract.GetScript", stateMachine.ContractGetCode)
	stateMachine.StateReader.Register("Neo.Contract.Destroy", stateMachine.ContractDestory)

	stateMachine.StateReader.Register("Neo.Storage.GetContext", stateMachine.StorageGetContext)
	stateMachine.StateReader.Register("Neo.Storage.Get", stateMachine.StorageGet)
	stateMachine.StateReader.Register("Neo.Storage.Put", stateMachine.StoragePut)
	stateMachine.StateReader.Register("Neo.Storage.Delete", stateMachine.StorageDelete)
	return &stateMachine
}

func (s *StateMachine) CreateAsset(engine *vm.ExecutionEngine) (bool, error) {
	tx := engine.GetCodeContainer().(*transaction.Transaction)
	assetId := tx.Hash()
	if vm.EvaluationStackCount(engine) < 7 {
		return false, errors.NewErr("[CreateAsset] Too few input parameters ")
	}
	assertType := asset.AssetType(vm.PopInt(engine))
	name := vm.PopByteArray(engine)
	if len(name) > 1024 {
		return false, ErrAssetNameInvalid
	}
	amount := vm.PopBigInt(engine)
	if amount.Int64() == 0 {
		return false, ErrAssetAmountInvalid
	}
	precision := vm.PopBigInt(engine)
	if precision.Int64() > 8 {
		return false, ErrAssetPrecisionInvalid
	}
	if amount.Int64() % int64(math.Pow(10, 8 - float64(precision.Int64()))) != 0 {
		return false, ErrAssetAmountInvalid
	}
	ownerByte := vm.PopByteArray(engine)
	owner, err := crypto.DecodePoint(ownerByte)
	if err != nil {
		return false, err
	}
	if result, err := s.StateReader.CheckWitnessPublicKey(engine, owner); !result {
		return result, err
	}
	adminByte := vm.PopByteArray(engine)
	admin, err := common.Uint160ParseFromBytes(adminByte)
	if err != nil {
		return false, err
	}

	assetState := &states.AssetState{
		AssetId:    assetId,
		AssetType:  asset.AssetType(assertType),
		Name:       string(name),
		Amount:     common.Fixed64(amount.Int64()),
		Precision:  byte(precision.Int64()),
		Admin:      admin,
		Owner:      owner,
		Expiration: ledger.DefaultLedger.Store.GetHeight() + 1 + 2000000,
		IsFrozen:   false,
	}
	state, err := s.CloneCache.GetOrAdd(store.ST_Asset, assetId.ToArray(), assetState); if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[CreateAsset] GetOrAdd error!")
	}
	vm.PushData(engine, state)
	return true, nil
}

func (s *StateMachine) ContractCreate(engine *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(engine) < 8 {
		return false, errors.NewErr("[ContractCreate] Too few input parameters ")
	}
	codeByte := vm.PopByteArray(engine)
	if len(codeByte) > 1024 * 1024 {
		return false, errors.NewErr("[ContractCreate] Code too long!")
	}
	parameters := vm.PopByteArray(engine)
	if len(parameters) > 252 {
		return false, errors.NewErr("[ContractCreate] Parameters too long!")
	}
	parameterList := make([]contract.ContractParameterType, 0)
	for _, v := range parameters {
		parameterList = append(parameterList, contract.ContractParameterType(v))
	}
	returnType := vm.PopInt(engine)
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
	funcCode := &code.FunctionCode{
		Code:           codeByte,
		ParameterTypes: parameterList,
		ReturnType:     contract.ContractParameterType(returnType),
	}
	contractState := &states.ContractState{
		Code:        funcCode,
		Name:        string(nameByte),
		Version:     string(versionByte),
		Author:      string(authorByte),
		Email:       string(emailByte),
		Description: string(descByte),
	}
	codeHash, err := common.Uint160ParseFromBytes(codeByte)
	if err != nil {
		return false, err
	}
	state, err := s.CloneCache.GetOrAdd(store.ST_Contract, codeHash.ToArray(), contractState); if err != nil {
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
	if len(codeByte) > 1024 * 1024 {
		return false, errors.NewErr("[ContractMigrate] Code too long!")
	}
	codeHash, err := common.ToCodeHash(codeByte)
	if err != nil {
		return false, err
	}
	item, err := s.CloneCache.Get(store.ST_Contract, codeHash.ToArray())
	if err != nil {
		return false, errors.NewErr("[ContractMigrate] Get Contract error!")
	}
	if item != nil {
		return false, errors.NewErr("[ContractMigrate] Migrate Contract has exist!")
	}

	parameters := vm.PopByteArray(engine)
	if len(parameters) > 252 {
		return false, errors.NewErr("[ContractMigrate] Parameters too long!")
	}
	parameterList := make([]contract.ContractParameterType, 0)
	for _, v := range parameters {
		parameterList = append(parameterList, contract.ContractParameterType(v))
	}
	returnType := vm.PopInt(engine)
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
	if len(emailByte) > 65536 {
		return false, errors.NewErr("[ContractMigrate] Desc too long!")
	}
	funcCode := &code.FunctionCode{
		Code:           codeByte,
		ParameterTypes: parameterList,
		ReturnType:     contract.ContractParameterType(returnType),
	}
	contractState := &states.ContractState{
		Code:        funcCode,
		Name:        string(nameByte),
		Version:     string(versionByte),
		Author:      string(authorByte),
		Email:       string(emailByte),
		Description: string(descByte),
	}
	s.CloneCache.Add(store.ST_Contract, codeHash.ToArray(), contractState)
	stateValues, err := s.CloneCache.Store.Find(store.ST_Contract, codeHash.ToArray())
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
		s.CloneCache.Add(store.ST_Storage, key.ToArray(), v.Value)
	}
	vm.PushData(engine, contractState)
	return s.ContractDestory(engine)
}

func (s *StateMachine) AssetRenew(engine *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(engine) < 2 {
		return false, errors.NewErr("[AssetRenew] Too few input parameters ")
	}
	data := vm.PopInteropInterface(engine)
	if data == nil {
		return false, errors.NewErr("[AssetRenew] Get Asset nil!")
	}
	years := vm.PopInt(engine)
	assetState := data.(*states.AssetState)
	if assetState.Expiration < ledger.DefaultLedger.Store.GetHeight() + 1 {
		assetState.Expiration = ledger.DefaultLedger.Store.GetHeight() + 1
	}
	assetState.Expiration += uint32(years) * 2000000
	vm.PushData(engine, assetState.Expiration)
	return true, nil
}

func (s *StateMachine) ContractDestory(engine *vm.ExecutionEngine) (bool, error) {
	data := engine.CurrentContext().CodeHash
	if data != nil {
		return false, nil
	}
	hash, err := common.Uint160ParseFromBytes(data)
	if err != nil {
		return false, err
	}
	item, err := s.CloneCache.Store.TryGet(store.ST_Contract, hash.ToArray())
	if err != nil {
		return false, err
	}
	if item == nil {
		return false, nil
	}
	s.CloneCache.Delete(store.ST_Contract, hash.ToArray())
	stateValues, err := s.CloneCache.Store.Find(store.ST_Contract, hash.ToArray())
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[ContractDestory] Find error!")
	}
	for _, v := range stateValues {
		s.CloneCache.Delete(store.ST_Storage, []byte(v.Key))
	}
	return true, nil
}

func (s *StateMachine) CheckStorageContext(context *StorageContext) (bool, error) {
	item, err := s.CloneCache.Get(store.ST_Contract, context.codeHash.ToArray())
	if err != nil {
		return false, err
	}
	if item == nil {
		return false, errors.NewErr(fmt.Sprintf("get contract by codehash=%v nil", context.codeHash))
	}
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
	item, err := s.CloneCache.Get(store.ST_Storage, k)
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
	s.CloneCache.Add(store.ST_Storage, k, &states.StorageItem{Value: value})
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
	s.CloneCache.Delete(store.ST_Storage, k)
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
	contractState := opInterface.(*states.ContractState)
	codeHash := contractState.Code.CodeHash()
	item, err := s.CloneCache.Store.TryGet(store.ST_Contract, codeHash.ToArray())
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[GetStorageContext] Get StorageContext nil")
	}
	if item == nil {
		return false, errors.NewErr(fmt.Sprintf("[GetStorageContext] Get contract by codehash:%v nil", codeHash))
	}
	currentHash, err := common.Uint160ParseFromBytes(engine.CurrentContext().GetCodeHash())
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[GetStorageContext] Get CurrentHash error")
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
