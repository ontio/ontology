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
	"encoding/json"
	"fmt"
	"github.com/ontio/ontology/common"
	"math/big"
	"strings"

	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/global_params"
	"github.com/ontio/ontology/smartcontract/service/native/ong"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/shard_stake"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	ntypes "github.com/ontio/ontology/vm/neovm/types"
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
	INIT_NAME                  = "init"
	DEPOSIT_GAS_NAME           = "depositGas"
	PEER_CONFIRM_WTIDHRAW_NAME = "peerConfirmWithdraw"
	COMMIT_DPOS_NAME           = "commitDpos"
	GET_SHARD_GAS_BALANCE_NAME = "getShardGasBalance"

	USER_WITHDRAW_GAS_NAME = "userWithdrawGas"
	WITHDRAW_RETRY_NAME    = "withdrawRetry"
	USER_WITHDRAW_SUCCESS  = "userWithdrawSuccess"
	SHARD_COMMIT_DPOS      = "shardCommitDpos"
	GET_USER_WITHDRAW_ID   = "getWithdrawId"
	GET_WITHDRAW_BY_ID     = "getWithdrawById"
	GET_UN_FINISH_WITHDRAW = "getUnFinishWithdraw"
)

var ShardGasMgmtVersion = shardmgmt.VERSION_CONTRACT_SHARD_MGMT

func InitShardGasManagement() {
	native.Contracts[utils.ShardGasMgmtContractAddress] = RegisterShardGasMgmtContract
}

func RegisterShardGasMgmtContract(native *native.NativeService) {
	// invoke at root
	native.Register(INIT_NAME, ShardGasMgmtInit)
	native.Register(DEPOSIT_GAS_NAME, DepositGasToShard)
	native.Register(PEER_CONFIRM_WTIDHRAW_NAME, PeerConfirmWithdraw)
	native.Register(COMMIT_DPOS_NAME, CommitDpos)
	native.Register(GET_SHARD_GAS_BALANCE_NAME, GetShardGasBalance)

	// invoke at child
	native.Register(USER_WITHDRAW_GAS_NAME, UserWithdrawGas)
	native.Register(WITHDRAW_RETRY_NAME, UserWithdrawRetry)
	native.Register(USER_WITHDRAW_SUCCESS, UserWithdrawSuccess)
	native.Register(SHARD_COMMIT_DPOS, ShardCommitDpos)
	native.Register(GET_USER_WITHDRAW_ID, GetUserWithdrawId)
	native.Register(GET_WITHDRAW_BY_ID, GetUserWithdrawById)
	native.Register(GET_UN_FINISH_WITHDRAW, GetUserUnFinishWithdraw)
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
		setVersion(native, contract)
		return utils.BYTE_TRUE, nil
	}

	if ver < ShardGasMgmtVersion {
		// make upgrade
		return utils.BYTE_FALSE, fmt.Errorf("upgrade TBD")
	} else if ver > ShardGasMgmtVersion {
		return utils.BYTE_FALSE, fmt.Errorf("version downgrade from %d to %d", ver, ShardGasMgmtVersion)
	}
	return utils.BYTE_TRUE, nil
}

func DepositGasToShard(native *native.NativeService) ([]byte, error) {
	if !native.ShardID.IsRootShard() {
		return utils.BYTE_FALSE, fmt.Errorf("DepositGasToShard: only can be invoked at root shard")
	}

	param := new(DepositGasParam)
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("DepositGasToShard: invalid param: %s", err)
	}

	if err := utils.ValidateOwner(native, param.User); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("DepositGasToShard: invalid user: %s", err)
	}
	if param.Amount > constants.ONG_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("DepositGasToShard: invalid amount")
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	if ok, err := checkVersion(native, contract); !ok || err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("DepositGasToShard: version check: %s", err)
	}
	if ok, err := checkShardID(native, param.ShardId); !ok || err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("DepositGasToShard: shardID check: %s", err)
	}
	shardGasBalance, err := getShardGasBalance(native, contract, param.ShardId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("DepositGasToShard: failed, err: %s", err)
	}
	setShardGasBalance(native, contract, param.ShardId, shardGasBalance+param.Amount)
	if err := ont.AppCallTransfer(native, utils.OngContractAddress, param.User, contract, param.Amount); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("DepositGasToShard: transfer ong failed, err: %s", err)
	}

	evt := &shardstates.DepositGasEvent{
		Height: native.Height,
		User:   param.User,
		Amount: param.Amount,
	}
	evt.ShardID = param.ShardId
	evt.SourceShardID = native.ShardID
	if err := shardmgmt.AddNotification(native, contract, evt); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("DepositGasToShard: add notification: %s", err)
	}

	return utils.BYTE_TRUE, nil
}

