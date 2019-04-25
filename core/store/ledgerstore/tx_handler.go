/*
 * Copyright (C) 2018 The ontology Authors
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

package ledgerstore

import (
	"bytes"
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/chainmgr/xshard_state"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/store"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/ontio/ontology/core/types"
	utils2 "github.com/ontio/ontology/core/utils"
	"github.com/ontio/ontology/core/xshard_types"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native/global_params"
	ninit "github.com/ontio/ontology/smartcontract/service/native/init"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/service/neovm"
	"github.com/ontio/ontology/smartcontract/storage"
	ntypes "github.com/ontio/ontology/vm/neovm/types"
	"math"
	"sort"
	"strconv"
)

//HandleDeployTransaction deal with smart contract deploy transaction
func HandleDeployTransaction(store store.LedgerStore, overlay *overlaydb.OverlayDB, cache *storage.CacheDB,
	tx *types.Transaction, header *types.Header, notify *event.ExecuteNotify) error {
	deploy := tx.Payload.(*payload.DeployCode)
	var (
		notifies    []*event.NotifyEventInfo
		gasConsumed uint64
		err         error
	)

	shardID, err := common.NewShardID(header.ShardID)
	if err != nil {
		return err
	}

	if tx.GasPrice != 0 {
		// init smart contract configuration info
		config := &smartcontract.Config{
			ShardID:   shardID,
			Time:      header.Timestamp,
			Height:    header.Height,
			Tx:        tx,
			BlockHash: header.Hash(),
		}
		createGasPrice, ok := neovm.GAS_TABLE.Load(neovm.CONTRACT_CREATE_NAME)
		if !ok {
			overlay.SetError(errors.NewErr("[HandleDeployTransaction] get CONTRACT_CREATE_NAME gas failed"))
			return nil
		}

		uintCodePrice, ok := neovm.GAS_TABLE.Load(neovm.UINT_DEPLOY_CODE_LEN_NAME)
		if !ok {
			overlay.SetError(errors.NewErr("[HandleDeployTransaction] get UINT_DEPLOY_CODE_LEN_NAME gas failed"))
			return nil
		}

		gasLimit := createGasPrice.(uint64) + calcGasByCodeLen(len(deploy.Code), uintCodePrice.(uint64))
		balance, err := isBalanceSufficient(tx.Payer, cache, config, store, gasLimit*tx.GasPrice)
		if err != nil {
			if err := costInvalidGas(tx.Payer, balance, config, cache, store, notify, shardID); err != nil {
				return err
			}
			return err
		}
		if tx.GasLimit < gasLimit {
			if err := costInvalidGas(tx.Payer, tx.GasLimit*tx.GasPrice, config, cache, store, notify, shardID); err != nil {
				return err
			}
			return fmt.Errorf("gasLimit insufficient, need:%d actual:%d", gasLimit, tx.GasLimit)

		}
		gasConsumed = gasLimit * tx.GasPrice
		notifies, err = chargeCostGas(tx.Payer, gasConsumed, config, cache, store, shardID)
		if err != nil {
			return err
		}
	}

	address := deploy.Address()
	log.Infof("deploy contract address:%s", address.ToHexString())
	// store contract message
	dep, err := cache.GetContract(address)
	if err != nil {
		return err
	}
	if dep == nil {
		cache.PutContract(deploy)
	}

	notify.Notify = append(notify.Notify, notifies...)
	notify.GasConsumed = gasConsumed
	notify.State = event.CONTRACT_STATE_SUCCESS
	return nil
}

//HandleChangeMetadataTransaction change contract metadata, only can be invoked by contract owner at root shard
func (self *StateStore) HandleChangeMetadataTransaction(store store.LedgerStore, overlay *overlaydb.OverlayDB,
	cache *storage.CacheDB, tx *types.Transaction, block *types.Block, notify *event.ExecuteNotify) error {
	shardID, err := types.NewShardID(block.Header.ShardID)
	if err != nil {
		return fmt.Errorf("generate shardId failed, %s", err)
	}
	cfg := &smartcontract.Config{
		ShardID:   shardID,
		Time:      block.Header.Timestamp,
		Height:    block.Header.Height,
		Tx:        tx,
		BlockHash: block.Hash(),
	}
	defer func() {
		if err != nil {
			if err := costInvalidGas(tx.Payer, tx.GasLimit*tx.GasPrice, cfg, overlay, store, notify, shardID); err != nil {
				log.Errorf("HandleChangeMetadataTransaction: costInvalidGas failed, err: %s", err)
			} else {
				cache.Commit()
			}
		}
	}()
	if tx.Version < common.VERSION_SUPPORT_SHARD {
		return fmt.Errorf("unsupport tx version")
	}
	newMeta, ok := tx.Payload.(*payload.MetaDataCode)
	if !ok {
		return fmt.Errorf("invalid payload")
	}
	meta, err := cache.GetMetaData(newMeta.Contract)
	if err != nil {
		return fmt.Errorf("read meta data form db failed, err: %s", err)
	}
	if meta == nil {
		return fmt.Errorf("contract cannot contain meta data")
	}
	checkWitness := false
	signerList, err := tx.GetSignatureAddresses()
	if err != nil {
		return fmt.Errorf("cannot get tx signer list, err: %s", err)
	}
	for _, signer := range signerList {
		if signer == meta.Owner {
			checkWitness = true
		}
	}
	if !checkWitness {
		return fmt.Errorf("tx cannot have owner signature")
	}
	// can only change the owner and active or freeze contract
	// cannot change shard info
	meta.Owner = newMeta.Owner
	meta.IsFrozen = newMeta.IsFrozen
	cache.PutMetaData(meta)
	gasConsumed := uint64(0)
	if tx.GasPrice > 0 {
		setMetaGas, ok := neovm.GAS_TABLE.Load(neovm.CONTRACT_SET_META_DATA_NAME)
		if !ok {
			return fmt.Errorf("estimate gas failed")
		}
		wholeGas := setMetaGas.(uint64)
		// init smart contract configuration info
		if tx.GasLimit < wholeGas {
			return fmt.Errorf("gasLimit insufficient, need:%d actual:%d", wholeGas, tx.GasLimit)
		}
		gasConsumed = tx.GasPrice * wholeGas
		notifies, err := chargeCostGas(tx.Payer, gasConsumed, cfg, cache, store, shardID)
		if err != nil {
			return err
		}
		notify.Notify = append(notify.Notify, notifies...)
	}
	notify.GasConsumed = gasConsumed
	notify.State = event.CONTRACT_STATE_SUCCESS
	cache.Commit()
	return nil
}

func resumeTxState(shardTxID xshard_state.ShardTxID, rspMsg *xshard_types.XShardTxRsp) (*xshard_state.TxState, *types.Transaction, error) {
	txState := xshard_state.CreateTxState(shardTxID).Clone()

	if txState.PendingReq == nil || txState.PendingReq.Req.IdxInTx != rspMsg.IdxInTx {
		// todo: system error or remote shard error
		return nil, nil, fmt.Errorf("invalid response id: %d", rspMsg.IdxInTx)
	}

	txState.OutReqResp = append(txState.OutReqResp, &xshard_state.XShardTxReqResp{Req: txState.PendingReq.Req, Resp: rspMsg})
	txState.PendingReq = nil

	txPayload := txState.TxPayload
	if txPayload == nil {
		return nil, nil, fmt.Errorf("failed to get tx payload")
	}

	tx, err := types.TransactionFromRawBytes(txPayload)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to re-init original tx: %s", err)
	}

	return txState, tx, nil
}

/*
	config := &smartcontract.Config{
		ShardID: common.NewShardIDUnchecked(tx.ShardID),
		Time:    uint32(time.Now().Unix()),
		Tx:      tx,
	}

	overlay := NewOverlayDB()
	cache := storage.NewCacheDB(overlay)
	if tx.TxType == types.Invoke {
		invoke := tx.Payload.(*payload.InvokeCode)

		sc := smartcontract.SmartContract{
			Config:           config,
			Store:            nil,
			MainShardTxState: txState,
			CacheDB:          cache,
			Gas:              100000000000000,
			PreExec:          true,
		}

		//start the smart contract executive function
		engine, _ := sc.NewExecuteEngine(invoke.Code)
		res, err := engine.Invoke()

		if err != nil {
			//if err == shardsysmsg.ErrYield {
			//	return txState, err
			//}
			// todo: handle error check
			if txState.PendingReq != nil {
				return txState, nil, shardsysmsg.ErrYield
			}
			return nil, nil, err
		}

		// xshard transaction has completed
		txState.WriteSet = cache.GetCache()

		return txState, res, nil
	}

	panic("unimplemented")
}
*/

