
package message

import (
	"encoding/json"
	"fmt"
)

func NewShardHelloMsg(localShard, targetShard uint64) (*CrossShardMsg, error) {
	hello := &ShardHelloMsg{
		TargetShardID: targetShard,
		SourceShardID: localShard,
	}
	payload, err := json.Marshal(hello)
	if err != nil {
		return nil, fmt.Errorf("marshal hello msg: %s", err)
	}

	return &CrossShardMsg{
		Version:SHARD_MSG_VERSION,
		Type: HELLO_MSG,
		Data: payload,
	}, nil
}

func NewShardHelloAckMsg() (*CrossShardMsg, error) {
	return nil, nil
}