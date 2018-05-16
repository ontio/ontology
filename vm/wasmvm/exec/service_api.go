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
