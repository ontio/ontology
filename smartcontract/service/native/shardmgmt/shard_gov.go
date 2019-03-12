package shardmgmt

import (
	"bytes"
	"fmt"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/global_params"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	CHANGE_MAX_AUTHORIZATION = "changeMaxAuthorization"
	CHANGE_PROPORTION        = "changeProportion" // node change proportion of stake user
	SET_MIN_STAKE            = "setMinStake"
	COMMIT_DPOS              = "commitDpos"
	USER_STAKE               = "userStake"
)

// TODO: quit node and withdraw unbound ong

func registerShardGov(native *native.NativeService) {
	native.Register(CHANGE_MAX_AUTHORIZATION, ChangeMaxAuthorization)
	native.Register(CHANGE_PROPORTION, ChangeProportion)
	native.Register(SET_MIN_STAKE, SetMinStake)
	native.Register(COMMIT_DPOS, CommitDpos)
	native.Register(USER_STAKE, UserStake)
}

func ChangeMaxAuthorization(native *native.NativeService) ([]byte, error) {
	cp := new(CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ChangeMaxAuthorization: invalid cmd param: %s", err)
	}
	params := new(ChangeMaxAuthorizationParam)
	if err := params.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ChangeMaxAuthorization: invalid param: %s", err)
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	shard, err := GetShardState(native, contract, params.ShardId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ChangeMaxAuthorization: get shard: %s", err)
	}
	shardPeerStakeInfo, paramPubkey, err := shard.GetPeerStakeInfo(params.PeerPubKey)
	if err != nil {
		return nil, fmt.Errorf("ChangeMaxAuthorization: peer not exist")
	}
	if err := utils.ValidateOwner(native, shardPeerStakeInfo.PeerOwner); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ChangeMaxAuthorization: check witness failed, err: %s", err)
	}
	shardPeerStakeInfo.MaxAuthorization = params.MaxAuthorization
	shard.Peers[paramPubkey] = shardPeerStakeInfo
	err = setShardState(native, contract, shard)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ChangeMaxAuthorization: failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func ChangeProportion(native *native.NativeService) ([]byte, error) {
	cp := new(CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ChangeProportion: invalid cmd param: %s", err)
	}
	params := new(ChangeProportionParam)
	if err := params.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ChangeProportion: invalid param: %s", err)
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	shard, err := GetShardState(native, contract, params.ShardId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ChangeProportion: get shard: %s", err)
	}
	shardPeerStakeInfo, paramPubkey, err := shard.GetPeerStakeInfo(params.PeerPubKey)
	if err != nil {
		return nil, fmt.Errorf("ChangeProportion: peer not exist")
	}
	if err := utils.ValidateOwner(native, shardPeerStakeInfo.PeerOwner); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ChangeProportion: check witness failed, err: %s", err)
	}
	shardPeerStakeInfo.Proportion = params.Proportion
	shard.Peers[paramPubkey] = shardPeerStakeInfo
	err = setShardState(native, contract, shard)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ChangeProportion: failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
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
	cp := new(CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SetMinStake: invalid cmd param: %s", err)
	}
	params := new(SetMinStakeParam)
	if err := params.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SetMinStake: invalid param: %s", err)
	}
	err = setUserMinStakeAmount(native, params.ShardId, params.Amount)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SetMinStake: failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func CommitDpos(native *native.NativeService) ([]byte, error) {
	cp := new(CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: invalid cmd param: %s", err)
	}
	params := new(CommitDposParam)
	if err := params.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: invalid param: %s", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	if err := utils.ValidateOwner(native, params.Peer); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: check witness failed, err: %s", err)
	}
	shard, err := GetShardState(native, contract, params.ShardId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: get shard: %s", err)
	}
	//check witness, only shard gas contract can call
	if native.ContextRef.CallingContext().ContractAddress != utils.ShardGasMgmtContractAddress {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: only shard gas contract can invoke")
	}
	err = divideFee(native, contract, shard, params.View, params.FeeAmount)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func UserStake(native *native.NativeService) ([]byte, error) {
	cp := new(CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserStake: invalid cmd param: %s", err)
	}
	param := new(UserStakeParam)
	if err := param.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserStake: invalid param: %s", err)
	}
	if len(param.PeerPubKey) != len(param.Amount) {
		return utils.BYTE_FALSE, fmt.Errorf("UserStake: stake peer num %d not equals stake amount num %d",
			len(param.PeerPubKey), len(param.Amount))
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	if err := utils.ValidateOwner(native, param.User); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserStake: check witness failed, err: %s", err)
	}
	if native.ContextRef.CallingContext().ContractAddress != utils.ShardStakeAddress {
		return utils.BYTE_FALSE, fmt.Errorf("UserStake: only shard stake contract can invoke")
	}
	shard, err := GetShardState(native, contract, param.ShardId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserStake: get shard: %s", err)
	}
	for index, peer := range param.PeerPubKey {
		shardPeerStakeInfo, paramPubkey, err := shard.GetPeerStakeInfo(peer)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("UserStake: failed, err: %s", err)
		}
		amount := param.Amount[index]
		if shardPeerStakeInfo.MaxAuthorization < amount+shardPeerStakeInfo.UserStakeAmount {
			return utils.BYTE_FALSE, fmt.Errorf("UserStake: larger than peer %s max authorization", peer)
		}
		shardPeerStakeInfo.UserStakeAmount += amount
		shardPeerStakeInfo.StakeAmount += amount
		shard.Peers[paramPubkey] = shardPeerStakeInfo
		err = ont.AppCallTransfer(native, shard.Config.StakeAssetAddress, param.User, contract, amount)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("UserStake: transfer stake asset failed, err: %s", err)
		}
		// TODO: update ong unbound time
	}
	err = setShardState(native, contract, shard)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserStake: update shard state failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}
