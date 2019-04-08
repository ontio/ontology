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

package shard_stake

import (
	"bytes"
	"fmt"

	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	INIT_SHARD               = "initShard"
	PEER_STAKE               = "peerInitStake"
	USER_STAKE               = "userStake"
	SET_MIN_STAKE            = "setMinStake"
	UNFREEZE_STAKE           = "unfreezeStake"
	WITHDRAW_STAKE           = "withdrawStake"
	WITHDRAW_FEE             = "withdrawFee"
	CHANGE_MAX_AUTHORIZATION = "changeMaxAuthorization"
	CHANGE_PROPORTION        = "changeProportion" // node change proportion of stake user
	COMMIT_DPOS              = "commitDpos"
	PEER_EXIT                = "peerExit"
	DELETE_PEER              = "deletePeer"
	WITHDRAW_ONG             = "withdrawOng"
)

func InitShardStake() {
	native.Contracts[utils.ShardStakeAddress] = RegisterShardStake
}

func RegisterShardStake(native *native.NativeService) {
	native.Register(INIT_SHARD, InitShard)
	native.Register(PEER_STAKE, PeerInitStake)
	native.Register(USER_STAKE, UserStake)
	native.Register(UNFREEZE_STAKE, UnfreezeStake)
	native.Register(WITHDRAW_STAKE, WithdrawStake)
	native.Register(SET_MIN_STAKE, SetMinStake)
	native.Register(WITHDRAW_FEE, WithdrawFee)
	native.Register(COMMIT_DPOS, CommitDpos)
	native.Register(CHANGE_MAX_AUTHORIZATION, ChangeMaxAuthorization)
	native.Register(CHANGE_PROPORTION, ChangeProportion)
	native.Register(DELETE_PEER, DeletePeer)
	native.Register(PEER_EXIT, PeerExit)
	native.Register(WITHDRAW_ONG, WithdrawOng)
}

func SetMinStake(native *native.NativeService) ([]byte, error) {
	if native.ContextRef.CallingContext().ContractAddress != utils.ShardMgmtContractAddress {
		return utils.BYTE_FALSE, fmt.Errorf("SetMinStake: only can be call by shard mgmt contract")
	}
	params := new(SetMinStakeParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SetMinStake: invalid param: %s", err)
	}
	setNodeMinStakeAmount(native, params.ShardId, params.Amount)
	return utils.BYTE_TRUE, nil
}

func InitShard(native *native.NativeService) ([]byte, error) {
	if native.ContextRef.CallingContext().ContractAddress != utils.ShardMgmtContractAddress {
		return utils.BYTE_FALSE, fmt.Errorf("PeerInitStake: only shard mgmt can invoke")
	}
	shardId, err := utils.DeserializeShardId(bytes.NewBuffer(native.Input))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerInitStake: failed, err: %s", err)
	}
	shardView := &utils.ChangeView{
		View:   0,
		Height: native.Height,
		TxHash: native.Tx.Hash(),
	}
	setShardView(native, shardId, shardView)
	return utils.BYTE_TRUE, nil
}

func PeerInitStake(native *native.NativeService) ([]byte, error) {
	params := new(PeerInitStakeParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerInitStake: invalid param: %s", err)
	}
	// only call by shard mgmt contract
	if native.ContextRef.CallingContext().ContractAddress != utils.ShardMgmtContractAddress {
		return utils.BYTE_FALSE, fmt.Errorf("PeerInitStake: only shard mgmt can invoke")
	}
	err := peerStake(native, params.ShardId, params.Value.PeerPubKey, params.PeerOwner, params.Value.Amount)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerInitStake: deserialize param pub key failed, err: %s", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	err = setShardStakeAssetAddr(native, contract, params.ShardId, params.StakeAssetAddr)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerInitStake: failed, err: %s", err)
	}
	minStakeAmount, err := GetNodeMinStakeAmount(native, params.ShardId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerInitStake: failed, err: %s", err)
	}
	if minStakeAmount > params.Value.Amount {
		return utils.BYTE_FALSE, fmt.Errorf("PeerInitStake: stake amount should be larger than min stake amount")
	}
	if params.StakeAssetAddr == utils.OntContractAddress {
		// save user unbound ong info
		unboundOngInfo := &UserUnboundOngInfo{
			StakeAmount: params.Value.Amount,
			Time:        native.Time,
		}
		setUserUnboundOngInfo(native, contract, params.PeerOwner, unboundOngInfo)
	}
	// transfer stake asset
	err = ont.AppCallTransfer(native, params.StakeAssetAddr, params.PeerOwner, contract, params.Value.Amount)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerInitStake: transfer stake asset failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

