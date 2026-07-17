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
	"fmt"

	"github.com/ontio/ontology/common"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/types"
)

type VbftBlock struct {
	Block      *types.Block
	EmptyBlock *types.Block
	Info       *vconfig.VbftBlockInfo
}

func (blk *VbftBlock) getProposer() uint32 {
	return blk.Info.Proposer
}

func (blk *VbftBlock) getBlockNum() uint32 {
	return blk.Block.Header.Height
}

func (blk *VbftBlock) getPrevBlockHash() common.Uint256 {
	return blk.Block.Header.PrevBlockHash
}

func initVbftBlock(block *types.Block) (*VbftBlock, error) {
	if block == nil {
		return nil, fmt.Errorf("nil block in initVbftBlock")
	}

	blkInfo, err := vconfig.VbftBlock(block.Header)
	if err != nil {
		return nil, fmt.Errorf("unmarshal blockInfo: %s", err)
	}

	return &VbftBlock{
		Block: block,
		Info:  blkInfo,
	}, nil
}
