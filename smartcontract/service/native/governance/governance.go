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
	"encoding/json"
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
	COMMIT_DPOS        = "commitDpos"
	VOTE_COMMIT_DPOS   = "voteCommitDpos"
	UPDATE_CONFIG      = "updateConfig"

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
)

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
	native.Register(COMMIT_DPOS, CommitDpos)
	native.Register(VOTE_COMMIT_DPOS, VoteCommitDpos)
	native.Register(UPDATE_CONFIG, UpdateConfig)
	native.Register("dataQuery", DataQuery)
}

func InitConfig(native *native.NativeService) ([]byte, error) {
	configuration := config.DefConfig.Genesis.VBFT
	contract := native.ContextRef.CurrentContext().ContractAddress

	//check the configuration
	if configuration.L < 16*configuration.K {
		return utils.BYTE_FALSE, errors.NewErr("[InitConfig] L is less than 16*K in config!")
	}
	view := new(big.Int).SetInt64(1)

	indexMap := make(map[uint32]struct{})
	var maxId uint32
	peers := []*PeerStakeInfo{}
	peerPoolMap := &PeerPoolMap{
		PeerPoolMap: make(map[string]*PeerPool),
	}
	for _, peer := range configuration.Peers {
		peerPool := new(PeerPool)
		_, ok := indexMap[peer.Index]
		if ok {
			return utils.BYTE_FALSE, errors.NewErr("[InitConfig] Peer index is duplicated!")
		}
		indexMap[peer.Index] = struct{}{}
		if peer.Index <= 0 {
			return utils.BYTE_FALSE, errors.NewErr("[InitConfig] Peer index in config must > 0!")
		}
		if peer.Index > maxId {
			maxId = peer.Index
		}
		peerPool.Index = peer.Index
		peerPool.PeerPubkey = peer.PeerPubkey
		peerPool.Address = peer.Address
		peerPool.InitPos = peer.InitPos
		peerPool.TotalPos = 0
		peerPool.Status = ConsensusStatus
		peerPoolMap.PeerPoolMap[peerPool.PeerPubkey] = peerPool
		value, err := json.Marshal(peerPoolMap)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Marshal] Marshal peerPoolMap error!")
		}
		peerPubkeyPrefix, err := hex.DecodeString(peerPool.PeerPubkey)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[hex.DecodeString] PeerPubkey format error!")
		}

		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), new(big.Int).Bytes()),
			&cstates.StorageItem{Value: value})
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), view.Bytes()),
			&cstates.StorageItem{Value: value})
		index := peerPool.Index
		buf := new(bytes.Buffer)
		err = serialization.WriteUint32(buf, index)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[serialization.WriteUint32] WriteUint32 error!")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_INDEX), peerPubkeyPrefix), &cstates.StorageItem{Value: buf.Bytes()})

		peers = append(peers, &PeerStakeInfo{
			Index:      peerPool.Index,
			PeerPubkey: peerPool.PeerPubkey,
			Stake:      peerPool.InitPos,
		})
	}

	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(CANDIDITE_INDEX)),
		&cstates.StorageItem{Value: new(big.Int).SetUint64(uint64(maxId + 1)).Bytes()})

	governanceView := &GovernanceView{
		View:       view,
		VoteCommit: false,
	}
	v, err := json.Marshal(governanceView)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Marshal] Marshal governanceView error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(GOVERNANCE_VIEW)), &cstates.StorageItem{Value: v})

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

	value, err := json.Marshal(config)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Marshal] Marshal config error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VBFT_CONFIG)), &cstates.StorageItem{Value: value})

	utils.AddCommonEvent(native, contract, INIT_CONFIG, true)

	return utils.BYTE_TRUE, nil
}

