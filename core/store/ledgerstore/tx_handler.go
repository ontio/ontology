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
	"github.com/ontio/ontology/core/chainmgr/xshard_state"
	utils2 "github.com/ontio/ontology/core/utils"
	"github.com/ontio/ontology/core/xshard_types"
	"math"
	"strconv"
	"time"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/store"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/ontio/ontology/core/types"
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
			if err := costInvalidGas(tx.Payer, balance, config, overlay, store, notify, shardID); err != nil {
				return err
			}
			return err
		}
		if tx.GasLimit < gasLimit {
			if err := costInvalidGas(tx.Payer, tx.GasLimit*tx.GasPrice, config, overlay, store, notify, shardID); err != nil {
				return err
			}
			return fmt.Errorf("gasLimit insufficient, need:%d actual:%d", gasLimit, tx.GasLimit)

		}
		gasConsumed = gasLimit * tx.GasPrice
		notifies, err = chargeCostGas(tx.Payer, gasConsumed, config, cache, store, shardID)
		if err != nil {
			return err
		}
		cache.Commit()
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
	cache.Commit()

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

func HandleShardCallTransaction(store store.LedgerStore, overlay *overlaydb.OverlayDB,
	cache *storage.CacheDB, xshardDB *storage.XShardDB,
	tx *types.Transaction, header *types.Header, notify *event.ExecuteNotify) error {
	shardCall := tx.Payload.(*payload.ShardCall)
	reqs := shardCall.Msgs
	for _, req := range reqs {
		var txState *xshard_state.TxState
		var shardTxID xshard_state.ShardTxID
		txCompleted := false
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

			err = HandleInvokeTransaction(store, overlay, cache, xshardDB, txState, subTx, header, notify)
			if overlay.Error() != nil {
				return err
			}
			if err != nil {
				log.Debugf("handle shard call tx error %s", err)
			}

		case xshard_types.EVENT_SHARD_TXREQ:
			msg := req.Msg.(*xshard_types.XShardTxReq)
			shardTxID = xshard_state.ShardTxID(string(req.SourceTxHash[:]))
			txState := xshard_state.CreateTxState(shardTxID).Clone()

			shardReq := txState.InReqResp[req.SourceShardID]
			lenReq := uint64(len(shardReq))
			if lenReq > 0 && msg.IdxInTx <= shardReq[lenReq-1].Req.IdxInTx {
				log.Debugf("handle shard call error: request is stale")
				continue
			}
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

			err = HandleInvokeTransaction(store, overlay, cache, xshardDB, txState, subTx, header, notify)

			// FIXME: invoke neo contract
			result, err := ctx.NativeCall(req.Msg.GetContract(), req.Msg.GetMethod(), req.Msg.GetArgs())

			log.Debugf("xshard req: method: %s, args: %v, result: %v, err: %s", req.Msg.GetMethod(), req.Msg.GetArgs(), result, err)

			// FIXME: save notification
			// FIXME: feeUsed
			rspMsg := &xshard_types.XShardTxRsp{
				IdxInTx: msg.IdxInTx,
				FeeUsed: 0,
				Result:  result.([]byte),
			}
			if err != nil {
				rspMsg.Error = true
			}
			txState.InReqResp[req.SourceShardID] = append(txState.InReqResp[req.SourceShardID],
				&xshard_state.XShardTxReqResp{Req: msg, Resp: rspMsg, Index: txState.TotalInReq})
			txState.TotalInReq += 1

			log.Debugf("process xshard request result: %v", result)
			// reset ctx.CacheDB
			ctx.CacheDB.Reset()
			// build TX-RSP
			if err := remoteSendShardMsg(ctx, req.SourceTxHash, req.SourceShardID, rspMsg); err != nil {
				return err
			}
			txState.SetTxExecutionPaused()
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
			//case xshard_types.EVENT_SHARD_TXRSP:
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
			//case xshard_types.EVENT_SHARD_PREPARE:
			//	shardTxID = xshard_state.ShardTxID(string(req.SourceTxHash[:]))
			//	if _, ok := ctx.SubShardTxState[shardTxID]; ok == false {
			//		ctx.SubShardTxState[shardTxID] = xshard_state.ShardTxInfo{
			//			Index: uint32(len(ctx.SubShardTxState)),
			//			State: xshard_state.CreateTxState(shardTxID).Clone(),
			//		}
			//	}
			//	tmp := ctx.MainShardTxState
			//	ctx.MainShardTxState = ctx.SubShardTxState[shardTxID].State
			//	if err = processXShardPrepareMsg(ctx, txState, req); err != nil {
			//		log.Errorf("process xshard prepare: %s", err)
			//	}
			//	ctx.MainShardTxState = tmp
			//	txCompleted = false
			//case xshard_types.EVENT_SHARD_PREPARED:
			//	shardTxID = xshard_state.ShardTxID(string(req.SourceTxHash[:]))
			//	txState = xshard_state.CreateTxState(shardTxID).Clone()
			//	if err = processXShardPreparedMsg(ctx, txState, req); err != nil {
			//		log.Errorf("process xshard prepared: %s", err)
			//	}
			//	// FIXME: completed with all-shards-prepared
			//	txCompleted = true
			//case xshard_types.EVENT_SHARD_COMMIT:
			//	shardTxID = xshard_state.ShardTxID(string(req.SourceTxHash[:]))
			//	txState = xshard_state.CreateTxState(shardTxID).Clone()
			//	if err = processXShardCommitMsg(ctx, txState, req); err != nil {
			//		log.Errorf("process xshard commit: %s", err)
			//	}
			//	txCompleted = true
			//case xshard_types.EVENT_SHARD_ABORT:
			//	shardTxID = xshard_state.ShardTxID(string(req.SourceTxHash[:]))
			//	txState = xshard_state.CreateTxState(shardTxID).Clone()
			//	if err = processXShardAbortMsg(ctx, txState, req); err != nil {
			//		log.Errorf("process xshard abort: %s", err)
			//	}
			//	txCompleted = true
		}

		if err != nil && err != ErrYield {
			// system error: abort the whole transaction process
			return utils.BYTE_FALSE, err
		}

		log.Debugf("DONE processing cross shard req %d(height: %d, type: %d)", evt.EventType, evt.FromHeight, req.Type)

		if txState != nil {
			xshard_state.PutTxState(shardTxID, txState)
		}
		// transaction should be completed, and be removed from txstate-db
		//todo : buggy
		if txCompleted {
			for _, s := range txState.GetTxShards() {
				log.Errorf("TODO: abort transaction %d on shard %d", common.ToHexString(req.SourceTxHash[:]), s)
			}
		}
	}

	sysTransFlag := bytes.Compare(code, ninit.COMMIT_DPOS_BYTES) == 0 || header.Height == 0

	isCharge := !sysTransFlag && tx.GasPrice != 0

	shardID, err := common.NewShardID(header.ShardID)
	if err != nil {
		return err
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
			overlay.SetError(errors.NewErr("[HandleInvokeTransaction] get UINT_INVOKE_CODE_LEN_NAME gas failed"))
			return nil
		}

		oldBalance, err = getBalanceFromNative(config, cache, store, tx.Payer)
		if err != nil {
			return err
		}

		minGas = neovm.MIN_TRANSACTION_GAS * tx.GasPrice

		if oldBalance < minGas {
			if err := costInvalidGas(tx.Payer, oldBalance, config, overlay, store, notify, shardID); err != nil {
				return err
			}
			return fmt.Errorf("balance gas: %d less than min gas: %d", oldBalance, minGas)
		}

		codeLenGasLimit = calcGasByCodeLen(len(invoke.Code), uintCodeGasPrice.(uint64))

		if oldBalance < codeLenGasLimit*tx.GasPrice {
			if err := costInvalidGas(tx.Payer, oldBalance, config, overlay, store, notify, shardID); err != nil {
				return err
			}
			return fmt.Errorf("balance gas insufficient: balance:%d < code length need gas:%d", oldBalance, codeLenGasLimit*tx.GasPrice)
		}

		if tx.GasLimit < codeLenGasLimit {
			if err := costInvalidGas(tx.Payer, tx.GasLimit*tx.GasPrice, config, overlay, store, notify, shardID); err != nil {
				return err
			}
			return fmt.Errorf("invoke transaction gasLimit insufficient: need%d actual:%d", tx.GasLimit, codeLenGasLimit)
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

	_, err = engine.Invoke()

	costGasLimit = availableGasLimit - sc.Gas
	if costGasLimit < neovm.MIN_TRANSACTION_GAS {
		costGasLimit = neovm.MIN_TRANSACTION_GAS
	}

	costGas = costGasLimit * tx.GasPrice
	if err != nil {
		if isCharge {
			if err := costInvalidGas(tx.Payer, costGas, config, overlay, store, notify, shardID); err != nil {
				return err
			}
		}

		if txState.ExecuteState == xshard_state.ExecYielded {
			blockHeight := header.Height
			for id := range txState.Shards {
				xshardDB.AddToShard(blockHeight, id)
			}
			_ = xshardDB.AddXShardReqsInBlock(blockHeight, &xshard_types.CommonShardMsg{
				SourceTxHash:  txState.PendingReq.SourceTxHash,
				SourceShardID: txState.PendingReq.SourceShardID,
				SourceHeight:  uint64(txState.PendingReq.SourceHeight),
				TargetShardID: txState.PendingReq.TargetShardID,
				Msg:           txState.PendingReq.Req,
			})

			//todo persist txstate
			xshardDB.Commit()
		} // todo prepared
		return err
	}

	var notifies []*event.NotifyEventInfo
	if isCharge {
		newBalance, err = getBalanceFromNative(config, cache, store, tx.Payer)
		if err != nil {
			return err
		}

		if newBalance < costGas {
			if err := costInvalidGas(tx.Payer, costGas, config, overlay, store, notify, shardID); err != nil {
				return err
			}
			return fmt.Errorf("gas insufficient, balance:%d < costGas:%d", newBalance, costGas)
		}

		notifies, err = chargeCostGas(tx.Payer, costGas, config, sc.CacheDB, store, shardID)
		if err != nil {
			return err
		}
	}
	notify.Notify = append(notify.Notify, sc.Notifications...)
	notify.Notify = append(notify.Notify, notifies...)
	notify.GasConsumed = costGas
	notify.State = event.CONTRACT_STATE_SUCCESS
	sc.CacheDB.Commit()
	//todo : add notify to xshardDB
	xshardDB.Commit()
	return nil
}

