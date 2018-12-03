///*
// * Copyright (C) 2018 The ontology Authors
// * This file is part of The ontology library.
// *
// * The ontology is free software: you can redistribute it and/or modify
// * it under the terms of the GNU Lesser General Public License as published by
// * the Free Software Foundation, either version 3 of the License, or
// * (at your option) any later version.
// *
// * The ontology is distributed in the hope that it will be useful,
// * but WITHOUT ANY WARRANTY; without even the implied warranty of
// * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// * GNU Lesser General Public License for more details.
// *
// * You should have received a copy of the GNU Lesser General Public License
// * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
// */
package wasmvm

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/store"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/context"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/states"
	"github.com/ontio/ontology/smartcontract/storage"
	"github.com/ontio/ontology/vm/wasmvm/exec"
	"github.com/ontio/ontology/vm/wasmvm/util"
)

type WasmVmService struct {
	Store   store.LedgerStore
	CacheDB *storage.CacheDB
	//CloneCache    *storage.CloneCache
	ContextRef    context.ContextRef
	Notifications []*event.NotifyEventInfo
	Code          []byte
	Tx            *types.Transaction
	Time          uint32
	Height        uint32
	BlockHash     common.Uint256
	Gas           *uint64
}

var (
	ERR_CHECK_STACK_SIZE  = errors.NewErr("[WasmVmService] vm over max stack size!")
	ERR_EXECUTE_CODE      = errors.NewErr("[WasmVmService] vm execute code invalid!")
	ERR_GAS_INSUFFICIENT  = errors.NewErr("[WasmVmService] gas insufficient")
	VM_EXEC_STEP_EXCEED   = errors.NewErr("[WasmVmService] vm execute step exceed!")
	CONTRACT_NOT_EXIST    = errors.NewErr("[WasmVmService] Get contract code from db fail")
	DEPLOYCODE_TYPE_ERROR = errors.NewErr("[WasmVmService] DeployCode type error!")
	VM_EXEC_FAULT         = errors.NewErr("[WasmVmService] vm execute state fault!")
)

