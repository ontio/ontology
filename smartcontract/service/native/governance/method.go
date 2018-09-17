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
	"sort"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/constants"
	cstates "github.com/ontio/ontology/core/states"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func registerCandidate(native *native.NativeService, flag string) error {
	params := new(RegisterCandidateParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//check auth of OntID
	err := appCallVerifyToken(native, contract, params.Caller, REGISTER_CANDIDATE, uint64(params.KeyNo))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallVerifyToken, verifyToken failed!")
	}

	//check witness
	err = utils.ValidateOwner(native, params.Address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "validateOwner, checkWitness error!")
	}

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getView, get view error!")
	}

	//check peerPubkey
	if err := validatePeerPubKeyFormat(params.PeerPubkey); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "invalid peer pubkey")
	}

	peerPubkeyPrefix, err := hex.DecodeString(params.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	//get black list
	blackList, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(BLACK_LIST), peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Get, get BlackList error!")
	}
	if blackList != nil {
		return errors.NewErr("registerCandidate, this Peer is in BlackList!")
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}

	//check if exist in PeerPool
	_, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if ok {
		return errors.NewErr("registerCandidate, peerPubkey is already in peerPoolMap!")
	}

	peerPoolItem := &PeerPoolItem{
		PeerPubkey: params.PeerPubkey,
		Address:    params.Address,
		InitPos:    uint64(params.InitPos),
		Status:     RegisterCandidateStatus,
	}
	peerPoolMap.PeerPoolMap[params.PeerPubkey] = peerPoolItem
	err = putPeerPoolMap(native, contract, view, peerPoolMap)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "putPeerPoolMap, put peerPoolMap error!")
	}

	//get globalParam
	globalParam, err := getGlobalParam(native, contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getGlobalParam, getGlobalParam error!")
	}

	switch flag {
	case "transfer":
		//ont transfer
		err = appCallTransferOnt(native, params.Address, utils.GovernanceContractAddress, uint64(params.InitPos))
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOnt, ont transfer error!")
		}

		//ong transfer
		err = appCallTransferOng(native, params.Address, utils.GovernanceContractAddress, globalParam.CandidateFee)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOng, ong transfer error!")
		}
	case "transferFrom":
		//ont transfer from
		err = appCallTransferFromOnt(native, utils.GovernanceContractAddress, params.Address, utils.GovernanceContractAddress, uint64(params.InitPos))
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferFromOnt, ont transfer error!")
		}

		//ong transfer from
		err = appCallTransferFromOng(native, utils.GovernanceContractAddress, params.Address, utils.GovernanceContractAddress, globalParam.CandidateFee)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferFromOng, ong transfer error!")
		}
	}

	//update total stake
	err = depositTotalStake(native, contract, params.Address, uint64(params.InitPos))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "depositTotalStake, depositTotalStake error!")
	}
	return nil
}

