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
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/go-interpreter/wagon/exec"
	"github.com/go-interpreter/wagon/wasm"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/core/types"
	utils2 "github.com/ontio/ontology/core/utils"
	"github.com/ontio/ontology/events"
	"github.com/ontio/ontology/smartcontract/service/wasmvm"
	"github.com/ontio/ontology/smartcontract/states"
	common3 "github.com/ontio/ontology/wasmtest/common"
)

const contractDir = "testwasmdata"
const testcaseMethod = "testcase"

func NewDeployWasmContract(signer *account.Account, code []byte) (*types.Transaction, error) {
	mutable := utils.NewDeployCodeTransaction(0, 100000000, code, payload.WASMVM_TYPE, "name", "version",
		"author", "email", "desc")

	err := utils.SignTransaction(signer, mutable)
	if err != nil {
		return nil, err
	}
	tx, err := mutable.IntoImmutable()
	return tx, err
}

func NewDeployNeoContract(signer *account.Account, code []byte) (*types.Transaction, error) {
	mutable := utils.NewDeployCodeTransaction(0, 100000000, code, payload.NEOVM_TYPE, "name", "version",
		"author", "email", "desc")

	err := utils.SignTransaction(signer, mutable)
	if err != nil {
		return nil, err
	}
	tx, err := mutable.IntoImmutable()
	return tx, err
}

func ExactTestCase(code []byte) [][]common3.TestCase {
	m, err := wasm.ReadModule(bytes.NewReader(code), func(name string) (*wasm.Module, error) {
		switch name {
		case "env":
			return wasmvm.NewHostModule(), nil
		}
		return nil, fmt.Errorf("module %q unknown", name)
	})
	checkErr(err)

	compiled, err := exec.CompileModule(m)
	checkErr(err)

	vm, err := exec.NewVMWithCompiled(compiled, 10*1024*1024)
	checkErr(err)

	param := common.NewZeroCopySink(nil)
	param.WriteString(testcaseMethod)
	host := &wasmvm.Runtime{Input: param.Bytes()}
	vm.HostData = host
	vm.RecoverPanic = true
	vm.AvaliableGas = &exec.Gas{GasLimit: 100000000000000, GasPrice: 0}
	vm.CallStackDepth = 1024

	entry := compiled.RawModule.Export.Entries["invoke"]
	index := int64(entry.Index)
	_, err = vm.ExecCode(index)
	checkErr(err)

	var testCase [][]common3.TestCase
	source := common.NewZeroCopySource(host.Output)
	jsonCase, _, _, _ := source.NextString()

	if len(jsonCase) == 0 {
		panic("failed to get testcase data from contract")
	}

	err = json.Unmarshal([]byte(jsonCase), &testCase)
	checkErr(err)

	return testCase
}

func LoadContracts(dir string) (map[string][]byte, error) {
	contracts := make(map[string][]byte)
	fnames, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return nil, err
	}
	for _, name := range fnames {
		if !(strings.HasSuffix(name, ".wasm") || strings.HasSuffix(name, ".avm")) {
			continue
		}
		raw, err := ioutil.ReadFile(name)
		if err != nil {
			return nil, err
		}
		contracts[path.Base(name)] = raw
	}

	return contracts, nil
}

