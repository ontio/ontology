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
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func genTestHeader() *Header {
	header := new(Header)
	header.Height = 321
	header.Bookkeepers, _ = config.DefConfig.GetBookkeepers()
	header.SigData = make([][]byte, 0)
	header.SigData = append(header.SigData,
		[]byte("1202028541d32f3b09180b00affe67a40516846c16663ccb916fd2db8106619f087527"))
	header.SigData = append(header.SigData,
		[]byte("120202dfb161f757921898ec2e30e3618d5c6646d993153b89312bac36d7688912c0ce"))
	header.SigData = append(header.SigData,
		[]byte("1202039dab38326268fe82fb7967fe2e7f5f6eaced6ec711148a66fbb8480c321c19dd"))
	return header
}

func genTestBlock() (*Block, error) {
	block := new(Block)
	block.Header = genTestHeader()
	block.Transactions = make([]*Transaction, 0)
	testTx , err := genTestTx(Invoke)
	block.Transactions = append(block.Transactions, testTx)
	block.RebuildMerkleRoot()
	return block, err
}

func TestBlock_Serialize_Deserialize(t *testing.T) {
	testBlock, err := genTestBlock()
	assert.Nil(t, err)
	bf := new(bytes.Buffer)
	err = testBlock.Serialize(bf)
	assert.Nil(t, err)

	deserializeBlock := new(Block)
	err = deserializeBlock.Deserialize(bf)
	deserializeBlock.RebuildMerkleRoot()
	assert.Nil(t, err)

	assert.Equal(t, deserializeBlock, testBlock)
}

func TestBlockFunc(t *testing.T) {
	testBlock, err := genTestBlock()
	assert.Nil(t, err)
	bf := new(bytes.Buffer)
	err = testBlock.Trim(bf)
	assert.Nil(t, err)

	trimBlock := new(Block)
	err = trimBlock.FromTrimmedData(bf)
	assert.Nil(t, err)

	assert.Equal(t, testBlock.Header.TransactionsRoot, trimBlock.Header.TransactionsRoot)

	if len(testBlock.ToArray()) < 0 {
		t.Fatal("block to array test failed")
	}

	if len(testBlock.Hash()) < 0 {
		t.Fatal("block to hash test failed")
	}
	assert.Equal(t, testBlock.Hash(), testBlock.Header.Hash())

	assert.Equal(t, testBlock.Type(), common.BLOCK)
}
