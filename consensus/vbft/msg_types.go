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
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
)

type MsgType uint8

const (
	BlockProposalMessage MsgType = 0
	BlockEndorseMessage  MsgType = 1
	BlockCommitMessage   MsgType = 2

	PeerHeartbeatMessage MsgType = 4

	ProposalFetchMessage MsgType = 7
	BlockSubmitMessage   MsgType = 10
)

func (self MsgType) String() string {
	switch self {
	case BlockProposalMessage:
		return "Proposal"
	case BlockEndorseMessage:
		return "Endorse"
	case BlockCommitMessage:
		return "Commit"
	case PeerHeartbeatMessage:
		return "Heartbeat"
	case ProposalFetchMessage:
		return "ProposalFetch"
	case BlockSubmitMessage:
		return "Submit"
	default:
		panic(fmt.Errorf("unknown msg type: %d", self))
	}
}

type ConsensusMsg interface {
	Type() MsgType
	GetBlockNum() uint32
	Serialize() ([]byte, error)
	Serialization(sink *common.ZeroCopySink)
	Deserialization(source *common.ZeroCopySource) error
}

type KeyProvider interface {
	GetPeerPubKey(peerIdx uint32) keypair.PublicKey
}

type BftConsensusMsg interface {
	ConsensusMsg
	Verify(pubs KeyProvider) error
}

type blockProposalMsgV2 struct {
	Proposer       uint32               `json:"leader"`
	VrfValue       []byte               `json:"vrf_value"`
	VrfProof       []byte               `json:"vrf_proof"`
	BlockHeight    uint32               `json:"block_height"`
	BlockTime      uint32               `json:"block_time"`
	Transactions   []*types.Transaction `json:"transactions"`
	BlockHash      common.Uint256       `json:"block_hash"`
	EmptyBlockHash common.Uint256       `json:"empty_block_hash"`
	Sig            []byte               `json:"sig"`
	EmptySig       []byte               `json:"empty_sig"`
}

func (msg *blockProposalMsgV2) Type() MsgType {
	return BlockProposalMessage
}

func (msg *blockProposalMsgV2) Verify(pubs KeyProvider) error {
	proposer := msg.Proposer
	pub := pubs.GetPeerPubKey(proposer)
	if pub == nil {
		return fmt.Errorf("unknown consensus node, index: %d", proposer)
	}
	sig, err := signature.Deserialize(msg.Sig)
	if err != nil {
		return fmt.Errorf("deserialize block sig: %s", err)
	}
	if !signature.Verify(pub, msg.BlockHash[:], sig) {
		return fmt.Errorf("failed to verify block sig")
	}

	sig, err = signature.Deserialize(msg.EmptySig)
	if err != nil {
		return fmt.Errorf("deserialize empty block sig: %s", err)
	}
	if !signature.Verify(pub, msg.EmptyBlockHash[:], sig) {
		return fmt.Errorf("failed to verify empty block sig")
	}

	return nil
}

func (msg *blockProposalMsgV2) GetBlockNum() uint32 {
	return msg.BlockHeight
}

func (msg *blockProposalMsgV2) Serialize() ([]byte, error) {
	panic("using serialization")
}

func (msg *blockProposalMsgV2) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(msg.Proposer)
	sink.WriteVarBytes(msg.VrfValue)
	sink.WriteVarBytes(msg.VrfProof)
	sink.WriteUint32(msg.BlockHeight)
	sink.WriteUint32(msg.BlockTime)
	sink.WriteVarUint(uint64(len(msg.Transactions)))
	for _, tx := range msg.Transactions {
		tx.Serialization(sink)
	}
	sink.WriteHash(msg.BlockHash)
	sink.WriteHash(msg.EmptyBlockHash)
	sink.WriteVarBytes(msg.Sig)
	sink.WriteVarBytes(msg.EmptySig)
}

func (msg *blockProposalMsgV2) Hash() common.Uint256 {
	sink := common.NewZeroCopySink(nil)
	sink.WriteUint32(msg.Proposer)
	sink.WriteHash(msg.BlockHash)
	sink.WriteHash(msg.EmptyBlockHash)
	return sha256.Sum256(sink.Bytes())
}