func (this *WasmVmService) Invoke() (interface{}, error) {

	if len(this.Code) == 0 {
		return nil, ERR_EXECUTE_CODE
	}

	stateMachine := NewWasmStateMachine()
	//register the "CallContract" function
	stateMachine.Register(exec.APPCALL_NAME, this.callContract)

	stateMachine.Register(exec.NATIVE_INVOKE_NAME, this.nativeInvoke)
	stateMachine.Register(exec.MARSHALNATIVEPARAMS_NAME, this.marshalNativeParams)
	//stateMachine.Register("ONT_MarshalNeoParams", this.marshalNeoParams)
	//runtime
	stateMachine.Register(exec.RUNTIME_CHECKWITNESS_NAME, this.runtimeCheckWitness)
	stateMachine.Register(exec.RUNTIME_NOTIFY_NAME, this.runtimeNotify)
	stateMachine.Register(exec.RUNTIME_CHECKSIG_NAME, this.runtimeCheckSig)
	stateMachine.Register(exec.RUNTIME_GETTIME_NAME, this.runtimeGetTime)
	stateMachine.Register(exec.RUNTIME_LOG_NAME, this.runtimeLog)
	stateMachine.Register(exec.RUNTIME_RAISEEXCEPTION_NAME, this.runtimeRaiseException)
	stateMachine.Register(exec.RUNTIME_GETCURRENTBLOCKHASH, this.runtimeGetCurrentBlockHash)
	stateMachine.Register(exec.RUNTIME_GETCODECONTAINER, this.runtimeGetCodeContainer)
	stateMachine.Register(exec.RUNTIME_GETEXECUTINGADDRESS, this.runtimeGetExecutingAddress)
	stateMachine.Register(exec.RUNTIME_GETCALLINGADDRESS, this.runtimeGetCallingAddress)
	stateMachine.Register(exec.RUNTIME_GETENTRYADDRESS, this.runtimeGetEntryAddress)
	stateMachine.Register(exec.RUNTIME_ADDRESSTOBASE58, this.runtimeAddressToBase58)
	stateMachine.Register(exec.RUNTIME_ADDRESSTOHEX, this.runtimeAddressToHex)

	//attribute
	stateMachine.Register(exec.ATTRIBUTE_GETUSAGE_NAME, this.attributeGetUsage)
	stateMachine.Register(exec.ATTRIBUTE_GETDATA_NAME, this.attributeGetData)
	//block
	stateMachine.Register(exec.BLOCK_GETHEADERHASH_NAME, this.blockGetCurrentHeaderHash)
	stateMachine.Register(exec.BLOCK_GETHEADERHEIGHT_NAME, this.blockGetCurrentHeaderHeight)
	stateMachine.Register(exec.BLOCK_GETBLOCKHASH_NAME, this.blockGetCurrentBlockHash)
	stateMachine.Register(exec.BLOCK_GETBLOCKHEIGHT_NAME, this.blockGetCurrentBlockHeight)
	stateMachine.Register(exec.BLOCK_GETTRANSACTIONBYHASH_NAME, this.blockGetTransactionByHash)
	stateMachine.Register(exec.BLOCK_GETTRANSACTIONCOUNT_NAME, this.blockGetTransactionCount)
	stateMachine.Register(exec.BLOCK_GETTRANSACTIONS, this.blockGetTransactions)

	//blockchain
	stateMachine.Register(exec.BLOCKCHAIN_GETHEGITH_NAME, this.blockChainGetHeight)
	stateMachine.Register(exec.BLOCKCHAIN_GETHEADERBYHEIGHT_NAME, this.blockChainGetHeaderByHeight)
	stateMachine.Register(exec.BLOCKCHAIN_GETHEADERBYHASH_NAME, this.blockChainGetHeaderByHash)
	stateMachine.Register(exec.BLOCKCHAIN_GETBLOCKBYHEIGHT_NAME, this.blockChainGetBlockByHeight)
	stateMachine.Register(exec.BLOCKCHAIN_GETBLOCKBYHASH_NAME, this.blockChainGetBlockByHash)
	stateMachine.Register(exec.BLOCKCHAIN_GETCONTRACT_NAME, this.blockChainGetContract)

	//header
	stateMachine.Register(exec.HEADER_GETHASH_NAME, this.headerGetHash)
	stateMachine.Register(exec.HEADER_GETVERSION_NAME, this.headerGetVersion)
	stateMachine.Register(exec.HEADER_GETPREVHASH_NAME, this.headerGetPrevHash)
	stateMachine.Register(exec.HEADER_GETMERKLEROOT_NAME, this.headerGetMerkleRoot)
	stateMachine.Register(exec.HEADER_GETINDEX_NAME, this.headerGetIndex)
	stateMachine.Register(exec.HEADER_GETTIMESTAMP_NAME, this.headerGetTimestamp)
	stateMachine.Register(exec.HEADER_GETCONSENSUSDATA_NAME, this.headerGetConsensusData)
	stateMachine.Register(exec.HEADER_GETNEXTCONSENSUS_NAME, this.headerGetNextConsensus)

	//storage
	stateMachine.Register(exec.STORAGE_PUT_NAME, this.putstore)
	stateMachine.Register(exec.STORAGE_GET_NAME, this.getstore)
	stateMachine.Register(exec.STORAGE_DELETE_NAME, this.deletestore)

	//transaction
	stateMachine.Register(exec.TRANSACTION_GETHASH_NAME, this.transactionGetHash)
	stateMachine.Register(exec.TRANSACTION_GETTYPE_NAME, this.transactionGetType)
	stateMachine.Register(exec.TRANSACTION_GETATTRIBUTES_NAME, this.transactionGetAttributes)

	//contract
	stateMachine.Register(exec.WASM_CONTRACT_CREATE_NAME, this.contractCreate)
	stateMachine.Register(exec.WASM_CONTRACT_MIGRATE_NAME, this.contractMigrate)
	stateMachine.Register(exec.WASM_CONTRACT_DELETE_NAME, this.contractDelete)

	engine := exec.NewExecutionEngine(
		new(util.ECDsaCrypto),
		stateMachine,
		this.Gas,
	)

	contract := &states.ContractInvokeParam{}
	contract.Deserialize(bytes.NewBuffer(this.Code))
	//addr := contract.Address

	code, err := this.Store.GetContractState(contract.Address)
	if err != nil {
		return nil, err
	}

	this.ContextRef.PushContext(&context.Context{ContractAddress: contract.Address, Code: code.Code})

	var caller common.Address
	if this.ContextRef.CallingContext() == nil {
		caller = common.Address{}
	} else {
		caller = this.ContextRef.CallingContext().ContractAddress
	}
	//this.ContextRef.PushContext(&context.Context{ContractAddress: contract.Address})

	res, err := engine.Call(caller, code.Code, contract.Method, contract.Args, contract.Version)

	if err != nil {
		return nil, err
	}

	//get the return message
	result, err := engine.GetVM().GetPointerMemory(uint64(binary.LittleEndian.Uint32(res)))
	if err != nil {
		return nil, err
	}

	this.ContextRef.PopContext()
	this.ContextRef.PushNotifications(this.Notifications)
	return result, nil
}

