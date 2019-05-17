/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package shard_stake

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

// set current+1 stake info to current stake info, only update view info, don't settle
func commitDpos(native *native.NativeService, param *CommitDposParam) error {
	shardId := param.ShardId
	currentChangeVIew, err := GetShardCurrentChangeView(native, shardId)
	if err != nil {
		return fmt.Errorf("commitDpos: failed, err: %s", shardId, err)
	}
	if param.Height <= currentChangeVIew.Height {
		return fmt.Errorf("commitDpos: param height unmatch")
	}
	currentView, err := GetShardCurrentViewIndex(native, shardId)
	if err != nil {
		return fmt.Errorf("commitDpos: failed, err: %s", shardId, err)
	}
	lastView := currentView - 1
	if err := handleDebt(native, View(lastView), param); err != nil {
		return fmt.Errorf("commitDpos: failed, err: %s", err)
	}
	feeAmount, shardHeight, shardBlockHash := param.FeeAmount, param.Height, param.Hash
	lastViewInfo, err := GetShardViewInfo(native, shardId, lastView)
	if err != nil {
		return fmt.Errorf("commitDpos: get shard %d last view info failed, err: %s", shardId, err)
	}
	currentViewInfo, err := GetShardViewInfo(native, shardId, currentView)
	if err != nil {
		return fmt.Errorf("commitDpos: get next view info failed, err: %s", err)
	}
	// empty currentViewInfo means that there are no pre-commit dpos
	if currentViewInfo.Peers == nil || len(currentViewInfo.Peers) == 0 {
		return fmt.Errorf("commitDpos: current view info is empty")
	}
	// update next view info
	setShardViewInfo(native, shardId, currentView+1, currentViewInfo)
	// settle current fee
	feeInfo := calPeerFee(lastViewInfo, feeAmount)
	for _, info := range feeInfo {
		peer := strings.ToLower(info.PeerPubKey)
		feeAmount := info.Amount
		peerInfo, ok := lastViewInfo.Peers[peer]
		if !ok {
			return fmt.Errorf("commitDpos: peer %s not exist at current view",
				hex.EncodeToString(keypair.SerializePublicKey(peer)))
		}
		peerInfo.WholeFee = feeAmount
		peerInfo.FeeBalance = feeAmount
		lastViewInfo.Peers[peer] = peerInfo
	}
	setShardViewInfo(native, shardId, lastView, lastViewInfo)
	shardView := &utils.ChangeView{
		View:   uint32(currentView),
		Height: shardHeight,
		TxHash: shardBlockHash,
	}
	// commit dpos
	setShardView(native, shardId, shardView)
	return nil
}

func handleDebt(native *native.NativeService, view View, param *CommitDposParam) error {
	wholeDebt := uint64(0)
	for debtShard, viewFeeInfo := range param.Debt {
		for view, fee := range viewFeeInfo {
			debtShardViewInfo, err := GetShardViewInfo(native, debtShard, View(view))
			if err != nil {
				return fmt.Errorf("handleDebt: failed, err: %s", err)
			}
			peerSplitFeeInfo := calPeerFee(debtShardViewInfo, fee)
			for _, amount := range peerSplitFeeInfo {
				debtShardViewInfo.Peers[amount.PeerPubKey].WholeFee += amount.Amount
				debtShardViewInfo.Peers[amount.PeerPubKey].FeeBalance += amount.Amount
			}
			setShardViewInfo(native, debtShard, View(view), debtShardViewInfo)
			wholeDebt += fee
		}
	}
	if param.FeeAmount < wholeDebt {
		return fmt.Errorf("handleDebt: whole fee not enough")
	}
	param.FeeAmount -= wholeDebt
	xshardFeeInfo := &XShardFeeInfo{
		Debt:   param.Debt,
		Income: param.Income,
	}
	setXShardFeeInfo(native, param.ShardId, view, xshardFeeInfo)
	return nil
}

