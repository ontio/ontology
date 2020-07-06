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
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func GetPublicKeyByID(srvc *native.NativeService) ([]byte, error) {
	args := common.NewZeroCopySource(srvc.Input)
	// arg0: ID
	arg0, err := utils.DecodeVarBytes(args)
	if err != nil {
		return nil, errors.New("get public key failed: argument 0 error")
	}
	// arg1: key ID
	arg1, err := utils.DecodeUint32(args)
	if err != nil {
		return nil, errors.New("get public key failed: argument 1 error")
	}

	key, err := encodeID(arg0)
	if err != nil {
		return nil, fmt.Errorf("get public key failed: %s", err)
	}

	pk, err := getPk(srvc, key, arg1)
	if err != nil {
		return nil, fmt.Errorf("get public key failed: %s", err)
	} else if pk == nil {
		return nil, errors.New("get public key failed: not found")
	} else if pk.revoked {
		return nil, errors.New("get public key failed: revoked")
	}

	return pk.key, nil
}

// deprecated
func GetDDO(srvc *native.NativeService) ([]byte, error) {
	log.Debug("GetDDO")
	source := common.NewZeroCopySource(srvc.Input)
	did, err := utils.DecodeVarBytes(source)
	if err != nil {
		return nil, fmt.Errorf("get id error, %s", err)
	}

	key, err := encodeID(did)
	if err != nil {
		return nil, err
	}
	// check state
	switch checkIDState(srvc, key) {
	case flag_not_exist:
		return nil, nil
	case flag_revoke:
		return nil, fmt.Errorf("id is already revoked")
	}
	// keys
	var0, err := GetPublicKeys(srvc)
	if err != nil {
		return nil, fmt.Errorf("get DDO error: %s", err)
	}

	sink := common.NewZeroCopySink(nil)
	sink.WriteVarBytes(var0)

	// attributes
	var1, err := GetAttributes(srvc)
	if err != nil {
		return nil, fmt.Errorf("get attribute error, %s", err)
	}
	sink.WriteVarBytes(var1)

	// old recovery
	// ignore error
	oldRec, _ := getOldRecovery(srvc, key)
	sink.WriteVarBytes(oldRec)

	// controller
	con, err := getController(srvc, key)
	var2 := []byte{}
	if err == nil {
		switch t := con.(type) {
		case []byte:
			var2 = t
		case *Group:
			var2 = t.ToJson()
		}
	}
	sink.WriteVarBytes(var2)

	// new recovery
	var3 := []byte{}
	rec, err := getRecovery(srvc, key)
	if rec != nil && err == nil {
		var3 = rec.ToJson()
	}
	sink.WriteVarBytes(var3)

	res := sink.Bytes()
	log.Debug("DDO:", hex.EncodeToString(res))
	return res, nil
}

// Deprecated
func GetPublicKeys(srvc *native.NativeService) ([]byte, error) {
	log.Debug("GetPublicKeys")
	args := common.NewZeroCopySource(srvc.Input)
	did, err := utils.DecodeVarBytes(args)
	if err != nil {
		return nil, fmt.Errorf("get public keys error: invalid argument, %s", err)
	}
	if len(did) == 0 {
		return nil, errors.New("get public keys error: invalid ID")
	}
	key, err := encodeID(did)
	if err != nil {
		return nil, fmt.Errorf("get public keys error: %s", err)
	}
	key = append(key, FIELD_PK)
	list, err := getAllPk(srvc, key)
	if err != nil {
		return nil, fmt.Errorf("get public keys error: %s", err)
	} else if list == nil {
		return nil, nil
	}

	sink := common.NewZeroCopySink(nil)
	for i, v := range list {
		if v.revoked {
			continue
		}
		sink.WriteUint32(uint32(i + 1))
		sink.WriteVarBytes(v.key)
	}

	return sink.Bytes(), nil
}

func GetPublicKeysJson(srvc *native.NativeService) ([]byte, error) {
	log.Debug("GetPublicKeysJson")
	args := common.NewZeroCopySource(srvc.Input)
	did, err := utils.DecodeVarBytes(args)
	if err != nil {
		return nil, fmt.Errorf("get public keys error: invalid argument, %s", err)
	}
	if len(did) == 0 {
		return nil, errors.New("get public keys error: invalid ID")
	}
	encId, err := encodeID(did)
	if err != nil {
		return nil, fmt.Errorf("get public keys error: %s", err)
	}
	if !isValid(srvc, encId) {
		return nil, nil
	}
	r, err := getAllPkJson(srvc, encId)
	if err != nil {
		return nil, fmt.Errorf("get public keys error: %s", err)
	} else if r == nil {
		return nil, nil
	}

	result, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal error: %s", err)
	}
	return result, nil
}