func RegisterSyncNode(native *native.NativeService) ([]byte, error) {
	params := new(RegisterSyncNodeParam)
	err := json.Unmarshal(native.Input, params)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Contract params Unmarshal error!")
	}

	//check initPos
	if params.InitPos <= 0 {
		return utils.BYTE_FALSE, errors.NewErr("[registerSyncNode] InitPos must > 0!")
	}

	//check witness
	err = utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[ValidateOwner] CheckWitness error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//check initPos
	if params.InitPos < MIN_INIT_STAKE {
		return utils.BYTE_FALSE, errors.NewErr(fmt.Sprintf("[RegisterSyncNode] InitPos must >= %v!", MIN_INIT_STAKE))
	}

	peerPubkeyPrefix, err := hex.DecodeString(params.PeerPubkey)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[hex.DecodeString] PeerPubkey format error!")
	}
	//get black list
	blackList, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(BLACK_LIST), peerPubkeyPrefix))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[native.CloneCache.Get] Get BlackList error!")
	}
	if blackList != nil {
		return utils.BYTE_FALSE, errors.NewErr("[RegisterSyncNode] This Peer is in BlackList!")
	}

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetView] Get view error!")
	}

	//check if PeerPool full
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetPeerPoolMap] Get peerPoolMap error!")
	}

	if len(peerPoolMap.PeerPoolMap) >= SYNC_NODE_NUM {
		return utils.BYTE_FALSE, errors.NewErr("[RegisterSyncNode] Sync node is full (7*7*7)!")
	}

	//check if exist in PeerPool
	_, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if ok {
		return utils.BYTE_FALSE, errors.NewErr("[RegisterSyncNode] PeerPubkey is already in peerPoolMap!")
	}

	peerPool := &PeerPool{
		PeerPubkey: params.PeerPubkey,
		Address:    params.Address,
		InitPos:    params.InitPos,
		Status:     RegisterSyncNodeStatus,
	}
	peerPoolMap.PeerPoolMap[params.PeerPubkey] = peerPool
	value, err := json.Marshal(peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Marshal] Marshal peerPoolMap error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), view.Bytes()), &cstates.StorageItem{Value: value})

	addressBytes, err := hex.DecodeString(params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[hex.DecodeString] Address format error!")
	}
	address, err := common.AddressParseFromBytes(addressBytes)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[common.AddressParseFromBytes] Address format error!")
	}
	//ont transfer
	err = AppCallTransferOnt(native, address, genesis.GovernanceContractAddress, params.InitPos)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[appCallTransferOnt] Ont transfer error!")
	}
	//ong transfer
	err = AppCallTransferOng(native, address, genesis.FeeSplitContractAddress, SYNC_NODE_FEE)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[appCallTransferOng] Ong transfer error!")
	}

	utils.AddCommonEvent(native, contract, REGISTER_SYNC_NODE, params)

	return utils.BYTE_TRUE, nil
}

func ApproveSyncNode(native *native.NativeService) ([]byte, error) {
	params := new(ApproveSyncNodeParam)
	err := json.Unmarshal(native.Input, params)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Contract params Unmarshal error!")
	}

	//TODO: check witness
	//err = utils.ValidateOwner(native, ADMIN_ADDRESS)
	//if err != nil {
	//	return errors.NewDetailErr(err, errors.ErrNoCode, "[ApproveSyncNode] CheckWitness error!")
	//}

	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetView] Get view error!")
	}

	//get peerPool
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetPeerPoolMap] Get peerPoolMap error!")
	}

	peerPool, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, errors.NewErr("[ApproveSyncNode] PeerPubkey is not in peerPoolMap!")
	}

	if peerPool.Status != RegisterSyncNodeStatus {
		return utils.BYTE_FALSE, errors.NewErr("[ApproveSyncNode] Peer status is not RegisterSyncNodeStatus!")
	}

	peerPool.Status = SyncNodeStatus

	peerPoolMap.PeerPoolMap[params.PeerPubkey] = peerPool
	value, err := json.Marshal(peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Marshal] Marshal peerPoolMap error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), view.Bytes()), &cstates.StorageItem{Value: value})

	utils.AddCommonEvent(native, contract, APPROVE_CANDIDATE, params)

	return utils.BYTE_TRUE, nil
}

func RegisterCandidate(native *native.NativeService) ([]byte, error) {
	params := new(RegisterCandidateParam)
	err := json.Unmarshal(native.Input, params)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Contract params Unmarshal error!")
	}

	//check witness
	err = utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[ValidateOwner] CheckWitness error!")
	}

	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetView] Get view error!")
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetPeerPoolMap] Get peerPoolMap error!")
	}

	peerPool, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, errors.NewErr("[RegisterCandidate] PeerPubkey is not in peerPoolMap!")
	}

	if peerPool.Address != params.Address {
		return utils.BYTE_FALSE, errors.NewErr("[RegisterCandidate] Peer is not registered by this address!")
	}
	if peerPool.Status != SyncNodeStatus {
		return utils.BYTE_FALSE, errors.NewErr("[RegisterCandidate] Peer status is not SyncNodeStatus!")
	}

	peerPool.Status = RegisterCandidateStatus

	peerPoolMap.PeerPoolMap[params.PeerPubkey] = peerPool
	value, err := json.Marshal(peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Marshal] Marshal peerPoolMap error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), view.Bytes()), &cstates.StorageItem{Value: value})

	addressBytes, err := hex.DecodeString(params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[hex.DecodeString] Address format error!")
	}
	address, err := common.AddressParseFromBytes(addressBytes)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[common.AddressParseFromBytes] Address format error!")
	}
	//ong transfer
	err = AppCallTransferOng(native, address, genesis.FeeSplitContractAddress, CANDIDATE_FEE)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[appCallTransferOng] Ong transfer error!")
	}

	utils.AddCommonEvent(native, contract, REGISTER_CANDIDATE, params)

	return utils.BYTE_TRUE, nil
}

