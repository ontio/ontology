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

package shardping

import (
	"bytes"
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/shard_sysmsg"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	"github.com/ontio/ontology/smartcontract/service/native/shardping/states"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

/////////
//
// Shard Ping test contract
//  (stateless test contract)
//
//	. Ping, response with Pong
//
/////////

const (
	// function name
	SHARD_PING_NAME      = "shardPing"
	SEND_SHARD_PING_NAME = "sendShardPing"
)

func InitShardPing() {
	native.Contracts[utils.ShardPingAddress] = RegisterShardPingContract
}

func RegisterShardPingContract(native *native.NativeService) {
	native.Register(SHARD_PING_NAME, ShardPingTest)
	native.Register(SEND_SHARD_PING_NAME, SendShardPingTest)
}

func ShardPingTest(native *native.NativeService) ([]byte, error) {
	cp := new(shardmgmt.CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ping shard, invalid cmd param: %s", err)
	}

	params := new(ShardPingParam)
	if err := params.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ping shard, invalid param: %s", err)
	}
	if params.ToShard != native.ShardID {
		return utils.BYTE_FALSE, fmt.Errorf("ping shard, invalid to shard: %d vs %d", params.ToShard, native.ShardID)
	}

	log.Infof("shard ping: from %d, to %d, param: %s", params.FromShard, params.ToShard, params.Param)
	return utils.BYTE_TRUE, nil
}

func SendShardPingTest(native *native.NativeService) ([]byte, error) {
	cp := new(shardmgmt.CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("send ping shard, invalid cmd param: %s", err)
	}

	params := new(ShardPingParam)
	if err := params.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("send ping shard, invalid param: %s", err)
	}
	if params.FromShard != native.ShardID {
		return utils.BYTE_FALSE, fmt.Errorf("send ping shard, invalid from shard: %d vs %d", params.FromShard, native.ShardID)
	}

	pingEvt := &shardping_events.SendShardPingEvent{
		Payload: "SendShardPingPayload",
	}
	sink := common.NewZeroCopySink(0)
	pingEvt.Serialization(sink)

	// call shard_sysmsg to send ping
	if err := appcallSendReq(native, params.ToShard, sink.Bytes()); err != nil {
		return utils.BYTE_FALSE, err
	}

	return utils.BYTE_TRUE, nil
}

func appcallSendReq(native *native.NativeService, toShard types.ShardID, payload []byte) error {
	paramBytes := new(bytes.Buffer)
	params := shardsysmsg.NotifyReqParam{
		ToShard: toShard,
		Args:    payload,
	}
	if err := params.Serialize(paramBytes); err != nil {
		return fmt.Errorf("send ping shard, marshal param: %s", err)
	}

	cmnBytes := new(bytes.Buffer)
	cmnParam := shardmgmt.CommonParam{paramBytes.Bytes()}
	if err := cmnParam.Serialize(cmnBytes); err != nil {
		return fmt.Errorf("send ping shard, marshal cmn param: %s", err)
	}

	if _, err := native.NativeCall(utils.ShardSysMsgContractAddress, shardsysmsg.REMOTE_NOTIFY, cmnBytes.Bytes()); err != nil {
		return fmt.Errorf("send ping shard, appcallSendReq: %s", err)
	}
	return nil
}
