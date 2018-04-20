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
	EncryptTypeFlag = cli.StringFlag{
		Name: "encrypt",
		Usage: `assign encrypt type when use create wallet, just as:
						SHA224withECDSA, SHA256withECDSA,
						SHA384withECDSA, SHA512withECDSA,
						SHA3-224withECDSA, SHA3-256withECDSA,
						SHA3-384withECDSA, SHA3-512withECDSA,
						RIPEMD160withECDSA, SM3withSM2, SHA512withEdDSA`,
	}

	WalletAddrFlag = cli.StringFlag{
		Name:  "addr",
		Usage: "wallet address string(base58)",
	}

	WalletNameFlag = cli.StringFlag{
		Name:  "name",
		Usage: "wallet name",
	}

	WalletUsedFlag = cli.StringFlag{
		Name:  "wallet",
		Usage: "which wallet will be used",
	}

	ConfigUsedFlag = cli.StringFlag{
		Name:  "config",
		Usage: "which config file will be used",
	}

	// RPC settings
	RPCEnabledFlag = cli.BoolFlag{
		Name:  "rpc",
		Usage: "enable rpc server? true or false",
	}

	WsEnabledFlag = cli.BoolFlag{
		Name:  "ws",
		Usage: "enable websocket server? true or false",
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
		Usage: "contract address that will be used when send raw transaction",
	}

	TransactionFromFlag = cli.StringFlag{
		Name:  "from",
		Usage: "address which transfer from",
	}
	TransactionToFlag = cli.StringFlag{
		Name:  "to",
		Usage: "address which transfer to",
	}
	TransactionValueFlag = cli.Int64Flag{
		Name:  "value",
		Usage: "transfer value",
	}
	UserPasswordFlag = cli.StringFlag{
		Name:  "password",
		Usage: "used when transfer",
	}

	DebugLevelFlag = cli.UintFlag{
		Name:  "debuglevel",
		Usage: "debug level(0~6) will be set",
	}

	ConsensusFlag = cli.StringFlag{
		Name:  "consensus",
		Usage: "consensus turn on/off",
	}

	//contract deploy
	ContractVmTypeFlag = cli.UintFlag{
		Name:  "type",
		Usage: "contract type ,value: NEOVM | WASM",
	}

	ContractStorageFlag = cli.BoolFlag{
		Name:  "store",
		Usage: "does this contract will be stored, value: true or false",
	}

	ContractCodeFlag = cli.StringFlag{
		Name:  "code",
		Usage: "directory of smart contract that will be deployed",
	}

	ContractNameFlag = cli.StringFlag{
		Name:  "cname",
		Usage: "contract name that will be deployed",
	}

	ContractVersionFlag = cli.StringFlag{
		Name:  "cversion",
		Usage: "contract version which will be deployed",
	}

	ContractAuthorFlag = cli.StringFlag{
		Name:  "author",
		Usage: "owner of deployed smart contract",
	}

	ContractEmailFlag = cli.StringFlag{
		Name:  "email",
		Usage: "owner email when deploy a smart contract",
	}

	ContractDescFlag = cli.StringFlag{
		Name:  "desc",
		Usage: "contract description when deploy one",
	}

	ContractParamsFlag = cli.StringFlag{
		Name:  "params",
		Usage: "contract parameter needed when invoded",
	}
	NonOptionFlag = cli.StringFlag{
		Name:  "optiion",
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
