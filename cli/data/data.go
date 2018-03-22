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

package data

//import (
//	"bytes"
//	"encoding/hex"
//	"fmt"
//	"github.com/Ontology/account"
//	. "github.com/Ontology/cli/common"
//	"github.com/Ontology/core/contract"
//	"github.com/Ontology/core/signature"
//	ctypes "github.com/Ontology/core/types"
//	"github.com/Ontology/http/base/rpc"
//	"github.com/urfave/cli"
//	"math/rand"
//	"os"
//	"strconv"
//	"github.com/Ontology/crypto"
//)
//
//func openWallet(name string, passwd []byte) account.Client {
//	if name == account.WalletFileName {
//		fmt.Println("Using default wallet: ", account.WalletFileName)
//	}
//	wallet := account.Open(name, passwd)
//	if wallet == nil {
//		fmt.Println("Failed to open wallet: ", name)
//		os.Exit(1)
//	}
//	return wallet
//}
//func newContractContextWithoutProgramHashes(data signature.SignableData) *contract.ContractContext {
//	return &contract.ContractContext{
//		Data:       data,
//		Codes:      make([][]byte, 1),
//		Parameters: make([][][]byte, 1),
//	}
//}
//func signTransaction(signer *account.Account, tx *ctypes.Transaction) error {
//	signature, err := crypto.Sign(signer.PrivKey(), tx.GetMessage())
//	if err != nil {
//		return nil, err
//	}
//	tx.Sigs = append(tx.Sigs, &ctypes.Sig{
//		PubKeys: []*crypto.PubKey{signer.PublicKey},
//		M: 1,
//		SigData: [][]byte{signature},
//	})
//	return nil
//}
//
//func readData(filepath string) ([]byte, error) {
//	if _, err := os.Stat(filepath); err != nil {
//		fmt.Printf("invalid file path:%s\n", err)
//		return nil, err
//	}
//
//	f, err := os.OpenFile(filepath, os.O_RDONLY, 0664)
//	defer f.Close()
//	if err != nil {
//		fmt.Printf("open file error:%s\n", err)
//		return nil, err
//	}
//	//read file
//	var payload []byte
//	var eof = false
//	for {
//		if eof {
//			break
//		}
//		buf := make([]byte, 1024)
//		nr, err := f.Read(buf[:])
//
//		switch true {
//		case nr < 0:
//			fmt.Fprintf(os.Stderr, "cat: error reading: %s\n", err.Error())
//			return nil, err
//		case nr == 0: // EOF
//			eof = true
//		case nr > 0:
//			payload = append(payload, buf...)
//
//		}
//	}
//	return payload, nil
//}
//
//func dataAction(c *cli.Context) error {
//	if c.NumFlags() == 0 {
//		cli.ShowSubcommandHelp(c)
//		return nil
//	}
//	upload := c.Bool("upload")
//	reg := c.Bool("reg")
//	get := c.Bool("get")
//	cat := c.Bool("cat")
//	if !upload && !reg && !get && !cat {
//		cli.ShowSubcommandHelp(c)
//		return nil
//	}
//
//	var resp []byte
//	//var txHex string
//	var err error
//	if upload {
//		filepath := c.String("file")
//
//		if filepath == "" {
//			cli.ShowSubcommandHelp(c)
//			return nil
//		}
//
//		payload, err := readData(filepath)
//		if err != nil {
//			fmt.Fprintln(os.Stderr, err)
//			return err
//		}
//
//		fmt.Println("data uploading...")
//		//tranfer data to node
//		resp, err = rpc.Call(Address(), "uploadDataFile", 0, []interface{}{payload})
//
//		if err != nil {
//			fmt.Fprintln(os.Stderr, err)
//			return err
//		}
//		fmt.Println("data uploaded")
//	}
//	if reg {
//		address := c.String("address")
//		if address == "" {
//			cli.ShowSubcommandHelp(c)
//			return nil
//		}
//		name := c.String("name")
//		if name == "" {
//			rbuf := make([]byte, 4)
//			rand.Read(rbuf)
//			name = "Ontology-" + hex.EncodeToString(rbuf)
//		}
//		//create transaction
//		var tx *transaction.Transaction
//
//		wallet := openWallet(c.String("wallet"), WalletPassword(c.String("password")))
//		admin, _ := wallet.GetDefaultAccount()
//
//		tx, _ = transaction.NewDataFileTransaction(address, name, "", admin.PubKey())
//		txAttr := transaction.NewTxAttribute(transaction.Nonce, []byte(strconv.FormatInt(rand.Int63(), 10)))
//		tx.Attributes = make([]*transaction.TxAttribute, 0)
//		tx.Attributes = append(tx.Attributes, &txAttr)
//
//		if err := signTransaction(admin, tx); err != nil {
//			fmt.Println("sign datafile transaction failed")
//			return err
//		}
//		var buffer bytes.Buffer
//		if err := tx.Serialize(&buffer); err != nil {
//			fmt.Println("serialize DataFileTransaction failed")
//			return err
//		}
//
//		txHex := hex.EncodeToString(buffer.Bytes())
//
//		resp, err = jsonrpc.Call(Address(), "regdatafile", 0, []interface{}{txHex})
//
//		if err != nil {
//			fmt.Fprintln(os.Stderr, err)
//			return err
//		}
//
//	}
//	if cat {
//		txhash := c.String("txhash")
//		if txhash == "" {
//			cli.ShowSubcommandHelp(c)
//			return nil
//		}
//		if txhash != "" {
//			resp, err = jsonrpc.Call(Address(), "catdatarecord", 0, []interface{}{txhash})
//
//		}
//
//	}
//	if get {
//		txhash := c.String("txhash")
//		if txhash == "" {
//			cli.ShowSubcommandHelp(c)
//			return nil
//		}
//		if txhash != "" {
//			resp, err = jsonrpc.Call(Address(), "getdataile", 0, []interface{}{txhash})
//			if err != nil {
//				fmt.Fprintln(os.Stderr, err)
//				return err
//			}
//		}
//
//	}
//
//	if err != nil {
//		fmt.Fprintln(os.Stderr, err)
//		return err
//	}
//	FormatOutput(resp)
//	return nil
//}
//
////NewCommand commands of ipfs and ipfs cluster
//func NewCommand() *cli.Command {
//	return &cli.Command{
//		Name:  "data",
//		Usage: "store and retrive data on chain",
//		UsageText: `
//This command can be used to manage data on chain.
//`,
//		ArgsUsage: "[args]",
//		Flags: []cli.Flag{
//
//			cli.BoolFlag{
//				Name:  "upload,u",
//				Usage: "upload data",
//			},
//			cli.BoolFlag{
//				Name:  "reg,r",
//				Usage: "reg data",
//			},
//			cli.BoolFlag{
//				Name:  "cat,c",
//				Usage: "cat data information",
//			},
//			cli.BoolFlag{
//				Name:  "get,g",
//				Usage: "download file",
//			},
//			cli.StringFlag{
//				Name:  "file,f",
//				Usage: "upload file name",
//			},
//			cli.StringFlag{
//				Name:  "name,n",
//				Usage: "reg name",
//			},
//			cli.StringFlag{
//				Name:  "wallet, w",
//				Value: account.WalletFileName,
//				Usage: "Wallet Name",
//			},
//			cli.StringFlag{
//				Name:  "password, p",
//				Usage: "password",
//			},
//			cli.StringFlag{
//				Name:  "txhash,t",
//				Usage: "transaction hash",
//			},
//			cli.StringFlag{
//				Name:  "address,a",
//				Usage: "ipfs data address",
//			},
//
//			// cli.StringFlag{
//			// 	Name:  "signature,s",
//			// 	Usage: "signature file of data file",
//			// },
//		},
//		Action: dataAction,
//		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
//			PrintError(c, err, "data")
//			return cli.NewExitError("", 1)
//		},
//	}
//}
