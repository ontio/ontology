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

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/chainmgr/xshard_state"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/service/neovm"
	"github.com/ontio/ontology/smartcontract/storage"
)

//
//  processXShardNotify : process as usual transaction, record fee debt from source shard
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
//  processXShardReq : load cached db, process request, save cached, record fee debt from source shard
//			 normal return
//
func processXShardReq(ctx *native.NativeService, req *shardstates.CommonShardMsg) error {
	if req.Msg.Type() != shardstates.EVENT_SHARD_TXREQ {
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

	if rspMsg, _ := xshard_state.GetTxResponse(req.SourceTxHash, reqMsg); rspMsg != nil {
		return nil
	}

	// FIXME: invoke neo contract
	result, err := ctx.NativeCall(req.Msg.GetContract(), req.Msg.GetMethod(), req.Msg.GetArgs())
	log.Debugf("xshard req: method: %s, args: %v, result: %v, err: %s", req.Msg.GetMethod(), req.Msg.GetArgs(), result, err)

	// FIXME: save notification
	// FIXME: feeUsed
	rspMsg := &shardstates.XShardTxRsp{
		IdxInTx: reqMsg.IdxInTx,
		FeeUsed: 0,
		Result:  result.([]byte),
	}
	if err != nil {
		rspMsg.Error = err.Error()
	}
	if err2 := xshard_state.PutTxRequest(req.SourceTxHash, nil, reqMsg); err2 != nil {
		return err2
	}
	if err2 := xshard_state.PutTxResponse(req.SourceTxHash, rspMsg); err2 != nil {
		return err2
	}

	log.Debugf("process xshard request result: %v", result)
	// reset ctx.CacheDB
	ctx.CacheDB.Reset()
	// build TX-RSP
	if err := remoteNotify(ctx, req.SourceTxHash, req.SourceShardID, rspMsg); err != nil {
		return err
	}
	return waitRemoteResponse(ctx, req.SourceTxHash)
}

//
//  processXShardRsp : load cached db, invoke PROCESS_XSHARD_RSP_FUNCNAME
//
func processXShardRsp(ctx *native.NativeService, msg *shardstates.CommonShardMsg) error {
	if msg.Msg.Type() != shardstates.EVENT_SHARD_TXRSP {
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
	if err := xshard_state.AddTxShard(tx, msg.SourceShardID); err != nil {
		return fmt.Errorf("failed to add shardID on response: %s", err)
	}
	if err := xshard_state.PutTxResponse(tx, rspMsg); err != nil {
		return fmt.Errorf("failed to store tx response: %s", err)
	}
	if err := xshard_state.SetTxExecutionContinued(tx); err != nil {
		return fmt.Errorf("failed to continue tx execution: %s", err)
	}

	log.Debugf("process xshard response result: %v", rspMsg.Result)

	// Get caller contract and cachedDB from tx
	txPayload, err := xshard_state.GetTxPayload(tx)
	if err != nil {
		return fmt.Errorf("failed to get tx payload on remote response: %s", err)
	}

	// FIXME: invoke neo contract
	// re-execute tx
	if err := xshard_state.SetNextReqIndex(tx, 0); err != nil {
		return fmt.Errorf("reset next request id: %s", err)
	}
	origTx, err := types.TransactionFromRawBytes(txPayload)
	if err != nil {
		return fmt.Errorf("failed to re-init original tx: %s", err)
	}
	invokeCode := origTx.Payload.(*payload.InvokeCode)
	engine, _ := ctx.ContextRef.NewExecuteEngine(invokeCode.Code)
	neo := engine.(*neovm.NeoVmService)
	neo.Tx = origTx
	_, resultErr := engine.Invoke()
	if resultErr != nil {
		// Txn failed, abort all transactions
		if _, err2 := abortTx(ctx, tx); err2 != nil {
			return fmt.Errorf("rwset verify %s, abort tx %v, err: %s", err, tx, err2)
		}
		return resultErr
	}

	log.Debugf("starting 2pc")

	// START 2PC COMMIT
	if err := xshard_state.VerifyStates(tx); err != nil {
		if _, err2 := abortTx(ctx, tx); err2 != nil {
			return fmt.Errorf("rwset verify %s, abort tx %v, err: %s", err, tx, err2)
		}
		return err
	}

	if err := xshard_state.SetTxPrepared(tx); err != nil {
		log.Errorf("set tx prepared: %s", err)
	}
	if _, err := sendPrepareRequest(ctx, tx); err != nil {
		return err
	}
	log.Debugf("starting 2pc, send prepare done")

	// lock contracts and save cached DB to Shard
	//if err := lockTxContracts(ctx, tx, result.([]byte), resultErr); err != nil {
	//	log.Errorf(" lock tx contract : %s", err)
	//	return err
	//}

	log.Debugf("starting 2pc, contract locked")

	// save Tx rwset and reset ctx.CacheDB
	xshard_state.UpdateTxResult(tx, ctx.CacheDB)
	ctx.CacheDB = storage.NewCacheDB(ctx.CacheDB.GetBackendDB())

	log.Debugf("starting 2pc, to wait response")

	return waitRemoteResponse(ctx, tx)
}
