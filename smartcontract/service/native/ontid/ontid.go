package ontid

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
)

var contractAddress = genesis.OntIDContractAddress[:]

func init() {
	native.Contracts[genesis.OntIDContractAddress] = RegisterIDContract
}

func RegisterIDContract(srvc *native.NativeService) {
	srvc.Register("regIDWithPublicKey", RegIdWithPublicKey)
	srvc.Register("addKey", AddKey)
	srvc.Register("removeKey", RemoveKey)
	srvc.Register("addRecovery", AddRecovery)
	srvc.Register("changeRecovery", ChangeRecovery)
	srvc.Register("addAttribute", AddAttribute)
	srvc.Register("removeAttribute", RemoveAttribute)
	srvc.Register("verifySignature", verifySignature)
	return
}

type publicKey struct {
	key     []byte
	revoked bool
}

func (this *publicKey) Serialize(w io.Writer) error {
	err := serialization.WriteVarBytes(w, this.key)
	if err != nil {
		return err
	}
	err = serialization.WriteBool(w, this.revoked)
	if err != nil {
		return err
	}
	return nil
}

func (this *publicKey) Deserialize(r io.Reader) error {
	data, err := serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}
	rev, err := serialization.ReadBool(r)
	if err != nil {
		return err
	}
	this.key = data
	this.revoked = rev
	return nil
}

func (this *publicKey) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	err := this.Serialize(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (this *publicKey) SetBytes(data []byte) error {
	buf := bytes.NewBuffer(data)
	return this.Deserialize(buf)
}

func RegIdWithPublicKey(srvc *native.NativeService) error {
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

func GetPublicKeyByID(srvc *native.NativeService) ([]byte, error) {
	args := bytes.NewBuffer(srvc.Input)
	// arg0: ID
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return nil, errors.New("get public key failed: argument 0 error")
	}
	// arg1: index
	arg1, err := serialization.ReadUint32(args)
	if err != nil {
		return nil, errors.New("get public key failed: argument 1 error")
	}

	key, err := encodeID(arg0)
	if err != nil {
		return nil, errors.New("get public key failed: " + err.Error())
	}

	pk, err := getPk(srvc, key, arg1)
	if err != nil {
		return nil, errors.New("get public key failed: " + err.Error())
	}
	if pk.revoked {
		return nil, errors.New("get public key failed: revoked")
	}

	return pk.key, nil
}

type attribute struct {
	key       []byte
	value     []byte
	valueType []byte
}

func (this *attribute) Value() ([]byte, error) {
	var buf bytes.Buffer
	err := serialization.WriteVarBytes(&buf, this.value)
	if err != nil {
		return nil, err
	}
	err = serialization.WriteVarBytes(&buf, this.valueType)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (this *attribute) SetValue(data []byte) error {
	buf := bytes.NewBuffer(data)
	val, err := serialization.ReadVarBytes(buf)
	if err != nil {
		return err
	}
	vt, err := serialization.ReadVarBytes(buf)
	if err != nil {
		return err
	}
	this.valueType = vt
	this.value = val
	return nil
}

func (this *attribute) Serialize(w io.Writer) error {
	err := serialization.WriteVarBytes(w, this.key)
	if err != nil {
		return err
	}
	err = serialization.WriteVarBytes(w, this.valueType)
	if err != nil {
		return err
	}
	err = serialization.WriteVarBytes(w, this.value)
	if err != nil {
		return err
	}
	return nil
}

func (this *attribute) Deserialize(r io.Reader) error {
	k, err := serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}
	vt, err := serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}
	v, err := serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}
	this.key = k
	this.value = v
	this.valueType = vt
	return nil
}

func RegIdWithAttributes(srvc *native.NativeService) error {
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
		t := new(attribute)
		err = t.Deserialize(buf)
		if err != nil {
			return errors.New("register ID with attributes error: parse attribute error, " + err.Error())
		}
		attr = append(attr, t)
	}
	for _, v := range attr {
		err = insertAttr(srvc, key, v)
		if err != nil {
			return errors.New("register ID with attributes error: store attributes error, " + err.Error())
		}
	}

	srvc.CloneCache.Add(common.ST_STORAGE, key, &states.StorageItem{Value: []byte{flag_exist}})

	triggerRegisterEvent(srvc, arg0)

	return nil
}

