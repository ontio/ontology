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

package shard_stake

import (
	"bytes"
	"fmt"

	"github.com/ontio/ontology/common"
	cstates "github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	USER_MAX_WITHDRAW_VIEW = 100 // one can withdraw 100 epoch dividends

	KEY_VIEW_INDEX = "view_index"
	KEY_VIEW_INFO  = "view_info"

	KEY_SHARD_STAKE_ASSET_ADDR = "shard_stake_asset"

	KEY_SHARD_VIEW_USER_STAKE = "shard_view_stake"     // user stake info at specific view index of shard
	KEY_SHARD_MIN_STAKE       = "shard_peer_min_stake" // peer min stake, ordinary user has not this limit

	KEY_VIEW_DIVIDED = "divided" // shard view has divided fee or not

	KEY_SHARD_USER_LAST_STAKE_VIEW    = "shard_last_stake_view"    // user latest stake influence view index
	KEY_SHARD_USER_LAST_WITHDRAW_VIEW = "shard_last_withdraw_view" // user latest withdraw view index, user's dividends at this view has not yet withdrawn

	KEY_UNBOUND_ONG = "unbound_ong"
)

func genShardDividedKey(contract common.Address, shardIdBytes []byte, viewBytes []byte) []byte {
	return utils.ConcatKey(contract, shardIdBytes, []byte(KEY_VIEW_DIVIDED), viewBytes)
}

func GenShardViewKey(shardIdBytes []byte) []byte {
	return append(shardIdBytes, []byte(KEY_VIEW_INDEX)...)
}

func genShardViewKey(contract common.Address, shardIdBytes []byte) []byte {
	return utils.ConcatKey(contract, shardIdBytes, GenShardViewKey(shardIdBytes))
}

func GenShardViewInfoKey(shardIdBytes []byte, viewBytes []byte) []byte {
	temp := append(shardIdBytes, viewBytes...)
	return append(temp, []byte(KEY_VIEW_INFO)...)
}

func genShardViewInfoKey(contract common.Address, shardIdBytes []byte, viewBytes []byte) []byte {
	return utils.ConcatKey(contract, GenShardViewInfoKey(shardIdBytes, viewBytes))
}

func genShardMinStakeKey(contract common.Address, shardIdBytes []byte) []byte {
	return utils.ConcatKey(contract, shardIdBytes, []byte(KEY_SHARD_MIN_STAKE))
}

func genShardStakeAssetAddrKey(contract common.Address, shardIdBytes []byte) []byte {
	return utils.ConcatKey(contract, shardIdBytes, []byte(KEY_SHARD_STAKE_ASSET_ADDR))
}

func genShardViewUserStakeKey(contract common.Address, shardIdBytes []byte, viewBytes []byte, user common.Address) []byte {
	return utils.ConcatKey(contract, shardIdBytes, viewBytes, []byte(KEY_SHARD_VIEW_USER_STAKE), user[:])
}

func genShardUserLastStakeViewKey(contract common.Address, shardIdBytes []byte, user common.Address) []byte {
	return utils.ConcatKey(contract, shardIdBytes, []byte(KEY_SHARD_USER_LAST_STAKE_VIEW), user[:])
}

func genShardUserLastWithdrawViewKey(contract common.Address, shardIdBytes []byte, user common.Address) []byte {
	return utils.ConcatKey(contract, shardIdBytes, []byte(KEY_SHARD_USER_LAST_WITHDRAW_VIEW), user[:])
}

func genUserUnboundOngKey(contract, user common.Address) []byte {
	return utils.ConcatKey(contract, []byte(KEY_UNBOUND_ONG), user[:])
}

func GetShardCurrentView(native *native.NativeService, id types.ShardID) (View, error) {
	shardIDBytes, err := utils.GetUint64Bytes(id.ToUint64())
	if err != nil {
		return 0, fmt.Errorf("GetShardCurrentView: ser shardId failed, err: %s", err)
	}
	key := genShardViewKey(utils.ShardStakeAddress, shardIDBytes)
	dataBytes, err := native.CacheDB.Get(key)
	if err != nil {
		return 0, fmt.Errorf("GetShardCurrentView: read db failed, err: %s", err)
	}
	if len(dataBytes) == 0 {
		return 0, fmt.Errorf("GetShardCurrentView: shard %d view not exist", id.ToUint64())
	}
	value, err := cstates.GetValueFromRawStorageItem(dataBytes)
	if err != nil {
		return 0, fmt.Errorf("GetShardCurrentView: parse store value failed, err: %s", err)
	}
	view, err := utils.GetBytesUint64(value)
	if err != nil {
		return 0, fmt.Errorf("GetShardCurrentView: deserialize value failed, err: %s", err)
	}
	return View(view), nil
}

