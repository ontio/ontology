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
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ontio/ontology/smartcontract/service/native/shardgas"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/chainmgr/xshard_state"
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
	// TODO: init sys-msg-queue Store
	return utils.BYTE_TRUE, nil
}

// RemoteNotify
// process requests from local-contract, to send notify-call to remote shard
func RemoteNotify(ctx *native.NativeService) ([]byte, error) {
	if ctx.ContextRef.IsPreExec() {
		return utils.BYTE_TRUE, nil
	}

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
	if err := remoteNotify(ctx, ctx.Tx.Hash(), reqParam.ToShard, msg); err != nil {
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

	cp := new(shardmgmt.CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(ctx.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remote invoke, invalid cmn param: %s", err)
	}

	reqParam := new(NotifyReqParam)
	if err := reqParam.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remote invoke, invalid param: %s", err)
	}

	log.Debugf("received remote invoke: %s", string(cp.Input))

	tx := ctx.Tx.Hash()
	reqIdx := xshard_state.GetNextReqIndex(tx)
	if reqIdx < 0 {
		return utils.BYTE_FALSE, xshard_state.ErrTooMuchRemoteReq
	}
	msg := &shardstates.XShardTxReq{
		IdxInTx:  reqIdx,
		Payer:    ctx.Tx.Payer,
		Fee:      0,
		Contract: reqParam.ToContract,
		Method:   reqParam.Method,
		Args:     reqParam.Args,
	}

	if err := xshard_state.ValidateTxRequest(tx, msg); err != nil {
		// TODO: abort transaction
		return utils.BYTE_FALSE, err
	}

	// get response from native-tx-statedb
	rspMsg, err := xshard_state.GetTxResponse(tx, msg)
	if rspMsg != nil {
		log.Debugf("remote invoke response available result: %v, err %s", rspMsg.Result, err)
		if err == nil {
			var resultErr error
			if rspMsg.Error != "" {
				resultErr = errors.New(rspMsg.Error)
			}
			return rspMsg.Result, resultErr
		} else if err != xshard_state.ErrNotFound {
			return utils.BYTE_FALSE, err
		}
	}

	// no response found in tx-statedb, send request
	txPayload := bytes.NewBuffer(nil)
	if err := ctx.Tx.Serialize(txPayload); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remote invoke, failed to get tx payload: %s", err)
	}
	if err := xshard_state.AddTxShard(tx, reqParam.ToShard); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remote invoke, failed to add shard: %s", err)
	}

	// put Tx-Request
	if err := xshard_state.PutTxRequest(tx, txPayload.Bytes(), msg); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remote invoke, put Tx request: %s", err)
	}
	if err := remoteNotify(ctx, tx, reqParam.ToShard, msg); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remote invoke, notify: %s", err)
	}
	// clean write-set
	ctx.CacheDB.Reset()
	if err := waitRemoteResponse(ctx, tx); err != nil {
		return utils.BYTE_FALSE, err
	}
	return nil, nil
}

// ProcessCrossShardMsg
// process remote-requests from remote shards
// including gas-message from root-chain, notify-call and transactional-call
func ProcessCrossShardMsg(ctx *native.NativeService) ([]byte, error) {
	if ctx.ContextRef.IsPreExec() {
		return utils.BYTE_TRUE, nil
	}

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
		case shardstates.EVENT_SHARD_MSG_COMMON:
			reqs, err := shardstates.DecodeShardCommonReqs(evt.Payload)
			if err != nil {
				return utils.BYTE_FALSE, fmt.Errorf("decode shard reqs: %s", err)
			}
			for _, req := range reqs {
				log.Debugf("processing cross shard req %d(height: %d, type: %d)", evt.EventType, evt.FromHeight, req.Type)
				txCompleted := false
				switch req.Type {
				case shardstates.EVENT_SHARD_NOTIFY:
					if err := processXShardNotify(ctx, req); err != nil {
						log.Errorf("process notify: %s", err)
					}
					txCompleted = true
				case shardstates.EVENT_SHARD_TXREQ:
					if err := processXShardReq(ctx, req); err != nil {
						log.Errorf("process xshard req: %s", err)
					}
					txCompleted = false
				case shardstates.EVENT_SHARD_TXRSP:
					if err := processXShardRsp(ctx, req); err != nil {
						log.Errorf("process xshard rsp: %s", err)
					}
					txCompleted = false
				case shardstates.EVENT_SHARD_PREPARE:
					if err := processXShardPrepareMsg(ctx, req); err != nil {
						log.Errorf("process xshard prepare: %s", err)
					}
					txCompleted = false
				case shardstates.EVENT_SHARD_PREPARED:
					if err := processXShardPreparedMsg(ctx, req); err != nil {
						log.Errorf("process xshard prepared: %s", err)
					}
					// FIXME: completed with all-shards-prepared
					txCompleted = true
				case shardstates.EVENT_SHARD_COMMIT:
					if err := processXShardCommitMsg(ctx, req); err != nil {
						log.Errorf("process xshard commit: %s", err)
					}
					txCompleted = true
				case shardstates.EVENT_SHARD_ABORT:
					if err := processXShardAbortMsg(ctx, req); err != nil {
						log.Errorf("process xshard abort: %s", err)
					}
					txCompleted = true
				}
				log.Debugf("DONE processing cross shard req %d(height: %d, type: %d)", evt.EventType, evt.FromHeight, req.Type)

				// transaction should be completed, and be removed from txstate-db
				if txCompleted {
					if shards, err := xshard_state.GetTxShards(req.SourceTxHash); err != xshard_state.ErrNotFound {
						for _, s := range shards {
							log.Errorf("TODO: abort transaction %d on shard %d", common.ToHexString(req.SourceTxHash[:]), s)
						}
					}
				} else {
					h := ctx.Tx.Hash()
					return utils.BYTE_FALSE, fmt.Errorf("tx %s not complete yet", hex.EncodeToString(h[:]))
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
