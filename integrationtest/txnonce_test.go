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
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"
	"strings"
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

const WingABI = "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"subtractedValue\",\"type\":\"uint256\"}],\"name\":\"decreaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"addedValue\",\"type\":\"uint256\"}],\"name\":\"increaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"mint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"burn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

const testContractDir = "./test-contract"

// Mainly test several scenarios
// 1. not enough ong for transfer, Will transaction nonce be updated
// 2. not enough ong for deploy contract, Will transaction nonce be updated
// 3. check ong transfer event, there should be two events of ong transfer in ong transfer transaction,
// others, there should be only one event for ong transfer.
func TestTxNonce(t *testing.T) {
	database, acct := NewLedger()
	gasPrice := uint64(500)
	gasLimit := uint64(200000)

	fromPrivateKey, toEthAddr := prepareEthAcct(database, acct, gasPrice, gasLimit)
	nonce := int64(0)
	transferAmt := 0.5 * 1000000000
	// enough ong for fee, enough ong for transfer
	checkEvmOngTransfer(database, acct, gasPrice, gasLimit, fromPrivateKey, toEthAddr, int64(transferAmt), nonce)

	// enough ong for fee, not enough ong for transfer
	checkEvmOngTransfer(database, acct, gasPrice, gasLimit, fromPrivateKey, toEthAddr, int64(transferAmt), nonce+1)

	// not enough ong for deploy evm contract
	checkDeployEvmContract(database, acct, gasPrice, gasLimit, int64(gasPrice*gasLimit)-1)

	// check ong transfer event
	//checkOngTransferEvent(database, acct)
}

func checkDeployEvmContract(database *ledger.Ledger, acct *account.Account, gasPrice, gasLimit uint64, ongAmt int64) {
	privateKey, err := crypto.GenerateKey()
	checkErr(err)
	ethAddr := crypto.PubkeyToAddress(privateKey.PublicKey)
	transferOng(database, gasPrice, gasLimit, acct, common2.Address(ethAddr), ongAmt)
	code := loadContract(testContractDir + "/wing_eth.evm")
	nonce := int64(0)
	evmTx := NewDeployEvmContract(privateKey, nonce, gasPrice, gasLimit, int64(gasPrice*gasLimit), code, WingABI)
	ontTx, err := types.TransactionFromEIP155(evmTx)
	checkErr(err)
	genBlock(database, acct, ontTx)
	acc, err := database.GetEthAccount(ethAddr)
	checkErr(err)
	if acc.Nonce != uint64(nonce+1) {
		panic(fmt.Sprintf("acc.Nonce: %d, nonce+1: %d", acc.Nonce, nonce+1))
	}
}

func checkOngTransferEvent(database *ledger.Ledger, acct *account.Account) {
	fromPrivateKey, toEthAddr := prepareEthAcct(database, acct, 500, 200000)
	tx := evmTransferOng(fromPrivateKey, 500, 200000, toEthAddr, 0, 1000000000)
	genBlock(database, acct, tx)
	fromEthAddr := crypto.PubkeyToAddress(fromPrivateKey.PublicKey)
	evt, err := database.GetEventNotifyByTx(tx.Hash())
	checkErr(err)
	parsed, err := abi.JSON(strings.NewReader(WingABI))
	checkErr(err)
	TransferID := common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
	if evt != nil {
		for _, noti := range evt.Notify {
			if noti.ContractAddress == utils.OngContractAddress {
				states, ok := noti.States.(string)
				if ok {
					data, err := hexutil.Decode(states)
					checkErr(err)
					source := common2.NewZeroCopySource(data)
					var storageLog types.StorageLog
					err = storageLog.Deserialization(source)
					checkErr(err)
					parsed.Unpack("Transfer", storageLog.Data)
					
					if bytes.Compare(storageLog.Topics[0][:], TransferID[:]) == 0 {
						panic("invalid TransferID")
					}
					fromHash := common.BytesToHash(fromEthAddr[:])
					if bytes.Compare(storageLog.Topics[1][:], fromHash[:]) == 0 {
						panic("invalid from address")
					}
					toHash := common.BytesToHash(toEthAddr[:])
					if bytes.Compare(storageLog.Topics[2][:], toHash[:]) == 0 {
						panic("invalid to address")
					}
				}
			}
		}
	}
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
