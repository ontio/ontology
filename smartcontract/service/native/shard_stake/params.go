package shard_stake

import (
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
	"io"
)

type UnfreezeFromShardParam struct {
	ShardId types.ShardID  `json:"shard_id"`
	Address common.Address `json:"address"`
	Amount  uint64         `json:"amount"`
}

func (this *UnfreezeFromShardParam) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *UnfreezeFromShardParam) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}
