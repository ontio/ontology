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

package shardgas

import (
	"bytes"
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	cstates "github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/shard_stake"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	KEY_VERSION               = "version"
	KEY_SHARD_GAS_AMOUNT      = "shard_balance"
	KEY_WITHDRA_CONFIRM_PEERS = "withdraw_confirm_peers"
	KEY_WITHDRA_CONFIRM_NUM   = "withdraw_confirm_num" // the num of peer confirm user withdraw

	KEY_COMMIT_DPOS_PEERS = "commit_dpos_peers"
	KEY_COMMIT_PEERS_NUM  = "commit_peers_num"
)

func genVersionKey(contract common.Address) []byte {
	return utils.ConcatKey(contract, []byte(KEY_VERSION))
}

func genShardGasBalance(contract common.Address, shardIDByte []byte) []byte {
	return utils.ConcatKey(contract, []byte(KEY_SHARD_GAS_AMOUNT), shardIDByte)
}

func genWithdrawConfirmNumKey(contract, user common.Address, shardIDBytes, idBytes []byte) []byte {
	return utils.ConcatKey(contract, shardIDBytes, user[:], []byte(KEY_WITHDRA_CONFIRM_NUM), idBytes[:])
}

func genPeerConfirmWithdrawKey(contract, user common.Address, peer string, shardIDBytes, idBytes []byte) []byte {
	return utils.ConcatKey(contract, shardIDBytes, user[:], []byte(KEY_WITHDRA_CONFIRM_PEERS), idBytes[:], []byte(peer))
}
func genCommitDposPeersNumKey(contract common.Address, shardIDBytes, viewBytes []byte) []byte {
	return utils.ConcatKey(contract, shardIDBytes, []byte(KEY_COMMIT_PEERS_NUM), viewBytes[:])
}

func genPeerCommitDposKey(contract common.Address, peer string, shardIDBytes, viewBytes []byte) []byte {
	return utils.ConcatKey(contract, shardIDBytes, []byte(KEY_COMMIT_DPOS_PEERS), viewBytes[:], []byte(peer))
}

func getVersion(native *native.NativeService, contract common.Address) (uint32, error) {
	versionBytes, err := native.CacheDB.Get(genVersionKey(contract))
	if err != nil {
		return 0, fmt.Errorf("get version: %s", err)
	}

	if versionBytes == nil {
		return 0, nil
	}

	value, err := cstates.GetValueFromRawStorageItem(versionBytes)
	if err != nil {
		return 0, fmt.Errorf("get versoin, deserialized from raw storage item: %s", err)
	}

	ver, err := serialization.ReadUint32(bytes.NewBuffer(value))
	if err != nil {
		return 0, fmt.Errorf("serialization.ReadUint32, deserialize version: %s", err)
	}
	return ver, nil
}

func setVersion(native *native.NativeService, contract common.Address) error {
	buf := new(bytes.Buffer)
	if err := serialization.WriteUint32(buf, ShardGasMgmtVersion); err != nil {
		return fmt.Errorf("failed to serialize version: %s", err)
	}

	native.CacheDB.Put(genVersionKey(contract), cstates.GenRawStorageItem(buf.Bytes()))
	return nil
}

func checkVersion(native *native.NativeService, contract common.Address) (bool, error) {
	ver, err := getVersion(native, contract)
	if err != nil {
		return false, err
	}
	return ver == ShardGasMgmtVersion, nil
}

func checkShardID(native *native.NativeService, shardID types.ShardID) (bool, error) {
	shardState, err := shardmgmt.GetShardState(native, utils.ShardMgmtContractAddress, shardID)
	if err != nil {
		return false, err
	}

	return shardState.State == shardstates.SHARD_STATE_ACTIVE, nil
}

func setShardGasBalance(native *native.NativeService, contract common.Address, shardId types.ShardID, amount uint64) error {
	shardIDBytes, err := utils.GetUint64Bytes(shardId.ToUint64())
	if err != nil {
		return fmt.Errorf("setShardGasBalance: serialize shardID: %s", err)
	}
	key := genShardGasBalance(contract, shardIDBytes)
	data, err := utils.GetUint64Bytes(amount)
	if err != nil {
		return fmt.Errorf("setShardGasBalance: serialize amount: %s", err)
	}
	native.CacheDB.Put(key, cstates.GenRawStorageItem(data))
	return nil
}

