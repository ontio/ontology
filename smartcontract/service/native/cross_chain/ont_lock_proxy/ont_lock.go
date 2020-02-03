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
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/cross_chain/cross_chain_manager"
	"github.com/ontio/ontology/smartcontract/service/native/global_params"
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
	native.Register(BIND_NAME, OntBind)
}

func OntBind(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	targetChainId, _, irregular, eof := source.NextVarUint()
	if irregular {
		return utils.BYTE_FALSE, fmt.Errorf("[OntBind] decode targetChainId NextVarUint error")
	}
	if eof {
		return utils.BYTE_FALSE, fmt.Errorf("[OntBind] decode targetChainId error")
	}
	targetChainContractHash, _, irregular, eof := source.NextVarBytes()
	if irregular {
		return utils.BYTE_FALSE, fmt.Errorf("[OntBind] decode targetChainContractHash NextVarBytes error")
	}
	if eof {
		return utils.BYTE_FALSE, fmt.Errorf("[OntBind] decode targetChainContractHash error:%s", io.ErrUnexpectedEOF)
	}
	operatorAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntBind] getAdmin, get admin error: %v", err)
	}
	if native.ContextRef.CheckWitness(operatorAddress) == false {
		return utils.BYTE_FALSE, errors.NewErr("[OntBind] authentication failed!")
	}
	storageItem := utils.GenVarBytesStorageItem(targetChainContractHash)
	native.CacheDB.Put(GenBindKey(utils.OntContractAddress, targetChainId), storageItem.ToArray())
	if config.DefConfig.Common.EnableEventLog {
		native.Notifications = append(native.Notifications,
			&event.NotifyEventInfo{
				ContractAddress: utils.OntLockContractAddress,
				States:          []interface{}{BIND_NAME, targetChainId, hex.EncodeToString(targetChainContractHash)},
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

	state := &ont.State{
		From:  lockParam.FromAddress,
		To:    lockContract,
		Value: lockParam.Args.Value,
	}
	_, _, err = ont.Transfer(native, ontContract, state)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	toContractAddress, err := utils.GetStorageVarBytes(native, GenBindKey(ontContract, lockParam.ToChainID))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntLock] get bind contract hash with chainID:%d error:%s", lockParam.ToChainID, err)
	}
	if len(toContractAddress) == 0 {
		return utils.BYTE_FALSE, fmt.Errorf("[OntLock] get bind contract hash with chainID:%d contractHash empty", lockParam.ToChainID)
	}
	AddLockNotifications(native, lockContract, toContractAddress, &lockParam)

	sink := common.NewZeroCopySink(nil)
	lockParam.Args.Serialization(sink)
	input := getCreateTxArgs(lockParam.ToChainID, toContractAddress, lockParam.Fee, UNLOCK_NAME, sink.Bytes())
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
	err := args.Deserialization(common.NewZeroCopySource(paramsBytes))
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

	contractHashBytes, err := utils.GetStorageVarBytes(native, GenBindKey(ontContract, fromChainId))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntUnlock] get bind contract hash with chainID:%d error:%s", fromChainId, err)
	}

	if !bytes.Equal(contractHashBytes, fromContractHashBytes) {
		return utils.BYTE_FALSE, fmt.Errorf("[OntUnlock] passed in contractHash NOT equal stored contractHash with chainID:%d, expect:%s, got:%s", fromChainId, contractHashBytes, fromContractHashBytes)
	}

	toAddress, err := common.AddressParseFromBytes(args.ToAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntUnlock] parse from bytes to address error:%s", err)
	}
	if args.Value == 0 {
		return utils.BYTE_TRUE, nil
	}
	_, _, err = ont.Transfer(native, ontContract, &ont.State{lockContract, toAddress, args.Value})
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	AddUnLockNotifications(native, lockContract, fromChainId, fromContractHashBytes, toAddress, args.Value)

	return utils.BYTE_TRUE, nil
}
