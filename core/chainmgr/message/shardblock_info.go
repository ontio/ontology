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
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/events/message"
)

//
// Marshal-Helper for transaction
//
type ShardBlockTx struct {
	Tx *types.Transaction
}

func (this *ShardBlockTx) Serialization(sink *common.ZeroCopySink) {
	this.Tx.Serialization(sink)
}

func (this *ShardBlockTx) Deserialization(source *common.ZeroCopySource) error {
	this.Tx = &types.Transaction{}
	return this.Tx.Deserialization(source)
}

//
// ShardBlockInfo contains:
//  .Block: block
//  .ShardTxs: Cross-Shard Tx from the block
//  .Events: shard events generated from the block (only for local block)
//
type ShardBlockInfo struct {
	FromShardID common.ShardID                   `json:"from_shard_id"`
	Height      uint32                           `json:"height"`
	Block       *types.Block                     `json:"block"`
	ShardTxs    map[common.ShardID]*ShardBlockTx `json:"shard_txs"` // indexed by ToShardID
}

func (this *ShardBlockInfo) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteUint64(this.FromShardID.ToUint64())
	sink.WriteUint32(this.Height)
	this.Block.Serialization(sink)
	return nil
}

func (this *ShardBlockInfo) Deserialization(source *common.ZeroCopySource) error {
	fromShard, eof := source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	id, err := common.NewShardID(fromShard)
	if err != nil {
		return fmt.Errorf("deserialization: generate from shard id failed, err: %s", err)
	}
	this.FromShardID = id
	this.Height, eof = source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.Block = &types.Block{}
	if err := this.Block.Deserialization(source); err != nil {
		return fmt.Errorf("deserialization: read header failed, err: %s", err)
	}
	eventNum, eof := source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	for i := uint64(0); i < eventNum; i++ {
		evt := &message.ShardEventState{}
		if err := evt.Deserialization(source); err != nil {
			return fmt.Errorf("deserialization: read event failed, index %d, err: %s", i, err)
		}
	}
	return nil
}
