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
	"math/big"

	scommon "github.com/ontio/ontology/core/store/common"
	vmtype "github.com/ontio/ontology/vm/neovm/types"
	"github.com/ontio/ontology/core/store"
	"github.com/ontio/ontology/smartcontract/storage"
	vm "github.com/ontio/ontology/vm/neovm"
	"github.com/ontio/ontology/smartcontract/context"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/states"
)

const (
	MAX_STACK_SIZE = 2 * 1024
	MAX_ARRAY_SIZE = 1024
	MAX_SIZE_FOR_BIGINTEGER = 32
)

var (
	// register all service for smart contract execute
	ServiceMap = map[string]Service{
		"Neo.Attribute.GetUsage": {Execute: AttributeGetUsage, Validator: validatorAttribute},
		"Neo.Attribute.GetData": {Execute: AttributeGetData, Validator: validatorAttribute},
		"Neo.Block.GetTransactionCount": {Execute: BlockGetTransactionCount, Validator: validatorBlock},
		"Neo.Block.GetTransactions": {Execute: BlockGetTransactions, Validator: validatorBlock},
		"Neo.Block.GetTransaction": {Execute: BlockGetTransaction, Validator: validatorBlockTransaction},
		"Neo.Blockchain.GetHeight": {Execute: BlockChainGetHeight},
		"Neo.Blockchain.GetHeader": {Execute: BlockChainGetHeader, Validator: validatorBlockChainHeader},
		"Neo.Blockchain.GetBlock": {Execute: BlockChainGetBlock, Validator: validatorBlockChainBlock},
		"Neo.Blockchain.GetTransaction": {Execute: BlockChainGetTransaction, Validator: validatorBlockChainTransaction},
		"Neo.Blockchain.GetContract": {Execute: BlockChainGetContract, Validator: validatorBlockChainContract},
		"Neo.Header.GetIndex": {Execute: HeaderGetIndex, Validator: validatorHeader},
		"Neo.Header.GetHash": {Execute: HeaderGetHash, Validator: validatorHeader},
		"Neo.Header.GetVersion": {Execute: HeaderGetVersion, Validator: validatorHeader},
		"Neo.Header.GetPrevHash": {Execute: HeaderGetPrevHash, Validator: validatorHeader},
		"Neo.Header.GetTimestamp": {Execute: HeaderGetTimestamp, Validator: validatorHeader},
		"Neo.Header.GetConsensusData": {Execute: HeaderGetConsensusData, Validator: validatorHeader},
		"Neo.Header.GetNextConsensus": {Execute: HeaderGetNextConsensus, Validator: validatorHeader},
		"Neo.Header.GetMerkleRoot": {Execute: HeaderGetMerkleRoot, Validator: validatorHeader},
		"Neo.Transaction.GetHash": {Execute: TransactionGetHash, Validator: validatorTransaction},
		"Neo.Transaction.GetType": {Execute: TransactionGetType, Validator: validatorTransaction},
		"Neo.Transaction.GetAttributes": {Execute: TransactionGetAttributes, Validator: validatorTransaction},
		"Neo.Contract.Create": {Execute: ContractCreate},
		"Neo.Contract.Migrate": {Execute: ContractMigrate},
		"Neo.Contract.GetStorageContext": {Execute: ContractGetStorageContext},
		"Neo.Contract.Destroy": {Execute: ContractDestory},
		"Neo.Contract.GetScript": {Execute: ContractGetCode, Validator: validatorGetCode},
		"Neo.Runtime.GetTime": {Execute: RuntimeGetTime},
		"Neo.Runtime.CheckWitness": {Execute: RuntimeCheckWitness, Validator: validatorCheckWitness},
		"Neo.Runtime.Notify": {Execute: RuntimeNotify, Validator: validatorNotify},
		"Neo.Runtime.Log": {Execute: RuntimeLog, Validator: validatorLog},
		"Neo.Runtime.CheckSig": {Execute: RuntimeCheckSig, Validator: validatorCheckSig},
		"Neo.Storage.Get": {Execute: StorageGet},
		"Neo.Storage.Put": {Execute: StoragePut},
		"Neo.Storage.Delete": {Execute: StorageDelete},
		"Neo.Storage.GetContext": {Execute: StorageGetContext},
		"System.ExecutionEngine.GetScriptContainer": {Execute: GetCodeContainer},
		"System.ExecutionEngine.GetExecutingScriptHash": {Execute: GetExecutingAddress},
		"System.ExecutionEngine.GetCallingScriptHash": {Execute: GetCallingAddress},
		"System.ExecutionEngine.GetEntryScriptHash": {Execute: GetEntryAddress},
	}
)

