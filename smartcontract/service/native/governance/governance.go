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

package governance

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	cstates "github.com/ontio/ontology/core/states"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/global_params"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	//status
	RegisterCandidateStatus Status = iota
	CandidateStatus
	ConsensusStatus
	QuitConsensusStatus
	QuitingStatus
	BlackConsensusStatus
	BlackStatus
)

const (
	//function name
	INIT_CONFIG         = "initConfig"
	REGISTER_CANDIDATE  = "registerCandidate"
	APPROVE_CANDIDATE   = "approveCandidate"
	REJECT_CANDIDATE    = "rejectCandidate"
	BLACK_NODE          = "blackNode"
	WHITE_NODE          = "whiteNode"
	QUIT_NODE           = "quitNode"
	VOTE_FOR_PEER       = "voteForPeer"
	UNVOTE_FOR_PEER     = "unVoteForPeer"
	WITHDRAW            = "withdraw"
	COMMIT_DPOS         = "commitDpos"
	UPDATE_CONFIG       = "updateConfig"
	UPDATE_GLOBAL_PARAM = "updateGlobalParam"
	CALL_SPLIT          = "callSplit"

	//key prefix
	GLOBAL_PARAM    = "globalParam"
	VBFT_CONFIG     = "vbftConfig"
	GOVERNANCE_VIEW = "governanceView"
	CANDIDITE_INDEX = "candidateIndex"
	PEER_POOL       = "peerPool"
	VOTE_INFO_POOL  = "voteInfoPool"
	PEER_INDEX      = "peerIndex"
	BLACK_LIST      = "blackList"
	TOTAL_STAKE     = "totalStake"

	//global
	PRECISE = 1000000
)

var Xi = []uint64{
	0, 100000, 200000, 300000, 400000, 500000, 600000, 700000, 800000, 900000, 1000000, 1100000, 1200000, 1300000, 1400000,
	1500000, 1600000, 1700000, 1800000, 1900000, 2000000, 2100000, 2200000, 2300000, 2400000, 2500000, 2600000, 2700000,
	2800000, 2900000, 3000000, 3100000, 3200000, 3300000, 3400000, 3500000, 3600000, 3700000, 3800000, 3900000, 4000000,
	4100000, 4200000, 4300000, 4400000, 4500000, 4600000, 4700000, 4800000, 4900000, 5000000, 5100000, 5200000, 5300000,
	5400000, 5500000, 5600000, 5700000, 5800000, 5900000, 6000000, 6100000, 6200000, 6300000, 6400000, 6500000, 6600000,
	6700000, 6800000, 6900000, 7000000, 7100000, 7200000, 7300000, 7400000, 7500000, 7600000, 7700000, 7800000, 7900000,
	8000000, 8100000, 8200000, 8300000, 8400000, 8500000, 8600000, 8700000, 8800000, 8900000, 9000000, 9100000, 9200000,
	9300000, 9400000, 9500000, 9600000, 9700000, 9800000, 9900000, 10000000,
}

var Yi = []uint64{
	0, 95123, 180968, 258213, 327493, 389401, 444491, 493282, 536257, 573866, 606531, 634645, 658574, 678660, 695220, 708550,
	718927, 726606, 731826, 734808, 735759, 734870, 732317, 728265, 722867, 716262, 708583, 699949, 690472, 680254, 669391,
	657969, 646069, 633765, 621124, 608209, 595076, 581778, 568361, 554869, 541342, 527814, 514317, 500882, 487534, 474297,
	461191, 448236, 435447, 422839, 410425, 398217, 386223, 374452, 362910, 351604, 340537, 329713, 319135, 308805, 298723,
	288890, 279306, 269969, 260879, 252033, 243429, 235066, 226939, 219045, 211382, 203945, 196731, 189736, 182955, 176384,
	170018, 163854, 157887, 152113, 146526, 141122, 135896, 130845, 125963, 121246, 116690, 112290, 108041, 103940, 99981,
	96162, 92477, 88923, 85496, 82192, 79006, 75936, 72977, 70126, 67380,
}

func InitGovernance() {
	native.Contracts[utils.GovernanceContractAddress] = RegisterGovernanceContract
}

func RegisterGovernanceContract(native *native.NativeService) {
	native.Register(INIT_CONFIG, InitConfig)
	native.Register(REGISTER_CANDIDATE, RegisterCandidate)
	native.Register(APPROVE_CANDIDATE, ApproveCandidate)
	native.Register(REJECT_CANDIDATE, RejectCandidate)
	native.Register(BLACK_NODE, BlackNode)
	native.Register(WHITE_NODE, WhiteNode)
	native.Register(QUIT_NODE, QuitNode)
	native.Register(VOTE_FOR_PEER, VoteForPeer)
	native.Register(UNVOTE_FOR_PEER, UnVoteForPeer)
	native.Register(WITHDRAW, Withdraw)
	native.Register(COMMIT_DPOS, CommitDpos)
	native.Register(UPDATE_CONFIG, UpdateConfig)
	native.Register(UPDATE_GLOBAL_PARAM, UpdateGlobalParam)
	native.Register(CALL_SPLIT, CallSplit)
}

