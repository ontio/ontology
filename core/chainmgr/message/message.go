package message

import (
	"encoding/json"
	"fmt"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/service/native/shardgas/states"
	"bytes"
)

const (
	SHARD_PROTOCOL_VERSION = 1
)

const (
	HELLO_MSG = iota
	CONFIG_MSG
	BLOCK_REQ_MSG
	BLOCK_RSP_MSG
	PEERINFO_REQ_MSG
	PEERINFO_RSP_MSG

	SHARD_CONTRACT_EVENT_MSG

	DISCONNECTED_MSG
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
	Config  []byte `json:"config"`

	// peer pk : ip-addr/port, (query ip-addr from p2p)
	// genesis config
}

func (msg *ShardConfigMsg) Type() int {
	return CONFIG_MSG
}

type ShardBlockReqMsg struct {
	ShardID  uint64 `json:"shard_id"`
	BlockNum uint64 `json:"block_num"`
}

func (msg *ShardBlockReqMsg) Type() int {
	return BLOCK_REQ_MSG
}

type ShardBlockRspMsg struct {
	ShardID     uint64                         `json:"shard_id"`
	Height      uint64                         `json:"height"`
	BlockHeader *ShardBlockHeader              `json:"block_header"`
	Events      []*shardstates.ShardEventState `json:"events"`
}

func (msg *ShardBlockRspMsg) Type() int {
	return BLOCK_RSP_MSG
}

type ShardGetPeerInfoReqMsg struct {
	PeerPubKey []byte `json:"peer_pub_key"`
}

func (msg *ShardGetPeerInfoReqMsg) Type() int {
	return PEERINFO_REQ_MSG
}

type ShardGetPeerInfoRspMsg struct {
	PeerPubKey  []byte `json:"peer_pub_key"`
	PeerAddress string `json:"peer_address"`
}

func (msg *ShardGetPeerInfoRspMsg) Type() int {
	return PEERINFO_RSP_MSG
}

type ShardDisconnectedMsg struct {
	Address string `json:"address"`
}

func (msg *ShardDisconnectedMsg) Type() int {
	return DISCONNECTED_MSG
}

type ShardContractEventMsg struct {
	FromShard uint64 `json:"from_shard"`
	EventType uint32 `json:"event_type"`
	EventData []byte `json:"event_data"`
}

func (msg *ShardContractEventMsg) Type() int {
	return SHARD_CONTRACT_EVENT_MSG
}

func DecodeShardMsg(msgtype int32, msgPayload []byte) (RemoteShardMsg, error) {
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
		msg := &ShardBlockReqMsg{}
		if err := json.Unmarshal(msgPayload, msg); err != nil {
			return nil, fmt.Errorf("unmarshal remote shard msg %d: %s", msgtype, err)
		}
		return msg, nil
	case BLOCK_RSP_MSG:
		msg := &ShardBlockRspMsg{}
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

func DecodeShardEvent(evtType uint32, evtPayload []byte) (shardstates.ShardMgmtEvent, error) {
	switch evtType {
	case shardgas_states.EVENT_SHARD_GAS_DEPOSIT:
		evt := &shardgas_states.DepositGasEvent{}
		if err := evt.Deserialize(bytes.NewBuffer(evtPayload)); err != nil {
			return nil, fmt.Errorf("unmarshal remote event: %s", err)
		}
		return evt, nil
	case shardgas_states.EVENT_SHARD_GAS_WITHDRAW_REQ:
	case shardgas_states.EVENT_SHARD_GAS_WITHDRAW_DONE:
		return nil, nil
	case shardstates.EVENT_SHARD_PEER_JOIN:
	case shardstates.EVENT_SHARD_PEER_LEAVE:
		return nil, nil
	}

	return nil, fmt.Errorf("unknown remote event type: %d", evtType)
}
