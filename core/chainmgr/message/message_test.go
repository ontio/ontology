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

package message_test

import (
	"testing"

	"github.com/ontio/ontology/core/chainmgr/message"
	"github.com/ontio/ontology/core/types"
)

func TestShardHelloMsg(t *testing.T) {
	hello := &message.ShardHelloMsg{TargetShardID: types.NewShardIDUnchecked(100),
		SourceShardID: types.NewShardIDUnchecked(200)}
	helloBytes := message.EncodeShardMsg(hello)
	msg2, err := message.DecodeShardMsg(message.HELLO_MSG, helloBytes)
	if err != nil {
		t.Fatalf("failed to decode hello: %s", err)
	}
	hello2, ok := msg2.(*message.ShardHelloMsg)
	if !ok {
		t.Fatalf("invalid hello msg type")
	}
	if hello.SourceShardID != hello2.SourceShardID {
		t.Fatalf("hello: invalid source shard id")
	}
	if hello.TargetShardID != hello2.TargetShardID {
		t.Fatalf("hello: invalid target shard id")
	}
}

func TestShardBlockRspMsg(t *testing.T) {
	blkHdr := newTestBlockHdr()
	tx := newTestShardTx(t, 1, 1000)
	rsp := &message.ShardBlockRspMsg{
		FromShardID: types.NewShardIDUnchecked(100),
		Height:      200,
		BlockHeader: blkHdr,
		Txs:         []*message.ShardBlockTx{tx},
	}

	msgBytes := message.EncodeShardMsg(rsp)

	msg2, err := message.DecodeShardMsg(message.BLOCK_RSP_MSG, msgBytes)
	if err != nil {
		t.Fatalf("failed to decode rsp msg: %s", err)
	}
	rsp2, ok := msg2.(*message.ShardBlockRspMsg)
	if !ok {
		t.Fatalf("invalid rsp msg type")
	}

	if rsp2.FromShardID != rsp.FromShardID {
		t.Fatalf("invalid from shard id")
	}
	if rsp2.Height != rsp.Height {
		t.Fatalf("invalid height")
	}
}
