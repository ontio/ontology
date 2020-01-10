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

package cross_chain_manager

import (
	"github.com/ontio/ontology/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateCrossChainTxParam(t *testing.T) {
	param := CreateCrossChainTxParam{
		ToChainID: 1,
		ToContractAddress: []byte{1, 2, 3, 4},
		Fee: 2,
		Method: "test",
		Args: []byte{1, 2, 3, 4},
	}
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	var p CreateCrossChainTxParam
	err := p.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, p, param)
}

func TestProcessCrossChainTxParam(t *testing.T) {
	param := ProcessCrossChainTxParam{
		Address: common.ADDRESS_EMPTY,
		FromChainID: 1,
		Height: 2,
		Proof: "test",
		Header: []byte{1, 2, 3, 4},
	}

	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	var p ProcessCrossChainTxParam
	err := p.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, param, p)
}

func TestOngUnlockParam(t *testing.T) {
	param := OngUnlockParam{
		FromChainID: 1,
		Address: common.ADDRESS_EMPTY,
		Amount: 1,
	}
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	var p OngUnlockParam
	err := p.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, param, p)
}