func UserWithdrawGas(native *native.NativeService) ([]byte, error) {
	if native.ShardID.IsRootShard() {
		return utils.BYTE_FALSE, fmt.Errorf("UserWithdrawGas: only can be invoked at child shard")
	}
	param := new(UserWithdrawGasParam)
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserWithdrawGas: invalid param: %s", err)
	}
	if err := utils.ValidateOwner(native, param.User); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserWithdrawGas: invalid user: %s", err)
	}
	balance, err := ong.GetOngBalance(native, param.User)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserWithdrawGas: get user balance failed, err: %s", err)
	}
	if balance < param.Amount {
		return utils.BYTE_FALSE, fmt.Errorf("UserWithdrawGas: user balance not enough")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	withdrawId, err := getUserWithdrawId(native, contract, param.User)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserWithdrawGas: failed, err: %s", err)
	}
	// withdraw id should be self-increment
	withdrawId++
	setUserWithdrawId(native, contract, param.User, withdrawId)
	// freeze user ong at this contract
	setUserWithdrawGas(native, contract, param.User, withdrawId, param.Amount)
	err = ont.AppCallTransfer(native, utils.OngContractAddress, param.User, utils.ShardSysMsgContractAddress, param.Amount)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserWithdrawGas: transfer ong failed, err: %s", err)
	}
	evt := &shardstates.WithdrawGasReqEvent{
		Height:     native.Height,
		User:       param.User,
		WithdrawId: withdrawId,
		Amount:     param.Amount,
	}
	rootShard, err := common.NewShardID(0)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserWithdrawGas: generate root shard id failed, err: %s", err)
	}
	evt.ShardID = rootShard
	evt.SourceShardID = native.ShardID
	if err := shardmgmt.AddNotification(native, contract, evt); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserWithdrawGas: add notification: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func UserWithdrawRetry(native *native.NativeService) ([]byte, error) {
	if native.ShardID.IsRootShard() {
		return utils.BYTE_FALSE, fmt.Errorf("UserWithdrawRetry: only can be invoked at child shard")
	}
	param := new(UserRetryWithdrawParam)
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserWithdrawRetry: invalid param: %s", err)
	}
	if err := utils.ValidateOwner(native, param.User); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserWithdrawRetry: invalid user: %s", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	frozenGasAmount, err := getUserWithdrawGas(native, contract, param.User, param.WithdrawId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserWithdrawRetry: failed, err: %s", err)
	}
	if frozenGasAmount == 0 {
		return utils.BYTE_FALSE, fmt.Errorf("UserWithdrawRetry: the withraw %d has withdrawn", param.WithdrawId)
	}
	evt := &shardstates.WithdrawGasReqEvent{
		Height:     native.Height,
		User:       param.User,
		WithdrawId: param.WithdrawId,
		Amount:     frozenGasAmount,
	}
	rootShard, err := common.NewShardID(0)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserWithdrawRetry: generate root shard id failed, err: %s", err)
	}
	evt.ShardID = rootShard
	evt.SourceShardID = native.ShardID
	if err := shardmgmt.AddNotification(native, contract, evt); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserWithdrawRetry: add notification: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func UserWithdrawSuccess(native *native.NativeService) ([]byte, error) {
	if native.ShardID.IsRootShard() {
		return utils.BYTE_FALSE, fmt.Errorf("UserWithdrawSuccess: only can be invoked at child shard")
	}
	if native.ContextRef.CallingContext().ContractAddress != utils.ShardSysMsgContractAddress {
		return utils.BYTE_FALSE, fmt.Errorf("UserWithdrawSuccess: only can be invoked by shard sys msg contract")
	}
	param := new(UserWithdrawSuccessParam)
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserWithdrawSuccess: invalid param: %s", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	frozenGasAmount, err := getUserWithdrawGas(native, contract, param.User, param.WithdrawId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserWithdrawSuccess: failed, err: %s", err)
	}
	if frozenGasAmount == 0 {
		return utils.BYTE_FALSE, fmt.Errorf("UserWithdrawSuccess: the withraw %d has withdrawn", param.WithdrawId)
	}
	setUserWithdrawGas(native, contract, param.User, param.WithdrawId, 0)
	return utils.BYTE_TRUE, nil
}

func ShardCommitDpos(native *native.NativeService) ([]byte, error) {
	if native.ShardID.IsRootShard() {
		return utils.BYTE_FALSE, fmt.Errorf("ShardCommitDpos: only can be invoked at child shard")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	balance, err := ong.GetOngBalance(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ShardCommitDpos: get shard fee balance failed, err: %s", err)
	}
	err = ont.AppCallTransfer(native, utils.OngContractAddress, contract, utils.ShardSysMsgContractAddress, balance)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ShardCommitDpos: transfer ong failed, err: %s", err)
	}
	// TODO: call shard mgmt shard commit dpos
	evt := &shardstates.ShardCommitDposEvent{
		Height:    native.Height,
		FeeAmount: balance,
	}
	rootShard, err := common.NewShardID(0)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ShardCommitDpos: generate root shard id failed, err: %s", err)
	}
	evt.ShardID = rootShard
	evt.SourceShardID = native.ShardID
	if err := shardmgmt.AddNotification(native, contract, evt); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ShardCommitDpos: add notification: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

