/*
 * Copyright (C) 2019 The ontology Authors
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
package oep4

import (
	"fmt"
	"io"
	"math/big"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type AssetId uint64

type Oep4 struct {
	Name        string
	Symbol      string
	Decimals    uint64
	TotalSupply *big.Int
}

func (this *Oep4) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.Name)
	sink.WriteString(this.Symbol)
	sink.WriteUint64(this.Decimals)
	sink.WriteVarBytes(common.BigIntToNeoBytes(this.TotalSupply))
}

func (this *Oep4) Deserialization(source *common.ZeroCopySource) error {
	var irr, eof bool
	this.Name, _, irr, eof = source.NextString()
	if irr {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.Symbol, _, irr, eof = source.NextString()
	if irr {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.Decimals, eof = source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	supply, _, irr, eof := source.NextVarBytes()
	if irr {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.TotalSupply = common.BigIntFromNeoBytes(supply)
	return nil
}

const (
	XSHARD_TRANSFER_PENDING  uint8 = 0x06
	XSHARD_TRANSFER_COMPLETE uint8 = 0x07
)

type XShardTransferState struct {
	ToShard   common.ShardID `json:"to_shard"`
	ToAccount common.Address `json:"to_account"`
	Amount    *big.Int       `json:"amount"`
	Status    uint8          `json:"status"`
}

func (this *XShardTransferState) Serialization(sink *common.ZeroCopySink) {
	utils.SerializationShardId(sink, this.ToShard)
	sink.WriteAddress(this.ToAccount)
	sink.WriteVarBytes(common.BigIntToNeoBytes(this.Amount))
	sink.WriteUint8(this.Status)
}

func (this *XShardTransferState) Deserialization(source *common.ZeroCopySource) error {
	var err error = nil
	this.ToShard, err = utils.DeserializationShardId(source)
	if err != nil {
		return fmt.Errorf("deserialization: read to shard failed, err: %s", err)
	}
	var irr, eof bool
	this.ToAccount, eof = source.NextAddress()
	if eof {
		return io.ErrUnexpectedEOF
	}
	amount, _, irr, eof := source.NextVarBytes()
	if irr {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.Amount = common.BigIntFromNeoBytes(amount)
	this.Status, eof = source.NextUint8()
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}
