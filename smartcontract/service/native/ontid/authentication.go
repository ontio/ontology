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
	if !isValid(srvc, encId) {
		return utils.BYTE_FALSE, errors.New("addNewAuthKey error: have not registered")
	}

	if err := checkWitnessByIndex(srvc, encId, params.SignIndex); err != nil {
		return utils.BYTE_FALSE, errors.New("verify signature failed: " + err.Error())
	}

	index, err := insertPk(srvc, encId, params.NewPublicKey.key, params.NewPublicKey.controller,
		false, true)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add auth key error, insertPk failed " + err.Error())
	}
	triggerAuthKeyEvent(srvc, "add", params.OntId, index)

	updateTimeAndClearProof(srvc, encId)
	return utils.BYTE_TRUE, nil
}

func addNewAuthKeyByRecovery(srvc *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(srvc.Input)
	// arg0: id
	arg0, err := utils.DecodeVarBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 0 error")
	}
	// arg1: new public key key
	key, err := utils.DecodeVarBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 1 error")
	}
	_, err = keypair.DeserializePublicKey(key)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add new auth key error: invalid key")
	}
	// arg2: new public key controller
	controller, err := utils.DecodeVarBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 1 error")
	}
	// arg3: signers
	arg2, err := utils.DecodeVarBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 2 error")
	}

	encId, err := encodeID(arg0)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	if !isValid(srvc, encId) {
		return utils.BYTE_FALSE, errors.New("add new auth key error: have not registered")
	}

	signers, err := deserializeSigners(arg2)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("signers error, %s", err)
	}

	rec, err := getRecovery(srvc, encId)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	if rec == nil {
		return utils.BYTE_FALSE, errors.New("recovery is not exist")
	}

	if !verifyGroupSignature(srvc, rec, signers) {
		return utils.BYTE_FALSE, errors.New("verification failed")
	}

	index, err := insertPk(srvc, encId, key, controller, false, true)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add auth key error, insertPk failed " + err.Error())
	}
	triggerAuthKeyEvent(srvc, "add", arg0, index)

	updateTimeAndClearProof(srvc, encId)
	return utils.BYTE_TRUE, nil
}

func addNewAuthKeyByController(srvc *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(srvc.Input)
	// arg0: id
	arg0, err := utils.DecodeVarBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 0")
	}

	// arg1: new public key key
	key, err := utils.DecodeVarBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 1 error")
	}
	_, err = keypair.DeserializePublicKey(key)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add new auth key error: invalid key")
	}
	// arg2: new public key controller
	controller, err := utils.DecodeVarBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 1 error")
	}

	encId, err := encodeID(arg0)
	if err != nil {
		return utils.BYTE_FALSE, errors.New(err.Error())
	}
	if !isValid(srvc, encId) {
		return utils.BYTE_FALSE, errors.New("add new auth key error: have not registered")
	}

	err = verifyControllerSignature(srvc, encId, source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("verifying signature failed")
	}

	index, err := insertPk(srvc, encId, key, controller, false, true)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add auth key error, insertPk failed " + err.Error())
	}
	triggerAuthKeyEvent(srvc, "add", arg0, index)

	updateTimeAndClearProof(srvc, encId)
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
	if !isValid(srvc, encId) {
		return utils.BYTE_FALSE, errors.New("setAuthKey error: have not registered")
	}

	if err := checkWitnessByIndex(srvc, encId, params.SignIndex); err != nil {
		return utils.BYTE_FALSE, errors.New("verify signature failed: " + err.Error())
	}

	err = changePkAuthentication(srvc, encId, params.Index, true)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add auth key error, changePkAuthentication failed " + err.Error())
	}
	triggerAuthKeyEvent(srvc, "set", params.OntId, params.Index)

	updateTimeAndClearProof(srvc, encId)
	return utils.BYTE_TRUE, nil
}

func setAuthKeyByRecovery(srvc *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(srvc.Input)
	// arg0: id
	arg0, err := utils.DecodeVarBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 0 error")
	}
	// arg1: index
	index, err := utils.DecodeVarUint(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 1 error")
	}
	// arg2: signers
	arg2, err := utils.DecodeVarBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 2 error")
	}

	encId, err := encodeID(arg0)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	if !isValid(srvc, encId) {
		return utils.BYTE_FALSE, errors.New("setAuthKeyByRecovery error: have not registered")
	}

	signers, err := deserializeSigners(arg2)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("signers error, %s", err)
	}

	rec, err := getRecovery(srvc, encId)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	if rec == nil {
		return utils.BYTE_FALSE, errors.New("recovery is not exist")
	}

	if !verifyGroupSignature(srvc, rec, signers) {
		return utils.BYTE_FALSE, errors.New("verification failed")
	}

	err = changePkAuthentication(srvc, encId, uint32(index), true)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add auth key error, changePkAuthentication failed " + err.Error())
	}
	triggerAuthKeyEvent(srvc, "set", arg0, uint32(index))

	updateTimeAndClearProof(srvc, encId)
	return utils.BYTE_TRUE, nil
}

