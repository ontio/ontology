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

func sid(id int) common.ShardID {
	return common.NewShardIDUnchecked(uint64(id))
}

func buildShardContexts(t *testing.T, numShards int) map[common.ShardID]*ShardContext {
	contract := RandomAddress()
	method := "executeShardCommand"
	InstallNativeContract(contract, map[string]native.Handler{
		method: ExecuteShardCommandApi,
	})

	shards := make(map[common.ShardID]*ShardContext, numShards)
	for i := 0; i < numShards; i++ {
		id := common.NewShardIDUnchecked(uint64(i))
		shards[id] = NewShardContext(id, contract, t)
	}

	return shards
}

func runFlowCommand(t *testing.T, shard common.ShardID, cmd ShardCommand, totalMsg int) {
	contract := RandomAddress()
	method := "executeShardCommand"
	InstallNativeContract(contract, map[string]native.Handler{
		method: ExecuteShardCommandApi,
	})

	shards := buildShardContexts(t, 100)

	totalShardMsg := RunShardTxToComplete(shards, shard, method, EncodeShardCommandToBytes(cmd))
	assert.Equal(t, totalMsg, totalShardMsg)
}

func TestRecurInvoke1(t *testing.T) {
	//shard0    -> invoke shard2 -> invoke shard3
	flow := MutliCommand{}.SubCmd(
		&GreetCommand{},
	).SubCmd(
		NewInvokeCommand(sid(2), NewInvokeCommand(sid(3), &GreetCommand{})),
	)

	// 2 req, 2 rep, 2 prep, 2 preped, 2 commit = 10
	runFlowCommand(t, sid(0), &flow, 10)
}

func TestRecurInvoke2(t *testing.T) {
	//shard0 -> invoke shard1
	//       -> invoke shard2 -> invoke shard3
	flow := MutliCommand{}.SubCmd(
		&GreetCommand{},
	).SubCmd(
		NewInvokeCommand(sid(1), &GreetCommand{}),
	).SubCmd(
		NewInvokeCommand(sid(2), NewInvokeCommand(sid(3), &GreetCommand{})),
	)

	// 3 req, 3 rep, 3 prep, 3 preped, 3 commit = 15
	runFlowCommand(t, sid(0), &flow, 15)
}

func TestRecurInvoke3(t *testing.T) {
	//        / -> req shard1 \
	// shard0                  --> req shard3 -> req other shard
	//        \ -> req shard2 /
	flow := MutliCommand{}.SubCmd(
		&GreetCommand{},
	).SubCmd(
		NewInvokeCommand(sid(1), NewInvokeCommand(sid(3), &GreetCommand{})),
	).SubCmd(
		NewInvokeCommand(sid(2), NewInvokeCommand(sid(3), &GreetCommand{})),
	)

	// 4 req, 4 rep, 4 prep, 4 preped, 4 commit = 20
	runFlowCommand(t, sid(0), &flow, 20)
}

func TestShardFlowPattern1(t *testing.T) {
	// shard0 -> invoke shard1
	//        -> invoke shard2 -> notify shard3
	flow := MutliCommand{}.SubCmd(
		&GreetCommand{},
	).SubCmd(
		NewInvokeCommand(sid(1), &GreetCommand{}),
	).SubCmd(
		NewInvokeCommand(sid(2), NewNotifyCommand(sid(3), &GreetCommand{})),
	)

	// 2 req, 2 rep, 2 prep, 2 preped, 2 commit, 1 notify = 11
	runFlowCommand(t, sid(0), &flow, 11)
}

func TestShardFlowPattern2(t *testing.T) {
	// shard0 -> notify3
	//        -> invoke shard2 -> notify shard3
	// 	      -> invoke shard1
	flow := MutliCommand{}.SubCmd(
		NewNotifyCommand(sid(3), &GreetCommand{}),
	).SubCmd(
		NewInvokeCommand(sid(2), NewNotifyCommand(sid(3), &GreetCommand{})),
	).SubCmd(
		NewInvokeCommand(sid(1), &GreetCommand{}),
	)

	// 2 req, 2 rep, 2 prep, 2 preped, 2 commit, 2 notify = 12
	runFlowCommand(t, sid(0), &flow, 12)
}
