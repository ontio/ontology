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
	"math/big"
	"sort"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/genesis"
	cstates "github.com/ontio/ontology/core/states"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	//status
	RegisterSyncNodeStatus Status = iota
	SyncNodeStatus
	RegisterCandidateStatus
	CandidateStatus
	ConsensusStatus
	QuitConsensusStatus
	QuitingStatus
)

const (
	//function name
	INIT_CONFIG        = "initConfig"
	REGISTER_SYNC_NODE = "registerSyncNode"
	APPROVE_SYNC_NODE  = "approveSyncNode"
	REGISTER_CANDIDATE = "registerCandidate"
	APPROVE_CANDIDATE  = "approveCandidate"
	BLACK_NODE         = "blackNode"
	WHITE_NODE         = "whiteNode"
	QUIT_NODE          = "quitNode"
	VOTE_FOR_PEER      = "voteForPeer"
	WITHDRAW           = "withdraw"
	COMMIT_DPOS        = "commitDpos"
	VOTE_COMMIT_DPOS   = "voteCommitDpos"
	UPDATE_CONFIG      = "updateConfig"
	CALL_SPLIT         = "callSplit"

	//key prefix
	VBFT_CONFIG      = "vbftConfig"
	GOVERNANCE_VIEW  = "governanceView"
	CANDIDITE_INDEX  = "candidateIndex"
	PEER_POOL        = "peerPool"
	VOTE_INFO_POOL   = "voteInfoPool"
	POS_FOR_COMMIT   = "posForCommit"
	VOTE_COMMIT_INFO = "voteCommitInfo"
	PEER_INDEX       = "peerIndex"
	BLACK_LIST       = "blackList"

	//global
	SYNC_NODE_FEE      = 50
	CANDIDATE_FEE      = 500
	MIN_INIT_STAKE     = 1000
	POS_COMMIT_TRIGGER = 100000
	CANDIDATE_NUM      = 7 * 7
	SYNC_NODE_NUM      = 7 * 7 * 7
	A                  = 0.5
	B                  = 0.45
	C                  = 0.05
	TOTAL_ONG          = 10000000000
	PRECISE            = 100000000
	YITA               = 5
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

func InitGovernance() {
	native.Contracts[genesis.GovernanceContractAddress] = RegisterGovernanceContract
}

func RegisterGovernanceContract(native *native.NativeService) {
	native.Register(INIT_CONFIG, InitConfig)
	native.Register(REGISTER_SYNC_NODE, RegisterSyncNode)
	native.Register(APPROVE_SYNC_NODE, ApproveSyncNode)
	native.Register(REGISTER_CANDIDATE, RegisterCandidate)
	native.Register(APPROVE_CANDIDATE, ApproveCandidate)
	native.Register(BLACK_NODE, BlackNode)
	native.Register(WHITE_NODE, WhiteNode)
	native.Register(QUIT_NODE, QuitNode)
	native.Register(VOTE_FOR_PEER, VoteForPeer)
	native.Register(WITHDRAW, Withdraw)
	native.Register(COMMIT_DPOS, CommitDpos)
	native.Register(VOTE_COMMIT_DPOS, VoteCommitDpos)
	native.Register(UPDATE_CONFIG, UpdateConfig)
	native.Register(CALL_SPLIT, CallSplit)
}

func InitConfig(native *native.NativeService) ([]byte, error) {
	configuration := config.DefConfig.Genesis.VBFT
	contract := native.ContextRef.CurrentContext().ContractAddress

	//check the configuration
	if configuration.L < 16*configuration.K {
		return utils.BYTE_FALSE, errors.NewErr("initConfig. L is less than 16*K in config!")
	}
	view := new(big.Int).SetInt64(1)

	indexMap := make(map[uint32]struct{})
	var maxId uint32
	peers := []*PeerStakeInfo{}
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
		bf := new(bytes.Buffer)
		if err := peerPoolMap.Serialize(bf); err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize peerPoolMap error!")
		}
		peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
		}

		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), new(big.Int).Bytes()),
			&cstates.StorageItem{Value: bf.Bytes()})
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), view.Bytes()),
			&cstates.StorageItem{Value: bf.Bytes()})
		index := peerPoolItem.Index
		buf := new(bytes.Buffer)
		err = serialization.WriteUint32(buf, index)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, writeUint32 error!")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_INDEX), peerPubkeyPrefix), &cstates.StorageItem{Value: buf.Bytes()})

		peers = append(peers, &PeerStakeInfo{
			Index:      peerPoolItem.Index,
			PeerPubkey: peerPoolItem.PeerPubkey,
			Stake:      peerPoolItem.InitPos,
		})
	}

	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(CANDIDITE_INDEX)),
		&cstates.StorageItem{Value: new(big.Int).SetUint64(uint64(maxId + 1)).Bytes()})

	governanceView := &GovernanceView{
		View:       view,
		VoteCommit: false,
	}
	bf := new(bytes.Buffer)
	if err := governanceView.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize governanceView error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(GOVERNANCE_VIEW)), &cstates.StorageItem{Value: bf.Bytes()})

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

	//posTable, chainPeers, err := calDposTable(native, config, peers)

	bf = new(bytes.Buffer)
	if err := config.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize config error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VBFT_CONFIG)), &cstates.StorageItem{Value: bf.Bytes()})

	utils.AddCommonEvent(native, contract, INIT_CONFIG, true)

	return utils.BYTE_TRUE, nil
}

