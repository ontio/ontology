package shard_stake

import (
	"encoding/hex"
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native"
	"strings"
)

// TODO: consider peer exit scenario

// set current+2 stake info to current+1 stake info, only update view info, don't settle
func commitDpos(native *native.NativeService, shardId types.ShardID, feeInfo map[string]uint64) error {
	currentView, err := GetShardCurrentView(native, shardId)
	if err != nil {
		return fmt.Errorf("commitDpos: get shard %d current view failed, err: %s", shardId, err)
	}
	currentViewInfo, err := GetShardViewInfo(native, shardId, currentView)
	if err != nil {
		return fmt.Errorf("commitDpos: get shard %d current view info failed, err: %s", shardId, err)
	}
	for peer, feeAmount := range feeInfo {
		peerInfo, ok := currentViewInfo.Peers[peer]
		if !ok {
			return fmt.Errorf("commitDpos: peer %s not exist at current view",
				hex.EncodeToString(keypair.SerializePublicKey(peer)))
		}
		peerInfo.WholeFee = feeAmount
		peerInfo.FeeBalance = feeAmount
		currentViewInfo.Peers[peer] = peerInfo
	}
	err = setShardViewInfo(native, shardId, currentView, currentViewInfo)
	if err != nil {
		return fmt.Errorf("commitDpos: update shard %d view info failed, err: %s", shardId, err)
	}
	nextView := currentView + 1
	nextTwoView := nextView + 1
	nextViewInfo, err := GetShardViewInfo(native, shardId, nextView)
	if err != nil {
		return fmt.Errorf("commitDpos: get next view info failed, err: %s", err)
	}
	if err := setShardViewInfo(native, shardId, nextTwoView, nextViewInfo); err != nil {
		return fmt.Errorf("commitDpos: set next two view info failed, err: %s", err)
	}
	err = setShardView(native, shardId, nextView)
	if err != nil {
		return fmt.Errorf("commitDpos: update shard %d view failed, err: %s", shardId, err)
	}
	return nil
}

