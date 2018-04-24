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

package native

import (
	"encoding/json"

	"encoding/hex"
	"github.com/ontio/ontology/core/genesis"
	cstates "github.com/ontio/ontology/core/states"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native/states"
	"math/big"
)

const (
	//function name
	REGISTER_CANDIDATE = "registerCandidate"
	APPROVE_CANDIDATE  = "approveCandidate"
	QUIT_CANDIDATE     = "quitCandidate"
	REGISTER_SYNC_NODE = "registerSyncNode"
	QUIT_SYNC_NODE     = "quitSyncNode"
	VOTE_FORP_EER        = "voteForPeer"

	//key prefix
	GOVERNANCE_VIEW = "governanceView"
	CANDIDITE_INDEX = "candidateIndex"
	CANDIDITE_POOL  = "candidatePool"
	REGISTER_POOL   = "registerPool"
	SYNC_NODE_POOL  = "syncNodePool"
	VOTE_INFO_POOL  = "voteInfoPool"
)

func init() {
	Contracts[genesis.GovernanceContractAddress] = RegisterGovernanceContract
}

func RegisterGovernanceContract(native *NativeService) {
	native.Register(REGISTER_CANDIDATE, RegisterCandidate)
	native.Register(APPROVE_CANDIDATE, ApproveCandidate)
	native.Register(QUIT_CANDIDATE, QuitCandidate)
	native.Register(REGISTER_SYNC_NODE, RegisterSyncNode)
	native.Register(QUIT_SYNC_NODE, QuitSyncNode)
	native.Register(VOTE_FORP_EER, VoteForPeer)
}

func RegisterCandidate(native *NativeService) error {
	params := new(states.RegisterCandidateParam)
	err := json.Unmarshal(native.Input, params)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[registerCandidate] Contract params Unmarshal error!")
	}

	//check witness
	err = validateOwner(native, params.Address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[registerCandidate] CheckWitness error!")
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	peerPubkeyPrefix, err := hex.DecodeString(params.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[registerCandidate] PeerPubkey format error!")
	}

	//registerPool storage
	//check registerPool
	v1, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(REGISTER_POOL), peerPubkeyPrefix))
	if v1 != nil {
		return errors.NewErr("[registerCandidate] PeerPubkey is already in registerPool!")
	}
	//check candidatePool
	v2, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(CANDIDITE_POOL), peerPubkeyPrefix))
	if v2 != nil {
		return errors.NewErr("[registerCandidate] PeerPubkey is already in candidatePool!")
	}

	registerPool := &states.RegisterPool{
		Address: params.Address,
		InitPos: params.InitPos,
	}
	value, err := json.Marshal(registerPool)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[registerCandidate] Marshal registerPool error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(REGISTER_POOL), peerPubkeyPrefix), &cstates.StorageItem{Value: value})

	addCommonEvent(native, contract, REGISTER_CANDIDATE, params)

	return nil
}

