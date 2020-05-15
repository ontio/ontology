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

type SpaceUpdate struct {
	SpaceOwner     common.Address
	Payer          common.Address
	NewVolume      uint64
	NewTimeExpired uint64
}

func (this *SpaceUpdate) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.SpaceOwner)
	utils.EncodeAddress(sink, this.Payer)
	utils.EncodeVarUint(sink, this.NewVolume)
	utils.EncodeVarUint(sink, this.NewTimeExpired)
}

func (this *SpaceUpdate) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.SpaceOwner, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.Payer, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.NewVolume, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.NewTimeExpired, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	return nil
}
