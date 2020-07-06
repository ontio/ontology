/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */
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

type serviceJson struct {
	Id              string `json:"id"`
	Type            string `json:"type"`
	ServiceEndpoint string `json:"serviceEndpoint"`
}

func addService(srvc *native.NativeService) ([]byte, error) {
	log.Debug("ID contract: addService")
	params := new(ServiceParam)
	if err := params.Deserialization(common.NewZeroCopySource(srvc.Input)); err != nil {
		return utils.BYTE_FALSE, errors.New("addService error: deserialization params error, " + err.Error())
	}
	encId, err := encodeID(params.OntId)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("addService error: " + err.Error())
	}
	if !isValid(srvc, encId) {
		return utils.BYTE_FALSE, errors.New("addService error: have not registered")
	}

	if err := checkWitnessByIndex(srvc, encId, params.Index); err != nil {
		return utils.BYTE_FALSE, errors.New("verify signature failed: " + err.Error())
	}
	err = putService(srvc, encId, params)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("addService error: putService failed: " + err.Error())
	}
	updateTimeAndClearProof(srvc, encId)
	return utils.BYTE_TRUE, nil
}

func updateService(srvc *native.NativeService) ([]byte, error) {
	log.Debug("ID contract: updateService")
	params := new(ServiceParam)
	if err := params.Deserialization(common.NewZeroCopySource(srvc.Input)); err != nil {
		return utils.BYTE_FALSE, errors.New("updateService error: deserialization params error, " + err.Error())
	}
	encId, err := encodeID(params.OntId)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("updateService error: " + err.Error())
	}
	if !isValid(srvc, encId) {
		return utils.BYTE_FALSE, errors.New("updateService error: have not registered")
	}

	if err := checkWitnessByIndex(srvc, encId, params.Index); err != nil {
		return utils.BYTE_FALSE, errors.New("verify signature failed: " + err.Error())
	}

	services, err := getServices(srvc, encId)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("updateService: get service error: " + err.Error())
	}
	if services == nil {
		return utils.BYTE_FALSE, errors.New("updateService: get services error: have not registered any service")
	}

	service := Service{
		ServiceId:      params.ServiceId,
		Type:           params.Type,
		ServiceEndpint: params.ServiceEndpint,
	}
	key := append(encId, FIELD_SERVICE)
	for i := 0; i < len(services); i++ {
		if bytes.Equal(services[i].ServiceId, service.ServiceId) {
			services[i] = service
			storeServices(services, srvc, key)
			updateTimeAndClearProof(srvc, encId)
			triggerServiceEvent(srvc, "update", params.OntId, params.ServiceId)
			return utils.BYTE_TRUE, nil
		}
	}
	return utils.BYTE_FALSE, errors.New("updateService: update service error: have not registered such service")
}

func removeService(srvc *native.NativeService) ([]byte, error) {
	log.Debug("ID contract: updateService")
	params := new(ServiceRemoveParam)
	if err := params.Deserialization(common.NewZeroCopySource(srvc.Input)); err != nil {
		return utils.BYTE_FALSE, errors.New("removeService error: deserialization params error, " + err.Error())
	}
	encId, err := encodeID(params.OntId)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("removeService error: " + err.Error())
	}
	if !isValid(srvc, encId) {
		return utils.BYTE_FALSE, errors.New("removeService error: have not registered")
	}

	if err := checkWitnessByIndex(srvc, encId, params.Index); err != nil {
		return utils.BYTE_FALSE, errors.New("verify signature failed: " + err.Error())
	}

	services, err := getServices(srvc, encId)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("removeService error: get service error: " + err.Error())
	}
	if services == nil {
		return utils.BYTE_FALSE, errors.New("removeService error: have not registered any service")
	}
	key := append(encId, FIELD_SERVICE)
	for i := 0; i < len(services); i++ {
		if bytes.Equal((services)[i].ServiceId, params.ServiceId) {
			services := append((services)[:i], (services)[i+1:]...)
			storeServices(services, srvc, key)
			updateTimeAndClearProof(srvc, encId)
			triggerServiceEvent(srvc, "remove", params.OntId, params.ServiceId)
			return utils.BYTE_TRUE, nil
		}
	}
	return utils.BYTE_FALSE, errors.New("removeService: remove service error: have not registered such service")
}

func getServices(srvc *native.NativeService, encId []byte) (Services, error) {
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
	return *services, nil
}

func checkServiceExist(services Services, service Service) bool {
	for i := 0; i < len(services); i++ {
		if bytes.Equal(services[i].ServiceId, service.ServiceId) {
			return true
		}
	}
	return false
}

func storeServices(services Services, srvc *native.NativeService, key []byte) {
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
		return fmt.Errorf("putService error: get storage error, %s", err)
	}
	services := new(Services)
	if servicesStore != nil {
		source := common.NewZeroCopySource(servicesStore.Value)
		err = services.Deserialization(source)
		if err != nil {
			return fmt.Errorf("deserialize services error, %s", err)
		}
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
	storeServices(*services, srvc, key)
	triggerServiceEvent(srvc, "add", params.OntId, params.ServiceId)
	return nil
}

func getServicesJson(srvc *native.NativeService, encId []byte) ([]*serviceJson, error) {
	services, err := getServices(srvc, encId)
	if err != nil {
		return nil, err
	}
	r := make([]*serviceJson, 0)
	for _, p := range services {
		service := new(serviceJson)

		ontId, err := decodeID(encId)
		if err != nil {
			return nil, err
		}
		service.Id = fmt.Sprintf("%s#%s", string(ontId), string(p.ServiceId))
		service.Type = string(p.Type)
		service.ServiceEndpoint = string(p.ServiceEndpint)
		r = append(r, service)
	}
	return r, nil
}
