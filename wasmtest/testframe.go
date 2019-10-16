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
	"math"
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
	"github.com/ontio/ontology/cmd"
	cmdcommon "github.com/ontio/ontology/cmd/common"
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
	common2 "github.com/ontio/ontology/http/base/common"
	"github.com/ontio/ontology/smartcontract/service/neovm"
	"github.com/ontio/ontology/smartcontract/service/wasmvm"
	"github.com/ontio/ontology/smartcontract/states"
	vmtypes "github.com/ontio/ontology/vm/neovm/types"
	common3 "github.com/ontio/ontology/wasmtest/common"
	"github.com/urfave/cli"
)

const (
	testcaseMethod = "testcase"
)

var batchMode bool

func ontologyCLI(ctx *cli.Context) {
	paramsStr, deploydir, contractfile := optionErrCheck(ctx)
	acct, database := initOntologyLedger(ctx)

	log.Info("loading contract")
	contract := getContact(deploydir, contractfile)

	log.Infof("deploying %d contracts", len(contract))
	deployContract(acct, database, contract)

	testContext := makeTestContext(acct, contract)

	contractfilename := path.Base(contractfile)

	if batchMode {
		if len(contractfile) == 0 {
			testWithbatchMode(acct, database, contract, testContext)
		} else {
			testSpecifiedContract(acct, database, contractfilename, contract[contractfilename], testContext)
		}
	} else if len(contractfile) != 0 {
		invokeSpecifiedContract(acct, database, contractfilename, paramsStr, testContext)
	} else {
		// only deploy contract.
	}
}

func getAccountByWallet(ctx *cli.Context) *account.Account {
	wallet, err := cmdcommon.OpenWallet(ctx)
	checkErr(err)
	passwd, err := cmdcommon.GetPasswd(ctx)
	checkErr(err)
	acct, err := wallet.GetDefaultAccount(passwd)
	checkErr(err)
	return acct
}

func invokeSpecifiedContract(acct *account.Account, database *ledger.Ledger, contractfile string, paramsStr string, testContext *common3.TestContext) {
	var tx *types.Transaction
	if strings.HasSuffix(contractfile, ".avm") || strings.HasSuffix(contractfile, ".avm.str") {
		testCase := common3.TestCase{Param: paramsStr}
		t, err := common3.GenNeoVMTransaction(testCase, testContext.AddrMap[contractfile], testContext)
		checkErr(err)
		tx = t
	} else if strings.HasSuffix(contractfile, ".wasm") || strings.HasSuffix(contractfile, ".wasm.str") {
		testCase := common3.TestCase{Param: paramsStr}
		t, err := common3.GenWasmTransaction(testCase, testContext.AddrMap[contractfile], testContext)
		checkErr(err)
		tx = t
	} else {
		panic("InvokeWasm: error suffix type")
	}

	res, err := database.PreExecuteContract(tx)
	log.Infof("testcase consume gas: %d", res.Gas)
	checkErr(err)
	block, _ := makeBlock(acct, []*types.Transaction{tx})
	err = database.AddBlock(block, common.UINT256_EMPTY)
	checkErr(err)
	if len(res.Notify) != 0 {
		js, _ := json.Marshal(res.Notify)
		log.Infof("Notify info : %s", string(js))
	}

	if res.Result != nil {
		js, _ := json.Marshal(res.Result)
		log.Infof("Return result: %s", string(js))
	}
}

func NewDeployWasmContract(signer *account.Account, code []byte) (*types.Transaction, error) {
	mutable, err := utils.NewDeployCodeTransaction(0, 100000000, code, payload.WASMVM_TYPE, "name", "version",
		"author", "email", "desc")
	if err != nil {
		return nil, err
	}
	err = utils.SignTransaction(signer, mutable)
	if err != nil {
		return nil, err
	}
	tx, err := mutable.IntoImmutable()
	return tx, err
}

func NewDeployNeoContract(signer *account.Account, code []byte) (*types.Transaction, error) {
	mutable, err := utils.NewDeployCodeTransaction(0, 100000000, code, payload.NEOVM_TYPE, "name", "version",
		"author", "email", "desc")
	if err != nil {
		return nil, err
	}
	err = utils.SignTransaction(signer, mutable)
	if err != nil {
		return nil, err
	}
	tx, err := mutable.IntoImmutable()
	return tx, err
}