func AddKey(srvc *native.NativeService) error {
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

	// arg2: operator's public key
	arg2, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("add key failed: argument 2 error, " + err.Error())
	}
	log.Debug("arg 2:", hex.EncodeToString(arg2))

	pub, err := keypair.DeserializePublicKey(arg2)
	if err != nil {
		return errors.New("add key failed: invalid public key, " + err.Error())
	}
	addr := types.AddressFromPubKey(pub)
	if !srvc.ContextRef.CheckWitness(addr) {
		return errors.New("add key failed: check witness failed")
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

func RemoveKey(srvc *native.NativeService) error {
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

	// arg2: operator's public key
	arg2, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("remove key failed: argument 2 error, " + err.Error())
	}
	pub, err := keypair.DeserializePublicKey(arg2)
	if err != nil {
		return errors.New("remove key failed: invalid public key, " + err.Error())
	}
	addr := types.AddressFromPubKey(pub)
	if !srvc.ContextRef.CheckWitness(addr) {
		return errors.New("remove key failed: check witness failed")
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

func AddRecovery(srvc *native.NativeService) error {
	args := bytes.NewBuffer(srvc.Input)
	// arg0: ID
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("add recovery failed: argument 0 error")
	}
	// arg1: recovery
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

	return nil
}

func ChangeRecovery(srvc *native.NativeService) error {
	args := bytes.NewBuffer(srvc.Input)
	// arg0: ID
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("change recovery failed: argument 0 error")
	}
	// arg1: public key of the new recovery
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

	return nil
}

func AddAttribute(srvc *native.NativeService) error {
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
		err = updateAttr(srvc, key, attr)
		if err != nil {
			return errors.New("add attribute failed: update attribute error, " + err.Error())
		}
		triggerAttributeEvent(srvc, "update", arg0, arg1)
	} else {
		err = insertAttr(srvc, key, attr)
		if err != nil {
			return errors.New("add attribute failed: " + err.Error())
		}

		triggerAttributeEvent(srvc, "add", arg0, arg1)
	}
	return nil
}

func RemoveAttribute(srvc *native.NativeService) error {
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

type DDO struct {
	Keys []publicKey `json:"Owners,omitempty"`
}

func GetDDO(srvc *native.NativeService) ([]byte, error) {
	var0, err := GetPublicKeys(srvc)
	if err != nil {
		return nil, errors.New("get DDO error: " + err.Error())
	}
	var1, err := GetAttributes(srvc)
	if err != nil {
		return nil, errors.New("get DDO error: " + err.Error())
	}

	var buf bytes.Buffer
	serialization.WriteVarBytes(&buf, var0)
	serialization.WriteVarBytes(&buf, var1)

	return buf.Bytes(), nil
}

func GetPublicKeys(srvc *native.NativeService) ([]byte, error) {
	did := srvc.Input
	if len(did) == 0 {
		return nil, errors.New("get public keys error: invalid ID")
	}
	key := append(did, field_pk)
	item, err := native.LinkedlistGetHead(srvc, key)
	if err != nil {
		return nil, errors.New("get public keys error: cannot get the list head, " + err.Error())
	} else if len(item) == 0 {
		return nil, errors.New("get public keys error: get list head failed")
	}

	var i uint = 0
	var res bytes.Buffer
	for len(item) > 0 {
		node, err := native.LinkedlistGetItem(srvc, key, item)
		if err != nil {
			return nil, errors.New("get public keys error: " + err.Error())
		} else if node == nil {
			return nil, errors.New("get public keys error: get list node failed")
		}

		var pk publicKey
		err = pk.SetBytes(node.GetPayload())
		if err != nil {
			return nil, errors.New("get public keys error: parse key error, " + err.Error())
		}
		serialization.WriteVarBytes(&res, pk.key)
		//TODO key id?

		i += 1
		item = node.GetNext()
	}

	return append([]byte{byte(i >> 8), byte(i & 0xff)}, res.Bytes()...), nil
}

func GetAttributes(srvc *native.NativeService) ([]byte, error) {
	did := srvc.Input
	if len(did) == 0 {
		return nil, errors.New("get attributes error: invalid ID")
	}
	key := append(did, field_attr)
	item, err := native.LinkedlistGetHead(srvc, key)
	if err != nil {
		return nil, errors.New("get attributes error: get list head error, " + err.Error())
	} else if len(item) == 0 {
		return nil, errors.New("get attributes error: cannot get list head")
	}

	var res bytes.Buffer
	var i uint16 = 0
	for len(item) > 0 {
		node, err := native.LinkedlistGetItem(srvc, key, item)
		if err != nil {

		} else if node == nil {

		}

		var attr attribute
		err = attr.SetValue(node.GetPayload())
		if err != nil {
			return nil, errors.New("get attributes error: parse attribute failed, " + err.Error())
		}

		var buf bytes.Buffer
		serialization.WriteVarBytes(&buf, attr.key)
		serialization.WriteVarBytes(&buf, attr.valueType)
		serialization.WriteVarBytes(&buf, attr.value)
		serialization.WriteVarBytes(&res, buf.Bytes())

		i += 1
		item = node.GetNext()
	}

	return append([]byte{byte(i >> 8), byte(i & 0xff)}, res.Bytes()...), nil
}

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
	field_pk byte = 1 + iota
	field_attr
	field_recovery
)

