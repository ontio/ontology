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

package ledgerstore

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/core/utils"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	nutils "github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/stretchr/testify/assert"
)

func TestExtractHeaderHeight(t *testing.T) {
	header := types.Header{
		Version:          uint32(0),
		PrevBlockHash:    common.UINT256_EMPTY,
		TransactionsRoot: common.UINT256_EMPTY,
		BlockRoot:        common.UINT256_EMPTY,
		Timestamp:        uint32(1),
		Height:           uint32(99999),
	}
	sink := common.NewZeroCopySink(nil)
	header.Serialization(sink)
	source := common.NewZeroCopySource(sink.Bytes())
	raw := types.RawHeader{}
	err := raw.Deserialization(source)
	assert.Nil(t, err)
	assert.Equal(t, uint32(99999), raw.Height)
}

func TestVersion(t *testing.T) {
	testBlockStore.NewBatch()
	version := byte(1)
	err := testBlockStore.SaveVersion(version)
	if err != nil {
		t.Errorf("SaveVersion error %s", err)
		return
	}
	err = testBlockStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo error %s", err)
		return
	}
	v, err := testBlockStore.GetVersion()
	if err != nil {
		t.Errorf("GetVersion error %s", err)
		return
	}
	if version != v {
		t.Errorf("TestVersion failed version %d != %d", v, version)
		return
	}
}

func TestCurrentBlock(t *testing.T) {
	blockHash := common.Uint256(sha256.Sum256([]byte("123456789")))
	blockHeight := uint32(1)
	testBlockStore.NewBatch()
	err := testBlockStore.SaveCurrentBlock(blockHeight, blockHash)
	if err != nil {
		t.Errorf("SaveCurrentBlock error %s", err)
		return
	}
	err = testBlockStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo error %s", err)
		return
	}
	hash, height, err := testBlockStore.GetCurrentBlock()
	if hash != blockHash {
		t.Errorf("TestCurrentBlock BlockHash %x != %x", hash, blockHash)
		return
	}
	if height != blockHeight {
		t.Errorf("TestCurrentBlock BlockHeight %x != %x", height, blockHeight)
		return
	}
}

func TestBlockHash(t *testing.T) {
	blockHash := common.Uint256(sha256.Sum256([]byte("123456789")))
	blockHeight := uint32(1)
	testBlockStore.NewBatch()
	testBlockStore.SaveBlockHash(blockHeight, blockHash)
	blockHash = common.Uint256(sha256.Sum256([]byte("234567890")))
	blockHeight = uint32(2)
	testBlockStore.SaveBlockHash(blockHeight, blockHash)
	err := testBlockStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo error %s", err)
		return
	}
	hash, err := testBlockStore.GetBlockHash(blockHeight)
	if err != nil {
		t.Errorf("GetBlockHash error %s", err)
		return
	}
	if hash != blockHash {
		t.Errorf("TestBlockHash failed BlockHash %x != %x", hash, blockHash)
		return
	}
}

func TestSaveTransaction(t *testing.T) {
	invoke := &payload.InvokeCode{}
	txTemp := &types.MutableTransaction{
		TxType:  types.InvokeNeo,
		Payload: invoke,
	}
	tx, err := txTemp.IntoImmutable()
	blockHeight := uint32(1)
	txHash := tx.Hash()

	exist, err := testBlockStore.ContainTransaction(txHash)
	if err != nil {
		t.Errorf("ContainTransaction error %s", err)
		return
	}
	if exist {
		t.Errorf("TestSaveTransaction ContainTransaction should be false.")
		return
	}

	testBlockStore.NewBatch()
	testBlockStore.SaveTransaction(tx, blockHeight)
	err = testBlockStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo error %s", err)
		return
	}

	tx1, height, err := testBlockStore.GetTransaction(txHash)
	if err != nil {
		t.Errorf("GetTransaction error %s", err)
		return
	}
	if blockHeight != height {
		t.Errorf("TestSaveTransaction failed BlockHeight %d != %d", height, blockHeight)
		return
	}
	if tx1.TxType != tx.TxType {
		t.Errorf("TestSaveTransaction failed TxType %d != %d", tx1.TxType, tx.TxType)
		return
	}
	tx1Hash := tx1.Hash()
	if txHash != tx1Hash {
		t.Errorf("TestSaveTransaction failed TxHash %x != %x", tx1Hash, txHash)
		return
	}

	exist, err = testBlockStore.ContainTransaction(txHash)
	if err != nil {
		t.Errorf("ContainTransaction error %s", err)
		return
	}
	if !exist {
		t.Errorf("TestSaveTransaction ContainTransaction should be true.")
		return
	}
}

