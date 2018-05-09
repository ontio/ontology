package ontid

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

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

func (this *publicKey) SetBytes(data []byte) error {
	buf := bytes.NewBuffer(data)
	return this.Deserialize(buf)
}
func (this *publicKey) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	err := this.Serialize(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func insertPk(srvc *native.NativeService, encID, pk []byte) error {
	var index uint32 = 0
	key1 := append(encID, field_pk)
	item, err := utils.LinkedlistGetHead(srvc, key1)
	if err == nil && item != nil {
		index = binary.LittleEndian.Uint32(item[len(key1):])
	}
	index += 1
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], index)
	key2 := buf[:]
	p := &publicKey{key: pk, revoked: false}
	val, err := p.Bytes()
	if err != nil {
		return errors.New("register ONT ID error: " + err.Error())
	}
	return utils.LinkedlistInsert(srvc, key1, key2, val)
}

func getPk(srvc *native.NativeService, encID []byte, index uint32) (*publicKey, error) {
	key1 := append(encID, field_pk)
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], index)
	key2 := buf[:]
	node, err := utils.LinkedlistGetItem(srvc, key1, key2)
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
		var pk publicKey
		err = pk.SetBytes(node.GetPayload())
		if err != nil {
			return nil, err
		}
		if bytes.Equal(pub, pk.key) {
			return key, nil
		}
		item = node.GetNext()
	}

	return nil, errors.New("public key not found")
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
