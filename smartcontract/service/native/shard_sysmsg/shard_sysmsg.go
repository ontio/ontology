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
	"errors"
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/chainmgr/xshard_state"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/shardgas"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/service/neovm"
)

/////////
//
// Shard-system contract
//
//	. manage user deposit gas on child shard
//	. process local shard request message
//	. process remote shard response message
//
/////////

var ErrYield = errors.New("transaction execution yielded")

const (
	// function names
	// 		processShardMsg: to process tx sent from remote shards
	//      remoteSendShardMsg/remoteInvoke: invoked by local-contract, trigger remote-tx event
	INIT_NAME               = "init"
	PROCESS_CROSS_SHARD_MSG = "processShardMsg"
	REMOTE_NOTIFY           = "remoteSendShardMsg"
	REMOTE_INVOKE           = "remoteInvoke"

	// key prefix
	KEY_SHARDS_IN_BLOCK = "shardsInBlock" // with block#, contains to-shard list
	KEY_REQS_IN_BLOCK   = "reqsInBlock"   // with block# - shard#, containers requests to shard#
)

func InitShardSystemMessageContract() {
	native.Contracts[utils.ShardSysMsgContractAddress] = RegisterShardSysMsgContract
}

func RegisterShardSysMsgContract(ctx *native.NativeService) {
	ctx.Register(INIT_NAME, ShardSysMsgInit)
	ctx.Register(PROCESS_CROSS_SHARD_MSG, ProcessCrossShardMsg)
	ctx.Register(REMOTE_NOTIFY, RemoteNotify)
	ctx.Register(REMOTE_INVOKE, RemoteInvoke)
}

func ShardSysMsgInit(ctx *native.NativeService) ([]byte, error) {
	// TODO: init sys-msg-queue Store
	return utils.BYTE_TRUE, nil
}

// RemoteNotify
// process requests from local-contract, to send notify-call to remote shard
func RemoteNotify(ctx *native.NativeService) ([]byte, error) {
	if ctx.ContextRef.IsPreExec() {
		return utils.BYTE_TRUE, nil
	}

	txHash := ctx.Tx.Hash()
	txState := xshard_state.CreateTxState(xshard_state.ShardTxID(string(txHash[:])))
	reqParam := new(NotifyReqParam)
	if err := reqParam.Deserialize(bytes.NewBuffer(ctx.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remote notify, invalid param: %s", err)
	}

	// send with minimal gas fee
	msg := &xshard_state.XShardNotify{
		NotifyID: txState.NumNotifies,
		Contract: reqParam.ToContract,
		Payer:    ctx.Tx.Payer,
		Fee:      neovm.MIN_TRANSACTION_GAS,
		Method:   reqParam.Method,
		Args:     reqParam.Args,
	}
	txState.NumNotifies += 1
	// todo: clean shardnotifies when replay transaction
	txState.ShardNotifies = append(txState.ShardNotifies, msg)
	//todo : clean this
	if err := remoteSendShardMsg(ctx, ctx.Tx.Hash(), reqParam.ToShard, msg); err != nil {
		return utils.BYTE_FALSE, err
	}
	return utils.BYTE_TRUE, nil
}

// RemoteInvoke
// process requests from local-contract, to send transactional-call to remote shard
func RemoteInvoke(ctx *native.NativeService) ([]byte, error) {
	if ctx.ContextRef.IsPreExec() {
		return utils.BYTE_TRUE, nil
	}

	reqParam := new(NotifyReqParam)
	if err := reqParam.Deserialize(bytes.NewBuffer(ctx.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remote invoke, invalid param: %s", err)
	}

	log.Debugf("received remote invoke: %s", string(ctx.Input))

	txHash := ctx.Tx.Hash()
	shardTxID := xshard_state.ShardTxID(string(txHash[:]))
	txState := xshard_state.CreateTxState(shardTxID).Clone()
	reqIdx := txState.NextReqID
	if reqIdx >= xshard_state.MaxRemoteReqPerTx {
		return utils.BYTE_FALSE, xshard_state.ErrTooMuchRemoteReq
	}
	msg := &xshard_state.XShardTxReq{
		IdxInTx:  uint64(reqIdx),
		Payer:    ctx.Tx.Payer,
		Fee:      0,
		Contract: reqParam.ToContract,
		Method:   reqParam.Method,
		Args:     reqParam.Args,
	}

	txState.NextReqID += 1

	if reqIdx < uint32(len(txState.OutReqResp)) {
		if xshard_state.IsXShardMsgEqual(msg, txState.OutReqResp[reqIdx].Req) == false {
			return utils.BYTE_FALSE, xshard_state.ErrMismatchedRequest
		}
		rspMsg := txState.OutReqResp[reqIdx].Resp
		var resultErr error
		if rspMsg.Error {
			resultErr = errors.New("remote invoke got error response")
		}
		return rspMsg.Result, resultErr
	}

	if len(txState.TxPayload) == 0 {
		txPayload := bytes.NewBuffer(nil)
		if err := ctx.Tx.Serialize(txPayload); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("remote invoke, failed to get tx payload: %s", err)
		}
		txState.TxPayload = txPayload.Bytes()
	}

	// no response found in tx-statedb, send request
	if err := txState.AddTxShard(reqParam.ToShard); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remote invoke, failed to add shard: %s", err)
	}

	txState.PendingReq = msg

	// put Tx-Request
	//todo: clean
	if err := remoteSendShardMsg(ctx, txHash, reqParam.ToShard, msg); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remote invoke, notify: %s", err)
	}

	xshard_state.PutTxState(shardTxID, txState)
	// clean write-set
	ctx.CacheDB.Reset()
	txState.SetTxExecutionPaused()

	return utils.BYTE_FALSE, ErrYield
}