func RegisterSyncNode(native *native.NativeService) ([]byte, error) {
	params := new(RegisterSyncNodeParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}
	address, err := common.AddressFromBase58(params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "common.AddressFromBase58, address format error!")
	}

	//check initPos
	if params.InitPos <= 0 {
		return utils.BYTE_FALSE, errors.NewErr("registerSyncNode, initPos must > 0!")
	}

	//check witness
	err = utils.ValidateOwner(native, address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "validateOwner, checkWitness error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//check initPos
	if params.InitPos < MIN_INIT_STAKE {
		return utils.BYTE_FALSE, errors.NewErr(fmt.Sprintf("registerSyncNode, initPos must >= %v!", MIN_INIT_STAKE))
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
		return utils.BYTE_FALSE, errors.NewErr("registerSyncNode, this Peer is in BlackList!")
	}

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getView, get view error!")
	}

	//check if PeerPool full
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}

	if len(peerPoolMap.PeerPoolMap) >= SYNC_NODE_NUM {
		return utils.BYTE_FALSE, errors.NewErr("registerSyncNode, sync node is full (7*7*7)!")
	}

	//check if exist in PeerPool
	_, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if ok {
		return utils.BYTE_FALSE, errors.NewErr("registerSyncNode, peerPubkey is already in peerPoolMap!")
	}

	peerPoolItem := &PeerPoolItem{
		PeerPubkey: params.PeerPubkey,
		Address:    address,
		InitPos:    params.InitPos,
		Status:     RegisterSyncNodeStatus,
	}
	peerPoolMap.PeerPoolMap[params.PeerPubkey] = peerPoolItem
	bf := new(bytes.Buffer)
	if err := peerPoolMap.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize peerPoolMap error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), view.Bytes()), &cstates.StorageItem{Value: bf.Bytes()})

	//ont transfer
	err = AppCallTransferOnt(native, address, genesis.GovernanceContractAddress, params.InitPos)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOnt, ont transfer error!")
	}
	//ong transfer
	err = AppCallTransferOng(native, address, genesis.GovernanceContractAddress, SYNC_NODE_FEE)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOng, ong transfer error!")
	}

	utils.AddCommonEvent(native, contract, REGISTER_SYNC_NODE, params)

	return utils.BYTE_TRUE, nil
}

func ApproveSyncNode(native *native.NativeService) ([]byte, error) {
	params := new(ApproveSyncNodeParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}

	//TODO: check witness
	//err = utils.ValidateOwner(native, ADMIN_ADDRESS)
	//if err != nil {
	//	return errors.NewDetailErr(err, errors.ErrNoCode, "approveSyncNode, checkWitness error!")
	//}

	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getView, get view error!")
	}

	//get peerPool
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}

	peerPoolItem, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, errors.NewErr("approveSyncNode, peerPubkey is not in peerPoolMap!")
	}

	if peerPoolItem.Status != RegisterSyncNodeStatus {
		return utils.BYTE_FALSE, errors.NewErr("approveSyncNode, peer status is not RegisterSyncNodeStatus!")
	}

	peerPoolItem.Status = SyncNodeStatus

	peerPoolMap.PeerPoolMap[params.PeerPubkey] = peerPoolItem
	bf := new(bytes.Buffer)
	if err := peerPoolMap.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize peerPoolMap error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), view.Bytes()), &cstates.StorageItem{Value: bf.Bytes()})

	utils.AddCommonEvent(native, contract, APPROVE_CANDIDATE, params)

	return utils.BYTE_TRUE, nil
}

