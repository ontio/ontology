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
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"bufio"
	"encoding/binary"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	clicommon "github.com/ontio/ontology/cli/common"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/core/utils"
	"github.com/ontio/ontology/http/base/rpc"
	states "github.com/ontio/ontology/smartcontract/service/native/ont"
	sstates "github.com/ontio/ontology/smartcontract/states"
	vmtypes "github.com/ontio/ontology/smartcontract/types"
	"github.com/urfave/cli"
)

func signTransaction(signer *account.Account, tx *types.Transaction) error {
	hash := tx.Hash()
	sign, _ := signature.Sign(signer, hash[:])
	tx.Sigs = append(tx.Sigs, &types.Sig{
		PubKeys: []keypair.PublicKey{signer.PublicKey},
		M:       1,
		SigData: [][]byte{sign},
	})
	return nil
}

func testAction(c *cli.Context) (err error) {
	txnNum := c.Int("num")
	passwd := c.String("password")
	genFile := c.Bool("gen")

	acct := account.Open(account.WALLET_FILENAME, []byte(passwd))
	acc := acct.GetDefaultAccount()
	if acc == nil {
		fmt.Println(" can not get default account")
		os.Exit(1)
	}
	if genFile {
		GenTransferFile(txnNum, acc, "transfer.txt")
		return nil
	}

	transferTest(txnNum, acc)

	return nil
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func Tx2Hex(tx *types.Transaction) string {
	var buffer bytes.Buffer
	tx.Serialize(&buffer)
	return hex.EncodeToString(buffer.Bytes())
}

func GenTransferFile(n int, acc *account.Account, fileName string) {
	f, err := os.Create(fileName)
	check(err)
	w := bufio.NewWriter(f)

	defer func() {
		w.Flush()
		f.Close()
	}()

	for i := 0; i < n; i++ {
		to := acc.Address
		binary.BigEndian.PutUint64(to[:], uint64(i))
		tx := NewOntTransferTransaction(acc.Address, to, 1)
		if err := signTransaction(acc, tx); err != nil {
			fmt.Println("signTransaction error:", err)
			os.Exit(1)
		}

		txhex := Tx2Hex(tx)
		_, _ = w.WriteString(fmt.Sprintf("%x,%s\n", tx.Hash(), txhex))
	}

}

func transferTest(n int, acc *account.Account) {
	if n <= 0 {
		n = 1
	}

	for i := 0; i < n; i++ {
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
		resp, err := rpc.Call(clicommon.RpcAddress(), "sendrawtransaction", 0,
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
		From:  from,
		To:    to,
		Value: uint64(value),
	})
	transfers := new(states.Transfers)
	transfers.States = sts

	bf := new(bytes.Buffer)

	if err := transfers.Serialize(bf); err != nil {
		fmt.Println("Serialize transfers struct error.")
		os.Exit(1)
	}

	cont := &sstates.Contract{
		Address: genesis.OntContractAddress,
		Method:  "transfer",
		Args:    bf.Bytes(),
	}

	ff := new(bytes.Buffer)
	if err := cont.Serialize(ff); err != nil {
		fmt.Println("Serialize contract struct error.")
		os.Exit(1)
	}

	tx := utils.NewInvokeTransaction(vmtypes.VmCode{
		VmType: vmtypes.Native,
		Code:   ff.Bytes(),
	})

	tx.Payer = from
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
			cli.StringFlag{
				Name:  "password, p",
				Usage: "wallet password",
				Value: "passwordtest",
			},
			cli.BoolFlag{
				Name:  "gen, g",
				Usage: "gen transaction to file",
			},
		},
		Action: testAction,
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			clicommon.PrintError(c, err, "test")
			return cli.NewExitError("", 1)
		},
	}
}