// TODO: candidate node and consensus node different rate
func calPeerFee(viewInfo *ViewInfo, wholeFee uint64) []*PeerAmount {
	feeInfo := make([]*PeerAmount, 0)
	wholeNodeStakeAmount := uint64(0)
	for _, info := range viewInfo.Peers {
		wholeNodeStakeAmount += info.UserStakeAmount + info.InitPos
	}
	for peer, info := range viewInfo.Peers {
		peerFee := (info.UserStakeAmount + info.InitPos) * wholeFee / wholeNodeStakeAmount
		feeInfo = append(feeInfo, &PeerAmount{PeerPubKey: peer, Amount: peerFee})
	}
	return feeInfo
}

func peerInitStake(native *native.NativeService, id common.ShardID, peerPubKey string, peerOwner common.Address,
	amount uint64) error {
	currentView, err := GetShardCurrentViewIndex(native, id)
	if err != nil {
		return fmt.Errorf("peerInitStake: get current view peer stake info failed, err: %s", err)
	}
	// if peer join after view 0, the stake should effective from next round
	if currentView > 0 {
		currentView++
	}
	nextView := currentView + 1
	initViewInfo, err := GetShardViewInfo(native, id, currentView)
	if err != nil {
		return fmt.Errorf("peerInitStake: get init view info failed, err: %s", err)
	}
	nextViewInfo, err := GetShardViewInfo(native, id, nextView)
	if err != nil {
		return fmt.Errorf("peerInitStake: get next view info failed, err: %s", err)
	}
	if initViewInfo.Peers == nil {
		initViewInfo.Peers = make(map[string]*PeerViewInfo)
		nextViewInfo.Peers = make(map[string]*PeerViewInfo)
	}
	if len(nextViewInfo.Peers) != len(initViewInfo.Peers) {
		nextViewInfo.Peers = initViewInfo.Peers
	}
	peerViewInfo, ok := initViewInfo.Peers[peerPubKey]
	if ok {
		return fmt.Errorf("peerInitStake: peer %s has already exist", peerPubKey)
	}
	peerViewInfo = &PeerViewInfo{
		PeerPubKey: peerPubKey,
		Owner:      peerOwner,
		InitPos:    amount,
		CanStake:   true, // default can stake asset
	}
	initViewInfo.Peers[peerPubKey] = peerViewInfo
	setShardViewInfo(native, id, currentView, initViewInfo)
	nextViewInfo.Peers[peerPubKey] = peerViewInfo
	setShardViewInfo(native, id, nextView, nextViewInfo)

	lastStakeView, err := getUserLastStakeView(native, id, peerOwner)
	if err != nil {
		return fmt.Errorf("reduceInitPos: failed, err: %s", err)
	}
	if lastStakeView > nextView {
		return fmt.Errorf("reduceInitPos: user last stake view %d and next view %d unmatch", lastStakeView, nextView)
	} else if lastStakeView == nextView {
		lastStakeView = currentView
	}
	lastUserStakeInfo, err := getShardViewUserStake(native, id, lastStakeView, peerOwner)
	if err != nil {
		return fmt.Errorf("reduceInitPos: get user last stake info failed, err: %s", err)
	}
	if isUserStakePeerEmpty(lastUserStakeInfo) {
		lastUserStakeInfo.Peers = make(map[string]*UserPeerStakeInfo)
	}
	if _, ok := lastUserStakeInfo.Peers[peerPubKey]; !ok {
		lastUserStakeInfo.Peers[peerPubKey] = &UserPeerStakeInfo{PeerPubKey: peerPubKey}
	}
	nextUserStakeInfo, err := getShardViewUserStake(native, id, nextView, peerOwner)
	if err != nil {
		return fmt.Errorf("reduceInitPos: get user next stake info failed, err: %s", err)
	}
	if isUserStakePeerEmpty(nextUserStakeInfo) {
		nextUserStakeInfo.Peers = lastUserStakeInfo.Peers
	}
	if _, ok := nextUserStakeInfo.Peers[peerPubKey]; !ok {
		nextUserStakeInfo.Peers[peerPubKey] = lastUserStakeInfo.Peers[peerPubKey]
	}
	// update user last stake view num
	setShardViewUserStake(native, id, currentView, peerOwner, lastUserStakeInfo)
	setShardViewUserStake(native, id, nextView, peerOwner, nextUserStakeInfo)
	setUserLastStakeView(native, id, peerOwner, nextView)
	return nil
}

