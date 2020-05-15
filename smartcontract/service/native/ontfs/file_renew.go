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
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type FileReNew struct {
	FileHash       []byte
	FileOwner      common.Address
	Payer          common.Address
	NewTimeExpired uint64
}

type FileReNewList struct {
	FilesReNew []FileReNew
}

func (this *FileReNew) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.FileHash)
	utils.EncodeAddress(sink, this.FileOwner)
	utils.EncodeAddress(sink, this.Payer)
	utils.EncodeVarUint(sink, this.NewTimeExpired)
}

func (this *FileReNew) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.FileHash, err = DecodeVarBytes(source)
	if err != nil {
		return err
	}
	this.FileOwner, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.Payer, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.NewTimeExpired, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	return nil
}

func (this *FileReNewList) Serialization(sink *common.ZeroCopySink) {
	fileReNewCount := uint64(len(this.FilesReNew))
	utils.EncodeVarUint(sink, fileReNewCount)

	for _, fileReNew := range this.FilesReNew {
		sinkTmp := common.NewZeroCopySink(nil)
		fileReNew.Serialization(sinkTmp)
		sink.WriteVarBytes(sinkTmp.Bytes())
	}
}

func (this *FileReNewList) Deserialization(source *common.ZeroCopySource) error {
	fileReNewCount, err := utils.DecodeVarUint(source)
	if err != nil {
		return err
	}

	for i := uint64(0); i < fileReNewCount; i++ {
		fileReNewTmp, err := DecodeVarBytes(source)
		if err != nil {
			return err
		}

		var fileReNew FileReNew
		src := common.NewZeroCopySource(fileReNewTmp)
		if err = fileReNew.Deserialization(src); err != nil {
			return err
		}
		this.FilesReNew = append(this.FilesReNew, fileReNew)
	}
	return nil
}
