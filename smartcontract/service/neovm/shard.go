/*
 * Copyright (C) 2019 The ontology Authors
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
	vm "github.com/ontio/ontology/vm/neovm"
	"math/big"
)

// ShardGetShardId push shardId to vm stack
func ShardGetShardId(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, service.ShardID.ToUint64())
	return nil
}

func NotifyRemoteShard(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 5 {
		return fmt.Errorf("too few input parameters")
	}
	shardId, err := vm.PopBigInt(engine)
	if err != nil {
		return fmt.Errorf("read shardId failed, err: %s", err)
	}
	target, err := common.NewShardID(shardId.Uint64())
	if err != nil {
		return fmt.Errorf("parse shardId failed, err: %s", err)
	}
	addr, err := vm.PopByteArray(engine)
	if err != nil {
		return fmt.Errorf("read dest contract failed, err: %s", err)
	}
	contract, err := common.AddressParseFromBytes(addr)
	if err != nil {
		return fmt.Errorf("parse dest contract failed, err: %s", err)
	}
	fee, err := vm.PopBigInt(engine)
	if err != nil {
		return fmt.Errorf("read fee failed, err: %s", err)
	}
	if fee.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("fee must larger than 0")
	}
	method, err := vm.PopByteArray(engine)
	if err != nil {
		return fmt.Errorf("read method failed, err: %s", err)
	}
	args, err := vm.PopByteArray(engine)
	if err != nil {
		return fmt.Errorf("read args failed, err: %s", err)
	}
	service.ContextRef.NotifyRemoteShard(target, contract, fee.Uint64(), string(method), args)
	vm.PushData(engine, true)
	return nil
}

func InvokeRemoteShard(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 4 {
		return fmt.Errorf("too few input parameters")
	}
	shardId, err := vm.PopBigInt(engine)
	if err != nil {
		return fmt.Errorf("read shardId failed, err: %s", err)
	}
	target, err := common.NewShardID(shardId.Uint64())
	if err != nil {
		return fmt.Errorf("parse shardId failed, err: %s", err)
	}
	addr, err := vm.PopByteArray(engine)
	if err != nil {
		return fmt.Errorf("read dest contract failed, err: %s", err)
	}
	contract, err := common.AddressParseFromBytes(addr)
	if err != nil {
		return fmt.Errorf("parse dest contract failed, err: %s", err)
	}
	method, err := vm.PopByteArray(engine)
	if err != nil {
		return fmt.Errorf("read method failed, err: %s", err)
	}
	args, err := vm.PopByteArray(engine)
	if err != nil {
		return fmt.Errorf("read args failed, err: %s", err)
	}
	_, err = service.ContextRef.InvokeRemoteShard(target, contract, string(method), args)
	if err == nil {
		vm.PushData(engine, true)
	}
	return err
}
