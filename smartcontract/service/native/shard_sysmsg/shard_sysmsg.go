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

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/chainmgr/xshard_state"
	"github.com/ontio/ontology/core/xshard_types"
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
	//      remoteSendShardMsg/RemoteInvokeApi: invoked by local-contract, trigger remote-tx event
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

// runtime api
func RemoteNotifyApi(ctx *native.NativeService, param *NotifyReqParam) {
	txState := ctx.MainShardTxState
	// send with minimal gas fee
	msg := &xshard_types.XShardNotify{
		NotifyID: txState.NumNotifies,
		Contract: param.ToContract,
		Payer:    ctx.Tx.Payer,
		Fee:      neovm.MIN_TRANSACTION_GAS,
		Method:   param.Method,
		Args:     param.Args,
	}
	txState.NumNotifies += 1
	// todo: clean shardnotifies when replay transaction
	txState.ShardNotifies = append(txState.ShardNotifies, msg)
}

// runtime api
func RemoteInvokeApi(ctx *native.NativeService, reqParam *NotifyReqParam) ([]byte, error) {
	txState := ctx.MainShardTxState
	reqIdx := txState.NextReqID
	if reqIdx >= xshard_state.MaxRemoteReqPerTx {
		return utils.BYTE_FALSE, xshard_state.ErrTooMuchRemoteReq
	}
	msg := &xshard_types.XShardTxReq{
		ShardMsgHeader: xshard_types.ShardMsgHeader{
			SourceShardID: ctx.ShardID,
			SourceHeight:  uint64(ctx.Height),
			TargetShardID: reqParam.ToShard,
			SourceTxHash:  ctx.Tx.Hash(),
		},
		IdxInTx:  uint64(reqIdx),
		Payer:    ctx.Tx.Payer,
		Fee:      neovm.MIN_TRANSACTION_GAS,
		Contract: reqParam.ToContract,
		Method:   reqParam.Method,
		Args:     reqParam.Args,
	}
	txState.NextReqID += 1

	if reqIdx < uint32(len(txState.OutReqResp)) {
		if xshard_types.IsXShardMsgEqual(msg, txState.OutReqResp[reqIdx].Req) == false {
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
	txState.ExecState = xshard_state.ExecYielded

	return utils.BYTE_FALSE, ErrYield
}

// RemoteNotify
// process requests from local-contract, to send notify-call to remote shard
func RemoteNotify(ctx *native.NativeService) ([]byte, error) {
	if ctx.ContextRef.IsPreExec() {
		return utils.BYTE_TRUE, nil
	}

	reqParam := new(NotifyReqParam)
	if err := reqParam.Deserialize(bytes.NewBuffer(ctx.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remote notify, invalid param: %s", err)
	}

	RemoteNotifyApi(ctx, reqParam)

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

	//todo: move out of testsuite
	return RemoteInvokeApi(ctx, reqParam)
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
		case xshard_types.EVENT_SHARD_MSG_COMMON:
			panic("unimplemented")
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
