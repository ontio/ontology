package exec

import (
	"bytes"
	"errors"
	"github.com/Ontology/vm/wasmvm/memory"
	"github.com/Ontology/vm/wasmvm/wasm"
	"github.com/Ontology/core/ledger"
	"github.com/Ontology/common"
	"github.com/Ontology/vm/types"
)

type IInteropService interface {
	Register(method string, handler func(*ExecutionEngine) (bool, error)) bool
	GetServiceMap() map[string]func(*ExecutionEngine) (bool, error)
}

type InteropService struct {
	serviceMap map[string]func(*ExecutionEngine) (bool, error)
}

func NewInteropService() *InteropService {
	service := InteropService{make(map[string]func(*ExecutionEngine) (bool, error))}

	//init some system functions
	service.Register("calloc", calloc)
	service.Register("malloc", malloc)
	service.Register("arrayLen", arrayLen)
	service.Register("memcpy", memcpy)
	service.Register("read_message", readMessage)
	service.Register("callContract", callContract)

	//===================add block apis below==================
	return &service
}

func (i *InteropService) Register(name string, handler func(*ExecutionEngine) (bool, error)) bool {
	if _, ok := i.serviceMap[name]; ok {
		return false
	}
	i.serviceMap[name] = handler
	return true
}

func (i *InteropService) Invoke(methodName string, engine *ExecutionEngine) (bool, error) {

	if v, ok := i.serviceMap[methodName]; ok {
		return v(engine)
	}
	return false, errors.New("Not supported method:" + methodName)
}

func (i *InteropService) MergeMap(mMap map[string]func(*ExecutionEngine) (bool, error)) bool {

	for k, v := range mMap {
		if _, ok := i.serviceMap[k]; !ok {
			i.serviceMap[k] = v
		}
	}
	return true
}

func (i *InteropService) Exists(name string) bool {
	_, ok := i.serviceMap[name]
	return ok
}
func (i *InteropService) GetServiceMap() map[string]func(*ExecutionEngine) (bool, error) {
	return i.serviceMap
}

//******************* basic functions ***************************
//TODO deside to replace the PUNKNOW type

//for the c language "calloc" function
func calloc(engine *ExecutionEngine) (bool, error) {

	envCall := engine.vm.envCall
	params := envCall.envParams

	if len(params) != 2 {
		return false, errors.New("parameter count error while call calloc")
	}
	count := int(params[0])
	length := int(params[1])
	//we don't know whats the alloc type here
	index, err := engine.vm.memory.MallocPointer(count*length, memory.P_UNKNOW)
	if err != nil {
		return false, err
	}

	//1. recover the vm context
	//2. if the call returns value,push the result to the stack
	engine.vm.ctx = envCall.envPreCtx
	if envCall.envReturns {
		engine.vm.pushUint64(uint64(index))
	}
	return true, nil
}

//for the c language "malloc" function
func malloc(engine *ExecutionEngine) (bool, error) {
	envCall := engine.vm.envCall
	params := envCall.envParams
	if len(params) != 1 {
		return false, errors.New("parameter count error while call calloc")
	}
	size := int(params[0])
	//we don't know whats the alloc type here
	index, err := engine.vm.memory.MallocPointer(size, memory.P_UNKNOW)
	if err != nil {
		return false, err
	}
	//1. recover the vm context
	//2. if the call returns value,push the result to the stack
	engine.vm.ctx = envCall.envPreCtx
	if envCall.envReturns {
		engine.vm.pushUint64(uint64(index))
	}
	return true, nil

}

//use arrayLen to replace 'sizeof'
func arrayLen(engine *ExecutionEngine) (bool, error) {
	envCall := engine.vm.envCall
	params := envCall.envParams
	if len(params) != 1 {
		return false, errors.New("parameter count error while call arrayLen")
	}

	pointer := params[0]

	tl, ok := engine.vm.memory.MemPoints[pointer]

	var result uint64
	if ok {
		switch tl.Ptype {
		case memory.P_INT8, memory.P_STRING:
			result = uint64(tl.Length / 1)
		case memory.P_INT16:
			result = uint64(tl.Length / 2)
		case memory.P_INT32, memory.P_FLOAT32:
			result = uint64(tl.Length / 4)
		case memory.P_INT64, memory.P_FLOAT64:
			result = uint64(tl.Length / 8)
		case memory.P_UNKNOW:
			//todo assume it's byte
			result = uint64(tl.Length / 1)
		default:
			result = uint64(0)
		}

	} else {
		result = uint64(0)
	}
	//1. recover the vm context
	//2. if the call returns value,push the result to the stack
	engine.vm.ctx = envCall.envPreCtx
	if envCall.envReturns {
		engine.vm.pushUint64(uint64(result))
	}
	return true, nil

}

