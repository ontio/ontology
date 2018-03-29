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

package transfer

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/urfave/cli"

	"github.com/Ontology/account"
	. "github.com/Ontology/cli/common"
	"github.com/Ontology/common"
	"github.com/Ontology/core/signature"
	ctypes "github.com/Ontology/core/types"
	cutils "github.com/Ontology/core/utils"
	"github.com/Ontology/http/base/rpc"
	"github.com/Ontology/smartcontract/service/native/states"
	vmtypes "github.com/Ontology/vm/types"
	"github.com/ontio/ontology-crypto/keypair"
)

func transferAction(c *cli.Context) error {
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}
	// wallet name is wallet.dat by default
	contract := c.String("contract")
	if contract == "" {
		fmt.Println("Invalid contract address: ", contract)
		os.Exit(1)
	}
	ct, _ := common.HexToBytes(contract)
	ctu, _ := common.AddressParseFromBytes(ct)
	from := c.String("from")
	if from == "" {
		fmt.Println("Invalid sender address: ", from)
		os.Exit(1)
	}
	f, _ := common.HexToBytes(from)
	fu, _ := common.AddressParseFromBytes(f)
	to := c.String("to")
	if to == "" {
		fmt.Println("Invalid revicer address: ", to)
		os.Exit(1)
	}
	t, _ := common.HexToBytes(to)
	tu, _ := common.AddressParseFromBytes(t)
	value := c.Int64("value")
	if value <= 0 {
		fmt.Println("Invalid ont amount: ", value)
		os.Exit(1)
	}

	var sts []*states.State
	sts = append(sts, &states.State{
		From:  fu,
		To:    tu,
		Value: big.NewInt(value),
	})
	transfers := &states.Transfers{
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

	passwd := c.String("password")

	acct := account.Open(account.WALLET_FILENAME, []byte(passwd))
	acc, err := acct.GetDefaultAccount()
	if err != nil {
		fmt.Println("GetDefaultAccount error:", err)
		os.Exit(1)
	}

	if err := signTransaction(acc, tx); err != nil {
		fmt.Println("signTransaction error:", err)
		os.Exit(1)
	}

	txbf := new(bytes.Buffer)
	if err := tx.Serialize(txbf); err != nil {
		fmt.Println("Serialize transaction error.")
		os.Exit(1)
	}
	resp, err := rpc.Call(RpcAddress(), "sendrawtransaction", 0,
		[]interface{}{hex.EncodeToString(txbf.Bytes())})

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

func signTransaction(signer *account.Account, tx *ctypes.Transaction) error {
	hash := tx.Hash()
	sign, _ := signature.Sign(signer.PrivateKey, hash[:])
	tx.Sigs = append(tx.Sigs, &ctypes.Sig{
		PubKeys: []keypair.PublicKey{signer.PublicKey},
		M:       1,
		SigData: [][]byte{sign},
	})
	return nil
}

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:        "transfer",
		Usage:       "user ont transfer",
		Description: "With nodectl transfer, you could transfer ont.",
		ArgsUsage:   "[args]",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "contract, c",
				Usage: "contract address",
			},
			cli.StringFlag{
				Name:  "from, f",
				Usage: "sender address",
			},
			cli.StringFlag{
				Name:  "to, t",
				Usage: "revicer address",
			},
			cli.Int64Flag{
				Name:  "value, v",
				Usage: "ont amount",
			},
			cli.StringFlag{
				Name:  "password, p",
				Usage: "wallet password",
				Value: "passwordtest",
			},
		},
		Action: transferAction,
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			return cli.NewExitError("", 1)
		},
	}
}
