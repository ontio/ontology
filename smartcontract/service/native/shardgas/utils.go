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
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func genWithdrawDelayKey(contract common.Address) []byte {
	return utils.ConcatKey(contract, []byte(KEY_WITHDRAW_DELAY))
}

func genVersionKey(contract common.Address) []byte {
	return utils.ConcatKey(contract, []byte(KEY_VERSION))
}

func genShardUserBalanceKey(contract common.Address, shardIDByte []byte, user common.Address) []byte {
	return utils.ConcatKey(contract, []byte(KEY_BALANCE), shardIDByte, user[:])
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

func setWithdrawDelay(native *native.NativeService, contract common.Address, delayHeight uint32) error {
	bf := new(bytes.Buffer)
	if err := serialization.WriteUint32(bf, delayHeight); err != nil {
		return fmt.Errorf("failed to serialize delay, err: %s", err)
	}
	native.CacheDB.Put(genWithdrawDelayKey(contract), cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}

func getWithdrawDelay(native *native.NativeService, contract common.Address) (uint32, error) {
	delayHeightBytes, err := native.CacheDB.Get(genWithdrawDelayKey(contract))
	if err != nil {
		return 0, fmt.Errorf("get withdraw delay height failed, err: %s", err)
	}
	if delayHeightBytes == nil || len(delayHeightBytes) == 0 {
		return 0, fmt.Errorf("withdraw delay height is empty")
	}
	value, err := cstates.GetValueFromRawStorageItem(delayHeightBytes)
	if err != nil {
		return 0, fmt.Errorf("get withdraw delay height, deseialize from raw storage item: %s", err)
	}
	delayHeight, err := serialization.ReadUint32(bytes.NewBuffer(value))
	if err != nil {
		return 0, fmt.Errorf("serialization.ReadUint64, deserialize withdraw delay height: %s", err)
	}
	return delayHeight, nil
}

func checkShardID(native *native.NativeService, shardID types.ShardID) (bool, error) {
	shardState, err := shardmgmt.GetShardState(native, utils.ShardMgmtContractAddress, shardID)
	if err != nil {
		return false, err
	}

	if shardState == nil {
		return false, fmt.Errorf("invalid shard %d", shardID)
	}

	return shardState.State == shardstates.SHARD_STATE_ACTIVE, nil
}

func getUserBalance(native *native.NativeService, contract common.Address, shardID types.ShardID,
	user common.Address) (*shardstates.UserGasInfo, error) {
	shardIDByte, err := shardutil.GetUint64Bytes(shardID.ToUint64())
	if err != nil {
		return nil, fmt.Errorf("ser ShardID %s", err)
	}
	dataBytes, err := native.CacheDB.Get(genShardUserBalanceKey(contract, shardIDByte, user))
	if err != nil {
		return nil, fmt.Errorf("get balance from db: %s", err)
	}
	if len(dataBytes) == 0 {
		return &shardstates.UserGasInfo{
			PendingWithdraw: make([]*shardstates.GasWithdrawInfo, 0),
		}, nil
	}
	value, err := cstates.GetValueFromRawStorageItem(dataBytes)
	if err != nil {
		return nil, fmt.Errorf("get balance, deserialize from rwa storage item: %s", err)
	}

	gasInfo := &shardstates.UserGasInfo{}
	if err := gasInfo.Deserialize(bytes.NewBuffer(value)); err != nil {
		return nil, fmt.Errorf("deserialize user balance: %s", err)
	}

	return gasInfo, nil
}

func setUserDeposit(native *native.NativeService, contract common.Address, shardID types.ShardID, user common.Address,
	userGas *shardstates.UserGasInfo) error {
	buf := new(bytes.Buffer)
	if err := userGas.Serialize(buf); err != nil {
		return fmt.Errorf("serialize user balance: %s", err)
	}

	shardIDByte, err := shardutil.GetUint64Bytes(shardID.ToUint64())
	if err != nil {
		return fmt.Errorf("ser ShardID %s", err)
	}

	native.CacheDB.Put(genShardUserBalanceKey(contract, shardIDByte, user), cstates.GenRawStorageItem(buf.Bytes()))
	return nil
}
