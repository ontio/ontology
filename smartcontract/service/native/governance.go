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
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"sort"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/genesis"
	cstates "github.com/ontio/ontology/core/states"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native/states"
)

const (
	//status
	RegisterSyncNodeStatus states.Status = iota
	SyncNodeStatus
	RegisterCandidateStatus
	CandidateStatus
	ConsensusStatus
	QuitStatus
	QuitConsensusStatus
)

const (
	//function name
	INIT_CONFIG        = "initConfig"
	REGISTER_SYNC_NODE = "registerSyncNode"
	APPROVE_SYNC_NODE  = "approveSyncNode"
	REGISTER_CANDIDATE = "registerCandidate"
	APPROVE_CANDIDATE  = "approveCandidate"
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

	//global
	SYNC_NODE_FEE  = 50
	CANDIDATE_FEE  = 500
	MIN_INIT_STAKE = 1000
	CandidateNum   = 7 * 7
	SyncNodeNum    = 7 * 7 * 7
)

func init() {
	Contracts[genesis.GovernanceContractAddress] = RegisterGovernanceContract
}

func RegisterGovernanceContract(native *NativeService) {
	native.Register(INIT_CONFIG, InitConfig)
	native.Register(REGISTER_SYNC_NODE, RegisterSyncNode)
	native.Register(APPROVE_SYNC_NODE, ApproveSyncNode)
	native.Register(REGISTER_CANDIDATE, RegisterCandidate)
	native.Register(APPROVE_CANDIDATE, ApproveCandidate)
	native.Register(QUIT_NODE, QuitNode)
	native.Register(VOTE_FOR_PEER, VoteForPeer)
	native.Register(COMMIT_DPOS, CommitDpos)
	native.Register(VOTE_COMMIT_DPOS, VoteCommitDpos)
	native.Register(UPDATE_CONFIG, UpdateConfig)
	native.Register("dataQuery", DataQuery)
}

func InitConfig(native *NativeService) error {
	consensusConfigFile := config.Parameters.ConsensusConfigPath
	// load dpos config
	file, err := ioutil.ReadFile(consensusConfigFile)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[initConfig] Failed to open config file!")
	}
	file = bytes.TrimPrefix(file, []byte("\xef\xbb\xbf"))

	configuration := new(states.Configuration)
	err = json.Unmarshal(file, configuration)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[initConfig] Contract params Unmarshal error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//check the configuration
	if configuration.L < 16*configuration.K {
		return errors.NewErr("[initConfig] L is less than 16*K in config!")
	}

	view := new(big.Int).SetInt64(1)

	value, err := json.Marshal(configuration)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[initConfig] Marshal candidatePool error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(VBFT_CONFIG)), &cstates.StorageItem{Value: value})

	initPeerPool := &states.PeerPoolList{}
	if err := json.Unmarshal(file, initPeerPool); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[initConfig] Failed to unmarshal config file!")
	}

	indexMap := make(map[uint32]struct{})
	var maxId uint32
	peers := []*states.PeerStakeInfo{}
	for _, peerPool := range initPeerPool.Peers {
		_, ok := indexMap[peerPool.Index]
		if ok {
			return errors.NewErr("[initConfig] Peer index is duplicated!")
		}
		indexMap[peerPool.Index] = struct{}{}
		if peerPool.Index <= 0 {
			return errors.NewErr("[initConfig] Peer index in config must > 0!")
		}
		if peerPool.Index > maxId {
			maxId = peerPool.Index
		}
		peerPool.TotalPos = 0
		peerPool.Status = ConsensusStatus
		value, err := json.Marshal(peerPool)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[initConfig] Marshal candidatePool error!")
		}
		peerPubkeyPrefix, err := hex.DecodeString(peerPool.PeerPubkey)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[initConfig] PeerPubkey format error!")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), new(big.Int).Bytes(), peerPubkeyPrefix),
			&cstates.StorageItem{Value: value})
		native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), view.Bytes(), peerPubkeyPrefix),
			&cstates.StorageItem{Value: value})

		peers = append(peers, &states.PeerStakeInfo{
			Index:      peerPool.Index,
			PeerPubkey: peerPool.PeerPubkey,
			Stake:      peerPool.InitPos,
		})
	}

	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(CANDIDITE_INDEX)),
		&cstates.StorageItem{Value: new(big.Int).SetUint64(uint64(maxId + 1)).Bytes()})

	governanceView := &states.GovernanceView{
		View:       view,
		VoteCommit: false,
	}
	v, err := json.Marshal(governanceView)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[initConfig] Marshal governanceView error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(GOVERNANCE_VIEW)), &cstates.StorageItem{Value: v})

	posTable, chainPeers, err := calDposTable(native, configuration, peers)
	fmt.Println("ChainPeers is :", chainPeers)
	fmt.Println("DPOS table is:", posTable)

	addCommonEvent(native, contract, INIT_CONFIG, true)
	return nil
}

