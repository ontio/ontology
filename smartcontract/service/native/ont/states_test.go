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
	"bytes"
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/ontio/ontology/core/types"
)

func TestState_Serialize(t *testing.T) {
	state := State{
		From:  types.AddressFromVmCode([]byte{1, 2, 3}),
		To:    types.AddressFromVmCode([]byte{4, 5, 6}),
		Value: 1,
	}
	bf := new(bytes.Buffer)
	if err := state.Serialize(bf); err != nil {
		t.Fatal("state serialize fail!")
	}

	state2 := State{}
	if err := state2.Deserialize(bf); err != nil {
		t.Fatal("state deserialize fail!")
	}

	assert.Equal(t, state, state2)
}
