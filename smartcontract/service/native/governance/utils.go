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
	"math/big"

	"github.com/ontio/ontology-crypto/vrf"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/serialization"
	vbftconfig "github.com/ontio/ontology/consensus/vbft/config"
	cstates "github.com/ontio/ontology/core/states"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/auth"
	"github.com/ontio/ontology/smartcontract/service/native/global_params"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func GetPeerPoolMap(native *native.NativeService, contract common.Address, view uint32) (*PeerPoolMap, error) {
	peerPoolMap := &PeerPoolMap{
		PeerPoolMap: make(map[string]*PeerPoolItem),
	}
	viewBytes, err := GetUint32Bytes(view)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "getUint32Bytes, getUint32Bytes error!")
	}
	peerPoolMapBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), viewBytes))
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get all peerPoolMap error!")
	}
	if peerPoolMapBytes == nil {
		return nil, errors.NewErr("getPeerPoolMap, peerPoolMap is nil!")
	}
	peerPoolMapStore, ok := peerPoolMapBytes.(*cstates.StorageItem)
	if !ok {
		return nil, errors.NewErr("getPeerPoolMap, peerPoolMapBytes is not available!")
	}
	if err := peerPoolMap.Deserialize(bytes.NewBuffer(peerPoolMapStore.Value)); err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize peerPoolMap error!")
	}
	return peerPoolMap, nil
}

func putPeerPoolMap(native *native.NativeService, contract common.Address, view uint32, peerPoolMap *PeerPoolMap) error {
	bf := new(bytes.Buffer)
	if err := peerPoolMap.Serialize(bf); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize peerPoolMap error!")
	}
	viewBytes, err := GetUint32Bytes(view)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getUint32Bytes, get viewBytes error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), viewBytes), &cstates.StorageItem{Value: bf.Bytes()})
	return nil
}

func GetGovernanceView(native *native.NativeService, contract common.Address) (*GovernanceView, error) {
	governanceViewBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(GOVERNANCE_VIEW)))
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "getGovernanceView, get governanceViewBytes error!")
	}
	governanceView := new(GovernanceView)
	if governanceViewBytes == nil {
		return nil, errors.NewErr("getGovernanceView, get nil governanceViewBytes!")
	} else {
		governanceViewStore, ok := governanceViewBytes.(*cstates.StorageItem)
		if !ok {
			return nil, errors.NewErr("getGovernanceView, governanceViewBytes is not available!")
		}
		if err := governanceView.Deserialize(bytes.NewBuffer(governanceViewStore.Value)); err != nil {
			return nil, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize governanceView error!")
		}
	}
	return governanceView, nil
}

func putGovernanceView(native *native.NativeService, contract common.Address, governanceView *GovernanceView) error {
	bf := new(bytes.Buffer)
	if err := governanceView.Serialize(bf); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize governanceView error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(GOVERNANCE_VIEW)), &cstates.StorageItem{Value: bf.Bytes()})
	return nil
}

func GetView(native *native.NativeService, contract common.Address) (uint32, error) {
	governanceView, err := GetGovernanceView(native, contract)
	if err != nil {
		return 0, errors.NewDetailErr(err, errors.ErrNoCode, "getView, getGovernanceView error!")
	}
	return governanceView.View, nil
}

func appCallTransferOng(native *native.NativeService, from common.Address, to common.Address, amount uint64) error {
	bf := new(bytes.Buffer)
	var sts []*ont.State
	sts = append(sts, &ont.State{
		From:  from,
		To:    to,
		Value: amount,
	})
	transfers := &ont.Transfers{
		States: sts,
	}
	err := transfers.Serialize(bf)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOng, transfers.Serialize error!")
	}

	if _, err := native.NativeCall(utils.OngContractAddress, "transfer", bf.Bytes()); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOng, appCall error!")
	}
	return nil
}

