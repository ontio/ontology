package ontid

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native"
)

func regIdWithPublicKey(srvc *native.NativeService) error {
	log.Debug("registerIdWithPublicKey")
	log.Debug("srvc.Input:", srvc.Input)
	// parse arguments
	args := bytes.NewBuffer(srvc.Input)
	// arg0: ID
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		log.Error(err)
		return errors.New("register ONT ID error: parsing argument 0 failed")
	}
	// arg1: public key
	arg1, err := serialization.ReadVarBytes(args)
	if err != nil {
		log.Error(err)
		return errors.New("register ONT ID error: parsing argument 1 failed")
	}

	log.Debug("arg 0:", hex.EncodeToString(arg0), string(arg0))
	log.Debug("arg 1:", hex.EncodeToString(arg1))

	if len(arg0) == 0 || len(arg1) == 0 {
		return errors.New("register ONT ID error: invalid argument")
	}

	key, err := encodeID(arg0)
	if err != nil {
		return errors.New("register ONT ID error: " + err.Error())
	}

	if checkIDExistence(srvc, key) {
		return errors.New("register ONT ID error: already registered")
	}

	public, err := keypair.DeserializePublicKey(arg1)
	if err != nil {
		log.Error(err)
		return errors.New("register ONT ID error: invalid public key")
	}
	addr := types.AddressFromPubKey(public)
	if !srvc.ContextRef.CheckWitness(addr) {
		return errors.New("register ONT ID error: checking witness failed")
	}

	// insert public key
	err = insertPk(srvc, key, arg1)
	if err != nil {
		return errors.New("register ONT ID error: store public key error, " + err.Error())
	}
	// set flags
	srvc.CloneCache.Add(common.ST_STORAGE, key, &states.StorageItem{Value: []byte{flag_exist}})

	triggerRegisterEvent(srvc, arg0)

	return nil
}

func regIdWithAttributes(srvc *native.NativeService) error {
	// parse arguments
	args := bytes.NewBuffer(srvc.Input)
	// arg0: ID
	arg0, err := serialization.ReadVarBytes(args)
	if len(arg0) == 0 {
		return errors.New("register ID with attributes error: argument 0 error")
	}
	// arg1: public key
	arg1, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("register ID with attributes error: argument 1 error, " + err.Error())
	}
	// arg2: attributes
	arg2, err := serialization.ReadVarBytes(args)
	if len(arg2) < 2 {
		return errors.New("register ID with attributes error: argument 2 error, " + err.Error())
	}

	key, err := encodeID(arg0)
	if err != nil {
		return errors.New("register ID with attributes error: " + err.Error())
	}

	if checkIDExistence(srvc, key) {
		return errors.New("register ID with attributes error: already registered")
	}
	public, err := keypair.DeserializePublicKey(arg1)
	if err != nil {
		return errors.New("register ID with attributes error: invalid public key: " + err.Error())
	}
	addr := types.AddressFromPubKey(public)
	if !srvc.ContextRef.CheckWitness(addr) {
		return errors.New("register ID with attributes error: check witness failed")
	}

	err = insertPk(srvc, key, arg1)
	if err != nil {
		return errors.New("register ID with attributes error: store pubic key error: " + err.Error())
	}

	// parse attributes
	buf := bytes.NewBuffer(arg2)
	attr := make([]*attribute, 0)
	for buf.Len() > 0 {
		buf1, err := serialization.ReadVarBytes(buf)
		t := new(attribute)
		err = t.Deserialize(bytes.NewBuffer(buf1))
		if err != nil {
			return errors.New("register ID with attributes error: parse attribute error, " + err.Error())
		}
		attr = append(attr, t)
	}
	for _, v := range attr {
		err = insertOrUpdateAttr(srvc, key, v)
		if err != nil {
			return errors.New("register ID with attributes error: store attributes error, " + err.Error())
		}
	}

	srvc.CloneCache.Add(common.ST_STORAGE, key, &states.StorageItem{Value: []byte{flag_exist}})

	triggerRegisterEvent(srvc, arg0)

	return nil
}

