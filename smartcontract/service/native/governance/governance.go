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
	"github.com/ontio/ontology/common/serialization"
	cstates "github.com/ontio/ontology/core/states"
	scommon "github.com/ontio/ontology/core/store/common"
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
	BlackStatus
)

const (
	//function name
	INIT_CONFIG                      = "initConfig"
	REGISTER_CANDIDATE               = "registerCandidate"
	REGISTER_CANDIDATE_TRANSFER_FROM = "registerCandidateTransferFrom"
	UNREGISTER_CANDIDATE             = "unRegisterCandidate"
	AUTHORIZE_FOR_PEER               = "authorizeForPeer"
	AUTHORIZE_FOR_PEER_TRANSFER_FROM = "authorizeForPeerTransferFrom"
	UNAUTHORIZE_FOR_PEER             = "unAuthorizeForPeer"
	APPROVE_CANDIDATE                = "approveCandidate"
	REJECT_CANDIDATE                 = "rejectCandidate"
	BLACK_NODE                       = "blackNode"
	WHITE_NODE                       = "whiteNode"
	QUIT_NODE                        = "quitNode"
	WITHDRAW                         = "withdraw"
	WITHDRAW_ONG                     = "withdrawOng"
	WITHDRAW_FEE                     = "withdrawFee"
	COMMIT_DPOS                      = "commitDpos"
	UPDATE_CONFIG                    = "updateConfig"
	UPDATE_GLOBAL_PARAM              = "updateGlobalParam"
	UPDATE_GLOBAL_PARAM2             = "updateGlobalParam2"
	UPDATE_SPLIT_CURVE               = "updateSplitCurve"
	TRANSFER_PENALTY                 = "transferPenalty"
	CHANGE_MAX_AUTHORIZATION         = "changeMaxAuthorization"
	SET_PEER_COST                    = "setPeerCost"
	ADD_INIT_POS                     = "addInitPos"
	REDUCE_INIT_POS                  = "reduceInitPos"
	SET_PROMISE_POS                  = "setPromisePos"

	//key prefix
	GLOBAL_PARAM      = "globalParam"
	GLOBAL_PARAM2     = "globalParam2"
	VBFT_CONFIG       = "vbftConfig"
	GOVERNANCE_VIEW   = "governanceView"
	CANDIDITE_INDEX   = "candidateIndex"
	PEER_POOL         = "peerPool"
	PEER_INDEX        = "peerIndex"
	BLACK_LIST        = "blackList"
	TOTAL_STAKE       = "totalStake"
	PENALTY_STAKE     = "penaltyStake"
	SPLIT_CURVE       = "splitCurve"
	PEER_ATTRIBUTES   = "peerAttributes"
	SPLIT_FEE         = "splitFee"
	SPLIT_FEE_ADDRESS = "splitFeeAddress"
	PROMISE_POS       = "promisPos"

	//global
	PRECISE = 1000000
)

// candidate fee must >= 1 ONG
var MIN_CANDIDATE_FEE = uint64(math.Pow(10, constants.ONG_DECIMALS))
var AUTHORIZE_INFO_POOL = []byte{118, 111, 116, 101, 73, 110, 102, 111, 80, 111, 111, 108}
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
	native.Register(REGISTER_CANDIDATE, RegisterCandidate)
	native.Register(REGISTER_CANDIDATE_TRANSFER_FROM, RegisterCandidateTransferFrom)
	native.Register(UNREGISTER_CANDIDATE, UnRegisterCandidate)
	native.Register(AUTHORIZE_FOR_PEER, AuthorizeForPeer)
	native.Register(AUTHORIZE_FOR_PEER_TRANSFER_FROM, AuthorizeForPeerTransferFrom)
	native.Register(UNAUTHORIZE_FOR_PEER, UnAuthorizeForPeer)
	native.Register(WITHDRAW, Withdraw)
	native.Register(QUIT_NODE, QuitNode)
	native.Register(WITHDRAW_ONG, WithdrawOng)
	native.Register(CHANGE_MAX_AUTHORIZATION, ChangeMaxAuthorization)
	native.Register(SET_PEER_COST, SetPeerCost)
	native.Register(WITHDRAW_FEE, WithdrawFee)
	native.Register(ADD_INIT_POS, AddInitPos)
	native.Register(REDUCE_INIT_POS, ReduceInitPos)

	native.Register(INIT_CONFIG, InitConfig)
	native.Register(APPROVE_CANDIDATE, ApproveCandidate)
	native.Register(REJECT_CANDIDATE, RejectCandidate)
	native.Register(BLACK_NODE, BlackNode)
	native.Register(WHITE_NODE, WhiteNode)
	native.Register(COMMIT_DPOS, CommitDpos)
	native.Register(UPDATE_CONFIG, UpdateConfig)
	native.Register(UPDATE_GLOBAL_PARAM, UpdateGlobalParam)
	native.Register(UPDATE_GLOBAL_PARAM2, UpdateGlobalParam2)
	native.Register(UPDATE_SPLIT_CURVE, UpdateSplitCurve)
	native.Register(TRANSFER_PENALTY, TransferPenalty)
	native.Register(SET_PROMISE_POS, SetPromisePos)
}

