package ontid

import (
	"errors"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

// TODO ADD TIME and PROOF
func addAuthKey(srvc *native.NativeService) ([]byte, error) {
	params := new(AddAuthKeyParam)
	if err := params.Deserialization(common.NewZeroCopySource(srvc.Input)); err != nil {
		return utils.BYTE_FALSE, errors.New("add auth key error: deserialization params error, " + err.Error())
	}
	encId, err := encodeID(params.OntId)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add auth key error: " + err.Error())
	}

	if checkIDState(srvc, encId) == flag_not_exist {
		return utils.BYTE_FALSE, errors.New("add auth key error: have not registered")
	}

	if err := checkWitnessByIndex(srvc, encId, params.SignIndex); err != nil {
		return utils.BYTE_FALSE, errors.New("verify signature failed: " + err.Error())
	}

	if params.IfNewPublicKey {
		index, err := insertPk(srvc, encId, params.NewPublicKey.key, params.NewPublicKey.controller,
			USE_ACCESS, ONLY_AUTHENTICATION, params.Proof)
		if err != nil {
			return utils.BYTE_FALSE, errors.New("add auth key error, insertPk failed " + err.Error())
		}
		triggerAuthKeyEvent(srvc, "add", params.OntId, index)
	} else {
		err = changePkAuthentication(srvc, encId, params.Index, BOTH, params.Proof)
		if err != nil {
			return utils.BYTE_FALSE, errors.New("add auth key error, changePkAuthentication failed " + err.Error())
		}
		triggerAuthKeyEvent(srvc, "add", params.OntId, params.Index)
	}

	updateProofAndTime(srvc, encId, params.Proof)
	return utils.BYTE_TRUE, nil
}

// TODO ADD TIME and PROOF
func removeAuthKey(srvc *native.NativeService) ([]byte, error) {
	params := new(Context)
	if err := params.Deserialization(common.NewZeroCopySource(srvc.Input)); err != nil {
		return utils.BYTE_FALSE, errors.New("remove auth key error: deserialization params error, " + err.Error())
	}
	encId, err := encodeID(params.OntId)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("remove auth key error: " + err.Error())
	}

	if checkIDState(srvc, encId) == flag_not_exist {
		return utils.BYTE_FALSE, errors.New("remove auth key error: have not registered")
	}

	if err := checkWitnessByIndex(srvc, encId, params.Index); err != nil {
		return utils.BYTE_FALSE, errors.New("verify signature failed: " + err.Error())
	}

	if err := revokeAuthKey(srvc, encId, params.Index, params.Proof); err != nil {
		return utils.BYTE_FALSE, errors.New("remove auth key error, revokeAuthKey failed: " + err.Error())
	}

	updateProofAndTime(srvc, encId, params.Proof)
	triggerAuthKeyEvent(srvc, "remove", params.OntId, params.Index)
	return utils.BYTE_TRUE, nil
}

type Authentication struct {
	Authentication []interface{}
}

func getAuthentication(srvc *native.NativeService, encID []byte) (*Authentication, error) {
	key := append(encID, FIELD_PK)
	publicKeys, err := getAllPk_Version1(srvc, encID, key)
	if err != nil {
		return nil, err
	}
	authentication := new(Authentication)
	for _, p := range publicKeys {
		if p.authentication == ONLY_AUTHENTICATION {
			authentication.Authentication = append(authentication.Authentication, p)
		}
		if p.authentication == BOTH {
			authentication.Authentication = append(authentication.Authentication, string(p.key))
		}
	}

	return authentication, nil
}
