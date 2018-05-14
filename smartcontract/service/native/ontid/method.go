package ontid

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func regIdWithPublicKey(srvc *native.NativeService) ([]byte, error) {
	log.Debug("registerIdWithPublicKey")
	log.Debug("srvc.Input:", srvc.Input)
	// parse arguments
	args := bytes.NewBuffer(srvc.Input)
	// arg0: ID
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("register ONT ID error: parsing argument 0 failed")
	}
	// arg1: public key
	arg1, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("register ONT ID error: parsing argument 1 failed")
	}

	log.Debug("arg 0:", hex.EncodeToString(arg0), string(arg0))
	log.Debug("arg 1:", hex.EncodeToString(arg1))

	if len(arg0) == 0 || len(arg1) == 0 {
		return utils.BYTE_FALSE, errors.New("register ONT ID error: invalid argument")
	}

	key, err := encodeID(arg0)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("register ONT ID error: " + err.Error())
	}

	if checkIDExistence(srvc, key) {
		return utils.BYTE_FALSE, errors.New("register ONT ID error: already registered")
	}

	public, err := keypair.DeserializePublicKey(arg1)
	if err != nil {
		log.Error(err)
		return utils.BYTE_FALSE, errors.New("register ONT ID error: invalid public key")
	}
	addr := types.AddressFromPubKey(public)
	if !srvc.ContextRef.CheckWitness(addr) {
		return utils.BYTE_FALSE, errors.New("register ONT ID error: checking witness failed")
	}

	// insert public key
	_, err = insertPk(srvc, key, arg1)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("register ONT ID error: store public key error, " + err.Error())
	}
	// set flags
	srvc.CloneCache.Add(common.ST_STORAGE, key, &states.StorageItem{Value: []byte{flag_exist}})

	triggerRegisterEvent(srvc, arg0)

	return utils.BYTE_TRUE, nil
}

func regIdWithAttributes(srvc *native.NativeService) ([]byte, error) {
	// parse arguments
	args := bytes.NewBuffer(srvc.Input)
	// arg0: ID
	arg0, err := serialization.ReadVarBytes(args)
	if len(arg0) == 0 {
		return utils.BYTE_FALSE, errors.New("register ID with attributes error: argument 0 error, " + err.Error())
	}
	// arg1: public key
	arg1, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("register ID with attributes error: argument 1 error, " + err.Error())
	}
	// arg2: attributes
	arg2, err := serialization.ReadVarBytes(args)
	if len(arg2) < 2 {
		return utils.BYTE_FALSE, errors.New("register ID with attributes error: argument 2 error, " + err.Error())
	}

	key, err := encodeID(arg0)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("register ID with attributes error: " + err.Error())
	}

	if checkIDExistence(srvc, key) {
		return utils.BYTE_FALSE, errors.New("register ID with attributes error: already registered")
	}
	public, err := keypair.DeserializePublicKey(arg1)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("register ID with attributes error: invalid public key: " + err.Error())
	}
	addr := types.AddressFromPubKey(public)
	if !srvc.ContextRef.CheckWitness(addr) {
		return utils.BYTE_FALSE, errors.New("register ID with attributes error: check witness failed")
	}

	_, err = insertPk(srvc, key, arg1)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("register ID with attributes error: store pubic key error: " + err.Error())
	}

	// parse attributes
	buf := bytes.NewBuffer(arg2)
	attr := make([]*attribute, 0)
	for buf.Len() > 0 {
		t := new(attribute)
		err = t.Deserialize(buf)
		if err != nil {
			return utils.BYTE_FALSE, errors.New("register ID with attributes error: parse attribute error, " + err.Error())
		}
		attr = append(attr, t)
	}
	for _, v := range attr {
		err = insertOrUpdateAttr(srvc, key, v)
		if err != nil {
			return utils.BYTE_FALSE, errors.New("register ID with attributes error: store attributes error, " + err.Error())
		}
	}

	srvc.CloneCache.Add(common.ST_STORAGE, key, &states.StorageItem{Value: []byte{flag_exist}})

	triggerRegisterEvent(srvc, arg0)

	return utils.BYTE_TRUE, nil
}

