package native

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/event"
)

func RegisterIDContract(srvc *NativeService) error {
	srvc.Register("RegIdWithPublicKey", RegIdWithPublicKey)
	srvc.Register("AddKey", AddKey)
	srvc.Register("RemoveKey", RemoveKey)
	srvc.Register("AddRecovery", AddRecovery)
	srvc.Register("ChangeRecovery", ChangeRecovery)
	srvc.Register("AddAttribute", AddAttribute)
	srvc.Register("RemoveAttribute", RemoveAttribute)
	srvc.Register("AddAttributeArray", AddAttributeArray)
	return nil
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

func RegIdWithPublicKey(srvc *NativeService) error {
	log.Debug("registerIdWithPublicKey")
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

	log.Debug("arg 0:", hex.EncodeToString(arg0))
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
	insertPk(srvc, key, arg1, 1)
	// set flags
	err = setNumOfField(srvc, key, field_num_of_pk, 1)
	if err != nil {
		return errors.New("register ONT ID error: set number of public key error: " + err.Error())
	}
	err = setNumOfField(srvc, key, field_num_of_attr, 0)
	if err != nil {
		return errors.New("register ONT ID error: set number of attribute error: " + err.Error())
	}
	srvc.CloneCache.Add(common.ST_STORAGE, key, &states.StorageItem{Value: []byte{flag_exist}})

	triggerRegisterEvent(srvc, arg0)

	return nil
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
	err = serialization.WriteVarBytes(w, this.value)
	if err != nil {
		return err
	}
	err = serialization.WriteVarBytes(w, this.valueType)
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
	v, err := serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}
	vt, err := serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}
	this.key = k
	this.value = v
	this.valueType = vt
	return nil
}

func RegIdWithAttributes(srvc *NativeService) error {
	// parse arguments
	args := bytes.NewBuffer(srvc.Input)
	// arg0: ID
	arg0, err := serialization.ReadVarBytes(args)
	if len(arg0) == 0 {
		return errors.New("register ID with attributes error: invalid id")
	}
	key, err := encodeID(arg0)
	if err != nil {
		return errors.New("register ID with attributes error: " + err.Error())
	}
	if checkIDExistence(srvc, key) {
		return errors.New("register ID with attributes error: already registered")
	}

	// arg1: public key
	arg1, err := serialization.ReadVarBytes(args)
	public, err := keypair.DeserializePublicKey(arg1)
	if err != nil {
		return errors.New("register ID with attributes error: invalid public key: " + err.Error())
	}
	addr := types.AddressFromPubKey(public)
	if !srvc.ContextRef.CheckWitness(addr) {
		return errors.New("register ID with attributes error: check witness failed")
	}

	// arg2: attributes number
	arg2, err := serialization.ReadVarBytes(args)
	if len(arg2) < 2 {
		return errors.New("register ID with attributes error: invalid attribute number")
	}
	num := int(binary.LittleEndian.Uint16(arg2))
	attr := make([]attribute, num)
	for i := 0; i < num; i++ {
		err = attr[i].Deserialize(args)
		if err != nil {
			return errors.New("register ID with attributes error: parse attribute error")
		}
	}

	err = insertPk(srvc, key, arg1, 1)
	if err != nil {
		return errors.New("register ID with attributes error: store pubic key error: " + err.Error())
	}

	for i := 0; i < num; i++ {
		err = insertAttr(srvc, key, &attr[i])
		if err != nil {
			return errors.New("register ID with attributes error: " + err.Error())
		}
	}

	err = setNumOfField(srvc, key, field_num_of_pk, 1)
	if err != nil {
		return errors.New("register ID with attributes error: set number of public key error: " + err.Error())
	}
	err = setNumOfField(srvc, key, field_num_of_attr, uint32(num))
	if err != nil {
		return errors.New("register ID with attributes error: set number of attribute error: " + err.Error())
	}
	srvc.CloneCache.Add(common.ST_STORAGE, key, &states.StorageItem{Value: []byte{flag_exist}})

	triggerRegisterEvent(srvc, arg0)

	return nil
}

func AddKey(srvc *NativeService) error {
	args := bytes.NewBuffer(srvc.Input)
	// arg0: id
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("add key failed: argument error, " + err.Error())
	}

	// arg1: public key
	arg1, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("add key failed: argument error, " + err.Error())
	}

	// arg2: operator's public key
	arg2, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("add key failed: argument error, " + err.Error())
	}
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
	if !isOwner(srvc, key, arg1) {
		return errors.New("add key failed: operator has no authorization")
	}

	item, err := findPk(srvc, key, arg1)
	if item != nil {
		return errors.New("add key failed: already exists")
	}

	n, err := getNumOfField(srvc, key, field_num_of_pk)
	if err != nil {
		return errors.New("add key failed: get number error, " + err.Error())
	}

	n += 1
	err = insertPk(srvc, key, arg1, uint16(n))
	if err != nil {
		return errors.New("add key failed: insert public key error, " + err.Error())
	}
	err = setNumOfField(srvc, key, field_num_of_pk, n)
	if err != nil {
		return errors.New("add key failed: set number error, " + err.Error())
	}

	return nil
}