func ApproveCandidate(native *NativeService) error {
	params := new(states.ApproveCandidateParam)
	err := json.Unmarshal(native.Input, params)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[approveCandidate] Contract params Unmarshal error!")
	}

	//TODO: check witness
	//err = validateOwner(native, params.Address)
	//if err != nil {
	//	return errors.NewDetailErr(err, errors.ErrNoCode, "[registerCandidate] CheckWitness error!")
	//}

	contract := native.ContextRef.CurrentContext().ContractAddress
	peerPubkeyPrefix, err := hex.DecodeString(params.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[approveCandidate] PeerPubkey format error!")
	}

	//get current view
	view, err := getGovernanceView(native, contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[approveCandidate] Get view error!")
	}

	//get registerPool
	registerPoolBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(REGISTER_POOL), peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[approveCandidate] Get registerPoolBytes error!")
	}
	registerPool := new(states.RegisterPool)
	if registerPoolBytes != nil {
		registerPoolStore, _ := registerPoolBytes.(*cstates.StorageItem)
		err := json.Unmarshal(registerPoolStore.Value, registerPool)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[approveCandidate] Unmarshal registerPool error!")
		}
	} else {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[approveCandidate] PeerPubkey is not in registerPool!")
	}

	//get index
	candidateIndexBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(CANDIDITE_INDEX)))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[approveCandidate] Get candidateIndex error!")
	}
	var candidateIndex *big.Int
	if candidateIndexBytes == nil {
		candidateIndex = new(big.Int).SetInt64(1)
	} else {
		candidateIndexStore, _ := candidateIndexBytes.(*cstates.StorageItem)
		candidateIndex = new(big.Int).SetBytes(candidateIndexStore.Value)
	}

	//candidatePool storage
	candidatePool := &states.CandidatePool{
		Index:   candidateIndex,
		Address: registerPool.Address,
		InitPos: registerPool.InitPos,
	}

	value, err := json.Marshal(candidatePool)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[approveCandidate] Marshal candidatePool error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(CANDIDITE_POOL), view.Bytes(), peerPubkeyPrefix), &cstates.StorageItem{Value: value})

	//update candidateIndex
	newCandidateIndex := new(big.Int).Add(candidateIndex, new(big.Int).SetInt64(1))
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(CANDIDITE_INDEX)), &cstates.StorageItem{Value: newCandidateIndex.Bytes()})

	//update registerPool
	native.CloneCache.Delete(scommon.ST_STORAGE, concatKey(contract, []byte(REGISTER_POOL), peerPubkeyPrefix))

	addCommonEvent(native, contract, APPROVE_CANDIDATE, params)

	return nil
}

func QuitCandidate(native *NativeService) error {
	params := new(states.QuitCandidateParam)
	err := json.Unmarshal(native.Input, params)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[quitCandidate] Contract params Unmarshal error!")
	}

	//check witness
	err = validateOwner(native, params.Address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[quitCandidate] CheckWitness error!")
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	peerPubkeyPrefix, err := hex.DecodeString(params.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[quitCandidate] PeerPubkey format error!")
	}

	//get current view
	view, err := getGovernanceView(native, contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[approveCandidate] Get view error!")
	}

	//get candidatePool
	candidatePoolBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(CANDIDITE_POOL), view.Bytes(), peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[quitCandidate] Get candidatePoolBytes error!")
	}
	candidatePool := new(states.CandidatePool)
	if candidatePoolBytes != nil {
		candidatePoolStore, _ := candidatePoolBytes.(*cstates.StorageItem)
		err := json.Unmarshal(candidatePoolStore.Value, candidatePool)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[quitCandidate] Unmarshal candidatePool error!")
		}
	} else {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[quitCandidate] PeerPubkey is not in candidatePool!")
	}
	if params.Address != candidatePool.Address {
		return errors.NewErr("[quitCandidate] PeerPubkey is not registered by this address!")
	}

	//delete candidatePool
	native.CloneCache.Delete(scommon.ST_STORAGE, concatKey(contract, []byte(CANDIDITE_POOL), view.Bytes(), peerPubkeyPrefix))

	addCommonEvent(native, contract, QUIT_CANDIDATE, params)

	return nil
}

func RegisterSyncNode(native *NativeService) error {
	params := new(states.RegisterSyncNodeParam)
	err := json.Unmarshal(native.Input, params)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[registerSyncNode] Contract params Unmarshal error!")
	}

	//check witness
	err = validateOwner(native, params.Address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[registerSyncNode] CheckWitness error!")
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	peerPubkeyPrefix, err := hex.DecodeString(params.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[registerSyncNode] PeerPubkey format error!")
	}

	//get current view
	view, err := getGovernanceView(native, contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[registerSyncNode] Get view error!")
	}

	//syncNodePool storage
	//check syncNodePool
	v, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(SYNC_NODE_POOL), peerPubkeyPrefix))
	if v != nil {
		return errors.NewErr("[registerSyncNode] PeerPubkey is already in syncNodePool!")
	}

	syncNodePool := &states.SyncNodePool{
		Address: params.Address,
		InitPos: params.InitPos,
	}

	value, err := json.Marshal(syncNodePool)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[registerSyncNode] Marshal syncNodePool error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(SYNC_NODE_POOL), view.Bytes(), peerPubkeyPrefix), &cstates.StorageItem{Value: value})

	addCommonEvent(native, contract, REGISTER_SYNC_NODE, params)

	return nil
}

