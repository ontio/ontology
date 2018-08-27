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
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
	"github.com/urfave/cli"
	"io"
	"os"
)

var ImportCommand = cli.Command{
	Name:      "import",
	Usage:     "Import blocks to DB from a file",
	ArgsUsage: "",
	Action:    importBlocks,
	Flags: []cli.Flag{
		utils.ImportFileFlag,
		utils.ImportEndHeightFlag,
		utils.DataDirFlag,
		utils.ConfigFlag,
		utils.NetworkIdFlag,
		utils.DisableEventLogFlag,
	},
	Description: "Note that import cmd doesn't support testmode",
}

func importBlocks(ctx *cli.Context) error {
	log.InitLog(log.InfoLog)

	_, err := SetOntologyConfig(ctx)
	if err != nil {
		fmt.Printf("SetOntologyConfig error:%s\n", err)
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	dbDir := utils.GetStoreDirPath(config.DefConfig.Common.DataDir, config.DefConfig.P2PNode.NetworkName)

	ledger.DefLedger, err = ledger.NewLedger(dbDir)
	if err != nil {
		return fmt.Errorf("NewLedger error:%s", err)
	}
	bookKeepers, err := config.DefConfig.GetBookkeepers()
	if err != nil {
		return fmt.Errorf("GetBookkeepers error:%s", err)
	}
	genesisConfig := config.DefConfig.Genesis
	genesisBlock, err := genesis.BuildGenesisBlock(bookKeepers, genesisConfig)
	if err != nil {
		return fmt.Errorf("genesisBlock error %s", err)
	}
	err = ledger.DefLedger.Init(bookKeepers, genesisBlock)
	if err != nil {
		return fmt.Errorf("Init ledger error:%s", err)
	}

	dataDir := ctx.String(utils.GetFlagName(utils.DataDirFlag))
	if dataDir == "" {
		fmt.Printf("missing datadir argument\n")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	importFile := ctx.String(utils.GetFlagName(utils.ImportFileFlag))
	if importFile == "" {
		fmt.Printf("missing import file argument\n")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	endBlockHeight := uint32(ctx.Uint(utils.GetFlagName(utils.ImportEndHeightFlag)))
	currBlockHeight := ledger.DefLedger.GetCurrentBlockHeight()

	if endBlockHeight > 0 && currBlockHeight >= endBlockHeight {
		fmt.Printf("CurrentBlockHeight:%d larger than or equal to EndBlockHeight:%d, No blocks to import.\n", currBlockHeight, endBlockHeight)
		return nil
	}

	ifile, err := os.OpenFile(importFile, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer ifile.Close()
	fReader := bufio.NewReader(ifile)

	metadata := utils.NewExportBlockMetadata()
	err = metadata.Deserialize(fReader)
	if err != nil {
		fmt.Printf("Metadata deserialize error:%s", err)
		return nil
	}
	if metadata.EndBlockHeight <= currBlockHeight {
		fmt.Printf("CurrentBlockHeight:%d larger than or equal to EndBlockHeight:%d, No blocks to import.\n", currBlockHeight, endBlockHeight)
		return nil
	}
	if endBlockHeight == 0 || endBlockHeight > metadata.EndBlockHeight {
		endBlockHeight = metadata.EndBlockHeight
	}

	startBlockHeight := metadata.StartBlockHeight
	if startBlockHeight > (currBlockHeight + 1) {
		fmt.Printf("Import block error: StartBlockHeight:%d larger than NextBlockHeight:%d\n", startBlockHeight, currBlockHeight+1)
		return nil
	}
	//progress bar
	uiprogress.Start()
	bar := uiprogress.AddBar(int(endBlockHeight - startBlockHeight + 1)).
		AppendCompleted().
		AppendElapsed().
		PrependFunc(func(b *uiprogress.Bar) string {
			return fmt.Sprintf("Block(%d/%d)", b.Current()+int(currBlockHeight), int(endBlockHeight))
		})

	fmt.Printf("Start import blocks\n")

	for i := uint32(startBlockHeight); i <= endBlockHeight; i++ {
		size, err := serialization.ReadUint32(fReader)
		if err != nil {
			return fmt.Errorf("Read block height:%d error:%s", i, err)
		}
		compressData := make([]byte, size)
		_, err = io.ReadFull(fReader, compressData)
		if err != nil {
			return fmt.Errorf("Read block data height:%d error:%s", i, err)
		}

		if i <= currBlockHeight {
			continue
		}

		blockData, err := utils.DecompressBlockData(compressData, metadata.CompressType)
		if err != nil {
			return fmt.Errorf("block height:%d decompress error:%s", i, err)
		}
		block, err := types.BlockFromRawBytes(blockData)
		if err != nil {
			return fmt.Errorf("block height:%d deserialize error:%s", i, err)
		}
		err = ledger.DefLedger.AddBlock(block)
		if err != nil {
			return fmt.Errorf("add block height:%d error:%s", i, err)
		}
		bar.Incr()
	}
	uiprogress.Stop()
	fmt.Printf("Import block complete, current block height:%d\n", ledger.DefLedger.GetCurrentBlockHeight())
	return nil
}
