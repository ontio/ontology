package shardmgmt

import (
	"github.com/ontio/ontology/common"
	"io"
)

type CreateShardParam struct {
	ParentShardID uint64         `json:"parent_shard_id"`
	Creator       common.Address `json:"creator"`
}

func (this *CreateShardParam) Serialize(w io.Writer) error {
	return serJson(w, this)
}

func (this *CreateShardParam) Deserialize(r io.Reader) error {
	return desJson(r, this)
}

type ConfigShardParam struct {
	ShardID        uint64 `json:"shard_id"`
	ConfigTestData []byte `json:"config_test_data"`
}

func (this *ConfigShardParam) Serialize(w io.Writer) error {
	return serJson(w, this)
}

func (this *ConfigShardParam) Deserialize(r io.Reader) error {
	return desJson(r, this)
}

type JoinShardParam struct {
	ShardID    uint64         `json:"shard_id"`
	PeerOwner  common.Address `json:"peer_owner"`
	PeerPubKey string         `json:"peer_pub_key"`
	StakeAmount uint64 `json:"stake_amount"`
}

func (this *JoinShardParam) Serialize(w io.Writer) error {
	return serJson(w, this)
}

func (this *JoinShardParam) Deserialize(r io.Reader) error {
	return desJson(r, this)
}
