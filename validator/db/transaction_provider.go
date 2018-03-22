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

package db

import (
	"github.com/Ontology/common"
	"github.com/Ontology/common/serialization"
	"github.com/Ontology/core/types"

	"io"
)

type TransactionMeta struct {
	BlockHeight uint32
	Spend       *FixedBitMap
}

func NewTransactionMeta(height uint32, outputs uint32) TransactionMeta {
	return TransactionMeta{
		BlockHeight: height,
		Spend:       NewFixedBitMap(outputs),
	}
}

func (self *TransactionMeta) DenoteSpent(index uint32) {
	self.Spend.Set(index)
}

func (self *TransactionMeta) DenoteUnspent(index uint32) {
	self.Spend.Unset(index)
}

func (self *TransactionMeta) Height() uint32 {
	return self.BlockHeight
}
func (self *TransactionMeta) IsSpent(idx uint32) bool {
	return self.Spend.Get(idx)
}

func (self *TransactionMeta) IsFullSpent() bool {
	return self.Spend.IsFullSet()
}

func (self *TransactionMeta) Serialize(w io.Writer) error {
	err := serialization.WriteUint32(w, self.BlockHeight)
	if err != nil {
		return err
	}

	err = self.Spend.Serialize(w)

	return err
}

func (self *TransactionMeta) Deserialize(r io.Reader) error {
	height, err := serialization.ReadUint32(r)
	if err != nil {
		return err
	}
	self.BlockHeight = height
	self.Spend = &FixedBitMap{}
	err = self.Spend.Deserialize(r)
	return err
}

type TransactionProvider interface {
	BestStateProvider
	ContainTransaction(hash common.Uint256) bool
	GetTransactionBytes(hash common.Uint256) ([]byte, error)
	GetTransaction(hash common.Uint256) (*types.Transaction, error)
	PersistBlock(block *types.Block) error
}

type TransactionMetaProvider interface {
	BestStateProvider
	GetTransactionMeta(hash common.Uint256) (TransactionMeta, error)
}
