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
package integrationtest

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	types2 "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ontio/ontology/account"
	utils2 "github.com/ontio/ontology/cmd/utils"
	common2 "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func TestTxNonce(t *testing.T) {
	database, acct := NewLedger()
	gasPrice := uint64(500)
	gasLimit := uint64(200000)

	fromPrivateKey, toEthAddr := prepareEthAcct(database, acct, gasPrice, gasLimit)
	nonce := int64(0)
	transferAmt := 0.5 * 1000000000
	checkEvmOngTransfer(database, acct, gasPrice, gasLimit, fromPrivateKey, toEthAddr, int64(transferAmt), nonce)
	checkEvmOngTransfer(database, acct, gasPrice, gasLimit, fromPrivateKey, toEthAddr, int64(transferAmt), nonce+1)
}

func checkEvmOngTransfer(database *ledger.Ledger, acct *account.Account, gasPrice, gasLimit uint64, fromPrivateKey *ecdsa.PrivateKey, toEthAddr common.Address, amt int64, nonce int64) {
	fromEthAddr := crypto.PubkeyToAddress(fromPrivateKey.PublicKey)
	before := ongBalanceOf(database, common2.Address(fromEthAddr))
	tx := evmTransferOng(fromPrivateKey, gasPrice, gasLimit, toEthAddr, nonce, amt)
	genBlock(database, acct, tx)
	evt, err := database.GetEventNotifyByTx(tx.Hash())
	checkErr(err)
	after := ongBalanceOf(database, common2.Address(fromEthAddr))
	var expect uint64
	if evt.State == 1 {
		expect = evt.GasConsumed + uint64(amt)
	} else {
		expect = evt.GasConsumed
	}
	if before-after != expect {
		panic(fmt.Sprintf("before:%d, after:%d, evt.GasConsumed:%d, transferAmt:%d",
			before, after, evt.GasConsumed, amt))
	}
	ethAcc, err := database.GetEthAccount(fromEthAddr)
	checkErr(err)
	if ethAcc.Nonce != uint64(nonce+1) {
		panic(fmt.Sprintf("ethAcc.Nonce:%d, nonce+1:%d", ethAcc.Nonce, nonce+1))
	}
}

func prepareEthAcct(database *ledger.Ledger, acct *account.Account, gasPrice, gasLimit uint64) (*ecdsa.PrivateKey, common.Address) {
	fromPrivateKey, err := crypto.GenerateKey()
	checkErr(err)
	fromEthAddr := crypto.PubkeyToAddress(fromPrivateKey.PublicKey)
	transferOng(database, gasPrice, gasLimit, acct, common2.Address(fromEthAddr), 1*1000000000)

	toPrivateKey, err := crypto.GenerateKey()
	checkErr(err)
	toEthAddr := crypto.PubkeyToAddress(toPrivateKey.PublicKey)
	return fromPrivateKey, toEthAddr
}

func genBlock(database *ledger.Ledger, acct *account.Account, tx *types.Transaction) {
	_, err := database.PreExecuteContract(tx)
	checkErr(err)
	block, _ := makeBlock(acct, []*types.Transaction{tx})
	err = database.AddBlock(block, nil, common2.UINT256_EMPTY)
	checkErr(err)
}

func transferOng(database *ledger.Ledger, gasPrice, gasLimit uint64, acct *account.Account, toAddr common2.Address, amount int64) {
	state := &ont.State{
		From:  acct.Address,
		To:    toAddr,
		Value: uint64(amount),
	}
	mutable := newNativeTx(utils.OngContractAddress, 0, gasPrice, gasLimit, "transfer", []interface{}{[]*ont.State{state}})
	err := utils2.SignTransaction(acct, mutable)
	checkErr(err)
	tx, err := mutable.IntoImmutable()
	checkErr(err)
	genBlock(database, acct, tx)
}

func ongBalanceOf(database *ledger.Ledger, acctAddr common2.Address) uint64 {
	mutable := newNativeTx(utils.OngContractAddress, 0, 0, 0, "balanceOf", []interface{}{acctAddr[:]})
	tx, err := mutable.IntoImmutable()
	checkErr(err)
	res, err := database.PreExecuteContract(tx)
	checkErr(err)
	data, err := hex.DecodeString(res.Result.(string))
	checkErr(err)
	balance := common2.BigIntFromNeoBytes(data)
	return balance.Uint64()
}

func evmTransferOng(testPrivateKey *ecdsa.PrivateKey, gasPrice, gasLimit uint64, toEthAddr common.Address, nonce int64, value int64) *types.Transaction {
	chainId := big.NewInt(int64(config.DefConfig.P2PNode.EVMChainId))
	opts, err := bind.NewKeyedTransactorWithChainID(testPrivateKey, chainId)
	opts.GasPrice = big.NewInt(int64(gasPrice))
	opts.Nonce = big.NewInt(nonce)
	opts.GasLimit = gasLimit
	opts.Value = big.NewInt(value)

	invokeTx := types2.NewTransaction(opts.Nonce.Uint64(), toEthAddr, opts.Value, opts.GasLimit, opts.GasPrice, []byte{})
	signedTx, err := opts.Signer(opts.From, invokeTx)
	checkErr(err)
	tx, err := types.TransactionFromEIP155(signedTx)
	checkErr(err)
	return tx
}
