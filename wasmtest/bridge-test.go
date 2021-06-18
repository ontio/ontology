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
package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	common4 "github.com/ethereum/go-ethereum/common"
	types2 "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
	common3 "github.com/ontio/ontology/wasmtest/common"
)

const WingABI = "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"subtractedValue\",\"type\":\"uint256\"}],\"name\":\"decreaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"addedValue\",\"type\":\"uint256\"}],\"name\":\"increaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"mint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"burn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

var testEthAddr common4.Address
var testPrivateKey *ecdsa.PrivateKey

func getBridgeContract(contract []Item) (bridge common3.ConAddr, wingErc20 common3.ConAddr, wingOep4 common3.ConAddr) {
	for _, item := range contract {
		file := item.File
		code := item.Contract
		if strings.HasSuffix(file, "wing_eth.evm") {
			ethAddr := crypto.CreateAddress(testEthAddr, 0)
			addr, _ := common.AddressParseFromBytes(ethAddr.Bytes())
			wingErc20 = common3.ConAddr{
				File:    file,
				Address: addr,
			}
			log.Infof("wingErc20 token address: %s", wingErc20.Address.ToHexString())
		} else if strings.HasSuffix(file, "bridge.avm") {
			bridge = common3.ConAddr{
				File:    file,
				Address: common.AddressFromVmCode(code),
			}
			log.Infof("bridge address: %s", bridge.Address.ToHexString())
		} else if strings.HasSuffix(file, "WingToken.avm") {
			wingOep4 = common3.ConAddr{
				File:    file,
				Address: common.AddressFromVmCode(code),
			}
			log.Infof("wingOep4 address: %s", wingOep4.Address.ToHexString())
		} else {
			continue
		}
	}
	return
}

func deployContract(contract []Item, acct *account.Account, database *ledger.Ledger) {
	txes := make([]*types.Transaction, 0, len(contract))
	nonce := int64(0)
	for _, item := range contract {
		file := item.File
		cont := item.Contract
		var tx *types.Transaction
		var err error
		if strings.HasSuffix(file, ".wasm") {
			tx, err = NewDeployWasmContract(acct, cont)
		} else if strings.HasSuffix(file, ".avm") {
			if file == "bridge2.avm" {
				// migrate contract
				continue
			}
			tx, err = NewDeployNeoContract(acct, cont)
		} else if strings.HasSuffix(file, "wing_eth.evm") {
			chainId := big.NewInt(int64(config.DefConfig.P2PNode.EVMChainId))
			testPrivateKeyStr := "59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d"
			testPrivateKey, err = crypto.HexToECDSA(testPrivateKeyStr)
			checkErr(err)
			testEthAddr = crypto.PubkeyToAddress(testPrivateKey.PublicKey)
			opts, err := bind.NewKeyedTransactorWithChainID(testPrivateKey, chainId)
			opts.GasPrice = big.NewInt(0)
			opts.Nonce = big.NewInt(nonce)
			opts.GasLimit = 8000000
			checkErr(err)
			ethtx, err := NewDeployEvmContract(opts, cont, WingABI)
			checkErr(err)
			tx, err = types.TransactionFromEIP155(ethtx)
			checkErr(err)
			_, err = tx.GetEIP155Tx()
			checkErr(err)
			nonce++
		}
		checkErr(err)
		_, err = database.PreExecuteContract(tx)
		//log.Infof("deploy %s consume gas: %d, %s", file, res.Gas, JsonString(res))
		checkErr(err)
		txes = append(txes, tx)
	}
	block, err := makeBlock(acct, txes)
	checkErr(err)
	err = database.AddBlock(block, nil, common.UINT256_EMPTY)
	checkErr(err)
}

func migrateBridge(bridge common3.ConAddr, newCode []byte, admin common.Address, database *ledger.Ledger, acct *account.Account) common3.ConAddr {
	te := common3.TestEnv{Witness: []common.Address{admin, acct.Address}}
	newCodeHex := hex.EncodeToString(newCode)
	//name, version, author, email, description
	param := fmt.Sprintf("[bytearray:%s,string:%s,string:%s,string:%s,string:%s,string:%s]",
		newCodeHex, "bridge", "1.0", "ontology", "@ont.io", "desc")
	tc := common3.NewTestCase(te, false, "migrate", param, "bool:true", "")
	tx, err := common3.GenNeoVMTransaction(tc, bridge.Address, &common3.TestContext{})
	checkErr(err)
	execTxCheckRes(tx, tc, database, bridge.Address, acct)
	newAddr := common.AddressFromVmCode(newCode)
	return common3.ConAddr{
		Address: newAddr,
	}
}

