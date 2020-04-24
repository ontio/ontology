package ontid

import (
	"errors"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func updateOrInsertProof(srvc *native.NativeService, encId []byte, proof []byte) {
	if srvc.Height < NEW_OWNER_BLOCK_HEIGHT {
		return
	}
	key := append(encId, FIELD_PROOF)
	item := states.StorageItem{}
	item.Value = proof
	item.StateVersion = _VERSION_0
	srvc.CacheDB.Put(key, item.ToArray())
}

func getProof(srvc *native.NativeService, encId []byte) (string, error) {
	key := append(encId, FIELD_PROOF)
	proofStore, err := utils.GetStorageItem(srvc, key)
	if err != nil {
		return "", errors.New("getProof error:" + err.Error())
	}
	source := common.NewZeroCopySource(proofStore.Value)
	proof, err := utils.DecodeVarBytes(source)
	if err != nil {
		return "", errors.New("DecodeVarBytes error:" + err.Error())
	}
	return string(proof), nil
}
