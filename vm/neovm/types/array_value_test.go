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

package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewArray(t *testing.T) {
	a := NewArrayValue()
	for i := 0; i < 1024; i++ {
		v := VmValueFromInt64(int64(i))
		err := a.Append(v)
		assert.Equal(t, err, nil)
	}
	v := VmValueFromInt64(int64(1024))
	err := a.Append(v)
	assert.NotNil(t, err)
}

func TestArrayValue_RemoveAt(t *testing.T) {
	a := NewArrayValue()
	for i := 0; i < 10; i++ {
		v := VmValueFromInt64(int64(i))
		err := a.Append(v)
		assert.Equal(t, err, nil)
	}
	err := a.RemoveAt(-1)
	assert.NotNil(t, err)
	err = a.RemoveAt(10)
	assert.NotNil(t, err)

	assert.Equal(t, a.Len(), int64(10))
	a.RemoveAt(0)
	assert.Equal(t, a.Len(), int64(9))
}