func ApproveCandidate(native *native.NativeService) ([]byte, error) {
	params := new(ApproveCandidateParam)
	err := json.Unmarshal(native.Input, params)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Contract params Unmarshal error!")
	}

	//TODO: check witness
	//err = utils.ValidateOwner(native, ADMIN_ADDRESS)
	//if err != nil {
	//	return errors.NewDetailErr(err, errors.ErrNoCode, "[ApproveCandidate] CheckWitness error!")
	//}

	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetView] Get view error!")
	}

	//check if peerPoolMap full
	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetPeerPoolMap] Get peerPoolMap error!")
	}

	num := 0
	for _, peerPool := range peerPoolMap.PeerPoolMap {
		if peerPool.Status == CandidateStatus || peerPool.Status == ConsensusStatus {
			num = num + 1
		}
	}
	if num >= CANDIDATE_NUM {
		return utils.BYTE_FALSE, errors.NewErr("[ApproveCandidate] Num of candidate node is full (7*7)!")
	}

	//get peerPool
	peerPool, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, errors.NewErr("[ApproveCandidate] PeerPubkey is not in peerPoolMap!")
	}

	if peerPool.Status != RegisterCandidateStatus {
		return utils.BYTE_FALSE, errors.NewErr("[ApproveCandidate] Peer status is not RegisterCandidateStatus!")
	}

	peerPool.Status = CandidateStatus
	peerPool.TotalPos = 0

	//check if has index
	peerPubkeyPrefix, err := hex.DecodeString(peerPool.PeerPubkey)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[hex.DecodeString] PeerPubkey format error!")
	}
	indexBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_INDEX), peerPubkeyPrefix))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[native.CloneCache.Get] Get indexBytes error!")
	}
	if indexBytes != nil {
		buf := bytes.NewBuffer(indexBytes.(*cstates.StorageItem).Value)
		peerPool.Index, err = serialization.ReadUint32(buf)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[serialization.ReadUint32] ReadUint32 error!")
		}
	} else {
		//get index
		candidateIndexBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(CANDIDITE_INDEX)))
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[native.CloneCache.Get] Get candidateIndex error!")
		}
		var candidateIndex uint64
		if candidateIndexBytes == nil {
			return utils.BYTE_FALSE, errors.NewErr("[ApproveCandidate] CandidateIndex is not init!")
		} else {
			candidateIndexStore, _ := candidateIndexBytes.(*cstates.StorageItem)
			candidateIndex = new(big.Int).SetBytes(candidateIndexStore.Value).Uint64()
		}
		peerPool.Index = uint32(candidateIndex)

		//update candidateIndex
		newCandidateIndex := candidateIndex + 1
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(CANDIDITE_INDEX)),
			&cstates.StorageItem{Value: new(big.Int).SetUint64(newCandidateIndex).Bytes()})

		buf := new(bytes.Buffer)
		err = serialization.WriteUint32(buf, peerPool.Index)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[serialization.WriteUint32] WriteUint32 error!")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_INDEX), peerPubkeyPrefix), &cstates.StorageItem{Value: buf.Bytes()})
	}
	peerPoolMap.PeerPoolMap[params.PeerPubkey] = peerPool
	value, err := json.Marshal(peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Marshal] Marshal peerPool error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), view.Bytes()), &cstates.StorageItem{Value: value})

	utils.AddCommonEvent(native, contract, APPROVE_CANDIDATE, params)

	return utils.BYTE_TRUE, nil
}

func BlackNode(native *native.NativeService) ([]byte, error) {
	params := new(BlackNodeParam)
	err := json.Unmarshal(native.Input, params)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Contract params Unmarshal error!")
	}

	//TODO: check witness
	//err = utils.ValidateOwner(native, ADMIN_ADDRESS)
	//if err != nil {
	//	return errors.NewDetailErr(err, errors.ErrNoCode, "[BlackNode] CheckWitness error!")
	//}
	contract := native.ContextRef.CurrentContext().ContractAddress

	peerPubkeyPrefix, err := hex.DecodeString(params.PeerPubkey)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[hex.DecodeString] PeerPubkey format error!")
	}
	//put peer into black list
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(BLACK_LIST), peerPubkeyPrefix), &cstates.StorageItem{Value: new(big.Int).SetUint64(1).Bytes()})

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetView] Get view error!")
	}
	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetPeerPoolMap] Get peerPoolMap error!")
	}

	peerPool, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, errors.NewErr("[QuitNode] PeerPubkey is not in peerPoolMap!")
	}

	//change peerPool status
	if peerPool.Status == ConsensusStatus {
		peerPool.Status = QuitConsensusStatus
	} else {
		peerPool.Status = QuitingStatus
	}

	peerPoolMap.PeerPoolMap[params.PeerPubkey] = peerPool
	value, err := json.Marshal(peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Marshal] Marshal peerPool error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), view.Bytes()), &cstates.StorageItem{Value: value})

	return utils.BYTE_TRUE, nil
}

