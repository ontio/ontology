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
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/core/types"
	p2pmsg "github.com/ontio/ontology/p2pserver/message/types"
)

type ConsensusMsgPayload struct {
	Type    MsgType `json:"type"`
	Len     uint32  `json:"len"`
	Payload []byte  `json:"payload"`
}

func DeserializeVbftMsg(msg *p2pmsg.ConsensusPayload) (ConsensusMsg, error) {
	msgPayload := msg.Data
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
	case PeerHeartbeatMessage:
		t := &peerHeartbeatMsg{}
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
			return nil, fmt.Errorf("failed to Deserialize msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case ProposalFetchMessage:
		t := &proposalFetchMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case BlockSubmitMessage:
		t := &blockSubmitMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		t.Submitter = uint32(msg.BookkeeperIndex)
		return t, nil
	}

	return nil, fmt.Errorf("unknown msg type: %d", m.Type)
}

func MustSerializeVbftMsg(msg ConsensusMsg) []byte {
	data, err := SerializeVbftMsg(msg)
	if err != nil {
		panic(err)
	}

	return data
}

// TODO: serialize should never fail.
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

func (self *Server) constructHeartbeatMsg() (*peerHeartbeatMsg, error) {
	vbftCtx := self.GetVbftContext()
	blkNum := vbftCtx.BlockNum - 1
	block := vbftCtx.PrevBlockInfo.Block

	bookkeepers := make([][]byte, 0)
	endorsePks := block.Header.Bookkeepers
	sigData := block.Header.SigData
	if len(endorsePks) == len(sigData) {
		for i := 0; i < len(endorsePks); i++ {
			bookkeepers = append(bookkeepers, keypair.SerializePublicKey(endorsePks[i]))
		}
	} else {
		log.Errorf("Invalid signature counts in block %d: %d vs %d", blkNum, len(endorsePks), len(sigData))
		sigData = make([][]byte, 0)
	}

	msg := &peerHeartbeatMsg{
		CommittedBlockNumber:   blkNum,
		CommittedBlockHash:     block.Hash(),
		CommittedBlockProposer: vbftCtx.PrevBlockInfo.Info.Proposer,
		Endorsers:              bookkeepers,
		EndorsersSig:           sigData,
		ChainConfigView:        vbftCtx.Config.View,
	}

	return msg, nil
}

func constructBlock(blkNum uint32, prevBlock *types.Block, txs []*types.Transaction, consensusPayload []byte, blockTime uint32, nonce uint64) *types.Block {
	var txHash []common.Uint256
	for _, t := range txs {
		txHash = append(txHash, t.Hash())
	}
	txRoot := common.ComputeMerkleRoot(txHash)
	blockRoot := ledger.DefLedger.GetBlockRootWithNewTxRoots(prevBlock.Header.Height, []common.Uint256{prevBlock.Header.TransactionsRoot, txRoot})

	blkHeader := &types.Header{
		PrevBlockHash:    prevBlock.Hash(),
		TransactionsRoot: txRoot,
		BlockRoot:        blockRoot,
		Timestamp:        blockTime,
		Height:           blkNum,
		ConsensusData:    nonce,
		ConsensusPayload: consensusPayload,
	}
	blk := &types.Block{
		Header:       blkHeader,
		Transactions: txs,
	}
	return blk
}

func (self *Server) SignBlock(blk *types.Block) error {
	blkHash := blk.Hash()
	sig, err := signature.Sign(self.account, blkHash[:])
	if err != nil {
		return fmt.Errorf("sign block failed, block hash:%s, error: %s", blkHash.ToHexString(), err)
	}
	blk.Header.Bookkeepers = []keypair.PublicKey{self.account.PublicKey}
	blk.Header.SigData = [][]byte{sig}
	return nil
}

func (self *Server) constructProposalMsg(vbftCtx *VbftContext, userTxs []*types.Transaction, chainconfig *vconfig.ChainConfig) (*blockProposalMsg, error) {
	prevBlk := vbftCtx.PrevBlockInfo.Block
	blkNum := vbftCtx.BlockNum
	blockTime := uint32(time.Now().Unix())
	if prevBlk.Header.Timestamp >= blockTime {
		blockTime = prevBlk.Header.Timestamp + 1
	}

	vrfValue, vrfProof, err := computeVrf(self.account.PrivateKey, blkNum, vbftCtx.PrevBlockInfo.Info.VrfValue)
	if err != nil {
		return nil, fmt.Errorf("failed to get vrf and proof: %s", err)
	}

	nonce := common.GetNonce()
	proposal := BuildProposalMsg(vbftCtx, userTxs, chainconfig, nonce, nonce, blockTime, self.Index, vrfValue, vrfProof)
	err = self.SignBlock(proposal.Block.EmptyBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to construct empty block: %s", err)
	}
	err = self.SignBlock(proposal.Block.Block)
	if err != nil {
		return nil, fmt.Errorf("failed to constuct blk: %s", err)
	}
	return proposal, nil
}