func setShardView(native *native.NativeService, id types.ShardID, view View) error {
	shardIDBytes, err := utils.GetUint64Bytes(id.ToUint64())
	if err != nil {
		return fmt.Errorf("setShardView: ser shardId failed, err: %s", err)
	}
	key := genShardViewKey(utils.ShardStakeAddress, shardIDBytes)
	value, err := utils.GetUint64Bytes(uint64(view))
	if err != nil {
		return fmt.Errorf("setShardView: ser view failed, err: %s", err)
	}
	native.CacheDB.Put(key, cstates.GenRawStorageItem(value))
	return nil
}

func GetShardViewInfo(native *native.NativeService, id types.ShardID, view View) (*ViewInfo, error) {
	shardIDBytes, err := utils.GetUint64Bytes(id.ToUint64())
	if err != nil {
		return nil, fmt.Errorf("GetShardViewInfo: ser shardId failed, err: %s", err)
	}
	viewBytes, err := utils.GetUint64Bytes(uint64(view))
	if err != nil {
		return nil, fmt.Errorf("GetShardViewInfo: ser view failed, err: %s", err)
	}
	key := genShardViewInfoKey(utils.ShardStakeAddress, shardIDBytes, viewBytes)
	dataBytes, err := native.CacheDB.Get(key)
	if err != nil {
		return nil, fmt.Errorf("GetShardViewInfo: read db failed, err: %s", err)
	}
	viewInfo := &ViewInfo{}
	if len(dataBytes) == 0 {
		return viewInfo, nil
	}
	storeValue, err := cstates.GetValueFromRawStorageItem(dataBytes)
	if err != nil {
		return nil, fmt.Errorf("GetShardViewInfo: parse store vale faield, err: %s", err)
	}
	err = viewInfo.Deserialize(bytes.NewBuffer(storeValue))
	if err != nil {
		return nil, fmt.Errorf("GetShardViewInfo: deserialize view info failed, err: %s", err)
	}
	return viewInfo, nil
}

