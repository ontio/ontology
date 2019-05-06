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
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/xshard_types"
	"github.com/ontio/ontology/smartcontract/service/native"
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
	//      remoteSendShardMsg/InvokeRemoteShard: invoked by local-contract, trigger remote-tx event
	INIT_NAME               = "init"
	PROCESS_CROSS_SHARD_MSG = "processShardMsg"
	REMOTE_NOTIFY           = "remoteSendShardMsg"
	REMOTE_INVOKE           = "remoteInvoke"
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
//todo remove this function
func RemoteNotify(ctx *native.NativeService) ([]byte, error) {
	if ctx.ContextRef.IsPreExec() {
		return utils.BYTE_TRUE, nil
	}

	reqParam := new(NotifyReqParam)
	if err := reqParam.Deserialize(bytes.NewBuffer(ctx.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remote notify, invalid param: %s", err)
	}

	ctx.NotifyRemoteShard(reqParam.ToShard, reqParam.ToContract, reqParam.Method, reqParam.Args)

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

	return ctx.InvokeRemoteShard(reqParam.ToShard, reqParam.ToContract, reqParam.Method, reqParam.Args)
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
		if evt.Version != utils.VERSION_CONTRACT_SHARD_MGMT {
			continue
		}

		switch evt.EventType {
		case xshard_types.EVENT_SHARD_MSG_COMMON:
			panic("unimplemented")
		default:
			return utils.BYTE_FALSE, fmt.Errorf("unknown event type: %d", evt.EventType)
		}
	}

	return utils.BYTE_TRUE, nil
}
