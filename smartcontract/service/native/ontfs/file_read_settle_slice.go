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

type FileReadSettleSlice struct {
	FileHash     []byte
	PayFrom      common.Address
	PayTo        common.Address
	SliceId      uint64
	PledgeHeight uint64
	Sig          []byte
	PubKey       []byte
}

func (this *FileReadSettleSlice) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.FileHash)
	utils.EncodeAddress(sink, this.PayFrom)
	utils.EncodeAddress(sink, this.PayTo)
	utils.EncodeVarUint(sink, this.SliceId)
	utils.EncodeVarUint(sink, this.PledgeHeight)
	sink.WriteVarBytes(this.Sig)
	sink.WriteVarBytes(this.PubKey)
}

func (this *FileReadSettleSlice) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.FileHash, err = DecodeVarBytes(source)
	if err != nil {
		return err
	}
	this.PayFrom, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.PayTo, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.SliceId, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.PledgeHeight, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.Sig, err = DecodeVarBytes(source)
	if err != nil {
		return err
	}
	this.PubKey, err = DecodeVarBytes(source)
	if err != nil {
		return err
	}
	return nil
}