func bridgeTest(bridge, wingOep4, wingErc20 common3.ConAddr, database *ledger.Ledger, acct *account.Account) {
	testPrivateKeyStr := "59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d"
	testPrivateKey, err := crypto.HexToECDSA(testPrivateKeyStr)
	checkErr(err)
	testEthAddr = crypto.PubkeyToAddress(testPrivateKey.PublicKey)

	// 调用 bridge init 方法
	admin, _ := common.AddressFromBase58("ARGK44mXXZfU6vcdSfFKMzjaabWxyog1qb")
	contractInit(admin, bridge, wingOep4, wingErc20, database, acct)
	txNonce := int64(2)
	txNonce = bridgeTestInner(admin, bridge, wingOep4, wingErc20, database, acct, txNonce)
	oep4BalanceBefore := oep4BalanceOf(database, wingOep4, bridge.Address)
	erc20BalanceBefore := erc20BalanceOf(database, wingErc20, ontAddrToEthAddr(bridge.Address), txNonce)

	newCode := loadContract(contractDir2 + "/" + "bridge2.avm")
	newBridge := migrateBridge(bridge, newCode, admin, database, acct)
	log.Infof("newBridge: %s", newBridge.Address.ToHexString())
	oep4BalanceAfter := oep4BalanceOf(database, wingOep4, newBridge.Address)
	erc20BalanceAfter := erc20BalanceOf(database, wingErc20, ontAddrToEthAddr(newBridge.Address), txNonce)
	ensureTrue(oep4BalanceBefore, oep4BalanceAfter)
	ensureTrue(erc20BalanceBefore, erc20BalanceAfter)
	bridgeTestInner(admin, newBridge, wingOep4, wingErc20, database, acct, txNonce)
}

func bridgeTestInner(admin common.Address, bridge, wingOep4, wingErc20 common3.ConAddr, database *ledger.Ledger, acct *account.Account, txNonce int64) int64 {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 100; i++ {
		amount := uint64(rand.Int63n(int64(1000000000)))
		// oep4 to erc20
		ethAcct, _ := common.AddressParseFromBytes(testEthAddr.Bytes())
		beforeAdmin := oep4BalanceOf(database, wingOep4, admin)
		beforeEthAcct := erc20BalanceOf(database, wingErc20, testEthAddr, txNonce)
		oep4ToErc20(bridge, admin, ethAcct, amount, database, acct, "oep4ToErc20")
		afterAdmin := oep4BalanceOf(database, wingOep4, admin)
		ensureTrue(afterAdmin, beforeAdmin-amount)
		afterEthAcct := erc20BalanceOf(database, wingErc20, testEthAddr, txNonce)
		ensureTrue(afterEthAcct, beforeEthAcct+amount)

		// erc20 to oep4
		erc20Approve(database, wingErc20, ontAddrToEthAddr(bridge.Address), amount, acct, txNonce)
		txNonce++
		beforeEthAcct = erc20BalanceOf(database, wingErc20, testEthAddr, txNonce)
		beforeAdmin = oep4BalanceOf(database, wingOep4, admin)
		oep4ToErc20(bridge, admin, ethAcct, amount, database, acct, "erc20ToOep4")
		afterEthAcct = erc20BalanceOf(database, wingErc20, testEthAddr, txNonce)
		ensureTrue(afterEthAcct, beforeEthAcct-amount)
		afterAdmin = oep4BalanceOf(database, wingOep4, admin)
		ensureTrue(afterAdmin, beforeAdmin+amount)
	}
	return txNonce
}
func oep4ToErc20(bridge common3.ConAddr, admin common.Address, ethAcct common.Address, amount uint64, database *ledger.Ledger, acct *account.Account, method string) {
	var param string
	if method == "oep4ToErc20" {
		param = fmt.Sprintf("[address:%s,address:%s,int:%d]", admin.ToBase58(), ethAcct.ToBase58(), amount)
	} else if method == "erc20ToOep4" {
		param = fmt.Sprintf("[address:%s,address:%s,int:%d]", ethAcct.ToBase58(), admin.ToBase58(), amount)
	} else {
		panic(method)
	}

	testContext := common3.TestContext{
		Admin:   admin,
		AddrMap: nil,
	}
	env := common3.TestEnv{
		Witness: []common.Address{admin, ethAcct},
	}
	tc := common3.NewTestCase(env, false, method, param, "bool:true", WingABI)
	tx, err := common3.GenNeoVMTransaction(tc, bridge.Address, &testContext)
	checkErr(err)
	log.Infof("method: %s", method)
	_, err = database.PreExecuteContract(tx)
	checkErr(err)
	//log.Infof("method: %s, pre: %s", method, JsonString(reee))
	execTxCheckRes(tx, tc, database, bridge.Address, acct)
}

