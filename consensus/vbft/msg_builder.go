package vbft

import (
	"encoding/json"
	"fmt"

	. "github.com/Ontology/common"
	"github.com/Ontology/core/types"
)

type ConsensusMsgPayload struct {
	Type    MsgType `json:"type"`
	Len     uint32  `json:"len"`
	Payload []byte  `json:"payload"`
}

func DeserializeVbftMsg(msgPayload []byte) (ConsensusMsg, error) {

	m := &ConsensusMsgPayload{}
	if err := json.Unmarshal(msgPayload, m); err != nil {
		return nil, err
	}
	if m.Len < uint32(len(m.Payload)) {
		return nil, fmt.Errorf("invalid payload length: %d", m.Len)
	}

	switch m.Type {
	case blockProposalMessage:
		t := &blockProposalMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case blockEndorseMessage:
		t := &blockEndorseMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case blockCommitMessage:
		t := &blockCommitMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case peerHandshakeMessage:
		t := &peerHandshakeMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case peerHeartbeatMessage:
		t := &peerHeartbeatMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case blockInfoFetchMessage:
		t := &BlockInfoFetchMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case blockInfoFetchRespMessage:
		t := &BlockInfoFetchRespMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case blockFetchMessage:
		t := &blockFetchMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case blockFetchRespMessage:
		t := &BlockFetchRespMsg{}
		if err := json.Unmarshal(m.Payload, t); err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg (type: %d): %s", m.Type, err)
		}
		return t, nil
	case proposalFetchMessage:
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

func DeserializeBlock(data []byte, blk *Block) error {
	return json.Unmarshal(data, blk)
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

	sig, err := SignMsg(self.privateKey, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to sign handshake msg: %s", err)
	}
	msg.Sig = sig
	return msg, nil
}

func (self *Server) constructHeartbeatMsg() (*peerHeartbeatMsg, error) {

	blkNum := self.GetCurrentBlockNo() - 1
	block, blockhash := self.blockPool.getSealedBlock(blkNum)
	if block == nil {
		return nil, fmt.Errorf("failed to get sealed block, current block: %d", self.GetCurrentBlockNo())
	}
	msg := &peerHeartbeatMsg{
		CommittedBlockNumber: blkNum,
		CommittedBlockHash:   blockhash,
		CommittedBlockLeader: block.getProposer(),
		ChainConfigView:      self.config.View,
	}

	sig, err := SignMsg(self.privateKey, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to sign heartbeat msg: %s", err)
	}
	msg.Sig = sig
	return msg, nil
}

func (self *Server) constructProposalMsg(blkNum uint64, txs []*types.Transaction) (*blockProposalMsg, error) {

	prevBlk, prevBlkHash := self.blockPool.getSealedBlock(blkNum - 1)
	if prevBlk == nil {
		return nil, fmt.Errorf("failed to get prevBlock (%d)", blkNum)
	}

	lastConfigBlkNum := prevBlk.Info.LastConfigBlockNum
	if prevBlk.Info.NewChainConfig != nil {
		lastConfigBlkNum = prevBlk.getBlockNum()
	}
	vbftBlkInfo := &VbftBlockInfo{
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
		Height:           uint32(blkNum),
		ConsensusPayload: consensusPayload,
		SigData:          [][]byte{{}, {}},
	}
	blk := &Block{
		Block: &types.Block{
			Header: blkHeader,
		},
		Info: vbftBlkInfo,
	}
	msg := &blockProposalMsg{
		Block: blk,
	}

	emptySig, err := SignMsg(self.privateKey, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to sign empty proposal: %s", err)
	}

	blk.Block.Transactions = txs
	sig, err := SignMsg(self.privateKey, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to sign proposal: %s", err)
	}

	msg.Block.Block.Header.SigData[0] = sig
	msg.Block.Block.Header.SigData[1] = emptySig
	return msg, nil
}

func (self *Server) constructEndorseMsg(proposal *blockProposalMsg, blkHash Uint256, forEmpty bool) (*blockEndorseMsg, error) {

	// TODO, support faultyMsg reporting

	msg := &blockEndorseMsg{
		Endorser:          self.Index,
		EndorsedProposer:  proposal.Block.getProposer(),
		BlockNum:          proposal.Block.getBlockNum(),
		EndorsedBlockHash: blkHash,
		EndorseForEmpty:   forEmpty,
	}
	sig, err := SignMsg(self.privateKey, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to sign endorse msg: %s", err)
	}
	msg.Sig = sig
	return msg, nil
}

func (self *Server) constructCommitMsg(proposal *blockProposalMsg, blkHash Uint256, forEmpty bool) (*blockCommitMsg, error) {

	// TODO, support faultyMsg reporting

	msg := &blockCommitMsg{
		Committer:       self.Index,
		BlockProposer:   proposal.Block.getProposer(),
		BlockNum:        proposal.Block.getBlockNum(),
		CommitBlockHash: blkHash,
		CommitForEmpty:  forEmpty,
	}

	sig, err := SignMsg(self.privateKey, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to sign commit msg: %s", err)
	}
	msg.Sig = sig
	return msg, nil
}

func (self *Server) constructBlockFetchMsg(blkNum uint64) (*blockFetchMsg, error) {
	msg := &blockFetchMsg{
		BlockNum: blkNum,
	}
	sig, err := SignMsg(self.privateKey, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to sign blockfetch msg: %s", err)
	}

	msg.Sig = sig
	return msg, nil
}

func (self *Server) constructBlockFetchRespMsg(blkNum uint64, blk *Block, blkHash Uint256) (*BlockFetchRespMsg, error) {
	msg := &BlockFetchRespMsg{
		BlockNumber: blkNum,
		BlockHash:   blkHash,
		BlockData:   blk,
	}
	sig, err := SignMsg(self.privateKey, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to sign blockfetch-rsp msg: %s", err)
	}
	msg.Sig = sig
	return msg, nil
}

func (self *Server) constructBlockInfoFetchMsg(startBlkNum uint64) (*BlockInfoFetchMsg, error) {

	msg := &BlockInfoFetchMsg{
		StartBlockNum: startBlkNum,
	}
	sig, err := SignMsg(self.privateKey, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to sign blockinfo fetch req msg: %s", err)
	}
	msg.Sig = sig
	return msg, nil
}

func (self *Server) constructBlockInfoFetchRespMsg(blockInfos []*BlockInfo_) (*BlockInfoFetchRespMsg, error) {
	msg := &BlockInfoFetchRespMsg{
		Blocks: blockInfos,
	}
	sig, err := SignMsg(self.privateKey, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to sign blockinfo fetch rsp msg: %s", err)
	}
	msg.Sig = sig
	return msg, nil
}

func (self *Server) constructProposalFetchMsg(blkNum uint64) (*proposalFetchMsg, error) {
	msg := &proposalFetchMsg{
		BlockNum: blkNum,
	}
	sig, err := SignMsg(self.privateKey, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to sign proposalFetch msg: %s", err)
	}

	msg.Sig = sig
	return msg, nil
}
