package ontid

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func regIdWithController(srvc *native.NativeService) ([]byte, error) {
	args := bytes.NewBuffer(srvc.Input)
	// arg0: ID
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("argument 0 error")
	}

	if !account.VerifyID(string(arg0)) {
		return utils.BYTE_FALSE, fmt.Errorf("invalid ID")
	}

	encId, err := encodeID(arg0)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	if checkIDExistence(srvc, encId) {
		return utils.BYTE_FALSE, fmt.Errorf("%s already registered", string(arg0))
	}

	// arg1: controller
	arg1, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("argument 1 error")
	}

	if bytes.Equal(arg1[:8], []byte("did:ont:")) {
		err = verifySingleController(srvc, arg1, args)
		if err != nil {
			return utils.BYTE_FALSE, err
		}
	} else {
		controller, err := deserializeGroup(arg1)
		if err != nil {
			return utils.BYTE_FALSE, errors.New("deserialize controller error")
		}
		err = verifyGroupController(srvc, controller, args)
		if err != nil {
			return utils.BYTE_FALSE, err
		}
	}

	key := append(encId, FIELD_CONTROLLER)
	utils.PutBytes(srvc, key, arg1)

	srvc.CacheDB.Put(key, states.GenRawStorageItem([]byte{flag_exist}))
	triggerRegisterEvent(srvc, arg0)
	return utils.BYTE_TRUE, nil
}

func verifyController(srvc *native.NativeService) ([]byte, error) {
	args := bytes.NewBuffer(srvc.Input)
	// arg0: ID
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("argument 0 error, %s", err)
	}

	key, err := encodeID(arg0)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	err = verifyControllerSignature(srvc, key, args)
	if err == nil {
		return utils.BYTE_TRUE, nil
	} else {
		return utils.BYTE_FALSE, fmt.Errorf("verification failed, %s", err)
	}
}

func removeController(srvc *native.NativeService) ([]byte, error) {
	args := bytes.NewBuffer(srvc.Input)
	// arg0: id
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("argument 0 error")
	}
	// arg1: public key index
	arg1, err := utils.ReadVarUint(args)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("argument 1 error")
	}
	encId, err := encodeID(arg0)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	pk, err := getPk(srvc, encId, uint32(arg1))
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	if pk.revoked {
		return utils.BYTE_FALSE, fmt.Errorf("authentication failed, public key is removed")
	}
	err = checkWitness(srvc, pk.key)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("checkWitness failed")
	}
	key := append(encId, FIELD_CONTROLLER)
	srvc.CacheDB.Delete(key)
	//TODO event
	return utils.BYTE_TRUE, nil
}

func addKeyByController(srvc *native.NativeService) ([]byte, error) {
	args := bytes.NewBuffer(srvc.Input)
	// arg0: id
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("argument 0 error")
	}

	// arg1: public key
	arg1, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("argument 1 error")
	}

	encId, err := encodeID(arg0)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	err = verifyControllerSignature(srvc, encId, args)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("verification failed, %s", err)
	}

	index, err := insertPk(srvc, encId, arg1)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("insertion failed, %s", err)
	}

	triggerPublicEvent(srvc, "add", arg0, arg1, index)
	return utils.BYTE_TRUE, nil
}

func removeKeyByController(srvc *native.NativeService) ([]byte, error) {
	args := bytes.NewBuffer(srvc.Input)
	// arg0: id
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 0")
	}

	// arg1: public key index
	arg1, err := utils.ReadVarUint(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 1")
	}

	encId, err := encodeID(arg0)
	if err != nil {
		return utils.BYTE_FALSE, errors.New(err.Error())
	}

	err = verifyControllerSignature(srvc, encId, args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("verifying signature failed")
	}

	pk, err := revokePkByIndex(srvc, encId, uint32(arg1))
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	triggerPublicEvent(srvc, "remove", arg0, pk, uint32(arg1))
	return utils.BYTE_TRUE, nil
}

func addAttributesByController(srvc *native.NativeService) ([]byte, error) {
	args := bytes.NewBuffer(srvc.Input)
	// arg0: id
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("argument 0 error")
	}

	// arg1: attributes
	num, err := utils.ReadVarUint(args)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("argument 1 error: %s", err)
	}
	var arg1 = make([]attribute, 0)
	for i := 0; i < int(num); i++ {
		var v attribute
		err = v.Deserialize(args)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("argument 1 error: %s", err)
		}
		arg1 = append(arg1, v)
	}

	encId, err := encodeID(arg0)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	err = verifyControllerSignature(srvc, encId, args)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("verification failed, %s", err)
	}

	err = batchInsertAttr(srvc, encId, arg1)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("insert attributes error, %s", err)
	}

	paths := getAttrKeys(arg1)
	triggerAttributeEvent(srvc, "add", arg0, paths)
	return utils.BYTE_TRUE, nil
}

func removeAttributeByController(srvc *native.NativeService) ([]byte, error) {
	args := bytes.NewBuffer(srvc.Input)
	// arg0: id
	arg0, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 0 error")
	}

	// arg1: path
	arg1, err := serialization.ReadVarBytes(args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("argument 1 error")
	}

	encId, err := encodeID(arg0)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	err = verifyControllerSignature(srvc, encId, args)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("verifying signature failed")
	}

	err = deleteAttr(srvc, encId, arg1)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	triggerAttributeEvent(srvc, "remove", arg0, [][]byte{arg1})
	return utils.BYTE_TRUE, nil
}

func getController(srvc *native.NativeService, encId []byte) (interface{}, error) {
	key := append(encId, FIELD_CONTROLLER)
	item, err := utils.GetStorageItem(srvc, key)
	if err != nil {
		return nil, err
	}

	if bytes.Equal(item.Value[:8], []byte("did:ont:")) {
		return item.Value, nil
	} else {
		return deserializeGroup(item.Value)
	}
}

func verifySingleController(srvc *native.NativeService, id []byte, args io.Reader) error {
	// public key index
	index, err := utils.ReadVarUint(args)
	if err != nil {
		return fmt.Errorf("index error, %s", err)
	}
	encId, err := encodeID(id)
	if err != nil {
		return err
	}
	pk, err := getPk(srvc, encId, uint32(index))
	if err != nil {
		return err
	}
	err = checkWitness(srvc, pk.key)
	if err != nil {
		return err
	}
	return nil
}

func verifyGroupController(srvc *native.NativeService, group *Group, args io.Reader) error {
	// signers
	buf, err := serialization.ReadVarBytes(args)
	if err != nil {
		return fmt.Errorf("signers error, %s", err)
	}
	signers, err := deserializeSigners(buf)
	if err != nil {
		return fmt.Errorf("signers error, %s", err)
	}
	if !verifyGroupSignature(srvc, group, signers) {
		return fmt.Errorf("verification failed")
	}
	return nil
}

func verifyControllerSignature(srvc *native.NativeService, encId []byte, args io.Reader) error {
	ctrl, err := getController(srvc, encId)
	if err != nil {
		return err
	}

	switch t := ctrl.(type) {
	case []byte:
		return verifySingleController(srvc, t, args)
	case *Group:
		return verifyGroupController(srvc, t, args)
	default:
		return fmt.Errorf("unknown controller type")
	}
}