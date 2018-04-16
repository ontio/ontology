package wasmvm

import (
	"bytes"
	"encoding/binary"
	"github.com/ontio/ontology/core/store"
	"github.com/ontio/ontology/smartcontract/storage"
	"github.com/ontio/ontology/smartcontract/context"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/core/types"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/states"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/vm/wasmvm/exec"
	"github.com/ontio/ontology/vm/wasmvm/util"

	"github.com/ontio/ontology-go-sdk/utils"
	stype  "github.com/ontio/ontology/smartcontract/types"
)

type WasmVmService struct {
	Store         store.LedgerStore
	CloneCache    *storage.CloneCache
	ContextRef    context.ContextRef
	Notifications []*event.NotifyEventInfo
	Tx            *types.Transaction
	Time          uint32
}

func NewWasmVmService(store store.LedgerStore,
						dbCache scommon.StateStore,
						tx *types.Transaction,
						time uint32,
						ctxRef context.ContextRef) *WasmVmService {
	var service WasmVmService
	service.Store = store
	service.CloneCache = storage.NewCloneCache(dbCache)
	service.Time = time
	service.Tx = tx
	service.ContextRef = ctxRef
	return &service
}

func (this *WasmVmService)Invoke() ([]byte,error){
	stateMachine := NewWasmStateMachine(this.Store, this.CloneCache,  this.Time)
	//register the "CallContract" function
	stateMachine.Register("CallContract",this.callContract)

	ctx := this.ContextRef.CurrentContext()
	engine := exec.NewExecutionEngine(
		this.Tx,
		new(util.ECDsaCrypto),
		stateMachine,
	)

	contract := &states.Contract{}
	contract.Deserialize(bytes.NewBuffer(ctx.Code.Code))
	addr := contract.Address
	if contract.Code == nil{
		dpcode, err := this.GetContractCodeFromAddress(addr)
		if err != nil {
			return nil,errors.NewErr("get contract  error")
		}
		contract.Code = dpcode
	}
	var caller common.Address
	if this.ContextRef.CallingContext() == nil {
		caller = common.Address{}
	} else {
		caller = this.ContextRef.CallingContext().ContractAddress
	}
	res, err := engine.Call(caller, contract.Code, contract.Method, contract.Args, contract.Version)
	if err != nil {
		return nil,err
	}

	//get the return message
	result, err := engine.GetVM().GetPointerMemory(uint64(binary.LittleEndian.Uint32(res)))
	if err != nil {
		return nil,err
	}
	this.CloneCache.Commit()
	this.ContextRef.PushNotifications(stateMachine.Notifications)
	return result,nil
}

//call contract will need 4 paramters
//0: contract address / contract code
//1: method name
//2: args
//3: isOffChain  "true" / "false"
func(this *WasmVmService)callContract(engine *exec.ExecutionEngine)(bool,error){

	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 4 {
		return false, errors.NewErr("[callContract]parameter count error while call readMessage")
	}
	contractAddressIdx := params[0]
	methodName, err := vm.GetPointerMemory(params[1])
	if err != nil {
		return false, errors.NewErr("[callContract]get Contract methodName failed:" + err.Error())
	}

	arg, err := vm.GetPointerMemory(params[2])
	if err != nil {
		return false, errors.NewErr("[callContract]get Contract arg failed:" + err.Error())
	}


	isoffchain,err := vm.GetPointerMemory(params[3])
	if err != nil {
		return false, errors.NewErr("[callContract]get Contract methodName failed:" + err.Error())
	}

	var result []byte
	if util.TrimBuffToString(isoffchain) == "false" {

		addr, err := vm.GetPointerMemory(contractAddressIdx)
		if err != nil {
			return false, errors.NewErr("[callContract]get Contract address failed:" + err.Error())
		}
		addrbytes, err := common.HexToBytes(util.TrimBuffToString(addr))
		if err != nil {
			return false, errors.NewErr("[callContract]get contract address error:" + err.Error())
		}
		contractAddress, err := common.AddressParseFromBytes(addrbytes)
		if err != nil {
			return false, errors.NewErr("[callContract]get contract address error:" + err.Error())
		}

		result ,err = this.ContextRef.AppCall(contractAddress,util.TrimBuffToString(methodName),nil,arg)
		if err != nil {
			return false, errors.NewErr("[callContract]AppCall failed:" + err.Error())
		}
	}else{
		offchainContractCode, err := vm.GetPointerMemory(contractAddressIdx)
		if err != nil {
			return false, errors.NewErr("[callContract]get Contract address failed:" + err.Error())
		}

		codestring := util.TrimBuffToString(offchainContractCode)
		codeAddress := utils.GetContractAddress(codestring,stype.WASMVM)

		contractBytes,err := common.HexToBytes(util.TrimBuffToString(offchainContractCode))
		if err != nil {
			 return false ,err
		}
		result ,err = this.ContextRef.AppCall(codeAddress,util.TrimBuffToString(methodName),contractBytes,arg)
		if err != nil {
			return false, errors.NewErr("[callContract]AppCall failed:" + err.Error())
		}
	}

	vm.RestoreCtx()
	if envCall.GetReturns() {
		idx,err := vm.SetPointerMemory(result)
		if err != nil {
			return false, errors.NewErr("[callContract]SetPointerMemory failed:" + err.Error())
		}
		vm.PushResult(uint64(idx))
	}

	return true ,nil
}

func (this *WasmVmService) GetContractCodeFromAddress(address common.Address) ([]byte, error) {

	dcode, err := this.Store.GetContractState(address)
	if err != nil {
		return nil, err
	}

	if dcode == nil {
		return nil,errors.NewErr("[GetContractCodeFromAddress] deployed code is nil")
	}

	return dcode.Code.Code, nil

}

func (this *WasmVmService) getContractFromAddr(addr []byte) ([]byte, error) {
	addrbytes, err := common.HexToBytes(util.TrimBuffToString(addr))
	if err != nil {
		return nil, errors.NewErr("get contract address error")
	}
	contactaddress, err := common.AddressParseFromBytes(addrbytes)
	if err != nil {
		return nil, errors.NewErr("get contract address error")
	}
	dpcode, err := this.GetContractCodeFromAddress(contactaddress)
	if err != nil {
		return nil, errors.NewErr("get contract  error")
	}
	return dpcode, nil
}
