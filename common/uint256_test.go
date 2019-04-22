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
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUint256_Serialize(t *testing.T) {
	var val Uint256
	val[1] = 245
	buf := bytes.NewBuffer(nil)
	err := val.Serialize(buf)
	assert.Nil(t, err)
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

func TestUint256Json(t *testing.T) {
	hash, err := Uint256FromHexString("5b622cfbde2948ae61242fd5d7ee1c84983459e142339316bdb6ab09faee2e02")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(hash)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("mashal hash is %s", string(data))
	newHash := &Uint256{}
	err = json.Unmarshal(data, newHash)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("unmarshal hash is %s", newHash.ToHexString())
	assert.Equal(t, hash.ToHexString(), newHash.ToHexString())
}
