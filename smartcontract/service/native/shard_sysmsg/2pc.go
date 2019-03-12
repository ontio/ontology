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

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/chainmgr/xshard_state"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/storage"
)

func processXShardPrepareMsg(ctx *native.NativeService, msg *shardstates.CommonShardMsg) error {
	if msg.GetType() != shardstates.EVENT_SHARD_PREPARE {
		return fmt.Errorf("invalid prepare type: %d", msg.GetType())
	}

	// check cached DB
	if !ctx.CacheDB.IsEmptyCache() {
		return fmt.Errorf("non-empty init db when processing prepare msg")
	}

	tx := msg.SourceTxHash
	reqs, err := xshard_state.GetTxRequests(tx)
	if err != nil {
		return fmt.Errorf("get tx requests on prepare %s: %s", common.ToHexString(tx[:]), err)
	}

	prepareOK := true
	for _, req := range reqs {
		result1, err1 := xshard_state.GetTxResponse(tx, req)
		if err1 == xshard_state.ErrNotFound {
			return fmt.Errorf("get tx response %d: %s", req.IdxInTx, err1)
		}
		result2, err2 := ctx.NativeCall(req.GetContract(), req.GetMethod(), req.GetArgs())
		if bytes.Compare(result1, result2.([]byte)) != 0 ||
			err1.Error() != err2.Error() {
			prepareOK = false
			break
		}
	}

	if !prepareOK {
		// we may have aborted the tx, send abort again
		abort := &shardstates.XShardCommitMsg{
			MsgType: shardstates.EVENT_SHARD_ABORT,
		}
		// TODO: clean TX resources
		remoteNotify(ctx, tx, msg.SourceShardID, abort)
		xshard_state.SetTxCommitted(tx, false)
		return fmt.Errorf("failed get tx db when processing prepare msg")
	}

	if err := lockTxContracts(ctx, tx, nil, nil); err != nil {
		return err
	}
	// response prepared
	pareparedMsg := &shardstates.XShardCommitMsg{
		MsgType: shardstates.EVENT_SHARD_PREPARED,
	}
	if _, err := remoteNotify(ctx, tx, msg.SourceShardID, pareparedMsg); err != nil {
		return fmt.Errorf("failed to add prepared response for tx %v: %s", tx, err)
	}

	// save tx rwset and reset ctx.CacheDB
	// TODO: add notification to cached DB
	xshard_state.UpdateTxResult(tx, ctx.CacheDB)
	ctx.CacheDB = storage.NewCacheDB(ctx.CacheDB.GetBackendDB())
	return waitRemoteResponse(ctx)
}

func processXShardPreparedMsg(ctx *native.NativeService, msg *shardstates.CommonShardMsg) error {
	if msg.GetType() != shardstates.EVENT_SHARD_PREPARED {
		return fmt.Errorf("invalid prepared type: %d", msg.GetType())
	}
	if !ctx.CacheDB.IsEmptyCache() {
		return fmt.Errorf("non-empty init db when processing prepared msg")
	}

	tx := msg.SourceTxHash
	txCommits, err := xshard_state.GetTxCommitState(tx)
	if err != nil {
		return fmt.Errorf("get Tx commit state: %s", err)
	}
	if _, present := txCommits[msg.SourceShardID.ToUint64()]; !present {
		return fmt.Errorf("invalid shard ID %d, in tx commit", msg.SourceShardID)
	}
	txCommits[msg.SourceShardID.ToUint64()] = xshard_state.TxPrepared

	if !txCommitReady(txCommits) {
		// wait for prepared from all shards
		return nil
	}

	if _, err := sendCommit(ctx, tx); err != nil {
		return fmt.Errorf("failed to commit tx %v: %s", tx, err)
	}

	// commit cached rwset
	txState, err := xshard_state.GetTxState(tx)
	if err != nil {
		return err
	}
	ctx.CacheDB = txState.DB
	xshard_state.SetTxCommitted(tx, true)
	unlockTxContract(ctx, tx)
	return nil
}

func processXShardCommitMsg(ctx *native.NativeService, msg *shardstates.CommonShardMsg) error {
	if !ctx.CacheDB.IsEmptyCache() {
		return fmt.Errorf("non-empty init db when processing prepared msg")
	}

	// update tx state
	// commit the cached rwset
	tx := msg.SourceTxHash
	txState, err := xshard_state.GetTxState(tx)
	if err != nil {
		return fmt.Errorf("get Tx state: %s", err)
	}
	ctx.CacheDB = txState.DB

	xshard_state.SetTxCommitted(tx, true)
	unlockTxContract(ctx, tx)
	return nil
}

func processXShardAbortMsg(ctx *native.NativeService, msg *shardstates.CommonShardMsg) error {
	if !ctx.CacheDB.IsEmptyCache() {
		return fmt.Errorf("non-empty init db when processing prepared msg")
	}

	// update tx state
	tx := msg.SourceTxHash
	xshard_state.SetTxCommitted(tx, false)
	return nil
}
