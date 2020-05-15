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
	"errors"
	"fmt"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func addNewAuthKey(srvc *native.NativeService) ([]byte, error) {
	params := new(AddNewAuthKeyParam)
	if err := params.Deserialization(common.NewZeroCopySource(srvc.Input)); err != nil {
		return utils.BYTE_FALSE, errors.New("add new auth key error: deserialization params error, " + err.Error())
	}
	_, err := keypair.DeserializePublicKey(params.NewPublicKey.key)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add new auth key error: invalid key")
	}
	encId, err := encodeID(params.OntId)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add new auth key error: " + err.Error())
	}

	if checkIDState(srvc, encId) == flag_not_exist {
		return utils.BYTE_FALSE, errors.New("add new auth key error: have not registered")
	}

	if err := checkWitnessByIndex(srvc, encId, params.SignIndex); err != nil {
		return utils.BYTE_FALSE, errors.New("verify signature failed: " + err.Error())
	}

	index, err := insertPk(srvc, encId, params.NewPublicKey.key, params.NewPublicKey.controller,
		USE_ACCESS, ONLY_AUTHENTICATION)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add auth key error, insertPk failed " + err.Error())
	}
	triggerAuthKeyEvent(srvc, "add", params.OntId, index)

	updateProofAndTime(srvc, encId, params.Proof)
	return utils.BYTE_TRUE, nil
}

func setAuthKey(srvc *native.NativeService) ([]byte, error) {
	params := new(SetAuthKeyParam)
	if err := params.Deserialization(common.NewZeroCopySource(srvc.Input)); err != nil {
		return utils.BYTE_FALSE, errors.New("set auth key error: deserialization params error, " + err.Error())
	}
	encId, err := encodeID(params.OntId)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("set auth key error: " + err.Error())
	}

	if checkIDState(srvc, encId) == flag_not_exist {
		return utils.BYTE_FALSE, errors.New("set auth key error: have not registered")
	}

	if err := checkWitnessByIndex(srvc, encId, params.SignIndex); err != nil {
		return utils.BYTE_FALSE, errors.New("verify signature failed: " + err.Error())
	}

	err = changePkAuthentication(srvc, encId, params.Index, BOTH, params.Proof)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add auth key error, changePkAuthentication failed " + err.Error())
	}
	triggerAuthKeyEvent(srvc, "set", params.OntId, params.Index)

	updateProofAndTime(srvc, encId, params.Proof)
	return utils.BYTE_TRUE, nil
}

func removeAuthKey(srvc *native.NativeService) ([]byte, error) {
	params := new(RemoveAuthKeyParam)
	if err := params.Deserialization(common.NewZeroCopySource(srvc.Input)); err != nil {
		return utils.BYTE_FALSE, errors.New("remove auth key error: deserialization params error, " + err.Error())
	}
	encId, err := encodeID(params.OntId)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("remove auth key error: " + err.Error())
	}

	if checkIDState(srvc, encId) == flag_not_exist {
		return utils.BYTE_FALSE, errors.New("remove auth key error: have not registered")
	}

	if err := checkWitnessByIndex(srvc, encId, params.SignIndex); err != nil {
		return utils.BYTE_FALSE, errors.New("verify signature failed: " + err.Error())
	}

	if err := revokeAuthKey(srvc, encId, params.Index, params.Proof); err != nil {
		return utils.BYTE_FALSE, errors.New("remove auth key error, revokeAuthKey failed: " + err.Error())
	}

	updateProofAndTime(srvc, encId, params.Proof)
	triggerAuthKeyEvent(srvc, "remove", params.OntId, params.Index)
	return utils.BYTE_TRUE, nil
}

func getAuthentication(srvc *native.NativeService, encId []byte) ([]interface{}, error) {
	key := append(encId, FIELD_PK)
	publicKeys, err := getAllPk_Version1(srvc, encId, key)
	if err != nil {
		return nil, err
	}
	authentication := make([]interface{}, 0)
	for index, p := range publicKeys {
		if !p.revoked {
			ontId, err := decodeID(encId)
			if err != nil {
				return nil, err
			}
			if p.authentication == ONLY_AUTHENTICATION {
				publicKey := new(publicKeyJson)
				publicKey.Id = fmt.Sprintf("%s#keys-%d", string(ontId), index+1)
				publicKey.Controller = string(p.controller)
				publicKey.Type, publicKey.PublicKeyHex, err = keyType(p.key)
				if err != nil {
					return nil, err
				}
				publicKey.Access = p.access
				authentication = append(authentication, publicKey)
			}
			if p.authentication == BOTH {
				authentication = append(authentication, fmt.Sprintf("%s#keys-%d", string(ontId), index+1))
			}
		}
	}
	return authentication, nil
}