var (
	ERR_CHECK_STACK_SIZE = errors.NewErr("[NeoVmService] vm over max stack size!")
	ERR_CHECK_ARRAY_SIZE = errors.NewErr("[NeoVmService] vm over max array size!")
	ERR_CHECK_BIGINTEGER = errors.NewErr("[NeoVmService] vm over max biginteger size!")
	ERR_CURRENT_CONTEXT_NIL = errors.NewErr("[NeoVmService] neovm service current context doesn't exist!")
	ERR_EXECUTE_CODE = errors.NewErr("[NeoVmService] vm execute code invalid!")
)

type (
	Execute func(service *NeoVmService, engine *vm.ExecutionEngine) error
	Validator func(engine *vm.ExecutionEngine) error
)

type Service struct {
	Execute   Execute
	Validator Validator
}

type NeoVmService struct {
	Store         store.LedgerStore
	CloneCache    *storage.CloneCache
	ContextRef    context.ContextRef
	Notifications []*event.NotifyEventInfo
	Tx            *types.Transaction
	Time          uint32
}

func NewNeoVmService(store store.LedgerStore, dbCache scommon.StateStore, tx *types.Transaction, time uint32, ctxRef context.ContextRef) *NeoVmService {
	var service NeoVmService
	service.Store = store
	service.CloneCache = storage.NewCloneCache(dbCache)
	service.Time = time
	service.Tx = tx
	service.ContextRef = ctxRef
	return &service
}

func (this *NeoVmService) Invoke() error {
	engine := vm.NewExecutionEngine()
	ctx := this.ContextRef.CurrentContext()
	if ctx == nil {
		return ERR_CURRENT_CONTEXT_NIL
	}
	if len(ctx.Code.Code) == 0 {
		return ERR_EXECUTE_CODE
	}
	engine.PushContext(vm.NewExecutionContext(engine, ctx.Code.Code))
	for {
		if len(engine.Contexts) == 0 || engine.Context == nil {
			break
		}
		if engine.Context.GetInstructionPointer() >= len(engine.Context.Code) {
			break
		}
		if err := engine.ExecuteCode(); err != nil {
			return err
		}
		if engine.Context.GetInstructionPointer() < len(engine.Context.Code) {
			if ok := checkStackSize(engine); !ok {
				return ERR_CHECK_STACK_SIZE
			}
			if ok := checkArraySize(engine); !ok {
				return ERR_CHECK_ARRAY_SIZE
			}
			if ok := checkBigIntegers(engine); !ok {
				return ERR_CHECK_BIGINTEGER
			}
		}
		switch engine.OpCode {
		case vm.SYSCALL:
			if err := this.SystemCall(engine); err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[NeoVmService] service system call error!")
			}
		case vm.APPCALL:
			c := new(states.Contract)
			if err := c.Deserialize(engine.Context.OpReader.Reader()); err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[NeoVmService] get contract parameters error!")
			}
			if _,err := this.ContextRef.AppCall(c.Address, c.Method, c.Code, c.Args); err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[NeoVmService] service app call error!")
			}
		default:
			if err := engine.StepInto(); err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[NeoVmService] vm execute error!")
			}
		}
	}
	this.ContextRef.PushNotifications(this.Notifications)
	this.CloneCache.Commit()
	return nil
}

