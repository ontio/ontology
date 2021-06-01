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
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ontio/ontology/common"
)

type EIP155Code struct {
	EIPTx *types.Transaction
}

func (self *EIP155Code) Deserialization(source *common.ZeroCopySource) error {
	code, err := source.ReadVarBytes()
	if err != nil {
		return err
	}
	tx := new(types.Transaction)
	err = rlp.DecodeBytes(code, tx)
	if err != nil {
		return err
	}

	self.EIPTx = tx
	return nil
}

func (self *EIP155Code) Serialization(sink *common.ZeroCopySink) {
	bts, err := rlp.EncodeToBytes(self.EIPTx)
	if err != nil {
		panic(err)
	}
	sink.WriteVarBytes(bts)
}
