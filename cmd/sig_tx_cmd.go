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
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ontio/ontology-crypto/keypair"
	cmdcom "github.com/ontio/ontology/cmd/common"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/core/types"
	"github.com/urfave/cli"
)

var MultiSigAddrCommand = cli.Command{
	Name:        "multisigaddr",
	Usage:       "Generate multi-signature address",
	Description: "Generate multi-signature address.",
	Action:      genMultiSigAddress,
	Flags: []cli.Flag{
		utils.AccountMultiMFlag,
		utils.AccountMultiPubKeyFlag,
	},
}

var MultiSigTxCommand = cli.Command{
	Name:        "multisigtx",
	Usage:       "Sign to multi-signature transaction",
	ArgsUsage:   "<rawtx>",
	Description: "Sign to multi-signature transaction.",
	Action:      multiSigToTx,
	Flags: []cli.Flag{
		utils.RPCPortFlag,
		utils.WalletFileFlag,
		utils.AccountMultiMFlag,
		utils.AccountMultiPubKeyFlag,
		utils.AccountAddressFlag,
		utils.SendTxFlag,
		utils.PrepareExecTransactionFlag,
	},
}

var SigTxCommand = cli.Command{
	Name:        "sigtx",
	Usage:       "Sign to transaction",
	ArgsUsage:   "<rawtx>",
	Description: "Sign to transaction.",
	Action:      sigToTx,
	Flags: []cli.Flag{
		utils.RPCPortFlag,
		utils.WalletFileFlag,
		utils.AccountAddressFlag,
		utils.SendTxFlag,
		utils.PrepareExecTransactionFlag,
	},
}