func memcpy(engine *ExecutionEngine) (bool, error) {
	envCall := engine.vm.envCall
	params := envCall.envParams
	if len(params) != 3 {
		return false, errors.New("parameter count error while call memcpy")
	}
	dest := int(params[0])
	src := int(params[1])
	length := int(params[2])

	if dest < src && dest+length > src {
		return false, errors.New("memcpy overlapped")
	}

	copy(engine.vm.memory.Memory[dest:dest+length], engine.vm.memory.Memory[src:src+length])

	//1. recover the vm context
	//2. if the call returns value,push the result to the stack
	engine.vm.ctx = envCall.envPreCtx
	if envCall.envReturns {
		engine.vm.pushUint64(uint64(1))
	}

	return true, nil //this return will be dropped in wasm
}

func readMessage(engine *ExecutionEngine) (bool, error) {
	envCall := engine.vm.envCall
	params := envCall.envParams
	if len(params) != 2 {
		return false, errors.New("parameter count error while call readMessage")
	}

	addr := int(params[0])
	length := int(params[1])

	msgBytes, err := engine.vm.GetMessageBytes()
	if err != nil {
		return false, err
	}
	if length != len(msgBytes) {
		return false, errors.New("readMessage length error")
	}
	copy(engine.vm.memory.Memory[addr:addr+length], msgBytes[:length])
	engine.vm.memory.MemPoints[uint64(addr)] = &memory.TypeLength{Ptype: memory.P_UNKNOW, Length: length}

	//1. recover the vm context
	//2. if the call returns value,push the result to the stack
	engine.vm.ctx = envCall.envPreCtx
	if envCall.envReturns {
		engine.vm.pushUint64(uint64(length))
	}

	return true, nil
}

//call other contract
func callContract(engine *ExecutionEngine) (bool, error) {
	envCall := engine.vm.envCall
	params := envCall.envParams
	if len(params) < 2 {
		return false, errors.New("parameter count error while call readMessage")
	}
	contractAddressIdx := params[0]
	addr, err := engine.vm.GetPointerMemory(contractAddressIdx)
	if err != nil {
		return false, errors.New("get Contract address failed")
	}
	//the contract codes
	contractBytes, err := getContractFromAddr(addr)
	if err != nil {
		return false, err
	}
	bf := bytes.NewBuffer(contractBytes)
	module, err := wasm.ReadModule(bf, emptyImporter)
	if err != nil {
		return false, errors.New("load Module failed")
	}

	methodName, err := engine.vm.GetPointerMemory(params[1])
	if err != nil {
		return false, errors.New("get Contract address failed")
	}
	//if has args
	var args []uint64
	if len(params) > 2 {
		args = params[2:]
	}

	res, err := engine.vm.CallContract(module, trimBuffToString(methodName), args...)
	if err != nil {
		return false, errors.New("call contract " + trimBuffToString(methodName) + " failed")
	}
	//engine.vm.RestoreStat()
	if envCall.envReturns {
		engine.vm.pushUint64(res)
	}

	return true, nil
}

func emptyImporter(name string) (*wasm.Module, error) {
	return nil, nil
}

func getContractFromAddr(addr []byte) ([]byte, error) {

	//todo get the contract code from ledger
	//just for test
/*		contract := trimBuffToString(addr)
		code, err := ioutil.ReadFile(fmt.Sprintf("./testdata2/%s.wasm",contract))
		if err != nil {
			fmt.Printf("./testdata2/%s.wasm is not exist",contract)
			return nil,err
		}

		return code,nil*/
	codeHash, err := common.Uint160ParseFromBytes(addr)
	if err != nil {
		return nil, errors.New("get address Code hash failed")
	}

	contract, err := ledger.DefLedger.GetContractState(codeHash)
	if err != nil {
		return nil, errors.New("get contract state failed")
	}

	if contract.VmType != types.WASMVM {
		return nil, errors.New(" contract is not a wasm contract")
	}

	return contract.Code, nil

}

//trim the '\00' byte
func trimBuffToString(bytes []byte) string {

	for i, b := range bytes {
		if b == 0 {
			return string(bytes[:i])
		}
	}
	return string(bytes)

}
