package ontid

import (
	"errors"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/service/native"
)

func updateOrInsertProof(srvc *native.NativeService, ontId []byte, proof []byte) error {
	encId, err := encodeID(ontId)
	if err != nil {
		return errors.New("updateOrInsertProof error: " + err.Error())
	}
	key := append(encId, FIELD_PROOF)
	item := states.StorageItem{}
	item.Value = proof
	item.StateVersion = _VERSION_0
	srvc.CacheDB.Put(key, item.ToArray())
	return nil
}
