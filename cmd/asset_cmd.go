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

package cmd

import (
	"fmt"
	"github.com/ontio/ontology/account"
	cmdCom "github.com/ontio/ontology/cmd/common"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/urfave/cli"
)

var AssetCommand = cli.Command{
	Name:         "asset",
	Usage:        "Handle assets",
	OnUsageError: cmdCom.CommonCommandErrorHandler,
	Description:  `asset control`,
	Subcommands: []cli.Command{
		{
			Action:       transfer,
			OnUsageError: cmdCom.CommonCommandErrorHandler,
			Name:         "transfer",
			Usage:        "Transfer ont to another account",
			ArgsUsage:    " ",
			Description:  "Transfer ont to another account. If from address doesnot specific, using default account",
			Flags: []cli.Flag{
				utils.TransactionFromFlag,
				utils.TransactionToFlag,
				utils.TransactionAmountFlag,
				utils.WalletFileFlag,
			},
		},
		{
			Action:       getBalance,
			OnUsageError: cmdCom.CommonCommandErrorHandler,
			Name:         "balance",
			Usage:        "Show balance of ont and ong of specified account",
			ArgsUsage:    "[address]",
			Flags: []cli.Flag{
				utils.AccountAddressFlag,
				utils.WalletFileFlag,
			},
		},
		{
			Action:       queryTransferStatus,
			OnUsageError: cmdCom.CommonCommandErrorHandler,
			Name:         "status",
			Usage:        "Display asset status",
			ArgsUsage:    "[address]",
			Description:  `Display asset transfer status of [address] or the default account if not specified.`,
			Flags: []cli.Flag{
				utils.TransactionHashFlag,
			},
		},
	},
}

func transfer(ctx *cli.Context) error {
	if !ctx.IsSet(utils.TransactionToFlag.Name) || !ctx.IsSet(utils.TransactionAmountFlag.Name) {
		return fmt.Errorf("Missing argument to or amount")
	}

	from := ctx.String(utils.TransactionFromFlag.Name)
	to := ctx.String(utils.TransactionToFlag.Name)
	amount := ctx.Uint(utils.TransactionAmountFlag.Name)

	wallet, err := cmdCom.OpenWallet(ctx)
	if err != nil {
		return fmt.Errorf("OpenWallet error:%s", err)
	}
	var signer *account.Account
	if from == "" {
		signer = wallet.GetDefaultAccount()
		if signer == nil {
			return fmt.Errorf("Please specific from address correctly")
		}
	} else {
		fromAddr, err := common.AddressFromBase58(from)
		if err != nil {
			return fmt.Errorf("Invalid from address:%s", from)
		}
		signer = wallet.GetAccountByAddress(fromAddr)
		if signer == nil {
			return fmt.Errorf("Cannot found account by address:%s", from)
		}
	}

	txHash, err := utils.Transfer(signer, to, amount)
	if err != nil {
		return fmt.Errorf("Transfer error:%s", err)
	}
	fmt.Printf("Transfer ONT\n")
	fmt.Printf("From:%s\n", signer.Address.ToBase58())
	fmt.Printf("To:%s\n", to)
	fmt.Printf("Amount:%d\n", amount)
	fmt.Printf("TxHash:%s\n", txHash)
	return nil
}

func getBalance(ctx *cli.Context) error {
	address := ""
	if ctx.IsSet(utils.AccountAddressFlag.Name) {
		address = ctx.String(utils.AccountAddressFlag.Name)
	}
	if address == "" {
		wallet, err := cmdCom.OpenWallet(ctx)
		if err != nil {
			return fmt.Errorf("OpenWallet error:%s", err)
		}
		defaultAcc := wallet.GetDefaultAccount()
		if defaultAcc == nil {
			return fmt.Errorf("GetDefaultAccount failed")
		}
		address = defaultAcc.Address.ToBase58()
	}
	balance, err := utils.GetBalance(address)
	if err != nil {
		return err
	}
	fmt.Printf("BalanceOf:%s\n", address)
	fmt.Printf("ONT:%s\n", balance.Ont)
	fmt.Printf("ONG:%s\n", balance.Ong)
	fmt.Printf("ONGApprove:%s\n", balance.OngAppove)
	return nil
}

func queryTransferStatus(ctx *cli.Context) error {
	if !ctx.IsSet(utils.TransactionHashFlag.Name) {
		return fmt.Errorf("Missing hash argument")
	}
	txHash := ctx.String(utils.TransactionHashFlag.Name)
	evtInfos, err := utils.GetSmartContractEvent(txHash)
	if err != nil {
		return fmt.Errorf("GetSmartContractEvent error:%s", err)
	}
	if len(evtInfos) == 0 {
		fmt.Println("Cannot find event log")
		return nil
	}
	for _, eventInfo := range evtInfos {
		states := eventInfo.States.([]interface{})
		fmt.Printf("Transaction:%s success\n", states[0])
		fmt.Printf("From:%s\n", states[1])
		fmt.Printf("To:%s\n", states[2])
		fmt.Printf("Amount:%v\n", states[3])
	}
	return nil
}
