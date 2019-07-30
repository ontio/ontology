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
	"strconv"

	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/core/types"
	httpcom "github.com/ontio/ontology/http/base/common"
	"github.com/urfave/cli"
)

var InfoCommand = cli.Command{
	Name:  "info",
	Usage: "Display information about the chain",
	Subcommands: []cli.Command{
		{
			Action:    txInfo,
			Name:      "tx",
			Usage:     "Display transaction information",
			ArgsUsage: "txHash",
			Flags: []cli.Flag{
				utils.RPCPortFlag,
			},
			Description: "Display transaction information",
		},
		{
			Action:    blockInfo,
			Name:      "block",
			Usage:     "Display block information",
			ArgsUsage: "<blochHash|height>",
			Flags: []cli.Flag{
				utils.RPCPortFlag,
			},
			Description: "Display block information",
		},
		{
			Action:      txStatus,
			Name:        "status",
			Usage:       "Display transaction status",
			ArgsUsage:   "<txhash>",
			Description: `Display status of transaction.`,
			Flags: []cli.Flag{
				utils.RPCPortFlag,
			},
		},
		{
			Action:      shardTxStatus,
			Name:        "shardtxstatus",
			Usage:       "Display shardtransaction status through sourceTxHash",
			ArgsUsage:   "<sourcetxhash>",
			Description: `Display shard transaction status with source transaction hash.`,
			Flags: []cli.Flag{
				utils.RPCPortFlag,
			},
		},
		{
			Action:      txState,
			Name:        "txstate",
			Usage:       "Display transaction state",
			ArgsUsage:   "<txhash, [notifyid]>",
			Description: `Display tx state if tx invoke remote shard, or tx is shard tx.`,
			Flags: []cli.Flag{
				utils.RPCPortFlag,
			},
		},
		{
			Action:      curBlockHeight,
			Name:        "curblockheight",
			Usage:       "Display the current block height",
			ArgsUsage:   "",
			Description: `Display the current block height.`,
			Flags: []cli.Flag{
				utils.RPCPortFlag,
			},
		},
		{
			Action:      getShardChainConfig,
			Name:        "shardChainConfig",
			Usage:       "Display chain config by shardID,block height",
			ArgsUsage:   "<shardID,height>",
			Description: `Display shard chain config by shardID,block height`,
			Flags: []cli.Flag{
				utils.RPCPortFlag,
			},
		},
	},
	Description: `Query information command can query information such as blocks, transactions, and transaction executions. 
You can use the ./Ontology info block --help command to view help information.`,
}

var ShowTxCommand = cli.Command{
	Action:    showTx,
	Name:      "showtx",
	Usage:     "Show info of raw transaction.",
	ArgsUsage: "<rawtx>",
	Flags: []cli.Flag{
		utils.RPCPortFlag,
	},
	Description: "Show info of raw transaction.",
}

