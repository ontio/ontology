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
	"fmt"
	"math/big"
	"sort"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/genesis"
	cstates "github.com/ontio/ontology/core/states"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native/states"
)

const (
	EXECUTE_SPLIT = "executeSplit"
	a             = 0.75
	b             = 0.2
	c             = 0.05
	TOTAL_ONG     = 10000000000
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

	//get peerPoolMap
	peerPoolMap, err := getPeerPoolMap(native, contract, view)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Get peerPoolMap error!")
	}
	peersCandidate := []*states.CandidateSplitInfo{}
	peersSyncNode := []*states.SyncNodeSplitInfo{}

	for _, peerPool := range peerPoolMap.PeerPoolMap {
		if peerPool.Status == CandidateStatus || peerPool.Status == ConsensusStatus {
			stake := peerPool.TotalPos + peerPool.InitPos
			peersCandidate = append(peersCandidate, &states.CandidateSplitInfo{
				PeerPubkey: peerPool.PeerPubkey,
				InitPos:    peerPool.InitPos,
				Address:    peerPool.Address,
				Stake:      float64(stake),
			})
		}
		if peerPool.Status == SyncNodeStatus || peerPool.Status == RegisterCandidateStatus {
			peersSyncNode = append(peersSyncNode, &states.SyncNodeSplitInfo{
				PeerPubkey: peerPool.PeerPubkey,
				InitPos:    peerPool.InitPos,
				Address:    peerPool.Address,
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
	avg := sum / float64(config.K)
	var sumS float64
	for i := 0; i < int(config.K); i++ {
		peersCandidate[i].S = (0.5 * peersCandidate[i].Stake) / (2 * avg)
		sumS += peersCandidate[i].S
	}

	//fee split of consensus peer
	fmt.Println("###############################################################")
	var splitAmount uint64
	remainCandidate := peersCandidate[0]
	for i := int(config.K) - 1; i >= 0; i-- {
		if peersCandidate[i].PeerPubkey > remainCandidate.PeerPubkey {
			remainCandidate = peersCandidate[i]
		}

		nodeAmount := uint64(TOTAL_ONG * a * peersCandidate[i].S / sumS)
		addressBytes, err := hex.DecodeString(peersCandidate[i].Address)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Address format error!")
		}
		address, err := common.AddressParseFromBytes(addressBytes)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Address format error!")
		}
		err = appCallApproveOng(native, genesis.FeeSplitContractAddress, address, new(big.Int).SetUint64(nodeAmount))
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Ong transfer error!")
		}
		fmt.Printf("Amount of node %v, address %v is %v: \n", i, peersCandidate[i].Address, nodeAmount)
		splitAmount += nodeAmount
	}
	//split remained amount
	remainAmount := TOTAL_ONG*a - splitAmount
	fmt.Println("Remained Amount is : ", remainAmount)
	remainAddressBytes, err := hex.DecodeString(remainCandidate.Address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Address format error!")
	}
	remainAddress, err := common.AddressParseFromBytes(remainAddressBytes)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Address format error!")
	}
	err = appCallApproveOng(native, genesis.FeeSplitContractAddress, remainAddress, new(big.Int).SetUint64(remainAmount))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Ong transfer error!")
	}
	fmt.Printf("Amount of address %v is: %d \n", remainCandidate.Address, remainAmount)

	//fee split of candidate peer
	fmt.Println("###############################################################")
	// cal s of each candidate node
	sum = 0
	for i := int(config.K); i < len(peersCandidate); i++ {
		sum += peersCandidate[i].Stake
	}
	avg = sum / float64(config.K)
	sumS = 0
	for i := int(config.K); i < len(peersCandidate); i++ {
		peersCandidate[i].S = (0.5 * peersCandidate[i].Stake) / (2 * avg)
		sumS += peersCandidate[i].S
	}
	splitAmount = 0
	remainCandidate = peersCandidate[int(config.K)]
	for i := int(config.K); i < len(peersCandidate); i++ {
		if peersCandidate[i].PeerPubkey > remainCandidate.PeerPubkey {
			remainCandidate = peersCandidate[i]
		}

		nodeAmount := uint64(TOTAL_ONG * b * peersCandidate[i].S / sumS)
		addressBytes, err := hex.DecodeString(peersCandidate[i].Address)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Address format error!")
		}
		address, err := common.AddressParseFromBytes(addressBytes)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Address format error!")
		}
		err = appCallApproveOng(native, genesis.FeeSplitContractAddress, address, new(big.Int).SetUint64(nodeAmount))
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Ong transfer error!")
		}
		fmt.Printf("Amount of node %v, address %v is %v: \n", i, peersCandidate[i].Address, nodeAmount)
		splitAmount += nodeAmount
	}
	//split remained amount
	remainAmount = TOTAL_ONG*b - splitAmount
	fmt.Println("Remained Amount is : ", remainAmount)
	remainAddressBytes, err = hex.DecodeString(remainCandidate.Address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Address format error!")
	}
	remainAddress, err = common.AddressParseFromBytes(remainAddressBytes)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Address format error!")
	}
	err = appCallApproveOng(native, genesis.FeeSplitContractAddress, remainAddress, new(big.Int).SetUint64(remainAmount))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Ong transfer error!")
	}
	fmt.Printf("Amount of address %v is: %d \n", remainCandidate.Address, remainAmount)

	//fee split of syncNode peer
	fmt.Println("###############################################################")
	// cal s of each candidate node
	sum = 0
	for i := 0; i < len(peersSyncNode); i++ {
		sum += float64(peersSyncNode[i].InitPos)
	}
	avg = sum / float64(config.K)
	sumS = 0
	for i := 0; i < len(peersSyncNode); i++ {
		peersSyncNode[i].S = (0.5 * float64(peersSyncNode[i].InitPos)) / (2 * avg)
		sumS += peersSyncNode[i].S
	}

	var splitSyncNodeAmount uint64
	for _, syncNodeSplitInfo := range peersSyncNode {
		amount := uint64(TOTAL_ONG * c * syncNodeSplitInfo.S / sumS)
		addressBytes, err := hex.DecodeString(syncNodeSplitInfo.Address)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Address format error!")
		}
		address, err := common.AddressParseFromBytes(addressBytes)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Address format error!")
		}
		err = appCallApproveOng(native, genesis.FeeSplitContractAddress, address, new(big.Int).SetUint64(amount))
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Ong transfer error!")
		}
		fmt.Printf("Amount of address %v is: %d \n", syncNodeSplitInfo.Address, amount)
		splitSyncNodeAmount += amount
	}
	remainSyncNodeAmount := TOTAL_ONG*c - splitSyncNodeAmount
	fmt.Println("RemainSyncNodeAmount is : ", remainSyncNodeAmount)

	// sort peers by peerPubkey
	sort.Slice(peersSyncNode, func(i, j int) bool {
		return peersSyncNode[i].PeerPubkey > peersSyncNode[j].PeerPubkey
	})

	addressBytes, err := hex.DecodeString(peersSyncNode[0].Address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Address format error!")
	}
	address, err := common.AddressParseFromBytes(addressBytes)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Address format error!")
	}
	err = appCallApproveOng(native, genesis.FeeSplitContractAddress, address, new(big.Int).SetUint64(remainSyncNodeAmount))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Ong transfer error!")
	}
	fmt.Printf("Amount of address %v is: %d \n", peersSyncNode[0].Address, remainSyncNodeAmount)

	addCommonEvent(native, genesis.FeeSplitContractAddress, EXECUTE_SPLIT, true)

	return nil
}
