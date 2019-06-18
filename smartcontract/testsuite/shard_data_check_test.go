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
	"github.com/ontio/ontology/common"
	common2 "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCallCounter(t *testing.T) {
	//shard0    -> invoke shard2
	flow := MutliCommand{}.SubCmd(
		NewInvokeCommand(sid(2), &CallCounterCommand{}),
	)

	shard := sid(1)
	cmd := &flow

	contract := RandomAddress()
	method := "executeShardCommand"
	InstallNativeContract(contract, map[string]native.Handler{
		method: ExecuteShardCommandApi,
	})

	shards := buildShardContexts(t, 100, contract)

	RunShardTxToComplete(shards, shard, method, EncodeShardCommandToBytes(cmd))

	key := []byte{byte(common2.ST_STORAGE)}
	key = append(key, "counter"...)

	val, err := shards[sid(2)].overlay.Get(key)
	assert.Nil(t, err)

	sink := common.NewZeroCopySink(0)
	sink.WriteUint64(1)
	assert.Equal(t, val, sink.Bytes())
}
