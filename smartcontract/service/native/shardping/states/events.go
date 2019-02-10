package shardping_events

import (
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
	"io"
)

type SendShardPingEvent struct {
	Payload string `json:"payload"`
}

func (this *SendShardPingEvent) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *SendShardPingEvent) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}
