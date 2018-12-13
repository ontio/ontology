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
	"fmt"

	"github.com/ontio/ontology-crypto/vrf"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	vbftconfig "github.com/ontio/ontology/consensus/vbft/config"
	cstates "github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/ongx"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/vm/neovm/types"
)

func GetPeerPoolMap(native *native.NativeService, contract common.Address) (*PeerPoolMap, error) {
	peerPoolMap := &PeerPoolMap{
		PeerPoolMap: make(map[string]*PeerPoolItem),
	}
	peerPoolMapBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(PEER_POOL)))
	if err != nil {
		return nil, fmt.Errorf("getPeerPoolMap, get all peerPoolMap error: %v", err)
	}
	if peerPoolMapBytes == nil {
		return nil, fmt.Errorf("getPeerPoolMap, peerPoolMap is nil")
	}
	item := cstates.StorageItem{}
	err = item.Deserialize(bytes.NewBuffer(peerPoolMapBytes))
	if err != nil {
		return nil, fmt.Errorf("deserialize PeerPoolMap error:%v", err)
	}
	peerPoolMapStore := item.Value
	if err := peerPoolMap.Deserialize(bytes.NewBuffer(peerPoolMapStore)); err != nil {
		return nil, fmt.Errorf("deserialize, deserialize peerPoolMap error: %v", err)
	}
	return peerPoolMap, nil
}

func putPeerPoolMap(native *native.NativeService, contract common.Address, peerPoolMap *PeerPoolMap) error {
	bf := new(bytes.Buffer)
	if err := peerPoolMap.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize peerPoolMap error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(PEER_POOL)), cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}

func GetGovernanceView(native *native.NativeService, contract common.Address) (*GovernanceView, error) {
	governanceViewBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(GOVERNANCE_VIEW)))
	if err != nil {
		return nil, fmt.Errorf("getGovernanceView, get governanceViewBytes error: %v", err)
	}
	governanceView := new(GovernanceView)
	if governanceViewBytes == nil {
		return nil, fmt.Errorf("getGovernanceView, get nil governanceViewBytes")
	} else {
		value, err := cstates.GetValueFromRawStorageItem(governanceViewBytes)
		if err != nil {
			return nil, fmt.Errorf("getGovernanceView, deserialize from raw storage item err:%v", err)
		}
		if err := governanceView.Deserialize(bytes.NewBuffer(value)); err != nil {
			return nil, fmt.Errorf("deserialize, deserialize governanceView error: %v", err)
		}
	}
	return governanceView, nil
}

func putGovernanceView(native *native.NativeService, contract common.Address, governanceView *GovernanceView) error {
	bf := bytes.NewBuffer(nil)
	if err := governanceView.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize governanceView error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(GOVERNANCE_VIEW)), cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}

func GetView(native *native.NativeService, contract common.Address) (uint32, error) {
	governanceView, err := GetGovernanceView(native, contract)
	if err != nil {
		return 0, fmt.Errorf("getView, getGovernanceView error: %v", err)
	}
	return governanceView.View, nil
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

func validatePeerPubKeyFormat(pubkey string) error {
	pk, err := vbftconfig.Pubkey(pubkey)
	if err != nil {
		return fmt.Errorf("failed to parse pubkey")
	}
	if !vrf.ValidatePublicKey(pk) {
		return fmt.Errorf("invalid for VRF")
	}
	return nil
}

func CheckVBFTConfig(configuration *config.VBFTConfig) error {
	if configuration.C == 0 {
		return fmt.Errorf("initConfig. C can not be 0 in config")
	}
	if int(configuration.K) != len(configuration.Peers) {
		return fmt.Errorf("initConfig. K must equal to length of peer in config")
	}
	if configuration.L < 16*configuration.K || configuration.L%configuration.K != 0 {
		return fmt.Errorf("initConfig. L can not be less than 16*K and K must be times of L in config")
	}
	if configuration.K < 2*configuration.C+1 {
		return fmt.Errorf("initConfig. K can not be less than 2*C+1 in config")
	}
	if configuration.N < configuration.K || configuration.K < 7 {
		return fmt.Errorf("initConfig. config not match N >= K >= 7")
	}
	if configuration.BlockMsgDelay < 5000 {
		return fmt.Errorf("initConfig. BlockMsgDelay must >= 5000")
	}
	if configuration.HashMsgDelay < 5000 {
		return fmt.Errorf("initConfig. HashMsgDelay must >= 5000")
	}
	if configuration.PeerHandshakeTimeout < 10 {
		return fmt.Errorf("initConfig. PeerHandshakeTimeout must >= 10")
	}
	if configuration.MinInitStake < 10000 {
		return fmt.Errorf("initConfig. MinInitStake must >= 10000")
	}
	if len(configuration.VrfProof) < 128 {
		return fmt.Errorf("initConfig. VrfProof must >= 128")
	}
	if len(configuration.VrfValue) < 128 {
		return fmt.Errorf("initConfig. VrfValue must >= 128")
	}

	indexMap := make(map[uint32]struct{})
	peerPubkeyMap := make(map[string]struct{})
	for _, peer := range configuration.Peers {
		_, ok := indexMap[peer.Index]
		if ok {
			return fmt.Errorf("initConfig, peer index is duplicated")
		}
		indexMap[peer.Index] = struct{}{}

		_, ok = peerPubkeyMap[peer.PeerPubkey]
		if ok {
			return fmt.Errorf("initConfig, peerPubkey is duplicated")
		}
		peerPubkeyMap[peer.PeerPubkey] = struct{}{}

		if peer.Index <= 0 {
			return fmt.Errorf("initConfig, peer index in config must > 0")
		}
		//check peerPubkey
		if err := validatePeerPubKeyFormat(peer.PeerPubkey); err != nil {
			return fmt.Errorf("invalid peer pubkey")
		}
		_, err := common.AddressFromBase58(peer.Address)
		if err != nil {
			return fmt.Errorf("common.AddressFromBase58, address format error: %v", err)
		}
	}
	return nil
}

func getConfig(native *native.NativeService, contract common.Address) (*Configuration, error) {
	config := new(Configuration)
	configBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(VBFT_CONFIG)))
	if err != nil {
		return nil, fmt.Errorf("native.CacheDB.Get, get configBytes error: %v", err)
	}
	if configBytes == nil {
		return nil, fmt.Errorf("getConfig, configBytes is nil")
	}
	value, err := cstates.GetValueFromRawStorageItem(configBytes)
	if err != nil {
		return nil, fmt.Errorf("getConfig, deserialize from raw storage item err:%v", err)
	}
	if err := config.Deserialize(bytes.NewBuffer(value)); err != nil {
		return nil, fmt.Errorf("deserialize, deserialize config error: %v", err)
	}
	return config, nil
}

