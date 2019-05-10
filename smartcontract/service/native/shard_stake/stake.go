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
	"encoding/json"
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"math/big"

	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	ntypes "github.com/ontio/ontology/vm/neovm/types"
)

const (
	INIT_SHARD               = "initShard"
	PEER_STAKE               = "peerInitStake"
	ADD_INIT_DOS             = "addInitPos"
	REDUCE_INIT_POS          = "reduceInitPos"
	USER_STAKE               = "userStake"
	UNFREEZE_STAKE           = "unfreezeStake"
	WITHDRAW_STAKE           = "withdrawStake"
	WITHDRAW_FEE             = "withdrawFee"
	CHANGE_MAX_AUTHORIZATION = "changeMaxAuthorization"
	CHANGE_PROPORTION        = "changeProportion" // node change proportion of stake user
	PRE_COMMIT_DPOS          = "preCommitDpos"
	COMMIT_DPOS              = "commitDpos"
	PEER_EXIT                = "peerExit"
	DELETE_PEER              = "deletePeer"
	WITHDRAW_ONG             = "withdrawOng"

	// for pre-execute
	GET_CURRENT_VIEW  = "getCurrentView"
	GET_PEER_INFO     = "getPeerInfo"
	GET_USER_INFO     = "getUserInfo"
	GET_IS_COMMITTING = "getIsCommitting"
)

func InitShardStake() {
	native.Contracts[utils.ShardStakeAddress] = RegisterShardStake
}

func RegisterShardStake(native *native.NativeService) {
	native.Register(INIT_SHARD, InitShard)
	native.Register(PEER_STAKE, PeerInitStake)
	native.Register(ADD_INIT_DOS, AddInitPos)
	native.Register(REDUCE_INIT_POS, ReduceInitPos)
	native.Register(USER_STAKE, UserStake)
	native.Register(UNFREEZE_STAKE, UnfreezeStake)
	native.Register(WITHDRAW_STAKE, WithdrawStake)
	native.Register(WITHDRAW_FEE, WithdrawFee)
	native.Register(PRE_COMMIT_DPOS, PreCommitDpos)
	native.Register(COMMIT_DPOS, CommitDpos)
	native.Register(CHANGE_MAX_AUTHORIZATION, ChangeMaxAuthorization)
	native.Register(CHANGE_PROPORTION, ChangeProportion)
	native.Register(DELETE_PEER, DeletePeer)
	native.Register(PEER_EXIT, PeerExit)
	native.Register(WITHDRAW_ONG, WithdrawOng)

	native.Register(GET_IS_COMMITTING, GetIsCommitting)
	native.Register(GET_CURRENT_VIEW, GetCurrentView)
	native.Register(GET_PEER_INFO, GetPeerInfo)
	native.Register(GET_USER_INFO, GetUserInfo)
}

func InitShard(native *native.NativeService) ([]byte, error) {
	if native.ContextRef.CallingContext().ContractAddress != utils.ShardMgmtContractAddress {
		return utils.BYTE_FALSE, fmt.Errorf("PeerInitStake: only shard mgmt can invoke")
	}
	param := &InitShardParam{}
	err := param.Deserialize(bytes.NewBuffer(native.Input))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerInitStake: failed, err: %s", err)
	}
	shardView := &utils.ChangeView{
		View:   0,
		Height: native.Height,
		TxHash: native.Tx.Hash(),
	}
	setShardView(native, param.ShardId, shardView)
	setNodeMinStakeAmount(native, param.ShardId, param.MinStake)
	setShardStakeAssetAddr(native, param.ShardId, param.StakeAssetAddr)
	return utils.BYTE_TRUE, nil
}

