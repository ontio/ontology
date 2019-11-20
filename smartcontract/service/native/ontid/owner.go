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
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const OWNER_TOTAL_SIZE = 1024 * 1024 // 1MB

type owner struct {
	key     []byte
	revoked bool
}

func (this *owner) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.key)
	sink.WriteBool(this.revoked)
}

func (this *owner) Deserialization(source *common.ZeroCopySource) error {
	v1, err := utils.DecodeVarBytes(source)
	if err != nil {
		return err
	}
	v2, err := utils.DecodeBool(source)
	if err != nil {
		return err
	}
	this.key = v1
	this.revoked = v2
	return nil
}

func getAllPk(srvc *native.NativeService, key []byte) ([]*owner, error) {
	val, err := utils.GetStorageItem(srvc, key)
	if err != nil {
		return nil, fmt.Errorf("get storage error, %s", err)
	}
	if val == nil {
		return nil, nil
	}
	source := common.NewZeroCopySource(val.Value)
	owners := make([]*owner, 0)
	for source.Len() > 0 {
		var t = new(owner)
		err = t.Deserialization(source)
		if err != nil {
			return nil, fmt.Errorf("deserialize owners error, %s", err)
		}
		owners = append(owners, t)
	}
	return owners, nil
}

func putAllPk(srvc *native.NativeService, key []byte, val []*owner) error {
	sink := common.NewZeroCopySink(nil)
	for _, i := range val {
		i.Serialization(sink)
	}
	var v states.StorageItem
	v.Value = sink.Bytes()
	if len(v.Value) > OWNER_TOTAL_SIZE {
		return errors.New("total key size is out of range")
	}
	srvc.CacheDB.Put(key, v.ToArray())
	return nil
}

func insertPk(srvc *native.NativeService, encID, pk []byte) (uint32, error) {
	key := append(encID, FIELD_PK)
	owners, err := getAllPk(srvc, key)
	if err != nil {
		owners = make([]*owner, 0)
	}
	for _, k := range owners {
		if bytes.Equal(k.key, pk) {
			return 0, errors.New("the key is already added")
		}
	}
	size := len(owners)
	owners = append(owners, &owner{pk, false})
	err = putAllPk(srvc, key, owners)
	if err != nil {
		return 0, err
	}
	return uint32(size + 1), nil
}

func getPk(srvc *native.NativeService, encID []byte, index uint32) (*owner, error) {
	key := append(encID, FIELD_PK)
	owners, err := getAllPk(srvc, key)
	if err != nil {
		return nil, err
	}
	if index < 1 || index > uint32(len(owners)) {
		return nil, errors.New("invalid key index")
	}
	return owners[index-1], nil
}

func findPk(srvc *native.NativeService, encID, pub []byte) (uint32, bool, error) {
	key := append(encID, FIELD_PK)
	owners, err := getAllPk(srvc, key)
	if err != nil {
		return 0, false, err
	}
	for i, v := range owners {
		if bytes.Equal(pub, v.key) {
			return uint32(i + 1), v.revoked, nil
		}
	}
	return 0, false, nil
}

func revokePk(srvc *native.NativeService, encID, pub []byte) (uint32, error) {
	key := append(encID, FIELD_PK)
	owners, err := getAllPk(srvc, key)
	if err != nil {
		return 0, err
	}
	var index uint32 = 0
	for i, v := range owners {
		if bytes.Equal(pub, v.key) {
			index = uint32(i + 1)
			if v.revoked {
				return index, errors.New("public key has already been revoked")
			}
			v.revoked = true
		}
	}
	if index == 0 {
		return 0, errors.New("revoke failed, public key not found")
	}
	putAllPk(srvc, key, owners)
	return index, nil
}

func revokePkByIndex(srvc *native.NativeService, encID []byte, index uint32) ([]byte, error) {
	key := append(encID, FIELD_PK)
	owners, err := getAllPk(srvc, key)
	if err != nil {
		return nil, err
	}
	if uint32(len(owners)) < index {
		return nil, errors.New("no such key")
	}
	index -= 1
	if owners[index].revoked {
		return nil, errors.New("already revoked")
	}
	owners[index].revoked = true
	putAllPk(srvc, key, owners)
	return owners[index].key, nil
}

func isOwner(srvc *native.NativeService, encID, pub []byte) bool {
	kID, revoked, err := findPk(srvc, encID, pub)
	if err != nil {
		log.Debug(err)
		return false
	}
	return kID != 0 && !revoked
}
