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

package neovm

import (
	"fmt"

	"github.com/ontio/ontology-crypto/keypair"
	scommon "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/core/store"
	"github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/context"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/storage"
	vm "github.com/ontio/ontology/vm/neovm"
)

var (
	// Register all service for smart contract execute
	ServiceMap = map[string]Service{
		ATTRIBUTE_GETUSAGE_NAME:              {Execute: AttributeGetUsage, Validator: validatorAttribute},
		ATTRIBUTE_GETDATA_NAME:               {Execute: AttributeGetData, Validator: validatorAttribute},
		BLOCK_GETTRANSACTIONCOUNT_NAME:       {Execute: BlockGetTransactionCount, Validator: validatorBlock},
		BLOCK_GETTRANSACTIONS_NAME:           {Execute: BlockGetTransactions, Validator: validatorBlock},
		BLOCK_GETTRANSACTION_NAME:            {Execute: BlockGetTransaction, Validator: validatorBlockTransaction},
		BLOCKCHAIN_GETHEIGHT_NAME:            {Execute: BlockChainGetHeight},
		BLOCKCHAIN_GETHEADER_NAME:            {Execute: BlockChainGetHeader, Validator: validatorBlockChainHeader},
		BLOCKCHAIN_GETBLOCK_NAME:             {Execute: BlockChainGetBlock, Validator: validatorBlockChainBlock},
		BLOCKCHAIN_GETTRANSACTION_NAME:       {Execute: BlockChainGetTransaction, Validator: validatorBlockChainTransaction},
		BLOCKCHAIN_GETCONTRACT_NAME:          {Execute: BlockChainGetContract, Validator: validatorBlockChainContract},
		BLOCKCHAIN_GETTRANSACTIONHEIGHT_NAME: {Execute: BlockChainGetTransactionHeight},
		HEADER_GETINDEX_NAME:                 {Execute: HeaderGetIndex, Validator: validatorHeader},
		HEADER_GETHASH_NAME:                  {Execute: HeaderGetHash, Validator: validatorHeader},
		HEADER_GETVERSION_NAME:               {Execute: HeaderGetVersion, Validator: validatorHeader},
		HEADER_GETPREVHASH_NAME:              {Execute: HeaderGetPrevHash, Validator: validatorHeader},
		HEADER_GETTIMESTAMP_NAME:             {Execute: HeaderGetTimestamp, Validator: validatorHeader},
		HEADER_GETCONSENSUSDATA_NAME:         {Execute: HeaderGetConsensusData, Validator: validatorHeader},
		HEADER_GETNEXTCONSENSUS_NAME:         {Execute: HeaderGetNextConsensus, Validator: validatorHeader},
		HEADER_GETMERKLEROOT_NAME:            {Execute: HeaderGetMerkleRoot, Validator: validatorHeader},
		TRANSACTION_GETHASH_NAME:             {Execute: TransactionGetHash, Validator: validatorTransaction},
		TRANSACTION_GETTYPE_NAME:             {Execute: TransactionGetType, Validator: validatorTransaction},
		TRANSACTION_GETATTRIBUTES_NAME:       {Execute: TransactionGetAttributes, Validator: validatorTransaction},
		CONTRACT_CREATE_NAME:                 {Execute: ContractCreate},
		CONTRACT_MIGRATE_NAME:                {Execute: ContractMigrate},
		CONTRACT_GETSTORAGECONTEXT_NAME:      {Execute: ContractGetStorageContext},
		CONTRACT_DESTROY_NAME:                {Execute: ContractDestory},
		CONTRACT_GETSCRIPT_NAME:              {Execute: ContractGetCode, Validator: validatorGetCode},
		RUNTIME_GETTIME_NAME:                 {Execute: RuntimeGetTime},
		RUNTIME_CHECKWITNESS_NAME:            {Execute: RuntimeCheckWitness, Validator: validatorCheckWitness},
		RUNTIME_NOTIFY_NAME:                  {Execute: RuntimeNotify, Validator: validatorNotify},
		RUNTIME_LOG_NAME:                     {Execute: RuntimeLog, Validator: validatorLog},
		RUNTIME_GETTRIGGER_NAME:              {Execute: RuntimeGetTrigger},
		RUNTIME_SERIALIZE_NAME:               {Execute: RuntimeSerialize, Validator: validatorSerialize},
		RUNTIME_DESERIALIZE_NAME:             {Execute: RuntimeDeserialize, Validator: validatorDeserialize},
		NATIVE_INVOKE_NAME:                   {Execute: NativeInvoke},
		STORAGE_GET_NAME:                     {Execute: StorageGet},
		STORAGE_PUT_NAME:                     {Execute: StoragePut},
		STORAGE_DELETE_NAME:                  {Execute: StorageDelete},
		STORAGE_GETCONTEXT_NAME:              {Execute: StorageGetContext},
		STORAGE_GETREADONLYCONTEXT_NAME:      {Execute: StorageGetReadOnlyContext},
		STORAGECONTEXT_ASREADONLY_NAME:       {Execute: StorageContextAsReadOnly},
		GETSCRIPTCONTAINER_NAME:              {Execute: GetCodeContainer},
		GETEXECUTINGSCRIPTHASH_NAME:          {Execute: GetExecutingAddress},
		GETCALLINGSCRIPTHASH_NAME:            {Execute: GetCallingAddress},
		GETENTRYSCRIPTHASH_NAME:              {Execute: GetEntryAddress},
	}
)