func InitConfig(native *native.NativeService) ([]byte, error) {
	configuration := new(config.VBFTConfig)
	if err := configuration.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	// check if initConfig is already execute
	governanceViewBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(GOVERNANCE_VIEW)))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getGovernanceView, get governanceViewBytes error!")
	}
	if governanceViewBytes != nil {
		return utils.BYTE_FALSE, errors.NewErr("initConfig. initConfig is already executed!")
	}

	//init globalParam
	globalParam := &GlobalParam{
		CandidateFee: 500,
		MinInitStake: 10000,
		CandidateNum: 7 * 7,
		A:            50,
		B:            50,
		Yita:         5,
	}
	bf := new(bytes.Buffer)
	if err := globalParam.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize globalParam error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(GLOBAL_PARAM)), &cstates.StorageItem{Value: bf.Bytes()})

	//check the configuration
	if configuration.C == 0 {
		return utils.BYTE_FALSE, errors.NewErr("initConfig. C can not be 0 in config!")
	}
	if int(configuration.K) != len(configuration.Peers) {
		return utils.BYTE_FALSE, errors.NewErr("initConfig. K must equal to length of peer in config!")
	}
	if configuration.L < 16*configuration.K {
		return utils.BYTE_FALSE, errors.NewErr("initConfig. L can not be less than 16*K in config!")
	}
	if configuration.K < 2*configuration.C+1 {
		return utils.BYTE_FALSE, errors.NewErr("initConfig. K can not be less than 2*C+1 in config!")
	}
	if configuration.N < configuration.K || configuration.K < 7 {
		return utils.BYTE_FALSE, errors.NewErr("initConfig. config not match N >= K >= 7!")
	}

	var view uint32 = 1

	indexMap := make(map[uint32]struct{})
	var maxId uint32
	//peers := []*PeerStakeInfo{}
	peerPoolMap := &PeerPoolMap{
		PeerPoolMap: make(map[string]*PeerPoolItem),
	}
	for _, peer := range configuration.Peers {
		peerPoolItem := new(PeerPoolItem)
		_, ok := indexMap[peer.Index]
		if ok {
			return utils.BYTE_FALSE, errors.NewErr("initConfig, peer index is duplicated!")
		}
		indexMap[peer.Index] = struct{}{}
		if peer.Index <= 0 {
			return utils.BYTE_FALSE, errors.NewErr("initConfig, peer index in config must > 0!")
		}
		if peer.Index > maxId {
			maxId = peer.Index
		}
		if peer.InitPos < uint64(globalParam.MinInitStake) {
			return utils.BYTE_FALSE, errors.NewErr(fmt.Sprintf("initConfig, initPos must >= %v!", globalParam.MinInitStake))
		}
		//check peerPubkey
		if err := validatePeerPubKeyFormat(peer.PeerPubkey); err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "invalid peer pubkey")
		}
		address, err := common.AddressFromBase58(peer.Address)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "common.AddressFromBase58, address format error!")
		}
		peerPoolItem.Index = peer.Index
		peerPoolItem.PeerPubkey = peer.PeerPubkey
		peerPoolItem.Address = address
		peerPoolItem.InitPos = peer.InitPos
		peerPoolItem.TotalPos = 0
		peerPoolItem.Status = ConsensusStatus
		peerPoolMap.PeerPoolMap[peerPoolItem.PeerPubkey] = peerPoolItem

		peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
		}
		index := peerPoolItem.Index
		indexBytes, err := GetUint32Bytes(index)
		if err != nil {
			return nil, errors.NewDetailErr(err, errors.ErrNoCode, "getUint32Bytes, getUint32Bytes error!")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_INDEX), peerPubkeyPrefix), &cstates.StorageItem{Value: indexBytes})

		//update total stake
		totalStake := &TotalStake{
			Address: address,
			Stake:   peer.InitPos,
		}
		bf := new(bytes.Buffer)
		if err := totalStake.Serialize(bf); err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize totalStake error!")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(TOTAL_STAKE), address[:]), &cstates.StorageItem{Value: bf.Bytes()})

		//peers = append(peers, &PeerStakeInfo{
		//	Index:      peerPoolItem.Index,
		//	PeerPubkey: peerPoolItem.PeerPubkey,
		//	Stake:      peerPoolItem.InitPos,
		//})
	}

	//init peer pool
	bf = new(bytes.Buffer)
	if err := peerPoolMap.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize peerPoolMap error!")
	}
	viewBytes, err := GetUint32Bytes(0)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "getUint32Bytes, get viewBytes error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), viewBytes), &cstates.StorageItem{Value: bf.Bytes()})
	viewBytes, err = GetUint32Bytes(view)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "getUint32Bytes, get viewBytes error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), viewBytes), &cstates.StorageItem{Value: bf.Bytes()})
	indexBytes, err := GetUint32Bytes(maxId + 1)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "getUint32Bytes, get indexBytes error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(CANDIDITE_INDEX)), &cstates.StorageItem{Value: indexBytes})

	//init governance view
	governanceView := &GovernanceView{
		View:   view,
		Height: native.Height,
		TxHash: native.Tx.Hash(),
	}
	bf = new(bytes.Buffer)
	if err := governanceView.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize governanceView error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(GOVERNANCE_VIEW)), &cstates.StorageItem{Value: bf.Bytes()})

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
	bf = new(bytes.Buffer)
	if err := config.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize config error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VBFT_CONFIG)), &cstates.StorageItem{Value: bf.Bytes()})

	//posTable, chainPeers, err := calDposTable(native, config, peers)

	return utils.BYTE_TRUE, nil
}

func RegisterCandidate(native *native.NativeService) ([]byte, error) {
	params := new(RegisterCandidateParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}
	address := params.Address

	//check witness
	err := utils.ValidateOwner(native, address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "validateOwner, checkWitness error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getView, get view error!")
	}

	//get globalParam
	globalParam, err := GetGlobalParam(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getGlobalParam, getGlobalParam error!")
	}

	//check initPos
	if params.InitPos < globalParam.MinInitStake {
		return utils.BYTE_FALSE, errors.NewErr(fmt.Sprintf("registerCandidate, initPos must >= %v!", globalParam.MinInitStake))
	}

	//check peerPubkey
	if err := validatePeerPubKeyFormat(params.PeerPubkey); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "invalid peer pubkey")
	}

	peerPubkeyPrefix, err := hex.DecodeString(params.PeerPubkey)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	//get black list
	blackList, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(BLACK_LIST), peerPubkeyPrefix))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Get, get BlackList error!")
	}
	if blackList != nil {
		return utils.BYTE_FALSE, errors.NewErr("registerCandidate, this Peer is in BlackList!")
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}

	//check if exist in PeerPool
	_, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if ok {
		return utils.BYTE_FALSE, errors.NewErr("registerCandidate, peerPubkey is already in peerPoolMap!")
	}

	peerPoolItem := &PeerPoolItem{
		PeerPubkey: params.PeerPubkey,
		Address:    address,
		InitPos:    uint64(params.InitPos),
		Status:     RegisterCandidateStatus,
	}
	peerPoolMap.PeerPoolMap[params.PeerPubkey] = peerPoolItem
	bf := new(bytes.Buffer)
	if err := peerPoolMap.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize peerPoolMap error!")
	}
	viewBytes, err := GetUint32Bytes(view)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "getUint32Bytes, get viewBytes error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), viewBytes), &cstates.StorageItem{Value: bf.Bytes()})

	//update total stake
	totalStakeBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(TOTAL_STAKE), address[:]))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Get, get indexBytes error!")
	}
	totalStake := &TotalStake{
		Address: address,
	}
	if totalStakeBytes != nil {
		if err := totalStake.Deserialize(bytes.NewBuffer(totalStakeBytes.(*cstates.StorageItem).Value)); err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize totalStake error!")
		}
	}
	totalStake.Stake = totalStake.Stake + uint64(params.InitPos)
	bf = new(bytes.Buffer)
	if err := totalStake.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize totalStake error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(TOTAL_STAKE), address[:]), &cstates.StorageItem{Value: bf.Bytes()})

	//ont transfer
	err = AppCallTransferOnt(native, address, utils.GovernanceContractAddress, uint64(params.InitPos))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOnt, ont transfer error!")
	}
	//ong transfer
	err = AppCallTransferOng(native, address, utils.GovernanceContractAddress, uint64(globalParam.CandidateFee))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOng, ong transfer error!")
	}

	return utils.BYTE_TRUE, nil
}

