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

package native

import (
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/context"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/states"
	sstates "github.com/ontio/ontology/smartcontract/states"
	"github.com/ontio/ontology/smartcontract/storage"
	ntypes "github.com/ontio/ontology/vm/neovm/types"
)

type (
	Handler         func(native *NativeService) ([]byte, error)
	RegisterService func(native *NativeService)
)

var (
	Contracts = make(map[common.Address]RegisterService)
)

// Native service struct
// Invoke a native smart contract, new a native service
type NativeService struct {
	CacheDB       *storage.CacheDB
	ServiceMap    map[string]Handler
	Notifications []*event.NotifyEventInfo
	InvokeParam   sstates.ContractInvokeParam
	Input         []byte
	Tx            *types.Transaction
	ShardID       types.ShardID
	ParentShardID uint64
	Height        uint32
	Time          uint32
	BlockHash     common.Uint256
	ContextRef    context.ContextRef
}

func (this *NativeService) Register(methodName string, handler Handler) {
	this.ServiceMap[methodName] = handler
}

func (this *NativeService) Invoke() (interface{}, error) {
	contract := this.InvokeParam
	services, ok := Contracts[contract.Address]
	if !ok {
		return false, fmt.Errorf("Native contract address %x haven't been registered.", contract.Address)
	}
	services(this)
	service, ok := this.ServiceMap[contract.Method]
	if !ok {
		return false, fmt.Errorf("Native contract %x doesn't support this function %s.",
			contract.Address, contract.Method)
	}
	args := this.Input
	this.Input = contract.Args
	this.ContextRef.PushContext(&context.Context{ContractAddress: contract.Address})
	notifications := this.Notifications
	this.Notifications = []*event.NotifyEventInfo{}
	result, err := service(this)
	if err != nil {
		return result, errors.NewDetailErr(err, errors.ErrNoCode, "[Invoke] Native serivce function execute error!")
	}
	this.ContextRef.PopContext()
	this.ContextRef.PushNotifications(this.Notifications)
	this.Notifications = notifications
	this.Input = args
	return result, nil
}

func (this *NativeService) NativeCall(address common.Address, method string, args []byte) (interface{}, error) {
	c := states.ContractInvokeParam{
		Address: address,
		Method:  method,
		Args:    args,
	}
	this.InvokeParam = c
	return this.Invoke()
}

func (this *NativeService) NeoVMCall(address common.Address, method string, args []byte) (interface{}, error) {
	dep, err := this.CacheDB.GetContract(address)
	if err != nil {
		return nil, errors.NewErr("[NeoVMCall] Get contract context error!")
	}
	log.Debugf("[NeoVMCall] native invoke neovm contract address:%s", address.ToHexString())
	if dep == nil {
		return nil, errors.NewErr("[NeoVMCall] native invoke neovm contract is nil")
	}
	stacks := make([]ntypes.StackItems, 2)
	stacks[0] = ntypes.NewByteArray([]byte(method))
	stacks[1] = ntypes.NewArray([]ntypes.StackItems{ntypes.NewByteArray(args)})
	engine, err := this.ContextRef.NewExecuteEngineWithElem(dep.Code, stacks)
	if err != nil {
		return nil, err
	}
	return engine.Invoke()
}
