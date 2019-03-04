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

package shardgas

import (
	"bytes"
	"fmt"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/vm/neovm/types"
	"math/big"

	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/global_params"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

/////////
//
// Shard-Gas management contract
//
//	. manage user deposit gas on parent
//	. shard tx fee split with request from shard
//
/////////

const (
	// function names
	INIT_NAME          = "init"
	SET_WITHDRAW_DELAY = "setWithdrawDelay"
	DEPOSIT_GAS_NAME   = "depositGas"
	WITHDRAS_GAS_NAME  = "withdrawGas"
	ACQUIRE_GAS_NAME   = "acquireWithdrawGas"
	GET_SHARD_BALANCE  = "getShardBalance"

	// Key prefix
	KEY_VERSION        = "version"
	KEY_BALANCE        = "balance"
	KEY_WITHDRAW_DELAY = "delay"
)

var ShardGasMgmtVersion = shardmgmt.VERSION_CONTRACT_SHARD_MGMT

func InitShardGasManagement() {
	native.Contracts[utils.ShardGasMgmtContractAddress] = RegisterShardGasMgmtContract
}

func RegisterShardGasMgmtContract(native *native.NativeService) {
	native.Register(INIT_NAME, ShardGasMgmtInit)
	native.Register(SET_WITHDRAW_DELAY, SetWithdrawDelay)
	native.Register(DEPOSIT_GAS_NAME, DepositGasToShard)
	native.Register(WITHDRAS_GAS_NAME, WithdrawGasFromShard)
	native.Register(ACQUIRE_GAS_NAME, AcquireWithdrawGasFromShard)
	native.Register(GET_SHARD_BALANCE, GetUserShardBalance)
}

func ShardGasMgmtInit(native *native.NativeService) ([]byte, error) {
	// check if admin
	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getAdmin, get admin error: %v", err)
	}

	//check witness
	if err := utils.ValidateOwner(native, adminAddress); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("init shard gas, checkWitness error: %v", err)
	}

	contract := native.ContextRef.CurrentContext().ContractAddress

	// check if shard-mgmt initialized
	ver, err := getVersion(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("init shard gas, get version: %s", err)
	}
	if ver == 0 {
		// initialize shardmgmt version
		if err := setVersion(native, contract); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("init shard gas version: %s", err)
		}

		return utils.BYTE_TRUE, nil
	}

	if ver < ShardGasMgmtVersion {
		// make upgrade
		return utils.BYTE_FALSE, fmt.Errorf("upgrade TBD")
	} else if ver > ShardGasMgmtVersion {
		return utils.BYTE_FALSE, fmt.Errorf("version downgrade from %d to %d", ver, ShardGasMgmtVersion)
	}

	err = setWithdrawDelay(native, contract, shardstates.DEFAULE_WITHDRAW_DELAY)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("save withdraw delay height failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func SetWithdrawDelay(native *native.NativeService) ([]byte, error) {
	// check if admin
	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getAdmin, get admin error: %v", err)
	}

	//check witness
	if err := utils.ValidateOwner(native, adminAddress); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("set withdraw delay, checkWitness error: %v", err)
	}

	cp := new(shardmgmt.CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("set withdraw delay, invalid cmd param: %s", err)
	}

	param := new(SetWithdrawDelayParam)
	if err := param.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("set withdraw delay, invalid param: %s", err)
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	err = setWithdrawDelay(native, contract, param.DelayHeight)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("set withdraw delay, save delay height: %s", err)
	}

	return utils.BYTE_TRUE, nil
}

func DepositGasToShard(native *native.NativeService) ([]byte, error) {
	cp := new(shardmgmt.CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deposit gas, invalid cmd param: %s", err)
	}

	param := new(DepositGasParam)
	if err := param.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deposit gas, invalid param: %s", err)
	}
	if err := utils.ValidateOwner(native, param.UserAddress); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deposit gas, invalid creator: %s", err)
	}
	if param.Amount > constants.ONG_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("deposit gas, invalid amount")
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	if ok, err := checkVersion(native, contract); !ok || err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deposit gas, version check: %s", err)
	}
	if ok, err := checkShardID(native, param.ShardID); !ok || err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deposit gas, shardID check: %s", err)
	}

	gasInfo, err := getUserBalance(native, contract, param.ShardID, param.UserAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deposit gas, get user balance: %s", err)
	}

	// transfer user ong to contract
	err = ont.AppCallTransfer(native, utils.OngContractAddress, param.UserAddress, contract, param.Amount)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deposit gas, transfer ong failed, err: %s", err)
	}

	gasInfo.Balance += param.Amount
	if err := setUserDeposit(native, contract, param.ShardID, param.UserAddress, gasInfo); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deposit gas, update user balance: %s", err)
	}

	evt := &shardstates.DepositGasEvent{
		Height: native.Height,
		User:   param.UserAddress,
		Amount: param.Amount,
	}
	evt.ShardID = param.ShardID
	evt.SourceShardID = native.ShardID
	if err := shardmgmt.AddNotification(native, contract, evt); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deposit gas, add notification: %s", err)
	}

	return utils.BYTE_TRUE, nil
}

