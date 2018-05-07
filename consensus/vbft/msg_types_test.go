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
	"testing"
	"time"

	"github.com/ontio/ontology/account"
	common "github.com/ontio/ontology/common"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
)

func constructProposalMsg(acc *account.Account) (*blockProposalMsg, error) {
	bookKeepingPayload := &payload.Bookkeeping{
		Nonce: uint64(time.Now().UnixNano()),
	}
	tx := &types.Transaction{
		TxType:     types.BookKeeping,
		Payload:    bookKeepingPayload,
		Attributes: []*types.TxAttribute{},
	}
	var txs []*types.Transaction
	txs = append(txs, tx)
	txHash := []common.Uint256{}
	for _, t := range txs {
		txHash = append(txHash, t.Hash())
	}
	txRoot, err := common.ComputeMerkleRoot(txHash)
	if err != nil {
		return nil, fmt.Errorf("compute hash root: %s", err)
	}
	vbftBlkInfo := &vconfig.VbftBlockInfo{
		Proposer:           1,
		LastConfigBlockNum: uint64(12),
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
		Height:           uint32(20),
		ConsensusData:    uint64(123456),
		ConsensusPayload: consensusPayload,
		SigData:          [][]byte{{}, {}},
	}
	blk := &Block{
		Block: &types.Block{
			Header: blkHeader,
		},
		Info: vbftBlkInfo,
	}
	blk.Block.Hash()
	msg := &blockProposalMsg{
		Block: blk,
	}
	emptySig, err := SignMsg(acc, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to sign empty proposal: %s", err)
	}

	blk.Block.Transactions = txs
	sig, err := SignMsg(acc, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to sign proposal: %s", err)
	}

	msg.Block.Block.Header.SigData[0] = sig
	msg.Block.Block.Header.SigData[1] = emptySig
	return msg, nil
}

func TestBlockProposalMsgVerify(t *testing.T) {
	acc := account.NewAccount("SHA256withECDSA")
	if acc == nil {
		t.Error("GetDefaultAccount error: acc is nil")
		return
	}
	msg, err := constructProposalMsg(acc)
	if err != nil {
		t.Errorf("constructProposalMsg failed:%v", err)
		return
	}
	err = msg.Verify(acc.PublicKey)
	if err != nil {
		t.Errorf("blockPropoaslMsg Verify Failed: %v", err)
		return
	}
	t.Log("TestBlockProposalMsgVerify Verify succ\n")
}

func constructEndorseMsg(acc *account.Account, proposal *blockProposalMsg, blkHash common.Uint256) (*blockEndorseMsg, error) {
	msg := &blockEndorseMsg{
		Endorser:          5,
		EndorsedProposer:  proposal.Block.getProposer(),
		BlockNum:          proposal.Block.getBlockNum(),
		EndorsedBlockHash: blkHash,
		EndorseForEmpty:   true,
	}
	return msg, nil
}

func TestBlockEndorseMsg(t *testing.T) {
	acc := account.NewAccount("SHA256withECDSA")
	if acc == nil {
		t.Error("GetDefaultAccount error: acc is nil")
		return
	}
	block, err := constructProposalMsg(acc)
	if err != nil {
		t.Errorf("TestBlockEndorseMsg failed: %v", err)
		return
	}
	blkHash, _ := HashBlock(block.Block)
	endorsemsg, err := constructEndorseMsg(acc, block, blkHash)
	if err != nil {
		t.Errorf("TestBlockEndorseMsg failed: %v", err)
		return
	}
	err = endorsemsg.Verify(acc.PublicKey)
	if err != nil {
		t.Errorf("TestBlockEndorseMsg Verify failed: %v", err)
		return
	}
	t.Log("TestBlockEndorseMsg succ")
}

