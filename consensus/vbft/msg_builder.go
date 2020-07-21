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
	cfg := self.GetChainConfig()
	msg := &peerHandshakeMsg{
		CommittedBlockNumber: blkNum,
		CommittedBlockHash:   blockhash,
		CommittedBlockLeader: block.getProposer(),
		ChainConfig:          &cfg,
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
		ChainConfigView:      self.GetChainConfig().View,
	}

	return msg, nil
}

func (self *Server) constructBlock(blkNum uint32, prevBlkHash common.Uint256, txs []*types.Transaction, consensusPayload []byte, blocktimestamp uint32) (*types.Block, error) {
	txHash := []common.Uint256{}
	for _, t := range txs {
		txHash = append(txHash, t.Hash())
	}
	lastBlock, _ := self.blockPool.getSealedBlock(blkNum - 1)
	if lastBlock == nil {
		log.Errorf("constructBlock getlastblock failed blknum:%d", blkNum-1)
		return nil, fmt.Errorf("constructBlock getlastblock failed blknum:%d", blkNum-1)
	}

	txRoot := common.ComputeMerkleRoot(txHash)
	blockRoot := ledger.DefLedger.GetBlockRootWithNewTxRoots(lastBlock.Block.Header.Height, []common.Uint256{lastBlock.Block.Header.TransactionsRoot, txRoot})

	blkHeader := &types.Header{
		PrevBlockHash:    prevBlkHash,
		TransactionsRoot: txRoot,
		BlockRoot:        blockRoot,
		Timestamp:        blocktimestamp,
		Height:           blkNum,
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
		return nil, fmt.Errorf("sign block failed, block hash:%s, error: %s", blkHash.ToHexString(), err)
	}
	blkHeader.Bookkeepers = []keypair.PublicKey{self.account.PublicKey}
	blkHeader.SigData = [][]byte{sig}

	return blk, nil
}

func (self *Server) constructCrossChainMsg(blkNum uint32) (*types.CrossChainMsg, error) {
	root, err := self.blockPool.getCrossStatesRoot(blkNum)
	if err != nil {
		return nil, err
	}
	log.Debugf("submitBlock height:%d statesroot:%+v", blkNum, root)
	if root == common.UINT256_EMPTY {
		return nil, nil
	}
	msg := &types.CrossChainMsg{
		Version:    types.CURR_CROSS_STATES_VERSION,
		Height:     blkNum,
		StatesRoot: root,
	}
	hash := msg.Hash()
	sig, err := signature.Sign(self.account, hash[:])
	if err != nil {
		return nil, fmt.Errorf("sign cross chain msg root failed,msg hash:%s,err:%s", hash.ToHexString(), err)
	}
	msg.SigData = append(msg.SigData, sig)
	return msg, nil
}

