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

package payload

import (
	"io"

	"github.com/ontio/ontology/common"
)

type MetaDataCode struct {
	OntVersion uint64
	Contract   common.Address
	Owner      common.Address
	AllShard   bool
	IsFrozen   bool
	ShardId    uint64

	// all contract that can be invoked by self, include nested invoke
	InvokedContract []common.Address
}

func NewDefaultMetaData() *MetaDataCode {
	return &MetaDataCode{
		OntVersion: common.VERSION_SUPPORT_SHARD,
		Contract:   common.ADDRESS_EMPTY,
		Owner:      common.ADDRESS_EMPTY,
		AllShard:   false,
		IsFrozen:   false,
		ShardId:    0,

		InvokedContract: make([]common.Address, 0),
	}
}

func (this *MetaDataCode) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.OntVersion)
	sink.WriteAddress(this.Contract)
	sink.WriteAddress(this.Owner)
	sink.WriteBool(this.AllShard)
	sink.WriteBool(this.IsFrozen)
	sink.WriteUint64(this.ShardId)
	sink.WriteUint64(uint64(len(this.InvokedContract)))
	for _, addr := range this.InvokedContract {
		sink.WriteAddress(addr)
	}
}

func (this *MetaDataCode) Deserialization(source *common.ZeroCopySource) error {
	var irr, eof bool
	this.OntVersion, eof = source.NextUint64()
	this.Contract, eof = source.NextAddress()
	this.Owner, eof = source.NextAddress()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.AllShard, irr, eof = source.NextBool()
	if irr {
		return common.ErrIrregularData
	}

	if eof {
		return io.ErrUnexpectedEOF
	}
	this.IsFrozen, irr, eof = source.NextBool()
	if irr {
		return common.ErrIrregularData
	}

	if eof {
		return io.ErrUnexpectedEOF
	}
	this.ShardId, eof = source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	addrNum, eof := source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.InvokedContract = make([]common.Address, 0)
	for i := uint64(0); i < addrNum; i++ {
		addr, eof := source.NextAddress()
		if eof {
			return io.ErrUnexpectedEOF
		}
		this.InvokedContract = append(this.InvokedContract, addr)
	}
	return nil
}

func (this *MetaDataCode) ToArray() []byte {
	sink := common.NewZeroCopySink(0)
	this.Serialization(sink)
	return sink.Bytes()
}
