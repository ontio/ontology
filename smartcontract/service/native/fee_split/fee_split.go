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

package fee_split

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"sort"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/genesis"
	cstates "github.com/ontio/ontology/core/states"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/governance"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	EXECUTE_SPLIT = "executeSplit"
	A             = 0.5
	B             = 0.45
	C             = 0.05
	TOTAL_ONG     = 10000000000
	PRECISE       = 100000000
	YITA          = 5
)

var Xi = []uint64{
	0, 100000000, 200000000, 300000000, 400000000, 500000000, 600000000, 700000000, 800000000, 900000000, 1000000000, 1100000000,
	1200000000, 1300000000, 1400000000, 1500000000, 1600000000, 1700000000, 1800000000, 1900000000, 2000000000, 2100000000, 2200000000,
	2300000000, 2400000000, 2500000000, 2600000000, 2700000000, 2800000000, 2900000000, 3000000000, 3100000000, 3200000000, 3300000000,
	3400000000, 3500000000, 3600000000, 3700000000, 3800000000, 3900000000, 4000000000, 4100000000, 4200000000, 4300000000, 4400000000,
	4500000000, 4600000000, 4700000000, 4800000000, 4900000000, 5000000000, 5100000000, 5200000000, 5300000000, 5400000000, 5500000000,
	5600000000, 5700000000, 5800000000, 5900000000, 6000000000, 6100000000, 6200000000, 6300000000, 6400000000, 6500000000, 6600000000,
	6700000000, 6800000000, 6900000000, 7000000000, 7100000000, 7200000000, 7300000000, 7400000000, 7500000000, 7600000000, 7700000000,
	7800000000, 7900000000, 8000000000, 8100000000, 8200000000, 8300000000, 8400000000, 8500000000, 8600000000, 8700000000, 8800000000,
	8900000000, 9000000000, 9100000000, 9200000000, 9300000000, 9400000000, 9500000000, 9600000000, 9700000000, 9800000000, 9900000000,
	10000000000,
}

var Yi = []uint64{
	0, 95122943, 180967484, 258212393, 327492302, 389400392, 444490933, 493281663, 536256037, 573865337, 606530660, 634644792,
	658573964, 678659510, 695219426, 708549830, 718926343, 726605385, 731825388, 734807945, 735758883, 734869274, 732316385, 728264570,
	722866109, 716261993, 708582662, 699948704, 690471500, 680253836, 669390481, 657968719, 646068858, 633764699, 621123982, 608208803,
	595075998, 581777516, 568360754, 554868880, 541341133, 527813105, 514316999, 500881879, 487533897, 474296511, 461190682, 448235063,
	435446176, 422838574, 410424994, 398216497, 386222607, 374451430, 362909769, 351603237, 340536351, 329712629, 319134677, 308804266,
	298722411, 288889439, 279305055, 269968400, 260878106, 252032351, 243428905, 235065173, 226938236, 219044892, 211381684, 203944942,
	196730802, 189735241, 182954096, 176383094, 170017867, 163853971, 157886910, 152112145, 146525112, 141121235, 135895939, 130844657,
	125962846, 121245989, 116689608, 112289270, 108040592, 103939247, 99980969, 96161560, 92476889, 88922898, 85495605, 82191105,
	79005572, 75935263, 72976515, 70125749, 67379470,
}

func InitFeeSplit() {
	native.Contracts[genesis.FeeSplitContractAddress] = RegisterFeeSplitContract
}

func RegisterFeeSplitContract(native *native.NativeService) {
	native.Register(EXECUTE_SPLIT, ExecuteSplit)
}