func RegisterCandidate(native *native.NativeService) ([]byte, error) {
	params := new(RegisterCandidateParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}
	address, err := common.AddressFromBase58(params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "common.AddressFromBase58, address format error!")
	}

	//check witness
	err = utils.ValidateOwner(native, address)
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
		return utils.BYTE_FALSE, errors.NewErr("registerCandidate, peerPubkey is not in peerPoolMap!")
	}

	if peerPoolItem.Address != address {
		return utils.BYTE_FALSE, errors.NewErr("registerCandidate, peer is not registered by this address!")
	}
	if peerPoolItem.Status != SyncNodeStatus {
		return utils.BYTE_FALSE, errors.NewErr("registerCandidate, peer status is not SyncNodeStatus!")
	}

	peerPoolItem.Status = RegisterCandidateStatus

	peerPoolMap.PeerPoolMap[params.PeerPubkey] = peerPoolItem
	bf := new(bytes.Buffer)
	if err := peerPoolMap.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize peerPoolMap error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), view.Bytes()), &cstates.StorageItem{Value: bf.Bytes()})

	//ong transfer
	err = AppCallTransferOng(native, address, genesis.GovernanceContractAddress, CANDIDATE_FEE)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOng, ong transfer error!")
	}

	utils.AddCommonEvent(native, contract, REGISTER_CANDIDATE, params)

	return utils.BYTE_TRUE, nil
}

func ApproveCandidate(native *native.NativeService) ([]byte, error) {
	params := new(ApproveCandidateParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}

	//TODO: check witness
	//err = utils.ValidateOwner(native, ADMIN_ADDRESS)
	//if err != nil {
	//	return errors.NewDetailErr(err, errors.ErrNoCode, "approveCandidate, checkWitness error!")
	//}

	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getView, get view error!")
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
	if num >= CANDIDATE_NUM {
		return utils.BYTE_FALSE, errors.NewErr("approveCandidate, num of candidate node is full (7*7)!")
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
		buf := bytes.NewBuffer(indexBytes.(*cstates.StorageItem).Value)
		peerPoolItem.Index, err = serialization.ReadUint32(buf)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, readUint32 error!")
		}
	} else {
		//get index
		candidateIndexBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(CANDIDITE_INDEX)))
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Get, get candidateIndex error!")
		}
		var candidateIndex uint64
		if candidateIndexBytes == nil {
			return utils.BYTE_FALSE, errors.NewErr("approveCandidate, candidateIndex is not init!")
		} else {
			candidateIndexStore, _ := candidateIndexBytes.(*cstates.StorageItem)
			candidateIndex = new(big.Int).SetBytes(candidateIndexStore.Value).Uint64()
		}
		peerPoolItem.Index = uint32(candidateIndex)

		//update candidateIndex
		newCandidateIndex := candidateIndex + 1
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(CANDIDITE_INDEX)),
			&cstates.StorageItem{Value: new(big.Int).SetUint64(newCandidateIndex).Bytes()})

		buf := new(bytes.Buffer)
		err = serialization.WriteUint32(buf, peerPoolItem.Index)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, writeUint32 error!")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_INDEX), peerPubkeyPrefix), &cstates.StorageItem{Value: buf.Bytes()})
	}
	peerPoolMap.PeerPoolMap[params.PeerPubkey] = peerPoolItem
	bf := new(bytes.Buffer)
	if err := peerPoolMap.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize peerPoolMap error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), view.Bytes()), &cstates.StorageItem{Value: bf.Bytes()})

	utils.AddCommonEvent(native, contract, APPROVE_CANDIDATE, params)

	return utils.BYTE_TRUE, nil
}

func BlackNode(native *native.NativeService) ([]byte, error) {
	params := new(BlackNodeParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}

	//TODO: check witness
	//err = utils.ValidateOwner(native, ADMIN_ADDRESS)
	//if err != nil {
	//	return errors.NewDetailErr(err, errors.ErrNoCode, "[BlackNode] CheckWitness error!")
	//}
	contract := native.ContextRef.CurrentContext().ContractAddress

	peerPubkeyPrefix, err := hex.DecodeString(params.PeerPubkey)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	//put peer into black list
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(BLACK_LIST), peerPubkeyPrefix), &cstates.StorageItem{Value: new(big.Int).SetUint64(1).Bytes()})

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
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), view.Bytes()), &cstates.StorageItem{Value: bf.Bytes()})

	return utils.BYTE_TRUE, nil
}

func WhiteNode(native *native.NativeService) ([]byte, error) {
	params := new(WhiteNodeParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}

	//TODO: check witness
	//err = utils.ValidateOwner(native, ADMIN_ADDRESS)
	//if err != nil {
	//	return errors.NewDetailErr(err, errors.ErrNoCode, "[WhiteNode] CheckWitness error!")
	//}
	contract := native.ContextRef.CurrentContext().ContractAddress

	peerPubkeyPrefix, err := hex.DecodeString(params.PeerPubkey)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	//remove peer from black list
	native.CloneCache.Delete(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(BLACK_LIST), peerPubkeyPrefix))

	return utils.BYTE_TRUE, nil
}

func QuitNode(native *native.NativeService) ([]byte, error) {
	params := new(QuitNodeParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}
	address, err := common.AddressFromBase58(params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "common.AddressFromBase58, address format error!")
	}

	//check witness
	err = utils.ValidateOwner(native, address)
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
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), view.Bytes()), &cstates.StorageItem{Value: bf.Bytes()})

	utils.AddCommonEvent(native, contract, QUIT_NODE, params)

	return utils.BYTE_TRUE, nil
}

