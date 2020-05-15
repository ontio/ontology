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

package neovm

import (
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/states"
	vm "github.com/ontio/ontology/vm/neovm"
)

func NativeInvoke(service *NeoVmService, engine *vm.Executor) error {
	version, err := engine.EvalStack.PopAsInt64()
	if err != nil {
		return err
	}
	address, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return err
	}
	addr, err := common.AddressParseFromBytes(address)
	if err != nil {
		return fmt.Errorf("invoke native contract:%s, address invalid", address)
	}
	method, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return err
	}
	if len(method) > METHOD_LENGTH_LIMIT {
		return fmt.Errorf("invoke native contract:%s method:%s too long, over max length 1024 limit", address, method)
	}
	args, err := engine.EvalStack.Pop()
	if err != nil {
		return err
	}
	sink := new(common.ZeroCopySink)
	if err := args.BuildParamToNative(sink); err != nil {
		return err
	}

	contract := states.ContractInvokeParam{
		Version: byte(version),
		Address: addr,
		Method:  string(method),
		Args:    sink.Bytes(),
	}

	nat := &native.NativeService{
		Store:       service.Store,
		CacheDB:     service.CacheDB,
		InvokeParam: contract,
		Tx:          service.Tx,
		Height:      service.Height,
		Time:        service.Time,
		ContextRef:  service.ContextRef,
		ServiceMap:  make(map[string]native.Handler),
		PreExec:     service.PreExec,
	}

	result, err := nat.Invoke()
	if err != nil {
		return err
	}
	return engine.EvalStack.PushBytes(result)
}
