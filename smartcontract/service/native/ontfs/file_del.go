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

type FileDel struct {
	FileHash []byte
}

type FileDelList struct {
	FilesDel []FileDel
}

func (this *FileDel) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.FileHash)
}

func (this *FileDel) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.FileHash, err = DecodeVarBytes(source)
	if err != nil {
		return err
	}
	return nil
}

func (this *FileDelList) Serialization(sink *common.ZeroCopySink) {
	fileDelCount := uint64(len(this.FilesDel))
	utils.EncodeVarUint(sink, fileDelCount)

	for _, fileDel := range this.FilesDel {
		sinkTmp := common.NewZeroCopySink(nil)
		fileDel.Serialization(sinkTmp)
		sink.WriteVarBytes(sinkTmp.Bytes())
	}
}

func (this *FileDelList) Deserialization(source *common.ZeroCopySource) error {
	fileDelCount, err := utils.DecodeVarUint(source)
	if err != nil {
		return err
	}

	for i := uint64(0); i < fileDelCount; i++ {
		fileDelTmp, err := DecodeVarBytes(source)
		if err != nil {
			return err
		}

		var fileDel FileDel
		src := common.NewZeroCopySource(fileDelTmp)
		if err = fileDel.Deserialization(src); err != nil {
			return err
		}
		this.FilesDel = append(this.FilesDel, fileDel)
	}
	return nil
}