// marshalNativeParams
// make parameter bytes for call native contract
func (this *WasmVmService) marshalNativeParams(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[callContract]parameter count error while call marshalNativeParams")
	}

	transferbytes, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}
	//transferbytes is a nested struct with states.Transfer
	//type Transfers struct {
	//	States  []*State		   -------->i32 pointer 4 bytes
	//}
	if len(transferbytes) != 4 {
		return false, errors.NewErr("[callContract]parameter format error while call marshalNativeParams")
	}
	transfer := &ont.Transfers{}

	statesAddr := binary.LittleEndian.Uint32(transferbytes[:4])
	statesbytes, err := vm.GetPointerMemory(uint64(statesAddr))
	if err != nil {
		return false, err
	}

	//statesbytes is slice of struct with states.
	//type State struct {
	//	From    common.Address  -------->i32 pointer 4 bytes
	//	To      common.Address  -------->i32 pointer 4 bytes
	//	Value   *big.Int        -------->i64 8 bytes
	//}
	//total is 4 + 4 + 8 = 24 bytes
	statecnt := len(statesbytes) / 16
	states := make([]ont.State, statecnt)

	for i := 0; i < statecnt; i++ {
		tmpbytes := statesbytes[i*16 : (i+1)*16]
		state := ont.State{}
		fromAddessBytes, err := vm.GetPointerMemory(uint64(binary.LittleEndian.Uint32(tmpbytes[:4])))
		if err != nil {
			return false, err
		}
		fromAddress, err := common.AddressFromBase58(util.TrimBuffToString(fromAddessBytes))
		if err != nil {
			return false, err
		}
		state.From = fromAddress

		toAddressBytes, err := vm.GetPointerMemory(uint64(binary.LittleEndian.Uint32(tmpbytes[4:8])))
		if err != nil {
			return false, err
		}
		toAddress, err := common.AddressFromBase58(util.TrimBuffToString(toAddressBytes))
		state.To = toAddress
		//tmpbytes[12:16] is padding
		amount := binary.LittleEndian.Uint64(tmpbytes[8:])
		state.Value = amount
		states[i] = state

	}
	transfer.States = states
	tbytes := new(bytes.Buffer)
	transfer.Serialize(tbytes)
	result, err := vm.SetPointerMemory(tbytes.Bytes())
	if err != nil {
		return false, err
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(result))
	return true, nil
}

//get contract code from address
func (this *WasmVmService) getContract(bs []byte) ([]byte, error) {
	address, err := common.AddressParseFromBytes(bs)
	dep, err := this.CacheDB.GetContract(address)
	if err != nil {
		return nil, errors.NewErr("[getContract] Get contract context error!")
	}
	log.Debugf("invoke contract address:%s", address.ToHexString())
	if dep == nil {
		return nil, CONTRACT_NOT_EXIST
	}
	return dep.Code, nil
}

