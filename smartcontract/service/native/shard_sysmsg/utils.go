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
	"bytes"
	"fmt"
	"sort"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/chainmgr/xshard_state"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/service/neovm"
)

func sendPrepareRequest(ctx *native.NativeService, tx common.Uint256) ([]byte, error) {
	toShards, err := xshard_state.GetTxShards(tx)
	if err != nil {
		return nil, err
	}

	for _, s := range toShards {
		msg := &shardstates.XShardCommitMsg{
			MsgType: shardstates.EVENT_SHARD_PREPARE,
		}
		if err := remoteNotify(ctx, tx, s, msg); err != nil {
			log.Errorf("send prepare to shard %d: %s", s.ToUint64(), err)
		}
	}

	return nil, nil
}

func abortTx(ctx *native.NativeService, tx common.Uint256) ([]byte, error) {
	// TODO: clean resources held by tx
	//

	// send abort message to all shards
	toShards, err := xshard_state.GetTxShards(tx)
	if err != nil {
		return nil, err
	}

	for _, s := range toShards {
		msg := &shardstates.XShardCommitMsg{
			MsgType: shardstates.EVENT_SHARD_ABORT,
		}
		if err := remoteNotify(ctx, tx, s, msg); err != nil {
			log.Errorf("send abort to shard %d: %s", s.ToUint64(), err)
		}
	}

	// FIXME: cleanup resources

	return nil, nil
}

func sendCommit(ctx *native.NativeService, tx common.Uint256) ([]byte, error) {
	toShards, err := xshard_state.GetTxShards(tx)
	if err != nil {
		return nil, err
	}

	for _, s := range toShards {
		msg := &shardstates.XShardCommitMsg{
			MsgType: shardstates.EVENT_SHARD_COMMIT,
		}
		if err := remoteNotify(ctx, tx, s, msg); err != nil {
			log.Errorf("send commit to shard %d: %s", s.ToUint64(), err)
		}
	}

	return nil, nil
}

func remoteNotify(ctx *native.NativeService, tx common.Uint256, toShard types.ShardID, msg shardstates.XShardMsg) error {
	if !ctx.ContextRef.CheckUseGas(neovm.REMOTE_NOTIFY_GAS) {
		return neovm.ERR_GAS_INSUFFICIENT
	}

	shardReq := &shardstates.CommonShardMsg{
		SourceShardID: ctx.ShardID,
		SourceHeight:  uint64(ctx.Height),
		TargetShardID: toShard,
		SourceTxHash:  tx,
		Msg:           msg,
	}

	// TODO: add evt to queue, update merkle root
	log.Debugf("to send remote notify type %d: from %d to %d", msg.Type(), ctx.ShardID, toShard)
	if err := addToShardsInBlock(ctx, toShard); err != nil {
		return fmt.Errorf("remote notify, failed to add to-shard to block: %s", err)
	}
	if err := addReqsInBlock(ctx, shardReq); err != nil {
		return fmt.Errorf("remote notify, failed to add req to block: %s", err)
	}

	return nil
}

func txCommitReady(tx common.Uint256, txState map[uint64]int) bool {
	t, err := xshard_state.GetTxState(tx)
	if err != nil {
		log.Errorf("shard get tx state: %s", err)
		return false
	}
	if t.State != xshard_state.TxPrepared {
		log.Errorf("shard tx state: %d", t.State)
		return false
	}
	for id, state := range txState {
		if state != xshard_state.TxPrepared {
			log.Errorf("shard %d not prepared: %d", id, state)
			return false
		}
	}
	return true
}

func lockTxContracts(ctx *native.NativeService, tx common.Uint256, result []byte, resultErr error) error {
	if result != nil {
		// save result/err to txstate-db
		if err := xshard_state.SetTxResult(tx, result, resultErr); err != nil {
			return fmt.Errorf("save Tx result: %s", err)
		}
	}

	contracts, err := xshard_state.GetTxContracts(tx)
	if err != nil {
		return fmt.Errorf("failed to get contract of tx %v", tx)
	}
	if len(contracts) > 1 {
		sort.Slice(contracts, func(i, j int) bool {
			return bytes.Compare(contracts[i][:], contracts[j][:]) > 0
		})
	}
	for _, c := range contracts {
		if err := xshard_state.LockContract(c); err != nil {
			// TODO: revert all locks
			return fmt.Errorf("failed to lock contract %v for tx %v", c, tx)
		}
	}

	return nil
}

func unlockTxContract(ctx *native.NativeService, tx common.Uint256) error {
	contracts, err := xshard_state.GetTxContracts(tx)
	if err != nil {
		return err
	}

	for _, c := range contracts {
		xshard_state.UnlockContract(c)
	}
	return nil
}

func waitRemoteResponse(ctx *native.NativeService, tx common.Uint256) error {
	// TODO: stop any further processing
	if err := xshard_state.SetTxExecutionPaused(tx); err != nil {
		return fmt.Errorf("set Tx execution paused: %s", err)
	}
	for ctx.ContextRef.CurrentContext() != ctx.ContextRef.EntryContext() {
		ctx.ContextRef.PopContext()
	}
	return nil
}
