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
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/storage"
)

//
// processXShardPrepareMsg
// processing PREPARE-phase of 2pc protocol
// 1. get requests from mq-kvstore
// 2. re-execute all request
// 3. if all results are consist
//     . lock related contracts
//     . store write-set
//     . response with PREPARED
// 4. otherwise, response with ABORT
//
func processXShardPrepareMsg(ctx *native.NativeService, msg *shardstates.CommonShardMsg) error {
	if msg.Msg.Type() != shardstates.EVENT_SHARD_PREPARE {
		return fmt.Errorf("invalid prepare type: %d", msg.GetType())
	}

	// check cached DB
	if !ctx.CacheDB.IsEmptyCache() {
		return fmt.Errorf("non-empty init db when processing prepare msg")
	}

	tx := msg.SourceTxHash
	reqs, err := xshard_state.GetTxRequests(tx)
	if err != nil {
		// no request available, the transaction should have been closed
		return fmt.Errorf("get tx requests on prepare %s: %s", common.ToHexString(tx[:]), err)
	}

	// sorting reqs with IdxInTx
	sort.Slice(reqs, func(i, j int) bool {
		return reqs[i].IdxInTx < reqs[j].IdxInTx
	})
	log.Debugf("process prepare : reqs: %d", len(reqs))

	// 1. re-execute all requests
	// 2. compare new responses with stored responses
	prepareOK := true
	for _, req := range reqs {
		rspMsg := xshard_state.GetTxResponse(tx, req)
		if rspMsg == nil {
			log.Errorf("can not find tx response at index: %d", req.IdxInTx)
			break
		}
		result2, err2 := ctx.NativeCall(req.GetContract(), req.GetMethod(), req.GetArgs())
		isError := false
		if err2 != nil {
			isError = true
		}
		if bytes.Compare(rspMsg.Result, result2.([]byte)) != 0 ||
			rspMsg.Error != isError {
			prepareOK = false
			break
		}
	}
	log.Debugf("process prepare : result: %v", prepareOK)

	if !prepareOK {
		// inconsistent response, abort
		abort := &shardstates.XShardCommitMsg{
			MsgType: shardstates.EVENT_SHARD_ABORT,
		}
		// TODO: clean TX resources
		if err := remoteNotify(ctx, tx, msg.SourceShardID, abort); err != nil {
			log.Errorf("remote notify: %s", err)
		}
		xshard_state.SetTxCommitted(tx, false)
		return fmt.Errorf("failed get tx db when processing prepare msg")
	}

	// response prepared
	pareparedMsg := &shardstates.XShardCommitMsg{
		MsgType: shardstates.EVENT_SHARD_PREPARED,
	}
	//if err := lockTxContracts(ctx, tx, nil, nil); err != nil {
	//	// FIXME
	//	return err
	//}
	if err := remoteNotify(ctx, tx, msg.SourceShardID, pareparedMsg); err != nil {
		return fmt.Errorf("failed to add prepared response for tx %v: %s", tx, err)
	}

	// save tx rwset and reset ctx.CacheDB
	// TODO: add notification to cached DB
	xshard_state.UpdateTxResult(tx, ctx.CacheDB)
	ctx.CacheDB = storage.NewCacheDB(ctx.CacheDB.GetBackendDB())
	return waitRemoteResponse(ctx, tx)
}

//
// processXShardPreparedMsg
// processing PREPARED message from remote shards
// 1. add prepared to shard-state
// 2. if not all remote shards prepared, continue waiting for more prepared
// 3. if all prepared,
//     . commit stored write-set
//     . release all resources
//
func processXShardPreparedMsg(ctx *native.NativeService, msg *shardstates.CommonShardMsg) error {
	if msg.Msg.Type() != shardstates.EVENT_SHARD_PREPARED {
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
	if _, present := txCommits[msg.SourceShardID]; !present {
		return fmt.Errorf("invalid shard ID %d, in tx commit", msg.SourceShardID)
	}
	txCommits[msg.SourceShardID] = xshard_state.TxPrepared

	if !txCommitReady(tx, txCommits) {
		// wait for prepared from all shards
		log.Error("commit not ready")
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
	ctx.CacheDB.SetCache(txState.WriteSet)
	xshard_state.SetTxCommitted(tx, true)
	unlockTxContract(ctx, tx)
	return nil
}

//
// processXShardCommitMsg
// processing COMMIT message
// 1. commit cached writeset
// 2. release all resources
//
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
	ctx.CacheDB.SetCache(txState.WriteSet)
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
