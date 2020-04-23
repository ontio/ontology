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

const (
	OWNER_TOTAL_SIZE       = 1024 * 1024 // 1MB
	NEW_OWNER_BLOCK_HEIGHT = 1000000

	ALL_ACCESS  = "all"
	CRUD_ACCESS = "crud"
	USE_ACCESS  = "use"

	ONLY_PUBLICKEY      = 0
	ONLY_AUTHENTICATION = 1
	BOTH                = 2
)

type publicKey struct {
	key            []byte
	revoked        bool
	controller     []byte
	access         string
	authentication uint8
}

func (this *publicKey) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.key)
	sink.WriteBool(this.revoked)
	sink.WriteVarBytes(this.controller)
	sink.WriteString(this.access)
	sink.WriteByte(this.authentication)
}

func (this *publicKey) Deserialization(source *common.ZeroCopySource) error {
	key, err := utils.DecodeVarBytes(source)
	if err != nil {
		return err
	}
	revoked, err := utils.DecodeBool(source)
	if err != nil {
		return err
	}
	controller, err := utils.DecodeVarBytes(source)
	if err != nil {
		return err
	}
	access, err := utils.DecodeString(source)
	if err != nil {
		return err
	}
	authentication, eof := source.NextByte()
	if eof {
		return fmt.Errorf("deserilize authentication error: eof")
	}

	this.key = key
	this.revoked = revoked
	this.controller = controller
	this.access = access
	this.authentication = authentication
	return nil
}

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