func blockInfo(ctx *cli.Context) error {
	SetRpcPort(ctx)
	if ctx.NArg() < 1 {
		PrintErrorMsg("Missing argument,BlockHash or height expected.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	var data []byte
	var err error
	var height int64
	arg := ctx.Args().First()
	if len(arg) > 30 {
		data, err = utils.GetBlock(arg)
		if err != nil {
			return fmt.Errorf("GetBlock error:%s", err)
		}
	} else {
		height, err = strconv.ParseInt(arg, 10, 64)
		if err != nil {
			return fmt.Errorf("arg:%s invalid block hash or block height", arg)
		}
		data, err = utils.GetBlock(height)
		if err != nil {
			return fmt.Errorf("GetBlock error:%s", err)
		}
	}
	PrintJsonData(data)
	return nil
}

func txInfo(ctx *cli.Context) error {
	SetRpcPort(ctx)
	if ctx.NArg() < 1 {
		PrintErrorMsg("Missing argument. TxHash expected.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	txInfo, err := utils.GetRawTransaction(ctx.Args().First())
	if err != nil {
		return fmt.Errorf("GetRawTransaction error:%s", err)
	}
	PrintJsonData(txInfo)
	return nil
}

func txStatus(ctx *cli.Context) error {
	SetRpcPort(ctx)
	if ctx.NArg() < 1 {
		PrintErrorMsg("Missing argument. TxHash expected.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	txHash := ctx.Args().First()
	evtInfos, err := utils.GetSmartContractEventInfo(txHash)
	if err != nil {
		return fmt.Errorf("GetSmartContractEvent error:%s", err)
	}
	if string(evtInfos) == "null" {
		PrintInfoMsg("Cannot get SmartContractEvent by TxHash:%s.", txHash)
		return nil
	}
	PrintInfoMsg("Transaction states:")
	PrintJsonData(evtInfos)
	return nil
}
func shardTxStatus(ctx *cli.Context) error {
	SetRpcPort(ctx)
	if ctx.NArg() < 1 {
		PrintErrorMsg("Missing argument. SourceTxHash expected.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	txHash := ctx.Args().First()
	evtInfos, err := utils.GetShardSmartContractEventInfo(txHash)
	if err != nil {
		return fmt.Errorf("GetSmartContractEvent error:%s", err)
	}
	if string(evtInfos) == "null" {
		PrintInfoMsg("Cannot get SmartContractEvent by TxHash:%s.", txHash)
		return nil
	}
	PrintInfoMsg("Transaction states:")
	PrintJsonData(evtInfos)
	return nil
}

func txState(ctx *cli.Context) error {
	SetRpcPort(ctx)
	if ctx.NArg() < 1 {
		PrintErrorMsg("Missing argument. TxHash expected.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	txHash := ctx.Args().First()
	var notifyId string
	var hasNotifyId bool
	if ctx.NArg() > 1 {
		args := ctx.Args().Tail()
		notifyId = args[0]
		hasNotifyId = true
	}
	txState, err := utils.GetShardTxState(txHash, notifyId, hasNotifyId)
	if err != nil {
		return fmt.Errorf("GetShardTxState error:%s", err)
	}
	if string(txState) == "null" {
		PrintInfoMsg("Cannot get shard transaction state by TxHash:%s.", txHash)
		return nil
	}
	PrintInfoMsg("Transaction states:")
	PrintJsonData(txState)
	return nil
}

func curBlockHeight(ctx *cli.Context) error {
	SetRpcPort(ctx)
	count, err := utils.GetBlockCount()
	if err != nil {
		return err
	}
	PrintInfoMsg("CurrentBlockHeight:%d", count-1)
	return nil
}

func showTx(ctx *cli.Context) error {
	SetRpcPort(ctx)
	if ctx.NArg() < 1 {
		PrintErrorMsg("Missing raw tx argument.")
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
	txInfo := httpcom.TransArryByteToHexString(tx)

	txHash := tx.Hash()
	height, _ := utils.GetTxHeight(txHash.ToHexString())
	txInfo.Height = height
	PrintJsonObject(txInfo)
	return nil
}

func getShardChainConfig(ctx *cli.Context) error {
	SetRpcPort(ctx)
	if ctx.NArg() < 2 {
		PrintErrorMsg("Missing argument. shardID,height expected.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	shardID, err := strconv.ParseUint(ctx.Args().First(), 10, 64)
	if err != nil {
		return fmt.Errorf("ParseUint shardID error:%s", err)
	}
	height, err := strconv.ParseUint(ctx.Args().Get(1), 10, 32)
	if err != nil {
		return fmt.Errorf("ParseUint height error:%s", err)
	}
	chainConfig, err := utils.GetShardChainConfig(shardID, height)
	if err != nil {
		return fmt.Errorf("GetShardChainConfig err:%s", err)
	}
	PrintJsonObject(chainConfig)
	return nil
}