//Init governance contract, include vbft config, global param and ontid admin.
func InitConfig(native *native.NativeService) ([]byte, error) {
	configuration := new(config.VBFTConfig)
	buf, err := serialization.ReadVarBytes(bytes.NewBuffer(native.Input))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("serialization.ReadVarBytes, contract params deserialize error: %v", err)
	}
	if err := configuration.Deserialize(bytes.NewBuffer(buf)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deserialize, contract params deserialize error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	// check if initConfig is already execute
	governanceViewBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(GOVERNANCE_VIEW)))
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

	//init globalParam
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
	err = putGlobalParam(native, contract, globalParam)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putGlobalParam, put globalParam error: %v", err)
	}

	var view uint32 = 1
	var maxId uint32

	peerPoolMap := &PeerPoolMap{
		PeerPoolMap: make(map[string]*PeerPoolItem),
	}
	for _, peer := range configuration.Peers {
		if peer.Index > maxId {
			maxId = peer.Index
		}
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

		peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("hex.DecodeString, peerPubkey format error: %v", err)
		}
		index := peerPoolItem.Index
		indexBytes, err := GetUint32Bytes(index)
		if err != nil {
			return nil, fmt.Errorf("getUint32Bytes, getUint32Bytes error: %v", err)
		}
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_INDEX), peerPubkeyPrefix), &cstates.StorageItem{Value: indexBytes})

		//update total stake
		err = depositTotalStake(native, contract, address, peerPoolItem.InitPos)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("depositTotalStake, depositTotalStake error: %v", err)
		}
	}

	//init peer pool
	err = putPeerPoolMap(native, contract, 0, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putPeerPoolMap, put peerPoolMap error: %v", err)
	}
	err = putPeerPoolMap(native, contract, view, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putPeerPoolMap, put peerPoolMap error: %v", err)
	}
	indexBytes, err := GetUint32Bytes(maxId + 1)
	if err != nil {
		return nil, fmt.Errorf("getUint32Bytes, get indexBytes error: %v", err)
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(CANDIDITE_INDEX)), &cstates.StorageItem{Value: indexBytes})

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

	//init admin OntID
	err = appCallInitContractAdmin(native, []byte(configuration.AdminOntID))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("appCallInitContractAdmin error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

//Register a candidate node, used by users.
//Users can register a candidate node with a authorized ontid.
//Candidate node can be authorized and become consensus node according to their pos.
//Candidate node can get ong bonus according to their pos.
func RegisterCandidate(native *native.NativeService) ([]byte, error) {
	err := registerCandidate(native, "transfer")
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("registerCandidate error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

//Register a candidate node, used by contracts.
//Contracts can register a candidate node with a authorized ontid after approving ont to governance contract before invoke this function.
//Candidate node can be authorized and become consensus node according to their pos.
//Candidate node can get ong bonus according to their pos.
func RegisterCandidateTransferFrom(native *native.NativeService) ([]byte, error) {
	err := registerCandidate(native, "transferFrom")
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("registerCandidateTransferFrom error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

//Unregister a registered candidate node, will remove node from pool, and unfreeze deposit ont.
func UnRegisterCandidate(native *native.NativeService) ([]byte, error) {
	params := new(UnRegisterCandidateParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deserialize, contract params deserialize error: %v", err)
	}
	address := params.Address
	contract := native.ContextRef.CurrentContext().ContractAddress

	//check witness
	err := utils.ValidateOwner(native, address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("validateOwner, checkWitness error: %v", err)
	}

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getView, get view error: %v", err)
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getPeerPoolMap, get peerPoolMap error: %v", err)
	}

	//check if exist in PeerPool
	peerPoolItem, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, fmt.Errorf("unRegisterCandidate, peerPubkey is not in peerPoolMap: %v", err)
	}

	if peerPoolItem.Status != RegisterCandidateStatus {
		return utils.BYTE_FALSE, fmt.Errorf("unRegisterCandidate, peer status is not RegisterCandidateStatus")
	}

	//check owner address
	if peerPoolItem.Address != params.Address {
		return utils.BYTE_FALSE, fmt.Errorf("unRegisterCandidate, address is not peer owner")
	}

	//unfreeze initPos
	authorizeInfo := &AuthorizeInfo{
		PeerPubkey:          peerPoolItem.PeerPubkey,
		Address:             peerPoolItem.Address,
		WithdrawUnfreezePos: peerPoolItem.InitPos,
	}
	err = putAuthorizeInfo(native, contract, authorizeInfo)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putAuthorizeInfo, put authorizeInfo error: %v", err)
	}

	delete(peerPoolMap.PeerPoolMap, params.PeerPubkey)
	err = putPeerPoolMap(native, contract, view, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putPeerPoolMap, put peerPoolMap error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

//Approve a registered candidate node
//Only approved candidate node can participate in consensus selection and get ong bonus.
func ApproveCandidate(native *native.NativeService) ([]byte, error) {
	params := new(ApproveCandidateParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deserialize, contract params deserialize error: %v", err)
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
		return utils.BYTE_FALSE, fmt.Errorf("approveCandidate, checkWitness error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getView, get view error: %v", err)
	}

	//get globalParam
	globalParam, err := getGlobalParam(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getGlobalParam, getGlobalParam error: %v", err)
	}

	//check if peerPoolMap full
	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getPeerPoolMap, get peerPoolMap error: %v", err)
	}

	num := 0
	for _, peerPoolItem := range peerPoolMap.PeerPoolMap {
		if peerPoolItem.Status == CandidateStatus || peerPoolItem.Status == ConsensusStatus {
			num = num + 1
		}
	}
	if num >= int(globalParam.CandidateNum) {
		return utils.BYTE_FALSE, fmt.Errorf("approveCandidate, num of candidate node is full")
	}

	//get peerPool
	peerPoolItem, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, fmt.Errorf("approveCandidate, peerPubkey is not in peerPoolMap")
	}

	//check initPos
	if peerPoolItem.InitPos < uint64(globalParam.MinInitStake) {
		return utils.BYTE_FALSE, fmt.Errorf("approveCandidate, initPos must >= %v", globalParam.MinInitStake)
	}

	if peerPoolItem.Status != RegisterCandidateStatus {
		return utils.BYTE_FALSE, fmt.Errorf("approveCandidate, peer status is not RegisterCandidateStatus")
	}

	peerPoolItem.Status = CandidateStatus
	peerPoolItem.TotalPos = 0

	//check if has index
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("hex.DecodeString, peerPubkey format error: %v", err)
	}
	indexBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_INDEX), peerPubkeyPrefix))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("native.CloneCache.Get, get indexBytes error: %v", err)
	}
	if indexBytes != nil {
		index, err := GetBytesUint32(indexBytes.(*cstates.StorageItem).Value)
		if err != nil {
			return nil, fmt.Errorf("GetBytesUint32, get index error: %v", err)
		}
		peerPoolItem.Index = index
	} else {
		//get candidate index
		candidateIndex, err := getCandidateIndex(native, contract)
		if err != nil {
			return nil, fmt.Errorf("getCandidateIndex, get candidateIndex error: %v", err)
		}
		peerPoolItem.Index = candidateIndex

		//update candidateIndex
		newCandidateIndex := candidateIndex + 1
		err = putCandidateIndex(native, contract, newCandidateIndex)
		if err != nil {
			return nil, fmt.Errorf("putCandidateIndex, put candidateIndex error: %v", err)
		}

		indexBytes, err := GetUint32Bytes(peerPoolItem.Index)
		if err != nil {
			return nil, fmt.Errorf("GetUint32Bytes, get indexBytes error: %v", err)
		}
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_INDEX), peerPubkeyPrefix), &cstates.StorageItem{Value: indexBytes})
	}
	peerPoolMap.PeerPoolMap[params.PeerPubkey] = peerPoolItem
	err = putPeerPoolMap(native, contract, view, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putPeerPoolMap, put peerPoolMap error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

//Reject a registered candidate node, remove node from pool and unfreeze deposit ont
//Only approved candidate node can participate in consensus selection and get ong bonus.
func RejectCandidate(native *native.NativeService) ([]byte, error) {
	params := new(RejectCandidateParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deserialize, contract params deserialize error: %v", err)
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
		return utils.BYTE_FALSE, fmt.Errorf("approveCandidate, checkWitness error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getView, get view error: %v", err)
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getPeerPoolMap, get peerPoolMap error: %v", err)
	}

	//draw back init pos
	peerPoolItem, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, fmt.Errorf("rejectCandidate, peerPubkey is not in peerPoolMap")
	}
	if peerPoolItem.Status != RegisterCandidateStatus {
		return utils.BYTE_FALSE, fmt.Errorf("rejectCandidate, peerPubkey is not RegisterCandidateStatus")
	}
	address := peerPoolItem.Address
	authorizeInfo, err := getAuthorizeInfo(native, contract, params.PeerPubkey, address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getAuthorizeInfo, get authorizeInfo error: %v", err)
	}
	authorizeInfo.WithdrawUnfreezePos = authorizeInfo.WithdrawUnfreezePos + peerPoolItem.InitPos
	err = putAuthorizeInfo(native, contract, authorizeInfo)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putAuthorizeInfo, put authorizeInfo error: %v", err)
	}

	//remove peerPubkey from peerPool
	delete(peerPoolMap.PeerPoolMap, params.PeerPubkey)
	err = putPeerPoolMap(native, contract, view, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putPeerPoolMap, put peerPoolMap error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

//Put a node into black list, remove node from pool
//Whole of initPos of black node will be punished, and several percent of authorize deposit will be punished too.
//Node in black list can't be registered.
func BlackNode(native *native.NativeService) ([]byte, error) {
	params := new(BlackNodeParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deserialize, contract params deserialize error: %v", err)
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
		return utils.BYTE_FALSE, fmt.Errorf("blackNode, checkWitness error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getView, get view error: %v", err)
	}
	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getPeerPoolMap, get peerPoolMap error: %v", err)
	}
	commit := false
	for _, peerPubkey := range params.PeerPubkeyList {
		peerPubkeyPrefix, err := hex.DecodeString(peerPubkey)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("hex.DecodeString, peerPubkey format error: %v", err)
		}
		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peerPubkey]
		if !ok {
			return utils.BYTE_FALSE, fmt.Errorf("blackNode, peerPubkey is not in peerPoolMap")
		}

		blackListItem := &BlackListItem{
			PeerPubkey: peerPoolItem.PeerPubkey,
			Address:    peerPoolItem.Address,
			InitPos:    peerPoolItem.InitPos,
		}
		bf := new(bytes.Buffer)
		if err := blackListItem.Serialize(bf); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("serialize, serialize blackListItem error: %v", err)
		}
		//put peer into black list
		native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(BLACK_LIST), peerPubkeyPrefix), &cstates.StorageItem{Value: bf.Bytes()})
		//change peerPool status
		if peerPoolItem.Status == ConsensusStatus {
			peerPoolItem.Status = BlackStatus
			peerPoolMap.PeerPoolMap[peerPubkey] = peerPoolItem
			err = putPeerPoolMap(native, contract, view, peerPoolMap)
			if err != nil {
				return utils.BYTE_FALSE, fmt.Errorf("putPeerPoolMap, put peerPoolMap error: %v", err)
			}
			commit = true
		} else {
			peerPoolItem.Status = BlackStatus
			peerPoolMap.PeerPoolMap[peerPubkey] = peerPoolItem
			err = putPeerPoolMap(native, contract, view, peerPoolMap)
			if err != nil {
				return utils.BYTE_FALSE, fmt.Errorf("putPeerPoolMap, put peerPoolMap error: %v", err)
			}
		}
	}
	//commitDpos
	if commit {
		// get config
		config, err := getConfig(native, contract)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("getConfig, get config error: %v", err)
		}
		err = executeCommitDpos(native, contract, config)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("executeCommitDpos, executeCommitDpos error: %v", err)
		}
	}
	return utils.BYTE_TRUE, nil
}

