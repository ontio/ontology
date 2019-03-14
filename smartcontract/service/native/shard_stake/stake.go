package shard_stake

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/global_params"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	PEER_STAKE               = "peerInitStake"
	USER_STAKE               = "userStake"
	SET_MIN_STAKE            = "setMinStake"
	UNFREEZE_STAKE           = "unfreezeStake"
	WITHDRAW_STAKE           = "withdrawStake"
	WITHDRAW_FEE             = "withdrawFee"
	CHANGE_MAX_AUTHORIZATION = "changeMaxAuthorization"
	CHANGE_PROPORTION        = "changeProportion" // node change proportion of stake user
	COMMIT_DPOS              = "commitDpos"
)

// TODO: quit node and withdraw unbound ong

func InitShardStake() {
	native.Contracts[utils.ShardStakeAddress] = RegisterShardStake
}

func RegisterShardStake(native *native.NativeService) {
	native.Register(PEER_STAKE, PeerInitStake)
	native.Register(USER_STAKE, UserStake)
	native.Register(UNFREEZE_STAKE, UnfreezeStake)
	native.Register(WITHDRAW_STAKE, WithdrawStake)
	native.Register(SET_MIN_STAKE, SetMinStake)
	native.Register(WITHDRAW_FEE, WithdrawFee)
	native.Register(COMMIT_DPOS, CommitDpos)
	native.Register(CHANGE_MAX_AUTHORIZATION, ChangeMaxAuthorization)
	native.Register(CHANGE_PROPORTION, ChangeProportion)
}

func SetMinStake(native *native.NativeService) ([]byte, error) {
	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SetMinStake: get admin error: %v", err)
	}
	//check witness
	if err := utils.ValidateOwner(native, adminAddress); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SetMinStake: checkWitness error: %v", err)
	}
	params := new(SetMinStakeParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SetMinStake: invalid param: %s", err)
	}
	err = setNodeMinStakeAmount(native, params.ShardId, params.Amount)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SetMinStake: failed, err: %s", err)
	}
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
	pubKeyData, err := hex.DecodeString(params.PeerPubKey)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerInitStake: decode param pub key failed, err: %s", err)
	}
	paramPubkey, err := keypair.DeserializePublicKey(pubKeyData)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerInitStake: deserialize param pub key failed, err: %s", err)
	}
	err = peerStake(native, params.ShardId, paramPubkey, params.PeerOwner, params.StakeAmount)
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
	if minStakeAmount > params.StakeAmount {
		return utils.BYTE_FALSE, fmt.Errorf("PeerInitStake: stake amount should be larger than min stake amount")
	}
	// transfer stake asset
	err = ont.AppCallTransfer(native, params.StakeAssetAddr, params.PeerOwner, contract, params.StakeAmount)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerInitStake: transfer stake asset failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func UserStake(native *native.NativeService) ([]byte, error) {
	param := new(UserStakeParam)
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserStake: invalid param: %s", err)
	}
	stakeInfo := make(map[string]uint64)
	for index, peer := range param.PeerPubKey {
		amount := param.Amount[index]
		stakeInfo[peer] = amount
	}
	err := userStake(native, param.ShardId, param.User, stakeInfo)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserStake: failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func UnfreezeStake(native *native.NativeService) ([]byte, error) {
	param := new(UnfreezeFromShardParam)
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UnfreezeStake: invalid param: %s", err)
	}
	if err := utils.ValidateOwner(native, param.Address); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UnfreezeStake: check witness failed, err: %s", err)
	}
	if len(param.Amount) != len(param.PeerPubKey) {
		return utils.BYTE_FALSE, fmt.Errorf("UnfreezeStake: stake amount num not match stake peer num")
	}
	stakeInfo := make(map[string]uint64)
	for index, peer := range param.PeerPubKey {
		amount := param.Amount[index]
		stakeInfo[peer] = amount
	}
	err := unfreezeStakeAsset(native, param.ShardId, param.Address, stakeInfo)
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
	_, err := withdrawStakeAsset(native, param.ShardId, param.User)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("WithdrawStake: failed, err: %s", err)
	}
	// TODO: transfer asset
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
	_, err := withdrawFee(native, param.ShardId, param.User)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("WithdrawFee: failed, err: %s", err)
	}
	// TODO: transfer asset
	return utils.BYTE_TRUE, nil
}

func CommitDpos(native *native.NativeService) ([]byte, error) {
	param := new(CommitDposParam)
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: invalid param: %s", err)
	}
	if len(param.PeerPubKey) != len(param.Amount) {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: peer pub key num not match amount num")
	}
	feeInfo := make(map[string]uint64)
	for index, peer := range param.PeerPubKey {
		feeInfo[peer] = param.Amount[index]
	}
	err := commitDpos(native, param.ShardId, feeInfo)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func ChangeMaxAuthorization(native *native.NativeService) ([]byte, error) {
	params := new(ChangeMaxAuthorizationParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ChangeMaxAuthorization: invalid param: %s", err)
	}
	err := changePeerInfo(native, params.ShardId, params.User, params.PeerPubKey, CHANGE_MAX_AUTHORIZATION, params.Amount);
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ChangeMaxAuthorization: failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func ChangeProportion(native *native.NativeService) ([]byte, error) {
	params := new(ChangeProportionParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ChangeProportion: invalid param: %s", err)
	}
	err := changePeerInfo(native, params.ShardId, params.User, params.PeerPubKey, CHANGE_PROPORTION, params.Amount);
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ChangeProportion: failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}
