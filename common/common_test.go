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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHexAndBytesTransfer(t *testing.T) {
	testBytes := []byte("10, 11, 12, 13, 14, 15, 16, 17, 18, 19")
	stringAfterTrans := ToHexString(testBytes)
	bytesAfterTrans, err := HexToBytes(stringAfterTrans)
	assert.Nil(t, err)
	assert.Equal(t, testBytes, bytesAfterTrans)
}

func TestGetNonce(t *testing.T) {
	nonce1 := GetNonce()
	nonce2 := GetNonce()
	assert.NotEqual(t, nonce1, nonce2)
}

func TestFileExisted(t *testing.T) {
	assert.True(t, FileExisted("common_test.go"))
	assert.True(t, FileExisted("common.go"))
	assert.False(t, FileExisted("../log/log.og"))
	assert.False(t, FileExisted("../log/log.go"))
	assert.True(t, FileExisted("./log/log.go"))
}
