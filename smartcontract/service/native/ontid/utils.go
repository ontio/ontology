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
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"

	"github.com/itchyny/base58-go"
	"github.com/ontio/ontology-crypto/keypair"
	com "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const flag_exist = 0x01

func checkIDExistence(srvc *native.NativeService, encID []byte) bool {
	val, err := srvc.CloneCache.Get(common.ST_STORAGE, encID)
	if err == nil {
		t, ok := val.(*states.StorageItem)
		if ok {
			if len(t.Value) > 0 && t.Value[0] == flag_exist {
				return true
			}
		}
	}
	return false
}

const (
	FIELD_PK byte = 1 + iota
	FIELD_ATTR
	FIELD_RECOVERY
)

func encodeID(id []byte) ([]byte, error) {
	length := len(id)
	if length == 0 || length > 255 {
		return nil, errors.New("encode ONT ID error: invalid ID length")
	}
	enc := []byte{byte(length)}
	enc = append(enc, id...)
	return enc, nil
}

func decodeID(data []byte) ([]byte, error) {
	if len(data) == 0 || len(data) != int(data[0])+1 {
		return nil, errors.New("decode ONT ID error: invalid data length")
	}
	return data[1:], nil
}

func verifyID(id []byte) bool {
	if len(id) < 9 {
		return false
	}
	if string(id[0:8]) != "did:ont:" {
		return false
	}
	buf, err := base58.BitcoinEncoding.Decode(id[8:])
	if err != nil {
		return false
	}
	bi, ok := new(big.Int).SetString(string(buf), 10)
	if !ok || bi == nil {
		return false
	}
	buf = bi.Bytes()
	// 1 byte version + 20 byte hash + 4 byte checksum
	if len(buf) != 25 {
		return false
	}
	pos := len(buf) - 4
	data := buf[:pos]
	checksum := buf[pos:]
	sum := sha256.Sum256(data)
	sum = sha256.Sum256(sum[:])
	if !bytes.Equal(sum[0:4], checksum) {
		return false
	}
	return true
}

func setRecovery(srvc *native.NativeService, encID []byte, recovery com.Address) error {
	key := append(encID, FIELD_RECOVERY)
	val := &states.StorageItem{Value: recovery[:]}
	srvc.CloneCache.Add(common.ST_STORAGE, key, val)
	return nil
}

func getRecovery(srvc *native.NativeService, encID []byte) ([]byte, error) {
	key := append(encID, FIELD_RECOVERY)
	item, err := utils.GetStorageItem(srvc, key)
	if err != nil {
		return nil, errors.New("get recovery error: " + err.Error())
	} else if item == nil {
		return nil, nil
	}
	return item.Value, nil
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
	addr, err := com.AddressParseFromBytes(key)
	if srvc.ContextRef.CheckWitness(addr) {
		return nil
	}

	return errors.New("check witness failed, " + hex.EncodeToString(key))
}
