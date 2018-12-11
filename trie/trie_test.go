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

package trie

import (
	"testing"
	"github.com/ontio/ontology/common"
	"github.com/stretchr/testify/assert"
)

func TestInsert(t *testing.T) {
	trie, err := New(common.Uint256{}, NewMemDatabase())
	assert.Nil(t, err)
	key := []byte("test")
	value := []byte("test")

	err = trie.TryUpdate(key, value)
	assert.Nil(t, err)

	result, err := trie.TryGet(key)
	assert.Nil(t, err)

	assert.Equal(t, result, value)

	trie.Delete(key)

	result, err = trie.TryGet(key)

	assert.Equal(t, result, []byte(nil))
}
