package ontid

import (
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/service/native"
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
