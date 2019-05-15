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
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
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
type CrossShardMsgInfo struct {
	ShardMsg *CrossShardMsg `json:"shard_msg"`
	ShardTx  *ShardBlockTx  `json:"shard_txs"`
}
type ShardBlockInfo struct {
	ShardMsg CrossShardMsg `json:"shard_msg"`
	ShardTx  *ShardBlockTx `json:"shard_txs"`
}

func (this *CrossShardMsgInfo) Serialization(sink *common.ZeroCopySink) error {
	this.ShardMsg.Serialization(sink)
	this.ShardTx.Tx.Serialization(sink)
	return nil
}

func (this *CrossShardMsgInfo) Deserialization(source *common.ZeroCopySource) error {
	var err error
	err = this.ShardMsg.Deserialization(source)
	if err != nil {
		return err
	}
	err = this.ShardTx.Deserialization(source)
	return err
}