func appCallTransferOnt(native *native.NativeService, from common.Address, to common.Address, amount uint64) error {
	bf := new(bytes.Buffer)
	var sts []*ont.State
	sts = append(sts, &ont.State{
		From:  from,
		To:    to,
		Value: amount,
	})
	transfers := &ont.Transfers{
		States: sts,
	}
	err := transfers.Serialize(bf)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOnt, transfers.Serialize error!")
	}

	if _, err := native.NativeCall(utils.OntContractAddress, "transfer", bf.Bytes()); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOnt, appCall error!")
	}
	return nil
}

func appCallApproveOng(native *native.NativeService, from common.Address, to common.Address, amount uint64) error {
	bf := new(bytes.Buffer)
	sts := &ont.State{
		From:  from,
		To:    to,
		Value: amount,
	}
	err := sts.Serialize(bf)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallApproveOng, transfers.Serialize error!")
	}

	if _, err := native.NativeCall(utils.OngContractAddress, "approve", bf.Bytes()); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallApproveOng, appCall error!")
	}
	return nil
}

func appCallTransferFromOng(native *native.NativeService, sender common.Address, from common.Address, to common.Address, amount uint64) error {
	bf := new(bytes.Buffer)
	params := &ont.TransferFrom{
		Sender: sender,
		From:   from,
		To:     to,
		Value:  amount,
	}
	err := params.Serialize(bf)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferFromOng, params serialize error!")
	}

	if _, err := native.NativeCall(utils.OngContractAddress, "transferFrom", bf.Bytes()); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferFromOng, appCall error!")
	}
	return nil
}

func getOngBalance(native *native.NativeService, address common.Address) (uint64, error) {
	bf := new(bytes.Buffer)
	err := utils.WriteAddress(bf, address)
	if err != nil {
		return 0, errors.NewDetailErr(err, errors.ErrNoCode, "getOngBalance, utils.WriteAddress error!")
	}

	value, err := native.NativeCall(utils.OngContractAddress, "balanceOf", bf.Bytes())
	if err != nil {
		return 0, errors.NewDetailErr(err, errors.ErrNoCode, "getOngBalance, appCall error!")
	}
	balance := new(big.Int).SetBytes(value.([]byte)).Uint64()
	return balance, nil
}

func splitCurve(native *native.NativeService, contract common.Address, pos uint64, avg uint64, yita uint64) (uint64, error) {
	xi := PRECISE * yita * 2 * pos / (avg * 10)
	index := xi / (PRECISE / 10)
	if index > uint64(len(Xi)-2) {
		index = uint64(len(Xi) - 2)
	}
	splitCurve, err := getSplitCurve(native, contract)
	if err != nil {
		return 0, errors.NewDetailErr(err, errors.ErrNoCode, "getSplitCurve, get splitCurve error!")
	}
	Yi := splitCurve.Yi
	s := ((Yi[index+1]-Yi[index])*xi + Yi[index]*Xi[index+1] - Yi[index+1]*Xi[index]) / (Xi[index+1] - Xi[index])
	return s, nil
}

func GetUint32Bytes(num uint32) ([]byte, error) {
	bf := new(bytes.Buffer)
	if err := serialization.WriteUint32(bf, num); err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize uint32 error!")
	}
	return bf.Bytes(), nil
}

func GetBytesUint32(b []byte) (uint32, error) {
	num, err := serialization.ReadUint32(bytes.NewBuffer(b))
	if err != nil {
		return 0, errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize uint32 error!")
	}
	return num, nil
}

func getGlobalParam(native *native.NativeService, contract common.Address) (*GlobalParam, error) {
	globalParamBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(GLOBAL_PARAM)))
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "getGlobalParam, get globalParamBytes error!")
	}
	globalParam := new(GlobalParam)
	if globalParamBytes == nil {
		return nil, errors.NewErr("getGlobalParam, get nil globalParamBytes!")
	} else {
		globalParamStore, ok := globalParamBytes.(*cstates.StorageItem)
		if !ok {
			return nil, errors.NewErr("getGlobalParam, globalParamBytes is not available!")
		}
		if err := globalParam.Deserialize(bytes.NewBuffer(globalParamStore.Value)); err != nil {
			return nil, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize globalParam error!")
		}
	}
	return globalParam, nil
}