func GetAttributes(srvc *native.NativeService) ([]byte, error) {
	log.Debug("GetAttributes")
	source := common.NewZeroCopySource(srvc.Input)
	did, err := utils.DecodeVarBytes(source)
	if err != nil {
		return nil, fmt.Errorf("get all attributes error: invalid argument, %s", err)
	}
	if len(did) == 0 {
		return nil, errors.New("get all attributes error: invalid ID")
	}
	key, err := encodeID(did)
	if err != nil {
		return nil, fmt.Errorf("get all attributes error: %s", err)
	}
	if !isValid(srvc, key) {
		return nil, nil
	}
	res, err := getAllAttr(srvc, key)
	if err != nil {
		return nil, fmt.Errorf("get all attributes error: %s", err)
	}

	return res, nil
}

func GetAttributeByKey(srvc *native.NativeService) ([]byte, error) {
	log.Debug("GetAttributeByKey")
	source := common.NewZeroCopySource(srvc.Input)
	did, err := utils.DecodeVarBytes(source)
	if err != nil {
		return nil, fmt.Errorf("get attributes by key error: invalid argument1, %s", err)
	}
	if len(did) == 0 {
		return nil, errors.New("get attributes by key error: invalid ID")
	}
	key, err := encodeID(did)
	if err != nil {
		return nil, fmt.Errorf("get attributes by key error: %s", err)
	}
	if !isValid(srvc, key) {
		return nil, nil
	}
	item, err := utils.DecodeVarBytes(source)
	if err != nil {
		return nil, fmt.Errorf("get attributes by key error: invalid argument2, %s", err)
	}
	res, err := getAttrByKey(srvc, key, item)
	if err != nil {
		return nil, fmt.Errorf("get attributes by key error: %s", err)
	}

	return res, nil
}

func GetAttributesJson(srvc *native.NativeService) ([]byte, error) {
	log.Debug("GetAttributesJson")
	source := common.NewZeroCopySource(srvc.Input)
	did, err := utils.DecodeVarBytes(source)
	if err != nil {
		return nil, fmt.Errorf("get public keys error: invalid argument, %s", err)
	}
	if len(did) == 0 {
		return nil, errors.New("get attributes error: invalid ID")
	}
	key, err := encodeID(did)
	if err != nil {
		return nil, fmt.Errorf("get public keys error: %s", err)
	}
	if !isValid(srvc, key) {
		return nil, nil
	}
	res, err := getAllAttrJson(srvc, key)
	if err != nil {
		return nil, fmt.Errorf("get attributes error: %s", err)
	}

	result, err := json.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal error: %s", err)
	}
	return result, nil
}

func GetKeyState(srvc *native.NativeService) ([]byte, error) {
	log.Debug("GetKeyState")
	source := common.NewZeroCopySource(srvc.Input)
	// arg0: ID
	arg0, _, irregular, eof := source.NextVarBytes()
	if irregular || eof {
		return nil, fmt.Errorf("get key state failed: argument 0 error")
	}
	// arg1: public key ID
	arg1, err := utils.DecodeVarUint(source)
	if err != nil {
		return nil, fmt.Errorf("get key state failed: argument 1 error, %s", err)
	}

	key, err := encodeID(arg0)
	if err != nil {
		return nil, fmt.Errorf("get key state failed: %s", err)
	}
	if !isValid(srvc, key) {
		return nil, nil
	}

	owner, err := getPk(srvc, key, uint32(arg1))
	if err != nil {
		return nil, fmt.Errorf("get key state failed: %s", err)
	} else if owner == nil {
		log.Debug("key state: not exist")
		return []byte("not exist"), nil
	}

	log.Debug("key state: ", owner.revoked)
	if owner.revoked {
		return []byte("revoked"), nil
	} else {
		return []byte("in use"), nil
	}
}

