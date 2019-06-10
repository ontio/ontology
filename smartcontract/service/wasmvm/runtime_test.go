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
package wasmvm

import (
	"github.com/ontio/ontology/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseEventInfo(t *testing.T) {
	states := []byte(`{"states":["transfer","AX3uShPqBpaeQoKvg2pjqN5N51zt3CZQbB","ANsQvWi9safntZfgrTmqNtSiaiBjV7t8M6",100]}`)
	contractAddress, _ := common.AddressFromBase58("APxJT5zUhFVHt4ywaP31Ro7ym9vae8AMur ")
	n := ParseEventInfo(contractAddress, states)
	n1, f := n.States.([]interface{})
	assert.True(t, f)
	assert.Equal(t, 4, len(n1))
	assert.Equal(t, "transfer", n1[0])
	assert.Equal(t, "AX3uShPqBpaeQoKvg2pjqN5N51zt3CZQbB", n1[1])
	assert.Equal(t, "ANsQvWi9safntZfgrTmqNtSiaiBjV7t8M6", n1[2])
	assert.Equal(t, 100, int(n1[3].(float64)))

	str := "some raw bytes ,not a json string"
	states = []byte(str)
	n = ParseEventInfo(contractAddress, states)
	_, f = n.States.([]interface{})
	assert.False(t, f)
	s, f := n.States.(string)
	assert.True(t, f)
	assert.Equal(t, str, s)
}