func putGlobalParam(native *native.NativeService, contract common.Address, globalParam *GlobalParam) error {
	bf := new(bytes.Buffer)
	if err := globalParam.Serialize(bf); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize globalParam error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(GLOBAL_PARAM)), &cstates.StorageItem{Value: bf.Bytes()})
	return nil
}

func validatePeerPubKeyFormat(pubkey string) error {
	pk, err := vbftconfig.Pubkey(pubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "failed to parse pubkey")
	}
	if !vrf.ValidatePublicKey(pk) {
		return errors.NewErr("invalid for VRF")
	}
	return nil
}

func CheckVBFTConfig(configuration *config.VBFTConfig) error {
	if configuration.C == 0 {
		return errors.NewErr("initConfig. C can not be 0 in config!")
	}
	if int(configuration.K) != len(configuration.Peers) {
		return errors.NewErr("initConfig. K must equal to length of peer in config!")
	}
	if configuration.L < 16*configuration.K {
		return errors.NewErr("initConfig. L can not be less than 16*K in config!")
	}
	if configuration.K < 2*configuration.C+1 {
		return errors.NewErr("initConfig. K can not be less than 2*C+1 in config!")
	}
	if configuration.N < configuration.K || configuration.K < 7 {
		return errors.NewErr("initConfig. config not match N >= K >= 7!")
	}
	if int(configuration.K) != len(configuration.Peers) {
		return errors.NewErr("initConfig. K must equal to length of peers!")
	}
	if configuration.BlockMsgDelay < 5000 {
		return errors.NewErr("initConfig. BlockMsgDelay must >= 5000!")
	}
	if configuration.HashMsgDelay < 5000 {
		return errors.NewErr("initConfig. HashMsgDelay must >= 5000!")
	}
	if configuration.PeerHandshakeTimeout < 10 {
		return errors.NewErr("initConfig. PeerHandshakeTimeout must >= 10!")
	}
	if configuration.MinInitStake < 10000 {
		return errors.NewErr("initConfig. MinInitStake must >= 10000!")
	}
	if len(configuration.VrfProof) < 128 {
		return errors.NewErr("initConfig. VrfProof must >= 128!")
	}
	if len(configuration.VrfValue) < 128 {
		return errors.NewErr("initConfig. VrfValue must >= 128!")
	}

	indexMap := make(map[uint32]struct{})
	peerPubkeyMap := make(map[string]struct{})
	for _, peer := range configuration.Peers {
		_, ok := indexMap[peer.Index]
		if ok {
			return errors.NewErr("initConfig, peer index is duplicated!")
		}
		indexMap[peer.Index] = struct{}{}

		_, ok = peerPubkeyMap[peer.PeerPubkey]
		if ok {
			return errors.NewErr("initConfig, peerPubkey is duplicated!")
		}
		peerPubkeyMap[peer.PeerPubkey] = struct{}{}

		if peer.Index <= 0 {
			return errors.NewErr("initConfig, peer index in config must > 0!")
		}
		//check peerPubkey
		if err := validatePeerPubKeyFormat(peer.PeerPubkey); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "invalid peer pubkey")
		}
		_, err := common.AddressFromBase58(peer.Address)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "common.AddressFromBase58, address format error!")
		}
	}
	return nil
}

func getConfig(native *native.NativeService, contract common.Address) (*Configuration, error) {
	config := new(Configuration)
	configBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VBFT_CONFIG)))
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Get, get configBytes error!")
	}
	if configBytes == nil {
		return nil, errors.NewErr("commitDpos, configBytes is nil!")
	}
	configStore, ok := configBytes.(*cstates.StorageItem)
	if !ok {
		return nil, errors.NewErr("getConfig, configBytes is not available!")
	}
	if err := config.Deserialize(bytes.NewBuffer(configStore.Value)); err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize config error!")
	}
	return config, nil
}

