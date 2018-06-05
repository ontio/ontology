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
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTxAttr_Serialize_Deserialize(t *testing.T) {
	attr := NewTxAttribute(Nonce, []byte("test transaction attribute"))
	bf := new(bytes.Buffer)
	err := attr.Serialize(bf)
	assert.Nil(t, err)

	desAttr := new(TxAttribute)
	err = desAttr.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, attr, *desAttr)
}

func TestTxAttr(t *testing.T) {
	attr := NewTxAttribute(Nonce, []byte("test transaction attribute"))
	assert.True(t, IsValidAttributeType(attr.Usage))
	assert.Equal(t, attr.GetSize(), uint32(0))
	assert.NotNil(t, attr.ToArray())

	attr = NewTxAttribute(Script, []byte("test transaction attribute"))
	assert.True(t, IsValidAttributeType(attr.Usage))
	assert.Equal(t, attr.GetSize(), uint32(0))
	assert.NotNil(t, attr.ToArray())

	attr = NewTxAttribute(DescriptionUrl, []byte("test transaction attribute"))
	assert.True(t, IsValidAttributeType(attr.Usage))
	assert.Condition(t, func() (success bool) {
		return attr.GetSize() > 0
	})
	assert.NotNil(t, attr.ToArray())

	attr = NewTxAttribute(Description, []byte("test transaction attribute"))
	assert.True(t, IsValidAttributeType(attr.Usage))
	assert.Equal(t, attr.GetSize(), uint32(0))
	assert.NotNil(t, attr.ToArray())

	attr = NewTxAttribute(0x79, []byte("test transaction attribute"))
	assert.False(t, IsValidAttributeType(attr.Usage))
	assert.Equal(t, attr.GetSize(), uint32(0))
	assert.NotNil(t, attr.ToArray())
}
