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
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract"
	"github.com/ontio/ontology/vm/neovm"
	"github.com/stretchr/testify/assert"
)

func TestConvertNeoVmTypeHexString(t *testing.T) {
	code := `00c57676c8681553797374656d2e52756e74696d652e4e6f74696679`

	hex, err := common.HexToBytes(code)

	if err != nil {
		t.Fatal("hex to byte error:", err)
	}

	config := &smartcontract.Config{
		Time:   10,
		Height: 10,
		Tx:     nil,
	}
	sc := smartcontract.SmartContract{
		Config: config,
		Gas:    100000,
	}
	engine, err := sc.NewExecuteEngine(hex, types.InvokeNeo)

	_, err = engine.Invoke()

	assert.Error(t, err, "over max parameters convert length")
}

func BenchmarkExecuteAdd(b *testing.B) {
	code := []byte{byte(neovm.PUSH1)}

	N := 50000
	for i := 0; i < N; i++ {
		code = append(code, byte(neovm.PUSH1), byte(neovm.ADD))
	}
	code = append(code, byte(neovm.RET))

	config := &smartcontract.Config{
		Time:   10,
		Height: 10,
		Tx:     nil,
	}

	for i := 0; i < b.N; i++ {
		sc := smartcontract.SmartContract{
			Config: config,
			Gas:    1000000,
		}
		engine, err := sc.NewExecuteEngine(code, types.InvokeNeo)
		if err != nil {
			panic(err)
		}
		_, err = engine.Invoke()
		if err != nil {
			panic(err)
		}
	}

}