func (msg *blockProposalMsgV2) Deserialization(source *common.ZeroCopySource) error {
	reader := source.Reader()
	msg.Proposer = reader.ReadUint32()
	msg.VrfValue = reader.ReadVarBytes()
	msg.VrfProof = reader.ReadVarBytes()
	msg.BlockHeight = reader.ReadUint32()
	msg.BlockTime = reader.ReadUint32()
	txnLen := reader.ReadVarUint()
	if reader.Error() != nil {
		return reader.Error()
	}
	msg.Transactions = make([]*types.Transaction, 0)
	for i := uint64(0); i < txnLen; i++ {
		tx := &types.Transaction{}
		err := tx.Deserialization(source)
		if err != nil {
			return err
		}
		msg.Transactions = append(msg.Transactions, tx)
	}
	msg.BlockHash = reader.ReadHash()
	msg.EmptyBlockHash = reader.ReadHash()
	msg.Sig = reader.ReadVarBytes()
	msg.EmptySig = reader.ReadVarBytes()

	return reader.Error()
}

type blockProposalMsg struct {
	Block *VbftBlock `json:"block"`
}

func (msg *blockProposalMsg) Type() MsgType {
	return BlockProposalMessage
}

func (msg *blockProposalMsg) ToV2() *blockProposalMsgV2 {
	return &blockProposalMsgV2{
		Proposer:       msg.Block.getProposer(),
		VrfValue:       msg.Block.Info.VrfValue,
		VrfProof:       msg.Block.Info.VrfProof,
		BlockHeight:    msg.Block.getBlockNum(),
		BlockTime:      msg.Block.Block.Header.Timestamp,
		Transactions:   msg.Block.Block.Transactions,
		BlockHash:      msg.Block.Block.Hash(),
		EmptyBlockHash: msg.Block.EmptyBlock.Hash(),
		Sig:            msg.Block.Block.Header.SigData[0],
		EmptySig:       msg.Block.EmptyBlock.Header.SigData[0],
	}
}

func (msg *blockProposalMsg) Verify(pubs KeyProvider) error {
	proposer := msg.Block.Info.Proposer
	pub := pubs.GetPeerPubKey(proposer)
	if pub == nil {
		return fmt.Errorf("unknown consensus node, index: %d", proposer)
	}
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
	return msg.Block.Serialize(), nil
}

func (msg *blockProposalMsg) UnmarshalJSON(data []byte) error {
	blk := &VbftBlock{}
	if err := blk.Deserialize(data); err != nil {
		return err
	}

	msg.Block = blk
	return nil
}

func (msg *blockProposalMsg) MarshalJSON() ([]byte, error) {
	return msg.Block.Serialize(), nil
}

func (msg *blockProposalMsg) Serialization(sink *common.ZeroCopySink) {
	msg.ToV2().Serialization(sink)
}

func (msg *blockProposalMsg) Deserialization(source *common.ZeroCopySource) error {
	panic("wrong execution path")
}

type blockEndorseMsg struct {
	Endorser          uint32         `json:"endorser"`
	EndorsedProposer  uint32         `json:"endorsed_proposer"`
	BlockNum          uint32         `json:"block_num"`
	EndorsedBlockHash common.Uint256 `json:"endorsed_block_hash"`
	EndorseForEmpty   bool           `json:"endorse_for_empty"`
	EndorserSig       []byte         `json:"endorser_sig"`
}

func (msg *blockEndorseMsg) Type() MsgType {
	return BlockEndorseMessage
}

func (msg *blockEndorseMsg) Verify(pubs KeyProvider) error {
	pub := pubs.GetPeerPubKey(msg.Endorser)
	if pub == nil {
		return fmt.Errorf("unknown consensus node, index: %d", msg.Endorser)
	}
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

func (msg *blockEndorseMsg) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(msg.Endorser)
	sink.WriteUint32(msg.EndorsedProposer)
	sink.WriteUint32(msg.BlockNum)
	sink.WriteHash(msg.EndorsedBlockHash)
	sink.WriteBool(msg.EndorseForEmpty)
	sink.WriteVarBytes(msg.EndorserSig)
}

func (msg *blockEndorseMsg) Deserialization(source *common.ZeroCopySource) error {
	reader := source.Reader()
	msg.Endorser = reader.ReadUint32()
	msg.EndorsedProposer = reader.ReadUint32()
	msg.BlockNum = reader.ReadUint32()
	msg.EndorsedBlockHash = reader.ReadHash()
	msg.EndorseForEmpty = reader.ReadBool()
	msg.EndorserSig = reader.ReadVarBytes()

	return reader.Error()
}

