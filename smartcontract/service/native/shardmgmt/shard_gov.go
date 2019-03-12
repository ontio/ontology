package shardmgmt

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	CHANGE_MAX_AUTHORIZATION = "changeMaxAuthorization"
	CHANGE_PROPORTION        = "changeProportion" // node change proportion of stake user
	COMMIT_DPOS              = "commitDpos"
)

func registerShardGov(native *native.NativeService) {
	native.Register(CHANGE_MAX_AUTHORIZATION, ChangeMaxAuthorization)
	native.Register(CHANGE_PROPORTION, ChangeProportion)
	native.Register(COMMIT_DPOS, CommitDpos)
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
	pubKeyData, err := hex.DecodeString(params.PeerPubKey)
	if err != nil {
		return nil, fmt.Errorf("ChangeMaxAuthorization: decode param pub key failed, err: %s", err)
	}
	paramPubkey, err := keypair.DeserializePublicKey(pubKeyData)
	if err != nil {
		return nil, fmt.Errorf("ChangeMaxAuthorization: deserialize param pub key failed, err: %s", err)
	}
	shardPeerStakeInfo, ok := shard.Peers[paramPubkey]
	if !ok {
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
	pubKeyData, err := hex.DecodeString(params.PeerPubKey)
	if err != nil {
		return nil, fmt.Errorf("ChangeProportion: decode param pub key failed, err: %s", err)
	}
	paramPubkey, err := keypair.DeserializePublicKey(pubKeyData)
	if err != nil {
		return nil, fmt.Errorf("ChangeProportion: deserialize param pub key failed, err: %s", err)
	}
	shardPeerStakeInfo, ok := shard.Peers[paramPubkey]
	if !ok {
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