func RemoveKey(srvc *NativeService) error {
	args := bytes.NewBuffer(srvc.Input)
	// arg0: id
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("remove key failed: argument error, " + err.Error())
	}

	// arg1: public key
	arg1, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("remove key failed: argument error, " + err.Error())
	}

	// arg2: operator's public key
	arg2, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("remove key failed: argument error, " + err.Error())
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
	ok, err := linkedlistDelete(srvc, key, key1)
	if err != nil {
		return errors.New("remove key failed: delete error, " + err.Error())
	} else if !ok {
		return errors.New("remove key failed: key not found")
	}

	return nil
}

func AddRecovery(srvc *NativeService) error {
	args := bytes.NewBuffer(srvc.Input)
	// arg0: ID
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("add recovery failed: argument error")
	}
	// arg1: recovery
	arg1, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("add recovery failed: argument error")
	}
	// arg2: operator's public key
	arg2, err := serialization.ReadVarBytes(args)
	if err != nil {
		return errors.New("add recovery failed: argument error")
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

func ChangeRecovery(srvc *NativeService) error {
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

func AddAttribute(srvc *NativeService) error {
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

		n, err := getNumOfField(srvc, key, field_num_of_attr)
		if err != nil {
			return errors.New("add attribute failed: get number error, " + err.Error())
		}
		n += 1
		err = setNumOfField(srvc, key, field_num_of_attr, uint32(n))
		if err != nil {
			return errors.New("add attribute failed: set number error, " + err.Error())
		}
		triggerAttributeEvent(srvc, "add", arg0, arg1)
	}
	return nil
}

func RemoveAttribute(srvc *NativeService) error {
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

	ok, err := linkedlistDelete(srvc, key, arg1)
	if err != nil {
		return errors.New("remove attribute failed: delete error, " + err.Error())
	} else if !ok {
		return errors.New("remove attribute failed: attribute not exist")
	}

	n, err := getNumOfField(srvc, key, field_num_of_attr)
	if err != nil {
		return errors.New("remove attribute failed: get number error, " + err.Error())
	}
	n += 1
	err = setNumOfField(srvc, key, field_num_of_attr, uint32(n))
	if err != nil {
		return errors.New("remove attribute failed: set number error, " + err.Error())
	}
	triggerAttributeEvent(srvc, "remove", arg0, arg1)
	return nil
}

func AddAttributeArray(srvc *NativeService) error {
	//TODO
	return nil
}

type DDO struct {
	Keys []publicKey `json:"Owners,omitempty"`
}

func GetDDO(srvc *NativeService) ([]byte, error) {
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

func GetPublicKeys(srvc *NativeService) ([]byte, error) {
	did := srvc.Input
	if len(did) == 0 {
		return nil, errors.New("get public keys error: invalid ID")
	}
	key := append(did, field_pk)
	item, err := linkedlistGetHead(srvc, key)
	if err != nil {
		return nil, errors.New("get public keys error: cannot get the list head, " + err.Error())
	} else if len(item) == 0 {
		return nil, errors.New("get public keys error: get list head failed")
	}

	var i uint16 = 0
	var res bytes.Buffer
	for len(item) > 0 {
		node, err := linkedlistGetItem(srvc, key, item)
		if err != nil {
			return nil, errors.New("get public keys error: " + err.Error())
		} else if node == nil {
			return nil, errors.New("get public keys error: get list node failed")
		}

		var pk publicKey
		err = pk.SetBytes(node.payload)
		if err != nil {
			return nil, errors.New("get public keys error: parse key error, " + err.Error())
		}
		serialization.WriteVarBytes(&res, pk.key)
		//TODO key id?

		i += 1
		item = node.next
	}

	return append([]byte{byte(i >> 8), byte(i & 0xff)}, res.Bytes()...), nil
}

func GetAttributes(srvc *NativeService) ([]byte, error) {
	did := srvc.Input
	if len(did) == 0 {
		return nil, errors.New("get attributes error: invalid ID")
	}
	key := append(did, field_attr)
	item, err := linkedlistGetHead(srvc, key)
	if err != nil {
		return nil, errors.New("get attributes error: get list head error, " + err.Error())
	} else if len(item) == 0 {
		return nil, errors.New("get attributes error: cannot get list head")
	}

	var res bytes.Buffer
	var i uint16 = 0
	for len(item) > 0 {
		node, err := linkedlistGetItem(srvc, key, item)
		if err != nil {

		} else if node == nil {

		}

		var attr attribute
		err = attr.SetValue(node.payload)
		if err != nil {
			return nil, errors.New("get attributes error: parse attribute failed, " + err.Error())
		}

		var buf bytes.Buffer
		serialization.WriteVarBytes(&buf, attr.key)
		serialization.WriteVarBytes(&buf, attr.valueType)
		serialization.WriteVarBytes(&buf, attr.value)
		serialization.WriteVarBytes(&res, buf.Bytes())

		i += 1
		item = node.next
	}

	return append([]byte{byte(i >> 8), byte(i & 0xff)}, res.Bytes()...), nil
}

const flag_exist = 0x01

func checkIDExistence(srvc *NativeService, encID []byte) bool {
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
	field_num_of_pk byte = 1 + iota
	field_pk
	field_num_of_attr
	field_attr_name
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
	return enc, nil
}

func decodeID(data []byte) ([]byte, error) {
	if len(data) == 0 || len(data) != int(data[0])+1 {
		return nil, errors.New("decode ONT ID error: invalid data length")
	}
	return data[1:], nil
}

func setNumOfField(srvc *NativeService, encID []byte, field byte, n uint32) error {
	key := append(encID, field)
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, n)
	val := &states.StorageItem{Value: buf}
	srvc.CloneCache.Add(common.ST_STORAGE, key, val)
	return nil
}

func getNumOfField(srvc *NativeService, encID []byte, field byte) (uint32, error) {
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

func insertPk(srvc *NativeService, encID, pk []byte, index uint16) error {
	key1 := append(encID, field_pk)
	key2 := append(key1, byte(index>>8))
	key2 = append(key2, byte(index&0xFF))
	p := &publicKey{key: pk, revoked: false}
	val, err := p.Bytes()
	if err != nil {
		return errors.New("register ONT ID error: " + err.Error())
	}
	return linkedlistInsert(srvc, key1, key2, val)
}

func findPk(srvc *NativeService, encID, pub []byte) ([]byte, error) {
	key := append(encID, field_pk)
	item, err := linkedlistGetHead(srvc, key)
	if err != nil {
		return nil, err
	}

	for len(item) > 0 {
		node, err := linkedlistGetItem(srvc, key, item)
		if err != nil {
			continue
		}
		var pk publicKey
		err = pk.SetBytes(node.payload)
		if err != nil {
			return nil, err
		}
		if bytes.Equal(pub, pk.key) {
			return key, nil
		}
		item = node.next
	}

	return nil, errors.New("public key not found")
}

func insertAttr(srvc *NativeService, encID []byte, attr *attribute) error {
	key := append(encID, field_attr)
	val, err := attr.Value()
	if err != nil {
		return errors.New("serialize attribute value error: " + err.Error())
	}
	err = linkedlistInsert(srvc, key, attr.key, val)
	if err != nil {
		return errors.New("store attribute error: " + err.Error())
	}
	return nil
}

func updateAttr(srvc *NativeService, encID []byte, attr *attribute) error {
	//TODO direct update instead of delete + insert
	key := append(encID, field_attr)
	ok, err := linkedlistDelete(srvc, key, attr.key)
	if err != nil {
		return err
	} else if !ok {
		return errors.New("delete old attribute failed")
	}

	val, err := attr.Value()
	if err != nil {
		return errors.New("serialize attribute value error: " + err.Error())
	}
	err = linkedlistInsert(srvc, key, attr.key, val)
	if err != nil {
		return errors.New("store attribute error: " + err.Error())
	}
	return nil
}

func findAttr(srvc *NativeService, encID, item []byte) (*LinkedlistNode, error) {
	key := append(encID, field_attr)
	return linkedlistGetItem(srvc, key, item)
}

func setRecovery(srvc *NativeService, encID, recovery []byte) error {
	key := append(encID, field_recovery)
	val := &states.StorageItem{Value: recovery}
	srvc.CloneCache.Add(common.ST_STORAGE, key, val)
	return nil
}

func getRecovery(srvc *NativeService, encID []byte) ([]byte, error) {
	key := append(encID, field_recovery)
	val, err := srvc.CloneCache.Get(common.ST_STORAGE, key)
	if err != nil {
		return nil, err
	}
	t, ok := val.(*states.StorageItem)
	if !ok {
		return nil, errors.New("get storage item error: invalid value type")
	}
	return t.Value, nil
}

func isOwner(srvc *NativeService, encID, pub []byte) bool {
	kid, err := findPk(srvc, encID, pub)
	if err != nil || len(kid) == 0 {
		return false
	}
	return true
}

func makeKeyID(did []byte, n int) []byte {
	//TODO
	return nil
}

func parseKeyID(kID []byte) (did []byte, n int) {
	//TODO
	return nil, 0
}

func checkWitness(srvc *NativeService, pub []byte) error {
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

func newEvent(srvc *NativeService, st interface{}) {
	e := event.NotifyEventInfo{}
	e.TxHash = srvc.Tx.Hash()
	e.ContractAddress = srvc.ContextRef.CurrentContext().ContractAddress
	e.States = st
	srvc.Notifications = append(srvc.Notifications, &e)
	return
}

func triggerRegisterEvent(srvc *NativeService, id []byte) {
	newEvent(srvc, []string{"Register", hex.EncodeToString(id)})
}

func triggerPublicEvent(srvc *NativeService, op string, id, pub []byte) {
	st := []string{"PublicKey", op, hex.EncodeToString(id), hex.EncodeToString(pub)}
	newEvent(srvc, st)
}

func triggerAttributeEvent(srvc *NativeService, op string, id, path []byte) {
	st := []string{"Attribute", op, hex.EncodeToString(id), hex.EncodeToString(path)}
	newEvent(srvc, st)
}
