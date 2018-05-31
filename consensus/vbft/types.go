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
	"github.com/ontio/ontology/common/serialization"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/types"
)

type Block struct {
	Block      *types.Block
	EmptyBlock *types.Block
	Info       *vconfig.VbftBlockInfo
}

func (blk *Block) getProposer() uint32 {
	return blk.Info.Proposer
}

func (blk *Block) getBlockNum() uint32 {
	return blk.Block.Header.Height
}

func (blk *Block) getPrevBlockHash() common.Uint256 {
	return blk.Block.Header.PrevBlockHash
}

func (blk *Block) getLastConfigBlockNum() uint32 {
	return blk.Info.LastConfigBlockNum
}

func (blk *Block) getNewChainConfig() *vconfig.ChainConfig {
	return blk.Info.NewChainConfig
}

//
// getVrfValue() is a helper function for participant selection.
//
func (blk *Block) getVrfValue() []byte {
	return blk.Info.VrfValue
}

func (blk *Block) getVrfProof() []byte {
	return blk.Info.VrfProof
}

func (blk *Block) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	if err := blk.Block.Serialize(buf); err != nil {
		return nil, fmt.Errorf("serialize block: %s", err)
	}

	payload := bytes.NewBuffer([]byte{})
	if err := serialization.WriteVarBytes(payload, buf.Bytes()); err != nil {
		return nil, fmt.Errorf("serialize block buf: %s", err)
	}

	if blk.EmptyBlock != nil {
		buf2 := bytes.NewBuffer([]byte{})
		if err := blk.EmptyBlock.Serialize(buf2); err != nil {
			return nil, fmt.Errorf("serialize empty block: %s", err)
		}
		if err := serialization.WriteVarBytes(payload, buf2.Bytes()); err != nil {
			return nil, fmt.Errorf("serialize empty block buf: %s", err)
		}
	}

	return payload.Bytes(), nil
}

func (blk *Block) Deserialize(data []byte) error {
	buf := bytes.NewBuffer(data)
	buf1, err := serialization.ReadVarBytes(buf)
	if err != nil {
		return fmt.Errorf("deserialize block buffer: %s", err)
	}

	block := &types.Block{}
	if err := block.Deserialize(bytes.NewBuffer(buf1)); err != nil {
		return fmt.Errorf("deserialize block: %s", err)
	}

	info := &vconfig.VbftBlockInfo{}
	if err := json.Unmarshal(block.Header.ConsensusPayload, info); err != nil {
		return fmt.Errorf("unmarshal vbft info: %s", err)
	}

	var emptyBlock *types.Block
	if buf.Len() > 0 {
		if buf2, err := serialization.ReadVarBytes(buf); err == nil {
			block2 := &types.Block{}
			if err := block2.Deserialize(bytes.NewBuffer(buf2)); err == nil {
				emptyBlock = block2
			}
		}
	}

	blk.Block = block
	blk.EmptyBlock = emptyBlock
	blk.Info = info

	return nil
}

func initVbftBlock(block *types.Block) (*Block, error) {
	if block == nil {
		return nil, fmt.Errorf("nil block in initVbftBlock")
	}

	blkInfo := &vconfig.VbftBlockInfo{}
	if err := json.Unmarshal(block.Header.ConsensusPayload, blkInfo); err != nil {
		return nil, fmt.Errorf("unmarshal blockInfo: %s", err)
	}

	return &Block{
		Block: block,
		Info:  blkInfo,
	}, nil
}