func WhiteNode(native *native.NativeService) ([]byte, error) {
	params := new(WhiteNodeParam)
	err := json.Unmarshal(native.Input, params)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Contract params Unmarshal error!")
	}

	//TODO: check witness
	//err = utils.ValidateOwner(native, ADMIN_ADDRESS)
	//if err != nil {
	//	return errors.NewDetailErr(err, errors.ErrNoCode, "[WhiteNode] CheckWitness error!")
	//}
	contract := native.ContextRef.CurrentContext().ContractAddress

	peerPubkeyPrefix, err := hex.DecodeString(params.PeerPubkey)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[hex.DecodeString] PeerPubkey format error!")
	}
	//remove peer from black list
	native.CloneCache.Delete(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(BLACK_LIST), peerPubkeyPrefix))

	return utils.BYTE_TRUE, nil
}

func QuitNode(native *native.NativeService) ([]byte, error) {
	params := new(QuitNodeParam)
	err := json.Unmarshal(native.Input, params)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Marshal] Contract params Unmarshal error!")
	}

	//check witness
	err = utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[ValidateOwner] CheckWitness error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetView] Get view error!")
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetPeerPoolMap] Get peerPoolMap error!")
	}

	peerPool, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, errors.NewErr("[QuitNode] PeerPubkey is not in peerPoolMap!")
	}

	if params.Address != peerPool.Address {
		return utils.BYTE_FALSE, errors.NewErr("[QuitNode] PeerPubkey is not registered by this address!")
	}

	//change peerPool status
	if peerPool.Status == ConsensusStatus {
		peerPool.Status = QuitConsensusStatus
	} else {
		peerPool.Status = QuitingStatus
	}

	peerPoolMap.PeerPoolMap[params.PeerPubkey] = peerPool
	value, err := json.Marshal(peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Marshal] Marshal peerPool error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), view.Bytes()), &cstates.StorageItem{Value: value})

	utils.AddCommonEvent(native, contract, QUIT_NODE, params)

	return utils.BYTE_TRUE, nil
}