func authorizeForPeer(native *native.NativeService, flag string) error {
	params := &AuthorizeForPeerParam{
		PeerPubkeyList: make([]string, 0),
		PosList:        make([]uint32, 0),
	}
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, contract params deserialize error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "validateOwner, checkWitness error!")
	}

	//get current view
	view, err := GetView(native, contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getView, get view error!")
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}

	//get globalParam
	globalParam, err := getGlobalParam(native, contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getGlobalParam, getGlobalParam error!")
	}

	var total uint64
	for i := 0; i < len(params.PeerPubkeyList); i++ {
		peerPubkey := params.PeerPubkeyList[i]
		pos := params.PosList[i]

		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peerPubkey]
		if !ok {
			return errors.NewErr("authorizeForPeer, peerPubkey is not in peerPoolMap!")
		}

		if peerPoolItem.Status != CandidateStatus && peerPoolItem.Status != ConsensusStatus {
			return errors.NewErr("authorizeForPeer, peerPubkey is not candidate and can not be authorized!")
		}

		authorizeInfo, err := getAuthorizeInfo(native, contract, peerPubkey, params.Address)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "getAuthorizeInfo, get authorizeInfo error!")
		}
		authorizeInfo.NewPos = authorizeInfo.NewPos + uint64(pos)
		total = total + uint64(pos)
		peerPoolItem.TotalPos = peerPoolItem.TotalPos + uint64(pos)
		if peerPoolItem.TotalPos > uint64(globalParam.PosLimit)*peerPoolItem.InitPos {
			return errors.NewErr("authorizeForPeer, pos of this peer is full!")
		}

		peerPoolMap.PeerPoolMap[peerPubkey] = peerPoolItem
		err = putAuthorizeInfo(native, contract, authorizeInfo)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "putAuthorizeInfo, put authorizeInfo error!")
		}
	}
	err = putPeerPoolMap(native, contract, view, peerPoolMap)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "putPeerPoolMap, put peerPoolMap error!")
	}

	switch flag {
	case "transfer":
		//ont transfer
		err = appCallTransferOnt(native, params.Address, utils.GovernanceContractAddress, total)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOnt, ont transfer error!")
		}
	case "transferFrom":
		//ont transfer from
		err = appCallTransferFromOnt(native, utils.GovernanceContractAddress, params.Address, utils.GovernanceContractAddress, total)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferFromOnt, ont transfer error!")
		}
	}

	//update total stake
	err = depositTotalStake(native, contract, params.Address, total)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "depositTotalStake, depositTotalStake error!")
	}

	return nil
}

func normalQuit(native *native.NativeService, contract common.Address, peerPoolItem *PeerPoolItem) error {
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	flag := false
	//draw back authorize pos
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, AUTHORIZE_INFO_POOL, peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
	}
	authorizeInfo := new(AuthorizeInfo)
	for _, v := range stateValues {
		authorizeInfoStore, ok := v.Value.(*cstates.StorageItem)
		if !ok {
			return errors.NewErr("authorizeInfoStore is not available!")
		}
		if err := authorizeInfo.Deserialize(bytes.NewBuffer(authorizeInfoStore.Value)); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize authorizeInfo error!")
		}
		authorizeInfo.WithdrawUnfreezePos = authorizeInfo.ConsensusPos + authorizeInfo.FreezePos + authorizeInfo.NewPos + authorizeInfo.WithdrawPos +
			authorizeInfo.WithdrawFreezePos + authorizeInfo.WithdrawUnfreezePos
		authorizeInfo.ConsensusPos = 0
		authorizeInfo.FreezePos = 0
		authorizeInfo.NewPos = 0
		authorizeInfo.WithdrawPos = 0
		authorizeInfo.WithdrawFreezePos = 0
		if authorizeInfo.Address == peerPoolItem.Address {
			flag = true
			authorizeInfo.WithdrawUnfreezePos = authorizeInfo.WithdrawUnfreezePos + peerPoolItem.InitPos
		}
		err = putAuthorizeInfo(native, contract, authorizeInfo)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "putAuthorizeInfo, put authorizeInfo error!")
		}
	}
	if flag == false {
		authorizeInfo := &AuthorizeInfo{
			PeerPubkey:          peerPoolItem.PeerPubkey,
			Address:             peerPoolItem.Address,
			WithdrawUnfreezePos: peerPoolItem.InitPos,
		}
		err = putAuthorizeInfo(native, contract, authorizeInfo)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "putAuthorizeInfo, put authorizeInfo error!")
		}
	}
	return nil
}

