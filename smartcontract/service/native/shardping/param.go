package shardping

import (
	"io"

	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
)

type ShardPingParam struct {
	FromShard uint64 `json:"from_shard"`
	ToShard   uint64 `json:"to_shard"`
	Param     string `json:"param"`
}

func (this *ShardPingParam) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *ShardPingParam) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}
