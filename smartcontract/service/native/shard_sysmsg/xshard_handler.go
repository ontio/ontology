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
//  notify : process as usual transaction, record fee debt from source shard
//			 normal return
//
func processXShardNotify(ctx *native.NativeService, req *shardstates.CommonShardMsg) error {

	// FIXME: invoke neo contract
	if _, err := ctx.NativeCall(req.Msg.GetContract(), req.Msg.GetMethod(), req.Msg.GetArgs()); err != nil {
		return err
	}

	// TODO: fee settlement
	return nil
}

//
//  xreq : load cached db, process request, save cached, record fee debt from source shard
//			 normal return
//
func processXShardReq(ctx *native.NativeService, req *shardstates.CommonShardMsg) error {
	if req.GetType() != shardstates.EVENT_SHARD_TXREQ {
		return fmt.Errorf("invalid request type: %d", req.GetType())
	}
	// check cached DB
	if !ctx.CacheDB.IsEmptyCache() {
		return fmt.Errorf("non-empty init db when processing shard common req")
	}

	// get txState
	reqMsg, ok := req.Msg.(*shardstates.XShardTxReq)
	if !ok || reqMsg == nil {
		return fmt.Errorf("invalid request message")
	}
	result, err := xshard_state.GetTxResponse(req.SourceTxHash, reqMsg)
	if err == xshard_state.ErrNotFound {
		// FIXME: invoke neo contract
		result2, err2 := ctx.NativeCall(req.Msg.GetContract(), req.Msg.GetMethod(), req.Msg.GetArgs())
		if err != nil {
			return err
		}

		result = result2.([]byte)
		err = err2
	} else if err != nil || result != nil {
		return err
	}

	// FIXME: save notification
	rspMsg := &shardstates.XShardTxRsp{
		IdxInTx: reqMsg.IdxInTx,
		FeeUsed: 0,
		Result:  result,
		Error:   err.Error(),
	}
	if err2 := xshard_state.PutTxRequest(req.SourceTxHash, nil, reqMsg); err2 != nil {
		return err2
	}
	if err2 := xshard_state.PutTxResponse(req.SourceTxHash, rspMsg); err2 != nil {
		return err2
	}

	log.Infof("process tx result: %v", result)
	// reset ctx.CacheDB
	ctx.CacheDB = storage.NewCacheDB(ctx.CacheDB.GetBackendDB())
	// build TX-RSP
	if _, err := remoteNotify(ctx, req.SourceTxHash, req.SourceShardID, rspMsg); err != nil {
		return err
	}
	return waitRemoteResponse(ctx)
}

//
//  xrsp : load cached db, invoke PROCESS_XSHARD_RSP_FUNCNAME
//
func processXShardRsp(ctx *native.NativeService, msg *shardstates.CommonShardMsg) error {
	if msg.GetType() != shardstates.EVENT_SHARD_TXRSP {
		return fmt.Errorf("invalid response type: %d", msg.GetType())
	}
	// get cached DB
	if !ctx.CacheDB.IsEmptyCache() {
		return fmt.Errorf("non-empty init db when processing shard common req")
	}

	rspMsg, ok := msg.Msg.(*shardstates.XShardTxRsp)
	if !ok || rspMsg == nil {
		return fmt.Errorf("invalid response message")
	}

	tx := msg.SourceTxHash
	if err := xshard_state.PutTxResponse(tx, rspMsg); err != nil {
		return fmt.Errorf("failed to store tx response: %s", err)
	}

	// Get caller contract and cachedDB from tx
	txPayload, err := xshard_state.GetTxPayload(tx)
	if err != nil {
		return fmt.Errorf("failed to get tx payload on remote response: %s", err)
	}

	// FIXME: invoke neo contract
	// re-execute tx
	engine, _ := ctx.ContextRef.NewExecuteEngine(txPayload)
	result, resultErr := engine.Invoke()
	if resultErr != nil {
		// Txn failed, abort all transactions
		if _, err2 := abortTx(ctx, tx); err2 != nil {
			return fmt.Errorf("rwset verify %s, abort tx %v, err: %s", err, tx, err2)
		}
		return resultErr
	}

	// START 2PC COMMIT
	if err := xshard_state.VerifyStates(tx); err != nil {
		if _, err2 := abortTx(ctx, tx); err2 != nil {
			return fmt.Errorf("rwset verify %s, abort tx %v, err: %s", err, tx, err2)
		}
		return err
	}

	xshard_state.SetTxPrepared(tx)
	if _, err := sendPrepareRequest(ctx, tx); err != nil {
		return err
	}

	// lock contracts and save cached DB to Shard
	if err := lockTxContracts(ctx, tx, result.([]byte), resultErr); err != nil {
		return err
	}

	// save Tx rwset and reset ctx.CacheDB
	xshard_state.UpdateTxResult(tx, ctx.CacheDB)
	ctx.CacheDB = storage.NewCacheDB(ctx.CacheDB.GetBackendDB())
	return waitRemoteResponse(ctx)
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
	sort.Slice(contracts, func(i, j int) bool {
		return bytes.Compare(contracts[i][:], contracts[j][:]) > 0
	})
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

func waitRemoteResponse(ctx *native.NativeService) error {
	// TODO: stop any further processing
	for ctx.ContextRef.CurrentContext() != ctx.ContextRef.EntryContext() {
		ctx.ContextRef.PopContext()
	}
	return nil
}
