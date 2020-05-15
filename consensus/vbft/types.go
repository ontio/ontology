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
	"encoding/json"
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/types"
)

type Block struct {
	Block              *types.Block
	EmptyBlock         *types.Block
	Info               *vconfig.VbftBlockInfo
	PrevExecMerkleRoot common.Uint256
	CrossChainMsg      *types.CrossChainMsg
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

func (blk *Block) getPrevExecMerkleRoot() common.Uint256 {
	return blk.PrevExecMerkleRoot
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

func (blk *Block) Serialize() []byte {
	payload := common.NewZeroCopySink(nil)
	payload.WriteVarBytes(common.SerializeToBytes(blk.Block))

	payload.WriteBool(blk.EmptyBlock != nil)
	if blk.EmptyBlock != nil {
		payload.WriteVarBytes(common.SerializeToBytes(blk.EmptyBlock))
	}
	payload.WriteHash(blk.PrevExecMerkleRoot)
	payload.WriteBool(blk.CrossChainMsg != nil)
	if blk.CrossChainMsg != nil {
		blk.CrossChainMsg.Serialization(payload)
	}
	return payload.Bytes()
}

func (blk *Block) Deserialize(data []byte) error {
	source := common.NewZeroCopySource(data)
	//buf := bytes.NewBuffer(data)
	buf1, _, irregular, eof := source.NextVarBytes()
	if irregular {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}

	block, err := types.BlockFromRawBytes(buf1)
	if err != nil {
		return fmt.Errorf("deserialize block: %s", err)
	}

	info := &vconfig.VbftBlockInfo{}
	if err := json.Unmarshal(block.Header.ConsensusPayload, info); err != nil {
		return fmt.Errorf("unmarshal vbft info: %s", err)
	}

	var emptyBlock *types.Block
	hasEmptyBlock, irr, eof := source.NextBool()
	if irr {
		return fmt.Errorf("read empty-block-bool: %s", common.ErrIrregularData)
	}
	if eof {
		return fmt.Errorf("read empty-block-bool: %s", io.ErrUnexpectedEOF)
	}
	if hasEmptyBlock {
		buf2, _, irregular, eof := source.NextVarBytes()
		if irregular || eof {
			return fmt.Errorf("read empty block failed: %v, %v", irregular, eof)
		}
		block2, err := types.BlockFromRawBytes(buf2)
		if err != nil {
			return fmt.Errorf("deserialize empty blk failed: %s", err)
		}
		emptyBlock = block2
	}

	var merkleRoot common.Uint256
	merkleRoot, eof = source.NextHash()
	if eof {
		return fmt.Errorf("block deserialize merkleRoot: %s", io.ErrUnexpectedEOF)
	}

	var crossChainMsg *types.CrossChainMsg
	// ignore eof for backward compatibility
	hasEmptyCCM, irr, _ := source.NextBool()
	if irr {
		return fmt.Errorf("read empty-crosschainmsg-bool.")
	}
	if hasEmptyCCM {
		crossChainMsg = new(types.CrossChainMsg)
		if err := crossChainMsg.Deserialization(source); err != nil {
			return err
		}
	}
	blk.Block = block
	blk.EmptyBlock = emptyBlock
	blk.Info = info
	blk.PrevExecMerkleRoot = merkleRoot
	blk.CrossChainMsg = crossChainMsg
	return nil
}

func initVbftBlock(block *types.Block, ccMsg *types.CrossChainMsg, prevExecMerkleRoot common.Uint256) (*Block, error) {
	if block == nil {
		return nil, fmt.Errorf("nil block in initVbftBlock")
	}

	blkInfo := &vconfig.VbftBlockInfo{}
	if err := json.Unmarshal(block.Header.ConsensusPayload, blkInfo); err != nil {
		return nil, fmt.Errorf("unmarshal blockInfo: %s", err)
	}

	return &Block{
		Block:              block,
		Info:               blkInfo,
		PrevExecMerkleRoot: prevExecMerkleRoot,
		CrossChainMsg:      ccMsg,
	}, nil
}
