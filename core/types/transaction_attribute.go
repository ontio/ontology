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

package types

import (
	"errors"
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
)

type TransactionAttributeUsage byte

const (
	Nonce          TransactionAttributeUsage = 0x00
	Script         TransactionAttributeUsage = 0x20
	DescriptionUrl TransactionAttributeUsage = 0x81
	Description    TransactionAttributeUsage = 0x90
)

func IsValidAttributeType(usage TransactionAttributeUsage) bool {
	return usage == Nonce || usage == Script ||
		usage == DescriptionUrl || usage == Description
}

type TxAttribute struct {
	Usage TransactionAttributeUsage
	Data  []byte
	Size  uint32
}

func NewTxAttribute(u TransactionAttributeUsage, d []byte) TxAttribute {
	tx := TxAttribute{u, d, 0}
	tx.Size = tx.GetSize()
	return tx
}

func (u *TxAttribute) GetSize() uint32 {
	if u.Usage == DescriptionUrl {
		return uint32(len([]byte{byte(0xff)}) + len([]byte{byte(0xff)}) + len(u.Data))
	}
	return 0
}

func (tx *TxAttribute) Serialization(sink *common.ZeroCopySink) error {
	if !IsValidAttributeType(tx.Usage) {
		return errors.New("Unsupported attribute Description.")
	}
	sink.WriteUint8(byte(tx.Usage))
	sink.WriteVarBytes(tx.Data)
	return nil
}

func (tx *TxAttribute) Deserialization(source *common.ZeroCopySource) error {
	val, eof := source.NextBytes(1)
	if eof {
		return fmt.Errorf("Transaction attribute Usage deserialization error: %s", io.ErrUnexpectedEOF)
	}
	tx.Usage = TransactionAttributeUsage(val[0])
	if !IsValidAttributeType(tx.Usage) {
		return errors.New("[TxAttribute] Unsupported attribute Description.")
	}
	var irregular bool
	tx.Data, _, irregular, eof = source.NextVarBytes()
	if irregular {
		return fmt.Errorf("Transaction attribute Data deserialization error: %s", common.ErrIrregularData)
	}
	if eof {
		return fmt.Errorf("Transaction attribute Data deserialization error: %s", io.ErrUnexpectedEOF)
	}
	return nil

}

func (tx *TxAttribute) ToArray() []byte {
	sink := common.NewZeroCopySink(nil)
	tx.Serialization(sink)
	return sink.Bytes()
}