func ApproveCandidate(native *native.NativeService) ([]byte, error) {
	params := new(ApproveCandidateParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}

	// get admin from database
	adminAddress := new(common.Address)
	admin, err := global_params.GetStorageAdmin(native, global_params.GetAdminKey(utils.ParamContractAddress, false))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getStorageAdmin, get admin error!")
	}
	copy(adminAddress[:], admin[:])
	err = utils.ValidateOwner(native, *adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "approveCandidate, checkWitness error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getView, get view error!")
	}

	//get globalParam
	globalParam, err := GetGlobalParam(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getGlobalParam, getGlobalParam error!")
	}

	//check if peerPoolMap full
	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}

	num := 0
	for _, peerPoolItem := range peerPoolMap.PeerPoolMap {
		if peerPoolItem.Status == CandidateStatus || peerPoolItem.Status == ConsensusStatus {
			num = num + 1
		}
	}
	if num >= int(globalParam.CandidateNum) {
		return utils.BYTE_FALSE, errors.NewErr("approveCandidate, num of candidate node is full!")
	}

	//get peerPool
	peerPoolItem, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, errors.NewErr("approveCandidate, peerPubkey is not in peerPoolMap!")
	}

	if peerPoolItem.Status != RegisterCandidateStatus {
		return utils.BYTE_FALSE, errors.NewErr("approveCandidate, peer status is not RegisterCandidateStatus!")
	}

	peerPoolItem.Status = CandidateStatus
	peerPoolItem.TotalPos = 0

	//check if has index
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	indexBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_INDEX), peerPubkeyPrefix))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Get, get indexBytes error!")
	}
	if indexBytes != nil {
		index, err := GetBytesUint32(indexBytes.(*cstates.StorageItem).Value)
		if err != nil {
			return nil, errors.NewDetailErr(err, errors.ErrNoCode, "GetBytesUint32, get index error!")
		}
		peerPoolItem.Index = index
	} else {
		//get index
		candidateIndexBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(CANDIDITE_INDEX)))
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Get, get candidateIndex error!")
		}
		var candidateIndex uint32
		if candidateIndexBytes == nil {
			return utils.BYTE_FALSE, errors.NewErr("approveCandidate, candidateIndex is not init!")
		} else {
			candidateIndexStore, _ := candidateIndexBytes.(*cstates.StorageItem)
			index, err := GetBytesUint32(candidateIndexStore.Value)
			if err != nil {
				return nil, errors.NewDetailErr(err, errors.ErrNoCode, "GetBytesUint32, get index error!")
			}
			candidateIndex = index
		}
		peerPoolItem.Index = uint32(candidateIndex)

		//update candidateIndex
		newCandidateIndex := candidateIndex + 1
		newCandidateIndexBytes, err := GetUint32Bytes(newCandidateIndex)
		if err != nil {
			return nil, errors.NewDetailErr(err, errors.ErrNoCode, "GetUint32Bytes, get newCandidateIndexBytes error!")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(CANDIDITE_INDEX)),
			&cstates.StorageItem{Value: newCandidateIndexBytes})

		indexBytes, err := GetUint32Bytes(peerPoolItem.Index)
		if err != nil {
			return nil, errors.NewDetailErr(err, errors.ErrNoCode, "GetUint32Bytes, get indexBytes error!")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_INDEX), peerPubkeyPrefix), &cstates.StorageItem{Value: indexBytes})
	}
	peerPoolMap.PeerPoolMap[params.PeerPubkey] = peerPoolItem
	bf := new(bytes.Buffer)
	if err := peerPoolMap.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize peerPoolMap error!")
	}
	viewBytes, err := GetUint32Bytes(view)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "GetUint32Bytes, get viewBytes error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), viewBytes), &cstates.StorageItem{Value: bf.Bytes()})

	return utils.BYTE_TRUE, nil
}

func RejectCandidate(native *native.NativeService) ([]byte, error) {
	params := new(RejectCandidateParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}

	// get admin from database
	adminAddress := new(common.Address)
	admin, err := global_params.GetStorageAdmin(native, global_params.GetAdminKey(utils.ParamContractAddress, false))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getStorageAdmin, get admin error!")
	}
	copy(adminAddress[:], admin[:])
	err = utils.ValidateOwner(native, *adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "approveCandidate, checkWitness error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getView, get view error!")
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}

	//draw back init pos
	peerPubkeyPrefix, err := hex.DecodeString(params.PeerPubkey)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	peerPoolItem, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, errors.NewErr("rejectCandidate, peerPubkey is not in peerPoolMap!")
	}
	address := peerPoolItem.Address
	voteInfoBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL),
		peerPubkeyPrefix, address[:]))
	voteInfo := &VoteInfo{
		PeerPubkey: params.PeerPubkey,
		Address:    address,
	}
	if voteInfoBytes != nil {
		voteInfoStore, _ := voteInfoBytes.(*cstates.StorageItem)
		if err := voteInfo.Deserialize(bytes.NewBuffer(voteInfoStore.Value)); err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfo error!")
		}
	}
	voteInfo.WithdrawUnfreezePos = voteInfo.WithdrawUnfreezePos + peerPoolItem.InitPos
	bf := new(bytes.Buffer)
	if err := voteInfo.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize voteInfo error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix,
		address[:]), &cstates.StorageItem{Value: bf.Bytes()})

	//remove peerPubkey from peerPool
	delete(peerPoolMap.PeerPoolMap, params.PeerPubkey)
	bf = new(bytes.Buffer)
	if err := peerPoolMap.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize peerPoolMap error!")
	}
	viewBytes, err := GetUint32Bytes(view)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "GetUint32Bytes, get viewBytes error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), viewBytes), &cstates.StorageItem{Value: bf.Bytes()})

	return utils.BYTE_TRUE, nil
}

func BlackNode(native *native.NativeService) ([]byte, error) {
	params := new(BlackNodeParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}

	// get admin from database
	adminAddress := new(common.Address)
	admin, err := global_params.GetStorageAdmin(native, global_params.GetAdminKey(utils.ParamContractAddress, false))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getStorageAdmin, get admin error!")
	}
	copy(adminAddress[:], admin[:])
	err = utils.ValidateOwner(native, *adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "blackNode, checkWitness error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	peerPubkeyPrefix, err := hex.DecodeString(params.PeerPubkey)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getView, get view error!")
	}
	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}

	peerPoolItem, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, errors.NewErr("blackNode, peerPubkey is not in peerPoolMap!")
	}

	blackListItem := &BlackListItem{
		PeerPubkey: peerPoolItem.PeerPubkey,
		Address:    peerPoolItem.Address,
		InitPos:    peerPoolItem.InitPos,
	}
	bf := new(bytes.Buffer)
	if err := blackListItem.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize blackListItem error!")
	}
	//put peer into black list
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(BLACK_LIST), peerPubkeyPrefix), &cstates.StorageItem{Value: bf.Bytes()})
	//change peerPool status
	if peerPoolItem.Status == ConsensusStatus {
		peerPoolItem.Status = BlackConsensusStatus
		peerPoolMap.PeerPoolMap[params.PeerPubkey] = peerPoolItem
		bf = new(bytes.Buffer)
		if err := peerPoolMap.Serialize(bf); err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize peerPoolMap error!")
		}
		viewBytes, err := GetUint32Bytes(view)
		if err != nil {
			return nil, errors.NewDetailErr(err, errors.ErrNoCode, "GetUint32Bytes, get viewBytes error!")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), viewBytes), &cstates.StorageItem{Value: bf.Bytes()})

		//commitDpos
		// get config
		config := new(Configuration)
		configBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VBFT_CONFIG)))
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Get, get configBytes error!")
		}
		if configBytes == nil {
			return utils.BYTE_FALSE, errors.NewErr("blackNode, configBytes is nil!")
		}
		configStore, _ := configBytes.(*cstates.StorageItem)
		if err := config.Deserialize(bytes.NewBuffer(configStore.Value)); err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize config error!")
		}
		err = executeCommitDpos(native, contract, config)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "executeCommitDpos, executeCommitDpos error!")
		}
	} else {
		peerPoolItem.Status = BlackStatus
		peerPoolMap.PeerPoolMap[params.PeerPubkey] = peerPoolItem
		bf = new(bytes.Buffer)
		if err := peerPoolMap.Serialize(bf); err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize peerPoolMap error!")
		}
		viewBytes, err := GetUint32Bytes(view)
		if err != nil {
			return nil, errors.NewDetailErr(err, errors.ErrNoCode, "GetUint32Bytes, get viewBytes error!")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), viewBytes), &cstates.StorageItem{Value: bf.Bytes()})
	}

	return utils.BYTE_TRUE, nil
}

