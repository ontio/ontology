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

package shardccmc

import (
	"bytes"
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	cstates "github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/shardccmc/states"
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
	if err := serialization.WriteUint32(buf, ShardCCMCVersion); err != nil {
		return fmt.Errorf("failed to serialize version: %s", err)
	}

	native.CacheDB.Put(utils.ConcatKey(contract, []byte(KEY_VERSION)), cstates.GenRawStorageItem(buf.Bytes()))
	return nil
}

func getCCMCState(native *native.NativeService, contract common.Address) (*ccmc_states.ShardCCMCState, error) {
	stateBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(KEY_CCMC_STATE)))
	if err != nil {
		return nil, fmt.Errorf("get ccmc global state: %s", err)
	}

	value, err := cstates.GetValueFromRawStorageItem(stateBytes)
	if err != nil {
		return nil, fmt.Errorf("get ccmc global state, deserialize from raw storage: %s", err)
	}

	globalState := &ccmc_states.ShardCCMCState{}
	if err := globalState.Deserialize(bytes.NewBuffer(value)); err != nil {
		return nil, fmt.Errorf("get ccmc global state: deserialize state: %s", err)
	}

	return globalState, nil
}

func setCCMCState(native *native.NativeService, contract common.Address, state *ccmc_states.ShardCCMCState) error {
	if state == nil {
		return fmt.Errorf("setCCMCState, nil state")
	}

	buf := new(bytes.Buffer)
	if err := state.Serialize(buf); err != nil {
		return fmt.Errorf("serialize ccmc global state: %s", err)
	}

	native.CacheDB.Put(utils.ConcatKey(contract, []byte(KEY_CCMC_STATE)), cstates.GenRawStorageItem(buf.Bytes()))
	return nil
}

func getCCInfo(native *native.NativeService, contract common.Address, CCID uint64) (*ccmc_states.ShardCCInfo, error) {
	ccidBytes, err := utils.GetUint64Bytes(CCID)
	if err != nil {
		return nil, fmt.Errorf("getCCInfo, serialize shardID: %s", err)
	}

	ccstateBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(KEY_CC_INFO), ccidBytes))
	if err != nil {
		return nil, fmt.Errorf("getCCInfo: %s", err)
	}
	if ccidBytes == nil {
		return nil, nil
	}

	value, err := cstates.GetValueFromRawStorageItem(ccstateBytes)
	if err != nil {
		return nil, fmt.Errorf("getCCInfo, deserialize from raw storage: %s", err)
	}

	state := &ccmc_states.ShardCCInfo{}
	if err := state.Deserialize(bytes.NewBuffer(value)); err != nil {
		return nil, fmt.Errorf("getCCInfo, deserialize CCInfo: %s", err)
	}

	return state, nil
}

func setCCInfo(native *native.NativeService, contract common.Address, state *ccmc_states.ShardCCInfo) error {
	if state == nil {
		return fmt.Errorf("setCCInfo, nil state")
	}

	ccidBytes, err := utils.GetUint64Bytes(state.CCID)
	if err != nil {
		return fmt.Errorf("setCCInfo, serialize shardID: %s", err)
	}

	buf := new(bytes.Buffer)
	if err := state.Serialize(buf); err != nil {
		return fmt.Errorf("serialize ccinfo: %s", err)
	}

	// set CC_STATE
	key := utils.ConcatKey(contract, []byte(KEY_CC_INFO), ccidBytes)
	native.CacheDB.Put(key, cstates.GenRawStorageItem(buf.Bytes()))

	// set CC_CONTRACT
	ccAddrKey := utils.ConcatKey(contract, []byte(KEY_CC_CONTRACT), state.ContractAddr[:])
	native.CacheDB.Put(ccAddrKey, cstates.GenRawStorageItem(ccidBytes))

	log.Infof("set ccstate %d , key %v, state: %s", state.ShardID, key, string(buf.Bytes()))
	return nil
}

func getCCID(native *native.NativeService, contract common.Address, addr common.Address) (uint64, error) {
	value, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(KEY_CC_CONTRACT), addr[:]))
	if err != nil {
		return 0, fmt.Errorf("getCCID: %s", err)
	}
	if value == nil {
		return 0, nil
	}

	ccidBytes, err := cstates.GetValueFromRawStorageItem(value)
	if err != nil {
		return 0, fmt.Errorf("getCCID, deserialize from raw storage: %s", err)
	}

	return utils.GetBytesUint64(ccidBytes)
}
