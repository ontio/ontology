package ontid

import (
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/service/native"
)

func updateTime(srvc *native.NativeService, key []byte) {
	if srvc.Height < NEW_OWNER_BLOCK_HEIGHT {
		return
	}
	item := states.StorageItem{}
	sink := common.NewZeroCopySink(nil)
	sink.WriteUint32(srvc.Time)
	item.Value = sink.Bytes()
	item.StateVersion = _VERSION_0
	srvc.CacheDB.Put(key, item.ToArray())
}
