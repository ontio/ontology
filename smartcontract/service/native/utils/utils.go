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
package utils

import (
	"bytes"
	"fmt"

	"github.com/ontio/ontology/common"
	cstates "github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/service/native"
)

func GetPeerPoolMap(native *native.NativeService, contract common.Address, view uint32, key string) (*PeerPoolMap, error) {
	peerPoolMap := &PeerPoolMap{
		PeerPoolMap: make(map[string]*PeerPoolItem),
	}
	viewBytes := GetUint32Bytes(view)
	peerPoolMapBytes, err := native.CacheDB.Get(ConcatKey(contract, []byte(key), viewBytes))
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

func PutPeerPoolMap(native *native.NativeService, contract common.Address, view uint32, peerPoolMap *PeerPoolMap, key string) error {
	bf := new(bytes.Buffer)
	if err := peerPoolMap.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize peerPoolMap error: %v", err)
	}
	viewBytes := GetUint32Bytes(view)
	native.CacheDB.Put(ConcatKey(contract, []byte(key), viewBytes), cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}

func GetChangeView(native *native.NativeService, contract common.Address, key []byte) (*ChangeView, error) {
	changeViewBytes, err := native.CacheDB.Get(ConcatKey(contract, key))
	if err != nil {
		return nil, fmt.Errorf("getChangeView, get changeViewBytes error: %v", err)
	}
	changeView := new(ChangeView)
	if changeViewBytes == nil {
		return nil, fmt.Errorf("getChangeView, get nil changeViewBytes")
	} else {
		value, err := cstates.GetValueFromRawStorageItem(changeViewBytes)
		if err != nil {
			return nil, fmt.Errorf("getChangeView, deserialize from raw storage item err:%v", err)
		}
		if err := changeView.Deserialization(common.NewZeroCopySource(value)); err != nil {
			return nil, fmt.Errorf("getChangeView, deserialize changeView error: %v", err)
		}
	}
	return changeView, nil
}

func GetView(native *native.NativeService, contract common.Address, key []byte) (uint32, error) {
	changeView, err := GetChangeView(native, contract, key)
	if err != nil {
		return 0, fmt.Errorf("getView, getView error: %v", err)
	}
	return changeView.View, nil
}

func PutChangeView(native *native.NativeService, contract common.Address, changeView *ChangeView, key []byte) {
	sink := common.NewZeroCopySink(0)
	changeView.Serialization(sink)
	native.CacheDB.Put(ConcatKey(contract, key), cstates.GenRawStorageItem(sink.Bytes()))
}

func GetConfig(native *native.NativeService, contract common.Address, key string) (*Configuration, error) {
	config := new(Configuration)
	configBytes, err := native.CacheDB.Get(ConcatKey(contract, []byte(key)))
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

func PutConfig(native *native.NativeService, contract common.Address, config *Configuration, key string) error {
	bf := new(bytes.Buffer)
	if err := config.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize config error: %v", err)
	}
	native.CacheDB.Put(ConcatKey(contract, []byte(key)), cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}

func GetPreConfig(native *native.NativeService, contract common.Address, key string) (*PreConfig, error) {
	preConfig := new(PreConfig)
	preConfigBytes, err := native.CacheDB.Get(ConcatKey(contract, []byte(key)))
	if err != nil {
		return nil, fmt.Errorf("native.CacheDB.Get, get preConfigBytes error: %v", err)
	}
	if preConfigBytes != nil {
		preConfigStore, err := cstates.GetValueFromRawStorageItem(preConfigBytes)
		if err != nil {
			return nil, fmt.Errorf("getConfig, preConfigBytes is not available")
		}
		if err := preConfig.Deserialize(bytes.NewBuffer(preConfigStore)); err != nil {
			return nil, fmt.Errorf("deserialize, deserialize preConfig error: %v", err)
		}
	}
	return preConfig, nil
}

func PutPreConfig(native *native.NativeService, contract common.Address, preConfig *PreConfig, key string) error {
	bf := new(bytes.Buffer)
	if err := preConfig.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize preConfig error: %v", err)
	}
	native.CacheDB.Put(ConcatKey(contract, []byte(key)), cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}