func handleShardAbortMsg(msg *xshard_types.CommonShardMsg, store store.LedgerStore, overlay *overlaydb.OverlayDB,
	cache *storage.CacheDB, xshardDB *storage.XShardDB, header *types.Header, notify *event.ExecuteNotify) {
	shardTxID := xshard_state.ShardTxID(string(msg.SourceTxHash[:]))
	txState := xshard_state.CreateTxState(shardTxID).Clone()

	// update tx state
	// commit the cached rwset
	cache.SetCache(txState.WriteSet)
	txState.State = xshard_state.TxAbort
	//unlockTxContract(ctx, tx)

	xshard_state.PutTxState(shardTxID, txState)
}

func handleShardCommitMsg(msg *xshard_types.CommonShardMsg, store store.LedgerStore, overlay *overlaydb.OverlayDB,
	cache *storage.CacheDB, xshardDB *storage.XShardDB, header *types.Header, notify *event.ExecuteNotify) {
	shardTxID := xshard_state.ShardTxID(string(msg.SourceTxHash[:]))
	txState := xshard_state.CreateTxState(shardTxID).Clone()

	// update tx state
	// commit the cached rwset
	cache.SetCache(txState.WriteSet)
	txState.State = xshard_state.TxCommit
	//unlockTxContract(ctx, tx)

	xshard_state.PutTxState(shardTxID, txState)
}

