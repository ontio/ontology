package shard_stake

import (
	"encoding/hex"
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
)

func peerStake(native *native.NativeService, id types.ShardID, peerPubKey keypair.PublicKey, peerOwner common.Address,
	amount uint64) error {
	initView := shardstates.View(0)
	info := &UserStakeInfo{Peers: make(map[keypair.PublicKey]*UserPeerStakeInfo)}
	info.Peers[peerPubKey] = &UserPeerStakeInfo{StakeAmount: amount}
	err := setShardViewUserStake(native, id, initView, peerOwner, info)
	if err != nil {
		return fmt.Errorf("peerStake: set init view peer stake info failed, err: %s", err)
	}
	nextView := initView + 1
	err = setShardViewUserStake(native, id, nextView, peerOwner, info)
	if err != nil {
		return fmt.Errorf("peerStake: set next view peer stake info failed, err: %s", err)
	}
	initViewInfo, err := getShardViewInfo(native, id, initView)
	if err != nil {
		return fmt.Errorf("peerStake: get init view info failed, err: %s", err)
	}
	nextViewInfo, err := getShardViewInfo(native, id, nextView)
	if err != nil {
		return fmt.Errorf("peerStake: get next view info failed, err: %s", err)
	}
	initViewInfo.Peers[peerPubKey].WholeStakeAmount = initViewInfo.Peers[peerPubKey].WholeStakeAmount + amount
	err = setShardViewInfo(native, id, initView, initViewInfo)
	if err != nil {
		return fmt.Errorf("peerStake: update init view info failed, err: %s", err)
	}
	nextViewInfo.Peers[peerPubKey].WholeStakeAmount = initViewInfo.Peers[peerPubKey].WholeStakeAmount + amount
	err = setShardViewInfo(native, id, nextView, nextViewInfo)
	if err != nil {
		return fmt.Errorf("peerStake: update current view info failed, err: %s", err)
	}
	// update user last stake view num
	err = setUserLastStakeView(native, id, peerOwner, nextView)
	if err != nil {
		return fmt.Errorf("peerStake: failed, err: %s", err)
	}
	return nil
}

func userStake(native *native.NativeService, id types.ShardID, user common.Address, stakeInfo map[string]uint64) error {
	// get view index
	lastStakeView, err := getUserLastStakeView(native, id, user)
	if err != nil {
		return fmt.Errorf("userStake: failed, err: %s", err)
	}
	currentView, err := getShardCurrentView(native, id)
	if err != nil {
		return fmt.Errorf("userStake: failed, err: %s", err)
	}
	nextView := currentView + 1

	userStakeInfo, err := getShardViewUserStake(native, id, lastStakeView, user)
	if err != nil {
		return fmt.Errorf("userStake: failed, err: %s", err)
	}
	shardViewInfo, err := getShardViewInfo(native, id, nextView)
	if err != nil {
		return fmt.Errorf("userStake: failed, err: %s", err)
	}

	// update user current stake info
	if lastStakeView < currentView {
		err = setShardViewUserStake(native, id, currentView, user, userStakeInfo)
		if err != nil {
			return fmt.Errorf("userStake: set current view user stake info failed, err: %s", err)
		}
	} else if lastStakeView > nextView {
		return fmt.Errorf("userStake: user last stake view %d and next view %d unmatch", lastStakeView, nextView)
	}

	for pubKeyString, amount := range stakeInfo {
		pubKeyData, err := hex.DecodeString(pubKeyString)
		if err != nil {
			return fmt.Errorf("userStake: decode pub key %s failed, err: %s", pubKeyString, err)
		}
		peer, err := keypair.DeserializePublicKey(pubKeyData)
		if err != nil {
			return fmt.Errorf("userStake: deserialize pub key %s failed, err: %s", pubKeyString, err)
		}
		userPeerStakeInfo, ok := userStakeInfo.Peers[peer]
		if !ok {
			userPeerStakeInfo = &UserPeerStakeInfo{}
		}
		userPeerStakeInfo.StakeAmount = userPeerStakeInfo.StakeAmount + amount
		userStakeInfo.Peers[peer] = userPeerStakeInfo

		shardPeerStakeInfo, ok := shardViewInfo.Peers[peer]
		if !ok {
			shardPeerStakeInfo = &PeerViewInfo{}
		}
		shardPeerStakeInfo.WholeStakeAmount = shardPeerStakeInfo.WholeStakeAmount + amount
		shardViewInfo.Peers[peer] = shardPeerStakeInfo
	}
	// update shard stake info and user stake info
	err = setShardViewUserStake(native, id, nextView, user, userStakeInfo)
	if err != nil {
		return fmt.Errorf("userStake: set next view user stake info failed, err: %s", err)
	}
	err = setShardViewInfo(native, id, nextView, shardViewInfo)
	if err != nil {
		return fmt.Errorf("userStake: failed, err: %s", err)
	}

	// update user last stake view num
	err = setUserLastStakeView(native, id, user, nextView)
	if err != nil {
		return fmt.Errorf("userStake: failed, err: %s", err)
	}
	return nil
}