func peerStake(native *native.NativeService, id types.ShardID, peerPubKey string, peerOwner common.Address,
	amount uint64) error {
	currentView, err := GetShardCurrentView(native, id)
	if err != nil {
		return fmt.Errorf("peerStake: get current view peer stake info failed, err: %s", err)
	}
	// if peer join after view 0, the stake should effective from next round
	if currentView > 0 {
		currentView++
	}
	info := &UserStakeInfo{Peers: make(map[string]*UserPeerStakeInfo)}
	info.Peers[peerPubKey] = &UserPeerStakeInfo{
		PeerPubKey:  peerPubKey,
		StakeAmount: amount,
	}
	err = setShardViewUserStake(native, id, currentView, peerOwner, info)
	if err != nil {
		return fmt.Errorf("peerStake: set init view peer stake info failed, err: %s", err)
	}
	nextView := currentView + 1
	err = setShardViewUserStake(native, id, nextView, peerOwner, info)
	if err != nil {
		return fmt.Errorf("peerStake: set next view peer stake info failed, err: %s", err)
	}
	initViewInfo, err := GetShardViewInfo(native, id, currentView)
	if err != nil {
		return fmt.Errorf("peerStake: get init view info failed, err: %s", err)
	}
	nextViewInfo, err := GetShardViewInfo(native, id, nextView)
	if err != nil {
		return fmt.Errorf("peerStake: get next view info failed, err: %s", err)
	}
	if initViewInfo.Peers == nil {
		initViewInfo.Peers = make(map[string]*PeerViewInfo)
		nextViewInfo.Peers = make(map[string]*PeerViewInfo)
	}
	peerViewInfo, ok := initViewInfo.Peers[peerPubKey]
	if ok {
		return fmt.Errorf("peerStake: peer %s has already exist", peerPubKey)
	}
	peerViewInfo = &PeerViewInfo{
		PeerPubKey:       peerPubKey,
		Owner:            peerOwner,
		WholeStakeAmount: amount,
		CanStake:         true, // default can stake asset
	}
	initViewInfo.Peers[peerPubKey] = peerViewInfo
	err = setShardViewInfo(native, id, currentView, initViewInfo)
	if err != nil {
		return fmt.Errorf("peerStake: update init view info failed, err: %s", err)
	}
	nextViewInfo.Peers[peerPubKey] = peerViewInfo
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
	currentView, err := GetShardCurrentView(native, id)
	if err != nil {
		return fmt.Errorf("userStake: failed, err: %s", err)
	}
	nextView := currentView + 1
	if lastStakeView > nextView {
		return fmt.Errorf("userStake: user last stake view %d and next view %d unmatch", lastStakeView, nextView)
	} else if lastStakeView == nextView {
		lastStakeView = currentView
	}
	lastUserStakeInfo, err := getShardViewUserStake(native, id, lastStakeView, user)
	if err != nil {
		return fmt.Errorf("userStake: get user last stake info failed, err: %s", err)
	}
	nextUserStakeInfo, err := getShardViewUserStake(native, id, nextView, user)
	if err != nil {
		return fmt.Errorf("userStake: get user next stake info failed, err: %s", err)
	}
	currentViewInfo, err := GetShardViewInfo(native, id, currentView)
	if err != nil {
		return fmt.Errorf("userStake: get current view info failed, err: %s", err)
	}
	nextViewInfo, err := GetShardViewInfo(native, id, nextView)
	if err != nil {
		return fmt.Errorf("userStake: get next view info failed, err: %s", err)
	}
	for pubKeyString, amount := range stakeInfo {
		currentPeerStakeInfo, ok := currentViewInfo.Peers[pubKeyString]
		if !ok {
			return fmt.Errorf("userStake: current view cannot find peer %s", pubKeyString)
		}
		nextPeerStakeInfo, ok := nextViewInfo.Peers[pubKeyString]
		if !ok {
			return fmt.Errorf("userStake: next view cannot find peer %s", pubKeyString)
		}
		if !nextPeerStakeInfo.CanStake {
			return fmt.Errorf("userStake: peer %s cannot stake", pubKeyString)
		}
		if nextPeerStakeInfo.MaxAuthorization < nextPeerStakeInfo.UserStakeAmount+amount {
			return fmt.Errorf("userStake: exceed peer %s authorization", pubKeyString)
		}
		lastUserPeerStakeInfo, ok := lastUserStakeInfo.Peers[pubKeyString]
		if !ok {
			lastUserPeerStakeInfo = &UserPeerStakeInfo{PeerPubKey: pubKeyString}
		}
		nextUserPeerStakeInfo, ok := nextUserStakeInfo.Peers[pubKeyString]
		if !ok {
			nextUserPeerStakeInfo = &UserPeerStakeInfo{PeerPubKey: pubKeyString,
				StakeAmount: lastUserPeerStakeInfo.StakeAmount, UnfreezeAmount: lastUserPeerStakeInfo.UnfreezeAmount}
		}
		lastUserPeerStakeInfo.CurrentViewStakeAmount += amount
		lastUserStakeInfo.Peers[pubKeyString] = lastUserPeerStakeInfo
		nextUserPeerStakeInfo.StakeAmount += amount
		nextUserStakeInfo.Peers[pubKeyString] = nextUserPeerStakeInfo
		currentPeerStakeInfo.CurrentViewStakeAmount += amount
		currentViewInfo.Peers[pubKeyString] = currentPeerStakeInfo
		nextPeerStakeInfo.WholeStakeAmount += amount
		nextPeerStakeInfo.UserStakeAmount += amount
		nextViewInfo.Peers[pubKeyString] = nextPeerStakeInfo
	}
	if err := setUserLastStakeView(native, id, user, nextView); err != nil {
		return fmt.Errorf("userStake: failed, err: %s", err)
	}
	if err := setShardViewUserStake(native, id, currentView, user, lastUserStakeInfo); err != nil {
		return fmt.Errorf("userStake: udpate current view user stake info failed, err: %s", err)
	}
	if err := setShardViewUserStake(native, id, nextView, user, nextUserStakeInfo); err != nil {
		return fmt.Errorf("userStake: udpate next view user stake info failed, err: %s", err)
	}
	if err := setShardViewInfo(native, id, currentView, currentViewInfo); err != nil {
		return fmt.Errorf("userStake: udpate current view info failed, err: %s", err)
	}
	if err := setShardViewInfo(native, id, nextView, nextViewInfo); err != nil {
		return fmt.Errorf("userStake: udpate current view info failed, err: %s", err)
	}
	return nil
}

