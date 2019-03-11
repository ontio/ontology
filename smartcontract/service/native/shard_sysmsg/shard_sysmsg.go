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

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
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

const (
	// function names
	// 		processShardMsg: to process tx sent from remote shards
	//      remoteNotify/remoteInvoke: invoked by local-contract, trigger remote-tx event
	INIT_NAME               = "init"
	PROCESS_CROSS_SHARD_MSG = "processShardMsg"
	REMOTE_NOTIFY           = "remoteNotify"
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
	// TODO: nothing to do yet
	return utils.BYTE_TRUE, nil
}

// RemoteNotify
// process requests from local-contract, to send notify-call to remote shard
func RemoteNotify(ctx *native.NativeService) ([]byte, error) {
	cp := new(shardmgmt.CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(ctx.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remote notify, invalid cmn param: %s", err)
	}

	reqParam := new(NotifyReqParam)
	if err := reqParam.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remote notify, invalid param: %s", err)
	}

	msg := &shardstates.XShardNotify{
		Contract: reqParam.ToContract,
		Payer:    ctx.Tx.Payer,
		Fee:      0,
		Method:   reqParam.Method,
		Args:     reqParam.Args,
	}
	return remoteNotify(ctx, ctx.Tx.Hash(), reqParam.ToShard, msg)
}

// RemoteInvoke
// process requests from local-contract, to send transactional-call to remote shard
func RemoteInvoke(ctx *native.NativeService) ([]byte, error) {
	cp := new(shardmgmt.CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(ctx.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remote invoke, invalid cmn param: %s", err)
	}

	reqParam := new(NotifyReqParam)
	if err := reqParam.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remote invoke, invalid param: %s", err)
	}

	txHash := ctx.Tx.Hash()
	msg := &shardstates.XShardTxReq{
		IdxInTx:  -1,
		Payer:    ctx.Tx.Payer,
		Fee:      0,
		Contract: reqParam.ToContract,
		Method:   reqParam.Method,
		Args:     reqParam.Args,
	}

	// get response from native-tx-statedb
	result, err := native.GetTxResponse(txHash, msg)
	if err != native.ErrNotFound {
		if err == native.ErrMismatchedRequest {
			// TODO: abort transaction
			return utils.BYTE_FALSE, err
		}
		return result, err
	}

	// no response found in tx-statedb, send request
	reqIdx := native.GetNextReqIndex(txHash)
	if reqIdx < 0 {
		return utils.BYTE_FALSE, native.ErrTooMuchRemoteReq
	}
	msg.IdxInTx = reqIdx
	txPayload := bytes.NewBuffer(nil)
	if err := ctx.Tx.Payload.Serialize(txPayload); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remote invoke, failed to get tx payload: %s", err)
	}
	if err := native.PutTxRequest(txHash, txPayload.Bytes(), msg); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remote invoke, put Tx request: %s", err)
	}
	if _, err := remoteNotify(ctx, txHash, reqParam.ToShard, msg); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remote invoke, notify: %s", err)
	}
	if err := waitRemoteResponse(ctx); err != nil {
		return utils.BYTE_FALSE, err
	}
	return nil, nil
}

// ProcessCrossShardMsg
// process remote-requests from remote shards
// including gas-message from root-chain, notify-call and transactional-call
func ProcessCrossShardMsg(ctx *native.NativeService) ([]byte, error) {
	cp := new(shardmgmt.CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(ctx.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("process cross shard, invalid cmd param: %s", err)
	}

	// FIXME: verify transaction from system
	// check block-execution is at shard-tx processing stage

	param := new(CrossShardMsgParam)
	if err := param.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		log.Errorf("cross-shard msg, invalid input: %s", err)
		return utils.BYTE_FALSE, fmt.Errorf("cross-shard msg, invalid input: %s", err)
	}

	for _, evt := range param.Events {
		log.Infof("processing cross shard msg %d(%d)", evt.EventType, evt.FromHeight)
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
		case shardstates.EVENT_SHARD_GAS_WITHDRAW_REQ:
			shardEvt, err := shardstates.DecodeShardGasEvent(evt.EventType, evt.Payload)
			if err != nil {
				return utils.BYTE_FALSE, fmt.Errorf("processing shard event %d: %s", evt.EventType, err)
			}

			if err := processShardGasWithdrawReq(shardEvt.(*shardstates.WithdrawGasReqEvent)); err != nil {
				return utils.BYTE_FALSE, fmt.Errorf("process gas deposit: %s", err)
			}
		case shardstates.EVENT_SHARD_MSG_COMMON:
			reqs, err := shardstates.DecodeShardCommonReqs(evt.Payload)
			if err != nil {
				return utils.BYTE_FALSE, fmt.Errorf("decode shard reqs: %s", err)
			}
			for _, req := range reqs {
				txCompleted := true
				switch req.Type {
				case shardstates.EVENT_SHARD_NOTIFY:
					if err := processXShardNotify(ctx, req); err != nil {
						return utils.BYTE_FALSE, fmt.Errorf("process notify: %s", err)
					}
				case shardstates.EVENT_SHARD_TXREQ:
					if err := processXShardReq(ctx, req); err != nil {
						return utils.BYTE_FALSE, fmt.Errorf("process req: %s", err)
					}
					txCompleted = false
				case shardstates.EVENT_SHARD_TXRSP:
					if err := processXShardRsp(ctx, req); err != nil {
						return utils.BYTE_FALSE, fmt.Errorf("process rsp: %s", err)
					}
				case shardstates.EVENT_SHARD_PREPARE:
					if err := processXShardPrepareMsg(ctx, req); err != nil {
						return utils.BYTE_FALSE, fmt.Errorf("process prepare-msg: %s", err)
					}
					txCompleted = false
				case shardstates.EVENT_SHARD_PREPARED:
					if err := processXShardPreparedMsg(ctx, req); err != nil {
						return utils.BYTE_FALSE, fmt.Errorf("process prepared-msg: %s", err)
					}
					txCompleted = false
				case shardstates.EVENT_SHARD_COMMIT:
					if err := processXShardCommitMsg(ctx, req); err != nil {
						return utils.BYTE_FALSE, fmt.Errorf("process commit-msg: %s", err)
					}
				case shardstates.EVENT_SHARD_ABORT:
					if err := processXShardAbortMsg(ctx, req); err != nil {
						return utils.BYTE_FALSE, fmt.Errorf("process abort-msg: %s", err)
					}
				}

				// transaction should be completed, and be removed from txstate-db
				if txCompleted {
					if shards, err := native.GetTxShards(req.SourceTxHash); err != native.ErrNotFound {
						for _, s := range shards {
							log.Errorf("TODO: abort transaction %d on shard %d", common.ToHexString(req.SourceTxHash[:]), s)
						}
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

func processShardGasWithdrawReq(evt *shardstates.WithdrawGasReqEvent) error {
	return nil
}
