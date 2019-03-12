package shard_stake

import (
	"fmt"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	USER_STAKE = "userStake"
)

func InitShardStake() {
	native.Contracts[utils.ShardStakeAddress] = RegisterShardStake
}

func RegisterShardStake(native *native.NativeService) {
	native.Register(USER_STAKE, UserStake)

}

func PeerInitStake(native *native.NativeService) ([]byte, error) {

}

func UserStake(native *native.NativeService) ([]byte, error) {
	_, err := native.NativeCall(utils.ShardMgmtContractAddress, shardmgmt.USER_STAKE, native.Input)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UserStake: call shardmgmt contarct failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func CommitDpos(native *native.NativeService) ([]byte, error) {
	return utils.BYTE_TRUE, nil
}
