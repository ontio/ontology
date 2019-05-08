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
package oep4

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/stretchr/testify/assert"
)

func TestOep4(t *testing.T) {
	state := &Oep4{
		Name:        "TestToken",
		Symbol:      "TT",
		Decimals:    12,
		TotalSupply: big.NewInt(1000000000),
	}
	sink := common.NewZeroCopySink(0)
	state.Serialization(sink)
	source := common.NewZeroCopySource(sink.Bytes())
	newState := &Oep4{}
	err := newState.Deserialization(source)
	assert.Nil(t, err)
}

func TestXShardTransferState(t *testing.T) {
	acc := account.NewAccount("")
	state := &XShardTransferState{
		ToShard:   common.NewShardIDUnchecked(39),
		ToAccount: acc.Address,
		Amount:    big.NewInt(384747),
		Status:    XSHARD_TRANSFER_COMPLETE,
	}
	sink := common.NewZeroCopySink(0)
	state.Serialization(sink)
	source := common.NewZeroCopySource(sink.Bytes())
	newState := &XShardTransferState{}
	err := newState.Deserialization(source)
	assert.Nil(t, err)
	data, err := json.Marshal(state)
	assert.Nil(t, err)
	t.Logf("marshal state is %s", string(data))
}
