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
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	cmdCom "github.com/ontio/ontology/cmd/common"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/password"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/signature"
	ctypes "github.com/ontio/ontology/core/types"
	cutils "github.com/ontio/ontology/core/utils"
	jrpc "github.com/ontio/ontology/http/base/rpc"
	nstates "github.com/ontio/ontology/smartcontract/service/native/states"
	"github.com/ontio/ontology/smartcontract/states"
	vmtypes "github.com/ontio/ontology/smartcontract/types"
	"github.com/urfave/cli"
)

var (
	AssetCommand = cli.Command{
		Name:         "asset",
		Action:       utils.MigrateFlags(assetCommand),
		Usage:        "Handle assets",
		OnUsageError: assetUsageError,
		Description:  `asset control`,
		Subcommands: []cli.Command{
			{
				Action:       transferAsset,
				OnUsageError: transferAssetUsageError,
				Name:         "transfer",
				Usage:        "Transfer asset to another account",
				ArgsUsage:    " ",
				Description:  `Transfer some asset to another account. Asset type is specified by its contract address. Default is the ont contract.`,
				Flags: []cli.Flag{
					utils.ContractAddrFlag,
					utils.TransactionFromFlag,
					utils.TransactionToFlag,
					utils.TransactionValueFlag,
					utils.AccountPassFlag,
					utils.AccountFileFlag,
				},
			},
			{
				Action:       queryTransferStatus,
				OnUsageError: transferAssetUsageError,
				Name:         "status",
				Usage:        "Display asset status",
				ArgsUsage:    "[address]",
				Description:  `Display asset transfer status of [address] or the default account if not specified.`,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "hash",
						Usage: "Specifies transaction hash `<hash>`",
					},
				},
			},
			{
				Action:       ontBalance,
				OnUsageError: balanceUsageError,
				Name:         "balance",
				Usage:        "Show balance of ont and ong of specified account",
				ArgsUsage:    "[address]",
				Flags: []cli.Flag{
					utils.AccountPassFlag,
					utils.AccountFileFlag,
				},
			},
		},
	}
)

func assetUsageError(context *cli.Context, err error, isSubcommand bool) error {
	fmt.Println(err.Error())
	fmt.Println("")
	cli.ShowSubcommandHelp(context)
	return nil
}

func assetCommand(ctx *cli.Context) error {
	fmt.Println("Error usage.\n")
	cli.ShowSubcommandHelp(ctx)
	return nil
}

func transferAssetUsageError(context *cli.Context, err error, isSubcommand bool) error {
	fmt.Println(err.Error())
	fmt.Println("")
	cli.ShowSubcommandHelp(context)
	return nil
}

func balanceUsageError(context *cli.Context, err error, isSubcommand bool) error {
	fmt.Println(err)
	fmt.Println("")
	cli.ShowSubcommandHelp(context)
	return nil
}

func signTransaction(signer *account.Account, tx *ctypes.Transaction) error {
	hash := tx.Hash()
	sign, _ := signature.Sign(signer, hash[:])
	tx.Sigs = append(tx.Sigs, &ctypes.Sig{
		PubKeys: []keypair.PublicKey{signer.PublicKey},
		M:       1,
		SigData: [][]byte{sign},
	})
	return nil
}

