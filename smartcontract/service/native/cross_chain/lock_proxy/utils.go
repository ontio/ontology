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

package lock_proxy

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/cross_chain/cross_chain_manager"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	LOCK_NAME               = "lock"
	UNLOCK_NAME             = "unlock"
	BIND_PROXY_NAME         = "bindProxy"
	BIND_ASSET_NAME         = "bindAsset"
	WITHDRAW_ONG_NAME       = "withdrawONG"
	GET_PROXY_HASH_NAME     = "getProxyHash"
	GET_ASSET_HASH_NAME     = "getAssetHash"
	GET_CROSSED_LIMIT_NAME  = "getCrossedLimit"
	GET_CROSSED_AMOUNT_NAME = "getCrossedAmount"

	TARGET_ASSET_HASH_PEFIX = "TargetAssetHash"
	CROSS_LIMIT_PREFIX      = "AssetCrossLimit"
	CROSS_AMOUNT_PREFIX     = "AssetCrossedAmount"
)

func AddLockNotifications(native *native.NativeService, contract, sourceAssetAddress common.Address, toChainId uint64, toContract []byte, targetAssetHash []byte, fromAddress common.Address, toAddress []byte, amount uint64) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          []interface{}{LOCK_NAME, hex.EncodeToString(sourceAssetAddress[:]), toChainId, hex.EncodeToString(toContract), hex.EncodeToString(targetAssetHash), fromAddress.ToBase58(), hex.EncodeToString(toAddress), amount},
		})
}
func AddUnLockNotifications(native *native.NativeService, contract common.Address, fromChainId uint64, fromProxyContract []byte, targetAssetHash common.Address, toAddress common.Address, amount uint64) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          []interface{}{UNLOCK_NAME, fromChainId, hex.EncodeToString(fromProxyContract), hex.EncodeToString(targetAssetHash[:]), toAddress.ToBase58(), amount},
		})
}

func getCreateTxArgs(toChainID uint64, contractHashBytes []byte, method string, argsBytes []byte) []byte {
	createCrossChainTxParam := &cross_chain_manager.CreateCrossChainTxParam{
		ToChainID:         toChainID,
		ToContractAddress: contractHashBytes,
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

func GenBindAssetHashKey(contract, assetContract common.Address, chainId uint64) []byte {
	sink := common.NewZeroCopySink(nil)
	sink.WriteUint64(chainId)
	chainIdBytes := sink.Bytes()
	temp := append(contract[:], []byte(BIND_ASSET_NAME)...)
	temp = append(temp, []byte(TARGET_ASSET_HASH_PEFIX)...)
	temp = append(temp, assetContract[:]...)
	return append(temp, chainIdBytes...)
}

func GenCrossedLimitKey(contract, assetContract common.Address, chainId uint64) []byte {
	sink := common.NewZeroCopySink(nil)
	sink.WriteUint64(chainId)
	chainIdBytes := sink.Bytes()
	temp := append(contract[:], []byte(BIND_ASSET_NAME)...)
	temp = append(temp, []byte(CROSS_LIMIT_PREFIX)...)
	temp = append(temp, assetContract[:]...)
	return append(temp, chainIdBytes...)
}

func GenCrossedAmountKey(contract, sourceContract common.Address, chainId uint64) []byte {
	sink := common.NewZeroCopySink(nil)
	sink.WriteUint64(chainId)
	chainIdBytes := sink.Bytes()
	temp := append(contract[:], []byte(CROSS_AMOUNT_PREFIX)...)
	temp = append(temp, sourceContract[:]...)
	return append(temp, chainIdBytes...)
}

func getAmount(native *native.NativeService, storgedKey []byte) (*big.Int, error) {
	valueBs, err := utils.GetStorageVarBytes(native, storgedKey)
	if err != nil {
		return nil, fmt.Errorf("getAmount, error:%s", err)
	}
	//value := common.BigIntFromNeoBytes(valueBs)
	value := big.NewInt(0).SetBytes(valueBs)
	return value, nil
}

func getAllowanceInput() []byte {
	sink := common.NewZeroCopySink(nil)
	sink.WriteAddress(utils.OntContractAddress)
	sink.WriteAddress(utils.LockProxyContractAddress)

	return sink.Bytes()
}

func getTransferFromInput(toAddress common.Address, value uint64) []byte {
	transferFromState := ont.TransferFrom{
		Sender: utils.LockProxyContractAddress,
		From:   utils.OntContractAddress,
		To:     toAddress,
		Value:  value,
	}
	sink := common.NewZeroCopySink(nil)
	transferFromState.Serialization(sink)
	return sink.Bytes()
}
