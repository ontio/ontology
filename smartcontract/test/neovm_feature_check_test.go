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

	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract"
	"github.com/ontio/ontology/vm/neovm"
	"github.com/ontio/ontology/vm/neovm/errors"
	"github.com/stretchr/testify/assert"
)

func TestHeight(t *testing.T) {
	byteCode0 := []byte{
		byte(neovm.NEWMAP),
		byte(neovm.PUSH0),
		byte(neovm.HASKEY),
	}

	byteCode1 := []byte{
		byte(neovm.NEWMAP),
		byte(neovm.KEYS),
	}

	byteCode2 := []byte{
		byte(neovm.NEWMAP),
		byte(neovm.VALUES),
	}

	bytecode := [...][]byte{byteCode0, byteCode1, byteCode2}

	disableHeight := config.GetOpcodeUpdateCheckHeight(config.DefConfig.P2PNode.NetworkId)
	heights := []uint32{10, disableHeight, disableHeight + 1}

	for _, height := range heights {
		config := &smartcontract.Config{Time: 10, Height: height}
		sc := smartcontract.SmartContract{Config: config, Gas: 100}
		expected := "[NeoVmService] vm execution error!: " + errors.ERR_NOT_SUPPORT_OPCODE.Error()
		if height > disableHeight {
			expected = ""
		}
		for i := 0; i < 3; i++ {
			engine, err := sc.NewExecuteEngine(bytecode[i], types.InvokeNeo)
			assert.Nil(t, err)

			_, err = engine.Invoke()
			if len(expected) > 0 {
				assert.EqualError(t, err, expected)
			} else {
				assert.Nil(t, err)
			}
		}
	}
}