func encodeID(id []byte) ([]byte, error) {
	length := len(id)
	if length == 0 || length > 255 {
		return nil, errors.New("encode ONT ID error: invalid ID length")
	}
	enc := []byte{byte(length)}
	enc = append(enc, id...)
	enc = append(contractAddress, enc...)
	return enc, nil
}

func decodeID(data []byte) ([]byte, error) {
	if len(data) == 0 || len(data) != int(data[0])+1 {
		return nil, errors.New("decode ONT ID error: invalid data length")
	}
	return data[1:], nil
}

func setNumOfField(srvc *native.NativeService, encID []byte, field byte, n uint32) error {
	key := append(encID, field)
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, n)
	val := &states.StorageItem{Value: buf}
	srvc.CloneCache.Add(common.ST_STORAGE, key, val)
	return nil
}

func getNumOfField(srvc *native.NativeService, encID []byte, field byte) (uint32, error) {
	key := append(encID, field)
	val, err := srvc.CloneCache.Get(common.ST_STORAGE, key)
	if err != nil {
		return 0, err
	}
	t, ok := val.(*states.StorageItem)
	if !ok {
		return 0, errors.New("get storage item error: invalid value type")
	}
	n := binary.LittleEndian.Uint32(t.Value)
	return n, nil
}

func insertPk(srvc *native.NativeService, encID, pk []byte) error {
	var index uint32 = 0
	key1 := append(encID, field_pk)
	item, err := native.LinkedlistGetHead(srvc, key1)
	if err == nil && item != nil {
		index = binary.LittleEndian.Uint32(item[len(key1):])
	}
	index += 1
	log.Debug("index:", index)
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], index)
	key2 := append(key1, buf[:]...)
	log.Debug("public key list id:", hex.EncodeToString(key1))
	log.Debug("public key id:", hex.EncodeToString(key2))
	p := &publicKey{key: pk, revoked: false}
	val, err := p.Bytes()
	if err != nil {
		return errors.New("register ONT ID error: " + err.Error())
	}
	return native.LinkedlistInsert(srvc, key1, key2, val)
}

func getPk(srvc *native.NativeService, encID []byte, index uint32) (*publicKey, error) {
	key1 := append(encID, field_pk)
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], index)
	key2 := append(key1, buf[:]...)
	node, err := native.LinkedlistGetItem(srvc, key1, key2)
	if err != nil {
		return nil, err
	}
	if len(node.GetPayload()) == 0 {
		return nil, errors.New("invalid public key data from storage")
	}

	var pk publicKey
	err = pk.SetBytes(node.GetPayload())
	if err != nil {
		return nil, errors.New("invalid public key data from storage")
	}

	return &pk, nil
}

func findPk(srvc *native.NativeService, encID, pub []byte) ([]byte, error) {
	key := append(encID, field_pk)
	item, err := native.LinkedlistGetHead(srvc, key)
	if err != nil {
		return nil, err
	}
	log.Debug("item:", hex.EncodeToString(item))

	for len(item) > 0 {
		node, err := native.LinkedlistGetItem(srvc, key, item)
		if err != nil {
			log.Debug(err)
			continue
		}
		var pk publicKey
		err = pk.SetBytes(node.GetPayload())
		if err != nil {
			return nil, err
		}
		log.Debug("pk.key:", hex.EncodeToString(pk.key))
		log.Debug("pub:", hex.EncodeToString(pub))
		if bytes.Equal(pub, pk.key) {
			return key, nil
		}
		item = node.GetNext()
	}

	return nil, errors.New("public key not found")
}