func unfreezeStakeAsset(native *native.NativeService, id types.ShardID, user common.Address, stakeInfo map[string]uint64) error {
	// get view index
	lastStakeView, err := getUserLastStakeView(native, id, user)
	if err != nil {
		return fmt.Errorf("unfreezeStakeAsset: failed, err: %s", err)
	}
	currentView, err := GetShardCurrentView(native, id)
	if err != nil {
		return fmt.Errorf("unfreezeStakeAsset: failed, err: %s", err)
	}
	nextView := currentView + 1
	if lastStakeView > nextView {
		return fmt.Errorf("unfreezeStakeAsset: user last stake view %d and next view %d unmatch",
			lastStakeView, nextView)
	} else if lastStakeView == nextView {
		lastStakeView = currentView
	}
	lastUserStakeInfo, err := getShardViewUserStake(native, id, lastStakeView, user)
	if err != nil {
		return fmt.Errorf("unfreezeStakeAsset: get user last stake info failed, err: %s", err)
	}
	nextUserStakeInfo, err := getShardViewUserStake(native, id, nextView, user)
	if err != nil {
		return fmt.Errorf("unfreezeStakeAsset: get user next stake info failed, err: %s", err)
	}
	if isUserStakePeerEmpty(lastUserStakeInfo) || isUserStakePeerEmpty(nextUserStakeInfo) {
		return fmt.Errorf("userStake: user stake peer info is empty")
	}
	nextViewInfo, err := GetShardViewInfo(native, id, nextView)
	if err != nil {
		return fmt.Errorf("unfreezeStakeAsset: get next view info failed, err: %s", err)
	}
	minStakeAmount, err := GetNodeMinStakeAmount(native, id)
	if err != nil {
		return fmt.Errorf("unfreezeStakeAsset: failed, err: %s", err)
	}
	for pubKeyString, amount := range stakeInfo {
		lastUserPeerStakeInfo, ok := lastUserStakeInfo.Peers[pubKeyString]
		if !ok {
			return fmt.Errorf("userStake: current view cannot find user stake peer %s", pubKeyString)
		}
		nextUserPeerStakeInfo, ok := nextUserStakeInfo.Peers[pubKeyString]
		if !ok {
			nextUserPeerStakeInfo = &UserPeerStakeInfo{PeerPubKey: pubKeyString,
				StakeAmount: lastUserPeerStakeInfo.StakeAmount, UnfreezeAmount: lastUserPeerStakeInfo.UnfreezeAmount}
		}
		if nextUserPeerStakeInfo.StakeAmount < amount {
			return fmt.Errorf("unfreezeStakeAsset: next user stake peer %s not enough", pubKeyString)
		}
		nextPeerStakeInfo, ok := nextViewInfo.Peers[pubKeyString]
		if !ok { // peer has already exit consensus and deleted
			nextUserPeerStakeInfo.UnfreezeAmount += nextUserPeerStakeInfo.StakeAmount
			nextUserPeerStakeInfo.StakeAmount = 0
		} else if nextPeerStakeInfo.WholeStakeAmount < amount {
			return fmt.Errorf("unfreezeStakeAsset: peer %s stake num not enough", pubKeyString)
		} else if nextPeerStakeInfo.Owner == user && minStakeAmount > nextUserPeerStakeInfo.StakeAmount-amount {
			return fmt.Errorf("unfreezeStakeAsset: peer %s owner stake amount not enough", pubKeyString)
		} else {
			nextUserPeerStakeInfo.StakeAmount -= amount
			nextUserPeerStakeInfo.UnfreezeAmount += amount
			nextPeerStakeInfo.WholeStakeAmount -= amount
			nextPeerStakeInfo.UserStakeAmount -= amount
			nextPeerStakeInfo.WholeUnfreezeAmount += amount
			nextViewInfo.Peers[pubKeyString] = nextPeerStakeInfo
		}
		nextUserStakeInfo.Peers[pubKeyString] = nextUserPeerStakeInfo
	}
	if err := setUserLastStakeView(native, id, user, nextView); err != nil {
		return fmt.Errorf("unfreezeStakeAsset: failed, err: %s", err)
	}
	// update user stake info from last to current
	if err := setShardViewUserStake(native, id, currentView, user, lastUserStakeInfo); err != nil {
		return fmt.Errorf("unfreezeStakeAsset: udpate current view user stake info failed, err: %s", err)
	}
	if err := setShardViewUserStake(native, id, nextView, user, nextUserStakeInfo); err != nil {
		return fmt.Errorf("unfreezeStakeAsset: udpate next view user stake info failed, err: %s", err)
	}
	if err := setShardViewInfo(native, id, nextView, nextViewInfo); err != nil {
		return fmt.Errorf("unfreezeStakeAsset: udpate current view info failed, err: %s", err)
	}
	return nil
}

