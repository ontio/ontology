package ontid

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

// TODO update time, proof
func addService(srvc *native.NativeService) ([]byte, error) {
	log.Debug("ID contract: addService")
	params := new(ServiceParam)
	if err := params.Deserialization(common.NewZeroCopySource(srvc.Input)); err != nil {
		return utils.BYTE_FALSE, errors.New("add service error: deserialization params error, " + err.Error())
	}
	encId, err := encodeID(params.OntId)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add service error: " + err.Error())
	}
	if checkIDState(srvc, encId) == flag_not_exist {
		return utils.BYTE_FALSE, errors.New("register ONT ID error: have not registered")
	}

	if err := checkWitnessByIndex(srvc, encId, params.Index); err != nil {
		return utils.BYTE_FALSE, errors.New("verify signature failed: " + err.Error())
	}
	err = putService(srvc, encId, params)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("putService failed: " + err.Error())
	}
	// TODO ADD PROOF
	return utils.BYTE_TRUE, nil
}

// TODO update time, proof
func updateService(srvc *native.NativeService) ([]byte, error) {
	log.Debug("ID contract: updateService")
	params := new(ServiceParam)
	if err := params.Deserialization(common.NewZeroCopySource(srvc.Input)); err != nil {
		return utils.BYTE_FALSE, errors.New("add service error: deserialization params error, " + err.Error())
	}
	encId, err := encodeID(params.OntId)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add service error: " + err.Error())
	}
	key := append(encId, FIELD_SERVICE)
	if checkIDState(srvc, encId) == flag_not_exist {
		return utils.BYTE_FALSE, errors.New("updateService error: have not registered")
	}

	if err := checkWitnessByIndex(srvc, encId, params.Index); err != nil {
		return utils.BYTE_FALSE, errors.New("verify signature failed: " + err.Error())
	}

	services, err := getServices(srvc, encId)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("get service error: " + err.Error())
	}
	if services == nil {
		return utils.BYTE_FALSE, errors.New("get services error: have not registered any service")
	}

	service := Service{
		ServiceId:      params.ServiceId,
		Type:           params.Type,
		ServiceEndpint: params.ServiceEndpint,
	}
	for i := 0; i < len(*services); i++ {
		if bytes.Equal((*services)[i].ServiceId, service.ServiceId) {
			(*services)[i] = service
			storeServices(services, srvc, key)
			return utils.BYTE_TRUE, nil
		}
	}
	return utils.BYTE_FALSE, nil
}

// TODO update time, proof
func removeService(srvc *native.NativeService) ([]byte, error) {
	log.Debug("ID contract: updateService")
	params := new(ServiceRemoveParam)
	if err := params.Deserialization(common.NewZeroCopySource(srvc.Input)); err != nil {
		return utils.BYTE_FALSE, errors.New("add service error: deserialization params error, " + err.Error())
	}
	encId, err := encodeID(params.OntId)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("add service error: " + err.Error())
	}
	key := append(encId, FIELD_SERVICE)
	if checkIDState(srvc, encId) == flag_not_exist {
		return utils.BYTE_FALSE, errors.New("updateService error: have not registered")
	}

	if err := checkWitnessByIndex(srvc, encId, params.Index); err != nil {
		return utils.BYTE_FALSE, errors.New("verify signature failed: " + err.Error())
	}

	services, err := getServices(srvc, encId)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("get service error: " + err.Error())
	}
	if services == nil {
		return utils.BYTE_FALSE, errors.New("get services error: have not registered any service")
	}
	for i := 0; i < len(*services); i++ {
		if bytes.Equal((*services)[i].ServiceId, params.ServiceId) {
			services := append((*services)[:i], (*services)[i+1:]...)
			storeServices(&services, srvc, key)
			return utils.BYTE_TRUE, nil
		}
	}
	return utils.BYTE_FALSE, nil
}

func getServices(srvc *native.NativeService, encId []byte) (*Services, error) {
	key := append(encId, FIELD_SERVICE)
	servicesStore, err := utils.GetStorageItem(srvc, key)
	if err != nil {
		return nil, errors.New("getServices error:" + err.Error())
	}
	if servicesStore == nil {
		return nil, nil
	}
	services := new(Services)
	if err := services.Deserialization(common.NewZeroCopySource(servicesStore.Value)); err != nil {
		return nil, err
	}
	return services, nil
}

func checkServiceExist(services Services, service Service) bool {
	for i := 0; i < len(services); i++ {
		if bytes.Equal(services[i].ServiceId, service.ServiceId) {
			return true
		}
	}
	return false
}

func storeServices(services *Services, srvc *native.NativeService, key []byte) {
	sink := common.NewZeroCopySink(nil)
	services.Serialization(sink)
	item := states.StorageItem{}
	item.Value = sink.Bytes()
	item.StateVersion = _VERSION_0
	srvc.CacheDB.Put(key, item.ToArray())
}

func putService(srvc *native.NativeService, encId []byte, params *ServiceParam) error {
	key := append(encId, FIELD_SERVICE)
	servicesStore, err := utils.GetStorageItem(srvc, key)
	if err != nil {
		return fmt.Errorf("get storage error, %s", err)
	}
	source := common.NewZeroCopySource(servicesStore.Value)
	services := new(Services)
	err = services.Deserialization(source)
	if err != nil {
		return fmt.Errorf("deserialize services error, %s", err)
	}

	service := Service{
		ServiceId:      params.ServiceId,
		Type:           params.Type,
		ServiceEndpint: params.ServiceEndpint,
	}
	if checkServiceExist(*services, service) {
		return fmt.Errorf("putService error: service has registered")
	}

	*services = append(*services, service)
	storeServices(services, srvc, key)
	return nil
}
