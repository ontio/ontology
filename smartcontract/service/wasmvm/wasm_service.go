package wasmvm

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"math/big"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/store"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	sccommon "github.com/ontio/ontology/smartcontract/common"
	"github.com/ontio/ontology/smartcontract/context"
	"github.com/ontio/ontology/smartcontract/event"
	nstates "github.com/ontio/ontology/smartcontract/service/native/states"
	"github.com/ontio/ontology/smartcontract/states"
	"github.com/ontio/ontology/smartcontract/storage"
	vmtypes "github.com/ontio/ontology/smartcontract/types"
	"github.com/ontio/ontology/vm/wasmvm/exec"
	"github.com/ontio/ontology/vm/wasmvm/util"
	"fmt"
)

type WasmVmService struct {
	Store         store.LedgerStore
	CloneCache    *storage.CloneCache
	ContextRef    context.ContextRef
	Notifications []*event.NotifyEventInfo
	Code          []byte
	Tx            *types.Transaction
	Time          uint32
}

func (this *WasmVmService) Invoke() (interface{}, error) {
	stateMachine := NewWasmStateMachine(this.Store, this.CloneCache, this.Time)
	//register the "CallContract" function
	stateMachine.Register("CallContract", this.callContract)
	stateMachine.Register("MarshalNativeParams", this.marshalNativeParams)
	stateMachine.Register("CheckWitness", this.CheckWitness)
	engine := exec.NewExecutionEngine(
		this.Tx,
		new(util.ECDsaCrypto),
		stateMachine,
	)

	contract := &states.Contract{}
	contract.Deserialize(bytes.NewBuffer(this.Code))
	addr := contract.Address

	if contract.Code == nil {
		dpcode, err := this.GetContractCodeFromAddress(addr)
		if err != nil {
			return nil, errors.NewErr("get contract  error")
		}
		contract.Code = dpcode
	}

	var caller common.Address
	if this.ContextRef.CallingContext() == nil {
		caller = common.Address{}
	} else {
		caller = this.ContextRef.CallingContext().ContractAddress
	}
	this.ContextRef.PushContext(&context.Context{ContractAddress: contract.Address})
	res, err := engine.Call(caller, contract.Code, contract.Method, contract.Args, contract.Version)

	if err != nil {
		return nil, err
	}

	//get the return message
	result, err := engine.GetVM().GetPointerMemory(uint64(binary.LittleEndian.Uint32(res)))
	if err != nil {
		return nil, err
	}
	this.ContextRef.PopContext()
	this.ContextRef.PushNotifications(stateMachine.Notifications)
	return result, nil
}

// marshalNativeParams
// make paramter bytes for call native contract
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
	//	Version byte               -------->i32  4 bytes
	//	States  []*State		   -------->i32 pointer 4 bytes
	//}
	if len(transferbytes) != 8 {
		return false, errors.NewErr("[callContract]parameter format error while call marshalNativeParams")
	}
	transfer := &nstates.Transfers{}
	tver := binary.LittleEndian.Uint32(transferbytes[:4])
	transfer.Version = byte(tver)

	statesAddr := binary.LittleEndian.Uint32(transferbytes[4:])
	statesbytes, err := vm.GetPointerMemory(uint64(statesAddr))
	if err != nil {
		return false, err
	}
	//statesbytes is slice of struct with states.
	//type State struct {
	//	Version byte            -------->i32 4 bytes
	//	From    common.Address  -------->i32 pointer 4 bytes
	//	To      common.Address  -------->i32 pointer 4 bytes extra padding 4 bytes
	//	Value   *big.Int        -------->i64 8 bytes
	//}
	//total is 4 + 4 + 4 + 4(dummy) + 8 = 24 bytes
	statecnt := len(statesbytes) / 24
	states := make([]*nstates.State, statecnt)

	for i := 0; i < statecnt; i++ {
		tmpbytes := statesbytes[i * 24 : (i + 1) * 24]
		state := &nstates.State{}
		state.Version = byte(binary.LittleEndian.Uint32(tmpbytes[:4]))
		fromAddessBytes, err := vm.GetPointerMemory(uint64(binary.LittleEndian.Uint32(tmpbytes[4:8])))
		if err != nil {
			return false, err
		}
		fromAddress, err := common.AddressFromBase58(util.TrimBuffToString(fromAddessBytes))
		if err != nil {
			return false, err
		}
		state.From = fromAddress

		toAddressBytes, err := vm.GetPointerMemory(uint64(binary.LittleEndian.Uint32(tmpbytes[8:12])))
		if err != nil {
			return false, err
		}
		toAddress, err := common.AddressFromBase58(util.TrimBuffToString(toAddressBytes))
		state.To = toAddress
		//tmpbytes[12:16] is padding
		amount := binary.LittleEndian.Uint64(tmpbytes[16:])
		state.Value = big.NewInt(int64(amount))
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

func (this *WasmVmService) CheckWitness(engine *exec.ExecutionEngine) (bool, error) {
	fmt.Println("=====CheckWitness start======")
	vm := engine.GetVM()

	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false ,errors.NewErr("[CheckWitness]get parameter count error!")
	}

	data,err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false ,errors.NewErr("[CheckWitness]" + err.Error())
	}
	address, err := common.AddressFromBase58(util.TrimBuffToString(data))
	if err != nil {
		return false ,errors.NewErr("[CheckWitness]" + err.Error())
	}
	chkRes := this.ContextRef.CheckWitness(address)

	res := 0
	if chkRes == true{
		res = 1
	}
	fmt.Printf("=====CheckWitness res is %d======",res)
	vm.RestoreCtx()
	if vm.GetEnvCall().GetReturns() {
		vm.PushResult(uint64(res))
	}
	return true, nil
}




