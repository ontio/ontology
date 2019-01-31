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

const (
	INIT_NAME               = "init"
	PROCESS_CROSS_SHARD_MSG = "processShardMsg"
)

func InitShardSystemMessageContract() {
	native.Contracts[utils.ShardSysMsgContractAddress] = RegisterShardSysMsgContract
}

func RegisterShardSysMsgContract(native *native.NativeService) {
	native.Register(INIT_NAME, ShardSysMsgInit)
	native.Register(PROCESS_CROSS_SHARD_MSG, ProcessCrossShardMsg)
}

func ShardSysMsgInit(native *native.NativeService) ([]byte, error) {
	return utils.BYTE_FALSE, nil
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