func ExecuteSplit(native *native.NativeService) ([]byte, error) {
	contract := genesis.GovernanceContractAddress
	//get current view
	cView, err := governance.GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, get view error!")
	}
	view := new(big.Int).Sub(cView, new(big.Int).SetInt64(1))

	//get peerPoolMap
	peerPoolMap, err := governance.GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, get peerPoolMap error!")
	}
	peersCandidate := []*CandidateSplitInfo{}
	peersSyncNode := []*SyncNodeSplitInfo{}

	for _, peerPool := range peerPoolMap.PeerPoolMap {
		if peerPool.Status == governance.CandidateStatus || peerPool.Status == governance.ConsensusStatus {
			stake := peerPool.TotalPos + peerPool.InitPos
			peersCandidate = append(peersCandidate, &CandidateSplitInfo{
				PeerPubkey: peerPool.PeerPubkey,
				InitPos:    peerPool.InitPos,
				Address:    peerPool.Address,
				Stake:      stake,
			})
		}
		if peerPool.Status == governance.SyncNodeStatus || peerPool.Status == governance.RegisterCandidateStatus {
			peersSyncNode = append(peersSyncNode, &SyncNodeSplitInfo{
				PeerPubkey: peerPool.PeerPubkey,
				InitPos:    peerPool.InitPos,
				Address:    peerPool.Address,
			})
		}
	}

	// get config
	config := new(governance.Configuration)
	configBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(governance.VBFT_CONFIG)))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, get configBytes error!")
	}
	if configBytes == nil {
		return utils.BYTE_FALSE, errors.NewErr("executeSplit, configBytes is nil!")
	}
	configStore, _ := configBytes.(*cstates.StorageItem)
	if err := config.Deserialize(bytes.NewBuffer(configStore.Value)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize config error!")
	}

	// sort peers by stake
	sort.Slice(peersCandidate, func(i, j int) bool {
		return peersCandidate[i].Stake > peersCandidate[j].Stake
	})

	// cal s of each consensus node
	var sum uint64
	for i := 0; i < int(config.K); i++ {
		sum += peersCandidate[i].Stake
	}
	avg := sum / uint64(config.K)
	var sumS uint64
	for i := 0; i < int(config.K); i++ {
		peersCandidate[i].S = splitCurve(peersCandidate[i].Stake, avg)
		sumS += peersCandidate[i].S
	}

	//fee split of consensus peer
	var splitAmount uint64
	remainCandidate := peersCandidate[0]
	for i := int(config.K) - 1; i >= 0; i-- {
		if peersCandidate[i].PeerPubkey > remainCandidate.PeerPubkey {
			remainCandidate = peersCandidate[i]
		}

		nodeAmount := uint64(TOTAL_ONG * A * peersCandidate[i].S / sumS)
		addressBytes, err := hex.DecodeString(peersCandidate[i].Address)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, address format error!")
		}
		address, err := common.AddressParseFromBytes(addressBytes)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, address format error!")
		}
		err = AppCallApproveOng(native, genesis.FeeSplitContractAddress, address, nodeAmount)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, ong transfer error!")
		}
		splitAmount += nodeAmount
	}
	//split remained amount
	remainAmount := TOTAL_ONG*A - splitAmount
	remainAddressBytes, err := hex.DecodeString(remainCandidate.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, address format error!")
	}
	remainAddress, err := common.AddressParseFromBytes(remainAddressBytes)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, address format error!")
	}
	err = AppCallApproveOng(native, genesis.FeeSplitContractAddress, remainAddress, remainAmount)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, ong transfer error!")
	}

	//fee split of candidate peer
	// cal s of each candidate node
	sum = 0
	for i := int(config.K); i < len(peersCandidate); i++ {
		sum += peersCandidate[i].Stake
	}
	splitAmount = 0
	remainCandidate = peersCandidate[int(config.K)]
	for i := int(config.K); i < len(peersCandidate); i++ {
		if peersCandidate[i].PeerPubkey > remainCandidate.PeerPubkey {
			remainCandidate = peersCandidate[i]
		}

		nodeAmount := uint64(TOTAL_ONG * B * peersCandidate[i].Stake / sum)
		addressBytes, err := hex.DecodeString(peersCandidate[i].Address)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, address format error!")
		}
		address, err := common.AddressParseFromBytes(addressBytes)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, address format error!")
		}
		err = AppCallApproveOng(native, genesis.FeeSplitContractAddress, address, nodeAmount)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, ong transfer error!")
		}
		splitAmount += nodeAmount
	}
	//split remained amount
	remainAmount = TOTAL_ONG*B - splitAmount
	remainAddressBytes, err = hex.DecodeString(remainCandidate.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, address format error!")
	}
	remainAddress, err = common.AddressParseFromBytes(remainAddressBytes)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, address format error!")
	}
	err = AppCallApproveOng(native, genesis.FeeSplitContractAddress, remainAddress, remainAmount)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, ong transfer error!")
	}

	//fee split of syncNode peer
	// cal s of each candidate node
	var splitSyncNodeAmount uint64
	for _, syncNodeSplitInfo := range peersSyncNode {
		amount := uint64(TOTAL_ONG * C / len(peersSyncNode))
		addressBytes, err := hex.DecodeString(syncNodeSplitInfo.Address)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, address format error!")
		}
		address, err := common.AddressParseFromBytes(addressBytes)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, address format error!")
		}
		err = AppCallApproveOng(native, genesis.FeeSplitContractAddress, address, amount)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Ong transfer error!")
		}
		splitSyncNodeAmount += amount
	}
	remainSyncNodeAmount := TOTAL_ONG*C - splitSyncNodeAmount

	// sort peers by peerPubkey
	sort.Slice(peersSyncNode, func(i, j int) bool {
		return peersSyncNode[i].PeerPubkey > peersSyncNode[j].PeerPubkey
	})

	addressBytes, err := hex.DecodeString(peersSyncNode[0].Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, address format error!")
	}
	address, err := common.AddressParseFromBytes(addressBytes)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, address format error!")
	}
	err = AppCallApproveOng(native, genesis.FeeSplitContractAddress, address, remainSyncNodeAmount)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, ong transfer error!")
	}

	utils.AddCommonEvent(native, genesis.FeeSplitContractAddress, EXECUTE_SPLIT, true)

	return utils.BYTE_TRUE, nil
}