type blockCommitMsg struct {
	Committer       uint32            `json:"committer"`
	BlockProposer   uint32            `json:"block_proposer"`
	BlockNum        uint32            `json:"block_num"`
	CommitBlockHash common.Uint256    `json:"commit_block_hash"`
	CommitForEmpty  bool              `json:"commit_for_empty"`
	EndorsersSig    map[uint32][]byte `json:"endorsers_sig"`
	CommitterSig    []byte            `json:"committer_sig"`
}

func (msg *blockCommitMsg) Type() MsgType {
	return BlockCommitMessage
}

func (msg *blockCommitMsg) Verify(pubs KeyProvider) error {
	pub := pubs.GetPeerPubKey(msg.Committer)
	if pub == nil {
		return fmt.Errorf("unknown consensus node, index: %d", msg.Committer)
	}
	hash := msg.CommitBlockHash
	sig, err := signature.Deserialize(msg.CommitterSig)
	if err != nil {
		return fmt.Errorf("deserialize block sig: %s", err)
	}
	if !signature.Verify(pub, hash[:], sig) {
		return fmt.Errorf("failed to verify block sig")
	}
	for peerIdx, endorserSig := range msg.EndorsersSig {
		p := pubs.GetPeerPubKey(peerIdx)
		if p == nil {
			return fmt.Errorf("unknown consensus node, index: %d", msg.Committer)
		}
		sig, err := signature.Deserialize(endorserSig)
		if err != nil {
			return fmt.Errorf("deserialize endorserSig sig:%s", err)
		}
		if !signature.Verify(p, hash[:], sig) {
			return fmt.Errorf("failed to verify endorserSig block sig")
		}
	}

	return nil
}

func (msg *blockCommitMsg) GetBlockNum() uint32 {
	return msg.BlockNum
}

func (msg *blockCommitMsg) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}

func (msg *blockCommitMsg) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(msg.Committer)
	sink.WriteUint32(msg.BlockProposer)
	sink.WriteUint32(msg.BlockNum)
	sink.WriteHash(msg.CommitBlockHash)
	sink.WriteBool(msg.CommitForEmpty)
	sink.WriteVarUint(uint64(len(msg.EndorsersSig)))
	keys := make([]uint32, 0, len(msg.EndorsersSig))
	for k := range msg.EndorsersSig {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	for _, k := range keys {
		sink.WriteUint32(k)
		sink.WriteVarBytes(msg.EndorsersSig[k])
	}
	sink.WriteVarBytes(msg.CommitterSig)
}

func (msg *blockCommitMsg) Deserialization(source *common.ZeroCopySource) error {
	reader := source.Reader()
	msg.Committer = reader.ReadUint32()
	msg.BlockProposer = reader.ReadUint32()
	msg.BlockNum = reader.ReadUint32()
	msg.CommitBlockHash = reader.ReadHash()
	msg.CommitForEmpty = reader.ReadBool()
	length := reader.ReadVarUint()
	if reader.Error() != nil {
		return reader.Error()
	}

	msg.EndorsersSig = make(map[uint32][]byte)
	for i := uint64(0); i < length; i++ {
		peerIdx := reader.ReadUint32()
		sig := reader.ReadVarBytes()
		if reader.Error() != nil {
			return reader.Error()
		}
		msg.EndorsersSig[peerIdx] = sig
	}
	msg.CommitterSig = reader.ReadVarBytes()

	return reader.Error()
}

type peerHeartbeatMsg struct {
	CommittedBlockNumber   uint32         `json:"committed_block_number"`
	CommittedBlockHash     common.Uint256 `json:"committed_block_hash"`
	CommittedBlockProposer uint32         `json:"committed_block_leader"`
	Endorsers              [][]byte       `json:"endorsers"`
	EndorsersSig           [][]byte       `json:"endorsers_sig"`
	ChainConfigView        uint32         `json:"chain_config_view"`
}

func (msg *peerHeartbeatMsg) Type() MsgType {
	return PeerHeartbeatMessage
}

func (msg *peerHeartbeatMsg) GetBlockNum() uint32 {
	return msg.CommittedBlockNumber
}

func (msg *peerHeartbeatMsg) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}

