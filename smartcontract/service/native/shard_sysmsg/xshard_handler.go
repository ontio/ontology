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
	// check cached DB
	if !ctx.CacheDB.IsEmptyCache() {
		return fmt.Errorf("non-empty init db when processing shard common req")
	}

	// get txState
	result, err := native.GetTxResponse(req.SourceTxHash, req)
	if err == native.ErrNotFound {
		// FIXME: invoke neo contract
		result2, err2 := ctx.NativeCall(req.Msg.GetContract(), req.Msg.GetMethod(), req.Msg.GetArgs())
		if err != nil {
			return err
		}
		if !req.IsTransactional() {
			// notification-call
			return nil
		}

		result = result2.([]byte)
		err = err2
		if err := native.PutTxResponse(req.SourceTxHash, req, result, err2); err != nil {
			return err
		}
	}

	// FIXME: save notification
	rspMsg := &shardstates.XShardTxRsp{
		FeeUsed: 0,
		Result:  result,
		Error:   err.Error(),
	}

	if err == nil {
		log.Infof("process tx result: %v", result)
		native.AddRemoteTxRsp(req.SourceTxHash, req.Msg.GetContract(), ctx.CacheDB, rspMsg)
		// reset ctx.CacheDB
		ctx.CacheDB = storage.NewCacheDB(ctx.CacheDB.GetBackendDB())
	}
	// build TX-RSP
	_, err = remoteNotify(ctx, req.SourceTxHash, req.SourceShardID, rspMsg)
	return err
}

//
//  xrsp : load cached db, invoke PROCESS_XSHARD_RSP_FUNCNAME
//
func processXShardRsp(ctx *native.NativeService, rsp *shardstates.CommonShardMsg) error {
	// get cached DB
	if !ctx.CacheDB.IsEmptyCache() {
		return fmt.Errorf("non-empty init db when processing shard common req")
	}
	// Get caller contract and cachedDB from tx
	txState, err := native.GetTxState(rsp.SourceTxHash)
	if err != nil {
		return fmt.Errorf("failed get tx db when processing shard req")
	}
	ctx.CacheDB = txState.DB

	// FIXME: invoke neo contract
	// re-execute tx
	result, err := ctx.NativeCall(txState.Caller, "", rsp.Msg.GetArgs())
	if err != nil {
		return err
	}

	// START 2PC COMMIT
	tx := rsp.SourceTxHash
	if err := native.VerifyStates(ctx, tx); err != nil {
		if _, err2 := abortTx(ctx, tx); err2 != nil {
			return fmt.Errorf("rwset verify %s, abort tx %v, err: %s", err, tx, err2)
		}
		return err
	}

	if _, err := sendPrepareRequest(ctx, tx); err != nil {
		return err
	}

	// save cached DB to Shard
	// FIXME: add notification to cached DB
	log.Infof("process tx result: %v", result)
	// . reset ctx.CacheDB
	ctx.CacheDB = storage.NewCacheDB(ctx.CacheDB.GetBackendDB())

	return err
}