func getShardGasBalance(native *native.NativeService, contract common.Address, shardId types.ShardID) (uint64, error) {
	shardIDBytes, err := utils.GetUint64Bytes(shardId.ToUint64())
	if err != nil {
		return 0, fmt.Errorf("getShardGasBalance: serialize shardID: %s", err)
	}
	key := genShardGasBalance(contract, shardIDBytes)
	storeValue, err := native.CacheDB.Get(key)
	if err != nil {
		return 0, fmt.Errorf("getShardGasBalance: read db failed, err: %s", err)
	}
	if len(storeValue) == 0 {
		return 0, nil
	}
	data, err := cstates.GetValueFromRawStorageItem(storeValue)
	if err != nil {
		return 0, fmt.Errorf("getShardGasBalance: parse store value failed, err: %s", err)
	}
	amount, err := utils.GetBytesUint64(data)
	if err != nil {
		return 0, fmt.Errorf("getShardGasBalance: dese value failed, err: %s", err)
	}
	return amount, nil
}

func getWithdrawConfirmNum(native *native.NativeService, contract, user common.Address, shardId types.ShardID,
	withdrawId uint64) (uint64, error) {
	shardIDBytes, err := utils.GetUint64Bytes(shardId.ToUint64())
	if err != nil {
		return 0, fmt.Errorf("getWithdrawConfirmNum: serialize shardID: %s", err)
	}
	idBytes, err := utils.GetUint64Bytes(withdrawId)
	if err != nil {
		return 0, fmt.Errorf("getWithdrawConfirmNum: serialize withdraw id: %s", err)
	}
	key := genWithdrawConfirmNumKey(contract, user, shardIDBytes, idBytes)
	storeValue, err := native.CacheDB.Get(key)
	if err != nil {
		return 0, fmt.Errorf("getWithdrawConfirmNum: read db failed, err: %s", err)
	}
	if len(storeValue) == 0 {
		return 0, nil
	}
	data, err := cstates.GetValueFromRawStorageItem(storeValue)
	if err != nil {
		return 0, fmt.Errorf("getWithdrawConfirmNum: parse store value failed, err: %s", err)
	}
	amount, err := utils.GetBytesUint64(data)
	if err != nil {
		return 0, fmt.Errorf("getWithdrawConfirmNum: dese value failed, err: %s", err)
	}
	return amount, nil
}

func setWithdrawConfirmNum(native *native.NativeService, contract, user common.Address, shardId types.ShardID,
	withdrawId, num uint64) error {
	shardIDBytes, err := utils.GetUint64Bytes(shardId.ToUint64())
	if err != nil {
		return fmt.Errorf("setWithdrawConfirmNum: serialize shardID: %s", err)
	}
	idBytes, err := utils.GetUint64Bytes(withdrawId)
	if err != nil {
		return fmt.Errorf("setWithdrawConfirmNum: serialize withdraw id: %s", err)
	}
	key := genWithdrawConfirmNumKey(contract, user, shardIDBytes, idBytes)
	value, err := utils.GetUint64Bytes(num)
	if err != nil {
		return fmt.Errorf("setWithdrawConfirmNum: serialize num failed, err: %s", err)
	}
	native.CacheDB.Put(key, cstates.GenRawStorageItem(value))
	return nil
}

func peerConfirmWithdraw(native *native.NativeService, contract, user common.Address, peer string, shardId types.ShardID,
	withdrawId uint64) error {
	shardIDBytes, err := utils.GetUint64Bytes(shardId.ToUint64())
	if err != nil {
		return fmt.Errorf("peerConfirmWithdraw: serialize shardId: %s", err)
	}
	idBytes, err := utils.GetUint64Bytes(withdrawId)
	if err != nil {
		return fmt.Errorf("peerConfirmWithdraw: serialize withdraw id: %s", err)
	}
	key := genPeerConfirmWithdrawKey(contract, user, peer, shardIDBytes, idBytes)
	data, err := utils.GetUint32Bytes(1)
	if err != nil {
		return fmt.Errorf("peerConfirmWithdraw: serialize confirm failed: %s", err)
	}
	native.CacheDB.Put(key, cstates.GenRawStorageItem(data))
	return nil
}

func isPeerConfirmWithdraw(native *native.NativeService, contract, user common.Address, peer string, shardId types.ShardID,
	withdrawId uint64) (bool, error) {
	shardIDBytes, err := utils.GetUint64Bytes(shardId.ToUint64())
	if err != nil {
		return false, fmt.Errorf("isPeerConfirmWithdraw: serialize shardId: %s", err)
	}
	idBytes, err := utils.GetUint64Bytes(withdrawId)
	if err != nil {
		return false, fmt.Errorf("isPeerConfirmWithdraw: serialize withdraw id: %s", err)
	}
	key := genPeerConfirmWithdrawKey(contract, user, peer, shardIDBytes, idBytes)
	storeValue, err := native.CacheDB.Get(key)
	if err != nil {
		return false, fmt.Errorf("isPeerConfirmWithdraw: read db failed, err: %s", err)
	}
	if len(storeValue) == 0 {
		return false, nil
	}
	data, err := cstates.GetValueFromRawStorageItem(storeValue)
	if err != nil {
		return false, fmt.Errorf("isPeerConfirmWithdraw: parse store value failed, err: %s", err)
	}
	num, err := utils.GetBytesUint32(data)
	if err != nil {
		return false, fmt.Errorf("isPeerConfirmWithdraw: dese value failed, err: %s", err)
	}
	return num == 1, nil
}

