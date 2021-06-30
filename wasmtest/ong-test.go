package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	types2 "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ontio/ontology/account"
	utils2 "github.com/ontio/ontology/cmd/utils"
	common2 "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
	cutils "github.com/ontio/ontology/core/utils"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"math/big"
	"time"
)

var gasPrice = uint64(500)
var gasLimit = uint64(200000)

func ongTest(database *ledger.Ledger, acct *account.Account) {
	ongBalanceOf(database, acct.Address)
	fromPrivateKey, err := crypto.GenerateKey()
	checkErr(err)
	fromEthAddr := crypto.PubkeyToAddress(fromPrivateKey.PublicKey)
	transferOng(database, acct, common2.Address(fromEthAddr), 1*1000000000)
	fromBalance1 := ongBalanceOf(database, common2.Address(fromEthAddr))
	log.Infof("from ong balance: %d", fromBalance1)
	toPrivateKey, err := crypto.GenerateKey()
	checkErr(err)
	toEthAddr := crypto.PubkeyToAddress(toPrivateKey.PublicKey)
	nonce := int64(0)
	transferAmt := 0.5 * 1000000000
	tx := evmTransferOng(fromPrivateKey, toEthAddr, nonce, int64(transferAmt))
	genBlock(database, acct, tx)
	evt, err := database.GetEventNotifyByTx(tx.Hash())
	checkErr(err)
	fromBalance2 := ongBalanceOf(database, common2.Address(fromEthAddr))
	if fromBalance1-fromBalance2 != evt.GasConsumed+uint64(transferAmt) {
		panic(fmt.Sprintf("fromBalance1:%d, fromBalance2:%d, evt.GasConsumed:%d, transferAmt:%d",
			fromBalance1, fromBalance2, evt.GasConsumed, uint64(transferAmt)))
	}
	ethAcc, err := database.GetEthAccount(fromEthAddr)
	checkErr(err)
	if ethAcc.Nonce != uint64(nonce+1) {
		panic(fmt.Sprintf("ethAcc.Nonce:%d, nonce+1:%d", ethAcc.Nonce, nonce+1))
	}

	tx = evmTransferOng(fromPrivateKey, toEthAddr, nonce+1, 0.5*1000000000)
	genBlock(database, acct, tx)
	evt, err = database.GetEventNotifyByTx(tx.Hash())
	checkErr(err)
	fromBalance3 := ongBalanceOf(database, common2.Address(fromEthAddr))

	if fromBalance2-fromBalance3 != evt.GasConsumed {
		panic(fmt.Sprintf("fromBalance2: %d, fromBalance3: %d, evt.GasConsumed:%d",
			fromBalance2, fromBalance3, evt.GasConsumed))
	}
	ethAcc, err = database.GetEthAccount(fromEthAddr)
	checkErr(err)
	if ethAcc.Nonce != uint64(nonce+2) {
		panic(fmt.Sprintf("ethAcc.Nonce: %d, nonce+2: %d", ethAcc.Nonce, nonce+2))
	}
}

func genBlock(database *ledger.Ledger, acct *account.Account, tx *types.Transaction) {
	_, err := database.PreExecuteContract(tx)
	checkErr(err)
	block, _ := makeBlock(acct, []*types.Transaction{tx})
	err = database.AddBlock(block, nil, common2.UINT256_EMPTY)
	checkErr(err)
}

func newNativeTx(contractAddress common2.Address, version byte, method string, params []interface{}) *types.MutableTransaction {
	invokeCode, err := cutils.BuildNativeInvokeCode(contractAddress, version, method, params)
	checkErr(err)
	invokePayload := &payload.InvokeCode{
		Code: invokeCode,
	}
	tx := &types.MutableTransaction{
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		TxType:   types.InvokeNeo,
		Nonce:    uint32(time.Now().Unix()),
		Payload:  invokePayload,
		Sigs:     nil,
	}
	return tx
}

func transferOng(database *ledger.Ledger, acct *account.Account, toAddr common2.Address, amount int64) {
	state := &ont.State{
		From:  acct.Address,
		To:    toAddr,
		Value: uint64(amount),
	}
	mutable := newNativeTx(utils.OngContractAddress, 0, "transfer", []interface{}{[]*ont.State{state}})
	err := utils2.SignTransaction(acct, mutable)
	checkErr(err)
	tx, err := mutable.IntoImmutable()
	checkErr(err)
	genBlock(database, acct, tx)
}

func ongBalanceOf(database *ledger.Ledger, acctAddr common2.Address) uint64 {
	mutable := newNativeTx(utils.OngContractAddress, 0, "balanceOf", []interface{}{acctAddr[:]})
	tx, err := mutable.IntoImmutable()
	checkErr(err)
	res, err := database.PreExecuteContract(tx)
	checkErr(err)
	data, err := hex.DecodeString(res.Result.(string))
	checkErr(err)
	balance := common2.BigIntFromNeoBytes(data)
	return balance.Uint64()
}

func evmTransferOng(testPrivateKey *ecdsa.PrivateKey, toEthAddr common.Address, nonce int64, value int64) *types.Transaction {
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