func WithdrawGasFromShard(native *native.NativeService) ([]byte, error) {
	cp := new(shardmgmt.CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("withdraw gas, invalid cmd param: %s", err)
	}

	params := new(WithdrawGasRequestParam)
	if err := params.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("withdraw gas, invalid param: %s", err)
	}
	if err := utils.ValidateOwner(native, params.UserAddress); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("withdraw gas, invalid creator: %s", err)
	}
	if params.Amount > constants.ONG_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("withdraw gas, invalid amount")
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	if ok, err := checkVersion(native, contract); !ok || err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("withdraw gas, version check: %s", err)
	}
	if ok, err := checkShardID(native, params.ShardID); !ok || err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("withdraw gas, shardID check: %s", err)
	}

	gasInfo, err := getUserBalance(native, contract, params.ShardID, params.UserAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("withdraw gas, get user balance: %s", err)
	}

	if gasInfo.Balance <= params.Amount {
		return utils.BYTE_FALSE, fmt.Errorf("withdraw gas, not enough balance for withdraw")
	}
	if len(gasInfo.PendingWithdraw) >= shardstates.CAP_PENDING_WITHDRAW {
		return utils.BYTE_FALSE, fmt.Errorf("withdraw gas, overlimited withdraw request")
	}

	gasInfo.Balance -= params.Amount
	gasInfo.WithdrawBalance += params.Amount
	gasInfo.PendingWithdraw = append(gasInfo.PendingWithdraw, &shardstates.GasWithdrawInfo{
		Height: native.Height,
		Amount: uint64(params.Amount),
	})

	if err := setUserDeposit(native, contract, params.ShardID, params.UserAddress, gasInfo); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("withdraw gas, update user balance: %s", err)
	}

	evt := &shardstates.WithdrawGasReqEvent{
		Height: native.Height,
		User:   params.UserAddress,
		Amount: params.Amount,
	}
	evt.SourceShardID = native.ShardID
	evt.ShardID = params.ShardID
	if err := shardmgmt.AddNotification(native, contract, evt); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("withdraw gas, add notification: %s", err)
	}

	return utils.BYTE_TRUE, nil
}

func AcquireWithdrawGasFromShard(native *native.NativeService) ([]byte, error) {
	cp := new(shardmgmt.CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("acquire gas, invalid cmd param: %s", err)
	}

	params := new(AcquireWithdrawGasParam)
	if err := params.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("acquire gas, invalid param: %s", err)
	}
	if err := utils.ValidateOwner(native, params.UserAddress); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("acquire gas, invalid creator: %s", err)
	}
	if params.Amount > constants.ONG_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("acquire gas, invalid amount")
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	if ok, err := checkVersion(native, contract); !ok || err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("acquire gas, version check: %s", err)
	}
	if ok, err := checkShardID(native, params.ShardID); !ok || err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("acquire gas, shardID check: %s", err)
	}

	gasInfo, err := getUserBalance(native, contract, params.ShardID, params.UserAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("acquire gas, get user balance: %s", err)
	}

	if params.Amount > gasInfo.WithdrawBalance {
		return utils.BYTE_FALSE, fmt.Errorf("acquire gas, not enough withdraw balance")
	}

	delayHeight, err := getWithdrawDelay(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("acquire gas, cannot get withdraw delay height, err: %s", err)
	}

	withdrawAmount := uint64(0)
	pendingWithdraw := make([]*shardstates.GasWithdrawInfo, 0)
	for _, w := range gasInfo.PendingWithdraw {
		if native.Height-w.Height > delayHeight {
			if withdrawAmount+w.Amount < params.Amount {
				withdrawAmount += w.Amount
			} else {
				w.Amount -= params.Amount - withdrawAmount
				withdrawAmount = params.Amount
				pendingWithdraw = append(pendingWithdraw, w)
			}
		} else {
			pendingWithdraw = append(pendingWithdraw, w)
		}
	}

	gasInfo.WithdrawBalance -= withdrawAmount
	gasInfo.PendingWithdraw = pendingWithdraw

	// transfer contract ong to user
	err = ont.AppCallTransfer(native, utils.OngContractAddress, contract, params.UserAddress, withdrawAmount)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("acquire gas, transfer ong failed, err: %s", err)
	}

	if err := setUserDeposit(native, contract, params.ShardID, params.UserAddress, gasInfo); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("acquire gas, update user balance: %s", err)
	}

	evt := &shardstates.WithdrawGasDoneEvent{
		Height: native.Height,
		User:   params.UserAddress,
		Amount: withdrawAmount,
	}
	evt.SourceShardID = native.ShardID
	evt.ShardID = params.ShardID
	if err := shardmgmt.AddNotification(native, contract, evt); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("acquire gas, add notification: %s", err)
	}

	return utils.BYTE_TRUE, nil
}

func GetUserShardBalance(native *native.NativeService) ([]byte, error) {
	cp := new(shardmgmt.CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("get balance, invalid cmd param: %s", err)
	}

	params := new(GetShardBalanceParam)
	if err := params.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("get balance, invalid param: %s", err)
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	balance, err := getUserBalance(native, contract, params.ShardId, params.UserAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("get balance, err: %s", err)
	}
	return types.BigIntToBytes(new(big.Int).SetUint64(balance.Balance)), nil
}
