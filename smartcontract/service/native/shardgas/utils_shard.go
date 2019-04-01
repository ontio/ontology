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

package shardgas

import (
	"fmt"

	"github.com/ontio/ontology/common"
	cstates "github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	KEY_USER_WITHDRAW_ID = "withdraw_id"
	KEY_USER_FROZEN_GAS  = "frozen_gas"
)

func genUserWithdrawId(contract, user common.Address) []byte {
	return utils.ConcatKey(contract, user[:], []byte(KEY_USER_WITHDRAW_ID))
}

func genUserFrozenGasKey(contract, user common.Address, withdrawIdBytes []byte) []byte {
	return utils.ConcatKey(contract, user[:], []byte(KEY_USER_FROZEN_GAS), withdrawIdBytes)
}

func getUserWithdrawId(native *native.NativeService, contract, user common.Address) (uint64, error) {
	key := genUserWithdrawId(contract, user)
	storeValue, err := native.CacheDB.Get(key)
	if err != nil {
		return 0, fmt.Errorf("getUserWithdrawId: read db failed, err: %s", err)
	}
	if len(storeValue) == 0 {
		return 0, nil
	}
	data, err := cstates.GetValueFromRawStorageItem(storeValue)
	if err != nil {
		return 0, fmt.Errorf("getUserWithdrawId: parse store value failed, err: %s", err)
	}
	id, err := utils.GetBytesUint64(data)
	if err != nil {
		return 0, fmt.Errorf("getUserWithdrawId: dese value failed, err: %s", err)
	}
	return id, nil
}

func setUserWithdrawId(native *native.NativeService, contract, user common.Address, id uint64) {
	key := genUserWithdrawId(contract, user)
	value := utils.GetUint64Bytes(id)
	native.CacheDB.Put(key, cstates.GenRawStorageItem(value))
}

func getUserWithdrawGas(native *native.NativeService, contract, user common.Address, withdrawId uint64) (uint64, error) {
	idBytes := utils.GetUint64Bytes(withdrawId)
	key := genUserFrozenGasKey(contract, user, idBytes)
	storeValue, err := native.CacheDB.Get(key)
	if err != nil {
		return 0, fmt.Errorf("getUserWithdrawGas: read db failed, err: %s", err)
	}
	if len(storeValue) == 0 {
		return 0, nil
	}
	data, err := cstates.GetValueFromRawStorageItem(storeValue)
	if err != nil {
		return 0, fmt.Errorf("getUserWithdrawGas: parse store value failed, err: %s", err)
	}
	amount, err := utils.GetBytesUint64(data)
	if err != nil {
		return 0, fmt.Errorf("getUserWithdrawGas: dese value failed, err: %s", err)
	}
	return amount, nil
}

func setUserWithdrawGas(native *native.NativeService, contract, user common.Address, withdrawId, amount uint64) {
	idBytes := utils.GetUint64Bytes(withdrawId)
	key := genUserFrozenGasKey(contract, user, idBytes)
	value := utils.GetUint64Bytes(amount)
	native.CacheDB.Put(key, cstates.GenRawStorageItem(value))
}
