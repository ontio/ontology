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
	"bytes"
	"fmt"

	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/events/message"
	"github.com/ontio/ontology/smartcontract/service/native/shard_stake"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	cstates "github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
	gov "github.com/ontio/ontology/smartcontract/service/native/governance"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	// key prefix
	KEY_VERSION      = "version"
	KEY_GLOBAL_STATE = "globalState"
	KEY_SHARD_STATE  = "shardState"

	KEY_SHARD_PEER_STATE = "peerState"
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

	ver, err := serialization.ReadUint32(bytes.NewBuffer(value))
	if err != nil {
		return 0, fmt.Errorf("serialization.ReadUint32, deserialize version: %s", err)
	}
	return ver, nil
}

func setVersion(native *native.NativeService, contract common.Address) error {
	buf := new(bytes.Buffer)
	if err := serialization.WriteUint32(buf, VERSION_CONTRACT_SHARD_MGMT); err != nil {
		return fmt.Errorf("failed to serialize version: %s", err)
	}

	native.CacheDB.Put(utils.ConcatKey(contract, []byte(KEY_VERSION)), cstates.GenRawStorageItem(buf.Bytes()))
	return nil
}

func checkVersion(native *native.NativeService, contract common.Address) (bool, error) {
	ver, err := getVersion(native, contract)
	if err != nil {
		return false, err
	}
	return ver == VERSION_CONTRACT_SHARD_MGMT, nil
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
	if err := globalState.Deserialize(bytes.NewBuffer(value)); err != nil {
		return nil, fmt.Errorf("get shardgmgmtm global state: deserialize state: %s", err)
	}

	return globalState, nil
}

func setGlobalState(native *native.NativeService, contract common.Address, state *shardstates.ShardMgmtGlobalState) error {
	if state == nil {
		return fmt.Errorf("setGlobalState, nil state")
	}

	buf := new(bytes.Buffer)
	if err := state.Serialize(buf); err != nil {
		return fmt.Errorf("serialize shardmgmt global state: %s", err)
	}

	native.CacheDB.Put(utils.ConcatKey(contract, []byte(KEY_GLOBAL_STATE)), cstates.GenRawStorageItem(buf.Bytes()))
	return nil
}

func GetShardState(native *native.NativeService, contract common.Address, shardID types.ShardID) (*shardstates.ShardState, error) {
	shardIDBytes, err := shardutil.GetUint64Bytes(shardID.ToUint64())
	if err != nil {
		return nil, fmt.Errorf("getShardState: serialize shardID: %s", err)
	}

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
	if err := state.Deserialize(bytes.NewBuffer(value)); err != nil {
		return nil, fmt.Errorf("getShardState: deserialize ShardState: %s", err)
	}

	return state, nil
}

func setShardState(native *native.NativeService, contract common.Address, state *shardstates.ShardState) error {
	shardIDBytes, err := shardutil.GetUint64Bytes(state.ShardID.ToUint64())
	if err != nil {
		return fmt.Errorf("setShardState: serialize shardID: %s", err)
	}

	buf := new(bytes.Buffer)
	if err := state.Serialize(buf); err != nil {
		return fmt.Errorf("setShardState: serialize shardstate: %s", err)
	}

	key := utils.ConcatKey(contract, []byte(KEY_SHARD_STATE), shardIDBytes)
	native.CacheDB.Put(key, cstates.GenRawStorageItem(buf.Bytes()))
	return nil
}

func AddNotification(native *native.NativeService, contract common.Address, info shardstates.ShardMgmtEvent) error {
	infoBuf := new(bytes.Buffer)
	if err := shardutil.SerJson(infoBuf, info); err != nil {
		return fmt.Errorf("addNotification, ser info: %s", err)
	}
	eventState := &message.ShardEventState{
		Version:    VERSION_CONTRACT_SHARD_MGMT,
		EventType:  info.GetType(),
		ToShard:    info.GetTargetShardID(),
		FromHeight: info.GetHeight(),
		Payload:    infoBuf.Bytes(),
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          eventState,
		})
	return nil
}

func setShardPeerState(native *native.NativeService, contract common.Address, shardId types.ShardID, state peerState,
	pubKey string) error {
	shardIDBytes, err := shardutil.GetUint64Bytes(shardId.ToUint64())
	if err != nil {
		return fmt.Errorf("setShardPeerState: serialize shardID: %s", err)
	}
	key := genPeerStateKey(contract, shardIDBytes, pubKey)
	native.CacheDB.Put(key, cstates.GenRawStorageItem([]byte(state)))
	return nil
}

