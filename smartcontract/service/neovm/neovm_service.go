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

	"github.com/ontio/ontology/core/store"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/context"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/states"
	"github.com/ontio/ontology/smartcontract/storage"
	vm "github.com/ontio/ontology/vm/neovm"
)

const (
	MAX_STACK_SIZE = 2 * 1024
)

var (
	// Register all service for smart contract execute
	ServiceMap = map[string]Service{
		ATTRIBUTE_GETUSAGE_NAME:         {Execute: AttributeGetUsage, Validator: validatorAttribute},
		ATTRIBUTE_GETDATA_NAME:          {Execute: AttributeGetData, Validator: validatorAttribute},
		BLOCK_GETTRANSACTIONCOUNT_NAME:  {Execute: BlockGetTransactionCount, Validator: validatorBlock},
		BLOCK_GETTRANSACTIONS_NAME:      {Execute: BlockGetTransactions, Validator: validatorBlock},
		BLOCK_GETTRANSACTION_NAME:       {Execute: BlockGetTransaction, Validator: validatorBlockTransaction},
		BLOCKCHAIN_GETHEIGHT_NAME:       {Execute: BlockChainGetHeight},
		BLOCKCHAIN_GETHEADER_NAME:       {Execute: BlockChainGetHeader, Validator: validatorBlockChainHeader},
		BLOCKCHAIN_GETBLOCK_NAME:        {Execute: BlockChainGetBlock, Validator: validatorBlockChainBlock},
		BLOCKCHAIN_GETTRANSACTION_NAME:  {Execute: BlockChainGetTransaction, Validator: validatorBlockChainTransaction},
		BLOCKCHAIN_GETCONTRACT_NAME:     {Execute: BlockChainGetContract, Validator: validatorBlockChainContract},
		HEADER_GETINDEX_NAME:            {Execute: HeaderGetIndex, Validator: validatorHeader},
		HEADER_GETHASH_NAME:             {Execute: HeaderGetHash, Validator: validatorHeader},
		HEADER_GETVERSION_NAME:          {Execute: HeaderGetVersion, Validator: validatorHeader},
		HEADER_GETPREVHASH_NAME:         {Execute: HeaderGetPrevHash, Validator: validatorHeader},
		HEADER_GETTIMESTAMP_NAME:        {Execute: HeaderGetTimestamp, Validator: validatorHeader},
		HEADER_GETCONSENSUSDATA_NAME:    {Execute: HeaderGetConsensusData, Validator: validatorHeader},
		HEADER_GETNEXTCONSENSUS_NAME:    {Execute: HeaderGetNextConsensus, Validator: validatorHeader},
		HEADER_GETMERKLEROOT_NAME:       {Execute: HeaderGetMerkleRoot, Validator: validatorHeader},
		TRANSACTION_GETHASH_NAME:        {Execute: TransactionGetHash, Validator: validatorTransaction},
		TRANSACTION_GETTYPE_NAME:        {Execute: TransactionGetType, Validator: validatorTransaction},
		TRANSACTION_GETATTRIBUTES_NAME:  {Execute: TransactionGetAttributes, Validator: validatorTransaction},
		CONTRACT_CREATE_NAME:            {Execute: ContractCreate},
		CONTRACT_MIGRATE_NAME:           {Execute: ContractMigrate},
		CONTRACT_GETSTORAGECONTEXT_NAME: {Execute: ContractGetStorageContext},
		CONTRACT_DESTROY_NAME:           {Execute: ContractDestory},
		CONTRACT_GETSCRIPT_NAME:         {Execute: ContractGetCode, Validator: validatorGetCode},
		RUNTIME_GETTIME_NAME:            {Execute: RuntimeGetTime},
		RUNTIME_CHECKWITNESS_NAME:       {Execute: RuntimeCheckWitness, Validator: validatorCheckWitness},
		RUNTIME_NOTIFY_NAME:             {Execute: RuntimeNotify, Validator: validatorNotify},
		RUNTIME_LOG_NAME:                {Execute: RuntimeLog, Validator: validatorLog},
		RUNTIME_CHECKSIG_NAME:           {Execute: RuntimeCheckSig, Validator: validatorCheckSig},
		STORAGE_GET_NAME:                {Execute: StorageGet},
		STORAGE_PUT_NAME:                {Execute: StoragePut},
		STORAGE_DELETE_NAME:             {Execute: StorageDelete},
		STORAGE_GETCONTEXT_NAME:         {Execute: StorageGetContext},
		GETSCRIPTCONTAINER_NAME:         {Execute: GetCodeContainer},
		GETEXECUTINGSCRIPTHASH_NAME:     {Execute: GetExecutingAddress},
		GETCALLINGSCRIPTHASH_NAME:       {Execute: GetCallingAddress},
		GETENTRYSCRIPTHASH_NAME:         {Execute: GetEntryAddress},
	}
)

var (
	ERR_CHECK_STACK_SIZE = errors.NewErr("[NeoVmService] vm over max stack size!")
	ERR_EXECUTE_CODE     = errors.NewErr("[NeoVmService] vm execute code invalid!")
	ERR_GAS_INSUFFICIENT = errors.NewErr("[NeoVmService] gas insufficient")
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
}

// Invoke a smart contract
func (this *NeoVmService) Invoke() (interface{}, error) {
	engine := vm.NewExecutionEngine()
	if len(this.Code) == 0 {
		return nil, ERR_EXECUTE_CODE
	}
	engine.PushContext(vm.NewExecutionContext(engine, this.Code))
	for {
		if len(engine.Contexts) == 0 || engine.Context == nil {
			break
		}
		if engine.Context.GetInstructionPointer() >= len(engine.Context.Code) {
			break
		}
		if err := engine.ExecuteCode(); err != nil {
			return nil, err
		}
		if err := engine.ValidateOp(); err != nil {
			return nil, err
		}

		if engine.Context.GetInstructionPointer() < len(engine.Context.Code) {
			if ok := checkStackSize(engine); !ok {
				return nil, ERR_CHECK_STACK_SIZE
			}
		}
		if !this.ContextRef.CheckUseGas(GasPrice(engine, engine.OpExec.Name)) {
			return nil, ERR_GAS_INSUFFICIENT
		}
		switch engine.OpCode {
		case vm.SYSCALL:
			if err := this.SystemCall(engine); err != nil {
				return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[NeoVmService] service system call error!")
			}
		case vm.APPCALL, vm.TAILCALL:
			c := new(states.Contract)
			if err := c.Deserialize(engine.Context.OpReader.Reader()); err != nil {
				return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[NeoVmService] get contract parameters error!")
			}
			result, err := this.ContextRef.AppCall(c.Address, c.Method, c.Code, c.Args)
			if err != nil {
				return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[NeoVmService] service app call error!")
			}
			if result != nil {
				vm.PushData(engine, result)
			}
		default:
			if err := engine.StepInto(); err != nil {
				return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[NeoVmService] vm execute error!")
			}
		}
	}
	this.ContextRef.PushNotifications(this.Notifications)
	if engine.EvaluationStack.Count() != 0 {
		return engine.EvaluationStack.Peek(0), nil
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
	if this.ContextRef.CheckUseGas(GasPrice(engine, serviceName)) {
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
	if uint32(size) > MAX_STACK_SIZE {
		return false
	}
	return true
}
