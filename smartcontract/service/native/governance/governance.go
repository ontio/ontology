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

//Governance contract:
//Users can apply for a candidate node to join consensus selection, deposit ONT to authorize for candidate nodes, quit selection and unAuthorize for candidate nodes through this contract.
//ONT deposited in the contract can get ONG bonus which come from transaction fee of the network.
package governance

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/constants"
	cstates "github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/global_params"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"sort"
)

const (
	//status
	RegisterCandidateStatus Status = iota
	CandidateStatus
	ConsensusStatus
	QuitConsensusStatus
	QuitingStatus
	BlackStatus
)

const (
	//function name
	INIT_CONFIG = "initConfig"
	COMMIT_DPOS = "commitDpos"
	//key prefix
	SIDE_CHAIN_ID   = "sideChainID"
	PEER_POOL       = "peerPool"
	VBFT_CONFIG     = "vbftConfig"
	GLOBAL_PARAM    = "globalParam"
	GLOBAL_PARAM2   = "globalParam2"
	SPLIT_CURVE     = "splitCurve"
	GOVERNANCE_VIEW = "governanceView"

	//global
	PRECISE           = 1000000
	//NEW_VERSION_VIEW  = 6
	//NEW_VERSION_BLOCK = 414100
)

// candidate fee must >= 1 ONG
//var MIN_CANDIDATE_FEE = uint64(math.Pow(10, constants.ONG_DECIMALS))
//var AUTHORIZE_INFO_POOL = []byte{118, 111, 116, 101, 73, 110, 102, 111, 80, 111, 111, 108}
var Xi = []uint32{
	0, 100000, 200000, 300000, 400000, 500000, 600000, 700000, 800000, 900000, 1000000, 1100000, 1200000, 1300000, 1400000,
	1500000, 1600000, 1700000, 1800000, 1900000, 2000000, 2100000, 2200000, 2300000, 2400000, 2500000, 2600000, 2700000,
	2800000, 2900000, 3000000, 3100000, 3200000, 3300000, 3400000, 3500000, 3600000, 3700000, 3800000, 3900000, 4000000,
	4100000, 4200000, 4300000, 4400000, 4500000, 4600000, 4700000, 4800000, 4900000, 5000000, 5100000, 5200000, 5300000,
	5400000, 5500000, 5600000, 5700000, 5800000, 5900000, 6000000, 6100000, 6200000, 6300000, 6400000, 6500000, 6600000,
	6700000, 6800000, 6900000, 7000000, 7100000, 7200000, 7300000, 7400000, 7500000, 7600000, 7700000, 7800000, 7900000,
	8000000, 8100000, 8200000, 8300000, 8400000, 8500000, 8600000, 8700000, 8800000, 8900000, 9000000, 9100000, 9200000,
	9300000, 9400000, 9500000, 9600000, 9700000, 9800000, 9900000, 10000000,
}

//Init governance contract address
func InitGovernance() {
	native.Contracts[utils.GovernanceContractAddress] = RegisterGovernanceContract
}

//Register methods of governance contract
func RegisterGovernanceContract(native *native.NativeService) {
	native.Register(INIT_CONFIG, InitConfig)
	native.Register(COMMIT_DPOS, CommitDpos)
}

