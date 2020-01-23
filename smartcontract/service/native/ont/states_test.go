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

package ont

import (
	"testing"

	"encoding/hex"
	"github.com/ontio/ontology/common"
	"github.com/stretchr/testify/assert"
)

func TestState_Serialize(t *testing.T) {
	state := State{
		From:  common.AddressFromVmCode([]byte{1, 2, 3}),
		To:    common.AddressFromVmCode([]byte{4, 5, 6}),
		Value: 1,
	}
	sink := common.NewZeroCopySink(nil)
	state.Serialization(sink)

	state2 := State{}
	source := common.NewZeroCopySource(sink.Bytes())
	if err := state2.Deserialization(source); err != nil {
		t.Fatal("state deserialize fail!")
	}

	assert.Equal(t, state, state2)
}

func TestLockParam_Serialize(t *testing.T) {
	fromAddr, _ := common.AddressFromBase58("709c937270e1d5a490718a2b4a230186bdd06a01")
	toAddr, _ := hex.DecodeString("709c937270e1d5a490718a2b4a230186bdd06a02")
	param := LockParam{
		ToChainID:   0,
		FromAddress: fromAddr,
		Fee:         1,
		Args: Args{
			ToAddress: toAddr,
			Value:     1,
		},
	}
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	param2 := LockParam{}
	source := common.NewZeroCopySource(sink.Bytes())
	if err := param2.Deserialization(source); err != nil {
		t.Fatal("state deserialize fail!")
	}

	assert.Equal(t, param, param2)
}
