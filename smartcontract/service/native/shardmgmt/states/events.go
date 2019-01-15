package shardstates

import (
	"io"

	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
)

const (
	EVENT_SHARD_CREATE = iota
	EVENT_SHARD_CONFIG_UPDATE
	EVENT_SHARD_PEER_JOIN
	EVENT_SHARD_ACTIVATED
	EVENT_SHARD_PEER_LEAVE
	EVENT_SHARD_DEPOSIT
)

type ShardMgmtEvent interface {
	serialization.SerializableData
	GetType() uint32
}

type CreateShardEvent struct {
	ShardID uint64 `json:"shard_id"`
}

func (evt *CreateShardEvent) GetType() uint32 {
	return EVENT_SHARD_CREATE
}

func (evt *CreateShardEvent) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, evt)
}

func (evt *CreateShardEvent) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, evt)
}

type ConfigShardEvent struct {
	ShardID uint64       `json:"shard_id"`
	Config  *ShardConfig `json:"config"`
}

func (evt *ConfigShardEvent) GetType() uint32 {
	return EVENT_SHARD_CONFIG_UPDATE
}

func (evt *ConfigShardEvent) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, evt)
}

func (evt *ConfigShardEvent) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, evt)
}

type JoinShardEvent struct {
	ShardID    uint64 `json:"shard_id"`
	PeerPubKey string `json:"peer_pub_key"`
}

func (evt *JoinShardEvent) GetType() uint32 {
	return EVENT_SHARD_PEER_JOIN
}

func (evt *JoinShardEvent) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, evt)
}

func (evt *JoinShardEvent) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, evt)
}

type ShardActiveEvent struct {
	ShardID uint64 `json:"shard_id"`
}

func (evt *ShardActiveEvent) GetType() uint32 {
	return EVENT_SHARD_ACTIVATED
}

func (evt *ShardActiveEvent) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, evt)
}

func (evt *ShardActiveEvent) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, evt)
}

type ShardEventState struct {
	Version   uint32 `json:"version"`
	EventType uint32 `json:"event_type"`
	Info      []byte `json:"info"`
}
