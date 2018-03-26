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

package test

import (
	"fmt"
	"os"

	. "github.com/Ontology/cli/common"
	"github.com/Ontology/http/base/rpc"
	"github.com/urfave/cli"
	"bytes"
	"encoding/hex"
	"math/big"
	"time"
	vmtypes "github.com/Ontology/vm/types"
	"github.com/Ontology/account"
	"github.com/Ontology/core/genesis"
	"github.com/Ontology/crypto"
	"github.com/Ontology/core/types"
	"github.com/Ontology/common"
	"github.com/Ontology/smartcontract/service/native/states"
	"encoding/json"
	"github.com/Ontology/core/utils"
)

func signTransaction(signer *account.Account, tx *types.Transaction) error {
	hash := tx.Hash()
	sign, _ := crypto.Sign(signer.PrivateKey, hash[:])
	tx.Sigs = append(tx.Sigs, &types.Sig{
		PubKeys: []*crypto.PubKey{signer.PublicKey},
		M: 1,
		SigData: [][]byte{sign},
	})
	return nil
}

func testAction(c *cli.Context) (err error) {
	txnNum := c.Int("num")

	transferTest(txnNum)

	return nil
}

func transferTest(n int) {
	if n <= 0 {
		n  = 1
	}
	acct := account.Open(account.WalletFileName, []byte("passwordtest"))
	acc, err := acct.GetDefaultAccount(); if err != nil {
		fmt.Println("GetDefaultAccount error:", err)
		os.Exit(1)
	}
	for i := 0; i < n; i ++ {

		tx := NewOntTransferTransaction(acc.Address, acc.Address, int64(i))
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
			return
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
		}
	}
}


func NewOntTransferTransaction(from, to common.Address, value int64) *types.Transaction {
	var sts []*states.State
	sts = append(sts, &states.State{
		From: from,
		To: to,
		Value: big.NewInt(value),
	})
	transfers := new(states.Transfers)
	transfers.States = sts

	bf := new(bytes.Buffer)

	if err := transfers.Serialize(bf); err != nil {
		fmt.Println("Serialize transfers struct error.")
		os.Exit(1)
	}

	cont := &states.Contract{
		Address: genesis.OntContractAddress,
		Method: "transfer",
		Args: bf.Bytes(),
	}

	ff := new(bytes.Buffer)
	if err := cont.Serialize(ff); err != nil {
		fmt.Println("Serialize contract struct error.")
		os.Exit(1)
	}

	tx := utils.NewInvokeTransaction(vmtypes.VmCode{
		VmType: vmtypes.Native,
		Code: ff.Bytes(),
	})

	tx.Nonce = uint32(time.Now().Unix())

	return tx
}

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:        "test",
		Usage:       "run test routine",
		Description: "With nodectl test, you could run simple tests.",
		ArgsUsage:   "[args]",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "num, n",
				Usage: "sample transaction numbers",
				Value: 1,
			},
		},
		Action: testAction,
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			PrintError(c, err, "test")
			return cli.NewExitError("", 1)
		},
	}
}