func VoteForPeer(native *native.NativeService) ([]byte, error) {
	params := new(VoteForPeerParam)
	err := json.Unmarshal(native.Input, params)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Contract params Unmarshal error!")
	}

	//check witness
	err = utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[ValidateOwner] CheckWitness error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	addressPrefix, err := hex.DecodeString(params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[hex.DecodeString] Address format error!")
	}

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetView] Get view error!")
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetPeerPoolMap] Get peerPoolMap error!")
	}

	var total int64
	for peerPubkey, pos := range params.VoteTable {
		peerPubkeyPrefix, err := hex.DecodeString(peerPubkey)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[hex.DecodeString] PeerPubkey format error!")
		}

		peerPool, ok := peerPoolMap.PeerPoolMap[peerPubkey]
		if !ok {
			return utils.BYTE_FALSE, errors.NewErr("[VoteForPeer] PeerPubkey is not in peerPoolMap!")
		}

		if peerPool.Status != CandidateStatus && peerPool.Status != ConsensusStatus {
			continue
		}

		voteInfoPoolBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL),
			peerPubkeyPrefix, addressPrefix))
		voteInfoPool := &VoteInfoPool{
			PeerPubkey: peerPubkey,
			Address:    params.Address,
		}
		if pos >= 0 {
			if voteInfoPoolBytes != nil {
				voteInfoPoolStore, _ := voteInfoPoolBytes.(*cstates.StorageItem)
				err = json.Unmarshal(voteInfoPoolStore.Value, voteInfoPool)
				if err != nil {
					return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Unmarshal voteInfoPool error!")
				}
				voteInfoPool.NewPos = voteInfoPool.NewPos + uint64(pos)
			} else {
				voteInfoPool.NewPos = uint64(pos)
			}
			total = total + pos
			peerPool.TotalPos = peerPool.TotalPos + uint64(pos)
		} else {
			if voteInfoPoolBytes != nil {
				voteInfoPoolStore, _ := voteInfoPoolBytes.(*cstates.StorageItem)
				err = json.Unmarshal(voteInfoPoolStore.Value, voteInfoPool)
				if err != nil {
					return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Unmarshal voteInfoPool error!")
				}
				temp := int64(voteInfoPool.NewPos) + pos
				if temp < 0 {
					prePos := int64(voteInfoPool.PrePos) + temp
					if prePos < 0 {
						voteInfoPool.PrePos = 0
						voteInfoPool.PreFreezePos = voteInfoPool.PreFreezePos + voteInfoPool.PrePos
						total = total - int64(voteInfoPool.NewPos)
						peerPool.TotalPos = peerPool.TotalPos - voteInfoPool.NewPos
						voteInfoPool.NewPos = 0
					} else {
						voteInfoPool.PrePos = uint64(prePos)
						voteInfoPool.PreFreezePos = uint64(int64(voteInfoPool.PreFreezePos) - temp)
						total = total - int64(voteInfoPool.NewPos)
						peerPool.TotalPos = peerPool.TotalPos - voteInfoPool.NewPos
						voteInfoPool.NewPos = 0
					}
				} else {
					voteInfoPool.NewPos = uint64(temp)
					total = total + pos
					peerPool.TotalPos = uint64(int64(peerPool.TotalPos) + pos)
				}
			}
		}
		peerPoolMap.PeerPoolMap[peerPubkey] = peerPool
		if voteInfoPool.PrePos == 0 && voteInfoPool.PreFreezePos == 0 && voteInfoPool.FreezePos == 0 && voteInfoPool.NewPos == 0 {
			native.CloneCache.Delete(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix, addressPrefix))
		} else {
			value, err := json.Marshal(voteInfoPool)
			if err != nil {
				return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Marshal] Marshal voteInfoPool error")
			}
			native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix,
				addressPrefix), &cstates.StorageItem{Value: value})
		}
	}
	value, err := json.Marshal(peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Marshal] Marshal peerPoolMap error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), view.Bytes()), &cstates.StorageItem{Value: value})

	addressBytes, err := hex.DecodeString(params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[hex.DecodeString] Address format error!")
	}
	address, err := common.AddressParseFromBytes(addressBytes)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[common.AddressParseFromBytes] Address format error!")
	}

	if total > 0 {
		//ont transfer
		err = AppCallTransferOnt(native, address, genesis.GovernanceContractAddress, uint64(total))
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[appCallTransferOnt] Ont transfer error!")
		}
	}
	if total < 0 {
		//ont transfer
		err = AppCallTransferOnt(native, genesis.GovernanceContractAddress, address, uint64(-total))
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[appCallTransferOnt] Ont transfer error!")
		}
	}

	utils.AddCommonEvent(native, contract, VOTE_FOR_PEER, params)

	return utils.BYTE_TRUE, nil
}

