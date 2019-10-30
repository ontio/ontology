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
	"encoding/hex"
	"errors"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func checkIDExistence(srvc *native.NativeService, encID []byte) bool {
	val, err := srvc.CacheDB.Get(encID)
	if err == nil {
		val, err := states.GetValueFromRawStorageItem(val)
		if err == nil {
			if len(val) > 0 && val[0] == flag_exist {
				return true
			}
		}
	}
	return false
}

const (
	flag_exist  byte = 0x01
	flag_revoke byte = 0x02

	FIELD_VERSION byte = 0
	FLAG_VERSION  byte = 0x01

	FIELD_PK         byte = 1
	FIELD_ATTR       byte = 2
	FIELD_RECOVERY   byte = 3
	FIELD_CONTROLLER byte = 4
)

func encodeID(id []byte) ([]byte, error) {
	length := len(id)
	if length == 0 || length > 255 {
		return nil, errors.New("encode ONT ID error: invalid ID length")
	}
	//enc := []byte{byte(length)}
	enc := append(utils.OntIDContractAddress[:], byte(length))
	enc = append(enc, id...)
	return enc, nil
}

func decodeID(data []byte) ([]byte, error) {
	prefix := len(utils.OntIDContractAddress)
	size := len(data)
	if size < prefix || size != int(data[prefix])+1+prefix {
		return nil, errors.New("decode ONT ID error: invalid data length")
	}
	return data[prefix+1:], nil
}

func checkWitness(srvc *native.NativeService, key []byte) error {
	// try as if key is a public key
	pk, err := keypair.DeserializePublicKey(key)
	if err == nil {
		addr := types.AddressFromPubKey(pk)
		if srvc.ContextRef.CheckWitness(addr) {
			return nil
		}
	}

	// try as if key is an address
	addr, err := common.AddressParseFromBytes(key)
	if srvc.ContextRef.CheckWitness(addr) {
		return nil
	}

	return errors.New("check witness failed, " + hex.EncodeToString(key))
}

func deleteID(srvc *native.NativeService, encID []byte) error {
	key := append(encID, FIELD_PK)
	srvc.CacheDB.Delete(key)

	key = append(encID, FIELD_CONTROLLER)
	srvc.CacheDB.Delete(key)

	key = append(encID, FIELD_RECOVERY)
	srvc.CacheDB.Delete(key)

	err := deleteAllAttr(srvc, encID)
	if err != nil {
		return err
	}

	//set flag to revoke
	utils.PutBytes(srvc, encID, []byte{flag_revoke})
	return nil
}
