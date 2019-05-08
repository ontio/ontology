/*
 * Copyright (C) 2019 The ontology Authors
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

package shardmgmt

import (
	"fmt"

	"github.com/ontio/ontology/common"
	cstates "github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/events/message"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	// key prefix
	KEY_VERSION      = "version"
	KEY_GLOBAL_STATE = "globalState"
	KEY_SHARD_STATE  = "shardState"

	KEY_SHARD_PEER_STATE = "peerState"

	KEY_RETRY_COMMIT_DPOS   = "retry_commit"
	KEY_XSHARD_HANDLING_FEE = "xshard_handling_fee"
)

type peerState string

const (
	state_default  peerState = "default"
	state_applied  peerState = "applied"
	state_approved peerState = "approved"
	state_joined   peerState = "joined"
)

func genPeerStateKey(contract common.Address, shardIdBytes []byte, pubKey string) []byte {
	return utils.ConcatKey(contract, shardIdBytes, []byte(KEY_SHARD_PEER_STATE), []byte(pubKey))
}

func genRetryCommitDposKey() []byte {
	return utils.ConcatKey(utils.ShardMgmtContractAddress, []byte(KEY_RETRY_COMMIT_DPOS))
}

func genXShardHandlingFeeKey() []byte {
	return utils.ConcatKey(utils.ShardMgmtContractAddress, []byte(KEY_XSHARD_HANDLING_FEE))
}

func getVersion(native *native.NativeService, contract common.Address) (uint32, error) {
	versionBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(KEY_VERSION)))
	if err != nil {
		return 0, fmt.Errorf("get version: %s", err)
	}

	if versionBytes == nil {
		return 0, nil
	}

	value, err := cstates.GetValueFromRawStorageItem(versionBytes)
	if err != nil {
		return 0, fmt.Errorf("get version, deserialized from raw storage item: %s", err)
	}

	ver, err := utils.GetBytesUint32(value)
	if err != nil {
		return 0, fmt.Errorf("serialization.ReadUint32, deserialize version: %s", err)
	}
	return ver, nil
}

func setVersion(native *native.NativeService, contract common.Address) {
	data := utils.GetUint32Bytes(utils.VERSION_CONTRACT_SHARD_MGMT)
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(KEY_VERSION)), cstates.GenRawStorageItem(data))
}

func checkVersion(native *native.NativeService, contract common.Address) (bool, error) {
	ver, err := getVersion(native, contract)
	if err != nil {
		return false, err
	}
	return ver == utils.VERSION_CONTRACT_SHARD_MGMT, nil
}

func getGlobalState(native *native.NativeService, contract common.Address) (*shardstates.ShardMgmtGlobalState, error) {
	stateBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(KEY_GLOBAL_STATE)))
	if err != nil {
		return nil, fmt.Errorf("get shardmgmt global state: %s", err)
	}

	value, err := cstates.GetValueFromRawStorageItem(stateBytes)
	if err != nil {
		return nil, fmt.Errorf("get shardmgmt global state, deserialize from raw storage: %s", err)
	}

	globalState := &shardstates.ShardMgmtGlobalState{}
	if err := globalState.Deserialization(common.NewZeroCopySource(value)); err != nil {
		return nil, fmt.Errorf("get shardgmgmtm global state: deserialize state: %s", err)
	}

	return globalState, nil
}

func setGlobalState(native *native.NativeService, contract common.Address, state *shardstates.ShardMgmtGlobalState) {
	sink := common.NewZeroCopySink(0)
	state.Serialization(sink)
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(KEY_GLOBAL_STATE)), cstates.GenRawStorageItem(sink.Bytes()))
}

func GetShardState(native *native.NativeService, contract common.Address, shardID common.ShardID) (*shardstates.ShardState, error) {
	shardIDBytes := utils.GetUint64Bytes(shardID.ToUint64())
	shardStateBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(KEY_SHARD_STATE), shardIDBytes))
	if err != nil {
		return nil, fmt.Errorf("getShardState: %s", err)
	}
	if shardStateBytes == nil || len(shardStateBytes) == 0 {
		return nil, fmt.Errorf("getShardState: shard %d not exist", shardID)
	}

	value, err := cstates.GetValueFromRawStorageItem(shardStateBytes)
	if err != nil {
		return nil, fmt.Errorf("getShardState: deserialize from raw storage: %s", err)
	}

	state := &shardstates.ShardState{}
	if err := state.Deserialization(common.NewZeroCopySource(value)); err != nil {
		return nil, fmt.Errorf("getShardState: deserialize ShardState: %s", err)
	}

	return state, nil
}

func setShardState(native *native.NativeService, contract common.Address, state *shardstates.ShardState) {
	shardIDBytes := utils.GetUint64Bytes(state.ShardID.ToUint64())
	sink := common.NewZeroCopySink(0)
	state.Serialization(sink)
	key := utils.ConcatKey(contract, []byte(KEY_SHARD_STATE), shardIDBytes)
	native.CacheDB.Put(key, cstates.GenRawStorageItem(sink.Bytes()))
}

func AddNotification(native *native.NativeService, contract common.Address, info shardstates.ShardMgmtEvent) error {
	sink := common.NewZeroCopySink(0)
	info.Serialization(sink)
	eventState := &message.ShardEventState{
		Version:    utils.VERSION_CONTRACT_SHARD_MGMT,
		EventType:  info.GetType(),
		ToShard:    info.GetTargetShardID(),
		FromHeight: info.GetHeight(),
		Payload:    sink.Bytes(),
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          eventState,
		})
	return nil
}

func setShardPeerState(native *native.NativeService, contract common.Address, shardId common.ShardID, state peerState,
	pubKey string) {
	shardIDBytes := utils.GetUint64Bytes(shardId.ToUint64())
	key := genPeerStateKey(contract, shardIDBytes, pubKey)
	native.CacheDB.Put(key, cstates.GenRawStorageItem([]byte(state)))
}

func getShardPeerState(native *native.NativeService, contract common.Address, shardId common.ShardID,
	pubKey string) (peerState, error) {
	shardIDBytes := utils.GetUint64Bytes(shardId.ToUint64())
	key := genPeerStateKey(contract, shardIDBytes, pubKey)
	data, err := native.CacheDB.Get(key)
	if err != nil {
		return state_default, fmt.Errorf("getShardPeerState: read db failed, err: %s", err)
	}
	if len(data) == 0 {
		return state_default, nil
	}
	value, err := cstates.GetValueFromRawStorageItem(data)
	if err != nil {
		return state_default, fmt.Errorf("getShardPeerState: parse store value failed, err: %s", err)
	}
	return peerState(value), nil
}

func setShardCommitDposInfo(native *native.NativeService, retry *shardstates.ShardCommitDposInfo) {
	sink := common.NewZeroCopySink(0)
	retry.Serialization(sink)
	native.CacheDB.Put(genRetryCommitDposKey(), cstates.GenRawStorageItem(sink.Bytes()))
}

func getShardCommitDposInfo(native *native.NativeService) (*shardstates.ShardCommitDposInfo, error) {
	raw, err := native.CacheDB.Get(genRetryCommitDposKey())
	if err != nil {
		return nil, fmt.Errorf("getShardCommitDposInfo: read db failed, err: %s", err)
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("getShardCommitDposInfo: store is empty")
	}
	storeValue, err := cstates.GetValueFromRawStorageItem(raw)
	if err != nil {
		return nil, fmt.Errorf("getShardCommitDposInfo: parse store value failed, err: %s", err)
	}
	source := common.NewZeroCopySource(storeValue)
	retry := &shardstates.ShardCommitDposInfo{}
	if err := retry.Deserialization(source); err != nil {
		return nil, fmt.Errorf("getShardCommitDposInfo: deserialize failed, err: %s", err)
	}
	return retry, nil
}

func setXShardHandlingFee(native *native.NativeService, feeInfo *shardstates.XShardHandlingFee) {
	key := genXShardHandlingFeeKey()
	sink := common.NewZeroCopySink(0)
	feeInfo.Serialization(sink)
	native.CacheDB.Put(key, cstates.GenRawStorageItem(sink.Bytes()))
}

func getXShardHandlingFee(native *native.NativeService) (*shardstates.XShardHandlingFee, error) {
	raw, err := native.CacheDB.Get(genXShardHandlingFeeKey())
	if err != nil {
		return nil, fmt.Errorf("getXShardHandlingFee: read db failed, err: %s", err)
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("getXShardHandlingFee: store is empty")
	}
	storeValue, err := cstates.GetValueFromRawStorageItem(raw)
	if err != nil {
		return nil, fmt.Errorf("getXShardHandlingFee: parse store value failed, err: %s", err)
	}
	source := common.NewZeroCopySource(storeValue)
	info := &shardstates.XShardHandlingFee{}
	if err := info.Deserialization(source); err != nil {
		return nil, fmt.Errorf("getXShardHandlingFee: deserialize failed, err: %s", err)
	}
	return info, nil
}