func VoteForPeer(native *native.NativeService) ([]byte, error) {
	params := &VoteForPeerParam{
		VoteTable: make(map[string]int64),
	}
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}
	address, err := common.AddressFromBase58(params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "common.AddressFromBase58, address format error!")
	}

	//check witness
	err = utils.ValidateOwner(native, address)
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

	var total int64
	for peerPubkey, pos := range params.VoteTable {
		peerPubkeyPrefix, err := hex.DecodeString(peerPubkey)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
		}

		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peerPubkey]
		if !ok {
			return utils.BYTE_FALSE, errors.NewErr("voteForPeer, peerPubkey is not in peerPoolMap!")
		}

		if peerPoolItem.Status != CandidateStatus && peerPoolItem.Status != ConsensusStatus {
			continue
		}

		voteInfoPoolBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL),
			peerPubkeyPrefix, address[:]))
		voteInfoPool := &VoteInfoPool{
			PeerPubkey: peerPubkey,
			Address:    address,
		}
		if pos >= 0 {
			if voteInfoPoolBytes != nil {
				voteInfoPoolStore, _ := voteInfoPoolBytes.(*cstates.StorageItem)
				if err := voteInfoPool.Deserialize(bytes.NewBuffer(voteInfoPoolStore.Value)); err != nil {
					return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfoPool error!")
				}
				voteInfoPool.NewPos = voteInfoPool.NewPos + uint64(pos)
			} else {
				voteInfoPool.NewPos = uint64(pos)
			}
			total = total + pos
			peerPoolItem.TotalPos = peerPoolItem.TotalPos + uint64(pos)
		} else {
			if voteInfoPoolBytes != nil {
				voteInfoPoolStore, _ := voteInfoPoolBytes.(*cstates.StorageItem)
				if err := voteInfoPool.Deserialize(bytes.NewBuffer(voteInfoPoolStore.Value)); err != nil {
					return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfoPool error!")
				}
				temp := int64(voteInfoPool.NewPos) + pos
				if temp < 0 {
					if peerPoolItem.Status == ConsensusStatus {
						consensusPos := int64(voteInfoPool.ConsensusPos) + temp
						if consensusPos < 0 {
							continue
						}
						newPos := voteInfoPool.NewPos
						voteInfoPool.NewPos = 0
						voteInfoPool.WithdrawUnfreezePos = voteInfoPool.WithdrawUnfreezePos + newPos
						voteInfoPool.ConsensusPos = uint64(consensusPos)
						voteInfoPool.WithdrawPos = uint64(int64(voteInfoPool.WithdrawPos) - temp)
						peerPoolItem.TotalPos = uint64(int64(peerPoolItem.TotalPos) + pos)
					}
					if peerPoolItem.Status == CandidateStatus {
						freezePos := int64(voteInfoPool.FreezePos) + temp
						if freezePos < 0 {
							continue
						}
						newPos := voteInfoPool.NewPos
						voteInfoPool.NewPos = 0
						voteInfoPool.WithdrawUnfreezePos = voteInfoPool.WithdrawUnfreezePos + newPos
						voteInfoPool.FreezePos = uint64(freezePos)
						voteInfoPool.WithdrawFreezePos = uint64(int64(voteInfoPool.WithdrawFreezePos) - temp)
						peerPoolItem.TotalPos = uint64(int64(peerPoolItem.TotalPos) + pos)
					}
				} else {
					voteInfoPool.NewPos = uint64(temp)
					voteInfoPool.WithdrawUnfreezePos = uint64(int64(voteInfoPool.WithdrawUnfreezePos) - pos)
					peerPoolItem.TotalPos = uint64(int64(peerPoolItem.TotalPos) + pos)
				}
			} else {
				continue
			}
		}
		peerPoolMap.PeerPoolMap[peerPubkey] = peerPoolItem
		bf := new(bytes.Buffer)
		if err := voteInfoPool.Serialize(bf); err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize voteInfoPool error!")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix,
			address[:]), &cstates.StorageItem{Value: bf.Bytes()})
	}
	bf := new(bytes.Buffer)
	if err := peerPoolMap.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize peerPoolMap error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), view.Bytes()), &cstates.StorageItem{Value: bf.Bytes()})

	//ont transfer
	err = AppCallTransferOnt(native, address, genesis.GovernanceContractAddress, uint64(total))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOnt, ont transfer error!")
	}

	utils.AddCommonEvent(native, contract, VOTE_FOR_PEER, params)

	return utils.BYTE_TRUE, nil
}