func addKey(srvc *native.NativeService) error {
	log.Debug("ID contract: AddKey")
	args := bytes.NewBuffer(srvc.Input)
	// arg0: id
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("add key failed: argument 0 error, " + err.Error())
	}
	log.Debug("arg 0:", hex.EncodeToString(arg0))

	// arg1: public key
	arg1, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("add key failed: argument 1 error, " + err.Error())
	}
	log.Debug("arg 1:", hex.EncodeToString(arg1))

	// arg2: operator's public key / address
	arg2, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("add key failed: argument 2 error, " + err.Error())
	}
	log.Debug("arg 2:", hex.EncodeToString(arg2))

	if err = checkWitness(srvc, arg2); err != nil {
		return errors.New("add key failed: check witness failed, " + err.Error())
	}

	key, err := encodeID(arg0)
	if err != nil {
		return errors.New("add key failed: " + err.Error())
	}
	if !checkIDExistence(srvc, key) {
		return errors.New("add key failed: ID not registered")
	}
	if !isOwner(srvc, key, arg2) {
		return errors.New("add key failed: operator has no authorization")
	}

	item, err := findPk(srvc, key, arg1)
	if item != nil {
		return errors.New("add key failed: already exists")
	}

	err = insertPk(srvc, key, arg1)
	if err != nil {
		return errors.New("add key failed: insert public key error, " + err.Error())
	}

	triggerPublicEvent(srvc, "add", arg0, arg1)

	return nil
}

func removeKey(srvc *native.NativeService) error {
	args := bytes.NewBuffer(srvc.Input)
	// arg0: id
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("remove key failed: argument 0 error, " + err.Error())
	}

	// arg1: public key
	arg1, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("remove key failed: argument 1 error, " + err.Error())
	}

	// arg2: operator's public key / address
	arg2, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("remove key failed: argument 2 error, " + err.Error())
	}
	if err = checkWitness(srvc, arg2); err != nil {
		return errors.New("remove key failed: check witness failed, " + err.Error())
	}

	key, err := encodeID(arg0)
	if err != nil {
		return errors.New("remove key failed: " + err.Error())
	}
	if !checkIDExistence(srvc, key) {
		return errors.New("remove key failed: ID not registered")
	}
	if !isOwner(srvc, key, arg1) {
		return errors.New("remove key failed: operator has no authorization")
	}

	key1, err := findPk(srvc, key, arg1)
	if err != nil {
		return errors.New("remove key failed: cannot find the key, " + err.Error())
	}
	ok, err := native.LinkedlistDelete(srvc, key, key1)
	if err != nil {
		return errors.New("remove key failed: delete error, " + err.Error())
	} else if !ok {
		return errors.New("remove key failed: key not found")
	}

	triggerPublicEvent(srvc, "remove", arg0, arg1)

	return nil
}

func addRecovery(srvc *native.NativeService) error {
	args := bytes.NewBuffer(srvc.Input)
	// arg0: ID
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("add recovery failed: argument 0 error")
	}
	// arg1: recovery address
	arg1, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("add recovery failed: argument 1 error")
	}
	// arg2: operator's public key
	arg2, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("add recovery failed: argument 2 error")
	}

	err = checkWitness(srvc, arg2)
	if err != nil {
		return errors.New("add recovery failed: " + err.Error())
	}

	key, err := encodeID(arg0)
	if err != nil {
		return errors.New("add recovery failed: " + err.Error())
	}
	if !checkIDExistence(srvc, key) {
		return errors.New("add recovery failed: ID not registered")
	}

	if !isOwner(srvc, key, arg2) {
		return errors.New("add recovery failed: not authorized")
	}

	re, err := getRecovery(srvc, key)
	if err != nil && len(re) > 0 {
		return errors.New("add recovery failed: already set recovery")
	}

	err = setRecovery(srvc, key, arg1)
	if err != nil {
		return errors.New("add recovery failed: " + err.Error())
	}

	triggerRecoveryEvent(srvc, "add", arg0, arg1)

	return nil
}

func changeRecovery(srvc *native.NativeService) error {
	args := bytes.NewBuffer(srvc.Input)
	// arg0: ID
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("change recovery failed: argument 0 error")
	}
	// arg1: new recovery address
	arg1, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("change recovery failed: argument 1 error")
	}
	// arg2: operator's public key, who should be the old recovery
	arg2, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("change recovery failed: argument 2 error")
	}

	key, err := encodeID(arg0)
	if err != nil {
		return errors.New("change recovery failed: " + err.Error())
	}
	err = checkWitness(srvc, arg2)
	if err != nil {
		return errors.New("change recovery failed: " + err.Error())
	}
	if !checkIDExistence(srvc, key) {
		return errors.New("change recovery failed: ID not registered")
	}
	re, err := getRecovery(srvc, key)
	if err != nil {
		return errors.New("change recovery failed: recovery not set")
	}
	if !bytes.Equal(re, arg2) {
		return errors.New("change recovery failed: operator is not the recovery")
	}
	err = setRecovery(srvc, key, arg1)
	if err != nil {
		return errors.New("change recovery failed: " + err.Error())
	}

	triggerRecoveryEvent(srvc, "change", arg0, arg1)
	return nil
}

