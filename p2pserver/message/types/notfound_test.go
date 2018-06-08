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

	cm "github.com/ontio/ontology/common"
)

func Uint256ParseFromBytes(f []byte) cm.Uint256 {
	if len(f) != 32 {
		return cm.Uint256{}
	}

	var hash [32]uint8
	for i := 0; i < 32; i++ {
		hash[i] = f[i]
	}
	return cm.Uint256(hash)
}

func TestNotFoundSerializationDeserialization(t *testing.T) {
	var msg NotFound
	str := "123456"
	hash := []byte(str)
	msg.Hash = Uint256ParseFromBytes(hash)

	MessageTest(t, &msg)
}