func PeerInitStake(native *native.NativeService) ([]byte, error) {
	param := new(PeerStakeParam)
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerInitStake: invalid param: %s", err)
	}
	// only call by shard mgmt contract
	if native.ContextRef.CallingContext().ContractAddress != utils.ShardMgmtContractAddress {
		return utils.BYTE_FALSE, fmt.Errorf("PeerInitStake: only shard mgmt can invoke")
	}
	if err := checkCommittingDpos(native, param.ShardId); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerInitStake: failed, err: %s", err)
	}
	err := peerInitStake(native, param.ShardId, param.Value.PeerPubKey, param.PeerOwner, param.Value.Amount)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerInitStake: deserialize param pub key failed, err: %s", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	minStakeAmount, err := GetNodeMinStakeAmount(native, param.ShardId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerInitStake: failed, err: %s", err)
	}
	if minStakeAmount > param.Value.Amount {
		return utils.BYTE_FALSE, fmt.Errorf("PeerInitStake: stake amount should be larger than min stake amount")
	}
	stakeAssetAddr, err := getShardStakeAssetAddr(native, param.ShardId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerInitStake: failed, err: %s", err)
	}
	if stakeAssetAddr == utils.OntContractAddress {
		// save user unbound ong info
		unboundOngInfo := &UserUnboundOngInfo{
			StakeAmount: param.Value.Amount,
			Time:        native.Time,
		}
		setUserUnboundOngInfo(native, param.PeerOwner, unboundOngInfo)
	}
	// transfer stake asset
	err = ont.AppCallTransfer(native, stakeAssetAddr, param.PeerOwner, contract, param.Value.Amount)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerInitStake: transfer stake asset failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func AddInitPos(native *native.NativeService) ([]byte, error) {
	param := new(PeerStakeParam)
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("AddInitPos: invalid param: %s", err)
	}
	if err := utils.ValidateOwner(native, param.PeerOwner); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("AddInitPos: check witness failed, err: %s", err)
	}
	if err := checkCommittingDpos(native, param.ShardId); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("AddInitPos: failed, err: %s", err)
	}
	if err := addInitPos(native, param.ShardId, param.PeerOwner, param.Value); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("AddInitPos: failed, err: %s", err)
	}
	stakeAssetAddr, err := getShardStakeAssetAddr(native, param.ShardId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("AddInitPos: failed, err: %s", err)
	}
	if stakeAssetAddr == utils.OntContractAddress {
		unboundOngInfo, err := getUserUnboundOngInfo(native, param.PeerOwner)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("AddInitPos: failed, err: %s", err)
		}
		if unboundOngInfo.Time == 0 {
			return utils.BYTE_FALSE, fmt.Errorf("AddInitPos: peer owner unboundong time is 0")
		}
		amount := utils.CalcUnbindOng(unboundOngInfo.StakeAmount,
			unboundOngInfo.Time-constants.GENESIS_BLOCK_TIMESTAMP, native.Time-constants.GENESIS_BLOCK_TIMESTAMP)
		unboundOngInfo.Balance += amount
		unboundOngInfo.Time = native.Time
		unboundOngInfo.StakeAmount += param.Value.Amount
		setUserUnboundOngInfo(native, param.PeerOwner, unboundOngInfo)
	}
	err = ont.AppCallTransfer(native, stakeAssetAddr, param.PeerOwner, utils.ShardStakeAddress, param.Value.Amount)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("AddInitPos: transfer stake asset failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func ReduceInitPos(native *native.NativeService) ([]byte, error) {
	param := new(PeerStakeParam)
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ReduceInitPos: invalid param: %s", err)
	}
	if err := checkCommittingDpos(native, param.ShardId); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ReduceInitPos: failed, err: %s", err)
	}
	if err := utils.ValidateOwner(native, param.PeerOwner); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ReduceInitPos: check witness failed, err: %s", err)
	}
	if err := reduceInitPos(native, param.ShardId, param.PeerOwner, param.Value); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ReduceInitPos: failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