func blackQuit(native *native.NativeService, contract common.Address, peerPoolItem *PeerPoolItem) error {
	// ont transfer to trigger unboundong
	err := appCallTransferOnt(native, utils.GovernanceContractAddress, utils.GovernanceContractAddress, peerPoolItem.InitPos)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOnt, ont transfer error!")
	}

	//update total stake
	err = withdrawTotalStake(native, contract, peerPoolItem.Address, peerPoolItem.InitPos)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "withdrawTotalStake, withdrawTotalStake error!")
	}

	initPos := peerPoolItem.InitPos
	var authorizePos uint64

	//get globalParam
	globalParam, err := getGlobalParam(native, contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getGlobalParam, getGlobalParam error!")
	}

	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	//draw back authorize pos
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, AUTHORIZE_INFO_POOL, peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
	}
	authorizeInfo := new(AuthorizeInfo)
	for _, v := range stateValues {
		authorizeInfoStore, ok := v.Value.(*cstates.StorageItem)
		if !ok {
			return errors.NewErr("authorizeInfoStore is not available!")
		}
		if err := authorizeInfo.Deserialize(bytes.NewBuffer(authorizeInfoStore.Value)); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize authorizeInfo error!")
		}
		total := authorizeInfo.ConsensusPos + authorizeInfo.FreezePos + authorizeInfo.NewPos + authorizeInfo.WithdrawPos + authorizeInfo.WithdrawFreezePos
		penalty := (uint64(globalParam.Penalty)*total + 99) / 100
		authorizeInfo.WithdrawUnfreezePos = total - penalty + authorizeInfo.WithdrawUnfreezePos
		authorizeInfo.ConsensusPos = 0
		authorizeInfo.FreezePos = 0
		authorizeInfo.NewPos = 0
		authorizeInfo.WithdrawPos = 0
		authorizeInfo.WithdrawFreezePos = 0
		address := authorizeInfo.Address
		err = putAuthorizeInfo(native, contract, authorizeInfo)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "putAuthorizeInfo, put authorizeInfo error!")
		}

		//update total stake
		err = withdrawTotalStake(native, contract, address, penalty)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "withdrawTotalStake, withdrawTotalStake error!")
		}
		authorizePos = authorizePos + penalty
	}

	//add penalty stake
	err = depositPenaltyStake(native, contract, peerPoolItem.PeerPubkey, initPos, authorizePos)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "depositPenaltyStake, deposit penaltyStake error!")
	}
	return nil
}

func consensusToConsensus(native *native.NativeService, contract common.Address, peerPoolItem *PeerPoolItem) error {
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	//update authorizeInfoPool
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, AUTHORIZE_INFO_POOL, peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
	}
	authorizeInfo := new(AuthorizeInfo)
	for _, v := range stateValues {
		authorizeInfoStore, ok := v.Value.(*cstates.StorageItem)
		if !ok {
			return errors.NewErr("authorizeInfoStore is not available!")
		}
		if err := authorizeInfo.Deserialize(bytes.NewBuffer(authorizeInfoStore.Value)); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize authorizeInfo error!")
		}
		if authorizeInfo.FreezePos != 0 {
			return errors.NewErr("commitPos, freezePos should be 0!")
		}
		newPos := authorizeInfo.NewPos
		authorizeInfo.ConsensusPos = authorizeInfo.ConsensusPos + newPos
		authorizeInfo.NewPos = 0
		withdrawPos := authorizeInfo.WithdrawPos
		withdrawFreezePos := authorizeInfo.WithdrawFreezePos
		authorizeInfo.WithdrawFreezePos = withdrawPos
		authorizeInfo.WithdrawUnfreezePos = authorizeInfo.WithdrawUnfreezePos + withdrawFreezePos
		authorizeInfo.WithdrawPos = 0

		err = putAuthorizeInfo(native, contract, authorizeInfo)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "putAuthorizeInfo, put authorizeInfo error!")
		}
	}
	return nil
}

