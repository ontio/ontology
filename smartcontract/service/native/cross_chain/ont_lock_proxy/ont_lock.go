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
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/cross_chain/cross_chain_manager"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"io"
)

func InitOntLock() {
	native.Contracts[utils.OntLockContractAddress] = RegisterOntLockContract
}

func RegisterOntLockContract(native *native.NativeService) {
	native.Register(LOCK_NAME, OntLock)
	native.Register(UNLOCK_NAME, OntUnlock)
	native.Register(BIND_PROXY_NAME, OntBindProxyHash)
	native.Register(BIND_ASSET_NAME, OntBindAssetHash)
}

func OntBindProxyHash(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)

	var bindParam BindProxyParam
	if err := bindParam.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntBindProxyHash] Deserialize BindProxyParam error:%s", err)
	}
	operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntBindProxyHash] get operator error: %v", err)
	}
	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntBindProxyHash], checkWitness error: %v", err)
	}
	storageItem := utils.GenVarBytesStorageItem(bindParam.TargetHash)
	native.CacheDB.Put(GenBindProxyKey(utils.OntLockContractAddress, bindParam.TargetChainId), storageItem.ToArray())
	if config.DefConfig.Common.EnableEventLog {
		native.Notifications = append(native.Notifications,
			&event.NotifyEventInfo{
				ContractAddress: utils.OntLockContractAddress,
				States:          []interface{}{BIND_PROXY_NAME, bindParam.TargetChainId, hex.EncodeToString(bindParam.TargetHash)},
			})
	}
	return utils.BYTE_TRUE, nil
}
func OntBindAssetHash(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)

	var bindParam BindAssetParam
	if err := bindParam.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntBindAssetHash] Deserialize BindAssetParam error:%s", err)
	}
	operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntBindAssetHash] get operator error: %v", err)
	}
	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntBindAssetHash], checkWitness error: %v", err)
	}
	storageItem := utils.GenVarBytesStorageItem(bindParam.TargetAssetHash)
	native.CacheDB.Put(GenBindAssetKey(utils.OntLockContractAddress, bindParam.SourceAssetHash[:], bindParam.TargetChainId), storageItem.ToArray())
	if config.DefConfig.Common.EnableEventLog {
		native.Notifications = append(native.Notifications,
			&event.NotifyEventInfo{
				ContractAddress: utils.OntLockContractAddress,
				States:          []interface{}{BIND_ASSET_NAME, hex.EncodeToString(bindParam.SourceAssetHash[:]), bindParam.TargetChainId, hex.EncodeToString(bindParam.TargetAssetHash)},
			})
	}
	return utils.BYTE_TRUE, nil
}

func OntLock(native *native.NativeService) ([]byte, error) {
	ontContract := utils.OntContractAddress
	lockContract := utils.OntLockContractAddress
	source := common.NewZeroCopySource(native.Input)

	var lockParam LockParam
	err := lockParam.Deserialization(source)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntLock] contract params deserialization error:%v", err)
	}

	if lockParam.Args.Value == 0 {
		return utils.BYTE_FALSE, nil
	}
	if lockParam.Args.Value > constants.ONT_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("[OntLock] ont amount:%d over totalSupply:%d", lockParam.Args.Value, constants.ONT_TOTAL_SUPPLY)
	}
	// currently, only support ont
	if !bytes.Equal(lockParam.Args.AssetHash, ontContract[:]) {
		return utils.BYTE_FALSE, fmt.Errorf("[OntLock] only support ont lock, expect:%s, but got:%s", hex.EncodeToString(ontContract[:]), hex.EncodeToString(lockParam.Args.AssetHash))
	}

	state := ont.State{
		From:  lockParam.FromAddress,
		To:    lockContract,
		Value: lockParam.Args.Value,
	}
	transferInput := getTransferInput(state)
	_, err = native.NativeCall(utils.OntContractAddress, ont.TRANSFER_NAME, transferInput)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	toContractAddress, err := utils.GetStorageVarBytes(native, GenBindProxyKey(lockContract, lockParam.ToChainID))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntLock] get bind contract hash with chainID:%d error:%s", lockParam.ToChainID, err)
	}
	if len(toContractAddress) == 0 {
		return utils.BYTE_FALSE, fmt.Errorf("[OntLock] get bind contract hash with chainID:%d contractHash empty", lockParam.ToChainID)
	}
	AddLockNotifications(native, lockContract, toContractAddress, &lockParam)

	sink := common.NewZeroCopySink(nil)
	lockParam.Args.SerializeForMultiChain(sink)
	input := getCreateTxArgs(lockParam.ToChainID, toContractAddress, UNLOCK_NAME, sink.Bytes())
	_, err = native.NativeCall(utils.CrossChainContractAddress, cross_chain_manager.CREATE_CROSS_CHAIN_TX, input)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntLock] createCrossChainTx, error:%s", err)
	}

	return utils.BYTE_TRUE, nil
}

