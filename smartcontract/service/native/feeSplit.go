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
	"github.com/ontio/ontology/common"
)

const (
	EXECUTE_SPLIT = "executeSplit"
	a = 0.75
	b = 0.2
	c = 0.05
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
			stake := peerPool.TotalPos + peerPool.InitPos
			peersCandidate = append(peersCandidate, &states.CandidateSplitInfo{
				PeerPubkey: peerPool.PeerPubkey,
				Stake:      float64(stake),
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
	var splitAmount uint64
	for i := int(config.K) - 1; i > 0; i-- {
		nodeAmount := TOTAL_ONG * a * peersCandidate[i].S / sumS
		fmt.Println("Amount of node i is:", nodeAmount)
		peerPubkeyPrefix, err := hex.DecodeString(peersCandidate[i].PeerPubkey)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] PeerPubkey format error!")
		}
		stateValues, err = native.CloneCache.Store.Find(scommon.ST_STORAGE, concatKey(contract, []byte(VOTE_INFO_POOL),
			view.Bytes(), peerPubkeyPrefix))
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Get all peerPool error!")
		}
		voteInfoPool := new(states.VoteInfoPool)
		for _, v := range stateValues {
			voteInfoPoolStore, _ := v.Value.(*cstates.StorageItem)
			err = json.Unmarshal(voteInfoPoolStore.Value, voteInfoPool)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Unmarshal voteInfoPool error!")
			}
			addressBytes, err := hex.DecodeString(voteInfoPool.Address)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Address format error!")
			}
			address, err := common.AddressParseFromBytes(addressBytes)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Address format error!")
			}
			pos := voteInfoPool.PrePos + voteInfoPool.PreFreezePos + voteInfoPool.FreezePos + voteInfoPool.NewPos
			amount := uint64(nodeAmount * float64(pos) / sum)

			//ong transfer
			err = appCallTransferOng(native, genesis.FeeSplitContractAddress, address, new(big.Int).SetUint64(pos))
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Ong transfer error!")
			}
			fmt.Printf("Amount of address %v is: %d", voteInfoPool.Address, amount)
			splitAmount = splitAmount + amount
		}
	}
	//split remained amount


	return nil
}










