package shardping

import (
	"bytes"
	"fmt"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/service/native/shardping/states"
	"github.com/ontio/ontology/smartcontract/service/native/shard_sysmsg"
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
	SHARD_PING_NAME = "shardPing"
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
	buf := new(bytes.Buffer)
	if err := pingEvt.Serialize(buf); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("send ping shard, serialize failed: %s", err)
	}

	// call shard_sysmsg to send ping
	if err := appcallSendReq(native, params.ToShard, buf.Bytes()); err != nil {
		return utils.BYTE_FALSE, err
	}

	return utils.BYTE_TRUE, nil
}

func appcallSendReq(native *native.NativeService, toShard uint64, payload []byte) error {
	buf := new(bytes.Buffer)
	params := shardsysmsg.NotifyReqParam{
		ToShard: toShard,
		Payload: payload,
	}
	if err := params.Serialize(buf); err != nil {
		return fmt.Errorf("send ping shard, marshal param: %s", err)
	}

	if _, err := native.NativeCall(utils.ShardSysMsgContractAddress, shardsysmsg.REMOTE_NOTIFY, buf.Bytes()); err != nil {
		return fmt.Errorf("send ping shard, appcallSendReq: %s", err)
	}
	return nil
}