func WhiteNode(native *native.NativeService) ([]byte, error) {
	params := new(WhiteNodeParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}

	// get admin from database
	adminAddress := new(common.Address)
	admin, err := global_params.GetStorageAdmin(native, global_params.GetAdminKey(utils.ParamContractAddress, false))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getStorageAdmin, get admin error!")
	}
	copy(adminAddress[:], admin[:])
	err = utils.ValidateOwner(native, *adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "whiteNode, checkWitness error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	peerPubkeyPrefix, err := hex.DecodeString(params.PeerPubkey)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}

	//get black list
	blackList := new(BlackListItem)
	blackListBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(BLACK_LIST), peerPubkeyPrefix))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Get, get BlackList error!")
	}
	if blackListBytes == nil {
		return utils.BYTE_FALSE, errors.NewErr("whiteNode, this Peer is not in BlackList!")
	}
	blackListStore, _ := blackListBytes.(*cstates.StorageItem)
	if err := blackList.Deserialize(bytes.NewBuffer(blackListStore.Value)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize blackList error!")
	}

	//draw back init pos
	address := blackList.Address
	voteInfoBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL),
		peerPubkeyPrefix, address[:]))
	voteInfo := &VoteInfo{
		PeerPubkey: blackList.PeerPubkey,
		Address:    address,
	}
	if voteInfoBytes != nil {
		voteInfoStore, _ := voteInfoBytes.(*cstates.StorageItem)
		if err := voteInfo.Deserialize(bytes.NewBuffer(voteInfoStore.Value)); err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfo error!")
		}
	}
	voteInfo.WithdrawUnfreezePos = voteInfo.WithdrawUnfreezePos + blackList.InitPos
	bf := new(bytes.Buffer)
	if err := voteInfo.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize voteInfo error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix,
		address[:]), &cstates.StorageItem{Value: bf.Bytes()})

	//remove peer from black list
	native.CloneCache.Delete(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(BLACK_LIST), peerPubkeyPrefix))

	return utils.BYTE_TRUE, nil
}

func QuitNode(native *native.NativeService) ([]byte, error) {
	params := new(QuitNodeParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}
	address := params.Address

	//check witness
	err := utils.ValidateOwner(native, address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "validateOwner, checkWitness error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getView, get view error!")
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}

	peerPoolItem, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, errors.NewErr("quitNode, peerPubkey is not in peerPoolMap!")
	}

	if address != peerPoolItem.Address {
		return utils.BYTE_FALSE, errors.NewErr("quitNode, peerPubkey is not registered by this address!")
	}

	//change peerPool status
	if peerPoolItem.Status == ConsensusStatus {
		peerPoolItem.Status = QuitConsensusStatus
	} else {
		peerPoolItem.Status = QuitingStatus
	}

	peerPoolMap.PeerPoolMap[params.PeerPubkey] = peerPoolItem
	bf := new(bytes.Buffer)
	if err := peerPoolMap.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize peerPoolMap error!")
	}
	viewBytes, err := GetUint32Bytes(view)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "GetUint32Bytes, get viewBytes error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), viewBytes), &cstates.StorageItem{Value: bf.Bytes()})

	return utils.BYTE_TRUE, nil
}

func VoteForPeer(native *native.NativeService) ([]byte, error) {
	params := &VoteForPeerParam{
		PeerPubkeyList: make([]string, 0),
		PosList:        make([]uint32, 0),
	}
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}
	address := params.Address

	//check witness
	err := utils.ValidateOwner(native, address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "validateOwner, checkWitness error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getView, get view error!")
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}

	var total uint64
	for i := 0; i < len(params.PeerPubkeyList); i++ {
		peerPubkey := params.PeerPubkeyList[i]
		pos := uint64(params.PosList[i])
		peerPubkeyPrefix, err := hex.DecodeString(peerPubkey)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
		}

		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peerPubkey]
		if !ok {
			return utils.BYTE_FALSE, errors.NewErr("voteForPeer, peerPubkey is not in peerPoolMap!")
		}

		if peerPoolItem.Status != CandidateStatus && peerPoolItem.Status != ConsensusStatus {
			return utils.BYTE_FALSE, errors.NewErr("voteForPeer, peerPubkey is not candidate and can not be voted!")
		}

		voteInfoBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL),
			peerPubkeyPrefix, address[:]))
		voteInfo := &VoteInfo{
			PeerPubkey: peerPubkey,
			Address:    address,
		}
		if voteInfoBytes != nil {
			voteInfoStore, _ := voteInfoBytes.(*cstates.StorageItem)
			if err := voteInfo.Deserialize(bytes.NewBuffer(voteInfoStore.Value)); err != nil {
				return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfo error!")
			}
			voteInfo.NewPos = voteInfo.NewPos + pos
		} else {
			voteInfo.NewPos = pos
		}
		total = total + pos
		peerPoolItem.TotalPos = peerPoolItem.TotalPos + pos

		peerPoolMap.PeerPoolMap[peerPubkey] = peerPoolItem
		bf := new(bytes.Buffer)
		if err := voteInfo.Serialize(bf); err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize voteInfo error!")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix,
			address[:]), &cstates.StorageItem{Value: bf.Bytes()})
	}
	bf := new(bytes.Buffer)
	if err := peerPoolMap.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize peerPoolMap error!")
	}
	viewBytes, err := GetUint32Bytes(view)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "GetUint32Bytes, get viewBytes error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), viewBytes), &cstates.StorageItem{Value: bf.Bytes()})

	//update total stake
	totalStakeBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(TOTAL_STAKE), address[:]))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Get, get indexBytes error!")
	}
	totalStake := &TotalStake{
		Address: address,
	}
	if totalStakeBytes != nil {
		if err := totalStake.Deserialize(bytes.NewBuffer(totalStakeBytes.(*cstates.StorageItem).Value)); err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize totalStake error!")
		}
	}
	totalStake.Stake = totalStake.Stake + total
	bf = new(bytes.Buffer)
	if err := totalStake.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize totalStake error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(TOTAL_STAKE), address[:]), &cstates.StorageItem{Value: bf.Bytes()})

	//ont transfer
	err = AppCallTransferOnt(native, address, utils.GovernanceContractAddress, total)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOnt, ont transfer error!")
	}

	return utils.BYTE_TRUE, nil
}

