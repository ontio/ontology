package shard_stake

import (
	"encoding/hex"
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native"
)

// TODO: consider peer exit scenario

// set current+2 stake info to current+1 stake info, only update view info, don't settle
func commitDpos(native *native.NativeService, shardId types.ShardID, feeInfo map[keypair.PublicKey]uint64, view View) error {
	currentView, err := getShardCurrentView(native, shardId)
	if err != nil {
		return fmt.Errorf("commitDpos: get shard %d current view failed, err: %s", shardId, err)
	}
	if view != currentView {
		return fmt.Errorf("commitDpos: the view %d not equals current view %d", view, currentView)
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

func peerStake(native *native.NativeService, id types.ShardID, peerPubKey keypair.PublicKey, peerOwner common.Address,
	amount uint64) error {
	currentView, err := getShardCurrentView(native, id)
	if err != nil {
		return fmt.Errorf("peerStake: get current view peer stake info failed, err: %s", err)
	}
	// if peer join after view 0, the stake should effective from next round
	if currentView > 0 {
		currentView++
	}
	info := &UserStakeInfo{Peers: make(map[keypair.PublicKey]*UserPeerStakeInfo)}
	pubKeyString := hex.EncodeToString(keypair.SerializePublicKey(peerPubKey))
	info.Peers[peerPubKey] = &UserPeerStakeInfo{
		PeerPubKey:  pubKeyString,
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
	peerViewInfo, ok := initViewInfo.Peers[peerPubKey]
	if ok {
		return fmt.Errorf("peerStake: peer %s has already exist", pubKeyString)
	}
	peerViewInfo = &PeerViewInfo{
		PeerPubKey:       pubKeyString,
		Owner:            peerOwner,
		WholeStakeAmount: amount,
		CanStake:         true, // default can stake asset
	}
	initViewInfo.Peers[peerPubKey] = peerViewInfo
	err = setShardViewInfo(native, id, currentView, initViewInfo)
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
		pubKeyData, err := hex.DecodeString(pubKeyString)
		if err != nil {
			return fmt.Errorf("userStake: decode pub key %s failed, err: %s", pubKeyString, err)
		}
		peer, err := keypair.DeserializePublicKey(pubKeyData)
		if err != nil {
			return fmt.Errorf("userStake: deserialize pub key %s failed, err: %s", pubKeyString, err)
		}
		currentPeerStakeInfo, ok := currentViewInfo.Peers[peer]
		if !ok {
			return fmt.Errorf("userStake: current view cannot find peer %s", pubKeyString)
		}
		nextPeerStakeInfo, ok := nextViewInfo.Peers[peer]
		if !ok {
			return fmt.Errorf("userStake: next view cannot find peer %s", pubKeyString)
		}
		if !nextPeerStakeInfo.CanStake {
			return fmt.Errorf("userStake: peer %s cannot stake", pubKeyString)
		}
		if nextPeerStakeInfo.MaxAuthorization < nextPeerStakeInfo.UserStakeAmount+amount {
			return fmt.Errorf("userStake: exceed peer %s authorization", pubKeyString)
		}
		lastUserPeerStakeInfo, ok := lastUserStakeInfo.Peers[peer]
		if !ok {
			lastUserPeerStakeInfo = &UserPeerStakeInfo{PeerPubKey: pubKeyString}
		}
		nextUserPeerStakeInfo, ok := nextUserStakeInfo.Peers[peer]
		if !ok {
			nextUserPeerStakeInfo = &UserPeerStakeInfo{PeerPubKey: pubKeyString,
				StakeAmount: lastUserPeerStakeInfo.StakeAmount, UnfreezeAmount: lastUserPeerStakeInfo.UnfreezeAmount}
		}
		lastUserPeerStakeInfo.CurrentViewStakeAmount += amount
		lastUserStakeInfo.Peers[peer] = lastUserPeerStakeInfo
		nextUserPeerStakeInfo.StakeAmount += amount
		nextUserStakeInfo.Peers[peer] = nextUserPeerStakeInfo
		currentPeerStakeInfo.CurrentViewStakeAmount += amount
		currentViewInfo.Peers[peer] = currentPeerStakeInfo
		nextPeerStakeInfo.WholeStakeAmount += amount
		nextPeerStakeInfo.UserStakeAmount += amount
		nextViewInfo.Peers[peer] = nextPeerStakeInfo
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
	currentView, err := getShardCurrentView(native, id)
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
		pubKeyData, err := hex.DecodeString(pubKeyString)
		if err != nil {
			return fmt.Errorf("unfreezeStakeAsset: decode pub key %s failed, err: %s", pubKeyString, err)
		}
		peer, err := keypair.DeserializePublicKey(pubKeyData)
		if err != nil {
			return fmt.Errorf("unfreezeStakeAsset: deserialize pub key %s failed, err: %s", pubKeyString, err)
		}
		nextPeerStakeInfo, ok := nextViewInfo.Peers[peer]
		if !ok {
			return fmt.Errorf("unfreezeStakeAsset: next view cannot find peer %s", pubKeyString)
		}
		if nextPeerStakeInfo.WholeStakeAmount < amount {
			return fmt.Errorf("unfreezeStakeAsset: peer %s stake num not enough", pubKeyString)
		}
		// use last stake to check first
		lastUserPeerStakeInfo, ok := lastUserStakeInfo.Peers[peer]
		if !ok {
			return fmt.Errorf("userStake: current view cannot find user stake peer %s", pubKeyString)
		}
		nextUserPeerStakeInfo, ok := nextUserStakeInfo.Peers[peer]
		if !ok {
			nextUserPeerStakeInfo = &UserPeerStakeInfo{PeerPubKey: pubKeyString,
				StakeAmount: lastUserPeerStakeInfo.StakeAmount, UnfreezeAmount: lastUserPeerStakeInfo.UnfreezeAmount}
		}
		if nextUserPeerStakeInfo.StakeAmount < amount {
			return fmt.Errorf("unfreezeStakeAsset: next user stake peer %s not enough", pubKeyString)
		}
		if nextPeerStakeInfo.Owner == user && minStakeAmount > nextUserPeerStakeInfo.StakeAmount-amount {
			return fmt.Errorf("unfreezeStakeAsset: peer %s owner stake amount not enough", pubKeyString)
		}
		nextUserPeerStakeInfo.StakeAmount -= amount
		nextUserPeerStakeInfo.UnfreezeAmount += amount
		nextUserStakeInfo.Peers[peer] = nextUserPeerStakeInfo
		nextPeerStakeInfo.WholeStakeAmount -= amount
		nextPeerStakeInfo.UserStakeAmount -= amount
		nextPeerStakeInfo.WholeUnfreezeAmount += amount
		nextViewInfo.Peers[peer] = nextPeerStakeInfo
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
	currentView, err := getShardCurrentView(native, id)
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
		currentPeerInfo, ok := currentViewInfo.Peers[peer]
		if !ok {
			return 0, fmt.Errorf("unfreezeStakeAsset: current view peer %s not exist", userPeerStakeInfo.PeerPubKey)
		}
		nextPeerInfo, ok := nextViewInfo.Peers[peer]
		if !ok {
			return 0, fmt.Errorf("unfreezeStakeAsset: next view peer %s not exist", userPeerStakeInfo.PeerPubKey)
		}
		currentPeerInfo.WholeUnfreezeAmount -= userPeerStakeInfo.UnfreezeAmount
		currentPeerInfo.CurrentViewStakeAmount -= userPeerStakeInfo.CurrentViewStakeAmount
		nextPeerInfo.WholeStakeAmount -= userPeerStakeInfo.CurrentViewStakeAmount
		nextPeerInfo.UserStakeAmount -= userPeerStakeInfo.CurrentViewStakeAmount
		nextPeerInfo.WholeUnfreezeAmount -= userPeerStakeInfo.UnfreezeAmount
		currentViewInfo.Peers[peer] = currentPeerInfo
		nextViewInfo.Peers[peer] = nextPeerInfo

		amount += userPeerStakeInfo.UnfreezeAmount + userPeerStakeInfo.CurrentViewStakeAmount
		nextUserPeerStakeInfo, ok := nextUserStakeInfo.Peers[peer]
		if !ok {
			nextUserPeerStakeInfo = &UserPeerStakeInfo{PeerPubKey: userPeerStakeInfo.PeerPubKey,
				StakeAmount: userPeerStakeInfo.StakeAmount,}
		} else {
			nextUserPeerStakeInfo.UnfreezeAmount -= userPeerStakeInfo.UnfreezeAmount
			nextUserPeerStakeInfo.StakeAmount -= userPeerStakeInfo.CurrentViewStakeAmount
		}
		userPeerStakeInfo.UnfreezeAmount = 0
		userPeerStakeInfo.CurrentViewStakeAmount = 0
		lastUserStakeInfo.Peers[peer] = userPeerStakeInfo
		nextUserStakeInfo.Peers[peer] = nextUserPeerStakeInfo
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
	currentView, err := getShardCurrentView(native, shardId)
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
	latestUserStakeInfo := &UserStakeInfo{Peers: make(map[keypair.PublicKey]*UserPeerStakeInfo)}
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
			if !ok {
				return 0, fmt.Errorf("withdrawFee: cannot get view %d peer %s stake info", i,
					hex.EncodeToString(keypair.SerializePublicKey(peer)))
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
	currentView, err := getShardCurrentView(native, shardId)
	if err != nil {
		return fmt.Errorf("changePeerInfo: failed, err: %s", err)
	}
	nextView := currentView + 1
	nextViewInfo, err := GetShardViewInfo(native, shardId, nextView)
	if err != nil {
		return fmt.Errorf("changePeerInfo: failed, err: %s", err)
	}
	peerInfo, pubKey, err := nextViewInfo.GetPeer(peerPubKey)
	if err != nil {
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
	nextViewInfo.Peers[pubKey] = peerInfo
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
