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
	"fmt"
	"io"
	"sort"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type View uint64 // shard consensus epoch index

type PeerViewInfo struct {
	PeerPubKey             string
	Owner                  common.Address
	CanStake               bool   // if user can stake peer
	WholeFee               uint64 // each epoch handling fee
	FeeBalance             uint64 // each epoch handling fee not be withdrawn
	WholeStakeAmount       uint64 // node + user stake amount at the last view
	WholeUnfreezeAmount    uint64 // all user can withdraw amount
	CurrentViewStakeAmount uint64 // current view user stake amount
	UserStakeAmount        uint64 // user stake amount
	MaxAuthorization       uint64 // max user stake amount
	Proportion             uint64 // proportion to user
}

func (this *PeerViewInfo) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubKey); err != nil {
		return fmt.Errorf("serialize peer public key failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.Owner); err != nil {
		return fmt.Errorf("serialize owner failed, err: %s", err)
	}
	if err := serialization.WriteBool(w, this.CanStake); err != nil {
		return fmt.Errorf("serialize can stake failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.WholeFee); err != nil {
		return fmt.Errorf("serialize whole fee failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.FeeBalance); err != nil {
		return fmt.Errorf("serialize fee balance failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.WholeStakeAmount); err != nil {
		return fmt.Errorf("serialize whole stake amount failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.WholeUnfreezeAmount); err != nil {
		return fmt.Errorf("serialize whole unfreeze amount failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.CurrentViewStakeAmount); err != nil {
		return fmt.Errorf("serialize current view stake amount failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.UserStakeAmount); err != nil {
		return fmt.Errorf("serialize user stake amount failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.MaxAuthorization); err != nil {
		return fmt.Errorf("serialize max authorization failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.Proportion); err != nil {
		return fmt.Errorf("serialize propotion failed, err: %s", err)
	}
	return nil
}

func (this *PeerViewInfo) Deserialize(r io.Reader) error {
	var err error = nil
	if this.PeerPubKey, err = serialization.ReadString(r); err != nil {
		return fmt.Errorf("deserialize: read peer pub key failed, err: %s", err)
	}
	if this.Owner, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read owner failed, err: %s", err)
	}
	if this.CanStake, err = serialization.ReadBool(r); err != nil {
		return fmt.Errorf("deserialize: read can stake failed, err: %s", err)
	}
	if this.WholeFee, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read whole fee failed, err: %s", err)
	}
	if this.FeeBalance, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read fee balance failed, err: %s", err)
	}
	if this.WholeStakeAmount, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read whole stake amount failed, err: %s", err)
	}
	if this.WholeUnfreezeAmount, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read whole unfreeze amount failed, err: %s", err)
	}
	if this.CurrentViewStakeAmount, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read current view stake amount failed, err: %s", err)
	}
	if this.UserStakeAmount, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read user stake amount failed, err: %s", err)
	}
	if this.MaxAuthorization, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read max authorization failed, err: %s", err)
	}
	if this.Proportion, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read proportion failed, err: %s", err)
	}
	return nil
}

type ViewInfo struct {
	Peers map[string]*PeerViewInfo
}

func (this *ViewInfo) Serialize(w io.Writer) error {
	err := utils.WriteVarUint(w, uint64(len(this.Peers)))
	if err != nil {
		return fmt.Errorf("serialize: wirte peers len faield, err: %s", err)
	}
	peerInfoList := make([]*PeerViewInfo, 0)
	for _, info := range this.Peers {
		peerInfoList = append(peerInfoList, info)
	}
	sort.SliceStable(peerInfoList, func(i, j int) bool {
		return peerInfoList[i].PeerPubKey > peerInfoList[j].PeerPubKey
	})
	for index, info := range peerInfoList {
		err = info.Serialize(w)
		if err != nil {
			return fmt.Errorf("serialize: index %d, err: %s", index, err)
		}
	}
	return nil
}