func UnVoteForPeer(native *native.NativeService) ([]byte, error) {
	params := &VoteForPeerParam{
		PeerPubkeyList: make([]string, 0),
		PosList:        make([]uint32, 0),
	}
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}
	address := params.Address

	//check witness
	err := utils.ValidateOwner(native, address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "validateOwner, checkWitness error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getView, get view error!")
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}

	for i := 0; i < len(params.PeerPubkeyList); i++ {
		peerPubkey := params.PeerPubkeyList[i]
		pos := uint64(params.PosList[i])
		peerPubkeyPrefix, err := hex.DecodeString(peerPubkey)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
		}

		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peerPubkey]
		if !ok {
			return utils.BYTE_FALSE, errors.NewErr("voteForPeer, peerPubkey is not in peerPoolMap!")
		}

		if peerPoolItem.Status != CandidateStatus && peerPoolItem.Status != ConsensusStatus {
			return utils.BYTE_FALSE, errors.NewErr("voteForPeer, peerPubkey is not candidate and can not be voted!")
		}

		voteInfoBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL),
			peerPubkeyPrefix, address[:]))
		voteInfo := &VoteInfo{
			PeerPubkey: peerPubkey,
			Address:    address,
		}
		if voteInfoBytes != nil {
			voteInfoStore, _ := voteInfoBytes.(*cstates.StorageItem)
			if err := voteInfo.Deserialize(bytes.NewBuffer(voteInfoStore.Value)); err != nil {
				return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfo error!")
			}
			if voteInfo.NewPos < pos {
				if peerPoolItem.Status == ConsensusStatus {
					if voteInfo.ConsensusPos < (pos - voteInfo.NewPos) {
						return utils.BYTE_FALSE, errors.NewErr("voteForPeer, your pos of this peerPubkey is not enough!")
					}
					consensusPos := voteInfo.ConsensusPos + voteInfo.NewPos - pos
					newPos := voteInfo.NewPos
					voteInfo.NewPos = 0
					voteInfo.WithdrawUnfreezePos = voteInfo.WithdrawUnfreezePos + newPos
					voteInfo.ConsensusPos = consensusPos
					voteInfo.WithdrawPos = voteInfo.WithdrawPos + pos - voteInfo.NewPos
					peerPoolItem.TotalPos = peerPoolItem.TotalPos - pos
				}
				if peerPoolItem.Status == CandidateStatus {
					if voteInfo.FreezePos < (pos - voteInfo.NewPos) {
						return utils.BYTE_FALSE, errors.NewErr("voteForPeer, your pos of this peerPubkey is not enough!")
					}
					freezePos := voteInfo.FreezePos + voteInfo.NewPos - pos
					newPos := voteInfo.NewPos
					voteInfo.NewPos = 0
					voteInfo.WithdrawUnfreezePos = voteInfo.WithdrawUnfreezePos + newPos
					voteInfo.FreezePos = uint64(freezePos)
					voteInfo.WithdrawFreezePos = voteInfo.WithdrawFreezePos + pos - voteInfo.NewPos
					peerPoolItem.TotalPos = peerPoolItem.TotalPos - pos
				}
			} else {
				temp := voteInfo.NewPos - pos
				voteInfo.NewPos = temp
				voteInfo.WithdrawUnfreezePos = voteInfo.WithdrawUnfreezePos + pos
				peerPoolItem.TotalPos = peerPoolItem.TotalPos - pos
			}
		} else {
			return utils.BYTE_FALSE, errors.NewErr("voteForPeer, your pos of this peerPubkey is not enough!")
		}

		peerPoolMap.PeerPoolMap[peerPubkey] = peerPoolItem
		bf := new(bytes.Buffer)
		if err := voteInfo.Serialize(bf); err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize voteInfo error!")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix,
			address[:]), &cstates.StorageItem{Value: bf.Bytes()})
	}
	bf := new(bytes.Buffer)
	if err := peerPoolMap.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize peerPoolMap error!")
	}
	viewBytes, err := GetUint32Bytes(view)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "GetUint32Bytes, get viewBytes error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), viewBytes), &cstates.StorageItem{Value: bf.Bytes()})

	return utils.BYTE_TRUE, nil
}

func Withdraw(native *native.NativeService) ([]byte, error) {
	params := &WithdrawParam{
		PeerPubkeyList: make([]string, 0),
		WithdrawList:   make([]uint32, 0),
	}
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}
	address := params.Address

	//check witness
	err := utils.ValidateOwner(native, address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "validateOwner, checkWitness error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getView, get view error!")
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}

	var total uint64
	for i := 0; i < len(params.PeerPubkeyList); i++ {
		peerPubkey := params.PeerPubkeyList[i]
		pos := params.WithdrawList[i]
		peerPubkeyPrefix, err := hex.DecodeString(peerPubkey)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
		}

		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peerPubkey]
		if !ok {
			return utils.BYTE_FALSE, errors.NewErr("withdraw, peerPubkey is not in peerPoolMap!")
		}

		if peerPoolItem.Status != CandidateStatus && peerPoolItem.Status != ConsensusStatus {
			return utils.BYTE_FALSE, errors.NewErr("withdraw, peerPubkey is not candidate and can not withdraw!")
		}

		voteInfoBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL),
			peerPubkeyPrefix, address[:]))
		voteInfo := new(VoteInfo)
		if voteInfoBytes != nil {
			voteInfoStore, _ := voteInfoBytes.(*cstates.StorageItem)
			if err := voteInfo.Deserialize(bytes.NewBuffer(voteInfoStore.Value)); err != nil {
				return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfo error!")
			}
			if voteInfo.WithdrawUnfreezePos < uint64(pos) {
				return utils.BYTE_FALSE, errors.NewErr("withdraw, your unfreeze withdraw pos of this peerPubkey is not enough!")
			} else {
				voteInfo.WithdrawUnfreezePos = voteInfo.WithdrawUnfreezePos - uint64(pos)
				total = total + uint64(pos)
				bf := new(bytes.Buffer)
				if err := voteInfo.Serialize(bf); err != nil {
					return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize voteInfo error!")
				}
				native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix,
					address[:]), &cstates.StorageItem{Value: bf.Bytes()})
			}
		} else {
			return utils.BYTE_FALSE, errors.NewErr("withdraw, your withdraw pos of this peerPubkey is not enough!")
		}
		if voteInfo.ConsensusPos == 0 && voteInfo.FreezePos == 0 && voteInfo.NewPos == 0 &&
			voteInfo.WithdrawPos == 0 && voteInfo.WithdrawFreezePos == 0 && voteInfo.WithdrawUnfreezePos == 0 {
			native.CloneCache.Delete(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix, address[:]))
		}
	}

	//update total stake
	totalStakeBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(TOTAL_STAKE), address[:]))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Get, get indexBytes error!")
	}
	totalStake := &TotalStake{
		Address: address,
	}
	if totalStakeBytes != nil {
		if err := totalStake.Deserialize(bytes.NewBuffer(totalStakeBytes.(*cstates.StorageItem).Value)); err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize totalStake error!")
		}
	}
	if totalStake.Stake < total {
		return utils.BYTE_FALSE, errors.NewErr("withdraw, ont deposit is not enough!")
	}
	totalStake.Stake = totalStake.Stake - total
	bf := new(bytes.Buffer)
	if err := totalStake.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize totalStake error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(TOTAL_STAKE), address[:]), &cstates.StorageItem{Value: bf.Bytes()})

	//ont transfer
	err = AppCallTransferOnt(native, utils.GovernanceContractAddress, address, total)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOnt, ont transfer error!")
	}

	return utils.BYTE_TRUE, nil
}