func RegisterSyncNode(native *NativeService) error {
	params := new(states.RegisterSyncNodeParam)
	err := json.Unmarshal(native.Input, params)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[registerSyncNode] Contract params Unmarshal error!")
	}

	//check initPos
	if params.InitPos < MIN_INIT_STAKE {
		return errors.NewErr(fmt.Sprintf("[registerSyncNode] InitPos must >= %v!", MIN_INIT_STAKE))
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

	//check if PeerPool full
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), view.Bytes()))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[registerSyncNode] Get all peerPool error!")
	}
	fmt.Println("Num of sync node is :", len(stateValues))
	if len(stateValues) >= SyncNodeNum {
		return errors.NewErr("[registerSyncNode] Sync node is full (7*7*7)!")
	}

	//check if exist in PeerPool
	v1, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), view.Bytes(), peerPubkeyPrefix))
	if v1 != nil {
		return errors.NewErr("[registerSyncNode] PeerPubkey is already in peerPool!")
	}

	peerPool := &states.PeerPool{
		PeerPubkey: params.PeerPubkey,
		Address:    params.Address,
		InitPos:    params.InitPos,
		Status:     RegisterSyncNodeStatus,
	}
	value, err := json.Marshal(peerPool)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[registerSyncNode] Marshal peerPool error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), view.Bytes(), peerPubkeyPrefix), &cstates.StorageItem{Value: value})

	addressBytes, err := hex.DecodeString(params.Address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[registerSyncNode] Address format error!")
	}
	address, err := common.AddressParseFromBytes(addressBytes)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[registerSyncNode] Address format error!")
	}
	//ont transfer
	err = appCallTransferOnt(native, address, genesis.GovernanceContractAddress, new(big.Int).SetUint64(params.InitPos))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[registerSyncNode] Ont transfer error!")
	}
	//ong transfer
	err = appCallTransferOng(native, address, genesis.GovernanceContractAddress, new(big.Int).SetInt64(SYNC_NODE_FEE))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[registerSyncNode] Ong transfer error!")
	}

	addCommonEvent(native, contract, REGISTER_SYNC_NODE, params)

	return nil
}

