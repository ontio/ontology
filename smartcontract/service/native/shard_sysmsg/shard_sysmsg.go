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
	"github.com/ontio/ontology/smartcontract/service/native"
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

func RegisterShardSysMsgContract(native *native.NativeService) {
	native.Register(INIT_NAME, ShardSysMsgInit)
	native.Register(PROCESS_CROSS_SHARD_MSG, ProcessCrossShardMsg)
	native.Register(REMOTE_NOTIFY, RemoteNotify)
	native.Register(REMOTE_INVOKE, RemoteInvoke)
}

func ShardSysMsgInit(native *native.NativeService) ([]byte, error) {
	// TODO: nothing to do yet
	return utils.BYTE_TRUE, nil
}

func ProcessCrossShardMsg(native *native.NativeService) ([]byte, error) {
	cp := new(shardmgmt.CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
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
		log.Errorf("processing cross shard msg %d(%d)", evt.EventType, evt.FromHeight)
		if evt.Version != shardmgmt.VERSION_CONTRACT_SHARD_MGMT {
			continue
		}

		shardEvt, err := shardstates.DecodeShardEvent(evt.EventType, evt.Payload)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("processing shard event %d", evt.EventType)
		}

		switch evt.EventType {
		case shardstates.EVENT_SHARD_GAS_DEPOSIT:
			if err := processShardGasDeposit(native, shardEvt.(*shardstates.DepositGasEvent)); err != nil {
				return utils.BYTE_FALSE, fmt.Errorf("process gas deposit: %s", err)
			}
		case shardstates.EVENT_SHARD_GAS_WITHDRAW_REQ:
			if err := processShardGasWithdrawReq(shardEvt.(*shardstates.WithdrawGasReqEvent)); err != nil {
				return utils.BYTE_FALSE, fmt.Errorf("process gas deposit: %s", err)
			}
		}
	}

	return utils.BYTE_TRUE, nil
}

func processShardGasDeposit(env *native.NativeService, evt *shardstates.DepositGasEvent) error {
	return appCallTransfer(env, utils.OngContractAddress, utils.ShardSysMsgContractAddress, evt.User, evt.Amount)
}

func processShardGasWithdrawReq(evt *shardstates.WithdrawGasReqEvent) error {
	return nil
}

func RemoteNotify(native *native.NativeService) ([]byte, error) {
	// TODO: verify invocation from other smart contract

	cp := new(shardmgmt.CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remote notify, invalid cmd param: %s", err)
	}

	params := new(NotifyReqParam)
	if err := params.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remote notify, invalid param: %s", err)
	}

	// TODO: add evt to queue, update merkle root
	log.Infof("to send remote notify: from %d to %d", native.ShardID, params.ToShard)
	if err := addToShardsInBlock(native, params.ToShard); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remote notify, failed to add to-shard to block: %s", err)
	}
	if err := addReqsInBlock(native, params); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remote notify, failed to add req to block: %s", err)
	}

	return utils.BYTE_TRUE, nil
}

func RemoteInvoke(native *native.NativeService) ([]byte, error) {
	// send evnt to chainmgr
	return utils.BYTE_FALSE, nil
}
