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

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	OWNER_TOTAL_SIZE = 1024 * 1024 // 1MB
)

type publicKeyJson struct {
	Id           string `json:"id"`
	Type         string `json:"type"`
	Controller   string `json:"controller"`
	PublicKeyHex string `json:"publicKeyHex"`
}

type publicKey struct {
	key              []byte
	revoked          bool
	controller       []byte
	isPkList         bool
	isAuthentication bool
}

func (this *publicKey) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.key)
	sink.WriteBool(this.revoked)
	sink.WriteVarBytes(this.controller)
	sink.WriteBool(this.isPkList)
	sink.WriteBool(this.isAuthentication)
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
	isPkList, err := utils.DecodeBool(source)
	if err != nil {
		return err
	}
	isAuthentication, err := utils.DecodeBool(source)
	if err != nil {
		return err
	}

	this.key = key
	this.revoked = revoked
	this.controller = controller
	this.isPkList = isPkList
	this.isAuthentication = isAuthentication
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

func getAllPk_Version1(srvc *native.NativeService, encId, key []byte) ([]*publicKey, error) {
	val, err := utils.GetStorageItem(srvc, key)
	if err != nil {
		return nil, fmt.Errorf("get storage error, %s", err)
	}
	if val == nil {
		return nil, nil
	}

	ontid, err := decodeID(encId)
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
			p.isPkList = true
			p.isAuthentication = true
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

func getAllPkJson(srvc *native.NativeService, encId []byte) ([]*publicKeyJson, error) {
	key := append(encId, FIELD_PK)
	publicKeys, err := getAllPk_Version1(srvc, encId, key)
	if err != nil {
		return nil, err
	}
	r := make([]*publicKeyJson, 0)
	for index, p := range publicKeys {
		if !p.revoked {
			publicKey := new(publicKeyJson)

			ontId, err := decodeID(encId)
			if err != nil {
				return nil, err
			}
			publicKey.Id = fmt.Sprintf("%s#keys-%d", string(ontId), index+1)
			publicKey.Controller = string(p.controller)
			publicKey.Type, publicKey.PublicKeyHex, err = keyType(p.key)
			if err != nil {
				return nil, err
			}
			r = append(r, publicKey)
		}
	}
	return r, nil
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

func insertPk(srvc *native.NativeService, encId, pk, controller []byte, isPkList, isAuthentication bool) (uint32, error) {
	key := append(encId, FIELD_PK)
	if srvc.Height < config.GetNewOntIdHeight() {
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
	} else {
		var err error
		publicKeys, err := getAllPk_Version1(srvc, encId, key)
		if err != nil {
			return 0, err
		}
		for _, k := range publicKeys {
			if bytes.Equal(k.key, pk) {
				return 0, errors.New("the key is already added")
			}
		}
		size := len(publicKeys)
		publicKeys = append(publicKeys, &publicKey{pk, false, controller, isPkList, isAuthentication})
		err = putAllPk_Version1(srvc, key, publicKeys)
		if err != nil {
			return 0, err
		}
		return uint32(size + 1), nil
	}
}

func changePkAuthentication(srvc *native.NativeService, encId []byte, index uint32, isAuthentication bool) error {
	key := append(encId, FIELD_PK)

	publicKeys, err := getAllPk_Version1(srvc, encId, key)
	if err != nil {
		return err
	}
	if index < 1 || index > uint32(len(publicKeys)) {
		return errors.New("invalid key index")
	}
	if publicKeys[index-1].revoked {
		return errors.New("key already revoked")
	}
	publicKeys[index-1].isAuthentication = isAuthentication
	err = putAllPk_Version1(srvc, key, publicKeys)
	if err != nil {
		return err
	}
	return nil
}

func getPk(srvc *native.NativeService, encId []byte, index uint32) (*publicKey, error) {
	key := append(encId, FIELD_PK)
	publicKeys, err := getAllPk_Version1(srvc, encId, key)
	if err != nil {
		return nil, err
	}
	if len(publicKeys) == 0 {
		return nil, fmt.Errorf("no record")
	}
	if index < 1 || index > uint32(len(publicKeys)) {
		return nil, errors.New("invalid key index")
	}
	return publicKeys[index-1], nil
}

func findPk(srvc *native.NativeService, encId, pub []byte) (uint32, bool, error) {
	key := append(encId, FIELD_PK)
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

func findPk_Version1(srvc *native.NativeService, encId, pub []byte) (uint32, bool, error) {
	key := append(encId, FIELD_PK)
	publicKeys, err := getAllPk_Version1(srvc, encId, key)
	if err != nil {
		return 0, false, err
	}
	for i, v := range publicKeys {
		if bytes.Equal(pub, v.key) && v.isAuthentication {
			return uint32(i + 1), v.revoked, nil
		}
	}
	return 0, false, nil
}

func revokePk(srvc *native.NativeService, encId, pub []byte) (uint32, error) {
	key := append(encId, FIELD_PK)
	if srvc.Height < config.GetNewOntIdHeight() {
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
		err = putAllPk(srvc, key, owners)
		if err != nil {
			return 0, err
		}
		return index, nil
	} else {
		publicKeys, err := getAllPk_Version1(srvc, encId, key)
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
		err = putAllPk_Version1(srvc, key, publicKeys)
		if err != nil {
			return 0, err
		}
		return index, nil
	}
}

func revokePkByIndex(srvc *native.NativeService, encId []byte, index uint32) ([]byte, error) {
	key := append(encId, FIELD_PK)
	if srvc.Height < config.GetNewOntIdHeight() {
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
		err = putAllPk(srvc, key, owners)
		if err != nil {
			return nil, err
		}
		return owners[index].key, nil
	} else {
		publicKeys, err := getAllPk_Version1(srvc, encId, key)
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
		err = putAllPk_Version1(srvc, key, publicKeys)
		if err != nil {
			return nil, err
		}
		return publicKeys[index].key, nil
	}
}

func isOwner(srvc *native.NativeService, encId, pub []byte) bool {
	kID, revoked, err := findPk_Version1(srvc, encId, pub)
	if err != nil {
		log.Debug(err)
		return false
	}
	return kID != 0 && !revoked
}

func keyType(publicKey []byte) (string, string, error) {
	switch keypair.KeyType(publicKey[0]) {
	case keypair.PK_P256_E, keypair.PK_P256_O, keypair.PK_P256_NC:
		return "EcdsaSecp256r1VerificationKey2019", hex.EncodeToString(publicKey), nil
	case keypair.PK_ECDSA:
		switch publicKey[1] {
		case keypair.P224:
			return "EcdsaSecp224r1VerificationKey2019", hex.EncodeToString(publicKey[2:]), nil
		case keypair.P256:
			return "EcdsaSecp256r1VerificationKey2019", hex.EncodeToString(publicKey[2:]), nil
		case keypair.P384:
			return "EcdsaSecp384r1VerificationKey2019", hex.EncodeToString(publicKey[2:]), nil
		case keypair.P521:
			return "EcdsaSecp521r1VerificationKey2019", hex.EncodeToString(publicKey[2:]), nil
		case keypair.SECP256K1:
			return "EcdsaSecp256k1VerificationKey2019", hex.EncodeToString(publicKey[2:]), nil
		default:
			return "", "", fmt.Errorf("unsupported type 1")
		}
	case keypair.PK_EDDSA:
		return "Ed25519VerificationKey2018", hex.EncodeToString(publicKey[2:]), nil
	case keypair.PK_SM2:
		return "SM2VerificationKey2019", hex.EncodeToString(publicKey[2:]), nil
	default:
		return "", "", fmt.Errorf("unsupported type 2")
	}
}
