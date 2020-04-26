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
	blk := &Block{
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

func TestBlockProposalMsgVerify(t *testing.T) {
	acc := account.NewAccount("SHA256withECDSA")
	if acc == nil {
		t.Error("GetDefaultAccount error: acc is nil")
		return
	}
	msg := constructProposalMsgTest(acc)
	err := msg.Verify(acc.PublicKey)
	if err != nil {
		t.Errorf("blockPropoaslMsg Verify Failed: %v", err)
		return
	}
	t.Log("TestBlockProposalMsgVerify Verify succ\n")
}

func constructEndorseMsg(acc *account.Account, proposal *blockProposalMsg, blkHash common.Uint256) (*blockEndorseMsg, error) {
	sig, _ := signature.Sign(acc, blkHash[:])
	msg := &blockEndorseMsg{
		Endorser:          5,
		EndorsedProposer:  proposal.Block.getProposer(),
		BlockNum:          proposal.Block.getBlockNum(),
		EndorsedBlockHash: blkHash,
		EndorseForEmpty:   true,
		EndorserSig:       sig,
	}
	return msg, nil
}

func TestBlockEndorseMsg(t *testing.T) {
	acc := account.NewAccount("SHA256withECDSA")
	if acc == nil {
		t.Error("GetDefaultAccount error: acc is nil")
		return
	}
	block := constructProposalMsgTest(acc)
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
	sig, _ := signature.Sign(acc, blkHash[:])
	msg := &blockCommitMsg{
		Committer:       5,
		BlockProposer:   proposal.Block.getProposer(),
		BlockNum:        proposal.Block.getBlockNum(),
		CommitBlockHash: blkHash,
		CommitForEmpty:  true,
		CommitterSig:    sig,
	}

	return msg, nil
}
func TestBlockCommitMsg(t *testing.T) {
	acc := account.NewAccount("SHA256withECDSA")
	if acc == nil {
		t.Error("GetDefaultAccount error: acc is nil")
		return
	}
	block := constructProposalMsgTest(acc)
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
		CommittedBlockNumber: 1,
		CommittedBlockHash:   common.Uint256{},
		CommittedBlockLeader: 3,
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
		StartBlockNum: 1,
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
		BlockNum: 1,
		Proposer: 1,
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
		BlockNum: 1,
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
		BlockNumber: 1,
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
	msg := constructProposalMsgTest(acc)
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
		BlockNum: 1,
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
	blk := &Block{
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

	blk2 := &Block{}
	if err := blk2.Deserialize(data); err != nil {
		t.Fatalf("deserialize blk: %s", err)
	}

	blk.EmptyBlock = nil
	data2 := blk.Serialize()
	blk3 := &Block{}
	if err := blk3.Deserialize(data2); err != nil {
		t.Fatalf("deserialize blk2: %s", err)
	}
}