func addAttribute(srvc *native.NativeService) error {
	args := bytes.NewBuffer(srvc.Input)
	// arg0: ID
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("add attribute failed: argument 0 error")
	}
	// arg1: path
	arg1, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("add attribute failed: argument 1 error")
	}
	// arg2: type
	arg2, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("add attribute failed: argument 2 error")
	}
	// arg3: value
	arg3, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("add attribute failed: argument 3 error")
	}
	// arg4: operator's public key
	arg4, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("add attribute failed: argument 4 error")
	}

	err = checkWitness(srvc, arg4)
	if err != nil {
		return errors.New("add attribute failed: " + err.Error())
	}
	key, err := encodeID(arg0)
	if err != nil {
		return errors.New("add attribute failed: " + err.Error())
	}
	if !checkIDExistence(srvc, key) {
		return errors.New("add attribute failed: ID not registered")
	}
	if !isOwner(srvc, key, arg4) {
		return errors.New("add attribute failed: no authorization")
	}

	attr := &attribute{key: arg1, valueType: arg2, value: arg3}

	node, err := findAttr(srvc, key, arg1)
	if node != nil {
		err = insertOrUpdateAttr(srvc, key, attr)
		if err != nil {
			return errors.New("add attribute failed: update attribute error, " + err.Error())
		}
		triggerAttributeEvent(srvc, "update", arg0, arg1)
	} else {
		err = insertOrUpdateAttr(srvc, key, attr)
		if err != nil {
			return errors.New("add attribute failed: " + err.Error())
		}

		triggerAttributeEvent(srvc, "add", arg0, arg1)
	}
	return nil
}

func removeAttribute(srvc *native.NativeService) error {
	args := bytes.NewBuffer(srvc.Input)
	// arg0: ID
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("remove attribute failed: argument 0 error")
	}
	// arg1: path
	arg1, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("remove attribute failed: argument 1 error")
	}
	// arg2: operator's public key
	arg2, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("remove attribute failed: argument 2 error")
	}

	err = checkWitness(srvc, arg2)
	if err != nil {
		return errors.New("remove attribute failed: " + err.Error())
	}
	key, err := encodeID(arg0)
	if err != nil {
		return errors.New("remove attribute failed: " + err.Error())
	}
	if !checkIDExistence(srvc, key) {
		return errors.New("remove attribute failed: ID not registered")
	}
	if !isOwner(srvc, key, arg2) {
		return errors.New("remove attribute failed: no authorization")
	}

	key1 := append(key, field_attr)
	ok, err := native.LinkedlistDelete(srvc, key1, arg1)
	if err != nil {
		return errors.New("remove attribute failed: delete error, " + err.Error())
	} else if !ok {
		return errors.New("remove attribute failed: attribute not exist")
	}

	triggerAttributeEvent(srvc, "remove", arg0, arg1)
	return nil
}

func verifySignature(srvc *native.NativeService) error {
	args := bytes.NewBuffer(srvc.Input)
	// arg0: ID
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("verify signature error: argument 0 error, " + err.Error())
	}
	// arg1: index of public key
	arg1, err := serialization.ReadUint32(args)
	if err != nil {
		return errors.New("verify signature error: argument 1 error, " + err.Error())
	}
	// arg2: message
	arg2, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("verify signature error: argument 2 error, " + err.Error())
	}
	// arg3: signature
	arg3, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("verify signature error: argument 3 error, " + err.Error())
	}

	key, err := encodeID(arg0)
	if err != nil {
		return errors.New("verify signature error: " + err.Error())
	}

	key1 := append(key, field_pk)
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], arg1)
	key1 = append(key, buf[:]...)

	val, err := srvc.CloneCache.Get(common.ST_STORAGE, key1)
	if err != nil {
		return errors.New("verify signature error: get key failed, " + err.Error())
	}

	item, ok := val.(*states.StorageItem)
	if !ok {
		return errors.New("verify signature error: invalid storage item")
	}

	var pk publicKey
	pk.SetBytes(item.Value)
	if err != nil {
		return errors.New("verify signature error: parse key error, " + err.Error())
	}

	pub, err := keypair.DeserializePublicKey(pk.key)
	if err != nil {
		return errors.New("verify signature error: deserialize public key error, " + err.Error())
	}

	sig, err := signature.Deserialize(arg3)
	if err != nil {
		return errors.New("verify signature error: deserialize signature error, " + err.Error())
	}
	if !signature.Verify(pub, arg2, sig) {
		return errors.New("verification failed")
	}

	return nil
}
