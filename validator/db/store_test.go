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

package db

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStore(t *testing.T) {
	store, err := NewStore("temp.db")
	assert.Nil(t, err)

	_, err = store.GetBestBlock()
	assert.NotNil(t, err)
}

func TestTransactionMeta(t *testing.T) {
	meta := NewTransactionMeta(10, 10)

	for i := uint32(0); i < 10; i++ {
		assert.False(t, meta.IsSpent(i))
		meta.DenoteSpent(i)
	}

	assert.True(t, meta.IsFullSpent())

	for i := uint32(0); i < 10; i++ {
		assert.True(t, meta.IsSpent(i))
		meta.DenoteUnspent(i)
	}
	assert.Equal(t, meta.Height(), uint32(10))

	data := bytes.NewBuffer(nil)
	meta.Serialize(data)
	meta2 := TransactionMeta{}
	meta2.Deserialize(data)
	assert.Equal(t, meta.Height(), meta2.Height())

	for i := uint32(0); i < 10; i++ {
		assert.Equal(t, meta.IsSpent(i), meta2.IsSpent(i))
	}

}
