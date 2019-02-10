package shardstates

import (
	"io"

	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
)

const (
	EVENT_SHARD_REQ_COMMON = iota + 256
)

type CommonShardReq struct {
	SourceShardID uint64 `json:"source_shard_id"`
	Height        uint64 `json:"height"`
	ShardID       uint64 `json:"shard_id"`
	Payload       []byte `json:"payload"`
}

func (evt *CommonShardReq) GetSourceShardID() uint64 {
	return evt.SourceShardID
}

func (evt *CommonShardReq) GetTargetShardID() uint64 {
	return evt.ShardID
}

func (evt *CommonShardReq) GetHeight() uint64 {
	return evt.Height
}

func (evt *CommonShardReq) GetType() uint32 {
	return EVENT_SHARD_REQ_COMMON
}

func (evt *CommonShardReq) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, evt)
}

func (evt *CommonShardReq) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, evt)
}