//Init governance contract, include vbft config, global param and ontid admin.
func InitConfig(native *native.NativeService) ([]byte, error) {
	configuration := new(config.VBFTConfig)
	//buf, err := serialization.ReadVarBytes(bytes.NewBuffer(native.Input))
	//if err != nil {
	//	return utils.BYTE_FALSE, fmt.Errorf("serialization.ReadVarBytes, contract params deserialize error: %v", err)
	//}
	//if err := configuration.Deserialize(bytes.NewBuffer(buf)); err != nil {
	if err := configuration.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deserialize, contract params deserialize error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	// check if initConfig is already execute
	governanceViewBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(GOVERNANCE_VIEW)))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getGovernanceView, get governanceViewBytes error: %v", err)
	}
	if governanceViewBytes != nil {
		return utils.BYTE_FALSE, fmt.Errorf("initConfig. initConfig is already executed")
	}

	//check the configuration
	err = CheckVBFTConfig(configuration)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("checkVBFTConfig failed: %v", err)
	}

	var view uint32 = 1

	peerPoolMap := &PeerPoolMap{
		PeerPoolMap: make(map[string]*PeerPoolItem),
	}
	for _, peer := range configuration.Peers {
		address, err := common.AddressFromBase58(peer.Address)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("common.AddressFromBase58, address format error: %v", err)
		}

		peerPoolItem := new(PeerPoolItem)
		peerPoolItem.Index = peer.Index
		peerPoolItem.PeerPubkey = peer.PeerPubkey
		peerPoolItem.Address = address
		peerPoolItem.InitPos = peer.InitPos
		peerPoolItem.TotalPos = 0
		peerPoolItem.Status = ConsensusStatus
		peerPoolMap.PeerPoolMap[peerPoolItem.PeerPubkey] = peerPoolItem
	}

	err = putSideChainID(native, contract, &SideChainID{configuration.SideChainID})
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putSideChainID, put sideChainID error: %v", err)
	}

	//init peer pool
	err = putPeerPoolMap(native, contract, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putPeerPoolMap, put peerPoolMap error: %v", err)
	}
	globalParam := &GlobalParam{
		CandidateFee: 500000000000,
		MinInitStake: configuration.MinInitStake,
		CandidateNum: 7 * 7,
		PosLimit:     20,
		A:            50,
		B:            50,
		Yita:         5,
		Penalty:      5,
	}
	if err = putGlobalParam(native, contract, globalParam) {
		return utils.BYTE_FALSE, fmt.Errorf("putGlobalParam, put globalParam error: %v", err)
	}

	//init governance view
	governanceView := &GovernanceView{
		View:   view,
		Height: native.Height,
		TxHash: native.Tx.Hash(),
	}
	err = putGovernanceView(native, contract, governanceView)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putGovernanceView, put governanceView error: %v", err)
	}

	//init config
	config := &Configuration{
		N:                    configuration.N,
		C:                    configuration.C,
		K:                    configuration.K,
		L:                    configuration.L,
		BlockMsgDelay:        configuration.BlockMsgDelay,
		HashMsgDelay:         configuration.HashMsgDelay,
		PeerHandshakeTimeout: configuration.PeerHandshakeTimeout,
		MaxBlockChangeView:   configuration.MaxBlockChangeView,
	}
	err = putConfig(native, contract, config)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putConfig, put config error: %v", err)
	}

	//init splitCurve
	splitCurve := &SplitCurve{
		Yi: []uint32{
			0, 95123, 180968, 258213, 327493, 389401, 444491, 493282, 536257, 573866, 606531, 634645, 658574, 678660, 695220, 708550,
			718927, 726606, 731826, 734808, 735759, 734870, 732317, 728265, 722867, 716262, 708583, 699949, 690472, 680254, 669391,
			657969, 646069, 633765, 621124, 608209, 595076, 581778, 568361, 554869, 541342, 527814, 514317, 500882, 487534, 474297,
			461191, 448236, 435447, 422839, 410425, 398217, 386223, 374452, 362910, 351604, 340537, 329713, 319135, 308805, 298723,
			288890, 279306, 269969, 260879, 252033, 243429, 235066, 226939, 219045, 211382, 203945, 196731, 189736, 182955, 176384,
			170018, 163854, 157887, 152113, 146526, 141122, 135896, 130845, 125963, 121246, 116690, 112290, 108041, 103940, 99981,
			96162, 92477, 88923, 85496, 82192, 79006, 75936, 72977, 70126, 67380,
		},
	}
	err = putSplitCurve(native, contract, splitCurve)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putSplitCurve, put splitCurve error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