func GetServiceJson(srvc *native.NativeService) ([]byte, error) {
	log.Debug("GetServiceJson")
	params := new(SearchServiceParam)
	source := common.NewZeroCopySource(srvc.Input)
	err := params.Deserialization(source)
	if err != nil {
		return nil, errors.New("GetService error: deserialization params error, " + err.Error())
	}
	encId, err := encodeID(params.OntId)
	if err != nil {
		return nil, fmt.Errorf("encodeID failed: %s", err)
	}
	if !isValid(srvc, encId) {
		return nil, nil
	}

	services, err := getServices(srvc, encId)
	if err != nil {
		return nil, errors.New("GetService error: getServices error, " + err.Error())
	}
	for i := 0; i < len(services); i++ {
		if bytes.Equal(services[i].ServiceId, params.ServiceId) {
			service := new(serviceJson)
			service.Id = fmt.Sprintf("%s#%s", string(params.OntId), string(params.ServiceId))
			service.Type = string(services[i].Type)
			service.ServiceEndpoint = string(services[i].ServiceEndpint)
			data, err := json.Marshal(service)
			if err != nil {
				return nil, errors.New("GetService error: json.Marshal error, " + err.Error())
			}
			return data, nil
		}
	}
	return nil, nil
}

func GetControllerJson(srvc *native.NativeService) ([]byte, error) {
	log.Debug("GetControllerJson")
	source := common.NewZeroCopySource(srvc.Input)
	// arg0: ID
	arg0, _, irregular, eof := source.NextVarBytes()
	if irregular || eof {
		return nil, fmt.Errorf("get key state failed: argument 0 error")
	}
	encId, err := encodeID(arg0)
	if err != nil {
		return nil, fmt.Errorf("encodeID failed: %s", err)
	}
	if !isValid(srvc, encId) {
		return nil, nil
	}
	r, err := getControllerJson(srvc, encId)
	if err != nil {
		return nil, fmt.Errorf("getControllerJson failed: %s", err)
	}
	return json.Marshal(r)
}

func GetDocumentJson(srvc *native.NativeService) ([]byte, error) {
	log.Debug("GetDocumentJson")
	source := common.NewZeroCopySource(srvc.Input)
	// arg0: ID
	arg0, _, irregular, eof := source.NextVarBytes()
	if irregular || eof {
		return nil, fmt.Errorf("get key state failed: argument 0 error")
	}
	encId, err := encodeID(arg0)
	if err != nil {
		return nil, fmt.Errorf("encodeID failed: %s", err)
	}
	if !isValid(srvc, encId) {
		return nil, nil
	}
	contexts, err := getContextsWithDefault(srvc, encId)
	if err != nil {
		return nil, fmt.Errorf("getContextsWithDefault failed: %s", err)
	}
	id := string(arg0)
	publicKey, err := getAllPkJson(srvc, encId)
	if err != nil {
		return nil, fmt.Errorf("getAllPkJson failed: %s", err)
	}
	authentication, err := getAuthentication(srvc, encId)
	if err != nil {
		return nil, fmt.Errorf("getAuthentication failed: %s", err)
	}
	controller, err := getControllerJson(srvc, encId)
	if err != nil {
		return nil, fmt.Errorf("getController failed: %s", err)
	}
	recovery, err := getRecoveryJson(srvc, encId)
	if err != nil {
		return nil, fmt.Errorf("getRecovery failed: %s", err)
	}
	service, err := getServicesJson(srvc, encId)
	if err != nil {
		return nil, fmt.Errorf("getServicesJson failed: %s", err)
	}
	attribute, err := getAllAttrJson(srvc, encId)
	if err != nil {
		return nil, fmt.Errorf("getAllAttrJson failed: %s", err)
	}
	created, err := getCreateTime(srvc, encId)
	if err != nil {
		return nil, fmt.Errorf("getCreateTime failed: %s", err)
	}
	updated, err := getUpdateTime(srvc, encId)
	if err != nil {
		return nil, fmt.Errorf("getUpdateTime failed: %s", err)
	}
	proof, err := getProof(srvc, encId)
	if err != nil {
		return nil, fmt.Errorf("getProof failed: %s", err)
	}
	document := new(Document)
	document.Contexts = contexts
	document.Id = id
	document.PublicKey = publicKey
	document.Authentication = authentication
	document.Controller = controller
	document.Recovery = recovery
	document.Service = service
	document.Attribute = attribute
	document.Created = created
	document.Updated = updated
	document.Proof = proof
	return json.Marshal(document)
}
