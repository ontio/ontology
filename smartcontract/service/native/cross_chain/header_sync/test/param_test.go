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

package test

import (
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native/cross_chain/header_sync"
	"github.com/stretchr/testify/assert"
)

func TestSyncBlockHeaderParam(t *testing.T) {
	param := header_sync.SyncBlockHeaderParam{
		Address: common.ADDRESS_EMPTY,
		Headers: [][]byte{{1}, {2}, {3}},
	}
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	var p header_sync.SyncBlockHeaderParam
	err := p.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, p, param)
}

func TestSyncGenesisHeaderParam(t *testing.T) {
	param := header_sync.SyncGenesisHeaderParam{
		GenesisHeader: []byte{1, 2, 3},
	}
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	var p header_sync.SyncGenesisHeaderParam
	err := p.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, p, param)
}
