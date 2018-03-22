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

package native

import (
	"github.com/Ontology/smartcontract/storage"
	scommon "github.com/Ontology/core/store/common"
	"bytes"
	"github.com/Ontology/common/serialization"
	"github.com/Ontology/core/types"
	"github.com/Ontology/smartcontract/event"
	"fmt"
)

type (
	Handler func(native *NativeService) (bool, error)
)

type NativeService struct {
	CloneCache *storage.CloneCache
	ServiceMap  map[string]Handler
	Notifications []*event.NotifyEventInfo
	Input []byte
	Tx *types.Transaction
}

func NewNativeService(dbCache scommon.IStateStore, input []byte, tx *types.Transaction) *NativeService {
	var nativeService NativeService
	nativeService.CloneCache = storage.NewCloneCache(dbCache)
	nativeService.Input = input
	nativeService.Tx = tx
	nativeService.ServiceMap = make(map[string]Handler)
	nativeService.Register("Token.Common.Transfer", Transfer)
	nativeService.Register("Token.Ont.Init", OntInit)
	return &nativeService
}

func(native *NativeService) Register(methodName string, handler Handler) {
	native.ServiceMap[methodName] = handler
}

func(native *NativeService) Invoke() (bool, error){
	bf := bytes.NewBuffer(native.Input)
	serviceName, err := serialization.ReadVarBytes(bf); if err != nil {
		return false, err
	}
	service, ok := native.ServiceMap[string(serviceName)]; if !ok {
		return false, fmt.Errorf("Native does not support this service:%s !",serviceName)
	}
	native.Input = bf.Bytes()
	return service(native)
}