func BuildProposalMsg(vbftCtx *VbftContext, userTxs []*types.Transaction, chainconfig *vconfig.ChainConfig, nonce, emptyNonce uint64, blockTime, proposer uint32, vrfValue, vrfProof []byte) *blockProposalMsg {
	prevBlk := vbftCtx.PrevBlockInfo.Block
	blkNum := vbftCtx.BlockNum

	lastConfigBlkNum := vbftCtx.PrevBlockInfo.Info.LastConfigBlockNum
	if vbftCtx.PrevBlockInfo.Info.NewChainConfig != nil {
		lastConfigBlkNum = prevBlk.Header.Height
	}
	vbftBlkInfo := &vconfig.VbftBlockInfo{
		Proposer:           proposer,
		VrfValue:           vrfValue,
		VrfProof:           vrfProof,
		LastConfigBlockNum: lastConfigBlkNum,
		NewChainConfig:     chainconfig,
	}
	consensusPayload, err := json.Marshal(vbftBlkInfo)
	if err != nil {
		panic(err)
	}
	var sysTxs []*types.Transaction
	if vbftCtx.NeedUpdateChainConfigTx() {
		sysTxs = append(sysTxs, CreateGovernaceTransaction(blkNum))
	}

	emptyBlk := constructBlock(blkNum, prevBlk, sysTxs, consensusPayload, blockTime, emptyNonce)
	blk := constructBlock(blkNum, prevBlk, append(sysTxs, userTxs...), consensusPayload, blockTime, nonce)
	msg := &blockProposalMsg{
		Block: &VbftBlock{
			Block:              blk,
			EmptyBlock:         emptyBlk,
			Info:               vbftBlkInfo,
			PrevExecMerkleRoot: vbftCtx.PrevBlockInfo.MerkleRoot,
		},
	}
	return msg
}

func (self *Server) constructEndorseMsg(proposal *blockProposalMsg, forEmpty bool) (*blockEndorseMsg, error) {

	// TODO, support faultyMsg reporting

	var endorserSig []byte
	var blkHash common.Uint256
	var err error
	if !forEmpty {
		blkHash = proposal.Block.Block.Hash()
	} else {
		if proposal.Block.EmptyBlock == nil {
			return nil, fmt.Errorf("blk %d proposal from %d has no empty proposal",
				proposal.GetBlockNum(), proposal.Block.getProposer())
		}

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
		EndorserSig:       endorserSig,
	}
	return msg, nil
}

func (self *Server) constructCommitMsg(proposal *blockProposalMsg, endorses map[uint32]*EndorseSigInfo, forEmpty bool) (*blockCommitMsg, error) {

	// TODO, support faultyMsg reporting

	var committerSig []byte
	var blkHash common.Uint256
	var err error

	if !forEmpty {
		blkHash = proposal.Block.Block.Hash()
	} else {
		if proposal.Block.EmptyBlock == nil {
			return nil, fmt.Errorf("blk %d proposal from %d has no empty proposal",
				proposal.GetBlockNum(), proposal.Block.getProposer())
		}

		blkHash = proposal.Block.EmptyBlock.Hash()
	}
	committerSig, err = signature.Sign(self.account, blkHash[:])
	if err != nil {
		return nil, fmt.Errorf("endorser failed to sign block. hash:%x, caused by: %s", blkHash, err)
	}

	endorsersSig := make(map[uint32][]byte)
	for endorser, e := range endorses {
		endorsersSig[endorser] = e.Signature
	}

	msg := &blockCommitMsg{
		Committer:       self.Index,
		BlockProposer:   proposal.Block.getProposer(),
		BlockNum:        proposal.Block.getBlockNum(),
		CommitBlockHash: blkHash,
		CommitForEmpty:  forEmpty,
		EndorsersSig:    endorsersSig,
		CommitterSig:    committerSig,
	}
	return msg, nil
}

func (self *Server) constructBlockFetchRespMsg(blkNum uint32, blk *VbftBlock, blkHash common.Uint256) *BlockFetchRespMsg {
	return &BlockFetchRespMsg{
		BlockNumber: blkNum,
		BlockHash:   blkHash,
		BlockData:   blk,
	}
}

func (self *Server) constructProposalFetchMsg(blkNum uint32, proposer uint32) *proposalFetchMsg {
	return &proposalFetchMsg{
		ProposerID: proposer,
		BlockNum:   blkNum,
	}
}

func (self *Server) constructBlockSubmitMsg(blkNum uint32, stateRoot common.Uint256) (*blockSubmitMsg, error) {
	submitSig, err := signature.Sign(self.account, stateRoot[:])
	if err != nil {
		return nil, fmt.Errorf("submit failed to sign stateroot hash:%x, err: %s", stateRoot, err)
	}
	msg := &blockSubmitMsg{
		BlockStateRoot: stateRoot,
		BlockNum:       blkNum,
		SubmitMsgSig:   submitSig,
	}
	return msg, nil
}
