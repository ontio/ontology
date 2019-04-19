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
package payload

import (
	"github.com/ontio/ontology/common"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvokeCode_Serialize(t *testing.T) {
	code := InvokeCode{
		Code: []byte{1, 2, 3},
	}

	bs := common.SerializeToBytes(&code)
	var code2 InvokeCode

	err := code2.Deserialization(common.NewZeroCopySource(bs))
	assert.Nil(t, err)
	assert.Equal(t, code, code2)

	buf := common.NewZeroCopySource(bs[:len(bs)-2])
	err = code.Deserialization(buf)

	assert.NotNil(t, err)
}
