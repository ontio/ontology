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
	"fmt"
	"strings"
	"testing"

	"github.com/ontio/ontology/common"
	scomm "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/ong"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/stretchr/testify/assert"
)

func WrapDbPut(handler native.Handler, account common.Address) native.Handler {
	wraperHandler := func(nat *native.NativeService) ([]byte, error) {
		key := nat.ContextRef.CurrentContext().ContractAddress[:]
		nat.CacheDB.Put(key, []byte{1})
		nat.CacheDB.Put(ont.GenBalanceKey(utils.OngContractAddress, account), []byte{1, 2, 3})
		return handler(nat)
	}

	return wraperHandler
}

func TestCallLockedAddress(t *testing.T) {
	shardAContract := RandomAddress()
	method := "remoteAddAndInc"
	InstallNativeContract(shardAContract, map[string]native.Handler{
		method: RemoteInvokeAddAndInc,
	})

	shardContext := NewShardContext(common.NewShardIDUnchecked(1), shardAContract, t)
	shardContext.LockedAddress[shardAContract] = struct{}{}
	_, _, err := shardContext.InvokeShardContractRaw(method, []interface{}{""})

	assert.NotNil(t, err)
	assert.True(t, strings.Contains(fmt.Sprintf("%s", err), "contract is locked to call"))
}

func TestCallLockedKeys(t *testing.T) {
	ong.InitOng()
	account := RandomAddress()
	shardContext := NewShardContext(common.NewShardIDUnchecked(1), utils.OngContractAddress, t)
	key := ont.GenBalanceKey(utils.OngContractAddress, account)
	shardContext.LockedKeys[string(scomm.ST_STORAGE)+string(key)] = struct{}{}
	method := ont.BALANCEOF_NAME
	_, _, err := shardContext.InvokeShardContractRaw(method, []interface{}{account})
	assert.NotNil(t, err)
	t.Log(err)
}

func runLockedFlowCommand(t *testing.T, shard common.ShardID, cmd ShardCommand, totalMsg int) {
	contract := RandomAddress()
	method := "executeShardCommand"
	account := RandomAddress()
	InstallNativeContract(contract, map[string]native.Handler{
		method: WrapDbPut(ExecuteShardCommandApi, account),
	})

	shards := buildShardContexts(t, 100, contract)

	totalShardMsg := RunShardTxToComplete(shards, shard, method, EncodeShardCommandToBytes(cmd))
	assert.Equal(t, totalMsg, totalShardMsg)
	// all shard contract lock should be released
	for _, shard := range shards {
		assert.Equal(t, 0, len(shard.LockedAddress))
	}

	assert.NotEqual(t, 0, len(shards[shard].LockHistory))
	_, ok := shards[shard].LockHistory[contract]
	assert.True(t, ok)
}

func TestLockedRecurInvoke1(t *testing.T) {
	//shard0    -> invoke shard2 -> invoke shard3
	flow := MutliCommand{}.SubCmd(
		&GreetCommand{},
	).SubCmd(
		NewInvokeCommand(sid(2), NewInvokeCommand(sid(3), &GreetCommand{})),
	)

	// 2 req, 2 rep, 2 prep, 2 preped, 2 commit = 10
	runLockedFlowCommand(t, sid(1), &flow, 10)
}
