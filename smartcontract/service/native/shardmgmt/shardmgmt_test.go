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

package shardmgmt_test

import (
	"bytes"
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/store/leveldbstore"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/ontio/ontology/smartcontract"
	"github.com/ontio/ontology/smartcontract/context"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/storage"
)

type TestSmartContract struct {
	smartcontract.SmartContract
}

func (this *TestSmartContract) CheckWitness(address common.Address) bool {
	return true
}

func TestShardMgmtInit(t *testing.T) {
	memback, _ := leveldbstore.NewMemLevelDBStore()
	overlay := overlaydb.NewOverlayDB(memback)

	cache := storage.NewCacheDB(overlay)
	contract := &TestSmartContract{}
	contract.PushContext(&context.Context{
		ContractAddress: utils.ShardMgmtContractAddress,
	})
	service := &native.NativeService{CacheDB: cache, ContextRef: contract}

	// init with empty db
	result, err := shardmgmt.ShardMgmtInit(service)
	if err != nil {
		t.Fatalf("shard mgmt init err: %s", err)
	}
	if bytes.Compare(result, utils.BYTE_TRUE) != 0 {
		t.Fatalf("shard mgmt init failed")
	}

	// init with initialized db
	result2, err := shardmgmt.ShardMgmtInit(service)
	if err != nil {
		t.Fatalf("shard mgmt init2 err: %s", err)
	}
	if bytes.Compare(result2, utils.BYTE_TRUE) != 0 {
		t.Fatalf("shard mgmt init2 failed")
	}
}
