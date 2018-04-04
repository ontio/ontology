package exec

import (
	"strconv"
	"errors"
	"github.com/ontio/ontology/vm/wasmvm/util"
)

func strToInt(engine *ExecutionEngine) (bool, error) {
	envCall := engine.vm.envCall
	params := envCall.envParams
	if len(params) != 1 {
		return false, errors.New("[jsonMashalParams]parameter count error")
	}

	addr := params[0]

	pBytes, err := engine.vm.GetPointerMemory(addr)
	if err != nil {
		return false, errors.New("[jsonMashalParams] GetPointerMemory err:" + err.Error())
	}

	str := util.TrimBuffToString(pBytes)
	i,err := strconv.Atoi(str)
	if err != nil{
		return false , errors.New("[jsonMashalParams] Atoi err:" + err.Error())
	}

	engine.vm.ctx = envCall.envPreCtx
	if envCall.envReturns {
		engine.vm.pushUint64(uint64(i))
	}
	return true, nil

}

func strToInt64(engine *ExecutionEngine) (bool, error) {
	envCall := engine.vm.envCall
	params := envCall.envParams
	if len(params) != 1 {
		return false, errors.New("[jsonMashalParams]parameter count error")
	}

	addr := params[0]

	pBytes, err := engine.vm.GetPointerMemory(addr)
	if err != nil {
		return false, errors.New("[jsonMashalParams] GetPointerMemory err:" + err.Error())
	}

	str := util.TrimBuffToString(pBytes)
	i,err := strconv.ParseInt(str,10,64)
	if err != nil{
		return false , errors.New("[jsonMashalParams] Atoi err:" + err.Error())
	}

	engine.vm.ctx = envCall.envPreCtx
	if envCall.envReturns {
		engine.vm.pushUint64(uint64(i))
	}
	return true, nil

}

func intToString(engine *ExecutionEngine) (bool, error) {
	envCall := engine.vm.envCall
	params := envCall.envParams
	if len(params) != 1 {
		return false, errors.New("[jsonMashalParams]parameter count error")
	}

	i := int(params[0])

	str := strconv.Itoa(i)

	idx,err := engine.vm.SetPointerMemory(str)
	if err != nil {
		return false, err
	}

	engine.vm.ctx = envCall.envPreCtx
	if envCall.envReturns {
		engine.vm.pushUint64(uint64(idx))
	}
	return true, nil

}


func int64ToString(engine *ExecutionEngine) (bool, error) {
	envCall := engine.vm.envCall
	params := envCall.envParams
	if len(params) != 1 {
		return false, errors.New("[jsonMashalParams]parameter count error")
	}

	i := int64(params[0])

	str := strconv.FormatInt(i,10)

	idx,err := engine.vm.SetPointerMemory(str)
	if err != nil {
		return false, err
	}

	engine.vm.ctx = envCall.envPreCtx
	if envCall.envReturns {
		engine.vm.pushUint64(uint64(idx))
	}
	return true, nil

}
