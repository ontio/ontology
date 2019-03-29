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

	"github.com/ontio/ontology/smartcontract/service/native/ont"

	"github.com/ontio/ontology/common"
	cstates "github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/auth"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func appCallTransferOnt(native *native.NativeService, from common.Address, to common.Address, amount uint64) error {
	err := ont.AppCallTransfer(native, utils.OntContractAddress, from, to, amount)
	if err != nil {
		return fmt.Errorf("appCallTransferOnt, appCallTransfer error: %v", err)
	}
	return nil
}

func appCallTransferOng(native *native.NativeService, from common.Address, to common.Address, amount uint64) error {
	err := ont.AppCallTransfer(native, utils.OngContractAddress, from, to, amount)
	if err != nil {
		return fmt.Errorf("appCallTransferOng, appCallTransfer error: %v", err)
	}
	return nil
}

func appCallTransferFromOnt(native *native.NativeService, sender common.Address, from common.Address, to common.Address,
	amount uint64) error {
	err := ont.AppCallTransferFrom(native, utils.OntContractAddress, sender, from, to, amount)
	if err != nil {
		return fmt.Errorf("appCallTransferFromOnt, appCallTransferFrom error: %v", err)
	}
	return nil
}

func appCallTransferFromOng(native *native.NativeService, sender common.Address, from common.Address, to common.Address,
	amount uint64) error {
	err := ont.AppCallTransferFrom(native, utils.OngContractAddress, sender, from, to, amount)
	if err != nil {
		return fmt.Errorf("appCallTransferFromOng, appCallTransferFrom error: %v", err)
	}
	return nil
}

func splitCurve(native *native.NativeService, contract common.Address, pos uint64, avg uint64, yita uint64) (uint64, error) {
	if avg == 0 {
		return 0, fmt.Errorf("splitCurve, avg stake is 0")
	}
	xi := PRECISE * yita * 2 * pos / (avg * 10)
	index := xi / (PRECISE / 10)
	if index > uint64(len(Xi)-2) {
		index = uint64(len(Xi) - 2)
	}
	splitCurve, err := getSplitCurve(native, contract)
	if err != nil {
		return 0, fmt.Errorf("getSplitCurve, get splitCurve error: %v", err)
	}
	Yi := splitCurve.Yi
	s := ((uint64(Yi[index+1])-uint64(Yi[index]))*xi + uint64(Yi[index])*uint64(Xi[index+1]) - uint64(Yi[index+1])*uint64(Xi[index])) / (uint64(Xi[index+1]) - uint64(Xi[index]))
	return s, nil
}

func getGlobalParam(native *native.NativeService, contract common.Address) (*GlobalParam, error) {
	globalParamBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(GLOBAL_PARAM)))
	if err != nil {
		return nil, fmt.Errorf("getGlobalParam, get globalParamBytes error: %v", err)
	}
	globalParam := new(GlobalParam)
	if globalParamBytes == nil {
		return nil, fmt.Errorf("getGlobalParam, get nil globalParamBytes")
	} else {
		value, err := cstates.GetValueFromRawStorageItem(globalParamBytes)
		if err != nil {
			return nil, fmt.Errorf("getGlobalParam, deserialize from raw storage item err:%v", err)
		}
		if err := globalParam.Deserialize(bytes.NewBuffer(value)); err != nil {
			return nil, fmt.Errorf("deserialize, deserialize globalParam error: %v", err)
		}
	}
	return globalParam, nil
}

func putGlobalParam(native *native.NativeService, contract common.Address, globalParam *GlobalParam) error {
	bf := new(bytes.Buffer)
	if err := globalParam.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize globalParam error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(GLOBAL_PARAM)), cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}