func contractInit(admin common.Address, bridge, wingOep4, wingErc20 common3.ConAddr, database *ledger.Ledger, acct *account.Account) {
	//wing oep4 init
	param := "int:1"
	testContext := common3.TestContext{
		Admin:   admin,
		AddrMap: nil,
	}
	te := common3.TestEnv{Witness: []common.Address{admin, acct.Address}}
	tc := common3.NewTestCase(te, false, "init", param, "bool:true", "")
	tx, err := common3.GenNeoVMTransaction(tc, wingOep4.Address, &testContext)
	checkErr(err)
	execTxCheckRes(tx, tc, database, wingOep4.Address, acct)

	// oep4 balanceOf
	ba := oep4BalanceOf(database, wingOep4, admin)
	ensureTrue(1000000000000000, ba)
	amount := uint64(100000000000000)
	oep4Transfer(database, wingOep4, admin, bridge.Address, amount, testContext, acct)
	ba = oep4BalanceOf(database, wingOep4, bridge.Address)
	ensureTrue(amount, ba)
	// bridge init
	param = fmt.Sprintf("[address:%s,address:%s]", wingOep4.Address.ToBase58(), wingErc20.Address.ToBase58())
	log.Infof("bridge init param: %s", param)
	tc = common3.NewTestCase(te, false, "init", param, "bool:true", "")
	tx, err = common3.GenNeoVMTransaction(tc, bridge.Address, &testContext)
	checkErr(err)
	execTxCheckRes(tx, tc, database, bridge.Address, acct)

	// bridge get_ont_address
	tc = common3.NewTestCase(te, false, "get_oep4_address", "int:1", "address:"+wingOep4.Address.ToBase58(), "")
	tx, err = common3.GenNeoVMTransaction(tc, bridge.Address, &testContext)
	checkErr(err)
	execTxCheckRes(tx, tc, database, bridge.Address, acct)

	// wing erc20 totalSupply
	evmTx, err := GenEVMTx(1, common4.BytesToAddress(wingErc20.Address[:]), "totalSupply", "")
	checkErr(err)
	tx, err = types.TransactionFromEIP155(evmTx)
	checkErr(err)
	res, err := database.PreExecuteContract(tx)
	checkErr(err)
	r := res.Result.([]byte)
	log.Infof("execute totalSupply: %v", JsonString(r))

	// wing erc20 name
	evmTx, err = GenEVMTx(1, common4.BytesToAddress(wingErc20.Address[:]), "name")
	checkErr(err)
	tx, err = types.TransactionFromEIP155(evmTx)
	checkErr(err)
	res, err = database.PreExecuteContract(tx)
	checkErr(err)
	parseEthResult("name", res.Result, WingABI)

	// wingErc20 balanceOf
	ba = erc20BalanceOf(database, wingErc20, testEthAddr, 1)
	ensureTrue(500000000000000, ba)
	erc20Transfer(database, wingErc20, ontAddrToEthAddr(bridge.Address), amount, acct, 1)
	ba = erc20BalanceOf(database, wingErc20, ontAddrToEthAddr(bridge.Address), 2)
	ensureTrue(amount, ba)
}

func ensureTrue(expect, actual uint64) {
	if expect != actual {
		panic(fmt.Sprintf("expect: %d, actual: %d", expect, actual))
	}
}

func ethAddrToOntAddr(ethAddr common4.Address) common.Address {
	addr, err := common.AddressParseFromBytes(ethAddr[:])
	checkErr(err)
	return addr
}
func ontAddrToEthAddr(ontAddr common.Address) common4.Address {
	return common4.BytesToAddress(ontAddr[:])
}