func TestHeaderIndexList(t *testing.T) {
	testBlockStore.NewBatch()
	startHeight := uint32(0)
	size := uint32(100)
	indexMap := make(map[uint32]common.Uint256, size)
	indexList := make([]common.Uint256, 0)
	for i := startHeight; i < size; i++ {
		hash := common.Uint256(sha256.Sum256([]byte(fmt.Sprintf("%v", i))))
		indexMap[i] = hash
		indexList = append(indexList, hash)
	}
	testBlockStore.SaveHeaderIndexList(startHeight, indexList)
	startHeight = uint32(100)
	size = uint32(100)
	indexMap = make(map[uint32]common.Uint256, size)
	for i := startHeight; i < size; i++ {
		hash := common.Uint256(sha256.Sum256([]byte(fmt.Sprintf("%v", i))))
		indexMap[i] = hash
		indexList = append(indexList, hash)
	}
	err := testBlockStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo error %s", err)
		return
	}

	totalMap, err := testBlockStore.GetHeaderIndexList()
	if err != nil {
		t.Errorf("GetHeaderIndexList error %s", err)
		return
	}

	for height, hash := range indexList {
		h, ok := totalMap[uint32(height)]
		if !ok {
			t.Errorf("TestHeaderIndexList failed height:%d hash not exist", height)
			return
		}
		if hash != h {
			t.Errorf("TestHeaderIndexList failed height:%d hash %x != %x", height, hash, h)
			return
		}
	}
}

func TestSaveHeader(t *testing.T) {
	acc1 := account.NewAccount("")
	acc2 := account.NewAccount("")
	bookkeeper, err := types.AddressFromBookkeepers([]keypair.PublicKey{acc1.PublicKey, acc2.PublicKey})
	if err != nil {
		t.Errorf("AddressFromBookkeepers error %s", err)
		return
	}
	header := &types.Header{
		Version:          123,
		PrevBlockHash:    common.Uint256{},
		TransactionsRoot: common.Uint256{},
		Timestamp:        uint32(uint32(time.Date(2017, time.February, 23, 0, 0, 0, 0, time.UTC).Unix())),
		Height:           uint32(1),
		ConsensusData:    123456789,
		NextBookkeeper:   bookkeeper,
	}
	block := &types.Block{
		Header:       header,
		Transactions: []*types.Transaction{},
	}
	blockHash := block.Hash()
	sysFee := common.Fixed64(1)

	testBlockStore.NewBatch()

	err = testBlockStore.SaveHeader(block, sysFee)
	if err != nil {
		t.Errorf("SaveHeader error %s", err)
		return
	}
	err = testBlockStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo error %s", err)
		return
	}

	h, err := testBlockStore.GetHeader(blockHash)
	if err != nil {
		t.Errorf("GetHeader error %s", err)
		return
	}

	headerHash := h.Hash()
	if blockHash != headerHash {
		t.Errorf("TestSaveHeader failed HeaderHash %x != %x", headerHash, blockHash)
		return
	}

	if header.Height != h.Height {
		t.Errorf("TestSaveHeader failed Height %d != %d", h.Height, header.Height)
		return
	}

	fee, err := testBlockStore.GetSysFeeAmount(blockHash)
	if err != nil {
		t.Errorf("TestSaveHeader SysFee %d != %d", fee, sysFee)
		return
	}
}

