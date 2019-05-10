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
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/stretchr/testify/assert"
)

func genPeerViewInfo() *PeerViewInfo {
	acc := account.NewAccount("")
	pubKey := hex.EncodeToString(keypair.SerializePublicKey(acc.PubKey()))
	return &PeerViewInfo{
		PeerPubKey:             pubKey,
		Owner:                  acc.Address,
		CanStake:               true,
		WholeFee:               7694272,
		FeeBalance:             2862,
		InitPos:                8746827,
		UserUnfreezeAmount:     35142,
		CurrentViewStakeAmount: 2111,
		UserStakeAmount:        6256,
		MaxAuthorization:       18818181,
		Proportion:             PEER_MAX_PROPORTION,
	}
}

func TestPeerViewInfo(t *testing.T) {
	state := genPeerViewInfo()
	sink := common.NewZeroCopySink(0)
	state.Serialization(sink)
	source := common.NewZeroCopySource(sink.Bytes())
	newState := &PeerViewInfo{}
	err := newState.Deserialization(source)
	assert.Nil(t, err)
	assert.Equal(t, state, newState)
}

func TestViewInfo(t *testing.T) {
	peerInfo1 := genPeerViewInfo()
	peerInfo2 := genPeerViewInfo()
	state := &ViewInfo{Peers: map[string]*PeerViewInfo{
		peerInfo1.PeerPubKey: peerInfo1,
		peerInfo2.PeerPubKey: peerInfo2,
	}}
	sink := common.NewZeroCopySink(0)
	state.Serialization(sink)
	source := common.NewZeroCopySource(sink.Bytes())
	newState := &ViewInfo{}
	err := newState.Deserialization(source)
	assert.Nil(t, err)
	assert.Equal(t, state, newState)
}

func genUserPeerStakeInfo() *UserPeerStakeInfo {
	acc := account.NewAccount("")
	pubKey := hex.EncodeToString(keypair.SerializePublicKey(acc.PubKey()))
	return &UserPeerStakeInfo{
		PeerPubKey:             pubKey,
		StakeAmount:            45532,
		CurrentViewStakeAmount: 2525,
		UnfreezeAmount:         1241,
	}
}

func TestUserPeerStakeInfo(t *testing.T) {
	state := genUserPeerStakeInfo()
	sink := common.NewZeroCopySink(0)
	state.Serialization(sink)
	source := common.NewZeroCopySource(sink.Bytes())
	newState := &UserPeerStakeInfo{}
	err := newState.Deserialization(source)
	assert.Nil(t, err)
	assert.Equal(t, state, newState)
}

func TestUserStakeInfo(t *testing.T) {
	peerInfo1 := genUserPeerStakeInfo()
	peerInfo2 := genUserPeerStakeInfo()
	state := &UserStakeInfo{Peers: map[string]*UserPeerStakeInfo{
		peerInfo1.PeerPubKey: peerInfo1,
		peerInfo2.PeerPubKey: peerInfo2,
	}}
	sink := common.NewZeroCopySink(0)
	state.Serialization(sink)
	source := common.NewZeroCopySource(sink.Bytes())
	newState := &UserStakeInfo{}
	err := newState.Deserialization(source)
	assert.Nil(t, err)
	assert.Equal(t, state, newState)
}

func TestUserUnboundOngInfo(t *testing.T) {
	state := &UserUnboundOngInfo{
		Time:        1433411,
		StakeAmount: 23445526,
		Balance:     1213131212,
	}
	sink := common.NewZeroCopySink(0)
	state.Serialization(sink)
	source := common.NewZeroCopySource(sink.Bytes())
	newState := &UserUnboundOngInfo{}
	err := newState.Deserialization(source)
	assert.Nil(t, err)
	assert.Equal(t, state, newState)
}