// callContract will need 4 paramters
//0: contract address
//1: contract code
//2: method name
//3: args
func (this *WasmVmService) callContract(engine *exec.ExecutionEngine) (bool, error) {

	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 4 {
		return false, errors.NewErr("[callContract]parameter count error while call readMessage")
	}

	var contractAddress common.Address
	var contractBytes []byte
	//get contract address
	contractAddressIdx := params[0]
	addr, err := vm.GetPointerMemory(contractAddressIdx)
	if err != nil {
		return false, errors.NewErr("[callContract]get Contract address failed:" + err.Error())
	}

	if addr != nil {
		addrbytes, err := common.HexToBytes(util.TrimBuffToString(addr))
		if err != nil {
			return false, errors.NewErr("[callContract]get contract address error:" + err.Error())
		}
		contractAddress, err = common.AddressParseFromBytes(addrbytes)
		if err != nil {
			return false, errors.NewErr("[callContract]get contract address error:" + err.Error())
		}

	}

	//get contract code

	codeIdx := params[1]

	offchainContractCode, err := vm.GetPointerMemory(codeIdx)
	if err != nil {
		return false, errors.NewErr("[callContract]get Contract address failed:" + err.Error())
	}

	if offchainContractCode != nil {
		contractBytes, err = common.HexToBytes(util.TrimBuffToString(offchainContractCode))
		if err != nil {
			return false, err

		}
		//compute the offchain code address
		codestring := util.TrimBuffToString(offchainContractCode)
		contractAddress = GetContractAddress(codestring, vmtypes.WASMVM)
	}

	//get method
	methodName, err := vm.GetPointerMemory(params[2])
	if err != nil {
		return false, errors.NewErr("[callContract]get Contract methodName failed:" + err.Error())
	}

	//get args
	arg, err := vm.GetPointerMemory(params[3])
	if err != nil {
		return false, errors.NewErr("[callContract]get Contract arg failed:" + err.Error())
	}

	result, err := this.ContextRef.AppCall(contractAddress, util.TrimBuffToString(methodName), contractBytes, arg)
	if err != nil {
		return false, errors.NewErr("[callContract]AppCall failed:" + err.Error())
	}
	vm.RestoreCtx()
	if envCall.GetReturns() {
		if contractAddress[0] == byte(vmtypes.NEOVM) {
			result = sccommon.ConvertNeoVmReturnTypes(result)
		}
		if contractAddress[0] == byte(vmtypes.Native) {
			bresult := result.(bool)
			if bresult == true {
				result = "true"
			} else {
				result = false
			}

		}
		if contractAddress[0] == byte(vmtypes.WASMVM) {
			//reserve for further process
		}

		idx, err := vm.SetPointerMemory(result.(string))
		if err != nil {
			return false, errors.NewErr("[callContract]SetPointerMemory failed:" + err.Error())
		}
		vm.PushResult(uint64(idx))
	}

	return true, nil
}

func (this *WasmVmService) GetContractCodeFromAddress(address common.Address) ([]byte, error) {

	dcode, err := this.Store.GetContractState(address)
	if err != nil {
		return nil, err
	}

	if dcode == nil {
		return nil, errors.NewErr("[GetContractCodeFromAddress] deployed code is nil")
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

//GetContractAddress return contract address
func GetContractAddress(code string, vmType vmtypes.VmType) common.Address {
	data, _ := hex.DecodeString(code)
	vmCode := &vmtypes.VmCode{
		VmType: vmType,
		Code:   data,
	}
	return vmCode.AddressFromVmCode()
}