func Withdraw(native *native.NativeService) ([]byte, error) {
	params := &WithdrawParam{
		WithdrawTable: make(map[string]uint64),
	}
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}
	address, err := common.AddressFromBase58(params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "common.AddressFromBase58, address format error!")
	}

	//check witness
	err = utils.ValidateOwner(native, address)
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
	for peerPubkey, pos := range params.WithdrawTable {
		peerPubkeyPrefix, err := hex.DecodeString(peerPubkey)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
		}

		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peerPubkey]
		if !ok {
			return utils.BYTE_FALSE, errors.NewErr("voteForPeer, peerPubkey is not in peerPoolMap!")
		}

		if peerPoolItem.Status != CandidateStatus && peerPoolItem.Status != ConsensusStatus {
			continue
		}

		voteInfoPoolBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL),
			peerPubkeyPrefix, address[:]))
		voteInfoPool := new(VoteInfoPool)
		if voteInfoPoolBytes != nil {
			voteInfoPoolStore, _ := voteInfoPoolBytes.(*cstates.StorageItem)
			if err := voteInfoPool.Deserialize(bytes.NewBuffer(voteInfoPoolStore.Value)); err != nil {
				return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfoPool error!")
			}
			if voteInfoPool.WithdrawUnfreezePos < pos {
				continue
			} else {
				voteInfoPool.WithdrawUnfreezePos = voteInfoPool.WithdrawUnfreezePos - pos
				total = total + pos
			}
		} else {
			continue
		}
		if voteInfoPool.ConsensusPos == 0 && voteInfoPool.FreezePos == 0 && voteInfoPool.NewPos == 0 &&
			voteInfoPool.WithdrawPos == 0 && voteInfoPool.WithdrawFreezePos == 0 && voteInfoPool.WithdrawUnfreezePos == 0 {
			native.CloneCache.Delete(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix, address[:]))
		}
	}

	//ont transfer
	err = AppCallTransferOnt(native, genesis.GovernanceContractAddress, address, total)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOnt, ont transfer error!")
	}

	utils.AddCommonEvent(native, contract, WITHDRAW, params)

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

	//TODO: check witness
	//err = utils.ValidateOwner(native, ADMIN_ADDRESS)
	//if err != nil {
	//	cycle := native.Height % config.MaxBlockChangeView == 0
	//	if !cycle && !governanceView.VoteCommit {
	//		return utils.BYTE_FALSE, errors.NewErr("[CommitDpos] Authentication Failed!")
	//	}
	//}

	err = executeCommitDpos(native, contract, config)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "commitDpos, commitDpos error!")
	}

	utils.AddCommonEvent(native, contract, COMMIT_DPOS, true)

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
	newView := new(big.Int).Add(view, new(big.Int).SetInt64(1))

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
		peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
		}

		if peerPoolItem.Status == QuitingStatus {
			//draw back init pos
			address := peerPoolItem.Address
			//ont transfer
			err = AppCallTransferOnt(native, genesis.GovernanceContractAddress, address, peerPoolItem.InitPos)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOnt, ont transfer error!")
			}

			//draw back vote pos
			stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix))
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
			}
			voteInfoPool := new(VoteInfoPool)
			for _, v := range stateValues {
				voteInfoPoolStore, _ := v.Value.(*cstates.StorageItem)
				if err := voteInfoPool.Deserialize(bytes.NewBuffer(voteInfoPoolStore.Value)); err != nil {
					return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfoPool error!")
				}
				pos := voteInfoPool.ConsensusPos + voteInfoPool.FreezePos + voteInfoPool.NewPos + voteInfoPool.WithdrawPos +
					voteInfoPool.WithdrawFreezePos + voteInfoPool.WithdrawUnfreezePos

				address := voteInfoPool.Address
				//ont transfer
				err = AppCallTransferOnt(native, genesis.GovernanceContractAddress, address, pos)
				if err != nil {
					return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOnt, ont transfer error!")
				}
				native.CloneCache.Delete(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix, address[:]))
			}
			delete(peerPoolMap.PeerPoolMap, peerPoolItem.PeerPubkey)
		}
		if peerPoolItem.Status == QuitConsensusStatus {
			peerPoolItem.Status = QuitingStatus
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
		//change peerPool status
		peerPubkeyPrefix, err := hex.DecodeString(peers[i].PeerPubkey)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
		}

		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peers[i].PeerPubkey]
		if !ok {
			return errors.NewErr("voteForPeer, peerPubkey is not in peerPoolMap!")
		}

		if peerPoolItem.Status == ConsensusStatus {
			//update voteInfoPool
			stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix))
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
			}
			voteInfoPool := new(VoteInfoPool)
			for _, v := range stateValues {
				voteInfoPoolStore, _ := v.Value.(*cstates.StorageItem)
				if err := voteInfoPool.Deserialize(bytes.NewBuffer(voteInfoPoolStore.Value)); err != nil {
					return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfoPool error!")
				}
				address := voteInfoPool.Address
				if voteInfoPool.FreezePos != 0 {
					return errors.NewErr("commitPos, freezePos should be 0!")
				}
				newPos := voteInfoPool.NewPos
				voteInfoPool.ConsensusPos = voteInfoPool.ConsensusPos + newPos
				voteInfoPool.NewPos = 0
				withdrawPos := voteInfoPool.WithdrawPos
				withdrawFreezePos := voteInfoPool.WithdrawFreezePos
				voteInfoPool.WithdrawFreezePos = withdrawPos
				voteInfoPool.WithdrawUnfreezePos = voteInfoPool.WithdrawUnfreezePos + withdrawFreezePos
				voteInfoPool.WithdrawPos = 0

				bf := new(bytes.Buffer)
				if err := voteInfoPool.Serialize(bf); err != nil {
					return errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize voteInfoPool error!")
				}
				native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix,
					address[:]), &cstates.StorageItem{Value: bf.Bytes()})
			}
		} else {
			//update voteInfoPool
			stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix))
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
			}
			voteInfoPool := new(VoteInfoPool)
			for _, v := range stateValues {
				voteInfoPoolStore, _ := v.Value.(*cstates.StorageItem)
				if err := voteInfoPool.Deserialize(bytes.NewBuffer(voteInfoPoolStore.Value)); err != nil {
					return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfoPool error!")
				}
				address := voteInfoPool.Address
				if voteInfoPool.ConsensusPos != 0 {
					return errors.NewErr("consensusPos, freezePos should be 0!")
				}

				voteInfoPool.ConsensusPos = voteInfoPool.ConsensusPos + voteInfoPool.FreezePos + voteInfoPool.NewPos
				voteInfoPool.NewPos = 0
				voteInfoPool.FreezePos = 0
				withdrawPos := voteInfoPool.WithdrawPos
				withdrawFreezePos := voteInfoPool.WithdrawFreezePos
				voteInfoPool.WithdrawFreezePos = withdrawPos
				voteInfoPool.WithdrawUnfreezePos = voteInfoPool.WithdrawUnfreezePos + withdrawFreezePos
				voteInfoPool.WithdrawPos = 0

				bf := new(bytes.Buffer)
				if err := voteInfoPool.Serialize(bf); err != nil {
					return errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize voteInfoPool error!")
				}
				native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix,
					address[:]), &cstates.StorageItem{Value: bf.Bytes()})
			}
		}
		peerPoolItem.Status = ConsensusStatus
		peerPoolMap.PeerPoolMap[peers[i].PeerPubkey] = peerPoolItem
	}

	//non consensus peers
	for i := int(config.K); i < len(peers); i++ {
		//change peerPool status
		peerPubkeyPrefix, err := hex.DecodeString(peers[i].PeerPubkey)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
		}

		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peers[i].PeerPubkey]
		if !ok {
			return errors.NewErr("voteForPeer, peerPubkey is not in peerPoolMap!")
		}

		if peerPoolItem.Status == ConsensusStatus {
			//update voteInfoPool
			stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix))
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
			}
			voteInfoPool := new(VoteInfoPool)
			for _, v := range stateValues {
				voteInfoPoolStore, _ := v.Value.(*cstates.StorageItem)
				if err := voteInfoPool.Deserialize(bytes.NewBuffer(voteInfoPoolStore.Value)); err != nil {
					return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfoPool error!")
				}
				address := voteInfoPool.Address
				if voteInfoPool.FreezePos != 0 {
					return errors.NewErr("commitPos, freezePos should be 0!")
				}
				voteInfoPool.FreezePos = voteInfoPool.ConsensusPos + voteInfoPool.NewPos
				voteInfoPool.NewPos = 0
				voteInfoPool.ConsensusPos = 0
				withdrawPos := voteInfoPool.WithdrawPos
				withdrawFreezePos := voteInfoPool.WithdrawFreezePos
				voteInfoPool.WithdrawFreezePos = withdrawPos
				voteInfoPool.WithdrawUnfreezePos = voteInfoPool.WithdrawUnfreezePos + withdrawFreezePos
				voteInfoPool.WithdrawPos = 0

				bf := new(bytes.Buffer)
				if err := voteInfoPool.Serialize(bf); err != nil {
					return errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize voteInfoPool error!")
				}
				native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix,
					address[:]), &cstates.StorageItem{Value: bf.Bytes()})
			}
		} else {
			//update voteInfoPool
			stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix))
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
			}
			voteInfoPool := new(VoteInfoPool)
			for _, v := range stateValues {
				voteInfoPoolStore, _ := v.Value.(*cstates.StorageItem)
				if err := voteInfoPool.Deserialize(bytes.NewBuffer(voteInfoPoolStore.Value)); err != nil {
					return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfoPool error!")
				}
				address := voteInfoPool.Address
				if voteInfoPool.ConsensusPos != 0 {
					return errors.NewErr("consensusPos, freezePos should be 0!")
				}

				newPos := voteInfoPool.NewPos
				freezePos := voteInfoPool.FreezePos
				voteInfoPool.NewPos = freezePos
				voteInfoPool.FreezePos = newPos
				withdrawPos := voteInfoPool.WithdrawPos
				withdrawFreezePos := voteInfoPool.WithdrawFreezePos
				voteInfoPool.WithdrawFreezePos = withdrawPos
				voteInfoPool.WithdrawUnfreezePos = voteInfoPool.WithdrawUnfreezePos + withdrawFreezePos
				voteInfoPool.WithdrawPos = 0

				bf := new(bytes.Buffer)
				if err := voteInfoPool.Serialize(bf); err != nil {
					return errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize voteInfoPool error!")
				}
				native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix,
					address[:]), &cstates.StorageItem{Value: bf.Bytes()})
			}
		}
		peerPoolItem.Status = CandidateStatus
		peerPoolMap.PeerPoolMap[peers[i].PeerPubkey] = peerPoolItem
	}
	bf := new(bytes.Buffer)
	if err := peerPoolMap.Serialize(bf); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize peerPoolMap error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), newView.Bytes()), &cstates.StorageItem{Value: bf.Bytes()})
	oldView := new(big.Int).Sub(view, new(big.Int).SetUint64(1))
	native.CloneCache.Delete(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), oldView.Bytes()))

	//get all vote for commit info
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_COMMIT_INFO)))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
	}

	voteCommitInfoPool := new(VoteCommitInfoPool)
	for _, v := range stateValues {
		voteCommitInfoPoolStore, _ := v.Value.(*cstates.StorageItem)
		if err := voteCommitInfoPool.Deserialize(bytes.NewBuffer(voteCommitInfoPoolStore.Value)); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteCommitInfoPool error!")
		}

		address := voteCommitInfoPool.Address
		//ont transfer
		err = AppCallTransferOnt(native, genesis.GovernanceContractAddress, address, voteCommitInfoPool.Pos)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOnt, ont transfer error!")
		}
		native.CloneCache.Delete(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_COMMIT_INFO), address[:]))
	}
	native.CloneCache.Delete(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(POS_FOR_COMMIT)))

	//posTable, chainPeers, err := calDposTable(native, config, peers)

	//update view
	governanceView = &GovernanceView{
		View:       newView,
		VoteCommit: false,
	}
	bf = new(bytes.Buffer)
	if err := governanceView.Serialize(bf); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize governanceView error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(GOVERNANCE_VIEW)), &cstates.StorageItem{Value: bf.Bytes()})

	return nil
}

