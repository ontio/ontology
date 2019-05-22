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

package xshard

import (
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
)

func newTestShardMsg(t *testing.T) *types.CrossShardMsg {
	shardMsg := &types.CrossShardMsg{
		FromShardID:   common.NewShardIDUnchecked(0),
		MsgHeight:     uint32(90),
		SignMsgHeight: uint32(100),
	}
	return shardMsg
}

func TestCrossShardPool(t *testing.T) {
	InitCrossShardPool(common.NewShardIDUnchecked(1), 100)
	shardMsg := newTestShardMsg(t)
	if err := AddCrossShardInfo(shardMsg, nil); err != nil {
		t.Fatalf("failed add CrossShardInfo:%s", err)
	}
}
