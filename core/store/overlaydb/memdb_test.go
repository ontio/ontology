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

package overlaydb

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIter(t *testing.T) {
	db := NewMemDB(0, 0)
	db.Put([]byte("aaa"), []byte("bbb"))
	iter := db.NewIterator(nil)
	assert.Equal(t, iter.First(), true)
	assert.Equal(t, iter.Last(), true)
	db.Delete([]byte("aaa"))
	assert.Equal(t, iter.First(), true)
	assert.Equal(t, len(iter.Value()), 0)
	assert.Equal(t, iter.Last(), true)
	assert.Equal(t, len(iter.Value()), 0)
}
