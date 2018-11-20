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
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
)

type Block struct {
	Header       *Header
	Transactions []*Transaction
}

func (b *Block) Serialize(w io.Writer) error {
	err := b.Header.Serialize(w)
	if err != nil {
		return err
	}

	err = serialization.WriteUint32(w, uint32(len(b.Transactions)))
	if err != nil {
		return fmt.Errorf("Block item Transactions length serialization failed: %s", err)
	}

	for _, transaction := range b.Transactions {
		err := transaction.Serialize(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Block) Serialization(sink *common.ZeroCopySink) error {
	err := b.Header.Serialization(sink)
	if err != nil {
		return err
	}

	sink.WriteUint32(uint32(len(b.Transactions)))
	for _, transaction := range b.Transactions {
		err := transaction.Serialization(sink)
		if err != nil {
			return err
		}
	}
	return nil
}

// if no error, ownership of param raw is transfered to Transaction
func BlockFromRawBytes(raw []byte) (*Block, error) {
	source := common.NewZeroCopySource(raw)
	block := &Block{}
	err := block.Deserialization(source)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (self *Block) Deserialization(source *common.ZeroCopySource) error {
	if self.Header == nil {
		self.Header = new(Header)
	}
	err := self.Header.Deserialization(source)
	if err != nil {
		return err
	}

	length, eof := source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}

	var hashes []common.Uint256
	mask := make(map[common.Uint256]bool)
	for i := uint32(0); i < length; i++ {
		transaction := new(Transaction)
		// note currently all transaction in the block shared the same source
		err := transaction.Deserialization(source)
		if err != nil {
			return err
		}
		txhash := transaction.Hash()
		if mask[txhash] {
			return errors.New("duplicated transaction in block")
		}
		mask[txhash] = true
		hashes = append(hashes, txhash)
		self.Transactions = append(self.Transactions, transaction)
	}

	root := common.ComputeMerkleRoot(hashes)
	if self.Header.TransactionsRoot != root {
		return errors.New("mismatched transaction root")
	}

	return nil
}

func (b *Block) ToArray() []byte {
	bf := new(bytes.Buffer)
	b.Serialize(bf)
	return bf.Bytes()
}

func (b *Block) Hash() common.Uint256 {
	return b.Header.Hash()
}

func (b *Block) Type() common.InventoryType {
	return common.BLOCK
}

func (b *Block) RebuildMerkleRoot() {
	txs := b.Transactions
	hashes := make([]common.Uint256, 0, len(txs))
	for _, tx := range txs {
		hashes = append(hashes, tx.Hash())
	}
	hash := common.ComputeMerkleRoot(hashes)
	b.Header.TransactionsRoot = hash
}