//get extends global params, for avoiding change default global param struct in store, add GlobalParam2 as extends struct
func getGlobalParam2(native *native.NativeService, contract common.Address) (*GlobalParam2, error) {
	//get globalParam
	globalParam, err := getGlobalParam(native, contract)
	if err != nil {
		return nil, fmt.Errorf("getGlobalParam, getGlobalParam error: %v", err)
	}

	globalParam2Bytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(GLOBAL_PARAM2)))
	if err != nil {
		return nil, fmt.Errorf("getGlobalParam2, get globalParam2Bytes error: %v", err)
	}
	globalParam2 := &GlobalParam2{
		MinAuthorizePos:      500,
		CandidateFeeSplitNum: globalParam.CandidateNum,
	}
	if globalParam2Bytes != nil {
		value, err := cstates.GetValueFromRawStorageItem(globalParam2Bytes)
		if err != nil {
			return nil, fmt.Errorf("getGlobalParam2, globalParam2Bytes is not available")
		}
		if err := globalParam2.Deserialize(bytes.NewBuffer(value)); err != nil {
			return nil, fmt.Errorf("deserialize, deserialize getGlobalParam2 error: %v", err)
		}
	}
	return globalParam2, nil
}

func putGlobalParam2(native *native.NativeService, contract common.Address, globalParam2 *GlobalParam2) error {
	bf := new(bytes.Buffer)
	if err := globalParam2.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize globalParam2 error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(GLOBAL_PARAM2)), cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}

func getCandidateIndex(native *native.NativeService, contract common.Address) (uint32, error) {
	candidateIndexBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(CANDIDITE_INDEX)))
	if err != nil {
		return 0, fmt.Errorf("native.CacheDB.Get, get candidateIndex error: %v", err)
	}
	if candidateIndexBytes == nil {
		return 0, fmt.Errorf("getCandidateIndex, candidateIndex is not init")
	} else {
		candidateIndexStore, err := cstates.GetValueFromRawStorageItem(candidateIndexBytes)
		if err != nil {
			return 0, fmt.Errorf("getCandidateIndex, deserialize from raw storage item err:%v", err)
		}
		candidateIndex, err := utils.GetBytesUint32(candidateIndexStore)
		if err != nil {
			return 0, fmt.Errorf("GetBytesUint32, get candidateIndex error: %v", err)
		}
		return candidateIndex, nil
	}
}

func putCandidateIndex(native *native.NativeService, contract common.Address, candidateIndex uint32) error {
	candidateIndexBytes, err := utils.GetUint32Bytes(candidateIndex)
	if err != nil {
		return fmt.Errorf("GetUint32Bytes, get candidateIndexBytes error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(CANDIDITE_INDEX)), cstates.GenRawStorageItem(candidateIndexBytes))
	return nil
}

func getSplitFee(native *native.NativeService, contract common.Address) (uint64, error) {
	splitFeeBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(SPLIT_FEE)))
	if err != nil {
		return 0, fmt.Errorf("native.CacheDB.Get, get splitFeeBytes error: %v", err)
	}
	var splitFee uint64 = 0
	if splitFeeBytes != nil {
		splitFeeStore, err := cstates.GetValueFromRawStorageItem(splitFeeBytes)
		if err != nil {
			return 0, fmt.Errorf("getSplitFee, splitFeeBytes is not available")
		}
		splitFee, err = utils.GetBytesUint64(splitFeeStore)
		if err != nil {
			return 0, fmt.Errorf("GetBytesUint64, get splitFee error: %v", err)
		}
	}
	return splitFee, nil
}

