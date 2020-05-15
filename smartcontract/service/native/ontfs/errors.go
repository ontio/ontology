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

package ontfs

import (
	"encoding/base64"
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type Errors struct {
	ObjectErrors map[string]string
}

func (this *Errors) AddObjectError(object string, errorString string) {
	if this.ObjectErrors == nil {
		this.ObjectErrors = make(map[string]string)
	}
	this.ObjectErrors[object] = errorString
}

func (this *Errors) ToString() string {
	sinkTmp := common.NewZeroCopySink(nil)

	errorCount := uint64(len(this.ObjectErrors))
	utils.EncodeVarUint(sinkTmp, errorCount)
	if errorCount == 0 {
		return base64.StdEncoding.EncodeToString(sinkTmp.Bytes())
	}

	for obj, error := range this.ObjectErrors {
		sinkTmp.WriteVarBytes([]byte(obj))
		sinkTmp.WriteVarBytes([]byte(error))
	}
	return base64.StdEncoding.EncodeToString(sinkTmp.Bytes())
}

func (this *Errors) FromString(errors string) error {
	errorsData, err := base64.StdEncoding.DecodeString(errors)
	if err != nil {
		return err
	}
	source := common.NewZeroCopySource(errorsData)
	errorCount, err := utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	if errorCount == 0 {
		return nil
	}

	for i := uint64(0); i < errorCount; i++ {
		obj, err := DecodeVarBytes(source)
		if err != nil {
			return err
		}
		error, err := DecodeVarBytes(source)
		if err != nil {
			return err
		}
		this.AddObjectError(string(obj), string(error))
	}
	return nil
}

func (this *Errors) AddErrorsEvent(native *native.NativeService) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	if !config.DefConfig.Common.EnableEventLog {
		return
	}

	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          this.ToString(),
		})
}

func (this *Errors) PrintErrors() {
	for obj, error := range this.ObjectErrors {
		fmt.Printf("[%s] error: %s\n", obj, error)
	}
}