func CommitDpos(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	//get governace view
	governanceView, err := GetGovernanceView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetGovernanceView] Get GovernanceView error!")
	}

	// get config
	config := new(Configuration)
	configBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VBFT_CONFIG)))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[native.CloneCache.Get] Get configBytes error!")
	}
	if configBytes == nil {
		return utils.BYTE_FALSE, errors.NewErr("[CommitDpos] ConfigBytes is nil!")
	}
	configStore, _ := configBytes.(*cstates.StorageItem)
	err = json.Unmarshal(configStore.Value, config)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Unmarshal config error!")
	}

	//TODO: check witness
	//err = utils.ValidateOwner(native, ADMIN_ADDRESS)
	//if err != nil {
	//	cycle := native.Height % config.MaxBlockChangeView == 0
	//	if !cycle && !governanceView.VoteCommit {
	//		return utils.BYTE_FALSE, errors.NewErr("[CommitDpos] Authentication Failed!")
	//	}
	//}

	//get current view
	view := governanceView.View
	newView := new(big.Int).Add(view, new(big.Int).SetInt64(1))

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetPeerPoolMap] Get peerPoolMap error!")
	}

	peers := []*PeerStakeInfo{}
	for _, peerPool := range peerPoolMap.PeerPoolMap {
		peerPubkeyPrefix, err := hex.DecodeString(peerPool.PeerPubkey)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[hex.DecodeString] PeerPubkey format error!")
		}

		if peerPool.Status == QuitingStatus {
			//draw back init pos
			addressBytes, err := hex.DecodeString(peerPool.Address)
			if err != nil {
				return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[hex.DecodeString] Address format error!")
			}
			address, err := common.AddressParseFromBytes(addressBytes)
			if err != nil {
				return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[common.AddressParseFromBytes] Address format error!")
			}
			//ont transfer
			err = AppCallTransferOnt(native, genesis.GovernanceContractAddress, address, peerPool.InitPos)
			if err != nil {
				return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[appCallTransferOnt] Ont transfer error!")
			}

			//draw back vote pos
			stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix))
			if err != nil {
				return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[native.CloneCache.Store.Find] Get all peerPool error!")
			}
			voteInfoPool := new(VoteInfoPool)
			for _, v := range stateValues {
				voteInfoPoolStore, _ := v.Value.(*cstates.StorageItem)
				err = json.Unmarshal(voteInfoPoolStore.Value, voteInfoPool)
				if err != nil {
					return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Unmarshal voteInfoPool error!")
				}
				pos := voteInfoPool.PrePos + voteInfoPool.PreFreezePos + voteInfoPool.FreezePos + voteInfoPool.NewPos

				addressBytes, err := hex.DecodeString(voteInfoPool.Address)
				if err != nil {
					return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[hex.DecodeString] Address format error!")
				}
				address, err := common.AddressParseFromBytes(addressBytes)
				if err != nil {
					return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[common.AddressParseFromBytes] Address format error!")
				}
				//ont transfer
				err = AppCallTransferOnt(native, genesis.GovernanceContractAddress, address, pos)
				if err != nil {
					return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[appCallTransferOnt] Ont transfer error!")
				}
				native.CloneCache.Delete(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix, addressBytes))
			}
			delete(peerPoolMap.PeerPoolMap, peerPool.PeerPubkey)
		}
		if peerPool.Status == QuitConsensusStatus {
			peerPool.Status = QuitingStatus
			peerPoolMap.PeerPoolMap[peerPool.PeerPubkey] = peerPool
		}

		if peerPool.Status == CandidateStatus || peerPool.Status == ConsensusStatus {
			stake := peerPool.TotalPos + peerPool.InitPos
			peers = append(peers, &PeerStakeInfo{
				Index:      peerPool.Index,
				PeerPubkey: peerPool.PeerPubkey,
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
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[hex.DecodeString] PeerPubkey format error!")
		}

		peerPool, ok := peerPoolMap.PeerPoolMap[peers[i].PeerPubkey]
		if !ok {
			return utils.BYTE_FALSE, errors.NewErr("[VoteForPeer] PeerPubkey is not in peerPoolMap!")
		}

		if peerPool.Status == ConsensusStatus {
			//update voteInfoPool
			stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix))
			if err != nil {
				return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[native.CloneCache.Store.Find] Get all peerPool error!")
			}
			voteInfoPool := new(VoteInfoPool)
			for _, v := range stateValues {
				voteInfoPoolStore, _ := v.Value.(*cstates.StorageItem)
				err = json.Unmarshal(voteInfoPoolStore.Value, voteInfoPool)
				if err != nil {
					return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Unmarshal voteInfoPool error!")
				}
				addressPrefix, err := hex.DecodeString(voteInfoPool.Address)
				if err != nil {
					errors.NewDetailErr(err, errors.ErrNoCode, "[hex.DecodeString] Address format error!")
				}
				freezePos := voteInfoPool.FreezePos
				newPos := voteInfoPool.NewPos
				preFreezePos := voteInfoPool.PreFreezePos
				voteInfoPool.PrePos = voteInfoPool.PrePos + newPos
				voteInfoPool.NewPos = freezePos
				voteInfoPool.FreezePos = preFreezePos
				voteInfoPool.PreFreezePos = 0

				value, err := json.Marshal(voteInfoPool)
				if err != nil {
					return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Marshal] Marshal voteInfoPool error")
				}
				native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix,
					addressPrefix), &cstates.StorageItem{Value: value})
			}
		} else {
			//update voteInfoPool
			stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix))
			if err != nil {
				return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[native.CloneCache.Store.Find] Get all peerPool error!")
			}
			voteInfoPool := new(VoteInfoPool)
			for _, v := range stateValues {
				voteInfoPoolStore, _ := v.Value.(*cstates.StorageItem)
				err = json.Unmarshal(voteInfoPoolStore.Value, voteInfoPool)
				if err != nil {
					return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Unmarshal voteInfoPool error!")
				}
				addressPrefix, err := hex.DecodeString(voteInfoPool.Address)
				if err != nil {
					return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[hex.DecodeString] Address format error!")
				}
				preFreezePos := voteInfoPool.PreFreezePos
				freezePos := voteInfoPool.FreezePos
				voteInfoPool.PrePos = voteInfoPool.NewPos
				voteInfoPool.NewPos = preFreezePos + freezePos
				voteInfoPool.PreFreezePos = 0
				voteInfoPool.FreezePos = 0

				value, err := json.Marshal(voteInfoPool)
				if err != nil {
					return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Marshal] Marshal voteInfoPool error")
				}
				native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix,
					addressPrefix), &cstates.StorageItem{Value: value})
			}
		}
		peerPool.Status = ConsensusStatus
		peerPoolMap.PeerPoolMap[peers[i].PeerPubkey] = peerPool
	}

	//non consensus peers
	for i := int(config.K); i < len(peers); i++ {
		//change peerPool status
		peerPubkeyPrefix, err := hex.DecodeString(peers[i].PeerPubkey)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[hex.DecodeString] PeerPubkey format error!")
		}

		peerPool, ok := peerPoolMap.PeerPoolMap[peers[i].PeerPubkey]
		if !ok {
			return utils.BYTE_FALSE, errors.NewErr("[VoteForPeer] PeerPubkey is not in peerPoolMap!")
		}

		if peerPool.Status == ConsensusStatus {
			//update voteInfoPool
			stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix))
			if err != nil {
				return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[native.CloneCache.Store.Find] Get all peerPool error!")
			}
			voteInfoPool := new(VoteInfoPool)
			for _, v := range stateValues {
				voteInfoPoolStore, _ := v.Value.(*cstates.StorageItem)
				err = json.Unmarshal(voteInfoPoolStore.Value, voteInfoPool)
				if err != nil {
					return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Unmarshal voteInfoPool error!")
				}
				addressPrefix, err := hex.DecodeString(voteInfoPool.Address)
				if err != nil {
					errors.NewDetailErr(err, errors.ErrNoCode, "[hex.DecodeString] Address format error!")
				}
				prePos := voteInfoPool.PrePos
				freezePos := voteInfoPool.FreezePos
				preFreezePos := voteInfoPool.PreFreezePos
				newPos := voteInfoPool.NewPos
				voteInfoPool.NewPos = freezePos
				voteInfoPool.FreezePos = newPos + prePos + preFreezePos
				voteInfoPool.PrePos = 0
				voteInfoPool.PreFreezePos = 0

				value, err := json.Marshal(voteInfoPool)
				if err != nil {
					return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Marshal] Marshal voteInfoPool error")
				}
				native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix,
					addressPrefix), &cstates.StorageItem{Value: value})
			}
		} else {
			//update voteInfoPool
			stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix))
			if err != nil {
				return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[native.CloneCache.Store.Find] Get all peerPool error!")
			}
			voteInfoPool := new(VoteInfoPool)
			for _, v := range stateValues {
				voteInfoPoolStore, _ := v.Value.(*cstates.StorageItem)
				err = json.Unmarshal(voteInfoPoolStore.Value, voteInfoPool)
				if err != nil {
					return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Unmarshal voteInfoPool error!")
				}
				addressPrefix, err := hex.DecodeString(voteInfoPool.Address)
				if err != nil {
					return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[hex.DecodeString] Address format error!")
				}
				preFreezePos := voteInfoPool.PreFreezePos
				newPos := voteInfoPool.NewPos
				freezePos := voteInfoPool.FreezePos
				voteInfoPool.NewPos = freezePos
				voteInfoPool.FreezePos = newPos + preFreezePos
				voteInfoPool.PreFreezePos = 0

				value, err := json.Marshal(voteInfoPool)
				if err != nil {
					return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Marshal] Marshal voteInfoPool error")
				}
				native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix,
					addressPrefix), &cstates.StorageItem{Value: value})
			}
		}
		peerPool.Status = CandidateStatus
		peerPoolMap.PeerPoolMap[peers[i].PeerPubkey] = peerPool
	}
	value, err := json.Marshal(peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Marshal] Marshal peerPoolMap error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), newView.Bytes()), &cstates.StorageItem{Value: value})
	oldView := new(big.Int).Sub(view, new(big.Int).SetUint64(1))
	native.CloneCache.Delete(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), oldView.Bytes()))

	//get all vote for commit info
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_COMMIT_INFO), view.Bytes()))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[native.CloneCache.Store.Find] Get all peerPool error!")
	}

	voteCommitInfoPool := new(VoteCommitInfoPool)
	for _, v := range stateValues {
		voteCommitInfoPoolStore, _ := v.Value.(*cstates.StorageItem)
		err = json.Unmarshal(voteCommitInfoPoolStore.Value, voteCommitInfoPool)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Unmarshal voteCommitInfoPool error!")
		}

		addressBytes, err := hex.DecodeString(voteCommitInfoPool.Address)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[hex.DecodeString] Address format error!")
		}
		address, err := common.AddressParseFromBytes(addressBytes)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[common.AddressParseFromBytes] Address format error!")
		}
		//ont transfer
		err = AppCallTransferOnt(native, genesis.GovernanceContractAddress, address, voteCommitInfoPool.Pos)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[appCallTransferOnt] Ont transfer error!")
		}
	}

	//posTable, chainPeers, err := calDposTable(native, config, peers)

	//update view
	governanceView = &GovernanceView{
		View:       newView,
		VoteCommit: false,
	}
	v, err := json.Marshal(governanceView)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Marshal] Marshal governanceView error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(GOVERNANCE_VIEW)), &cstates.StorageItem{Value: v})

	utils.AddCommonEvent(native, contract, COMMIT_DPOS, true)

	return utils.BYTE_TRUE, nil
}

