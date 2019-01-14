package shardstates

const (
	EVENT_SHARD_CREATE = iota
	EVENT_SHARD_CONFIG_UPDATE
	EVENT_SHARD_PEER_JOIN
	EVENT_SHARD_PEER_LEAVE
	EVENT_SHARD_DEPOSIT
)

type ShardMgmtEvent interface {
	GetType() uint32
}

type CreateShardEvent struct {
	ShardID uint64 `json:"shard_id"`
}

func (evt *CreateShardEvent) GetType() uint32 {
	return EVENT_SHARD_CREATE
}

type ConfigShardEvent struct {
	ShardID uint64       `json:"shard_id"`
	Config  *ShardConfig `json:"config"`
}

func (evt *ConfigShardEvent) GetType() uint32 {
	return EVENT_SHARD_CONFIG_UPDATE
}

type JoinShardEvent struct {
	ShardID    uint64 `json:"shard_id"`
	PeerPubKey string `json:"peer_pub_key"`
}

func (evt *JoinShardEvent) GetType() uint32 {
	return EVENT_SHARD_PEER_JOIN
}

type ShardEventState struct {
	Version   uint32 `json:"version"`
	EventType uint32 `json:"event_type"`
	Info      []byte `json:"info"`
}
