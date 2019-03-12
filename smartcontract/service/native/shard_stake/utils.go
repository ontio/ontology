package shard_stake

import (
	"bytes"
	"fmt"
	"github.com/ontio/ontology/common"
	cstates "github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native"
	sstates "github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	sutils "github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	USER_MAX_WITHDRAW_VIEW = 100 // one can withdraw 100 epoch dividends

	KEY_VIEW_INDEX            = "view_index"
	KEY_VIEW_INFO             = "view_info"
	KEY_SHARD_VIEW_USER_STAKE = "shard_view_stake" // user stake info at specific view index of shard

	KEY_SHARD_USER_LAST_STAKE_VIEW    = "shard_last_stake_view"    // user latest stake influence view index
	KEY_SHARD_USER_LAST_WITHDRAW_VIEW = "shard_last_withdraw_view" // user latest withdraw view index, user's dividends at this view has not yet withdrawn
)

func genShardViewKey(contract common.Address, shardIdBytes []byte) []byte {
	return utils.ConcatKey(contract, shardIdBytes, []byte(KEY_VIEW_INDEX))
}

func genShardViewInfoKey(contract common.Address, shardIdBytes []byte, viewBytes []byte) []byte {
	return utils.ConcatKey(contract, shardIdBytes, viewBytes, []byte(KEY_VIEW_INFO))
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

func getShardCurrentView(native *native.NativeService, id types.ShardID) (sstates.View, error) {
	shardIDBytes, err := sutils.GetUint64Bytes(id.ToUint64())
	if err != nil {
		return 0, fmt.Errorf("getShardCurrentView: ser shardId failed, err: %s", err)
	}
	key := genShardViewKey(utils.ShardStakeAddress, shardIDBytes)
	dataBytes, err := native.CacheDB.Get(key)
	if err != nil {
		return 0, fmt.Errorf("getShardCurrentView: read db failed, err: %s", err)
	}
	if len(dataBytes) == 0 {
		return 0, fmt.Errorf("getShardCurrentView: shard %d view not exist", id.ToUint64())
	}
	value, err := cstates.GetValueFromRawStorageItem(dataBytes)
	if err != nil {
		return 0, fmt.Errorf("getShardCurrentView: parse store value failed, err: %s", err)
	}
	view, err := sutils.GetBytesUint64(value)
	if err != nil {
		return 0, fmt.Errorf("getShardCurrentView: deserialize value failed, err: %s", err)
	}
	return sstates.View(view), nil
}

func setShardView(native *native.NativeService, id types.ShardID, view sstates.View) error {
	shardIDBytes, err := sutils.GetUint64Bytes(id.ToUint64())
	if err != nil {
		return fmt.Errorf("setShardView: ser shardId failed, err: %s", err)
	}
	key := genShardViewKey(utils.ShardStakeAddress, shardIDBytes)
	value, err := sutils.GetUint64Bytes(uint64(view))
	if err != nil {
		return fmt.Errorf("setShardView: ser view failed, err: %s", err)
	}
	native.CacheDB.Put(key, cstates.GenRawStorageItem(value))
	return nil
}

func getShardViewInfo(native *native.NativeService, id types.ShardID, view sstates.View) (*ViewInfo, error) {
	shardIDBytes, err := sutils.GetUint64Bytes(id.ToUint64())
	if err != nil {
		return nil, fmt.Errorf("getShardViewInfo: ser shardId failed, err: %s", err)
	}
	viewBytes, err := sutils.GetUint64Bytes(uint64(view))
	if err != nil {
		return nil, fmt.Errorf("getShardViewInfo: ser view failed, err: %s", err)
	}
	key := genShardViewInfoKey(utils.ShardStakeAddress, shardIDBytes, viewBytes)
	dataBytes, err := native.CacheDB.Get(key)
	if err != nil {
		return nil, fmt.Errorf("getShardViewInfo: read db failed, err: %s", err)
	}
	viewInfo := &ViewInfo{}
	if len(dataBytes) == 0 {
		return viewInfo, nil
	}
	storeValue, err := cstates.GetValueFromRawStorageItem(dataBytes)
	if err != nil {
		return nil, fmt.Errorf("getShardViewInfo: parse store vale faield, err: %s", err)
	}
	err = viewInfo.Deserialize(bytes.NewBuffer(storeValue))
	if err != nil {
		return nil, fmt.Errorf("getShardViewInfo: deserialize view info failed, err: %s", err)
	}
	return viewInfo, nil
}

func setShardViewInfo(native *native.NativeService, id types.ShardID, view sstates.View, info *ViewInfo) error {
	shardIDBytes, err := sutils.GetUint64Bytes(id.ToUint64())
	if err != nil {
		return fmt.Errorf("setShardViewInfo: ser shardId failed, err: %s", err)
	}
	viewBytes, err := sutils.GetUint64Bytes(uint64(view))
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

func getShardViewUserStake(native *native.NativeService, id types.ShardID, view sstates.View,
	user common.Address) (*UserStakeInfo, error) {
	shardIDBytes, err := sutils.GetUint64Bytes(id.ToUint64())
	if err != nil {
		return nil, fmt.Errorf("getShardViewUserStake: ser shardId failed, err: %s", err)
	}
	viewBytes, err := sutils.GetUint64Bytes(uint64(view))
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

func setShardViewUserStake(native *native.NativeService, id types.ShardID, view sstates.View, user common.Address,
	info *UserStakeInfo) error {
	shardIDBytes, err := sutils.GetUint64Bytes(id.ToUint64())
	if err != nil {
		return fmt.Errorf("setShardViewUserStake: ser shardId failed, err: %s", err)
	}
	viewBytes, err := sutils.GetUint64Bytes(uint64(view))
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

func getUserLastStakeView(native *native.NativeService, id types.ShardID, user common.Address) (sstates.View, error) {
	shardIDBytes, err := sutils.GetUint64Bytes(id.ToUint64())
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
	view, err := sutils.GetBytesUint64(data)
	if err != nil {
		return 0, fmt.Errorf("getShardViewUserStake: dese value failed, err: %s", err)
	}
	return sstates.View(view), nil
}

func setUserLastStakeView(native *native.NativeService, id types.ShardID, user common.Address, view sstates.View) error {
	shardIDBytes, err := sutils.GetUint64Bytes(id.ToUint64())
	if err != nil {
		return fmt.Errorf("setUserLastStakeView: ser shardId failed, err: %s", err)
	}
	key := genShardUserLastStakeViewKey(utils.ShardStakeAddress, shardIDBytes, user)
	viewBytes, err := sutils.GetUint64Bytes(uint64(view))
	if err != nil {
		return fmt.Errorf("setUserLastStakeView: ser view failed, err: %s", err)
	}
	native.CacheDB.Put(key, cstates.GenRawStorageItem(viewBytes))
	return nil
}

func getUserLastWithdrawView(native *native.NativeService, id types.ShardID, user common.Address) (sstates.View, error) {
	shardIDBytes, err := sutils.GetUint64Bytes(id.ToUint64())
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
	view, err := sutils.GetBytesUint64(data)
	if err != nil {
		return 0, fmt.Errorf("getUserLastWithdrawView: dese value failed, err: %s", err)
	}
	return sstates.View(view), nil
}

func setUserLastWithdrawView(native *native.NativeService, id types.ShardID, user common.Address, view sstates.View) error {
	shardIDBytes, err := sutils.GetUint64Bytes(id.ToUint64())
	if err != nil {
		return fmt.Errorf("setUserLastWithdrawView: ser shardId failed, err: %s", err)
	}
	key := genShardUserLastWithdrawViewKey(utils.ShardStakeAddress, shardIDBytes, user)
	data, err := sutils.GetUint64Bytes(uint64(view))
	if err != nil {
		return fmt.Errorf("setUserLastWithdrawView: ser view failed, err: %s", err)
	}
	native.CacheDB.Put(key, cstates.GenRawStorageItem(data))
	return nil
}
