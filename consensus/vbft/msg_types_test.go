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
	"testing"
	"time"

	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/core/types"
)

func constructProposalMsgTest(acc *account.Account) *blockProposalMsg {
	txRoot := common.ComputeMerkleRoot(nil)
	vbftBlkInfo := &vconfig.VbftBlockInfo{
		Proposer:           1,
		LastConfigBlockNum: 12,
		NewChainConfig:     nil,
	}
	consensusPayload, err := json.Marshal(vbftBlkInfo)
	if err != nil {
		return nil
	}
	blkHeader := &types.Header{
		PrevBlockHash:    common.Uint256{},
		TransactionsRoot: txRoot,
		Timestamp:        uint32(time.Now().Unix()),
		Height:           uint32(20),
		ConsensusData:    uint64(123456),
		ConsensusPayload: consensusPayload,
		SigData:          [][]byte{{}, {}},
	}
	hash := blkHeader.Hash()
	sigdata, _ := signature.Sign(acc, hash[:])
	blkHeader.SigData[0] = sigdata
	blk := &VbftBlock{
		Block: &types.Block{
			Header:       blkHeader,
			Transactions: nil,
		},
		EmptyBlock: &types.Block{
			Header:       blkHeader,
			Transactions: nil,
		},
		Info:               vbftBlkInfo,
		PrevExecMerkleRoot: common.Uint256{},
	}
	msg := &blockProposalMsg{
		Block: blk,
	}

	return msg
}

func constructBlock() (*VbftBlock, error) {
	var txs []*types.Transaction
	txRoot := common.ComputeMerkleRoot(nil)
	vbftBlkInfo := &vconfig.VbftBlockInfo{
		Proposer:           1,
		LastConfigBlockNum: 1,
		NewChainConfig:     nil,
	}
	consensusPayload, err := json.Marshal(vbftBlkInfo)
	if err != nil {
		return nil, err
	}
	blkHeader := &types.Header{
		PrevBlockHash:    common.Uint256{},
		TransactionsRoot: txRoot,
		Timestamp:        uint32(time.Now().Unix()),
		Height:           uint32(1),
		ConsensusData:    uint64(123456),
		ConsensusPayload: consensusPayload,
		SigData:          [][]byte{{}, {}},
	}
	blk := &VbftBlock{
		Block: &types.Block{
			Header:       blkHeader,
			Transactions: txs,
		},
		EmptyBlock: &types.Block{
			Header:       blkHeader,
			Transactions: nil,
		},
		Info:               vbftBlkInfo,
		PrevExecMerkleRoot: common.Uint256{},
	}
	blk.Block.Hash()
	blk.Block.Transactions = txs
	return blk, nil
}
func TestBlockFetchRespMsgSerialize(t *testing.T) {
	blk, err := constructBlock()
	if err != nil {
		t.Errorf("constructBlock failed: %v", err)
		return
	}
	blockfetchrespmsg := &BlockFetchRespMsg{
		BlockNumber: 1,
		BlockHash:   common.Uint256{},
		BlockData:   blk,
	}
	_, err = blockfetchrespmsg.Serialize()
	if err != nil {
		t.Errorf("BlockFetchRespMsg Serialize failed: %v", err)
		return
	}
	t.Logf("BlockFetchRespMsg Serialize succ")
}

func TestBlockFetchRespMsgDeserialize(t *testing.T) {
	blk, err := constructBlock()
	if err != nil {
		t.Errorf("constructBlock failed: %v", err)
		return
	}
	blockfetchrespmsg := &BlockFetchRespMsg{
		BlockNumber: 1,
		BlockHash:   common.Uint256{},
		BlockData:   blk,
	}
	msg, err := blockfetchrespmsg.Serialize()
	if err != nil {
		t.Errorf("BlockFetchRespMsg Serialize failed: %v", err)
		return
	}
	respmsg := &BlockFetchRespMsg{}
	err = respmsg.Deserialize(msg)
	if err != nil {
		t.Errorf("BlockFetchRespMsg Deserialize failed: %v", err)
		return
	}
	t.Logf("BlockFetchRespMsg Serialize succ: %v\n", respmsg.BlockNumber)
}

func TestBlockSerialization(t *testing.T) {
	blk, err := constructBlock()
	if err != nil {
		t.Errorf("constructBlock failed: %v", err)
		return
	}

	data := blk.Serialize()

	blk2 := &VbftBlock{}
	if err := blk2.Deserialize(data); err != nil {
		t.Fatalf("deserialize blk: %s", err)
	}

	blk.EmptyBlock = nil
	data2 := blk.Serialize()
	blk3 := &VbftBlock{}
	if err := blk3.Deserialize(data2); err != nil {
		t.Fatalf("deserialize blk2: %s", err)
	}
}