func GenNeoTextCaseTransaction(contract common.Address, database *ledger.Ledger) [][]common3.TestCase {
	params := make([]interface{}, 0)
	method := string("testcase")
	// neovm entry api is def Main(method, args). and testcase method api need no other args, so pass a random args to entry api.
	operation := 1
	params = append(params, method)
	params = append(params, operation)
	tx, err := common2.NewNeovmInvokeTransaction(0, 100000000, contract, params)
	imt, err := tx.IntoImmutable()
	if err != nil {
		panic(err)
	}
	res, err := database.PreExecuteContract(imt)
	if err != nil {
		panic(err)
	}

	ret := res.Result.(string)
	jsonCase, err := common.HexToBytes(ret)

	if err != nil {
		panic(err)
	}
	if len(jsonCase) == 0 {
		panic("failed to get testcase data from contract")
	}
	var testCase [][]common3.TestCase
	err = json.Unmarshal([]byte(jsonCase), &testCase)
	if err != nil {
		panic("failed Unmarshal")
	}
	return testCase
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
	envGasLimit := uint64(100000000000000)
	envExecStep := uint64(100000000000000)
	vm.AvaliableGas = &exec.Gas{GasLimit: &envGasLimit, GasPrice: 0, GasFactor: 5, ExecStep: &envExecStep}
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

func LoadContractsByDir(dir string, contracts map[string][]byte) {
	fnames, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return
	}

	for _, name := range fnames {
		if !strings.HasSuffix(name, ".wasm") && !strings.HasSuffix(name, ".avm") && !strings.HasSuffix(name, ".wasm.str") && !strings.HasSuffix(name, ".avm.str") {
			continue
		}
		raw, err := ioutil.ReadFile(name)
		if err != nil {
			log.Errorf("Read %s : %s", err)
		}

		if strings.HasSuffix(name, ".str") {
			code, err := hex.DecodeString(strings.TrimSpace(string(raw)))
			checkErr(err)
			contracts[path.Base(name)] = code
		} else {
			contracts[path.Base(name)] = raw
		}
	}
}