func addKey(srvc *native.NativeService) ([]byte, error) {
	log.Debug("ID contract: AddKey")
	args := bytes.NewBuffer(srvc.Input)
	// arg0: id
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add key failed: argument 0 error, " + err.Error())
	}
	log.Debug("arg 0:", hex.EncodeToString(arg0))

	// arg1: public key
	arg1, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add key failed: argument 1 error, " + err.Error())
	}
	log.Debug("arg 1:", hex.EncodeToString(arg1))

	// arg2: operator's public key / address
	arg2, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add key failed: argument 2 error, " + err.Error())
	}
	log.Debug("arg 2:", hex.EncodeToString(arg2))

	if err = checkWitness(srvc, arg2); err != nil {
		return utils.BYTE_FALSE, errors.New("add key failed: check witness failed, " + err.Error())
	}

	key, err := encodeID(arg0)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add key failed: " + err.Error())
	}
	if !checkIDExistence(srvc, key) {
		return utils.BYTE_FALSE, errors.New("add key failed: ID not registered")
	}
	if !isOwner(srvc, key, arg2) {
		return utils.BYTE_FALSE, errors.New("add key failed: operator has no authorization")
	}

	item, err := findPk(srvc, key, arg1)
	if item != nil {
		return utils.BYTE_FALSE, errors.New("add key failed: already exists")
	}

	keyID, err := insertPk(srvc, key, arg1)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add key failed: insert public key error, " + err.Error())
	}

	triggerPublicEvent(srvc, "add", arg0, arg1, keyID)

	return utils.BYTE_TRUE, nil
}

func removeKey(srvc *native.NativeService) ([]byte, error) {
	args := bytes.NewBuffer(srvc.Input)
	// arg0: id
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remove key failed: argument 0 error, %s", err)
	}

	// arg1: public key
	arg1, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remove key failed: argument 1 error, %s", err)
	}

	// arg2: operator's public key / address
	arg2, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remove key failed: argument 2 error, %s", err)
	}
	if err = checkWitness(srvc, arg2); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remove key failed: check witness failed, %S", err)
	}

	key, err := encodeID(arg0)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remove key failed: %s", err)
	}
	if !checkIDExistence(srvc, key) {
		return utils.BYTE_FALSE, errors.New("remove key failed: ID not registered")
	}
	var auth bool = false
	rec, err := getRecovery(srvc, key)
	if len(rec) > 0 {
		auth = bytes.Equal(rec, arg2)
	}
	if !auth {
		if !isOwner(srvc, key, arg2) {
			return utils.BYTE_FALSE, errors.New("remove key failed: operator has no authorization")
		}
	}

	keyID, err := revokePk(srvc, key, arg1)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remove key failed: %s", err)
	}

	triggerPublicEvent(srvc, "remove", arg0, arg1, keyID)

	return utils.BYTE_TRUE, nil
}

func addRecovery(srvc *native.NativeService) ([]byte, error) {
	args := bytes.NewBuffer(srvc.Input)
	// arg0: ID
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add recovery failed: argument 0 error")
	}
	// arg1: recovery address
	arg1, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add recovery failed: argument 1 error")
	}
	// arg2: operator's public key
	arg2, err := serialization.ReadVarBytes(args)
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
	if !checkIDExistence(srvc, key) {
		return utils.BYTE_FALSE, errors.New("add recovery failed: ID not registered")
	}

	if !isOwner(srvc, key, arg2) {
		return utils.BYTE_FALSE, errors.New("add recovery failed: not authorized")
	}

	re, err := getRecovery(srvc, key)
	if err != nil && len(re) > 0 {
		return utils.BYTE_FALSE, errors.New("add recovery failed: already set recovery")
	}

	err = setRecovery(srvc, key, arg1)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add recovery failed: " + err.Error())
	}

	triggerRecoveryEvent(srvc, "add", arg0, arg1)

	return utils.BYTE_TRUE, nil
}

func changeRecovery(srvc *native.NativeService) ([]byte, error) {
	args := bytes.NewBuffer(srvc.Input)
	// arg0: ID
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("change recovery failed: argument 0 error")
	}
	// arg1: new recovery address
	arg1, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("change recovery failed: argument 1 error")
	}
	// arg2: operator's public key, who should be the old recovery
	arg2, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("change recovery failed: argument 2 error")
	}

	key, err := encodeID(arg0)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("change recovery failed: " + err.Error())
	}
	err = checkWitness(srvc, arg2)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("change recovery failed: " + err.Error())
	}
	if !checkIDExistence(srvc, key) {
		return utils.BYTE_FALSE, errors.New("change recovery failed: ID not registered")
	}
	re, err := getRecovery(srvc, key)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("change recovery failed: recovery not set")
	}
	if !bytes.Equal(re, arg2) {
		return utils.BYTE_FALSE, errors.New("change recovery failed: operator is not the recovery")
	}
	err = setRecovery(srvc, key, arg1)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("change recovery failed: " + err.Error())
	}

	triggerRecoveryEvent(srvc, "change", arg0, arg1)
	return utils.BYTE_TRUE, nil
}

