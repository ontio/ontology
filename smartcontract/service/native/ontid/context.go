package ontid

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

//
var _DefaultContexts = [][]byte{[]byte("https://www.w3.org/ns/did/v1"), []byte("https://ontid.ont.io/did/v1")}

func addContext(srvc *native.NativeService) ([]byte, error) {
	params := new(Context)
	if err := params.Deserialization(common.NewZeroCopySource(srvc.Input)); err != nil {
		return utils.BYTE_FALSE, errors.New("addContext error: deserialization params error, " + err.Error())
	}
	encId, err := encodeID(params.OntId)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("addContext error: " + err.Error())
	}

	if checkIDState(srvc, encId) == flag_not_exist {
		return utils.BYTE_FALSE, errors.New("register ONT ID error: have not registered")
	}

	if err := checkWitnessByIndex(srvc, encId, params.Index); err != nil {
		return utils.BYTE_FALSE, errors.New("verify signature failed: " + err.Error())
	}
	key := append(encId, FIELD_CONTEXT)

	if err := putContexts(srvc, key, params); err != nil {
		return utils.BYTE_FALSE, errors.New("addContext error: putContexts failed: " + err.Error())
	}
	updateProofAndTime(srvc, encId, params.Proof)
	return utils.BYTE_TRUE, nil
}

func removeContext(srvc *native.NativeService) ([]byte, error) {
	params := new(Context)
	if err := params.Deserialization(common.NewZeroCopySource(srvc.Input)); err != nil {
		return utils.BYTE_FALSE, errors.New("addContext error: deserialization params error, " + err.Error())
	}
	encId, err := encodeID(params.OntId)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("removeContext error: " + err.Error())
	}

	if checkIDState(srvc, encId) == flag_not_exist {
		return utils.BYTE_FALSE, errors.New("register ONT ID error: have not registered")
	}

	if err := checkWitnessByIndex(srvc, encId, params.Index); err != nil {
		return utils.BYTE_FALSE, errors.New("verify signature failed: " + err.Error())
	}
	key := append(encId, FIELD_CONTEXT)

	if err := deleteContexts(srvc, key, params); err != nil {
		return utils.BYTE_FALSE, errors.New("removeContext error: deleteContexts failed: " + err.Error())
	}
	updateProofAndTime(srvc, encId, params.Proof)
	return utils.BYTE_TRUE, nil
}

func deleteContexts(srvc *native.NativeService, key []byte, params *Context) error {
	contexts, err := getContexts(srvc, key)
	if err != nil {
		return fmt.Errorf("deleteContexts error: getContexts error, %s", err)
	}
	repeat := getRepeatContexts(contexts, params)
	var remove [][]byte
	for i := 0; i < len(params.Contexts); i++ {
		if _, ok := repeat[common.ToHexString(params.Contexts[i])]; ok {
			contexts = append(contexts[:i], contexts[i+1:]...)
			remove = append(remove, params.Contexts[i])
		}
	}
	triggerContextEvent(srvc, "remove", params.OntId, remove)
	err = storeContexts(contexts, srvc, key)
	if err != nil {
		return fmt.Errorf("deleteContexts error: storeContexts error, %s", err)
	}
	return nil
}

func putContexts(srvc *native.NativeService, key []byte, params *Context) error {
	contexts, err := getContexts(srvc, key)
	if err != nil {
		return fmt.Errorf("putContexts error: getContexts failed, %s", err)
	}
	repeat := getRepeatContexts(contexts, params)
	var add [][]byte
	for i := 0; i < len(params.Contexts); i++ {
		if _, ok := repeat[common.ToHexString(params.Contexts[i])]; !ok {
			if (bytes.Equal(params.Contexts[i], _DefaultContexts[0])) && (bytes.Equal(params.Contexts[i], _DefaultContexts[1])) {
				contexts = append(contexts, params.Contexts[i])
				add = append(add, params.Contexts[i])
			}
		}
	}
	triggerContextEvent(srvc, "add", params.OntId, add)
	err = storeContexts(contexts, srvc, key)
	if err != nil {
		return fmt.Errorf("putContexts error: storeContexts failed, %s", err)
	}
	return nil
}

func getRepeatContexts(contexts [][]byte, params *Context) map[string]bool {
	repeat := make(map[string]bool)
	for i := 0; i < len(contexts); i++ {
		for j := 0; j < len(params.Contexts); j++ {
			if bytes.Equal(contexts[i], params.Contexts[j]) {
				repeat[common.ToHexString(params.Contexts[j])] = true
			}
		}
	}
	return repeat
}

func getContexts(srvc *native.NativeService, key []byte) ([][]byte, error) {
	contextsStore, err := utils.GetStorageItem(srvc, key)
	if err != nil {
		return nil, errors.New("getContexts error:" + err.Error())
	}
	if contextsStore == nil {
		return nil, nil
	}
	contexts := new(Contexts)
	if err := contexts.Deserialization(common.NewZeroCopySource(contextsStore.Value)); err != nil {
		return nil, err
	}
	return *contexts, nil
}

func getContextsWithDefault(srvc *native.NativeService, key []byte) ([][]byte, error) {
	contexts, err := getContexts(srvc, key)
	if err != nil {
		return nil, fmt.Errorf("getContextsWithDefault error, %s", err)
	}
	contexts = append(_DefaultContexts, contexts...)
	return contexts, nil
}

func storeContexts(contexts Contexts, srvc *native.NativeService, key []byte) error {
	sink := common.NewZeroCopySink(nil)
	contexts.Serialization(sink)
	item := states.StorageItem{}
	item.Value = sink.Bytes()
	item.StateVersion = _VERSION_0
	srvc.CacheDB.Put(key, item.ToArray())
	return nil
}
