package shardmgmt

const (
	EVENT_SHARD_CREATE = iota
	EVENT_SHARD_CONFIG_UPDATE
	EVENT_SHARD_PEER_JOIN
	EVENT_SHARD_PEER_LEAVE
	EVENT_SHARD_DEPOSIT
)

type shardEvent interface {
	getType() uint32
}

type createShardEvent struct {
	ShardID uint64 `json:"shard_id"`
}

func (evt *createShardEvent) getType() uint32 {
	return EVENT_SHARD_CREATE
}

type configShardEvent struct {
	ShardID uint64       `json:"shard_id"`
	Config  *shardConfig `json:"config"`
}

func (evt *configShardEvent) getType() uint32 {
	return EVENT_SHARD_CONFIG_UPDATE
}

type joinShardEvent struct {
	ShardID    uint64 `json:"shard_id"`
	PeerPubKey string `json:"peer_pub_key"`
}

func (evt *joinShardEvent) getType() uint32 {
	return EVENT_SHARD_PEER_JOIN
}

type shardEventState struct {
	Version   uint32 `json:"version"`
	EventType uint32 `json:"event_type"`
	Info      []byte `json:"info"`
}
