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

package common

import (
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"testing"
	"time"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
	"github.com/stretchr/testify/assert"

	ethcomm "github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	sysconfig "github.com/ontio/ontology/common/config"
	txtypes "github.com/ontio/ontology/core/types"
)

var (
	txn *types.Transaction
)

func init() {
	log.InitLog(log.InfoLog, log.Stdout)

	mutable := &types.MutableTransaction{
		TxType:  types.InvokeNeo,
		Nonce:   uint32(time.Now().Unix()),
		Payload: &payload.InvokeCode{Code: []byte{}},
	}

	txn, _ = mutable.IntoImmutable()
}

func TestTxPool(t *testing.T) {
	txPool := NewTxPool()

	txEntry := &VerifiedTx{
		Tx:             txn,
		VerifiedHeight: 10,
	}

	ret := txPool.AddTxList(txEntry)
	if !ret.Success() {
		t.Error("Failed to add tx to the pool")
		return
	}

	ret = txPool.AddTxList(txEntry)
	if ret.Success() {
		t.Error("Failed to add tx to the pool")
		return
	}

	txList, oldTxList := txPool.GetTxPool(true, 0)
	for _, v := range txList {
		assert.NotNil(t, v)
	}

	for _, v := range oldTxList {
		assert.NotNil(t, v)
	}

	entry := txPool.GetTransaction(txn.Hash())
	if entry == nil {
		t.Error("Failed to get the transaction")
		return
	}

	assert.Equal(t, txn.Hash(), entry.Hash())

	status := txPool.GetTxStatus(txn.Hash())
	if status == nil {
		t.Error("failed to get the status")
		return
	}

	assert.Equal(t, txn.Hash(), status.Hash)

	count := txPool.GetTransactionCount()
	assert.Equal(t, count, 1)

	txPool.CleanCompletedTransactionList([]*types.Transaction{txn}, 0)
}

func genTxWithNonceAndPrice(nonce uint64, gp int64) *txtypes.Transaction {
	privateKey, _ := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19")

	value := big.NewInt(1000000000)
	gaslimit := uint64(21000)
	gasPrice := big.NewInt(gp)

	toAddress := ethcomm.HexToAddress("0x4592d8f8d7b001e72cb26a73e4fa1806a51ac79d")

	var data []byte
	tx := etypes.NewTransaction(nonce, toAddress, value, gaslimit, gasPrice, data)

	chainId := big.NewInt(int64(sysconfig.DefConfig.P2PNode.EVMChainId))
	signedTx, err := etypes.SignTx(tx, etypes.NewEIP155Signer(chainId), privateKey)
	if err != nil {
		return nil
	}

	otx, err := txtypes.TransactionFromEIP155(signedTx)
	if err != nil {
		//fmt.Printf("err:%s\n", err.Error())
		return nil
	}
	return otx
}

func TestTXPool_NextNonce(t *testing.T) {
	txPool := NewTxPool()
	addr := genTxWithNonceAndPrice(uint64(0), 2500).Payer
	ledgerNonce := uint64(0)
	//empty case nonce is 0
	assert.Equal(t, txPool.NextNonce(addr, ledgerNonce), uint64(0))

	//no 0 nonce case , nonce is 0
	tx := genTxWithNonceAndPrice(uint64(2), 2500)
	txn := &VerifiedTx{
		Tx:             tx,
		VerifiedHeight: 100,
		Nonce:          uint64(105),
	}
	txPool.AddTxList(txn)
	assert.Equal(t, txPool.NextNonce(addr, ledgerNonce), uint64(0))

	//consecutive case ,nonce is the last + 1
	for i := 0; i < 100; i++ {
		tx := genTxWithNonceAndPrice(uint64(i), 2500)
		txn := &VerifiedTx{
			Tx:             tx,
			VerifiedHeight: 100,
			Nonce:          uint64(i),
		}

		txPool.AddTxList(txn)
	}
	assert.Equal(t, txPool.NextNonce(addr, ledgerNonce), uint64(100))

	//no consecutive case nonce is still 100
	tx = genTxWithNonceAndPrice(uint64(105), 2500)
	txn = &VerifiedTx{
		Tx:             tx,
		VerifiedHeight: 100,
		Nonce:          uint64(105),
	}
	txPool.AddTxList(txn)
	assert.Equal(t, txPool.NextNonce(addr, ledgerNonce), uint64(100))

}
