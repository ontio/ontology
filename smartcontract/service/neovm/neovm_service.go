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
	"bytes"
	"fmt"
	scommon "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/store"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/context"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/storage"
	vm "github.com/ontio/ontology/vm/neovm"
	vmty "github.com/ontio/ontology/vm/neovm/types"
	"io"
)

var (
	// Register all service for smart contract execute
	ServiceMap = map[string]Service{
		ATTRIBUTE_GETUSAGE_NAME:              {Execute: AttributeGetUsage},
		ATTRIBUTE_GETDATA_NAME:               {Execute: AttributeGetData},
		BLOCK_GETTRANSACTIONCOUNT_NAME:       {Execute: BlockGetTransactionCount},
		BLOCK_GETTRANSACTIONS_NAME:           {Execute: BlockGetTransactions},
		BLOCK_GETTRANSACTION_NAME:            {Execute: BlockGetTransaction},
		BLOCKCHAIN_GETHEIGHT_NAME:            {Execute: BlockChainGetHeight},
		BLOCKCHAIN_GETHEADER_NAME:            {Execute: BlockChainGetHeader},
		BLOCKCHAIN_GETBLOCK_NAME:             {Execute: BlockChainGetBlock},
		BLOCKCHAIN_GETTRANSACTION_NAME:       {Execute: BlockChainGetTransaction},
		BLOCKCHAIN_GETCONTRACT_NAME:          {Execute: BlockChainGetContract},
		BLOCKCHAIN_GETTRANSACTIONHEIGHT_NAME: {Execute: BlockChainGetTransactionHeight},
		HEADER_GETINDEX_NAME:                 {Execute: HeaderGetIndex},
		HEADER_GETHASH_NAME:                  {Execute: HeaderGetHash},
		HEADER_GETVERSION_NAME:               {Execute: HeaderGetVersion},
		HEADER_GETPREVHASH_NAME:              {Execute: HeaderGetPrevHash},
		HEADER_GETTIMESTAMP_NAME:             {Execute: HeaderGetTimestamp},
		HEADER_GETCONSENSUSDATA_NAME:         {Execute: HeaderGetConsensusData},
		HEADER_GETNEXTCONSENSUS_NAME:         {Execute: HeaderGetNextConsensus},
		HEADER_GETMERKLEROOT_NAME:            {Execute: HeaderGetMerkleRoot},
		TRANSACTION_GETHASH_NAME:             {Execute: TransactionGetHash},
		TRANSACTION_GETTYPE_NAME:             {Execute: TransactionGetType},
		TRANSACTION_GETATTRIBUTES_NAME:       {Execute: TransactionGetAttributes},
		CONTRACT_CREATE_NAME:                 {Execute: ContractCreate},
		CONTRACT_MIGRATE_NAME:                {Execute: ContractMigrate},
		CONTRACT_GETSTORAGECONTEXT_NAME:      {Execute: ContractGetStorageContext},
		CONTRACT_DESTROY_NAME:                {Execute: ContractDestory},
		CONTRACT_GETSCRIPT_NAME:              {Execute: ContractGetCode},
		RUNTIME_GETTIME_NAME:                 {Execute: RuntimeGetTime},
		RUNTIME_CHECKWITNESS_NAME:            {Execute: RuntimeCheckWitness},
		RUNTIME_NOTIFY_NAME:                  {Execute: RuntimeNotify},
		RUNTIME_LOG_NAME:                     {Execute: RuntimeLog},
		RUNTIME_GETTRIGGER_NAME:              {Execute: RuntimeGetTrigger},
		RUNTIME_SERIALIZE_NAME:               {Execute: RuntimeSerialize},
		RUNTIME_DESERIALIZE_NAME:             {Execute: RuntimeDeserialize},
		RUNTIME_VERIFYMUTISIG_NAME:           {Execute: RuntimeVerifyMutiSig},
		NATIVE_INVOKE_NAME:                   {Execute: NativeInvoke},
		WASM_INVOKE_NAME:                     {Execute: WASMInvoke},
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

		RUNTIME_BASE58TOADDRESS_NAME:     {Execute: RuntimeBase58ToAddress},
		RUNTIME_ADDRESSTOBASE58_NAME:     {Execute: RuntimeAddressToBase58},
		RUNTIME_GETCURRENTBLOCKHASH_NAME: {Execute: RuntimeGetCurrentBlockHash},
	}
)

var (
	ERR_CHECK_STACK_SIZE  = errors.NewErr("[NeoVmService] vm execution exceeded the max stack size!")
	ERR_EXECUTE_CODE      = errors.NewErr("[NeoVmService] vm execution code was invalid!")
	ERR_GAS_INSUFFICIENT  = errors.NewErr("[NeoVmService] insufficient gas for transaction!")
	VM_EXEC_STEP_EXCEED   = errors.NewErr("[NeoVmService] vm execution exceeded the step limit!")
	CONTRACT_NOT_EXIST    = errors.NewErr("[NeoVmService] the given contract does not exist!")
	DEPLOYCODE_TYPE_ERROR = errors.NewErr("[NeoVmService] deploy code type error!")
	VM_EXEC_FAULT         = errors.NewErr("[NeoVmService] vm execution encountered a state fault!")
)

type (
	Execute func(service *NeoVmService, engine *vm.Executor) error
)

type Service struct {
	Execute Execute
}

