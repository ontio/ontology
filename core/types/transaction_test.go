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
	"testing"

	"github.com/ontio/ontology/core/payload"
	"github.com/stretchr/testify/assert"
)

func TestTransaction_Deserialize_Hash(t *testing.T) {
	invoke := &payload.InvokeCode{Code: []byte{1, 2, 3}}
	tx := Transaction{
		TxType:  Invoke,
		Payload: invoke,
	}
	assert.Nil(t, tx.hash)
	hash := tx.Hash()

	buf := bytes.NewBuffer(nil)
	err := tx.Serialize(buf)
	assert.Nil(t, err)

	var tx2 Transaction
	err = tx2.Deserialize(buf)
	assert.Nil(t, err)
	assert.Equal(t, hash, *tx2.hash)
}