func unConsensusToConsensus(native *native.NativeService, contract common.Address, peerPoolItem *PeerPoolItem) error {
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	//update authorizeInfoPool
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, AUTHORIZE_INFO_POOL, peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
	}
	authorizeInfo := new(AuthorizeInfo)
	for _, v := range stateValues {
		authorizeInfoStore, ok := v.Value.(*cstates.StorageItem)
		if !ok {
			return errors.NewErr("authorizeInfoStore is not available!")
		}
		if err := authorizeInfo.Deserialize(bytes.NewBuffer(authorizeInfoStore.Value)); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize authorizeInfo error!")
		}
		if authorizeInfo.ConsensusPos != 0 {
			return errors.NewErr("consensusPos, freezePos should be 0!")
		}

		authorizeInfo.ConsensusPos = authorizeInfo.ConsensusPos + authorizeInfo.FreezePos + authorizeInfo.NewPos
		authorizeInfo.NewPos = 0
		authorizeInfo.FreezePos = 0
		withdrawPos := authorizeInfo.WithdrawPos
		withdrawFreezePos := authorizeInfo.WithdrawFreezePos
		authorizeInfo.WithdrawFreezePos = withdrawPos
		authorizeInfo.WithdrawUnfreezePos = authorizeInfo.WithdrawUnfreezePos + withdrawFreezePos
		authorizeInfo.WithdrawPos = 0

		err = putAuthorizeInfo(native, contract, authorizeInfo)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "putAuthorizeInfo, put authorizeInfo error!")
		}
	}
	return nil
}

func consensusToUnConsensus(native *native.NativeService, contract common.Address, peerPoolItem *PeerPoolItem) error {
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	//update authorizeInfoPool
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, AUTHORIZE_INFO_POOL, peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
	}
	authorizeInfo := new(AuthorizeInfo)
	for _, v := range stateValues {
		authorizeInfoStore, ok := v.Value.(*cstates.StorageItem)
		if !ok {
			return errors.NewErr("authorizeInfoStore is not available!")
		}
		if err := authorizeInfo.Deserialize(bytes.NewBuffer(authorizeInfoStore.Value)); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize authorizeInfo error!")
		}
		if authorizeInfo.FreezePos != 0 {
			return errors.NewErr("commitPos, freezePos should be 0!")
		}

		authorizeInfo.FreezePos = authorizeInfo.ConsensusPos + authorizeInfo.NewPos
		authorizeInfo.NewPos = 0
		authorizeInfo.ConsensusPos = 0
		withdrawPos := authorizeInfo.WithdrawPos
		withdrawFreezePos := authorizeInfo.WithdrawFreezePos
		authorizeInfo.WithdrawFreezePos = withdrawPos
		authorizeInfo.WithdrawUnfreezePos = authorizeInfo.WithdrawUnfreezePos + withdrawFreezePos
		authorizeInfo.WithdrawPos = 0

		err = putAuthorizeInfo(native, contract, authorizeInfo)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "putAuthorizeInfo, put authorizeInfo error!")
		}
	}
	return nil
}

func unConsensusToUnConsensus(native *native.NativeService, contract common.Address, peerPoolItem *PeerPoolItem) error {
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	//update authorizeInfoPool
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(contract, AUTHORIZE_INFO_POOL, peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
	}
	authorizeInfo := new(AuthorizeInfo)
	for _, v := range stateValues {
		authorizeInfoStore, ok := v.Value.(*cstates.StorageItem)
		if !ok {
			return errors.NewErr("authorizeInfoStore is not available!")
		}
		if err := authorizeInfo.Deserialize(bytes.NewBuffer(authorizeInfoStore.Value)); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize authorizeInfo error!")
		}
		if authorizeInfo.ConsensusPos != 0 {
			return errors.NewErr("consensusPos, freezePos should be 0!")
		}

		newPos := authorizeInfo.NewPos
		freezePos := authorizeInfo.FreezePos
		authorizeInfo.NewPos = 0
		authorizeInfo.FreezePos = newPos + freezePos
		withdrawPos := authorizeInfo.WithdrawPos
		withdrawFreezePos := authorizeInfo.WithdrawFreezePos
		authorizeInfo.WithdrawFreezePos = withdrawPos
		authorizeInfo.WithdrawUnfreezePos = authorizeInfo.WithdrawUnfreezePos + withdrawFreezePos
		authorizeInfo.WithdrawPos = 0

		err = putAuthorizeInfo(native, contract, authorizeInfo)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "putAuthorizeInfo, put authorizeInfo error!")
		}
	}
	return nil
}