func getShardPeerState(native *native.NativeService, contract common.Address, shardId types.ShardID,
	pubKey string) (peerState, error) {
	shardIDBytes, err := shardutil.GetUint64Bytes(shardId.ToUint64())
	if err != nil {
		return state_default, fmt.Errorf("getShardPeerState: serialize shardID: %s", err)
	}
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

func getRootCurrentViewPeerItem(native *native.NativeService, pubKey string) (*utils.PeerPoolItem, error) {
	peerPoolMap, err := getRootCurrentViewPeerMap(native)
	if err != nil {
		return nil, fmt.Errorf("getRootCurrentViewPeerItem: failed, err: %s", err)
	}
	item, ok := peerPoolMap.PeerPoolMap[pubKey]
	if !ok {
		return nil, fmt.Errorf("getRootCurrentViewPeerItem: peer not exist")
	}
	return item, nil
}

func getRootCurrentViewPeerMap(native *native.NativeService) (*utils.PeerPoolMap, error) {
	//get current view
	view, err := utils.GetView(native, utils.GovernanceContractAddress, gov.GOVERNANCE_VIEW)
	if err != nil {
		return nil, fmt.Errorf("getRootCurrentViewPeerMap: get view error: %s", err)
	}
	//get peerPoolMap
	peerPoolMap, err := utils.GetPeerPoolMap(native, utils.GovernanceContractAddress, view, gov.PEER_POOL)
	if err != nil {
		return nil, fmt.Errorf("getRootCurrentViewPeerMap: get peerPoolMap error: %s", err)
	}
	return peerPoolMap, nil
}

func setNodeMinStakeAmount(native *native.NativeService, id types.ShardID, amount uint64) error {
	setMinStakeParam := &shard_stake.SetMinStakeParam{
		ShardId: id,
		Amount:  amount,
	}
	bf := new(bytes.Buffer)
	if err := setMinStakeParam.Serialize(bf); err != nil {
		return fmt.Errorf("setNodeMinStakeAmount: failed, err: %s", err)
	}
	if _, err := native.NativeCall(utils.ShardStakeAddress, shard_stake.SET_MIN_STAKE, bf.Bytes()); err != nil {
		return fmt.Errorf("setNodeMinStakeAmount: failed, err: %s", err)
	}
	return nil
}

func initStakeContractShard(native *native.NativeService, id types.ShardID) error {
	param := &shard_stake.InitShardParam{
		ShardId: id,
	}
	bf := new(bytes.Buffer)
	if err := param.Serialize(bf); err != nil {
		return fmt.Errorf("initStakeContractShard: failed, err: %s", err)
	}
	if _, err := native.NativeCall(utils.ShardStakeAddress, shard_stake.INIT_SHARD, bf.Bytes()); err != nil {
		return fmt.Errorf("initStakeContractShard: failed, err: %s", err)
	}
	return nil
}

func peerInitStake(native *native.NativeService, param *JoinShardParam, stakeAssetAddr common.Address) error {
	callParam := &shard_stake.PeerInitStakeParam{
		ShardId:        param.ShardID,
		StakeAssetAddr: stakeAssetAddr,
		PeerOwner:      param.PeerOwner,
		PeerPubKey:     param.PeerPubKey,
		StakeAmount:    param.StakeAmount,
	}
	bf := new(bytes.Buffer)
	if err := callParam.Serialize(bf); err != nil {
		return fmt.Errorf("peerInitStake: failed, err: %s", err)
	}
	if _, err := native.NativeCall(utils.ShardStakeAddress, shard_stake.PEER_STAKE, bf.Bytes()); err != nil {
		return fmt.Errorf("peerInitStake: failed, err: %s", err)
	}
	return nil
}

func peerExit(native *native.NativeService, shardId types.ShardID, peer string) error {
	param := &shard_stake.PeerExitParam{
		ShardId: shardId,
		Peer:    peer,
	}
	bf := new(bytes.Buffer)
	if err := param.Serialize(bf); err != nil {
		return fmt.Errorf("peerExit: failed, err: %s", err)
	}
	if _, err := native.NativeCall(utils.ShardStakeAddress, shard_stake.PEER_EXIT, bf.Bytes()); err != nil {
		return fmt.Errorf("peerExit: failed, err: %s", err)
	}
	return nil
}

func deletePeer(native *native.NativeService, shardId types.ShardID, peers []string) error {
	param := &shard_stake.DeletePeerParam{
		ShardId: shardId,
		Peers:   peers,
	}
	bf := new(bytes.Buffer)
	if err := param.Serialize(bf); err != nil {
		return fmt.Errorf("deletePeer: failed, err: %s", err)
	}
	if _, err := native.NativeCall(utils.ShardStakeAddress, shard_stake.DELETE_PEER, bf.Bytes()); err != nil {
		return fmt.Errorf("deletePeer: failed, err: %s", err)
	}
	return nil
}

func commitDpos(native *native.NativeService, shardId types.ShardID, amount []uint64, peers []string) error {
	param := &shard_stake.CommitDposParam{
		ShardId:    shardId,
		PeerPubKey: peers,
		Amount:     amount,
	}
	bf := new(bytes.Buffer)
	if err := param.Serialize(bf); err != nil {
		return fmt.Errorf("commitDpos: failed, err: %s", err)
	}
	if _, err := native.NativeCall(utils.ShardStakeAddress, shard_stake.COMMIT_DPOS, bf.Bytes()); err != nil {
		return fmt.Errorf("commitDpos: failed, err: %s", err)
	}
	return nil
}
