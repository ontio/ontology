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

func setRecovery(srvc *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(srvc.Input)
	// arg0: ID
	arg0, err := utils.DecodeVarBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("set recovery failed: argument 0 error")
	}
	// arg1: recovery struct
	arg1, err := utils.DecodeVarBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("set recovery failed: argument 1 error")
	}
	// arg2: operator's public key index
	arg2, err := utils.DecodeVarUint(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("set recovery failed: argument 2 error")
	}

	encId, err := encodeID(arg0)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("set recovery failed: " + err.Error())
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

	re, err = putRecovery(srvc, encId, arg1)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("set recovery failed: " + err.Error())
	}

	newEvent(srvc, []interface{}{"recovery", "set", string(arg0), re.ToJson()})
	return utils.BYTE_TRUE, nil
}

func updateRecovery(srvc *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(srvc.Input)
	// arg0: ID
	arg0, err := utils.DecodeVarBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 0 error")
	}
	// arg1: new recovery
	arg1, err := utils.DecodeVarBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 1 error")
	}
	// arg2: signers
	arg2, err := utils.DecodeVarBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 2 error")
	}

	key, err := encodeID(arg0)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("update recovery failed: " + err.Error())
	}
	re, err := getRecovery(srvc, key)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("update recovery failed: recovery not set")
	}
	signers, err := deserializeSigners(arg2)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("signers error: " + err.Error())
	}

	if !verifyGroupSignature(srvc, re, signers) {
		return utils.BYTE_FALSE, errors.New("verification failed")
	}
	re, err = putRecovery(srvc, key, arg1)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("update recovery failed: " + err.Error())
	}

	newEvent(srvc, []interface{}{"Recovery", "update", string(arg0), re.ToJson()})
	return utils.BYTE_TRUE, nil
}

func addKeyByRecovery(srvc *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(srvc.Input)
	// arg0: id
	arg0, err := utils.DecodeVarBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 0 error")
	}
	// arg1: public key
	arg1, err := utils.DecodeVarBytes(source)
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
	source := common.NewZeroCopySource(srvc.Input)
	// arg0: id
	arg0, err := utils.DecodeVarBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 0 error")
	}
	// arg1: public key index
	arg1, err := utils.DecodeVarUint(source)
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

func putRecovery(srvc *native.NativeService, encID, data []byte) (*Group, error) {
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

// deprecated
// retain for conpatibility
func addRecovery(srvc *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(srvc.Input)
	// arg0: ID
	arg0, err := utils.DecodeVarBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add recovery failed: argument 0 error")
	}
	// arg1: recovery address
	arg1, err := utils.DecodeAddress(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add recovery failed: argument 1 error")
	}
	// arg2: operator's public key
	arg2, err := utils.DecodeVarBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add recovery failed: argument 2 error")
	}

	err = checkWitness(srvc, arg2)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add recovery failed: " + err.Error())
	}

	key, err := encodeID(arg0)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add recovery failed: " + err.Error())
	}
	if !isValid(srvc, key) {
		return utils.BYTE_FALSE, errors.New("add recovery failed: ID not registered")
	}
	if !isOwner(srvc, key, arg2) {
		return utils.BYTE_FALSE, errors.New("add recovery failed: not authorized")
	}

	re, err := getOldRecovery(srvc, key)
	if err == nil && len(re) > 0 {
		return utils.BYTE_FALSE, errors.New("add recovery failed: already set recovery")
	}

	err = setOldRecovery(srvc, key, arg1)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add recovery failed: " + err.Error())
	}

	triggerRecoveryEvent(srvc, "add", arg0, arg1)

	return utils.BYTE_TRUE, nil
}

// deprecated
// retain for conpatibility
func changeRecovery(srvc *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(srvc.Input)
	// arg0: ID
	arg0, err := utils.DecodeVarBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("change recovery failed: argument 0 error")
	}
	// arg1: new recovery address
	arg1, err := utils.DecodeAddress(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("change recovery failed: argument 1 error")
	}
	// arg2: operator's address, who should be the old recovery
	arg2, err := utils.DecodeAddress(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("change recovery failed: argument 2 error")
	}

	key, err := encodeID(arg0)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("change recovery failed: " + err.Error())
	}
	re, err := getOldRecovery(srvc, key)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("change recovery failed: recovery not set")
	}
	if !bytes.Equal(re, arg2[:]) {
		return utils.BYTE_FALSE, errors.New("change recovery failed: operator is not the recovery")
	}
	err = checkWitness(srvc, arg2[:])
	if err != nil {
		return utils.BYTE_FALSE, errors.New("change recovery failed: " + err.Error())
	}
	if !isValid(srvc, key) {
		return utils.BYTE_FALSE, errors.New("change recovery failed: ID not registered")
	}
	err = setOldRecovery(srvc, key, arg1)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("change recovery failed: " + err.Error())
	}

	triggerRecoveryEvent(srvc, "change", arg0, arg1)
	return utils.BYTE_TRUE, nil
}

// deprecated
// retain for conpatibility
func setOldRecovery(srvc *native.NativeService, encID []byte, recovery common.Address) error {
	key := append(encID, FIELD_RECOVERY)
	val := states.StorageItem{Value: recovery[:]}
	srvc.CacheDB.Put(key, val.ToArray())
	return nil
}

// deprecated
// retain for conpatibility
func getOldRecovery(srvc *native.NativeService, encID []byte) ([]byte, error) {
	key := append(encID, FIELD_RECOVERY)
	item, err := utils.GetStorageItem(srvc, key)
	if err != nil {
		return nil, errors.New("get recovery error: " + err.Error())
	} else if item == nil {
		return nil, nil
	}
	return item.Value, nil
}
