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
	"io"
	"os"

	"github.com/gosuri/uiprogress"
	"github.com/urfave/cli"

	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
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

	cfg, err := SetOntologyConfig(ctx)
	if err != nil {
		PrintErrorMsg("SetOntologyConfig error:%s", err)
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	dbDir := utils.GetStoreDirPath(config.DefConfig.Common.DataDir, config.DefConfig.P2PNode.NetworkName)

	stateHashHeight := config.GetStateHashCheckHeight(cfg.P2PNode.NetworkId)
	ledger.DefLedger, err = ledger.NewLedger(dbDir, stateHashHeight)
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
		return fmt.Errorf("BuildGenesisBlock error %s", err)
	}
	err = ledger.DefLedger.Init(bookKeepers, genesisBlock)
	if err != nil {
		return fmt.Errorf("init ledger error:%s", err)
	}

	dataDir := ctx.String(utils.GetFlagName(utils.DataDirFlag))
	if dataDir == "" {
		PrintErrorMsg("Missing %s argument.", utils.DataDirFlag.Name)
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	importFile := ctx.String(utils.GetFlagName(utils.ImportFileFlag))
	if importFile == "" {
		PrintErrorMsg("Missing %s argument.", utils.ImportFileFlag.Name)
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	endBlockHeight := uint32(ctx.Uint(utils.GetFlagName(utils.ImportEndHeightFlag)))
	currBlockHeight := ledger.DefLedger.GetCurrentBlockHeight()

	if endBlockHeight > 0 && currBlockHeight >= endBlockHeight {
		PrintWarnMsg("CurrentBlockHeight:%d larger than or equal to EndBlockHeight:%d, No blocks to import.", currBlockHeight, endBlockHeight)
		return nil
	}

	ifile, err := os.OpenFile(importFile, os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("OpenFile error:%s", err)
	}
	defer ifile.Close()
	fReader := bufio.NewReader(ifile)

	metadata := utils.NewExportBlockMetadata()
	err = metadata.Deserialize(fReader)
	if err != nil {
		return fmt.Errorf("block data file metadata deserialize error:%s", err)
	}
	if metadata.EndBlockHeight <= currBlockHeight {
		PrintWarnMsg("CurrentBlockHeight:%d larger than or equal to EndBlockHeight:%d, No blocks to import.", currBlockHeight, endBlockHeight)
		return nil
	}
	if endBlockHeight == 0 || endBlockHeight > metadata.EndBlockHeight {
		endBlockHeight = metadata.EndBlockHeight
	}

	startBlockHeight := metadata.StartBlockHeight
	if startBlockHeight > (currBlockHeight + 1) {
		return fmt.Errorf("import block error: StartBlockHeight:%d larger than NextBlockHeight:%d", startBlockHeight, currBlockHeight+1)
	}
	//progress bar
	uiprogress.Start()
	bar := uiprogress.AddBar(int(endBlockHeight - startBlockHeight + 1)).
		AppendCompleted().
		AppendElapsed().
		PrependFunc(func(b *uiprogress.Bar) string {
			return fmt.Sprintf("Block(%d/%d)", b.Current()+int(currBlockHeight), int(endBlockHeight))
		})

	PrintInfoMsg("Start import blocks.")

	for i := startBlockHeight; i <= endBlockHeight; i++ {
		size, err := serialization.ReadUint32(fReader)
		if err != nil {
			return fmt.Errorf("read block size:%d error:%s", i, err)
		}
		compressData := make([]byte, size)
		_, err = io.ReadFull(fReader, compressData)
		if err != nil {
			return fmt.Errorf("read block data height:%d error:%s", i, err)
		}
		crossMsgSize, err := serialization.ReadUint32(fReader)
		if err != nil {
			return fmt.Errorf("read cross chain msg height:%d error:%s", i, err)
		}
		var crossMsgCompressData []byte
		if crossMsgSize != 0 {
			crossMsgCompressData = make([]byte, size)
			_, err = io.ReadFull(fReader, crossMsgCompressData)
			if err != nil {
				return fmt.Errorf("read block data height:%d error:%s", i, err)
			}
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
		execResult, err := ledger.DefLedger.ExecuteBlock(block)
		if err != nil {
			return fmt.Errorf("block height:%d ExecuteBlock error:%s", i, err)
		}

		var crossChainMsg *types.CrossChainMsg
		if crossMsgSize != 0 {
			crossChainMsg = new(types.CrossChainMsg)
			crossChainData, err := utils.DecompressBlockData(crossMsgCompressData, metadata.CompressType)
			if err != nil {
				return fmt.Errorf("block height:%d decompress error:%s", i, err)
			}
			if err := crossChainMsg.Deserialization(common.NewZeroCopySource(crossChainData)); err != nil {
				return fmt.Errorf("block height:%d decompress error:%s", i, err)
			}
		}
		err = ledger.DefLedger.SubmitBlock(block, crossChainMsg, execResult)
		if err != nil {
			return fmt.Errorf("SubmitBlock block height:%d error:%s", i, err)
		}
		bar.Incr()
	}
	uiprogress.Stop()
	PrintInfoMsg("Import block completed, current block height:%d.", ledger.DefLedger.GetCurrentBlockHeight())
	return nil
}