func addInitPos(native *native.NativeService, id common.ShardID, owner common.Address, info *PeerAmount) error {
	currentView, err := GetShardCurrentViewIndex(native, id)
	if err != nil {
		return fmt.Errorf("addInitPos: failed, err: %s", err)
	}
	nextView := currentView + 1
	nextViewInfo, err := GetShardViewInfo(native, id, nextView)
	if err != nil {
		return fmt.Errorf("addInitPos: get view info failed, err: %s", err)
	}
	pubKeyString := strings.ToLower(info.PeerPubKey)
	peerInfo, ok := nextViewInfo.Peers[pubKeyString]
	if !ok {
		return fmt.Errorf("addInitPos: node %s not exist", pubKeyString)
	}
	if peerInfo.Owner != owner {
		return fmt.Errorf("addInitPos: user %s isn't peer owner %s", owner.ToBase58(), peerInfo.Owner.ToBase58())
	}
	peerInfo.InitPos += info.Amount
	nextViewInfo.Peers[pubKeyString] = peerInfo
	setShardViewInfo(native, id, nextView, nextViewInfo)
	return nil
}

func reduceInitPos(native *native.NativeService, id common.ShardID, owner common.Address, info *PeerAmount) error {
	currentView, err := GetShardCurrentViewIndex(native, id)
	if err != nil {
		return fmt.Errorf("reduceInitPos: failed, err: %s", err)
	}
	nextView := currentView + 1
	nextViewInfo, err := GetShardViewInfo(native, id, nextView)
	if err != nil {
		return fmt.Errorf("reduceInitPos: get next view info failed, err: %s", err)
	}
	pubKeyString := strings.ToLower(info.PeerPubKey)
	peerInfo, ok := nextViewInfo.Peers[pubKeyString]
	if !ok {
		return fmt.Errorf("reduceInitPos: node %s not exist in next view", pubKeyString)
	}
	if peerInfo.Owner != owner {
		return fmt.Errorf("reduceInitPos: user %s isn't peer owner %s", owner.ToBase58(), peerInfo.Owner.ToBase58())
	}
	if peerInfo.InitPos < info.Amount {
		return fmt.Errorf("reduceInitPos: init pos not enough")
	}
	minNodeStake, err := GetNodeMinStakeAmount(native, id)
	if err != nil {
		return fmt.Errorf("reduceInitPos: failed, err: %s", err)
	}
	if peerInfo.InitPos-info.Amount < minNodeStake {
		return fmt.Errorf("reduceInitPos: should more than min stake amout")
	}
	peerInfo.InitPos -= info.Amount
	peerInfo.UserUnfreezeAmount += info.Amount
	nextViewInfo.Peers[pubKeyString] = peerInfo
	setShardViewInfo(native, id, nextView, nextViewInfo)
	lastStakeView, err := getUserLastStakeView(native, id, owner)
	if err != nil {
		return fmt.Errorf("reduceInitPos: failed, err: %s", err)
	}
	if lastStakeView > nextView {
		return fmt.Errorf("reduceInitPos: user last stake view %d and next view %d unmatch", lastStakeView, nextView)
	} else if lastStakeView == nextView {
		lastStakeView = currentView
	}
	lastUserStakeInfo, err := getShardViewUserStake(native, id, lastStakeView, owner)
	if err != nil {
		return fmt.Errorf("reduceInitPos: get user last stake info failed, err: %s", err)
	}
	if isUserStakePeerEmpty(lastUserStakeInfo) {
		lastUserStakeInfo.Peers = make(map[string]*UserPeerStakeInfo)
	}
	if _, ok := lastUserStakeInfo.Peers[pubKeyString]; !ok {
		lastUserStakeInfo.Peers[pubKeyString] = &UserPeerStakeInfo{PeerPubKey: pubKeyString}
	}
	nextUserStakeInfo, err := getShardViewUserStake(native, id, nextView, owner)
	if err != nil {
		return fmt.Errorf("reduceInitPos: get user next stake info failed, err: %s", err)
	}
	if isUserStakePeerEmpty(nextUserStakeInfo) {
		copyUserStakeInfo(lastUserStakeInfo, nextUserStakeInfo)
	}
	if _, ok := nextUserStakeInfo.Peers[pubKeyString]; !ok {
		nextUserStakeInfo.Peers[pubKeyString] = lastUserStakeInfo.Peers[pubKeyString]
	}
	nextUserStakeInfo.Peers[pubKeyString].UnfreezeAmount += info.Amount
	setShardViewUserStake(native, id, currentView, owner, lastUserStakeInfo)
	setShardViewUserStake(native, id, nextView, owner, nextUserStakeInfo)
	setUserLastStakeView(native, id, owner, nextView)
	return nil
}