var (
	ERR_CHECK_STACK_SIZE  = errors.NewErr("[NeoVmService] vm over max stack size!")
	ERR_EXECUTE_CODE      = errors.NewErr("[NeoVmService] vm execute code invalid!")
	ERR_GAS_INSUFFICIENT  = errors.NewErr("[NeoVmService] gas insufficient")
	CONTRACT_NOT_EXIST    = errors.NewErr("[NeoVmService] Get contract code from db fail")
	DEPLOYCODE_TYPE_ERROR = errors.NewErr("[NeoVmService] DeployCode type error!")
)

type (
	Execute   func(service *NeoVmService, engine *vm.ExecutionEngine) error
	Validator func(engine *vm.ExecutionEngine) error
)

type Service struct {
	Execute   Execute
	Validator Validator
}

// NeoVmService is a struct for smart contract provide interop service
type NeoVmService struct {
	Store         store.LedgerStore
	CloneCache    *storage.CloneCache
	ContextRef    context.ContextRef
	Notifications []*event.NotifyEventInfo
	Code          []byte
	Tx            *types.Transaction
	Time          uint32
	Height        uint32
	Engine        *vm.ExecutionEngine
}

// Invoke a smart contract
func (this *NeoVmService) Invoke() (interface{}, error) {
	if len(this.Code) == 0 {
		return nil, ERR_EXECUTE_CODE
	}
	this.ContextRef.PushContext(&context.Context{ContractAddress: types.AddressFromVmCode(this.Code), Code: this.Code})
	this.Engine.PushContext(vm.NewExecutionContext(this.Engine, this.Code))
	for {
		if len(this.Engine.Contexts) == 0 || this.Engine.Context == nil {
			break
		}
		if this.Engine.Context.GetInstructionPointer() >= len(this.Engine.Context.Code) {
			break
		}
		if err := this.Engine.ExecuteCode(); err != nil {
			return nil, err
		}
		if this.Engine.Context.GetInstructionPointer() < len(this.Engine.Context.Code) {
			if ok := checkStackSize(this.Engine); !ok {
				return nil, ERR_CHECK_STACK_SIZE
			}
		}
		if this.Engine.OpCode >= vm.PUSHBYTES1 && this.Engine.OpCode <= vm.PUSHBYTES75 {
			if !this.ContextRef.CheckUseGas(OPCODE_GAS) {
				return nil, ERR_GAS_INSUFFICIENT
			}
		} else {
			if err := this.Engine.ValidateOp(); err != nil {
				return nil, err
			}
			if !this.ContextRef.CheckUseGas(GasPrice(this.Engine, this.Engine.OpExec.Name)) {
				return nil, ERR_GAS_INSUFFICIENT
			}
		}
		switch this.Engine.OpCode {
		case vm.VERIFY:
			pubKey := vm.PopByteArray(this.Engine)
			key, err := keypair.DeserializePublicKey(pubKey)
			if err != nil {
				return nil, err
			}
			sig := vm.PopByteArray(this.Engine)
			data := vm.PopByteArray(this.Engine)
			if err := signature.Verify(key, data, sig); err != nil {
				vm.PushData(this.Engine, false)
			} else {
				vm.PushData(this.Engine, true)
			}
		case vm.SYSCALL:
			if err := this.SystemCall(this.Engine); err != nil {
				return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[NeoVmService] service system call error!")
			}
		case vm.APPCALL, vm.TAILCALL:
			address := this.Engine.Context.OpReader.ReadBytes(20)
			code, err := this.getContract(address)
			if err != nil {
				return nil, err
			}
			service, err := this.ContextRef.NewExecuteEngine(code)
			if err != nil {
				return nil, err
			}
			this.Engine.EvaluationStack.CopyTo(service.(*NeoVmService).Engine.EvaluationStack)
			result, err := service.Invoke()
			if err != nil {
				return nil, err
			}
			if result != nil {
				vm.PushData(this.Engine, result)
			}
		default:
			if err := this.Engine.StepInto(); err != nil {
				return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[NeoVmService] vm execute error!")
			}
		}
	}
	this.ContextRef.PopContext()
	this.ContextRef.PushNotifications(this.Notifications)
	if this.Engine.EvaluationStack.Count() != 0 {
		return this.Engine.EvaluationStack.Peek(0), nil
	}
	return nil, nil
}