// NeoVmService is a struct for smart contract provide interop service
type NeoVmService struct {
	Store         store.LedgerStore
	CacheDB       *storage.CacheDB
	ContextRef    context.ContextRef
	Notifications []*event.NotifyEventInfo
	Code          []byte
	GasTable      map[string]uint64
	Tx            *types.Transaction
	Time          uint32
	Height        uint32
	BlockHash     scommon.Uint256
	Engine        *vm.Executor
	PreExec       bool
}

// Invoke a smart contract
func (this *NeoVmService) Invoke() (interface{}, error) {
	if len(this.Code) == 0 {
		return nil, ERR_EXECUTE_CODE
	}
	this.ContextRef.PushContext(&context.Context{ContractAddress: scommon.AddressFromVmCode(this.Code), Code: this.Code})
	var gasTable [256]uint64
	for {
		//check the execution step count
		if this.PreExec && !this.ContextRef.CheckExecStep() {
			return nil, VM_EXEC_STEP_EXCEED
		}
		if this.Engine.Context == nil {
			break
		}
		if this.Engine.Context.GetInstructionPointer() >= len(this.Engine.Context.Code) {
			break
		}
		opCode, eof := this.Engine.Context.ReadOpCode()
		if eof {
			return nil, io.EOF
		}

		price := gasTable[opCode]
		if opCode >= vm.PUSHBYTES1 && opCode <= vm.PUSHBYTES75 {
			price = OPCODE_GAS
		} else if price == 0 {
			opExec := vm.OpExecList[opCode]
			p, err := GasPrice(this.GasTable, this.Engine, opExec.Name)
			if err != nil {
				return nil, err
			}
			price = p
			// note: this works because the gas fee for opcode is constant
			gasTable[opCode] = price
		}

		if !this.ContextRef.CheckUseGas(price) {
			return nil, ERR_GAS_INSUFFICIENT
		}

		switch opCode {
		case vm.SYSCALL:
			if err := this.SystemCall(this.Engine); err != nil {
				return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[NeoVmService] service system call error!")
			}
		case vm.APPCALL:
			address, err := this.Engine.Context.OpReader.ReadBytes(20)
			if err != nil {
				return nil, fmt.Errorf("[Appcall] read contract address error:%v", err)
			}
			if bytes.Compare(address, scommon.ADDRESS_EMPTY[:]) == 0 {
				if this.Engine.EvalStack.Count() < 1 {
					return nil, fmt.Errorf("[Appcall] too few input parameters: %d", this.Engine.EvalStack.Count())
				}
				address, err = this.Engine.EvalStack.PopAsBytes()
				if err != nil {
					return nil, fmt.Errorf("[Appcall] pop contract address error:%v", err)
				}
				if len(address) != 20 {
					return nil, fmt.Errorf("[Appcall] pop contract address len != 20:%x", address)
				}
			}
			addr, err := scommon.AddressParseFromBytes(address)
			if err != nil {
				return nil, err
			}
			code, err := this.GetNeoContract(addr)
			if err != nil {
				return nil, err
			}
			service, err := this.ContextRef.NewExecuteEngine(code, types.InvokeNeo)
			if err != nil {
				return nil, err
			}
			err = this.Engine.EvalStack.CopyTo(service.(*NeoVmService).Engine.EvalStack)
			if err != nil {
				return nil, fmt.Errorf("[Appcall] EvalStack CopyTo error:%x", err)
			}
			result, err := service.Invoke()
			if err != nil {
				return nil, err
			}
			if result != nil {
				val := result.(*vmty.VmValue)
				err := this.Engine.EvalStack.Push(*val)
				if err != nil {
					return nil, err
				}
			}
		default:
			state, err := this.Engine.ExecuteOp(opCode, this.Engine.Context)
			if err != nil {
				return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[NeoVmService] vm execution error!")
			}
			if state == vm.FAULT {
				return nil, VM_EXEC_FAULT
			}
		}
	}
	this.ContextRef.PopContext()
	this.ContextRef.PushNotifications(this.Notifications)
	if this.Engine.EvalStack.Count() != 0 {
		val, err := this.Engine.EvalStack.Peek(0)
		if err != nil {
			return nil, err
		}
		return &val, nil
	}
	return nil, nil
}

// SystemCall provide register service for smart contract to interaction with blockchain
func (this *NeoVmService) SystemCall(engine *vm.Executor) error {
	serviceName, err := engine.Context.OpReader.ReadVarString(vm.MAX_BYTEARRAY_SIZE)
	if err != nil {
		return err
	}
	service, ok := ServiceMap[serviceName]
	if !ok {
		return errors.NewErr(fmt.Sprintf("[SystemCall] the given service is not supported: %s", serviceName))
	}
	price, err := GasPrice(this.GasTable, engine, serviceName)
	if err != nil {
		return err
	}
	if !this.ContextRef.CheckUseGas(price) {
		return ERR_GAS_INSUFFICIENT
	}
	if err := service.Execute(this, engine); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[SystemCall] service execution error!")
	}
	return nil
}

func (this *NeoVmService) GetNeoContract(address scommon.Address) ([]byte, error) {
	dep, err := this.CacheDB.GetContract(address)
	if err != nil {
		return nil, errors.NewErr("[getNeoContract] get contract context error!")
	}
	log.Debugf("invoke contract address:%s", address.ToHexString())
	if dep == nil {
		return nil, CONTRACT_NOT_EXIST
	}
	return dep.GetNeoCode()
}
