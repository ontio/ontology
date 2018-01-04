package neovm

import (
	. "github.com/Ontology/vm/neovm/errors"
	"github.com/Ontology/common/log"
	"github.com/Ontology/common"
)

type IInteropService interface {
	Register(method string, handler func(*ExecutionEngine) (bool, error)) bool
	GetServiceMap() map[string]func(*ExecutionEngine) (bool, error)
}

type InteropService struct {
	serviceMap map[string]func(*ExecutionEngine) (bool, error)
}

func NewInteropService() *InteropService {
	var i InteropService
	i.serviceMap = make(map[string]func(*ExecutionEngine) (bool, error), 0)
	i.Register("System.ExecutionEngine.GetScriptContainer", i.GetCodeContainer)
	i.Register("System.ExecutionEngine.GetExecutingScriptHash", i.GetExecutingCodeHash)
	i.Register("System.ExecutionEngine.GetCallingScriptHash", i.GetCallingCodeHash)
	i.Register("System.ExecutionEngine.GetEntryScriptHash", i.GetEntryCodeHash)
	return &i
}

func (is *InteropService) Register(methodName string, handler func(*ExecutionEngine) (bool, error)) bool {
	if _, ok := is.serviceMap[methodName]; ok {
		return false
	}
	is.serviceMap[methodName] = handler
	return true
}

func (i *InteropService) MergeMap(dictionary map[string]func(*ExecutionEngine) (bool, error)) {
	for k, v := range dictionary {
		if _, ok := i.serviceMap[k]; !ok {
			i.serviceMap[k] = v
		}
	}
}

func (i *InteropService) GetServiceMap() map[string]func(*ExecutionEngine) (bool, error) {
	return i.serviceMap
}

func (i *InteropService) Invoke(methodName string, engine *ExecutionEngine) (bool, error) {
	if v, ok := i.serviceMap[methodName]; ok {
		log.Error("Invoke MethodName:", methodName)
		return v(engine)
	}
	return false, ErrNotSupportService
}

func (i *InteropService) GetCodeContainer(engine *ExecutionEngine) (bool, error) {
	PushData(engine, engine.codeContainer)
	return true, nil
}

func (i *InteropService) GetExecutingCodeHash(engine *ExecutionEngine) (bool, error) {
	code, err := engine.ExecutingCode()
	if err != nil {
		return false, err
	}
	codeHash, err := common.ToCodeHash(code)
	if err != nil {
		return false, err
	}
	PushData(engine, &codeHash)
	return true, nil
}

func (i *InteropService) GetCallingCodeHash(engine *ExecutionEngine) (bool, error) {
	context, err := engine.CallingContext()
	if err != nil {
		return false, err
	}
	codeHash, err := context.GetCodeHash()
	if err != nil {
		return false, err
	}
	PushData(engine, &codeHash)
	return true, nil
}
func (i *InteropService) GetEntryCodeHash(engine *ExecutionEngine) (bool, error) {
	context, err := engine.EntryContext()
	if err != nil {
		return false, err
	}
	codeHash, err := context.GetCodeHash()
	if err != nil {
		return false, err
	}
	PushData(engine, &codeHash)
	return true, nil
}
