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

package utils

import (
	"strings"

	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/smartcontract/service/neovm"
	"github.com/urfave/cli"
)

const (
	DEFAULT_EXPORT_FILE   = "./OntBlocks.dat"
	DEFAULT_ABI_PATH      = "./abi"
	DEFAULT_EXPORT_HEIGHT = 0
	DEFAULT_WALLET_PATH   = "./wallet_data"
)

var (
	//Ontology setting
	ConfigFlag = cli.StringFlag{
		Name:  "config",
		Usage: "Genesis block config `<file>`. If doesn't specifies, use main net config as default.",
	}
	LogLevelFlag = cli.UintFlag{
		Name:  "loglevel",
		Usage: "Set the log level to `<level>` (0~6). 0:Trace 1:Debug 2:Info 3:Warn 4:Error 5:Fatal 6:MaxLevel",
		Value: config.DEFAULT_LOG_LEVEL,
	}
	DisableLogFileFlag = cli.BoolFlag{
		Name:  "disable-log-file",
		Usage: "Discard log output to file",
	}
	LogDirFlag = cli.StringFlag{
		Name:  "log-dir",
		Usage: "log output to the file",
		Value: log.PATH,
	}
	DisableEventLogFlag = cli.BoolFlag{
		Name:  "disable-event-log",
		Usage: "Discard event log output by smart contract execution",
	}
	WasmVerifyMethodFlag = cli.BoolFlag{
		Name:  "enable-wasmjit-verifier",
		Usage: "Enable wasmjit verifier to verify wasm contract",
	}
	WalletFileFlag = cli.StringFlag{
		Name:  "wallet,w",
		Value: config.DEFAULT_WALLET_FILE_NAME,
		Usage: "Wallet `<file>`",
	}
	ImportFileFlag = cli.StringFlag{
		Name:  "import-file",
		Usage: "Path of import `<file>`",
		Value: DEFAULT_EXPORT_FILE,
	}
	ImportEndHeightFlag = cli.UintFlag{
		Name:  "end-height",
		Usage: "Stop import block `<height>` of the import.",
		Value: DEFAULT_EXPORT_HEIGHT,
	}
	DataDirFlag = cli.StringFlag{
		Name:  "data-dir",
		Usage: "Block data storage `<path>`",
		Value: config.DEFAULT_DATA_DIR,
	}

	//Consensus setting
	EnableConsensusFlag = cli.BoolFlag{
		Name:  "enable-consensus",
		Usage: "Start consensus module",
	}
	MaxTxInBlockFlag = cli.IntFlag{
		Name:  "max-tx-in-block",
		Usage: "Max transaction `<number>` in block",
		Value: config.DEFAULT_MAX_TX_IN_BLOCK,
	}
	GasLimitFlag = cli.Uint64Flag{
		Name:  "gaslimit",
		Usage: "Min gas limit `<value>` of transaction to be accepted by tx pool.",
		Value: neovm.MIN_TRANSACTION_GAS,
	}
	GasPriceFlag = cli.Uint64Flag{
		Name:  "gasprice",
		Usage: "Min gas price `<value>` of transaction to be accepted by tx pool.",
		Value: config.DEFAULT_GAS_PRICE,
	}

	//Test Mode setting
	EnableTestModeFlag = cli.BoolFlag{
		Name:  "testmode",
		Usage: "Single node network for testing. In test mode, will start rpc, rest, web socket server, and set default gasprice to 0",
	}
	TestModeGenBlockTimeFlag = cli.UintFlag{
		Name:  "testmode-gen-block-time",
		Usage: "Block-out `<time>`(s) in test mode.",
		Value: config.DEFAULT_GEN_BLOCK_TIME,
	}

	//P2P setting
	ReservedPeersOnlyFlag = cli.BoolFlag{
		Name:  "reserved-only",
		Usage: "Connect reserved peers only. Reserved peers are configured with --reserved-file.",
	}
	ReservedPeersFileFlag = cli.StringFlag{
		Name:  "reserved-file",
		Usage: "Reserved peers `<file>`",
		Value: config.DEFAULT_RESERVED_FILE,
	}
	NetworkIdFlag = cli.UintFlag{
		Name:  "networkid",
		Usage: "Network id `<number>`. 1=ontology main net, 2=polaris test net, 3=testmode, and other for custom network",
		Value: config.NETWORK_ID_MAIN_NET,
	}
	NodePortFlag = cli.UintFlag{
		Name:  "nodeport",
		Usage: "P2P network port `<number>`",
		Value: config.DEFAULT_NODE_PORT,
	}
	HttpInfoPortFlag = cli.UintFlag{
		Name:  "httpinfo-port",
		Usage: "The listening port of http server for viewing node information `<number>`",
		Value: config.DEFAULT_HTTP_INFO_PORT,
	}
	MaxConnInBoundFlag = cli.UintFlag{
		Name:  "max-conn-in-bound",
		Usage: "Max connection `<number>` in bound",
		Value: config.DEFAULT_MAX_CONN_IN_BOUND,
	}
	MaxConnOutBoundFlag = cli.UintFlag{
		Name:  "max-conn-out-bound",
		Usage: "Max connection `<number>` out bound",
		Value: config.DEFAULT_MAX_CONN_OUT_BOUND,
	}
	MaxConnInBoundForSingleIPFlag = cli.UintFlag{
		Name:  "max-conn-in-bound-single-ip",
		Usage: "Max connection `<number>` in bound for single ip",
		Value: config.DEFAULT_MAX_CONN_IN_BOUND_FOR_SINGLE_IP,
	}
	// RPC settings
	RPCDisabledFlag = cli.BoolFlag{
		Name:  "disable-rpc",
		Usage: "Shut down the rpc server.",
	}
	RPCPortFlag = cli.UintFlag{
		Name:  "rpcport",
		Usage: "Json rpc server listening port `<number>`",
		Value: config.DEFAULT_RPC_PORT,
	}
	RPCLocalEnableFlag = cli.BoolFlag{
		Name:  "localrpc",
		Usage: "Enable local rpc server",
	}
	RPCLocalProtFlag = cli.UintFlag{
		Name:  "localrpcport",
		Usage: "Json rpc local server listening port `<number>`",
		Value: config.DEFAULT_RPC_LOCAL_PORT,
	}

	//Websocket setting
	WsEnabledFlag = cli.BoolFlag{
		Name:  "ws",
		Usage: "Enable web socket server",
	}
	WsPortFlag = cli.UintFlag{
		Name:  "wsport",
		Usage: "Ws server listening port `<number>`",
		Value: config.DEFAULT_WS_PORT,
	}

	//Restful setting
	RestfulEnableFlag = cli.BoolFlag{
		Name:  "rest",
		Usage: "Enable restful api server",
	}
	RestfulPortFlag = cli.UintFlag{
		Name:  "restport",
		Usage: "Restful server listening port `<number>`",
		Value: config.DEFAULT_REST_PORT,
	}
	RestfulMaxConnsFlag = cli.UintFlag{
		Name:  "restmaxconns",
		Usage: "Restful server maximum connections `<number>`",
		Value: config.DEFAULT_REST_MAX_CONN,
	}

	//Account setting
	AccountPassFlag = cli.StringFlag{
		Name:   "password,p",
		Hidden: true,
		Usage:  "Account `<password>` when Ontology node starts.",
	}
	AccountAddressFlag = cli.StringFlag{
		Name:  "account,a",
		Usage: "Account `<address>` when the Ontology node starts. If not specific, using default account instead",
	}
	AccountDefaultFlag = cli.BoolFlag{
		Name:  "default,d",
		Usage: "Default settings to create a new account (equal to '-t ecdsa -b 256 -s SHA256withECDSA')",
	}
	AccountTypeFlag = cli.StringFlag{
		Name:  "type,t",
		Usage: "Specifies the `<key-type>` by signature algorithm.",
	}
	AccountKeylenFlag = cli.StringFlag{
		Name:  "bit-length,b",
		Usage: "Specifies the `<bit-length>` of key",
	}
	AccountSigSchemeFlag = cli.StringFlag{
		Name:  "signature-scheme,s",
		Usage: "Specifies the signature scheme `<scheme>`",
	}
	AccountSetDefaultFlag = cli.BoolFlag{
		Name:  "as-default,d",
		Usage: "Set the specified account to default account of wallet",
	}
	AccountQuantityFlag = cli.UintFlag{
		Name:  "number,n",
		Value: 1,
		Usage: "Specifies the `<quantity>` of account to create.",
	}
	AccountSourceFileFlag = cli.StringFlag{
		Name:  "source,s",
		Usage: "Source wallet `<file>` to import",
	}
	AccountLabelFlag = cli.StringFlag{
		Name:  "label,l",
		Usage: "Set account `<label>` for easy and fast use of accounts.",
	}
	AccountKeyFlag = cli.StringFlag{
		Name:  "key,k",
		Usage: "Use `<private key>` (hex encoding) of the account",
	}
	AccountVerboseFlag = cli.BoolFlag{
		Name:  "verbose,v",
		Usage: "Display accounts with details",
	}
	AccountChangePasswdFlag = cli.BoolFlag{
		Name:  "change-passwd",
		Usage: "Change account password",
	}
	AccountLowSecurityFlag = cli.BoolFlag{
		Name:  "low-security",
		Usage: "Change account to low protection strength for low performance devices",
	}
	AccountWIFFlag = cli.BoolFlag{
		Name:  "wif",
		Usage: "Import WIF keys from the source file specified by --source option",
	}
	AccountMultiMFlag = cli.UintFlag{
		Name:  "m",
		Usage: "Min signature `<number>` of multi signature address",
		Value: 1,
	}
	AccountMultiPubKeyFlag = cli.StringFlag{
		Name:  "pubkey",
		Usage: "Pub key list of multi `<addresses>`, separate addreses with comma `,`",
	}
	IdentityFlag = cli.BoolFlag{
		Name:  "ontid",
		Usage: "create an ONT ID instead of account",
	}

	//SmartContract setting
	ContractAddrFlag = cli.StringFlag{
		Name:  "address",
		Usage: "Contract `<address>`",
	}
	ContractVmTypeFlag = cli.UintFlag{
		Name:  "vmtype",
		Usage: "The Contract type: 1 for Neovm ,3 for Wasmvm",
		Value: 1,
	}
	ContractCodeFileFlag = cli.StringFlag{
		Name:  "code",
		Usage: "File path of contract code `<path>`",
	}
	ContractNameFlag = cli.StringFlag{
		Name:  "name",
		Usage: "Specifies contract name to `<name>`",
	}
	ContractVersionFlag = cli.StringFlag{
		Name:  "version",
		Usage: "Specifies contract version to `<ver>`",
	}
	ContractAuthorFlag = cli.StringFlag{
		Name:  "author",
		Usage: "Set `<address>` as the contract owner",
		Value: "",
	}
	ContractEmailFlag = cli.StringFlag{
		Name:  "email",
		Usage: "Set `<email>` owner's email address",
		Value: "",
	}
	ContractDescFlag = cli.StringFlag{
		Name:  "desc",
		Usage: "Set `<text>` as the description of the contract",
		Value: "",
	}
	ContractParamsFlag = cli.StringFlag{
		Name:  "params",
		Usage: "Contract parameters list to invoke. separate params with comma ','",
	}
	ContractPrepareDeployFlag = cli.BoolFlag{
		Name:  "prepare,p",
		Usage: "Prepare deploy contract without commit to ledger",
	}
	ContractPrepareInvokeFlag = cli.BoolFlag{
		Name:  "prepare,p",
		Usage: "Prepare invoke contract without commit to ledger",
	}
	ContractReturnTypeFlag = cli.StringFlag{
		Name:  "return",
		Usage: "Return `<type>` of contract. bytearray(hexstring), string, int, boolean",
	}

	//information cmd settings
	BlockHashInfoFlag = cli.StringFlag{
		Name:  "hash",
		Usage: "Get block info by block hash",
	}
	BlockHeightInfoFlag = cli.UintFlag{
		Name:  "height",
		Usage: "Get block info by block height",
	}

	//Transfer setting
	TransactionAssetFlag = cli.StringFlag{
		Name:  "asset",
		Usage: "Asset of ONT or ONG",
		Value: ASSET_ONT,
	}
	TransactionFromFlag = cli.StringFlag{
		Name:  "from",
		Usage: "Transfer-out account `<address>`",
	}
	TransactionToFlag = cli.StringFlag{
		Name:  "to",
		Usage: "Transfer-in account `<address>`",
	}
	TransactionAmountFlag = cli.StringFlag{
		Name:  "amount",
		Usage: "Transfer `<amount>`. Float number",
	}
	TransactionHashFlag = cli.StringFlag{
		Name:  "hash",
		Usage: "Transaction `<hash>`",
	}
	TransactionGasPriceFlag = cli.Uint64Flag{
		Name:  "gasprice",
		Usage: "Gas price of transaction",
		Value: config.DEFAULT_GAS_PRICE,
	}
	TransactionGasLimitFlag = cli.Uint64Flag{
		Name:  "gaslimit",
		Usage: "Gas limit of the transaction",
		Value: neovm.MIN_TRANSACTION_GAS,
	}
	TransactionPayerFlag = cli.StringFlag{
		Name:  "payer",
		Usage: "Transaction fee payer `<address>`,Default is the signer address",
	}

	//Asset setting
	ApproveAssetFromFlag = cli.StringFlag{
		Name:  "from",
		Usage: "Transfer-out account `<address>`",
	}
	ApproveAssetToFlag = cli.StringFlag{
		Name:  "to",
		Usage: "Transfer-in account `<address>`",
	}
	ApproveAssetFlag = cli.StringFlag{
		Name:  "asset",
		Usage: "Asset of ONT of ONG to approve",
		Value: "ont",
	}
	ApproveAmountFlag = cli.StringFlag{
		Name:  "amount",
		Usage: "Amount of approve. Float number",
	}
	TransferFromAmountFlag = cli.StringFlag{
		Name:  "amount",
		Usage: "Amount of transfer from. Float number",
	}
	TransferFromSenderFlag = cli.StringFlag{
		Name:  "sender",
		Usage: "Sender account `<address>` of transfer from transaction, Default is transfer-to account",
	}
	SendTxFlag = cli.BoolFlag{
		Name:  "send,s",
		Usage: "Send raw transaction to Ontology",
	}
	ForceSendTxFlag = cli.BoolFlag{
		Name:  "force,f",
		Usage: "Force to send transaction",
	}
	RawTransactionFlag = cli.StringFlag{
		Name:  "raw-tx",
		Usage: "Raw `<transaction>` encode with hex string",
	}
	PrepareExecTransactionFlag = cli.BoolFlag{
		Name:  "prepare,p",
		Usage: "Prepare execute transaction, without commit to ledger",
	}
	WithdrawONGReceiveAccountFlag = cli.StringFlag{
		Name:  "receive",
		Usage: "ONG receive `<address>`，Default the same with owner account",
	}
	WithdrawONGAmountFlag = cli.StringFlag{
		Name:  "amount",
		Usage: "Withdraw amount `<number>`, Float number. Default withdraw all",
	}
	ForceTxFlag = cli.BoolFlag{
		Name:  "force,f",
		Usage: "Force to send transaction",
	}

	//Cli setting
	CliAddressFlag = cli.StringFlag{
		Name:  "cliaddress",
		Usage: "Rpc bind `<address>`",
		Value: config.DEFUALT_CLI_RPC_ADDRESS,
	}
	CliRpcPortFlag = cli.UintFlag{
		Name:  "cliport",
		Usage: "Rpc bind port `<number>`",
		Value: config.DEFAULT_CLI_RPC_PORT,
	}
	CliABIPathFlag = cli.StringFlag{
		Name:  "abi",
		Usage: "Abi `<file>` path",
		Value: DEFAULT_ABI_PATH,
	}
	CliWalletDirFlag = cli.StringFlag{
		Name:  "walletdir",
		Usage: "Wallet data `<path>`",
		Value: DEFAULT_WALLET_PATH,
	}

	//Export setting
	ExportFileFlag = cli.StringFlag{
		Name:  "export-file",
		Usage: "Export `<file>` path",
		Value: DEFAULT_EXPORT_FILE,
	}
	ExportStartHeightFlag = cli.UintFlag{
		Name:  "start-height",
		Usage: "Start block height `<number>` to export",
		Value: DEFAULT_EXPORT_HEIGHT,
	}
	ExportEndHeightFlag = cli.UintFlag{
		Name:  "end-height",
		Usage: "Stop block height `<number>` to export",
		Value: DEFAULT_EXPORT_HEIGHT,
	}
	ExportSpeedFlag = cli.StringFlag{
		Name:  "export-speed",
		Usage: "Export block speed `<level>` (h|m|l), h for high speed, m for middle speed and l for low speed",
		Value: "m",
	}

	//PreExecute switcher
	TxpoolPreExecDisableFlag = cli.BoolFlag{
		Name:  "disable-tx-pool-pre-exec",
		Usage: "Disable preExecute in tx pool",
	}

	//local PreExecute switcher
	DisableSyncVerifyTxFlag = cli.BoolFlag{
		Name:  "disable-sync-verify-tx",
		Usage: "Disable sync verify transaction in interface",
	}

	DisableBroadcastNetTxFlag = cli.BoolFlag{
		Name:  "disable-broadcast-net-tx",
		Usage: "Disable broadcast tx from network in tx pool",
	}

	NonOptionFlag = cli.StringFlag{
		Name:  "option",
		Usage: "this command does not need option, please run directly",
	}
)

//GetFlagName deal with short flag, and return the flag name whether flag name have short name
func GetFlagName(flag cli.Flag) string {
	name := flag.GetName()
	if name == "" {
		return ""
	}
	return strings.TrimSpace(strings.Split(name, ",")[0])
}