// return withdraw amount
func withdrawStakeAsset(native *native.NativeService, id types.ShardID, user common.Address) (uint64, error) {
	// get view index
	lastStakeView, err := getUserLastStakeView(native, id, user)
	if err != nil {
		return 0, fmt.Errorf("unfreezeStakeAsset: failed, err: %s", err)
	}
	currentView, err := GetShardCurrentView(native, id)
	if err != nil {
		return 0, fmt.Errorf("unfreezeStakeAsset: failed, err: %s", err)
	}
	nextView := currentView + 1
	if lastStakeView > nextView {
		return 0, fmt.Errorf("unfreezeStakeAsset: user last stake view %d and next view %d unmatch",
			lastStakeView, nextView)
	} else if lastStakeView == nextView {
		lastStakeView = currentView
	}
	lastUserStakeInfo, err := getShardViewUserStake(native, id, lastStakeView, user)
	if err != nil {
		return 0, fmt.Errorf("unfreezeStakeAsset: get user last stake info failed, err: %s", err)
	}
	if isUserStakePeerEmpty(lastUserStakeInfo) {
		return 0, fmt.Errorf("userStake: user stake peer info is empty")
	}
	nextUserStakeInfo, err := getShardViewUserStake(native, id, nextView, user)
	if err != nil {
		return 0, fmt.Errorf("unfreezeStakeAsset: get user next stake info failed, err: %s", err)
	}
	currentViewInfo, err := GetShardViewInfo(native, id, currentView)
	if err != nil {
		return 0, fmt.Errorf("userStake: get current view info failed, err: %s", err)
	}
	nextViewInfo, err := GetShardViewInfo(native, id, nextView)
	if err != nil {
		return 0, fmt.Errorf("unfreezeStakeAsset: get next view info failed, err: %s", err)
	}
	amount := uint64(0)
	for peer, userPeerStakeInfo := range lastUserStakeInfo.Peers {
		if nextPeerInfo, ok := nextViewInfo.Peers[peer]; ok {
			nextPeerInfo.WholeStakeAmount -= userPeerStakeInfo.CurrentViewStakeAmount
			nextPeerInfo.UserStakeAmount -= userPeerStakeInfo.CurrentViewStakeAmount
			nextPeerInfo.WholeUnfreezeAmount -= userPeerStakeInfo.UnfreezeAmount
			nextViewInfo.Peers[peer] = nextPeerInfo
			nextUserPeerStakeInfo, ok := nextUserStakeInfo.Peers[peer]
			if !ok {
				nextUserPeerStakeInfo = &UserPeerStakeInfo{PeerPubKey: userPeerStakeInfo.PeerPubKey,
					StakeAmount: userPeerStakeInfo.StakeAmount,}
			} else {
				nextUserPeerStakeInfo.UnfreezeAmount -= userPeerStakeInfo.UnfreezeAmount
				nextUserPeerStakeInfo.StakeAmount -= userPeerStakeInfo.CurrentViewStakeAmount
			}
			nextUserStakeInfo.Peers[peer] = nextUserPeerStakeInfo
		} else {
			delete(nextUserStakeInfo.Peers, peer)
		}

		amount += userPeerStakeInfo.UnfreezeAmount + userPeerStakeInfo.CurrentViewStakeAmount
		if currentPeerInfo, ok := currentViewInfo.Peers[peer]; ok {
			currentPeerInfo.WholeUnfreezeAmount -= userPeerStakeInfo.UnfreezeAmount
			currentPeerInfo.CurrentViewStakeAmount -= userPeerStakeInfo.CurrentViewStakeAmount
			currentViewInfo.Peers[peer] = currentPeerInfo
			userPeerStakeInfo.UnfreezeAmount = 0
			userPeerStakeInfo.CurrentViewStakeAmount = 0
			lastUserStakeInfo.Peers[peer] = userPeerStakeInfo
		} else {
			delete(lastUserStakeInfo.Peers, peer)
		}
	}
	if err := setUserLastStakeView(native, id, user, nextView); err != nil {
		return 0, fmt.Errorf("unfreezeStakeAsset: failed, err: %s", err)
	}
	// update user stake info from last to current
	if err := setShardViewUserStake(native, id, currentView, user, lastUserStakeInfo); err != nil {
		return 0, fmt.Errorf("unfreezeStakeAsset: udpate current view user stake info failed, err: %s", err)
	}
	if err := setShardViewUserStake(native, id, nextView, user, nextUserStakeInfo); err != nil {
		return 0, fmt.Errorf("unfreezeStakeAsset: udpate next view user stake info failed, err: %s", err)
	}
	if err := setShardViewInfo(native, id, nextView, nextViewInfo); err != nil {
		return 0, fmt.Errorf("unfreezeStakeAsset: udpate current view info failed, err: %s", err)
	}
	return amount, nil
}

