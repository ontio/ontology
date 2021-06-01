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
package proc

import (
	"crypto/ecdsa"
	"fmt"
	ethcomm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ontio/ontology/common"
	txtypes "github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	tc "github.com/ontio/ontology/txnpool/common"
	vt "github.com/ontio/ontology/validator/types"
	"github.com/stretchr/testify/assert"

	"math/big"
	"testing"
	"time"
)

func GenTx(nonce uint64) *txtypes.Transaction {
	privateKey, _ := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19")
	//assert.Nil(t, err)

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		//assert.True(t, ok)
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	fmt.Printf("addr:%s\n", fromAddress.Hex())

	ontAddress, _ := common.AddressParseFromBytes(fromAddress[:])
	//assert.Nil(t, err)
	fmt.Printf("ont addr:%s\n", ontAddress.ToBase58())

	value := big.NewInt(1000000000)
	gaslimit := uint64(21000)
	gasPrice := big.NewInt(2500)

	toAddress := ethcomm.HexToAddress("0x4592d8f8d7b001e72cb26a73e4fa1806a51ac79d")

	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gaslimit, gasPrice, data)

	chainId := big.NewInt(0)
	signedTx, _ := types.SignTx(tx, types.NewEIP155Signer(chainId), privateKey)
	//assert.Nil(t, err)

	otx, _ := txtypes.TransactionFromEIP155(signedTx)
	//assert.Nil(t, err)
	return otx
}

func Test_GenEIP155tx(t *testing.T) {
	otx := GenTx(0)

	fmt.Printf("1. otx.payer:%s\n", otx.Payer.ToBase58())

	assert.True(t, otx.TxType == txtypes.EIP155)

	t.Log("Starting test tx")
	var s *TXPoolServer
	s = NewTxPoolServer(tc.MAX_WORKER_NUM, true, false)
	if s == nil {
		t.Error("Test case: new tx pool server failed")
		return
	}
	defer s.Stop()

	f := s.assignTxToWorker(otx, sender, nil)
	assert.True(t, f)

	tmptx := s.pendingEipTxs[otx.Payer].txs.Get(0)
	assert.True(t, tmptx.TxType == txtypes.EIP155 && tmptx.Nonce == 0)

	time.Sleep(10 * time.Second)
	txEntry := &tc.TXEntry{
		Tx:    otx,
		Attrs: []*tc.TXAttr{},
	}
	//fmt.Printf("before %s nonce is :%d\n",otx.Payer.ToBase58(),s.pendingNonces.get(otx.Payer))
	f = s.addTxList(txEntry)
	assert.True(t, f)
	//fmt.Printf("after %s nonce is :%d\n",otx.Payer.ToBase58(),s.pendingNonces.get(otx.Payer))

	tmptx2 := s.eipTxPool[otx.Payer].txs.Get(0)
	assert.True(t, tmptx2.TxType == txtypes.EIP155 && tmptx2.Nonce == 0)
	//tmptx = s.pendingEipTxs[otx.Payer].txs.Get(0)
	//assert.Nil(t,tmptx)

	ret := s.checkTx(otx.Hash())
	if ret == false {
		t.Error("Failed to check the tx")
		return
	}

	entry := s.getTransaction(otx.Hash())
	if entry == nil {
		t.Error("Failed to get the transaction")
		return
	}

	pendingNonce := s.pendingNonces.get(otx.Payer)
	assert.Equal(t, pendingNonce, uint64(1))

	t.Log("Ending test tx")

}

func Test_AssignRsp2Worker(t *testing.T) {
	t.Log("Starting assign response to the worker testing")
	var s *TXPoolServer
	s = NewTxPoolServer(tc.MAX_WORKER_NUM, true, false)
	if s == nil {
		t.Error("Test case: new tx pool server failed")
		return
	}

	defer s.Stop()

	s.assignRspToWorker(nil)

	statelessRsp := &vt.CheckResponse{
		WorkerId: 0,
		ErrCode:  errors.ErrNoError,
		Hash:     txn.Hash(),
		Type:     vt.Stateless,
		Height:   0,
	}

	statefulRsp := &vt.CheckResponse{
		WorkerId: 0,
		ErrCode:  errors.ErrUnknown,
		Hash:     txn.Hash(),
		Type:     vt.Stateful,
		Height:   0,
	}
	s.assignRspToWorker(statelessRsp)
	s.assignRspToWorker(statefulRsp)

	statelessRsp = &vt.CheckResponse{
		WorkerId: 0,
		ErrCode:  errors.ErrUnknown,
		Hash:     txn.Hash(),
		Type:     vt.Stateless,
		Height:   0,
	}
	s.assignRspToWorker(statelessRsp)

	t.Log("Ending assign response to the worker testing")
}
