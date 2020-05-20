/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or * (at your option) any later version.
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
	"fmt"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func isValid(srvc *native.NativeService, encId []byte) bool {
	return checkIDState(srvc, encId) == flag_valid
}

func checkIDState(srvc *native.NativeService, encId []byte) byte {
	val, err := srvc.CacheDB.Get(encId)
	if err == nil {
		val, err := states.GetValueFromRawStorageItem(val)
		if err == nil {
			if len(val) > 0 {
				return val[0]
			}
		}
	}
	return flag_not_exist
}

const (
	flag_not_exist byte = 0x00
	flag_valid     byte = 0x01
	flag_revoke    byte = 0x02

	FIELD_VERSION byte = 0
	FLAG_VERSION  byte = 0x01

	FIELD_PK         byte = 1
	FIELD_ATTR       byte = 2
	FIELD_RECOVERY   byte = 3
	FIELD_CONTROLLER byte = 4
	FIELD_SERVICE    byte = 5
	FIELD_CREATED    byte = 6
	FIELD_UPDATED    byte = 7
	FIELD_PROOF      byte = 8
	FIELD_CONTEXT    byte = 9
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
	if err == nil && srvc.ContextRef.CheckWitness(addr) {
		return nil
	}

	return errors.New("check witness failed, " + hex.EncodeToString(key))
}

func checkWitnessByIndex(srvc *native.NativeService, encId []byte, index uint32) error {
	pk, err := getPk(srvc, encId, index)
	if err != nil {
		return err
	} else if pk.revoked {
		return errors.New("revoked key")
	}

	//verify access
	if !pk.isAuthentication {
		return fmt.Errorf("pk do not have access")
	}

	return checkWitness(srvc, pk.key)
}

func checkWitnessWithoutAuth(srvc *native.NativeService, encId []byte, index uint32) error {
	pk, err := getPk(srvc, encId, index)
	if err != nil {
		return err
	} else if pk.revoked {
		return errors.New("revoked key")
	}

	return checkWitness(srvc, pk.key)
}

func deleteID(srvc *native.NativeService, encId []byte) error {
	key := append(encId, FIELD_PK)
	srvc.CacheDB.Delete(key)

	key = append(encId, FIELD_CONTROLLER)
	srvc.CacheDB.Delete(key)

	key = append(encId, FIELD_RECOVERY)
	srvc.CacheDB.Delete(key)
	if srvc.Height >= config.GetNewOntIdHeight() {
		key = append(encId, FIELD_SERVICE)
		srvc.CacheDB.Delete(key)

		key = append(encId, FIELD_CREATED)
		srvc.CacheDB.Delete(key)

		key = append(encId, FIELD_UPDATED)
		srvc.CacheDB.Delete(key)

		key = append(encId, FIELD_PROOF)
		srvc.CacheDB.Delete(key)

		key = append(encId, FIELD_CONTEXT)
		srvc.CacheDB.Delete(key)
	}
	err := deleteAllAttr(srvc, encId)
	if err != nil {
		return err
	}

	//set flag to revoke
	utils.PutBytes(srvc, encId, []byte{flag_revoke})
	return nil
}

func updateTimeAndClearProof(srvc *native.NativeService, encId []byte) {
	key := append(encId, FIELD_UPDATED)
	updateTime(srvc, key)
}

func createTimeAndClearProof(srvc *native.NativeService, encId []byte) {
	key := append(encId, FIELD_CREATED)
	updateTime(srvc, key)
}
