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
package testsuite

import (
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/stretchr/testify/assert"
)

func TestShardFlow(t *testing.T) {
	contract := RandomAddress()
	method := "executeShardCommand"
	InstallNativeContract(contract, map[string]native.Handler{
		method: ExecuteShardCommandApi,
	})

	shards := make(map[common.ShardID]*ShardContext, 3)
	shard0 := common.NewShardIDUnchecked(0)
	shard1 := common.NewShardIDUnchecked(1)
	shard2 := common.NewShardIDUnchecked(2)
	shard3 := common.NewShardIDUnchecked(3)

	shards[shard0] = NewShardContext(shard0, contract, t)
	shards[shard1] = NewShardContext(shard1, contract, t)
	shards[shard2] = NewShardContext(shard2, contract, t)
	shards[shard3] = NewShardContext(shard3, contract, t)

	// shard0 -> invoke shard1
	//        -> invoke shard2 -> notify shard3
	flow := MutliCommand{}.SubCmd(
		&GreetCommand{},
	).SubCmd(
		NewInvokeCommand(shard1, &GreetCommand{}),
	).SubCmd(
		NewInvokeCommand(shard2, NewNotifyCommand(shard3, &GreetCommand{})),
	)

	totalShardMsg := RunShardTxToComplete(shards, shard0, method, EncodeShardCommandToBytes(&flow))
	// 2 req, 2 rep, 2 prep, 2 preped, 2 commit, 1 notify = 11
	assert.Equal(t, 11, totalShardMsg)
	flow = MutliCommand{}.SubCmd(
		NewNotifyCommand(shard3, &GreetCommand{}),
	).SubCmd(
		NewInvokeCommand(shard2, NewNotifyCommand(shard3, &GreetCommand{})),
	).SubCmd(
		NewInvokeCommand(shard1, &GreetCommand{}),
	)

	totalShardMsg = RunShardTxToComplete(shards, shard0, method, EncodeShardCommandToBytes(&flow))
	// 2 req, 2 rep, 2 prep, 2 preped, 2 commit, 2 notify = 12
	assert.Equal(t, 12, totalShardMsg)
}
