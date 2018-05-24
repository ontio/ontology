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
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/smartcontract/service/neovm"
	"github.com/urfave/cli"
	"strings"
)

var (
	//Ontology setting
	ConfigFlag = cli.StringFlag{
		Name:  "config",
		Usage: "Use `<filename>` as the genesis config file",
		Value: config.DEFAULT_CONFIG_FILE_NAME,
	}
	LogLevelFlag = cli.UintFlag{
		Name:  "loglevel",
		Usage: "Set the log level to `<level>` (0~6). 0:Debug 1:Info 2:Warn 3:Error 4:Fatal 5:Trace 6:MaxLevel",
		Value: config.DEFAULT_LOG_LEVEL,
	}
	DisableEventLogFlag = cli.BoolFlag{
		Name:  "disableeventlog",
		Usage: "If set disableeventlog flag, will not record event log output by smart contract",
	}
	WalletFileFlag = cli.StringFlag{
		Name:  "wallet,w",
		Value: config.DEFAULT_WALLET_FILE_NAME,
		Usage: "Use `<filename>` as the wallet",
	}

	//Consensus setting
	DisableConsensusFlag = cli.BoolFlag{
		Name:  "disableconsensus",
		Usage: "If set disableconsensus, will not start consensus module",
	}
	MaxTxInBlockFlag = cli.IntFlag{
		Name:  "maxtxinblock",
		Usage: "Max transaction number in block",
		Value: config.DEFAULT_MAX_TX_IN_BLOCK,
	}
	GasLimitFlag = cli.Uint64Flag{
		Name:  "gaslimit",
		Usage: "The gas limit required by the transaction pool for a new transaction",
		Value: config.DEFAULT_GAS_LIMIT,
	}
	GasPriceFlag = cli.Uint64Flag{
		Name:  "gasprice",
		Usage: "The gas price enforced by the transaction pool for a new transaction",
		Value: config.DEFAULT_GAS_PRICE,
	}

	//Test Mode setting
	EnableTestModeFlag = cli.BoolFlag{
		Name:  "testmode",
		Usage: "Runin test mode will start solo consensus. If enable testmode flag, ontology will ignore the consensus type config in genesis",
	}
	TestModeGenBlockTimeFlag = cli.UintFlag{
		Name:  "testmodegenblocktime",
		Usage: "Interval of generate block in test mode, unit(s)",
		Value: config.DEFAULT_GEN_BLOCK_TIME,
	}

	//P2P setting
	NodePortFlag = cli.UintFlag{
		Name:  "nodeport",
		Usage: "P2P node listening port",
		Value: config.DEFAULT_NODE_PORT,
	}
	DualPortSupportFlag = cli.BoolFlag{
		Name:  "dualport",
		Usage: "Enable dual port support. Means p2p node port difference with consensus port",
	}
	ConsensusPortFlag = cli.UintFlag{
		Name:  "consensusport",
		Usage: "Consensus listening port",
		Value: config.DEFAULT_CONSENSUS_PORT,
	}

	// RPC settings
	RPCDisabledFlag = cli.BoolFlag{
		Name:  "disablerpc",
		Usage: "Disable Json rpc server",
	}
	RPCPortFlag = cli.UintFlag{
		Name:  "rpcport",
		Usage: "Json rpc server listening port",
		Value: config.DEFAULT_RPC_PORT,
	}
	RPCLocalEnableFlag = cli.BoolFlag{
		Name:  "localrpc",
		Usage: "Enable local rpc server",
	}
	RPCLocalProtFlag = cli.UintFlag{
		Name:  "rpclocalport",
		Usage: "Json rpc local server listening port",
		Value: config.DEFAULT_RPC_LOCAL_PORT,
	}

	//Websocket setting
	WsEnabledFlag = cli.BoolFlag{
		Name:  "ws",
		Usage: "Enable websocket server",
	}
	WsPortFlag = cli.UintFlag{
		Name:  "wsport",
		Usage: "Ws server listening port",
		Value: config.DEFAULT_WS_PORT,
	}

	//Restful setting
	RestfulEnableFlag = cli.BoolFlag{
		Name:  "rest",
		Usage: "Enable restful api server",
	}
	RestfulPortFlag = cli.UintFlag{
		Name:  "restport",
		Usage: "Restful server listening port",
		Value: config.DEFAULT_REST_PORT,
	}

	//Account setting
	AccountPassFlag = cli.StringFlag{
		Name:   "password,p",
		Hidden: true,
		Usage:  "Specifies `<password>` for the account",
	}
	AccountAddressFlag = cli.StringFlag{
		Name:  "account,a",
		Usage: "Address of account. Address can be `<address>` in base58 code , label or index. if not specific, will use default account",
	}
	AccountDefaultFlag = cli.BoolFlag{
		Name:  "default,d",
		Usage: "Use default settings (equal to '-t ecdsa -b 256 -s SHA256withECDSA')",
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
		Usage: "Set the specified account to default",
	}
	AccountQuantityFlag = cli.UintFlag{
		Name:  "number,n",
		Value: 1,
		Usage: "Specifies the `<quantity>` of account to create.",
	}
	AccountSourceFileFlag = cli.StringFlag{
		Name:  "source,s",
		Usage: "Use `<filename>` as the source wallet file to import",
	}
	AccountLabelFlag = cli.StringFlag{
		Name:  "label,l",
		Usage: "Use `<label>` for the account",
	}
	AccountKeyFlag = cli.StringFlag{
		Name:  "key,k",
		Usage: "Use `<private key(hex encoding)>` of the account",
	}
	AccountVerboseFlag = cli.BoolFlag{
		Name:  "verbose,v",
		Usage: "Display accounts with details",
	}
	AccountChangePasswdFlag = cli.BoolFlag{
		Name:  "changepasswd",
		Usage: "Change account password",
	}

	//SmartContract setting
	ContractAddrFlag = cli.StringFlag{
		Name:  "address",
		Usage: "Contract address",
	}
	ContractStorageFlag = cli.BoolFlag{
		Name:  "needstore",
		Usage: "Is need use store in contract",
	}
	ContractCodeFileFlag = cli.StringFlag{
		Name:  "code",
		Usage: "File path of contract code `<path>`",
	}
	ContractNameFlag = cli.StringFlag{
		Name:  "name",
		Usage: "Specifies contract name to `<name>`",
	}
	ContractVersionFlag = cli.IntFlag{
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
		Usage: "Invoke contract parameters list. use comma ',' to split params, and must add type prefix to params. Param type support bytearray(hexstring), string, integer, boolean,For example: string:foo,int:0,bool:true; If parameter is an object array, enclose array with '[]'. For example:  string:foo,[int:0,bool:true]",
	}
	ContractPrepareInvokeFlag = cli.BoolFlag{
		Name:  "prepare,p",
		Usage: "Prepare invoke contract without commit to ledger",
	}
	ContranctReturnTypeFlag = cli.StringFlag{
		Name:  "return",
		Usage: "Return type of contract.Return type support bytearray(hexstring), string, integer, boolean. If return type is object array, enclose array with '[]'. For example [string,int,bool,string]. Only prepare invoke need this flag.",
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
		Usage: "Asset to transfer `<ont|ong>`",
		Value: ASSET_ONT,
	}
	TransactionFromFlag = cli.StringFlag{
		Name:  "from",
		Usage: "`<address>` which sends the asset. address also can be index, label",
	}
	TransactionToFlag = cli.StringFlag{
		Name:  "to",
		Usage: "`<address>` which receives the asset",
	}
	TransactionAmountFlag = cli.StringFlag{
		Name:  "amount",
		Usage: "Specifies `<amount>` as the transferred amount",
	}
	TransactionHashFlag = cli.StringFlag{
		Name:  "hash",
		Usage: "Transaction <hash>",
	}
	TransactionGasPriceFlag = cli.Uint64Flag{
		Name:  "gasprice",
		Usage: "Gas price of transaction",
		Value: 0,
	}
	TransactionGasLimitFlag = cli.Uint64Flag{
		Name:  "gaslimit",
		Usage: "Gas limit of tansaction",
		Value: neovm.TRANSACTION_GAS,
	}

	//Asset setting
	ApproveAssetFromFlag = cli.StringFlag{
		Name:  "from",
		Usage: "Approve from address",
	}
	ApproveAssetToFlag = cli.StringFlag{
		Name:  "to",
		Usage: "Approve to address",
	}
	ApproveAssetFlag = cli.StringFlag{
		Name:  "asset",
		Usage: "Approve asset <ont|ong>",
		Value: "ont",
	}
	ApproveAmountFlag = cli.StringFlag{
		Name:  "amount",
		Usage: "Approve amount",
	}
	TransferFromAmountFlag = cli.StringFlag{
		Name:  "amount",
		Usage: "Transfer from amount",
	}
	TransferFromSenderFlag = cli.StringFlag{
		Name:  "sender",
		Usage: "Sender of transfer from transaction, if empty sender is to address",
	}

	//Cli setting
	CliEnableRpcFlag = cli.BoolFlag{
		Name:  "clirpc",
		Usage: "Enable cli rpc",
	}
	CliRpcPortFlag = cli.UintFlag{
		Name:  "cliport",
		Usage: "Cli rpc port",
		Value: config.DEFAULT_CLI_RPC_PORT,
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
