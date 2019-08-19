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

	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func addRecovery(srvc *native.NativeService) ([]byte, error) {
	args := bytes.NewBuffer(srvc.Input)
	// arg0: ID
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add recovery failed: argument 0 error")
	}
	// arg1: recovery struct
	arg1, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add recovery failed: argument 1 error")
	}
	// arg2: operator's public key index
	arg2, err := utils.ReadVarUint(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add recovery failed: argument 2 error")
	}

	encId, err := encodeID(arg0)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add recovery failed: " + err.Error())
	}
	pk, err := getPk(srvc, encId, uint32(arg2))
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	if pk.revoked {
		return utils.BYTE_FALSE, errors.New("authentication failed, public key is revoked")
	}
	err = checkWitness(srvc, pk.key)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("checkWitness failed")
	}

	re, err := getRecovery(srvc, encId)
	if err == nil && re != nil {
		return utils.BYTE_FALSE, errors.New("recovery is already set")
	}

	re, err = setRecovery(srvc, encId, arg1)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add recovery failed: " + err.Error())
	}

	newEvent(srvc, []interface{}{"recovery", "add", string(arg0), re.ToJson()})
	return utils.BYTE_TRUE, nil
}

func changeRecovery(srvc *native.NativeService) ([]byte, error) {
	args := bytes.NewBuffer(srvc.Input)
	// arg0: ID
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 0 error")
	}
	// arg1: new recovery
	arg1, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 1 error")
	}
	// arg2: signers
	arg2, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 2 error")
	}

	key, err := encodeID(arg0)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("change recovery failed: " + err.Error())
	}
	re, err := getRecovery(srvc, key)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("change recovery failed: recovery not set")
	}
	signers, err := deserializeSigners(arg2)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("signers error: " + err.Error())
	}

	if !verifyGroupSignature(srvc, re, signers) {
		return utils.BYTE_FALSE, errors.New("verification failed")
	}
	re, err = setRecovery(srvc, key, arg1)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("change recovery failed: " + err.Error())
	}

	newEvent(srvc, []interface{}{"Recovery", "change", string(arg0), re.ToJson()})
	return utils.BYTE_TRUE, nil
}

func addKeyByRecovery(srvc *native.NativeService) ([]byte, error) {
	args := bytes.NewBuffer(srvc.Input)
	// arg0: id
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 0 error")
	}
	// arg1: public key
	arg1, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 1 error")
	}
	// arg2: signers
	arg2, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 2 error")
	}

	encId, err := encodeID(arg0)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	signers, err := deserializeSigners(arg2)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("signers error, %s", err)
	}

	rec, err := getRecovery(srvc, encId)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	if !verifyGroupSignature(srvc, rec, signers) {
		return utils.BYTE_FALSE, errors.New("verification failed")
	}

	index, err := insertPk(srvc, encId, arg1)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	triggerPublicEvent(srvc, "add", arg0, arg1, index)
	return utils.BYTE_TRUE, nil
}

func removeKeyByRecovery(srvc *native.NativeService) ([]byte, error) {
	args := bytes.NewBuffer(srvc.Input)
	// arg0: id
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 0 error")
	}
	// arg1: public key index
	arg1, err := utils.ReadVarUint(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 1 error")
	}
	// arg2: signers
	arg2, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 2 error")
	}

	encId, err := encodeID(arg0)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	signers, err := deserializeSigners(arg2)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("signers error, %s", err)
	}

	rec, err := getRecovery(srvc, encId)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	if !verifyGroupSignature(srvc, rec, signers) {
		return utils.BYTE_FALSE, errors.New("verification failed")
	}

	pk, err := revokePkByIndex(srvc, encId, uint32(arg1))
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	triggerPublicEvent(srvc, "remove", arg0, pk, uint32(arg1))
	return utils.BYTE_TRUE, nil
}

func setRecovery(srvc *native.NativeService, encID, data []byte) (*Group, error) {
	rec, err := deserializeGroup(data)
	if err != nil {
		return nil, err
	}
	err = validateMembers(srvc, rec)
	if err != nil {
		return nil, fmt.Errorf("invalid recovery member, %s", err)
	}
	key := append(encID, FIELD_RECOVERY)
	utils.PutBytes(srvc, key, data)
	return rec, nil
}

func getRecovery(srvc *native.NativeService, encID []byte) (*Group, error) {
	key := append(encID, FIELD_RECOVERY)
	item, err := utils.GetStorageItem(srvc, key)
	if err != nil {
		return nil, err
	} else if item == nil {
		return nil, errors.New("empty storage item")
	}
	return deserializeGroup(item.Value)
}