func depositTotalStake(native *native.NativeService, contract common.Address, address common.Address, stake uint64) error {
	totalStake, err := getTotalStake(native, contract, address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getTotalStake, get totalStake error!")
	}

	preStake := totalStake.Stake
	preTimeOffset := totalStake.TimeOffset
	timeOffset := native.Time - constants.GENESIS_BLOCK_TIMESTAMP

	amount := utils.CalcUnbindOng(preStake, preTimeOffset, timeOffset)
	err = appCallTransferFromOng(native, utils.GovernanceContractAddress, utils.OntContractAddress, totalStake.Address, amount)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferFromOng, transfer from ong error!")
	}

	totalStake.Stake = preStake + stake
	totalStake.TimeOffset = timeOffset

	err = putTotalStake(native, contract, totalStake)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "putTotalStake, put totalStake error!")
	}
	return nil
}

func withdrawTotalStake(native *native.NativeService, contract common.Address, address common.Address, stake uint64) error {
	totalStake, err := getTotalStake(native, contract, address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getTotalStake, get totalStake error!")
	}
	if totalStake.Stake < stake {
		return errors.NewErr("withdraw, ont deposit is not enough!")
	}

	preStake := totalStake.Stake
	preTimeOffset := totalStake.TimeOffset
	timeOffset := native.Time - constants.GENESIS_BLOCK_TIMESTAMP

	amount := utils.CalcUnbindOng(preStake, preTimeOffset, timeOffset)
	err = appCallTransferFromOng(native, utils.GovernanceContractAddress, utils.OntContractAddress, totalStake.Address, amount)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferFromOng, transfer from ong error!")
	}

	totalStake.Stake = preStake - stake
	totalStake.TimeOffset = timeOffset

	err = putTotalStake(native, contract, totalStake)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "putTotalStake, put totalStake error!")
	}
	return nil
}

func depositPenaltyStake(native *native.NativeService, contract common.Address, peerPubkey string, initPos uint64, authorizePos uint64) error {
	penaltyStake, err := getPenaltyStake(native, contract, peerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getPenaltyStake, get penaltyStake error!")
	}

	preInitPos := penaltyStake.InitPos
	preAuthorizePos := penaltyStake.AuthorizePos
	preStake := preInitPos + preAuthorizePos
	preTimeOffset := penaltyStake.TimeOffset
	preAmount := penaltyStake.Amount
	timeOffset := native.Time - constants.GENESIS_BLOCK_TIMESTAMP

	amount := utils.CalcUnbindOng(preStake, preTimeOffset, timeOffset)

	penaltyStake.Amount = preAmount + amount
	penaltyStake.InitPos = preInitPos + initPos
	penaltyStake.AuthorizePos = preAuthorizePos + authorizePos
	penaltyStake.TimeOffset = timeOffset

	err = putPenaltyStake(native, contract, penaltyStake)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "putPenaltyStake, put penaltyStake error!")
	}
	return nil
}

