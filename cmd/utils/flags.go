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
	"github.com/urfave/cli"
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
	MaxTxInBlockFlag = cli.IntFlag{
		Name:  "maxtxinblock",
		Usage: "Max transaction number in block",
		Value: config.DEFAULT_MAX_TX_IN_BLOCK,
	}
	DisableEventLogFlag = cli.BoolFlag{
		Name:  "disableeventlog",
		Usage: "If set disableeventlog flag, will not record event log output by smart contract",
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
	RPCEnabledFlag = cli.BoolFlag{
		Name:  "rpc",
		Usage: "Enable Json rpc server",
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
	WalletFileFlag = cli.StringFlag{
		Name:  "wallet",
		Value: config.DEFAULT_WALLET_FILE_NAME,
		Usage: "Use `<filename>` as the wallet",
	}
	AccountPassFlag = cli.StringFlag{
		Name:   "password,p",
		Hidden: true,
		Usage:  "Specifies `<password>` for the account",
	}
	AccountAddressFlag = cli.StringFlag{
		Name:  "address",
		Usage: "Address of account, if not specific, will use default account",
	}
	AccountDefaultFlag = cli.BoolFlag{
		Name:  "default,d",
		Usage: "Use default settings (equal to '-t ecdsa -b 256 -s SHA256withECDSA')",
	}
	AccountTypeFlag = cli.StringFlag{
		Name:  "type,t",
		Usage: "Specifies the `<key-type>` by signature algorithm",
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
	AccountVerboseFlag = cli.BoolFlag{
		Name:  "verbose,v",
		Usage: "Display accounts with details",
	}

	//SmartContract setting
	ContractAddrFlag = cli.StringFlag{
		Name:  "address",
		Usage: "Contract address",
	}
	ContractVmTypeFlag = cli.StringFlag{
		Name:  "vmtype",
		Value: "neovm",
		Usage: "Specifies contract type to one of `<neovm|wasm>`",
	}
	ContractStorageFlag = cli.BoolFlag{
		Name:  "needstore",
		Usage: "Need use store in contract",
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
		Value: "",
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
		Usage: "Invoke contract parameters list. use comma ',' to split params, and must add type prefix to params.0:bytearray(hexstring), 1:string, 2:integer, 3:boolean,For example: 1:foo,2:0,3:true;If parameter is an object array, enclose array with '[]', For example:  1:foo,[2:0,3:true]",
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
	TransactionFromFlag = cli.StringFlag{
		Name:  "from",
		Usage: "`<address>` which sends the asset",
	}
	TransactionToFlag = cli.StringFlag{
		Name:  "to",
		Usage: "`<address>` which receives the asset",
	}
	TransactionAmountFlag = cli.Int64Flag{
		Name:  "amount",
		Usage: "Specifies `<amount>` as the transferred amount",
	}
	TransactionHashFlag = cli.StringFlag{
		Name:  "hash",
		Usage: "Transaction hash",
	}

	NonOptionFlag = cli.StringFlag{
		Name:  "option",
		Usage: "this command does not need option, please run directly",
	}
)

func MigrateFlags(action func(ctx *cli.Context) error) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		for _, name := range ctx.FlagNames() {
			if ctx.IsSet(name) {
				ctx.GlobalSet(name, ctx.String(name))
			}
		}
		return action(ctx)
	}
}
