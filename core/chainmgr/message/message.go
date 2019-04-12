/*
 * Copyright (C) 2019 The ontology Authors
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

package message

import (
	"fmt"
	"github.com/ontio/ontology/common"
	"io"

	"github.com/ontio/ontology/core/types"
)

const (
	SHARD_PROTOCOL_VERSION = 1
)

const (
	HELLO_MSG = iota
	CONFIG_MSG
	BLOCK_REQ_MSG
	BLOCK_RSP_MSG
	PEERINFO_REQ_MSG
	PEERINFO_RSP_MSG

	DISCONNECTED_MSG
)

type RemoteShardMsg interface {
	Type() int
	Serialization(sink *common.ZeroCopySink)
	Deserialization(source *common.ZeroCopySource) error
}

type ShardHelloMsg struct {
	TargetShardID types.ShardID `json:"target_shard_id"`
	SourceShardID types.ShardID `json:"source_shard_id"`
}

func (msg *ShardHelloMsg) Type() int {
	return HELLO_MSG
}

func (this *ShardHelloMsg) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.TargetShardID.ToUint64())
	sink.WriteUint64(this.SourceShardID.ToUint64())
}

func (this *ShardHelloMsg) Deserialization(source *common.ZeroCopySource) error {
	target, eof := source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	targetShardID, err := types.NewShardID(target)
	if err != nil {
		return fmt.Errorf("generate target shard id failed, err: %s", err)
	}
	this.TargetShardID = targetShardID
	sourceId, eof := source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	sourceShardId, err := types.NewShardID(sourceId)
	if err != nil {
		return fmt.Errorf("generate source shard id failed, err: %s", err)
	}
	this.SourceShardID = sourceShardId
	return nil
}

type SibShardInfo struct {
	SeedList []string
	GasPrice uint64
	GasLimit uint64
}

func (this *SibShardInfo) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(uint64(len(this.SeedList)))
	sink.WriteUint64(this.GasPrice)
	sink.WriteUint64(this.GasLimit)
}

func (this *SibShardInfo) Deserialization(source *common.ZeroCopySource) error {
	seedNum, eof := source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.SeedList = make([]string, seedNum)
	for i := uint64(0); i < seedNum; i++ {
		seed, _, irr, eof := source.NextString()
		if irr {
			return fmt.Errorf("read seed failed, index %d, err: %s", i, common.ErrIrregularData)
		}
		if eof {
			return fmt.Errorf("read seed failed, index %d, err: %s", i, io.ErrUnexpectedEOF)
		}
		this.SeedList[i] = seed
	}
	this.GasPrice, eof = source.NextUint64()
	this.GasLimit, eof = source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}

type ShardConfigMsg struct {
	Account   []byte                   `json:"account"`
	SibShards map[uint64]*SibShardInfo `json:"sib_shards"`
	Config    []byte                   `json:"config"`

	// peer pk : ip-addr/port, (query ip-addr from p2p)
	// genesis config
}

func (msg *ShardConfigMsg) Type() int {
	return CONFIG_MSG
}

func (this *ShardConfigMsg) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.Account)
	sink.WriteUint64(uint64(len(this.SibShards)))
	for id, info := range this.SibShards {
		sink.WriteUint64(id)
		info.Serialization(sink)
	}
	sink.WriteVarBytes(this.Config)
}

func (this *ShardConfigMsg) Deserialization(source *common.ZeroCopySource) error {
	var irr, eof bool
	this.Account, _, irr, eof = source.NextVarBytes()
	if irr {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	sibNum, eof := source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	for i := uint64(0); i < sibNum; i++ {
		info := &SibShardInfo{}
		if err := info.Deserialization(source); err != nil {
			return fmt.Errorf("read sib info failed, index %d, err: %s", i, err)
		}
	}
	this.Config, _, irr, eof = source.NextVarBytes()
	if irr {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}

type ShardBlockReqMsg struct {
	ShardID     types.ShardID `json:"shard_id"`
	BlockHeight uint32        `json:"block_num"`
}

func (msg *ShardBlockReqMsg) Type() int {
	return BLOCK_REQ_MSG
}

func (this *ShardBlockReqMsg) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.ShardID.ToUint64())
	sink.WriteUint32(this.BlockHeight)
}

func (this *ShardBlockReqMsg) Deserialization(source *common.ZeroCopySource) error {
	shardId, eof := source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	id, err := types.NewShardID(shardId)
	if err != nil {
		return fmt.Errorf("generate shard id failed, err: %s", err)
	}
	this.ShardID = id
	this.BlockHeight, eof = source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}

type ShardBlockRspMsg struct {
	FromShardID types.ShardID     `json:"from_shard_id"`
	Height      uint32            `json:"height"`
	BlockHeader *ShardBlockHeader `json:"block_header"`
	Txs         []*ShardBlockTx   `json:"txs"`
}

func (msg *ShardBlockRspMsg) Type() int {
	return BLOCK_RSP_MSG
}

func (this *ShardBlockRspMsg) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.FromShardID.ToUint64())
	sink.WriteUint32(this.Height)
	this.BlockHeader.Serialization(sink)
	sink.WriteUint64(uint64(len(this.Txs)))
	for _, tx := range this.Txs {
		tx.Serialization(sink)
	}
}

func (this *ShardBlockRspMsg) Deserialization(source *common.ZeroCopySource) error {
	shardId, eof := source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	id, err := types.NewShardID(shardId)
	if err != nil {
		return fmt.Errorf("generate shard id failed, err: %s", err)
	}
	this.FromShardID = id
	this.Height, eof = source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	if err := this.BlockHeader.Deserialization(source); err != nil {
		return fmt.Errorf("read header failed, err: %s", err)
	}
	txNum, eof := source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.Txs = make([]*ShardBlockTx, txNum)
	for i := uint64(0); i < txNum; i++ {
		tx := &ShardBlockTx{}
		if err := tx.Deserialization(source); err != nil {
			return fmt.Errorf("read tx failed, index %d, err: %s", i, err)
		}
		this.Txs[i] = tx
	}
	return nil
}

type ShardGetPeerInfoReqMsg struct {
	PeerPubKey []byte `json:"peer_pub_key"`
}

func (msg *ShardGetPeerInfoReqMsg) Type() int {
	return PEERINFO_REQ_MSG
}

func (this *ShardGetPeerInfoReqMsg) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.PeerPubKey)
}

func (this *ShardGetPeerInfoReqMsg) Deserialization(source *common.ZeroCopySource) error {
	pubKey, _, irr, eof := source.NextVarBytes()
	if irr {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.PeerPubKey = pubKey
	return nil
}

type ShardGetPeerInfoRspMsg struct {
	PeerPubKey  []byte `json:"peer_pub_key"`
	PeerAddress string `json:"peer_address"`
}

func (msg *ShardGetPeerInfoRspMsg) Type() int {
	return PEERINFO_RSP_MSG
}

func (this *ShardGetPeerInfoRspMsg) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.PeerPubKey)
	sink.WriteString(this.PeerAddress)
}

func (this *ShardGetPeerInfoRspMsg) Deserialization(source *common.ZeroCopySource) error {
	pubKey, _, irr, eof := source.NextVarBytes()
	if irr {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.PeerPubKey = pubKey
	this.PeerAddress, _, irr, eof = source.NextString()
	if irr {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}

type ShardDisconnectedMsg struct {
	Address string `json:"address"`
}

func (msg *ShardDisconnectedMsg) Type() int {
	return DISCONNECTED_MSG
}

func (this *ShardDisconnectedMsg) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.Address)
}

func (this *ShardDisconnectedMsg) Deserialization(source *common.ZeroCopySource) error {
	address, _, irr, eof := source.NextString()
	if irr {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.Address = address
	return nil
}

func EncodeShardMsg(msg RemoteShardMsg) []byte {
	sink := common.NewZeroCopySink(0)
	msg.Serialization(sink)
	return sink.Bytes()
}

func DecodeShardMsg(msgtype int32, msgPayload []byte) (RemoteShardMsg, error) {
	switch msgtype {
	case HELLO_MSG:
		msg := &ShardHelloMsg{}
		if err := msg.Deserialization(common.NewZeroCopySource(msgPayload)); err != nil {
			return nil, fmt.Errorf("decode remote shard msg %d: %s", msgtype, err)
		}
		return msg, nil
	case CONFIG_MSG:
		msg := &ShardConfigMsg{}
		if err := msg.Deserialization(common.NewZeroCopySource(msgPayload)); err != nil {
			return nil, fmt.Errorf("decode remote shard msg %d: %s", msgtype, err)
		}
		return msg, nil
	case BLOCK_REQ_MSG:
		msg := &ShardBlockReqMsg{}
		if err := msg.Deserialization(common.NewZeroCopySource(msgPayload)); err != nil {
			return nil, fmt.Errorf("decode remote shard msg %d: %s", msgtype, err)
		}
		return msg, nil
	case BLOCK_RSP_MSG:
		msg := &ShardBlockRspMsg{}
		if err := msg.Deserialization(common.NewZeroCopySource(msgPayload)); err != nil {
			return nil, fmt.Errorf("decode remote shard msg %d: %s", msgtype, err)
		}
		return msg, nil
	case PEERINFO_REQ_MSG:
		msg := &ShardGetPeerInfoReqMsg{}
		if err := msg.Deserialization(common.NewZeroCopySource(msgPayload)); err != nil {
			return nil, fmt.Errorf("decode remote shard msg %d: %s", msgtype, err)
		}
		return msg, nil
	case PEERINFO_RSP_MSG:
		msg := &ShardGetPeerInfoRspMsg{}
		if err := msg.Deserialization(common.NewZeroCopySource(msgPayload)); err != nil {
			return nil, fmt.Errorf("decode remote shard msg %d: %s", msgtype, err)
		}
		return msg, nil
	}
	return nil, fmt.Errorf("unknown remote shard msg type: %d", msgtype)
}