func VoteCommitDpos(native *native.NativeService) ([]byte, error) {
	params := new(VoteCommitDposParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}
	address, err := common.AddressFromBase58(params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "common.AddressFromBase58, address format error!")
	}

	//check witness
	err = utils.ValidateOwner(native, address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "validateOwner, checkWitness error!")
	}

	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "getView, get view error!")
	}

	//get voteCommitInfo
	voteCommitInfoPool := new(VoteCommitInfoPool)
	voteCommitInfoPoolBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_COMMIT_INFO), address[:]))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Get, get voteCommitInfoBytes error!")
	}
	if voteCommitInfoPoolBytes != nil {
		voteCommitInfoPoolStore, _ := voteCommitInfoPoolBytes.(*cstates.StorageItem)
		if err := voteCommitInfoPool.Deserialize(bytes.NewBuffer(voteCommitInfoPoolStore.Value)); err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteCommitInfoPool error!")
		}
	}
	pos := int64(voteCommitInfoPool.Pos) + params.Pos
	if pos < 0 {
		return utils.BYTE_FALSE, errors.NewErr("voteCommitDpos, remain pos is negative!")
	}
	voteCommitInfoPool.Pos = uint64(pos)
	bf := new(bytes.Buffer)
	if err := voteCommitInfoPool.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize voteCommitInfoPool error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_COMMIT_INFO), address[:]), &cstates.StorageItem{Value: bf.Bytes()})

	//get total pos for commit
	posCommitBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(POS_FOR_COMMIT)))
	posCommit := new(big.Int)
	if posCommitBytes != nil {
		posCommitStore, _ := posCommitBytes.(*cstates.StorageItem)
		posCommit = new(big.Int).SetBytes(posCommitStore.Value)
	}
	newPosCommit := uint64(posCommit.Int64() + params.Pos)

	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(POS_FOR_COMMIT)), &cstates.StorageItem{Value: new(big.Int).SetUint64(newPosCommit).Bytes()})

	if newPosCommit >= POS_COMMIT_TRIGGER {
		governanceView := &GovernanceView{
			View:       view,
			VoteCommit: true,
		}
		bf := new(bytes.Buffer)
		if err := governanceView.Serialize(bf); err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize governanceView error!")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(GOVERNANCE_VIEW)), &cstates.StorageItem{Value: bf.Bytes()})

		// get config
		config := new(Configuration)
		configBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VBFT_CONFIG)))
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Get, get configBytes error!")
		}
		if configBytes == nil {
			return utils.BYTE_FALSE, errors.NewErr("voteCommitDpos, configBytes is nil!")
		}
		configStore, _ := configBytes.(*cstates.StorageItem)
		if err := config.Deserialize(bytes.NewBuffer(configStore.Value)); err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize config error!")
		}

		err = executeCommitDpos(native, contract, config)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "commitDpos, commitDpos error!")
		}
	}

	//ont transfer
	if params.Pos > 0 {
		err = AppCallTransferOnt(native, address, genesis.GovernanceContractAddress, uint64(params.Pos))
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOnt, ont transfer error!")
		}
	}
	if params.Pos < 0 {
		err = AppCallTransferOnt(native, genesis.GovernanceContractAddress, address, uint64(-params.Pos))
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOnt, ont transfer error!")
		}
	}

	utils.AddCommonEvent(native, contract, VOTE_COMMIT_DPOS, params)

	return utils.BYTE_TRUE, nil
}

