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
	"io/ioutil"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	types2 "github.com/ethereum/go-ethereum/core/types"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/core/types"
	cutils "github.com/ontio/ontology/core/utils"
	"github.com/ontio/ontology/events"
)

func NewLedger() (*ledger.Ledger, *account.Account) {
	datadir := "testdata"
	err := os.RemoveAll(datadir)
	defer func() {
		_ = os.RemoveAll(datadir)
		_ = os.RemoveAll(log.PATH)
	}()
	checkErr(err)
	log.Trace("Node version: ", config.Version)

	acct := account.NewAccount("")
	buf := keypair.SerializePublicKey(acct.PublicKey)
	config.DefConfig.Genesis.ConsensusType = "solo"
	config.DefConfig.Genesis.SOLO.GenBlockTime = 3
	config.DefConfig.Genesis.SOLO.Bookkeepers = []string{hex.EncodeToString(buf)}
	config.DefConfig.P2PNode.NetworkId = 3

	bookkeepers := []keypair.PublicKey{acct.PublicKey}
	//Init event hub
	events.Init()

	log.Info("1. Loading the Ledger")
	genblock, err := genesis.BuildGenesisBlock(bookkeepers, config.DefConfig.Genesis)
	checkErr(err)
	database, err := ledger.InitLedger(datadir, 1000000, bookkeepers, genblock)
	checkErr(err)
	ledger.DefLedger = database
	return database, acct
}

func makeBlock(acc *account.Account, txs []*types.Transaction) (*types.Block, error) {
	nextBookkeeper, err := types.AddressFromBookkeepers([]keypair.PublicKey{acc.PublicKey})
	if err != nil {
		return nil, fmt.Errorf("GetBookkeeperAddress error:%s", err)
	}
	prevHash := ledger.DefLedger.GetCurrentBlockHash()
	height := ledger.DefLedger.GetCurrentBlockHeight()

	nonce := uint64(height)
	var txHash []common.Uint256
	for _, t := range txs {
		txHash = append(txHash, t.Hash())
	}

	txRoot := common.ComputeMerkleRoot(txHash)

	blockRoot := ledger.DefLedger.GetBlockRootWithNewTxRoots(height+1, []common.Uint256{txRoot})
	header := &types.Header{
		Version:          0,
		PrevBlockHash:    prevHash,
		TransactionsRoot: txRoot,
		BlockRoot:        blockRoot,
		Timestamp:        constants.GENESIS_BLOCK_TIMESTAMP + height + 1,
		Height:           height + 1,
		ConsensusData:    nonce,
		NextBookkeeper:   nextBookkeeper,
	}
	block := &types.Block{
		Header:       header,
		Transactions: txs,
	}

	blockHash := block.Hash()

	sig, err := signature.Sign(acc, blockHash[:])
	if err != nil {
		return nil, fmt.Errorf("signature, Sign error:%s", err)
	}

	block.Header.Bookkeepers = []keypair.PublicKey{acc.PublicKey}
	block.Header.SigData = [][]byte{sig}
	return block, nil
}

func newNativeTx(contractAddress common.Address, version byte, gasPrice, gasLimit uint64, method string, params []interface{}) *types.MutableTransaction {
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

func loadContract(filePath string) []byte {
	if common.FileExisted(filePath) {
		raw, err := ioutil.ReadFile(filePath)
		checkErr(err)
		code, err := hex.DecodeString(strings.TrimSpace(string(raw)))
		if err != nil {
			return raw
		} else {
			return code
		}
	} else {
		panic("no existed file:" + filePath)
	}
}

func NewDeployEvmContract(privateKey *ecdsa.PrivateKey, nonce int64, gasPrice, gasLimit uint64, msgValue int64, code []byte, jsonABI string, params ...interface{}) *types2.Transaction {
	chainId := big.NewInt(int64(config.DefConfig.P2PNode.EVMChainId))
	opts, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
	checkErr(err)
	opts.GasPrice = big.NewInt(int64(gasPrice))
	opts.Nonce = big.NewInt(nonce)
	opts.GasLimit = gasLimit
	parsed, err := abi.JSON(strings.NewReader(jsonABI))
	checkErr(err)
	input, err := parsed.Pack("", params...)
	checkErr(err)
	input = append(code, input...)
	deployTx := types2.NewContractCreation(opts.Nonce.Uint64(), big.NewInt(msgValue), opts.GasLimit, opts.GasPrice, input)
	signedTx, err := opts.Signer(opts.From, deployTx)
	checkErr(err)
	return signedTx
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
