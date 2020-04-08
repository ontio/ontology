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
package common

import (
	"encoding/json"
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/stretchr/testify/assert"
)

func TestTestEnv(t *testing.T) {
	a := TestEnv{Witness: []common.Address{common.ADDRESS_EMPTY}}

	encoded, _ := json.Marshal(&a)
	assert.Equal(t, string(encoded), `{"witness":["AFmseVrdL9f9oyCzZefL9tG6UbvhPbdYzM"]}`)

	var b TestEnv
	err := json.Unmarshal(encoded, &b)
	assert.Nil(t, err)
	assert.Equal(t, a, b)
}

func TestTestCase(t *testing.T) {
	a := TestEnv{Witness: []common.Address{common.ADDRESS_EMPTY}}
	ts := TestCase{Env: a, Method: "func1", Param: "int:100, bool:true", Expect: "int:10"}

	encoded, _ := json.Marshal(ts)

	assert.Equal(t, string(encoded), `{"env":{"witness":["AFmseVrdL9f9oyCzZefL9tG6UbvhPbdYzM"]},"needcontext":false,"method":"func1","param":"int:100, bool:true","expected":"int:10","notify":""}`)
}