func UpdateConfig(native *native.NativeService) ([]byte, error) {
	configuration := new(Configuration)
	if err := configuration.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize configuration error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//check the configuration
	if configuration.L < 16*configuration.K {
		return utils.BYTE_FALSE, errors.NewErr("updateConfig, L is less than 16*K in config!")
	}

	bf := new(bytes.Buffer)
	if err := configuration.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize configuration error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VBFT_CONFIG)), &cstates.StorageItem{Value: bf.Bytes()})

	utils.AddCommonEvent(native, contract, UPDATE_CONFIG, configuration)

	return utils.BYTE_TRUE, nil
}

func CallSplit(native *native.NativeService) ([]byte, error) {
	//TODO: check witness
	//err = utils.ValidateOwner(native, ADMIN_ADDRESS)
	//if err != nil {
	//	return errors.NewDetailErr(err, errors.ErrNoCode, "callSplit, checkWitness error!")
	//}

	contract := genesis.GovernanceContractAddress
	//get current view
	cView, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "callSplit, get view error!")
	}
	view := new(big.Int).Sub(cView, new(big.Int).SetInt64(1))

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "callSplit, get peerPoolMap error!")
	}

	err = executeSplit(native, contract, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, executeSplitp error!")
	}

	utils.AddCommonEvent(native, genesis.GovernanceContractAddress, CALL_SPLIT, true)

	return utils.BYTE_TRUE, nil
}

