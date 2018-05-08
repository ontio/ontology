package exec

import (
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/vm/wasmvm/util"
	"strconv"
)

func strToInt(engine *ExecutionEngine) (bool, error) {
	envCall := engine.vm.envCall
	params := envCall.envParams
	if len(params) != 1 {
		return false, errors.NewErr("[jsonMashalParams]parameter count error")
	}

	addr := params[0]

	pBytes, err := engine.vm.GetPointerMemory(addr)
	if err != nil {
		return false, errors.NewErr("[jsonMashalParams] GetPointerMemory err:" + err.Error())
	}
	if pBytes == nil || len(pBytes) == 0 {
		engine.vm.ctx = envCall.envPreCtx
		if envCall.envReturns {
			engine.vm.pushUint64(uint64(0))
		}
		return true, nil
	}

	str := util.TrimBuffToString(pBytes)
	i, err := strconv.Atoi(str)
	if err != nil {
		return false, errors.NewErr("[jsonMashalParams] Atoi err:" + err.Error())
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
		return false, errors.NewErr("[jsonMashalParams]parameter count error")
	}

	addr := params[0]

	pBytes, err := engine.vm.GetPointerMemory(addr)
	if err != nil {
		return false, errors.NewErr("[jsonMashalParams] GetPointerMemory err:" + err.Error())
	}

	if pBytes == nil || len(pBytes) == 0 {
		engine.vm.ctx = envCall.envPreCtx
		if envCall.envReturns {
			engine.vm.pushUint64(uint64(0))
		}
		return true, nil
	}

	str := util.TrimBuffToString(pBytes)
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return false, errors.NewErr("[jsonMashalParams] Atoi err:" + err.Error())
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
		return false, errors.NewErr("[jsonMashalParams]parameter count error")
	}

	i := int(params[0])

	str := strconv.Itoa(i)

	idx, err := engine.vm.SetPointerMemory(str)
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
	if len(params) != 2 {
		return false, errors.NewErr("[jsonMashalParams]parameter count error")
	}
	i := int64(params[0])
	radix := int(params[1])
	str := strconv.FormatInt(i, radix)
	idx, err := engine.vm.SetPointerMemory(str)
	if err != nil {
		return false, err
	}

	engine.vm.ctx = envCall.envPreCtx
	if envCall.envReturns {
		engine.vm.pushUint64(uint64(idx))
	}
	return true, nil

}

func int64Add(engine *ExecutionEngine) (bool, error) {
	envCall := engine.vm.envCall
	params := envCall.envParams
	if len(params) != 2 {
		return false, errors.NewErr("[jsonMashalParams]parameter count error")
	}
	i := int64(params[0])
	j := int64(params[1])
	sum := i + j

	engine.vm.ctx = envCall.envPreCtx
	if envCall.envReturns {
		engine.vm.pushUint64(uint64(sum))
	}
	return true, nil
}

func int64Subtract(engine *ExecutionEngine) (bool, error) {
	envCall := engine.vm.envCall
	params := envCall.envParams
	if len(params) != 2 {
		return false, errors.NewErr("[jsonMashalParams]parameter count error")
	}
	i := int64(params[0])
	j := int64(params[1])
	sub := i - j

	engine.vm.ctx = envCall.envPreCtx
	if envCall.envReturns {
		engine.vm.pushUint64(uint64(sub))
	}
	return true, nil
}