func init() {
	runtime.GOMAXPROCS(4)
	batchMode = true
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func execTxCheckRes(tx *types.Transaction, testCase common3.TestCase, database *ledger.Ledger, addr common.Address, acct *account.Account) {
	res, err := database.PreExecuteContract(tx)
	log.Infof("testcase consume gas: %d", res.Gas)
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

var (
	ByteCodeFlag = cli.StringFlag{
		Name:  "file,f",
		Usage: "the contract filename to be tested.",
	}
	DirFlag = cli.StringFlag{
		Name:  "dir,d",
		Usage: "deploy all contract in specified directory.",
	}
	ContractParamsFlag = cli.StringFlag{
		Name:  "param,p",
		Usage: "specify contract param when set option file(-f).",
	}
	BatchModeFlag = cli.BoolFlag{
		Name:  "batch,b",
		Usage: "batch mode to test all contract file specified by option dir(-d).",
	}
	LogLevelFlag = cli.UintFlag{
		Name:  "loglevel,l",
		Usage: "set the log levela.",
		Value: log.InfoLog,
	}
	WalletFlag = cli.BoolFlag{
		Name:  "walletuse,w",
		Usage: "use wallet.data default account. need input passwd",
	}
	StepMaxFlag = cli.BoolFlag{
		Name:  "maxstep,m",
		Usage: "set vm maxstep",
	}
)

func setupAPP() *cli.App {
	app := cli.NewApp()
	app.Usage = "ontology CLI"
	app.Action = ontologyCLI
	app.Version = config.Version
	app.Copyright = "Copyright in 2019 The Ontology Authors"
	app.Flags = []cli.Flag{
		ByteCodeFlag,
		DirFlag,
		ContractParamsFlag,
		BatchModeFlag,
		LogLevelFlag,
		WalletFlag,
		StepMaxFlag,
	}
	app.Before = func(context *cli.Context) error {
		runtime.GOMAXPROCS(runtime.NumCPU())
		return nil
	}
	return app
}

func main() {
	if err := setupAPP().Run(os.Args); err != nil {
		cmd.PrintErrorMsg(err.Error())
		os.Exit(1)
	}
}

func testSpecifiedContract(acct *account.Account, database *ledger.Ledger, file string, cont []byte, testContext *common3.TestContext) {
	log.Infof("exacting testcase from %s", file)
	addr := testContext.AddrMap[file]

	if strings.HasSuffix(file, ".avm") || strings.HasSuffix(file, ".avm.str") {
		testCases := GenNeoTextCaseTransaction(addr, database)
		for _, testCase := range testCases[0] { // only handle group 0 currently
			val, _ := json.Marshal(testCase)
			log.Info("executing testcase: ", string(val))
			tx, err := common3.GenNeoVMTransaction(testCase, addr, testContext)
			checkErr(err)

			execTxCheckRes(tx, testCase, database, addr, acct)
		}
	} else if strings.HasSuffix(file, ".wasm") || strings.HasSuffix(file, ".wasm.str") {
		testCases := ExactTestCase(cont)
		for _, testCase := range testCases[0] { // only handle group 0 currently
			val, _ := json.Marshal(testCase)
			log.Info("executing testcase: ", string(val))
			tx, err := common3.GenWasmTransaction(testCase, addr, testContext)
			checkErr(err)

			execTxCheckRes(tx, testCase, database, addr, acct)
		}
	} else {
		panic("testSpecifiedContract: error suffix contract name")
	}
}

func testWithbatchMode(acct *account.Account, database *ledger.Ledger, contract map[string][]byte, testContext *common3.TestContext) {
	for file, cont := range contract {
		testSpecifiedContract(acct, database, file, cont, testContext)
	}

	log.Info("contract test succeed")
}

func makeTestContext(acct *account.Account, contract map[string][]byte) *common3.TestContext {
	addrMap := make(map[string]common.Address)
	for file, code := range contract {
		addrMap[path.Base(file)] = common.AddressFromVmCode(code)
	}

	testContext := common3.TestContext{
		Admin:   acct.Address,
		AddrMap: addrMap,
	}
	return &testContext
}

func getContact(deploydir string, contractfile string) map[string][]byte {

	contract := make(map[string][]byte)
	if len(deploydir) != 0 {
		LoadContractsByDir(deploydir, contract)
	}

	if len(contractfile) != 0 {
		if !strings.HasSuffix(contractfile, ".wasm") && !strings.HasSuffix(contractfile, ".avm") && !strings.HasSuffix(contractfile, ".wasm.str") && !strings.HasSuffix(contractfile, ".avm.str") {
			log.Errorf("%s file suffix error: must be .wasm/.avm/.wasm.str/.avm.str.", contractfile)
			panic("")
		}
		raw, err := ioutil.ReadFile(contractfile)
		checkErr(err)
		if strings.HasSuffix(contractfile, ".str") {
			code, err := hex.DecodeString(strings.TrimSpace(string(raw)))
			checkErr(err)
			contract[path.Base(contractfile)] = code
		} else {
			contract[path.Base(contractfile)] = raw
		}
	}

	if len(contract) == 0 {
		panic("error: no contract to test.")
	}

	return contract
}

func deployContract(acct *account.Account, database *ledger.Ledger, contract map[string][]byte) {
	txes := make([]*types.Transaction, 0, len(contract))
	for file, cont := range contract {
		var tx *types.Transaction
		var err error
		if strings.HasSuffix(file, ".wasm") || strings.HasSuffix(file, ".wasm.str") {
			tx, err = NewDeployWasmContract(acct, cont)
		} else if strings.HasSuffix(file, ".avm") || strings.HasSuffix(file, ".avm.str") {
			tx, err = NewDeployNeoContract(acct, cont)
		} else {
			panic("error file suffix")
		}

		checkErr(err)

		res, err := database.PreExecuteContract(tx)
		log.Infof("deploy %s consume gas: %d", file, res.Gas)
		checkErr(err)
		txes = append(txes, tx)
	}

	block, _ := makeBlock(acct, txes)
	err := database.AddBlock(block, common.UINT256_EMPTY)
	checkErr(err)
}

func initOntologyLedger(ctx *cli.Context) (*account.Account, *ledger.Ledger) {
	datadir := "testdata"
	err := os.RemoveAll(datadir)
	defer func() {
		_ = os.RemoveAll(datadir)
		_ = os.RemoveAll(log.PATH)
	}()

	usewallet := ctx.GlobalBool(utils.GetFlagName(WalletFlag))
	var acct *account.Account
	if usewallet {
		acct = getAccountByWallet(ctx)
	} else {
		acct = account.NewAccount("")
	}

	buf := keypair.SerializePublicKey(acct.PublicKey)
	config.DefConfig.Genesis.ConsensusType = "solo"
	config.DefConfig.Genesis.SOLO.GenBlockTime = 3
	config.DefConfig.Genesis.SOLO.Bookkeepers = []string{hex.EncodeToString(buf)}
	config.DefConfig.P2PNode.NetworkId = 0

	bookkeepers := []keypair.PublicKey{acct.PublicKey}
	//Init event hub
	events.Init()

	//log.Info("1. Loading the Ledger")
	database, err := ledger.NewLedger(datadir, 1000000)
	checkErr(err)
	ledger.DefLedger = database
	genblock, err := genesis.BuildGenesisBlock(bookkeepers, config.DefConfig.Genesis)
	checkErr(err)
	err = database.Init(bookkeepers, genblock)
	checkErr(err)
	return acct, database
}

func optionErrCheck(ctx *cli.Context) (string, string, string) {
	LogLevel := ctx.Uint(utils.GetFlagName(LogLevelFlag))
	log.InitLog(int(LogLevel), log.PATH, log.Stdout)
	maxstep := ctx.GlobalBool(utils.GetFlagName(StepMaxFlag))

	if maxstep {
		neovm.VM_STEP_LIMIT = math.MaxInt32
		config.DEFAULT_WASM_MAX_STEPCOUNT = math.MaxUint64
	}

	batchMode = ctx.GlobalBool(utils.GetFlagName(BatchModeFlag))
	deploydir := ctx.String(utils.GetFlagName(DirFlag))
	paramsStr := ctx.String(utils.GetFlagName(ContractParamsFlag))
	contractfile := ctx.String(utils.GetFlagName(ByteCodeFlag))

	// error check.
	if batchMode {
		if len(paramsStr) != 0 {
			panic("can not use --param/-p option when in batchMode(-b) .")
		}

		if len(deploydir) == 0 && len(contractfile) == 0 {
			panic("must specify the contract directory --directory(-d) or contractfile(--file/-f) when in batchMode(-b).")
		}
	} else {
		if len(paramsStr) != 0 && len(contractfile) == 0 {
			panic("must specify the contract file(--file/-f) when use --param/-p .")
		}

		if len(deploydir) == 0 && len(paramsStr) == 0 && len(contractfile) == 0 {
			panic("please specify some option. use --help/-h .")
		}

		// here check the param rightness
		_, err := utils.ParseParams(paramsStr)
		if err != nil {
			log.Errorf("praseParam err: %s", err)
		}
	}

	if len(contractfile) != 0 {
		log.Infof("test file %s with args: %s", path.Base(contractfile), paramsStr)
	}

	return paramsStr, deploydir, contractfile
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
			if execEnv.Tx.TxType == types.InvokeNeo {
				val := buildNeoVmValueFromExpect(expect)
				cv, err := val.ConvertNeoVmValueHexString()
				checkErr(err)
				assertEq(cv, result.Result)
			} else if execEnv.Tx.TxType == types.InvokeWasm {
				exp, err := utils2.BuildWasmContractParam(expect)
				checkErr(err)
				assertEq(ret, hex.EncodeToString(exp))
			} else {
				panic("error tx type")
			}
		}
		if len(testCase.Notify) != 0 {
			js, _ := json.Marshal(result.Notify)
			assertEq(true, strings.Contains(string(js), testCase.Notify))
		}
	}
}

func buildNeoVmValueFromExpect(expectlist []interface{}) *vmtypes.VmValue {
	if len(expectlist) > 1 {
		panic("only support return one value")
	}
	expect := expectlist[0]

	switch expect.(type) {
	case string:
		val, err := vmtypes.VmValueFromBytes([]byte(expect.(string)))
		if err != nil {
			panic(err)
		}
		return &val
	case []byte:
		val, err := vmtypes.VmValueFromBytes(expect.([]byte))
		if err != nil {
			panic(err)
		}
		return &val
	case int64:
		val := vmtypes.VmValueFromInt64(expect.(int64))
		return &val
	case bool:
		val := vmtypes.VmValueFromBool(expect.(bool))
		return &val
	case common.Address:
		addr := expect.(common.Address)
		val, err := vmtypes.VmValueFromBytes(addr[:])
		if err != nil {
			panic(err)
		}
		return &val
	default:
		fmt.Printf("unspport param type %s", reflect.TypeOf(expect))
		panic("unspport param type")
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
