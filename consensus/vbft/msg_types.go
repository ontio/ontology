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
	"errors"
	"fmt"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
)

type MsgType uint8

const (
	BlockProposalMessage MsgType = iota
	BlockEndorseMessage
	BlockCommitMessage

	PeerHandshakeMessage
	PeerHeartbeatMessage

	BlockInfoFetchMessage
	BlockInfoFetchRespMessage
	ProposalFetchMessage
	BlockFetchMessage
	BlockFetchRespMessage
)

type ConsensusMsg interface {
	Type() MsgType
	Verify(pub keypair.PublicKey) error
	GetBlockNum() uint32
	Serialize() ([]byte, error)
}

type blockProposalMsg struct {
	Block *Block `json:"block"`
}

func (msg *blockProposalMsg) Type() MsgType {
	return BlockProposalMessage
}

func (msg *blockProposalMsg) Verify(pub keypair.PublicKey) error {
	// verify block
	if len(msg.Block.Block.Header.SigData) == 0 {
		return errors.New("no sigdata in block")
	}
	sigdata := msg.Block.Block.Header.SigData[0]
	hash := msg.Block.Block.Hash()

	sig, err := signature.Deserialize(sigdata)
	if err != nil {
		return fmt.Errorf("deserialize block sig: %s", err)
	}
	if !signature.Verify(pub, hash[:], sig) {
		return fmt.Errorf("failed to verify block sig")
	}

	// verify empty block
	if msg.Block.EmptyBlock != nil {
		if len(msg.Block.EmptyBlock.Header.SigData) == 0 {
			return errors.New("no sigdata in empty block")
		}
		sigdata := msg.Block.EmptyBlock.Header.SigData[0]
		hash := msg.Block.EmptyBlock.Hash()
		sig, err := signature.Deserialize(sigdata)
		if err != nil {
			return fmt.Errorf("deserialize empty block sig: %s", err)
		}
		if !signature.Verify(pub, hash[:], sig) {
			return fmt.Errorf("failed to verify empty block sig")
		}
	}

	return nil
}

func (msg *blockProposalMsg) GetBlockNum() uint32 {
	return msg.Block.Block.Header.Height
}

func (msg *blockProposalMsg) Serialize() ([]byte, error) {
	return msg.Block.Serialize()
}

func (msg *blockProposalMsg) UnmarshalJSON(data []byte) error {
	blk := &Block{}
	if err := blk.Deserialize(data); err != nil {
		return err
	}

	msg.Block = blk
	return nil
}

func (msg *blockProposalMsg) MarshalJSON() ([]byte, error) {
	return msg.Block.Serialize()
}

type FaultyReport struct {
	FaultyID      uint32         `json:"faulty_id"`
	FaultyMsgHash common.Uint256 `json:"faulty_block_hash"`
}

type blockEndorseMsg struct {
	Endorser          uint32          `json:"endorser"`
	EndorsedProposer  uint32          `json:"endorsed_proposer"`
	BlockNum          uint32          `json:"block_num"`
	EndorsedBlockHash common.Uint256  `json:"endorsed_block_hash"`
	EndorseForEmpty   bool            `json:"endorse_for_empty"`
	FaultyProposals   []*FaultyReport `json:"faulty_proposals"`
	ProposerSig       []byte          `json:"proposer_sig"`
	EndorserSig       []byte          `json:"endorser_sig"`
}

func (msg *blockEndorseMsg) Type() MsgType {
	return BlockEndorseMessage
}

func (msg *blockEndorseMsg) Verify(pub keypair.PublicKey) error {
	hash := msg.EndorsedBlockHash
	sig, err := signature.Deserialize(msg.EndorserSig)
	if err != nil {
		return fmt.Errorf("deserialize block sig: %s", err)
	}
	if !signature.Verify(pub, hash[:], sig) {
		return fmt.Errorf("failed to verify block sig")
	}
	return nil
}

func (msg *blockEndorseMsg) GetBlockNum() uint32 {
	return msg.BlockNum
}

func (msg *blockEndorseMsg) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}

type blockCommitMsg struct {
	Committer       uint32            `json:"committer"`
	BlockProposer   uint32            `json:"block_proposer"`
	BlockNum        uint32            `json:"block_num"`
	CommitBlockHash common.Uint256    `json:"commit_block_hash"`
	CommitForEmpty  bool              `json:"commit_for_empty"`
	FaultyVerifies  []*FaultyReport   `json:"faulty_verifies"`
	ProposerSig     []byte            `json:"proposer_sig"`
	EndorsersSig    map[uint32][]byte `json:"endorsers_sig"`
	CommitterSig    []byte            `json:"committer_sig"`
}

func (msg *blockCommitMsg) Type() MsgType {
	return BlockCommitMessage
}

func (msg *blockCommitMsg) Verify(pub keypair.PublicKey) error {
	hash := msg.CommitBlockHash
	sig, err := signature.Deserialize(msg.CommitterSig)
	if err != nil {
		return fmt.Errorf("deserialize block sig: %s", err)
	}
	if !signature.Verify(pub, hash[:], sig) {
		return fmt.Errorf("failed to verify block sig")
	}

	return nil
}

func (msg *blockCommitMsg) GetBlockNum() uint32 {
	return msg.BlockNum
}