func constructCommitMsg(acc *account.Account, proposal *blockProposalMsg, blkHash common.Uint256) (*blockCommitMsg, error) {
	msg := &blockCommitMsg{
		Committer:       5,
		BlockProposer:   proposal.Block.getProposer(),
		BlockNum:        proposal.Block.getBlockNum(),
		CommitBlockHash: blkHash,
		CommitForEmpty:  true,
	}

	return msg, nil
}
func TestBlockCommitMsg(t *testing.T) {
	acc := account.NewAccount("SHA256withECDSA")
	if acc == nil {
		t.Error("GetDefaultAccount error: acc is nil")
		return
	}
	block, err := constructProposalMsg(acc)
	if err != nil {
		t.Errorf("TestBlockCommitMsg failed: %v", err)
		return
	}
	blkHash, _ := HashBlock(block.Block)
	commitmsg, err := constructCommitMsg(acc, block, blkHash)
	if err != nil {
		t.Errorf("TestBlockCommitMsg failed: %v", err)
		return
	}
	err = commitmsg.Verify(acc.PublicKey)
	if err != nil {
		t.Errorf("TestBlockCommitMsg Verify failed: %v", err)
		return
	}
	t.Log("TestBlockCommitMsg succ")
}

func constructHandshakeMsg(acc *account.Account) (*peerHandshakeMsg, error) {
	cc := &vconfig.ChainConfig{}
	msg := &peerHandshakeMsg{
		CommittedBlockNumber: uint64(1),
		CommittedBlockHash:   common.Uint256{},
		CommittedBlockLeader: uint32(3),
		ChainConfig:          cc,
	}
	return msg, nil
}
func TestPeerHandshakeMsg(t *testing.T) {
	acc := account.NewAccount("SHA256withECDSA")
	if acc == nil {
		t.Error("GetDefaultAccount error: acc is nil")
		return
	}
	msg, err := constructHandshakeMsg(acc)
	if err != nil {
		t.Errorf("constructHandshakeMsg failed: %v", err)
		return
	}
	err = msg.Verify(acc.PublicKey)
	if err != nil {
		t.Errorf("peerHandshakeMsg Verify failed: %v\n", err)
		return
	}
	t.Log("TestPeerHandshakeMsg succ")
}

func constructHeartbeatMsg(acc *account.Account) (*peerHeartbeatMsg, error) {
	msg := &peerHeartbeatMsg{
		CommittedBlockNumber: 5,
		CommittedBlockHash:   common.Uint256{},
		CommittedBlockLeader: uint32(1),
		ChainConfigView:      uint32(1),
	}

	return msg, nil
}
func TestPeerHeartbeatMsg(t *testing.T) {

	acc := account.NewAccount("SHA256withECDSA")
	if acc == nil {
		t.Error("GetDefaultAccount error: acc is nil")
		return
	}
	msg, err := constructHeartbeatMsg(acc)
	if err != nil {
		t.Errorf("constructHeartbeatMsg failed %v", err)
		return
	}
	err = msg.Verify(acc.PublicKey)
	if err != nil {
		t.Errorf("TestPeerHeartbeatMsg Verify failed %v", err)
		return
	}
	t.Log("TestPeerHeartbeatMsg succ")
}

func constructBlockInfoFetchMsg(acc *account.Account) (*BlockInfoFetchMsg, error) {
	msg := &BlockInfoFetchMsg{
		StartBlockNum: uint64(1),
	}

	return msg, nil
}

func TestBlockInfoFetchMsg(t *testing.T) {
	acc := account.NewAccount("SHA256withECDSA")
	if acc == nil {
		t.Error("GetDefaultAccount error: acc is nil")
		return
	}
	msg, err := constructBlockInfoFetchMsg(acc)
	if err != nil {
		t.Errorf("constructBlockInfoFetchMsg failed: %v", err)
		return
	}
	err = msg.Verify(acc.PublicKey)
	if err != nil {
		t.Errorf("TestBlockInfoFetchMsg Verify failed %v", err)
		return
	}
	t.Log("TestBlockInfoFetchMsg succ")
}

func constructBlockInfoFetchRespMsg(acc *account.Account) (*BlockInfoFetchRespMsg, error) {
	blockInfo := &BlockInfo_{
		BlockNum: uint64(1),
		Proposer: uint32(1),
	}
	var blockInfos []*BlockInfo_
	blockInfos = append(blockInfos, blockInfo)
	msg := &BlockInfoFetchRespMsg{
		Blocks: blockInfos,
	}
	return msg, nil
}

