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
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
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

func (self *Server) constructProposalMsg(blkNum uint64, txs []*types.Transaction) (*blockProposalMsg, error) {

	prevBlk, prevBlkHash := self.blockPool.getSealedBlock(blkNum - 1)
	if prevBlk == nil {
		return nil, fmt.Errorf("failed to get prevBlock (%d)", blkNum)
	}

	txHash := []common.Uint256{}
	for _, t := range txs {
		txHash = append(txHash, t.Hash())
	}
	txRoot, err := common.ComputeMerkleRoot(txHash)
	if err != nil {
		return nil, fmt.Errorf("compute hash root: %s", err)
	}
	blockRoot := ledger.DefLedger.GetBlockRootWithNewTxRoot(txRoot)

	lastConfigBlkNum := prevBlk.Info.LastConfigBlockNum
	if prevBlk.Info.NewChainConfig != nil {
		lastConfigBlkNum = prevBlk.getBlockNum()
	}
	vbftBlkInfo := &vconfig.VbftBlockInfo{
		Proposer:           self.Index,
		LastConfigBlockNum: lastConfigBlkNum,
		NewChainConfig:     nil,
	}
	consensusPayload, err := json.Marshal(vbftBlkInfo)
	if err != nil {
		return nil, err
	}
	blkHeader := &types.Header{
		PrevBlockHash:    prevBlkHash,
		TransactionsRoot: txRoot,
		BlockRoot:        blockRoot,
		Timestamp:        uint32(time.Now().Unix()),
		Height:           uint32(blkNum),
		ConsensusData:    uint64(self.Index),
		ConsensusPayload: consensusPayload,
		SigData:          [][]byte{{}, {}},
	}
	blk := &Block{
		Block: &types.Block{
			Header: blkHeader,
		},
		Info: vbftBlkInfo,
	}
	blk.Block.Hash() // update block header hash
	msg := &blockProposalMsg{
		Block: blk,
	}

	emptySig, err := SignMsg(self.account, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to sign empty proposal: %s", err)
	}

	blk.Block.Transactions = txs
	sig, err := SignMsg(self.account, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to sign proposal: %s", err)
	}

	msg.Block.Block.Header.Bookkeepers = []keypair.PublicKey{self.account.PublicKey, self.account.PublicKey}
	msg.Block.Block.Header.SigData = [][]byte{sig, emptySig}
	return msg, nil
}

func (self *Server) constructEndorseMsg(proposal *blockProposalMsg, blkHash common.Uint256, forEmpty bool) (*blockEndorseMsg, error) {

	// TODO, support faultyMsg reporting

	var proposerSig, endorserSig []byte
	var err error
	if !forEmpty {
		proposerSig = proposal.Block.Block.Header.SigData[0]

		buf := new(bytes.Buffer)
		proposal.Block.Block.Header.SerializeUnsigned(buf)
		endorserSig, err = signature.Sign(self.account, buf.Bytes())
		if err != nil {
			return nil, fmt.Errorf("endorser failed to sign blkheader: %s", err)
		}
	} else {
		proposerSig = proposal.Block.Block.Header.SigData[1]

		txroot := proposal.Block.Block.Header.TransactionsRoot
		proposal.Block.Block.Header.TransactionsRoot = common.UINT256_EMPTY
		buf := new(bytes.Buffer)
		proposal.Block.Block.Header.SerializeUnsigned(buf)
		proposal.Block.Block.Header.TransactionsRoot = txroot
		endorserSig, err = signature.Sign(self.account, buf.Bytes())
		if err != nil {
			return nil, fmt.Errorf("endorser failed to sign blkheader: %s", err)
		}
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

func (self *Server) constructCommitMsg(proposal *blockProposalMsg, endorses []*blockEndorseMsg, blkHash common.Uint256, forEmpty bool) (*blockCommitMsg, error) {

	// TODO, support faultyMsg reporting

	var proposerSig, committerSig []byte
	var err error

	if !forEmpty {
		proposerSig = proposal.Block.Block.Header.SigData[0]
		buf := new(bytes.Buffer)
		proposal.Block.Block.Header.SerializeUnsigned(buf)
		committerSig, err = signature.Sign(self.account, buf.Bytes())
		if err != nil {
			return nil, fmt.Errorf("committer failed to sign blkheader: %s", err)
		}
	} else {
		proposerSig = proposal.Block.Block.Header.SigData[1]

		txroot := proposal.Block.Block.Header.TransactionsRoot
		proposal.Block.Block.Header.TransactionsRoot = common.UINT256_EMPTY
		buf := new(bytes.Buffer)
		proposal.Block.Block.Header.SerializeUnsigned(buf)
		proposal.Block.Block.Header.TransactionsRoot = txroot
		committerSig, err = signature.Sign(self.account, buf.Bytes())
		if err != nil {
			return nil, fmt.Errorf("endorser failed to sign blkheader: %s", err)
		}
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

func (self *Server) constructBlockFetchMsg(blkNum uint64) (*blockFetchMsg, error) {
	msg := &blockFetchMsg{
		BlockNum: blkNum,
	}
	return msg, nil
}

func (self *Server) constructBlockFetchRespMsg(blkNum uint64, blk *Block, blkHash common.Uint256) (*BlockFetchRespMsg, error) {
	msg := &BlockFetchRespMsg{
		BlockNumber: blkNum,
		BlockHash:   blkHash,
		BlockData:   blk,
	}
	return msg, nil
}

func (self *Server) constructBlockInfoFetchMsg(startBlkNum uint64) (*BlockInfoFetchMsg, error) {

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

func (self *Server) constructProposalFetchMsg(blkNum uint64, proposer uint32) (*proposalFetchMsg, error) {
	msg := &proposalFetchMsg{
		ProposerID: proposer,
		BlockNum:   blkNum,
	}
	return msg, nil
}