func ApproveSyncNode(native *NativeService) error {
	params := new(states.ApproveSyncNodeParam)
	err := json.Unmarshal(native.Input, params)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[approveSyncNode] Contract params Unmarshal error!")
	}

	//TODO: check witness
	//err = validateOwner(native, params.Address)
	//if err != nil {
	//	return errors.NewDetailErr(err, errors.ErrNoCode, "[registerCandidate] CheckWitness error!")
	//}

	contract := native.ContextRef.CurrentContext().ContractAddress
	peerPubkeyPrefix, err := hex.DecodeString(params.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[approveSyncNode] PeerPubkey format error!")
	}

	//get current view
	view, err := getGovernanceView(native, contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[approveSyncNode] Get view error!")
	}

	//get peerPool
	peerPoolBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), view.Bytes(), peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[approveSyncNode] Get peerPoolBytes error!")
	}
	peerPool := new(states.PeerPool)
	if peerPoolBytes != nil {
		peerPoolStore, _ := peerPoolBytes.(*cstates.StorageItem)
		err := json.Unmarshal(peerPoolStore.Value, peerPool)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[approveSyncNode] Unmarshal peerPool error!")
		}
	} else {
		return errors.NewErr("[approveSyncNode] PeerPubkey is not in peerPool!")
	}

	if peerPool.Status != RegisterSyncNodeStatus {
		return errors.NewErr("[approveSyncNode] Peer status is not RegisterSyncNodeStatus!")
	}

	peerPool.Status = SyncNodeStatus

	value, err := json.Marshal(peerPool)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[approveSyncNode] Marshal peerPool error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), view.Bytes(), peerPubkeyPrefix), &cstates.StorageItem{Value: value})

	addCommonEvent(native, contract, APPROVE_CANDIDATE, params)

	return nil
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

	//get current view
	view, err := getGovernanceView(native, contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[registerCandidate] Get view error!")
	}

	//syncNodePool storage
	//check syncNodePool
	peerPoolBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), view.Bytes(), peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[registerCandidate] Get peerPoolBytes error!")
	}
	if peerPoolBytes == nil {
		return errors.NewErr("[registerCandidate] PeerPubkey is not in peerPool!")
	}
	peerPool := new(states.PeerPool)
	peerPoolStore, _ := peerPoolBytes.(*cstates.StorageItem)
	err = json.Unmarshal(peerPoolStore.Value, peerPool)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[registerCandidate] Unmarshal peerPool error!")
	}
	if peerPool.Address != params.Address {
		return errors.NewErr("[registerCandidate] Peer is not registered by this address!")
	}
	if peerPool.Status != SyncNodeStatus {
		return errors.NewErr("[registerCandidate] Peer status is not SyncNodeStatus!")
	}

	peerPool.Status = RegisterCandidateStatus

	value, err := json.Marshal(peerPool)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[registerCandidate] Marshal syncNodePool error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), view.Bytes(), peerPubkeyPrefix), &cstates.StorageItem{Value: value})

	addressBytes, err := hex.DecodeString(params.Address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[registerCandidate] Address format error!")
	}
	address, err := common.AddressParseFromBytes(addressBytes)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[registerCandidate] Address format error!")
	}
	//ong transfer
	err = appCallTransferOng(native, address, genesis.GovernanceContractAddress, new(big.Int).SetInt64(CANDIDATE_FEE))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[registerCandidate] Ong transfer error!")
	}

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

	//check if PeerPool full
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), view.Bytes()))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[approveCandidate] Get all peerPool error!")
	}
	peerPool := new(states.PeerPool)
	num := 0
	for _, v := range stateValues {
		peerPoolStore, _ := v.Value.(*cstates.StorageItem)
		err = json.Unmarshal(peerPoolStore.Value, peerPool)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[approveCandidate] Unmarshal peerPool error!")
		}
		if peerPool.Status == CandidateStatus || peerPool.Status == ConsensusStatus {
			num = num + 1
		}
	}
	fmt.Println("Num of candidate node is :", num)
	if num >= CandidateNum {
		return errors.NewErr("[approveCandidate] Num of candidate node is full (7*7)!")
	}

	//get peerPool
	peerPoolBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), view.Bytes(), peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[approveCandidate] Get peerPoolBytes error!")
	}
	peerPool = new(states.PeerPool)
	if peerPoolBytes != nil {
		peerPoolStore, _ := peerPoolBytes.(*cstates.StorageItem)
		err := json.Unmarshal(peerPoolStore.Value, peerPool)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[approveCandidate] Unmarshal peerPool error!")
		}
	} else {
		return errors.NewErr("[approveCandidate] PeerPubkey is not in peerPool!")
	}

	if peerPool.Status != RegisterCandidateStatus {
		return errors.NewErr("[approveCandidate] Peer status is not RegisterCandidateStatus!")
	}

	peerPool.Status = CandidateStatus
	peerPool.TotalPos = 0

	//get index
	candidateIndexBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(CANDIDITE_INDEX)))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[approveCandidate] Get candidateIndex error!")
	}
	var candidateIndex uint64
	if candidateIndexBytes == nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[approveCandidate] CandidateIndex is not init!")
	} else {
		candidateIndexStore, _ := candidateIndexBytes.(*cstates.StorageItem)
		candidateIndex = new(big.Int).SetBytes(candidateIndexStore.Value).Uint64()
	}

	peerPool.Index = uint32(candidateIndex)

	value, err := json.Marshal(peerPool)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[approveCandidate] Marshal peerPool error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), view.Bytes(), peerPubkeyPrefix),
		&cstates.StorageItem{Value: value})

	//update candidateIndex
	newCandidateIndex := candidateIndex + 1
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(CANDIDITE_INDEX)),
		&cstates.StorageItem{Value: new(big.Int).SetUint64(newCandidateIndex).Bytes()})

	addCommonEvent(native, contract, APPROVE_CANDIDATE, params)

	return nil
}

