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
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type FileTransfer struct {
	FileHash []byte
	OriOwner common.Address
	NewOwner common.Address
}

type FileTransferList struct {
	FilesTransfer []FileTransfer
}

func (this *FileTransfer) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.FileHash)
	utils.EncodeAddress(sink, this.OriOwner)
	utils.EncodeAddress(sink, this.NewOwner)
}

func (this *FileTransfer) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.FileHash, err = DecodeVarBytes(source)
	if err != nil {
		return err
	}
	this.OriOwner, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.NewOwner, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	return nil
}

func (this *FileTransferList) Serialization(sink *common.ZeroCopySink) {
	fileTransCount := uint64(len(this.FilesTransfer))
	utils.EncodeVarUint(sink, fileTransCount)

	for _, fileTrans := range this.FilesTransfer {
		sinkTmp := common.NewZeroCopySink(nil)
		fileTrans.Serialization(sinkTmp)
		sink.WriteVarBytes(sinkTmp.Bytes())
	}
}

func (this *FileTransferList) Deserialization(source *common.ZeroCopySource) error {
	fileTransCount, err := utils.DecodeVarUint(source)
	if err != nil {
		return err
	}

	for i := uint64(0); i < fileTransCount; i++ {
		fileTransTmp, err := DecodeVarBytes(source)
		if err != nil {
			return err
		}

		var fileTrans FileTransfer
		src := common.NewZeroCopySource(fileTransTmp)
		if err = fileTrans.Deserialization(src); err != nil {
			return err
		}
		this.FilesTransfer = append(this.FilesTransfer, fileTrans)
	}
	return nil
}

func setFileOwner(native *native.NativeService, fileHash []byte, fileOwner common.Address) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	fileOwnerKey := GenFsFileOwnerKey(contract, fileHash)
	utils.PutBytes(native, fileOwnerKey, fileOwner[:])
}

func getFileOwner(native *native.NativeService, fileHash []byte) (common.Address, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	fileOwnerKey := GenFsFileOwnerKey(contract, fileHash)

	item, err := utils.GetStorageItem(native, fileOwnerKey)
	if err != nil || item == nil || item.Value == nil {
		return common.Address{}, fmt.Errorf("getFileOwner GetStorageItem error")
	}
	return common.AddressParseFromBytes(item.Value)
}

func delFileOwner(native *native.NativeService, fileHash []byte) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	fileOwnerKey := GenFsFileOwnerKey(contract, fileHash)
	native.CacheDB.Delete(fileOwnerKey)
}