// pre-execute tx
func GetUserWithdrawId(native *native.NativeService) ([]byte, error) {
	if native.ShardID.IsRootShard() {
		return utils.BYTE_FALSE, fmt.Errorf("GetUserWithdrawId: only can be invoked at child shard")
	}
	user, err := utils.ReadAddress(bytes.NewBuffer(native.Input))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetUserWithdrawId: read param failed, err: %s", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	withdrawId, err := getUserWithdrawId(native, contract, user)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetUserWithdrawId: failed, err: %s", err)
	}

	return ntypes.BigIntToBytes(new(big.Int).SetUint64(withdrawId)), nil
}

func GetUserWithdrawById(native *native.NativeService) ([]byte, error) {
	if native.ShardID.IsRootShard() {
		return utils.BYTE_FALSE, fmt.Errorf("GetUserWithdrawById: only can be invoked at child shard")
	}
	param := &GetWithdrawByIdParam{}
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetUserWithdrawById: failed, err: %s", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	amount, err := getUserWithdrawGas(native, contract, param.User, param.WithdrawId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetUserWithdrawById: failed, err: %s", err)
	}
	return ntypes.BigIntToBytes(new(big.Int).SetUint64(amount)), nil
}

func GetUserUnFinishWithdraw(native *native.NativeService) ([]byte, error) {
	if native.ShardID.IsRootShard() {
		return utils.BYTE_FALSE, fmt.Errorf("GetUserUnFinishWithdraw: only can be invoked at child shard")
	}
	user, err := utils.ReadAddress(bytes.NewBuffer(native.Input))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetUserUnFinishWithdraw: read param failed, err: %s", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	withdrawId, err := getUserWithdrawId(native, contract, user)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetUserUnFinishWithdraw: failed, err: %s", err)
	}
	result := make(map[uint64]uint64)
	// withdraw id start from 1
	for i := uint64(1); i < withdrawId; i++ {
		amount, _ := getUserWithdrawGas(native, contract, user, i)
		if amount > 0 {
			result[i] = amount
		}
	}
	data, _ := json.Marshal(result)
	return data, nil
}

func PeerConfirmWithdraw(native *native.NativeService) ([]byte, error) {
	if !native.ShardID.IsRootShard() {
		return utils.BYTE_FALSE, fmt.Errorf("PeerConfirmWithdraw: only can be invoked at root shard")
	}
	param := new(PeerWithdrawGasParam)
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerConfirmWithdraw: invalid param: %s", err)
	}
	if err := utils.ValidateOwner(native, param.Signer); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerConfirmWithdraw: invalid peer signer: %s", err)
	}
	shard, err := shardmgmt.GetShardState(native, utils.ShardMgmtContractAddress, param.ShardId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerConfirmWithdraw: get shard state failed, err: %s", err)
	}
	shardPeerInfo, ok := shard.Peers[strings.ToLower(param.PeerPubKey)]
	if !ok {
		return utils.BYTE_FALSE, fmt.Errorf("PeerConfirmWithdraw: peer not exist at shard")
	}
	if shardPeerInfo.NodeType != shardstates.CONSENSUS_NODE {
		return utils.BYTE_FALSE, fmt.Errorf("PeerConfirmWithdraw: peer not consensus node")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	oldConfirmedNum, err := getWithdrawConfirmNum(native, contract, param.User, param.ShardId, param.WithdrawId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerConfirmWithdraw: failed, err: %s", err)
	}
	newConfirmedNum := oldConfirmedNum
	isPeerConfirmed, err := isPeerConfirmWithdraw(native, contract, param.User, param.PeerPubKey, param.ShardId,
		param.WithdrawId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerConfirmWithdraw: failed, err: %s", err)
	}
	if !isPeerConfirmed {
		newConfirmedNum++
		peerConfirmWithdraw(native, contract, param.User, param.PeerPubKey, param.ShardId, param.WithdrawId)
		setWithdrawConfirmNum(native, contract, param.User, param.ShardId, param.WithdrawId, newConfirmedNum)
	}
	required := shard.Config.VbftCfg.K - shard.Config.VbftCfg.C
	if config.DefConfig.Genesis.ConsensusType == config.CONSENSUS_TYPE_SOLO {
		required = 1
	}
	if uint32(newConfirmedNum) < required {
		return utils.BYTE_TRUE, nil
	} else {
		if uint32(oldConfirmedNum) < required {
			shardBalance, err := getShardGasBalance(native, contract, param.ShardId)
			if err != nil {
				return utils.BYTE_FALSE, fmt.Errorf("PeerConfirmWithdraw: failed, err: %s", err)
			}
			if shardBalance < param.Amount {
				return utils.BYTE_FALSE, fmt.Errorf("PeerConfirmWithdraw: shard balance not enough")
			}
			setShardGasBalance(native, contract, param.ShardId, shardBalance-param.Amount)
			err = ont.AppCallTransfer(native, utils.OngContractAddress, contract, param.User, param.Amount)
			if err != nil {
				return utils.BYTE_FALSE, fmt.Errorf("PeerConfirmWithdraw: transfer ong failed, err: %s", err)
			}
		}
		evt := &shardstates.WithdrawGasDoneEvent{
			Height:     native.Height,
			User:       param.User,
			WithdrawId: param.WithdrawId,
		}
		evt.ShardID = param.ShardId
		evt.SourceShardID = native.ShardID
		if err := shardmgmt.AddNotification(native, contract, evt); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("PeerConfirmWithdraw: add notification: %s", err)
		}
	}
	return utils.BYTE_TRUE, nil
}