//Remove a node from black list, allow it to be registered
func WhiteNode(native *native.NativeService) ([]byte, error) {
	params := new(WhiteNodeParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deserialize, contract params deserialize error: %v", err)
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
		return utils.BYTE_FALSE, fmt.Errorf("whiteNode, checkWitness error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	peerPubkeyPrefix, err := hex.DecodeString(params.PeerPubkey)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("hex.DecodeString, peerPubkey format error: %v", err)
	}

	//check black list
	blackListBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(BLACK_LIST), peerPubkeyPrefix))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("native.CloneCache.Get, get BlackList error: %v", err)
	}
	if blackListBytes == nil {
		return utils.BYTE_FALSE, fmt.Errorf("whiteNode, this Peer is not in BlackList: %v", err)
	}

	//remove peer from black list
	native.CloneCache.Delete(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(BLACK_LIST), peerPubkeyPrefix))

	return utils.BYTE_TRUE, nil
}

//Quit a registered node, used by node owner.
//Remove node from pool and unfreeze deposit next epoch(candidate node) / next next epoch(consensus node)
func QuitNode(native *native.NativeService) ([]byte, error) {
	params := new(QuitNodeParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deserialize, contract params deserialize error: %v", err)
	}
	address := params.Address

	//check witness
	err := utils.ValidateOwner(native, address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("validateOwner, checkWitness error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getView, get view error: %v", err)
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getPeerPoolMap, get peerPoolMap error: %v", err)
	}

	//get config
	config, err := getConfig(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getConfig, get config error: %v", err)
	}

	peerPoolItem, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, fmt.Errorf("quitNode, peerPubkey is not in peerPoolMap")
	}

	if address != peerPoolItem.Address {
		return utils.BYTE_FALSE, fmt.Errorf("quitNode, peerPubkey is not registered by this address")
	}
	if peerPoolItem.Status != ConsensusStatus && peerPoolItem.Status != CandidateStatus {
		return utils.BYTE_FALSE, fmt.Errorf("quitNode, peerPubkey is not CandidateStatus or ConsensusStatus")
	}

	//check peers num
	num := 0
	for _, peerPoolItem := range peerPoolMap.PeerPoolMap {
		if peerPoolItem.Status == CandidateStatus || peerPoolItem.Status == ConsensusStatus {
			num = num + 1
		}
	}
	if num <= int(config.K) {
		return utils.BYTE_FALSE, fmt.Errorf("quitNode, num of peers is less than K")
	}

	//change peerPool status
	if peerPoolItem.Status == ConsensusStatus {
		peerPoolItem.Status = QuitConsensusStatus
	} else {
		peerPoolItem.Status = QuitingStatus
	}

	peerPoolMap.PeerPoolMap[params.PeerPubkey] = peerPoolItem
	err = putPeerPoolMap(native, contract, view, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putPeerPoolMap, put peerPoolMap error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

//Authorize for a node by depositing ONT in this governance contract, used by users
func AuthorizeForPeer(native *native.NativeService) ([]byte, error) {
	err := authorizeForPeer(native, "transfer")
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("authorizeForPeer error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

//Authorize for a node by depositing ONT in this governance contract, used by contracts
func AuthorizeForPeerTransferFrom(native *native.NativeService) ([]byte, error) {
	err := authorizeForPeer(native, "transferFrom")
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("authorizeForPeerTransferFrom error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

//UnAuthorize for a node by redeeming ONT from this governance contract
func UnAuthorizeForPeer(native *native.NativeService) ([]byte, error) {
	params := &AuthorizeForPeerParam{
		PeerPubkeyList: make([]string, 0),
		PosList:        make([]uint32, 0),
	}
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deserialize, contract params deserialize error: %v", err)
	}
	address := params.Address

	//check witness
	err := utils.ValidateOwner(native, address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("validateOwner, checkWitness error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getView, get view error: %v", err)
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getPeerPoolMap, get peerPoolMap error: %v", err)
	}

	//get globalParam2
	globalParam2, err := getGlobalParam2(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getGlobalParam2, getGlobalParam2 error: %v", err)
	}

	for i := 0; i < len(params.PeerPubkeyList); i++ {
		peerPubkey := params.PeerPubkeyList[i]
		pos := params.PosList[i]

		//check pos
		if pos < globalParam2.MinAuthorizePos || pos%globalParam2.MinAuthorizePos != 0 {
			return utils.BYTE_FALSE, fmt.Errorf("unAuthorizeForPeer, pos must be times of %d", globalParam2.MinAuthorizePos)
		}

		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peerPubkey]
		if !ok {
			return utils.BYTE_FALSE, fmt.Errorf("unAuthorizeForPeer, peerPubkey is not in peerPoolMap")
		}

		if peerPoolItem.Status != CandidateStatus && peerPoolItem.Status != ConsensusStatus {
			return utils.BYTE_FALSE, fmt.Errorf("unAuthorizeForPeer, peerPubkey is not candidate and can not be authorized")
		}

		authorizeInfo, err := getAuthorizeInfo(native, contract, peerPubkey, address)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("getAuthorizeInfo, get authorizeInfo error: %v", err)
		}
		if authorizeInfo.NewPos < uint64(pos) {
			if peerPoolItem.Status == ConsensusStatus {
				if authorizeInfo.ConsensusPos < (uint64(pos) - authorizeInfo.NewPos) {
					return utils.BYTE_FALSE, fmt.Errorf("unAuthorizeForPeer, your pos of this peerPubkey is not enough")
				}
				consensusPos := authorizeInfo.ConsensusPos + authorizeInfo.NewPos - uint64(pos)
				newPos := authorizeInfo.NewPos
				authorizeInfo.NewPos = 0
				authorizeInfo.WithdrawUnfreezePos = authorizeInfo.WithdrawUnfreezePos + newPos
				authorizeInfo.ConsensusPos = consensusPos
				authorizeInfo.WithdrawPos = authorizeInfo.WithdrawPos + uint64(pos) - authorizeInfo.NewPos
				peerPoolItem.TotalPos = peerPoolItem.TotalPos - uint64(pos)
			}
			if peerPoolItem.Status == CandidateStatus {
				if authorizeInfo.FreezePos < (uint64(pos) - authorizeInfo.NewPos) {
					return utils.BYTE_FALSE, fmt.Errorf("unAuthorizeForPeer, your pos of this peerPubkey is not enough")
				}
				freezePos := authorizeInfo.FreezePos + authorizeInfo.NewPos - uint64(pos)
				newPos := authorizeInfo.NewPos
				authorizeInfo.NewPos = 0
				authorizeInfo.WithdrawUnfreezePos = authorizeInfo.WithdrawUnfreezePos + newPos
				authorizeInfo.FreezePos = freezePos
				authorizeInfo.WithdrawFreezePos = authorizeInfo.WithdrawFreezePos + uint64(pos) - authorizeInfo.NewPos
				peerPoolItem.TotalPos = peerPoolItem.TotalPos - uint64(pos)
			}
		} else {
			temp := authorizeInfo.NewPos - uint64(pos)
			authorizeInfo.NewPos = temp
			authorizeInfo.WithdrawUnfreezePos = authorizeInfo.WithdrawUnfreezePos + uint64(pos)
			peerPoolItem.TotalPos = peerPoolItem.TotalPos - uint64(pos)
		}

		peerPoolMap.PeerPoolMap[peerPubkey] = peerPoolItem
		err = putAuthorizeInfo(native, contract, authorizeInfo)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("putAuthorizeInfo, put authorizeInfo error: %v", err)
		}
	}
	err = putPeerPoolMap(native, contract, view, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putPeerPoolMap, put peerPoolMap error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

//Withdraw unfreezed ONT deposited in this governance contract.
func Withdraw(native *native.NativeService) ([]byte, error) {
	params := &WithdrawParam{
		PeerPubkeyList: make([]string, 0),
		WithdrawList:   make([]uint32, 0),
	}
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deserialize, contract params deserialize error: %v", err)
	}
	address := params.Address

	//check witness
	err := utils.ValidateOwner(native, address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("validateOwner, checkWitness error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	var total uint64
	for i := 0; i < len(params.PeerPubkeyList); i++ {
		peerPubkey := params.PeerPubkeyList[i]
		pos := params.WithdrawList[i]
		peerPubkeyPrefix, err := hex.DecodeString(peerPubkey)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("hex.DecodeString, peerPubkey format error: %v", err)
		}

		authorizeInfo, err := getAuthorizeInfo(native, contract, peerPubkey, address)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("getAuthorizeInfo, get authorizeInfo error: %v", err)
		}
		if authorizeInfo.WithdrawUnfreezePos < uint64(pos) {
			return utils.BYTE_FALSE, fmt.Errorf("withdraw, your unfreeze withdraw pos of this peerPubkey is not enough")
		} else {
			authorizeInfo.WithdrawUnfreezePos = authorizeInfo.WithdrawUnfreezePos - uint64(pos)
			total = total + uint64(pos)
			err = putAuthorizeInfo(native, contract, authorizeInfo)
			if err != nil {
				return utils.BYTE_FALSE, fmt.Errorf("putAuthorizeInfo, put authorizeInfo error: %v", err)
			}
		}
		if authorizeInfo.ConsensusPos == 0 && authorizeInfo.FreezePos == 0 && authorizeInfo.NewPos == 0 &&
			authorizeInfo.WithdrawPos == 0 && authorizeInfo.WithdrawFreezePos == 0 && authorizeInfo.WithdrawUnfreezePos == 0 {
			native.CloneCache.Delete(scommon.ST_STORAGE, utils.ConcatKey(contract, AUTHORIZE_INFO_POOL, peerPubkeyPrefix, address[:]))
		}
	}

	//ont transfer
	err = appCallTransferOnt(native, utils.GovernanceContractAddress, address, total)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("appCallTransferOnt, ont transfer error: %v", err)
	}

	//update total stake
	err = withdrawTotalStake(native, contract, address, total)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("withdrawTotalStake, withdrawTotalStake error: %v", err)
	}

	return utils.BYTE_TRUE, nil
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

	err = executeCommitDpos(native, contract, config)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("executeCommitDpos, executeCommitDpos error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

//Update VBFT config
func UpdateConfig(native *native.NativeService) ([]byte, error) {
	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getAdmin, get admin error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("updateConfig, checkWitness error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get globalParam
	globalParam, err := getGlobalParam(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getGlobalParam, getGlobalParam error: %v", err)
	}

	configuration := new(Configuration)
	if err := configuration.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deserialize, deserialize configuration error: %v", err)
	}

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getView, get view error: %v", err)
	}
	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getPeerPoolMap, get peerPoolMap error: %v", err)
	}
	candidateNum := 0
	for _, peerPoolItem := range peerPoolMap.PeerPoolMap {
		if peerPoolItem.Status == CandidateStatus || peerPoolItem.Status == ConsensusStatus {
			candidateNum = candidateNum + 1
		}
	}

	//check the configuration
	if configuration.C == 0 {
		return utils.BYTE_FALSE, fmt.Errorf("updateConfig. C can not be 0 in config")
	}
	if int(configuration.K) > candidateNum {
		return utils.BYTE_FALSE, fmt.Errorf("updateConfig. K can not be larger than num of candidate peer in config")
	}
	if configuration.L < 16*configuration.K || configuration.L%configuration.K != 0 {
		return utils.BYTE_FALSE, fmt.Errorf("updateConfig. L can not be less than 16*K and K must be times of L in config")
	}
	if configuration.K < 2*configuration.C+1 {
		return utils.BYTE_FALSE, fmt.Errorf("updateConfig. K can not be less than 2*C+1 in config")
	}
	if 4*configuration.K > globalParam.CandidateNum {
		return utils.BYTE_FALSE, fmt.Errorf("updateConfig. 4*K can not be more than candidateNum")
	}
	if configuration.N < configuration.K || configuration.K < 7 {
		return utils.BYTE_FALSE, fmt.Errorf("updateConfig. config not match N >= K >= 7")
	}
	if configuration.BlockMsgDelay < 5000 {
		return utils.BYTE_FALSE, fmt.Errorf("updateConfig. BlockMsgDelay must >= 5000")
	}
	if configuration.HashMsgDelay < 5000 {
		return utils.BYTE_FALSE, fmt.Errorf("updateConfig. HashMsgDelay must >= 5000")
	}
	if configuration.PeerHandshakeTimeout < 10 {
		return utils.BYTE_FALSE, fmt.Errorf("updateConfig. PeerHandshakeTimeout must >= 10")
	}
	err = putConfig(native, contract, configuration)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putConfig, put config error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

//Update global params of this governance contract
func UpdateGlobalParam(native *native.NativeService) ([]byte, error) {
	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getAdmin, get admin error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("updateGlobalParam, checkWitness error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	// get config
	config, err := getConfig(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getConfig, get config error: %v", err)
	}

	globalParam := new(GlobalParam)
	if err := globalParam.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deserialize, deserialize globalParam error: %v", err)
	}

	//check the globalParam
	if (globalParam.A + globalParam.B) != 100 {
		return utils.BYTE_FALSE, fmt.Errorf("updateGlobalParam. A + B must equal to 100")
	}
	if globalParam.Yita == 0 {
		return utils.BYTE_FALSE, fmt.Errorf("updateGlobalParam. Yita must > 0")
	}
	if globalParam.Penalty > 100 {
		return utils.BYTE_FALSE, fmt.Errorf("updateGlobalParam. Penalty must <= 100")
	}
	if globalParam.PosLimit < 1 {
		return utils.BYTE_FALSE, fmt.Errorf("updateGlobalParam. PosLimit must >= 1")
	}
	if globalParam.CandidateNum < 4*config.K {
		return utils.BYTE_FALSE, fmt.Errorf("updateGlobalParam. CandidateNum must >= 4*K")
	}
	if globalParam.CandidateFee != 0 && globalParam.CandidateFee < MIN_CANDIDATE_FEE {
		return utils.BYTE_FALSE, fmt.Errorf("updateGlobalParam. CandidateFee must >= %d", MIN_CANDIDATE_FEE)
	}
	err = putGlobalParam(native, contract, globalParam)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putGlobalParam, put globalParam error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

//Update global params of this governance contract
func UpdateGlobalParam2(native *native.NativeService) ([]byte, error) {
	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getAdmin, get admin error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("updateGlobalParam2, checkWitness error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	globalParam2 := new(GlobalParam2)
	if err := globalParam2.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deserialize, deserialize globalParam2 error: %v", err)
	}

	//check the globalParam
	if globalParam2.MinAuthorizePos == 0 {
		return utils.BYTE_FALSE, fmt.Errorf("globalParam2.MinAuthorizePos can not be 0")
	}
	// get config
	config, err := getConfig(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getConfig, get config error: %v", err)
	}
	if globalParam2.CandidateFeeSplitNum < config.K {
		return utils.BYTE_FALSE, fmt.Errorf("globalParam2.CandidateFeeSplitNum can not be less than config.K")
	}

	err = putGlobalParam2(native, contract, globalParam2)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putGlobalParam2, put globalParam2 error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

//Update split curve
func UpdateSplitCurve(native *native.NativeService) ([]byte, error) {
	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getAdmin, get admin error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("updateSplitCurve, checkWitness error: %v", err)
	}

	splitCurve := new(SplitCurve)
	if err := splitCurve.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deserialize, deserialize splitCurve error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	err = putSplitCurve(native, contract, splitCurve)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putSplitCurve, put splitCurve error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

//Transfer all punished ONT of a black node to a certain address
func TransferPenalty(native *native.NativeService) ([]byte, error) {
	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getAdmin, get admin error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("transferPenalty, checkWitness error: %v", err)
	}

	param := new(TransferPenaltyParam)
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deserialize, deserialize transferPenaltyParam error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	err = withdrawPenaltyStake(native, contract, param.PeerPubkey, param.Address)
	if err != nil {
		return nil, fmt.Errorf("withdrawPenaltyStake, withdraw penaltyStake error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

//Withdraw unbounded ONG according to deposit ONT in this governance contract
func WithdrawOng(native *native.NativeService) ([]byte, error) {
	params := new(WithdrawOngParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deserialize, deserialize transferPenaltyParam error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("withdrawOng, checkWitness error: %v", err)
	}

	// ont transfer to trigger unboundong
	err = appCallTransferOnt(native, utils.GovernanceContractAddress, utils.GovernanceContractAddress, 1)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("appCallTransferOnt, ont transfer error: %v", err)
	}

	totalStake, err := getTotalStake(native, contract, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getTotalStake, get totalStake error: %v", err)
	}

	preTimeOffset := totalStake.TimeOffset
	timeOffset := native.Time - constants.GENESIS_BLOCK_TIMESTAMP

	amount := utils.CalcUnbindOng(totalStake.Stake, preTimeOffset, timeOffset)
	err = appCallTransferFromOng(native, utils.GovernanceContractAddress, utils.OntContractAddress, totalStake.Address, amount)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("appCallTransferFromOng, transfer from ong error: %v", err)
	}

	totalStake.TimeOffset = timeOffset

	err = putTotalStake(native, contract, totalStake)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putTotalStake, put totalStake error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

//Change the status if node can receive authorization from ont holders
func ChangeMaxAuthorization(native *native.NativeService) ([]byte, error) {
	params := new(ChangeMaxAuthorizationParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deserialize, deserialize changeMaxAuthorizationParam error: %v", err)
	}

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("validateOwner, checkWitness error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//check if is peer owner
	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getView, get view error: %v", err)
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getPeerPoolMap, get peerPoolMap error: %v", err)
	}

	peerPoolItem, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, fmt.Errorf("changeMaxAuthorization, peerPubkey is not in peerPoolMap")
	}
	if peerPoolItem.Address != params.Address {
		return utils.BYTE_FALSE, fmt.Errorf("address is not peer owner")
	}

	peerAttributes, err := getPeerAttributes(native, contract, params.PeerPubkey)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getPeerAttributes error: %v", err)
	}
	peerAttributes.MaxAuthorize = params.MaxAuthorize

	err = putPeerAttributes(native, contract, peerAttributes)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putPeerAttributes error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

//Set node cost, node can take some percentage of fee before split
func SetPeerCost(native *native.NativeService) ([]byte, error) {
	params := new(SetPeerCostParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deserialize, deserialize setPeerCostParam error: %v", err)
	}
	if params.PeerCost >= 100 {
		return utils.BYTE_FALSE, fmt.Errorf("peerCost must >= 0 and <= 100")
	}

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("validateOwner, checkWitness error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//check if is peer owner
	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getView, get view error: %v", err)
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getPeerPoolMap, get peerPoolMap error: %v", err)
	}
	peerPoolItem, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, fmt.Errorf("setPeerCost, peerPubkey is not in peerPoolMap")
	}
	if peerPoolItem.Address != params.Address {
		return utils.BYTE_FALSE, fmt.Errorf("address is not peer owner")
	}

	peerAttributes, err := getPeerAttributes(native, contract, params.PeerPubkey)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getPeerAttributes error: %v", err)
	}
	newPeerCost := peerAttributes.NewPeerCost
	//check set cost view
	if view-peerAttributes.SetCostView >= 2 {
		peerAttributes.OldPeerCost = newPeerCost
	}
	peerAttributes.NewPeerCost = uint64(params.PeerCost)
	peerAttributes.SetCostView = view

	err = putPeerAttributes(native, contract, peerAttributes)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putPeerAttributes error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

//Withdraw split fee of address
func WithdrawFee(native *native.NativeService) ([]byte, error) {
	params := new(WithdrawFeeParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deserialize, deserialize withdrawFeeParam error: %v", err)
	}

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("validateOwner, checkWitness error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	splitFeeAddress, err := getSplitFeeAddress(native, contract, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getSplitFeeAddress, getSplitFeeAddress error: %v", err)
	}
	fee := splitFeeAddress.Amount

	//ong transfer
	err = appCallTransferOng(native, utils.GovernanceContractAddress, params.Address, fee)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("appCallTransferOng, ong transfer error: %v", err)
	}

	//delete from splitFeeAddress
	native.CloneCache.Delete(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(SPLIT_FEE_ADDRESS), params.Address[:]))

	//update splitFee
	splitFee, err := getSplitFee(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getSplitFee, getSplitFee error: %v", err)
	}
	if splitFee < fee {
		return utils.BYTE_FALSE, fmt.Errorf("withdrawFee, splitFee is not enough")
	}
	newSplitFee := splitFee - fee
	err = putSplitFee(native, contract, newSplitFee)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putSplitFee, put splitFee error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