func (this *NeoVmService) SystemCall(engine *vm.ExecutionEngine) error {
	serviceName := engine.Context.OpReader.ReadVarString()
	service, ok := ServiceMap[serviceName]
	if !ok {
		return errors.NewErr("[SystemCall] service not support!")
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
		switch engine.OpCode{
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

func checkArraySize(engine *vm.ExecutionEngine) bool {
	switch engine.OpCode {
	case vm.PACK:
	case vm.NEWARRAY:
	case vm.NEWSTRUCT:
		if engine.EvaluationStack.Count() == 0 {
			return false
		}
		size := vm.PeekInt(engine)
		if size > MAX_ARRAY_SIZE {
			return false
		}
	}
	return true
}

func checkBigIntegers(engine *vm.ExecutionEngine) bool {
	switch engine.OpCode {
	case vm.INC:
		if engine.EvaluationStack.Count() == 0 {
			return false
		}
		x := vm.PeekBigInteger(engine)
		if !checkBigInteger(x) || !checkBigInteger(new(big.Int).Add(x, big.NewInt(1))) {
			return false
		}
	case vm.DEC:
		if engine.EvaluationStack.Count() == 0 {
			return false
		}
		x := vm.PeekBigInteger(engine)
		if !checkBigInteger(x) || (x.Sign() < 0 && !checkBigInteger(new(big.Int).Sub(x, big.NewInt(1)))) {
			return false
		}
	case vm.ADD:
		if engine.EvaluationStack.Count() < 2 {
			return false
		}
		x2 := vm.PeekBigInteger(engine)
		x1 := vm.PeekNBigInt(1, engine)
		if !checkBigInteger(x1) || !checkBigInteger(x2) || !checkBigInteger(new(big.Int).Add(x1, x2)) {
			return false
		}
	case vm.SUB:
		if engine.EvaluationStack.Count() < 2 {
			return false
		}
		x2 := vm.PeekBigInteger(engine)
		x1 := vm.PeekNBigInt(1, engine)
		if !checkBigInteger(x1) || !checkBigInteger(x2) || !checkBigInteger(new(big.Int).Sub(x1, x2)) {
			return false
		}
	case vm.MUL:
		if engine.EvaluationStack.Count() < 2 {
			return false
		}
		x2 := vm.PeekBigInteger(engine)
		x1 := vm.PeekNBigInt(1, engine)
		lx2 := len(vmtype.ConvertBigIntegerToBytes(x2))
		lx1 := len(vmtype.ConvertBigIntegerToBytes(x1))
		if lx2 > MAX_SIZE_FOR_BIGINTEGER || lx1 > MAX_SIZE_FOR_BIGINTEGER || (lx1 + lx2) > MAX_SIZE_FOR_BIGINTEGER {
			return false
		}
	case vm.DIV:
		if engine.EvaluationStack.Count() < 2 {
			return false
		}
		x2 := vm.PeekBigInteger(engine)
		x1 := vm.PeekNBigInt(1, engine)
		if !checkBigInteger(x2) || !checkBigInteger(x1) {
			return false
		}
		if x2.Sign() == 0 {
			return false
		}
	case vm.MOD:
		if engine.EvaluationStack.Count() < 2 {
			return false
		}
		x2 := vm.PeekBigInteger(engine)
		x1 := vm.PeekNBigInt(1, engine)
		if !checkBigInteger(x2) || !checkBigInteger(x1) {
			return false
		}
		if x2.Sign() == 0 {
			return false
		}
	}
	return true
}

func checkBigInteger(value *big.Int) bool {
	if value == nil {
		return false
	}
	if len(vmtype.ConvertBigIntegerToBytes(value)) > MAX_SIZE_FOR_BIGINTEGER {
		return false
	}
	return true
}