func setShardViewInfo(native *native.NativeService, id types.ShardID, view View, info *ViewInfo) error {
	shardIDBytes, err := utils.GetUint64Bytes(id.ToUint64())
	if err != nil {
		return fmt.Errorf("setShardViewInfo: ser shardId failed, err: %s", err)
	}
	viewBytes, err := utils.GetUint64Bytes(uint64(view))
	if err != nil {
		return fmt.Errorf("setShardViewInfo: ser view failed, err: %s", err)
	}
	key := genShardViewInfoKey(utils.ShardStakeAddress, shardIDBytes, viewBytes)
	bf := new(bytes.Buffer)
	err = info.Serialize(bf)
	if err != nil {
		return fmt.Errorf("setShardViewInfo: ser view info failed, err: %s", err)
	}
	native.CacheDB.Put(key, cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}

func getShardViewUserStake(native *native.NativeService, id types.ShardID, view View, user common.Address) (*UserStakeInfo,
	error) {
	shardIDBytes, err := utils.GetUint64Bytes(id.ToUint64())
	if err != nil {
		return nil, fmt.Errorf("getShardViewUserStake: ser shardId failed, err: %s", err)
	}
	viewBytes, err := utils.GetUint64Bytes(uint64(view))
	if err != nil {
		return nil, fmt.Errorf("getShardViewUserStake: ser view failed, err: %s", err)
	}
	key := genShardViewUserStakeKey(utils.ShardStakeAddress, shardIDBytes, viewBytes, user)
	dataBytes, err := native.CacheDB.Get(key)
	if err != nil {
		return nil, fmt.Errorf("getShardViewUserStake: read db failed, err: %s", err)
	}
	info := &UserStakeInfo{}
	if len(dataBytes) == 0 {
		return info, nil
	}
	value, err := cstates.GetValueFromRawStorageItem(dataBytes)
	if err != nil {
		return nil, fmt.Errorf("getShardViewUserStake: parse store info failed, err: %s", err)
	}
	err = info.Deserialize(bytes.NewBuffer(value))
	if err != nil {
		return nil, fmt.Errorf("getShardViewUserStake: dese info failed, err: %s", err)
	}
	return info, nil
}

func setShardViewUserStake(native *native.NativeService, id types.ShardID, view View, user common.Address,
	info *UserStakeInfo) error {
	shardIDBytes, err := utils.GetUint64Bytes(id.ToUint64())
	if err != nil {
		return fmt.Errorf("setShardViewUserStake: ser shardId failed, err: %s", err)
	}
	viewBytes, err := utils.GetUint64Bytes(uint64(view))
	if err != nil {
		return fmt.Errorf("setShardViewUserStake: ser view failed, err: %s", err)
	}
	key := genShardViewUserStakeKey(utils.ShardStakeAddress, shardIDBytes, viewBytes, user)
	bf := new(bytes.Buffer)
	err = info.Serialize(bf)
	if err != nil {
		return fmt.Errorf("setShardViewUserStake: ser info failed, err: %s", err)
	}
	native.CacheDB.Put(key, cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}

func getUserLastStakeView(native *native.NativeService, id types.ShardID, user common.Address) (View, error) {
	shardIDBytes, err := utils.GetUint64Bytes(id.ToUint64())
	if err != nil {
		return 0, fmt.Errorf("getUserLastStakeView: ser shardId failed, err: %s", err)
	}
	key := genShardUserLastStakeViewKey(utils.ShardStakeAddress, shardIDBytes, user)
	storeValue, err := native.CacheDB.Get(key)
	if err != nil {
		return 0, fmt.Errorf("getUserLastStakeView: ser shardId failed, err: %s", err)
	}
	if len(storeValue) == 0 {
		return 0, nil
	}
	data, err := cstates.GetValueFromRawStorageItem(storeValue)
	if err != nil {
		return 0, fmt.Errorf("getUserLastStakeView: parse store value failed, err: %s", err)
	}
	view, err := utils.GetBytesUint64(data)
	if err != nil {
		return 0, fmt.Errorf("getShardViewUserStake: dese value failed, err: %s", err)
	}
	return View(view), nil
}

func setUserLastStakeView(native *native.NativeService, id types.ShardID, user common.Address, view View) error {
	shardIDBytes, err := utils.GetUint64Bytes(id.ToUint64())
	if err != nil {
		return fmt.Errorf("setUserLastStakeView: ser shardId failed, err: %s", err)
	}
	key := genShardUserLastStakeViewKey(utils.ShardStakeAddress, shardIDBytes, user)
	viewBytes, err := utils.GetUint64Bytes(uint64(view))
	if err != nil {
		return fmt.Errorf("setUserLastStakeView: ser view failed, err: %s", err)
	}
	native.CacheDB.Put(key, cstates.GenRawStorageItem(viewBytes))
	return nil
}

func getUserLastWithdrawView(native *native.NativeService, id types.ShardID, user common.Address) (View, error) {
	shardIDBytes, err := utils.GetUint64Bytes(id.ToUint64())
	if err != nil {
		return 0, fmt.Errorf("getUserLastWithdrawView: ser shardId failed, err: %s", err)
	}
	key := genShardUserLastWithdrawViewKey(utils.ShardStakeAddress, shardIDBytes, user)
	storeValue, err := native.CacheDB.Get(key)
	if err != nil {
		return 0, fmt.Errorf("getUserLastWithdrawView: ser shardId failed, err: %s", err)
	}
	if len(storeValue) == 0 {
		return 0, nil
	}
	data, err := cstates.GetValueFromRawStorageItem(storeValue)
	if err != nil {
		return 0, fmt.Errorf("getUserLastWithdrawView: parse store value failed, err: %s", err)
	}
	view, err := utils.GetBytesUint64(data)
	if err != nil {
		return 0, fmt.Errorf("getUserLastWithdrawView: dese value failed, err: %s", err)
	}
	return View(view), nil
}

func setUserLastWithdrawView(native *native.NativeService, id types.ShardID, user common.Address, view View) error {
	shardIDBytes, err := utils.GetUint64Bytes(id.ToUint64())
	if err != nil {
		return fmt.Errorf("setUserLastWithdrawView: ser shardId failed, err: %s", err)
	}
	key := genShardUserLastWithdrawViewKey(utils.ShardStakeAddress, shardIDBytes, user)
	data, err := utils.GetUint64Bytes(uint64(view))
	if err != nil {
		return fmt.Errorf("setUserLastWithdrawView: ser view failed, err: %s", err)
	}
	native.CacheDB.Put(key, cstates.GenRawStorageItem(data))
	return nil
}

func GetNodeMinStakeAmount(native *native.NativeService, id types.ShardID) (uint64, error) {
	shardIDBytes, err := utils.GetUint64Bytes(id.ToUint64())
	if err != nil {
		return 0, fmt.Errorf("GetNodeMinStakeAmount: ser shardId failed, err: %s", err)
	}
	key := genShardMinStakeKey(utils.ShardStakeAddress, shardIDBytes)
	storeValue, err := native.CacheDB.Get(key)
	if err != nil {
		return 0, fmt.Errorf("GetNodeMinStakeAmount: read db failed, err: %s", err)
	}
	if len(storeValue) == 0 {
		return 0, nil
	}
	data, err := cstates.GetValueFromRawStorageItem(storeValue)
	if err != nil {
		return 0, fmt.Errorf("GetNodeMinStakeAmount: parse store value failed, err: %s", err)
	}
	amount, err := utils.GetBytesUint64(data)
	if err != nil {
		return 0, fmt.Errorf("GetNodeMinStakeAmount: dese value failed, err: %s", err)
	}
	return amount, nil
}

func setNodeMinStakeAmount(native *native.NativeService, id types.ShardID, amount uint64) error {
	shardIDBytes, err := utils.GetUint64Bytes(id.ToUint64())
	if err != nil {
		return fmt.Errorf("setNodeMinStakeAmount: ser shardId failed, err: %s", err)
	}
	key := genShardMinStakeKey(utils.ShardStakeAddress, shardIDBytes)
	data, err := utils.GetUint64Bytes(amount)
	if err != nil {
		return fmt.Errorf("setNodeMinStakeAmount: ser view failed, err: %s", err)
	}
	native.CacheDB.Put(key, cstates.GenRawStorageItem(data))
	return nil
}

func setShardStakeAssetAddr(native *native.NativeService, contract common.Address, shardId types.ShardID,
	addr common.Address) error {
	shardIDBytes, err := utils.GetUint64Bytes(shardId.ToUint64())
	if err != nil {
		return fmt.Errorf("setShardStakeAssetAddr: serialize shardID: %s", err)
	}
	bf := new(bytes.Buffer)
	err = addr.Serialize(bf)
	if err != nil {
		return fmt.Errorf("setShardStakeAssetAddr: serialize addr: %s", err)
	}
	key := genShardStakeAssetAddrKey(contract, shardIDBytes)
	native.CacheDB.Put(key, cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}

func getShardStakeAssetAddr(native *native.NativeService, contract common.Address, shardId types.ShardID) (common.Address,
	error) {
	addr := common.Address{}
	shardIDBytes, err := utils.GetUint64Bytes(shardId.ToUint64())
	if err != nil {
		return addr, fmt.Errorf("getShardStakeAssetAddr: serialize shardID: %s", err)
	}
	key := genShardStakeAssetAddrKey(contract, shardIDBytes)
	storeValue, err := native.CacheDB.Get(key)
	if err != nil {
		return addr, fmt.Errorf("getShardStakeAssetAddr: read db failed, err: %s", err)
	}
	if len(storeValue) == 0 {
		return addr, nil
	}
	data, err := cstates.GetValueFromRawStorageItem(storeValue)
	if err != nil {
		return addr, fmt.Errorf("getShardStakeAssetAddr: parse db value failed, err: %s", err)
	}
	err = addr.Deserialize(bytes.NewBuffer(data))
	if err != nil {
		return addr, fmt.Errorf("getShardStakeAssetAddr: deserialize value failed, err: %s", err)
	}
	return addr, nil
}

func getUserUnboundOngInfo(native *native.NativeService, contract, user common.Address) (*UserUnboundOngInfo, error) {
	key := genUserUnboundOngKey(contract, user)
	storeValue, err := native.CacheDB.Get(key)
	if err != nil {
		return nil, fmt.Errorf("getUserUnboundOngInfo: read db failed, err: %s", err)
	}
	info := &UserUnboundOngInfo{}
	if len(storeValue) == 0 {
		return info, nil
	}
	data, err := cstates.GetValueFromRawStorageItem(storeValue)
	if err != nil {
		return nil, fmt.Errorf("getUserUnboundOngInfo: parse db value failed, err: %s", err)
	}
	if err := info.Deserialize(bytes.NewBuffer(data)); err != nil {
		return nil, fmt.Errorf("getUserUnboundOngInfo: deserialize failed, err: %s", err)
	}
	return info, nil
}

func setUserUnboundOngInfo(native *native.NativeService, contract, user common.Address, info *UserUnboundOngInfo) error {
	key := genUserUnboundOngKey(contract, user)
	bf := new(bytes.Buffer)
	if err := info.Serialize(bf); err != nil {
		return fmt.Errorf("setUserUnboundOngInfo: failed, err: %s", err)
	}
	native.CacheDB.Put(key, cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}