//add init pos of a node
func AddInitPos(native *native.NativeService) ([]byte, error) {
	params := new(ChangeInitPosParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deserialize, deserialize changeInitPosParam error: %v", err)
	}

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("validateOwner, checkWitness error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//check if is peer owner
	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getView, get view error: %v", err)
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getPeerPoolMap, get peerPoolMap error: %v", err)
	}
	peerPoolItem, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, fmt.Errorf("addInitPos, peerPubkey is not in peerPoolMap")
	}
	if peerPoolItem.Address != params.Address {
		return utils.BYTE_FALSE, fmt.Errorf("address is not peer owner")
	}

	peerPoolMap.PeerPoolMap[params.PeerPubkey].InitPos = peerPoolItem.InitPos + uint64(params.Pos)
	err = putPeerPoolMap(native, contract, view, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putPeerPoolMap error: %v", err)
	}

	//ont transfer
	err = appCallTransferOnt(native, params.Address, utils.GovernanceContractAddress, uint64(params.Pos))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("appCallTransferOnt, ont transfer error: %v", err)
	}

	//update total stake
	err = depositTotalStake(native, contract, params.Address, uint64(params.Pos))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("depositTotalStake, depositTotalStake error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

//reduce init pos of a node
func ReduceInitPos(native *native.NativeService) ([]byte, error) {
	params := new(ChangeInitPosParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deserialize, deserialize changeInitPosParam error: %v", err)
	}

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("validateOwner, checkWitness error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//check if is peer owner
	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getView, get view error: %v", err)
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getPeerPoolMap, get peerPoolMap error: %v", err)
	}
	peerPoolItem, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, fmt.Errorf("reduceInitPos, peerPubkey is not in peerPoolMap")
	}
	if peerPoolItem.Address != params.Address {
		return utils.BYTE_FALSE, fmt.Errorf("address is not peer owner")
	}
	newInitPos := peerPoolItem.InitPos - uint64(params.Pos)
	//get globalParam
	globalParam, err := getGlobalParam(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getGlobalParam, getGlobalParam error: %v", err)
	}
	if newInitPos < peerPoolItem.TotalPos/uint64(globalParam.PosLimit) {
		return utils.BYTE_FALSE, fmt.Errorf("initPos must more than totalPos/posLimit")
	}
	//get promise pos
	promisePos, err := getPromisePos(native, contract, params.PeerPubkey)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getPromisePos, getPromisePos error: %v", err)
	}
	if newInitPos < uint64(promisePos.PromisePos) {
		return utils.BYTE_FALSE, fmt.Errorf("initPos must more than promise pos")
	}

	peerPoolMap.PeerPoolMap[params.PeerPubkey].InitPos = newInitPos
	err = putPeerPoolMap(native, contract, view, peerPoolMap)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putPeerPoolMap error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

//set promise pos of a node
func SetPromisePos(native *native.NativeService) ([]byte, error) {
	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getAdmin, get admin error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("setPromisePos, checkWitness error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	promisePos := new(PromisePos)
	if err := promisePos.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deserialize, contract params deserialize error: %v", err)
	}
	//update promise pos
	err = putPromisePos(native, contract, promisePos)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("putPromisePos, put promisePos error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}