func userStake(native *native.NativeService, id common.ShardID, user common.Address, stakeInfo []*PeerAmount) error {
	// get view index
	lastStakeView, err := getUserLastStakeView(native, id, user)
	if err != nil {
		return fmt.Errorf("userStake: failed, err: %s", err)
	}
	currentView, err := GetShardCurrentViewIndex(native, id)
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
	if isUserStakePeerEmpty(lastUserStakeInfo) { // user stake peer firstly
		lastUserStakeInfo.Peers = make(map[string]*UserPeerStakeInfo)
	}
	nextUserStakeInfo, err := getShardViewUserStake(native, id, nextView, user)
	if err != nil {
		return fmt.Errorf("userStake: get user next stake info failed, err: %s", err)
	}
	if isUserStakePeerEmpty(nextUserStakeInfo) {
		copyUserStakeInfo(lastUserStakeInfo, nextUserStakeInfo)
	}
	currentViewInfo, err := GetShardViewInfo(native, id, currentView)
	if err != nil {
		return fmt.Errorf("userStake: get current view info failed, err: %s", err)
	}
	nextViewInfo, err := GetShardViewInfo(native, id, nextView)
	if err != nil {
		return fmt.Errorf("userStake: get next view info failed, err: %s", err)
	}
	for _, info := range stakeInfo {
		pubKeyString := strings.ToLower(info.PeerPubKey)
		amount := info.Amount
		currentPeerStakeInfo, ok := currentViewInfo.Peers[pubKeyString]
		if !ok {
			return fmt.Errorf("userStake: current view cannot find peer %s", pubKeyString)
		}
		if currentPeerStakeInfo.Owner == user {
			return fmt.Errorf("userStake: cannot stake self node %s at current", pubKeyString)
		}
		nextPeerStakeInfo, ok := nextViewInfo.Peers[pubKeyString]
		if !ok {
			return fmt.Errorf("userStake: next view cannot find peer %s", pubKeyString)
		}
		if !nextPeerStakeInfo.CanStake {
			return fmt.Errorf("userStake: peer %s cannot stake", pubKeyString)
		}
		if nextPeerStakeInfo.Owner == user {
			return fmt.Errorf("userStake: cannot stake self node %s at next", pubKeyString)
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
		nextPeerStakeInfo.UserStakeAmount += amount
		nextViewInfo.Peers[pubKeyString] = nextPeerStakeInfo
	}
	setUserLastStakeView(native, id, user, nextView)
	setShardViewUserStake(native, id, currentView, user, lastUserStakeInfo)
	setShardViewUserStake(native, id, nextView, user, nextUserStakeInfo)
	setShardViewInfo(native, id, currentView, currentViewInfo)
	setShardViewInfo(native, id, nextView, nextViewInfo)
	return nil
}

func unfreezeStakeAsset(native *native.NativeService, id common.ShardID, user common.Address, unFreezeInfo []*PeerAmount) error {
	// get view index
	lastStakeView, err := getUserLastStakeView(native, id, user)
	if err != nil {
		return fmt.Errorf("unfreezeStakeAsset: failed, err: %s", err)
	}
	currentView, err := GetShardCurrentViewIndex(native, id)
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
		return fmt.Errorf("unfreezeStakeAsset: user stake peer info is empty")
	}
	nextViewInfo, err := GetShardViewInfo(native, id, nextView)
	if err != nil {
		return fmt.Errorf("unfreezeStakeAsset: get next view info failed, err: %s", err)
	}
	for _, info := range unFreezeInfo {
		pubKeyString := strings.ToLower(info.PeerPubKey)
		amount := info.Amount
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
		if !ok {
			return fmt.Errorf("unfreezeStakeAsset: next view cannot find peer %s", pubKeyString)
		} else if nextPeerStakeInfo.Owner == user {
			return fmt.Errorf("unfreezeStakeAsset: cannot unfreeze self node %s at next", pubKeyString)
		} else if nextPeerStakeInfo.UserStakeAmount < amount {
			return fmt.Errorf("unfreezeStakeAsset: peer %s stake num not enough", pubKeyString)
		} else if nextPeerStakeInfo.InitPos == 0 {
			// peer has already exit consensus, user should also unfreeze stake asset, cannot withdraw asset straightly
			nextUserPeerStakeInfo.UnfreezeAmount += nextUserPeerStakeInfo.StakeAmount
			nextUserPeerStakeInfo.StakeAmount = 0
		} else {
			nextUserPeerStakeInfo.StakeAmount -= amount
			nextUserPeerStakeInfo.UnfreezeAmount += amount
			nextPeerStakeInfo.UserStakeAmount -= amount
			nextPeerStakeInfo.UserUnfreezeAmount += amount
			nextViewInfo.Peers[pubKeyString] = nextPeerStakeInfo
		}
		nextUserStakeInfo.Peers[pubKeyString] = nextUserPeerStakeInfo
	}
	setUserLastStakeView(native, id, user, nextView)
	// update user stake info from last to current
	setShardViewUserStake(native, id, currentView, user, lastUserStakeInfo)
	setShardViewUserStake(native, id, nextView, user, nextUserStakeInfo)
	setShardViewInfo(native, id, nextView, nextViewInfo)
	return nil
}