func setAuthKeyByController(srvc *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(srvc.Input)
	// arg0: id
	arg0, err := utils.DecodeVarBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 0")
	}

	// arg1: index
	index, err := utils.DecodeVarUint(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 1 error")
	}

	encId, err := encodeID(arg0)
	if err != nil {
		return utils.BYTE_FALSE, errors.New(err.Error())
	}
	if !isValid(srvc, encId) {
		return utils.BYTE_FALSE, errors.New("setAuthKeyByController error: have not registered")
	}

	err = verifyControllerSignature(srvc, encId, source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("verifying signature failed")
	}

	err = changePkAuthentication(srvc, encId, uint32(index), true)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add auth key error, changePkAuthentication failed " + err.Error())
	}
	triggerAuthKeyEvent(srvc, "set", arg0, uint32(index))

	updateTimeAndClearProof(srvc, encId)
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
	if !isValid(srvc, encId) {
		return utils.BYTE_FALSE, errors.New("removeAuthKey error: have not registered")
	}

	if err := checkWitnessByIndex(srvc, encId, params.SignIndex); err != nil {
		return utils.BYTE_FALSE, errors.New("verify signature failed: " + err.Error())
	}

	if err := changePkAuthentication(srvc, encId, uint32(params.Index), false); err != nil {
		return utils.BYTE_FALSE, errors.New("remove auth key error, revokeAuthKey failed: " + err.Error())
	}

	updateTimeAndClearProof(srvc, encId)
	triggerAuthKeyEvent(srvc, "remove", params.OntId, params.Index)
	return utils.BYTE_TRUE, nil
}

func removeAuthKeyByRecovery(srvc *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(srvc.Input)
	// arg0: id
	arg0, err := utils.DecodeVarBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 0 error")
	}
	// arg1: index
	index, err := utils.DecodeVarUint(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 1 error")
	}
	// arg2: signers
	arg2, err := utils.DecodeVarBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 2 error")
	}

	encId, err := encodeID(arg0)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	if !isValid(srvc, encId) {
		return utils.BYTE_FALSE, errors.New("removeAuthKeyByRecovery error: have not registered")
	}

	signers, err := deserializeSigners(arg2)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("signers error, %s", err)
	}

	rec, err := getRecovery(srvc, encId)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	if rec == nil {
		return utils.BYTE_FALSE, errors.New("recovery is not exist")
	}

	if !verifyGroupSignature(srvc, rec, signers) {
		return utils.BYTE_FALSE, errors.New("verification failed")
	}

	if err := changePkAuthentication(srvc, encId, uint32(index), false); err != nil {
		return utils.BYTE_FALSE, errors.New("remove auth key error, revokeAuthKey failed: " + err.Error())
	}

	updateTimeAndClearProof(srvc, encId)
	triggerAuthKeyEvent(srvc, "remove", arg0, uint32(index))
	return utils.BYTE_TRUE, nil
}

func removeAuthKeyByController(srvc *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(srvc.Input)
	// arg0: id
	arg0, err := utils.DecodeVarBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 0")
	}

	// arg1: index
	index, err := utils.DecodeVarUint(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 1 error")
	}

	encId, err := encodeID(arg0)
	if err != nil {
		return utils.BYTE_FALSE, errors.New(err.Error())
	}
	if !isValid(srvc, encId) {
		return utils.BYTE_FALSE, errors.New("removeAuthKeyByController error: have not registered")
	}

	err = verifyControllerSignature(srvc, encId, source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("verifying signature failed")
	}

	if err := changePkAuthentication(srvc, encId, uint32(index), false); err != nil {
		return utils.BYTE_FALSE, errors.New("remove auth key error, revokeAuthKey failed: " + err.Error())
	}

	updateTimeAndClearProof(srvc, encId)
	triggerAuthKeyEvent(srvc, "remove", arg0, uint32(index))
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
			if p.isAuthentication && !p.isPkList {
				publicKey := new(publicKeyJson)
				publicKey.Id = fmt.Sprintf("%s#keys-%d", string(ontId), index+1)
				publicKey.Controller = string(p.controller)
				publicKey.Type, publicKey.PublicKeyHex, err = keyType(p.key)
				if err != nil {
					return nil, err
				}
				authentication = append(authentication, publicKey)
			}
			if p.isAuthentication && p.isPkList {
				authentication = append(authentication, fmt.Sprintf("%s#keys-%d", string(ontId), index+1))
			}
		}
	}
	return authentication, nil
}