// peer quit consensus completed, only call by shard mgmt at commitDpos
func PeerExit(native *native.NativeService) ([]byte, error) {
	param := new(PeerExitParam)
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerExit: invalid param: %s", err)
	}
	if native.ContextRef.CallingContext().ContractAddress != utils.ShardMgmtContractAddress {
		return utils.BYTE_FALSE, fmt.Errorf("PeerExit: only can be invoked by shardmgmt contract")
	}
	currentView, err := GetShardCurrentView(native, param.ShardId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerExit: failed, err: %s", err)
	}
	// get the next view info
	nextView := currentView + 1
	viewInfo, err := GetShardViewInfo(native, param.ShardId, nextView)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerExit: failed, err: %s", err)
	}
	if peerViewInfo, ok := viewInfo.Peers[param.Peer]; !ok {
		return utils.BYTE_FALSE, fmt.Errorf("PeerExit: peer %s not exist", peerViewInfo.PeerPubKey)
	} else {
		peerViewInfo.CanStake = false
	}
	setShardViewInfo(native, param.ShardId, nextView, viewInfo)
	return utils.BYTE_TRUE, nil
}

// peer quit consensus completed, only call by shard mgmt at commitDpos
func DeletePeer(native *native.NativeService) ([]byte, error) {
	param := new(DeletePeerParam)
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("DeletePeer: invalid param: %s", err)
	}
	if native.ContextRef.CallingContext().ContractAddress != utils.ShardMgmtContractAddress {
		return utils.BYTE_FALSE, fmt.Errorf("DeletePeer: only can be invoked by shardmgmt contract")
	}
	currentView, err := GetShardCurrentView(native, param.ShardId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("DeletePeer: failed, err: %s", err)
	}
	// get the next view info
	nextView := currentView + 1
	viewInfo, err := GetShardViewInfo(native, param.ShardId, nextView)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("DeletePeer: failed, err: %s", err)
	}
	for _, peer := range param.Peers {
		delete(viewInfo.Peers, peer)
	}
	setShardViewInfo(native, param.ShardId, nextView, viewInfo)
	return utils.BYTE_TRUE, nil
}

func UserStake(native *native.NativeService) ([]byte, error) {
	param := new(UserStakeParam)
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserStake: invalid param: %s", err)
	}
	if err := utils.ValidateOwner(native, param.User); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserStake: check witness failed, err: %s", err)
	}
	err := userStake(native, param.ShardId, param.User, param.Value)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserStake: failed, err: %s", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	wholeAmount := uint64(0)
	for _, value := range param.Value {
		wholeAmount += value.Amount
	}
	stakeAssetAddr, err := getShardStakeAssetAddr(native, contract, param.ShardId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserStake: failed, err: %s", err)
	}
	if stakeAssetAddr == utils.OntContractAddress {
		unboundOngInfo, err := getUserUnboundOngInfo(native, contract, param.User)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("UserStake: failed, err: %s", err)
		}
		if unboundOngInfo.Time == 0 {
			// user stake firstly
			unboundOngInfo.Time = native.Time
			unboundOngInfo.StakeAmount = wholeAmount
		} else {
			amount := utils.CalcUnbindOng(unboundOngInfo.StakeAmount,
				unboundOngInfo.Time-constants.GENESIS_BLOCK_TIMESTAMP, native.Time-constants.GENESIS_BLOCK_TIMESTAMP)
			unboundOngInfo.Balance += amount
			unboundOngInfo.Time = native.Time
			unboundOngInfo.StakeAmount += wholeAmount
		}
		setUserUnboundOngInfo(native, contract, param.User, unboundOngInfo)
	}
	if err := ont.AppCallTransfer(native, stakeAssetAddr, param.User, contract, wholeAmount); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserStake: transfer stake asset failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func UnfreezeStake(native *native.NativeService) ([]byte, error) {
	param := new(UnfreezeFromShardParam)
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UnfreezeStake: invalid param: %s", err)
	}
	if err := utils.ValidateOwner(native, param.User); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UnfreezeStake: check witness failed, err: %s", err)
	}
	err := unfreezeStakeAsset(native, param.ShardId, param.User, param.Value)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UnfreezeStake: faield, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func WithdrawStake(native *native.NativeService) ([]byte, error) {
	param := new(WithdrawStakeAssetParam)
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("WithdrawStake: invalid param: %s", err)
	}
	if err := utils.ValidateOwner(native, param.User); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("WithdrawStake: check witness failed, err: %s", err)
	}
	num, err := withdrawStakeAsset(native, param.ShardId, param.User)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("WithdrawStake: failed, err: %s", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	stakeAssetAddr, err := getShardStakeAssetAddr(native, contract, param.ShardId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserStake: failed, err: %s", err)
	}
	if stakeAssetAddr == utils.OntContractAddress {
		unboundOngInfo, err := getUserUnboundOngInfo(native, contract, param.User)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("UserStake: failed, err: %s", err)
		}
		if unboundOngInfo.Time == 0 {
			// user stake firstly
			unboundOngInfo.Time = native.Time
			unboundOngInfo.StakeAmount = num
		} else if unboundOngInfo.StakeAmount > num {
			amount := utils.CalcUnbindOng(unboundOngInfo.StakeAmount, unboundOngInfo.Time-constants.GENESIS_BLOCK_TIMESTAMP,
				native.Time-constants.GENESIS_BLOCK_TIMESTAMP)
			unboundOngInfo.Balance += amount
			unboundOngInfo.Time = native.Time
			unboundOngInfo.StakeAmount += num
		} else {
			return utils.BYTE_FALSE, fmt.Errorf("UserStake: user stake whole amount not enough")
		}
		setUserUnboundOngInfo(native, contract, param.User, unboundOngInfo)
	}
	err = ont.AppCallTransfer(native, utils.OntContractAddress, contract, param.User, num)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("WithdrawStake: transfer ont failed, amount %d, err: %s", num, err)
	}
	return utils.BYTE_TRUE, nil
}

