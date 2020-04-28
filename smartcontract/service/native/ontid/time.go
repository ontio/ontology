package ontid

import (
	"errors"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
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

func getUpdateTime(srvc *native.NativeService, encId []byte) (uint32, error) {
	key := append(encId, FIELD_UPDATED)
	return getTime(srvc, key)
}

func getCreateTime(srvc *native.NativeService, encId []byte) (uint32, error) {
	key := append(encId, FIELD_CREATED)
	return getTime(srvc, key)
}

func getTime(srvc *native.NativeService, key []byte) (uint32, error) {
	timeStore, err := utils.GetStorageItem(srvc, key)
	if err != nil {
		return 0, errors.New("getTime error:" + err.Error())
	}
	var createTime uint32 = 0
	if timeStore != nil {
		source := common.NewZeroCopySource(timeStore.Value)
		createTime, err = utils.DecodeUint32(source)
		if err != nil {
			return 0, errors.New("DecodeUint32 error:" + err.Error())
		}
	}
	return createTime, nil
}
