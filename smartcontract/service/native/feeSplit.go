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

	"github.com/ontio/ontology/core/genesis"
	cstates "github.com/ontio/ontology/core/states"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native/states"
	"math/big"
	"sort"
	"fmt"
	"encoding/hex"
)

const (
	EXECUTE_SPLIT = "executeSplit"
	TOTAL_ONG = 10000000000
)

func init() {
	Contracts[genesis.FeeSplitContractAddress] = RegisterFeeSplitContract
}

func RegisterFeeSplitContract(native *NativeService) {
	native.Register(EXECUTE_SPLIT, ExecuteSplit)
}

func ExecuteSplit(native *NativeService) error {
	contract := genesis.GovernanceContractAddress
	//get current view
	cView, err := getGovernanceView(native, contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Get view error!")
	}
	view := new(big.Int).Sub(cView, new(big.Int).SetInt64(1))

	//get all peerPool
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), view.Bytes()))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Get all peerPool error!")
	}
	peersCandidate := []*states.CandidateSplitInfo{}
	peerPool := new(states.PeerPool)
	for _, v := range stateValues {
		peerPoolStore, _ := v.Value.(*cstates.StorageItem)
		err = json.Unmarshal(peerPoolStore.Value, peerPool)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Unmarshal peerPool error!")
		}
		if peerPool.Status == CandidateStatus || peerPool.Status == ConsensusStatus {
			stake := new(big.Int).Add(peerPool.TotalPos, peerPool.InitPos)
			peersCandidate = append(peersCandidate, &states.CandidateSplitInfo{
				PeerPubkey: peerPool.PeerPubkey,
				Stake:      float64(stake.Uint64()),
			})
		}
	}
	// get config
	config := new(states.Configuration)
	configBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(VBFT_CONFIG)))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Get configBytes error!")
	}
	if configBytes == nil {
		return errors.NewErr("[executeSplit] ConfigBytes is nil!")
	}
	configStore, _ := configBytes.(*cstates.StorageItem)
	err = json.Unmarshal(configStore.Value, config)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Unmarshal config error!")
	}

	// sort peers by stake
	sort.Slice(peersCandidate, func(i, j int) bool {
		return peersCandidate[i].Stake > peersCandidate[j].Stake
	})

	// cal s of each consensus node
	var sum float64
	for i := 0; i < int(config.K); i++ {
		sum += peersCandidate[i].Stake
	}
	avg := sum/float64(config.K)
	var sumS float64
	for i := 0; i < int(config.K); i++ {
		peersCandidate[i].S = (0.5 * peersCandidate[i].Stake) / (2 * avg)
		sumS += peersCandidate[i].S
	}

	//fee split of consensus peer
	for i := int(config.K) - 1; i > 0; i-- {
		amount := uint64(TOTAL_ONG * peersCandidate[i].S / sumS)
		fmt.Println("Amount of node i is:", amount)
		peerPubkeyPrefix, err := hex.DecodeString(peersCandidate[i].PeerPubkey)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] PeerPubkey format error!")
		}
		stateValues, err = native.CloneCache.Store.Find(scommon.ST_STORAGE, concatKey(contract, []byte(VOTE_INFO_POOL),
			view.Bytes(), peerPubkeyPrefix))
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Get all peerPool error!")
		}
		voteInfoPool := new(states.VoteInfoPool)
		for _, v := range stateValues {
			voteInfoPoolStore, _ := v.Value.(*cstates.StorageItem)
			err = json.Unmarshal(voteInfoPoolStore.Value, voteInfoPool)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Unmarshal voteInfoPool error!")
			}
			//addressPrefix, err := hex.DecodeString(voteInfoPool.Address)
			//if err != nil {
			//	errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Address format error!")
			//}
		}
	}


	return nil
}