func CommitDpos(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	// get config
	config := new(Configuration)
	configBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VBFT_CONFIG)))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Get, get configBytes error!")
	}
	if configBytes == nil {
		return utils.BYTE_FALSE, errors.NewErr("commitDpos, configBytes is nil!")
	}
	configStore, _ := configBytes.(*cstates.StorageItem)
	if err := config.Deserialize(bytes.NewBuffer(configStore.Value)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize config error!")
	}

	//get governace view
	governanceView, err := GetGovernanceView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getGovernanceView, get GovernanceView error!")
	}

	// get admin from database
	adminAddress := new(common.Address)
	admin, err := global_params.GetStorageAdmin(native, global_params.GetAdminKey(utils.ParamContractAddress, false))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getStorageAdmin, get admin error!")
	}
	copy(adminAddress[:], admin[:])
	err = utils.ValidateOwner(native, *adminAddress)
	if err != nil {
		cycle := (native.Height - governanceView.Height) >= config.MaxBlockChangeView
		if !cycle {
			return utils.BYTE_FALSE, errors.NewErr("commitDpos, authentication Failed!")
		}
	}

	err = executeCommitDpos(native, contract, config)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "executeCommitDpos, executeCommitDpos error!")
	}

	return utils.BYTE_TRUE, nil
}

func executeCommitDpos(native *native.NativeService, contract common.Address, config *Configuration) error {
	//get governace view
	governanceView, err := GetGovernanceView(native, contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getGovernanceView, get GovernanceView error!")
	}

	//get current view
	view := governanceView.View
	newView := view + 1

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}

	//feeSplit first
	err = executeSplit(native, contract, peerPoolMap)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, executeSplit error!")
	}

	peers := []*PeerStakeInfo{}
	for _, peerPoolItem := range peerPoolMap.PeerPoolMap {
		if peerPoolItem.Status == QuitingStatus {
			err = normalQuit(native, contract, peerPoolItem)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "normalQuit, normalQuit error!")
			}
			delete(peerPoolMap.PeerPoolMap, peerPoolItem.PeerPubkey)
		}
		if peerPoolItem.Status == BlackStatus {
			err = blackQuit(native, contract, peerPoolItem)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "blackQuit, blackQuit error!")
			}
			delete(peerPoolMap.PeerPoolMap, peerPoolItem.PeerPubkey)
		}
		if peerPoolItem.Status == QuitConsensusStatus {
			peerPoolItem.Status = QuitingStatus
			peerPoolMap.PeerPoolMap[peerPoolItem.PeerPubkey] = peerPoolItem
		}
		if peerPoolItem.Status == BlackConsensusStatus {
			peerPoolItem.Status = BlackStatus
			peerPoolMap.PeerPoolMap[peerPoolItem.PeerPubkey] = peerPoolItem
		}

		if peerPoolItem.Status == CandidateStatus || peerPoolItem.Status == ConsensusStatus {
			stake := peerPoolItem.TotalPos + peerPoolItem.InitPos
			peers = append(peers, &PeerStakeInfo{
				Index:      peerPoolItem.Index,
				PeerPubkey: peerPoolItem.PeerPubkey,
				Stake:      stake,
			})
		}
	}

	// sort peers by stake
	sort.Slice(peers, func(i, j int) bool {
		return peers[i].Stake > peers[j].Stake
	})

	// consensus peers
	for i := 0; i < int(config.K); i++ {
		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peers[i].PeerPubkey]
		if !ok {
			return errors.NewErr("voteForPeer, peerPubkey is not in peerPoolMap!")
		}

		if peerPoolItem.Status == ConsensusStatus {
			err = consensusToConsensus(native, contract, peerPoolItem)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "consensusToConsensus, consensusToConsensus error!")
			}
		} else {
			err = unConsensusToConsensus(native, contract, peerPoolItem)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "unConsensusToConsensus, unConsensusToConsensus error!")
			}
		}
		peerPoolItem.Status = ConsensusStatus
		peerPoolMap.PeerPoolMap[peers[i].PeerPubkey] = peerPoolItem
	}

	//non consensus peers
	for i := int(config.K); i < len(peers); i++ {
		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peers[i].PeerPubkey]
		if !ok {
			return errors.NewErr("voteForPeer, peerPubkey is not in peerPoolMap!")
		}

		if peerPoolItem.Status == ConsensusStatus {
			err = consensusToUnConsensus(native, contract, peerPoolItem)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "consensusToUnConsensus, consensusToUnConsensus error!")
			}
		} else {
			err = unConsensusToUnConsensus(native, contract, peerPoolItem)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "unConsensusToUnConsensus, unConsensusToUnConsensus error!")
			}
		}
		peerPoolItem.Status = CandidateStatus
		peerPoolMap.PeerPoolMap[peers[i].PeerPubkey] = peerPoolItem
	}
	bf := new(bytes.Buffer)
	if err := peerPoolMap.Serialize(bf); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize peerPoolMap error!")
	}
	newViewBytes, err := GetUint32Bytes(newView)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "GetUint32Bytes, get newViewBytes error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), newViewBytes), &cstates.StorageItem{Value: bf.Bytes()})
	oldView := view - 1
	oldViewBytes, err := GetUint32Bytes(oldView)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "GetUint32Bytes, get oldViewBytes error!")
	}
	native.CloneCache.Delete(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), oldViewBytes))

	//update view
	governanceView = &GovernanceView{
		View:   governanceView.View + 1,
		Height: native.Height,
		TxHash: native.Tx.Hash(),
	}
	bf = new(bytes.Buffer)
	if err := governanceView.Serialize(bf); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize governanceView error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(GOVERNANCE_VIEW)), &cstates.StorageItem{Value: bf.Bytes()})

	//posTable, chainPeers, err := calDposTable(native, config, peers)

	return nil
}