func handleShardPreparedMsg(msg *xshard_types.CommonShardMsg, store store.LedgerStore, overlay *overlaydb.OverlayDB,
	cache *storage.CacheDB, xshardDB *storage.XShardDB, header *types.Header, notify *event.ExecuteNotify) {
	shardTxID := xshard_state.ShardTxID(string(msg.SourceTxHash[:]))
	txState := xshard_state.CreateTxState(shardTxID).Clone()

	if _, present := txState.Shards[msg.SourceShardID]; !present {
		log.Infof("invalid shard ID %d, in tx commit", msg.SourceShardID)
		return
	}
	txState.Shards[msg.SourceShardID] = xshard_state.TxPrepared

	if !txState.IsCommitReady() {
		// wait for prepared from all shards
		log.Info("commit not ready")
		return
	}

	//todo send commit
	//if _, err := sendCommit(ctx, txState, tx); err != nil {
	//	return fmt.Errorf("failed to commit tx %v: %s", tx, err)
	//}

	// commit cached rwset
	cache.SetCache(txState.WriteSet)
	txState.State = xshard_state.TxCommit
	//todo:
	//unlockTxContract(ctx, tx)
	xshard_state.PutTxState(shardTxID, txState)
}

func handleShardPrepareMsg(prepMsg *xshard_types.CommonShardMsg, store store.LedgerStore, overlay *overlaydb.OverlayDB,
	cache *storage.CacheDB, xshardDB *storage.XShardDB, header *types.Header, notify *event.ExecuteNotify) {
	shardTxID := xshard_state.ShardTxID(string(prepMsg.SourceTxHash[:]))
	txState := xshard_state.CreateTxState(shardTxID).Clone()

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

		subTx, err := buildNativeInvokeTx(req.Payer, req.Contract, req.Method, req.Args, req.Fee)
		if err != nil {
			return
		}

		cache.Reset()
		result2, err2 := executeTransaction(store, overlay, cache, txState, subTx, header, notify)

		isError := false
		var res []byte
		if err2 != nil {
			isError = true
		} else {
			res, _ = result2.(*ntypes.ByteArray).GetByteArray() // todo
		}
		if bytes.Compare(rspMsg.Result, res) != 0 || rspMsg.Error != isError {
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
		_ = xshardDB.AddXShardMsgInBlock(header.Height, &xshard_types.CommonShardMsg{
			SourceShardID: prepMsg.TargetShardID,
			SourceHeight:  uint64(header.Height),
			TargetShardID: prepMsg.SourceShardID,
			//SourceTxHash  common.Uint256
			Type: xshard_types.EVENT_SHARD_ABORT,
			//Payload       []byte //todo
			Msg: abort,
		})

		txState.State = xshard_state.TxAbort
		return
	}

	// response prepared
	pareparedMsg := &xshard_types.XShardCommitMsg{
		MsgType: xshard_types.EVENT_SHARD_PREPARED,
	}
	//if err := lockTxContracts(ctx, tx, nil, nil); err != nil {
	//	// FIXME
	//	return err
	//}
	_ = xshardDB.AddXShardMsgInBlock(header.Height, &xshard_types.CommonShardMsg{
		SourceShardID: prepMsg.TargetShardID,
		SourceHeight:  uint64(header.Height),
		TargetShardID: prepMsg.SourceShardID,
		//SourceTxHash  common.Uint256
		Type: xshard_types.EVENT_SHARD_ABORT,
		//Payload       []byte //todo
		Msg: pareparedMsg,
	})

	// save tx rwset and reset ctx.CacheDB
	// TODO: add notification to cached DB
	txState.WriteSet = cache.GetCache()
	cache = storage.NewCacheDB(cache.GetBackendDB())
	return
}

