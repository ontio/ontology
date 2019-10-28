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

package utils

import (
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

type State struct {
	From   common.Address
	To     common.Address
	Amount uint64
}

func TestBuildWasmContractParam(t *testing.T) {
	addr, _ := common.AddressFromBase58("ANT97HNwurK2LE2LEiU72MsSD684nPyJMX")
	state := State{
		From:   addr,
		To:     addr,
		Amount: 1,
	}

	param := []interface{}{"transferMulti", state, state}
	bs, _ := BuildWasmContractParam(param)
	fmt.Println("bs1:", common.ToHexString(bs))

	param = []interface{}{"transferMulti", Tuple(state, state)}
	sliBs, _ := BuildWasmContractParam(param)
	fmt.Println("bs2:", common.ToHexString(sliBs))

	assert.Equal(t, bs, sliBs)

	param = []interface{}{"testNum", []interface{}{[]interface{}{1, 2, 3}}}
	bs3, _ := BuildWasmContractParam(param)
	fmt.Println("bs3:", common.ToHexString(bs3))

	param = []interface{}{"testNum", []interface{}{[]int{1, 2, 3}}}
	bs4, _ := BuildWasmContractParam(param)
	fmt.Println("bs4:", common.ToHexString(bs4))

	assert.Equal(t, bs3, bs4)

	type ABC struct {
		A uint64
		B common.Address
		C string
	}

	param = []interface{}{"testABC", ABC{1, addr, "test"}}
	struct1, _ := BuildWasmContractParam(param)

	fmt.Println("bs6:", common.ToHexString(struct1))

	param = []interface{}{"testABC", 1, addr, "test"}
	struct2, _ := BuildWasmContractParam(param)
	fmt.Println("bs7:", common.ToHexString(struct2))
	assert.Equal(t, struct1, struct2)
}
