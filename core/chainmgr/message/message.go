
package message

import (
	"encoding/json"
	"fmt"
)

const (
	SHARD_MSG_VERSION = 1
)

const (
	HELLO_MSG = iota
	CONFIG_MSG
	BLOCK_REQ_MSG
	BLOCK_RSP_MSG
	PEERINFO_REQ_MSG
	PEERINFO_RSP_MSG
)

type RemoteShardMsg interface {
	Type() int
}

type ShardHelloMsg struct {
	TargetShardID uint64 `json:"target_shard_id"`
	SourceShardID uint64 `json:"source_shard_id"`
}

func (msg *ShardHelloMsg) Type() int {
	return HELLO_MSG
}

type ShardConfigMsg struct {
	Account []byte `json:"account"`
	Config []byte `json:"config"`
}

func (msg *ShardConfigMsg) Type() int {
	return CONFIG_MSG
}

type ShardGetGenesisBlockReqMsg struct {
	ShardID uint64 `json:"shard_id"`
}

func (msg *ShardGetGenesisBlockReqMsg) Type() int {
	return BLOCK_REQ_MSG
}

type ShardGetGenesisBlockRspMsg struct {

}

func (msg *ShardGetGenesisBlockRspMsg) Type() int {
	return BLOCK_RSP_MSG
}

type ShardGetPeerInfoReqMsg struct {

}

func (msg *ShardGetPeerInfoReqMsg) Type() int {
	return PEERINFO_REQ_MSG
}

type ShardGetPeerInfoRspMsg struct {

}

func (msg *ShardGetPeerInfoRspMsg) Type() int {
	return PEERINFO_RSP_MSG
}

func Decode(msgtype int32, msgPayload []byte) (RemoteShardMsg, error) {
	switch msgtype {
	case HELLO_MSG:
		msg := &ShardHelloMsg{}
		if err := json.Unmarshal(msgPayload, msg); err != nil {
			return nil, fmt.Errorf("unmarshal remote shard msg %d: %s", msgtype, err)
		}
		return msg, nil
	case CONFIG_MSG:
		msg := &ShardConfigMsg{}
		if err := json.Unmarshal(msgPayload, msg); err != nil {
			return nil, fmt.Errorf("unmarshal remote shard msg %d: %s", msgtype, err)
		}
		return msg, nil
	case BLOCK_REQ_MSG:
		msg := &ShardGetGenesisBlockReqMsg{}
		if err := json.Unmarshal(msgPayload, msg); err != nil {
			return nil, fmt.Errorf("unmarshal remote shard msg %d: %s", msgtype, err)
		}
		return msg, nil
	case BLOCK_RSP_MSG:
		msg := &ShardGetGenesisBlockRspMsg{}
		if err := json.Unmarshal(msgPayload, msg); err != nil {
			return nil, fmt.Errorf("unmarshal remote shard msg %d: %s", msgtype, err)
		}
		return msg, nil
	case PEERINFO_REQ_MSG:
		msg := &ShardGetPeerInfoReqMsg{}
		if err := json.Unmarshal(msgPayload, msg); err != nil {
			return nil, fmt.Errorf("unmarshal remote shard msg %d: %s", msgtype, err)
		}
		return msg, nil
	case PEERINFO_RSP_MSG:
		msg := &ShardGetPeerInfoRspMsg{}
		if err := json.Unmarshal(msgPayload, msg); err != nil {
			return nil, fmt.Errorf("unmarshal remote shard msg %d: %s", msgtype, err)
		}
		return msg, nil
	}
	return nil, fmt.Errorf("unknown remote shard msg type: %d", msgtype)
}
