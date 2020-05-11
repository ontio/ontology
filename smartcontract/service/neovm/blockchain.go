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
	"math/big"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/errors"
	vm "github.com/ontio/ontology/vm/neovm"
)

// BlockChainGetHeight put blockchain's height to vm stack
func BlockChainGetHeight(service *NeoVmService, engine *vm.Executor) error {
	return engine.EvalStack.PushUint32(service.Height - 1)
}

func BlockChainGetHeightNew(service *NeoVmService, engine *vm.Executor) error {
	return engine.EvalStack.PushUint32(service.Height)
}

func BlockChainGetHeaderNew(service *NeoVmService, engine *vm.Executor) error {
	data, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return err
	}
	b := common.BigIntFromNeoBytes(data)
	if b.Cmp(big.NewInt(int64(service.Height))) != 0 {
		return errors.NewErr("can only get current block header")
	}

	header := &HeaderValue{Height: service.Height, Timestamp: service.Time, Hash: service.BlockHash}
	return engine.EvalStack.PushAsInteropValue(header)
}

// BlockChainGetContract put blockchain's contract to vm stack
func BlockChainGetContract(service *NeoVmService, engine *vm.Executor) error {
	b, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return err
	}
	address, err := common.AddressParseFromBytes(b)
	if err != nil {
		return err
	}
	item, err := service.Store.GetContractState(address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetContract] GetContract error!")
	}
	err = engine.EvalStack.PushAsInteropValue(item)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetContract] PushAsInteropValue error!")
	}
	return nil
}