func QuitNode(native *NativeService) error {
	params := new(states.QuitNodeParam)
	err := json.Unmarshal(native.Input, params)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[quitNode] Contract params Unmarshal error!")
	}

	//check witness
	err = validateOwner(native, params.Address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[quitNode] CheckWitness error!")
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	peerPubkeyPrefix, err := hex.DecodeString(params.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[quitNode] PeerPubkey format error!")
	}

	//get current view
	view, err := getGovernanceView(native, contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[quitNode] Get view error!")
	}

	//get peerPool
	peerPoolBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), view.Bytes(), peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[quitNode] Get peerPoolBytes error!")
	}
	peerPool := new(states.PeerPool)
	if peerPoolBytes != nil {
		peerPoolStore, _ := peerPoolBytes.(*cstates.StorageItem)
		err := json.Unmarshal(peerPoolStore.Value, peerPool)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[quitNode] Unmarshal peerPool error!")
		}
	} else {
		return errors.NewErr("[quitNode] PeerPubkey is not in peerPool!")
	}
	if params.Address != peerPool.Address {
		return errors.NewErr("[quitNode] PeerPubkey is not registered by this address!")
	}

	//change peerPool status
	if peerPool.Status == ConsensusStatus {
		peerPool.Status = QuitConsensusStatus
	} else {
		peerPool.Status = QuitStatus
	}
	peerPool.QuitView = view

	value, err := json.Marshal(peerPool)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[quitNode] Marshal peerPool error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), view.Bytes(), peerPubkeyPrefix), &cstates.StorageItem{Value: value})

	addCommonEvent(native, contract, QUIT_NODE, params)

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

	var total int64
	for peerPubkey, pos := range params.VoteTable {
		peerPubkeyPrefix, err := hex.DecodeString(peerPubkey)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[voteForPeer] PeerPubkey format error!")
		}

		//get peerPool
		peerPoolBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), view.Bytes(), peerPubkeyPrefix))
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[voteForPeer] Get peerPoolBytes error!")
		}
		peerPool := new(states.PeerPool)
		if peerPoolBytes == nil {
			continue
		}
		peerPoolStore, _ := peerPoolBytes.(*cstates.StorageItem)
		err = json.Unmarshal(peerPoolStore.Value, peerPool)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[voteForPeer] Unmarshal peerPool error!")
		}
		if peerPool.Status != CandidateStatus && peerPool.Status != ConsensusStatus {
			continue
		}

		voteInfoPoolBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(VOTE_INFO_POOL),
			view.Bytes(), peerPubkeyPrefix, addressPrefix))
		voteInfoPool := &states.VoteInfoPool{
			PeerPubkey: peerPubkey,
			Address:    params.Address,
		}
		if pos >= 0 {
			if voteInfoPoolBytes != nil {
				voteInfoPoolStore, _ := voteInfoPoolBytes.(*cstates.StorageItem)
				err = json.Unmarshal(voteInfoPoolStore.Value, voteInfoPool)
				if err != nil {
					return errors.NewDetailErr(err, errors.ErrNoCode, "[voteForPeer] Unmarshal voteInfoPool error!")
				}
				voteInfoPool.NewPos = voteInfoPool.NewPos + uint64(pos)
			} else {
				voteInfoPool.NewPos = uint64(pos)
			}
			total = total + pos
			peerPool.TotalPos = peerPool.TotalPos + uint64(pos)
			value, err := json.Marshal(voteInfoPool)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[quitSyncNode] Marshal voteInfoPool error")
			}
			native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(VOTE_INFO_POOL), view.Bytes(),
				peerPubkeyPrefix, addressPrefix), &cstates.StorageItem{Value: value})

			value, err = json.Marshal(peerPool)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[voteForPeer] Marshal peerPool error")
			}
			native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), view.Bytes(), peerPubkeyPrefix),
				&cstates.StorageItem{Value: value})
		} else {
			if voteInfoPoolBytes != nil {
				voteInfoPoolStore, _ := voteInfoPoolBytes.(*cstates.StorageItem)
				err = json.Unmarshal(voteInfoPoolStore.Value, voteInfoPool)
				if err != nil {
					return errors.NewDetailErr(err, errors.ErrNoCode, "[voteForPeer] Unmarshal voteInfoPool error!")
				}
				temp := int64(voteInfoPool.NewPos) + pos
				if temp < 0 {
					prePos := int64(voteInfoPool.PrePos) + temp
					if prePos < 0 {
						continue
					}
					voteInfoPool.PrePos = uint64(prePos)
					voteInfoPool.PreFreezePos = uint64(int64(voteInfoPool.PreFreezePos) - temp)
					total = total - int64(voteInfoPool.NewPos)
					peerPool.TotalPos = peerPool.TotalPos - voteInfoPool.NewPos
					voteInfoPool.NewPos = 0
				} else {
					voteInfoPool.NewPos = uint64(temp)
					total = total + pos
					peerPool.TotalPos = uint64(int64(peerPool.TotalPos) + pos)
				}

				value, err := json.Marshal(voteInfoPool)
				if err != nil {
					return errors.NewDetailErr(err, errors.ErrNoCode, "[quitSyncNode] Marshal voteInfoPool error")
				}
				native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(VOTE_INFO_POOL), view.Bytes(),
					peerPubkeyPrefix, addressPrefix), &cstates.StorageItem{Value: value})

				value, err = json.Marshal(peerPool)
				if err != nil {
					return errors.NewDetailErr(err, errors.ErrNoCode, "[voteForPeer] Marshal peerPool error")
				}
				native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), view.Bytes(), peerPubkeyPrefix), &cstates.StorageItem{Value: value})
			}
		}
	}
	fmt.Println("Total is :", total)

	addressBytes, err := hex.DecodeString(params.Address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[voteForPeer] Address format error!")
	}
	address, err := common.AddressParseFromBytes(addressBytes)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[voteForPeer] Address format error!")
	}

	if total > 0 {
		//ont transfer
		err = appCallTransferOnt(native, address, genesis.GovernanceContractAddress, new(big.Int).SetInt64(total))
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[voteForPeer] Ont transfer error!")
		}
	}
	if total < 0 {
		//ont transfer
		err = appCallTransferOnt(native, genesis.GovernanceContractAddress, address, new(big.Int).Abs(new(big.Int).SetInt64(total)))
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[voteForPeer] Ont transfer error!")
		}
	}

	addCommonEvent(native, contract, VOTE_FOR_PEER, params)

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

	newView := new(big.Int).Add(view, new(big.Int).SetInt64(1))

	//get all peerPool
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), view.Bytes()))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Get all peerPool error!")
	}

	peers := []*states.PeerStakeInfo{}
	peerPool := new(states.PeerPool)
	for _, v := range stateValues {
		peerPoolStore, _ := v.Value.(*cstates.StorageItem)
		err = json.Unmarshal(peerPoolStore.Value, peerPool)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Unmarshal peerPool error!")
		}
		peerPubkeyPrefix, err := hex.DecodeString(peerPool.PeerPubkey)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] PeerPubkey format error!")
		}

		native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), newView.Bytes(), peerPubkeyPrefix), v.Value)

		if peerPool.Status == QuitStatus {
			quitView := peerPool.QuitView

			//draw back init pos
			addressBytes, err := hex.DecodeString(peerPool.Address)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Address format error!")
			}
			address, err := common.AddressParseFromBytes(addressBytes)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Address format error!")
			}
			//ont transfer
			err = appCallTransferOnt(native, genesis.GovernanceContractAddress, address, new(big.Int).SetUint64(peerPool.InitPos))
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[voteForPeer] Ont transfer error!")
			}
			fmt.Printf("############################## draw back init pos %v, to address %v \n", peerPool.InitPos, peerPool.Address)

			//draw back vote pos
			stateValues, err = native.CloneCache.Store.Find(scommon.ST_STORAGE, concatKey(contract, []byte(VOTE_INFO_POOL), quitView.Bytes(), peerPubkeyPrefix))
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
				pos := voteInfoPool.PrePos + voteInfoPool.PreFreezePos + voteInfoPool.FreezePos + voteInfoPool.NewPos

				addressBytes, err := hex.DecodeString(voteInfoPool.Address)
				if err != nil {
					return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Address format error!")
				}
				address, err := common.AddressParseFromBytes(addressBytes)
				if err != nil {
					return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Address format error!")
				}
				//ont transfer
				err = appCallTransferOnt(native, genesis.GovernanceContractAddress, address, new(big.Int).SetUint64(pos))
				if err != nil {
					return errors.NewDetailErr(err, errors.ErrNoCode, "[voteForPeer] Ont transfer error!")
				}
				fmt.Printf("########################### draw back vote pos %v, to address %v \n", pos, voteInfoPool.Address)
			}

			native.CloneCache.Delete(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), newView.Bytes(), peerPubkeyPrefix))
		}
		if peerPool.Status == QuitConsensusStatus {
			peerPool.Status = QuitStatus
			value, err := json.Marshal(peerPool)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Marshal peerPool error")
			}
			native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), newView.Bytes(), peerPubkeyPrefix),
				&cstates.StorageItem{Value: value})
		}

		if peerPool.Status == CandidateStatus || peerPool.Status == ConsensusStatus {
			stake := peerPool.TotalPos + peerPool.InitPos
			peers = append(peers, &states.PeerStakeInfo{
				Index:      peerPool.Index,
				PeerPubkey: peerPool.PeerPubkey,
				Stake:      stake,
			})
		}
	}

	// get config
	config := new(states.Configuration)
	configBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(VBFT_CONFIG)))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Get configBytes error!")
	}
	if configBytes == nil {
		return errors.NewErr("[commitDpos] ConfigBytes is nil!")
	}
	configStore, _ := configBytes.(*cstates.StorageItem)
	err = json.Unmarshal(configStore.Value, config)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Unmarshal config error!")
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
			return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] PeerPubkey format error!")
		}
		peerPoolBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), view.Bytes(), peerPubkeyPrefix))
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Get peerPoolBytes error!")
		}

		peerPool := new(states.PeerPool)
		if peerPoolBytes == nil {
			return errors.NewErr("[commitDpos] PeerPoolBytes consensus is nil!")
		}
		peerPoolStore, _ := peerPoolBytes.(*cstates.StorageItem)
		err = json.Unmarshal(peerPoolStore.Value, peerPool)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Unmarshal peerPool error!")
		}

		if peerPool.Status == ConsensusStatus {
			//update voteInfoPool
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
				addressPrefix, err := hex.DecodeString(voteInfoPool.Address)
				if err != nil {
					errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Address format error!")
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
					return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Marshal voteInfoPool error")
				}
				native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(VOTE_INFO_POOL), newView.Bytes(),
					peerPubkeyPrefix, addressPrefix), &cstates.StorageItem{Value: value})
			}
		} else {
			//update voteInfoPool
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
				addressPrefix, err := hex.DecodeString(voteInfoPool.Address)
				if err != nil {
					errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Address format error!")
				}
				preFreezePos := voteInfoPool.PreFreezePos
				freezePos := voteInfoPool.FreezePos
				voteInfoPool.PrePos = voteInfoPool.NewPos
				voteInfoPool.NewPos = preFreezePos + freezePos
				voteInfoPool.PreFreezePos = 0
				voteInfoPool.FreezePos = 0

				value, err := json.Marshal(voteInfoPool)
				if err != nil {
					return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Marshal voteInfoPool error")
				}
				native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(VOTE_INFO_POOL), newView.Bytes(),
					peerPubkeyPrefix, addressPrefix), &cstates.StorageItem{Value: value})
			}
		}
		peerPool.Status = ConsensusStatus
		value, err := json.Marshal(peerPool)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Marshal peerPool error")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), newView.Bytes(), peerPubkeyPrefix),
			&cstates.StorageItem{Value: value})
	}

	//non consensus peers
	for i := int(config.K); i < len(peers); i++ {
		//change peerPool status
		peerPubkeyPrefix, err := hex.DecodeString(peers[i].PeerPubkey)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] PeerPubkey format error!")
		}

		peerPoolBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), view.Bytes(), peerPubkeyPrefix))
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Get peerPoolBytes error!")
		}
		peerPool := new(states.PeerPool)
		if peerPoolBytes == nil {
			return errors.NewErr("[commitDpos] PeerPoolBytes non consensus is nil!")
		}
		peerPoolStore, _ := peerPoolBytes.(*cstates.StorageItem)
		err = json.Unmarshal(peerPoolStore.Value, peerPool)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Unmarshal peerPool error!")
		}
		if peerPool.Status == ConsensusStatus {
			//update voteInfoPool
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
				addressPrefix, err := hex.DecodeString(voteInfoPool.Address)
				if err != nil {
					errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Address format error!")
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
					return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Marshal voteInfoPool error")
				}
				native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(VOTE_INFO_POOL), newView.Bytes(),
					peerPubkeyPrefix, addressPrefix), &cstates.StorageItem{Value: value})
			}
		} else {
			//update voteInfoPool
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
				addressPrefix, err := hex.DecodeString(voteInfoPool.Address)
				if err != nil {
					errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Address format error!")
				}
				preFreezePos := voteInfoPool.PreFreezePos
				newPos := voteInfoPool.NewPos
				freezePos := voteInfoPool.FreezePos
				voteInfoPool.NewPos = freezePos
				voteInfoPool.FreezePos = newPos + preFreezePos
				voteInfoPool.PreFreezePos = 0

				value, err := json.Marshal(voteInfoPool)
				if err != nil {
					return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Marshal voteInfoPool error")
				}
				native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(VOTE_INFO_POOL), newView.Bytes(),
					peerPubkeyPrefix, addressPrefix), &cstates.StorageItem{Value: value})
			}
		}
		peerPool.Status = CandidateStatus
		value, err := json.Marshal(peerPool)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Marshal peerPool error")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), newView.Bytes(), peerPubkeyPrefix),
			&cstates.StorageItem{Value: value})
	}

	//get all vote for commit info
	stateValues, err = native.CloneCache.Store.Find(scommon.ST_STORAGE, concatKey(contract, []byte(VOTE_COMMIT_INFO), view.Bytes()))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Get all peerPool error!")
	}

	voteCommitInfoPool := new(states.VoteCommitInfoPool)
	for _, v := range stateValues {
		voteCommitInfoPoolStore, _ := v.Value.(*cstates.StorageItem)
		err = json.Unmarshal(voteCommitInfoPoolStore.Value, voteCommitInfoPool)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Unmarshal voteCommitInfoPool error!")
		}

		addressBytes, err := hex.DecodeString(voteCommitInfoPool.Address)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Address format error!")
		}
		address, err := common.AddressParseFromBytes(addressBytes)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Address format error!")
		}
		//ont transfer
		err = appCallTransferOnt(native, genesis.GovernanceContractAddress, address, new(big.Int).SetUint64(voteCommitInfoPool.Pos))
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[voteForPeer] Ont transfer error!")
		}
	}

	posTable, chainPeers, err := calDposTable(native, config, peers)
	fmt.Println("ChainPeers is :", chainPeers)
	fmt.Println("DPOS table is:", posTable)

	//update view
	governanceView := &states.GovernanceView{
		View:       newView,
		VoteCommit: false,
	}
	v, err := json.Marshal(governanceView)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Marshal governanceView error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(GOVERNANCE_VIEW)), &cstates.StorageItem{Value: v})

	addCommonEvent(native, contract, COMMIT_DPOS, true)

	return nil
}

