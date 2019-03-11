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

package shardsysmsg

import (
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func remoteTxResponse(ctx *native.NativeService, tx common.Uint256, toShard types.ShardID, result []byte, err string) ([]byte, error) {
	msg := &shardstates.XShardTxRsp{
		FeeUsed: 0,
		Result:  result,
		Error:   err,
	}
	return remoteNotify(ctx, tx, toShard, msg)
}

func sendPrepareRequest(ctx *native.NativeService, tx common.Uint256) ([]byte, error) {
	toShards, err := native.GetTxShards(tx)
	if err != nil {
		return nil, err
	}

	for _, s := range toShards {
		msg := &shardstates.XShardCommitMsg{
			MsgType: shardstates.EVENT_SHARD_PREPARE,
		}
		remoteNotify(ctx, tx, s, msg)
	}

	return nil, nil
}

func sendPreparedResponse(ctx *native.NativeService, shard types.ShardID, tx common.Uint256) ([]byte, error) {
	msg := &shardstates.XShardCommitMsg{
		MsgType: shardstates.EVENT_SHARD_PREPARED,
	}
	return remoteNotify(ctx, tx, shard, msg)
}

func abortTx(ctx *native.NativeService, tx common.Uint256) ([]byte, error) {
	// TODO: clean resources held by tx
	//

	// send abort message to all shards
	toShards, err := native.GetTxShards(tx)
	if err != nil {
		return nil, err
	}

	for _, s := range toShards {
		msg := &shardstates.XShardCommitMsg{
			MsgType: shardstates.EVENT_SHARD_ABORT,
		}
		remoteNotify(ctx, tx, s, msg)
	}

	return nil, nil
}

func sendCommit(ctx *native.NativeService, tx common.Uint256) ([]byte, error) {
	toShards, err := native.GetTxShards(tx)
	if err != nil {
		return nil, err
	}

	for _, s := range toShards {
		msg := &shardstates.XShardCommitMsg{
			MsgType: shardstates.EVENT_SHARD_COMMIT,
		}
		remoteNotify(ctx, tx, s, msg)
	}

	return nil, nil
}

func remoteNotify(ctx *native.NativeService, tx common.Uint256, toShard types.ShardID, msg shardstates.XShardMsg) ([]byte, error) {
	shardReq := &shardstates.CommonShardMsg{
		SourceShardID: ctx.ShardID,
		SourceHeight:  uint64(ctx.Height),
		TargetShardID: toShard,
		SourceTxHash:  tx,
		Msg:           msg,
	}

	// TODO: add evt to queue, update merkle root
	log.Errorf("to send remote notify: from %d to %d", ctx.ShardID, toShard)
	if err := addToShardsInBlock(ctx, toShard); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remote notify, failed to add to-shard to block: %s", err)
	}
	if err := addReqsInBlock(ctx, shardReq); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remote notify, failed to add req to block: %s", err)
	}

	if shardReq.IsTransactional() {
		// TODO: save tx to pending queue, abort the current execution
		if err := native.AddRemoteTxReq(tx, msg); err != nil {
			// TODO: remove added remoteNotify msg
			return utils.BYTE_FALSE, err
		}
		// TODO: stop any further processing
		for ctx.ContextRef.CurrentContext() != ctx.ContextRef.EntryContext() {
			ctx.ContextRef.PopContext()
		}
	}
	return utils.BYTE_TRUE, nil
}
