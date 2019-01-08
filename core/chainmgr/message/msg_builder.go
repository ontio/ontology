package message

import (
	"encoding/json"
	"fmt"
	"github.com/ontio/ontology-eventbus/actor"
)

func NewShardHelloMsg(localShard, targetShard uint64, sender *actor.PID) (*CrossShardMsg, error) {
	hello := &ShardHelloMsg{
		TargetShardID: targetShard,
		SourceShardID: localShard,
	}
	payload, err := json.Marshal(hello)
	if err != nil {
		return nil, fmt.Errorf("marshal hello msg: %s", err)
	}

	return &CrossShardMsg{
		Version: SHARD_MSG_VERSION,
		Type:    HELLO_MSG,
		Sender:  sender,
		Data:    payload,
	}, nil
}

func NewShardHelloAckMsg(accPayload []byte, sender *actor.PID) (*CrossShardMsg, error) {
	ack := &ShardHelloAckMsg{
		Account: accPayload,
	}
	payload, err := json.Marshal(ack)
	if err != nil {
		return nil, fmt.Errorf("marshal hello ack msg: %s", err)
	}

	return &CrossShardMsg{
		Version: SHARD_MSG_VERSION,
		Type:    HELLO_ACK_MSG,
		Sender:  sender,
		Data:    payload,
	}, nil
}