func executeSplit(native *native.NativeService, contract common.Address, peerPoolMap *PeerPoolMap) error {
	peersCandidate := []*CandidateSplitInfo{}
	peersSyncNode := []*SyncNodeSplitInfo{}

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
		if peerPoolItem.Status == SyncNodeStatus || peerPoolItem.Status == RegisterCandidateStatus {
			peersSyncNode = append(peersSyncNode, &SyncNodeSplitInfo{
				PeerPubkey: peerPoolItem.PeerPubkey,
				InitPos:    peerPoolItem.InitPos,
				Address:    peerPoolItem.Address,
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
		peersCandidate[i].S = splitCurve(peersCandidate[i].Stake, avg)
		sumS += peersCandidate[i].S
	}

	//fee split of consensus peer
	var splitAmount uint64
	for i := int(config.K) - 1; i >= 0; i-- {
		nodeAmount := uint64(TOTAL_ONG * A * peersCandidate[i].S / sumS)
		address := peersCandidate[i].Address
		err = AppCallApproveOng(native, genesis.GovernanceContractAddress, address, nodeAmount)
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
		nodeAmount := uint64(TOTAL_ONG * B * peersCandidate[i].Stake / sum)
		address := peersCandidate[i].Address
		err = AppCallApproveOng(native, genesis.GovernanceContractAddress, address, nodeAmount)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, ong transfer error!")
		}
		splitAmount += nodeAmount
	}

	//fee split of syncNode peer
	// cal s of each candidate node
	var splitSyncNodeAmount uint64
	for _, syncNodeSplitInfo := range peersSyncNode {
		amount := uint64(TOTAL_ONG * C / len(peersSyncNode))
		address := syncNodeSplitInfo.Address
		err = AppCallApproveOng(native, genesis.GovernanceContractAddress, address, amount)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Ong transfer error!")
		}
		splitSyncNodeAmount += amount
	}
	return nil
}
