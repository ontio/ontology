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
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

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

	native.CacheDB.Put(utils.ConcatKey(contract, []byte(KEY_VERSION)), cstates.GenRawStorageItem(buf.Bytes()))
	return nil
}

func checkVersion(native *native.NativeService, contract common.Address) (bool, error) {
	ver, err := getVersion(native, contract)
	if err != nil {
		return false, err
	}
	return ver == ShardGasMgmtVersion, nil
}

func checkShardID(native *native.NativeService, shardID uint64) (bool, error) {
	shardState, err := shardmgmt.GetShardState(native, utils.ShardMgmtContractAddress, shardID)
	if err != nil {
		return false, err
	}

	if shardState == nil {
		return false, fmt.Errorf("invalid shard %d", shardID)
	}

	return shardState.State == shardstates.SHARD_STATE_ACTIVE, nil
}

func getUserBalance(native *native.NativeService, contract common.Address, shardID uint64, user common.Address) (*shardstates.UserGasInfo, error) {
	shardIDByte, err := shardutil.GetUint64Bytes(shardID)
	if err != nil {
		return nil, fmt.Errorf("ser ShardID %s", err)
	}
	keyBytes := utils.ConcatKey(contract, []byte(KEY_BALANCE), shardIDByte, user[:])
	dataBytes, err := native.CacheDB.Get(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("get balance from db: %s", err)
	}
	if len(dataBytes) == 0 {
		return &shardstates.UserGasInfo{
			PendingWithdraw: make([]*shardstates.GasWithdrawInfo, 0),
		}, nil
	}

	gasInfo := &shardstates.UserGasInfo{}
	if err := user.Deserialize(bytes.NewBuffer(dataBytes)); err != nil {
		return nil, fmt.Errorf("deserialize user balance: %s", err)
	}

	return gasInfo, nil
}

func setUserDeposit(native *native.NativeService, contract common.Address, shardID uint64, user common.Address, userGas *shardstates.UserGasInfo) error {
	buf := new(bytes.Buffer)
	if err := userGas.Serialize(buf); err != nil {
		return fmt.Errorf("serialize user balance: %s", err)
	}

	shardIDByte, err := shardutil.GetUint64Bytes(shardID)
	if err != nil {
		return fmt.Errorf("ser ShardID %s", err)
	}
	keyBytes := utils.ConcatKey(contract, []byte(KEY_BALANCE), shardIDByte, user[:])

	native.CacheDB.Put(keyBytes, cstates.GenRawStorageItem(buf.Bytes()))
	return nil
}
