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
	"encoding/hex"
	"fmt"
	ethcomm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ontio/ontology/common"
	sysconfig "github.com/ontio/ontology/common/config"
	txtypes "github.com/ontio/ontology/core/types"
	tc "github.com/ontio/ontology/txnpool/common"
	"github.com/stretchr/testify/assert"

	"math/big"
	"testing"
	"time"
)

func initCfg() {
	sysconfig.DefConfig.P2PNode.EVMChainId = 5851
}

func genTxWithNonceAndPrice(nonce uint64, gp int64) *txtypes.Transaction {
	privateKey, _ := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19")
	//assert.Nil(t, err)

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	fmt.Printf("addr:%s\n", fromAddress.Hex())

	ontAddress, _ := common.AddressParseFromBytes(fromAddress[:])
	fmt.Printf("ont addr:%s\n", ontAddress.ToBase58())

	value := big.NewInt(1000000000)
	gaslimit := uint64(21000)
	gasPrice := big.NewInt(gp)

	toAddress := ethcomm.HexToAddress("0x4592d8f8d7b001e72cb26a73e4fa1806a51ac79d")
	toontAddress, _ := common.AddressParseFromBytes(toAddress[:])
	fmt.Printf("to ont addr:%s\n", toontAddress.ToBase58())

	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gaslimit, gasPrice, data)

	chainId := big.NewInt(int64(sysconfig.DefConfig.P2PNode.EVMChainId))
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainId), privateKey)
	if err != nil {
		fmt.Printf("err:%s\n", err.Error())
		return nil
	}

	bt, _ := rlp.EncodeToBytes(signedTx)
	fmt.Printf("rlptx:%s", hex.EncodeToString(bt))

	otx, err := txtypes.TransactionFromEIP155(signedTx)
	if err != nil {
		fmt.Printf("err:%s\n", err.Error())
		return nil
	}
	return otx
}

func Test_ethtxRLP(t *testing.T) {
	initCfg()
	genTxWithNonceAndPrice(0, 2500)

}

func Test_From(t *testing.T) {
	initCfg()
	otx1 := genTxWithNonceAndPrice(0, 2500)
	//fmt.Println(otx.Payer)
	otx2 := genTxWithNonceAndPrice(0, 2500)
	//fmt.Println(otx.Payer)
	otx3 := genTxWithNonceAndPrice(0, 3000)
	//fmt.Println(otx.Payer)
	otx4 := genTxWithNonceAndPrice(1, 2500)
	assert.Equal(t, otx1.Payer, otx2.Payer)
	assert.Equal(t, otx2.Payer, otx3.Payer)
	assert.Equal(t, otx3.Payer, otx4.Payer)
	assert.Equal(t, otx1.Payer, otx4.Payer)

}

func Test_GenEIP155tx(t *testing.T) {
	initCfg()

	otx := genTxWithNonceAndPrice(0, 2500)

	fmt.Printf("1. otx.payer:%s\n", otx.Payer.ToBase58())

	assert.True(t, otx.TxType == txtypes.EIP155)

	t.Log("Starting test tx")
	var s *TXPoolServer
	s = NewTxPoolServer(true, false)
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

func Test_higherNonce(t *testing.T) {
	initCfg()
	otx1 := genTxWithNonceAndPrice(0, 2500)
	var s *TXPoolServer
	s = NewTxPoolServer(true, false)
	if s == nil {
		t.Error("Test case: new tx pool server failed")
		return
	}
	defer s.Stop()

	f := s.assignTxToWorker(otx1, sender, nil)
	assert.True(t, f)

	tmptx := s.pendingEipTxs[otx1.Payer].txs.Get(0)
	assert.True(t, tmptx.GasPrice == 2500)

	otx2 := genTxWithNonceAndPrice(1, 2500)
	assert.Equal(t, otx1.Payer, otx2.Payer)
	f = s.assignTxToWorker(otx2, sender, nil)
	assert.True(t, f)

	tmptx = s.pendingEipTxs[otx1.Payer].txs.Get(1)
	assert.True(t, tmptx.GasPrice == 2500 && tmptx.Nonce == 1)

	otx3 := genTxWithNonceAndPrice(0, 3000)
	f = s.assignTxToWorker(otx3, sender, nil)
	assert.True(t, f)

	tmptx = s.pendingEipTxs[otx1.Payer].txs.Get(0)
	assert.True(t, tmptx.GasPrice == 3000 && tmptx.Nonce == 0)

	time.Sleep(10 * time.Second)
	txEntry1 := &tc.TXEntry{
		Tx:    otx1,
		Attrs: []*tc.TXAttr{},
	}
	//fmt.Printf("before %s nonce is :%d\n",otx.Payer.ToBase58(),s.pendingNonces.get(otx.Payer))
	f = s.addTxList(txEntry1)
	assert.True(t, f)
	tmptx1 := s.eipTxPool[otx1.Payer].txs.Get(0)
	assert.True(t, tmptx1.GasPrice == 2500)

	time.Sleep(10 * time.Second)
	txEntry2 := &tc.TXEntry{
		Tx:    otx2,
		Attrs: []*tc.TXAttr{},
	}
	//fmt.Printf("before %s nonce is :%d\n",otx.Payer.ToBase58(),s.pendingNonces.get(otx.Payer))
	f = s.addTxList(txEntry2)
	assert.True(t, f)
	tmptx2 := s.eipTxPool[otx1.Payer].txs.Get(1)
	assert.True(t, tmptx2.GasPrice == 2500)

	time.Sleep(10 * time.Second)
	txEntry3 := &tc.TXEntry{
		Tx:    otx3,
		Attrs: []*tc.TXAttr{},
	}
	//fmt.Printf("before %s nonce is :%d\n",otx.Payer.ToBase58(),s.pendingNonces.get(otx.Payer))
	f = s.addTxList(txEntry3)
	assert.True(t, f)
	tmptx3 := s.eipTxPool[otx1.Payer].txs.Get(0)
	assert.True(t, tmptx3.GasPrice == 3000)
	assert.Nil(t, s.txPool.GetTransaction(otx1.Hash()))
}

//
