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
package common

import (
	"fmt"
	"math/big"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	ctypes "github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/neovm"
	ntypes "github.com/ontio/ontology/vm/neovm/types"
)

func CrossChainNeoVMCall(this *native.NativeService, address common.Address, method string, args []byte,
	fromContractAddress []byte, fromChainID uint64) (interface{}, error) {
	dep, err := this.CacheDB.GetContract(address)
	if err != nil {
		return nil, errors.NewErr("[NeoVMCall] Get contract context error!")
	}
	log.Debugf("[NeoVMCall] native invoke neovm contract address:%s", address.ToHexString())
	if dep == nil {
		return nil, errors.NewErr("[NeoVMCall] native invoke neovm contract is nil")
	}
	m, err := ntypes.VmValueFromBytes([]byte(method))
	if err != nil {
		return nil, err
	}
	array := ntypes.NewArrayValue()
	a, err := ntypes.VmValueFromBytes(args)
	if err != nil {
		return nil, err
	}
	if err := array.Append(a); err != nil {
		return nil, err
	}
	fca, err := ntypes.VmValueFromBytes(fromContractAddress)
	if err != nil {
		return nil, err
	}
	if err := array.Append(fca); err != nil {
		return nil, err
	}
	fci, err := ntypes.VmValueFromBigInt(new(big.Int).SetUint64(fromChainID))
	if err != nil {
		return nil, err
	}
	if err := array.Append(fci); err != nil {
		return nil, err
	}
	if !this.ContextRef.CheckUseGas(neovm.NATIVE_INVOKE_GAS) {
		return nil, fmt.Errorf("[CrossChainNeoVMCall], check use gaslimit insufficient！")
	}
	engine, err := this.ContextRef.NewExecuteEngine(dep.GetRawCode(), ctypes.InvokeNeo)
	if err != nil {
		return nil, err
	}
	evalStack := engine.(*neovm.NeoVmService).Engine.EvalStack
	if err := evalStack.Push(ntypes.VmValueFromArrayVal(array)); err != nil {
		return nil, err
	}
	if err := evalStack.Push(m); err != nil {
		return nil, err
	}
	return engine.Invoke()
}