func buildNativeInvokeTx(payer, contract common.Address, method string, msg []byte, fee uint64) (*types.Transaction, error) {
	invokCode, err := utils2.BuildNativeInvokeCode(contract, 0, method, []interface{}{msg})
	if err != nil {
		return nil, err
	}
	invokePayload := &payload.InvokeCode{
		Code: invokCode,
	}
	tx := &types.MutableTransaction{
		Version:  0,
		GasPrice: 0,
		GasLimit: fee,
		TxType:   types.Invoke,
		//Nonce:    header.Timestamp,
		Payer:   payer,
		Payload: invokePayload,
		Sigs:    make([]types.Sig, 0, 0),
	}
	subTx, err := tx.IntoImmutable()
	if err != nil {
		return nil, err
	}

	return subTx, nil
}

func handleShardReqMsg(req *xshard_types.CommonShardMsg, store store.LedgerStore, overlay *overlaydb.OverlayDB,
	cache *storage.CacheDB, xshardDB *storage.XShardDB, header *types.Header, notify *event.ExecuteNotify) {
	msg := req.Msg.(*xshard_types.XShardTxReq)
	shardTxID := xshard_state.ShardTxID(string(req.SourceTxHash[:]))
	txState := xshard_state.CreateTxState(shardTxID).Clone()

	shardReq := txState.InReqResp[req.SourceShardID]
	lenReq := uint64(len(shardReq))
	if lenReq > 0 && msg.IdxInTx <= shardReq[lenReq-1].Req.IdxInTx {
		log.Debugf("handle shard call error: request is stale")
		return
	}

	subTx, err := buildNativeInvokeTx(msg.Payer, msg.Contract, msg.Method, msg.Args, msg.Fee)
	if err != nil {
		return
	}

	cache.Reset()
	result, err := executeTransaction(store, overlay, cache, txState, subTx, header, notify)

	log.Debugf("xshard req: method: %s, args: %v, result: %v, err: %s", req.Msg.GetMethod(), req.Msg.GetArgs(), result, err)

	rspMsg := &xshard_types.XShardTxRsp{
		IdxInTx: msg.IdxInTx,
		FeeUsed: 0,
	}
	if err != nil {
		rspMsg.Error = true // todo pending case
	}
	res, _ := result.(*ntypes.ByteArray).GetByteArray() // todo
	rspMsg.Result = res

	// FIXME: save notification
	// FIXME: feeUsed
	txState.InReqResp[req.SourceShardID] = append(txState.InReqResp[req.SourceShardID],
		&xshard_state.XShardTxReqResp{Req: msg, Resp: rspMsg, Index: txState.TotalInReq})
	txState.TotalInReq += 1

	xshardDB.AddXShardMsgInBlock(header.Height, &xshard_types.CommonShardMsg{
		//SourceShardID :
		//SourceHeight  uint64
		//TargetShardID common.ShardID
		//SourceTxHash  common.Uint256
		//Type          uint64
		//Payload       []byte //todo
		Msg: rspMsg,
	})

	xshard_state.PutTxState(shardTxID, txState)
	log.Debugf("process xshard request result: %v", result)
	//	if _, ok := ctx.SubShardTxState[shardTxID]; ok == false {
	//		ctx.SubShardTxState[shardTxID] = xshard_state.ShardTxInfo{
	//			Index: uint32(len(ctx.SubShardTxState)),
	//			State: xshard_state.CreateTxState(shardTxID).Clone(),
	//		}
	//	}
	//	tmp := ctx.MainShardTxState
	//	ctx.MainShardTxState = ctx.SubShardTxState[shardTxID].State
	//	if err = processXShardReq(ctx, req); err != nil {
	//		log.Errorf("process xshard req: %s", err)
	//	}
	//	ctx.MainShardTxState = tmp
	//	txCompleted = false

}
func handleShardRespMsg(req *xshard_types.CommonShardMsg, store store.LedgerStore, overlay *overlaydb.OverlayDB,
	cache *storage.CacheDB, xshardDB *storage.XShardDB, header *types.Header, notify *event.ExecuteNotify) {
	msg := req.Msg.(*xshard_types.XShardTxRsp)
	shardTxID := xshard_state.ShardTxID(string(req.SourceTxHash[:]))
	txState := xshard_state.CreateTxState(shardTxID).Clone()
	txState, subTx, err := resumeTxState(shardTxID, msg)
	if err != nil {
		return
	}
	// re-execute tx
	txState.NextReqID = 0
	cache.Reset()
	_, err = executeTransaction(store, overlay, cache, txState, subTx, header, notify)

	if err != nil {
		if txState.ExecuteState == xshard_state.ExecYielded {
			blockHeight := header.Height
			for id := range txState.Shards {
				xshardDB.AddToShard(blockHeight, id)
			}
			_ = xshardDB.AddXShardMsgInBlock(blockHeight, &xshard_types.CommonShardMsg{
				SourceTxHash:  txState.PendingReq.SourceTxHash,
				SourceShardID: txState.PendingReq.SourceShardID,
				SourceHeight:  uint64(txState.PendingReq.SourceHeight),
				TargetShardID: txState.PendingReq.TargetShardID,
				Msg:           txState.PendingReq.Req,
			})

			//todo persist txstate
			xshardDB.Commit()
		}

		return
	}

	// tx completed send prepare msg
	for _, s := range txState.GetTxShards() {
		msg := &xshard_types.XShardCommitMsg{
			MsgType: xshard_types.EVENT_SHARD_PREPARE,
		}
		_ = xshardDB.AddXShardMsgInBlock(header.Height, &xshard_types.CommonShardMsg{
			SourceTxHash:  subTx.Hash(),
			SourceShardID: common.NewShardIDUnchecked(header.ShardID),
			SourceHeight:  uint64(header.Height),
			TargetShardID: s,
			Msg:           msg,
		})
	}

	log.Debugf("starting 2pc, send prepare done")

	// lock contracts and save cached DB to Shard
	//if err := lockTxContracts(ctx, tx, result.([]byte), resultErr); err != nil {
	//	log.Errorf(" lock tx contract : %s", err)
	//	return err
	//}

	log.Debugf("starting 2pc, contract locked")

	// save Tx rwset and reset ctx.CacheDB
	txState.WriteSet = cache.GetCache()
	cache = storage.NewCacheDB(cache.GetBackendDB())

	log.Debugf("starting 2pc, to wait response")
	return
	//	shardTxID = xshard_state.ShardTxID(string(req.SourceTxHash[:]))
	//	if _, ok := ctx.SubShardTxState[shardTxID]; ok == false {
	//		ctx.SubShardTxState[shardTxID] = xshard_state.ShardTxInfo{
	//			Index: uint32(len(ctx.SubShardTxState)),
	//			State: xshard_state.CreateTxState(shardTxID).Clone(),
	//		}
	//	}
	//	tmp := ctx.MainShardTxState
	//	ctx.MainShardTxState = ctx.SubShardTxState[shardTxID].State
	//	if err = processXShardRsp(ctx, txState, req); err != nil {
	//		log.Errorf("process xshard rsp: %s", err)
	//	}
	//	ctx.MainShardTxState = tmp
	//	txCompleted = false
}

