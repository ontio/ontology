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
	"io"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type owner struct {
	key     []byte
	revoked bool
}

func (this *owner) Serialize(w io.Writer) error {
	if err := serialization.WriteVarBytes(w, this.key); err != nil {
		return err
	}
	if err := serialization.WriteBool(w, this.revoked); err != nil {
		return err
	}
	return nil
}

func (this *owner) Deserialize(r io.Reader) error {
	v1, err := serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}
	v2, err := serialization.ReadBool(r)
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
	buf := bytes.NewBuffer(val.Value)
	owners := make([]*owner, 0)
	for buf.Len() > 0 {
		var t = new(owner)
		err = t.Deserialize(buf)
		if err != nil {
			return nil, fmt.Errorf("deserialize owners error, %s", err)
		}
		owners = append(owners, t)
	}
	return owners, nil
}

func putAllPk(srvc *native.NativeService, key []byte, val []*owner) error {
	var buf bytes.Buffer
	for _, i := range val {
		err := i.Serialize(&buf)
		if err != nil {
			return fmt.Errorf("serialize owner error, %s", err)
		}
	}
	var v states.StorageItem
	v.Value = buf.Bytes()
	srvc.CloneCache.Add(common.ST_STORAGE, key, &v)
	return nil
}

func insertPk(srvc *native.NativeService, encID, pk []byte) (uint32, error) {
	key := append(encID, FIELD_PK)
	owners, err := getAllPk(srvc, key)
	if err != nil {
		owners = make([]*owner, 0)
	}
	size := len(owners)
	if size >= 0xFFFFFFFF {
		//FIXME currently the limit is for all the keys, including the
		//      revoked ones.
		return 0, errors.New("reach the max limit, cannot add more keys")
	}
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
	if index > uint32(len(owners)) {
		return nil, nil
	}
	return owners[index-1], nil
}

func findPk(srvc *native.NativeService, encID, pub []byte) (uint32, error) {
	key := append(encID, FIELD_PK)
	owners, err := getAllPk(srvc, key)
	if err != nil {
		return 0, err
	}
	for i, v := range owners {
		if bytes.Equal(pub, v.key) {
			return uint32(i + 1), nil
		}
	}
	return 0, nil
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
			v.revoked = true
			index = uint32(i)
		}
	}
	if index == 0 {
		return 0, errors.New("revoke failed, public key not found")
	}
	err = putAllPk(srvc, key, owners)
	if err != nil {
		return 0, err
	}
	return index, nil
}

func isOwner(srvc *native.NativeService, encID, pub []byte) bool {
	kID, err := findPk(srvc, encID, pub)
	if err != nil {
		log.Debug(err)
		return false
	}
	if kID == 0 {
		return false
	}
	return true
}
