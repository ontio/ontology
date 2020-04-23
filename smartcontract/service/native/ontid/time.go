package ontid

import (
	"bytes"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/service/native"
)

func updateTime(srvc *native.NativeService, key []byte) {
	item := states.StorageItem{}
	bf := new(bytes.Buffer)
	serialization.WriteUint32(bf, srvc.Time)
	item.Value = bf.Bytes()
	item.StateVersion = _VERSION_0
	srvc.CacheDB.Put(key, item.ToArray())
}