func QuitSyncNode(native *NativeService) error {
	params := new(states.QuitSyncNodeParam)
	err := json.Unmarshal(native.Input, params)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[quitSyncNode] Contract params Unmarshal error!")
	}

	//check witness
	err = validateOwner(native, params.Address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[quitSyncNode] CheckWitness error!")
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	peerPubkeyPrefix, err := hex.DecodeString(params.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[quitSyncNode] PeerPubkey format error!")
	}

	//get current view
	view, err := getGovernanceView(native, contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[quitSyncNode] Get view error!")
	}

	//get syncNodePool
	syncNodePoolBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(SYNC_NODE_POOL), view.Bytes(), peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[quitSyncNode] Get syncNodePoolBytes error!")
	}
	syncNodePool := new(states.SyncNodePool)
	if syncNodePoolBytes != nil {
		candidatePoolStore, _ := syncNodePoolBytes.(*cstates.StorageItem)
		err := json.Unmarshal(candidatePoolStore.Value, syncNodePool)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[quitSyncNode] Unmarshal syncNodePool error!")
		}
	} else {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[quitSyncNode] PeerPubkey is not in syncNodePool!")
	}
	if params.Address != syncNodePool.Address {
		return errors.NewErr("[quitSyncNode] PeerPubkey is not registered by this address!")
	}

	//delete syncNodePool
	native.CloneCache.Delete(scommon.ST_STORAGE, concatKey(contract, []byte(SYNC_NODE_POOL), view.Bytes(), peerPubkeyPrefix))

	addCommonEvent(native, contract, QUIT_SYNC_NODE, params)

	return nil
}

func VoteForPeer(native *NativeService) error {
	params := new(states.VoteForPeerParam)
	err := json.Unmarshal(native.Input, params)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[voteForPeer] Contract params Unmarshal error!")
	}

	//check witness
	err = validateOwner(native, params.Address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[voteForPeer] CheckWitness error!")
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	addressPrefix, err := hex.DecodeString(params.Address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[voteForPeer] Address format error!")
	}

	//get current view
	view, err := getGovernanceView(native, contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[voteForPeer] Get view error!")
	}

	//get voteInfoPool
	voteInfoPoolBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(VOTE_INFO_POOL), view.Bytes(), addressPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[voteForPeer] Get voteInfoPoolBytes error!")
	}
	voteInfoPool := new(states.VoteInfoPool)
	newTotal := new(big.Int)
	for _, item := range params.VoteTable {
		newTotal = new(big.Int).Add(newTotal, item)
	}
	if voteInfoPoolBytes != nil {
		voteInfoPoolStore, _ := voteInfoPoolBytes.(*cstates.StorageItem)
		err := json.Unmarshal(voteInfoPoolStore.Value, voteInfoPool)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[voteForPeer] Unmarshal voteInfoPool error!")
		}
		if newTotal.Cmp(voteInfoPool.Total) > 0 {
			//TODO: transfer ont
		}
		if newTotal.Cmp(voteInfoPool.Total) < 0 {
			//TODO: transfer ont
		}
	}
	//TODO: transfer ont(newTotal)

	newVoteInfoPool := &states.VoteInfoPool{
		Address:params.Address,
		Total:newTotal,
		VoteTable:params.VoteTable,
	}
	value, err := json.Marshal(newVoteInfoPool)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[registerSyncNode] Marshal newVoteInfoPool error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(VOTE_INFO_POOL), view.Bytes(), addressPrefix), &cstates.StorageItem{Value: value})

	addCommonEvent(native, contract, VOTE_FORP_EER, params)

	return nil
}

func CommitDpos(native *NativeService) error {
	return nil
}
