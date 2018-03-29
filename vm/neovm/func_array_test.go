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

package neovm

import (
	"testing"
	"math/big"

	"github.com/Ontology/vm/neovm/types"
)

func TestOpArraySize(t *testing.T) {
	engine.opCode = ARRAYSIZE

	bs := []byte{0x51, 0x52}
	i := big.NewInt(1)

	is := []types.StackItemInterface{types.NewByteArray(bs), types.NewInteger(i)}
	PushData(engine, is);

	_, err := opArraySize(engine)

	if err != nil {
		t.Fatal(err)
	}

	t.Log("op array size result 2, execute result:", engine.GetEvaluationStack().Peek(0).GetStackItem().GetBigInteger())
}

func TestOpPack(t *testing.T) {
	engine.opCode = PACK

	bs := []byte{0x51, 0x52}
	i := big.NewInt(1)
	n := 2

	PushData(engine, bs)

	PushData(engine, i)

	PushData(engine, n)

	if _, err := opPack(engine); err != nil {
		t.Fatal(err)
	}
	array := engine.GetEvaluationStack().Peek(0).GetStackItem().GetArray()

	for _, v := range array {
		t.Log("value:", v.GetByteArray())
	}
}

func TestOpUnPack(t *testing.T) {
	engine.opCode = UNPACK

	if _, err := opUnpack(engine); err != nil {
		t.Fatal(err)
	}
	t.Log(engine.GetEvaluationStack().Pop().GetStackItem().GetBigInteger())
	t.Log(engine.GetEvaluationStack().Pop().GetStackItem().GetBigInteger())
	t.Log(engine.GetEvaluationStack().Pop().GetStackItem().GetByteArray())

}

func TestOpPickItem(t *testing.T) {
	engine.opCode = PICKITEM

	bs := []byte{0x51, 0x52}
	i := big.NewInt(1)

	is := []types.StackItemInterface{types.NewByteArray(bs), types.NewInteger(i)}
	PushData(engine, is)

	PushData(engine, 0)

	if _, err := opPickItem(engine); err != nil {
		t.Fatal(err)
	}
	t.Log(engine.GetEvaluationStack().Pop().GetStackItem().GetByteArray())

}