func getAllPk_Version1(srvc *native.NativeService, encID, key []byte) ([]*publicKey, error) {
	val, err := utils.GetStorageItem(srvc, key)
	if err != nil {
		return nil, fmt.Errorf("get storage error, %s", err)
	}
	if val == nil {
		return nil, nil
	}

	ontid, err := decodeID(encID)
	if err != nil {
		return nil, fmt.Errorf("decodeID error, %s", err)
	}
	source := common.NewZeroCopySource(val.Value)
	publicKeys := make([]*publicKey, 0)
	switch val.StateVersion {
	case _VERSION_0:
		for source.Len() > 0 {
			var t = new(owner)
			var p = new(publicKey)
			err = t.Deserialization(source)
			if err != nil {
				return nil, fmt.Errorf("deserialize owners error, %s", err)
			}
			p.key = t.key
			p.controller = ontid
			p.access = ALL_ACCESS
			p.revoked = t.revoked
			publicKeys = append(publicKeys, p)
		}
	case _VERSION_1:
		for source.Len() > 0 {
			var t = new(publicKey)
			err = t.Deserialization(source)
			if err != nil {
				return nil, fmt.Errorf("deserialize owners error, %s", err)
			}
			publicKeys = append(publicKeys, t)
		}
	}

	return publicKeys, nil
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

func putAllPk_Version1(srvc *native.NativeService, key []byte, val []*publicKey) error {
	sink := common.NewZeroCopySink(nil)
	for _, i := range val {
		i.Serialization(sink)
	}
	var v states.StorageItem
	v.Value = sink.Bytes()
	v.StateVersion = _VERSION_1
	if len(v.Value) > OWNER_TOTAL_SIZE {
		return errors.New("total key size is out of range")
	}
	srvc.CacheDB.Put(key, v.ToArray())
	return nil
}

func insertPk(srvc *native.NativeService, encID, pk, controller []byte, access string, authentication uint8,
	proof []byte) (uint32, error) {
	key := append(encID, FIELD_PK)
	if srvc.Height < NEW_OWNER_BLOCK_HEIGHT {
		owners, err := getAllPk(srvc, key)
		if err != nil {
			return 0, err
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
	} else {
		var err error
		publicKeys, err := getAllPk_Version1(srvc, encID, key)
		if err != nil {
			return 0, err
		}
		for _, k := range publicKeys {
			if bytes.Equal(k.key, pk) {
				return 0, errors.New("the key is already added")
			}
		}
		size := len(publicKeys)
		a, err := validateAccess(access)
		if err != nil {
			return 0, err
		}
		publicKeys = append(publicKeys, &publicKey{pk, false, controller, a, authentication})
		err = putAllPk_Version1(srvc, key, publicKeys)
		if err != nil {
			return 0, err
		}

		//:TODO update proof
		return uint32(size + 1), nil
	}
}

func changePkAuthentication(srvc *native.NativeService, encID []byte, index uint32, authentication uint8, proof []byte) error {
	key := append(encID, FIELD_PK)

	publicKeys, err := getAllPk_Version1(srvc, encID, key)
	if err != nil {
		return err
	}
	publicKeys[index].authentication = authentication
	err = putAllPk_Version1(srvc, key, publicKeys)
	if err != nil {
		return err
	}

	//:TODO update proof
	return nil
}

func getPk(srvc *native.NativeService, encID []byte, index uint32) (*publicKey, error) {
	key := append(encID, FIELD_PK)
	publicKeys, err := getAllPk_Version1(srvc, encID, key)
	if err != nil {
		return nil, err
	}
	if index < 1 || index > uint32(len(publicKeys)) {
		return nil, errors.New("invalid key index")
	}
	return publicKeys[index-1], nil
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

func findPk_Version1(srvc *native.NativeService, encID, pub []byte) (uint32, bool, error) {
	key := append(encID, FIELD_PK)
	publicKeys, err := getAllPk_Version1(srvc, encID, key)
	if err != nil {
		return 0, false, err
	}
	for i, v := range publicKeys {
		if bytes.Equal(pub, v.key) && v.access != USE_ACCESS {
			return uint32(i + 1), v.revoked, nil
		}
	}
	return 0, false, nil
}

func revokeAuthKey(srvc *native.NativeService, encID []byte, index uint32, proof []byte) error {
	key := append(encID, FIELD_PK)

	publicKeys, err := getAllPk_Version1(srvc, encID, key)
	if err != nil {
		return err
	}
	publicKeys[index].revoked = true
	err = putAllPk_Version1(srvc, key, publicKeys)
	if err != nil {
		return err
	}

	//:TODO update proof
	return nil
}

func revokePk(srvc *native.NativeService, encID, pub, proof []byte) (uint32, error) {
	key := append(encID, FIELD_PK)
	if srvc.Height < NEW_OWNER_BLOCK_HEIGHT {
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
	} else {
		publicKeys, err := getAllPk_Version1(srvc, encID, key)
		if err != nil {
			return 0, err
		}
		var index uint32 = 0
		for i, v := range publicKeys {
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
		putAllPk_Version1(srvc, key, publicKeys)

		//:TODO update proof
		return index, nil
	}
}

func revokePkByIndex(srvc *native.NativeService, encID []byte, index uint32, proof []byte) ([]byte, error) {
	key := append(encID, FIELD_PK)
	if srvc.Height < NEW_OWNER_BLOCK_HEIGHT {
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
	} else {
		publicKeys, err := getAllPk_Version1(srvc, encID, key)
		if err != nil {
			return nil, err
		}
		if uint32(len(publicKeys)) < index {
			return nil, errors.New("no such key")
		}
		index -= 1
		if publicKeys[index].revoked {
			return nil, errors.New("already revoked")
		}
		publicKeys[index].revoked = true
		putAllPk_Version1(srvc, key, publicKeys)

		//:TODO update proof
		return publicKeys[index].key, nil
	}
}

func setKeyAccessByIndex(srvc *native.NativeService, encID []byte, index uint32, access string, proof []byte) ([]byte, error) {
	key := append(encID, FIELD_PK)
	publicKeys, err := getAllPk_Version1(srvc, encID, key)
	if err != nil {
		return nil, err
	}
	if uint32(len(publicKeys)) < index {
		return nil, errors.New("no such key")
	}
	index -= 1
	if publicKeys[index].revoked {
		return nil, errors.New("already revoked")
	}
	a, err := validateAccess(access)
	if err != nil {
		return nil, err
	}
	publicKeys[index].access = a
	putAllPk_Version1(srvc, key, publicKeys)

	//:TODO update proof
	return publicKeys[index].key, nil
}

func isOwner(srvc *native.NativeService, encID, pub []byte) bool {
	kID, revoked, err := findPk_Version1(srvc, encID, pub)
	if err != nil {
		log.Debug(err)
		return false
	}
	return kID != 0 && !revoked
}

func validateAccess(access string) (string, error) {
	switch access {
	case "":
		return ALL_ACCESS, nil
	case ALL_ACCESS:
		return ALL_ACCESS, nil
	case CRUD_ACCESS:
		return CRUD_ACCESS, nil
	case USE_ACCESS:
		return USE_ACCESS, nil
	default:
		return "", fmt.Errorf("access type is not supported")
	}
}