// SystemCall provide register service for smart contract to interaction with blockchain
func (this *NeoVmService) SystemCall(engine *vm.ExecutionEngine) error {
	serviceName := engine.Context.OpReader.ReadVarString()
	service, ok := ServiceMap[serviceName]
	if !ok {
		return errors.NewErr(fmt.Sprintf("[SystemCall] service not support: %s", serviceName))
	}
	if !this.ContextRef.CheckUseGas(GasPrice(engine, serviceName)) {
		return ERR_GAS_INSUFFICIENT
	}
	if service.Validator != nil {
		if err := service.Validator(engine); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[SystemCall] service validator error!")
		}
	}

	if err := service.Execute(this, engine); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[SystemCall] service execute error!")
	}
	return nil
}

func (this *NeoVmService) getContract(address []byte) ([]byte, error) {
	item, err := this.CloneCache.Store.TryGet(common.ST_CONTRACT, address)
	if err != nil {
		return nil, errors.NewErr("[getContract] Get contract context error!")
	}
	log.Infof("invoke contract address:%x", scommon.ToArrayReverse(address))
	if item == nil {
		return nil, CONTRACT_NOT_EXIST
	}
	contract, ok := item.Value.(*payload.DeployCode)
	if !ok {
		return nil, DEPLOYCODE_TYPE_ERROR
	}
	return contract.Code, nil
}

func checkStackSize(engine *vm.ExecutionEngine) bool {
	size := 0
	if engine.OpCode < vm.PUSH16 {
		size = 1
	} else {
		switch engine.OpCode {
		case vm.DEPTH, vm.DUP, vm.OVER, vm.TUCK:
			size = 1
		case vm.UNPACK:
			if engine.EvaluationStack.Count() == 0 {
				return false
			}
			size = len(vm.PeekStackItem(engine).GetArray())
		}
	}
	size += engine.EvaluationStack.Count() + engine.AltStack.Count()
	if size > MAX_STACK_SIZE {
		return false
	}
	return true
}
