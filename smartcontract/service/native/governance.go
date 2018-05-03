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
	"math/big"

	"fmt"
	"io/ioutil"
	"sort"

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
	PEER_POOL        = "peerePool"
	VOTE_INFO_POOL   = "voteInfoPool"
	POS_FOR_COMMIT   = "posForCommit"
	FORCE_COMMIT     = "forceCommit"
	VOTE_COMMIT_INFO = "voteCommitInfo"

	//global
	ConsensusNum = 7
	CandidateNum = 7 * 7
	SyncNodeNum  = 7 * 7 * 6
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

	config := new(states.Configuration)
	err = json.Unmarshal(file, config)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[initConfig] Contract params Unmarshal error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	//TODO: check the config
	value, err := json.Marshal(config)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[initConfig] Marshal candidatePool error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(VBFT_CONFIG)), &cstates.StorageItem{Value: value})

	initPeerPool := &states.PeerPoolList{}
	if err := json.Unmarshal(file, initPeerPool); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[initConfig] Failed to unmarshal config file!")
	}

	for _, peerPool := range initPeerPool.Peers {
		peerPool.TotalPos = new(big.Int)
		peerPool.Status = ConsensusStatus
		value, err := json.Marshal(peerPool)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[initConfig] Marshal candidatePool error!")
		}
		peerPubkeyPrefix, err := hex.DecodeString(peerPool.PeerPubkey)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[initConfig] PeerPubkey format error!")
		}
		native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), peerPubkeyPrefix), &cstates.StorageItem{Value: value})
	}

	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(CANDIDITE_INDEX)), &cstates.StorageItem{Value: new(big.Int).SetInt64(8).Bytes()})

	governanceView := &states.GovernanceView{
		View:       new(big.Int).SetInt64(1),
		VoteCommit: false,
	}
	v, err := json.Marshal(governanceView)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[initConfig] Marshal governanceView error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(GOVERNANCE_VIEW)), &cstates.StorageItem{Value: v})

	addCommonEvent(native, contract, INIT_CONFIG, true)
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

	//check PeerPool
	v1, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), peerPubkeyPrefix))
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
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), peerPubkeyPrefix), &cstates.StorageItem{Value: value})

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

	//get peerPool
	peerPoolBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), peerPubkeyPrefix))
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
		return errors.NewDetailErr(err, errors.ErrNoCode, "[approveSyncNode] PeerPubkey is not in peerPool!")
	}

	if peerPool.Status != RegisterSyncNodeStatus {
		return errors.NewErr("[approveSyncNode] Peer status is not RegisterSyncNodeStatus!")
	}

	peerPool.Status = SyncNodeStatus

	value, err := json.Marshal(peerPool)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[approveSyncNode] Marshal peerPool error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), peerPubkeyPrefix), &cstates.StorageItem{Value: value})

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

	//syncNodePool storage
	//check syncNodePool
	peerPoolBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), peerPubkeyPrefix))
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
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), peerPubkeyPrefix), &cstates.StorageItem{Value: value})

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

	//get peerPool
	peerPoolBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[approveCandidate] Get peerPoolBytes error!")
	}
	peerPool := new(states.PeerPool)
	if peerPoolBytes != nil {
		peerPoolStore, _ := peerPoolBytes.(*cstates.StorageItem)
		err := json.Unmarshal(peerPoolStore.Value, peerPool)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[approveCandidate] Unmarshal peerPool error!")
		}
	} else {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[approveCandidate] PeerPubkey is not in peerPool!")
	}

	if peerPool.Status != RegisterCandidateStatus {
		return errors.NewErr("[approveCandidate] Peer status is not RegisterCandidateStatus!")
	}

	peerPool.Status = CandidateStatus
	peerPool.TotalPos = new(big.Int)

	//get index
	candidateIndexBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(CANDIDITE_INDEX)))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[approveCandidate] Get candidateIndex error!")
	}
	var candidateIndex *big.Int
	if candidateIndexBytes == nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[approveCandidate] CandidateIndex is not init!")
	} else {
		candidateIndexStore, _ := candidateIndexBytes.(*cstates.StorageItem)
		candidateIndex = new(big.Int).SetBytes(candidateIndexStore.Value)
	}

	peerPool.Index = candidateIndex

	value, err := json.Marshal(peerPool)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[approveCandidate] Marshal peerPool error")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), peerPubkeyPrefix), &cstates.StorageItem{Value: value})

	//update candidateIndex
	newCandidateIndex := new(big.Int).Add(candidateIndex, new(big.Int).SetInt64(1))
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(CANDIDITE_INDEX)), &cstates.StorageItem{Value: newCandidateIndex.Bytes()})

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

	//get peerPool
	peerPoolBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), peerPubkeyPrefix))
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
		return errors.NewDetailErr(err, errors.ErrNoCode, "[quitNode] PeerPubkey is not in peerPool!")
	}
	if params.Address != peerPool.Address {
		return errors.NewErr("[quitNode] PeerPubkey is not registered by this address!")
	}

	//get current view
	view, err := getGovernanceView(native, contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[quitNode] Get view error!")
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
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), peerPubkeyPrefix), &cstates.StorageItem{Value: value})

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

	total := new(big.Int)
	for peerPubkey, pos := range params.VoteTable {
		peerPubkeyPrefix, err := hex.DecodeString(peerPubkey)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[voteForPeer] PeerPubkey format error!")
		}

		//get peerPool
		peerPoolBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), peerPubkeyPrefix))
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
			PeerPubkey:   peerPubkey,
			Address:      params.Address,
			PrePos:       new(big.Int),
			FreezePos:    new(big.Int),
			NewPos:       new(big.Int),
			PreFreezePos: new(big.Int),
		}
		if pos.Sign() >= 0 {
			if voteInfoPoolBytes != nil {
				voteInfoPoolStore, _ := voteInfoPoolBytes.(*cstates.StorageItem)
				err = json.Unmarshal(voteInfoPoolStore.Value, voteInfoPool)
				if err != nil {
					return errors.NewDetailErr(err, errors.ErrNoCode, "[voteForPeer] Unmarshal voteInfoPool error!")
				}
				voteInfoPool.NewPos = new(big.Int).Add(voteInfoPool.NewPos, pos)
			} else {
				voteInfoPool.NewPos = pos
			}
			total = new(big.Int).Add(total, pos)
			peerPool.TotalPos = new(big.Int).Add(peerPool.TotalPos, pos)
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
			native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), peerPubkeyPrefix), &cstates.StorageItem{Value: value})
		} else {
			if voteInfoPoolBytes != nil {
				voteInfoPoolStore, _ := voteInfoPoolBytes.(*cstates.StorageItem)
				err = json.Unmarshal(voteInfoPoolStore.Value, voteInfoPool)
				if err != nil {
					return errors.NewDetailErr(err, errors.ErrNoCode, "[voteForPeer] Unmarshal voteInfoPool error!")
				}
				temp := new(big.Int).Add(voteInfoPool.NewPos, pos)
				if temp.Sign() < 0 {
					voteInfoPool.PrePos = new(big.Int).Add(voteInfoPool.PrePos, temp)
					if voteInfoPool.PrePos.Sign() < 0 {
						continue
					}
					voteInfoPool.PreFreezePos = new(big.Int).Sub(voteInfoPool.PreFreezePos, temp)
					total = new(big.Int).Sub(total, voteInfoPool.NewPos)
					voteInfoPool.NewPos = new(big.Int)
					peerPool.TotalPos = new(big.Int).Add(peerPool.TotalPos, temp)
				} else {
					voteInfoPool.NewPos = new(big.Int).Add(voteInfoPool.NewPos, pos)
					total = new(big.Int).Add(total, pos)
					peerPool.TotalPos = new(big.Int).Add(peerPool.TotalPos, pos)
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
				native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), peerPubkeyPrefix), &cstates.StorageItem{Value: value})
			}
		}
	}
	fmt.Println("Total is :", total)
	if total.Sign() > 0 {
		//TODO: pay
	}
	if total.Sign() < 0 {
		//TODO: withdraw
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
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL)))
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

		if peerPool.Status == QuitStatus {
			//get qiutView
			peerPoolBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), peerPubkeyPrefix))
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
			quitView := peerPool.QuitView

			//draw back init pos
			//TODO: Draw back init pos
			fmt.Printf("############################## draw back init pos %v, to address %v \n", peerPool.InitPos, peerPool.Address)

			//draw back vote pos
			//TODO: Draw back vote pos
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
				pos1 := new(big.Int).Add(voteInfoPool.PrePos, voteInfoPool.PreFreezePos)
				pos2 := new(big.Int).Add(voteInfoPool.FreezePos, voteInfoPool.NewPos)
				pos := new(big.Int).Add(pos1, pos2)
				fmt.Printf("########################### draw back vote pos %v, to address %v \n", pos, voteInfoPool.Address)
			}

			native.CloneCache.Delete(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), peerPubkeyPrefix))
		}
		if peerPool.Status == QuitConsensusStatus {
			peerPool.Status = QuitStatus
			value, err := json.Marshal(peerPool)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Marshal peerPool error")
			}
			native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), peerPubkeyPrefix), &cstates.StorageItem{Value: value})
		}

		if peerPool.Status == CandidateStatus || peerPool.Status == ConsensusStatus {
			stake := new(big.Int).Add(peerPool.TotalPos, peerPool.InitPos)
			peers = append(peers, &states.PeerStakeInfo{
				Index:      uint32(peerPool.Index.Uint64()),
				PeerPubkey: peerPool.PeerPubkey,
				Stake:      stake.Uint64(),
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

	// get stake sum of top-k peers
	var sum uint64
	for i := 0; i < int(config.K); i++ {
		sum += peers[i].Stake

		//change peerPool status
		peerPubkeyPrefix, err := hex.DecodeString(peers[i].PeerPubkey)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] PeerPubkey format error!")
		}
		peerPoolBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), peerPubkeyPrefix))
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
			stateValues, err = native.CloneCache.Store.Find(scommon.ST_STORAGE, concatKey(contract, []byte(VOTE_INFO_POOL), view.Bytes(), peerPubkeyPrefix))
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
				voteInfoPool.PrePos = new(big.Int).Add(voteInfoPool.PrePos, newPos)
				voteInfoPool.NewPos = freezePos
				voteInfoPool.FreezePos = preFreezePos
				voteInfoPool.PreFreezePos = new(big.Int)

				value, err := json.Marshal(voteInfoPool)
				if err != nil {
					return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Marshal voteInfoPool error")
				}
				native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(VOTE_INFO_POOL), newView.Bytes(),
					peerPubkeyPrefix, addressPrefix), &cstates.StorageItem{Value: value})
			}
		} else {
			//update voteInfoPool
			stateValues, err = native.CloneCache.Store.Find(scommon.ST_STORAGE, concatKey(contract, []byte(VOTE_INFO_POOL), view.Bytes(), peerPubkeyPrefix))
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
				voteInfoPool.NewPos = new(big.Int).Add(preFreezePos, freezePos)
				voteInfoPool.PreFreezePos = new(big.Int)
				voteInfoPool.FreezePos = new(big.Int)

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
		native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), peerPubkeyPrefix), &cstates.StorageItem{Value: value})
	}

	//non consensus peers
	for i := int(config.K); i < len(peers); i++ {
		//change peerPool status
		peerPubkeyPrefix, err := hex.DecodeString(peers[i].PeerPubkey)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] PeerPubkey format error!")
		}

		peerPoolBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), peerPubkeyPrefix))
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
			stateValues, err = native.CloneCache.Store.Find(scommon.ST_STORAGE, concatKey(contract, []byte(VOTE_INFO_POOL), view.Bytes(), peerPubkeyPrefix))
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
				voteInfoPool.NewPos = new(big.Int).Add(voteInfoPool.NewPos, freezePos)
				voteInfoPool.FreezePos = new(big.Int).Add(prePos, preFreezePos)
				voteInfoPool.PrePos = new(big.Int)
				voteInfoPool.PreFreezePos = new(big.Int)

				value, err := json.Marshal(voteInfoPool)
				if err != nil {
					return errors.NewDetailErr(err, errors.ErrNoCode, "[commitDpos] Marshal voteInfoPool error")
				}
				native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(VOTE_INFO_POOL), newView.Bytes(),
					peerPubkeyPrefix, addressPrefix), &cstates.StorageItem{Value: value})
			}
		} else {
			//update voteInfoPool
			stateValues, err = native.CloneCache.Store.Find(scommon.ST_STORAGE, concatKey(contract, []byte(VOTE_INFO_POOL), view.Bytes(), peerPubkeyPrefix))
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
				voteInfoPool.NewPos = new(big.Int).Add(voteInfoPool.NewPos, voteInfoPool.PrePos)
				voteInfoPool.NewPos = new(big.Int).Add(voteInfoPool.NewPos, voteInfoPool.FreezePos)
				voteInfoPool.PrePos = new(big.Int)
				voteInfoPool.FreezePos = preFreezePos
				voteInfoPool.PreFreezePos = new(big.Int)

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
		native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL), peerPubkeyPrefix), &cstates.StorageItem{Value: value})
	}

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
	voteCommitInfo := new(big.Int)
	voteCommitInfoBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(VOTE_COMMIT_INFO), view.Bytes(), addressPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[voteCommitDpos] Get voteCommitInfoBytes error!")
	}
	if voteCommitInfoBytes != nil {
		voteCommitStore, _ := voteCommitInfoBytes.(*cstates.StorageItem)
		voteCommitInfo = new(big.Int).SetBytes(voteCommitStore.Value)
	}
	newVoteCommitInfo := new(big.Int).Add(voteCommitInfo, params.Pos)
	if newVoteCommitInfo.Sign() < 0 {
		return errors.NewErr("[voteCommitDpos] Remain pos is negative!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(VOTE_COMMIT_INFO), view.Bytes(), addressPrefix), &cstates.StorageItem{Value: newVoteCommitInfo.Bytes()})

	//get total pos for commit
	posCommitBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(POS_FOR_COMMIT), view.Bytes()))
	posCommit := new(big.Int)
	if posCommitBytes != nil {
		posCommitStore, _ := posCommitBytes.(*cstates.StorageItem)
		posCommit = new(big.Int).SetBytes(posCommitStore.Value)
	}
	newPosCommit := new(big.Int).Add(posCommit, params.Pos)

	if newPosCommit.Cmp(new(big.Int).SetInt64(100000)) >= 0 {
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

	return nil
}

func UpdateConfig(native *NativeService) error {
	config := new(states.Configuration)
	err := json.Unmarshal(native.Input, config)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[updateConfig] Contract params Unmarshal error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//TODO: check the config

	value, err := json.Marshal(config)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[updateConfig] Marshal config error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, concatKey(contract, []byte(VBFT_CONFIG)), &cstates.StorageItem{Value: value})

	addCommonEvent(native, contract, UPDATE_CONFIG, config)

	return nil
}

func DataQuery(native *NativeService) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	//get all peerPool
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, concatKey(contract, []byte(PEER_POOL)))
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

	//get current view
	view, err := getGovernanceView(native, contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[DataQuery] Get view error!")
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
