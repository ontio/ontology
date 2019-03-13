package shard_stake

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	PEER_STAKE     = "peerInitStake"
	USER_STAKE     = "userStake"
	UNFREEZE_STAKE = "unfreezeStake"
	WITHDRAW_STAKE = "withdrawStake"
	WITHDRAW_FEE   = "withdrawFee"
)

func InitShardStake() {
	native.Contracts[utils.ShardStakeAddress] = RegisterShardStake
}

func RegisterShardStake(native *native.NativeService) {
	native.Register(PEER_STAKE, PeerInitStake)
	native.Register(USER_STAKE, UserStake)
	native.Register(UNFREEZE_STAKE, UnfreezeStake)
	native.Register(WITHDRAW_STAKE, WithdrawStake)
	native.Register(WITHDRAW_FEE, WithdrawFee)
}

func PeerInitStake(native *native.NativeService) ([]byte, error) {
	cp := new(shardmgmt.CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerInitStake: invalid cmd param: %s", err)
	}
	params := new(shardmgmt.JoinShardParam)
	if err := params.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("PeerInitStake: invalid param: %s", err)
	}
	// only call by shard mgmt contract
	if native.ContextRef.CallingContext().ContractAddress != utils.ShardMgmtContractAddress {
		return utils.BYTE_FALSE, fmt.Errorf("PeerInitStake: only shard mgmt can invoke")
	}
	pubKeyData, err := hex.DecodeString(params.PeerPubKey)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("JoinShard: decode param pub key failed, err: %s", err)
	}
	paramPubkey, err := keypair.DeserializePublicKey(pubKeyData)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("JoinShard: deserialize param pub key failed, err: %s", err)
	}
	err = peerStake(native, params.ShardID, paramPubkey, params.PeerOwner, params.StakeAmount)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("JoinShard: deserialize param pub key failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func UserStake(native *native.NativeService) ([]byte, error) {
	_, err := native.NativeCall(utils.ShardMgmtContractAddress, shardmgmt.USER_STAKE, native.Input)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserStake: call shardmgmt contarct failed, err: %s", err)
	}
	cp := new(shardmgmt.CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserStake: invalid cmd param: %s", err)
	}
	param := new(shardmgmt.UserStakeParam)
	if err := param.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserStake: invalid param: %s", err)
	}
	stakeInfo := make(map[string]uint64)
	for index, peer := range param.PeerPubKey {
		amount := param.Amount[index]
		stakeInfo[peer] = amount
	}
	err = userStake(native, param.ShardId, param.User, stakeInfo)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserStake: failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func UnfreezeStake(native *native.NativeService) ([]byte, error) {
	cp := new(shardmgmt.CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UnfreezeStake: invalid cmd param: %s", err)
	}
	param := new(UnfreezeFromShardParam)
	if err := param.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
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
	cp := new(shardmgmt.CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("WithdrawStake: invalid cmd param: %s", err)
	}
	param := new(WithdrawStakeAssetParam)
	if err := param.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
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
	cp := new(shardmgmt.CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("WithdrawFee: invalid cmd param: %s", err)
	}
	param := new(WithdrawFeeParam)
	if err := param.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
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
	return utils.BYTE_TRUE, nil
}