func putConfig(native *native.NativeService, contract common.Address, config *Configuration) error {
	bf := new(bytes.Buffer)
	if err := config.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize config error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(VBFT_CONFIG)), cstates.GenRawStorageItem(bf.Bytes()))
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

func getOngBalance(native *native.NativeService, address common.Address) (uint64, error) {
	bf := new(bytes.Buffer)
	err := utils.WriteAddress(bf, address)
	if err != nil {
		return 0, fmt.Errorf("getOngBalance, utils.WriteAddress error: %v", err)
	}
	sink := common.ZeroCopySink{}
	utils.EncodeAddress(&sink, address)

	value, err := native.NativeCall(utils.OngContractAddress, "balanceOf", sink.Bytes())
	if err != nil {
		return 0, fmt.Errorf("getOngBalance, appCall error: %v", err)
	}
	balance := types.BigIntFromBytes(value.([]byte)).Uint64()
	return balance, nil
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

func getSyncAddress(native *native.NativeService) (common.Address, error) {
	key := append(utils.OngContractAddress[:], ongx.SYNC_ADDRESS...)
	syncAddressBytes, err := native.CacheDB.Get(key)
	if err != nil {
		return common.Address{}, fmt.Errorf("getSyncAddress, get address from cache error:%s", err)
	}
	if syncAddressBytes == nil {
		return common.Address{}, fmt.Errorf("getSyncAddress, get nil syncAddressBytes")
	}
	syncAddressStore, err := cstates.GetValueFromRawStorageItem(syncAddressBytes)
	if err != nil {
		return common.Address{}, fmt.Errorf("getSyncAddress, deserialize from raw storage item err:%v", err)
	}

	syncAddress := new(ongx.SyncAddress)
	if err := syncAddress.Deserialize(common.NewZeroCopySource(syncAddressStore)); err != nil {
		return common.Address{}, fmt.Errorf("getSyncAddress, deserialize syncAddress error: %v", err)
	}
	return syncAddress.SyncAddress, nil
}

func appCallTransferOng(native *native.NativeService, from common.Address, to common.Address, amount uint64) error {
	err := appCallTransfer(native, utils.OngContractAddress, from, to, amount)
	if err != nil {
		return fmt.Errorf("appCallTransferOng, appCallTransfer error: %v", err)
	}
	return nil
}

func appCallTransfer(native *native.NativeService, contract common.Address, from common.Address, to common.Address, amount uint64) error {
	var sts []ongx.State
	sts = append(sts, ongx.State{
		From:  from,
		To:    to,
		Value: amount,
	})
	transfers := ongx.Transfers{
		States: sts,
	}
	sink := common.NewZeroCopySink(nil)
	transfers.Serialization(sink)

	if _, err := native.NativeCall(contract, "transfer", sink.Bytes()); err != nil {
		return fmt.Errorf("appCallTransfer, appCall error: %v", err)
	}
	return nil
}