func (msg *blockCommitMsg) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}

type peerHandshakeMsg struct {
	CommittedBlockNumber uint32               `json:"committed_block_number"`
	CommittedBlockHash   common.Uint256       `json:"committed_block_hash"`
	CommittedBlockLeader uint32               `json:"committed_block_leader"`
	ChainConfig          *vconfig.ChainConfig `json:"chain_config"`
}

func (msg *peerHandshakeMsg) Type() MsgType {
	return PeerHandshakeMessage
}

func (msg *peerHandshakeMsg) Verify(pub keypair.PublicKey) error {

	return nil
}

func (msg *peerHandshakeMsg) GetBlockNum() uint32 {
	return 0
}

func (msg *peerHandshakeMsg) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}

type peerHeartbeatMsg struct {
	CommittedBlockNumber uint32         `json:"committed_block_number"`
	CommittedBlockHash   common.Uint256 `json:"committed_block_hash"`
	CommittedBlockLeader uint32         `json:"committed_block_leader"`
	Endorsers            [][]byte       `json:"endorsers"`
	EndorsersSig         [][]byte       `json:"endorsers_sig"`
	ChainConfigView      uint32         `json:"chain_config_view"`
}

func (msg *peerHeartbeatMsg) Type() MsgType {
	return PeerHeartbeatMessage
}

func (msg *peerHeartbeatMsg) Verify(pub keypair.PublicKey) error {
	return nil
}

func (msg *peerHeartbeatMsg) GetBlockNum() uint32 {
	return 0
}

func (msg *peerHeartbeatMsg) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}

type BlockInfoFetchMsg struct {
	StartBlockNum uint32 `json:"start_block_num"`
}

func (msg *BlockInfoFetchMsg) Type() MsgType {
	return BlockInfoFetchMessage
}

func (msg *BlockInfoFetchMsg) Verify(pub keypair.PublicKey) error {
	return nil
}

func (msg *BlockInfoFetchMsg) GetBlockNum() uint32 {
	return 0
}

func (msg *BlockInfoFetchMsg) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}

type BlockInfo_ struct {
	BlockNum   uint32            `json:"block_num"`
	Proposer   uint32            `json:"proposer"`
	Signatures map[uint32][]byte `json:"signatures"`
}

// to fetch committed block from neighbours
type BlockInfoFetchRespMsg struct {
	Blocks []*BlockInfo_ `json:"blocks"`
}

func (msg *BlockInfoFetchRespMsg) Type() MsgType {
	return BlockInfoFetchRespMessage
}

func (msg *BlockInfoFetchRespMsg) Verify(pub keypair.PublicKey) error {
	return nil
}

func (msg *BlockInfoFetchRespMsg) GetBlockNum() uint32 {
	return 0
}

func (msg *BlockInfoFetchRespMsg) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}

// block fetch msg is to fetch block which could have not been committed or endorsed
type blockFetchMsg struct {
	BlockNum uint32 `json:"block_num"`
}

func (msg *blockFetchMsg) Type() MsgType {
	return BlockFetchMessage
}

func (msg *blockFetchMsg) Verify(pub keypair.PublicKey) error {
	return nil
}

func (msg *blockFetchMsg) GetBlockNum() uint32 {
	return 0
}

func (msg *blockFetchMsg) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}

type BlockFetchRespMsg struct {
	BlockNumber uint32         `json:"block_number"`
	BlockHash   common.Uint256 `json:"block_hash"`
	BlockData   *Block         `json:"block_data"`
}

func (msg *BlockFetchRespMsg) Type() MsgType {
	return BlockFetchRespMessage
}

func (msg *BlockFetchRespMsg) Verify(pub keypair.PublicKey) error {
	return nil
}

func (msg *BlockFetchRespMsg) GetBlockNum() uint32 {
	return 0
}

func (msg *BlockFetchRespMsg) Serialize() ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	serialization.WriteUint32(buffer, msg.BlockNumber)
	msg.BlockHash.Serialize(buffer)
	blockbuff, err := msg.BlockData.Serialize()
	if err != nil {
		return nil, err
	}
	buffer.Write(blockbuff)
	return buffer.Bytes(), nil
}

func (msg *BlockFetchRespMsg) Deserialize(data []byte) error {
	buffer := bytes.NewBuffer(data)
	blocknum, err := serialization.ReadUint32(buffer)
	if err != nil {
		return err
	}
	msg.BlockNumber = blocknum
	err = msg.BlockHash.Deserialize(buffer)
	if err != nil {
		return err
	}
	blk := &Block{}
	if err := blk.Deserialize(buffer.Bytes()); err != nil {
		return fmt.Errorf("unmarshal block type: %s", err)
	}
	msg.BlockData = blk
	return nil
}

// proposal fetch msg is to fetch proposal when peer failed to get proposal locally
type proposalFetchMsg struct {
	ProposerID uint32 `json:"proposer_id"`
	BlockNum   uint32 `json:"block_num"`
}

func (msg *proposalFetchMsg) Type() MsgType {
	return ProposalFetchMessage
}

func (msg *proposalFetchMsg) Verify(pub keypair.PublicKey) error {
	return nil
}

func (msg *proposalFetchMsg) GetBlockNum() uint32 {
	return 0
}

func (msg *proposalFetchMsg) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}
