package shardstates

import (
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
)

const (
	SHARD_STATE_CREATED    = iota
	SHARD_STATE_CONFIGURED // all parameter configured
	SHARD_STATE_READY      // all peers joined
	SHARD_STATE_ACTIVE     // started
	SHARD_STATE_ARCHIVED
)

type ShardMgmtGlobalState struct {
	NextShardID uint64 `json:"next_shard_id"`
}

func (this *ShardMgmtGlobalState) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *ShardMgmtGlobalState) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

type ShardConfig struct {
	Config               []byte         `json:"config"`
	StakeContractAddress common.Address `json:"stake_contract_address"`
}

func (this *ShardConfig) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *ShardConfig) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

type ShardState struct {
	ShardID        uint64         `json:"shard_id"`
	Creator        common.Address `json:"creator"`
	State          uint32         `json:"state"`
	Config         *ShardConfig   `json:"config"`
	PeerPubkeyList []string       `json:"peer_pubkey_list"`
}

func (this *ShardState) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *ShardState) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