func userUnfreezeStakeAsset(native *native.NativeService, id types.ShardID, user common.Address, stakeInfo map[string]uint64) error {
	// get view index
	lastStakeView, err := getUserLastStakeView(native, id, user)
	if err != nil {
		return fmt.Errorf("userUnfreezeStakeAsset: failed, err: %s", err)
	}
	currentView, err := getShardCurrentView(native, id)
	if err != nil {
		return fmt.Errorf("userUnfreezeStakeAsset: failed, err: %s", err)
	}
	nextView := currentView + 1

	// read user stake info and view stake info
	userStakeInfo, err := getShardViewUserStake(native, id, lastStakeView, user)
	if err != nil {
		return fmt.Errorf("userUnfreezeStakeAsset: failed, err: %s", err)
	}
	shardViewInfo, err := getShardViewInfo(native, id, nextView)
	if err != nil {
		return fmt.Errorf("userUnfreezeStakeAsset: failed, err: %s", err)
	}
	if lastStakeView < currentView {
		// update current user stake info
		err = setShardViewUserStake(native, id, currentView, user, userStakeInfo)
		if err != nil {
			return fmt.Errorf("userUnfreezeStakeAsset: set current view user stake info failed, err: %s", err)
		}
	} else if lastStakeView > nextView {
		return fmt.Errorf("userUnfreezeStakeAsset: user last stake view %d and next view %d unmatch",
			lastStakeView, nextView)
	}
	for pubKeyString, amount := range stakeInfo {
		pubKeyData, err := hex.DecodeString(pubKeyString)
		if err != nil {
			return fmt.Errorf("userUnfreezeStakeAsset: decode param pub key failed, err: %s", err)
		}
		peer, err := keypair.DeserializePublicKey(pubKeyData)
		if err != nil {
			return fmt.Errorf("userUnfreezeStakeAsset: deserialize param pub key failed, err: %s", err)
		}
		userPeerStakeInfo, ok := userStakeInfo.Peers[peer]
		if !ok {
			userPeerStakeInfo = &UserPeerStakeInfo{}
		}
		if userPeerStakeInfo.StakeAmount < amount {
			return fmt.Errorf("userUnfreezeStakeAsset: stake amount %d not enough",
				userPeerStakeInfo.StakeAmount)
		}
		userPeerStakeInfo.StakeAmount -= amount
		userPeerStakeInfo.UnfreezeAmount += amount
		userStakeInfo.Peers[peer] = userPeerStakeInfo

		shardPeerStakeInfo, ok := shardViewInfo.Peers[peer]
		if !ok {
			shardPeerStakeInfo = &PeerViewInfo{}
		}
		if shardPeerStakeInfo.WholeStakeAmount < amount {
			return fmt.Errorf("userUnfreezeStakeAsset: whole stake amount %d not enough",
				shardPeerStakeInfo.WholeStakeAmount)
		}
		shardPeerStakeInfo.WholeStakeAmount -= amount
		shardPeerStakeInfo.WholeUnfreezeAmount += amount
		shardViewInfo.Peers[peer] = shardPeerStakeInfo
	}

	// update next stake info
	err = setShardViewUserStake(native, id, nextView, user, userStakeInfo)
	if err != nil {
		return fmt.Errorf("userUnfreezeStakeAsset: failed, err: %s", err)
	}
	err = setShardViewInfo(native, id, nextView, shardViewInfo)
	if err != nil {
		return fmt.Errorf("userUnfreezeStakeAsset: failed, err: %s", err)
	}

	// update user last stake view num
	err = setUserLastStakeView(native, id, user, nextView)
	if err != nil {
		return fmt.Errorf("userUnfreezeStakeAsset: failed, err: %s", err)
	}
	return nil
}
