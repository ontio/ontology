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
	"encoding/hex"
	"encoding/json"
	"math/big"

	"github.com/ontio/ontology/core/genesis"
	cstates "github.com/ontio/ontology/core/states"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native/states"
	"sort"
	"fmt"
	"math"
)

const (
	//function name
	REGISTER_CANDIDATE = "registerCandidate"
	APPROVE_CANDIDATE  = "approveCandidate"
	QUIT_CANDIDATE     = "quitCandidate"
	REGISTER_SYNC_NODE = "registerSyncNode"
	QUIT_SYNC_NODE     = "quitSyncNode"
	VOTE_FORP_EER      = "voteForPeer"
	COMMIT_DPOS        = "commitDpos"

	//key prefix
	INIT_CONFIG     = "initConfig"
	GOVERNANCE_VIEW = "governanceView"
	CANDIDITE_INDEX = "candidateIndex"
	CANDIDITE_POOL  = "candidatePool"
	REGISTER_POOL   = "registerPool"
	SYNC_NODE_POOL  = "syncNodePool"
	VOTE_INFO_POOL  = "voteInfoPool"

	//global
	configK = 7
	configL = 112
)

func init() {
	Contracts[genesis.GovernanceContractAddress] = RegisterGovernanceContract
}

func RegisterGovernanceContract(native *NativeService) {
	native.Register(INIT_CONFIG, InitConfig)
	native.Register(REGISTER_CANDIDATE, RegisterCandidate)
	native.Register(APPROVE_CANDIDATE, ApproveCandidate)
	native.Register(QUIT_CANDIDATE, QuitCandidate)
	native.Register(REGISTER_SYNC_NODE, RegisterSyncNode)
	native.Register(QUIT_SYNC_NODE, QuitSyncNode)
	native.Register(VOTE_FORP_EER, VoteForPeer)
	native.Register(COMMIT_DPOS, CommitDpos)
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
		Index:      candidateIndex,
		PeerPubkey: params.PeerPubkey,
		Address:    registerPool.Address,
		InitPos:    registerPool.InitPos,
		TotalPos:   new(big.Int),
	}

	value, err := json.Marshal(candidatePool)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[approveCandidate] Marshal candidatePool error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(CANDIDITE_POOL), peerPubkeyPrefix), &cstates.StorageItem{Value: value})

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

	//get candidatePool
	candidatePoolBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(CANDIDITE_POOL), peerPubkeyPrefix))
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
	native.CloneCache.Delete(scommon.ST_STORAGE, concatKey(contract, []byte(CANDIDITE_POOL), peerPubkeyPrefix))

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

	//syncNodePool storage
	//check syncNodePool
	v, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(SYNC_NODE_POOL), peerPubkeyPrefix))
	if v != nil {
		return errors.NewErr("[registerSyncNode] PeerPubkey is already in syncNodePool!")
	}

	syncNodePool := &states.SyncNodePool{
		Address:  params.Address,
		InitPos:  params.InitPos,
		TotalPos: new(big.Int),
	}

	value, err := json.Marshal(syncNodePool)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[registerSyncNode] Marshal syncNodePool error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(SYNC_NODE_POOL), peerPubkeyPrefix), &cstates.StorageItem{Value: value})

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

	//get syncNodePool
	syncNodePoolBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(SYNC_NODE_POOL), peerPubkeyPrefix))
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
	native.CloneCache.Delete(scommon.ST_STORAGE, concatKey(contract, []byte(SYNC_NODE_POOL), peerPubkeyPrefix))

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

	total := new(big.Int)
	for peerPubkey, pos := range params.VoteTable {
		peerPubkeyPrefix, err := hex.DecodeString(peerPubkey)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[voteForPeer] PeerPubkey format error!")
		}

		//get candidatePool
		candidatePoolBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(CANDIDITE_POOL), peerPubkeyPrefix))
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[voteForPeer] Get candidatePoolBytes error!")
		}
		candidatePool := new(states.CandidatePool)
		if candidatePoolBytes == nil {
			continue
		}
		candidatePoolStore, _ := candidatePoolBytes.(*cstates.StorageItem)
		err = json.Unmarshal(candidatePoolStore.Value, candidatePool)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[voteForPeer] Unmarshal candidatePool error!")
		}

		voteBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(VOTE_INFO_POOL),
			view.Bytes(), peerPubkeyPrefix, addressPrefix))
		vote := new(big.Int)
		if voteBytes != nil {
			voteStore, _ := voteBytes.(*cstates.StorageItem)
			temp := new(big.Int).SetBytes(voteStore.Value)
			vote = new(big.Int).Add(pos, temp)
		} else {
			vote = pos
		}
		if vote.Cmp(new(big.Int)) >= 0 {
			total = new(big.Int).Add(total, pos)
			native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(VOTE_INFO_POOL), view.Bytes(),
				peerPubkeyPrefix, addressPrefix), &cstates.StorageItem{Value: vote.Bytes()})

			candidatePool.TotalPos = new(big.Int).Add(candidatePool.TotalPos, pos)
			value, err := json.Marshal(candidatePool)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[voteForPeer] Marshal candidatePool error")
			}
			native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(CANDIDITE_POOL), peerPubkeyPrefix), &cstates.StorageItem{Value: value})
		}
	}
	if total.Cmp(new(big.Int)) > 0 {
		//TODO: pay
	}
	if total.Cmp(new(big.Int)) < 0 {
		//TODO: withdraw
	}

	addCommonEvent(native, contract, VOTE_FORP_EER, params)

	return nil
}

