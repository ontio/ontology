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
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
)

type Block struct {
	Header       *Header
	Transactions []*Transaction

	hash *common.Uint256
}

func (b *Block) Serialize(w io.Writer) error {
	b.Header.Serialize(w)
	err := serialization.WriteUint32(w, uint32(len(b.Transactions)))
	if err != nil {
		return fmt.Errorf("Block item Transactions length serialization failed: %s", err)
	}

	for _, transaction := range b.Transactions {
		transaction.Serialize(w)
	}
	return nil
}

func (b *Block) Deserialize(r io.Reader) error {
	if b.Header == nil {
		b.Header = new(Header)
	}
	b.Header.Deserialize(r)

	//Transactions
	length, err := serialization.ReadUint32(r)
	if err != nil {
		return err
	}

	var hashes = make([]common.Uint256, 0, length)
	for i := uint32(0); i < length; i++ {
		transaction := new(Transaction)
		transaction.Deserialize(r)
		txhash := transaction.Hash()
		b.Transactions = append(b.Transactions, transaction)
		hashes = append(hashes, txhash)
	}

	b.Header.TransactionsRoot = common.ComputeMerkleRoot(hashes)

	return nil
}

func (b *Block) Trim(w io.Writer) error {
	b.Header.Serialize(w)
	err := serialization.WriteUint32(w, uint32(len(b.Transactions)))
	if err != nil {
		return fmt.Errorf("Block item Transactions length serialization failed: %s", err)
	}
	for _, transaction := range b.Transactions {
		temp := *transaction
		hash := temp.Hash()
		hash.Serialize(w)
	}
	return nil
}

func (b *Block) FromTrimmedData(r io.Reader) error {
	if b.Header == nil {
		b.Header = new(Header)
	}
	b.Header.Deserialize(r)

	//Transactions
	var i uint32
	Len, err := serialization.ReadUint32(r)
	if err != nil {
		return err
	}
	var txhash common.Uint256
	var hashes []common.Uint256
	for i = 0; i < Len; i++ {
		txhash.Deserialize(r)
		transaction := new(Transaction)
		transaction.SetHash(txhash)
		b.Transactions = append(b.Transactions, transaction)
		hashes = append(hashes, txhash)
	}

	b.Header.TransactionsRoot = common.ComputeMerkleRoot(hashes)

	return nil
}

func (b *Block) ToArray() []byte {
	bf := new(bytes.Buffer)
	b.Serialize(bf)
	return bf.Bytes()
}

func (b *Block) Hash() common.Uint256 {
	if b.hash == nil {
		b.hash = new(common.Uint256)
		*b.hash = b.Header.Hash()
	}
	return *b.hash
}

func (b *Block) Type() common.InventoryType {
	return common.BLOCK
}

func (b *Block) RebuildMerkleRoot() error {
	txs := b.Transactions
	hashes := []common.Uint256{}
	for _, tx := range txs {
		hashes = append(hashes, tx.Hash())
	}
	hash := common.ComputeMerkleRoot(hashes)
	b.Header.TransactionsRoot = hash
	return nil
}

func (bd *Block) SerializeUnsigned(w io.Writer) error {
	return bd.Header.SerializeUnsigned(w)
}
