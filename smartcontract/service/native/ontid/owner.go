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
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func insertPk(srvc *native.NativeService, encID, pk []byte) (uint32, error) {
	var i uint32 = 0
	key1 := append(encID, FIELD_PK)
	item, err := utils.LinkedlistGetHead(srvc, key1)
	if err == nil && item != nil {
		i = binary.LittleEndian.Uint32(item)
	}
	i += 1

	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], i)
	key2 := buf[:]
	err = utils.LinkedlistInsert(srvc, key1, key2, pk)
	if err != nil {
		return 0, err
	}

	key3 := append(encID, FIELD_PK_STATE)
	key3 = append(key3, key2...)
	utils.PutBytes(srvc, key3, []byte{1})
	return i, nil
}

func getPk(srvc *native.NativeService, encID []byte, index uint32) ([]byte, error) {
	key1 := append(encID, FIELD_PK)
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], index)
	key2 := buf[:]
	node, err := utils.LinkedlistGetItem(srvc, key1, key2)
	if err != nil {
		return nil, err
	}
	data := node.GetPayload()
	if len(data) == 0 {
		return nil, errors.New("invalid public key data from storage")
	}

	return data, nil
}

func findPk(srvc *native.NativeService, encID, pub []byte) ([]byte, error) {
	key := append(encID, FIELD_PK)
	item, err := utils.LinkedlistGetHead(srvc, key)
	if err != nil {
		return nil, err
	}

	for len(item) > 0 {
		node, err := utils.LinkedlistGetItem(srvc, key, item)
		if err != nil {
			log.Debug(err)
			continue
		}
		if bytes.Equal(pub, node.GetPayload()) {
			return item, nil
		}
		item = node.GetNext()
	}

	return nil, errors.New("public key not found")
}

func revokePk(srvc *native.NativeService, encID, pub []byte) (uint32, error) {
	keyID, err := findPk(srvc, encID, pub)
	if err != nil {
		return 0, fmt.Errorf("cannot find the key, %s", err)
	}
	key := append(encID, FIELD_PK_STATE)
	key = append(key, keyID...)
	utils.PutBytes(srvc, key, []byte{0})
	return binary.LittleEndian.Uint32(keyID), nil
}

func isOwner(srvc *native.NativeService, encID, pub []byte) bool {
	kid, err := findPk(srvc, encID, pub)
	if err != nil || len(kid) == 0 {
		log.Debug(err)
		return false
	}
	return true
}