func VoteCommitDpos(native *native.NativeService) ([]byte, error) {
	params := new(VoteCommitDposParam)
	err := json.Unmarshal(native.Input, params)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Contract params Unmarshal error!")
	}

	//check witness
	err = utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[ValidateOwner] CheckWitness error!")
	}

	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetView] Get view error!")
	}

	addressPrefix, err := hex.DecodeString(params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[hex.DecodeString] Address format error!")
	}

	//get voteCommitInfo
	voteCommitInfoPool := new(VoteCommitInfoPool)
	voteCommitInfoPoolBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_COMMIT_INFO), view.Bytes(), addressPrefix))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[native.CloneCache.Get] Get voteCommitInfoBytes error!")
	}
	if voteCommitInfoPoolBytes != nil {
		voteCommitInfoPoolStore, _ := voteCommitInfoPoolBytes.(*cstates.StorageItem)
		err = json.Unmarshal(voteCommitInfoPoolStore.Value, voteCommitInfoPool)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Unmarshal voteCommitInfoPool error!")
		}
	}
	pos := int64(voteCommitInfoPool.Pos) + params.Pos
	if pos < 0 {
		return utils.BYTE_FALSE, errors.NewErr("[VoteCommitDpos] Remain pos is negative!")
	}
	voteCommitInfoPool.Pos = uint64(pos)
	v, err := json.Marshal(voteCommitInfoPool)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Marshal] Marshal governanceView error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_COMMIT_INFO), view.Bytes(), addressPrefix), &cstates.StorageItem{Value: v})

	//get total pos for commit
	posCommitBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(POS_FOR_COMMIT), view.Bytes()))
	posCommit := new(big.Int)
	if posCommitBytes != nil {
		posCommitStore, _ := posCommitBytes.(*cstates.StorageItem)
		posCommit = new(big.Int).SetBytes(posCommitStore.Value)
	}
	newPosCommit := posCommit.Int64() + params.Pos

	if newPosCommit >= POS_COMMIT_TRIGGER {
		governanceView := &GovernanceView{
			View:       view,
			VoteCommit: true,
		}
		v, err := json.Marshal(governanceView)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Marshal] Marshal governanceView error")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(GOVERNANCE_VIEW)), &cstates.StorageItem{Value: v})
	}

	addressBytes, err := hex.DecodeString(params.Address)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[hex.DecodeString] Address format error!")
	}
	address, err := common.AddressParseFromBytes(addressBytes)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[common.AddressParseFromBytes] Address format error!")
	}

	//ont transfer
	if params.Pos > 0 {
		err = AppCallTransferOnt(native, address, genesis.GovernanceContractAddress, uint64(params.Pos))
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[appCallTransferOnt] Ont transfer error!")
		}
	}
	if params.Pos < 0 {
		err = AppCallTransferOnt(native, genesis.GovernanceContractAddress, address, uint64(-params.Pos))
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[appCallTransferOnt] Ont transfer error!")
		}
	}

	return utils.BYTE_TRUE, nil
}