func executeTransaction(store store.LedgerStore, overlay *overlaydb.OverlayDB,
	cache *storage.CacheDB, txState *xshard_state.TxState,
	tx *types.Transaction, header *types.Header, notify *event.ExecuteNotify) (result interface{}, err error) {
	shardID, err := common.NewShardID(header.ShardID)
	if err != nil {
		return nil, err
	}
	config := &smartcontract.Config{
		ShardID:   shardID,
		Time:      header.Timestamp,
		Height:    header.Height,
		Tx:        tx,
		BlockHash: header.Hash(),
	}

	if tx.TxType == types.Invoke {
		invoke := tx.Payload.(*payload.InvokeCode)

		sc := smartcontract.SmartContract{
			Config:           config,
			Store:            nil,
			MainShardTxState: txState,
			CacheDB:          cache,
			Gas:              100000000000000,
		}

		//start the smart contract executive function
		engine, _ := sc.NewExecuteEngine(invoke.Code)
		res, err := engine.Invoke()
		return res, err
	}

	panic("unimplemented")
}

func HandleShardCallTransaction(store store.LedgerStore, overlay *overlaydb.OverlayDB,
	cache *storage.CacheDB, xshardDB *storage.XShardDB,
	tx *types.Transaction, header *types.Header, notify *event.ExecuteNotify) error {
	shardCall := tx.Payload.(*payload.ShardCall)
	reqs := shardCall.Msgs
	for _, req := range reqs {
		var shardTxID xshard_state.ShardTxID
		switch req.Type {
		case xshard_types.EVENT_SHARD_NOTIFY:
			msg := req.Msg.(*xshard_types.XShardNotify)
			nid := msg.NotifyID
			sink := common.NewZeroCopySink(0)
			sink.WriteBytes(req.SourceTxHash[:]) //todo : use shard tx id
			sink.WriteUint32(nid)
			shardTxID = xshard_state.ShardTxID(string(sink.Bytes()))

			txState := xshard_state.CreateTxState(shardTxID)

			invokCode, err := utils2.BuildNativeInvokeCode(msg.Contract, 0, msg.Method, []interface{}{msg.Args})
			if err != nil {
				return nil
			}
			invokePayload := &payload.InvokeCode{
				Code: invokCode,
			}
			tx := &types.MutableTransaction{
				Version:  0,
				GasPrice: 0,
				GasLimit: msg.Fee,
				TxType:   types.Invoke,
				Nonce:    header.Timestamp,
				Payer:    msg.Payer,
				Payload:  invokePayload,
				Sigs:     make([]types.Sig, 0, 0),
			}
			subTx, err := tx.IntoImmutable()
			if err != nil {
				return nil
			}

			_, err = HandleInvokeTransaction(store, overlay, cache, xshardDB, txState, subTx, header, notify)
			if overlay.Error() != nil {
				return err
			}
			if err != nil {
				log.Debugf("handle shard call tx error %s", err)
			}

		case xshard_types.EVENT_SHARD_TXREQ:
			handleShardReqMsg(req, store, overlay, cache, xshardDB, header, notify)
		case xshard_types.EVENT_SHARD_TXRSP:
			handleShardRespMsg(req, store, overlay, cache, xshardDB, header, notify)
		case xshard_types.EVENT_SHARD_PREPARE:
			handleShardPrepareMsg(req, store, overlay, cache, xshardDB, header, notify)
		case xshard_types.EVENT_SHARD_PREPARED:
			handleShardPreparedMsg(req, store, overlay, cache, xshardDB, header, notify)
		case xshard_types.EVENT_SHARD_COMMIT:
			handleShardCommitMsg(req, store, overlay, cache, xshardDB, header, notify)
		case xshard_types.EVENT_SHARD_ABORT:
			handleShardAbortMsg(req, store, overlay, cache, xshardDB, header, notify)
		}

		//log.Debugf("DONE processing cross shard req %d(height: %d, type: %d)", evt.EventType, evt.FromHeight, req.Type)
		//
		//// transaction should be completed, and be removed from txstate-db
		////todo : buggy
		//if txCompleted {
		//	for _, s := range txState.GetTxShards() {
		//		log.Errorf("TODO: abort transaction %d on shard %d", common.ToHexString(req.SourceTxHash[:]), s)
		//	}
		//}
	}

	return nil
}

