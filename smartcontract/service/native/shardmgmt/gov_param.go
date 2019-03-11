package shardmgmt

import (
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
	"io"
)

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
