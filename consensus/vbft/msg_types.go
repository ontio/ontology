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

	common "github.com/Ontology/common"
	vconfig "github.com/Ontology/consensus/vbft/config"
	"github.com/Ontology/core/types"
	"github.com/Ontology/crypto"
	"github.com/Ontology/common/serialization"
)

type MsgType uint8

const (
	blockProposalMessage MsgType = iota
	blockEndorseMessage
	blockCommitMessage

	peerHandshakeMessage
	peerHeartbeatMessage

	blockInfoFetchMessage
	blockInfoFetchRespMessage
	proposalFetchMessage
	blockFetchMessage
	blockFetchRespMessage
)

type ConsensusMsg interface {
	Type() MsgType
	Verify(pub *crypto.PubKey) error
	GetBlockNum() uint64
	Serialize() ([]byte, error)
}

type blockProposalMsg struct {
	Block *Block `json:"block"`
}

func (msg *blockProposalMsg) Type() MsgType {
	return blockProposalMessage
}

func (msg *blockProposalMsg) Verify(pub *crypto.PubKey) error {

	// FIXME

	return nil
}

func (msg *blockProposalMsg) GetBlockNum() uint64 {
	return uint64(msg.Block.Block.Header.Height)
}

func (msg *blockProposalMsg) Serialize() ([]byte, error) {
	return msg.Block.Serialize()
}

func (msg *blockProposalMsg) UnmarshalJSON(data []byte) error {
	buf := bytes.NewBuffer(data)

	blk := &types.Block{}
	if err := blk.Deserialize(buf); err != nil {
		return fmt.Errorf("unmarshal block type: %s", err)
	}

	block, err := initVbftBlock(blk)
	if err != nil {
		return fmt.Errorf("init vbft block: %s", err)
	}
	msg.Block = block
	return nil
}

func (msg *blockProposalMsg) MarshalJSON() ([]byte, error) {
	return msg.Block.Serialize()
}

type FaultyReport struct {
	FaultyID      uint32  `json:"faulty_id"`
	FaultyMsgHash common.Uint256 `json:"faulty_block_hash"`
}

type blockEndorseMsg struct {
	Endorser          uint32          `json:"endorser"`
	EndorsedProposer  uint32          `json:"endorsed_proposer"`
	BlockNum          uint64          `json:"block_num"`
	EndorsedBlockHash common.Uint256         `json:"endorsed_block_hash"`
	EndorseForEmpty   bool            `json:"endorse_for_empty"`
	FaultyProposals   []*FaultyReport `json:"faulty_proposals"`
	Sig               []byte          `json:"sig"`
}

func (msg *blockEndorseMsg) Type() MsgType {
	return blockEndorseMessage
}

func (msg *blockEndorseMsg) Verify(pub *crypto.PubKey) error {
	sig := msg.Sig
	msg.Sig = nil

	defer func() {
		msg.Sig = sig
	}()

	if data, err := msg.Serialize(); err != nil {
		return fmt.Errorf("failed to serialize endorse msg: %s", err)
	} else if err := crypto.Verify(*pub, data, sig); err != nil {
		return fmt.Errorf("failed to verify endorse msg: %s", err)
	}

	return nil
}

func (msg *blockEndorseMsg) GetBlockNum() uint64 {
	return msg.BlockNum
}

func (msg *blockEndorseMsg) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}

type blockCommitMsg struct {
	Committer       uint32          `json:"committer"`
	BlockProposer   uint32          `json:"block_proposer"`
	BlockNum        uint64          `json:"block_num"`
	CommitBlockHash common.Uint256         `json:"commit_block_hash"`
	CommitForEmpty  bool            `json:"commit_for_empty"`
	FaultyVerifies  []*FaultyReport `json:"faulty_verifies"`
	Sig             []byte          `json:"sig"`
}

func (msg *blockCommitMsg) Type() MsgType {
	return blockCommitMessage
}

func (msg *blockCommitMsg) Verify(pub *crypto.PubKey) error {
	sig := msg.Sig
	msg.Sig = nil

	defer func() {
		msg.Sig = sig
	}()

	if data, err := msg.Serialize(); err != nil {
		return fmt.Errorf("failed to serialize commit msg: %s", err)
	} else if err := crypto.Verify(*pub, data, sig); err != nil {
		return fmt.Errorf("failed to verify commit msg: %s", err)
	}

	return nil
}