func getViewCommitNum(native *native.NativeService, contract common.Address, shardId types.ShardID,
	view shard_stake.View) (uint64, error) {
	shardIDBytes, err := utils.GetUint64Bytes(shardId.ToUint64())
	if err != nil {
		return 0, fmt.Errorf("getViewCommitNum: serialize shardID: %s", err)
	}
	viewBytes, err := utils.GetUint64Bytes(uint64(view))
	if err != nil {
		return 0, fmt.Errorf("getViewCommitNum: serialize withdraw id: %s", err)
	}
	key := genCommitDposPeersNumKey(contract, shardIDBytes, viewBytes)
	storeValue, err := native.CacheDB.Get(key)
	if err != nil {
		return 0, fmt.Errorf("getViewCommitNum: read db failed, err: %s", err)
	}
	if len(storeValue) == 0 {
		return 0, nil
	}
	data, err := cstates.GetValueFromRawStorageItem(storeValue)
	if err != nil {
		return 0, fmt.Errorf("getViewCommitNum: parse store value failed, err: %s", err)
	}
	amount, err := utils.GetBytesUint64(data)
	if err != nil {
		return 0, fmt.Errorf("getViewCommitNum: dese value failed, err: %s", err)
	}
	return amount, nil
}

func setViewCommitNum(native *native.NativeService, contract common.Address, shardId types.ShardID,
	view shard_stake.View, num uint64) error {
	shardIDBytes, err := utils.GetUint64Bytes(shardId.ToUint64())
	if err != nil {
		return fmt.Errorf("setViewCommitNum: serialize shardID: %s", err)
	}
	viewBytes, err := utils.GetUint64Bytes(uint64(view))
	if err != nil {
		return fmt.Errorf("setViewCommitNum: serialize withdraw id: %s", err)
	}
	key := genCommitDposPeersNumKey(contract, shardIDBytes, viewBytes)
	value, err := utils.GetUint64Bytes(num)
	if err != nil {
		return fmt.Errorf("setViewCommitNum: serialize num failed, err: %s", err)
	}
	native.CacheDB.Put(key, cstates.GenRawStorageItem(value))
	return nil
}

func peerCommitView(native *native.NativeService, contract common.Address, peer string, shardId types.ShardID,
	view shard_stake.View) error {
	shardIDBytes, err := utils.GetUint64Bytes(shardId.ToUint64())
	if err != nil {
		return fmt.Errorf("peerCommitView: serialize shardId: %s", err)
	}
	viewBytes, err := utils.GetUint64Bytes(uint64(view))
	if err != nil {
		return fmt.Errorf("peerCommitView: serialize withdraw id: %s", err)
	}
	key := genPeerCommitDposKey(contract, peer, shardIDBytes, viewBytes)
	data, err := utils.GetUint32Bytes(1)
	if err != nil {
		return fmt.Errorf("peerCommitView: serialize confirm failed: %s", err)
	}
	native.CacheDB.Put(key, cstates.GenRawStorageItem(data))
	return nil
}

func isPeerCommitView(native *native.NativeService, contract common.Address, peer string, shardId types.ShardID,
	view shard_stake.View) (bool, error) {
	shardIDBytes, err := utils.GetUint64Bytes(shardId.ToUint64())
	if err != nil {
		return false, fmt.Errorf("isPeerCommitView: serialize shardId: %s", err)
	}
	viewBytes, err := utils.GetUint64Bytes(uint64(view))
	if err != nil {
		return false, fmt.Errorf("isPeerCommitView: serialize withdraw id: %s", err)
	}
	key := genPeerCommitDposKey(contract, peer, shardIDBytes, viewBytes)
	storeValue, err := native.CacheDB.Get(key)
	if err != nil {
		return false, fmt.Errorf("isPeerCommitView: read db failed, err: %s", err)
	}
	if len(storeValue) == 0 {
		return false, nil
	}
	data, err := cstates.GetValueFromRawStorageItem(storeValue)
	if err != nil {
		return false, fmt.Errorf("isPeerCommitView: parse store value failed, err: %s", err)
	}
	num, err := utils.GetBytesUint32(data)
	if err != nil {
		return false, fmt.Errorf("isPeerCommitView: dese value failed, err: %s", err)
	}
	return num == 1, nil
}
