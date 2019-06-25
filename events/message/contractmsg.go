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
package message

import (
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/payload"
)

type MetaDataEvent struct {
	Version  uint32
	Height   uint32
	MetaData *payload.MetaDataCode
}

type ContractEvent struct {
	Version       uint32
	DeployHeight  uint32
	Contract      *payload.DeployCode
	Destroyed     bool
	DestroyHeight uint32
}

func (this *ContractEvent) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(this.Version)
	sink.WriteUint32(this.DeployHeight)
	this.Contract.Serialization(sink)
	sink.WriteBool(this.Destroyed)
	sink.WriteUint32(this.DestroyHeight)
}

func (this *ContractEvent) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	this.Version, eof = source.NextUint32()
	this.DeployHeight, eof = source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.Contract = &payload.DeployCode{}
	err := this.Contract.Deserialization(source)
	if err != nil {
		return err
	}
	var irr bool
	this.Destroyed, irr, eof = source.NextBool()
	if irr {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.DestroyHeight, eof = source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}
