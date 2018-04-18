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
	"os"

	sdkutils "github.com/ontio/ontology-go-sdk/utils"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/urfave/cli"
	"errors"
)

var (
	InfoCommand = cli.Command{
		Action:   utils.MigrateFlags(infoCommand),
		Name:     "info",
		Usage:    "ontology info [block|chain|transaction|version] [OPTION]",
		Flags:    append(NodeFlags, InfoFlags...),
		Category: "INFO COMMANDS",
		Subcommands: []cli.Command{
			blockCommandSet,
			txCommandSet,
			versionCommand,
		},
		Description: ``,
	}
)

func showInfoHelp() {
	var infoHelp = `
   Name:
      ontology info                    Show blockchain information

   Usage:
      ontology info [command options] [args]

   Description:
      With ontology info, you can look up blocks, transactions, etc.

   Command:
      version

      block
         --hash value                  block hash value
         --height value                block height value

      tx
         --hash value                  transaction hash value

`
	fmt.Println(infoHelp)
}

func infoCommand(context *cli.Context) error {
	showInfoHelp()
	return nil
}

func blockInfoUsageError(context *cli.Context, err error, isSubcommand bool) error {
	fmt.Println("Error:", err.Error())
	showBlockInfoHelp()
	return nil
}

func getCurrentBlockHeight(ctx *cli.Context) error {
	height, err := ontSdk.Rpc.GetBlockCount()
	if nil != err {
		log.Fatalf("Get block height information is error:  %s", err.Error())
		return err
	}
	fmt.Println("Current blockchain height: ", height)
	return nil
}

var blockCommandSet = cli.Command{
	Action:       utils.MigrateFlags(blockInfoCommand),
	Name:         "block",
	Usage:        "./ontology info block [OPTION]",
	Flags:        append(NodeFlags, InfoFlags...),
	OnUsageError: blockInfoUsageError,
	Category:     "INFO COMMANDS",
	Description:  ``,
	Subcommands: []cli.Command{
		{
			Action:      utils.MigrateFlags(getCurrentBlockHeight),
			Name:        "count",
			Usage:       "issue asset by command",
			Category:    "INFO COMMANDS",
			Description: ``,
		},
	},
}

func txInfoUsageError(context *cli.Context, err error, isSubcommand bool) error {
	fmt.Println("Error:", err.Error())
	showTxInfoHelp()
	return nil
}

var txCommandSet = cli.Command{
	Action:       utils.MigrateFlags(txInfoCommand),
	Name:         "tx",
	Usage:        "ontology info tx [OPTION]\n",
	Flags:        append(NodeFlags, InfoFlags...),
	OnUsageError: txInfoUsageError,
	Category:     "INFO COMMANDS",
	Description:  ``,
}

func versionInfoUsageError(context *cli.Context, err error, isSubcommand bool) error {
	fmt.Println("Error:", err.Error())
	showVersionInfoHelp()
	return nil
}

var versionCommand = cli.Command{
	Action:       utils.MigrateFlags(versionInfoCommand),
	Name:         "version",
	Usage:        "ontology info version\n",
	OnUsageError: versionInfoUsageError,
	Category:     "INFO COMMANDS",
	Description:  ``,
}

func showVersionInfoHelp() {
	var versionInfoHelp = `
   Name:
      ontology info version            Show ontology node version

   Usage:
      ontology info version

   Description:
      With this command, you can look up the ontology node version.

`
	fmt.Println(versionInfoHelp)
}

func versionInfoCommand(ctx *cli.Context) error {
	version, err := ontSdk.Rpc.GetVersion()
	if nil != err {
		log.Fatalf("Get version information is error:  %s", err.Error())
		return err
	}
	fmt.Println("Node version: ", version)
	return nil
}

func showBlockInfoHelp() {
	var blockInfoHelp = `
   Name:
      ontology info block             Show blockchain information

   Usage:
      ontology info block [command options] [args]

   Description:
      With this command, you can look up block information.

   Options:
      --hash value                    block hash value
      --height value                  block height value
`
	fmt.Println(blockInfoHelp)
}

func blockInfoCommand(ctx *cli.Context) error {
	if ctx.IsSet(utils.HeightInfoFlag.Name) {
		height := ctx.Int(utils.HeightInfoFlag.Name)
		fmt.Println("blockInfo height: ", height)
		if height >= 0 {
			block, err := ontSdk.Rpc.GetBlockByHeight(uint32(height))

			if err != nil {
				log.Fatalf("Get block by height(%d) is error:%s", height, err.Error())
				return err
			}
			if block == nil || block.Header == nil {
				log.Fatalf("Get block by height(%d), the block or block.Header is nil", height)
				return errors.New("GetBlockByHeight: the block or block.Header is nil ")
			}

			echoBlockGracefully(block)
			return nil
		}
	} else if ctx.IsSet(utils.HashInfoFlag.Name) {
		blockHash := ctx.String(utils.HashInfoFlag.Name)
		fmt.Println("blockInfo blockHash: ", blockHash)
		if "" != blockHash {
			var hash common.Uint256
			hex, err := hex.DecodeString(blockHash)
			if err != nil {
				log.Fatalf("Decode string error, blockHash:%s, err:%s", blockHash, err.Error())
				return err
			}
			if err := hash.Deserialize(bytes.NewReader(hex)); err != nil {
				log.Fatalf("Deserialize hex error,hex:%s, err:%s", hex, err.Error())
				return err
			}
			block, err := ontSdk.Rpc.GetBlockByHash(hash)
			if err != nil {
				log.Fatalf("GetBlock GetBlockFromStore BlockHash:%x error:%s", hash, err)
				return err
			}
			if block == nil || block.Header == nil {
				return errors.New("GetBlockByHash: the block or block.Header is nil ")
			}

			echoBlockGracefully(block)
			return nil
		}
	}
	showBlockInfoHelp()
	return nil
}

func showTxInfoHelp() {
	var txInfoHelp = `
   Name:
      ontology info tx               Show transaction information

   Usage:
      ontology info tx [command options] [args]

   Description:
      With this command, you can look up transaction information.

   Options:
      --hash value                   transaction hash value

`
	fmt.Println(txInfoHelp)
}

func txInfoCommand(ctx *cli.Context) error {
	if ctx.IsSet(utils.HashInfoFlag.Name) {
		txHash := ctx.String(utils.HashInfoFlag.Name)
		ontInitTx, err := sdkutils.ParseUint256FromHexString(txHash)
		if err != nil {
			log.Errorf("ParseUint256FromHexString error:%s", err)
			return err
		}

		tx, err := ontSdk.Rpc.GetRawTransaction(ontInitTx)
		if err != nil {
			log.Errorf("GetRawTransaction error:%s", err)
			return err
		}

		echoBlockGracefully(tx)
		return nil
	}
	showTxInfoHelp()
	return nil
}

func echoBlockGracefully(block interface{}) {
	jsons, errs := json.Marshal(block)
	if errs != nil {
		log.Fatalf("Marshal json err:%s", errs.Error())
	}

	var out bytes.Buffer
	err := json.Indent(&out, jsons, "", "\t")
	if err != nil {
		log.Fatalf("Gracefully format json err: %s", err.Error())
	}
	out.WriteTo(os.Stdout)
}
