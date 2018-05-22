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
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/core/types"
)

type ConsensusMsgPayload struct {
	Type    MsgType `json:"type"`
	Len     uint32  `json:"len"`
	Payload []byte  `json:"payload"`
}

func DeserializeVbftMsg(msgPayload []byte) (ConsensusMsg, error) {

	m := &ConsensusMsgPayload{}
	if err := json.Unmarshal(msgPayload, m); err != nil {
		return nil, fmt.Errorf("unmarshal consensus msg payload: %s", err)
	}
	if m.Len < uint32(len(m.Payload)) {
		return nil, fmt.Errorf("invalid payload length: %d", m.Len)
	}

	switch m.Type {
	case BlockProposalMessage:
		t := &blockProposalMsg{}
		if err := t.UnmarshalJSON(m.Payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case BlockEndorseMessage:
		t := &blockEndorseMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case BlockCommitMessage:
		t := &blockCommitMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case PeerHandshakeMessage:
		t := &peerHandshakeMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case PeerHeartbeatMessage:
		t := &peerHeartbeatMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case BlockInfoFetchMessage:
		t := &BlockInfoFetchMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case BlockInfoFetchRespMessage:
		t := &BlockInfoFetchRespMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case BlockFetchMessage:
		t := &blockFetchMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case BlockFetchRespMessage:
		t := &BlockFetchRespMsg{}
		if err := t.Deserialize(m.Payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case ProposalFetchMessage:
		t := &proposalFetchMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	}

	return nil, fmt.Errorf("unknown msg type: %d", m.Type)
}

func SerializeVbftMsg(msg ConsensusMsg) ([]byte, error) {

	payload, err := msg.Serialize()
	if err != nil {
		return nil, err
	}

	return json.Marshal(&ConsensusMsgPayload{
		Type:    msg.Type(),
		Len:     uint32(len(payload)),
		Payload: payload,
	})
}

func (self *Server) constructHandshakeMsg() (*peerHandshakeMsg, error) {

	blkNum := self.GetCurrentBlockNo() - 1
	block, blockhash := self.blockPool.getSealedBlock(blkNum)
	if block == nil {
		return nil, fmt.Errorf("failed to get sealed block, current block: %d", self.GetCurrentBlockNo())
	}
	msg := &peerHandshakeMsg{
		CommittedBlockNumber: blkNum,
		CommittedBlockHash:   blockhash,
		CommittedBlockLeader: block.getProposer(),
		ChainConfig:          self.config,
	}

	return msg, nil
}

func (self *Server) constructHeartbeatMsg() (*peerHeartbeatMsg, error) {

	blkNum := self.GetCurrentBlockNo() - 1
	block, blockhash := self.blockPool.getSealedBlock(blkNum)
	if block == nil {
		return nil, fmt.Errorf("failed to get sealed block, current block: %d", self.GetCurrentBlockNo())
	}

	bookkeepers := make([][]byte, 0)
	endorsePks := block.Block.Header.Bookkeepers
	sigData := block.Block.Header.SigData
	if len(endorsePks) == len(sigData) {
		for i := 0; i < len(endorsePks); i++ {
			bookkeepers = append(bookkeepers, keypair.SerializePublicKey(endorsePks[i]))
		}
	} else {
		log.Errorf("Invalid signature counts in block %d: %d vs %d", blkNum, len(endorsePks), len(sigData))
		sigData = make([][]byte, 0)
	}

	msg := &peerHeartbeatMsg{
		CommittedBlockNumber: blkNum,
		CommittedBlockHash:   blockhash,
		CommittedBlockLeader: block.getProposer(),
		Endorsers:            bookkeepers,
		EndorsersSig:         sigData,
		ChainConfigView:      self.config.View,
	}

	return msg, nil
}

func (self *Server) constructBlock(blkNum uint32, prevBlkHash common.Uint256, txs []*types.Transaction, consensusPayload []byte) (*types.Block, error) {
	txHash := []common.Uint256{}
	for _, t := range txs {
		txHash = append(txHash, t.Hash())
	}
	txRoot := common.ComputeMerkleRoot(txHash)
	blockRoot := ledger.DefLedger.GetBlockRootWithNewTxRoot(txRoot)

	blkHeader := &types.Header{
		PrevBlockHash:    prevBlkHash,
		TransactionsRoot: txRoot,
		BlockRoot:        blockRoot,
		Timestamp:        uint32(time.Now().Unix()),
		Height:           uint32(blkNum),
		ConsensusData:    common.GetNonce(),
		ConsensusPayload: consensusPayload,
	}
	blk := &types.Block{
		Header:       blkHeader,
		Transactions: txs,
	}
	blkHash := blk.Hash()
	sig, err := signature.Sign(self.account, blkHash[:])
	if err != nil {
		return nil, fmt.Errorf("sign block failed, block hash：%x, error: %s", blkHash, err)
	}
	blkHeader.Bookkeepers = []keypair.PublicKey{self.account.PublicKey}
	blkHeader.SigData = [][]byte{sig}

	return blk, nil
}

func (self *Server) constructProposalMsg(blkNum uint32, sysTxs, userTxs []*types.Transaction, chainconfig *vconfig.ChainConfig) (*blockProposalMsg, error) {

	prevBlk, prevBlkHash := self.blockPool.getSealedBlock(blkNum - 1)
	if prevBlk == nil {
		return nil, fmt.Errorf("failed to get prevBlock (%d)", blkNum)
	}

	lastConfigBlkNum := prevBlk.Info.LastConfigBlockNum
	if prevBlk.Info.NewChainConfig != nil {
		lastConfigBlkNum = prevBlk.getBlockNum()
	}
	if chainconfig != nil {
		lastConfigBlkNum = blkNum
	}
	vbftBlkInfo := &vconfig.VbftBlockInfo{
		Proposer:           self.Index,
		LastConfigBlockNum: lastConfigBlkNum,
		NewChainConfig:     chainconfig,
	}
	consensusPayload, err := json.Marshal(vbftBlkInfo)
	if err != nil {
		return nil, err
	}

	emptyBlk, err := self.constructBlock(blkNum, prevBlkHash, sysTxs, consensusPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to construct empty block: %s", err)
	}
	blk, err := self.constructBlock(blkNum, prevBlkHash, append(sysTxs, userTxs...), consensusPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to constuct blk: %s", err)
	}

	msg := &blockProposalMsg{
		Block: &Block{
			Block:      blk,
			EmptyBlock: emptyBlk,
			Info:       vbftBlkInfo,
		},
	}

	return msg, nil
}

func (self *Server) constructEndorseMsg(proposal *blockProposalMsg, forEmpty bool) (*blockEndorseMsg, error) {

	// TODO, support faultyMsg reporting

	var proposerSig, endorserSig []byte
	var blkHash common.Uint256
	var err error
	if !forEmpty {
		proposerSig = proposal.Block.Block.Header.SigData[0]
		blkHash = proposal.Block.Block.Hash()

	} else {
		if proposal.Block.EmptyBlock == nil {
			return nil, fmt.Errorf("blk %d proposal from %d has no empty proposal",
				proposal.GetBlockNum(), proposal.Block.getProposer())
		}

		proposerSig = proposal.Block.EmptyBlock.Header.SigData[0]
		blkHash = proposal.Block.EmptyBlock.Hash()
	}
	endorserSig, err = signature.Sign(self.account, blkHash[:])
	if err != nil {
		return nil, fmt.Errorf("endorser failed to sign block. hash:%x, err: %s", blkHash, err)
	}

	msg := &blockEndorseMsg{
		Endorser:          self.Index,
		EndorsedProposer:  proposal.Block.getProposer(),
		BlockNum:          proposal.Block.getBlockNum(),
		EndorsedBlockHash: blkHash,
		EndorseForEmpty:   forEmpty,
		ProposerSig:       proposerSig,
		EndorserSig:       endorserSig,
	}

	return msg, nil
}

func (self *Server) constructCommitMsg(proposal *blockProposalMsg, endorses []*blockEndorseMsg, forEmpty bool) (*blockCommitMsg, error) {

	// TODO, support faultyMsg reporting

	var proposerSig, committerSig []byte
	var blkHash common.Uint256
	var err error

	if !forEmpty {
		proposerSig = proposal.Block.Block.Header.SigData[0]
		blkHash = proposal.Block.Block.Hash()
	} else {
		if proposal.Block.EmptyBlock == nil {
			return nil, fmt.Errorf("blk %d proposal from %d has no empty proposal",
				proposal.GetBlockNum(), proposal.Block.getProposer())
		}

		proposerSig = proposal.Block.EmptyBlock.Header.SigData[0]
		blkHash = proposal.Block.EmptyBlock.Hash()
	}
	committerSig, err = signature.Sign(self.account, blkHash[:])
	if err != nil {
		return nil, fmt.Errorf("endorser failed to sign block. hash:%x, caused by: %s", blkHash, err)
	}

	endorsersSig := make(map[uint32][]byte)
	for _, e := range endorses {
		endorsersSig[e.Endorser] = e.EndorserSig
	}

	msg := &blockCommitMsg{
		Committer:       self.Index,
		BlockProposer:   proposal.Block.getProposer(),
		BlockNum:        proposal.Block.getBlockNum(),
		CommitBlockHash: blkHash,
		CommitForEmpty:  forEmpty,
		ProposerSig:     proposerSig,
		EndorsersSig:    endorsersSig,
		CommitterSig:    committerSig,
	}

	return msg, nil
}

func (self *Server) constructBlockFetchMsg(blkNum uint32) (*blockFetchMsg, error) {
	msg := &blockFetchMsg{
		BlockNum: blkNum,
	}
	return msg, nil
}

func (self *Server) constructBlockFetchRespMsg(blkNum uint32, blk *Block, blkHash common.Uint256) (*BlockFetchRespMsg, error) {
	msg := &BlockFetchRespMsg{
		BlockNumber: blkNum,
		BlockHash:   blkHash,
		BlockData:   blk,
	}
	return msg, nil
}

func (self *Server) constructBlockInfoFetchMsg(startBlkNum uint32) (*BlockInfoFetchMsg, error) {

	msg := &BlockInfoFetchMsg{
		StartBlockNum: startBlkNum,
	}
	return msg, nil
}

func (self *Server) constructBlockInfoFetchRespMsg(blockInfos []*BlockInfo_) (*BlockInfoFetchRespMsg, error) {
	msg := &BlockInfoFetchRespMsg{
		Blocks: blockInfos,
	}
	return msg, nil
}

func (self *Server) constructProposalFetchMsg(blkNum uint32, proposer uint32) (*proposalFetchMsg, error) {
	msg := &proposalFetchMsg{
		ProposerID: proposer,
		BlockNum:   blkNum,
	}
	return msg, nil
}
