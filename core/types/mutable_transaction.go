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

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/payload"
)

type MutableTransaction struct {
	Version      byte
	TxType       TransactionType
	Nonce        uint32
	GasPrice     uint64
	GasLimit     uint64
	Payer        common.Address
	Payload      Payload
	ExtraSigners []common.Address
	Sigs         []Sig
}

// output has no reference to self
func (self *MutableTransaction) IntoImmutable() (*Transaction, error) {
	sink := common.NewZeroCopySink(nil)
	err := self.serialize(sink)
	if err != nil {
		return nil, err
	}

	return TransactionFromRawBytes(sink.Bytes())
}

func (self *MutableTransaction) Hash() common.Uint256 {
	tx, err := self.IntoImmutable()
	if err != nil {
		return common.UINT256_EMPTY
	}
	return tx.Hash()
}

// Serialize the Transaction
func (tx *MutableTransaction) serialize(sink *common.ZeroCopySink) error {
	err := tx.serializeUnsigned(sink)
	if err != nil {
		return err
	}

	sink.WriteVarUint(uint64(len(tx.Sigs)))
	for _, sig := range tx.Sigs {
		err = sig.Serialization(sink)
		if err != nil {
			return err
		}
	}

	return nil
}

func (tx *MutableTransaction) serializeUnsigned(sink *common.ZeroCopySink) error {
	sink.WriteByte(byte(tx.Version))
	sink.WriteByte(byte(tx.TxType))
	sink.WriteUint32(tx.Nonce)
	sink.WriteUint64(tx.GasPrice)
	sink.WriteUint64(tx.GasLimit)
	sink.WriteBytes(tx.Payer[:])

	//Payload
	if tx.Payload == nil {
		return errors.New("transaction payload is nil")
	}
	switch pl := tx.Payload.(type) {
	case *payload.DeployCode:
		pl.Serialization(sink)
	case *payload.InvokeCode:
		pl.Serialization(sink)
	default:
		return errors.New("wrong transaction payload type")
	}
	sink.WriteVarUint(uint64(len(tx.ExtraSigners)))
	for _, addr := range tx.ExtraSigners {
		sink.WriteAddress(addr)
	}

	return nil
}
