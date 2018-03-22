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

package bookkeeper

//import (
//	"bytes"
//	"encoding/hex"
//	"fmt"
//	"github.com/Ontology/account"
//	. "github.com/Ontology/cli/common"
//	"github.com/Ontology/core/contract"
//	"github.com/Ontology/core/signature"
//	"github.com/Ontology/core/transaction"
//	"github.com/Ontology/crypto"
//	"github.com/Ontology/http/httpjsonrpc"
//	"math/rand"
//	"os"
//	"strconv"
//
//	"github.com/urfave/cli"
//)
//
//func makeBookkeeperTransaction(pubkey *crypto.PubKey, op bool, cert []byte, issuer *account.Account) (string, error) {
//	tx, _ := transaction.NewBookKeeperTransaction(pubkey, op, cert, issuer.PubKey())
//	attr := transaction.NewTxAttribute(transaction.Nonce, []byte(strconv.FormatInt(rand.Int63(), 10)))
//	tx.Attributes = make([]*transaction.TxAttribute, 0)
//	tx.Attributes = append(tx.Attributes, &attr)
//	if err := signTransaction(issuer, tx); err != nil {
//		fmt.Println("Sign regist transaction failed.")
//		return "", err
//	}
//	var buffer bytes.Buffer
//	if err := tx.Serialize(&buffer); err != nil {
//		fmt.Println("Serialize bookkeeper transaction failed.")
//		return "", err
//	}
//	return hex.EncodeToString(buffer.Bytes()), nil
//}
//
//func newContractContextWithoutProgramHashes(data signature.SignableData) *contract.ContractContext {
//	return &contract.ContractContext{
//		Data:       data,
//		Codes:      make([][]byte, 1),
//		Parameters: make([][][]byte, 1),
//	}
//}
//
//func signTransaction(signer *account.Account, tx *transaction.Transaction) error {
//	signature, err := signature.SignBySigner(tx, signer)
//	if err != nil {
//		fmt.Println("SignBySigner failed.")
//		return err
//	}
//	transactionContract, err := contract.CreateSignatureContract(signer.PubKey())
//	if err != nil {
//		fmt.Println("CreateSignatureContract failed.")
//		return err
//	}
//	transactionContractContext := contract.NewContractContext(tx)
//	if err := transactionContractContext.AddContract(transactionContract, signer.PubKey(), signature); err != nil {
//		fmt.Println("AddContract failed")
//		return err
//	}
//	tx.SetPrograms(transactionContractContext.GetPrograms())
//	return nil
//}
//
//func assetAction(c *cli.Context) error {
//	if c.NumFlags() == 0 {
//		cli.ShowSubcommandHelp(c)
//		return nil
//	}
//	var pubkeyHex []byte
//	var err error
//	var add bool
//	addPubkey := c.String("add")
//	subPubkey := c.String("sub")
//	if addPubkey == "" && subPubkey == "" {
//		fmt.Println("missing --add or --sub")
//		return nil
//	}
//
//	if addPubkey != "" {
//		pubkeyHex, err = hex.DecodeString(addPubkey)
//		add = true
//	}
//	if subPubkey != "" {
//		if pubkeyHex != nil {
//			fmt.Println("using --add or --sub")
//			return nil
//		}
//		pubkeyHex, err = hex.DecodeString(subPubkey)
//		add = false
//	}
//	if err != nil {
//		fmt.Println("Invalid public key in hex")
//		return nil
//	}
//	pubkey, err := crypto.DecodePoint(pubkeyHex)
//	if err != nil {
//		fmt.Println("Invalid public key")
//		return nil
//	}
//	cert := c.String("cert")
//
//	wallet := account.Open(account.WalletFileName, WalletPassword(c.String("password")))
//	if wallet == nil {
//		fmt.Println("Failed to open wallet.")
//		os.Exit(1)
//	}
//
//	acc, _ := wallet.GetDefaultAccount()
//	txHex, err := makeBookkeeperTransaction(pubkey, add, []byte(cert), acc)
//	if err != nil {
//		return err
//	}
//
//	resp, err := jsonrpc.Call(Address(), "sendrawtransaction", 0, []interface{}{txHex})
//	if err != nil {
//		fmt.Fprintln(os.Stderr, err)
//		return err
//	}
//
//	FormatOutput(resp)
//
//	return nil
//}
//
//func NewCommand() *cli.Command {
//	return &cli.Command{
//		Name:        "bookkeeper",
//		Usage:       "add or remove bookkeeper",
//		Description: "With nodectl bookkeeper, you could add or remove bookkeeper.",
//		ArgsUsage:   "[args]",
//		Flags: []cli.Flag{
//			cli.StringFlag{
//				Name:  "add, a",
//				Usage: "add a bookkeeper",
//			},
//			cli.StringFlag{
//				Name:  "sub, s",
//				Usage: "sub a bookkeeper",
//			},
//			cli.StringFlag{
//				Name:  "cert, c",
//				Usage: "authorized certificate",
//			},
//		},
//		Action: assetAction,
//		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
//			PrintError(c, err, "bookkeeper")
//			return cli.NewExitError("", 1)
//		},
//	}
//}