func putConfig(native *native.NativeService, contract common.Address, config *Configuration) error {
	bf := new(bytes.Buffer)
	if err := config.Serialize(bf); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize config error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VBFT_CONFIG)), &cstates.StorageItem{Value: bf.Bytes()})
	return nil
}

func getCandidateIndex(native *native.NativeService, contract common.Address) (uint32, error) {
	candidateIndexBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(CANDIDITE_INDEX)))
	if err != nil {
		return 0, errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Get, get candidateIndex error!")
	}
	if candidateIndexBytes == nil {
		return 0, errors.NewErr("approveCandidate, candidateIndex is not init!")
	} else {
		candidateIndexStore, ok := candidateIndexBytes.(*cstates.StorageItem)
		if !ok {
			return 0, errors.NewErr("getCandidateIndex, candidateIndexBytes is not available!")
		}
		candidateIndex, err := GetBytesUint32(candidateIndexStore.Value)
		if err != nil {
			return 0, errors.NewDetailErr(err, errors.ErrNoCode, "GetBytesUint32, get candidateIndex error!")
		}
		return candidateIndex, nil
	}
}

func putCandidateIndex(native *native.NativeService, contract common.Address, candidateIndex uint32) error {
	candidateIndexBytes, err := GetUint32Bytes(candidateIndex)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "GetUint32Bytes, get newCandidateIndexBytes error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(CANDIDITE_INDEX)),
		&cstates.StorageItem{Value: candidateIndexBytes})
	return nil
}

func getVoteInfo(native *native.NativeService, contract common.Address, peerPubkey string, address common.Address) (*VoteInfo, error) {
	peerPubkeyPrefix, err := hex.DecodeString(peerPubkey)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	voteInfoBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL),
		peerPubkeyPrefix, address[:]))
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "get voteInfoBytes error!")
	}
	voteInfo := &VoteInfo{
		PeerPubkey: peerPubkey,
		Address:    address,
	}
	if voteInfoBytes != nil {
		voteInfoStore, ok := voteInfoBytes.(*cstates.StorageItem)
		if !ok {
			return nil, errors.NewErr("getVoteInfo, voteInfoBytes is not available!")
		}
		if err := voteInfo.Deserialize(bytes.NewBuffer(voteInfoStore.Value)); err != nil {
			return nil, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfo error!")
		}
	}
	return voteInfo, nil
}

func putVoteInfo(native *native.NativeService, contract common.Address, voteInfo *VoteInfo) error {
	peerPubkeyPrefix, err := hex.DecodeString(voteInfo.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	bf := new(bytes.Buffer)
	if err := voteInfo.Serialize(bf); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize voteInfo error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix,
		voteInfo.Address[:]), &cstates.StorageItem{Value: bf.Bytes()})
	return nil
}

func getPenaltyStake(native *native.NativeService, contract common.Address, peerPubkey string) (*PenaltyStake, error) {
	peerPubkeyPrefix, err := hex.DecodeString(peerPubkey)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	penaltyStakeBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PENALTY_STAKE),
		peerPubkeyPrefix))
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "get voteInfoBytes error!")
	}
	penaltyStake := &PenaltyStake{
		PeerPubkey: peerPubkey,
	}
	if penaltyStakeBytes != nil {
		penaltyStakeStore, ok := penaltyStakeBytes.(*cstates.StorageItem)
		if !ok {
			return nil, errors.NewErr("getPenaltyStake, penaltyStakeBytes is not available!")
		}
		if err := penaltyStake.Deserialize(bytes.NewBuffer(penaltyStakeStore.Value)); err != nil {
			return nil, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfo error!")
		}
	}
	return penaltyStake, nil
}

func putPenaltyStake(native *native.NativeService, contract common.Address, penaltyStake *PenaltyStake) error {
	peerPubkeyPrefix, err := hex.DecodeString(penaltyStake.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	bf := new(bytes.Buffer)
	if err := penaltyStake.Serialize(bf); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize voteInfo error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PENALTY_STAKE), peerPubkeyPrefix),
		&cstates.StorageItem{Value: bf.Bytes()})
	return nil
}

