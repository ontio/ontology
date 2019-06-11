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
package neovm

import (
	"crypto/rand"
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/store/leveldbstore"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/ontio/ontology/smartcontract/storage"
	"github.com/stretchr/testify/assert"
)

func randomAddress() common.Address {
	var addr common.Address
	_, _ = rand.Read(addr[:])

	return addr
}

func TestCheckInvokedContract(t *testing.T) {
	memback, err := leveldbstore.NewMemLevelDBStore()
	if err != nil {
		t.Fatal(err)
	}
	overlay := overlaydb.NewOverlayDB(memback)
	cache := storage.NewCacheDB(overlay)
	// forth level meta
	forthMeta := make([]*payload.MetaDataCode, 0)
	for i := 0; i < 6; i++ {
		meta := payload.NewDefaultMetaData()
		meta.Contract = randomAddress()
		meta.AllShard = true
		forthMeta = append(forthMeta, meta)
		cache.PutMetaData(meta)
	}
	// third level meta
	thirdMeta := make([]*payload.MetaDataCode, 4)
	thirdMeta[0] = &payload.MetaDataCode{
		Contract:        randomAddress(),
		AllShard:        true,
		InvokedContract: []common.Address{forthMeta[0].Contract, forthMeta[1].Contract},
	}
	cache.PutMetaData(thirdMeta[0])
	thirdMeta[1] = &payload.MetaDataCode{
		Contract:        randomAddress(),
		AllShard:        true,
		InvokedContract: []common.Address{forthMeta[2].Contract, forthMeta[3].Contract},
	}
	cache.PutMetaData(thirdMeta[1])
	thirdMeta[2] = &payload.MetaDataCode{
		Contract:        randomAddress(),
		AllShard:        true,
		InvokedContract: []common.Address{forthMeta[4].Contract},
	}
	cache.PutMetaData(thirdMeta[2])
	thirdMeta[3] = &payload.MetaDataCode{
		Contract:        randomAddress(),
		AllShard:        true,
		InvokedContract: []common.Address{forthMeta[5].Contract},
	}
	cache.PutMetaData(thirdMeta[3])
	// second level meta
	secondMeta := make([]*payload.MetaDataCode, 2)
	secondMeta[0] = &payload.MetaDataCode{
		Contract:        randomAddress(),
		AllShard:        true,
		InvokedContract: []common.Address{thirdMeta[0].Contract, thirdMeta[1].Contract},
	}
	cache.PutMetaData(secondMeta[0])
	secondMeta[1] = &payload.MetaDataCode{
		Contract:        randomAddress(),
		AllShard:        true,
		InvokedContract: []common.Address{thirdMeta[2].Contract, thirdMeta[3].Contract},
	}
	cache.PutMetaData(secondMeta[1])
	// top level meta
	topMeta := &payload.MetaDataCode{
		Contract:        randomAddress(),
		AllShard:        true,
		InvokedContract: []common.Address{secondMeta[0].Contract, secondMeta[1].Contract},
	}
	err = CheckInvokedContract(topMeta, cache)
	if err != nil {
		t.Fatal(err)
	}
	cache.PutMetaData(topMeta)
	// test dependent recycle
	forthMeta[0].InvokedContract = []common.Address{topMeta.Contract}
	err = CheckInvokedContract(forthMeta[0], cache)
	assert.NotNil(t, err)
	t.Log(err)
}
