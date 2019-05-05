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
	"github.com/ontio/ontology/core/xshard_types"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/shardgas"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/testsuite"
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

	reqParam := new(NotifyReqParam)
	if err := reqParam.Deserialize(bytes.NewBuffer(ctx.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remote notify, invalid param: %s", err)
	}

	testsuite.RemoteNotify(ctx, reqParam)

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
	return testsuite.RemoteInvoke(ctx, reqParam)
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
