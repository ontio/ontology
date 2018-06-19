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
	"bytes"
	"github.com/ontio/ontology/vm/neovm/types"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestSerializeStackItemUnDeterministic(t *testing.T) {
	item := types.NewMap()
	k1 := types.NewInteger(big.NewInt(1))
	kb, _ := k1.GetByteArray()
	k2 := types.NewByteArray(kb)
	item.Add(k1, k1)
	item.Add(k2, k2)

	buf := bytes.NewBuffer(nil)
	err := SerializeStackItem(item, buf)
	assert.Nil(t, err)

	N := 1000
	failed := 0
	for i := 0; i < N; i++ {
		buf2 := bytes.NewBuffer(nil)
		err := SerializeStackItem(item, buf2)
		assert.Nil(t, err)

		if bytes.Equal(buf.Bytes(), buf2.Bytes()) == false {
			failed += 1
		}
	}

	assert.Equal(t, 0, failed)
}