//HandleInvokeTransaction deal with smart contract invoke transaction
func HandleInvokeTransaction(store store.LedgerStore, overlay *overlaydb.OverlayDB,
	cache *storage.CacheDB, xshardDB *storage.XShardDB, txState *xshard_state.TxState,
	tx *types.Transaction, header *types.Header, notify *event.ExecuteNotify) (result interface{}, err error) {
	invoke := tx.Payload.(*payload.InvokeCode)
	code := invoke.Code
	sysTransFlag := bytes.Compare(code, ninit.COMMIT_DPOS_BYTES) == 0 || header.Height == 0

	isCharge := !sysTransFlag && tx.GasPrice != 0

	shardID, e := common.NewShardID(header.ShardID)
	if e != nil {
		err = e
		return
	}
	// init smart contract configuration info
	config := &smartcontract.Config{
		ShardID:   shardID,
		Time:      header.Timestamp,
		Height:    header.Height,
		Tx:        tx,
		BlockHash: header.Hash(),
	}

	var (
		costGasLimit      uint64
		costGas           uint64
		oldBalance        uint64
		newBalance        uint64
		codeLenGasLimit   uint64
		availableGasLimit uint64
		minGas            uint64
	)

	availableGasLimit = tx.GasLimit
	if isCharge {
		uintCodeGasPrice, ok := neovm.GAS_TABLE.Load(neovm.UINT_INVOKE_CODE_LEN_NAME)
		if !ok {
			err = errors.NewErr("[HandleInvokeTransaction] get UINT_INVOKE_CODE_LEN_NAME gas failed")
			overlay.SetError(err)
			return
		}

		oldBalance, err = getBalanceFromNative(config, cache, store, tx.Payer)
		if err != nil {
			return
		}

		minGas = neovm.MIN_TRANSACTION_GAS * tx.GasPrice

		if oldBalance < minGas {
			if err = costInvalidGas(tx.Payer, oldBalance, config, cache, store, notify, shardID); err != nil {
				return
			}
			return nil, fmt.Errorf("balance gas: %d less than min gas: %d", oldBalance, minGas)
		}

		codeLenGasLimit = calcGasByCodeLen(len(invoke.Code), uintCodeGasPrice.(uint64))

		if oldBalance < codeLenGasLimit*tx.GasPrice {
			if err = costInvalidGas(tx.Payer, oldBalance, config, cache, store, notify, shardID); err != nil {
				return
			}
			return nil, fmt.Errorf("balance gas insufficient: balance:%d < code length need gas:%d", oldBalance, codeLenGasLimit*tx.GasPrice)
		}

		if tx.GasLimit < codeLenGasLimit {
			if err = costInvalidGas(tx.Payer, tx.GasLimit*tx.GasPrice, config, cache, store, notify, shardID); err != nil {
				return
			}
			return nil, fmt.Errorf("invoke transaction gasLimit insufficient: need%d actual:%d", tx.GasLimit, codeLenGasLimit)
		}

		maxAvaGasLimit := oldBalance / tx.GasPrice
		if availableGasLimit > maxAvaGasLimit {
			availableGasLimit = maxAvaGasLimit
		}
	}

	//init smart contract info
	sc := smartcontract.SmartContract{
		Config:           config,
		CacheDB:          cache,
		MainShardTxState: txState,
		Store:            store,
		Gas:              availableGasLimit - codeLenGasLimit,
	}

	//start the smart contract executive function
	engine, _ := sc.NewExecuteEngine(invoke.Code)

	result, err = engine.Invoke()

	costGasLimit = availableGasLimit - sc.Gas
	if costGasLimit < neovm.MIN_TRANSACTION_GAS {
		costGasLimit = neovm.MIN_TRANSACTION_GAS
	}

	costGas = costGasLimit * tx.GasPrice
	if err != nil {
		if isCharge {
			if err = costInvalidGas(tx.Payer, costGas, config, cache, store, notify, shardID); err != nil {
				return
			}
		}

		if txState.ExecuteState == xshard_state.ExecYielded {
			blockHeight := header.Height
			for id := range txState.Shards {
				xshardDB.AddToShard(blockHeight, id)
			}
			_ = xshardDB.AddXShardMsgInBlock(blockHeight, &xshard_types.CommonShardMsg{
				SourceTxHash:  txState.PendingReq.SourceTxHash,
				SourceShardID: txState.PendingReq.SourceShardID,
				SourceHeight:  uint64(txState.PendingReq.SourceHeight),
				TargetShardID: txState.PendingReq.TargetShardID,
				Msg:           txState.PendingReq.Req,
			})

			//todo persist txstate
			xshardDB.Commit()
		} // todo prepared
		return nil, err
	}

	var notifies []*event.NotifyEventInfo
	if isCharge {
		newBalance, err = getBalanceFromNative(config, cache, store, tx.Payer)
		if err != nil {
			return nil, err
		}

		if newBalance < costGas {
			if err = costInvalidGas(tx.Payer, costGas, config, cache, store, notify, shardID); err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("gas insufficient, balance:%d < costGas:%d", newBalance, costGas)
		}

		notifies, err = chargeCostGas(tx.Payer, costGas, config, sc.CacheDB, store, shardID)
		if err != nil {
			return nil, err
		}
	}
	notify.Notify = append(notify.Notify, sc.Notifications...)
	notify.Notify = append(notify.Notify, notifies...)
	notify.GasConsumed = costGas
	notify.State = event.CONTRACT_STATE_SUCCESS
	sc.CacheDB.Commit()
	//todo : add notify to xshardDB
	xshardDB.Commit()
	return result, nil
}