func putSplitFee(native *native.NativeService, contract common.Address, splitFee uint64) error {
	splitFeeBytes, err := utils.GetUint64Bytes(splitFee)
	if err != nil {
		return fmt.Errorf("GetUint64Bytes, get splitFeeBytes error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(SPLIT_FEE)), cstates.GenRawStorageItem(splitFeeBytes))
	return nil
}

func getSplitFeeAddress(native *native.NativeService, contract common.Address, address common.Address) (*SplitFeeAddress, error) {
	splitFeeAddressBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(SPLIT_FEE_ADDRESS), address[:]))
	if err != nil {
		return nil, fmt.Errorf("native.CacheDB.Get, get splitFeeAddressBytes error: %v", err)
	}
	splitFeeAddress := &SplitFeeAddress{
		Address: address,
	}
	if splitFeeAddressBytes != nil {
		splitFeeAddressStore, err := cstates.GetValueFromRawStorageItem(splitFeeAddressBytes)
		if err != nil {
			return nil, fmt.Errorf("getSplitFeeAddress, splitFeeAddressBytes is not available")
		}
		err = splitFeeAddress.Deserialize(bytes.NewBuffer(splitFeeAddressStore))
		if err != nil {
			return nil, fmt.Errorf("deserialize, deserialize splitFeeAddress error: %v", err)
		}
	}
	return splitFeeAddress, nil
}

func putSplitFeeAddress(native *native.NativeService, contract common.Address, address common.Address, splitFeeAddress *SplitFeeAddress) error {
	bf := new(bytes.Buffer)
	if err := splitFeeAddress.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize splitFeeAddress error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(SPLIT_FEE_ADDRESS), address[:]),
		cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}

func getAuthorizeInfo(native *native.NativeService, contract common.Address, peerPubkey string, address common.Address) (*AuthorizeInfo, error) {
	peerPubkeyPrefix, err := hex.DecodeString(peerPubkey)
	if err != nil {
		return nil, fmt.Errorf("hex.DecodeString, peerPubkey format error: %v", err)
	}
	authorizeInfoBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, AUTHORIZE_INFO_POOL,
		peerPubkeyPrefix, address[:]))
	if err != nil {
		return nil, fmt.Errorf("get authorizeInfoBytes error: %v", err)
	}
	authorizeInfo := &AuthorizeInfo{
		PeerPubkey: peerPubkey,
		Address:    address,
	}
	if authorizeInfoBytes != nil {
		authorizeInfoStore, err := cstates.GetValueFromRawStorageItem(authorizeInfoBytes)
		if err != nil {
			return nil, fmt.Errorf("getAuthorizeInfo, deserialize from raw storage item err:%v", err)
		}
		if err := authorizeInfo.Deserialize(bytes.NewBuffer(authorizeInfoStore)); err != nil {
			return nil, fmt.Errorf("deserialize, deserialize authorizeInfo error: %v", err)
		}
	}
	return authorizeInfo, nil
}

func putAuthorizeInfo(native *native.NativeService, contract common.Address, authorizeInfo *AuthorizeInfo) error {
	peerPubkeyPrefix, err := hex.DecodeString(authorizeInfo.PeerPubkey)
	if err != nil {
		return fmt.Errorf("hex.DecodeString, peerPubkey format error: %v", err)
	}
	bf := new(bytes.Buffer)
	if err := authorizeInfo.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize authorizeInfo error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, AUTHORIZE_INFO_POOL, peerPubkeyPrefix,
		authorizeInfo.Address[:]), cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}

func getPenaltyStake(native *native.NativeService, contract common.Address, peerPubkey string) (*PenaltyStake, error) {
	peerPubkeyPrefix, err := hex.DecodeString(peerPubkey)
	if err != nil {
		return nil, fmt.Errorf("hex.DecodeString, peerPubkey format error: %v", err)
	}
	penaltyStakeBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(PENALTY_STAKE),
		peerPubkeyPrefix))
	if err != nil {
		return nil, fmt.Errorf("get authorizeInfoBytes error: %v", err)
	}
	penaltyStake := &PenaltyStake{
		PeerPubkey: peerPubkey,
	}
	if penaltyStakeBytes != nil {
		penaltyStakeStore, err := cstates.GetValueFromRawStorageItem(penaltyStakeBytes)
		if err != nil {
			return nil, fmt.Errorf("getPenaltyStake, deserialize from raw storage item err:%v", err)
		}
		if err := penaltyStake.Deserialize(bytes.NewBuffer(penaltyStakeStore)); err != nil {
			return nil, fmt.Errorf("deserialize, deserialize authorizeInfo error: %v", err)
		}
	}
	return penaltyStake, nil
}