func withdrawPenaltyStake(native *native.NativeService, contract common.Address, peerPubkey string, address common.Address) error {
	penaltyStake, err := getPenaltyStake(native, contract, peerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getPenaltyStake, get penaltyStake error!")
	}

	preStake := penaltyStake.InitPos + penaltyStake.AuthorizePos
	preTimeOffset := penaltyStake.TimeOffset
	preAmount := penaltyStake.Amount
	timeOffset := native.Time - constants.GENESIS_BLOCK_TIMESTAMP

	amount := utils.CalcUnbindOng(preStake, preTimeOffset, timeOffset)

	//ont transfer
	err = appCallTransferOnt(native, utils.GovernanceContractAddress, address, preStake)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOnt, ont transfer error!")
	}
	//ong approve
	err = appCallTransferFromOng(native, utils.GovernanceContractAddress, utils.OntContractAddress, address, amount+preAmount)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferFromOng, transfer from ong error!")
	}

	peerPubkeyPrefix, err := hex.DecodeString(peerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	native.CloneCache.Delete(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PENALTY_STAKE), peerPubkeyPrefix))
	return nil
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
	peerPoolMapSplit, err := GetPeerPoolMap(native, contract, view-1)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}

	//feeSplit first
	err = executeSplit(native, contract, peerPoolMapSplit)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, executeSplit error!")
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, contract, view)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}

	var peers []*PeerStakeInfo
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

		if peerPoolItem.Status == CandidateStatus || peerPoolItem.Status == ConsensusStatus {
			stake := peerPoolItem.TotalPos + peerPoolItem.InitPos
			peers = append(peers, &PeerStakeInfo{
				Index:      peerPoolItem.Index,
				PeerPubkey: peerPoolItem.PeerPubkey,
				Stake:      stake,
			})
		}
	}
	if len(peers) < int(config.K) {
		return errors.NewErr("commitDpos, num of peers is less than K!")
	}

	// sort peers by stake
	sort.SliceStable(peers, func(i, j int) bool {
		if peers[i].Stake > peers[j].Stake {
			return true
		} else if peers[i].Stake == peers[j].Stake {
			return peers[i].PeerPubkey > peers[j].PeerPubkey
		}
		return false
	})

	// consensus peers
	for i := 0; i < int(config.K); i++ {
		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peers[i].PeerPubkey]
		if !ok {
			return errors.NewErr("commitDpos, peerPubkey is not in peerPoolMap!")
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
			return errors.NewErr("authorizeForPeer, peerPubkey is not in peerPoolMap!")
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
	err = putPeerPoolMap(native, contract, newView, peerPoolMap)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "putPeerPoolMap, put peerPoolMap error!")
	}
	oldView := view - 1
	oldViewBytes, err := GetUint32Bytes(oldView)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "GetUint32Bytes, get oldViewBytes error!")
	}
	native.CloneCache.Delete(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), oldViewBytes))

	//update view
	governanceView = &GovernanceView{
		View:   newView,
		Height: native.Height,
		TxHash: native.Tx.Hash(),
	}
	err = putGovernanceView(native, contract, governanceView)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "putGovernanceView, put governanceView error!")
	}

	return nil
}

func executeSplit(native *native.NativeService, contract common.Address, peerPoolMap *PeerPoolMap) error {
	balance, err := getOngBalance(native, utils.GovernanceContractAddress)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, getOngBalance error!")
	}
	//get globalParam
	globalParam, err := getGlobalParam(native, contract)
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
	config, err := getConfig(native, contract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getConfig, get config error!")
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
			return errors.NewDetailErr(err, errors.ErrNoCode, "splitCurve, calculate splitCurve error!")
		}
		sumS += peersCandidate[i].S
	}
	if sumS == 0 {
		return errors.NewErr("executeSplit, sumS is 0!")
	}

	//fee split of consensus peer
	for i := int(config.K) - 1; i >= 0; i-- {
		nodeAmount := balance * uint64(globalParam.A) / 100 * peersCandidate[i].S / sumS
		address := peersCandidate[i].Address
		err = appCallTransferOng(native, utils.GovernanceContractAddress, address, nodeAmount)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, ong transfer error!")
		}
	}

	//fee split of candidate peer
	// cal s of each candidate node
	sum = 0
	for i := int(config.K); i < len(peersCandidate); i++ {
		sum += peersCandidate[i].Stake
	}
	if sum == 0 {
		return nil
	}
	for i := int(config.K); i < len(peersCandidate); i++ {
		nodeAmount := balance * uint64(globalParam.B) / 100 * peersCandidate[i].Stake / sum
		address := peersCandidate[i].Address
		err = appCallTransferOng(native, utils.GovernanceContractAddress, address, nodeAmount)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, ong transfer error!")
		}
	}

	return nil
}