func VoteCommitDpos(native *NativeService) error {
	params := new(states.VoteCommitDposParam)
	err := json.Unmarshal(native.Input, params)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[voteCommitDpos] Contract params Unmarshal error!")
	}

	//check witness
	err = validateOwner(native, params.Address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[voteCommitDpos] CheckWitness error!")
	}

	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := getGovernanceView(native, contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[voteCommitDpos] Get view error!")
	}

	addressPrefix, err := hex.DecodeString(params.Address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[voteCommitDpos] Address format error!")
	}

	//get voteCommitInfo
	voteCommitInfoPool := new(states.VoteCommitInfoPool)
	voteCommitInfoPoolBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(VOTE_COMMIT_INFO), view.Bytes(), addressPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[voteCommitDpos] Get voteCommitInfoBytes error!")
	}
	if voteCommitInfoPoolBytes != nil {
		voteCommitInfoPoolStore, _ := voteCommitInfoPoolBytes.(*cstates.StorageItem)
		err = json.Unmarshal(voteCommitInfoPoolStore.Value, voteCommitInfoPool)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[voteCommitDpos] Unmarshal voteCommitInfoPool error!")
		}
	}
	pos := int64(voteCommitInfoPool.Pos) + params.Pos
	if pos < 0 {
		return errors.NewErr("[voteCommitDpos] Remain pos is negative!")
	}
	voteCommitInfoPool.Pos = uint64(pos)
	v, err := json.Marshal(voteCommitInfoPool)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Marshal governanceView error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(VOTE_COMMIT_INFO), view.Bytes(), addressPrefix), &cstates.StorageItem{Value: v})

	//get total pos for commit
	posCommitBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(POS_FOR_COMMIT), view.Bytes()))
	posCommit := new(big.Int)
	if posCommitBytes != nil {
		posCommitStore, _ := posCommitBytes.(*cstates.StorageItem)
		posCommit = new(big.Int).SetBytes(posCommitStore.Value)
	}
	newPosCommit := posCommit.Int64() + params.Pos

	if newPosCommit >= 100000 {
		governanceView := &states.GovernanceView{
			View:       view,
			VoteCommit: true,
		}
		v, err := json.Marshal(governanceView)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[voteCommitDpos] Marshal governanceView error")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(GOVERNANCE_VIEW)), &cstates.StorageItem{Value: v})
	}

	addressBytes, err := hex.DecodeString(params.Address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[voteCommitDpos] Address format error!")
	}
	address, err := common.AddressParseFromBytes(addressBytes)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[voteCommitDpos] Address format error!")
	}

	//ont transfer
	if params.Pos > 0 {
		err = appCallTransferOnt(native, address, genesis.GovernanceContractAddress, new(big.Int).SetInt64(params.Pos))
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[voteCommitDpos] Ont transfer error!")
		}
	}
	if params.Pos < 0 {
		err = appCallTransferOnt(native, address, genesis.GovernanceContractAddress, new(big.Int).Abs(new(big.Int).SetInt64(params.Pos)))
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[voteCommitDpos] Ont transfer error!")
		}
	}

	return nil
}