func addAttribute(srvc *native.NativeService) ([]byte, error) {
	args := bytes.NewBuffer(srvc.Input)
	// arg0: ID
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add attribute failed: argument 0 error")
	}
	// arg1: path
	arg1, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add attribute failed: argument 1 error")
	}
	// arg2: type
	arg2, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add attribute failed: argument 2 error")
	}
	// arg3: value
	arg3, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add attribute failed: argument 3 error")
	}
	// arg4: operator's public key
	arg4, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add attribute failed: argument 4 error")
	}

	err = checkWitness(srvc, arg4)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add attribute failed: " + err.Error())
	}
	key, err := encodeID(arg0)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add attribute failed: " + err.Error())
	}
	if !checkIDExistence(srvc, key) {
		return utils.BYTE_FALSE, errors.New("add attribute failed: ID not registered")
	}
	if !isOwner(srvc, key, arg4) {
		return utils.BYTE_FALSE, errors.New("add attribute failed: no authorization")
	}

	attr := &attribute{key: arg1, valueType: arg2, value: arg3}

	node, err := findAttr(srvc, key, arg1)
	if node != nil {
		err = insertOrUpdateAttr(srvc, key, attr)
		if err != nil {
			return utils.BYTE_FALSE, errors.New("add attribute failed: update attribute error, " + err.Error())
		}
		triggerAttributeEvent(srvc, "update", arg0, arg1)
	} else {
		err = insertOrUpdateAttr(srvc, key, attr)
		if err != nil {
			return utils.BYTE_FALSE, errors.New("add attribute failed: " + err.Error())
		}

		triggerAttributeEvent(srvc, "add", arg0, arg1)
	}
	return utils.BYTE_TRUE, nil
}

func removeAttribute(srvc *native.NativeService) ([]byte, error) {
	args := bytes.NewBuffer(srvc.Input)
	// arg0: ID
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("remove attribute failed: argument 0 error")
	}
	// arg1: path
	arg1, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("remove attribute failed: argument 1 error")
	}
	// arg2: operator's public key
	arg2, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("remove attribute failed: argument 2 error")
	}

	err = checkWitness(srvc, arg2)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("remove attribute failed: " + err.Error())
	}
	key, err := encodeID(arg0)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("remove attribute failed: " + err.Error())
	}
	if !checkIDExistence(srvc, key) {
		return utils.BYTE_FALSE, errors.New("remove attribute failed: ID not registered")
	}
	if !isOwner(srvc, key, arg2) {
		return utils.BYTE_FALSE, errors.New("remove attribute failed: no authorization")
	}

	key1 := append(key, FIELD_ATTR)
	ok, err := utils.LinkedlistDelete(srvc, key1, arg1)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("remove attribute failed: delete error, " + err.Error())
	} else if !ok {
		return utils.BYTE_FALSE, errors.New("remove attribute failed: attribute not exist")
	}

	triggerAttributeEvent(srvc, "remove", arg0, arg1)
	return utils.BYTE_TRUE, nil
}

func verifySignature(srvc *native.NativeService) ([]byte, error) {
	args := bytes.NewBuffer(srvc.Input)
	// arg0: ID
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("verify signature error: argument 0 error, " + err.Error())
	}
	// arg1: index of public key
	arg1, err := serialization.ReadUint32(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("verify signature error: argument 1 error, " + err.Error())
	}

	key, err := encodeID(arg0)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("verify signature error: " + err.Error())
	}

	key1 := append(key, FIELD_PK)
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], arg1)

	node, err := utils.LinkedlistGetItem(srvc, key1, buf[:])
	if err != nil {
		return utils.BYTE_FALSE, errors.New("verify signature error: get key failed, " + err.Error())
	}

	err = checkWitness(srvc, node.GetPayload())
	if err != nil {
		return utils.BYTE_FALSE, errors.New("verify signature failed: " + err.Error())
	}

	return utils.BYTE_TRUE, nil
}
