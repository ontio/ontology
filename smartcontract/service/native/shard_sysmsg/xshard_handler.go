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
	"github.com/ontio/ontology/core/xshard_types"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/neovm"
	"github.com/ontio/ontology/smartcontract/storage"
)

//
//  processXShardNotify : process as usual transaction, record fee debt from source shard
//			 normal return
//
func processXShardNotify(ctx *native.NativeService, req *xshard_types.CommonShardMsg) error {

	// TODO: invoke neo contract
	//builder := neovm.NewParamsBuilder(new(bytes.Buffer))
	//params := make([]interface{}, 0)
	//params = append(params, req.Msg.GetMethod())
	//for _, arg := range req.Msg.GetArgs() {
	//	params = append(params, arg)
	//}
	//err := utils.BuildNeoVMParam(builder, params)
	//if err != nil {
	//	return err
	//}
	//args := append(builder.ToArray(), byte(neovm.APPCALL))
	//contract := req.Msg.GetContract()
	//args = append(args, contract[:]...)
	//
	//engine, _ := ctx.ContextRef.NewExecuteEngine(args)
	//_, err = engine.Invoke()
	//
	//log.Errorf("XSHARD NOTIFY on method: %s", req.Msg.GetMethod())
	//
	//return err

	if _, err := ctx.NativeCall(req.Msg.GetContract(), req.Msg.GetMethod(), req.Msg.GetArgs()); err != nil {
		return err
	}
	return nil
}

//
//  processXShardReq : load cached db, process request, save cached, record fee debt from source shard
//			 normal return
//
func processXShardReq(ctx *native.NativeService, req *xshard_types.CommonShardMsg) error {
	if req.Msg.Type() != xshard_types.EVENT_SHARD_TXREQ {
		return fmt.Errorf("invalid request type: %d", req.GetType())
	}
	// check cached DB
	if !ctx.CacheDB.IsEmptyCache() {
		return fmt.Errorf("non-empty init db when processing shard common req")
	}

	txState := ctx.MainShardTxState

	// get txState
	reqMsg, ok := req.Msg.(*xshard_types.XShardTxReq)
	if !ok || reqMsg == nil {
		return fmt.Errorf("invalid request message")
	}

	shardReq := txState.InReqResp[req.SourceShardID]
	lenReq := uint64(len(shardReq))
	if lenReq > 0 && reqMsg.IdxInTx <= shardReq[lenReq-1].Req.IdxInTx {
		return nil
	}

	// FIXME: invoke neo contract
	result, err := ctx.NativeCall(req.Msg.GetContract(), req.Msg.GetMethod(), req.Msg.GetArgs())

	log.Debugf("xshard req: method: %s, args: %v, result: %v, err: %s", req.Msg.GetMethod(), req.Msg.GetArgs(), result, err)

	// FIXME: save notification
	// FIXME: feeUsed
	rspMsg := &xshard_types.XShardTxRsp{
		IdxInTx: reqMsg.IdxInTx,
		FeeUsed: 0,
		Result:  result.([]byte),
	}
	if err != nil {
		rspMsg.Error = true
	}
	txState.InReqResp[req.SourceShardID] = append(txState.InReqResp[req.SourceShardID],
		&xshard_state.XShardTxReqResp{Req: reqMsg, Resp: rspMsg, Index: txState.TotalInReq})
	txState.TotalInReq += 1

	log.Debugf("process xshard request result: %v", result)
	// reset ctx.CacheDB
	ctx.CacheDB.Reset()
	// build TX-RSP
	if err := remoteSendShardMsg(ctx, req.SourceTxHash, req.SourceShardID, rspMsg); err != nil {
		return err
	}
	txState.SetTxExecutionPaused()
	return nil
}

//
//  processXShardRsp : load cached db, invoke PROCESS_XSHARD_RSP_FUNCNAME
//
func processXShardRsp(ctx *native.NativeService, txState *xshard_state.TxState, msg *xshard_types.CommonShardMsg) error {
	if msg.Msg.Type() != xshard_types.EVENT_SHARD_TXRSP {
		return fmt.Errorf("invalid response type: %d", msg.GetType())
	}
	// get cached DB
	if !ctx.CacheDB.IsEmptyCache() {
		return fmt.Errorf("non-empty init db when processing shard common req")
	}

	rspMsg := msg.Msg.(*xshard_types.XShardTxRsp)
	tx := msg.SourceTxHash
	if txState.PendingReq == nil || txState.PendingReq.Req.IdxInTx != rspMsg.IdxInTx {
		// todo: system error or remote shard error
		return fmt.Errorf("invalid response id: %d", rspMsg.IdxInTx)
	}

	txState.OutReqResp = append(txState.OutReqResp, &xshard_state.XShardTxReqResp{Req: txState.PendingReq.Req, Resp: rspMsg})
	txState.PendingReq = nil
	txState.SetTxExecutionContinued()

	log.Debugf("process xshard response result: %v", rspMsg.Result)

	// Get caller contract and cachedDB from tx
	txPayload := txState.TxPayload
	if txPayload == nil {
		return fmt.Errorf("failed to get tx payload")
	}

	// FIXME: invoke neo contract
	// re-execute tx
	txState.NextReqID = 0
	//todo: add txState in NeoVmService
	xshard_state.PutTxState(txState.TxID, txState)

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
		if _, err2 := abortTx(ctx, txState, tx); err2 != nil {
			return fmt.Errorf("rwset verify %s, abort tx %v, err: %s", err, tx, err2)
		}
		return resultErr
	}

	log.Debugf("starting 2pc")

	// START 2PC COMMIT
	if err := txState.VerifyStates(); err != nil {
		if _, err2 := abortTx(ctx, txState, tx); err2 != nil {
			return fmt.Errorf("rwset verify %s, abort tx %v, err: %s", err, tx, err2)
		}
		return err
	}

	if err := txState.SetTxPrepared(); err != nil {
		log.Errorf("set tx prepared: %s", err)
	}
	if _, err := sendPrepareRequest(ctx, txState, tx); err != nil {
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
	txState.WriteSet = ctx.CacheDB.GetCache()
	ctx.CacheDB = storage.NewCacheDB(ctx.CacheDB.GetBackendDB())

	log.Debugf("starting 2pc, to wait response")

	return nil
}