func insertAttr(srvc *native.NativeService, encID []byte, attr *attribute) error {
	key := append(encID, field_attr)
	val, err := attr.Value()
	if err != nil {
		return errors.New("serialize attribute value error: " + err.Error())
	}
	err = native.LinkedlistInsert(srvc, key, attr.key, val)
	if err != nil {
		return errors.New("store attribute error: " + err.Error())
	}
	return nil
}

func updateAttr(srvc *native.NativeService, encID []byte, attr *attribute) error {
	//TODO direct update instead of delete + insert
	key := append(encID, field_attr)
	ok, err := native.LinkedlistDelete(srvc, key, attr.key)
	if err != nil {
		return err
	} else if !ok {
		return errors.New("delete old attribute failed")
	}

	val, err := attr.Value()
	if err != nil {
		return errors.New("serialize attribute value error: " + err.Error())
	}
	err = native.LinkedlistInsert(srvc, key, attr.key, val)
	if err != nil {
		return errors.New("store attribute error: " + err.Error())
	}
	return nil
}

func findAttr(srvc *native.NativeService, encID, item []byte) (*native.LinkedlistNode, error) {
	key := append(encID, field_attr)
	return native.LinkedlistGetItem(srvc, key, item)
}

func setRecovery(srvc *native.NativeService, encID, recovery []byte) error {
	key := append(encID, field_recovery)
	val := &states.StorageItem{Value: recovery}
	srvc.CloneCache.Add(common.ST_STORAGE, key, val)
	return nil
}

func getRecovery(srvc *native.NativeService, encID []byte) ([]byte, error) {
	key := append(encID, field_recovery)
	item, err := getStorageItem(srvc, key)
	if err != nil {
		return nil, errors.New("get recovery error: " + err.Error())
	}
	return item.Value, nil
}

func getOwnerKey(srvc *native.NativeService, encID []byte, index uint32) (*publicKey, error) {
	key := append(encID, field_pk)
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], index)
	key = append(key, buf[:]...)
	item, err := getStorageItem(srvc, key)
	if err != nil {
		return nil, errors.New("get owner key error: " + err.Error())
	}

	var pk publicKey
	err = pk.SetBytes(item.Value)
	if err != nil {
		return nil, errors.New("get owner key error: invalid data")
	}

	return &pk, nil
}

func isOwner(srvc *native.NativeService, encID, pub []byte) bool {
	kid, err := findPk(srvc, encID, pub)
	if err != nil || len(kid) == 0 {
		log.Debug(err)
		return false
	}
	return true
}

func checkWitness(srvc *native.NativeService, pub []byte) error {
	pk, err := keypair.DeserializePublicKey(pub)
	if err != nil {
		return errors.New("invalid public key, " + err.Error())
	}
	addr := types.AddressFromPubKey(pk)
	if !srvc.ContextRef.CheckWitness(addr) {
		return errors.New("check witness failed")
	}
	return nil
}

func newEvent(srvc *native.NativeService, st interface{}) {
	e := event.NotifyEventInfo{}
	e.TxHash = srvc.Tx.Hash()
	e.ContractAddress = srvc.ContextRef.CurrentContext().ContractAddress
	e.States = st
	srvc.Notifications = append(srvc.Notifications, &e)
	return
}

func triggerRegisterEvent(srvc *native.NativeService, id []byte) {
	newEvent(srvc, []string{"Register", hex.EncodeToString(id)})
}

func triggerPublicEvent(srvc *native.NativeService, op string, id, pub []byte) {
	st := []string{"PublicKey", op, hex.EncodeToString(id), hex.EncodeToString(pub)}
	newEvent(srvc, st)
}

func triggerAttributeEvent(srvc *native.NativeService, op string, id, path []byte) {
	st := []string{"Attribute", op, string(id), string(path)}
	newEvent(srvc, st)
}

func getStorageItem(srvc *native.NativeService, key []byte) (*states.StorageItem, error) {
	val, err := srvc.CloneCache.Get(common.ST_STORAGE, key)
	if err != nil {
		return nil, err
	}
	t, ok := val.(*states.StorageItem)
	if !ok {
		return nil, errors.New("invalid value type")
	}
	return t, nil
}

func makeKeyID(id []byte, index uint32) []byte {
	encID, _ := encodeID(id)
	key := append(encID, field_pk)
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], index)
	return append(key, buf[:]...)
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