func (msg *peerHeartbeatMsg) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(msg.CommittedBlockNumber)
	sink.WriteHash(msg.CommittedBlockHash)
	sink.WriteUint32(msg.CommittedBlockProposer)
	sink.WriteVarUint(uint64(len(msg.Endorsers)))
	for _, endorser := range msg.Endorsers {
		sink.WriteVarBytes(endorser)
	}
	sink.WriteVarUint(uint64(len(msg.EndorsersSig)))
	for _, sig := range msg.EndorsersSig {
		sink.WriteVarBytes(sig)
	}
	sink.WriteUint32(msg.ChainConfigView)
}

func (msg *peerHeartbeatMsg) Deserialization(source *common.ZeroCopySource) error {
	reader := source.Reader()
	msg.CommittedBlockNumber = reader.ReadUint32()
	msg.CommittedBlockHash = reader.ReadHash()
	msg.CommittedBlockProposer = reader.ReadUint32()
	endorserLen := reader.ReadVarUint()
	if reader.Error() != nil {
		return reader.Error()
	}

	for i := uint64(0); i < endorserLen; i++ {
		msg.Endorsers = append(msg.Endorsers, reader.ReadVarBytes())
		if reader.Error() != nil {
			return reader.Error()
		}
	}
	sigLen := reader.ReadVarUint()
	if reader.Error() != nil {
		return reader.Error()
	}
	for i := uint64(0); i < sigLen; i++ {
		msg.EndorsersSig = append(msg.EndorsersSig, reader.ReadVarBytes())
		if reader.Error() != nil {
			return reader.Error()
		}
	}
	msg.ChainConfigView = reader.ReadUint32()

	return reader.Error()
}

// proposal fetch msg is to fetch proposal when peer failed to get proposal locally
type proposalFetchMsg struct {
	ProposerID uint32 `json:"proposer_id"`
	BlockNum   uint32 `json:"block_num"`
}

func (msg *proposalFetchMsg) Type() MsgType {
	return ProposalFetchMessage
}

func (msg *proposalFetchMsg) GetBlockNum() uint32 {
	return 0
}

func (msg *proposalFetchMsg) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}

func (msg *proposalFetchMsg) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(msg.ProposerID)
	sink.WriteUint32(msg.BlockNum)
}

func (msg *proposalFetchMsg) Deserialization(source *common.ZeroCopySource) error {
	reader := source.Reader()
	msg.ProposerID = reader.ReadUint32()
	msg.BlockNum = reader.ReadUint32()
	return reader.Error()
}

type blockSubmitMsg struct {
	Submitter      uint32
	BlockStateRoot common.Uint256 `json:"block_state_root"`
	BlockNum       uint32         `json:"block_num"`
	SubmitMsgSig   []byte         `json:"submit_msg_sig"`
}

func (msg *blockSubmitMsg) Type() MsgType {
	return BlockSubmitMessage
}

func (msg *blockSubmitMsg) Verify(pubs KeyProvider) error {
	pub := pubs.GetPeerPubKey(msg.Submitter)
	if pub == nil {
		return fmt.Errorf("unknown consensus node, index: %d", msg.Submitter)
	}

	hash := msg.BlockStateRoot
	sig, err := signature.Deserialize(msg.SubmitMsgSig)
	if err != nil {
		return fmt.Errorf("deserialize submitmsg sig: %s", err)
	}
	if !signature.Verify(pub, hash[:], sig) {
		return fmt.Errorf("failed to verify submit sig")
	}
	return nil
}

func (msg *blockSubmitMsg) GetBlockNum() uint32 {
	return msg.BlockNum
}

func (msg *blockSubmitMsg) Serialize() ([]byte, error) {
	return json.Marshal(msg)
}

func (msg *blockSubmitMsg) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(msg.Submitter)
	sink.WriteHash(msg.BlockStateRoot)
	sink.WriteUint32(msg.BlockNum)
	sink.WriteVarBytes(msg.SubmitMsgSig)
}

func (msg *blockSubmitMsg) Hash() common.Uint256 {
	sink := common.NewZeroCopySink(nil)
	sink.WriteUint32(msg.Submitter)
	sink.WriteHash(msg.BlockStateRoot)
	sink.WriteUint32(msg.BlockNum)
	return sha256.Sum256(sink.Bytes())
}

func (msg *blockSubmitMsg) Deserialization(source *common.ZeroCopySource) error {
	reader := source.Reader()
	msg.Submitter = reader.ReadUint32()
	msg.BlockStateRoot = reader.ReadHash()
	msg.BlockNum = reader.ReadUint32()
	msg.SubmitMsgSig = reader.ReadVarBytes()

	return reader.Error()
}