func (self *Server) constructProposalMsg(blkNum uint32, sysTxs, userTxs []*types.Transaction, chainconfig *vconfig.ChainConfig) (*blockProposalMsg, error) {

	prevBlk, prevBlkHash := self.blockPool.getSealedBlock(blkNum - 1)
	if prevBlk == nil {
		return nil, fmt.Errorf("failed to get prevBlock (%d)", blkNum-1)
	}
	blocktimestamp := uint32(time.Now().Unix())
	if prevBlk.Block.Header.Timestamp >= blocktimestamp {
		blocktimestamp = prevBlk.Block.Header.Timestamp + 1
	}

	vrfValue, vrfProof, err := computeVrf(self.account.PrivateKey, blkNum, prevBlk.getVrfValue())
	if err != nil {
		return nil, fmt.Errorf("failed to get vrf and proof: %s", err)
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
		VrfValue:           vrfValue,
		VrfProof:           vrfProof,
		LastConfigBlockNum: lastConfigBlkNum,
		NewChainConfig:     chainconfig,
	}
	consensusPayload, err := json.Marshal(vbftBlkInfo)
	if err != nil {
		return nil, err
	}

	emptyBlk, err := self.constructBlock(blkNum, prevBlkHash, sysTxs, consensusPayload, blocktimestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to construct empty block: %s", err)
	}
	blk, err := self.constructBlock(blkNum, prevBlkHash, append(sysTxs, userTxs...), consensusPayload, blocktimestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to constuct blk: %s", err)
	}
	merkleRoot, err := self.blockPool.getExecMerkleRoot(blkNum - 1)
	if err != nil {
		return nil, fmt.Errorf("failed to GetExecMerkleRoot: %s,blkNum:%d", err, blkNum-1)
	}
	crossChainMsg, err := self.constructCrossChainMsg(blkNum - 1)
	if err != nil {
		return nil, fmt.Errorf("failed to crossChainMsgHash :%s,blkNum:%d", err, (blkNum - 1))
	}
	msg := &blockProposalMsg{
		Block: &Block{
			Block:              blk,
			EmptyBlock:         emptyBlk,
			Info:               vbftBlkInfo,
			PrevExecMerkleRoot: merkleRoot,
			CrossChainMsg:      crossChainMsg,
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
		proposerSig = proposal.BlockProposerSig
		blkHash = proposal.Block.Block.Hash()

	} else {
		if proposal.Block.EmptyBlock == nil {
			return nil, fmt.Errorf("blk %d proposal from %d has no empty proposal",
				proposal.GetBlockNum(), proposal.Block.getProposer())
		}

		proposerSig = proposal.EmptyBlockProposerSig
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
	if proposal.Block.CrossChainMsg != nil {
		hash := proposal.Block.CrossChainMsg.Hash()
		sig, err := signature.Sign(self.account, hash[:])
		if err != nil {
			return nil, fmt.Errorf("sign cross chain msg root failed,msg hash:%s,err:%s", hash.ToHexString(), err)
		}
		msg.CrossChainMsgEndorserSig = sig
		msg.CrossChainMsgHash = hash
	}
	return msg, nil
}

func (self *Server) constructCommitMsg(proposal *blockProposalMsg, endorses []*blockEndorseMsg, forEmpty bool) (*blockCommitMsg, error) {

	// TODO, support faultyMsg reporting

	var proposerSig, committerSig []byte
	var blkHash common.Uint256
	var err error

	if !forEmpty {
		proposerSig = proposal.BlockProposerSig
		blkHash = proposal.Block.Block.Hash()
	} else {
		if proposal.Block.EmptyBlock == nil {
			return nil, fmt.Errorf("blk %d proposal from %d has no empty proposal",
				proposal.GetBlockNum(), proposal.Block.getProposer())
		}

		proposerSig = proposal.EmptyBlockProposerSig
		blkHash = proposal.Block.EmptyBlock.Hash()
	}
	committerSig, err = signature.Sign(self.account, blkHash[:])
	if err != nil {
		return nil, fmt.Errorf("endorser failed to sign block. hash:%x, caused by: %s", blkHash, err)
	}

	commitCrossChain := true
	endorsersSig := make(map[uint32][]byte)
	crossChainEndorserSig := make(map[uint32][]byte)
	var ccmCommitSig []byte
	for _, e := range endorses {
		endorsersSig[e.Endorser] = e.EndorserSig
		crossChainEndorserSig[e.Endorser] = e.CrossChainMsgEndorserSig
		if e.Endorser == self.Index {
			commitCrossChain = false
			ccmCommitSig = e.CrossChainMsgEndorserSig
		}
	}

	var hash common.Uint256
	if proposal.Block.CrossChainMsg != nil {
		hash = proposal.Block.CrossChainMsg.Hash()
	}

	msg := &blockCommitMsg{
		Committer:                 self.Index,
		BlockProposer:             proposal.Block.getProposer(),
		BlockNum:                  proposal.Block.getBlockNum(),
		CommitBlockHash:           blkHash,
		CommitForEmpty:            forEmpty,
		ProposerSig:               proposerSig,
		EndorsersSig:              endorsersSig,
		CommitterSig:              committerSig,
		CommitCCMHash:             hash,
		CrossChainMsgEndorserSig:  crossChainEndorserSig,
		CrossChainMsgCommitterSig: ccmCommitSig,
	}

	if proposal.Block.CrossChainMsg != nil && commitCrossChain {
		sig, err := signature.Sign(self.account, hash[:])
		if err != nil {
			return nil, fmt.Errorf("sign cross chain msg root failed,msg hash:%s,err:%s", hash.ToHexString(), err)
		}
		msg.CrossChainMsgCommitterSig = sig
	}
	return msg, nil
}

func (self *Server) constructBlockFetchMsg(blkNum uint32) *blockFetchMsg {
	return &blockFetchMsg{
		BlockNum: blkNum,
	}
}

func (self *Server) constructBlockFetchRespMsg(blkNum uint32, blk *Block, blkHash common.Uint256) *BlockFetchRespMsg {
	return &BlockFetchRespMsg{
		BlockNumber: blkNum,
		BlockHash:   blkHash,
		BlockData:   blk,
	}
}

func (self *Server) constructBlockInfoFetchMsg(startBlkNum uint32) *BlockInfoFetchMsg {
	return &BlockInfoFetchMsg{
		StartBlockNum: startBlkNum,
	}
}

func (self *Server) constructBlockInfoFetchRespMsg(blockInfos []*BlockInfo_) *BlockInfoFetchRespMsg {
	return &BlockInfoFetchRespMsg{
		Blocks: blockInfos,
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
