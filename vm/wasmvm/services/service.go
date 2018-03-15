package services
/*

import (
	"errors"
	"fmt"
	"github.com/Ontology/vm/wasmvm/memory"
)

type EnvHandler func(*memory.VMmemory, []uint64) (uint64,error)

type InterOpService struct {
	serviceMap map[string]EnvHandler
}

func NewInterOpService() *InterOpService {
	i := &InterOpService{serviceMap: make(map[string]EnvHandler)}
	i.Register("addOne", func(mem *memory.VMmemory, p []uint64) (uint64,error) { return p[0] + 1,nil })
	i.Register("getBlockHeight", blockChainGetHeight)
	i.Register("getString", i.GetString)
	i.Register("getAcct",getAddress)
	i.Register("getBlockHash",blockChainGetBlockHash)
	//calloc and malloc
	i.Register("calloc",calloc)
	i.Register("malloc",malloc)
	i.Register("memcpy",memcpy)
	i.Register("arrayLen",arrayLen)
	i.Register("read_message",readMessage)


	return i
}

func (i *InterOpService) Register(name string, handler EnvHandler) bool {
	if _, ok := i.serviceMap[name]; ok {
		return false
	}
	i.serviceMap[name] = handler
	return true
}

func (i *InterOpService) Invoke(methodName string, params []uint64, mem *memory.VMmemory) (uint64, error) {

	if v, ok := i.serviceMap[methodName]; ok {
		return v(mem, params)
	}
	return uint64(0), errors.New("Not supported method:" + methodName)
}

func (i *InterOpService) MergeMap(mMap map[string]EnvHandler) bool {

	for k, v := range mMap {
		if _, ok := i.serviceMap[k]; !ok {
			i.serviceMap[k] = v
		}
	}
	return true
}

func (i *InterOpService) Exists(name string) bool {
	_, ok := i.serviceMap[name]
	return ok
}

//todo modify with the real function
func (i *InterOpService) GetBlockHeight(mem *memory.VMmemory, param []uint64) uint64 {
	return uint64(108)
}

func (i *InterOpService) GetString(memory *memory.VMmemory, param []uint64) (uint64,error) {
	mem := memory.Memory
	if len(mem) == 0 || len(param) != 2 || len(mem) < int(param[0]+param[1]) {
		return uint64(0),errors.New("parameter is not right")
	}
	start := int(param[0])
	length := int(param[1])
	fmt.Printf("start is %d,length is %d\n", start, length)
	str := string(mem[start : start+length])

	fmt.Println("string is :" + str)
	return uint64(len(str)),nil
}
*/