// return the amount that user could withdraw
func withdrawFee(native *native.NativeService, shardId types.ShardID, user common.Address) (uint64, error) {
	userWithdrawView, err := getUserLastWithdrawView(native, shardId, user)
	if err != nil {
		return 0, fmt.Errorf("withdrawFee: failed, err: %s", err)
	}
	currentView, err := GetShardCurrentView(native, shardId)
	if err != nil {
		return 0, fmt.Errorf("withdrawFee: failed, err: %s", err)
	}
	if currentView == 0 {
		return 0, fmt.Errorf("withdrawFee: init view not support dividends")
	}
	// withdraw view at [userWithdrawView+1, currentView)
	dividends := uint64(0)
	i := userWithdrawView
	count := 0
	latestUserStakeInfo := &UserStakeInfo{Peers: make(map[string]*UserPeerStakeInfo)}
	lastStakeView, err := getUserLastStakeView(native, shardId, user)
	if err != nil {
		return 0, fmt.Errorf("withdrawFee: failed, err: %s", err)
	}
	if lastStakeView <= userWithdrawView {
		latestUserStakeInfo, err = getShardViewUserStake(native, shardId, lastStakeView, user)
		if err != nil {
			return 0, fmt.Errorf("withdrawFee: get user latest view stake info failed, err: %s", err)
		}
	}
	for ; i < currentView && count < USER_MAX_WITHDRAW_VIEW; i++ {
		userStake, err := getShardViewUserStake(native, shardId, i, user)
		if err != nil {
			return 0, fmt.Errorf("withdrawFee: failed, view %d, err: %s", i, err)
		}
		if !isUserStakePeerEmpty(userStake) {
			if !isUserStakePeerEmpty(latestUserStakeInfo) {
				continue
			} else {
				userStake = latestUserStakeInfo
			}
		}
		viewStake, err := GetShardViewInfo(native, shardId, i)
		if err != nil {
			return 0, fmt.Errorf("withdrawFee: failed, view %d, err: %s", i, err)
		}
		for peer, info := range userStake.Peers {
			peerStakeInfo, ok := viewStake.Peers[peer]
			if !ok { // peer has already exit consensus
				continue
			}
			if peerStakeInfo.FeeBalance == 0 {
				continue
			}
			peerDivide := info.StakeAmount * peerStakeInfo.WholeFee / peerStakeInfo.WholeStakeAmount
			peerStakeInfo.FeeBalance = peerStakeInfo.FeeBalance - peerDivide
			viewStake.Peers[peer] = peerStakeInfo
			dividends += peerDivide
		}
		err = setShardViewInfo(native, shardId, i, viewStake)
		if err != nil {
			return 0, fmt.Errorf("withdrawFee: failed, view %d, err: %s", i, err)
		}
		count++
		latestUserStakeInfo = userStake
	}
	err = setUserLastWithdrawView(native, shardId, user, i)
	if err != nil {
		return 0, fmt.Errorf("withdrawFee: failed, view %d, err: %s", i, err)
	}
	err = setShardViewUserStake(native, shardId, i, user, latestUserStakeInfo)
	if err != nil {
		return 0, fmt.Errorf("withdrawFee: failed, view %d, err: %s", i, err)
	}
	return dividends, nil
}