func CommitDpos(native *native.NativeService) ([]byte, error) {
	if !native.ShardID.IsRootShard() {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: only can be invoked at root")
	}
	param := new(CommitDposParam)
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: invalid param: %s", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	shard, err := shardmgmt.GetShardState(native, utils.ShardMgmtContractAddress, param.ShardID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: get shard state failed, err: %s", err)
	}
	shardBalance, err := getShardGasBalance(native, contract, param.ShardID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: failed, err: %s", err)
	}
	if shardBalance < param.FeeAmount {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: shard gas balance not enough")
	}
	shardPeerInfo, ok := shard.Peers[strings.ToLower(param.PeerPubKey)]
	if !ok {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: peer not exist at shard")
	}
	if shardPeerInfo.NodeType != shardstates.CONSENSUS_NODE {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: peer not consensus node")
	}
	shardCurrentView, err := shard_stake.GetShardCurrentView(native, param.ShardID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: failed, err: %s", err)
	}
	oldCommitAmount, err := getViewCommitNum(native, contract, param.ShardID, shardCurrentView)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: failed, err: %s", err)
	}
	newCommitNum := oldCommitAmount
	isPeerCommited, err := isPeerCommitView(native, contract, param.PeerPubKey, param.ShardID, shardCurrentView)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: failed, err: %s", err)
	}
	if !isPeerCommited {
		newCommitNum++
		if err := peerCommitView(native, contract, param.PeerPubKey, param.ShardID, shardCurrentView); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: failed, err: %s", err)
		}
		if err := setViewCommitNum(native, contract, param.ShardID, shardCurrentView, newCommitNum); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: failed, err: %s", err)
		}
	}
	required := shard.Config.VbftCfg.K - shard.Config.VbftCfg.C
	if config.DefConfig.Genesis.ConsensusType == config.CONSENSUS_TYPE_SOLO {
		required = 1
	}
	if uint32(newCommitNum) < required {
		return utils.BYTE_TRUE, nil
	} else if uint32(oldCommitAmount) < required {
		err = ont.AppCallTransfer(native, utils.OngContractAddress, contract, utils.ShardStakeAddress, param.FeeAmount)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: transfer ong failed, err: %s", err)
		}
		// call shard mgmt commit dpos
		bf := new(bytes.Buffer)
		if err := param.CommitDposParam.Serialize(bf); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: serialize mgmt commit dpos param failed, err: %s", err)
		}
		if _, err = native.NativeCall(utils.ShardMgmtContractAddress, shardmgmt.COMMIT_DPOS_NAME, bf.Bytes()); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: call shard mgmt failed, err: %s", err)
		}
	}
	return utils.BYTE_TRUE, nil
}

// pre-execute tx
func GetShardGasBalance(native *native.NativeService) ([]byte, error) {
	if !native.ShardID.IsRootShard() {
		return utils.BYTE_FALSE, fmt.Errorf("GetShardGasBalance: only can be invoked at root")
	}
	bf := bytes.NewBuffer(native.Input)
	param, err := utils.ReadVarUint(bf)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetShardGasBalance: deserialize param failed, err: %s", err)
	}
	shardId, err := common.NewShardID(param)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetShardGasBalance: generate root shard id failed, err: %s", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	shardBalance, err := getShardGasBalance(native, contract, shardId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetShardGasBalance: failed, err: %s", err)
	}
	return ntypes.BigIntToBytes(new(big.Int).SetUint64(shardBalance)), nil
}
