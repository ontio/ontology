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
	"bufio"
	"fmt"
	"github.com/gosuri/uiprogress"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/urfave/cli"
	"os"
	"time"
)

var ExportCommand = cli.Command{
	Name:      "export",
	Usage:     "Export blocks in DB to a file",
	ArgsUsage: "",
	Action:    exportBlocks,
	Flags: []cli.Flag{
		utils.RPCPortFlag,
		utils.ExportFileFlag,
		utils.ExportHeightFlag,
		utils.ExportSpeedFlag,
	},
	Description: "",
}

func exportBlocks(ctx *cli.Context) error {
	SetRpcPort(ctx)
	exportFile := ctx.String(utils.GetFlagName(utils.ExportFileFlag))
	if exportFile == "" {
		fmt.Printf("Missing file argumen\n")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	if common.FileExisted(exportFile) {
		return fmt.Errorf("File:%s has already exist", exportFile)
	}
	endHeight := ctx.Uint(utils.GetFlagName(utils.ExportHeightFlag))
	blockCount, err := utils.GetBlockCount()
	if err != nil {
		return fmt.Errorf("GetBlockCount error:%s", err)
	}
	if endHeight == 0 || endHeight >= uint(blockCount) {
		endHeight = uint(blockCount) - 1
	}
	speed := ctx.String(utils.GetFlagName(utils.ExportSpeedFlag))
	var sleepTime time.Duration
	switch speed {
	case "h":
		sleepTime = 0
	case "m":
		sleepTime = time.Millisecond * 2
	default:
		sleepTime = time.Millisecond * 5
	}

	ef, err := os.OpenFile(exportFile, os.O_RDWR|os.O_CREATE, 0664)
	if err != nil {
		return fmt.Errorf("Open file:%s error:%s", exportFile, err)
	}
	defer ef.Close()
	fWriter := bufio.NewWriter(ef)

	metadata := utils.NewExportBlockMetadata()
	metadata.BlockHeight = uint32(endHeight)
	err = metadata.Serialize(fWriter)
	if err != nil {
		return fmt.Errorf("Write export metadata error:%s", err)
	}

	//progress bar
	uiprogress.Start()
	bar := uiprogress.AddBar(int(endHeight)).
		AppendCompleted().
		AppendElapsed().
		PrependFunc(func(b *uiprogress.Bar) string {
			return fmt.Sprintf("Block(%d/%d)", b.Current(), int(endHeight))
		})

	fmt.Printf("Start export.\n")
	for i := uint32(0); i <= uint32(endHeight); i++ {
		blockData, err := utils.GetBlockData(i)
		if err != nil {
			return fmt.Errorf("Get block:%d error:%s", i, err)
		}
		data, err := utils.CompressBlockData(blockData, metadata.CompressType)
		if err != nil {
			return fmt.Errorf("Compress block height:%d error:%s", i, err)
		}
		err = serialization.WriteUint32(fWriter, uint32(len(data)))
		if err != nil {
			return fmt.Errorf("write block data height:%d len:%d error:%s", i, uint32(len(data)), err)
		}
		_, err = fWriter.Write(data)
		if err != nil {
			return fmt.Errorf("write block data height:%d error:%s", i, err)
		}
		if sleepTime > 0 {
			time.Sleep(sleepTime)
		}
		bar.Incr()
	}
	uiprogress.Stop()

	err = fWriter.Flush()
	if err != nil {
		return fmt.Errorf("Export flush file error:%s", err)
	}
	fmt.Printf("Export blocks successfully.\n")
	fmt.Printf("Total blocks:%d\n", endHeight+1)
	fmt.Printf("Export file:%s\n", exportFile)
	return nil
}