func CommitDpos(native *NativeService) error {
	//TODO: check witness
	//err = validateOwner(native, params.Address)
	//if err != nil {
	//	return errors.NewDetailErr(err, errors.ErrNoCode, "[registerCandidate] CheckWitness error!")
	//}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := getGovernanceView(native, contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Get view error!")
	}

	//get all candidatePool
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, concatKey(contract, []byte(CANDIDITE_POOL)))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Get all candidatePool error!")
	}

	peers := []*states.PeerStakeInfo{}
	candidatePool := new(states.CandidatePool)
	for _, v := range stateValues {
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] PeerPubkey format error!")
		}
		candidatePoolStore, _ := v.Value.(*cstates.StorageItem)
		err = json.Unmarshal(candidatePoolStore.Value, candidatePool)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Unmarshal candidatePool error!")
		}
		//fmt.Println(candidatePool)
		peers = append(peers, &states.PeerStakeInfo{
			Index:      candidatePool.Index.Uint64(),
			PeerPubkey: candidatePool.PeerPubkey,
			Stake:      candidatePool.TotalPos.Uint64(),
		})

		//update candidatePool
		candidatePool.TotalPos = new(big.Int)
		value, err := json.Marshal(candidatePool)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Marshal candidatePool error")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, []byte(v.Key), &cstates.StorageItem{Value: value})
	}

	// sort peers by stake
	sort.Slice(peers, func(i, j int) bool {
		return peers[i].Stake > peers[j].Stake
	})

	// get stake sum of top-k peers
	var sum uint64
	for i := 0; i < int(configK); i++ {
		sum += peers[i].Stake
	}

	// calculate peer ranks
	scale := configL/configK - 1
	if scale <= 0 {
		return errors.NewErr("[commitDpos] L is equal or less than K!")
	}

	peerRanks := make([]uint64, 0)
	for i := 0; i < int(configK); i++ {
		if peers[i].Stake == 0 {
			return errors.NewErr(fmt.Sprintf("[commitDpos] peers rank %d, has zero stake!", i))
		}
		s := uint64(math.Ceil(float64(peers[i].Stake) * float64(scale) * float64(configK) / float64(sum)))
		peerRanks = append(peerRanks, s)
	}

	// calculate dpos table
	dposTable := make([]uint64, 0)
	for i := 0; i < int(configK); i++ {
		for j := uint64(0); j < peerRanks[i]; j++ {
			dposTable = append(dposTable, peers[i].Index)
		}
	}

	// shuffle
	for i := len(dposTable) - 1; i > 0; i-- {
		h, err := my_hash(native.Tx.Hash(), native.Height, peers[dposTable[i]-1].PeerPubkey, i)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Failed to calculate hash value!")
		}
		j := h % uint64(i)
		dposTable[i], dposTable[j] = dposTable[j], dposTable[i]
	}
	fmt.Println("DPOS table is:", dposTable)

	//update view
	view = new(big.Int).Add(view, new(big.Int).SetInt64(1))
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(GOVERNANCE_VIEW)), &cstates.StorageItem{Value: view.Bytes()})

	addCommonEvent(native, contract, COMMIT_DPOS, true)

	return nil
}