func WithdrawFee(native *native.NativeService) ([]byte, error) {
	param := new(WithdrawFeeParam)
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("WithdrawFee: invalid param: %s", err)
	}
	if err := utils.ValidateOwner(native, param.User); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("WithdrawFee: check witness failed, err: %s", err)
	}
	amount, err := withdrawFee(native, param.ShardId, param.User)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("WithdrawFee: failed, err: %s", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	err = ont.AppCallTransfer(native, utils.OngContractAddress, contract, param.User, amount)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("WithdrawFee: transfer ong failed, amount %d, err: %s", amount, err)
	}
	return utils.BYTE_TRUE, nil
}

func CommitDpos(native *native.NativeService) ([]byte, error) {
	if native.ContextRef.CallingContext().ContractAddress != utils.ShardMgmtContractAddress {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: only shard mgmt contract can invoke")
	}
	param := new(CommitDposParam)
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: invalid param: %s", err)
	}
	err := commitDpos(native, param.ShardId, param.Value)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func ChangeMaxAuthorization(native *native.NativeService) ([]byte, error) {
	param := new(ChangeMaxAuthorizationParam)
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ChangeMaxAuthorization: invalid param: %s", err)
	}
	if err := utils.ValidateOwner(native, param.User); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ChangeMaxAuthorization: check witness failed, err: %s", err)
	}
	err := changePeerInfo(native, param.ShardId, param.User, param.Value, CHANGE_MAX_AUTHORIZATION)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ChangeMaxAuthorization: failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func ChangeProportion(native *native.NativeService) ([]byte, error) {
	param := new(ChangeProportionParam)
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ChangeProportion: invalid param: %s", err)
	}
	if err := utils.ValidateOwner(native, param.User); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ChangeProportion: check witness failed, err: %s", err)
	}
	if param.Value.Amount > 100 {
		return utils.BYTE_FALSE, fmt.Errorf("ChangeProportion: proportion larger than 100")
	}
	err := changePeerInfo(native, param.ShardId, param.User, param.Value, CHANGE_PROPORTION)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ChangeProportion: failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func WithdrawOng(native *native.NativeService) ([]byte, error) {
	user, err := utils.ReadAddress(bytes.NewBuffer(native.Input))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("WithdrawOng: invalid param: %s", err)
	}
	if err := utils.ValidateOwner(native, user); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("WithdrawOng: check witness failed, err: %s", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	unboundOngInfo, err := getUserUnboundOngInfo(native, contract, user)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("WithdrawOng: failed, err: %s", err)
	}
	if unboundOngInfo.Time == 0 {
		return utils.BYTE_FALSE, fmt.Errorf("WithdrawOng: user not stake")
	}
	amount := utils.CalcUnbindOng(unboundOngInfo.StakeAmount, unboundOngInfo.Time-constants.GENESIS_BLOCK_TIMESTAMP,
		native.Time-constants.GENESIS_BLOCK_TIMESTAMP)
	wholeAmount := amount + unboundOngInfo.Balance
	unboundOngInfo.Balance = 0
	unboundOngInfo.Time = native.Time
	setUserUnboundOngInfo(native, contract, user, unboundOngInfo)
	if err := ont.AppCallTransfer(native, utils.OntContractAddress, contract, contract, 1); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("WithdrawOng: transfer ont failed, err: %s", err)
	}
	if err := ont.AppCallTransferFrom(native, utils.OngContractAddress, contract, utils.OntContractAddress, user,
		wholeAmount); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("WithdrawOng: transfer ong failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}