func (msg *blockCommitMsg) GetBlockNum() uint64 {
	return msg.BlockNum
}

func (msg *blockCommitMsg) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}

type peerHandshakeMsg struct {
	CommittedBlockNumber uint64               `json:"committed_block_number"`
	CommittedBlockHash   common.Uint256              `json:"committed_block_hash"`
	CommittedBlockLeader uint32               `json:"committed_block_leader"`
	ChainConfig          *vconfig.ChainConfig `json:"chain_config"`
	Sig                  []byte               `json:"sig"`
}

func (msg *peerHandshakeMsg) Type() MsgType {
	return peerHandshakeMessage
}

func (msg *peerHandshakeMsg) Verify(pub *crypto.PubKey) error {
	sig := msg.Sig
	msg.Sig = nil

	defer func() {
		msg.Sig = sig
	}()

	if data, err := msg.Serialize(); err != nil {
		return fmt.Errorf("failed to serialize handshake msg: %s", err)
	} else if err := crypto.Verify(*pub, data, sig); err != nil {
		return fmt.Errorf("failed to verify handshake msg: %s, data: %v", err, data)
	}

	return nil
}

func (msg *peerHandshakeMsg) GetBlockNum() uint64 {
	return 0
}

func (msg *peerHandshakeMsg) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}

type peerHeartbeatMsg struct {
	CommittedBlockNumber uint64  `json:"committed_block_number"`
	CommittedBlockHash   common.Uint256 `json:"committed_block_hash"`
	CommittedBlockLeader uint32  `json:"committed_block_leader"`
	ChainConfigView      uint32  `json:"chain_config_view"`
	Sig                  []byte  `json:"sig"`
}

func (msg *peerHeartbeatMsg) Type() MsgType {
	return peerHeartbeatMessage
}

func (msg *peerHeartbeatMsg) Verify(pub *crypto.PubKey) error {
	sig := msg.Sig
	msg.Sig = nil

	defer func() {
		msg.Sig = sig
	}()

	if data, err := msg.Serialize(); err != nil {
		return fmt.Errorf("failed to serialize heartbeat msg: %s", err)
	} else if err := crypto.Verify(*pub, data, sig); err != nil {
		return fmt.Errorf("failed to verify heartbeat msg: %s", err)
	}

	return nil
}

func (msg *peerHeartbeatMsg) GetBlockNum() uint64 {
	return 0
}

func (msg *peerHeartbeatMsg) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}

type BlockInfoFetchMsg struct {
	StartBlockNum uint64 `json:"start_block_num"`
	Sig           []byte `json:"sig"`
}

func (msg *BlockInfoFetchMsg) Type() MsgType {
	return blockInfoFetchMessage
}

func (msg *BlockInfoFetchMsg) Verify(pub *crypto.PubKey) error {
	sig := msg.Sig
	msg.Sig = nil

	defer func() {
		msg.Sig = sig
	}()

	if data, err := msg.Serialize(); err != nil {
		return fmt.Errorf("failed to serialize blockinfo fetch msg: %s", err)
	} else if err := crypto.Verify(*pub, data, sig); err != nil {
		return fmt.Errorf("failed to verify blockinfo fetch msg: %s", err)
	}

	return nil
}

func (msg *BlockInfoFetchMsg) GetBlockNum() uint64 {
	return 0
}

func (msg *BlockInfoFetchMsg) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}

type BlockInfo_ struct {
	BlockNum uint64 `json:"block_num"`
	Proposer uint32 `json:"proposer"`
}

// to fetch committed block from neighbours
type BlockInfoFetchRespMsg struct {
	Blocks []*BlockInfo_ `json:"blocks"`
	Sig    []byte        `json:"sig"`
}

func (msg *BlockInfoFetchRespMsg) Type() MsgType {
	return blockInfoFetchRespMessage
}

func (msg *BlockInfoFetchRespMsg) Verify(pub *crypto.PubKey) error {
	sig := msg.Sig
	msg.Sig = nil

	defer func() {
		msg.Sig = sig
	}()

	if data, err := msg.Serialize(); err != nil {
		return fmt.Errorf("failed to serialize blockinfo resp msg: %s", err)
	} else if err := crypto.Verify(*pub, data, sig); err != nil {
		return fmt.Errorf("failed to verify blockinfo resp msg: %s", err)
	}

	return nil
}