func getTotalStake(native *native.NativeService, contract common.Address, address common.Address) (*TotalStake, error) {
	totalStakeBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(TOTAL_STAKE),
		address[:]))
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "get voteInfoBytes error!")
	}
	totalStake := &TotalStake{
		Address: address,
	}
	if totalStakeBytes != nil {
		totalStakeStore, ok := totalStakeBytes.(*cstates.StorageItem)
		if !ok {
			return nil, errors.NewErr("getTotalStake, totalStakeStore is not available!")
		}
		if err := totalStake.Deserialize(bytes.NewBuffer(totalStakeStore.Value)); err != nil {
			return nil, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfo error!")
		}
	}
	return totalStake, nil
}

func putTotalStake(native *native.NativeService, contract common.Address, totalStake *TotalStake) error {
	bf := new(bytes.Buffer)
	if err := totalStake.Serialize(bf); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize voteInfo error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(TOTAL_STAKE), totalStake.Address[:]),
		&cstates.StorageItem{Value: bf.Bytes()})
	return nil
}

func getAdmin(native *native.NativeService) (common.Address, error) {
	adminAddress := new(common.Address)
	admin, err := global_params.GetStorageRole(native, global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return common.Address{}, errors.NewDetailErr(err, errors.ErrNoCode, "getStorageAdmin, get admin error!")
	}
	copy(adminAddress[:], admin[:])
	return *adminAddress, nil
}

func getSplitCurve(native *native.NativeService, contract common.Address) (*SplitCurve, error) {
	splitCurveBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(SPLIT_CURVE)))
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "getSplitCurve, get splitCurveBytes error!")
	}
	splitCurve := new(SplitCurve)
	if splitCurveBytes == nil {
		return nil, errors.NewErr("getSplitCurve, get nil splitCurveBytes!")
	} else {
		splitCurveStore, ok := splitCurveBytes.(*cstates.StorageItem)
		if !ok {
			return nil, errors.NewErr("getSplitCurve, splitCurveBytes is not available!")
		}
		if err := splitCurve.Deserialize(bytes.NewBuffer(splitCurveStore.Value)); err != nil {
			return nil, errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize splitCurve error!")
		}
	}
	return splitCurve, nil
}

func putSplitCurve(native *native.NativeService, contract common.Address, splitCurve *SplitCurve) error {
	bf := new(bytes.Buffer)
	if err := splitCurve.Serialize(bf); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialize, serialize splitCurve error!")
	}
	native.CloneCache.Add(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(SPLIT_CURVE)), &cstates.StorageItem{Value: bf.Bytes()})
	return nil
}

func appCallInitContractAdmin(native *native.NativeService, adminOntID []byte) error {
	bf := new(bytes.Buffer)
	params := &auth.InitContractAdminParam{
		AdminOntID: adminOntID,
	}
	err := params.Serialize(bf)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallInitContractAdmin, param serialize error!")
	}

	if _, err := native.NativeCall(utils.AuthContractAddress, "initContractAdmin", bf.Bytes()); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallInitContractAdmin, appCall error!")
	}
	return nil
}

func appCallVerifyToken(native *native.NativeService, contract common.Address, caller []byte, fn string, keyNo uint64) error {
	bf := new(bytes.Buffer)
	params := &auth.VerifyTokenParam{
		ContractAddr: contract,
		Caller:       caller,
		Fn:           fn,
		KeyNo:        keyNo,
	}
	err := params.Serialize(bf)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallVerifyToken, param serialize error!")
	}

	ok, err := native.NativeCall(utils.AuthContractAddress, "verifyToken", bf.Bytes())
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallVerifyToken, appCall error!")
	}
	if !bytes.Equal(ok.([]byte), utils.BYTE_TRUE) {
		return errors.NewErr("appCallVerifyToken, verifyToken failed!")
	}
	return nil
}