func putPenaltyStake(native *native.NativeService, contract common.Address, penaltyStake *PenaltyStake) error {
	peerPubkeyPrefix, err := hex.DecodeString(penaltyStake.PeerPubkey)
	if err != nil {
		return fmt.Errorf("hex.DecodeString, peerPubkey format error: %v", err)
	}
	bf := new(bytes.Buffer)
	if err := penaltyStake.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize authorizeInfo error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(PENALTY_STAKE), peerPubkeyPrefix),
		cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}

func getTotalStake(native *native.NativeService, contract common.Address, address common.Address) (*TotalStake, error) {
	totalStakeBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(TOTAL_STAKE),
		address[:]))
	if err != nil {
		return nil, fmt.Errorf("get authorizeInfoBytes error: %v", err)
	}
	totalStake := &TotalStake{
		Address: address,
	}
	if totalStakeBytes != nil {
		totalStakeStore, err := cstates.GetValueFromRawStorageItem(totalStakeBytes)
		if err != nil {
			return nil, fmt.Errorf("getTotalStake, deserialize from raw storage item err:%v", err)
		}
		if err := totalStake.Deserialize(bytes.NewBuffer(totalStakeStore)); err != nil {
			return nil, fmt.Errorf("deserialize, deserialize authorizeInfo error: %v", err)
		}
	}
	return totalStake, nil
}

func putTotalStake(native *native.NativeService, contract common.Address, totalStake *TotalStake) error {
	bf := new(bytes.Buffer)
	if err := totalStake.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize authorizeInfo error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(TOTAL_STAKE), totalStake.Address[:]),
		cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}

func getSplitCurve(native *native.NativeService, contract common.Address) (*SplitCurve, error) {
	splitCurveBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(SPLIT_CURVE)))
	if err != nil {
		return nil, fmt.Errorf("getSplitCurve, get splitCurveBytes error: %v", err)
	}
	splitCurve := new(SplitCurve)
	if splitCurveBytes == nil {
		return nil, fmt.Errorf("getSplitCurve, get nil splitCurveBytes")
	} else {
		splitCurveStore, err := cstates.GetValueFromRawStorageItem(splitCurveBytes)
		if err != nil {
			return nil, fmt.Errorf("getSplitCurve, deserialize from raw storage item err:%v", err)
		}
		if err := splitCurve.Deserialize(bytes.NewBuffer(splitCurveStore)); err != nil {
			return nil, fmt.Errorf("deserialize, deserialize splitCurve error: %v", err)
		}
	}
	return splitCurve, nil
}

func putSplitCurve(native *native.NativeService, contract common.Address, splitCurve *SplitCurve) error {
	bf := new(bytes.Buffer)
	if err := splitCurve.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize splitCurve error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(SPLIT_CURVE)), cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}

func appCallInitContractAdmin(native *native.NativeService, adminOntID []byte) error {
	bf := new(bytes.Buffer)
	params := &auth.InitContractAdminParam{
		AdminOntID: adminOntID,
	}
	err := params.Serialize(bf)
	if err != nil {
		return fmt.Errorf("appCallInitContractAdmin, param serialize error: %v", err)
	}

	if _, err := native.NativeCall(utils.AuthContractAddress, "initContractAdmin", bf.Bytes()); err != nil {
		return fmt.Errorf("appCallInitContractAdmin, appCall error: %v", err)
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
		return fmt.Errorf("appCallVerifyToken, param serialize error: %v", err)
	}

	ok, err := native.NativeCall(utils.AuthContractAddress, "verifyToken", bf.Bytes())
	if err != nil {
		return fmt.Errorf("appCallVerifyToken, appCall error: %v", err)
	}
	if !bytes.Equal(ok.([]byte), utils.BYTE_TRUE) {
		return fmt.Errorf("appCallVerifyToken, verifyToken failed")
	}
	return nil
}

