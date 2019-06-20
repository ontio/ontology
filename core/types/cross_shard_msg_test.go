/*
 * Copyright (C) 2018 The ontology Authors
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
package types

import (
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/stretchr/testify/assert"
)

func TestCrossShardMsgHash_Serialize(t *testing.T) {
	sigData := make(map[uint64][]byte)
	sigData[0] = []byte("123456")
	sigData[1] = []byte("345678")
	crossShardMsgHash := &CrossShardMsgHash{
		ShardID: common.NewShardIDUnchecked(0),
		MsgHash: common.Uint256{1, 2, 3},
		SigData: sigData,
	}
	sink := common.NewZeroCopySink(0)
	crossShardMsgHash.Serialization(sink)
	msg := sink.Bytes()
	var shardMsg CrossShardMsgHash
	source := common.NewZeroCopySource(msg)
	err := shardMsg.Deserialization(source)
	assert.Nil(t, err)
	assert.Equal(t, len(crossShardMsgHash.SigData), len(shardMsg.SigData))
}
