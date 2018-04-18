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
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
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
		Usage:        "ontology asset [transfer] [OPTION]",
		ArgsUsage:    "",
		Category:     "ASSET COMMANDS",
		OnUsageError: assetUsageError,
		Description:  `asset controll`,
		Subcommands: []cli.Command{
			{
				Action:       utils.MigrateFlags(transferAsset),
				OnUsageError: transferAssetUsageError,
				Name:         "transfer",
				Usage:        "ontology asset transfer [OPTION]\n",
				Flags:        append(NodeFlags, ContractFlags...),
				Category:     "ASSET COMMANDS",
				Description:  ``,
			},
		},
	}
)

func assetUsageError(context *cli.Context, err error, isSubcommand bool) error {
	fmt.Println(err.Error())
	showAssetHelp()
	return nil
}

func assetCommand(ctx *cli.Context) error {
	showAssetHelp()
	return nil
}

func showAssetHelp() {
	var assetHelp = `
   Name:
      ontology asset                       asset operation

   Usage:
      ontology asset [command options] [args]

   Description:
      With this command, you can control assert through transaction.

   Command:
      transfer
         --caddr     value                 smart contract address
         --from      value                 wallet address base58, which will transfer from
         --to        value                 wallet address base58, which will transfer to
         --value     value                 how much asset will be transfered
         --password  value                 use password who transfer from
`
	fmt.Println(assetHelp)
}

func transferAssetUsageError(context *cli.Context, err error, isSubcommand bool) error {
	fmt.Println(err.Error())
	showAssetTransferHelp()
	return nil
}

func showAssetTransferHelp() {
	var assetTransferHelp = `
   Name:
      ontology asset transfer              asset transfer

   Usage:
      ontology asset transfer [command options] [args]

   Description:
      With this command, you can transfer assert through transaction.

   Command:
      --caddr     value                    smart contract address
      --from      value                    wallet address base58, which will transfer from
      --to        value                    wallet address base58, which will transfer to
      --value     value                    how much asset will be transfered
      --password  value                    use password who transfer from
`
	fmt.Println(assetTransferHelp)
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
	if !ctx.IsSet(utils.ContractAddrFlag.Name) || !ctx.IsSet(utils.TransactionFromFlag.Name) || !ctx.IsSet(utils.TransactionToFlag.Name) || !ctx.IsSet(utils.TransactionValueFlag.Name) || !ctx.IsSet(utils.UserPasswordFlag.Name) {
		showAssetTransferHelp()
		return nil
	}
	contract := ctx.String(utils.ContractAddrFlag.Name)
	if contract == "" {
		fmt.Println("Invalid contract address: ", contract)
		os.Exit(1)
	}
	ct, _ := common.HexToBytes(contract)
	ctu, _ := common.AddressParseFromBytes(ct)
	from := ctx.String(utils.TransactionFromFlag.Name)
	if from == "" {
		fmt.Println("Invalid sender address: ", from)
		os.Exit(1)
	}
	f, _ := common.HexToBytes(from)
	fu, _ := common.AddressParseFromBytes(f)
	to := ctx.String(utils.TransactionToFlag.Name)
	if to == "" {
		fmt.Println("Invalid revicer address: ", to)
		os.Exit(1)
	}
	t, _ := common.HexToBytes(to)
	tu, _ := common.AddressParseFromBytes(t)
	value := ctx.Int64(utils.TransactionValueFlag.Name)
	if value <= 0 {
		fmt.Println("Invalid ont amount: ", value)
		os.Exit(1)
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
		os.Exit(1)
	}

	cont := &states.Contract{
		Address: ctu,
		Method:  "transfer",
		Args:    bf.Bytes(),
	}

	ff := new(bytes.Buffer)

	if err := cont.Serialize(ff); err != nil {
		fmt.Println("Serialize contract struct error.")
		os.Exit(1)
	}

	tx := cutils.NewInvokeTransaction(vmtypes.VmCode{
		VmType: vmtypes.Native,
		Code:   ff.Bytes(),
	})

	tx.Nonce = uint32(time.Now().Unix())

	passwd := ctx.String(utils.UserPasswordFlag.Name)

	acct := account.Open(account.WALLET_FILENAME, []byte(passwd))
	acc := acct.GetDefaultAccount()

	if err := signTransaction(acc, tx); err != nil {
		fmt.Println("signTransaction error:", err)
		os.Exit(1)
	}

	txbf := new(bytes.Buffer)
	if err := tx.Serialize(txbf); err != nil {
		fmt.Println("Serialize transaction error.")
		os.Exit(1)
	}
	resp, err := jrpc.Call(rpcAddress(), "sendrawtransaction", 0, []interface{}{hex.EncodeToString(txbf.Bytes())})

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	r := make(map[string]interface{})
	err = json.Unmarshal(resp, &r)
	if err != nil {
		fmt.Println("Unmarshal JSON failed")
		os.Exit(1)
	}
	switch r["result"].(type) {
	case map[string]interface{}:

	case string:
		fmt.Println(r["result"].(string))
		os.Exit(1)
	}

	return nil
}
