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
package common

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUint256_Serialize(t *testing.T) {
	var val Uint256
	val[1] = 245
	buf := bytes.NewBuffer(nil)
	err := val.Serialize(buf)
	assert.NotNil(t, err)
}

func TestUint256_Deserialize(t *testing.T) {
	var val Uint256
	val[1] = 245
	buf := bytes.NewBuffer(nil)
	val.Serialize(buf)

	var val2 Uint256
	val2.Deserialize(buf)

	assert.Equal(t, val, val2)

	buf = bytes.NewBuffer([]byte{1, 2, 3})
	err := val2.Deserialize(buf)

	assert.NotNil(t, err)
}

func TestUint256ParseFromBytes(t *testing.T) {
	buf := []byte{1, 2, 3}

	_, err := Uint256ParseFromBytes(buf)

	assert.NotNil(t, err)
}
