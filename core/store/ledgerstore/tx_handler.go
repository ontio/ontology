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
	"math"
	"sort"
	"strconv"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/chainmgr/xshard_state"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/store"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/ontio/ontology/core/types"
	cutils "github.com/ontio/ontology/core/utils"
	"github.com/ontio/ontology/core/xshard_types"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/global_params"
	ninit "github.com/ontio/ontology/smartcontract/service/native/init"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
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
func HandleChangeMetadataTransaction(store store.LedgerStore, overlay *overlaydb.OverlayDB,
	cache *storage.CacheDB, tx *types.Transaction, header *types.Header, notify *event.ExecuteNotify) error {
	shardID, err := common.NewShardID(header.ShardID)
	if err != nil {
		return fmt.Errorf("generate shardId failed, %s", err)
	}
	cfg := &smartcontract.Config{
		ShardID:   shardID,
		Time:      header.Timestamp,
		Height:    header.Height,
		Tx:        tx,
		BlockHash: header.Hash(),
	}
	defer func() {
		if err != nil {
			if err := costInvalidGas(tx.Payer, tx.GasLimit*tx.GasPrice, cfg, cache, store, notify, shardID); err != nil {
				log.Errorf("HandleChangeMetadataTransaction: costInvalidGas failed, err: %s", err)
			}
			cache.Commit()
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

func resumeTxState(txState *xshard_state.TxState, rspMsg *xshard_types.XShardTxRsp) (*types.Transaction, error) {
	if txState.PendingOutReq == nil || txState.PendingOutReq.IdxInTx != rspMsg.IdxInTx {
		// todo: system error or remote shard error
		return nil, fmt.Errorf("invalid response id: %d", rspMsg.IdxInTx)
	}

	txState.OutReqResp = append(txState.OutReqResp, &xshard_state.XShardTxReqResp{Req: txState.PendingOutReq, Resp: rspMsg})
	txState.PendingOutReq = nil

	txPayload := txState.TxPayload
	if txPayload == nil {
		return nil, fmt.Errorf("failed to get tx payload")
	}

	tx, err := types.TransactionFromRawBytes(txPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to re-init original tx: %s", err)
	}

	return tx, nil
}

func handleShardAbortMsg(msg *xshard_types.XShardAbortMsg, store store.LedgerStore, overlay *overlaydb.OverlayDB,
	cache *storage.CacheDB, xshardDB *storage.XShardDB, header *types.Header, notify *event.TransactionNotify) {
	shardTxID := msg.ShardTxID

	txState, err := xshardDB.GetXShardState(shardTxID)
	if err != nil {
		return
	}
	// update tx state
	txState.ExecState = xshard_state.ExecAborted
	//unlockTxContract(ctx, tx)

	xshardDB.SetXShardState(txState)
}

func handleShardCommitMsg(msg *xshard_types.XShardCommitMsg, store store.LedgerStore, overlay *overlaydb.OverlayDB,
	cache *storage.CacheDB, xshardDB *storage.XShardDB, header *types.Header, notify *event.TransactionNotify) {
	shardTxID := msg.ShardTxID
	txState, err := xshardDB.GetXShardState(shardTxID)
	if err != nil {
		return
	}
	if txState.ExecState == xshard_state.ExecCommited {
		return
	}

	// update tx state
	// commit the cached rwset
	cache.SetCache(txState.WriteSet)
	txState.ExecState = xshard_state.ExecCommited
	//todo : determine which tx the notification belong
	notify.ContractEvent.Notify = append(notify.ContractEvent.Notify, txState.Notify.Notify...)

	for _, msg := range txState.ShardNotifies {
		notify.ShardMsg = append(notify.ShardMsg, msg)
	}
	for _, shard := range txState.GetTxShards() {
		cmt := &xshard_types.XShardCommitMsg{
			ShardMsgHeader: xshard_types.ShardMsgHeader{
				SourceShardID: common.NewShardIDUnchecked(header.ShardID),
				TargetShardID: shard,
				SourceTxHash:  msg.SourceTxHash,
				ShardTxID:     shardTxID,
			},
		}
		notify.ShardMsg = append(notify.ShardMsg, cmt)
	}

	// todo:
	//unlockTxContract(ctx, tx)

	xshardDB.SetXShardState(txState)
}

func handleShardPreparedMsg(msg *xshard_types.XShardPreparedMsg, store store.LedgerStore, overlay *overlaydb.OverlayDB,
	cache *storage.CacheDB, xshardDB *storage.XShardDB, header *types.Header, notify *event.TransactionNotify) {
	shardTxID := msg.ShardTxID
	txState, err := xshardDB.GetXShardState(shardTxID)
	if err != nil {
		return
	}

	if _, present := txState.Shards[msg.SourceShardID]; !present {
		log.Infof("invalid shard ID %d, in tx commit", msg.SourceShardID)
		return
	}
	txState.Shards[msg.SourceShardID] = xshard_state.ExecPrepared

	if !txState.IsCommitReady() {
		xshardDB.SetXShardState(txState)
		// wait for prepared from all shards
		return
	}
	if txState.PendingPrepare != nil {
		prepMsg := txState.PendingPrepare
		txState.PendingPrepare = nil
		// response prepared
		preparedMsg := &xshard_types.XShardPreparedMsg{
			ShardMsgHeader: xshard_types.ShardMsgHeader{
				SourceShardID: common.NewShardIDUnchecked(header.ShardID),
				TargetShardID: prepMsg.SourceShardID,
				SourceTxHash:  prepMsg.SourceTxHash,
				ShardTxID:     shardTxID,
			},
		}

		notify.ShardMsg = append(notify.ShardMsg, preparedMsg)
		return
	}

	for _, shard := range txState.GetTxShards() {
		cmt := &xshard_types.XShardCommitMsg{
			ShardMsgHeader: xshard_types.ShardMsgHeader{
				SourceShardID: common.NewShardIDUnchecked(header.ShardID),
				TargetShardID: shard,
				SourceTxHash:  msg.SourceTxHash,
				ShardTxID:     shardTxID,
			},
		}
		notify.ShardMsg = append(notify.ShardMsg, cmt)
	}

	// commit cached rwset
	cache.SetCache(txState.WriteSet)
	txState.ExecState = xshard_state.ExecCommited
	//todo : determine which tx the notification belong
	notify.ContractEvent.Notify = append(notify.ContractEvent.Notify, txState.Notify.Notify...)
	for _, notifyMsg := range txState.ShardNotifies {
		notify.ShardMsg = append(notify.ShardMsg, notifyMsg)
	}
	// todo:
	//unlockTxContract(ctx, tx)
	xshardDB.SetXShardState(txState)
}

func handleShardPrepareMsg(prepMsg *xshard_types.XShardPrepareMsg, store store.LedgerStore, overlay *overlaydb.OverlayDB,
	cache *storage.CacheDB, xshardDB *storage.XShardDB, header *types.Header, notify *event.TransactionNotify) {
	shardTxID := prepMsg.ShardTxID
	txState, err := xshardDB.GetXShardState(shardTxID)
	if err != nil {
		return
	}
	if txState.ExecState == xshard_state.ExecPrepared {
		// this case will happen when the transaction flow is as follows:
		//        / -> req shard2 \
		// shard1                  --> req shard3 -> req other shard
		//        \ -> req shard4 /
		// so shard3 may recieve PrepareMsg from shard2 and shard4 concurrenty, we just reply prepared directly
		// because even if shard3 will aborted finally, it will propagate it to shard1 by reply to the first preprare msg
		preparedMsg := &xshard_types.XShardPreparedMsg{
			ShardMsgHeader: xshard_types.ShardMsgHeader{
				SourceShardID: common.NewShardIDUnchecked(header.ShardID),
				TargetShardID: prepMsg.SourceShardID,
				SourceTxHash:  prepMsg.SourceTxHash,
				ShardTxID:     shardTxID,
			},
		}
		notify.ShardMsg = append(notify.ShardMsg, preparedMsg)

		return
	}
	txState.ExecState = xshard_state.ExecNone

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
	txState.ShardNotifies = nil
	// 1. re-execute all requests
	// 2. compare new responses with stored responses
	contractEvent := &event.ExecuteNotify{}
	prepareOK := true
	for _, val := range reqResp {
		req := val.Req
		rspMsg := val.Resp

		subTx, err := buildTx(req.Payer, req.Contract, req.Method, []interface{}{req.Args}, header.ShardID, req.Fee,
			header.Timestamp)
		if err != nil {
			return
		}

		cache.Reset()
		result2, _, err2 := execShardTransaction(req.SourceShardID, store, cache, txState, subTx, header, contractEvent)

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
		abort := &xshard_types.XShardAbortMsg{
			ShardMsgHeader: xshard_types.ShardMsgHeader{
				SourceShardID: common.NewShardIDUnchecked(header.ShardID),
				TargetShardID: prepMsg.SourceShardID,
				SourceTxHash:  prepMsg.SourceTxHash,
				ShardTxID:     shardTxID,
			},
		}

		// TODO: clean TX resources
		notify.ShardMsg = append(notify.ShardMsg, abort)

		txState.ExecState = xshard_state.ExecAborted
		xshardDB.SetXShardState(txState)
		return
	}

	// save tx rwset and reset ctx.CacheDB
	// TODO: add notification to cached DB
	txState.WriteSet = cache.GetCache()
	txState.Notify = contractEvent
	cache = storage.NewCacheDB(cache.GetBackendDB())
	txState.ExecState = xshard_state.ExecPrepared

	reqShards := txState.GetTxShards()
	if len(reqShards) != 0 {
		for _, s := range txState.GetTxShards() {
			msg := &xshard_types.XShardPrepareMsg{
				ShardMsgHeader: xshard_types.ShardMsgHeader{
					SourceShardID: common.NewShardIDUnchecked(header.ShardID),
					TargetShardID: s,
					SourceTxHash:  prepMsg.SourceTxHash,
					ShardTxID:     shardTxID,
				},
			}
			notify.ShardMsg = append(notify.ShardMsg, msg)
		}

		txState.PendingPrepare = prepMsg
	} else {
		// response prepared
		preparedMsg := &xshard_types.XShardPreparedMsg{
			ShardMsgHeader: xshard_types.ShardMsgHeader{
				SourceShardID: common.NewShardIDUnchecked(header.ShardID),
				TargetShardID: prepMsg.SourceShardID,
				SourceTxHash:  prepMsg.SourceTxHash,
				ShardTxID:     shardTxID,
			},
		}
		//if err := lockTxContracts(ctx, tx, nil, nil); err != nil {
		//	// FIXME
		//	return err
		//}
		notify.ShardMsg = append(notify.ShardMsg, preparedMsg)
	}

	xshardDB.SetXShardState(txState)
	return
}

// todo: handle pending case
func handleShardNotifyMsg(msg *xshard_types.XShardNotify, store store.LedgerStore, overlay *overlaydb.OverlayDB,
	cache *storage.CacheDB, xshardDB *storage.XShardDB, header *types.Header, notify *event.TransactionNotify) {
	shardId, err := common.NewShardID(header.ShardID)
	if err != nil {
		log.Debugf("handle shard notify check header shardId %s", err)
		return
	}
	nid := msg.NotifyID
	sink := common.NewZeroCopySink(0)
	sink.WriteBytes([]byte(msg.ShardTxID))
	sink.WriteUint32(nid)
	shardTxID := xshard_types.ShardTxID(string(sink.Bytes()))
	txState := xshard_state.CreateTxState(shardTxID)
	txState.ExecState = xshard_state.ExecNone

	tx, err := buildTx(msg.Payer, msg.Contract, msg.Method, []interface{}{msg.Args}, header.ShardID, msg.Fee,
		header.Timestamp)
	if err != nil {
		log.Debugf("handle shard notify failed %s", err)
		return
	}

	cache.Reset()
	result, gasConsume, err := execShardTransaction(msg.SourceShardID, store, cache, txState, tx, header, notify.ContractEvent)
	if err != nil && txState.ExecState == xshard_state.ExecYielded {
		notify.ShardMsg = append(notify.ShardMsg, txState.PendingOutReq)

		txState.ShardNotifies = nil
		xshardDB.SetXShardState(txState)
		return
	}

	if gasConsume < neovm.MIN_TRANSACTION_GAS {
		gasConsume = neovm.MIN_TRANSACTION_GAS
	}
	if tx.GasPrice > 0 {
		// add payer permission
		tx.SignedAddr = append(tx.SignedAddr, tx.Payer) // TODO: consider risk
		cfg := &smartcontract.Config{
			ShardID:   shardId,
			Time:      header.Timestamp,
			Height:    header.Height,
			Tx:        tx,
			BlockHash: header.Hash(),
		}
		minGas := tx.GasPrice * neovm.MIN_TRANSACTION_GAS
		if err != nil {
			log.Debugf("handle shard notify error %s", err)
			if err := costInvalidGas(tx.Payer, minGas, cfg, cache, store, notify.ContractEvent, shardId); err != nil {
				log.Debugf("handle shard notify: tx failed, cost invalid gas failed, %s", err)
				return
			}
			return
		} else {
			if chargeNotifies, err := chargeCostGas(tx.Payer, gasConsume*tx.GasPrice, cfg, cache, store, shardId); err != nil {
				log.Debugf("handle shard notify: charge failed, %s", err)
				if err := costInvalidGas(tx.Payer, minGas, cfg, cache, store, notify.ContractEvent, shardId); err != nil {
					log.Debugf("handle shard notify: cost invalid gas failed, %s", err)
					return
				}
			} else {
				notify.ContractEvent.Notify = append(notify.ContractEvent.Notify, chargeNotifies...)
			}
		}
	}

	// no shard transaction and tx completed
	for _, notifyMsg := range txState.ShardNotifies {
		notify.ShardMsg = append(notify.ShardMsg, notifyMsg)
	}

	log.Debugf("process xshard notify result: %v", result)
}

func handleShardReqMsg(msg *xshard_types.XShardTxReq, store store.LedgerStore, overlay *overlaydb.OverlayDB,
	cache *storage.CacheDB, xshardDB *storage.XShardDB, header *types.Header, notify *event.TransactionNotify) {
	shardTxID := msg.ShardTxID
	txState, err := xshardDB.GetXShardState(shardTxID)
	if err != nil {
		return
	}
	txState.ExecState = xshard_state.ExecNone

	shardReq := txState.InReqResp[msg.SourceShardID]
	lenReq := uint64(len(shardReq))
	if lenReq > 0 && msg.IdxInTx <= shardReq[lenReq-1].Req.IdxInTx {
		log.Debugf("handle shard req error: request is stale")
		return
	}

	subTx, err := buildTx(msg.Payer, msg.Contract, msg.Method, []interface{}{msg.Args}, header.ShardID, msg.Fee,
		header.Timestamp)
	if err != nil {
		log.Debugf("handle shard req error: %s", err)
		return
	}

	cache.Reset()
	evts := &event.ExecuteNotify{}
	result, feeUsed, err := execShardTransaction(msg.SourceShardID, store, cache, txState, subTx, header, evts)
	log.Debugf("xshard msg: method: %s, args: %v, result: %v, err: %s", msg.GetMethod(), msg.GetArgs(), result, err)

	if feeUsed < neovm.MIN_TRANSACTION_GAS {
		feeUsed = neovm.MIN_TRANSACTION_GAS
	}
	rspMsg := &xshard_types.XShardTxRsp{
		ShardMsgHeader: xshard_types.ShardMsgHeader{
			SourceShardID: common.NewShardIDUnchecked(header.ShardID),
			TargetShardID: msg.SourceShardID,
			SourceTxHash:  msg.SourceTxHash,
			ShardTxID:     msg.ShardTxID,
		},
		IdxInTx: msg.IdxInTx,
		FeeUsed: feeUsed,
	}
	if subTx.GasPrice > 0 {
		feeParam := &shardmgmt.XShardHandlingFeeParam{
			IsDebt:  false,
			ShardId: msg.SourceShardID,
			Fee:     feeUsed * subTx.GasPrice,
		}
		recordXShardHandlingFee(msg.TargetShardID, txState, feeParam, store, cache, header, notify)
	}
	if err != nil {
		if txState.ExecState == xshard_state.ExecYielded {
			txState.PendingInReq = msg
			notify.ShardMsg = append(notify.ShardMsg, txState.PendingOutReq)

			txState.ShardNotifies = nil
			xshardDB.SetXShardState(txState)
			return
		}
		rspMsg.Error = true // todo pending case
	} else {
		res, _ := result.(*ntypes.ByteArray).GetByteArray() // todo
		rspMsg.Result = res
	}

	// FIXME: save notification
	// FIXME: feeUsed
	txState.InReqResp[msg.SourceShardID] = append(txState.InReqResp[msg.SourceShardID],
		&xshard_state.XShardTxReqResp{Req: msg, Resp: rspMsg, Index: txState.TotalInReq})
	txState.TotalInReq += 1

	notify.ShardMsg = append(notify.ShardMsg, rspMsg)

	txState.ShardNotifies = nil
	xshardDB.SetXShardState(txState)
	log.Debugf("process xshard request result: %v", result)
}

func handleShardRespMsg(msg *xshard_types.XShardTxRsp, store store.LedgerStore, overlay *overlaydb.OverlayDB,
	cache *storage.CacheDB, xshardDB *storage.XShardDB, header *types.Header, notify *event.TransactionNotify) {
	shardId, err := common.NewShardID(header.ShardID)
	if err != nil {
		log.Debugf("handle shard resp check shardId failed, err: %s", err)
		return
	}
	shardTxID := msg.ShardTxID
	txState, err := xshardDB.GetXShardState(shardTxID)
	if err != nil {
		return
	}
	txState.ExecState = xshard_state.ExecNone
	if msg.FeeUsed < neovm.MIN_TRANSACTION_GAS {
		msg.FeeUsed = neovm.MIN_TRANSACTION_GAS
	}
	subTx, err := resumeTxState(txState, msg)
	if err != nil {
		return
	}
	// re-execute tx
	txState.NextReqID = 0
	txState.ShardNotifies = nil
	cache.Reset()
	isChargeFailed := false
	if subTx.GasPrice > 0 {
		// charge req handling fee
		cfg := &smartcontract.Config{
			ShardID:   shardId,
			Time:      header.Timestamp,
			Height:    header.Height,
			Tx:        subTx, // original tx has payer signature
			BlockHash: header.Hash(),
		}
		notifies, err := chargeCostGas(subTx.Payer, msg.FeeUsed*subTx.GasPrice, cfg, cache, store, shardId)
		if err != nil {
			log.Debugf("handle shard resp, charge xshard handling fee failed, err: %s", err)
			fee := neovm.MIN_TRANSACTION_GAS * subTx.GasPrice
			err = costInvalidGas(subTx.Payer, fee, cfg, cache, store, notify.ContractEvent, shardId)
			if err != nil {
				log.Debugf("handle shard resp, cost invalid handling fee failed, err: %s", err)
			} else {
				feeParam := &shardmgmt.XShardHandlingFeeParam{
					IsDebt:  true,
					ShardId: msg.SourceShardID,
					Fee:     fee,
				}
				recordXShardHandlingFee(shardId, txState, feeParam, store, cache, header, notify)
			}
			isChargeFailed = true
		} else {
			notify.ContractEvent.Notify = append(notify.ContractEvent.Notify, notifies...)
			feeParam := &shardmgmt.XShardHandlingFeeParam{
				IsDebt:  true,
				ShardId: msg.SourceShardID,
				Fee:     msg.FeeUsed * subTx.GasPrice,
			}
			recordXShardHandlingFee(shardId, txState, feeParam, store, cache, header, notify)
		}
	}

	evts := &event.ExecuteNotify{
		TxHash: msg.SourceTxHash, // todo
	}
	result, _, err := execShardTransaction(msg.SourceShardID, store, cache, txState, subTx, header, evts)
	if err != nil && txState.ExecState == xshard_state.ExecYielded {
		notify.ShardMsg = append(notify.ShardMsg, txState.PendingOutReq)
		txState.ShardNotifies = nil

		xshardDB.SetXShardState(txState)
		return
	}

	if txState.PendingInReq != nil {
		reqMsg := txState.PendingInReq
		rspMsg := &xshard_types.XShardTxRsp{
			ShardMsgHeader: xshard_types.ShardMsgHeader{
				SourceShardID: common.NewShardIDUnchecked(header.ShardID),
				TargetShardID: reqMsg.SourceShardID,
				SourceTxHash:  reqMsg.SourceTxHash,
				ShardTxID:     reqMsg.ShardTxID,
			},
			IdxInTx: reqMsg.IdxInTx,
			FeeUsed: 0,
		}
		if err != nil {
			rspMsg.Error = true // todo pending case
		} else {
			res, _ := result.(*ntypes.ByteArray).GetByteArray() // todo
			rspMsg.Result = res
		}

		txState.InReqResp[reqMsg.SourceShardID] = append(txState.InReqResp[reqMsg.SourceShardID],
			&xshard_state.XShardTxReqResp{Req: reqMsg, Resp: rspMsg, Index: txState.TotalInReq})
		txState.PendingInReq = nil
		txState.TotalInReq += 1

		notify.ShardMsg = append(notify.ShardMsg, rspMsg)

		txState.ShardNotifies = nil
		xshardDB.SetXShardState(txState)

		return
	}

	if err != nil || isChargeFailed {
		for _, s := range txState.GetTxShards() {
			abort := &xshard_types.XShardAbortMsg{
				ShardMsgHeader: xshard_types.ShardMsgHeader{
					SourceShardID: common.NewShardIDUnchecked(header.ShardID),
					TargetShardID: s,
					SourceTxHash:  msg.SourceTxHash,
					ShardTxID:     shardTxID,
				},
			}
			notify.ShardMsg = append(notify.ShardMsg, abort)
		}

		txState.ExecState = xshard_state.ExecAborted
		xshardDB.SetXShardState(txState)

		return
	}

	// tx completed send prepare msg
	for _, s := range txState.GetTxShards() {
		msg := &xshard_types.XShardPrepareMsg{
			ShardMsgHeader: xshard_types.ShardMsgHeader{
				SourceShardID: common.NewShardIDUnchecked(header.ShardID),
				TargetShardID: s,
				SourceTxHash:  subTx.Hash(),
				ShardTxID:     shardTxID,
			},
		}
		notify.ShardMsg = append(notify.ShardMsg, msg)
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
	txState.Notify = evts
	txState.ExecState = xshard_state.ExecPrepared
	cache = storage.NewCacheDB(cache.GetBackendDB())

	xshardDB.SetXShardState(txState)

	log.Debugf("starting 2pc, to wait response")
	return
}

func execShardTransaction(fromShard common.ShardID, store store.LedgerStore, cache *storage.CacheDB,
	txState *xshard_state.TxState, tx *types.Transaction, header *types.Header,
	notify *event.ExecuteNotify) (result interface{}, gasConsume uint64, err error) {
	result = nil
	gasConsume = 0
	shardID, err := common.NewShardID(header.ShardID)
	if err != nil {
		return
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
		uintCodeGasPrice, ok := neovm.GAS_TABLE.Load(neovm.UINT_INVOKE_CODE_LEN_NAME)
		if !ok {
			err = fmt.Errorf("execShardTransaction: get UINT_INVOKE_CODE_LEN_NAME gas failed")
			return
		}
		codeLenGasLimit := calcGasByCodeLen(len(invoke.Code), uintCodeGasPrice.(uint64))
		if tx.GasLimit < codeLenGasLimit {
			err = fmt.Errorf("execShardTransaction: gas not enough")
			return
		}
		sc := smartcontract.SmartContract{
			Config:       config,
			Store:        store,
			ShardTxState: txState,
			IsShardCall:  true,
			FromShard:    fromShard,
			CacheDB:      cache,
			Gas:          tx.GasLimit - codeLenGasLimit,
		}

		//start the smart contract executive function
		engine, _ := sc.NewExecuteEngine(invoke.Code)
		res, err := engine.Invoke()
		notify.Notify = append(notify.Notify, sc.Notifications...)
		gasConsume = tx.GasLimit - sc.Gas
		return res, gasConsume, err
	}

	panic("unimplemented")
}

func HandleShardCallTransaction(store store.LedgerStore, overlay *overlaydb.OverlayDB,
	cache *storage.CacheDB, xshardDB *storage.XShardDB,
	msgs []xshard_types.CommonShardMsg, header *types.Header, notify *event.TransactionNotify) error {
	for _, req := range msgs {
		switch msg := req.(type) {
		case *xshard_types.XShardNotify:
			handleShardNotifyMsg(msg, store, overlay, cache, xshardDB, header, notify)
		case *xshard_types.XShardTxReq:
			handleShardReqMsg(msg, store, overlay, cache, xshardDB, header, notify)
		case *xshard_types.XShardTxRsp:
			handleShardRespMsg(msg, store, overlay, cache, xshardDB, header, notify)
		case *xshard_types.XShardPrepareMsg:
			handleShardPrepareMsg(msg, store, overlay, cache, xshardDB, header, notify)
		case *xshard_types.XShardPreparedMsg:
			handleShardPreparedMsg(msg, store, overlay, cache, xshardDB, header, notify)
		case *xshard_types.XShardCommitMsg:
			handleShardCommitMsg(msg, store, overlay, cache, xshardDB, header, notify)
		case *xshard_types.XShardAbortMsg:
			handleShardAbortMsg(msg, store, overlay, cache, xshardDB, header, notify)
		}
	}

	return nil
}

//HandleInvokeTransaction deal with smart contract invoke transaction
func HandleInvokeTransaction(store store.LedgerStore, overlay *overlaydb.OverlayDB,
	cache *storage.CacheDB, xshardDB *storage.XShardDB, tx *types.Transaction, header *types.Header,
	notify *event.TransactionNotify) (result interface{}, err error) {
	invoke := tx.Payload.(*payload.InvokeCode)
	code := invoke.Code
	sysTransFlag := bytes.Compare(code, ninit.COMMIT_DPOS_BYTES) == 0 || header.Height == 0
	txHash := tx.Hash()
	txState := xshard_state.CreateTxState(xshard_types.ShardTxID(string(txHash[:])))

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
			if err = costInvalidGas(tx.Payer, oldBalance, config, cache, store, notify.ContractEvent, shardID); err != nil {
				return
			}
			return nil, fmt.Errorf("balance gas: %d less than min gas: %d", oldBalance, minGas)
		}

		codeLenGasLimit = calcGasByCodeLen(len(invoke.Code), uintCodeGasPrice.(uint64))

		if oldBalance < codeLenGasLimit*tx.GasPrice {
			if err = costInvalidGas(tx.Payer, oldBalance, config, cache, store, notify.ContractEvent, shardID); err != nil {
				return
			}
			return nil, fmt.Errorf("balance gas insufficient: balance:%d < code length need gas:%d", oldBalance, codeLenGasLimit*tx.GasPrice)
		}

		if tx.GasLimit < codeLenGasLimit {
			if err = costInvalidGas(tx.Payer, tx.GasLimit*tx.GasPrice, config, cache, store, notify.ContractEvent, shardID); err != nil {
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
		Config:       config,
		CacheDB:      cache,
		ShardTxState: txState,
		Store:        store,
		Gas:          availableGasLimit - codeLenGasLimit,
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
			if err = costInvalidGas(tx.Payer, costGas, config, cache, store, notify.ContractEvent, shardID); err != nil {
				return
			}
		}

		if txState.ExecState == xshard_state.ExecYielded {
			notify.ShardMsg = append(notify.ShardMsg, txState.PendingOutReq)

			txState.ShardNotifies = nil
			xshardDB.SetXShardState(txState)
			return nil, nil
		}

		return nil, err
	}

	var notifies []*event.NotifyEventInfo
	if isCharge {
		newBalance, err = getBalanceFromNative(config, cache, store, tx.Payer)
		if err != nil {
			return nil, err
		}

		if newBalance < costGas {
			if err = costInvalidGas(tx.Payer, costGas, config, cache, store, notify.ContractEvent, shardID); err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("gas insufficient, balance:%d < costGas:%d", newBalance, costGas)
		}

		notifies, err = chargeCostGas(tx.Payer, costGas, config, sc.CacheDB, store, shardID)
		if err != nil {
			return nil, err
		}
	}
	cnotify := notify.ContractEvent
	cnotify.Notify = append(cnotify.Notify, sc.Notifications...)
	cnotify.Notify = append(cnotify.Notify, notifies...)
	cnotify.GasConsumed = costGas
	cnotify.State = event.CONTRACT_STATE_SUCCESS
	sc.CacheDB.Commit()
	for _, msg := range txState.ShardNotifies {
		notify.ShardMsg = append(notify.ShardMsg, msg)
	}

	return result, nil
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
		contractAddr = utils.ShardMgmtContractAddress
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
	n, gasPriceParam := params.GetParam(genesis.NAME_GAS_PRICE)
	if n != -1 && gasPriceParam.Value != "" {
		neovm.GAS_PRICE, _ = strconv.ParseUint(gasPriceParam.Value, 10, 64)
	}
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

func buildTx(payer, contract common.Address, method string, args []interface{}, shardId, gasLimit uint64,
	nonce uint32) (*types.Transaction, error) {
	invokeCode := []byte{}
	var err error = nil
	if _, ok := native.Contracts[contract]; ok {
		invokeCode, err = cutils.BuildNativeInvokeCode(contract, 0, method, args)
	} else {
		invokeCode, err = cutils.BuildNeoVMInvokeCode(contract, []interface{}{method, args})
	}
	if err != nil {
		return nil, fmt.Errorf("buildTx: build invoke failed, err: %s", err)
	}
	invokePayload := &payload.InvokeCode{
		Code: invokeCode,
	}
	mutable := &types.MutableTransaction{
		Version:  common.CURR_TX_VERSION,
		GasPrice: neovm.GAS_PRICE,
		ShardID:  shardId,
		GasLimit: gasLimit,
		TxType:   types.Invoke,
		Nonce:    nonce,
		Payer:    payer,
		Payload:  invokePayload,
		Sigs:     make([]types.Sig, 0, 0),
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		return nil, fmt.Errorf("buildTx: build tx failed, err: %s", err)
	}
	return tx, nil
}

func recordXShardHandlingFee(selfShard common.ShardID, txState *xshard_state.TxState,
	feeParam *shardmgmt.XShardHandlingFeeParam, store store.LedgerStore, cache *storage.CacheDB, header *types.Header,
	notify *event.TransactionNotify) {
	chargeFeeTx, err := buildTx(common.ADDRESS_EMPTY, utils.ShardMgmtContractAddress,
		shardmgmt.UPDATE_XSHARD_HANDLING_FEE, []interface{}{feeParam}, header.ShardID, 0, 0)
	if err != nil {
		log.Debugf("handle shard resp, build xshard fee tx error: %s", err)
	} else {
		_, _, err = execShardTransaction(selfShard, store, cache, txState, chargeFeeTx, header, notify.ContractEvent)
		if err != nil {
			log.Debugf("handle shard resp, record xshard fee tx error: %s", err)
		}
	}
}