func TestBlockInfoFetchRespMsg(t *testing.T) {
	acc := account.NewAccount("SHA256withECDSA")
	if acc == nil {
		t.Error("GetDefaultAccount error: acc is nil")
		return
	}
	msg, err := constructBlockInfoFetchRespMsg(acc)
	if err != nil {
		t.Errorf("constructBlockInfoFetchRespMsg failed: %v", err)
		return
	}
	err = msg.Verify(acc.PublicKey)
	if err != nil {
		t.Errorf("TestBlockInfoFetchRespMsg Verify failed %v", err)
		return
	}
	t.Log("TestBlockInfoFetchRespMsg succ")
}

func constructBlockFetchMsg(acc *account.Account) (*blockFetchMsg, error) {
	msg := &blockFetchMsg{
		BlockNum: uint64(1),
	}

	return msg, nil
}
func TestBlockFetchMsg(t *testing.T) {
	acc := account.NewAccount("SHA256withECDSA")
	if acc == nil {
		t.Error("GetDefaultAccount error: acc is nil")
		return
	}
	msg, err := constructBlockFetchMsg(acc)
	if err != nil {
		t.Errorf("constructBlockFetchMsg failed: %v", err)
		return
	}
	err = msg.Verify(acc.PublicKey)
	if err != nil {
		t.Errorf("TestBlockFetchMsg Verify failed %v", err)
		return
	}
	t.Log("TestBlockFetchMsg succ")
}

func constructBlockFetchRespMsg(acc *account.Account, blk *Block) (*BlockFetchRespMsg, error) {
	msg := &BlockFetchRespMsg{
		BlockNumber: uint64(1),
		BlockHash:   common.Uint256{},
		BlockData:   blk,
	}

	return msg, nil
}

func TestBlockFetchRespMsg(t *testing.T) {
	acc := account.NewAccount("SHA256withECDSA")
	if acc == nil {
		t.Error("GetDefaultAccount error: acc is nil")
		return
	}
	msg, err := constructProposalMsg(acc)
	if err != nil {
		t.Errorf("constructProposalMsg failed:%v", err)
		return
	}
	respmsg, err := constructBlockFetchRespMsg(acc, msg.Block)
	if err != nil {
		t.Errorf("constructBlockFetchMsg failed :%v", err)
		return
	}
	err = respmsg.Verify(acc.PublicKey)
	if err != nil {
		t.Errorf("blockFetchRespMsg Verify Failed: %v", err)
		return
	}
	t.Log("TestBlockFetchRespMsg Verify succ\n")
}

func constructProposalFetchMsg(acc *account.Account) (*proposalFetchMsg, error) {
	msg := &proposalFetchMsg{
		BlockNum: uint64(1),
	}
	return msg, nil
}
func TestProposalFetchMsg(t *testing.T) {
	acc := account.NewAccount("SHA256withECDSA")
	if acc == nil {
		t.Error("GetDefaultAccount error: acc is nil")
		return
	}
	msg, err := constructProposalFetchMsg(acc)
	if err != nil {
		t.Errorf("constructProposalFetchMsg failed: %v", err)
		return
	}
	err = msg.Verify(acc.PublicKey)
	if err != nil {
		t.Errorf("TestProposalFetchMsg Verify failed %v", err)
		return
	}
	t.Log("TestProposalFetchMsg succ")
}

func constructBlock() (*Block, error) {
	bookKeepingPayload := &payload.Bookkeeping{
		Nonce: uint64(time.Now().UnixNano()),
	}
	tx := &types.Transaction{
		TxType:     types.BookKeeping,
		Payload:    bookKeepingPayload,
		Attributes: []*types.TxAttribute{},
	}
	var txs []*types.Transaction
	txs = append(txs, tx)
	txHash := []common.Uint256{}
	for _, t := range txs {
		txHash = append(txHash, t.Hash())
	}
	txRoot, err := common.ComputeMerkleRoot(txHash)
	if err != nil {
		return nil, fmt.Errorf("compute hash root: %s", err)
	}
	vbftBlkInfo := &vconfig.VbftBlockInfo{
		Proposer:           1,
		LastConfigBlockNum: uint64(1),
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
	blk := &Block{
		Block: &types.Block{
			Header: blkHeader,
			Transactions:txs,
		},
		Info: vbftBlkInfo,
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
		BlockNumber: uint64(1),
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
		BlockNumber: uint64(1),
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
