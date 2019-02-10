/*
 * Copyright (C) 2019 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package shardsysmsg

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
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

type NotifyReqParam struct {
	ToShard uint64 `json:"to_shard"`
	Payload []byte `json:"payload"`
}

func (this *NotifyReqParam) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *NotifyReqParam) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}