func transferAsset(ctx *cli.Context) error {
	if !ctx.IsSet(utils.TransactionFromFlag.Name) || !ctx.IsSet(utils.TransactionToFlag.Name) || !ctx.IsSet(utils.TransactionValueFlag.Name) {
		fmt.Println("Missing argument.\n")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	ctu := genesis.OntContractAddress
	if ctx.IsSet(utils.ContractAddrFlag.Name) {
		contract := ctx.String(utils.ContractAddrFlag.Name)
		ct, err := common.HexToBytes(contract)
		if err != nil {
			fmt.Println("Parase contract address error, from hex to bytes")
			return err
		}

		ctu, err = common.AddressParseFromBytes(ct)
		if err != nil {
			fmt.Println("Parase contract address error, please use correct smart contract address")
			return err
		}
	}

	from := ctx.String(utils.TransactionFromFlag.Name)
	fu, err := common.AddressFromBase58(from)
	if err != nil {
		fmt.Println("Parase transfer-from address error, make sure you are using base58 address")
		return err
	}

	to := ctx.String(utils.TransactionToFlag.Name)
	tu, err := common.AddressFromBase58(to)
	if err != nil {
		fmt.Println("Parase transfer-to address error, make sure you are using base58 address")
		return err
	}

	value := ctx.Int64("value")
	if value <= 0 {
		fmt.Println("Value must be int type and bigger than zero. Invalid ont amount: ", value)
		return errors.New("Value is invalid")
	}

	var passwd []byte
	var filename string = account.WALLET_FILENAME
	if ctx.IsSet("file") {
		filename = ctx.String("file")
	}
	if !common.FileExisted(filename) {
		fmt.Println(filename, "not found.")
		return errors.New("Asset transfer failed.")
	}
	if ctx.IsSet("password") {
		passwd = []byte(ctx.String("password"))
	} else {
		passwd, err = password.GetAccountPassword()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return errors.New("input password error")
		}
	}
	acct := account.Open(filename, passwd)
	for i, _ := range passwd {
		passwd[i] = 0
	}
	if nil == acct {
		fmt.Println("Open account failed, please check your input password and make sure your wallet.dat exist")
		return errors.New("Get Account Error")
	}

	acc := acct.GetAccountByAddress(fu)
	if nil == acc {
		fmt.Println("Get account by address error")
		return errors.New("Get Account Error")
	}

	var sts []*nstates.State
	sts = append(sts, &nstates.State{
		From:  fu,
		To:    tu,
		Value: big.NewInt(value),
	})
	transfers := &nstates.Transfers{
		States: sts,
	}
	bf := new(bytes.Buffer)

	if err := transfers.Serialize(bf); err != nil {
		fmt.Println("Serialize transfers struct error.")
		return err
	}

	cont := &states.Contract{
		Address: ctu,
		Method:  "transfer",
		Args:    bf.Bytes(),
	}

	ff := new(bytes.Buffer)

	if err := cont.Serialize(ff); err != nil {
		fmt.Println("Serialize contract struct error.")
		return err
	}

	tx := cutils.NewInvokeTransaction(vmtypes.VmCode{
		VmType: vmtypes.Native,
		Code:   ff.Bytes(),
	})

	tx.Nonce = uint32(time.Now().Unix())

	if err := signTransaction(acc, tx); err != nil {
		fmt.Println("signTransaction error:", err)
		return err
	}

	txbf := new(bytes.Buffer)
	if err := tx.Serialize(txbf); err != nil {
		fmt.Println("Serialize transaction error.")
		return err
	}

	resp, err := jrpc.Call(rpcAddress(), "sendrawtransaction", 0,
		[]interface{}{hex.EncodeToString(txbf.Bytes())})

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	r := make(map[string]interface{})
	err = json.Unmarshal(resp, &r)
	if err != nil {
		fmt.Println("Unmarshal JSON failed")
		return err
	}

	switch r["result"].(type) {
	case map[string]interface{}:

	case string:
		time.Sleep(10 * time.Second)
		resp, err := ontSdk.Rpc.GetSmartContractEventWithHexString(r["result"].(string))
		if err != nil {
			fmt.Printf("Please query transfer status manually by hash :%s", r["result"].(string))
			return err
		}
		fmt.Println("\nAsset Transfer Result:")
		cmdCom.EchoJsonDataGracefully(resp)
		return nil
	}

	fmt.Printf("Please query transfer status manually by hash :%s", r["result"].(string))
	return nil
}

func queryTransferStatus(ctx *cli.Context) error {
	if !ctx.IsSet("hash") {
		fmt.Println("Missing transaction hash.")
		cli.ShowSubcommandHelp(ctx)
	}

	trHash := ctx.String("hash")
	resp, err := ontSdk.Rpc.GetSmartContractEventWithHexString(trHash)
	if err != nil {
		fmt.Println("Parase contract address error, from hex to bytes")
		return err
	}
	cmdCom.EchoJsonDataGracefully(resp)
	return nil
}

func ontBalance(ctx *cli.Context) error {
	var filename string = account.WALLET_FILENAME
	if ctx.IsSet("file") {
		filename = ctx.String("file")
	}

	var base58Addr string
	if ctx.NArg() == 0 {
		var passwd []byte
		var err error
		if ctx.IsSet("password") {
			passwd = []byte(ctx.String("password"))
		} else {
			passwd, err = password.GetAccountPassword()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return errors.New("input password error")
			}
		}
		acct := account.Open(filename, passwd)
		for i, _ := range passwd {
			passwd[i] = 0
		}
		if acct == nil {
			return errors.New("open wallet error")
		}
		dac := acct.GetDefaultAccount()
		if dac == nil {
			return errors.New("cannot get the default account")
		}
		base58Addr = dac.Address.ToBase58()
	} else {
		base58Addr = ctx.Args().First()
	}
	balance, err := ontSdk.Rpc.GetBalanceWithBase58(base58Addr)
	if nil != err {
		fmt.Printf("Get Balance with base58 err: %s", err.Error())
		return err
	}
	fmt.Printf("ONT: %d; ONG: %d; ONGAppove: %d\n Address(base58): %s\n", balance.Ont.Int64(), balance.Ong.Int64(), balance.OngAppove.Int64(), base58Addr)
	return nil
}
