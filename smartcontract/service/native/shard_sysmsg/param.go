package shardsysmsg

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
)

type CrossShardMsgParam struct {
	Events []*shardstates.ShardEventState
}

func (this *CrossShardMsgParam) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, uint32(len(this.Events))); err != nil {
		return fmt.Errorf("construct shardTx, write evt count: %s", err)
	}
	for _, evt := range this.Events {
		evtBytes, err := json.Marshal(evt)
		if err != nil {
			return fmt.Errorf("construct shardTx, marshal evt: %s", err)
		}
		if err := serialization.WriteVarBytes(w, evtBytes); err != nil {
			return fmt.Errorf("construct shardTx, write evt: %s", err)
		}
	}
	return nil
}

func (this *CrossShardMsgParam) Deserialize(r io.Reader) error {
	evtCnt, err := serialization.ReadUint32(r)
	if err != nil {
		return fmt.Errorf("des - CrossShardMsg: %s", err)
	}
	evts := make([]*shardstates.ShardEventState, 0)
	for i := uint32(0); i < evtCnt; i++ {
		evtBytes, err := serialization.ReadVarBytes(r)
		if err != nil {
			return fmt.Errorf("des - CrossShardMsg, read bytes: %s", err)
		}
		evt := &shardstates.ShardEventState{}
		if err := json.Unmarshal(evtBytes, evt); err != nil {
			return fmt.Errorf("des - CrossShardMsg, unmarshal: %s", err)
		}
		evts = append(evts, evt)
	}

	this.Events = evts
	return nil
}
