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

package shardsysmsg_test

import (
	"bytes"
	"github.com/ontio/ontology/events/message"
	"testing"

	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/shard_sysmsg"
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
	toShardID, err := types.NewShardID(3)
	if err != nil {
		t.Fatalf("invalid shard id")
	}
	evt := &message.ShardEventState{
		Version:    1,
		EventType:  2,
		ToShard:    toShardID,
		FromHeight: 4,
		Payload:    []byte("test"),
	}
	param := &shardsysmsg.CrossShardMsgParam{
		Events: []*message.ShardEventState{evt},
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