func init() {
	log.InitLog(log.InfoLog, log.PATH, log.Stdout)
	runtime.GOMAXPROCS(4)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
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

	bookkeepers := []keypair.PublicKey{acct.PublicKey}
	//Init event hub
	events.Init()

	log.Info("1. Loading the Ledger")
	database, err := ledger.NewLedger(datadir, 1000000)
	checkErr(err)
	ledger.DefLedger = database
	genblock, err := genesis.BuildGenesisBlock(bookkeepers, config.DefConfig.Genesis)
	checkErr(err)
	err = database.Init(bookkeepers, genblock)
	checkErr(err)

	log.Info("loading wasm contract")
	contract, err := LoadContracts(contractDir)
	checkErr(err)

	log.Infof("deploying %d wasm contracts", len(contract))
	txes := make([]*types.Transaction, 0, len(contract))
	for file, cont := range contract {
		if strings.HasSuffix(file, ".wasm") {
			tx, err := NewDeployWasmContract(acct, cont)
			checkErr(err)
			txes = append(txes, tx)
		} else if strings.HasSuffix(file, ".avm") {
			tx, err := NewDeployNeoContract(acct, cont)
			checkErr(err)
			txes = append(txes, tx)
		}
	}
	block, _ := makeBlock(acct, txes)
	err = database.AddBlock(block, common.UINT256_EMPTY)
	checkErr(err)

	addrMap := make(map[string]common.Address)
	for file, code := range contract {
		addrMap[path.Base(file)] = common.AddressFromVmCode(code)
	}

	testContext := common3.TestContext{
		Admin:   acct.Address,
		AddrMap: addrMap,
	}

	for file, cont := range contract {
		if !strings.HasSuffix(file, ".wasm") {
			continue
		}

		log.Infof("exacting testcase from %s", file)
		testCases := ExactTestCase(cont)
		addr := common.AddressFromVmCode(cont)
		for _, testCase := range testCases[0] { // only handle group 0 currently
			val, _ := json.Marshal(testCase)
			log.Info("executing testcase: ", string(val))
			tx, err := common3.GenWasmTransaction(testCase, addr, &testContext)
			checkErr(err)

			res, err := database.PreExecuteContract(tx)
			checkErr(err)

			height := database.GetCurrentBlockHeight()
			header, err := database.GetHeaderByHeight(height)
			checkErr(err)
			blockTime := header.Timestamp + 1

			execEnv := ExecEnv{Time: blockTime, Height: height + 1, Tx: tx, BlockHash: header.Hash(), Contract: addr}
			checkExecResult(testCase, res, execEnv)

			block, _ := makeBlock(acct, []*types.Transaction{tx})
			err = database.AddBlock(block, common.UINT256_EMPTY)
			checkErr(err)
		}
	}

	log.Info("contract test succeed")
}

type ExecEnv struct {
	Contract  common.Address
	Time      uint32
	Height    uint32
	Tx        *types.Transaction
	BlockHash common.Uint256
}

func checkExecResult(testCase common3.TestCase, result *states.PreExecResult, execEnv ExecEnv) {
	assertEq(result.State, byte(1))
	ret := result.Result.(string)
	switch testCase.Method {
	case "timestamp":
		sink := common.NewZeroCopySink(nil)
		sink.WriteUint64(uint64(execEnv.Time))
		assertEq(ret, hex.EncodeToString(sink.Bytes()))
	case "block_height":
		sink := common.NewZeroCopySink(nil)
		sink.WriteUint32(uint32(execEnv.Height))
		assertEq(ret, hex.EncodeToString(sink.Bytes()))
	case "self_address", "entry_address":
		assertEq(ret, hex.EncodeToString(execEnv.Contract[:]))
	case "caller_address":
		assertEq(ret, hex.EncodeToString(common.ADDRESS_EMPTY[:]))
	case "current_txhash":
		hash := execEnv.Tx.Hash()
		assertEq(ret, hex.EncodeToString(hash[:]))
	case "current_blockhash":
		assertEq(ret, hex.EncodeToString(execEnv.BlockHash[:]))
	//case "sha256":
	//	let data :&[u8]= source.read().unwrap();
	//	sink.write(runtime::sha256(&data))
	//}
	default:
		if len(testCase.Expect) != 0 {
			expect, err := utils.ParseParams(testCase.Expect)
			checkErr(err)
			exp, err := utils2.BuildWasmContractParam(expect)
			checkErr(err)
			assertEq(ret, hex.EncodeToString(exp))
		}
	}
}

func GenAccounts(num int) []*account.Account {
	var accounts []*account.Account
	for i := 0; i < num; i++ {
		acc := account.NewAccount("")
		accounts = append(accounts, acc)
	}
	return accounts
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

func assertEq(a interface{}, b interface{}) {
	if reflect.DeepEqual(a, b) == false {
		panic(fmt.Sprintf("not equal: a= %v, b=%v", a, b))
	}
}
