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
	"github.com/ontio/ontology/common/config"
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
		Name:        "asset",
		Usage:       "controll wallet,just as create delete action .etc",
		ArgsUsage:   "",
		Category:    "ASSET COMMANDS",
		Description: `[reg/issue/transfer]`,
		Subcommands: []cli.Command{
			{
				Action:      utils.MigrateFlags(regAsset),
				Name:        "reg",
				Usage:       "register asset",
				Flags:       append(append(NodeFlags, RpcFlags...), WhisperFlags...),
				Category:    "ASSET COMMANDS",
				Description: ``,
			},
			{
				Action:      utils.MigrateFlags(issueAsset),
				Name:        "issue",
				Usage:       "issue asset by command",
				Flags:       append(append(NodeFlags, RpcFlags...), WhisperFlags...),
				Category:    "ASSET COMMANDS",
				Description: ``,
			},
			{
				Action:      utils.MigrateFlags(transferAsset),
				Name:        "transfer",
				Usage:       "transfer asset",
				Flags:       append(append(NodeFlags, RpcFlags...), WhisperFlags...),
				Category:    "ASSET COMMANDS",
				Description: ``,
			},
		},
	}
)

func regAsset(ctx *cli.Context) error {
	//TODO
	return nil
}

func issueAsset(ctx *cli.Context) error {
	//TODO
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
	config.Init(ctx)
	contract := ctx.GlobalString(utils.ContractAddrFlag.Name)
	if contract == "" {
		fmt.Println("Invalid contract address: ", contract)
		os.Exit(1)
	}
	ct, _ := common.HexToBytes(contract)
	ctu, _ := common.AddressParseFromBytes(ct)
	from := ctx.GlobalString(utils.TransactionFromFlag.Name)
	if from == "" {
		fmt.Println("Invalid sender address: ", from)
		os.Exit(1)
	}
	f, _ := common.HexToBytes(from)
	fu, _ := common.AddressParseFromBytes(f)
	to := ctx.GlobalString(utils.TransactionToFlag.Name)
	if to == "" {
		fmt.Println("Invalid revicer address: ", to)
		os.Exit(1)
	}
	t, _ := common.HexToBytes(to)
	tu, _ := common.AddressParseFromBytes(t)
	value := ctx.GlobalInt64(utils.TransactionValueFlag.Name)
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

	passwd := ctx.GlobalString(utils.UserPasswordFlag.Name)

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
	resp, err := jrpc.Call(localRpcAddress(), "sendrawtransaction", 0, []interface{}{hex.EncodeToString(txbf.Bytes())})

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
