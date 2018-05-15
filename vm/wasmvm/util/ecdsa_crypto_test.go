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
package util

import (
	"fmt"
	"testing"
)

func TestECDsaCrypto_Hash160(t *testing.T) {
	ecdsa := &ECDsaCrypto{}
	b := []byte("test string")
	res := ecdsa.Hash160(b)
	if len(res) != 20 {
		t.Error("TestECDsaCrypto_Hash160 length is not 20")
	}
	b = []byte(nil)
	res = ecdsa.Hash160(b)
	if len(res) != 20 {
		t.Error("TestECDsaCrypto_Hash160 length is not 20")
	}

}

func TestECDsaCrypto_Hash256(t *testing.T) {
	ecdsa := &ECDsaCrypto{}
	b := []byte("test string")
	res := ecdsa.Hash256(b)
	fmt.Println(res)
	if len(res) != 32 {
		t.Error("TestECDsaCrypto_Hash160 length is not 20")
	}
	b = []byte(nil)
	res = ecdsa.Hash256(b)
	if len(res) != 32 {
		t.Error("TestECDsaCrypto_Hash160 length is not 20")
	}
}