// return withdraw amount
func withdrawStakeAsset(native *native.NativeService, id common.ShardID, user common.Address) (uint64, error) {
	// get view index
	lastStakeView, err := getUserLastStakeView(native, id, user)
	if err != nil {
		return 0, fmt.Errorf("withdrawStakeAsset: failed, err: %s", err)
	}
	currentView, err := GetShardCurrentViewIndex(native, id)
	if err != nil {
		return 0, fmt.Errorf("withdrawStakeAsset: failed, err: %s", err)
	}
	nextView := currentView + 1
	if lastStakeView > nextView {
		return 0, fmt.Errorf("withdrawStakeAsset: user last stake view %d and next view %d unmatch",
			lastStakeView, nextView)
	} else if lastStakeView == nextView {
		lastStakeView = currentView
	}
	lastUserStakeInfo, err := getShardViewUserStake(native, id, lastStakeView, user)
	if err != nil {
		return 0, fmt.Errorf("withdrawStakeAsset: get user last stake info failed, err: %s", err)
	}
	if isUserStakePeerEmpty(lastUserStakeInfo) {
		return 0, fmt.Errorf("withdrawStakeAsset: user stake peer info is empty")
	}
	nextUserStakeInfo, err := getShardViewUserStake(native, id, nextView, user)
	if err != nil {
		return 0, fmt.Errorf("withdrawStakeAsset: get user next stake info failed, err: %s", err)
	}
	if isUserStakePeerEmpty(nextUserStakeInfo) {
		copyUserStakeInfo(lastUserStakeInfo, nextUserStakeInfo)
	}
	currentViewInfo, err := GetShardViewInfo(native, id, currentView)
	if err != nil {
		return 0, fmt.Errorf("withdrawStakeAsset: get current view info failed, err: %s", err)
	}
	nextViewInfo, err := GetShardViewInfo(native, id, nextView)
	if err != nil {
		return 0, fmt.Errorf("withdrawStakeAsset: get next view info failed, err: %s", err)
	}
	amount := uint64(0)
	for peer, userPeerStakeInfo := range lastUserStakeInfo.Peers {
		if nextPeerInfo, ok := nextViewInfo.Peers[peer]; ok {
			nextPeerInfo.UserStakeAmount -= userPeerStakeInfo.CurrentViewStakeAmount
			nextPeerInfo.UserUnfreezeAmount -= userPeerStakeInfo.UnfreezeAmount
			nextViewInfo.Peers[peer] = nextPeerInfo
			nextUserPeerStakeInfo, ok := nextUserStakeInfo.Peers[peer]
			if !ok {
				nextUserPeerStakeInfo = &UserPeerStakeInfo{PeerPubKey: userPeerStakeInfo.PeerPubKey,
					StakeAmount: userPeerStakeInfo.StakeAmount}
			} else {
				nextUserPeerStakeInfo.UnfreezeAmount -= userPeerStakeInfo.UnfreezeAmount
				nextUserPeerStakeInfo.StakeAmount -= userPeerStakeInfo.CurrentViewStakeAmount
			}
			nextUserStakeInfo.Peers[peer] = nextUserPeerStakeInfo
		}
		amount += userPeerStakeInfo.UnfreezeAmount + userPeerStakeInfo.CurrentViewStakeAmount
		if currentPeerInfo, ok := currentViewInfo.Peers[peer]; ok {
			currentPeerInfo.UserUnfreezeAmount -= userPeerStakeInfo.UnfreezeAmount
			currentPeerInfo.CurrentViewStakeAmount -= userPeerStakeInfo.CurrentViewStakeAmount
			currentViewInfo.Peers[peer] = currentPeerInfo
		}
		userPeerStakeInfo.CurrentViewStakeAmount = 0
		userPeerStakeInfo.UnfreezeAmount = 0
		lastUserStakeInfo.Peers[peer] = userPeerStakeInfo
	}
	setUserLastStakeView(native, id, user, nextView)
	// update user stake info from last to current
	setShardViewUserStake(native, id, currentView, user, lastUserStakeInfo)
	setShardViewUserStake(native, id, nextView, user, nextUserStakeInfo)
	setShardViewInfo(native, id, nextView, nextViewInfo)
	return amount, nil
}