func CommitDpos(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	address, err := getSyncAddress(native)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getSyncAddress, get syncAddress error: %v", err)
	}
	err = utils.ValidateOwner(native, address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("validateOwner, checkWitness error: %v", err)
	}
	splitFee(native, contract)
	commitDposParam := new(CommitDposParam)
	buf, err := serialization.ReadVarBytes(bytes.NewBuffer(native.Input))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("serialization.ReadVarBytes, contract params deserialize error: %v", err)
	}
	if err := commitDposParam.Deserialize(bytes.NewBuffer(buf)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deserialize, contract params deserialize error: %v", err)
	}
	//input governance view
	err = putGovernanceView(native, contract, commitDposParam.GovernanceView)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putGovernanceView, put governanceView error: %v", err)
	}
	//input configuration
	err = putConfig(native, contract, commitDposParam.Configuration)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putConfig, put config error: %v", err)
	}
	//input global param
	err = putGlobalParam(native, contract, commitDposParam.GlobalParam)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putGlobalParam, put globalParam error: %v", err)
	}
	//input global param2
	err = putGlobalParam2(native, contract, commitDposParam.GlobalParam2)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putGlobalParam2, put globalParam2 error: %v", err)
	}
	//input split curve
	err = putSplitCurve(native, contract, commitDposParam.SplitCurve)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putSplitCurve, put splitCurve error: %v", err)
	}
	//input peer pool map
	peerPoolMap := &PeerPoolMap{
		PeerPoolMap: make(map[string]*PeerPoolItem),
	}
	nodeList := commitDposParam.SideChainNodeInfo.NodeInfoMap
	for k := range commitDposParam.PeerPoolMap.PeerPoolMap {
		if _, ok := nodeList[k]; ok {
			//exist
			peerPoolMap.PeerPoolMap[k] = commitDposParam.PeerPoolMap.PeerPoolMap[k]
		}
	}
	// get config
	config, err := getConfig(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getConfig, get config error: %v", err)
	}
	if len(peerPoolMap.PeerPoolMap) < int(config.K) {
		return utils.BYTE_FALSE, fmt.Errorf("length of peer pool map is less than config.K")
	}
	err = putPeerPoolMap(native, contract, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putPeerPoolMap, put peerPoolMap error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func splitFee(native *native.NativeService, contract common.Address) error {
	// get config
	config, err := getConfig(native, contract)
	if err != nil {
		return fmt.Errorf("getConfig, get config error: %v", err)
	}
	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract)
	if err != nil {
		return fmt.Errorf("splitFee, get peerPoolMap error: %v", err)
	}
	balance, err := getOngBalance(native, utils.GovernanceContractAddress)
	if err != nil {
		return fmt.Errorf("splitFee, getOngBalance error: %v", err)
	}
	//get globalParam
	globalParam, err := getGlobalParam(native, contract)
	if err != nil {
		return fmt.Errorf("getGlobalParam, getGlobalParam error: %v", err)
	}

	peersCandidate := []*CandidateSplitInfo{}

	for _, peerPoolItem := range peerPoolMap.PeerPoolMap {
		if peerPoolItem.Status == CandidateStatus || peerPoolItem.Status == ConsensusStatus {
			stake := peerPoolItem.TotalPos + peerPoolItem.InitPos
			peersCandidate = append(peersCandidate, &CandidateSplitInfo{
				PeerPubkey: peerPoolItem.PeerPubkey,
				InitPos:    peerPoolItem.InitPos,
				Address:    peerPoolItem.Address,
				Stake:      stake,
			})
		}
	}
	// sort peers by stake
	sort.SliceStable(peersCandidate, func(i, j int) bool {
		if peersCandidate[i].Stake > peersCandidate[j].Stake {
			return true
		} else if peersCandidate[i].Stake == peersCandidate[j].Stake {
			return peersCandidate[i].PeerPubkey > peersCandidate[j].PeerPubkey
		}
		return false
	})
	// cal s of each consensus node
	var sum uint64
	for i := 0; i < int(config.K); i++ {
		sum += peersCandidate[i].Stake
	}
	// if sum = 0, means consensus peer in config, do not split
	if sum < uint64(config.K) {
		return nil
	}
	avg := sum / uint64(config.K)
	var sumS uint64
	for i := 0; i < int(config.K); i++ {
		peersCandidate[i].S, err = splitCurve(native, contract, peersCandidate[i].Stake, avg, uint64(globalParam.Yita))
		if err != nil {
			return fmt.Errorf("splitCurve, calculate splitCurve error: %v", err)
		}
		sumS += peersCandidate[i].S
	}
	if sumS == 0 {
		return fmt.Errorf("splitFee, sumS is 0")
	}
	//fee split of consensus peer
	for i := 0; i < int(config.K); i++ {
		nodeAmount := balance * uint64(globalParam.A) / 100 * peersCandidate[i].S / sumS
		address := peersCandidate[i].Address
		err = appCallTransferOng(native, utils.GovernanceContractAddress, address, nodeAmount)
		if err != nil {
			return fmt.Errorf("splitFee, ong transfer error: %v", err)
		}
	}

	//fee split of candidate peer
	//cal s of each candidate node
	//get globalParam2
	globalParam2, err := getGlobalParam2(native, contract)
	if err != nil {
		return fmt.Errorf("getGlobalParam2, getGlobalParam2 error: %v", err)
	}
	var length int
	if int(globalParam2.CandidateFeeSplitNum) >= len(peersCandidate) {
		length = len(peersCandidate)
	} else {
		length = int(globalParam2.CandidateFeeSplitNum)
	}
	sum = 0
	for i := int(config.K); i < length; i++ {
		sum += peersCandidate[i].Stake
	}
	if sum == 0 {
		return nil
	}
	for i := int(config.K); i < length; i++ {
		nodeAmount := balance * uint64(globalParam.B) / 100 * peersCandidate[i].Stake / sum
		address := peersCandidate[i].Address
		err = appCallTransferOng(native, utils.GovernanceContractAddress, address, nodeAmount)
		if err != nil {
			return fmt.Errorf("splitFee, ong transfer error: %v", err)
		}
	}
	return nil
}

//Go to next consensus epoch
func CommitDpos(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	// get config
	config, err := getConfig(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getConfig, get config error: %v", err)
	}

	//get governace view
	governanceView, err := GetGovernanceView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getGovernanceView, get GovernanceView error: %v", err)
	}

	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getAdmin, get admin error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		cycle := (native.Height - governanceView.Height) >= config.MaxBlockChangeView
		if !cycle {
			return utils.BYTE_FALSE, fmt.Errorf("commitDpos, authentication Failed")
		}
	}

	err = executeCommitDpos(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("executeCommitDpos, executeCommitDpos error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}