// peer quit consensus, user cannot stake, only call by shard mgmt at commitDpos
func PeerExit(native *native.NativeService) ([]byte, error) {
	param := new(PeerExitParam)
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerExit: invalid param: %s", err)
	}
	if native.ContextRef.CallingContext().ContractAddress != utils.ShardMgmtContractAddress {
		return utils.BYTE_FALSE, fmt.Errorf("PeerExit: only can be invoked by shardmgmt contract")
	}
	if err := checkCommittingDpos(native, param.ShardId); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerExit: failed, err: %s", err)
	}
	currentView, err := GetShardCurrentView(native, param.ShardId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerExit: failed, err: %s", err)
	}
	currentViewInfo, err := GetShardViewInfo(native, param.ShardId, currentView)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerExit: get current view info failed, err: %s", err)
	}
	currentPeerInfo, ok := currentViewInfo.Peers[param.Peer]
	if !ok {
		return utils.BYTE_FALSE, fmt.Errorf("PeerExit: peer %s not exist", currentPeerInfo.PeerPubKey)
	} else {
		currentPeerInfo.CanStake = false
	}
	currentViewInfo.Peers[param.Peer] = currentPeerInfo
	setShardViewInfo(native, param.ShardId, currentView, currentViewInfo)
	// get the next view info
	nextView := currentView + 1
	nextViewInfo, err := GetShardViewInfo(native, param.ShardId, nextView)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerExit: get next view info failed, err: %s", err)
	}
	nextPeerInfo, ok := nextViewInfo.Peers[param.Peer]
	if !ok {
		nextPeerInfo = currentPeerInfo
	} else {
		nextPeerInfo.CanStake = false
	}
	nextViewInfo.Peers[param.Peer] = nextPeerInfo
	setShardViewInfo(native, param.ShardId, nextView, nextViewInfo)
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
	if err := checkCommittingDpos(native, param.ShardId); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("DeletePeer: failed, err: %s", err)
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
	// peer could withdraw stake asset at new consensus epoch
	for _, peer := range param.Peers {
		if peerInfo, ok := viewInfo.Peers[peer]; ok {
			lastStakeView, err := getUserLastStakeView(native, param.ShardId, peerInfo.Owner)
			if err != nil {
				return utils.BYTE_FALSE, fmt.Errorf("DeletePeer: peer %s, err: %s", peer, err)
			}
			if lastStakeView > nextView {
				return utils.BYTE_FALSE, fmt.Errorf("DeletePeer: peer %s last stake view %d and next view %d unmatch",
					peer, lastStakeView, nextView)
			}
			ownerStakeInfo, err := getShardViewUserStake(native, param.ShardId, lastStakeView, peerInfo.Owner)
			if err != nil {
				return utils.BYTE_FALSE, fmt.Errorf("DeletePeer: peer %s, err: %s", peer, err)
			}
			ownerStakeSelfInfo, ok := ownerStakeInfo.Peers[peer]
			if ok {
				ownerStakeSelfInfo.UnfreezeAmount += peerInfo.InitPos
			} else {
				ownerStakeSelfInfo = &UserPeerStakeInfo{PeerPubKey: peer, UnfreezeAmount: peerInfo.InitPos}
			}
			peerInfo.UserUnfreezeAmount += peerInfo.InitPos
			peerInfo.InitPos = 0
			setUserLastStakeView(native, param.ShardId, peerInfo.Owner, nextView)
			setShardViewUserStake(native, param.ShardId, nextView, peerInfo.Owner, ownerStakeInfo)
		} else {
			return utils.BYTE_FALSE, fmt.Errorf("DeletePeer: peer %s not exist", peer)
		}
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
	if err := checkCommittingDpos(native, param.ShardId); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserStake: failed, err: %s", err)
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
	stakeAssetAddr, err := getShardStakeAssetAddr(native, param.ShardId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserStake: failed, err: %s", err)
	}
	if stakeAssetAddr == utils.OntContractAddress {
		unboundOngInfo, err := getUserUnboundOngInfo(native, param.User)
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
		setUserUnboundOngInfo(native, param.User, unboundOngInfo)
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
	if err := checkCommittingDpos(native, param.ShardId); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UnfreezeStake: failed, err: %s", err)
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
	if err := checkCommittingDpos(native, param.ShardId); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("WithdrawStake: failed, err: %s", err)
	}
	if err := utils.ValidateOwner(native, param.User); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("WithdrawStake: check witness failed, err: %s", err)
	}
	num, err := withdrawStakeAsset(native, param.ShardId, param.User)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("WithdrawStake: failed, err: %s", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	stakeAssetAddr, err := getShardStakeAssetAddr(native, param.ShardId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserStake: failed, err: %s", err)
	}
	if stakeAssetAddr == utils.OntContractAddress {
		unboundOngInfo, err := getUserUnboundOngInfo(native, param.User)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("UserStake: failed, err: %s", err)
		}
		if unboundOngInfo.Time == 0 {
			// user stake firstly
			unboundOngInfo.Time = native.Time
			unboundOngInfo.StakeAmount = num
		} else if unboundOngInfo.StakeAmount >= num {
			amount := utils.CalcUnbindOng(unboundOngInfo.StakeAmount, unboundOngInfo.Time-constants.GENESIS_BLOCK_TIMESTAMP,
				native.Time-constants.GENESIS_BLOCK_TIMESTAMP)
			unboundOngInfo.Balance += amount
			unboundOngInfo.Time = native.Time
			unboundOngInfo.StakeAmount += num
		} else {
			return utils.BYTE_FALSE, fmt.Errorf("UserStake: user stake whole amount not enough")
		}
		setUserUnboundOngInfo(native, param.User, unboundOngInfo)
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
	if err := checkCommittingDpos(native, param.ShardId); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("WithdrawStake: failed, err: %s", err)
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

func PreCommitDpos(native *native.NativeService) ([]byte, error) {
	if native.ContextRef.CallingContext().ContractAddress != utils.ShardMgmtContractAddress {
		return utils.BYTE_FALSE, fmt.Errorf("PreCommitDpos: only shard mgmt contract can invoke")
	}
	shardId, err := utils.DeserializeShardId(bytes.NewReader(native.Input))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PreCommitDpos: deserialize shard id faield, err: %s", err)
	}
	if err := checkCommittingDpos(native, shardId); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("WithdrawStake: failed, err: %s", err)
	}
	setShardCommitting(native, shardId, true)
	currentView, err := GetShardCurrentView(native, shardId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PreCommitDpos: faield, err: %s", err)
	}
	setShardView(native, shardId, &utils.ChangeView{View: uint32(currentView)})
	return utils.BYTE_TRUE, nil
}

