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

package vbft

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ontio/ontology/common"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/types"
)

type Block struct {
	Block *types.Block           `json:"block"`
	Info  *vconfig.VbftBlockInfo `json:"info"`
}

func (blk *Block) getProposer() uint32 {
	return blk.Info.Proposer
}

func (blk *Block) getBlockNum() uint64 {
	return uint64(blk.Block.Header.Height)
}

func (blk *Block) getPrevBlockHash() common.Uint256 {
	return blk.Block.Header.PrevBlockHash
}

func (blk *Block) getLastConfigBlockNum() uint64 {
	return blk.Info.LastConfigBlockNum
}

func (blk *Block) getNewChainConfig() *vconfig.ChainConfig {
	return blk.Info.NewChainConfig
}

func (blk *Block) isEmpty() bool {
	return blk.Block.Transactions == nil || len(blk.Block.Transactions) == 0
}

func (blk *Block) Serialize() ([]byte, error) {
	infoData, err := json.Marshal(blk.Info)
	if err != nil {
		return nil, fmt.Errorf("marshal blockInfo: %s", err)
	}

	blk.Block.Header.ConsensusPayload = infoData

	buf := bytes.NewBuffer([]byte{})
	if err := blk.Block.Serialize(buf); err != nil {
		return nil, fmt.Errorf("serialize block type: %s", err)
	}

	return buf.Bytes(), nil
}

func initVbftBlock(block *types.Block) (*Block, error) {
	blkInfo := &vconfig.VbftBlockInfo{}
	if err := json.Unmarshal(block.Header.ConsensusPayload, blkInfo); err != nil {
		return nil, fmt.Errorf("unmarshal blockInfo: %s", err)
	}

	return &Block{
		Block: block,
		Info:  blkInfo,
	}, nil
}
