package shardsysmsg_test

import (
	"bytes"
	"testing"

	"github.com/ontio/ontology/smartcontract/service/native/shard_sysmsg"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
)

func Test_ParamSerialize(t *testing.T) {
	param := &shardsysmsg.CrossShardMsgParam{}

	buf := new(bytes.Buffer)
	if err := param.Serialize(buf); err != nil {
		t.Fatalf("serialize param: %s", err)
	}

	param2 := &shardsysmsg.CrossShardMsgParam{}
	if err := param2.Deserialize(buf); err != nil {
		t.Fatalf("deserialize param: %s", err)
	}
}

func Test_ParamSerialize2(t *testing.T) {
	evt := &shardstates.ShardEventState{
		Version:    1,
		EventType:  2,
		ToShard:    3,
		FromHeight: 4,
		Payload:    []byte("test"),
	}
	param := &shardsysmsg.CrossShardMsgParam{
		Events: []*shardstates.ShardEventState{evt},
	}

	buf := new(bytes.Buffer)
	if err := param.Serialize(buf); err != nil {
		t.Fatalf("serialize param: %s", err)
	}

	param2 := &shardsysmsg.CrossShardMsgParam{}
	if err := param2.Deserialize(buf); err != nil {
		t.Fatalf("deserialize param: %s", err)
	}

	if len(param2.Events) != 1 {
		t.Fatalf("mismatch events %d", len(param2.Events))
	}
	evt2 := param2.Events[0]
	if evt2.Version != 1 {
		t.Fatalf("mismatch event version")
	}
}
