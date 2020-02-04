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

package ont_lock_proxy

import (
	"encoding/hex"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/cross_chain/cross_chain_manager"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
)

const (
	LOCK_NAME       = "lock"
	UNLOCK_NAME     = "unlock"
	BIND_PROXY_NAME = "bindProxy"
	BIND_ASSET_NAME = "bindAsset"
)

func AddLockNotifications(native *native.NativeService, contract common.Address, toContract []byte, state *LockParam) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          []interface{}{LOCK_NAME, state.FromAddress.ToBase58(), state.ToChainID, hex.EncodeToString(toContract), hex.EncodeToString(state.Args.ToAddress), state.Args.Value},
		})
}
func AddUnLockNotifications(native *native.NativeService, contract common.Address, fromChainId uint64, fromContract []byte, toAddress common.Address, amount uint64) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          []interface{}{UNLOCK_NAME, fromChainId, hex.EncodeToString(fromContract), toAddress.ToBase58(), amount},
		})
}

func getCreateTxArgs(toChainID uint64, contractHashBytes []byte, fee uint64, method string, argsBytes []byte) []byte {
	createCrossChainTxParam := &cross_chain_manager.CreateCrossChainTxParam{
		ToChainID:         toChainID,
		ToContractAddress: contractHashBytes,
		Fee:               fee,
		Method:            method,
		Args:              argsBytes,
	}
	sink := common.NewZeroCopySink(nil)
	createCrossChainTxParam.Serialization(sink)
	return sink.Bytes()
}

func getTransferInput(state ont.State) []byte {
	var transfers ont.Transfers
	transfers.States = []ont.State{state}
	sink := common.NewZeroCopySink(nil)
	transfers.Serialization(sink)
	return sink.Bytes()
}

func GenBindProxyKey(contract common.Address, chainId uint64) []byte {
	sink := common.NewZeroCopySink(nil)
	sink.WriteUint64(chainId)
	chainIdBytes := sink.Bytes()
	temp := append(contract[:], []byte(BIND_PROXY_NAME)...)
	return append(temp, chainIdBytes...)
}

func GenBindAssetKey(contract common.Address, assetContract []byte, chainId uint64) []byte {
	sink := common.NewZeroCopySink(nil)
	sink.WriteUint64(chainId)
	chainIdBytes := sink.Bytes()
	temp := append(contract[:], assetContract...)
	temp = append(temp, []byte(BIND_ASSET_NAME)...)
	return append(temp, chainIdBytes...)
}
