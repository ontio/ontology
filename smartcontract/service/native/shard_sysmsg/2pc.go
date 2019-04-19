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
	"github.com/ontio/ontology/core/xshard_types"
	"sort"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/chainmgr/xshard_state"
	"github.com/ontio/ontology/smartcontract/service/native"
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
func processXShardPrepareMsg(ctx *native.NativeService, txState *xshard_state.TxState, msg *xshard_types.CommonShardMsg) error {
	if msg.Msg.Type() != xshard_types.EVENT_SHARD_PREPARE {
		return fmt.Errorf("invalid prepare type: %d", msg.GetType())
	}

	// check cached DB
	if !ctx.CacheDB.IsEmptyCache() {
		return fmt.Errorf("non-empty init db when processing prepare msg")
	}

	tx := msg.SourceTxHash
	var reqResp []*xshard_state.XShardTxReqResp
	for _, shardReq := range txState.InReqResp {
		reqResp = append(reqResp, shardReq...)
	}

	// sorting reqResp with IdxInTx
	sort.Slice(reqResp, func(i, j int) bool {
		return reqResp[i].Index < reqResp[j].Index
	})
	log.Debugf("process prepare : reqs: %d", len(reqResp))

	txState.NextReqID = 0
	// 1. re-execute all requests
	// 2. compare new responses with stored responses
	prepareOK := true
	for _, val := range reqResp {
		req := val.Req
		rspMsg := val.Resp
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
		abort := &xshard_types.XShardCommitMsg{
			MsgType: xshard_types.EVENT_SHARD_ABORT,
		}
		// TODO: clean TX resources
		if err := remoteSendShardMsg(ctx, tx, msg.SourceShardID, abort); err != nil {
			log.Errorf("remote notify: %s", err)
		}

		txState.State = xshard_state.TxAbort
		return fmt.Errorf("failed get tx db when processing prepare msg")
	}

	// response prepared
	pareparedMsg := &xshard_types.XShardCommitMsg{
		MsgType: xshard_types.EVENT_SHARD_PREPARED,
	}
	//if err := lockTxContracts(ctx, tx, nil, nil); err != nil {
	//	// FIXME
	//	return err
	//}
	if err := remoteSendShardMsg(ctx, tx, msg.SourceShardID, pareparedMsg); err != nil {
		return fmt.Errorf("failed to add prepared response for tx %v: %s", tx, err)
	}

	// save tx rwset and reset ctx.CacheDB
	// TODO: add notification to cached DB
	txState.WriteSet = ctx.CacheDB.GetCache()
	ctx.CacheDB = storage.NewCacheDB(ctx.CacheDB.GetBackendDB())
	return ErrYield
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
func processXShardPreparedMsg(ctx *native.NativeService, txState *xshard_state.TxState, msg *xshard_types.CommonShardMsg) error {
	if msg.Msg.Type() != xshard_types.EVENT_SHARD_PREPARED {
		return fmt.Errorf("invalid prepared type: %d", msg.GetType())
	}
	if !ctx.CacheDB.IsEmptyCache() {
		return fmt.Errorf("non-empty init db when processing prepared msg")
	}

	tx := msg.SourceTxHash
	if _, present := txState.Shards[msg.SourceShardID]; !present {
		return fmt.Errorf("invalid shard ID %d, in tx commit", msg.SourceShardID)
	}
	txState.Shards[msg.SourceShardID] = xshard_state.TxPrepared

	if !txCommitReady(txState) {
		// wait for prepared from all shards
		log.Error("commit not ready")
		return nil
	}

	if _, err := sendCommit(ctx, txState, tx); err != nil {
		return fmt.Errorf("failed to commit tx %v: %s", tx, err)
	}

	// commit cached rwset
	ctx.CacheDB.SetCache(txState.WriteSet)
	txState.State = xshard_state.TxCommit
	//todo:
	unlockTxContract(ctx, tx)
	return nil
}

//
// processXShardCommitMsg
// processing COMMIT message
// 1. commit cached writeset
// 2. release all resources
//
func processXShardCommitMsg(ctx *native.NativeService, txState *xshard_state.TxState, msg *xshard_types.CommonShardMsg) error {
	if !ctx.CacheDB.IsEmptyCache() {
		return fmt.Errorf("non-empty init db when processing prepared msg")
	}

	// update tx state
	// commit the cached rwset
	tx := msg.SourceTxHash
	ctx.CacheDB.SetCache(txState.WriteSet)
	txState.State = xshard_state.TxCommit
	unlockTxContract(ctx, tx)
	return nil
}

func processXShardAbortMsg(ctx *native.NativeService, txState *xshard_state.TxState, msg *xshard_types.CommonShardMsg) error {
	if !ctx.CacheDB.IsEmptyCache() {
		return fmt.Errorf("non-empty init db when processing prepared msg")
	}

	// update tx state
	txState.State = xshard_state.TxAbort
	return nil
}
