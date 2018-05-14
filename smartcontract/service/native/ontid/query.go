package ontid

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/ontio/ontology/common/log"
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
	// arg1: key ID
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

	return pk, nil
}

func GetDDO(srvc *native.NativeService) ([]byte, error) {
	log.Debug("GetDDO")
	var0, err := GetPublicKeys(srvc)
	if err != nil {
		return nil, fmt.Errorf("get DDO error: %s", err)
	}
	var buf bytes.Buffer
	serialization.WriteVarBytes(&buf, var0)

	var1, err := GetAttributes(srvc)
	if err == nil {
		serialization.WriteVarBytes(&buf, var1)
	}

	res := buf.Bytes()
	log.Debug("DDO:", hex.EncodeToString(res))
	return res, nil
}

func GetPublicKeys(srvc *native.NativeService) ([]byte, error) {
	log.Debug("GetPublicKeys")
	did := srvc.Input
	if len(did) == 0 {
		return nil, errors.New("get public keys error: invalid ID")
	}
	key, err := encodeID(did)
	if err != nil {
		return nil, fmt.Errorf("get public keys error: %s", err)
	}
	key = append(key, FIELD_PK)
	item, err := utils.LinkedlistGetHead(srvc, key)
	if err != nil {
		return nil, fmt.Errorf("get public keys error: cannot get the list head, %s", err)
	} else if len(item) == 0 {
		return nil, errors.New("get public keys error: get list head failed")
	}

	var res bytes.Buffer
	for len(item) > 0 {
		node, err := utils.LinkedlistGetItem(srvc, key, item)
		if err != nil {
			return nil, fmt.Errorf("get public keys error: %s", err)
		} else if node == nil {
			return nil, errors.New("get public keys error: get list node failed")
		}

		keyID := binary.LittleEndian.Uint32(item)
		err = serialization.WriteUint32(&res, keyID)
		if err != nil {
			return nil, fmt.Errorf("get public keys error: serialize error, %s", err)
		}
		err = serialization.WriteVarBytes(&res, node.GetPayload())
		if err != nil {
			return nil, fmt.Errorf("get public keys error: serialize error, %s", err)
		}

		item = node.GetNext()
	}

	return res.Bytes(), nil
}

func GetAttributes(srvc *native.NativeService) ([]byte, error) {
	log.Debug("GetAttributes")
	did := srvc.Input
	if len(did) == 0 {
		return nil, errors.New("get attributes error: invalid ID")
	}
	key, err := encodeID(did)
	if err != nil {
		return nil, fmt.Errorf("get public keys error: %s", err)
	}
	key = append(key, FIELD_ATTR)
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

	return res.Bytes(), nil
}

func GetKeyState(srvc *native.NativeService) ([]byte, error) {
	log.Debug("GetKeyState")
	args := bytes.NewBuffer(srvc.Input)
	// arg0: ID
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return nil, fmt.Errorf("get key status failed: argument 0 error, %s", err)
	}
	// arg1: public key ID
	arg1, err := serialization.ReadUint32(args)
	if err != nil {
		return nil, fmt.Errorf("get key status failed: argument 1 error, %s", err)
	}

	key, err := encodeID(arg0)
	if err != nil {
		return nil, fmt.Errorf("get key status failed: %s", err)
	}

	key = append(key, FIELD_PK_STATE)
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], arg1)
	key = append(key, buf[:]...)

	item, err := utils.GetStorageItem(srvc, key)
	if err != nil || item == nil {
		log.Debug("key not exist")
		return []byte("not exist"), nil
	}

	log.Debug("key state: ", item.Value)
	if item.Value[0] == 1 {
		return []byte("in use"), nil
	} else {
		return []byte("revoked"), nil
	}
}