// return the amount that user could withdraw
func withdrawFee(native *native.NativeService, shardId common.ShardID, user common.Address) (uint64, error) {
	userWithdrawView, err := getUserLastWithdrawView(native, shardId, user)
	if err != nil {
		return 0, fmt.Errorf("withdrawFee: failed, err: %s", err)
	}
	currentView, err := GetShardCurrentViewIndex(native, shardId)
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
		if isUserStakePeerEmpty(userStake) {
			if isUserStakePeerEmpty(latestUserStakeInfo) {
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
				return 0, fmt.Errorf("withdrawFee: failed, view %d, peer %s not exist", i, peer)
			}
			if peerStakeInfo.FeeBalance == 0 {
				continue
			}
			dividends := uint64(0)
			wholeFee := new(big.Int).SetUint64(peerStakeInfo.WholeFee)
			userProportion := new(big.Int).SetUint64(peerStakeInfo.Proportion)
			peerProportion := new(big.Int).SetUint64(PEER_MAX_PROPORTION - peerStakeInfo.Proportion)
			proportionBase := new(big.Int).SetUint64(PEER_MAX_PROPORTION)
			if user == peerStakeInfo.Owner { // peer owner
				if peerStakeInfo.UserStakeAmount == 0 {
					dividends = peerStakeInfo.WholeFee
				} else {
					temp := wholeFee.Mul(wholeFee, peerProportion)
					dividends = temp.Div(temp, proportionBase).Uint64()
				}
			} else {
				temp := wholeFee.Mul(wholeFee, userProportion)
				temp.Mul(temp, new(big.Int).SetUint64(info.StakeAmount))
				temp.Div(temp, new(big.Int).SetUint64(peerStakeInfo.UserStakeAmount))
				// wholeFee * proportion * stakeAmount / allStakeAmount / PEER_MAX_PROPORTION
				dividends = temp.Div(temp, proportionBase).Uint64()
			}
			peerStakeInfo.FeeBalance = peerStakeInfo.FeeBalance - dividends
			viewStake.Peers[peer] = peerStakeInfo
			dividends += dividends
		}
		setShardViewInfo(native, shardId, i, viewStake)
		count++
		latestUserStakeInfo = userStake
	}
	setUserLastWithdrawView(native, shardId, user, i)
	setShardViewUserStake(native, shardId, i, user, latestUserStakeInfo)
	return dividends, nil
}

