package ontid

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

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

func GetDDO(srvc *native.NativeService) ([]byte, error) {
	var0, err := GetPublicKeys(srvc)
	if err != nil {
		return nil, fmt.Errorf("get DDO error: %s", err)
	}
	var1, err := GetAttributes(srvc)
	if err != nil {
		return nil, fmt.Errorf("get DDO error: %s", err)
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
	item, err := utils.LinkedlistGetHead(srvc, key)
	if err != nil {
		return nil, fmt.Errorf("get public keys error: cannot get the list head, %s", err)
	} else if len(item) == 0 {
		return nil, errors.New("get public keys error: get list head failed")
	}

	var i uint = 0
	var res bytes.Buffer
	for len(item) > 0 {
		node, err := utils.LinkedlistGetItem(srvc, key, item)
		if err != nil {
			return nil, fmt.Errorf("get public keys error: %s", err)
		} else if node == nil {
			return nil, errors.New("get public keys error: get list node failed")
		}

		var pk publicKey
		err = pk.SetBytes(node.GetPayload())
		if err != nil {
			return nil, fmt.Errorf("get public keys error: parse key error, %s", err)
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
	item, err := utils.LinkedlistGetHead(srvc, key)
	if err != nil {
		return nil, fmt.Errorf("get attributes error: get list head error, %s", err)
	} else if len(item) == 0 {
		return nil, errors.New("get attributes error: cannot get list head")
	}

	var res bytes.Buffer
	var i uint16 = 0
	for len(item) > 0 {
		node, err := utils.LinkedlistGetItem(srvc, key, item)
		if err != nil {
			return nil, fmt.Errorf("get attributes error: get storage item error, %s", err)
		} else if node == nil {
			return nil, fmt.Errorf("get attributes error: storage item, not exists")
		}

		var attr attribute
		err = attr.SetValue(node.GetPayload())
		if err != nil {
			return nil, fmt.Errorf("get attributes error: parse attribute failed, %s", err)
		}
		attr.key = item
		err = attr.Serialize(&res)
		if err != nil {
			return nil, fmt.Errorf("get attributes error: serialize error, %s", err)
		}

		i += 1
		item = node.GetNext()
	}

	return append([]byte{byte(i >> 8), byte(i & 0xff)}, res.Bytes()...), nil
}