func UpdateConfig(native *native.NativeService) ([]byte, error) {
	// get admin from database
	adminAddress := new(common.Address)
	admin, err := global_params.GetStorageAdmin(native, global_params.GetAdminKey(utils.ParamContractAddress, false))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getStorageAdmin, get admin error!")
	}
	copy(adminAddress[:], admin[:])
	err = utils.ValidateOwner(native, *adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "updateConfig, checkWitness error!")
	}

	configuration := new(Configuration)
	if err := configuration.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize configuration error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getView, get view error!")
	}
	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}
	candidateNum := 0
	for _, peerPoolItem := range peerPoolMap.PeerPoolMap {
		if peerPoolItem.Status == CandidateStatus || peerPoolItem.Status == ConsensusStatus {
			candidateNum = candidateNum + 1
		}
	}

	//check the configuration
	if configuration.C == 0 {
		return utils.BYTE_FALSE, errors.NewErr("updateConfig. C can not be 0 in config!")
	}
	if int(configuration.K) > candidateNum {
		return utils.BYTE_FALSE, errors.NewErr("updateConfig. K can not be larger than num of candidate peer in config!")
	}
	if configuration.L < 16*configuration.K || configuration.L%configuration.K != 0 {
		return utils.BYTE_FALSE, errors.NewErr("updateConfig. L can not be less than 16*K and K must be times of L in config!")
	}
	if configuration.K < 2*configuration.C+1 {
		return utils.BYTE_FALSE, errors.NewErr("updateConfig. K can not be less than 2*C+1 in config!")
	}
	if configuration.N < configuration.K || configuration.K < 7 {
		return utils.BYTE_FALSE, errors.NewErr("updateConfig. config not match N >= K >= 7!")
	}

	bf := new(bytes.Buffer)
	if err := configuration.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize configuration error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VBFT_CONFIG)), &cstates.StorageItem{Value: bf.Bytes()})

	return utils.BYTE_TRUE, nil
}

func UpdateGlobalParam(native *native.NativeService) ([]byte, error) {
	// get admin from database
	adminAddress := new(common.Address)
	admin, err := global_params.GetStorageAdmin(native, global_params.GetAdminKey(utils.ParamContractAddress, false))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getStorageAdmin, get admin error!")
	}
	copy(adminAddress[:], admin[:])
	err = utils.ValidateOwner(native, *adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "updateGlobalParam, checkWitness error!")
	}

	globalParam := new(GlobalParam)
	if err := globalParam.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize globalParam error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//check the globalParam
	if (globalParam.A + globalParam.B) != 100 {
		return utils.BYTE_FALSE, errors.NewErr("updateGlobalParam. A + B must equal to 100!")
	}
	if globalParam.Yita == 0 {
		return utils.BYTE_FALSE, errors.NewErr("updateGlobalParam. Yita must > 0!")
	}

	bf := new(bytes.Buffer)
	if err := globalParam.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize globalParam error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(GLOBAL_PARAM)), &cstates.StorageItem{Value: bf.Bytes()})

	return utils.BYTE_TRUE, nil
}

func CallSplit(native *native.NativeService) ([]byte, error) {
	// get admin from database
	adminAddress := new(common.Address)
	admin, err := global_params.GetStorageAdmin(native, global_params.GetAdminKey(utils.ParamContractAddress, false))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getStorageAdmin, get admin error!")
	}
	copy(adminAddress[:], admin[:])
	err = utils.ValidateOwner(native, *adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "callSplit, checkWitness error!")
	}

	contract := utils.GovernanceContractAddress
	//get current view
	cView, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "callSplit, get view error!")
	}
	view := cView - 1

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "callSplit, get peerPoolMap error!")
	}

	err = executeSplit(native, contract, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, executeSplitp error!")
	}

	return utils.BYTE_TRUE, nil
}

func executeSplit(native *native.NativeService, contract common.Address, peerPoolMap *PeerPoolMap) error {
	balance, err := GetOngBalance(native, utils.GovernanceContractAddress)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, getOngBalance error!")
	}
	//get globalParam
	globalParam, err := GetGlobalParam(native, contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getGlobalParam, getGlobalParam error!")
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

	// get config
	config := new(Configuration)
	configBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VBFT_CONFIG)))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, get configBytes error!")
	}
	if configBytes == nil {
		return errors.NewErr("executeSplit, configBytes is nil!")
	}
	configStore, _ := configBytes.(*cstates.StorageItem)
	if err := config.Deserialize(bytes.NewBuffer(configStore.Value)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize config error!")
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
		peersCandidate[i].S = splitCurve(peersCandidate[i].Stake, avg, globalParam.Yita)
		sumS += peersCandidate[i].S
	}

	//fee split of consensus peer
	var splitAmount uint64
	for i := int(config.K) - 1; i >= 0; i-- {
		nodeAmount := balance * uint64(globalParam.A) / 100 * peersCandidate[i].S / sumS
		address := peersCandidate[i].Address
		err = AppCallApproveOng(native, utils.GovernanceContractAddress, address, nodeAmount)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, ong transfer error!")
		}
		splitAmount += nodeAmount
	}

	//fee split of candidate peer
	// cal s of each candidate node
	sum = 0
	for i := int(config.K); i < len(peersCandidate); i++ {
		sum += peersCandidate[i].Stake
	}
	splitAmount = 0
	for i := int(config.K); i < len(peersCandidate); i++ {
		nodeAmount := balance * uint64(globalParam.B) / 100 * peersCandidate[i].Stake / sum
		address := peersCandidate[i].Address
		err = AppCallApproveOng(native, utils.GovernanceContractAddress, address, nodeAmount)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, ong transfer error!")
		}
		splitAmount += nodeAmount
	}

	return nil
}

func normalQuit(native *native.NativeService, contract common.Address, peerPoolItem *PeerPoolItem) error {
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	flag := false
	//draw back vote pos
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
	}
	voteInfo := new(VoteInfo)
	for _, v := range stateValues {
		voteInfoStore, _ := v.Value.(*cstates.StorageItem)
		if err := voteInfo.Deserialize(bytes.NewBuffer(voteInfoStore.Value)); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfo error!")
		}
		voteInfo.WithdrawUnfreezePos = voteInfo.ConsensusPos + voteInfo.FreezePos + voteInfo.NewPos + voteInfo.WithdrawPos +
			voteInfo.WithdrawFreezePos + voteInfo.WithdrawUnfreezePos
		voteInfo.ConsensusPos = 0
		voteInfo.FreezePos = 0
		voteInfo.NewPos = 0
		voteInfo.WithdrawPos = 0
		voteInfo.WithdrawFreezePos = 0
		if voteInfo.Address == peerPoolItem.Address {
			flag = true
			voteInfo.WithdrawUnfreezePos = voteInfo.WithdrawUnfreezePos + peerPoolItem.InitPos
		}
		bf := new(bytes.Buffer)
		if err := voteInfo.Serialize(bf); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize voteInfo error!")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix,
			voteInfo.Address[:]), &cstates.StorageItem{Value: bf.Bytes()})
	}
	if flag == false {
		voteInfo := &VoteInfo{
			PeerPubkey:          peerPoolItem.PeerPubkey,
			Address:             peerPoolItem.Address,
			WithdrawUnfreezePos: peerPoolItem.InitPos,
		}
		bf := new(bytes.Buffer)
		if err := voteInfo.Serialize(bf); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize voteInfo error!")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix,
			peerPoolItem.Address[:]), &cstates.StorageItem{Value: bf.Bytes()})
	}
	return nil
}

