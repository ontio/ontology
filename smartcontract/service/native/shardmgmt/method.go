package shardmgmt

import (
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func divideFee(native *native.NativeService, contract common.Address, shard *shardstates.ShardState,
	view shardstates.View, feeAmount uint64) error {
	isViewDivided, err := isViewDivided(native, contract, shard.ShardID, view)
	if err != nil {
		return fmt.Errorf("divideFee: failed, err: %s", err)
	}
	if isViewDivided {
		return nil
	}
	consensusStakeAmount, candidateStakeAmount := uint64(0), uint64(0)
	for _, stakeInfo := range shard.Peers {
		if stakeInfo.NodeType == shardstates.CONSENSUS_NODE {
			consensusStakeAmount += stakeInfo.StakeAmount
		} else {
			candidateStakeAmount += stakeInfo.StakeAmount
		}
	}
	for _, stakeInfo := range shard.Peers {
		nodeAmount := uint64(0)
		userAmount := uint64(0)
		if stakeInfo.NodeType == shardstates.CONSENSUS_NODE {
			nodeAmount = feeAmount / 2 * stakeInfo.StakeAmount * (100 - stakeInfo.Proportion) / 100 / consensusStakeAmount
			userAmount = feeAmount / 2 * stakeInfo.StakeAmount * stakeInfo.Proportion / 100 / consensusStakeAmount
		} else {
			candidateStakeAmount += stakeInfo.StakeAmount
			nodeAmount = feeAmount / 2 * (stakeInfo.StakeAmount * (100 - stakeInfo.Proportion) / candidateStakeAmount)
			userAmount = feeAmount / 2 * stakeInfo.StakeAmount * stakeInfo.Proportion / 100 / candidateStakeAmount
		}
		err = ont.AppCallTransfer(native, utils.OngContractAddress, contract, stakeInfo.PeerOwner,
			nodeAmount)
		if err != nil {
			return fmt.Errorf("divideFee: transfer node fee failed, peer %s, err: %s", stakeInfo.PeerPubKey, err)
		}
		// TODO: call shard stake contract method
		err = ont.AppCallTransfer(native, utils.OngContractAddress, contract, utils.ShardStakeAddress,
			userAmount)
		if err != nil {
			return fmt.Errorf("divideFee: transfer user fee failed, peer %s, err: %s", stakeInfo.PeerPubKey, err)
		}
	}
	err = setViewDivided(native, contract, shard.ShardID, view)
	if err != nil {
		return fmt.Errorf("divideFee: failed, err: %s", err)
	}
	return nil
}