func erc20Transfer(database *ledger.Ledger, wingErc20 common3.ConAddr, to common4.Address, amount uint64, acct *account.Account, txNonce int64) {
	evmTx, err := GenEVMTx(txNonce, common4.BytesToAddress(wingErc20.Address[:]), "transfer", to, big.NewInt(0).SetUint64(amount))
	checkErr(err)
	tx, err := types.TransactionFromEIP155(evmTx)
	checkErr(err)
	tc := common3.NewTestCase(common3.TestEnv{}, false, "transfer", "", "bool:true", WingABI)
	execTxCheckRes(tx, tc, database, wingErc20.Address, acct)
}

func erc20Approve(database *ledger.Ledger, wingErc20 common3.ConAddr, to common4.Address, amount uint64, acct *account.Account, txNonce int64) {
	evmTx, err := GenEVMTx(txNonce, common4.BytesToAddress(wingErc20.Address[:]), "approve", to, big.NewInt(0).SetUint64(amount))
	checkErr(err)
	tx, err := types.TransactionFromEIP155(evmTx)
	checkErr(err)
	tc := common3.NewTestCase(common3.TestEnv{}, false, "approve", "", "bool:true", WingABI)
	execTxCheckRes(tx, tc, database, wingErc20.Address, acct)
}

func erc20BalanceOf(database *ledger.Ledger, wingErc20 common3.ConAddr, addr common4.Address, txNonce int64) uint64 {
	evmTx, err := GenEVMTx(txNonce, common4.BytesToAddress(wingErc20.Address[:]), "balanceOf", addr)
	checkErr(err)
	tx, err := types.TransactionFromEIP155(evmTx)
	checkErr(err)
	res, err := database.PreExecuteContract(tx)
	checkErr(err)
	data := parseEthResult("balanceOf", res.Result, WingABI)
	d, ok := data.(*big.Int)
	if !ok {
		panic(data)
	}
	return d.Uint64()
}

func oep4BalanceOf(database *ledger.Ledger, wingOep4 common3.ConAddr, addr common.Address) uint64 {
	tc := common3.NewTestCase(common3.TestEnv{}, false, "balanceOf", fmt.Sprintf("[address:%s]", addr.ToBase58()), "", "")
	tx, err := common3.GenNeoVMTransaction(tc, wingOep4.Address, &common3.TestContext{})
	checkErr(err)
	res, err := database.PreExecuteContract(tx)
	checkErr(err)
	data := res.Result.(string)
	d, _ := hex.DecodeString(data)
	return common.BigIntFromNeoBytes(d).Uint64()
}

func oep4Transfer(database *ledger.Ledger, wingOep4 common3.ConAddr, from, to common.Address, amount uint64, testContext common3.TestContext, acct *account.Account) {
	tc := common3.NewTestCase(common3.TestEnv{}, false, "transfer", fmt.Sprintf("[address:%s,address:%s,int:%d]", from.ToBase58(), to.ToBase58(), amount), "bool:true", "")
	tx, err := common3.GenNeoVMTransaction(tc, wingOep4.Address, &testContext)
	checkErr(err)
	execTxCheckRes(tx, tc, database, wingOep4.Address, acct)
}

func parseEthResult(method string, data interface{}, jsonAbi string) interface{} {
	r := data.([]byte)
	parsed, _ := abi.JSON(strings.NewReader(jsonAbi))
	eee, err := parsed.Unpack(method, r)
	checkErr(err)
	log.Infof("method: %s, result: %v", method, eee)
	return eee[0]
}

func GenEVMTx(nonce int64, contractAddr common4.Address, method string, params ...interface{}) (*types2.Transaction, error) {
	chainId := big.NewInt(int64(config.DefConfig.P2PNode.EVMChainId))
	opts, err := bind.NewKeyedTransactorWithChainID(testPrivateKey, chainId)
	opts.GasPrice = big.NewInt(0)
	opts.Nonce = big.NewInt(nonce)
	opts.GasLimit = 8000000

	checkErr(err)
	parsed, err := abi.JSON(strings.NewReader(WingABI))
	checkErr(err)
	input, err := parsed.Pack(method, params...)
	deployTx := types2.NewTransaction(opts.Nonce.Uint64(), contractAddr, opts.Value, opts.GasLimit, opts.GasPrice, input)
	signedTx, err := opts.Signer(opts.From, deployTx)
	checkErr(err)
	return signedTx, err
}