func TestBlock(t *testing.T) {
	acc1 := account.NewAccount("")
	acc2 := account.NewAccount("")
	bookkeeper, err := types.AddressFromBookkeepers([]keypair.PublicKey{acc1.PublicKey, acc2.PublicKey})
	if err != nil {
		t.Errorf("AddressFromBookkeepers error %s", err)
		return
	}
	header := &types.Header{
		Version:          123,
		PrevBlockHash:    common.Uint256{},
		TransactionsRoot: common.Uint256{},
		Timestamp:        uint32(uint32(time.Date(2017, time.February, 23, 0, 0, 0, 0, time.UTC).Unix())),
		Height:           uint32(2),
		ConsensusData:    1234567890,
		NextBookkeeper:   bookkeeper,
	}

	tx1, err := transferTx(acc1.Address, acc2.Address, 10)
	if err != nil {
		t.Errorf("TestBlock transferTx error:%s", err)
		return
	}

	block := &types.Block{
		Header:       header,
		Transactions: []*types.Transaction{tx1},
	}
	blockHash := block.Hash()
	tx1Hash := tx1.Hash()

	testBlockStore.NewBatch()

	err = testBlockStore.SaveBlock(block)
	if err != nil {
		t.Errorf("SaveHeader error %s", err)
		return
	}
	err = testBlockStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo error %s", err)
		return
	}

	b, err := testBlockStore.GetBlock(blockHash)
	if err != nil {
		t.Errorf("GetBlock error %s", err)
		return
	}

	hash := b.Hash()
	if hash != blockHash {
		t.Errorf("TestBlock failed BlockHash %x != %x ", hash, blockHash)
		return
	}
	exist, err := testBlockStore.ContainTransaction(tx1Hash)
	if err != nil {
		t.Errorf("ContainTransaction error %s", err)
		return
	}
	if !exist {
		t.Errorf("TestBlock failed transaction %x should exist", tx1Hash)
		return
	}

	if len(block.Transactions) != len(b.Transactions) {
		t.Errorf("TestBlock failed Transaction size %d != %d ", len(b.Transactions), len(block.Transactions))
		return
	}
	if b.Transactions[0].Hash() != tx1Hash {
		t.Errorf("TestBlock failed transaction1 hash %x != %x", b.Transactions[0].Hash(), tx1Hash)
		return
	}
}

func transferTx(from, to common.Address, amount uint64) (*types.Transaction, error) {
	var sts []ont.State
	sts = append(sts, ont.State{
		From:  from,
		To:    to,
		Value: amount,
	})
	var cversion byte
	return invokeSmartContractTx(0, 30000, cversion, nutils.OntContractAddress, "transfer", []interface{}{sts})
}

func invokeSmartContractTx(gasPrice,
	gasLimit uint64,
	cversion byte,
	contractAddress common.Address,
	method string,
	args []interface{}) (*types.Transaction, error) {

	invokCode, err := utils.BuildNativeInvokeCode(contractAddress, cversion, method, args)
	if err != nil {
		return nil, err
	}
	return newInvokeTransaction(gasPrice, gasLimit, invokCode), nil
}

func newInvokeTransaction(gasPirce, gasLimit uint64, code []byte) *types.Transaction {
	invokePayload := &payload.InvokeCode{
		Code: code,
	}
	tx := &types.MutableTransaction{
		Version:  0,
		GasPrice: gasPirce,
		GasLimit: gasLimit,
		TxType:   types.InvokeNeo,
		Nonce:    uint32(time.Now().Unix()),
		Payload:  invokePayload,
		Sigs:     make([]types.Sig, 0, 0),
	}
	res, err := tx.IntoImmutable()
	if err != nil {
		return nil
	}
	return res
}
