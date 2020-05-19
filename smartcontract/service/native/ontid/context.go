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
package ontid

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

//
var _DefaultContexts = [][]byte{[]byte("https://www.w3.org/ns/did/v1"), []byte("https://ontid.ont.io/did/v1")}

func addContext(srvc *native.NativeService) ([]byte, error) {
	params := new(Context)
	if err := params.Deserialization(common.NewZeroCopySource(srvc.Input)); err != nil {
		return utils.BYTE_FALSE, errors.New("addContext error: deserialization params error, " + err.Error())
	}
	encId, err := encodeID(params.OntId)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("addContext error: " + err.Error())
	}
	if !isValid(srvc, encId) {
		return utils.BYTE_FALSE, errors.New("addContext error: have not registered")
	}

	if err := checkWitnessByIndex(srvc, encId, params.Index); err != nil {
		return utils.BYTE_FALSE, errors.New("verify signature failed: " + err.Error())
	}
	key := append(encId, FIELD_CONTEXT)

	if err := putContexts(srvc, key, params); err != nil {
		return utils.BYTE_FALSE, errors.New("addContext error: putContexts failed: " + err.Error())
	}
	updateTimeAndClearProof(srvc, encId)
	return utils.BYTE_TRUE, nil
}

func removeContext(srvc *native.NativeService) ([]byte, error) {
	params := new(Context)
	if err := params.Deserialization(common.NewZeroCopySource(srvc.Input)); err != nil {
		return utils.BYTE_FALSE, errors.New("addContext error: deserialization params error, " + err.Error())
	}
	encId, err := encodeID(params.OntId)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("removeContext error: " + err.Error())
	}
	if !isValid(srvc, encId) {
		return utils.BYTE_FALSE, errors.New("removeContext error: have not registered")
	}

	if err := checkWitnessByIndex(srvc, encId, params.Index); err != nil {
		return utils.BYTE_FALSE, errors.New("verify signature failed: " + err.Error())
	}
	key := append(encId, FIELD_CONTEXT)

	if err := deleteContexts(srvc, key, params); err != nil {
		return utils.BYTE_FALSE, errors.New("removeContext error: deleteContexts failed: " + err.Error())
	}
	updateTimeAndClearProof(srvc, encId)
	return utils.BYTE_TRUE, nil
}

func deleteContexts(srvc *native.NativeService, key []byte, params *Context) error {
	contexts, err := getContexts(srvc, key)
	if err != nil {
		return fmt.Errorf("deleteContexts error: getContexts error, %s", err)
	}
	coincidence := getCoincidence(contexts, params)
	var remove [][]byte
	var remain [][]byte
	for i := 0; i < len(contexts); i++ {
		if _, ok := coincidence[common.ToHexString(contexts[i])]; !ok {
			remain = append(remain, contexts[i])
		} else {
			remove = append(remove, contexts[i])
		}
	}
	triggerContextEvent(srvc, "remove", params.OntId, remove)
	err = storeContexts(remain, srvc, key)
	if err != nil {
		return fmt.Errorf("deleteContexts error: storeContexts error, %s", err)
	}
	return nil
}

func putContexts(srvc *native.NativeService, key []byte, params *Context) error {
	contexts, err := getContexts(srvc, key)
	if err != nil {
		return fmt.Errorf("putContexts error: getContexts failed, %s", err)
	}
	var add [][]byte
	removeDuplicate(params)
	coincidence := getCoincidence(contexts, params)
	for i := 0; i < len(params.Contexts); i++ {
		if (!bytes.Equal(params.Contexts[i], _DefaultContexts[0])) && (!bytes.Equal(params.Contexts[i], _DefaultContexts[1])) {
			if _, ok := coincidence[common.ToHexString(params.Contexts[i])]; !ok {
				contexts = append(contexts, params.Contexts[i])
				add = append(add, params.Contexts[i])
			}
		}
	}
	triggerContextEvent(srvc, "add", params.OntId, add)
	err = storeContexts(contexts, srvc, key)
	if err != nil {
		return fmt.Errorf("putContexts error: storeContexts failed, %s", err)
	}
	return nil
}

func getCoincidence(contexts [][]byte, params *Context) map[string]bool {
	repeat := make(map[string]bool)
	for i := 0; i < len(contexts); i++ {
		for j := 0; j < len(params.Contexts); j++ {
			if bytes.Equal(contexts[i], params.Contexts[j]) {
				repeat[common.ToHexString(params.Contexts[j])] = true
			}
		}
	}
	return repeat
}

func removeDuplicate(params *Context) {
	repeat := make(map[string]bool)
	var res [][]byte
	for i := 0; i < len(params.Contexts); i++ {
		if _, ok := repeat[common.ToHexString(params.Contexts[i])]; !ok {
			res = append(res, params.Contexts[i])
			repeat[common.ToHexString(params.Contexts[i])] = true
		}
	}
	params.Contexts = res
}

func getContexts(srvc *native.NativeService, key []byte) ([][]byte, error) {
	contextsStore, err := utils.GetStorageItem(srvc, key)
	if err != nil {
		return nil, errors.New("getContexts error:" + err.Error())
	}
	if contextsStore == nil {
		return nil, nil
	}
	contexts := new(Contexts)
	if err := contexts.Deserialization(common.NewZeroCopySource(contextsStore.Value)); err != nil {
		return nil, err
	}
	return *contexts, nil
}

func getContextsWithDefault(srvc *native.NativeService, encId []byte) ([]string, error) {
	key := append(encId, FIELD_CONTEXT)
	contexts, err := getContexts(srvc, key)
	if err != nil {
		return nil, fmt.Errorf("getContextsWithDefault error, %s", err)
	}
	contexts = append(_DefaultContexts, contexts...)
	var res []string
	for i := 0; i < len(contexts); i++ {
		res = append(res, string(contexts[i]))
	}
	return res, nil
}

func storeContexts(contexts Contexts, srvc *native.NativeService, key []byte) error {
	sink := common.NewZeroCopySink(nil)
	contexts.Serialization(sink)
	item := states.StorageItem{}
	item.Value = sink.Bytes()
	item.StateVersion = _VERSION_0
	srvc.CacheDB.Put(key, item.ToArray())
	return nil
}