// change peer max authorization and proportion
func changePeerInfo(native *native.NativeService, shardId types.ShardID, peerOwner common.Address, peerPubKey string,
	methodName string, amount uint64) error {
	currentView, err := GetShardCurrentView(native, shardId)
	if err != nil {
		return fmt.Errorf("changePeerInfo: failed, err: %s", err)
	}
	nextView := currentView + 1
	nextViewInfo, err := GetShardViewInfo(native, shardId, nextView)
	if err != nil {
		return fmt.Errorf("changePeerInfo: failed, err: %s", err)
	}
	peerInfo, ok := nextViewInfo.Peers[strings.ToLower(peerPubKey)]
	if !ok {
		return fmt.Errorf("changePeerInfo: failed, err: %s", err)
	}
	if peerInfo.Owner != peerOwner {
		return fmt.Errorf("changePeerInfo: peer owner not match")
	}
	switch methodName {
	case CHANGE_MAX_AUTHORIZATION:
		peerInfo.MaxAuthorization = amount
	case CHANGE_PROPORTION:
		peerInfo.Proportion = amount
	default:
		return fmt.Errorf("changePeerInfo: unsupport change field")
	}
	nextViewInfo.Peers[peerPubKey] = peerInfo
	if err := setShardViewInfo(native, shardId, nextView, nextViewInfo); err != nil {
		return fmt.Errorf("changePeerInfo: field, err: %s", err)
	}
	return nil
}

func isUserStakePeerEmpty(info *UserStakeInfo) bool {
	if info.Peers == nil || len(info.Peers) == 0 {
		return false
	}
	for _, stakeInfo := range info.Peers {
		if stakeInfo.StakeAmount != 0 || stakeInfo.UnfreezeAmount != 0 {
			return true
		}
	}
	return false
}