func UpdateConfig(native *NativeService) error {
	configuration := new(states.Configuration)
	err := json.Unmarshal(native.Input, configuration)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[updateConfig] Contract params Unmarshal error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//check the configuration
	if configuration.L < 16*configuration.K {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[updateConfig] L is less than 16*K in config!")
	}

	value, err := json.Marshal(configuration)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[updateConfig] Marshal config error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(VBFT_CONFIG)), &cstates.StorageItem{Value: value})

	addCommonEvent(native, contract, UPDATE_CONFIG, configuration)

	return nil
}

func DataQuery(native *NativeService) error {
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := getGovernanceView(native, contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[DataQuery] Get view error!")
	}

	//get all peerPool
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), view.Bytes()))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[DataQuery] Get all peerPool error!")
	}

	peerPool := new(states.PeerPool)
	for _, v := range stateValues {
		peerPoolStore, _ := v.Value.(*cstates.StorageItem)
		err = json.Unmarshal(peerPoolStore.Value, peerPool)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[DataQuery] Unmarshal peerPool error!")
		}
		fmt.Println("PeerPool is : ", peerPool)
	}

	fmt.Println("view :", view)
	//update voteInfoPool
	stateValues, err = native.CloneCache.Store.Find(scommon.ST_STORAGE, concatKey(contract, []byte(VOTE_INFO_POOL), view.Bytes()))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[DataQuery] Get all peerPool error!")
	}
	voteInfoPool := new(states.VoteInfoPool)
	for _, v := range stateValues {
		voteInfoPoolStore, _ := v.Value.(*cstates.StorageItem)
		err = json.Unmarshal(voteInfoPoolStore.Value, voteInfoPool)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[DataQuery] Unmarshal voteInfoPool error!")
		}
		fmt.Println("VoteInfoPool is : ", voteInfoPool)
	}

	return nil
}