func UpdateConfig(native *native.NativeService) ([]byte, error) {
	configuration := new(Configuration)
	err := json.Unmarshal(native.Input, configuration)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Contract params Unmarshal error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//check the configuration
	if configuration.L < 16*configuration.K {
		return utils.BYTE_FALSE, errors.NewErr("[UpdateConfig] L is less than 16*K in config!")
	}

	value, err := json.Marshal(configuration)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Marshal] Marshal config error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VBFT_CONFIG)), &cstates.StorageItem{Value: value})

	utils.AddCommonEvent(native, contract, UPDATE_CONFIG, configuration)

	return utils.BYTE_TRUE, nil
}

func DataQuery(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetView] Get view error!")
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetPeerPoolMap] Get peerPoolMap error!")
	}

	for _, peerPool := range peerPoolMap.PeerPoolMap {
		fmt.Println("PeerPool is : ", peerPool)
	}

	fmt.Println("view :", view)
	//update voteInfoPool
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL)))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[native.CloneCache.Store.Find] Get all peerPool error!")
	}
	voteInfoPool := new(VoteInfoPool)
	for _, v := range stateValues {
		voteInfoPoolStore, _ := v.Value.(*cstates.StorageItem)
		err = json.Unmarshal(voteInfoPoolStore.Value, voteInfoPool)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[json.Unmarshal] Unmarshal voteInfoPool error!")
		}
		fmt.Println("VoteInfoPool is : ", voteInfoPool)
	}

	return utils.BYTE_TRUE, nil
}
