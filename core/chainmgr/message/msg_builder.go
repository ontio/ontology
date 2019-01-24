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
	"encoding/json"
	"fmt"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/core/types"
)

func NewShardHelloMsg(localShard, targetShard uint64, sender *actor.PID) (*CrossShardMsg, error) {
	hello := &ShardHelloMsg{
		TargetShardID: targetShard,
		SourceShardID: localShard,
	}
	payload, err := json.Marshal(hello)
	if err != nil {
		return nil, fmt.Errorf("marshal hello msg: %s", err)
	}

	return &CrossShardMsg{
		Version: SHARD_PROTOCOL_VERSION,
		Type:    HELLO_MSG,
		Sender:  sender,
		Data:    payload,
	}, nil
}

func NewShardConfigMsg(accPayload []byte, configPayload []byte, sender *actor.PID) (*CrossShardMsg, error) {
	ack := &ShardConfigMsg{
		Account: accPayload,
		Config:  configPayload,
	}
	payload, err := json.Marshal(ack)
	if err != nil {
		return nil, fmt.Errorf("marshal hello ack msg: %s", err)
	}

	return &CrossShardMsg{
		Version: SHARD_PROTOCOL_VERSION,
		Type:    CONFIG_MSG,
		Sender:  sender,
		Data:    payload,
	}, nil
}

func NewShardBlockRspMsg(fromShardID, toShardID uint64, blkInfo *ShardBlockInfo, sender *actor.PID) (*CrossShardMsg, error) {
	blkRsp := &ShardBlockRspMsg{
		FromShardID: fromShardID,
		Height:      blkInfo.Height,
		BlockHeader: blkInfo.Header,
		Txs:         []*ShardBlockTx{blkInfo.ShardTxs[toShardID]},
	}

	payload, err := json.Marshal(blkRsp)
	if err != nil {
		return nil, fmt.Errorf("marshal shard block rsp msg: %s", err)
	}

	return &CrossShardMsg{
		Version: SHARD_PROTOCOL_VERSION,
		Type:    BLOCK_RSP_MSG,
		Sender:  sender,
		Data:    payload,
	}, nil
}

func NewShardBlockInfo(shardID uint64, blk *types.Block) (*ShardBlockInfo, error) {
	if blk == nil {
		return nil, fmt.Errorf("newShardBlockInfo, nil block")
	}

	blockInfo := &ShardBlockInfo{
		FromShardID: shardID,
		Height:      uint64(blk.Header.Height),
		State:       ShardBlockNew,
		Header: &ShardBlockHeader{
			Header: blk.Header,
		},
	}

	// TODO: add event from block to blockInfo

	return blockInfo, nil
}

func NewShardBlockInfoFromRemote(ShardID uint64, msg *ShardBlockRspMsg) (*ShardBlockInfo, error) {
	if msg == nil {
		return nil, fmt.Errorf("newShardBlockInfo, nil msg")
	}

	blockInfo := &ShardBlockInfo{
		FromShardID: msg.FromShardID,
		Height:      uint64(msg.BlockHeader.Header.Height),
		State:       ShardBlockReceived,
		Header: &ShardBlockHeader{
			Header: msg.BlockHeader.Header,
		},
		ShardTxs: make(map[uint64]*ShardBlockTx),
	}

	if len(msg.Txs) > 0 {
		blockInfo.ShardTxs[ShardID] = msg.Txs[0]
	}
	// TODO: add event from msg to blockInfo

	return blockInfo, nil
}
