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
		ArgsUsage:    "",
		OnUsageError: assetUsageError,
		Description:  `asset control`,
		Subcommands: []cli.Command{
			{
				Action:       utils.MigrateFlags(transferAsset),
				OnUsageError: transferAssetUsageError,
				Name:         "transfer",
				Usage:        "Transfer asset to another account",
				Flags:        append(NodeFlags, ContractFlags...),
				Description:  ``,
			},
			{
				Action:       utils.MigrateFlags(queryTransferStatus),
				OnUsageError: transferAssetUsageError,
				Name:         "status",
				Usage:        "Display asset status",
				Flags:        append(append(NodeFlags, ContractFlags...), InfoFlags...),
				Description:  ``,
			},
			{
				Action:       ontBalance,
				OnUsageError: balanceUsageError,
				Name:         "ont-balance",
				Usage:        "Show balance of ont and ong of specified account",
				ArgsUsage:    "[address]",
				Flags: []cli.Flag{
					utils.UserPasswordFlag,
					utils.AccountFileFlag,
				},
			},
		},
	}
)

func assetUsageError(context *cli.Context, err error, isSubcommand bool) error {
	fmt.Println(err.Error())
	cli.ShowSubcommandHelp(context)
	return nil
}

func assetCommand(ctx *cli.Context) error {
	showAssetHelp()
	return nil
}

func transferAssetUsageError(context *cli.Context, err error, isSubcommand bool) error {
	fmt.Println(err.Error())
	showAssetTransferHelp()
	return nil
}

func balanceUsageError(context *cli.Context, err error, isSubcommand bool) error {
	fmt.Println(err)
	showAssetTransferHelp()
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
	if !ctx.IsSet(utils.ContractAddrFlag.Name) || !ctx.IsSet(utils.TransactionFromFlag.Name) || !ctx.IsSet(utils.TransactionToFlag.Name) || !ctx.IsSet(utils.TransactionValueFlag.Name) {
		showAssetTransferHelp()
		return nil
	}
	contract := ctx.GlobalString(utils.ContractAddrFlag.Name)
	ct, err := common.HexToBytes(contract)
	if err != nil {
		fmt.Println("Parase contract address error, from hex to bytes")
		return err
	}

	ctu, err := common.AddressParseFromBytes(ct)
	if err != nil {
		fmt.Println("Parase contract address error, please use correct smart contract address")
		return err
	}

	from := ctx.GlobalString(utils.TransactionFromFlag.Name)
	fu, err := common.AddressFromBase58(from)
	if err != nil {
		fmt.Println("Parase transfer-from address error, make sure you are using base58 address")
		return err
	}

	to := ctx.GlobalString(utils.TransactionToFlag.Name)
	tu, err := common.AddressFromBase58(to)
	if err != nil {
		fmt.Println("Parase transfer-to address error, make sure you are using base58 address")
		return err
	}

	value := ctx.Int64(utils.TransactionValueFlag.Name)
	if value <= 0 {
		fmt.Println("Value must be int type and bigger than zero. Invalid ont amount: ", value)
		return errors.New("Value is invalid")
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

	var passwd []byte
	if ctx.IsSet(utils.UserPasswordFlag.Name) {
		passwd = []byte(ctx.GlobalString(utils.UserPasswordFlag.Name))
	} else {
		passwd, err = password.GetAccountPassword()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return errors.New("input password error")
		}
	}

	var filename string = account.WALLET_FILENAME
	if ctx.IsSet(utils.AccountFileFlag.Name) {
		filename = ctx.String(utils.AccountFileFlag.Name)
	}
	acct := account.Open(filename, passwd)
	acc := acct.GetDefaultAccount()

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
	if !ctx.IsSet(utils.HashInfoFlag.Name) {
		showQueryAssetTransferHelp()
	}

	trHash := ctx.GlobalString(utils.HashInfoFlag.Name)
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
	if ctx.IsSet(utils.AccountFileFlag.Name) {
		filename = ctx.String(utils.AccountFileFlag.Name)
	}

	var base58Addr string
	if ctx.NArg() == 0 {
		var passwd []byte
		var err error
		if ctx.IsSet(utils.UserPasswordFlag.Name) {
			passwd = []byte(ctx.GlobalString(utils.UserPasswordFlag.Name))
		} else {
			passwd, err = password.GetAccountPassword()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return errors.New("input password error")
			}
		}
		acct := account.Open(filename, passwd)
		base58Addr = acct.GetDefaultAccount().Address.ToBase58()
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