func getPeerAttributes(native *native.NativeService, contract common.Address, peerPubkey string) (*PeerAttributes, error) {
	peerPubkeyPrefix, err := hex.DecodeString(peerPubkey)
	if err != nil {
		return nil, fmt.Errorf("hex.DecodeString, peerPubkey format error: %v", err)
	}
	peerAttributesBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(PEER_ATTRIBUTES), peerPubkeyPrefix))
	if err != nil {
		return nil, fmt.Errorf("getPeerAttributes, native.CacheDB.Get error: %v", err)
	}
	peerAttributes := &PeerAttributes{
		PeerPubkey:   peerPubkey,
		MaxAuthorize: 0,
		T2PeerCost:   100,
		T1PeerCost:   100,
		TPeerCost:    100,
	}
	if peerAttributesBytes != nil {
		peerAttributesStore, err := cstates.GetValueFromRawStorageItem(peerAttributesBytes)
		if err != nil {
			return nil, fmt.Errorf("getPeerAttributes, peerAttributesStore is not available")
		}
		if err := peerAttributes.Deserialize(bytes.NewBuffer(peerAttributesStore)); err != nil {
			return nil, fmt.Errorf("deserialize, deserialize peerAttributes error: %v", err)
		}
	}
	return peerAttributes, nil
}

func getPeerCost(native *native.NativeService, contract common.Address, peerPubkey string) (uint64, error) {
	//get peerAttributes
	peerAttributes, err := getPeerAttributes(native, contract, peerPubkey)
	if err != nil {
		return 0, fmt.Errorf("getPeerAttributes error: %v", err)
	}

	return peerAttributes.TPeerCost, nil
}

func putPeerAttributes(native *native.NativeService, contract common.Address, peerAttributes *PeerAttributes) error {
	peerPubkeyPrefix, err := hex.DecodeString(peerAttributes.PeerPubkey)
	if err != nil {
		return fmt.Errorf("hex.DecodeString, peerPubkey format error: %v", err)
	}
	bf := new(bytes.Buffer)
	if err := peerAttributes.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize peerAttributes error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(PEER_ATTRIBUTES), peerPubkeyPrefix), cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}

func getPromisePos(native *native.NativeService, contract common.Address, peerPubkey string) (*PromisePos, error) {
	peerPubkeyPrefix, err := hex.DecodeString(peerPubkey)
	if err != nil {
		return nil, fmt.Errorf("hex.DecodeString, peerPubkey format error: %v", err)
	}
	promisePosBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(PROMISE_POS), peerPubkeyPrefix))
	if err != nil {
		return nil, fmt.Errorf("get promisePosBytes error: %v", err)
	}
	promisePosStore, err := cstates.GetValueFromRawStorageItem(promisePosBytes)
	if err != nil {
		return nil, fmt.Errorf("get value from promisePosBytes err:%v", err)
	}
	promisePos := new(PromisePos)
	if err := promisePos.Deserialize(bytes.NewBuffer(promisePosStore)); err != nil {
		return nil, fmt.Errorf("deserialize, deserialize promisePos error: %v", err)
	}
	return promisePos, nil
}

func putPromisePos(native *native.NativeService, contract common.Address, promisePos *PromisePos) error {
	peerPubkeyPrefix, err := hex.DecodeString(promisePos.PeerPubkey)
	if err != nil {
		return fmt.Errorf("hex.DecodeString, peerPubkey format error: %v", err)
	}
	bf := new(bytes.Buffer)
	if err := promisePos.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize promisePos error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(PROMISE_POS), peerPubkeyPrefix),
		cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}