func SaveNotify(eventStore scommon.EventStore, txHash common.Uint256, notify *event.ExecuteNotify) error {
	if !config.DefConfig.Common.EnableEventLog {
		return nil
	}
	if err := eventStore.SaveEventNotifyByTx(txHash, notify); err != nil {
		return fmt.Errorf("SaveEventNotifyByTx error %s", err)
	}
	event.PushSmartCodeEvent(txHash, 0, event.EVENT_NOTIFY, notify)
	return nil
}

func genNativeTransferCode(from, to common.Address, value uint64) []byte {
	transfer := ont.Transfers{States: []ont.State{{From: from, To: to, Value: value}}}
	return common.SerializeToBytes(&transfer)
}

// check whether payer ong balance sufficient
func isBalanceSufficient(payer common.Address, cache *storage.CacheDB, config *smartcontract.Config, store store.LedgerStore, gas uint64) (uint64, error) {
	balance, err := getBalanceFromNative(config, cache, store, payer)
	if err != nil {
		return 0, err
	}
	if balance < gas {
		return 0, fmt.Errorf("payer gas insufficient, need %d , only have %d", gas, balance)
	}
	return balance, nil
}

func chargeCostGas(payer common.Address, gas uint64, config *smartcontract.Config,
	cache *storage.CacheDB, store store.LedgerStore, shardID common.ShardID) ([]*event.NotifyEventInfo, error) {
	contractAddr := utils.GovernanceContractAddress
	if !shardID.IsRootShard() {
		contractAddr = utils.ShardGasMgmtContractAddress
	}
	params := genNativeTransferCode(payer, contractAddr, gas)

	sc := smartcontract.SmartContract{
		Config:  config,
		CacheDB: cache,
		Store:   store,
		Gas:     math.MaxUint64,
	}

	service, _ := sc.NewNativeService()
	_, err := service.NativeCall(utils.OngContractAddress, "transfer", params)
	if err != nil {
		return nil, err
	}
	return sc.Notifications, nil
}

