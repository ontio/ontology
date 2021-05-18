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

package system

import (
	"fmt"
	"math/big"

	config2 "github.com/ontio/ontology/common/config"

	common2 "github.com/ethereum/go-ethereum/common"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/event"
	evm2 "github.com/ontio/ontology/smartcontract/service/evm"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/ong"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/storage"
	"github.com/ontio/ontology/vm/evm"
	"github.com/ontio/ontology/vm/evm/params"
)

const (
	EvmInvokeName = "evmInvoke"
)

func InitSystem() {
	native.Contracts[utils.SystemContractAddress] = RegisterSystemContract
}

func RegisterSystemContract(native *native.NativeService) {
	native.Register(EvmInvokeName, EVMInvoke)
}

func EVMInvoke(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)

	caller, err := utils.DecodeAddress(source)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("evm invoke decode contract address error: %v", err)
	}
	target, err := utils.DecodeAddress(source)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("evm invoke decode contract address error: %v", err)
	}
	input, err := utils.DecodeVarBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("evm invoke decode input error: %v", err)
	}

	if !native.ContextRef.CheckWitness(caller) {
		return utils.BYTE_FALSE, fmt.Errorf("evm invoke error: verify witness failed for caller: %s", caller.ToBase58())
	}

	// Create a new context to be used in the EVM environment
	blockContext := evm2.NewEVMBlockContext(native.Height, native.Time, native.Store)
	gasLeft, gasPrice := native.ContextRef.GetGasInfo()
	txctx := evm.TxContext{
		Origin:   common2.Address(native.Tx.Payer),
		GasPrice: big.NewInt(0).SetUint64(gasPrice),
	}
	statedb := storage.NewStateDB(native.CacheDB, common2.Hash(native.Tx.Hash()), common2.Hash(native.BlockHash), ong.OngBalanceHandle{})
	config := params.GetChainConfig(config2.DefConfig.P2PNode.EVMChainId)
	vmenv := evm.NewEVM(blockContext, txctx, statedb, config, evm.Config{})

	callerCtx := native.ContextRef.CallingContext()
	if callerCtx == nil {
		return utils.BYTE_FALSE, fmt.Errorf("evm invoke must have a caller")
	}
	ret, leftGas, err := vmenv.Call(evm.AccountRef(caller), common2.Address(target), input, gasLeft, big.NewInt(0))
	gasUsed := gasLeft - leftGas
	refund := gasUsed / 2
	if refund > statedb.GetRefund() {
		refund = statedb.GetRefund()
	}
	gasUsed -= refund
	enoughGas := native.ContextRef.CheckUseGas(gasUsed)
	if !enoughGas {
		return utils.BYTE_FALSE, fmt.Errorf("evm invoke out of gas")
	}
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("invoke evm error:%s, return: %x", err, ret)
	}

	for _, log := range statedb.GetLogs() {
		native.Notifications = append(native.Notifications, event.NotifyEventInfoFromEvmLog(log))
	}

	err = statedb.CommitToCacheDB()
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	return ret, nil
}