func (msg *BlockInfoFetchRespMsg) GetBlockNum() uint64 {
	return 0
}

func (msg *BlockInfoFetchRespMsg) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}

// block fetch msg is to fetch block which could have not been committed or endorsed
type blockFetchMsg struct {
	BlockNum uint64 `json:"block_num"`
	Sig      []byte `json:"sig"`
}

func (msg *blockFetchMsg) Type() MsgType {
	return blockFetchMessage
}

func (msg *blockFetchMsg) Verify(pub *crypto.PubKey) error {
	sig := msg.Sig
	msg.Sig = nil

	defer func() {
		msg.Sig = sig
	}()

	if data, err := msg.Serialize(); err != nil {
		return fmt.Errorf("failed to serialize blockfetch msg: %s", err)
	} else if err := crypto.Verify(*pub, data, sig); err != nil {
		return fmt.Errorf("failed to verify blockfetch msg: %s", err)
	}

	return nil
}

func (msg *blockFetchMsg) GetBlockNum() uint64 {
	return 0
}

func (msg *blockFetchMsg) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}

type BlockFetchRespMsg struct {
	BlockNumber uint64  `json:"block_number"`
	BlockHash   common.Uint256 `json:"block_hash"`
	BlockData   *Block  `json:"block_data"`
	Sig         []byte  `json:"sig"`
}

func (msg *BlockFetchRespMsg) Type() MsgType {
	return blockFetchRespMessage
}

func (msg *BlockFetchRespMsg) Verify(pub *crypto.PubKey) error {
	sig := msg.Sig
	msg.Sig = nil

	defer func() {
		msg.Sig = sig
	}()

	if data, err := msg.Serialize(); err != nil {
		return fmt.Errorf("failed to serialize blockfetch rsp msg: %s", err)
	} else if err := crypto.Verify(*pub, data, sig); err != nil {
		return fmt.Errorf("failed to verify blockfetch rsp msg: %s", err)
	}

	return nil
}

func (msg *BlockFetchRespMsg) GetBlockNum() uint64 {
	return 0
}

func (msg *BlockFetchRespMsg) Serialize() ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	serialization.WriteUint64(buffer,msg.BlockNumber)
	msg.BlockHash.Serialize(buffer)
	blockbuff,err := msg.BlockData.Serialize()
	if err != nil {
		return nil ,err
	}
	buffer.Write(blockbuff)
	serialization.WriteVarBytes(buffer,msg.Sig)
	return buffer.Bytes(),nil
}

func (msg *BlockFetchRespMsg) Deserialize(data []byte) error {
	buffer := bytes.NewBuffer(data)
	blocknum, err := serialization.ReadUint64(buffer)
	if err != nil {
		return err
	}
	msg.BlockNumber = blocknum
	err = msg.BlockHash.Deserialize(buffer)
	if err != nil {
		return err
	}
	blk := &types.Block{}
	if err := blk.Deserialize(buffer); err != nil {
		return fmt.Errorf("unmarshal block type: %s", err)
	}
	block, err := initVbftBlock(blk)
	if err != nil {
		return fmt.Errorf("init vbft block: %s", err)
	}
	msg.BlockData = block
	sig, err := serialization.ReadVarBytes(buffer)
	if err != nil {
		return err
	}
	msg.Sig = sig
	return nil
}

// proposal fetch msg is to fetch proposal when peer failed to get proposal locally
type proposalFetchMsg struct {
	BlockNum uint64 `json:"block_num"`
	Sig      []byte `json:"sig"`
}

func (msg *proposalFetchMsg) Type() MsgType {
	return proposalFetchMessage
}

func (msg *proposalFetchMsg) Verify(pub *crypto.PubKey) error {
	sig := msg.Sig
	msg.Sig = nil

	defer func() {
		msg.Sig = sig
	}()

	if data, err := msg.Serialize(); err != nil {
		return fmt.Errorf("failed to serialize blockfetch msg: %s", err)
	} else if err := crypto.Verify(*pub, data, sig); err != nil {
		return fmt.Errorf("failed to verify blockfetch msg: %s", err)
	}

	return nil
}

func (msg *proposalFetchMsg) GetBlockNum() uint64 {
	return 0
}

func (msg *proposalFetchMsg) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}