func OntUnlock(native *native.NativeService) ([]byte, error) {

	//  this method cannot be invoked by anybody except verifyTxManagerContract
	if !native.ContextRef.CheckWitness(utils.CrossChainContractAddress) {
		return utils.BYTE_FALSE, fmt.Errorf("[OntUnlock] should be invoked by VerirfyTxManager Contract, checkwitness failed!")
	}
	ontContract := utils.OntContractAddress
	lockContract := utils.OntLockContractAddress
	source := common.NewZeroCopySource(native.Input)

	paramsBytes, _, irregular, eof := source.NextVarBytes()
	if irregular {
		return utils.BYTE_FALSE, fmt.Errorf("[OntUnlock] input params varbytes error")
	}
	if eof {
		return utils.BYTE_FALSE, fmt.Errorf("[OntUnlock] input params varbytes error:%s", io.ErrUnexpectedEOF)
	}
	var args Args
	err := args.DeserializeForMultiChain(common.NewZeroCopySource(paramsBytes))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntUnlock] deserialize args error:%s", err)
	}

	fromContractHashBytes, _, irregular, eof := source.NextVarBytes()
	if irregular {
		return utils.BYTE_FALSE, fmt.Errorf("[OntUnlock] input from contract hash varbytes error")
	}
	if eof {
		return utils.BYTE_FALSE, fmt.Errorf("[OntUnlock] input deseriaize from contract hash error!")
	}

	fromChainId, eof := source.NextUint64()
	if eof {
		return utils.BYTE_FALSE, fmt.Errorf("[OntUnlock] input deseriaize from chainID error!")
	}

	contractHashBytes, err := utils.GetStorageVarBytes(native, GenBindProxyKey(lockContract, fromChainId))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntUnlock] get bind contract hash with chainID:%d error:%s", fromChainId, err)
	}

	if !bytes.Equal(contractHashBytes, fromContractHashBytes) {
		return utils.BYTE_FALSE, fmt.Errorf("[OntUnlock] passed in contractHash NOT equal stored contractHash with chainID:%d, expect:%s, got:%s", fromChainId, contractHashBytes, fromContractHashBytes)
	}
	// currently, only support ont
	if !bytes.Equal(args.AssetHash, ontContract[:]) {
		return utils.BYTE_FALSE, fmt.Errorf("[OntUnlock] target asset hash, expect:%s, got:%s", hex.EncodeToString(ontContract[:]), hex.EncodeToString(args.AssetHash))
	}

	toAddress, err := common.AddressParseFromBytes(args.ToAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntUnlock] parse from bytes to address error:%s", err)
	}
	if args.Value == 0 {
		return utils.BYTE_TRUE, nil
	}

	transferInput := getTransferInput(ont.State{lockContract, toAddress, args.Value})
	_, err = native.NativeCall(utils.OntContractAddress, ont.TRANSFER_NAME, transferInput)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	AddUnLockNotifications(native, lockContract, fromChainId, fromContractHashBytes, toAddress, args.Value)

	return utils.BYTE_TRUE, nil
}
