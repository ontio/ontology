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
	"github.com/urfave/cli"
)

var (
	ConfigUsedFlag = cli.StringFlag{
		Name:  "config",
		Usage: "Use `<filename>` as the config file",
	}

	// RPC settings
	RPCEnabledFlag = cli.BoolFlag{
		Name:  "rpc",
		Usage: "Enable rpc server",
	}

	WsEnabledFlag = cli.BoolFlag{
		Name:  "ws",
		Usage: "Enable websocket server",
	}

	//information cmd settings
	HashInfoFlag = cli.StringFlag{
		Name:  "hash",
		Usage: "transaction or block hash value",
	}

	HeightInfoFlag = cli.StringFlag{
		Name:  "height",
		Usage: "block height value",
	}

	//send raw transaction
	ContractAddrFlag = cli.StringFlag{
		Name:  "caddr",
		Usage: "contract `<address>` of the asset",
	}

	TransactionFromFlag = cli.StringFlag{
		Name:  "from",
		Usage: "`<address>` which sends the asset",
	}
	TransactionToFlag = cli.StringFlag{
		Name:  "to",
		Usage: "`<address>` which receives the asset",
	}
	TransactionValueFlag = cli.Int64Flag{
		Name:  "value",
		Usage: "Specifies `<value>` as the transferred amount",
	}

	DebugLevelFlag = cli.UintFlag{
		Name:  "debuglevel",
		Usage: "Set the log level to `<level>` (0~6)",
	}

	ConsensusFlag = cli.StringFlag{
		Name:  "consensus",
		Usage: "Turn `<on | off>` the consensus",
	}

	//contract deploy
	ContractVmTypeFlag = cli.StringFlag{
		Name:  "type",
		Value: "neovm",
		Usage: "Specifies contract type to one of `<neovm|wasm>`",
	}

	ContractStorageFlag = cli.BoolFlag{
		Name:  "store",
		Usage: "Store the contract",
	}

	ContractCodeFlag = cli.StringFlag{
		Name:  "code",
		Usage: "Input contracts from `<path>`",
	}

	ContractNameFlag = cli.StringFlag{
		Name:  "cname",
		Usage: "Specifies contract name to `<name>`",
	}

	ContractVersionFlag = cli.StringFlag{
		Name:  "cversion",
		Usage: "Specifies contract version to `<ver>`",
	}

	ContractAuthorFlag = cli.StringFlag{
		Name:  "author",
		Usage: "Set `<address>` as the contract owner",
	}

	ContractEmailFlag = cli.StringFlag{
		Name:  "email",
		Usage: "Set `<email>` owner's email address",
	}

	ContractDescFlag = cli.StringFlag{
		Name:  "desc",
		Usage: "Set `<text>` as the description of the contract",
	}

	ContractParamsFlag = cli.StringFlag{
		Name:  "params",
		Usage: "Specifies contract parameters `<list>` when invoked",
	}
	NonOptionFlag = cli.StringFlag{
		Name:  "option",
		Usage: "this command does not need option, please run directly",
	}
	//account management
	AccountQuantityFlag = cli.UintFlag{
		Name:  "number,n",
		Usage: "Specifies the `<quantity>` of account to create, default is 1.",
	}
	AccountVerboseFlag = cli.BoolFlag{
		Name:  "verbose,v",
		Usage: "Display accounts with details",
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
	AccountPassFlag = cli.StringFlag{
		Name:   "password,p",
		Hidden: true,
		Usage:  "Specifies `<password>` for the account",
	}
	AccountDefaultFlag = cli.BoolFlag{
		Name:  "default,d",
		Usage: "Use default settings (equal to '-t ecdsa -b 256 -s SHA256withECDSA')",
	}
	AccountSetDefaultFlag = cli.BoolFlag{
		Name:  "as-default,d",
		Usage: "Set the specified account to default",
	}
	AccountFileFlag = cli.StringFlag{
		Name:  "file,f",
		Value: "wallet.dat",
		Usage: "Use `<filename>` as the wallet",
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
