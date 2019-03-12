package shardmgmt

import (
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
	"io"
)

type SetMinStakeParam struct {
	ShardId types.ShardID `json:"shard_id"`
	Amount  uint64        `json:"amount"`
}

func (this *SetMinStakeParam) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *SetMinStakeParam) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

type ChangeMaxAuthorizationParam struct {
	ShardId          types.ShardID `json:"shard_id"`
	PeerPubKey       string        `json:"peer_pub_key"`
	MaxAuthorization uint64        `json:"max_authorization"`
}

func (this *ChangeMaxAuthorizationParam) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *ChangeMaxAuthorizationParam) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

type ChangeProportionParam struct {
	ShardId    types.ShardID `json:"shard_id"`
	PeerPubKey string        `json:"peer_pub_key"`
	Proportion uint64        `json:"proportion"`
}

func (this *ChangeProportionParam) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *ChangeProportionParam) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

type CommitDposParam struct {
	ShardId   types.ShardID      `json:"shard_id"`
	Peer      common.Address     `json:"peer"`
	FeeAmount uint64             `json:"fee_amount"`
	NewConfig *config.VBFTConfig `json:"new_config"`
	View      shardstates.View   `json:"view"`
}

func (this *CommitDposParam) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *CommitDposParam) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

type UserStakeParam struct {
	ShardId    types.ShardID  `json:"shard_id"`
	User       common.Address `json:"user"`
	PeerPubKey []string       `json:"peer_pub_key"`
	Amount     []uint64       `json:"amount"`
}

func (this *UserStakeParam) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *UserStakeParam) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}