//HandleInvokeTransaction deal with smart contract invoke transaction
func HandleInvokeTransaction(store store.LedgerStore, overlay *overlaydb.OverlayDB,
	cache *storage.CacheDB, xshardDB *storage.XShardDB, txState *xshard_state.TxState,
	tx *types.Transaction, header *types.Header, notify *event.ExecuteNotify) error {
	invoke := tx.Payload.(*payload.InvokeCode)
	code := invoke.Code
	sysTransFlag := bytes.Compare(code, ninit.COMMIT_DPOS_BYTES) == 0 || header.Height == 0

	isCharge := !sysTransFlag && tx.GasPrice != 0

	shardID, err := common.NewShardID(header.ShardID)
	if err != nil {
		return err
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
			overlay.SetError(errors.NewErr("[HandleInvokeTransaction] get UINT_INVOKE_CODE_LEN_NAME gas failed"))
			return nil
		}

		oldBalance, err = getBalanceFromNative(config, cache, store, tx.Payer)
		if err != nil {
			return err
		}

		minGas = neovm.MIN_TRANSACTION_GAS * tx.GasPrice

		if oldBalance < minGas {
			if err := costInvalidGas(tx.Payer, oldBalance, config, overlay, store, notify, shardID); err != nil {
				return err
			}
			return fmt.Errorf("balance gas: %d less than min gas: %d", oldBalance, minGas)
		}

		codeLenGasLimit = calcGasByCodeLen(len(invoke.Code), uintCodeGasPrice.(uint64))

		if oldBalance < codeLenGasLimit*tx.GasPrice {
			if err := costInvalidGas(tx.Payer, oldBalance, config, overlay, store, notify, shardID); err != nil {
				return err
			}
			return fmt.Errorf("balance gas insufficient: balance:%d < code length need gas:%d", oldBalance, codeLenGasLimit*tx.GasPrice)
		}

		if tx.GasLimit < codeLenGasLimit {
			if err := costInvalidGas(tx.Payer, tx.GasLimit*tx.GasPrice, config, overlay, store, notify, shardID); err != nil {
				return err
			}
			return fmt.Errorf("invoke transaction gasLimit insufficient: need%d actual:%d", tx.GasLimit, codeLenGasLimit)
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

	_, err = engine.Invoke()

	costGasLimit = availableGasLimit - sc.Gas
	if costGasLimit < neovm.MIN_TRANSACTION_GAS {
		costGasLimit = neovm.MIN_TRANSACTION_GAS
	}

	costGas = costGasLimit * tx.GasPrice
	if err != nil {
		if isCharge {
			if err := costInvalidGas(tx.Payer, costGas, config, overlay, store, notify, shardID); err != nil {
				return err
			}
		}

		if txState.ExecuteState == xshard_state.ExecYielded {
			blockHeight := header.Height
			for id := range txState.Shards {
				xshardDB.AddToShard(blockHeight, id)
			}
			_ = xshardDB.AddXShardReqsInBlock(blockHeight, &xshard_types.CommonShardMsg{
				SourceTxHash:  txState.PendingReq.SourceTxHash,
				SourceShardID: txState.PendingReq.SourceShardID,
				SourceHeight:  uint64(txState.PendingReq.SourceHeight),
				TargetShardID: txState.PendingReq.TargetShardID,
				Msg:           txState.PendingReq.Req,
			})

			//todo persist txstate
			xshardDB.Commit()
		} // todo prepared
		return err
	}

	var notifies []*event.NotifyEventInfo
	if isCharge {
		newBalance, err = getBalanceFromNative(config, cache, store, tx.Payer)
		if err != nil {
			return err
		}

		if newBalance < costGas {
			if err := costInvalidGas(tx.Payer, costGas, config, overlay, store, notify, shardID); err != nil {
				return err
			}
			return fmt.Errorf("gas insufficient, balance:%d < costGas:%d", newBalance, costGas)
		}

		notifies, err = chargeCostGas(tx.Payer, costGas, config, sc.CacheDB, store, shardID)
		if err != nil {
			return err
		}
	}
	notify.Notify = append(notify.Notify, sc.Notifications...)
	notify.Notify = append(notify.Notify, notifies...)
	notify.GasConsumed = costGas
	notify.State = event.CONTRACT_STATE_SUCCESS
	sc.CacheDB.Commit()
	//todo : add notify to xshardDB
	xshardDB.Commit()
	return nil
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
	tr := new(bytes.Buffer)
	transfer.Serialize(tr)
	return tr.Bytes()
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

func costInvalidGas(address common.Address, gas uint64, config *smartcontract.Config, overlay *overlaydb.OverlayDB,
	store store.LedgerStore, notify *event.ExecuteNotify, shardID common.ShardID) error {
	cache := storage.NewCacheDB(overlay)
	notifies, err := chargeCostGas(address, gas, config, cache, store, shardID)
	if err != nil {
		return err
	}
	cache.Commit()
	notify.GasConsumed = gas
	notify.Notify = append(notify.Notify, notifies...)
	return nil
}

func calcGasByCodeLen(codeLen int, codeGas uint64) uint64 {
	return uint64(codeLen/neovm.PER_UNIT_CODE_LEN) * codeGas
}