//native invoke
func (this *WasmVmService) nativeInvoke(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 4 {
		return false, errors.NewErr("[nativeInvoke]parameter count error while call readMessage")
	}
	//get version
	version := params[0]

	//get contractAddress
	contractAddressIdx := params[1]
	addr, err := vm.GetPointerMemory(contractAddressIdx)
	if err != nil {
		return false, errors.NewErr("[nativeInvoke]get Contract address failed:" + err.Error())
	}

	if len(addr) == 0 {
		return false, errors.NewErr("[nativeInvoke]No native contract address found!")
	}
	addrbytes, err := common.HexToBytes(util.TrimBuffToString(addr))

	contractAddress, err := common.AddressParseFromBytes(addrbytes)

	//get method
	methodIdx := params[2]
	method, err := vm.GetPointerMemory(methodIdx)
	if err != nil {
		return false, errors.NewErr("[nativeInvoke]get method failed!")
	}
	methodName := util.TrimBuffToString(method)

	//get args
	argsIdx := params[3]
	argsbytes, err := vm.GetPointerMemory(argsIdx)
	contract := states.ContractInvokeParam{
		Version: byte(version),
		Address: contractAddress,
		Method:  methodName,
		Args:    argsbytes,
	}

	native := &native.NativeService{
		CacheDB:     this.CacheDB,
		InvokeParam: contract,
		Tx:          this.Tx,
		Height:      this.Height,
		Time:        this.Time,
		ContextRef:  this.ContextRef,
		ServiceMap:  make(map[string]native.Handler),
	}

	result, err := native.Invoke()
	if err != nil {
		return false, errors.NewErr("[nativeInvoke]AppCall failed:" + err.Error())
	}

	if envCall.GetReturns() {
		//res = fmt.Sprintf("%s", result)

		idx, err := vm.SetPointerMemory(result)
		if err != nil {
			return false, errors.NewErr("[callContract]SetPointerMemory failed:" + err.Error())
		}
		vm.PushResult(uint64(idx))
	}

	return true, nil

}

// callContract
// need 3 parameters
//0: contract address
//1: method name
//2: args
func (this *WasmVmService) callContract(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 3 {
		return false, errors.NewErr("[callContract]parameter count error while call readMessage")
	}

	var contractAddress common.Address
	//var contractBytes []byte
	//get contract address
	contractAddressIdx := params[0]
	addr, err := vm.GetPointerMemory(contractAddressIdx)
	if err != nil {
		return false, errors.NewErr("[callContract]get Contract address failed:" + err.Error())
	}

	if addr != nil {
		//addrbytes, err := common.HexToBytes(util.TrimBuffToString(addr))
		//if err != nil {
		//	return false, errors.NewErr("[callContract]get contract address error:" + err.Error())
		//}
		contractAddress, err = common.AddressFromBase58(util.TrimBuffToString(addr))
		if err != nil {
			return false, errors.NewErr("[callContract]get contract address error:" + err.Error())
		}

	}

	methodName, err := vm.GetPointerMemory(params[1])
	if err != nil {
		return false, errors.NewErr("[callContract]get Contract methodName failed:" + err.Error())
	}
	//get args
	args, err := vm.GetPointerMemory(params[2])

	if err != nil {
		return false, errors.NewErr("[callContract]get Contract arg failed:" + err.Error())
	}
	this.ContextRef.PushContext(&context.Context{
		Code:            vm.VMCode,
		ContractAddress: vm.ContractAddress})

	bf := new(bytes.Buffer)
	contract := states.ContractInvokeParam{
		Version: 1, //fix to > 0
		Address: contractAddress,
		Method:  string(methodName),
		Args:    args,
	}

	if err := contract.Serialize(bf); err != nil {
		return false, err
	}

	this.Code = bf.Bytes()
	result, err := this.Invoke()
	if err != nil {
		return false, errors.NewErr("[callContract]AppCall failed:" + err.Error())
	}
	this.ContextRef.PopContext()
	vm.RestoreCtx()
	//var res string
	if envCall.GetReturns() {
		//res = fmt.Sprintf("%s", result)

		idx, err := vm.SetPointerMemory(result)
		if err != nil {
			return false, errors.NewErr("[callContract]SetPointerMemory failed:" + err.Error())
		}
		vm.PushResult(uint64(idx))
	}

	return true, nil
}

//GetContractAddress return contract address
func GetContractAddress(code string) (common.Address, error) {
	data, err := hex.DecodeString(code)
	if err != nil {
		return common.Address{}, err
	}

	return common.AddressFromVmCode(data), nil
}