// change peer max authorization and proportion
func changePeerInfo(native *native.NativeService, shardId common.ShardID, peerOwner common.Address, info *PeerAmount,
	methodName string) error {
	currentView, err := GetShardCurrentViewIndex(native, shardId)
	if err != nil {
		return fmt.Errorf("changePeerInfo: failed, err: %s", err)
	}
	nextView := currentView + 1
	nextViewInfo, err := GetShardViewInfo(native, shardId, nextView)
	if err != nil {
		return fmt.Errorf("changePeerInfo: failed, err: %s", err)
	}
	peerInfo, ok := nextViewInfo.Peers[strings.ToLower(info.PeerPubKey)]
	if !ok {
		return fmt.Errorf("changePeerInfo: peer not exist in next view")
	}
	if peerInfo.Owner != peerOwner {
		return fmt.Errorf("changePeerInfo: peer owner not match")
	}
	if !peerInfo.CanStake {
		return fmt.Errorf("changePeerInfo: peer is exiting consensus")
	}
	if peerInfo.InitPos == 0 {
		return fmt.Errorf("changePeerInfo: peer exited consensus")
	}
	switch methodName {
	case CHANGE_MAX_AUTHORIZATION:
		peerInfo.MaxAuthorization = info.Amount
	case CHANGE_PROPORTION:
		peerInfo.Proportion = info.Amount
	default:
		return fmt.Errorf("changePeerInfo: unsupport change field")
	}
	nextViewInfo.Peers[strings.ToLower(info.PeerPubKey)] = peerInfo
	setShardViewInfo(native, shardId, nextView, nextViewInfo)
	return nil
}

func isUserStakePeerEmpty(info *UserStakeInfo) bool {
	if info.Peers == nil || len(info.Peers) == 0 {
		return true
	}
	return false
}

func copyUserStakeInfo(lastUserStakeInfo, nextUserStakeInfo *UserStakeInfo) {
	nextUserStakeInfo.Peers = make(map[string]*UserPeerStakeInfo)
	for peer, info := range lastUserStakeInfo.Peers {
		nextUserStakeInfo.Peers[peer] = &UserPeerStakeInfo{
			PeerPubKey:             info.PeerPubKey,
			StakeAmount:            info.StakeAmount,
			CurrentViewStakeAmount: info.CurrentViewStakeAmount,
			UnfreezeAmount:         info.UnfreezeAmount,
		}
	}
}
