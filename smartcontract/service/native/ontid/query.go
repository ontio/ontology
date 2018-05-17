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
	"errors"
	"fmt"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/service/native"
)

func GetPublicKeyByID(srvc *native.NativeService) ([]byte, error) {
	args := bytes.NewBuffer(srvc.Input)
	// arg0: ID
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return nil, errors.New("get public key failed: argument 0 error")
	}
	// arg1: key ID
	arg1, err := serialization.ReadUint32(args)
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

func GetDDO(srvc *native.NativeService) ([]byte, error) {
	log.Debug("GetDDO")
	var0, err := GetPublicKeys(srvc)
	if err != nil {
		return nil, fmt.Errorf("get DDO error: %s", err)
	} else if var0 == nil {
		log.Debug("DDO: null")
		return nil, nil
	}
	var buf bytes.Buffer
	serialization.WriteVarBytes(&buf, var0)

	var1, err := GetAttributes(srvc)
	serialization.WriteVarBytes(&buf, var1)

	args := bytes.NewBuffer(srvc.Input)
	did, _ := serialization.ReadVarBytes(args)
	key, _ := encodeID(did)
	var2, err := getRecovery(srvc, key)
	serialization.WriteVarBytes(&buf, var2)

	res := buf.Bytes()
	log.Debug("DDO:", hex.EncodeToString(res))
	return res, nil
}

func GetPublicKeys(srvc *native.NativeService) ([]byte, error) {
	log.Debug("GetPublicKeys")
	args := bytes.NewBuffer(srvc.Input)
	did, err := serialization.ReadVarBytes(args)
	if err != nil {
		return nil, fmt.Errorf("get public keys error: invalid argument", err)
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

	var res bytes.Buffer
	for i, v := range list {
		if v.revoked {
			continue
		}
		err = serialization.WriteUint32(&res, uint32(i+1))
		if err != nil {
			return nil, fmt.Errorf("get public keys error: %s", err)
		}
		err = serialization.WriteVarBytes(&res, v.key)
		if err != nil {
			return nil, fmt.Errorf("get public keys error: %s", err)
		}
	}

	return res.Bytes(), nil
}

func GetAttributes(srvc *native.NativeService) ([]byte, error) {
	log.Debug("GetAttributes")
	args := bytes.NewBuffer(srvc.Input)
	did, err := serialization.ReadVarBytes(args)
	if err != nil {
		return nil, fmt.Errorf("get public keys error: invalid argument", err)
	}
	if len(did) == 0 {
		return nil, errors.New("get attributes error: invalid ID")
	}
	key, err := encodeID(did)
	if err != nil {
		return nil, fmt.Errorf("get public keys error: %s", err)
	}
	res, err := getAllAttr(srvc, key)
	if err != nil {
		return nil, fmt.Errorf("get attributes error: %s", err)
	}

	return res, nil
}

func GetKeyState(srvc *native.NativeService) ([]byte, error) {
	log.Debug("GetKeyState")
	args := bytes.NewBuffer(srvc.Input)
	// arg0: ID
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return nil, fmt.Errorf("get key state failed: argument 0 error, %s", err)
	}
	// arg1: public key ID
	arg1, err := serialization.ReadUint32(args)
	if err != nil {
		return nil, fmt.Errorf("get key state failed: argument 1 error, %s", err)
	}

	key, err := encodeID(arg0)
	if err != nil {
		return nil, fmt.Errorf("get key state failed: %s", err)
	}

	owner, err := getPk(srvc, key, arg1)
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