func genMultiSigAddress(ctx *cli.Context) error {
	pkstr := strings.TrimSpace(strings.Trim(ctx.String(utils.GetFlagName(utils.AccountMultiPubKeyFlag)), ","))
	m := ctx.Uint(utils.GetFlagName(utils.AccountMultiMFlag))
	if pkstr == "" || m == 0 {
		PrintErrorMsg("Missing argument. %s or %s expected.",
			utils.GetFlagName(utils.AccountMultiMFlag),
			utils.GetFlagName(utils.AccountMultiPubKeyFlag))
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	pks := strings.Split(pkstr, ",")
	pubKeys := make([]keypair.PublicKey, 0, len(pks))
	for _, pk := range pks {
		pk := strings.TrimSpace(pk)
		if pk == "" {
			continue
		}
		data, err := hex.DecodeString(pk)
		pubKey, err := keypair.DeserializePublicKey(data)
		if err != nil {
			return fmt.Errorf("invalid pub key:%s", pk)
		}
		pubKeys = append(pubKeys, pubKey)
	}
	pkSize := len(pubKeys)
	if !(1 <= m && int(m) <= pkSize && pkSize > 1 && pkSize <= constants.MULTI_SIG_MAX_PUBKEY_SIZE) {
		PrintErrorMsg("Invalid argument. %s must > 1 and <= %d, and m must > 0 and < number of pub key.",
			utils.GetFlagName(utils.AccountMultiPubKeyFlag),
			constants.MULTI_SIG_MAX_PUBKEY_SIZE)
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	addr, err := types.AddressFromMultiPubKeys(pubKeys, int(m))
	if err != nil {
		return err
	}

	PrintInfoMsg("Pub key list:")
	for i, pubKey := range pubKeys {
		addr := types.AddressFromPubKey(pubKey)
		PrintInfoMsg("Index %d Address:%s PubKey:%x ", i+1, addr.ToBase58(), keypair.SerializePublicKey(pubKey))
	}
	PrintInfoMsg("\nMultiSigAddress:%s", addr.ToBase58())
	return nil
}

func multiSigToTx(ctx *cli.Context) error {
	SetRpcPort(ctx)
	pkstr := strings.TrimSpace(strings.Trim(ctx.String(utils.GetFlagName(utils.AccountMultiPubKeyFlag)), ","))
	m := ctx.Uint(utils.GetFlagName(utils.AccountMultiMFlag))
	if pkstr == "" || m == 0 {
		PrintErrorMsg("Missing argument. %s or %s expected.",
			utils.GetFlagName(utils.AccountMultiMFlag),
			utils.GetFlagName(utils.AccountMultiPubKeyFlag))
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	pks := strings.Split(pkstr, ",")
	pubKeys := make([]keypair.PublicKey, 0, len(pks))
	for _, pk := range pks {
		pk := strings.TrimSpace(pk)
		if pk == "" {
			continue
		}
		data, err := hex.DecodeString(pk)
		pubKey, err := keypair.DeserializePublicKey(data)
		if err != nil {
			return fmt.Errorf("invalid pub key:%s", pk)
		}
		pubKeys = append(pubKeys, pubKey)
	}
	pkSize := len(pubKeys)
	if !(1 <= m && int(m) <= pkSize && pkSize > 1 && pkSize <= constants.MULTI_SIG_MAX_PUBKEY_SIZE) {
		PrintErrorMsg("Invalid argument. %s must > 1 and <= %d, and m must > 0 and < number of pub key.",
			utils.GetFlagName(utils.AccountMultiPubKeyFlag),
			constants.MULTI_SIG_MAX_PUBKEY_SIZE)
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	if ctx.NArg() < 1 {
		PrintErrorMsg("Missing <rawtx> argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	rawTx := ctx.Args().First()
	txData, err := hex.DecodeString(rawTx)
	if err != nil {
		return fmt.Errorf("RawTx hex decode error:%s", err)
	}
	tx, err := types.TransactionFromRawBytes(txData)
	if err != nil {
		return fmt.Errorf("TransactionFromRawBytes error:%s", err)
	}

	mutTx, err := tx.IntoMutable()
	if err != nil {
		return fmt.Errorf("IntoMutable error:%s", err)
	}

	acc, err := cmdcom.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("GetAccount error:%s", err)
	}
	err = utils.MultiSigTransaction(mutTx, uint16(m), pubKeys, acc)
	if err != nil {
		return fmt.Errorf("MultiSigTransaction error:%s", err)
	}

	tx, err = mutTx.IntoImmutable()
	if err != nil {
		return fmt.Errorf("IntoImmutable error:%s", err)
	}
	sink := common.ZeroCopySink{}
	tx.Serialization(&sink)

	rawTx = hex.EncodeToString(sink.Bytes())
	PrintInfoMsg("RawTx after multi signed:")
	PrintInfoMsg(rawTx)
	PrintInfoMsg("")

	if ctx.IsSet(utils.GetFlagName(utils.PrepareExecTransactionFlag)) {
		preResult, err := utils.PrepareSendRawTransaction(rawTx)
		if err != nil {
			return err
		}
		if preResult.State == 0 {
			return fmt.Errorf("prepare execute transaction failed. %v", preResult)
		}
		PrintInfoMsg("Prepare execute transaction success.")
		PrintInfoMsg("Gas limit:%d", preResult.Gas)
		PrintInfoMsg("Result:%v", preResult.Result)
		return nil
	}

	if ctx.IsSet(utils.GetFlagName(utils.SendTxFlag)) {
		txHash, err := utils.SendRawTransactionData(rawTx)
		if err != nil {
			return err
		}
		PrintInfoMsg("Send transaction success.")
		PrintInfoMsg("  TxHash:%s", txHash)
		PrintInfoMsg("\nTip:")
		PrintInfoMsg("  Using './ontology info status %s' to query transaction status.", txHash)
	}
	return nil
}

func sigToTx(ctx *cli.Context) error {
	SetRpcPort(ctx)
	if ctx.NArg() < 1 {
		PrintErrorMsg("Missing <rawtx> argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	rawTx := ctx.Args().First()
	txData, err := hex.DecodeString(rawTx)
	if err != nil {
		return fmt.Errorf("RawTx hex decode error:%s", err)
	}
	tx, err := types.TransactionFromRawBytes(txData)
	if err != nil {
		return fmt.Errorf("TransactionFromRawBytes error:%s", err)
	}

	mutTx, err := tx.IntoMutable()
	if err != nil {
		return fmt.Errorf("IntoMutable error:%s", err)
	}

	acc, err := cmdcom.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("GetAccount error:%s", err)
	}

	err = utils.SignTransaction(acc, mutTx)
	if err != nil {
		return fmt.Errorf("SignTransaction error:%s", err)
	}

	tx, err = mutTx.IntoImmutable()
	if err != nil {
		return fmt.Errorf("IntoImmutable error:%s", err)
	}
	sink := common.ZeroCopySink{}
	tx.Serialization(&sink)

	rawTx = hex.EncodeToString(sink.Bytes())
	PrintInfoMsg("RawTx after signed:")
	PrintInfoMsg(rawTx)
	PrintInfoMsg("")

	if ctx.IsSet(utils.GetFlagName(utils.PrepareExecTransactionFlag)) {
		preResult, err := utils.PrepareSendRawTransaction(rawTx)
		if err != nil {
			return err
		}
		if preResult.State == 0 {
			return fmt.Errorf("prepare execute transaction failed. %v", preResult)
		}
		PrintInfoMsg("Prepare execute transaction success.")
		PrintInfoMsg("Gas limit:%d", preResult.Gas)
		PrintInfoMsg("Result:%v", preResult.Result)
		return nil
	}

	if ctx.IsSet(utils.GetFlagName(utils.SendTxFlag)) {
		txHash, err := utils.SendRawTransactionData(rawTx)
		if err != nil {
			return err
		}
		PrintInfoMsg("Send transaction success.")
		PrintInfoMsg("  TxHash:%s", txHash)
		PrintInfoMsg("\nTip:")
		PrintInfoMsg("  Using './ontology info status %s' to query transaction status.", txHash)
	}
	return nil
}