// ProcessCrossShardMsg
// process remote-requests from remote shards
// including gas-message from root-chain, notify-call and transactional-call
func ProcessCrossShardMsg(ctx *native.NativeService) ([]byte, error) {
	if ctx.ContextRef.IsPreExec() {
		return utils.BYTE_TRUE, nil
	}

	// FIXME: verify transaction from system
	// check block-execution is at shard-tx processing stage

	param := new(CrossShardMsgParam)
	if err := param.Deserialize(bytes.NewBuffer(ctx.Input)); err != nil {
		log.Errorf("cross-shard msg, invalid input: %s", err)
		return utils.BYTE_FALSE, fmt.Errorf("cross-shard msg, invalid input: %s", err)
	}

	for _, evt := range param.Events {
		log.Debugf("processing cross shard msg %d(height: %d)", evt.EventType, evt.FromHeight)
		if evt.Version != shardmgmt.VERSION_CONTRACT_SHARD_MGMT {
			continue
		}

		switch evt.EventType {
		case shardstates.EVENT_SHARD_GAS_DEPOSIT:
			shardEvt, err := shardstates.DecodeShardGasEvent(evt.EventType, evt.Payload)
			if err != nil {
				return utils.BYTE_FALSE, fmt.Errorf("processing shard event %d: %s", evt.EventType, err)
			}

			if err := processShardGasDeposit(ctx, shardEvt.(*shardstates.DepositGasEvent)); err != nil {
				return utils.BYTE_FALSE, fmt.Errorf("process gas deposit: %s", err)
			}
		case shardstates.EVENT_SHARD_GAS_WITHDRAW_DONE:
			shardEvt, err := shardstates.DecodeShardGasEvent(evt.EventType, evt.Payload)
			if err != nil {
				return utils.BYTE_FALSE, fmt.Errorf("processing shard event %d: %s", evt.EventType, err)
			}

			if err := processShardGasWithdrawDone(ctx, shardEvt.(*shardstates.WithdrawGasDoneEvent)); err != nil {
				return utils.BYTE_FALSE, fmt.Errorf("process gas deposit: %s", err)
			}
		case xshard_state.EVENT_SHARD_MSG_COMMON:
			reqs, err := xshard_state.DecodeShardCommonReqs(evt.Payload)
			if err != nil {
				return utils.BYTE_FALSE, fmt.Errorf("decode shard reqs: %s", err)
			}
			for _, req := range reqs {
				log.Debugf("processing cross shard req %d(height: %d, type: %d)", evt.EventType, evt.FromHeight, req.Type)
				var txState *xshard_state.TxState
				var shardTxID xshard_state.ShardTxID
				txCompleted := false
				switch req.Type {
				case xshard_state.EVENT_SHARD_NOTIFY:
					nid := req.Msg.(*xshard_state.XShardNotify).NotifyID
					sink := common.NewZeroCopySink(0)
					sink.WriteBytes(req.SourceTxHash[:]) //todo : use shard tx id
					sink.WriteUint32(nid)
					shardTxID = xshard_state.ShardTxID(string(sink.Bytes()))

					txState := xshard_state.CreateTxState(shardTxID).Clone()
					if err = processXShardNotify(ctx, txState, req); err != nil {
						log.Errorf("process notify: %s", err)
					}
					txCompleted = true
				case xshard_state.EVENT_SHARD_TXREQ:
					shardTxID = xshard_state.ShardTxID(string(req.SourceTxHash[:]))
					txState = xshard_state.CreateTxState(shardTxID).Clone()
					if err = processXShardReq(ctx, txState, req); err != nil {
						log.Errorf("process xshard req: %s", err)
					}
					txCompleted = false
				case xshard_state.EVENT_SHARD_TXRSP:
					shardTxID = xshard_state.ShardTxID(string(req.SourceTxHash[:]))
					txState = xshard_state.CreateTxState(shardTxID).Clone()
					if err = processXShardRsp(ctx, txState, req); err != nil {
						log.Errorf("process xshard rsp: %s", err)
					}
					txCompleted = false
				case xshard_state.EVENT_SHARD_PREPARE:
					shardTxID = xshard_state.ShardTxID(string(req.SourceTxHash[:]))
					txState = xshard_state.CreateTxState(shardTxID).Clone()
					if err = processXShardPrepareMsg(ctx, txState, req); err != nil {
						log.Errorf("process xshard prepare: %s", err)
					}
					txCompleted = false
				case xshard_state.EVENT_SHARD_PREPARED:
					shardTxID = xshard_state.ShardTxID(string(req.SourceTxHash[:]))
					txState = xshard_state.CreateTxState(shardTxID).Clone()
					if err = processXShardPreparedMsg(ctx, txState, req); err != nil {
						log.Errorf("process xshard prepared: %s", err)
					}
					// FIXME: completed with all-shards-prepared
					txCompleted = true
				case xshard_state.EVENT_SHARD_COMMIT:
					shardTxID = xshard_state.ShardTxID(string(req.SourceTxHash[:]))
					txState = xshard_state.CreateTxState(shardTxID).Clone()
					if err = processXShardCommitMsg(ctx, txState, req); err != nil {
						log.Errorf("process xshard commit: %s", err)
					}
					txCompleted = true
				case xshard_state.EVENT_SHARD_ABORT:
					shardTxID = xshard_state.ShardTxID(string(req.SourceTxHash[:]))
					txState = xshard_state.CreateTxState(shardTxID).Clone()
					if err = processXShardAbortMsg(ctx, txState, req); err != nil {
						log.Errorf("process xshard abort: %s", err)
					}
					txCompleted = true
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
		default:
			return utils.BYTE_FALSE, fmt.Errorf("unknown event type: %d", evt.EventType)
		}
	}

	return utils.BYTE_TRUE, nil
}

func processShardGasDeposit(ctx *native.NativeService, evt *shardstates.DepositGasEvent) error {
	return ont.AppCallTransfer(ctx, utils.OngContractAddress, utils.ShardSysMsgContractAddress, evt.User, evt.Amount)
}

func processShardGasWithdrawDone(ctx *native.NativeService, evt *shardstates.WithdrawGasDoneEvent) error {
	param := &shardgas.UserWithdrawSuccessParam{
		User:       evt.User,
		WithdrawId: evt.WithdrawId,
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	if err != nil {
		return fmt.Errorf("processShardGasWithdrawDone: failed, err: %s", err)
	}
	_, err = ctx.NativeCall(utils.ShardGasMgmtContractAddress, shardgas.USER_WITHDRAW_SUCCESS, bf.Bytes())
	if err != nil {
		return fmt.Errorf("processShardGasWithdrawDone: call shard gas contract failed, err: %s", err)
	}
	return nil
}
