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
	PEER_STAKE         = "peerInitStake"
	USER_STAKE         = "userStake"
	SET_PEER_MIN_STAKE = "setMinStake"
)

func InitShardStake() {
	native.Contracts[utils.ShardStakeAddress] = RegisterShardStake
}

func RegisterShardStake(native *native.NativeService) {
	native.Register(PEER_STAKE, PeerInitStake)
	native.Register(USER_STAKE, UserStake)

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
	return utils.BYTE_TRUE, nil
}

func CommitDpos(native *native.NativeService) ([]byte, error) {
	return utils.BYTE_TRUE, nil
}
