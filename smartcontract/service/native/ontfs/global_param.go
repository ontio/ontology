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
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type FsGlobalParam struct {
	MinDownLoadFee           uint64 //min download fee for single task
	NodeMinVolume            uint64 //min total volume with fsNode
	NodePerKbPledge          uint64 //fsNode's pledge for participant
	GasPerKbForRead          uint64 //cost for ontfs-sdk read from fsNode
	GasPerKbForSaveWithFile  uint64 //cost for ontfs-sdk save from fsNode
	GasPerKbForSaveWithSpace uint64 //cost for ontfs-sdk save from fsNode
}

func (this *FsGlobalParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.MinDownLoadFee)
	utils.EncodeVarUint(sink, this.NodeMinVolume)
	utils.EncodeVarUint(sink, this.NodePerKbPledge)
	utils.EncodeVarUint(sink, this.GasPerKbForRead)
	utils.EncodeVarUint(sink, this.GasPerKbForSaveWithFile)
	utils.EncodeVarUint(sink, this.GasPerKbForSaveWithSpace)
}

func (this *FsGlobalParam) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.MinDownLoadFee, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.NodeMinVolume, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.NodePerKbPledge, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.GasPerKbForRead, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.GasPerKbForSaveWithFile, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.GasPerKbForSaveWithSpace, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	return err
}

func setGlobalParam(native *native.NativeService, globalParam *FsGlobalParam) {
	globalParamKey := GenGlobalParamKey(native.ContextRef.CurrentContext().ContractAddress)
	sink := common.NewZeroCopySink(nil)
	globalParam.Serialization(sink)
	utils.PutBytes(native, globalParamKey, sink.Bytes())
}

func getGlobalParam(native *native.NativeService) (*FsGlobalParam, error) {
	var globalParam FsGlobalParam

	globalParamKey := GenGlobalParamKey(native.ContextRef.CurrentContext().ContractAddress)
	item, err := utils.GetStorageItem(native, globalParamKey)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "getGlobalParam GetStorageItem error!")
	}
	if item == nil {
		globalParam = FsGlobalParam{
			MinDownLoadFee:           DefaultMinDownLoadFee,
			NodeMinVolume:            DefaultNodeMinVolume,
			NodePerKbPledge:          DefaultNodePerKbPledge,
			GasPerKbForRead:          DefaultGasPerKbForRead,
			GasPerKbForSaveWithFile:  DefaultGasPerKbForSaveWithFile,
			GasPerKbForSaveWithSpace: DefaultGasPerKbForSaveWithSpace,
		}
		return &globalParam, nil
	}

	source := common.NewZeroCopySource(item.Value)
	if err := globalParam.Deserialization(source); err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "getGlobalParam Deserialization error!")
	}
	return &globalParam, nil
}