func (this *ViewInfo) Deserialize(r io.Reader) error {
	num, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialze: read peers num failed, err: %s", err)
	}
	this.Peers = make(map[string]*PeerViewInfo)
	for i := uint64(0); i < num; i++ {
		info := &PeerViewInfo{}
		err = info.Deserialize(r)
		if err != nil {
			return fmt.Errorf("deserialize: index %d, err: %s", i, err)
		}
		this.Peers[info.PeerPubKey] = info
	}
	return nil
}

type UserPeerStakeInfo struct {
	PeerPubKey             string
	StakeAmount            uint64
	CurrentViewStakeAmount uint64
	UnfreezeAmount         uint64
}

type UserStakeInfo struct {
	Peers map[string]*UserPeerStakeInfo
}

func (this *UserStakeInfo) Serialize(w io.Writer) error {
	err := utils.WriteVarUint(w, uint64(len(this.Peers)))
	if err != nil {
		return fmt.Errorf("serialize: wirte peers len faield, err: %s", err)
	}
	userPeerInfoList := make([]*UserPeerStakeInfo, 0)
	for _, info := range this.Peers {
		userPeerInfoList = append(userPeerInfoList, info)
	}
	sort.SliceStable(userPeerInfoList, func(i, j int) bool {
		return userPeerInfoList[i].PeerPubKey > userPeerInfoList[j].PeerPubKey
	})
	for index, info := range userPeerInfoList {
		err = serialization.WriteString(w, info.PeerPubKey)
		if err != nil {
			return fmt.Errorf("serialize peer public key failed, index %d, err: %s", index, err)
		}
		err = utils.WriteVarUint(w, info.StakeAmount)
		if err != nil {
			return fmt.Errorf("serialize stake amount failed, index %d, err: %s", index, err)
		}
		err = utils.WriteVarUint(w, info.CurrentViewStakeAmount)
		if err != nil {
			return fmt.Errorf("serialize current view stake amount failed, index %d, err: %s", index, err)
		}
		err = utils.WriteVarUint(w, info.UnfreezeAmount)
		if err != nil {
			return fmt.Errorf("serialize unfreeze amount failed, index %d, err: %s", index, err)
		}
	}
	return nil
}

func (this *UserStakeInfo) Deserialize(r io.Reader) error {
	num, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialze: read peers num failed, err: %s", err)
	}
	this.Peers = make(map[string]*UserPeerStakeInfo)
	for i := uint64(0); i < num; i++ {
		info := &UserPeerStakeInfo{}
		peerPubKey, err := serialization.ReadString(r)
		if err != nil {
			return fmt.Errorf("deserialze: read peer pub key failed, index %d, err: %s", i, err)
		}
		info.PeerPubKey = peerPubKey
		stakeAmount, err := utils.ReadVarUint(r)
		if err != nil {
			return fmt.Errorf("deserialze: deserialize stake amount failed, err: %s", err)
		}
		info.StakeAmount = stakeAmount
		currentViewStakeAmount, err := utils.ReadVarUint(r)
		if err != nil {
			return fmt.Errorf("deserialze: deserialize current view stake amount failed, err: %s", err)
		}
		info.CurrentViewStakeAmount = currentViewStakeAmount
		unfreezeAmount, err := utils.ReadVarUint(r)
		if err != nil {
			return fmt.Errorf("deserialze: deserialize unfreeze amount failed, err: %s", err)
		}
		info.UnfreezeAmount = unfreezeAmount
		this.Peers[peerPubKey] = info
	}
	return nil
}

type UserUnboundOngInfo struct {
	Time        uint32
	StakeAmount uint64
	Balance     uint64
}

func (this *UserUnboundOngInfo) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, uint64(this.Time)); err != nil {
		return fmt.Errorf("serialize: write time failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.StakeAmount); err != nil {
		return fmt.Errorf("serialize: write amount failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.Balance); err != nil {
		return fmt.Errorf("serialize: write ong balance failed, err: %s", err)
	}
	return nil
}

func (this *UserUnboundOngInfo) Deserialize(r io.Reader) error {
	time, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("serialize: read time failed, err: %s", err)
	}
	this.Time = uint32(time)
	amount, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("serialize: read amount failed, err: %s", err)
	}
	this.StakeAmount = amount
	balance, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("serialize: read ong balance failed, err: %s", err)
	}
	this.Balance = balance
	return nil
}