func blackQuit(native *native.NativeService, contract common.Address, peerPoolItem *PeerPoolItem) error {
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	//draw back vote pos
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
	}
	voteInfo := new(VoteInfo)
	for _, v := range stateValues {
		voteInfoStore, _ := v.Value.(*cstates.StorageItem)
		if err := voteInfo.Deserialize(bytes.NewBuffer(voteInfoStore.Value)); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfo error!")
		}
		voteInfo.WithdrawUnfreezePos = voteInfo.ConsensusPos + voteInfo.FreezePos + voteInfo.NewPos + voteInfo.WithdrawPos +
			voteInfo.WithdrawFreezePos + voteInfo.WithdrawUnfreezePos
		voteInfo.ConsensusPos = 0
		voteInfo.FreezePos = 0
		voteInfo.NewPos = 0
		voteInfo.WithdrawPos = 0
		voteInfo.WithdrawFreezePos = 0
		address := voteInfo.Address
		bf := new(bytes.Buffer)
		if err := voteInfo.Serialize(bf); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize voteInfo error!")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix,
			address[:]), &cstates.StorageItem{Value: bf.Bytes()})
	}
	return nil
}

func consensusToConsensus(native *native.NativeService, contract common.Address, peerPoolItem *PeerPoolItem) error {
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	//update voteInfoPool
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
	}
	voteInfo := new(VoteInfo)
	for _, v := range stateValues {
		voteInfoStore, _ := v.Value.(*cstates.StorageItem)
		if err := voteInfo.Deserialize(bytes.NewBuffer(voteInfoStore.Value)); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfo error!")
		}
		address := voteInfo.Address
		if voteInfo.FreezePos != 0 {
			return errors.NewErr("commitPos, freezePos should be 0!")
		}
		newPos := voteInfo.NewPos
		voteInfo.ConsensusPos = voteInfo.ConsensusPos + newPos
		voteInfo.NewPos = 0
		withdrawPos := voteInfo.WithdrawPos
		withdrawFreezePos := voteInfo.WithdrawFreezePos
		voteInfo.WithdrawFreezePos = withdrawPos
		voteInfo.WithdrawUnfreezePos = voteInfo.WithdrawUnfreezePos + withdrawFreezePos
		voteInfo.WithdrawPos = 0

		bf := new(bytes.Buffer)
		if err := voteInfo.Serialize(bf); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize voteInfo error!")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix,
			address[:]), &cstates.StorageItem{Value: bf.Bytes()})
	}
	return nil
}

func unConsensusToConsensus(native *native.NativeService, contract common.Address, peerPoolItem *PeerPoolItem) error {
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	//update voteInfoPool
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
	}
	voteInfo := new(VoteInfo)
	for _, v := range stateValues {
		voteInfoStore, _ := v.Value.(*cstates.StorageItem)
		if err := voteInfo.Deserialize(bytes.NewBuffer(voteInfoStore.Value)); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfo error!")
		}
		address := voteInfo.Address
		if voteInfo.ConsensusPos != 0 {
			return errors.NewErr("consensusPos, freezePos should be 0!")
		}

		voteInfo.ConsensusPos = voteInfo.ConsensusPos + voteInfo.FreezePos + voteInfo.NewPos
		voteInfo.NewPos = 0
		voteInfo.FreezePos = 0
		withdrawPos := voteInfo.WithdrawPos
		withdrawFreezePos := voteInfo.WithdrawFreezePos
		voteInfo.WithdrawFreezePos = withdrawPos
		voteInfo.WithdrawUnfreezePos = voteInfo.WithdrawUnfreezePos + withdrawFreezePos
		voteInfo.WithdrawPos = 0

		bf := new(bytes.Buffer)
		if err := voteInfo.Serialize(bf); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize voteInfo error!")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix,
			address[:]), &cstates.StorageItem{Value: bf.Bytes()})
	}
	return nil
}

func consensusToUnConsensus(native *native.NativeService, contract common.Address, peerPoolItem *PeerPoolItem) error {
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	//update voteInfoPool
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
	}
	voteInfo := new(VoteInfo)
	for _, v := range stateValues {
		voteInfoStore, _ := v.Value.(*cstates.StorageItem)
		if err := voteInfo.Deserialize(bytes.NewBuffer(voteInfoStore.Value)); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfo error!")
		}
		address := voteInfo.Address
		if voteInfo.FreezePos != 0 {
			return errors.NewErr("commitPos, freezePos should be 0!")
		}
		voteInfo.FreezePos = voteInfo.ConsensusPos + voteInfo.NewPos
		voteInfo.NewPos = 0
		voteInfo.ConsensusPos = 0
		withdrawPos := voteInfo.WithdrawPos
		withdrawFreezePos := voteInfo.WithdrawFreezePos
		voteInfo.WithdrawFreezePos = withdrawPos
		voteInfo.WithdrawUnfreezePos = voteInfo.WithdrawUnfreezePos + withdrawFreezePos
		voteInfo.WithdrawPos = 0

		bf := new(bytes.Buffer)
		if err := voteInfo.Serialize(bf); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize voteInfo error!")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix,
			address[:]), &cstates.StorageItem{Value: bf.Bytes()})
	}
	return nil
}

func unConsensusToUnConsensus(native *native.NativeService, contract common.Address, peerPoolItem *PeerPoolItem) error {
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	//update voteInfoPool
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
	}
	voteInfo := new(VoteInfo)
	for _, v := range stateValues {
		voteInfoStore, _ := v.Value.(*cstates.StorageItem)
		if err := voteInfo.Deserialize(bytes.NewBuffer(voteInfoStore.Value)); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfo error!")
		}
		address := voteInfo.Address
		if voteInfo.ConsensusPos != 0 {
			return errors.NewErr("consensusPos, freezePos should be 0!")
		}

		newPos := voteInfo.NewPos
		freezePos := voteInfo.FreezePos
		voteInfo.NewPos = 0
		voteInfo.FreezePos = newPos + freezePos
		withdrawPos := voteInfo.WithdrawPos
		withdrawFreezePos := voteInfo.WithdrawFreezePos
		voteInfo.WithdrawFreezePos = withdrawPos
		voteInfo.WithdrawUnfreezePos = voteInfo.WithdrawUnfreezePos + withdrawFreezePos
		voteInfo.WithdrawPos = 0

		bf := new(bytes.Buffer)
		if err := voteInfo.Serialize(bf); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize voteInfo error!")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix,
			address[:]), &cstates.StorageItem{Value: bf.Bytes()})
	}
	return nil
}