func refreshGlobalParam(config *smartcontract.Config, cache *storage.CacheDB, store store.LedgerStore) error {
	bf := new(bytes.Buffer)
	if err := utils.WriteVarUint(bf, uint64(len(neovm.GAS_TABLE_KEYS))); err != nil {
		return fmt.Errorf("write gas_table_keys length error:%s", err)
	}
	for _, value := range neovm.GAS_TABLE_KEYS {
		if err := serialization.WriteString(bf, value); err != nil {
			return fmt.Errorf("serialize param name error:%s", value)
		}
	}

	sc := smartcontract.SmartContract{
		Config:  config,
		CacheDB: cache,
		Store:   store,
		Gas:     math.MaxUint64,
	}

	service, _ := sc.NewNativeService()
	result, err := service.NativeCall(utils.ParamContractAddress, "getGlobalParam", bf.Bytes())
	if err != nil {
		return err
	}
	params := new(global_params.Params)
	if err := params.Deserialize(bytes.NewBuffer(result.([]byte))); err != nil {
		return fmt.Errorf("deserialize global params error:%s", err)
	}
	neovm.GAS_TABLE.Range(func(key, value interface{}) bool {
		n, ps := params.GetParam(key.(string))
		if n != -1 && ps.Value != "" {
			pu, err := strconv.ParseUint(ps.Value, 10, 64)
			if err != nil {
				log.Errorf("[refreshGlobalParam] failed to parse uint %v\n", ps.Value)
			} else {
				neovm.GAS_TABLE.Store(key, pu)

			}
		}
		return true
	})
	return nil
}

func getBalanceFromNative(config *smartcontract.Config, cache *storage.CacheDB, store store.LedgerStore, address common.Address) (uint64, error) {
	bf := new(bytes.Buffer)
	if err := utils.WriteAddress(bf, address); err != nil {
		return 0, err
	}
	sc := smartcontract.SmartContract{
		Config:  config,
		CacheDB: cache,
		Store:   store,
		Gas:     math.MaxUint64,
	}

	service, _ := sc.NewNativeService()
	result, err := service.NativeCall(utils.OngContractAddress, ont.BALANCEOF_NAME, bf.Bytes())
	if err != nil {
		return 0, err
	}
	return ntypes.BigIntFromBytes(result.([]byte)).Uint64(), nil
}

func costInvalidGas(address common.Address, gas uint64, config *smartcontract.Config, cache *storage.CacheDB,
	store store.LedgerStore, notify *event.ExecuteNotify, shardID common.ShardID) error {
	cache.Reset()
	notifies, err := chargeCostGas(address, gas, config, cache, store, shardID)
	if err != nil {
		return err
	}
	notify.GasConsumed = gas
	notify.Notify = append(notify.Notify, notifies...)
	return nil
}

func calcGasByCodeLen(codeLen int, codeGas uint64) uint64 {
	return uint64(codeLen/neovm.PER_UNIT_CODE_LEN) * codeGas
}