func CommitDpos(native *native.NativeService) ([]byte, error) {
	data, err := serialization.ReadVarBytes(bytes.NewReader(native.Input))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("decode input failed, err: %s", err)
	}
	param := new(CommitDposParam)
	if err := param.Deserialization(common.NewZeroCopySource(data)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: invalid param: %s", err)
	}
	isCommitting, err := isShardCommitting(native, param.ShardId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: failed, err: %s", err)
	}
	if !isCommitting {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: shard doesn't per-commit")
	}
	setShardCommitting(native, param.ShardId, false)
	if err := commitDpos(native, param.ShardId, param.FeeAmount, param.Height, param.Hash); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func ChangeMaxAuthorization(native *native.NativeService) ([]byte, error) {
	param := new(ChangeMaxAuthorizationParam)
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ChangeMaxAuthorization: invalid param: %s", err)
	}
	if err := checkCommittingDpos(native, param.ShardId); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ChangeMaxAuthorization: failed, err: %s", err)
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
	if err := checkCommittingDpos(native, param.ShardId); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ChangeProportion: failed, err: %s", err)
	}
	if err := utils.ValidateOwner(native, param.User); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ChangeProportion: check witness failed, err: %s", err)
	}
	if param.Value.Amount > PEER_MAX_PROPORTION {
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
	unboundOngInfo, err := getUserUnboundOngInfo(native, user)
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
	setUserUnboundOngInfo(native, user, unboundOngInfo)
	if err := ont.AppCallTransfer(native, utils.OntContractAddress, contract, contract, 1); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("WithdrawOng: transfer ont failed, err: %s", err)
	}
	if err := ont.AppCallTransferFrom(native, utils.OngContractAddress, contract, utils.OntContractAddress, user,
		wholeAmount); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("WithdrawOng: transfer ong failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func GetIsCommitting(native *native.NativeService) ([]byte, error) {
	shardId, err := utils.DeserializeShardId(bytes.NewReader(native.Input))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetIsCommitting: read shardId failed, err: %s", err)
	}
	isCommitting, err := isShardCommitting(native, shardId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetIsCommitting: failed, err: %s", err)
	}
	if isCommitting {
		return utils.BYTE_TRUE, nil
	}
	return utils.BYTE_FALSE, nil
}

func GetCurrentView(native *native.NativeService) ([]byte, error) {
	shardId, err := utils.DeserializeShardId(bytes.NewReader(native.Input))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetCurrentView: read shardId failed, err: %s", err)
	}
	currentView, err := GetShardCurrentView(native, shardId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetCurrentView: failed, err: %s", err)
	}
	return ntypes.BigIntToBytes(new(big.Int).SetUint64(uint64(currentView))), nil
}

func GetPeerInfo(native *native.NativeService) ([]byte, error) {
	param := &GetPeerInfoParam{}
	if err := param.Deserialize(bytes.NewReader(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetPeerInfo: failed, err: %s", err)
	}
	peerInfo, err := GetShardViewInfo(native, param.ShardId, View(param.View))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetPeerInfo: failed, err: %s", err)
	}
	data, err := json.Marshal(peerInfo)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetPeerInfo: marshal peer info failed, err: %s", err)
	}
	return data, nil
}

// return user stake info at nearest of param view
func GetUserInfo(native *native.NativeService) ([]byte, error) {
	param := &GetUserStakeInfoParam{}
	if err := param.Deserialize(bytes.NewReader(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetUserInfo: failed, err: %s", err)
	}
	result := &UserStakeInfo{Peers: make(map[string]*UserPeerStakeInfo)}
	for i := param.View; ; i-- {
		info, err := getShardViewUserStake(native, param.ShardId, View(i), param.User)
		if err != nil {
			log.Debugf("GetUserInfo: get view %d info failed, err: %s", i, err)
			continue
		}
		if !isUserStakePeerEmpty(info) {
			result = info
			break
		}
		if i == 0 {
			break
		}
	}
	data, err := json.Marshal(result)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetUserInfo: marshal info failed, err: %s", err)
	}
	return data, nil
}
